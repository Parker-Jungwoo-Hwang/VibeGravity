# 03 Gpt Pro Ingest Recall Kernel Http

Generated: 2026-04-25

This file is part of the GPT-Pro review material bundle for VibeGravity.

## Included Sources

- `internal/httpapi/router.go`
- `internal/httpapi/router_test.go`
- `internal/ingest/doc.go`
- `internal/ingest/service.go`
- `internal/ingest/service_test.go`
- `internal/kernel/doc.go`
- `internal/kernel/service.go`
- `internal/kernel/service_test.go`
- `internal/recall/assembler.go`
- `internal/recall/assembler_test.go`
- `internal/recall/doc.go`
- `internal/recall/freshness.go`

## Source Contents


<!-- Source: internal/httpapi/router.go | bytes=11550 | lines=417 | sha16=e4817702894f30ef -->

```go
// ============================================================
// FILE     : internal/httpapi/router.go
// PURPOSE  : Defines the HTTP router and initial transport handlers.
// LAYER    : interface
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : App, NewRouter
// DEPENDS  : internal/core, errors, fmt, net/http, strconv, strings, time, github.com/go-chi/chi/v5, pgxpool
// USED_BY  : cmd/server, tests
// ------------------------------------------------------------
// AGENT_NOTE: HTTP handlers must preserve the core service semantics.
// ============================================================

// Package httpapi provides the HTTP transport layer for VibeGravity.
package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

// App holds dependencies for the HTTP API.
type App struct {
	Service core.VibeGravityService
	DBPool  *pgxpool.Pool
}

// NewRouter creates and configures the HTTP router.
func NewRouter(app *App) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/healthz", app.Healthz)

	r.Route("/v1", func(r chi.Router) {
		r.Post("/prefetch", app.Prefetch)
		r.Post("/sync-turn", app.SyncTurn)
		r.Post("/documents", app.AddDocument)
		r.Post("/search/memories", app.SearchMemories)
		r.Post("/search/documents", app.SearchDocuments)
		r.Post("/notes", app.AddNote)
		r.Post("/plans", app.CreatePlan)
		r.Patch("/plans/{id}", app.UpdatePlan)
		r.Post("/memory/correct", app.CorrectMemory)
		r.Get("/memory/{id}/explain", app.ExplainMemory)
		r.Get("/timeline", app.GetTimeline)
	})

	return r
}

// Healthz returns 200 OK if the application and database are healthy.
func (a *App) Healthz(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if a.DBPool == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"status": "error",
			"error":  "database unavailable",
		})
		return
	}

	if err := a.DBPool.Ping(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"status": "error",
			"error":  "database unavailable",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

// Prefetch handles recall pack requests.
func (a *App) Prefetch(w http.ResponseWriter, r *http.Request) {
	if a.Service == nil {
		writeError(w, http.StatusServiceUnavailable, "service unavailable")
		return
	}
	var req core.PrefetchRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON request")
		return
	}
	resp, err := a.Service.Prefetch(r.Context(), &req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// SyncTurn handles hot-path turn ingest requests.
func (a *App) SyncTurn(w http.ResponseWriter, r *http.Request) {
	if a.Service == nil {
		writeError(w, http.StatusServiceUnavailable, "service unavailable")
		return
	}
	var req core.SyncTurnRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON request")
		return
	}
	resp, err := a.Service.SyncTurn(r.Context(), &req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, resp)
}

// AddDocument handles document creation requests.
func (a *App) AddDocument(w http.ResponseWriter, r *http.Request) {
	if a.Service == nil {
		writeError(w, http.StatusServiceUnavailable, "service unavailable")
		return
	}
	var req core.AddDocumentRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON request")
		return
	}
	resp, err := a.Service.AddDocument(r.Context(), &req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

// SearchMemories handles manual memory search requests.
func (a *App) SearchMemories(w http.ResponseWriter, r *http.Request) {
	if a.Service == nil {
		writeError(w, http.StatusServiceUnavailable, "service unavailable")
		return
	}
	var req core.SearchMemoriesRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON request")
		return
	}
	resp, err := a.Service.SearchMemories(r.Context(), &req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// SearchDocuments handles manual document search requests.
func (a *App) SearchDocuments(w http.ResponseWriter, r *http.Request) {
	if a.Service == nil {
		writeError(w, http.StatusServiceUnavailable, "service unavailable")
		return
	}
	var req core.SearchDocumentsRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON request")
		return
	}
	resp, err := a.Service.SearchDocuments(r.Context(), &req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// AddNote handles manual note creation requests.
func (a *App) AddNote(w http.ResponseWriter, r *http.Request) {
	if a.Service == nil {
		writeError(w, http.StatusServiceUnavailable, "service unavailable")
		return
	}
	var req core.AddNoteRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON request")
		return
	}
	resp, err := a.Service.AddNote(r.Context(), &req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

// CreatePlan handles structured plan creation requests.
func (a *App) CreatePlan(w http.ResponseWriter, r *http.Request) {
	if a.Service == nil {
		writeError(w, http.StatusServiceUnavailable, "service unavailable")
		return
	}
	var req core.CreatePlanRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON request")
		return
	}
	resp, err := a.Service.CreatePlan(r.Context(), &req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

// UpdatePlan handles structured plan patch requests.
func (a *App) UpdatePlan(w http.ResponseWriter, r *http.Request) {
	if a.Service == nil {
		writeError(w, http.StatusServiceUnavailable, "service unavailable")
		return
	}
	var req core.UpdatePlanRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON request")
		return
	}
	req.PlanID = chi.URLParam(r, "id")
	resp, err := a.Service.UpdatePlan(r.Context(), &req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// CorrectMemory handles human correction requests.
func (a *App) CorrectMemory(w http.ResponseWriter, r *http.Request) {
	if a.Service == nil {
		writeError(w, http.StatusServiceUnavailable, "service unavailable")
		return
	}
	var req core.CorrectMemoryRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON request")
		return
	}
	resp, err := a.Service.CorrectMemory(r.Context(), &req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// ExplainMemory handles provenance lookup requests.
func (a *App) ExplainMemory(w http.ResponseWriter, r *http.Request) {
	if a.Service == nil {
		writeError(w, http.StatusServiceUnavailable, "service unavailable")
		return
	}
	req := core.ExplainMemoryRequest{
		TenantID:        r.URL.Query().Get("tenant_id"),
		WorkspaceID:     r.URL.Query().Get("workspace_id"),
		MemoryID:        chi.URLParam(r, "id"),
		EntityID:        r.URL.Query().Get("entity_id"),
		VisibleGroupIDs: parseStringList(r.URL.Query()["visible_group_ids"]),
	}
	resp, err := a.Service.ExplainMemory(r.Context(), &req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// GetTimeline handles timeline view requests.
func (a *App) GetTimeline(w http.ResponseWriter, r *http.Request) {
	if a.Service == nil {
		writeError(w, http.StatusServiceUnavailable, "service unavailable")
		return
	}
	req, err := timelineRequestFromQuery(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	resp, err := a.Service.GetTimeline(r.Context(), req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func decodeJSON(r *http.Request, out any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(out)
}

func timelineRequestFromQuery(r *http.Request) (*core.GetTimelineRequest, error) {
	values := r.URL.Query()
	req := &core.GetTimelineRequest{
		TenantID:    values.Get("tenant_id"),
		WorkspaceID: values.Get("workspace_id"),
		EntityID:    values.Get("entity_id"),
	}
	scopes, err := parseTimelineScopes(values["scopes"])
	if err != nil {
		return nil, err
	}
	req.Scopes = scopes
	from, err := parseOptionalTime(values.Get("from"), "from")
	if err != nil {
		return nil, err
	}
	req.From = from
	to, err := parseOptionalTime(values.Get("to"), "to")
	if err != nil {
		return nil, err
	}
	req.To = to
	if limitValue := strings.TrimSpace(values.Get("limit")); limitValue != "" {
		limit, err := strconv.Atoi(limitValue)
		if err != nil {
			return nil, fmt.Errorf("invalid limit")
		}
		req.Limit = limit
	}
	return req, nil
}

func parseTimelineScopes(values []string) ([]core.MemoryScope, error) {
	scopes := make([]core.MemoryScope, 0, len(values))
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			scope := core.MemoryScope(part)
			switch scope {
			case core.MemoryScopeAgentPrivate, core.MemoryScopeWorkspaceShared,
				core.MemoryScopeGroupShared, core.MemoryScopeSessionScratch:
				scopes = append(scopes, scope)
			default:
				return nil, fmt.Errorf("invalid scope %q", part)
			}
		}
	}
	return scopes, nil
}

func parseStringList(values []string) []string {
	items := make([]string, 0, len(values))
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			item := strings.TrimSpace(part)
			if item != "" {
				items = append(items, item)
			}
		}
	}
	return items
}

func parseOptionalTime(value, field string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, fmt.Errorf("invalid %s", field)
	}
	parsed = parsed.UTC()
	return &parsed, nil
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, core.ErrInvalidArgument):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, core.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, core.ErrConflict):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, core.ErrNotImplemented):
		writeError(w, http.StatusNotImplemented, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal error")
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{
		"error": message,
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

```



<!-- Source: internal/httpapi/router_test.go | bytes=11633 | lines=304 | sha16=89f5d8be3435d63b -->

```go
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

```



<!-- Source: internal/ingest/doc.go | bytes=786 | lines=16 | sha16=f59d228ec9f4c8f5 -->

```go
// ============================================================
// FILE     : internal/ingest/doc.go
// PURPOSE  : Provides package documentation for the sync_turn hot write path.
// LAYER    : application
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : package ingest
// DEPENDS  : plans/05_runtime-contracts_ingest-recall-apply.md
// USED_BY  : core service implementations, HTTP API, Hermes adapter
// ------------------------------------------------------------
// AGENT_NOTE: Keep sync_turn fast: normalize, validate, dedupe, insert raw events, enqueue jobs, ack.
// ============================================================

// Package ingest owns the sync_turn hot path for raw event writes and job enqueueing.
package ingest

```



<!-- Source: internal/ingest/service.go | bytes=6599 | lines=216 | sha16=902c77ef6248ec9e -->

```go
// ============================================================
// FILE     : internal/ingest/service.go
// PURPOSE  : Implements the sync_turn hot path for raw events and worker jobs.
// LAYER    : application
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : Dependencies, Service, NewService
// DEPENDS  : internal/core, internal/store
// USED_BY  : internal/kernel, tests
// ------------------------------------------------------------
// AGENT_NOTE: This path must not perform reasoning, graph updates, profile updates, or dreaming.
// ============================================================

package ingest

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/store"
)

const defaultSyncSource = "sync_turn"

// Dependencies collects stores and runtime hooks needed by the ingest service.
type Dependencies struct {
	RawEvents store.RawEventStore
	Jobs      store.JobStore
	Clock     func() time.Time
}

// Service owns the sync_turn hot path.
type Service struct {
	rawEvents store.RawEventStore
	jobs      store.JobStore
	clock     func() time.Time
}

// NewService builds an ingest service.
func NewService(deps Dependencies) (*Service, error) {
	if deps.RawEvents == nil {
		return nil, fmt.Errorf("%w: ingest raw event store is required", core.ErrInvalidArgument)
	}
	if deps.Jobs == nil {
		return nil, fmt.Errorf("%w: ingest job store is required", core.ErrInvalidArgument)
	}
	clock := deps.Clock
	if clock == nil {
		clock = time.Now
	}
	return &Service{
		rawEvents: deps.RawEvents,
		jobs:      deps.Jobs,
		clock:     clock,
	}, nil
}

// SyncTurn records raw turn events and enqueues one process_turn_event job.
func (s *Service) SyncTurn(ctx context.Context, req *core.SyncTurnRequest) (*core.SyncTurnResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("%w: sync turn request is required", core.ErrInvalidArgument)
	}
	if err := validateSyncTurnRequest(req); err != nil {
		return nil, err
	}

	now := s.clock().UTC()
	events, err := s.buildRawEvents(req, now)
	if err != nil {
		return nil, err
	}

	eventIDs, err := s.rawEvents.AppendRawEvents(ctx, events)
	if err != nil {
		if errors.Is(err, core.ErrDuplicate) {
			return &core.SyncTurnResponse{
				Status:         "accepted",
				SessionID:      req.SessionID,
				EventIDs:       nil,
				JobIDs:         nil,
				DuplicateCount: len(events),
			}, nil
		}
		return nil, fmt.Errorf("append raw events: %w", err)
	}
	if len(eventIDs) > len(events) {
		return nil, fmt.Errorf("%w: raw event store returned too many ids", core.ErrConflict)
	}

	jobIDs := make([]string, 0, 1)
	if len(eventIDs) > 0 {
		payload, err := json.Marshal(map[string]string{
			"session_id": req.SessionID,
			"actor_id":   req.ActorID,
			"source":     defaultSyncSource,
		})
		if err != nil {
			return nil, fmt.Errorf("encode ingest job payload: %w", err)
		}
		jobs := []*core.IngestJob{{
			ID:          stableID("job", req.TenantID, req.WorkspaceID, req.SessionID, req.IdempotencyKey, strings.Join(eventIDs, ",")),
			TenantID:    req.TenantID,
			WorkspaceID: req.WorkspaceID,
			JobKind:     core.JobKindProcessTurnEvent,
			Status:      "queued",
			RawEventIDs: eventIDs,
			PayloadJSON: payload,
			Attempts:    0,
			AvailableAt: now,
			CreatedAt:   now,
			UpdatedAt:   now,
		}}
		jobIDs, err = s.jobs.EnqueueJobs(ctx, jobs)
		if err != nil {
			return nil, fmt.Errorf("enqueue ingest jobs: %w", err)
		}
	}

	return &core.SyncTurnResponse{
		Status:         "accepted",
		SessionID:      req.SessionID,
		EventIDs:       eventIDs,
		JobIDs:         jobIDs,
		DuplicateCount: len(events) - len(eventIDs),
	}, nil
}

func validateSyncTurnRequest(req *core.SyncTurnRequest) error {
	required := map[string]string{
		"tenant_id":       req.TenantID,
		"workspace_id":    req.WorkspaceID,
		"session_id":      req.SessionID,
		"actor_id":        req.ActorID,
		"idempotency_key": req.IdempotencyKey,
	}
	for name, value := range required {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%w: %s is required", core.ErrInvalidArgument, name)
		}
	}
	if len(req.TurnEvents) == 0 {
		return fmt.Errorf("%w: at least one turn event is required", core.ErrInvalidArgument)
	}
	for i, event := range req.TurnEvents {
		if strings.TrimSpace(event.EventKind) == "" {
			return fmt.Errorf("%w: turn_events[%d].event_kind is required", core.ErrInvalidArgument, i)
		}
		if len(normalizePayload(event.PayloadJSON)) > 0 && !json.Valid(normalizePayload(event.PayloadJSON)) {
			return fmt.Errorf("%w: turn_events[%d].payload_json must be valid JSON", core.ErrInvalidArgument, i)
		}
	}
	return nil
}

func (s *Service) buildRawEvents(req *core.SyncTurnRequest, now time.Time) ([]*core.RawEvent, error) {
	events := make([]*core.RawEvent, 0, len(req.TurnEvents))
	for i, event := range req.TurnEvents {
		payload := normalizePayload(event.PayloadJSON)
		if !json.Valid(payload) {
			return nil, fmt.Errorf("%w: turn_events[%d].payload_json must be valid JSON", core.ErrInvalidArgument, i)
		}
		source := strings.TrimSpace(event.Source)
		if source == "" {
			source = defaultSyncSource
		}
		occurredAt := event.OccurredAt
		if occurredAt.IsZero() {
			occurredAt = now
		}
		fingerprint := strings.TrimSpace(event.Fingerprint)
		if fingerprint == "" {
			fingerprint = stableID("fp", req.TenantID, req.WorkspaceID, req.SessionID, event.EventKind, source, string(payload))
		}
		idempotencyKey := fmt.Sprintf("%s:%03d:%s", req.IdempotencyKey, i, fingerprint)
		events = append(events, &core.RawEvent{
			ID:             stableID("evt", req.TenantID, source, idempotencyKey),
			TenantID:       req.TenantID,
			WorkspaceID:    req.WorkspaceID,
			SessionID:      req.SessionID,
			ActorID:        req.ActorID,
			EventKind:      event.EventKind,
			Source:         source,
			IdempotencyKey: idempotencyKey,
			Fingerprint:    fingerprint,
			OccurredAt:     occurredAt.UTC(),
			PayloadJSON:    append(json.RawMessage(nil), payload...),
			CreatedAt:      now,
		})
	}
	return events, nil
}

func normalizePayload(payload json.RawMessage) json.RawMessage {
	if len(payload) == 0 {
		return json.RawMessage(`{}`)
	}
	return append(json.RawMessage(nil), payload...)
}

func stableID(prefix string, parts ...string) string {
	h := sha256.New()
	for _, part := range parts {
		_, _ = h.Write([]byte(part))
		_, _ = h.Write([]byte{0})
	}
	sum := h.Sum(nil)
	return prefix + "_" + hex.EncodeToString(sum)[:24]
}

```



