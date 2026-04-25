// ============================================================
// FILE     : internal/worker/processor_test.go
// PURPOSE  : Verifies worker job claiming, dispatch, completion, and retry handoff.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : TestProcessorRunOnce_ClaimsOneJob, TestProcessorRunOnce_UnknownJobKindFailsSafely, worker processor tests
// DEPENDS  : internal/worker, internal/core, internal/graph, internal/reasoning
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: These tests lock the pipeline skeleton only; they must not assert memory extraction quality.
// ============================================================

package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/graph"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/reasoning"
)

func TestProcessorRunOnce_ClaimsOneJob(t *testing.T) {
	t.Parallel()

	job := testProcessTurnJob()
	jobs := &fakeWorkerJobStore{claimed: []*core.IngestJob{job}}
	rawEvents := fakeRawEventsForJob(job)
	processor := newTestProcessor(t, jobs, rawEvents, &fakeReasoner{}, &fakeApplyEngine{})

	result, err := processor.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce returned error: %v", err)
	}

	if jobs.claimWorkerID != "worker:test" {
		t.Fatalf("unexpected worker ID: %s", jobs.claimWorkerID)
	}
	if jobs.claimLimit != 1 {
		t.Fatalf("expected default claim limit 1, got %d", jobs.claimLimit)
	}
	if result.Claimed != 1 || result.Completed != 1 || result.Failed != 0 {
		t.Fatalf("unexpected run result: %#v", result)
	}
}

func TestProcessorRunOnce_UnknownJobKindFailsSafely(t *testing.T) {
	t.Parallel()

	job := testProcessTurnJob()
	job.JobKind = core.JobKind("unknown_job")
	jobs := &fakeWorkerJobStore{claimed: []*core.IngestJob{job}}
	rawEvents := fakeRawEventsForJob(job)
	processor := newTestProcessor(t, jobs, rawEvents, &fakeReasoner{}, &fakeApplyEngine{})

	result, err := processor.RunOnce(context.Background())
	if err == nil {
		t.Fatalf("expected RunOnce to report unsupported job kind")
	}

	if result.Claimed != 1 || result.Completed != 0 || result.Failed != 1 {
		t.Fatalf("unexpected run result: %#v", result)
	}
	if len(jobs.completed) != 0 {
		t.Fatalf("expected no completed jobs, got %v", jobs.completed)
	}
	if len(jobs.failed) != 1 {
		t.Fatalf("expected one failed job, got %d", len(jobs.failed))
	}
	if !strings.Contains(jobs.failed[0].err.Error(), "unsupported job kind") {
		t.Fatalf("unexpected failure error: %v", jobs.failed[0].err)
	}
	if len(rawEvents.requestedIDs) != 0 {
		t.Fatalf("unknown job kind should not load raw events, got %v", rawEvents.requestedIDs)
	}
}

func TestProcessorProcessTurnEvent_LoadsRawEvents(t *testing.T) {
	t.Parallel()

	job := testProcessTurnJob()
	jobs := &fakeWorkerJobStore{claimed: []*core.IngestJob{job}}
	rawEvents := fakeRawEventsForJob(job)
	reasoner := &fakeReasoner{}
	processor := newTestProcessor(t, jobs, rawEvents, reasoner, &fakeApplyEngine{})

	if _, err := processor.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce returned error: %v", err)
	}

	if !reflect.DeepEqual(rawEvents.requestedIDs, job.RawEventIDs) {
		t.Fatalf("unexpected raw event IDs: got %v want %v", rawEvents.requestedIDs, job.RawEventIDs)
	}
	if reasoner.received == nil {
		t.Fatalf("expected reasoner to receive an envelope")
	}
	if len(reasoner.received.RawEvents) != len(job.RawEventIDs) {
		t.Fatalf("expected %d raw events, got %d", len(job.RawEventIDs), len(reasoner.received.RawEvents))
	}
	if reasoner.received.Stage2.RequiredOutputName != reasoning.StageNameResolve {
		t.Fatalf("unexpected stage 2 output contract: %s", reasoner.received.Stage2.RequiredOutputName)
	}
	if reasoner.received.Stage2.RequiredOutputSchema != reasoning.Stage2ResolveOutputSchemaV0 {
		t.Fatalf("unexpected stage 2 output schema: %s", reasoner.received.Stage2.RequiredOutputSchema)
	}
}

