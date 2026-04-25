// ============================================================
// FILE     : internal/store/postgres/store.go
// PURPOSE  : Provides the PostgreSQL store root backed by pgxpool.
// LAYER    : infra
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : Store, NewStore
// DEPENDS  : github.com/jackc/pgx/v5/pgxpool, internal/store
// USED_BY  : cmd/server, cmd/worker
// ------------------------------------------------------------
// AGENT_NOTE: Implement storage interfaces without changing core domain semantics.
// ============================================================

// Package postgres implements the VibeGravity store interfaces using PostgreSQL and pgxpool.
package postgres

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/store"
)

var _ store.RawEventStore = (*Store)(nil)
var _ store.JobStore = (*Store)(nil)
var _ store.JobMetricsStore = (*Store)(nil)
var _ store.MemoryStore = (*Store)(nil)
var _ store.CorrectionStore = (*Store)(nil)
var _ store.TimelineStore = (*Store)(nil)
var _ store.ProfileStore = (*Store)(nil)
var _ store.NoteStore = (*Store)(nil)
var _ store.PlanStore = (*Store)(nil)
var _ store.DocumentStore = (*Store)(nil)
var _ store.SessionSummaryStore = (*Store)(nil)
var _ store.DreamingStore = (*Store)(nil)
var _ store.GroupStore = (*Store)(nil)

// Store implements the core storage interfaces for VibeGravity.
type Store struct {
	pool *pgxpool.Pool
}

// NewStore creates a new PostgreSQL store instance.
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{
		pool: pool,
	}
}