<!-- Source: internal/ingest/service_test.go | bytes=5981 | lines=205 | sha16=77803e93b205356b -->

```go
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

```



<!-- Source: internal/kernel/doc.go | bytes=804 | lines=16 | sha16=79b40337a919e494 -->

```go
// ============================================================
// FILE     : internal/kernel/doc.go
// PURPOSE  : Provides package documentation for the concrete VibeGravity service composition.
// LAYER    : application
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : package kernel
// DEPENDS  : internal/core, internal/ingest, internal/recall
// USED_BY  : cmd/server, tests, future Hermes and MCP adapters
// ------------------------------------------------------------
// AGENT_NOTE: Keep this package as orchestration glue; product rules belong in the domain packages it composes.
// ============================================================

// Package kernel composes VibeGravity application services behind the core contract.
package kernel

```



<!-- Source: internal/kernel/service.go | bytes=23991 | lines=739 | sha16=07000eff9f8d3ade -->

```go
// ============================================================
// FILE     : internal/kernel/service.go
// PURPOSE  : Implements core.VibeGravityService by delegating to product services.
// LAYER    : application
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : Dependencies, Service, NewService
// DEPENDS  : encoding/json, internal/core, internal/ingest, internal/recall, internal/store
// USED_BY  : cmd/server, tests, future Hermes and MCP adapters
// ------------------------------------------------------------
// AGENT_NOTE: Do not hide product behavior here; route calls to the package that owns the contract.
// ============================================================

package kernel

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/ingest"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/recall"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/store"
)

// Dependencies contains the product services composed by the kernel.
type Dependencies struct {
	Ingest      *ingest.Service
	Recall      *recall.Assembler
	Notes       store.NoteStore
	Plans       store.PlanStore
	Memories    store.MemoryStore
	Corrections store.CorrectionStore
	Timeline    store.TimelineStore
	Documents   store.DocumentStore
}

// Service is the concrete v1 VibeGravity service.
type Service struct {
	ingest      *ingest.Service
	recall      *recall.Assembler
	notes       store.NoteStore
	plans       store.PlanStore
	memories    store.MemoryStore
	corrections store.CorrectionStore
	timeline    store.TimelineStore
	documents   store.DocumentStore
}

var _ core.VibeGravityService = (*Service)(nil)

// NewService creates the concrete VibeGravity service.
func NewService(deps Dependencies) (*Service, error) {
	if deps.Ingest == nil {
		return nil, fmt.Errorf("%w: ingest service is required", core.ErrInvalidArgument)
	}
	if deps.Recall == nil {
		return nil, fmt.Errorf("%w: recall assembler is required", core.ErrInvalidArgument)
	}
	return &Service{
		ingest:      deps.Ingest,
		recall:      deps.Recall,
		notes:       deps.Notes,
		plans:       deps.Plans,
		memories:    deps.Memories,
		corrections: deps.Corrections,
		timeline:    deps.Timeline,
		documents:   deps.Documents,
	}, nil
}

// Prefetch assembles a next-turn recall pack.
func (s *Service) Prefetch(ctx context.Context, req *core.PrefetchRequest) (*core.PrefetchResponse, error) {
	return s.recall.Prefetch(ctx, req)
}

// SyncTurn records turn events on the hot path.
func (s *Service) SyncTurn(ctx context.Context, req *core.SyncTurnRequest) (*core.SyncTurnResponse, error) {
	return s.ingest.SyncTurn(ctx, req)
}

const documentChunkMaxRunes = 1800

// AddDocument stores a document and its initial lexical retrieval chunks.
func (s *Service) AddDocument(ctx context.Context, req *core.AddDocumentRequest) (*core.AddDocumentResponse, error) {
	if s.documents == nil {
		return nil, fmt.Errorf("%w: add document", core.ErrNotImplemented)
	}
	if req == nil {
		return nil, fmt.Errorf("%w: add document request is required", core.ErrInvalidArgument)
	}
	if err := requireFields(map[string]string{
		"tenant_id":    req.TenantID,
		"workspace_id": req.WorkspaceID,
		"source":       req.Source,
		"title":        req.Title,
		"content":      req.Content,
	}); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	document := &core.Document{
		TenantID:     req.TenantID,
		WorkspaceID:  req.WorkspaceID,
		Source:       req.Source,
		Title:        req.Title,
		Fingerprint:  valueOr(req.Fingerprint, documentFingerprint(req)),
		MetadataJSON: jsonOrEmpty(req.MetadataJSON),
		VersionHint:  req.VersionHint,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	chunks := buildDocumentChunks("", req.Content, now)
	if err := s.documents.AddDocumentWithChunks(ctx, document, chunks); err != nil {
		return nil, err
	}
	chunkIDs := make([]string, 0, len(chunks))
	for _, chunk := range chunks {
		chunkIDs = append(chunkIDs, chunk.ID)
	}
	return &core.AddDocumentResponse{DocumentID: document.ID, ChunkIDs: chunkIDs, Status: "created"}, nil
}

// SearchMemories delegates memory search to storage.
func (s *Service) SearchMemories(ctx context.Context, req *core.SearchMemoriesRequest) (*core.SearchMemoriesResponse, error) {
	if s.memories == nil {
		return nil, fmt.Errorf("%w: search memories", core.ErrNotImplemented)
	}
	if req == nil {
		return nil, fmt.Errorf("%w: search memories request is required", core.ErrInvalidArgument)
	}
	if err := requireFields(map[string]string{
		"tenant_id":    req.TenantID,
		"workspace_id": req.WorkspaceID,
	}); err != nil {
		return nil, err
	}
	if len(req.Scopes) == 0 {
		req.Scopes = []core.MemoryScope{
			core.MemoryScopeAgentPrivate,
			core.MemoryScopeWorkspaceShared,
			core.MemoryScopeSessionScratch,
		}
	}
	if len(req.ArtifactClasses) == 0 {
		req.ArtifactClasses = []core.ArtifactClass{
			core.ArtifactClassContext,
			core.ArtifactClassKnowledge,
			core.ArtifactClassTimeline,
			core.ArtifactClassPlan,
		}
	}
	return s.memories.SearchMemories(ctx, req)
}

// SearchDocuments delegates document search to storage.
func (s *Service) SearchDocuments(ctx context.Context, req *core.SearchDocumentsRequest) (*core.SearchDocumentsResponse, error) {
	if s.documents == nil {
		return nil, fmt.Errorf("%w: search documents", core.ErrNotImplemented)
	}
	if req == nil {
		return nil, fmt.Errorf("%w: search documents request is required", core.ErrInvalidArgument)
	}
	if err := requireFields(map[string]string{
		"tenant_id":    req.TenantID,
		"workspace_id": req.WorkspaceID,
	}); err != nil {
		return nil, err
	}
	return s.documents.SearchDocuments(ctx, req)
}

// AddNote creates a human-authored recall control note.
func (s *Service) AddNote(ctx context.Context, req *core.AddNoteRequest) (*core.AddNoteResponse, error) {
	if s.notes == nil {
		return nil, fmt.Errorf("%w: add note", core.ErrNotImplemented)
	}
	if req == nil {
		return nil, fmt.Errorf("%w: add note request is required", core.ErrInvalidArgument)
	}
	if err := requireFields(map[string]string{
		"tenant_id":       req.TenantID,
		"workspace_id":    req.WorkspaceID,
		"scope":           string(req.Scope),
		"owner_entity_id": req.OwnerEntityID,
		"text":            req.Text,
	}); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	note := &core.Note{
		TenantID:      req.TenantID,
		WorkspaceID:   req.WorkspaceID,
		NoteKind:      valueOr(req.NoteKind, "operator"),
		Scope:         req.Scope,
		OwnerEntityID: req.OwnerEntityID,
		Text:          req.Text,
		Pinned:        req.Pinned,
		ExpiresAt:     req.ExpiresAt,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := s.notes.AddNote(ctx, note); err != nil {
		return nil, err
	}
	return &core.AddNoteResponse{NoteID: note.ID, Status: "created"}, nil
}

// CreatePlan creates a structured plan and its initial items.
func (s *Service) CreatePlan(ctx context.Context, req *core.CreatePlanRequest) (*core.CreatePlanResponse, error) {
	if s.plans == nil {
		return nil, fmt.Errorf("%w: create plan", core.ErrNotImplemented)
	}
	if req == nil {
		return nil, fmt.Errorf("%w: create plan request is required", core.ErrInvalidArgument)
	}
	if err := requireFields(map[string]string{
		"tenant_id":       req.TenantID,
		"workspace_id":    req.WorkspaceID,
		"title":           req.Title,
		"scope":           string(req.Scope),
		"owner_entity_id": req.OwnerEntityID,
	}); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	plan := &core.Plan{
		TenantID:      req.TenantID,
		WorkspaceID:   req.WorkspaceID,
		Title:         req.Title,
		Status:        valueOr(req.Status, "active"),
		Scope:         req.Scope,
		OwnerEntityID: req.OwnerEntityID,
		EvidenceJSON:  jsonOrEmpty(req.EvidenceJSON),
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	items := make([]*core.PlanItem, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, &core.PlanItem{
			ID:           item.ID,
			Title:        item.Title,
			Status:       valueOr(item.Status, "open"),
			EvidenceJSON: jsonOrEmpty(item.EvidenceJSON),
			CreatedAt:    now,
			UpdatedAt:    now,
		})
	}
	if err := s.plans.CreatePlan(ctx, plan, items); err != nil {
		return nil, err
	}
	itemIDs := make([]string, 0, len(items))
	for _, item := range items {
		itemIDs = append(itemIDs, item.ID)
	}
	return &core.CreatePlanResponse{PlanID: plan.ID, ItemIDs: itemIDs, Status: "created"}, nil
}

// UpdatePlan updates a structured plan and optionally replaces provided items.
func (s *Service) UpdatePlan(ctx context.Context, req *core.UpdatePlanRequest) (*core.UpdatePlanResponse, error) {
	if s.plans == nil {
		return nil, fmt.Errorf("%w: update plan", core.ErrNotImplemented)
	}
	if req == nil {
		return nil, fmt.Errorf("%w: update plan request is required", core.ErrInvalidArgument)
	}
	if err := requireFields(map[string]string{
		"tenant_id":    req.TenantID,
		"workspace_id": req.WorkspaceID,
		"plan_id":      req.PlanID,
	}); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	plan := &core.Plan{
		ID:           req.PlanID,
		TenantID:     req.TenantID,
		WorkspaceID:  req.WorkspaceID,
		EvidenceJSON: req.EvidenceJSON,
		UpdatedAt:    now,
	}
	if req.Title != nil {
		plan.Title = strings.TrimSpace(*req.Title)
		if plan.Title == "" {
			return nil, fmt.Errorf("%w: title cannot be empty", core.ErrInvalidArgument)
		}
	}
	if req.Status != nil {
		plan.Status = strings.TrimSpace(*req.Status)
		if plan.Status == "" {
			return nil, fmt.Errorf("%w: status cannot be empty", core.ErrInvalidArgument)
		}
	}
	items := make([]*core.PlanItem, 0, len(req.Items))
	if req.Items != nil {
		for _, item := range req.Items {
			title := strings.TrimSpace(item.Title)
			if title == "" {
				return nil, fmt.Errorf("%w: plan item title is required", core.ErrInvalidArgument)
			}
			items = append(items, &core.PlanItem{
				ID:           item.ID,
				Title:        title,
				Status:       valueOr(item.Status, "open"),
				EvidenceJSON: jsonOrEmpty(item.EvidenceJSON),
				CreatedAt:    now,
				UpdatedAt:    now,
			})
		}
	} else {
		items = nil
	}
	if err := s.plans.UpdatePlan(ctx, plan, items); err != nil {
		return nil, err
	}
	return &core.UpdatePlanResponse{PlanID: req.PlanID, Status: "updated"}, nil
}

// CorrectMemory records human correction intent and applies an operator-driven supersession.
func (s *Service) CorrectMemory(ctx context.Context, req *core.CorrectMemoryRequest) (*core.CorrectMemoryResponse, error) {
	if s.memories == nil || s.corrections == nil {
		return nil, fmt.Errorf("%w: correct memory", core.ErrNotImplemented)
	}
	supersessions, ok := s.memories.(correctionSupersessionStore)
	if !ok {
		return nil, fmt.Errorf("%w: correct memory supersession store", core.ErrNotImplemented)
	}
	if req == nil {
		return nil, fmt.Errorf("%w: correct memory request is required", core.ErrInvalidArgument)
	}
	if err := requireFields(map[string]string{
		"tenant_id":       req.TenantID,
		"workspace_id":    req.WorkspaceID,
		"memory_id":       req.MemoryID,
		"operator_id":     req.OperatorID,
		"idempotency_key": req.IdempotencyKey,
		"correction_text": req.CorrectionText,
	}); err != nil {
		return nil, err
	}
	memory, err := s.memories.GetMemory(ctx, req.MemoryID)
	if err != nil {
		return nil, err
	}
	if memory == nil {
		return nil, core.ErrNotFound
	}
	if memory.TenantID != req.TenantID || memory.WorkspaceID != req.WorkspaceID {
		return nil, core.ErrNotFound
	}
	if memory.Status != core.MemoryStatusActive || !memory.LatestFlag {
		return nil, fmt.Errorf("%w: correction target memory must be active latest", core.ErrConflict)
	}
	now := time.Now().UTC()
	payload, err := correctionPayload(req)
	if err != nil {
		return nil, err
	}
	event := &core.RawEvent{
		TenantID:       req.TenantID,
		WorkspaceID:    req.WorkspaceID,
		SessionID:      "correction:" + req.MemoryID,
		ActorID:        req.OperatorID,
		EventKind:      "memory_correction",
		Source:         "operator_correction",
		IdempotencyKey: req.IdempotencyKey,
		Fingerprint:    correctionFingerprint(req),
		OccurredAt:     now,
		PayloadJSON:    payload,
		CreatedAt:      now,
	}
	correction := &core.MemoryCorrection{
		TenantID:       req.TenantID,
		WorkspaceID:    req.WorkspaceID,
		MemoryID:       req.MemoryID,
		OperatorID:     req.OperatorID,
		IdempotencyKey: req.IdempotencyKey,
		CorrectionText: strings.TrimSpace(req.CorrectionText),
		EvidenceJSON:   jsonOrEmpty(req.EvidenceJSON),
		Status:         "recorded",
		CreatedAt:      now,
	}
	recorded, err := s.corrections.RecordMemoryCorrection(ctx, event, correction)
	if err != nil {
		return nil, err
	}
	replacement, trace, edge, err := buildCorrectionSupersession(memory, recorded, now)
	if err != nil {
		return nil, err
	}
	if err := supersessions.CreateMemoryWithTraceAndUpdateEdge(ctx, replacement, trace, edge); err != nil {
		return nil, err
	}
	return &core.CorrectMemoryResponse{
		MemoryID:           req.MemoryID,
		RawEventID:         recorded.RawEventID,
		CorrectionID:       recorded.ID,
		CorrectionRecorded: true,
		TraceWritten:       true,
		Status:             "applied",
	}, nil
}

type correctionSupersessionStore interface {
	CreateMemoryWithTraceAndUpdateEdge(ctx context.Context, memory *core.Memory, trace *core.MemoryTrace, edge *core.MemoryEdge) error
}

const (
	timelineDefaultLimit = 50
	timelineMaxLimit     = 100
)

// GetTimeline assembles a read-only operator timeline over existing artifacts.
func (s *Service) GetTimeline(ctx context.Context, req *core.GetTimelineRequest) (*core.GetTimelineResponse, error) {
	if s.timeline == nil {
		return nil, fmt.Errorf("%w: get timeline", core.ErrNotImplemented)
	}
	if req == nil {
		return nil, fmt.Errorf("%w: get timeline request is required", core.ErrInvalidArgument)
	}
	if err := requireFields(map[string]string{
		"tenant_id":    req.TenantID,
		"workspace_id": req.WorkspaceID,
		"entity_id":    req.EntityID,
	}); err != nil {
		return nil, err
	}
	if req.From != nil && req.To != nil && req.From.After(*req.To) {
		return nil, fmt.Errorf("%w: from must be before to", core.ErrInvalidArgument)
	}
	normalized := *req
	scopes, err := normalizeTimelineScopes(req.Scopes)
	if err != nil {
		return nil, err
	}
	normalized.Scopes = scopes
	if normalized.Limit == 0 {
		normalized.Limit = timelineDefaultLimit
	}
	if normalized.Limit < 0 || normalized.Limit > timelineMaxLimit {
		return nil, fmt.Errorf("%w: limit must be between 1 and %d", core.ErrInvalidArgument, timelineMaxLimit)
	}
	return s.timeline.GetTimeline(ctx, &normalized)
}

// ExplainMemory delegates provenance lookup to storage.
func (s *Service) ExplainMemory(ctx context.Context, req *core.ExplainMemoryRequest) (*core.ExplainMemoryResponse, error) {
	if s.memories == nil {
		return nil, fmt.Errorf("%w: explain memory", core.ErrNotImplemented)
	}
	if req == nil {
		return nil, fmt.Errorf("%w: explain memory request is required", core.ErrInvalidArgument)
	}
	if err := requireFields(map[string]string{
		"tenant_id":    req.TenantID,
		"workspace_id": req.WorkspaceID,
		"memory_id":    req.MemoryID,
	}); err != nil {
		return nil, err
	}
	return s.memories.ExplainMemory(ctx, req)
}

func requireFields(fields map[string]string) error {
	for name, value := range fields {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%w: %s is required", core.ErrInvalidArgument, name)
		}
	}
	return nil
}

func jsonOrEmpty(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 {
		return json.RawMessage(`{}`)
	}
	return raw
}

func valueOr(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func documentFingerprint(req *core.AddDocumentRequest) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{
		req.TenantID,
		req.WorkspaceID,
		req.Source,
		req.Title,
		req.Content,
	}, "\x00")))
	return fmt.Sprintf("sha256:%x", sum[:])
}

func correctionPayload(req *core.CorrectMemoryRequest) (json.RawMessage, error) {
	payload := struct {
		MemoryID       string          `json:"memory_id"`
		OperatorID     string          `json:"operator_id"`
		CorrectionText string          `json:"correction_text"`
		EvidenceJSON   json.RawMessage `json:"evidence_json"`
	}{
		MemoryID:       req.MemoryID,
		OperatorID:     req.OperatorID,
		CorrectionText: strings.TrimSpace(req.CorrectionText),
		EvidenceJSON:   jsonOrEmpty(req.EvidenceJSON),
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal correction payload: %w", err)
	}
	return data, nil
}

func correctionFingerprint(req *core.CorrectMemoryRequest) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{
		req.TenantID,
		req.WorkspaceID,
		req.MemoryID,
		req.OperatorID,
		req.IdempotencyKey,
		strings.TrimSpace(req.CorrectionText),
		string(jsonOrEmpty(req.EvidenceJSON)),
	}, "\x00")))
	return fmt.Sprintf("sha256:%x", sum[:])
}

func buildCorrectionSupersession(target *core.Memory, correction *core.MemoryCorrection, correctedAt time.Time) (*core.Memory, *core.MemoryTrace, *core.MemoryEdge, error) {
	if target == nil {
		return nil, nil, nil, fmt.Errorf("%w: correction target memory is required", core.ErrInvalidArgument)
	}
	if correction == nil {
		return nil, nil, nil, fmt.Errorf("%w: recorded correction is required", core.ErrInvalidArgument)
	}
	replacement := &core.Memory{
		ID:            correctionSupersessionID(target.ID, correction.IdempotencyKey),
		TenantID:      target.TenantID,
		WorkspaceID:   target.WorkspaceID,
		Scope:         target.Scope,
		GroupID:       target.GroupID,
		OwnerEntityID: target.OwnerEntityID,
		Kind:          target.Kind,
		ArtifactClass: target.ArtifactClass,
		Text:          strings.TrimSpace(correction.CorrectionText),
		Fingerprint:   correctionSupersessionFingerprint(target, correction),
		Confidence:    1.0,
		Status:        core.MemoryStatusActive,
		ValidFrom:     correctedAt.UTC(),
		LatestFlag:    true,
		MetadataJSON:  correctionSupersessionMetadata(target, correction),
		CreatedAt:     correctedAt.UTC(),
		UpdatedAt:     correctedAt.UTC(),
	}
	appliedOperations, err := correctionSupersessionOperation(target, replacement, correction)
	if err != nil {
		return nil, nil, nil, err
	}
	trace := &core.MemoryTrace{
		MemoryID:               replacement.ID,
		RawEventIDs:            []string{correction.RawEventID},
		ReasoningJobID:         "correction:" + correction.ID,
		ReasoningStage:         "operator_correction",
		CandidateSnapshotJSON:  json.RawMessage(`{"source":"operator_correction"}`),
		AppliedOperationsJSON:  appliedOperations,
		OperatorCorrectionFlag: true,
		RelatedDocumentIDs:     []string{},
		CreatedAt:              correctedAt.UTC(),
	}
	edge := &core.MemoryEdge{
		FromMemoryID:   replacement.ID,
		ToMemoryID:     target.ID,
		EdgeKind:       core.EdgeKindUpdates,
		Confidence:     1.0,
		CreatedByJobID: "correction:" + correction.ID,
		CreatedAt:      correctedAt.UTC(),
	}
	return replacement, trace, edge, nil
}

func correctionSupersessionID(targetMemoryID, idempotencyKey string) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{targetMemoryID, idempotencyKey}, "\x00")))
	return fmt.Sprintf("mem_corr_%x", sum[:12])
}

func correctionSupersessionFingerprint(target *core.Memory, correction *core.MemoryCorrection) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{
		target.TenantID,
		target.WorkspaceID,
		target.ID,
		string(target.Scope),
		target.OwnerEntityID,
		string(target.Kind),
		string(target.ArtifactClass),
		correction.IdempotencyKey,
		strings.TrimSpace(correction.CorrectionText),
	}, "\x00")))
	return fmt.Sprintf("fp_%x", sum[:16])
}

func correctionSupersessionMetadata(target *core.Memory, correction *core.MemoryCorrection) json.RawMessage {
	data, err := json.Marshal(struct {
		Source         string          `json:"source"`
		CorrectionID   string          `json:"correction_id"`
		TargetMemoryID string          `json:"target_memory_id"`
		OperatorID     string          `json:"operator_id"`
		EvidenceJSON   json.RawMessage `json:"evidence_json"`
	}{
		Source:         "operator_correction",
		CorrectionID:   correction.ID,
		TargetMemoryID: target.ID,
		OperatorID:     correction.OperatorID,
		EvidenceJSON:   jsonOrEmpty(correction.EvidenceJSON),
	})
	if err != nil {
		return json.RawMessage(`{"source":"operator_correction"}`)
	}
	return data
}

func correctionSupersessionOperation(target *core.Memory, replacement *core.Memory, correction *core.MemoryCorrection) (json.RawMessage, error) {
	data, err := json.Marshal([]struct {
		OperationID     string        `json:"operation_id"`
		Kind            string        `json:"kind"`
		MemoryID        string        `json:"memory_id"`
		TargetMemoryID  string        `json:"target_memory_id"`
		RawEventIDs     []string      `json:"raw_event_ids"`
		OperatorID      string        `json:"operator_id"`
		EdgeKind        core.EdgeKind `json:"edge_kind"`
		CorrectionID    string        `json:"correction_id"`
		CorrectionState string        `json:"correction_state"`
	}{{
		OperationID:     "operator_correction:" + correction.ID,
		Kind:            "update_memory",
		MemoryID:        replacement.ID,
		TargetMemoryID:  target.ID,
		RawEventIDs:     []string{correction.RawEventID},
		OperatorID:      correction.OperatorID,
		EdgeKind:        core.EdgeKindUpdates,
		CorrectionID:    correction.ID,
		CorrectionState: "applied",
	}})
	if err != nil {
		return nil, fmt.Errorf("marshal correction supersession operation: %w", err)
	}
	return json.RawMessage(data), nil
}

func normalizeTimelineScopes(scopes []core.MemoryScope) ([]core.MemoryScope, error) {
	if len(scopes) == 0 {
		return []core.MemoryScope{
			core.MemoryScopeAgentPrivate,
			core.MemoryScopeWorkspaceShared,
			core.MemoryScopeSessionScratch,
		}, nil
	}
	seen := make(map[core.MemoryScope]struct{}, len(scopes))
	normalized := make([]core.MemoryScope, 0, len(scopes))
	for _, scope := range scopes {
		switch scope {
		case core.MemoryScopeAgentPrivate, core.MemoryScopeWorkspaceShared, core.MemoryScopeSessionScratch:
			if _, ok := seen[scope]; !ok {
				seen[scope] = struct{}{}
				normalized = append(normalized, scope)
			}
		case core.MemoryScopeGroupShared:
			continue
		default:
			return nil, fmt.Errorf("%w: unsupported timeline scope %q", core.ErrInvalidArgument, scope)
		}
	}
	return normalized, nil
}

func buildDocumentChunks(documentID, content string, now time.Time) []*core.DocumentChunk {
	paragraphs := strings.Split(strings.TrimSpace(content), "\n\n")
	chunks := make([]*core.DocumentChunk, 0, len(paragraphs))
	var builder strings.Builder
	flush := func() {
		text := strings.TrimSpace(builder.String())
		if text == "" {
			builder.Reset()
			return
		}
		chunks = append(chunks, &core.DocumentChunk{
			DocumentID:     documentID,
			ChunkIndex:     len(chunks),
			Text:           text,
			MetadataJSON:   json.RawMessage(`{}`),
			EmbeddingModel: "pending",
			EmbeddingDims:  0,
			CreatedAt:      now,
			UpdatedAt:      now,
		})
		builder.Reset()
	}
	for _, paragraph := range paragraphs {
		paragraph = strings.TrimSpace(paragraph)
		if paragraph == "" {
			continue
		}
		if builder.Len() > 0 && builder.Len()+len(paragraph)+2 > documentChunkMaxRunes {
			flush()
		}
		if len([]rune(paragraph)) > documentChunkMaxRunes {
			flush()
			for _, part := range splitRunes(paragraph, documentChunkMaxRunes) {
				builder.WriteString(part)
				flush()
			}
			continue
		}
		if builder.Len() > 0 {
			builder.WriteString("\n\n")
		}
		builder.WriteString(paragraph)
	}
	flush()
	return chunks
}

func splitRunes(text string, maxRunes int) []string {
	runes := []rune(text)
	parts := make([]string, 0, (len(runes)/maxRunes)+1)
	for start := 0; start < len(runes); start += maxRunes {
		end := start + maxRunes
		if end > len(runes) {
			end = len(runes)
		}
		parts = append(parts, string(runes[start:end]))
	}
	return parts
}

```