func TestProcessorProcessTurnEvent_PassesStageEnvelopeToReasoner(t *testing.T) {
	t.Parallel()

	job := testProcessTurnJob()
	jobs := &fakeWorkerJobStore{claimed: []*core.IngestJob{job}}
	rawEvents := fakeRawEventsForJob(job)
	reasoner := &fakeReasoner{}
	processor := newTestProcessor(t, jobs, rawEvents, reasoner, &fakeApplyEngine{})

	if _, err := processor.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce returned error: %v", err)
	}

	if reasoner.received == nil {
		t.Fatalf("expected reasoner to receive an envelope")
	}
	if reasoner.received.Stage1.JobID != job.ID {
		t.Fatalf("expected Stage 1 envelope for job %s, got %#v", job.ID, reasoner.received.Stage1)
	}
	stage2 := reasoner.received.Stage2
	if stage2.RequiredOutputSchema != reasoning.Stage2ResolveOutputSchemaV0 {
		t.Fatalf("expected required output schema %q, got %q", reasoning.Stage2ResolveOutputSchemaV0, stage2.RequiredOutputSchema)
	}
	if len(stage2.RelevantMemories) != 0 {
		t.Fatalf("worker should leave retrieved Stage 2 context to the reasoner, got %#v", stage2.RelevantMemories)
	}
}

func TestProcessorRunOnce_CompletesJobOnSuccess(t *testing.T) {
	t.Parallel()

	job := testProcessTurnJob()
	jobs := &fakeWorkerJobStore{claimed: []*core.IngestJob{job}}
	applyEngine := &fakeApplyEngine{}
	processor := newTestProcessor(t, jobs, fakeRawEventsForJob(job), &fakeReasoner{}, applyEngine)

	result, err := processor.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce returned error: %v", err)
	}

	if !reflect.DeepEqual(jobs.completed, []string{job.ID}) {
		t.Fatalf("unexpected completed jobs: %v", jobs.completed)
	}
	if len(jobs.failed) != 0 {
		t.Fatalf("expected no failed jobs, got %d", len(jobs.failed))
	}
	if applyEngine.received == nil || applyEngine.received.JobID != job.ID {
		t.Fatalf("expected apply request for job %s, got %#v", job.ID, applyEngine.received)
	}
	if result.Completed != 1 {
		t.Fatalf("expected one completed job, got %#v", result)
	}
}

func TestProcessorRunOnce_FailsJobOnReasoningOrApplyError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		reasoner  reasoning.Orchestrator
		apply     graph.ApplyEngine
		wantError string
	}{
		{
			name:      "reasoning",
			reasoner:  &fakeReasoner{err: errors.New("codex bridge unavailable")},
			apply:     &fakeApplyEngine{},
			wantError: "reason process_turn_event",
		},
		{
			name:      "apply",
			reasoner:  &fakeReasoner{},
			apply:     &fakeApplyEngine{err: errors.New("apply validation failed")},
			wantError: "apply process_turn_event",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			job := testProcessTurnJob()
			jobs := &fakeWorkerJobStore{claimed: []*core.IngestJob{job}}
			processor := newTestProcessor(t, jobs, fakeRawEventsForJob(job), tt.reasoner, tt.apply)

			result, err := processor.RunOnce(context.Background())
			if err == nil {
				t.Fatalf("expected RunOnce to return an error")
			}

			if !strings.Contains(err.Error(), tt.wantError) {
				t.Fatalf("expected error to contain %q, got %v", tt.wantError, err)
			}
			if result.Claimed != 1 || result.Completed != 0 || result.Failed != 1 {
				t.Fatalf("unexpected run result: %#v", result)
			}
			if len(jobs.completed) != 0 {
				t.Fatalf("expected no completed jobs, got %v", jobs.completed)
			}
			if len(jobs.failed) != 1 || jobs.failed[0].jobID != job.ID {
				t.Fatalf("unexpected failed jobs: %#v", jobs.failed)
			}
		})
	}
}

