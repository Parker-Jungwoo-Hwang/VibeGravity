// ============================================================
// FILE     : internal/hermes/provider_test.go
// PURPOSE  : Verifies Hermes provider lifecycle hooks delegate to VibeGravity core semantics.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : Hermes provider adapter tests
// DEPENDS  : context, errors, strings, testing, internal/core
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: These tests use a fake core service; they do not call a real Hermes runtime.
// ============================================================

package hermes

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

func TestProviderDelegatesPrefetchAndRendersContext(t *testing.T) {
	t.Parallel()

	service := &fakeService{
		prefetchResp: &core.PrefetchResponse{
			Blocks: []core.RecallBlock{
				{Kind: "pinned_note", Priority: 100, Text: "Keep Hermes first.", Scope: core.MemoryScopeWorkspaceShared, Source: "notes", Freshness: "stored"},
				{Kind: "active_plan", Priority: 95, Text: "Finish V1 core semantics."},
			},
			Meta: core.RecallMeta{EstimatedTokens: 12, Sources: []string{"notes", "plans"}},
		},
	}
	provider := newTestProvider(t, service)

	resp, err := provider.Prefetch(context.Background(), &core.PrefetchRequest{
		TenantID: "tenant_1", WorkspaceID: "workspace_1", SessionID: "session_1", ActorID: "agent:hermes-main",
	})
	if err != nil {
		t.Fatalf("Prefetch returned error: %v", err)
	}
	if service.prefetchCalls != 1 {
		t.Fatalf("expected one prefetch call, got %d", service.prefetchCalls)
	}
	rendered := provider.RenderContext(resp)
	if !strings.Contains(rendered, "[pinned_note:100:workspace_shared:notes:stored] Keep Hermes first.") {
		t.Fatalf("rendered context lost pinned note: %q", rendered)
	}
	if !strings.Contains(rendered, "[active_plan:95] Finish V1 core semantics.") {
		t.Fatalf("rendered context lost active plan: %q", rendered)
	}
}

func TestProviderDelegatesSyncTurn(t *testing.T) {
	t.Parallel()

	service := &fakeService{syncResp: &core.SyncTurnResponse{Status: "accepted", EventIDs: []string{"evt_1"}, JobIDs: []string{"job_1"}}}
	provider := newTestProvider(t, service)

	resp, err := provider.SyncTurn(context.Background(), &core.SyncTurnRequest{
		TenantID: "tenant_1", WorkspaceID: "workspace_1", SessionID: "session_1", ActorID: "agent:hermes-main",
	})
	if err != nil {
		t.Fatalf("SyncTurn returned error: %v", err)
	}
	if service.syncCalls != 1 {
		t.Fatalf("expected one sync call, got %d", service.syncCalls)
	}
	if resp.Status != "accepted" || len(resp.JobIDs) != 1 {
		t.Fatalf("unexpected sync response: %#v", resp)
	}
}

func TestProviderAvailabilityReflectsPrefetchHealth(t *testing.T) {
	t.Parallel()

	okProvider := newTestProvider(t, &fakeService{prefetchResp: &core.PrefetchResponse{}})
	if !okProvider.IsAvailable(context.Background(), &core.PrefetchRequest{TenantID: "tenant_1"}) {
		t.Fatalf("expected provider to be available when prefetch succeeds")
	}

	failingProvider := newTestProvider(t, &fakeService{prefetchErr: errors.New("database unavailable")})
	if failingProvider.IsAvailable(context.Background(), &core.PrefetchRequest{TenantID: "tenant_1"}) {
		t.Fatalf("expected provider to be unavailable when prefetch fails")
	}
}

func TestProviderToolsExposeV1HermesSurface(t *testing.T) {
	t.Parallel()

	provider := newTestProvider(t, &fakeService{})
	tools := provider.GetTools()
	names := make(map[string]bool, len(tools))
	for _, tool := range tools {
		names[tool.Name] = true
	}
	for _, want := range []string{"recall_preview", "search_memory", "add_note", "show_plan", "explain_memory", "correct_memory", "view_timeline", "degraded_status"} {
		if !names[want] {
			t.Fatalf("expected provider tool %q in %#v", want, tools)
		}
	}
}

