// ============================================================
// FILE     : internal/ingest/service_test.go
// PURPOSE  : Verifies sync_turn idempotency, validation, and job enqueue behavior.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : TestServiceSyncTurn_EnqueuesJobForNewEvents, TestServiceSyncTurn_ReplayedTurnIsDuplicate
// DEPENDS  : internal/ingest, internal/core
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: Keep these tests focused on the API hot path; reasoning belongs to worker tests.
// ============================================================

package ingest

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

func TestServiceSyncTurn_EnqueuesJobForNewEvents(t *testing.T) {
	t.Parallel()

	rawEvents := newFakeRawEventStore()
	jobs := &fakeJobStore{}
	service := newTestService(t, rawEvents, jobs)

	resp, err := service.SyncTurn(context.Background(), testSyncTurnRequest())
	if err != nil {
		t.Fatalf("SyncTurn returned error: %v", err)
	}

	if resp.Status != "accepted" {
		t.Fatalf("unexpected status: %s", resp.Status)
	}
	if len(resp.EventIDs) != 2 {
		t.Fatalf("expected 2 event IDs, got %d", len(resp.EventIDs))
	}
	if len(resp.JobIDs) != 1 {
		t.Fatalf("expected 1 job ID, got %d", len(resp.JobIDs))
	}
	if resp.DuplicateCount != 0 {
		t.Fatalf("expected duplicate count 0, got %d", resp.DuplicateCount)
	}
	if len(rawEvents.events) != 2 {
		t.Fatalf("expected 2 stored events, got %d", len(rawEvents.events))
	}
	if len(jobs.jobs) != 1 {
		t.Fatalf("expected 1 stored job, got %d", len(jobs.jobs))
	}
	if jobs.jobs[0].JobKind != core.JobKindProcessTurnEvent {
		t.Fatalf("unexpected job kind: %s", jobs.jobs[0].JobKind)
	}
	if len(jobs.jobs[0].RawEventIDs) != 2 {
		t.Fatalf("expected job to reference 2 events, got %d", len(jobs.jobs[0].RawEventIDs))
	}
}

func TestServiceSyncTurn_ReplayedTurnIsDuplicate(t *testing.T) {
	t.Parallel()

	rawEvents := newFakeRawEventStore()
	jobs := &fakeJobStore{}
	service := newTestService(t, rawEvents, jobs)
	req := testSyncTurnRequest()

	if _, err := service.SyncTurn(context.Background(), req); err != nil {
		t.Fatalf("first SyncTurn returned error: %v", err)
	}
	resp, err := service.SyncTurn(context.Background(), req)
	if err != nil {
		t.Fatalf("second SyncTurn returned error: %v", err)
	}

	if len(resp.EventIDs) != 0 {
		t.Fatalf("expected duplicate replay to return no new events, got %d", len(resp.EventIDs))
	}
	if len(resp.JobIDs) != 0 {
		t.Fatalf("expected duplicate replay to enqueue no jobs, got %d", len(resp.JobIDs))
	}
	if resp.DuplicateCount != 2 {
		t.Fatalf("expected duplicate count 2, got %d", resp.DuplicateCount)
	}
	if len(jobs.jobs) != 1 {
		t.Fatalf("expected only the first call to enqueue a job, got %d jobs", len(jobs.jobs))
	}
}

func TestServiceSyncTurn_ValidatesRequiredFields(t *testing.T) {
	t.Parallel()

	service := newTestService(t, newFakeRawEventStore(), &fakeJobStore{})
	req := testSyncTurnRequest()
	req.TenantID = ""

	_, err := service.SyncTurn(context.Background(), req)
	if !errors.Is(err, core.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}

func newTestService(t *testing.T, rawEvents *fakeRawEventStore, jobs *fakeJobStore) *Service {
	t.Helper()
	service, err := NewService(Dependencies{
		RawEvents: rawEvents,
		Jobs:      jobs,
		Clock: func() time.Time {
			return time.Date(2026, time.April, 24, 0, 0, 0, 0, time.UTC)
		},
	})
	if err != nil {
		t.Fatalf("NewService returned error: %v", err)
	}
	return service
}

func testSyncTurnRequest() *core.SyncTurnRequest {
	return &core.SyncTurnRequest{
		TenantID:       "tenant_1",
		WorkspaceID:    "workspace_1",
		SessionID:      "session_1",
		ActorID:        "agent:hermes-main",
		IdempotencyKey: "turn_1",
		TurnEvents: []core.RawEventPayload{
			{
				EventKind:   "user_message",
				Source:      "hermes",
				Fingerprint: "fp_user",
				OccurredAt:  time.Date(2026, time.April, 24, 0, 1, 0, 0, time.UTC),
				PayloadJSON: json.RawMessage(`{"text":"Remember the plan."}`),
			},
			{
				EventKind:   "assistant_message",
				Source:      "hermes",
				Fingerprint: "fp_assistant",
				OccurredAt:  time.Date(2026, time.April, 24, 0, 2, 0, 0, time.UTC),
				PayloadJSON: json.RawMessage(`{"text":"Acknowledged."}`),
			},
		},
	}
}

type fakeRawEventStore struct {
	seen   map[string]struct{}
	events []*core.RawEvent
}

func newFakeRawEventStore() *fakeRawEventStore {
	return &fakeRawEventStore{
		seen: make(map[string]struct{}),
	}
}

func (s *fakeRawEventStore) AppendRawEvents(_ context.Context, events []*core.RawEvent) ([]string, error) {
	ids := make([]string, 0, len(events))
	for _, event := range events {
		key := event.TenantID + "\x00" + event.Source + "\x00" + event.IdempotencyKey
		if _, ok := s.seen[key]; ok {
			continue
		}
		s.seen[key] = struct{}{}
		s.events = append(s.events, event)
		ids = append(ids, event.ID)
	}
	return ids, nil
}

func (s *fakeRawEventStore) GetRawEvents(_ context.Context, _ []string) ([]*core.RawEvent, error) {
	return nil, core.ErrNotImplemented
}

type fakeJobStore struct {
	jobs []*core.IngestJob
}

func (s *fakeJobStore) EnqueueJobs(_ context.Context, jobs []*core.IngestJob) ([]string, error) {
	ids := make([]string, 0, len(jobs))
	for _, job := range jobs {
		s.jobs = append(s.jobs, job)
		ids = append(ids, job.ID)
	}
	return ids, nil
}

func (s *fakeJobStore) ClaimJobs(context.Context, string, int) ([]*core.IngestJob, error) {
	return nil, core.ErrNotImplemented
}

func (s *fakeJobStore) CompleteJob(context.Context, string) error {
	return core.ErrNotImplemented
}

func (s *fakeJobStore) FailJob(context.Context, string, error) error {
	return core.ErrNotImplemented
}

func (s *fakeJobStore) BlockJob(context.Context, string, error) error {
	return core.ErrNotImplemented
}