func TestProcessorRunOnce_BlocksJobWhenApplyNotImplemented(t *testing.T) {
	t.Parallel()

	job := testProcessTurnJob()
	jobs := &fakeWorkerJobStore{claimed: []*core.IngestJob{job}}
	applyErr := fmt.Errorf("%w: operations[0].kind %q is validation-only in store-backed apply", core.ErrNotImplemented, reasoning.OperationKindUpdateMemory)
	applyEngine := &fakeApplyEngine{err: applyErr}
	processor := newTestProcessor(t, jobs, fakeRawEventsForJob(job), &fakeReasoner{}, applyEngine)

	result, err := processor.RunOnce(context.Background())
	if err == nil {
		t.Fatalf("expected RunOnce to return an error")
	}
	if !errors.Is(err, core.ErrNotImplemented) {
		t.Fatalf("expected RunOnce error to wrap ErrNotImplemented, got %v", err)
	}
	if len(jobs.completed) != 0 {
		t.Fatalf("expected unsupported apply work not to complete the job, got %v", jobs.completed)
	}
	if len(jobs.failed) != 0 {
		t.Fatalf("expected unsupported apply work not to use retry failure path, got %#v", jobs.failed)
	}
	if len(jobs.blocked) != 1 || !errors.Is(jobs.blocked[0].err, core.ErrNotImplemented) {
		t.Fatalf("expected blocked job to record ErrNotImplemented, got %#v", jobs.blocked)
	}
	if !strings.Contains(jobs.blocked[0].err.Error(), "unsupported apply work") {
		t.Fatalf("expected blocked record to explain unsupported apply work, got %v", jobs.blocked[0].err)
	}
	if result.Claimed != 1 || result.Completed != 0 || result.Failed != 0 || result.Blocked != 1 {
		t.Fatalf("unexpected run result: %#v", result)
	}
	if len(result.Failures) != 1 || result.Failures[0].JobID != job.ID {
		t.Fatalf("expected per-job failure report for %s, got %#v", job.ID, result.Failures)
	}
	if !strings.Contains(result.Failures[0].Error, "unsupported apply work") {
		t.Fatalf("expected failure report to include unsupported apply work, got %#v", result.Failures[0])
	}
}

func TestProcessorRunOnce_RejectsIncompleteRawEventBundleBeforeReasoning(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		mutate      func(*fakeWorkerRawEventStore)
		wantWrapped error
	}{
		{
			name: "missing requested event",
			mutate: func(rawEvents *fakeWorkerRawEventStore) {
				delete(rawEvents.events, "evt_2")
			},
			wantWrapped: core.ErrNotFound,
		},
		{
			name: "returned event id mismatch",
			mutate: func(rawEvents *fakeWorkerRawEventStore) {
				rawEvents.events["evt_2"].ID = "evt_unrequested"
			},
			wantWrapped: core.ErrConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			job := testProcessTurnJob()
			jobs := &fakeWorkerJobStore{claimed: []*core.IngestJob{job}}
			rawEvents := fakeRawEventsForJob(job)
			tt.mutate(rawEvents)
			reasoner := &fakeReasoner{}
			applyEngine := &fakeApplyEngine{}
			processor := newTestProcessor(t, jobs, rawEvents, reasoner, applyEngine)

			result, err := processor.RunOnce(context.Background())
			if err == nil {
				t.Fatalf("expected RunOnce to return an error")
			}
			if !errors.Is(err, tt.wantWrapped) {
				t.Fatalf("expected error to wrap %v, got %v", tt.wantWrapped, err)
			}
			if !strings.Contains(err.Error(), "raw event bundle") {
				t.Fatalf("expected raw event bundle error, got %v", err)
			}
			if reasoner.received != nil {
				t.Fatalf("reasoner should not run for incomplete raw event bundle")
			}
			if applyEngine.received != nil {
				t.Fatalf("apply should not run for incomplete raw event bundle")
			}
			if len(jobs.completed) != 0 {
				t.Fatalf("expected incomplete raw event bundle not to complete the job, got %v", jobs.completed)
			}
			if len(jobs.failed) != 1 {
				t.Fatalf("expected one failed job, got %#v", jobs.failed)
			}
			if result.Claimed != 1 || result.Completed != 0 || result.Failed != 1 {
				t.Fatalf("unexpected run result: %#v", result)
			}
		})
	}
}

