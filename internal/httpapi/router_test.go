// ============================================================
// FILE     : internal/httpapi/router_test.go
// PURPOSE  : Verifies HTTP transport handlers delegate to the core service contract.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : TestPrefetchHandler_CallsService, TestSyncTurnHandler_CallsService
// DEPENDS  : internal/httpapi, internal/core
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: Keep handler tests about transport behavior; product rules belong in service tests.
// ============================================================

package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

func TestPrefetchHandler_CallsService(t *testing.T) {
	t.Parallel()

	service := &fakeVibeGravityService{}
	router := NewRouter(&App{Service: service})
	body := `{"tenant_id":"tenant_1","workspace_id":"workspace_1","session_id":"session_1","actor_id":"agent:hermes-main","query":"next","budget_tokens":2200,"mode":"default"}`

	req := httptest.NewRequest(http.MethodPost, "/v1/prefetch", strings.NewReader(body))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if service.prefetchReq == nil || service.prefetchReq.ActorID != "agent:hermes-main" {
		t.Fatalf("service did not receive prefetch request: %#v", service.prefetchReq)
	}
	if !strings.Contains(rr.Body.String(), "pinned_note") {
		t.Fatalf("response body did not include recall block: %s", rr.Body.String())
	}
}

func TestSyncTurnHandler_CallsService(t *testing.T) {
	t.Parallel()

	service := &fakeVibeGravityService{}
	router := NewRouter(&App{Service: service})
	body := `{"tenant_id":"tenant_1","workspace_id":"workspace_1","session_id":"session_1","actor_id":"agent:hermes-main","idempotency_key":"turn_1","turn_events":[{"event_kind":"user_message","source":"hermes","fingerprint":"fp_1","occurred_at":"2026-04-24T00:00:00Z","payload_json":{"text":"hello"}}]}`

	req := httptest.NewRequest(http.MethodPost, "/v1/sync-turn", strings.NewReader(body))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d: %s", rr.Code, rr.Body.String())
	}
	if service.syncReq == nil || service.syncReq.IdempotencyKey != "turn_1" {
		t.Fatalf("service did not receive sync request: %#v", service.syncReq)
	}
	if !strings.Contains(rr.Body.String(), "accepted") {
		t.Fatalf("response body did not include accepted status: %s", rr.Body.String())
	}
}

func TestHealthz_ReturnsUnavailableWithoutDBPool(t *testing.T) {
	t.Parallel()

	router := NewRouter(&App{})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d: %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "database unavailable") {
		t.Fatalf("response body did not explain database unavailability: %s", rr.Body.String())
	}
}

func TestAddDocumentHandler_CallsService(t *testing.T) {
	t.Parallel()

	service := &fakeVibeGravityService{}
	router := NewRouter(&App{Service: service})
	body := `{"tenant_id":"tenant_1","workspace_id":"workspace_1","source":"operator_upload","title":"Runtime Notes","content":"important context"}`

	req := httptest.NewRequest(http.MethodPost, "/v1/documents", strings.NewReader(body))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
	if service.addDocumentReq == nil || service.addDocumentReq.Title != "Runtime Notes" {
		t.Fatalf("service did not receive add document request: %#v", service.addDocumentReq)
	}
	if !strings.Contains(rr.Body.String(), "doc_1") {
		t.Fatalf("response body did not include document id: %s", rr.Body.String())
	}
}

func TestUpdatePlanHandler_UsesPathID(t *testing.T) {
	t.Parallel()

	service := &fakeVibeGravityService{}
	router := NewRouter(&App{Service: service})
	body := `{"tenant_id":"tenant_1","workspace_id":"workspace_1","plan_id":"body_value","status":"active"}`

	req := httptest.NewRequest(http.MethodPatch, "/v1/plans/plan_path", strings.NewReader(body))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if service.updatePlanReq == nil || service.updatePlanReq.PlanID != "plan_path" {
		t.Fatalf("service did not receive path plan id: %#v", service.updatePlanReq)
	}
	if !strings.Contains(rr.Body.String(), "updated") {
		t.Fatalf("response body did not include updated status: %s", rr.Body.String())
	}
}

func TestCorrectMemoryHandler_CallsService(t *testing.T) {
	t.Parallel()

	service := &fakeVibeGravityService{}
	router := NewRouter(&App{Service: service})
	body := `{"tenant_id":"tenant_1","workspace_id":"workspace_1","memory_id":"mem_1","operator_id":"operator_1","idempotency_key":"correction_1","correction_text":"Use the newer fact."}`

	req := httptest.NewRequest(http.MethodPost, "/v1/memory/correct", strings.NewReader(body))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if service.correctMemoryReq == nil || service.correctMemoryReq.MemoryID != "mem_1" {
		t.Fatalf("service did not receive correction request: %#v", service.correctMemoryReq)
	}
}