<!-- Source: internal/kernel/service_test.go | bytes=18655 | lines=560 | sha16=5bb741f0b98b04c5 -->

```go
// ============================================================
// FILE     : internal/kernel/service_test.go
// PURPOSE  : Verifies kernel-level document and plan API behavior.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : kernel service tests
// DEPENDS  : context, testing, internal/core
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: Keep these tests focused on service composition, not PostgreSQL details.
// ============================================================

package kernel

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

func TestAddDocumentStoresDocumentAndChunks(t *testing.T) {
	t.Parallel()

	documents := &fakeDocumentStore{}
	service := &Service{documents: documents}

	resp, err := service.AddDocument(context.Background(), &core.AddDocumentRequest{
		TenantID:    "tenant_1",
		WorkspaceID: "workspace_1",
		Source:      "operator_upload",
		Title:       "Runtime Notes",
		Content:     strings.Repeat("A", documentChunkMaxRunes+5),
	})
	if err != nil {
		t.Fatalf("AddDocument returned error: %v", err)
	}
	if resp.DocumentID != "doc_test" || resp.Status != "created" {
		t.Fatalf("unexpected response: %#v", resp)
	}
	if documents.document == nil || documents.document.Fingerprint == "" {
		t.Fatalf("document was not stored with a fingerprint: %#v", documents.document)
	}
	if len(documents.chunks) != 2 || len(resp.ChunkIDs) != 2 {
		t.Fatalf("expected long content to become two chunks, chunks=%d resp=%#v", len(documents.chunks), resp)
	}
	if documents.chunks[0].DocumentID != "doc_test" || documents.chunks[0].ChunkIndex != 0 {
		t.Fatalf("first chunk not linked/indexed correctly: %#v", documents.chunks[0])
	}
	if documents.atomicWrites != 1 || documents.separateDocumentWrites != 0 || documents.separateChunkWrites != 0 {
		t.Fatalf("document ingestion must use one atomic store call: %#v", documents)
	}
}

func TestAddDocumentDoesNotReportSuccessWhenAtomicStoreFails(t *testing.T) {
	t.Parallel()

	storeErr := errors.New("chunk insert failed")
	documents := &fakeDocumentStore{atomicErr: storeErr}
	service := &Service{documents: documents}

	resp, err := service.AddDocument(context.Background(), &core.AddDocumentRequest{
		TenantID:    "tenant_1",
		WorkspaceID: "workspace_1",
		Source:      "operator_upload",
		Title:       "Runtime Notes",
		Content:     "chunk body",
	})
	if !errors.Is(err, storeErr) {
		t.Fatalf("AddDocument error = %v, want %v", err, storeErr)
	}
	if resp != nil {
		t.Fatalf("AddDocument returned response on failed atomic write: %#v", resp)
	}
	if documents.document != nil || len(documents.chunks) != 0 {
		t.Fatalf("failed atomic write must not be treated as committed, document=%#v chunks=%#v", documents.document, documents.chunks)
	}
}

func TestUpdatePlanDelegatesPatchAndItems(t *testing.T) {
	t.Parallel()

	plans := &fakePlanStore{}
	service := &Service{plans: plans}
	title := "Ship Work Pack 03"
	status := "active"

	resp, err := service.UpdatePlan(context.Background(), &core.UpdatePlanRequest{
		TenantID:    "tenant_1",
		WorkspaceID: "workspace_1",
		PlanID:      "plan_1",
		Title:       &title,
		Status:      &status,
		Items: []core.PlanItemInput{{
			Title: "Wire document API",
		}},
	})
	if err != nil {
		t.Fatalf("UpdatePlan returned error: %v", err)
	}
	if resp.PlanID != "plan_1" || resp.Status != "updated" {
		t.Fatalf("unexpected response: %#v", resp)
	}
	if plans.updatedPlan == nil || plans.updatedPlan.Title != title || plans.updatedPlan.Status != status {
		t.Fatalf("plan update was not delegated: %#v", plans.updatedPlan)
	}
	if len(plans.updatedItems) != 1 || plans.updatedItems[0].Status != "open" {
		t.Fatalf("plan items were not normalized/delegated: %#v", plans.updatedItems)
	}
}

func TestCorrectMemoryValidatesRequiredFields(t *testing.T) {
	t.Parallel()

	service := &Service{
		memories:    &fakeKernelMemoryStore{},
		corrections: &fakeCorrectionStore{},
	}

	_, err := service.CorrectMemory(context.Background(), &core.CorrectMemoryRequest{
		TenantID:       "tenant_1",
		WorkspaceID:    "workspace_1",
		MemoryID:       "mem_1",
		OperatorID:     "operator_1",
		IdempotencyKey: "correction_1",
		CorrectionText: "   ",
	})
	if !errors.Is(err, core.ErrInvalidArgument) {
		t.Fatalf("expected invalid argument for blank correction text, got %v", err)
	}
}

func TestCorrectMemoryReturnsNotFoundForMissingMemory(t *testing.T) {
	t.Parallel()

	service := &Service{
		memories:    &fakeKernelMemoryStore{err: core.ErrNotFound},
		corrections: &fakeCorrectionStore{},
	}

	_, err := service.CorrectMemory(context.Background(), validCorrectionRequest())
	if !errors.Is(err, core.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestCorrectMemoryRecordsRawEventAndCorrection(t *testing.T) {
	t.Parallel()

	corrections := &fakeCorrectionStore{}
	service := &Service{
		memories:    &fakeKernelMemoryStore{memory: validCorrectionTargetMemory()},
		corrections: corrections,
	}

	resp, err := service.CorrectMemory(context.Background(), validCorrectionRequest())
	if err != nil {
		t.Fatalf("CorrectMemory returned error: %v", err)
	}
	if resp.MemoryID != "mem_1" || resp.RawEventID != "evt_correction" || resp.CorrectionID != "corr_1" || !resp.CorrectionRecorded || !resp.TraceWritten || resp.Status != "applied" {
		t.Fatalf("unexpected correction response: %#v", resp)
	}
	if corrections.event == nil || corrections.event.EventKind != "memory_correction" {
		t.Fatalf("raw correction event was not recorded: %#v", corrections.event)
	}
	if corrections.event.Source != "operator_correction" || corrections.event.ActorID != "operator_1" {
		t.Fatalf("raw correction event source/actor mismatch: %#v", corrections.event)
	}
	if corrections.correction == nil || corrections.correction.CorrectionText != "Use the newer fact." {
		t.Fatalf("operator-visible correction artifact was not recorded: %#v", corrections.correction)
	}
	var payload map[string]any
	if err := json.Unmarshal(corrections.event.PayloadJSON, &payload); err != nil {
		t.Fatalf("correction event payload is not JSON: %v", err)
	}
	if payload["memory_id"] != "mem_1" || payload["correction_text"] != "Use the newer fact." {
		t.Fatalf("correction payload lost intent: %#v", payload)
	}
	if corrections.correction == nil || corrections.correction.Status != "recorded" {
		t.Fatalf("correction artifact should be recorded before supersession: %#v", corrections.correction)
	}
	memories := service.memories.(*fakeKernelMemoryStore)
	if memories.updateMemory == nil || memories.updateTrace == nil || memories.updateEdge == nil {
		t.Fatalf("correction did not apply graph supersession: memory=%#v trace=%#v edge=%#v", memories.updateMemory, memories.updateTrace, memories.updateEdge)
	}
	if memories.updateMemory.Text != "Use the newer fact." || memories.updateMemory.Scope != core.MemoryScopeWorkspaceShared || memories.updateMemory.OwnerEntityID != "agent:hermes-main" {
		t.Fatalf("replacement memory did not preserve target boundary and corrected text: %#v", memories.updateMemory)
	}
	if memories.updateTrace.RawEventIDs[0] != "evt_correction" || !memories.updateTrace.OperatorCorrectionFlag || memories.updateTrace.ReasoningStage != "operator_correction" {
		t.Fatalf("correction trace did not preserve operator provenance: %#v", memories.updateTrace)
	}
	if memories.updateEdge.FromMemoryID != memories.updateMemory.ID || memories.updateEdge.ToMemoryID != "mem_1" || memories.updateEdge.EdgeKind != core.EdgeKindUpdates {
		t.Fatalf("correction updates edge mismatch: %#v", memories.updateEdge)
	}
}

func TestCorrectMemoryIdempotentRetryReturnsRecordedArtifact(t *testing.T) {
	t.Parallel()

	corrections := &fakeCorrectionStore{
		recorded: &core.MemoryCorrection{
			ID:             "corr_existing",
			MemoryID:       "mem_1",
			RawEventID:     "evt_existing",
			IdempotencyKey: "correction_1",
			CorrectionText: "Use the newer fact.",
			OperatorID:     "operator_1",
			Status:         "recorded",
		},
	}
	service := &Service{
		memories:    &fakeKernelMemoryStore{memory: validCorrectionTargetMemory()},
		corrections: corrections,
	}

	resp, err := service.CorrectMemory(context.Background(), validCorrectionRequest())
	if err != nil {
		t.Fatalf("CorrectMemory retry returned error: %v", err)
	}
	if resp.RawEventID != "evt_existing" || resp.CorrectionID != "corr_existing" || resp.Status != "applied" || !resp.TraceWritten {
		t.Fatalf("idempotent retry did not return existing correction artifact: %#v", resp)
	}
}

func TestCorrectMemoryRejectsNonLatestTargetBeforeRecordingCorrection(t *testing.T) {
	t.Parallel()

	target := validCorrectionTargetMemory()
	target.Status = core.MemoryStatusSuperseded
	target.LatestFlag = false
	corrections := &fakeCorrectionStore{}
	service := &Service{
		memories:    &fakeKernelMemoryStore{memory: target},
		corrections: corrections,
	}

	_, err := service.CorrectMemory(context.Background(), validCorrectionRequest())
	if !errors.Is(err, core.ErrConflict) {
		t.Fatalf("expected ErrConflict for non-latest correction target, got %v", err)
	}
	if corrections.event != nil || corrections.correction != nil {
		t.Fatalf("non-latest target must not record correction side effects: event=%#v correction=%#v", corrections.event, corrections.correction)
	}
}

func TestCorrectMemoryDoesNotReportSuccessWhenSupersessionFails(t *testing.T) {
	t.Parallel()

	storeErr := errors.New("supersession failed")
	memories := &fakeKernelMemoryStore{memory: validCorrectionTargetMemory(), updateErr: storeErr}
	service := &Service{
		memories:    memories,
		corrections: &fakeCorrectionStore{},
	}

	resp, err := service.CorrectMemory(context.Background(), validCorrectionRequest())
	if !errors.Is(err, storeErr) {
		t.Fatalf("CorrectMemory error = %v, want %v", err, storeErr)
	}
	if resp != nil {
		t.Fatalf("failed supersession must not return success: %#v", resp)
	}
}

func TestGetTimelineDefaultsScopesLimitAndDelegates(t *testing.T) {
	t.Parallel()

	timeline := &fakeTimelineStore{}
	service := &Service{timeline: timeline}

	resp, err := service.GetTimeline(context.Background(), &core.GetTimelineRequest{
		TenantID:    "tenant_1",
		WorkspaceID: "workspace_1",
		EntityID:    "agent:hermes-main",
	})
	if err != nil {
		t.Fatalf("GetTimeline returned error: %v", err)
	}
	if resp == nil || len(resp.Items) != 1 || resp.Items[0].ID != "tl_1" {
		t.Fatalf("unexpected timeline response: %#v", resp)
	}
	if timeline.req == nil || timeline.req.Limit != timelineDefaultLimit {
		t.Fatalf("timeline request was not defaulted: %#v", timeline.req)
	}
	wantScopes := []core.MemoryScope{
		core.MemoryScopeAgentPrivate,
		core.MemoryScopeWorkspaceShared,
		core.MemoryScopeSessionScratch,
	}
	if !sameMemoryScopes(timeline.req.Scopes, wantScopes) {
		t.Fatalf("timeline scopes = %#v, want %#v", timeline.req.Scopes, wantScopes)
	}
}

func TestGetTimelineRejectsInvalidScopeAndLimit(t *testing.T) {
	t.Parallel()

	service := &Service{timeline: &fakeTimelineStore{}}

	_, err := service.GetTimeline(context.Background(), &core.GetTimelineRequest{
		TenantID:    "tenant_1",
		WorkspaceID: "workspace_1",
		EntityID:    "agent:hermes-main",
		Scopes:      []core.MemoryScope{"public"},
	})
	if !errors.Is(err, core.ErrInvalidArgument) {
		t.Fatalf("expected invalid scope error, got %v", err)
	}

	_, err = service.GetTimeline(context.Background(), &core.GetTimelineRequest{
		TenantID:    "tenant_1",
		WorkspaceID: "workspace_1",
		EntityID:    "agent:hermes-main",
		Limit:       timelineMaxLimit + 1,
	})
	if !errors.Is(err, core.ErrInvalidArgument) {
		t.Fatalf("expected invalid limit error, got %v", err)
	}
}

func TestGetTimelineExcludesGroupSharedUntilMembershipFiltering(t *testing.T) {
	t.Parallel()

	timeline := &fakeTimelineStore{}
	service := &Service{timeline: timeline}

	_, err := service.GetTimeline(context.Background(), &core.GetTimelineRequest{
		TenantID:    "tenant_1",
		WorkspaceID: "workspace_1",
		EntityID:    "agent:hermes-main",
		Scopes: []core.MemoryScope{
			core.MemoryScopeGroupShared,
			core.MemoryScopeWorkspaceShared,
		},
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("GetTimeline returned error: %v", err)
	}
	if !sameMemoryScopes(timeline.req.Scopes, []core.MemoryScope{core.MemoryScopeWorkspaceShared}) {
		t.Fatalf("group_shared should be excluded from timeline scopes, got %#v", timeline.req.Scopes)
	}
}

func TestExplainMemoryDelegatesVisibilityFields(t *testing.T) {
	t.Parallel()

	memories := &fakeKernelMemoryStore{}
	service := &Service{memories: memories}

	_, err := service.ExplainMemory(context.Background(), &core.ExplainMemoryRequest{
		TenantID:        "tenant_1",
		WorkspaceID:     "workspace_1",
		MemoryID:        "mem_1",
		EntityID:        "agent:hermes-main",
		VisibleGroupIDs: []string{"group_design"},
	})
	if err != nil {
		t.Fatalf("ExplainMemory returned error: %v", err)
	}
	if memories.explainReq == nil || memories.explainReq.EntityID != "agent:hermes-main" {
		t.Fatalf("explain visibility fields were not delegated: %#v", memories.explainReq)
	}
	if got := memories.explainReq.VisibleGroupIDs; len(got) != 1 || got[0] != "group_design" {
		t.Fatalf("visible group ids were not delegated: %#v", memories.explainReq)
	}
}

func validCorrectionRequest() *core.CorrectMemoryRequest {
	return &core.CorrectMemoryRequest{
		TenantID:       "tenant_1",
		WorkspaceID:    "workspace_1",
		MemoryID:       "mem_1",
		OperatorID:     "operator_1",
		IdempotencyKey: "correction_1",
		CorrectionText: "Use the newer fact.",
	}
}

func validCorrectionTargetMemory() *core.Memory {
	return &core.Memory{
		ID:            "mem_1",
		TenantID:      "tenant_1",
		WorkspaceID:   "workspace_1",
		Scope:         core.MemoryScopeWorkspaceShared,
		OwnerEntityID: "agent:hermes-main",
		Kind:          core.MemoryKindFact,
		ArtifactClass: core.ArtifactClassKnowledge,
		Text:          "Old fact.",
		Confidence:    0.7,
		Status:        core.MemoryStatusActive,
		LatestFlag:    true,
	}
}

func sameMemoryScopes(got, want []core.MemoryScope) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

type fakeDocumentStore struct {
	document               *core.Document
	chunks                 []*core.DocumentChunk
	atomicErr              error
	atomicWrites           int
	separateDocumentWrites int
	separateChunkWrites    int
}

func (s *fakeDocumentStore) AddDocumentWithChunks(_ context.Context, document *core.Document, chunks []*core.DocumentChunk) error {
	s.atomicWrites++
	if s.atomicErr != nil {
		return s.atomicErr
	}
	document.ID = "doc_test"
	for i, chunk := range chunks {
		chunk.DocumentID = document.ID
		chunk.ID = "chunk_test_" + string(rune('a'+i))
	}
	s.document = document
	s.chunks = chunks
	return nil
}

func (s *fakeDocumentStore) AddDocument(_ context.Context, document *core.Document) error {
	s.separateDocumentWrites++
	document.ID = "doc_test"
	s.document = document
	return nil
}

func (s *fakeDocumentStore) AddDocumentChunks(_ context.Context, chunks []*core.DocumentChunk) error {
	s.separateChunkWrites++
	for i, chunk := range chunks {
		chunk.ID = "chunk_test_" + string(rune('a'+i))
	}
	s.chunks = chunks
	return nil
}

func (s *fakeDocumentStore) SearchDocuments(context.Context, *core.SearchDocumentsRequest) (*core.SearchDocumentsResponse, error) {
	return nil, core.ErrNotImplemented
}

type fakePlanStore struct {
	updatedPlan  *core.Plan
	updatedItems []*core.PlanItem
}

func (s *fakePlanStore) CreatePlan(context.Context, *core.Plan, []*core.PlanItem) error {
	return core.ErrNotImplemented
}

func (s *fakePlanStore) UpdatePlan(_ context.Context, plan *core.Plan, items []*core.PlanItem) error {
	s.updatedPlan = plan
	s.updatedItems = items
	return nil
}

func (s *fakePlanStore) GetActivePlans(context.Context, *core.GetActivePlansRequest) ([]*core.Plan, error) {
	return nil, core.ErrNotImplemented
}

type fakeKernelMemoryStore struct {
	memory       *core.Memory
	updateMemory *core.Memory
	updateTrace  *core.MemoryTrace
	updateEdge   *core.MemoryEdge
	explainReq   *core.ExplainMemoryRequest
	err          error
	updateErr    error
}

func (s *fakeKernelMemoryStore) UpsertMemory(context.Context, *core.Memory) error {
	return core.ErrNotImplemented
}

func (s *fakeKernelMemoryStore) GetMemory(context.Context, string) (*core.Memory, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.memory, nil
}

func (s *fakeKernelMemoryStore) SearchMemories(context.Context, *core.SearchMemoriesRequest) (*core.SearchMemoriesResponse, error) {
	return nil, core.ErrNotImplemented
}

func (s *fakeKernelMemoryStore) UpsertMemoryEdge(context.Context, *core.MemoryEdge) error {
	return core.ErrNotImplemented
}

func (s *fakeKernelMemoryStore) WriteMemoryTrace(context.Context, *core.MemoryTrace) error {
	return core.ErrNotImplemented
}

func (s *fakeKernelMemoryStore) ExplainMemory(_ context.Context, req *core.ExplainMemoryRequest) (*core.ExplainMemoryResponse, error) {
	s.explainReq = req
	return &core.ExplainMemoryResponse{MemoryID: req.MemoryID}, nil
}

func (s *fakeKernelMemoryStore) CreateMemoryWithTraceAndUpdateEdge(_ context.Context, memory *core.Memory, trace *core.MemoryTrace, edge *core.MemoryEdge) error {
	if s.updateErr != nil {
		return s.updateErr
	}
	s.updateMemory = memory
	s.updateTrace = trace
	s.updateEdge = edge
	return nil
}

type fakeCorrectionStore struct {
	event      *core.RawEvent
	correction *core.MemoryCorrection
	recorded   *core.MemoryCorrection
}

type fakeTimelineStore struct {
	req *core.GetTimelineRequest
}

func (s *fakeTimelineStore) GetTimeline(_ context.Context, req *core.GetTimelineRequest) (*core.GetTimelineResponse, error) {
	reqCopy := *req
	reqCopy.Scopes = append([]core.MemoryScope(nil), req.Scopes...)
	s.req = &reqCopy
	return &core.GetTimelineResponse{
		Items: []core.TimelineItem{{
			ID:            "tl_1",
			Kind:          core.MemoryKindCorrection,
			ArtifactClass: core.ArtifactClassTimeline,
			Text:          "Correction for memory mem_1: Use the newer fact.",
			MemoryID:      "mem_1",
			RawEventID:    "evt_1",
		}},
	}, nil
}

func (s *fakeCorrectionStore) RecordMemoryCorrection(_ context.Context, event *core.RawEvent, correction *core.MemoryCorrection) (*core.MemoryCorrection, error) {
	s.event = event
	s.correction = correction
	if s.recorded != nil {
		return s.recorded, nil
	}
	correction.ID = "corr_1"
	correction.RawEventID = "evt_correction"
	correction.Status = "recorded"
	return correction, nil
}

```



