// ============================================================
// FILE     : internal/mcp/surface_test.go
// PURPOSE  : Verifies the MCP-style tool surface delegates to shared core semantics.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : MCP surface adapter tests
// DEPENDS  : context, encoding/json, errors, strings, testing, internal/core
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: These tests do not start a protocol server; they lock tool-to-core mapping.
// ============================================================

package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

func TestSurfaceListsV1Tools(t *testing.T) {
	t.Parallel()

	surface := newTestSurface(t, &fakeService{})
	tools := surface.Tools()
	names := make(map[string]bool, len(tools))
	for _, tool := range tools {
		names[tool.Name] = true
	}
	for _, want := range []string{"prefetch", "recall_preview", "sync_turn", "search_memory", "add_note", "correct_memory", "view_timeline", "explain_memory", "degraded_status"} {
		if !names[want] {
			t.Fatalf("expected tool %q in %#v", want, tools)
		}
	}
}

func TestSurfaceCallsRecallPreviewAlias(t *testing.T) {
	t.Parallel()

	service := &fakeService{prefetchResp: &core.PrefetchResponse{Blocks: []core.RecallBlock{{Kind: "memory", Priority: 90, Text: "Preview"}}}}
	surface := newTestSurface(t, service)

	raw, err := surface.Call(context.Background(), "recall_preview", json.RawMessage(`{"tenant_id":"tenant_1","workspace_id":"workspace_1"}`))
	if err != nil {
		t.Fatalf("Call returned error: %v", err)
	}
	if service.prefetchCalls != 1 {
		t.Fatalf("expected recall_preview to delegate to prefetch, got %d calls", service.prefetchCalls)
	}
	if !strings.Contains(string(raw), "Preview") {
		t.Fatalf("expected encoded recall preview output, got %s", string(raw))
	}
}

func TestSurfaceCallsPrefetch(t *testing.T) {
	t.Parallel()

	service := &fakeService{prefetchResp: &core.PrefetchResponse{Blocks: []core.RecallBlock{{Kind: "note", Priority: 100, Text: "Pinned"}}}}
	surface := newTestSurface(t, service)

	raw, err := surface.Call(context.Background(), "prefetch", json.RawMessage(`{"tenant_id":"tenant_1","workspace_id":"workspace_1"}`))
	if err != nil {
		t.Fatalf("Call returned error: %v", err)
	}
	if service.prefetchCalls != 1 {
		t.Fatalf("expected one prefetch call, got %d", service.prefetchCalls)
	}
	if !strings.Contains(string(raw), "Pinned") {
		t.Fatalf("expected encoded prefetch output, got %s", string(raw))
	}
}

func TestSurfaceCallsCorrectMemory(t *testing.T) {
	t.Parallel()

	service := &fakeService{correctResp: &core.CorrectMemoryResponse{MemoryID: "mem_1", CorrectionRecorded: true, Status: "recorded"}}
	surface := newTestSurface(t, service)

	raw, err := surface.Call(context.Background(), "correct_memory", json.RawMessage(`{"tenant_id":"tenant_1","workspace_id":"workspace_1","memory_id":"mem_1"}`))
	if err != nil {
		t.Fatalf("Call returned error: %v", err)
	}
	if service.correctCalls != 1 {
		t.Fatalf("expected one correction call, got %d", service.correctCalls)
	}
	if !strings.Contains(string(raw), `"correction_recorded":true`) {
		t.Fatalf("expected encoded correction output, got %s", string(raw))
	}
}

func TestSurfaceCallsViewTimeline(t *testing.T) {
	t.Parallel()

	service := &fakeService{timelineResp: &core.GetTimelineResponse{Items: []core.TimelineItem{{ID: "item_1", MemoryID: "mem_1", Text: "Corrected project rule"}}}}
	surface := newTestSurface(t, service)

	raw, err := surface.Call(context.Background(), "view_timeline", json.RawMessage(`{"tenant_id":"tenant_1","workspace_id":"workspace_1","entity_id":"agent:hermes-main"}`))
	if err != nil {
		t.Fatalf("Call returned error: %v", err)
	}
	if service.timelineCalls != 1 {
		t.Fatalf("expected one timeline call, got %d", service.timelineCalls)
	}
	if !strings.Contains(string(raw), "Corrected project rule") {
		t.Fatalf("expected encoded timeline output, got %s", string(raw))
	}
}

func TestSurfaceCallsExplainMemory(t *testing.T) {
	t.Parallel()

	service := &fakeService{explainResp: &core.ExplainMemoryResponse{MemoryID: "mem_1", Trace: core.MemoryTraceResult{ReasoningJobID: "job_1"}}}
	surface := newTestSurface(t, service)

	raw, err := surface.Call(context.Background(), "explain_memory", json.RawMessage(`{"tenant_id":"tenant_1","workspace_id":"workspace_1","memory_id":"mem_1","entity_id":"agent:hermes-main","visible_group_ids":["group_design"]}`))
	if err != nil {
		t.Fatalf("Call returned error: %v", err)
	}
	if service.explainCalls != 1 {
		t.Fatalf("expected one explain call, got %d", service.explainCalls)
	}
	if service.explainReq == nil || service.explainReq.EntityID != "agent:hermes-main" {
		t.Fatalf("expected explain visibility request, got %#v", service.explainReq)
	}
	if got := service.explainReq.VisibleGroupIDs; len(got) != 1 || got[0] != "group_design" {
		t.Fatalf("expected visible group ids, got %#v", service.explainReq)
	}
	if !strings.Contains(string(raw), `"reasoning_job_id":"job_1"`) {
		t.Fatalf("expected encoded explain output, got %s", string(raw))
	}
}