func TestProcessorRunOnce_RejectsAmbiguousRawEventActorBeforeStage2Sources(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		mutate      func(*fakeWorkerRawEventStore)
		wantWrapped error
		wantError   string
	}{
		{
			name: "empty actor",
			mutate: func(rawEvents *fakeWorkerRawEventStore) {
				rawEvents.events["evt_2"].ActorID = ""
			},
			wantWrapped: core.ErrInvalidArgument,
			wantError:   "empty actor_id",
		},
		{
			name: "mixed actors",
			mutate: func(rawEvents *fakeWorkerRawEventStore) {
				rawEvents.events["evt_2"].ActorID = "agent:other"
			},
			wantWrapped: core.ErrConflict,
			wantError:   "mixed actor_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			job := testProcessTurnJob()
			jobs := &fakeWorkerJobStore{claimed: []*core.IngestJob{job}}
			rawEvents := fakeRawEventsForJob(job)
			tt.mutate(rawEvents)
			reasoner := &fakeReasoner{}
			applyEngine := &fakeApplyEngine{}
			processor := newTestProcessor(t, jobs, rawEvents, reasoner, applyEngine)

			result, err := processor.RunOnce(context.Background())
			if err == nil {
				t.Fatalf("expected RunOnce to reject ambiguous raw event actors")
			}
			if !errors.Is(err, tt.wantWrapped) {
				t.Fatalf("expected error to wrap %v, got %v", tt.wantWrapped, err)
			}
			if !strings.Contains(err.Error(), tt.wantError) {
				t.Fatalf("expected error to contain %q, got %v", tt.wantError, err)
			}
			if reasoner.received != nil {
				t.Fatalf("reasoner should not run for ambiguous raw event actors")
			}
			if applyEngine.received != nil {
				t.Fatalf("apply should not run for ambiguous raw event actors")
			}
			if len(jobs.completed) != 0 {
				t.Fatalf("expected ambiguous actor bundle not to complete the job, got %v", jobs.completed)
			}
			if len(jobs.failed) != 1 {
				t.Fatalf("expected one failed job, got %#v", jobs.failed)
			}
			if result.Claimed != 1 || result.Completed != 0 || result.Failed != 1 {
				t.Fatalf("unexpected run result: %#v", result)
			}
		})
	}
}

func TestProcessorRunOnce_ReportsAppliedOperationCounts(t *testing.T) {
	t.Parallel()

	job := testProcessTurnJob()
	jobs := &fakeWorkerJobStore{claimed: []*core.IngestJob{job}}
	applyEngine := &fakeApplyEngine{result: &graph.ApplyResult{
		AppliedOperationCount: 2,
		MemoryIDs:             []string{"mem_1", "mem_2"},
		TraceWritten:          true,
	}}
	processor := newTestProcessor(t, jobs, fakeRawEventsForJob(job), &fakeReasoner{}, applyEngine)

	result, err := processor.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce returned error: %v", err)
	}

	if result.AppliedOperationCount != 2 {
		t.Fatalf("unexpected applied operation count: got %d want 2", result.AppliedOperationCount)
	}
	if result.MemoryIDCount != 2 {
		t.Fatalf("unexpected memory id count: got %d want 2", result.MemoryIDCount)
	}
	if result.TraceWrittenCount != 1 {
		t.Fatalf("unexpected trace written count: got %d want 1", result.TraceWrittenCount)
	}
}