func TestProviderCallToolDelegatesRecallPreview(t *testing.T) {
	t.Parallel()

	service := &fakeService{prefetchResp: &core.PrefetchResponse{Blocks: []core.RecallBlock{{Text: "Use scoped recall."}}}}
	provider := newTestProvider(t, service)

	raw, err := provider.CallTool(context.Background(), "recall_preview", json.RawMessage(`{"tenant_id":"tenant_1","workspace_id":"workspace_1"}`))
	if err != nil {
		t.Fatalf("CallTool returned error: %v", err)
	}
	if service.prefetchCalls != 1 {
		t.Fatalf("expected one prefetch call, got %d", service.prefetchCalls)
	}
	if service.prefetchReq == nil || service.prefetchReq.TenantID != "tenant_1" || service.prefetchReq.WorkspaceID != "workspace_1" {
		t.Fatalf("expected recall preview request to pass through unchanged, got %#v", service.prefetchReq)
	}
	if !strings.Contains(string(raw), "Use scoped recall.") {
		t.Fatalf("expected encoded recall preview output, got %s", string(raw))
	}
}

func TestProviderCallToolDelegatesTrustLoopTools(t *testing.T) {
	t.Parallel()

	service := &fakeService{
		searchResp:   &core.SearchMemoriesResponse{Memories: []core.MemoryResult{{MemoryID: "mem_1", Text: "Project rule"}}},
		addNoteResp:  &core.AddNoteResponse{NoteID: "note_1", Status: "created"},
		explainResp:  &core.ExplainMemoryResponse{MemoryID: "mem_1", Trace: core.MemoryTraceResult{ReasoningJobID: "job_1"}},
		correctResp:  &core.CorrectMemoryResponse{MemoryID: "mem_1", CorrectionRecorded: true},
		timelineResp: &core.GetTimelineResponse{Items: []core.TimelineItem{{ID: "item_1", Text: "Correction recorded"}}},
	}
	provider := newTestProvider(t, service)

	cases := []struct {
		name string
		raw  json.RawMessage
		want string
	}{
		{name: "search_memory", raw: json.RawMessage(`{"tenant_id":"tenant_1","workspace_id":"workspace_1","query":"rule"}`), want: "Project rule"},
		{name: "add_note", raw: json.RawMessage(`{"tenant_id":"tenant_1","workspace_id":"workspace_1","text":"Pin this"}`), want: "note_1"},
		{name: "explain_memory", raw: json.RawMessage(`{"tenant_id":"tenant_1","workspace_id":"workspace_1","memory_id":"mem_1"}`), want: "job_1"},
		{name: "correct_memory", raw: json.RawMessage(`{"tenant_id":"tenant_1","workspace_id":"workspace_1","memory_id":"mem_1"}`), want: "correction_recorded"},
		{name: "view_timeline", raw: json.RawMessage(`{"tenant_id":"tenant_1","workspace_id":"workspace_1","entity_id":"agent:hermes-main"}`), want: "Correction recorded"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := provider.CallTool(context.Background(), tc.name, tc.raw)
			if err != nil {
				t.Fatalf("CallTool returned error: %v", err)
			}
			if !strings.Contains(string(raw), tc.want) {
				t.Fatalf("expected output to contain %q, got %s", tc.want, string(raw))
			}
		})
	}
	if service.explainReq == nil || service.explainReq.MemoryID != "mem_1" || service.explainReq.WorkspaceID != "workspace_1" {
		t.Fatalf("expected explain request to pass through unchanged, got %#v", service.explainReq)
	}
	if service.correctReq == nil || service.correctReq.MemoryID != "mem_1" || service.correctReq.WorkspaceID != "workspace_1" {
		t.Fatalf("expected correction request to pass through unchanged, got %#v", service.correctReq)
	}
	if service.timelineReq == nil || service.timelineReq.EntityID != "agent:hermes-main" || service.timelineReq.WorkspaceID != "workspace_1" {
		t.Fatalf("expected timeline request to pass through unchanged, got %#v", service.timelineReq)
	}
}

func TestProviderCallToolReturnsDegradedStatusFromPrefetchMeta(t *testing.T) {
	t.Parallel()

	service := &fakeService{prefetchResp: &core.PrefetchResponse{Meta: core.RecallMeta{Freshness: "stale", Degraded: true, DegradedReasons: []string{"worker backlog"}}}}
	provider := newTestProvider(t, service)

	raw, err := provider.CallTool(context.Background(), "degraded_status", json.RawMessage(`{"tenant_id":"tenant_1","workspace_id":"workspace_1"}`))
	if err != nil {
		t.Fatalf("CallTool returned error: %v", err)
	}
	if service.prefetchCalls != 1 {
		t.Fatalf("expected one prefetch call, got %d", service.prefetchCalls)
	}
	if service.prefetchReq == nil || service.prefetchReq.TenantID != "tenant_1" || service.prefetchReq.WorkspaceID != "workspace_1" {
		t.Fatalf("expected degraded status request to pass through unchanged, got %#v", service.prefetchReq)
	}
	if !strings.Contains(string(raw), `"freshness":"stale"`) || !strings.Contains(string(raw), "worker backlog") {
		t.Fatalf("expected encoded recall meta, got %s", string(raw))
	}
}

