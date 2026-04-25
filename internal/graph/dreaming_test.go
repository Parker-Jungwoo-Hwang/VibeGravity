// ============================================================
// FILE     : internal/graph/dreaming_test.go
// PURPOSE  : Verifies dreaming session summaries and tier-promotion orchestration.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : graph dreaming tests
// DEPENDS  : context, testing, time, internal/core
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: Dreaming tests should prove no duplicate memory creation is required.
// ============================================================

package graph

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

func TestDreamSessionWritesSummaryAndPromotesExistingMemories(t *testing.T) {
	t.Parallel()

	store := &fakeGraphDreamingStore{
		input: &core.DreamingSessionInput{
			RawEventIDs: []string{"evt_1", "evt_2"},
			Memories: []*core.Memory{
				{ID: "mem_1", Text: "User prefers narrow work-pack slices."},
				{ID: "mem_2", Text: "Workspace decisions are kept file-backed."},
			},
		},
		promotions: map[core.DreamingTier]*core.DreamingPromotionResult{
			core.DreamingTierMidTerm:       {PromotedCount: 2, MemoryIDs: []string{"mem_1", "mem_2"}},
			core.DreamingTierLongTerm:      {PromotedCount: 1, MemoryIDs: []string{"mem_2"}},
			core.DreamingTierUltraLongTerm: {PromotedCount: 0},
		},
	}
	service := newTestGraphDreamingService(t, store)

	result, err := service.DreamSession(context.Background(), &core.DreamSessionRequest{
		JobID:       "job_dream_1",
		TenantID:    "tenant_1",
		WorkspaceID: "workspace_1",
		SessionID:   "session_1",
	})
	if err != nil {
		t.Fatalf("DreamSession returned error: %v", err)
	}

	if !result.SessionSummaryWritten || result.MidTermPromoted != 2 || result.LongTermPromoted != 1 {
		t.Fatalf("unexpected dreaming result: %#v", result)
	}
	if store.summary == nil {
		t.Fatalf("expected session summary to be written")
	}
	if !strings.Contains(store.summary.SummaryText, "2 raw events and 2 derived memories") {
		t.Fatalf("unexpected summary text: %s", store.summary.SummaryText)
	}
	if len(store.requests) != 3 {
		t.Fatalf("expected mid, long, and ultra promotion requests, got %#v", store.requests)
	}
	if store.requests[0].Tier != core.DreamingTierMidTerm || len(store.requests[0].MemoryIDs) != 2 {
		t.Fatalf("unexpected mid-term request: %#v", store.requests[0])
	}
}

func TestDreamWorkspacePromotesStableTiersOnly(t *testing.T) {
	t.Parallel()

	store := &fakeGraphDreamingStore{
		promotions: map[core.DreamingTier]*core.DreamingPromotionResult{
			core.DreamingTierLongTerm:      {PromotedCount: 3},
			core.DreamingTierUltraLongTerm: {PromotedCount: 1},
		},
	}
	service := newTestGraphDreamingService(t, store)

	result, err := service.DreamWorkspace(context.Background(), &core.DreamWorkspaceRequest{
		JobID:       "job_dream_workspace",
		TenantID:    "tenant_1",
		WorkspaceID: "workspace_1",
	})
	if err != nil {
		t.Fatalf("DreamWorkspace returned error: %v", err)
	}

	if result.LongTermPromoted != 3 || result.UltraLongTermPromoted != 1 {
		t.Fatalf("unexpected workspace dreaming result: %#v", result)
	}
	if len(store.requests) != 2 {
		t.Fatalf("expected two promotion requests, got %#v", store.requests)
	}
	for _, req := range store.requests {
		if !req.RequireStableKind {
			t.Fatalf("workspace promotion must require stable kind: %#v", req)
		}
	}
}

func newTestGraphDreamingService(t *testing.T, store *fakeGraphDreamingStore) *DreamingService {
	t.Helper()
	service, err := NewDreamingService(DreamingDependencies{
		Store: store,
		Clock: func() time.Time {
			return time.Date(2026, time.April, 24, 2, 0, 0, 0, time.UTC)
		},
	})
	if err != nil {
		t.Fatalf("NewDreamingService returned error: %v", err)
	}
	return service
}

type fakeGraphDreamingStore struct {
	input      *core.DreamingSessionInput
	summary    *core.SessionSummary
	promotions map[core.DreamingTier]*core.DreamingPromotionResult
	requests   []*core.DreamingPromotionRequest
}

func (s *fakeGraphDreamingStore) LoadDreamingSessionInput(context.Context, *core.DreamSessionRequest) (*core.DreamingSessionInput, error) {
	if s.input != nil {
		return s.input, nil
	}
	return &core.DreamingSessionInput{}, nil
}

func (s *fakeGraphDreamingStore) PromoteMemories(_ context.Context, req *core.DreamingPromotionRequest) (*core.DreamingPromotionResult, error) {
	s.requests = append(s.requests, req)
	if s.promotions != nil {
		if result, ok := s.promotions[req.Tier]; ok {
			return result, nil
		}
	}
	return &core.DreamingPromotionResult{}, nil
}

func (s *fakeGraphDreamingStore) UpsertSessionSummary(_ context.Context, summary *core.SessionSummary) error {
	s.summary = summary
	return nil
}

func (s *fakeGraphDreamingStore) GetSessionSummary(context.Context, string) (*core.SessionSummary, error) {
	return nil, core.ErrNotFound
}