func TestProcessorRunOnce_ProcessesDreamSessionJob(t *testing.T) {
	t.Parallel()

	job := testProcessTurnJob()
	job.JobKind = core.JobKindDreamSession
	job.RawEventIDs = nil
	job.PayloadJSON = json.RawMessage(`{"session_id":"session_1"}`)
	jobs := &fakeWorkerJobStore{claimed: []*core.IngestJob{job}}
	dreamingStore := &fakeDreamingStore{
		input: &core.DreamingSessionInput{
			RawEventIDs: []string{"evt_1"},
			Memories: []*core.Memory{{
				ID:         "mem_1",
				Text:       "User prefers contract-first implementation.",
				Confidence: 0.9,
			}},
		},
		promotions: map[core.DreamingTier]*core.DreamingPromotionResult{
			core.DreamingTierMidTerm:       {PromotedCount: 1, MemoryIDs: []string{"mem_1"}},
			core.DreamingTierLongTerm:      {PromotedCount: 1, MemoryIDs: []string{"mem_1"}},
			core.DreamingTierUltraLongTerm: {PromotedCount: 0},
		},
	}
	processor := newTestProcessorWithDreaming(t, jobs, fakeRawEventsForJob(testProcessTurnJob()), &fakeReasoner{}, &fakeApplyEngine{}, dreamingStore)

	result, err := processor.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce returned error: %v", err)
	}

	if result.Completed != 1 || result.SessionDreamCount != 2 {
		t.Fatalf("unexpected dream session result: %#v", result)
	}
	if dreamingStore.summary == nil || dreamingStore.summary.SessionID != "session_1" {
		t.Fatalf("expected session summary to be written, got %#v", dreamingStore.summary)
	}
	if len(dreamingStore.requests) == 0 || dreamingStore.requests[0].Tier != core.DreamingTierMidTerm {
		t.Fatalf("expected mid-term promotion request, got %#v", dreamingStore.requests)
	}
}

func TestProcessorRunOnce_ProcessesDreamWorkspaceJob(t *testing.T) {
	t.Parallel()

	job := testProcessTurnJob()
	job.JobKind = core.JobKindDreamWorkspace
	job.RawEventIDs = nil
	job.PayloadJSON = json.RawMessage(`{}`)
	jobs := &fakeWorkerJobStore{claimed: []*core.IngestJob{job}}
	dreamingStore := &fakeDreamingStore{
		promotions: map[core.DreamingTier]*core.DreamingPromotionResult{
			core.DreamingTierLongTerm:      {PromotedCount: 2, MemoryIDs: []string{"mem_1", "mem_2"}},
			core.DreamingTierUltraLongTerm: {PromotedCount: 1, MemoryIDs: []string{"mem_3"}},
		},
	}
	processor := newTestProcessorWithDreaming(t, jobs, fakeRawEventsForJob(testProcessTurnJob()), &fakeReasoner{}, &fakeApplyEngine{}, dreamingStore)

	result, err := processor.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce returned error: %v", err)
	}

	if result.Completed != 1 || result.WorkspaceDreamCount != 3 {
		t.Fatalf("unexpected dream workspace result: %#v", result)
	}
	if len(dreamingStore.requests) != 2 {
		t.Fatalf("expected long and ultra-long promotion requests, got %#v", dreamingStore.requests)
	}
}

func newTestProcessor(
	t *testing.T,
	jobs *fakeWorkerJobStore,
	rawEvents *fakeWorkerRawEventStore,
	reasoner reasoning.Orchestrator,
	applyEngine graph.ApplyEngine,
) *Processor {
	t.Helper()
	processor, err := NewProcessor(Dependencies{
		WorkerID:    "worker:test",
		Jobs:        jobs,
		RawEvents:   rawEvents,
		Reasoner:    reasoner,
		ApplyEngine: applyEngine,
		Dreaming:    newTestDreamingService(t, &fakeDreamingStore{}),
	})
	if err != nil {
		t.Fatalf("NewProcessor returned error: %v", err)
	}
	return processor
}

