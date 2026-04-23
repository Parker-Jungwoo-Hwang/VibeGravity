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
)

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