func TestProviderCallToolRejectsUnknownInvalidAndUnbackedTools(t *testing.T) {
	t.Parallel()

	provider := newTestProvider(t, &fakeService{})
	if _, err := provider.CallTool(context.Background(), "unknown", json.RawMessage(`{}`)); !errors.Is(err, core.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument for unknown tool, got %v", err)
	}
	if _, err := provider.CallTool(context.Background(), "recall_preview", json.RawMessage(`{`)); !errors.Is(err, core.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument for invalid JSON, got %v", err)
	}
	if _, err := provider.CallTool(context.Background(), "show_plan", json.RawMessage(`{}`)); !errors.Is(err, core.ErrNotImplemented) {
		t.Fatalf("expected ErrNotImplemented for show_plan, got %v", err)
	}
}

func newTestProvider(t *testing.T, service core.VibeGravityService) *Provider {
	t.Helper()

	provider, err := NewProvider(service)
	if err != nil {
		t.Fatalf("NewProvider returned error: %v", err)
	}
	return provider
}

type fakeService struct {
	prefetchCalls int
	syncCalls     int
	searchCalls   int
	addNoteCalls  int
	explainCalls  int
	correctCalls  int
	timelineCalls int
	prefetchResp  *core.PrefetchResponse
	prefetchErr   error
	syncResp      *core.SyncTurnResponse
	syncErr       error
	searchResp    *core.SearchMemoriesResponse
	addNoteResp   *core.AddNoteResponse
	explainResp   *core.ExplainMemoryResponse
	correctResp   *core.CorrectMemoryResponse
	timelineResp  *core.GetTimelineResponse
	prefetchReq   *core.PrefetchRequest
	explainReq    *core.ExplainMemoryRequest
	correctReq    *core.CorrectMemoryRequest
	timelineReq   *core.GetTimelineRequest
}

func (s *fakeService) Prefetch(_ context.Context, req *core.PrefetchRequest) (*core.PrefetchResponse, error) {
	s.prefetchCalls++
	s.prefetchReq = req
	return s.prefetchResp, s.prefetchErr
}

func (s *fakeService) SyncTurn(context.Context, *core.SyncTurnRequest) (*core.SyncTurnResponse, error) {
	s.syncCalls++
	return s.syncResp, s.syncErr
}

func (s *fakeService) AddDocument(context.Context, *core.AddDocumentRequest) (*core.AddDocumentResponse, error) {
	return nil, core.ErrNotImplemented
}

func (s *fakeService) SearchMemories(context.Context, *core.SearchMemoriesRequest) (*core.SearchMemoriesResponse, error) {
	s.searchCalls++
	return s.searchResp, nil
}

func (s *fakeService) SearchDocuments(context.Context, *core.SearchDocumentsRequest) (*core.SearchDocumentsResponse, error) {
	return nil, core.ErrNotImplemented
}

func (s *fakeService) AddNote(context.Context, *core.AddNoteRequest) (*core.AddNoteResponse, error) {
	s.addNoteCalls++
	return s.addNoteResp, nil
}

func (s *fakeService) CreatePlan(context.Context, *core.CreatePlanRequest) (*core.CreatePlanResponse, error) {
	return nil, core.ErrNotImplemented
}

func (s *fakeService) UpdatePlan(context.Context, *core.UpdatePlanRequest) (*core.UpdatePlanResponse, error) {
	return nil, core.ErrNotImplemented
}

func (s *fakeService) CorrectMemory(_ context.Context, req *core.CorrectMemoryRequest) (*core.CorrectMemoryResponse, error) {
	s.correctCalls++
	s.correctReq = req
	return s.correctResp, nil
}

func (s *fakeService) GetTimeline(_ context.Context, req *core.GetTimelineRequest) (*core.GetTimelineResponse, error) {
	s.timelineCalls++
	s.timelineReq = req
	return s.timelineResp, nil
}

func (s *fakeService) ExplainMemory(_ context.Context, req *core.ExplainMemoryRequest) (*core.ExplainMemoryResponse, error) {
	s.explainCalls++
	s.explainReq = req
	return s.explainResp, nil
}