func newTestProcessorWithDreaming(
	t *testing.T,
	jobs *fakeWorkerJobStore,
	rawEvents *fakeWorkerRawEventStore,
	reasoner reasoning.Orchestrator,
	applyEngine graph.ApplyEngine,
	dreamingStore *fakeDreamingStore,
) *Processor {
	t.Helper()
	processor, err := NewProcessor(Dependencies{
		WorkerID:    "worker:test",
		Jobs:        jobs,
		RawEvents:   rawEvents,
		Reasoner:    reasoner,
		ApplyEngine: applyEngine,
		Dreaming:    newTestDreamingService(t, dreamingStore),
		Clock: func() time.Time {
			return time.Date(2026, time.April, 24, 1, 0, 0, 0, time.UTC)
		},
	})
	if err != nil {
		t.Fatalf("NewProcessor returned error: %v", err)
	}
	return processor
}

func newTestDreamingService(t *testing.T, store *fakeDreamingStore) *graph.DreamingService {
	t.Helper()
	service, err := graph.NewDreamingService(graph.DreamingDependencies{
		Store: store,
		Clock: func() time.Time {
			return time.Date(2026, time.April, 24, 1, 0, 0, 0, time.UTC)
		},
	})
	if err != nil {
		t.Fatalf("NewDreamingService returned error: %v", err)
	}
	return service
}

func testProcessTurnJob() *core.IngestJob {
	return &core.IngestJob{
		ID:          "job_1",
		TenantID:    "tenant_1",
		WorkspaceID: "workspace_1",
		JobKind:     core.JobKindProcessTurnEvent,
		Status:      "running",
		RawEventIDs: []string{"evt_1", "evt_2"},
		PayloadJSON: json.RawMessage(`{"session_id":"session_1"}`),
		CreatedAt:   time.Date(2026, time.April, 24, 0, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, time.April, 24, 0, 0, 0, 0, time.UTC),
	}
}

func fakeRawEventsForJob(job *core.IngestJob) *fakeWorkerRawEventStore {
	events := make(map[string]*core.RawEvent, len(job.RawEventIDs))
	for i, eventID := range job.RawEventIDs {
		events[eventID] = &core.RawEvent{
			ID:          eventID,
			TenantID:    job.TenantID,
			WorkspaceID: job.WorkspaceID,
			SessionID:   "session_1",
			ActorID:     "agent:hermes-main",
			EventKind:   "message",
			Source:      "hermes",
			Fingerprint: "fp_" + eventID,
			OccurredAt:  time.Date(2026, time.April, 24, 0, i, 0, 0, time.UTC),
			PayloadJSON: json.RawMessage(`{"text":"hello"}`),
			CreatedAt:   time.Date(2026, time.April, 24, 0, i, 0, 0, time.UTC),
		}
	}
	return &fakeWorkerRawEventStore{events: events}
}

type failedJob struct {
	jobID string
	err   error
}

type fakeWorkerJobStore struct {
	claimed       []*core.IngestJob
	claimWorkerID string
	claimLimit    int
	completed     []string
	failed        []failedJob
	blocked       []failedJob
}

func (s *fakeWorkerJobStore) EnqueueJobs(context.Context, []*core.IngestJob) ([]string, error) {
	return nil, core.ErrNotImplemented
}

func (s *fakeWorkerJobStore) ClaimJobs(_ context.Context, workerID string, limit int) ([]*core.IngestJob, error) {
	s.claimWorkerID = workerID
	s.claimLimit = limit
	return s.claimed, nil
}

func (s *fakeWorkerJobStore) CompleteJob(_ context.Context, jobID string) error {
	s.completed = append(s.completed, jobID)
	return nil
}

func (s *fakeWorkerJobStore) FailJob(_ context.Context, jobID string, err error) error {
	s.failed = append(s.failed, failedJob{
		jobID: jobID,
		err:   err,
	})
	return nil
}

