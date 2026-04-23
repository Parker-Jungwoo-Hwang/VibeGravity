// ============================================================
// FILE     : internal/httpapi/router.go
// PURPOSE  : Defines the HTTP router and initial transport handlers.
// LAYER    : interface
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : App, NewRouter
// DEPENDS  : internal/core, github.com/go-chi/chi/v5, pgxpool
// USED_BY  : cmd/server, tests
// ------------------------------------------------------------
// AGENT_NOTE: HTTP handlers must preserve the core service semantics.
// ============================================================

// Package httpapi provides the HTTP transport layer for VibeGravity.
package httpapi

import (
	"encoding/json"
	"net/http"

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
		// Skeleton routes for core service
		r.Post("/prefetch", app.Prefetch)
		r.Post("/sync-turn", app.SyncTurn)
		// Other endpoints...
	})

	return r
}

// Healthz returns 200 OK if the application and database are healthy.
func (a *App) Healthz(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Ping the database
	if err := a.DBPool.Ping(ctx); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{
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

// Prefetch handler placeholder
func (a *App) Prefetch(w http.ResponseWriter, _ *http.Request) {
	// Not implemented
	w.WriteHeader(http.StatusNotImplemented)
}

// SyncTurn handler placeholder
func (a *App) SyncTurn(w http.ResponseWriter, _ *http.Request) {
	// Not implemented
	w.WriteHeader(http.StatusNotImplemented)
}