<!-- Source: internal/recall/assembler.go | bytes=19005 | lines=720 | sha16=01d9ebe708bd528b -->

```go
// ============================================================
// FILE     : internal/recall/assembler.go
// PURPOSE  : Assembles budget-aware typed recall blocks for prefetch requests.
// LAYER    : application
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : Dependencies, Assembler, NewAssembler
// DEPENDS  : internal/core, internal/store
// USED_BY  : internal/kernel, tests
// ------------------------------------------------------------
// AGENT_NOTE: Keep recall typed before rendering so Hermes, MCP, and HTTP share one meaning.
// ============================================================

package recall

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/store"
)

const (
	defaultBudgetTokens = 2200
	smallBudgetTokens   = 1000
	richBudgetTokens    = 4000
)

// Dependencies collects optional recall candidate stores.
type Dependencies struct {
	Notes     store.NoteStore
	Plans     store.PlanStore
	Memories  store.MemoryStore
	Documents store.DocumentStore
	Profiles  store.ProfileStore
	Summaries store.SessionSummaryStore
	Groups    store.GroupStore
	Freshness FreshnessProvider
	Clock     func() time.Time
}

// Assembler builds prefetch recall packs from typed candidate pools.
type Assembler struct {
	notes     store.NoteStore
	plans     store.PlanStore
	memories  store.MemoryStore
	documents store.DocumentStore
	profiles  store.ProfileStore
	summaries store.SessionSummaryStore
	groups    store.GroupStore
	freshness FreshnessProvider
	clock     func() time.Time
}

type candidateBlock struct {
	block  core.RecallBlock
	source string
	rank   float64
}

// NewAssembler builds a recall assembler. Missing stores are treated as degraded mode.
func NewAssembler(deps Dependencies) *Assembler {
	clock := deps.Clock
	if clock == nil {
		clock = time.Now
	}
	return &Assembler{
		notes:     deps.Notes,
		plans:     deps.Plans,
		memories:  deps.Memories,
		documents: deps.Documents,
		profiles:  deps.Profiles,
		summaries: deps.Summaries,
		groups:    deps.Groups,
		freshness: deps.Freshness,
		clock:     clock,
	}
}

// Prefetch assembles a budget-aware typed recall pack.
func (a *Assembler) Prefetch(ctx context.Context, req *core.PrefetchRequest) (*core.PrefetchResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("%w: prefetch request is required", core.ErrInvalidArgument)
	}
	if err := validatePrefetchRequest(req); err != nil {
		return nil, err
	}

	groupIDs, err := a.visibleGroupIDs(ctx, req)
	if err != nil {
		return nil, err
	}
	controlScopes := baseVisibleScopes()
	memoryScopes := visibleScopes(groupIDs)
	candidates := make([]candidateBlock, 0, 16)

	candidates, err = a.addPinnedNotes(ctx, req, controlScopes, candidates)
	if err != nil {
		return nil, err
	}
	candidates, err = a.addActivePlans(ctx, req, controlScopes, candidates)
	if err != nil {
		return nil, err
	}
	candidates, err = a.addProfiles(ctx, req, candidates)
	if err != nil {
		return nil, err
	}
	candidates, err = a.addSessionSummary(ctx, req, candidates)
	if err != nil {
		return nil, err
	}
	candidates, err = a.addMemories(ctx, req, memoryScopes, groupIDs, candidates)
	if err != nil {
		return nil, err
	}
	candidates, err = a.addDocuments(ctx, req, candidates)
	if err != nil {
		return nil, err
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].block.Priority != candidates[j].block.Priority {
			return candidates[i].block.Priority > candidates[j].block.Priority
		}
		return candidates[i].rank > candidates[j].rank
	})

	blocks, sources, estimated := packCandidates(candidates, budgetFor(req))
	freshness := a.recallFreshness(ctx, req)
	blocks = applyRecallFreshness(blocks, freshness)
	degradedReasons := append(a.degradedReasons(), freshness.Reasons...)
	degradedReasons = uniqueStrings(degradedReasons)
	return &core.PrefetchResponse{
		Blocks: blocks,
		Meta: core.RecallMeta{
			EstimatedTokens:     estimated,
			Sources:             sources,
			Freshness:           freshness.Freshness,
			FreshnessLagSeconds: freshness.LagSeconds,
			Degraded:            len(degradedReasons) > 0,
			DegradedReasons:     degradedReasons,
		},
	}, nil
}

func (a *Assembler) recallFreshness(ctx context.Context, req *core.PrefetchRequest) Freshness {
	if a.freshness == nil {
		return Freshness{Freshness: "stored"}
	}
	state, err := a.freshness.RecallFreshness(ctx, req)
	if err != nil {
		return Freshness{
			Freshness:       "degraded",
			Reasons:         []string{"recall_freshness_probe_unavailable"},
			AffectedSources: derivedRecallSources(),
		}
	}
	if strings.TrimSpace(state.Freshness) == "" {
		state.Freshness = "stored"
	}
	return state
}

func (a *Assembler) degradedReasons() []string {
	var reasons []string
	if a.notes == nil {
		reasons = append(reasons, "notes_unavailable")
	}
	if a.plans == nil {
		reasons = append(reasons, "plans_unavailable")
	}
	if a.memories == nil {
		reasons = append(reasons, "memories_unavailable")
	}
	if a.documents == nil {
		reasons = append(reasons, "documents_unavailable")
	}
	if a.profiles == nil {
		reasons = append(reasons, "profiles_unavailable")
	}
	if a.summaries == nil {
		reasons = append(reasons, "session_summaries_unavailable")
	}
	if a.groups == nil {
		reasons = append(reasons, "group_membership_unavailable")
	}
	return reasons
}

func validatePrefetchRequest(req *core.PrefetchRequest) error {
	required := map[string]string{
		"tenant_id":    req.TenantID,
		"workspace_id": req.WorkspaceID,
		"session_id":   req.SessionID,
		"actor_id":     req.ActorID,
	}
	for name, value := range required {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%w: %s is required", core.ErrInvalidArgument, name)
		}
	}
	return nil
}

func visibleScopes(groupIDs []string) []core.MemoryScope {
	scopes := baseVisibleScopes()
	if len(groupIDs) > 0 {
		scopes = append(scopes, core.MemoryScopeGroupShared)
	}
	return scopes
}

func baseVisibleScopes() []core.MemoryScope {
	return []core.MemoryScope{
		core.MemoryScopeAgentPrivate,
		core.MemoryScopeWorkspaceShared,
		core.MemoryScopeSessionScratch,
	}
}

func (a *Assembler) visibleGroupIDs(ctx context.Context, req *core.PrefetchRequest) ([]string, error) {
	if a.groups == nil {
		return []string{}, nil
	}
	memberships, err := a.groups.ListMembershipsForEntity(ctx, req.TenantID, req.WorkspaceID, req.ActorID)
	if errors.Is(err, core.ErrNotFound) {
		return []string{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("list visible memory groups: %w", err)
	}
	return membershipGroupIDs(memberships), nil
}

func membershipGroupIDs(memberships []*core.MemoryGroupMembership) []string {
	if len(memberships) == 0 {
		return []string{}
	}
	groupIDs := make([]string, 0, len(memberships))
	seen := make(map[string]struct{}, len(memberships))
	for _, membership := range memberships {
		if membership == nil {
			continue
		}
		groupID := strings.TrimSpace(membership.GroupID)
		if groupID == "" {
			continue
		}
		if _, ok := seen[groupID]; ok {
			continue
		}
		seen[groupID] = struct{}{}
		groupIDs = append(groupIDs, groupID)
	}
	return groupIDs
}

func (a *Assembler) addPinnedNotes(ctx context.Context, req *core.PrefetchRequest, scopes []core.MemoryScope, candidates []candidateBlock) ([]candidateBlock, error) {
	if a.notes == nil {
		return candidates, nil
	}
	notes, err := a.notes.ListPinnedNotes(ctx, &core.ListPinnedNotesRequest{
		TenantID:      req.TenantID,
		WorkspaceID:   req.WorkspaceID,
		OwnerEntityID: req.ActorID,
		Scopes:        scopes,
	})
	if errors.Is(err, core.ErrNotFound) {
		return candidates, nil
	}
	if err != nil {
		return nil, fmt.Errorf("list pinned notes: %w", err)
	}
	now := a.clock().UTC()
	for _, note := range notes {
		if note == nil || strings.TrimSpace(note.Text) == "" {
			continue
		}
		if note.ExpiresAt != nil && !note.ExpiresAt.After(now) {
			continue
		}
		candidates = append(candidates, candidateBlock{
			block: core.RecallBlock{
				ID:            note.ID,
				Kind:          "pinned_note",
				Priority:      100,
				Text:          note.Text,
				Scope:         note.Scope,
				Source:        "notes",
				SourceID:      note.ID,
				Status:        "pinned",
				Freshness:     "stored",
				OwnerEntityID: note.OwnerEntityID,
			},
			source: "notes",
		})
	}
	return candidates, nil
}

func (a *Assembler) addActivePlans(ctx context.Context, req *core.PrefetchRequest, scopes []core.MemoryScope, candidates []candidateBlock) ([]candidateBlock, error) {
	if a.plans == nil {
		return candidates, nil
	}
	plans, err := a.plans.GetActivePlans(ctx, &core.GetActivePlansRequest{
		TenantID:      req.TenantID,
		WorkspaceID:   req.WorkspaceID,
		OwnerEntityID: req.ActorID,
		Scopes:        scopes,
	})
	if errors.Is(err, core.ErrNotFound) {
		return candidates, nil
	}
	if err != nil {
		return nil, fmt.Errorf("list active plans: %w", err)
	}
	for _, plan := range plans {
		if plan == nil || strings.TrimSpace(plan.Title) == "" || suppressedPlanStatus(plan.Status) {
			continue
		}
		candidates = append(candidates, candidateBlock{
			block: core.RecallBlock{
				ID:            plan.ID,
				Kind:          "active_plan",
				Priority:      95,
				Text:          plan.Title,
				Scope:         plan.Scope,
				Source:        "plans",
				SourceID:      plan.ID,
				Status:        plan.Status,
				Freshness:     "stored",
				OwnerEntityID: plan.OwnerEntityID,
			},
			source: "plans",
		})
	}
	return candidates, nil
}

func (a *Assembler) addProfiles(ctx context.Context, req *core.PrefetchRequest, candidates []candidateBlock) ([]candidateBlock, error) {
	if a.profiles == nil {
		return candidates, nil
	}
	profileTargets := []struct {
		entityID string
		scope    core.MemoryScope
	}{
		{entityID: req.ActorID, scope: core.MemoryScopeAgentPrivate},
		{entityID: "workspace:" + req.WorkspaceID, scope: core.MemoryScopeWorkspaceShared},
	}
	for _, target := range profileTargets {
		profile, err := a.profiles.GetProfile(ctx, target.entityID, target.scope)
		if errors.Is(err, core.ErrNotFound) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("get profile: %w", err)
		}
		if profile == nil {
			continue
		}
		if text := jsonText(profile.StaticJSON); text != "" {
			candidates = append(candidates, candidateBlock{
				block: core.RecallBlock{
					ID:        target.entityID,
					Kind:      "profile_static",
					Priority:  90,
					Text:      text,
					Scope:     target.scope,
					Source:    "profile",
					SourceID:  target.entityID,
					Status:    "snapshot",
					Freshness: "stored",
				},
				source: "profile",
			})
		}
		if text := jsonText(profile.DynamicJSON); text != "" {
			candidates = append(candidates, candidateBlock{
				block: core.RecallBlock{
					ID:        target.entityID,
					Kind:      "profile_dynamic",
					Priority:  85,
					Text:      text,
					Scope:     target.scope,
					Source:    "profile",
					SourceID:  target.entityID,
					Status:    "snapshot",
					Freshness: "stored",
				},
				source: "profile",
			})
		}
	}
	return candidates, nil
}

func (a *Assembler) addSessionSummary(ctx context.Context, req *core.PrefetchRequest, candidates []candidateBlock) ([]candidateBlock, error) {
	if a.summaries == nil {
		return candidates, nil
	}
	summary, err := a.summaries.GetSessionSummary(ctx, req.SessionID)
	if errors.Is(err, core.ErrNotFound) {
		return candidates, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get session summary: %w", err)
	}
	if summary != nil && strings.TrimSpace(summary.SummaryText) != "" {
		candidates = append(candidates, candidateBlock{
			block: core.RecallBlock{
				ID:        summary.ID,
				Kind:      "session_summary",
				Priority:  80,
				Text:      summary.SummaryText,
				Scope:     core.MemoryScopeSessionScratch,
				Source:    "session_summaries",
				SourceID:  summary.ID,
				Status:    "summary",
				Freshness: "stored",
			},
			source: "session_summaries",
		})
	}
	return candidates, nil
}

func (a *Assembler) addMemories(ctx context.Context, req *core.PrefetchRequest, scopes []core.MemoryScope, groupIDs []string, candidates []candidateBlock) ([]candidateBlock, error) {
	if a.memories == nil || strings.TrimSpace(req.Query) == "" {
		return candidates, nil
	}
	resp, err := a.memories.SearchMemories(ctx, &core.SearchMemoriesRequest{
		TenantID:        req.TenantID,
		WorkspaceID:     req.WorkspaceID,
		OwnerEntityID:   req.ActorID,
		VisibleGroupIDs: groupIDs,
		Query:           req.Query,
		Scopes:          scopes,
		ArtifactClasses: []core.ArtifactClass{
			core.ArtifactClassContext,
			core.ArtifactClassKnowledge,
			core.ArtifactClassTimeline,
			core.ArtifactClassPlan,
		},
	})
	if errors.Is(err, core.ErrNotFound) {
		return candidates, nil
	}
	if err != nil {
		return nil, fmt.Errorf("search memories: %w", err)
	}
	if resp == nil {
		return candidates, nil
	}
	for i, memory := range resp.Memories {
		if i >= 8 {
			break
		}
		if strings.TrimSpace(memory.Text) == "" || !memory.LatestFlag {
			continue
		}
		rank := scoreMemoryCandidate(req.Query, memory, a.clock().UTC())
		candidates = append(candidates, candidateBlock{
			block: core.RecallBlock{
				ID:            memory.MemoryID,
				Kind:          "memory",
				Priority:      70 + int(rank),
				Text:          memory.Text,
				Scope:         memory.Scope,
				Source:        "memories",
				SourceID:      memory.MemoryID,
				Status:        memoryRecallStatus(memory),
				Freshness:     "stored",
				OwnerEntityID: memory.OwnerEntityID,
			},
			source: "memories",
			rank:   rank,
		})
	}
	return candidates, nil
}

func (a *Assembler) addDocuments(ctx context.Context, req *core.PrefetchRequest, candidates []candidateBlock) ([]candidateBlock, error) {
	if a.documents == nil || strings.TrimSpace(req.Query) == "" {
		return candidates, nil
	}
	resp, err := a.documents.SearchDocuments(ctx, &core.SearchDocumentsRequest{
		TenantID:    req.TenantID,
		WorkspaceID: req.WorkspaceID,
		Query:       req.Query,
	})
	if errors.Is(err, core.ErrNotFound) {
		return candidates, nil
	}
	if err != nil {
		return nil, fmt.Errorf("search documents: %w", err)
	}
	if resp == nil {
		return candidates, nil
	}
	for i, chunk := range resp.Chunks {
		if i >= 5 {
			break
		}
		if strings.TrimSpace(chunk.Text) == "" {
			continue
		}
		rank := scoreDocumentCandidate(req.Query, chunk)
		candidates = append(candidates, candidateBlock{
			block: core.RecallBlock{
				ID:        chunk.ChunkID,
				Kind:      "document",
				Priority:  60 + int(rank),
				Text:      chunk.Text,
				Scope:     core.MemoryScopeWorkspaceShared,
				Source:    "documents",
				SourceID:  chunk.DocumentID,
				Status:    "supporting_context",
				Freshness: "stored",
			},
			source: "documents",
			rank:   rank,
		})
	}
	return candidates, nil
}

func packCandidates(candidates []candidateBlock, budget int) ([]core.RecallBlock, []string, int) {
	blocks := make([]core.RecallBlock, 0, len(candidates))
	sources := make([]string, 0, 8)
	seenSource := make(map[string]struct{})
	seenText := make(map[string]struct{})
	estimated := 0

	for _, candidate := range candidates {
		text := strings.TrimSpace(candidate.block.Text)
		if text == "" {
			continue
		}
		dedupKey := strings.ToLower(text)
		if _, ok := seenText[dedupKey]; ok {
			continue
		}
		remaining := budget - estimated
		if remaining <= 0 {
			break
		}
		blockBudget := minInt(remaining, maxBlockTokens(budget))
		candidate.block.Text = truncateToBudget(text, blockBudget)
		if candidate.block.Text == "" {
			continue
		}
		tokenCost := estimateTokens(candidate.block.Text)
		if estimated+tokenCost > budget {
			continue
		}
		blocks = append(blocks, candidate.block)
		seenText[dedupKey] = struct{}{}
		estimated += tokenCost
		if candidate.source != "" {
			if _, ok := seenSource[candidate.source]; !ok {
				sources = append(sources, candidate.source)
				seenSource[candidate.source] = struct{}{}
			}
		}
	}
	return blocks, sources, estimated
}

func maxBlockTokens(budget int) int {
	if budget <= 8 {
		return budget
	}
	limit := budget * 45 / 100
	if limit < 8 {
		return 8
	}
	return limit
}

func budgetFor(req *core.PrefetchRequest) int {
	if req.BudgetTokens > 0 {
		return req.BudgetTokens
	}
	switch strings.ToLower(strings.TrimSpace(req.Mode)) {
	case "small":
		return smallBudgetTokens
	case "rich":
		return richBudgetTokens
	default:
		return defaultBudgetTokens
	}
}

func estimateTokens(text string) int {
	words := strings.Fields(text)
	if len(words) == 0 {
		return 0
	}
	return len(words) + len(words)/3 + 1
}

func scoreMemoryCandidate(query string, memory core.MemoryResult, now time.Time) float64 {
	score := lexicalOverlapScore(query, memory.Text) * 20
	if memory.Confidence > 0 {
		score += memory.Confidence * 3
	}
	if !memory.ValidFrom.IsZero() {
		age := now.Sub(memory.ValidFrom)
		switch {
		case age < 0:
			score++
		case age <= 24*time.Hour:
			score += 3
		case age <= 7*24*time.Hour:
			score += 1.5
		}
	}
	switch memory.Kind {
	case core.MemoryKindConstraint, core.MemoryKindDecision, core.MemoryKindProcedure, core.MemoryKindTaskState:
		score++
	}
	return score
}

func scoreDocumentCandidate(query string, chunk core.DocumentChunkResult) float64 {
	score := lexicalOverlapScore(query, chunk.Text) * 12
	if chunk.Score > 0 {
		score += chunk.Score * 4
	}
	return score
}

func memoryRecallStatus(memory core.MemoryResult) string {
	if !memory.LatestFlag {
		return "suppressed"
	}
	return "active"
}

func lexicalOverlapScore(query, text string) float64 {
	queryTerms := recallTerms(query)
	if len(queryTerms) == 0 {
		return 0
	}
	textTerms := recallTerms(text)
	matches := 0
	for term := range queryTerms {
		if _, ok := textTerms[term]; ok {
			matches++
		}
	}
	return float64(matches) / float64(len(queryTerms))
}

func recallTerms(text string) map[string]struct{} {
	terms := make(map[string]struct{})
	for _, raw := range strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	}) {
		if len(raw) < 3 {
			continue
		}
		terms[raw] = struct{}{}
	}
	return terms
}

func minInt(left, right int) int {
	if left < right {
		return left
	}
	return right
}

func truncateToBudget(text string, budget int) string {
	if budget <= 0 {
		return ""
	}
	if estimateTokens(text) <= budget {
		return text
	}
	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}
	maxWords := (budget - 1) * 3 / 4
	if maxWords <= 0 {
		return ""
	}
	if maxWords > len(words) {
		maxWords = len(words)
	}
	return strings.Join(words[:maxWords], " ") + "..."
}

func jsonText(raw []byte) string {
	text := strings.TrimSpace(string(raw))
	if text == "" || text == "{}" || text == "null" {
		return ""
	}
	return text
}

func suppressedPlanStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "done", "completed", "archived", "deleted", "cancelled":
		return true
	default:
		return false
	}
}

```



