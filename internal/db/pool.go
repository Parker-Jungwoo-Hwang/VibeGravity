// ============================================================
// FILE     : internal/db/pool.go
// PURPOSE  : Builds PostgreSQL connection pools with pgvector registration.
// LAYER    : infra
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : NewPool
// DEPENDS  : internal/config, github.com/jackc/pgx/v5, github.com/pgvector/pgvector-go
// USED_BY  : cmd/server, cmd/worker, cmd/cli, tests
// ------------------------------------------------------------
// AGENT_NOTE: PostgreSQL is canonical; keep pgvector setup explicit here.
// ============================================================

// Package db manages the database connection pool using pgxpool.
package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pgxvector "github.com/pgvector/pgvector-go/pgx"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/config"
)

// NewPool initializes a new PostgreSQL connection pool.
// It configures the pgxpool to register pgvector types automatically on new connections.
func NewPool(ctx context.Context, cfg config.Config) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Register pgvector types on each connection
	poolConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		return pgxvector.RegisterTypes(ctx, conn)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}