func TestSurfaceCallsDegradedStatus(t *testing.T) {
	t.Parallel()

	service := &fakeService{prefetchResp: &core.PrefetchResponse{Meta: core.RecallMeta{Freshness: "stale", Degraded: true, DegradedReasons: []string{"worker backlog"}}}}
	surface := newTestSurface(t, service)

	raw, err := surface.Call(context.Background(), "degraded_status", json.RawMessage(`{"tenant_id":"tenant_1","workspace_id":"workspace_1","session_id":"session_1","actor_id":"agent:hermes-main"}`))
	if err != nil {
		t.Fatalf("Call returned error: %v", err)
	}
	if service.prefetchCalls != 1 {
		t.Fatalf("expected one prefetch call, got %d", service.prefetchCalls)
	}
	if !strings.Contains(string(raw), `"freshness":"stale"`) || !strings.Contains(string(raw), "worker backlog") {
		t.Fatalf("expected encoded recall meta, got %s", string(raw))
	}
}

func TestSurfaceRejectsUnknownToolAndInvalidJSON(t *testing.T) {
	t.Parallel()

	surface := newTestSurface(t, &fakeService{})
	if _, err := surface.Call(context.Background(), "unknown", json.RawMessage(`{}`)); !errors.Is(err, core.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument for unknown tool, got %v", err)
	}
	if _, err := surface.Call(context.Background(), "prefetch", json.RawMessage(`{`)); !errors.Is(err, core.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument for invalid JSON, got %v", err)
	}
}

func newTestSurface(t *testing.T, service core.VibeGravityService) *Surface {
	t.Helper()

	surface, err := NewSurface(service)
	if err != nil {
		t.Fatalf("NewSurface returned error: %v", err)
	}
	return surface
}

type fakeService struct {
	prefetchCalls int
	correctCalls  int
	timelineCalls int
	explainCalls  int
	prefetchResp  *core.PrefetchResponse
	correctResp   *core.CorrectMemoryResponse
	timelineResp  *core.GetTimelineResponse
	explainResp   *core.ExplainMemoryResponse
	explainReq    *core.ExplainMemoryRequest
}

func (s *fakeService) Prefetch(context.Context, *core.PrefetchRequest) (*core.PrefetchResponse, error) {
	s.prefetchCalls++
	return s.prefetchResp, nil
}

func (s *fakeService) SyncTurn(context.Context, *core.SyncTurnRequest) (*core.SyncTurnResponse, error) {
	return &core.SyncTurnResponse{Status: "accepted"}, nil
}

func (s *fakeService) AddDocument(context.Context, *core.AddDocumentRequest) (*core.AddDocumentResponse, error) {
	return &core.AddDocumentResponse{Status: "created"}, nil
}

func (s *fakeService) SearchMemories(context.Context, *core.SearchMemoriesRequest) (*core.SearchMemoriesResponse, error) {
	return &core.SearchMemoriesResponse{}, nil
}

func (s *fakeService) SearchDocuments(context.Context, *core.SearchDocumentsRequest) (*core.SearchDocumentsResponse, error) {
	return &core.SearchDocumentsResponse{}, nil
}

func (s *fakeService) AddNote(context.Context, *core.AddNoteRequest) (*core.AddNoteResponse, error) {
	return &core.AddNoteResponse{Status: "created"}, nil
}

func (s *fakeService) CreatePlan(context.Context, *core.CreatePlanRequest) (*core.CreatePlanResponse, error) {
	return &core.CreatePlanResponse{Status: "created"}, nil
}

func (s *fakeService) UpdatePlan(context.Context, *core.UpdatePlanRequest) (*core.UpdatePlanResponse, error) {
	return &core.UpdatePlanResponse{Status: "updated"}, nil
}

func (s *fakeService) CorrectMemory(context.Context, *core.CorrectMemoryRequest) (*core.CorrectMemoryResponse, error) {
	s.correctCalls++
	return s.correctResp, nil
}

func (s *fakeService) GetTimeline(context.Context, *core.GetTimelineRequest) (*core.GetTimelineResponse, error) {
	s.timelineCalls++
	return s.timelineResp, nil
}

func (s *fakeService) ExplainMemory(_ context.Context, req *core.ExplainMemoryRequest) (*core.ExplainMemoryResponse, error) {
	s.explainCalls++
	s.explainReq = req
	return s.explainResp, nil
}