<!-- Source: internal/recall/assembler_test.go | bytes=18185 | lines=552 | sha16=3f556040623b812a -->

```go
// ============================================================
// FILE     : internal/recall/assembler_test.go
// PURPOSE  : Verifies prefetch typed block assembly, priority, scopes, and budget behavior.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : recall assembler behavior tests
// DEPENDS  : internal/recall, internal/core
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: Recall tests should assert typed blocks before any Hermes text rendering exists.
// ============================================================

package recall

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

func TestAssemblerPrefetch_PrioritizesManualControls(t *testing.T) {
	t.Parallel()

	notes := &fakeNoteStore{notes: []*core.Note{{
		ID:            "note_1",
		Text:          "Always prefer the Go-first plan.",
		Scope:         core.MemoryScopeWorkspaceShared,
		OwnerEntityID: "workspace:workspace_1",
		Pinned:        true,
	}}}
	plans := &fakePlanStore{plans: []*core.Plan{{
		ID:            "plan_1",
		Title:         "Implement sync_turn before worker reasoning.",
		Status:        "active",
		Scope:         core.MemoryScopeAgentPrivate,
		OwnerEntityID: "agent:hermes-main",
	}}}
	assembler := NewAssembler(Dependencies{
		Notes: notes,
		Plans: plans,
		Profiles: &fakeProfileStore{profiles: map[string]*core.Profile{
			"agent:hermes-main|agent_private": {
				EntityID:   "agent:hermes-main",
				Scope:      core.MemoryScopeAgentPrivate,
				StaticJSON: json.RawMessage(`{"style":"brief"}`),
			},
		}},
	})

	resp, err := assembler.Prefetch(context.Background(), testPrefetchRequest())
	if err != nil {
		t.Fatalf("Prefetch returned error: %v", err)
	}

	gotKinds := recallKinds(resp.Blocks)
	wantKinds := []string{"pinned_note", "active_plan", "profile_static"}
	if !reflect.DeepEqual(gotKinds, wantKinds) {
		t.Fatalf("unexpected block kinds: got %v want %v", gotKinds, wantKinds)
	}
	if resp.Meta.EstimatedTokens <= 0 {
		t.Fatalf("expected positive token estimate")
	}
	if !reflect.DeepEqual(resp.Meta.Sources, []string{"notes", "plans", "profile"}) {
		t.Fatalf("unexpected sources: %v", resp.Meta.Sources)
	}
	if resp.Blocks[0].Scope != core.MemoryScopeWorkspaceShared || resp.Blocks[0].Source != "notes" || resp.Blocks[0].SourceID != "note_1" || resp.Blocks[0].Status != "pinned" || resp.Blocks[0].Freshness != "stored" {
		t.Fatalf("pinned note block did not expose trust metadata: %#v", resp.Blocks[0])
	}
	if resp.Blocks[1].Scope != core.MemoryScopeAgentPrivate || resp.Blocks[1].Source != "plans" || resp.Blocks[1].SourceID != "plan_1" || resp.Blocks[1].Status != "active" {
		t.Fatalf("active plan block did not expose trust metadata: %#v", resp.Blocks[1])
	}
	if notes.lastReq.OwnerEntityID != "agent:hermes-main" || notes.lastReq.TenantID != "tenant_1" {
		t.Fatalf("pinned notes request was not actor scoped: %#v", notes.lastReq)
	}
	if plans.lastReq.OwnerEntityID != "agent:hermes-main" || plans.lastReq.TenantID != "tenant_1" {
		t.Fatalf("active plans request was not actor scoped: %#v", plans.lastReq)
	}
}

func TestAssemblerPrefetch_UsesScopeAwareMemorySearch(t *testing.T) {
	t.Parallel()

	memories := &fakeMemoryStore{resp: &core.SearchMemoriesResponse{
		Memories: []core.MemoryResult{{
			MemoryID:      "mem_1",
			Text:          "VibeGravity keeps private and shared memory separate.",
			Scope:         core.MemoryScopeWorkspaceShared,
			OwnerEntityID: "workspace:workspace_1",
			LatestFlag:    true,
		}},
	}}
	assembler := NewAssembler(Dependencies{
		Memories: memories,
	})

	resp, err := assembler.Prefetch(context.Background(), testPrefetchRequest())
	if err != nil {
		t.Fatalf("Prefetch returned error: %v", err)
	}

	if len(resp.Blocks) != 1 || resp.Blocks[0].Kind != "memory" {
		t.Fatalf("expected one memory block, got %#v", resp.Blocks)
	}
	wantScopes := []core.MemoryScope{
		core.MemoryScopeAgentPrivate,
		core.MemoryScopeWorkspaceShared,
		core.MemoryScopeSessionScratch,
	}
	if !reflect.DeepEqual(memories.lastReq.Scopes, wantScopes) {
		t.Fatalf("unexpected scopes: got %v want %v", memories.lastReq.Scopes, wantScopes)
	}
	if memories.lastReq.OwnerEntityID != "agent:hermes-main" {
		t.Fatalf("memory search request was not actor scoped: %#v", memories.lastReq)
	}
	if resp.Blocks[0].Source != "memories" || resp.Blocks[0].SourceID != "mem_1" || resp.Blocks[0].Scope != core.MemoryScopeWorkspaceShared || resp.Blocks[0].Status != "active" {
		t.Fatalf("memory block did not expose source and scope metadata: %#v", resp.Blocks[0])
	}
}

func TestAssemblerPrefetch_IncludesGroupSharedMemoriesForMemberActor(t *testing.T) {
	t.Parallel()

	groupID := "group_design"
	memories := &fakeMemoryStore{resp: &core.SearchMemoriesResponse{
		Memories: []core.MemoryResult{{
			MemoryID:   "mem_group",
			Text:       "Design group agreed to keep MCP as the first external protocol.",
			Scope:      core.MemoryScopeGroupShared,
			GroupID:    &groupID,
			LatestFlag: true,
		}},
	}}
	assembler := NewAssembler(Dependencies{
		Memories: memories,
		Groups: &fakeGroupStore{memberships: []*core.MemoryGroupMembership{{
			GroupID:  "group_design",
			EntityID: "agent:hermes-main",
		}}},
	})

	resp, err := assembler.Prefetch(context.Background(), testPrefetchRequest())
	if err != nil {
		t.Fatalf("Prefetch returned error: %v", err)
	}

	if len(resp.Blocks) != 1 || resp.Blocks[0].Kind != "memory" {
		t.Fatalf("expected group memory block, got %#v", resp.Blocks)
	}
	wantScopes := []core.MemoryScope{
		core.MemoryScopeAgentPrivate,
		core.MemoryScopeWorkspaceShared,
		core.MemoryScopeSessionScratch,
		core.MemoryScopeGroupShared,
	}
	if !reflect.DeepEqual(memories.lastReq.Scopes, wantScopes) {
		t.Fatalf("unexpected scopes: got %v want %v", memories.lastReq.Scopes, wantScopes)
	}
	if !reflect.DeepEqual(memories.lastReq.VisibleGroupIDs, []string{"group_design"}) {
		t.Fatalf("expected visible group ids, got %#v", memories.lastReq.VisibleGroupIDs)
	}
	if resp.Blocks[0].Scope != core.MemoryScopeGroupShared {
		t.Fatalf("group memory block should expose group scope: %#v", resp.Blocks[0])
	}
}

func TestAssemblerPrefetch_MarksMissingStoresAsDegraded(t *testing.T) {
	t.Parallel()

	assembler := NewAssembler(Dependencies{
		Notes: &fakeNoteStore{notes: []*core.Note{{
			ID:            "note_1",
			Text:          "Keep Hermes Memory trust loop visible.",
			Scope:         core.MemoryScopeWorkspaceShared,
			OwnerEntityID: "workspace:workspace_1",
			Pinned:        true,
		}}},
	})

	resp, err := assembler.Prefetch(context.Background(), testPrefetchRequest())
	if err != nil {
		t.Fatalf("Prefetch returned error: %v", err)
	}

	if !resp.Meta.Degraded {
		t.Fatalf("expected degraded metadata when stores are unavailable: %#v", resp.Meta)
	}
	if !containsString(resp.Meta.DegradedReasons, "memories_unavailable") {
		t.Fatalf("expected memories_unavailable reason, got %#v", resp.Meta.DegradedReasons)
	}
}

func TestAssemblerPrefetch_MarksDerivedRecallStaleFromBacklogFreshness(t *testing.T) {
	t.Parallel()

	lagSeconds := int64(120)
	assembler := NewAssembler(Dependencies{
		Notes: &fakeNoteStore{notes: []*core.Note{{
			ID:            "note_1",
			Text:          "Manual operator guardrail stays current.",
			Scope:         core.MemoryScopeWorkspaceShared,
			OwnerEntityID: "workspace:workspace_1",
			Pinned:        true,
		}}},
		Memories: &fakeMemoryStore{resp: &core.SearchMemoriesResponse{
			Memories: []core.MemoryResult{{
				MemoryID:      "mem_1",
				Text:          "Worker-derived memory may lag during Codex outage.",
				Scope:         core.MemoryScopeAgentPrivate,
				OwnerEntityID: "agent:hermes-main",
				LatestFlag:    true,
			}},
		}},
		Freshness: fakeFreshnessProvider{state: Freshness{
			Freshness:       "stale",
			LagSeconds:      &lagSeconds,
			Reasons:         []string{"worker_backlog_stale", "codex_or_worker_retry_backlog"},
			AffectedSources: []string{"memories"},
		}},
	})

	resp, err := assembler.Prefetch(context.Background(), testPrefetchRequest())
	if err != nil {
		t.Fatalf("Prefetch returned error: %v", err)
	}

	if !resp.Meta.Degraded {
		t.Fatalf("expected degraded meta from freshness provider: %#v", resp.Meta)
	}
	if resp.Meta.Freshness != "stale" || resp.Meta.FreshnessLagSeconds == nil || *resp.Meta.FreshnessLagSeconds != lagSeconds {
		t.Fatalf("unexpected freshness metadata: %#v", resp.Meta)
	}
	if !containsString(resp.Meta.DegradedReasons, "worker_backlog_stale") || !containsString(resp.Meta.DegradedReasons, "codex_or_worker_retry_backlog") {
		t.Fatalf("missing freshness degraded reasons: %#v", resp.Meta.DegradedReasons)
	}
	if resp.Blocks[0].Kind != "pinned_note" || resp.Blocks[0].Freshness != "stored" {
		t.Fatalf("manual note freshness should remain stored: %#v", resp.Blocks)
	}
	if resp.Blocks[1].Kind != "memory" || resp.Blocks[1].Freshness != "stale" {
		t.Fatalf("derived memory freshness should be stale: %#v", resp.Blocks)
	}
}

func TestBacklogFreshnessProvider_MarksRetryBacklogStale(t *testing.T) {
	t.Parallel()

	lagSeconds := int64(90)
	provider := BacklogFreshnessProvider{
		Jobs: fakeJobMetricsStore{metrics: &core.JobBacklogMetrics{
			Counts: core.JobStatusCounts{
				Queued:      1,
				ReadyQueued: 1,
			},
			OldestQueuedAgeSeconds:  &lagSeconds,
			RetryableQueuedAttempts: 1,
		}},
		StaleAfter: time.Minute,
	}

	state, err := provider.RecallFreshness(context.Background(), testPrefetchRequest())
	if err != nil {
		t.Fatalf("RecallFreshness returned error: %v", err)
	}
	if state.Freshness != "stale" || state.LagSeconds == nil || *state.LagSeconds != lagSeconds {
		t.Fatalf("unexpected backlog freshness state: %#v", state)
	}
	if !containsString(state.Reasons, "worker_backlog_stale") || !containsString(state.Reasons, "codex_or_worker_retry_backlog") {
		t.Fatalf("unexpected backlog freshness reasons: %#v", state.Reasons)
	}
}

func TestBacklogFreshnessProvider_MarksLongRunningJobsStale(t *testing.T) {
	t.Parallel()

	queuedLagSeconds := int64(45)
	runningLagSeconds := int64(180)
	provider := BacklogFreshnessProvider{
		Jobs: fakeJobMetricsStore{metrics: &core.JobBacklogMetrics{
			Counts: core.JobStatusCounts{
				Queued:      1,
				ReadyQueued: 1,
				Running:     1,
			},
			OldestQueuedAgeSeconds:  &queuedLagSeconds,
			OldestRunningAgeSeconds: &runningLagSeconds,
		}},
		StaleAfter: time.Minute,
	}

	state, err := provider.RecallFreshness(context.Background(), testPrefetchRequest())
	if err != nil {
		t.Fatalf("RecallFreshness returned error: %v", err)
	}
	if state.Freshness != "stale" || state.LagSeconds == nil || *state.LagSeconds != runningLagSeconds {
		t.Fatalf("expected running job lag to drive stale freshness: %#v", state)
	}
	if containsString(state.Reasons, "worker_backlog_stale") {
		t.Fatalf("queued job younger than threshold should not mark backlog stale: %#v", state.Reasons)
	}
	if !containsString(state.Reasons, "worker_running_stale") {
		t.Fatalf("missing running stale reason: %#v", state.Reasons)
	}
}

func TestAssemblerPrefetch_TruncatesToBudget(t *testing.T) {
	t.Parallel()

	assembler := NewAssembler(Dependencies{
		Notes: &fakeNoteStore{notes: []*core.Note{{
			Text:   "one two three four five six seven eight nine ten eleven twelve",
			Pinned: true,
		}}},
	})
	req := testPrefetchRequest()
	req.BudgetTokens = 4

	resp, err := assembler.Prefetch(context.Background(), req)
	if err != nil {
		t.Fatalf("Prefetch returned error: %v", err)
	}
	if len(resp.Blocks) != 1 {
		t.Fatalf("expected one truncated block, got %d", len(resp.Blocks))
	}
	if resp.Meta.EstimatedTokens > req.BudgetTokens {
		t.Fatalf("estimated tokens exceeded budget: got %d budget %d", resp.Meta.EstimatedTokens, req.BudgetTokens)
	}
}

func TestAssemblerPrefetch_PreservesPlanWhenPinnedNoteIsLong(t *testing.T) {
	t.Parallel()

	assembler := NewAssembler(Dependencies{
		Notes: &fakeNoteStore{notes: []*core.Note{{
			Text:   strings.Repeat("manual guardrail ", 120),
			Pinned: true,
		}}},
		Plans: &fakePlanStore{plans: []*core.Plan{{
			Title:  "Finish recall token budgeting before graph quality work.",
			Status: "active",
		}}},
	})
	req := testPrefetchRequest()
	req.BudgetTokens = 40

	resp, err := assembler.Prefetch(context.Background(), req)
	if err != nil {
		t.Fatalf("Prefetch returned error: %v", err)
	}

	gotKinds := recallKinds(resp.Blocks)
	wantKinds := []string{"pinned_note", "active_plan"}
	if !reflect.DeepEqual(gotKinds, wantKinds) {
		t.Fatalf("unexpected block kinds: got %v want %v", gotKinds, wantKinds)
	}
	if resp.Meta.EstimatedTokens > req.BudgetTokens {
		t.Fatalf("estimated tokens exceeded budget: got %d budget %d", resp.Meta.EstimatedTokens, req.BudgetTokens)
	}
}

func TestAssemblerPrefetch_RanksMemoryByRelevanceConfidenceAndRecency(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 24, 0, 0, 0, 0, time.UTC)
	memories := &fakeMemoryStore{resp: &core.SearchMemoriesResponse{
		Memories: []core.MemoryResult{
			{
				MemoryID:   "old_generic",
				Text:       "General workspace context for prior implementation work.",
				Confidence: 0.95,
				ValidFrom:  now.Add(-72 * time.Hour),
				LatestFlag: true,
			},
			{
				MemoryID:   "recent_recall",
				Text:       "Recall token budgeting must preserve active plans and suppress noisy context.",
				Confidence: 0.80,
				ValidFrom:  now.Add(-1 * time.Hour),
				LatestFlag: true,
			},
		},
	}}
	assembler := NewAssembler(Dependencies{
		Memories: memories,
		Clock:    func() time.Time { return now },
	})
	req := testPrefetchRequest()
	req.Query = "recall token budgeting quality"

	resp, err := assembler.Prefetch(context.Background(), req)
	if err != nil {
		t.Fatalf("Prefetch returned error: %v", err)
	}
	if len(resp.Blocks) < 2 {
		t.Fatalf("expected both memory candidates, got %#v", resp.Blocks)
	}
	if resp.Blocks[0].Text != "Recall token budgeting must preserve active plans and suppress noisy context." {
		t.Fatalf("expected relevant recent memory first, got %#v", resp.Blocks)
	}
}

func testPrefetchRequest() *core.PrefetchRequest {
	return &core.PrefetchRequest{
		TenantID:     "tenant_1",
		WorkspaceID:  "workspace_1",
		SessionID:    "session_1",
		ActorID:      "agent:hermes-main",
		Query:        "What should Hermes remember next?",
		BudgetTokens: 2200,
		Mode:         "default",
	}
}

func recallKinds(blocks []core.RecallBlock) []string {
	kinds := make([]string, 0, len(blocks))
	for _, block := range blocks {
		kinds = append(kinds, block.Kind)
	}
	return kinds
}

func containsString(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

type fakeNoteStore struct {
	lastReq *core.ListPinnedNotesRequest
	notes   []*core.Note
}

func (s *fakeNoteStore) AddNote(context.Context, *core.Note) error {
	return core.ErrNotImplemented
}

func (s *fakeNoteStore) ListPinnedNotes(_ context.Context, req *core.ListPinnedNotesRequest) ([]*core.Note, error) {
	reqCopy := *req
	reqCopy.Scopes = append([]core.MemoryScope(nil), req.Scopes...)
	s.lastReq = &reqCopy
	return s.notes, nil
}

type fakePlanStore struct {
	lastReq *core.GetActivePlansRequest
	plans   []*core.Plan
}

func (s *fakePlanStore) CreatePlan(context.Context, *core.Plan, []*core.PlanItem) error {
	return core.ErrNotImplemented
}

func (s *fakePlanStore) UpdatePlan(context.Context, *core.Plan, []*core.PlanItem) error {
	return core.ErrNotImplemented
}

func (s *fakePlanStore) GetActivePlans(_ context.Context, req *core.GetActivePlansRequest) ([]*core.Plan, error) {
	reqCopy := *req
	reqCopy.Scopes = append([]core.MemoryScope(nil), req.Scopes...)
	s.lastReq = &reqCopy
	return s.plans, nil
}

type fakeProfileStore struct {
	profiles map[string]*core.Profile
}

func (s *fakeProfileStore) GetProfile(_ context.Context, entityID string, scope core.MemoryScope) (*core.Profile, error) {
	profile, ok := s.profiles[entityID+"|"+string(scope)]
	if !ok {
		return nil, core.ErrNotFound
	}
	return profile, nil
}

func (s *fakeProfileStore) UpsertProfile(context.Context, *core.Profile) error {
	return core.ErrNotImplemented
}

type fakeMemoryStore struct {
	lastReq *core.SearchMemoriesRequest
	resp    *core.SearchMemoriesResponse
}

func (s *fakeMemoryStore) UpsertMemory(context.Context, *core.Memory) error {
	return core.ErrNotImplemented
}

func (s *fakeMemoryStore) GetMemory(context.Context, string) (*core.Memory, error) {
	return nil, core.ErrNotImplemented
}

func (s *fakeMemoryStore) SearchMemories(_ context.Context, req *core.SearchMemoriesRequest) (*core.SearchMemoriesResponse, error) {
	s.lastReq = req
	return s.resp, nil
}

func (s *fakeMemoryStore) UpsertMemoryEdge(context.Context, *core.MemoryEdge) error {
	return core.ErrNotImplemented
}

func (s *fakeMemoryStore) WriteMemoryTrace(context.Context, *core.MemoryTrace) error {
	return core.ErrNotImplemented
}

func (s *fakeMemoryStore) ExplainMemory(context.Context, *core.ExplainMemoryRequest) (*core.ExplainMemoryResponse, error) {
	return nil, core.ErrNotImplemented
}

type fakeGroupStore struct {
	memberships []*core.MemoryGroupMembership
}

func (s *fakeGroupStore) CreateMemoryGroup(context.Context, *core.MemoryGroup) error {
	return core.ErrNotImplemented
}

func (s *fakeGroupStore) AddMembership(context.Context, *core.MemoryGroupMembership) error {
	return core.ErrNotImplemented
}

func (s *fakeGroupStore) ListMemberships(context.Context, string) ([]*core.MemoryGroupMembership, error) {
	return nil, core.ErrNotImplemented
}

func (s *fakeGroupStore) ListMembershipsForEntity(context.Context, string, string, string) ([]*core.MemoryGroupMembership, error) {
	return s.memberships, nil
}

type fakeFreshnessProvider struct {
	state Freshness
}

func (p fakeFreshnessProvider) RecallFreshness(context.Context, *core.PrefetchRequest) (Freshness, error) {
	return p.state, nil
}

type fakeJobMetricsStore struct {
	metrics *core.JobBacklogMetrics
}

func (s fakeJobMetricsStore) GetJobBacklogMetrics(context.Context, *core.JobBacklogMetricsRequest) (*core.JobBacklogMetrics, error) {
	return s.metrics, nil
}

```