func (s *fakeWorkerJobStore) BlockJob(_ context.Context, jobID string, err error) error {
	s.blocked = append(s.blocked, failedJob{
		jobID: jobID,
		err:   err,
	})
	return nil
}

type fakeWorkerRawEventStore struct {
	events       map[string]*core.RawEvent
	requestedIDs []string
	err          error
}

func (s *fakeWorkerRawEventStore) AppendRawEvents(context.Context, []*core.RawEvent) ([]string, error) {
	return nil, core.ErrNotImplemented
}

func (s *fakeWorkerRawEventStore) GetRawEvents(_ context.Context, ids []string) ([]*core.RawEvent, error) {
	s.requestedIDs = append([]string(nil), ids...)
	if s.err != nil {
		return nil, s.err
	}
	events := make([]*core.RawEvent, 0, len(ids))
	for _, id := range ids {
		if event, ok := s.events[id]; ok {
			events = append(events, event)
		}
	}
	return events, nil
}

type fakeDreamingStore struct {
	input      *core.DreamingSessionInput
	summary    *core.SessionSummary
	promotions map[core.DreamingTier]*core.DreamingPromotionResult
	requests   []*core.DreamingPromotionRequest
}

func (s *fakeDreamingStore) LoadDreamingSessionInput(context.Context, *core.DreamSessionRequest) (*core.DreamingSessionInput, error) {
	if s.input != nil {
		return s.input, nil
	}
	return &core.DreamingSessionInput{}, nil
}

func (s *fakeDreamingStore) PromoteMemories(_ context.Context, req *core.DreamingPromotionRequest) (*core.DreamingPromotionResult, error) {
	s.requests = append(s.requests, req)
	if s.promotions != nil {
		if result, ok := s.promotions[req.Tier]; ok {
			return result, nil
		}
	}
	return &core.DreamingPromotionResult{}, nil
}

func (s *fakeDreamingStore) UpsertSessionSummary(_ context.Context, summary *core.SessionSummary) error {
	s.summary = summary
	return nil
}

func (s *fakeDreamingStore) GetSessionSummary(context.Context, string, string, string) (*core.SessionSummary, error) {
	return nil, core.ErrNotFound
}

type fakeReasoner struct {
	received *reasoning.ProcessTurnEnvelope
	err      error
	result   *reasoning.ProcessTurnResult
}

func (s *fakeReasoner) ProcessTurn(_ context.Context, envelope *reasoning.ProcessTurnEnvelope) (*reasoning.ProcessTurnResult, error) {
	s.received = envelope
	if s.err != nil {
		return nil, s.err
	}
	if s.result != nil {
		return s.result, nil
	}
	return successfulReasoningResult(), nil
}

type fakeApplyEngine struct {
	received *graph.ApplyRequest
	err      error
	result   *graph.ApplyResult
}

func (s *fakeApplyEngine) Apply(_ context.Context, req *graph.ApplyRequest) (*graph.ApplyResult, error) {
	s.received = req
	if s.err != nil {
		return nil, s.err
	}
	if s.result != nil {
		return s.result, nil
	}
	return &graph.ApplyResult{
		AppliedOperationCount: 0,
		MemoryIDs:             []string{},
		TraceWritten:          false,
	}, nil
}

func successfulReasoningResult() *reasoning.ProcessTurnResult {
	return &reasoning.ProcessTurnResult{
		Stage1: reasoning.Stage1Output{
			CandidateEntities: []reasoning.CandidateEntity{},
			CandidateMemories: []reasoning.CandidateMemory{},
		},
		Stage2: reasoning.Stage2Output{
			Operations:     []reasoning.GraphOperation{},
			ProfileDelta:   json.RawMessage(`{}`),
			SessionSummary: "",
			PlanDelta:      json.RawMessage(`{}`),
			Trace: reasoning.Trace{
				SchemaVersion: "v0",
				Stage:         reasoning.StageNameResolve,
				Codes:         []string{"test"},
				MetadataJSON:  json.RawMessage(`{}`),
			},
		},
	}
}