func TestGetTimelineHandler_CallsService(t *testing.T) {
	t.Parallel()

	service := &fakeVibeGravityService{}
	router := NewRouter(&App{Service: service})

	req := httptest.NewRequest(http.MethodGet, "/v1/timeline?tenant_id=tenant_1&workspace_id=workspace_1&entity_id=agent:hermes-main&scopes=agent_private,workspace_shared&from=2026-04-24T00:00:00Z&to=2026-04-25T00:00:00Z&limit=25", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if service.timelineReq == nil || service.timelineReq.EntityID != "agent:hermes-main" {
		t.Fatalf("service did not receive timeline request: %#v", service.timelineReq)
	}
	if got := service.timelineReq.Scopes; len(got) != 2 || got[0] != core.MemoryScopeAgentPrivate || got[1] != core.MemoryScopeWorkspaceShared {
		t.Fatalf("handler did not parse scopes: %#v", got)
	}
	if service.timelineReq.From == nil || service.timelineReq.To == nil || service.timelineReq.Limit != 25 {
		t.Fatalf("handler did not parse from/to/limit: %#v", service.timelineReq)
	}
	if !service.timelineReq.From.Equal(time.Date(2026, 4, 24, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected from time: %v", service.timelineReq.From)
	}
}

func TestGetTimelineHandler_RejectsInvalidQuery(t *testing.T) {
	t.Parallel()

	service := &fakeVibeGravityService{}
	router := NewRouter(&App{Service: service})

	req := httptest.NewRequest(http.MethodGet, "/v1/timeline?tenant_id=tenant_1&workspace_id=workspace_1&entity_id=agent:hermes-main&limit=not-a-number", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
	if service.timelineReq != nil {
		t.Fatalf("service should not receive invalid timeline request: %#v", service.timelineReq)
	}
}

func TestExplainMemoryHandler_CallsServiceWithVisibility(t *testing.T) {
	t.Parallel()

	service := &fakeVibeGravityService{}
	router := NewRouter(&App{Service: service})

	req := httptest.NewRequest(http.MethodGet, "/v1/memory/mem_1/explain?tenant_id=tenant_1&workspace_id=workspace_1&entity_id=agent:hermes-main&visible_group_ids=group_design,group_ops", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if service.explainReq == nil || service.explainReq.MemoryID != "mem_1" || service.explainReq.EntityID != "agent:hermes-main" {
		t.Fatalf("service did not receive explain visibility request: %#v", service.explainReq)
	}
	if got := service.explainReq.VisibleGroupIDs; len(got) != 2 || got[0] != "group_design" || got[1] != "group_ops" {
		t.Fatalf("handler did not parse visible group ids: %#v", service.explainReq)
	}
}

type fakeVibeGravityService struct {
	prefetchReq      *core.PrefetchRequest
	syncReq          *core.SyncTurnRequest
	addDocumentReq   *core.AddDocumentRequest
	updatePlanReq    *core.UpdatePlanRequest
	correctMemoryReq *core.CorrectMemoryRequest
	timelineReq      *core.GetTimelineRequest
	explainReq       *core.ExplainMemoryRequest
}

func (s *fakeVibeGravityService) Prefetch(_ context.Context, req *core.PrefetchRequest) (*core.PrefetchResponse, error) {
	s.prefetchReq = req
	return &core.PrefetchResponse{
		Blocks: []core.RecallBlock{{
			Kind:     "pinned_note",
			Priority: 100,
			Text:     "Keep the Hermes-first plan visible.",
		}},
		Meta: core.RecallMeta{
			EstimatedTokens: 8,
			Sources:         []string{"notes"},
		},
	}, nil
}

func (s *fakeVibeGravityService) SyncTurn(_ context.Context, req *core.SyncTurnRequest) (*core.SyncTurnResponse, error) {
	s.syncReq = req
	return &core.SyncTurnResponse{
		Status:         "accepted",
		SessionID:      req.SessionID,
		EventIDs:       []string{"evt_1"},
		JobIDs:         []string{"job_1"},
		DuplicateCount: 0,
	}, nil
}

func (s *fakeVibeGravityService) AddDocument(_ context.Context, req *core.AddDocumentRequest) (*core.AddDocumentResponse, error) {
	s.addDocumentReq = req
	return &core.AddDocumentResponse{
		DocumentID: "doc_1",
		ChunkIDs:   []string{"chunk_1"},
		Status:     "created",
	}, nil
}

func (s *fakeVibeGravityService) SearchMemories(context.Context, *core.SearchMemoriesRequest) (*core.SearchMemoriesResponse, error) {
	return nil, core.ErrNotImplemented
}

func (s *fakeVibeGravityService) SearchDocuments(context.Context, *core.SearchDocumentsRequest) (*core.SearchDocumentsResponse, error) {
	return nil, core.ErrNotImplemented
}

func (s *fakeVibeGravityService) AddNote(context.Context, *core.AddNoteRequest) (*core.AddNoteResponse, error) {
	return nil, core.ErrNotImplemented
}

func (s *fakeVibeGravityService) CreatePlan(context.Context, *core.CreatePlanRequest) (*core.CreatePlanResponse, error) {
	return nil, core.ErrNotImplemented
}

func (s *fakeVibeGravityService) UpdatePlan(_ context.Context, req *core.UpdatePlanRequest) (*core.UpdatePlanResponse, error) {
	s.updatePlanReq = req
	return &core.UpdatePlanResponse{PlanID: req.PlanID, Status: "updated"}, nil
}

func (s *fakeVibeGravityService) CorrectMemory(_ context.Context, req *core.CorrectMemoryRequest) (*core.CorrectMemoryResponse, error) {
	s.correctMemoryReq = req
	return &core.CorrectMemoryResponse{
		MemoryID:           req.MemoryID,
		RawEventID:         "evt_correction",
		CorrectionID:       "corr_1",
		CorrectionRecorded: true,
		TraceWritten:       false,
		Status:             "accepted",
	}, nil
}

func (s *fakeVibeGravityService) GetTimeline(_ context.Context, req *core.GetTimelineRequest) (*core.GetTimelineResponse, error) {
	s.timelineReq = req
	return &core.GetTimelineResponse{Items: []core.TimelineItem{}}, nil
}

func (s *fakeVibeGravityService) ExplainMemory(_ context.Context, req *core.ExplainMemoryRequest) (*core.ExplainMemoryResponse, error) {
	s.explainReq = req
	return &core.ExplainMemoryResponse{MemoryID: req.MemoryID}, nil
}