<!-- Source: internal/recall/doc.go | bytes=802 | lines=16 | sha16=93e9a6d62beac4d7 -->

```go
// ============================================================
// FILE     : internal/recall/doc.go
// PURPOSE  : Provides package documentation for budget-aware recall pack assembly.
// LAYER    : application
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : package recall
// DEPENDS  : plans/05_runtime-contracts_ingest-recall-apply.md
// USED_BY  : core service implementations, HTTP API, Hermes adapter, MCP tools
// ------------------------------------------------------------
// AGENT_NOTE: Build typed blocks before rendering and keep recall scope-aware and budget-aware.
// ============================================================

// Package recall owns prefetch candidate assembly, ranking, suppression, and token packing.
package recall

```



<!-- Source: internal/recall/freshness.go | bytes=4483 | lines=146 | sha16=80255c896c486b53 -->

```go
// ============================================================
// FILE     : internal/recall/freshness.go
// PURPOSE  : Converts worker/Codex freshness signals into recall-visible metadata.
// LAYER    : application
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : FreshnessProvider, Freshness, BacklogFreshnessProvider
// DEPENDS  : context, time, internal/core, internal/store
// USED_BY  : internal/recall, cmd/server, cmd/cli, tests
// ------------------------------------------------------------
// AGENT_NOTE: Freshness is operator visibility only; it must not mutate graph state.
// ============================================================

package recall

import (
	"context"
	"time"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/store"
)

const defaultBacklogStaleAfter = 30 * time.Second

// FreshnessProvider reports whether stored recall may lag behind raw turn input.
type FreshnessProvider interface {
	RecallFreshness(ctx context.Context, req *core.PrefetchRequest) (Freshness, error)
}

// Freshness is a narrow recall-only view of worker/Codex freshness.
type Freshness struct {
	Freshness       string
	LagSeconds      *int64
	Reasons         []string
	AffectedSources []string
}

// BacklogFreshnessProvider maps read-only worker backlog metrics to recall freshness.
type BacklogFreshnessProvider struct {
	Jobs       store.JobMetricsStore
	StaleAfter time.Duration
	Clock      func() time.Time
}

// RecallFreshness returns stale metadata when ready or retrying worker jobs imply
// derived memories may not include the latest raw events.
func (p BacklogFreshnessProvider) RecallFreshness(ctx context.Context, req *core.PrefetchRequest) (Freshness, error) {
	if p.Jobs == nil {
		return Freshness{Freshness: "stored"}, nil
	}
	now := time.Now().UTC()
	if p.Clock != nil {
		now = p.Clock().UTC()
	}
	metrics, err := p.Jobs.GetJobBacklogMetrics(ctx, &core.JobBacklogMetricsRequest{
		TenantID:     req.TenantID,
		WorkspaceID:  req.WorkspaceID,
		GeneratedNow: now,
	})
	if err != nil {
		return Freshness{}, err
	}
	staleAfter := p.StaleAfter
	if staleAfter == 0 {
		staleAfter = defaultBacklogStaleAfter
	}
	state := Freshness{Freshness: "stored"}
	if metrics == nil {
		return state, nil
	}
	if metrics.OldestQueuedAgeSeconds != nil {
		age := *metrics.OldestQueuedAgeSeconds
		if metrics.Counts.ReadyQueued > 0 && age >= int64(staleAfter.Seconds()) {
			state.Freshness = "stale"
			state.LagSeconds = maxLagSeconds(state.LagSeconds, age)
			state.Reasons = append(state.Reasons, "worker_backlog_stale")
			state.AffectedSources = derivedRecallSources()
		}
	}
	if metrics.OldestRunningAgeSeconds != nil {
		age := *metrics.OldestRunningAgeSeconds
		if metrics.Counts.Running > 0 && age >= int64(staleAfter.Seconds()) {
			state.Freshness = "stale"
			state.LagSeconds = maxLagSeconds(state.LagSeconds, age)
			state.Reasons = append(state.Reasons, "worker_running_stale")
			state.AffectedSources = derivedRecallSources()
		}
	}
	if metrics.RetryableQueuedAttempts > 0 {
		if state.Freshness == "stored" {
			state.Freshness = "stale"
			state.AffectedSources = derivedRecallSources()
		}
		state.Reasons = append(state.Reasons, "codex_or_worker_retry_backlog")
	}
	return state, nil
}

func maxLagSeconds(current *int64, candidate int64) *int64 {
	if current != nil && *current >= candidate {
		return current
	}
	value := candidate
	return &value
}

func applyRecallFreshness(blocks []core.RecallBlock, freshness Freshness) []core.RecallBlock {
	if freshness.Freshness == "" || freshness.Freshness == "stored" || len(freshness.AffectedSources) == 0 {
		return blocks
	}
	affected := make(map[string]struct{}, len(freshness.AffectedSources))
	for _, source := range freshness.AffectedSources {
		affected[source] = struct{}{}
	}
	for i := range blocks {
		if _, ok := affected[blocks[i].Source]; ok {
			blocks[i].Freshness = freshness.Freshness
		}
	}
	return blocks
}

func derivedRecallSources() []string {
	return []string{"memories", "profile", "session_summaries"}
}

func uniqueStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

```
