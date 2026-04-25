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
