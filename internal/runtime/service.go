// ============================================================
// FILE     : internal/runtime/service.go
// PURPOSE  : Composes the shared VibeGravity service and job stores for process entrypoints.
// LAYER    : infra
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : OpenService, OpenHTTPApp, OpenJobOperatorStore
// DEPENDS  : internal/config, internal/db, internal/httpapi, internal/ingest, internal/kernel, internal/recall, internal/store/postgres
// USED_BY  : cmd/server, cmd/cli, cmd/vibegravity
// ------------------------------------------------------------
// AGENT_NOTE: Keep HTTP, MCP, and CLI service wiring identical through this package.
// ============================================================

package runtime

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/config"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/db"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/httpapi"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/ingest"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/kernel"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/recall"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/store/postgres"
)

// JobOperatorStore is the runtime store surface used by operator CLI commands.
type JobOperatorStore interface {
	ListBlockedJobs(ctx context.Context, limit int) ([]*core.IngestJob, error)
	RequeueBlockedJob(ctx context.Context, jobID string, reason string) error
	GetJobBacklogMetrics(ctx context.Context, req *core.JobBacklogMetricsRequest) (*core.JobBacklogMetrics, error)
}

// OpenService opens PostgreSQL and composes the shared core service.
func OpenService(ctx context.Context, cfg config.Config) (core.VibeGravityService, func(), error) {
	pool, err := db.NewPool(ctx, cfg)
	if err != nil {
		return nil, func() {}, err
	}
	service, err := NewServiceFromPool(pool)
	if err != nil {
		pool.Close()
		return nil, func() {}, err
	}
	return service, pool.Close, nil
}

// OpenHTTPApp opens PostgreSQL and returns an HTTP app that shares runtime service wiring.
func OpenHTTPApp(ctx context.Context, cfg config.Config) (*httpapi.App, func(), error) {
	pool, err := db.NewPool(ctx, cfg)
	if err != nil {
		return nil, func() {}, err
	}
	service, err := NewServiceFromPool(pool)
	if err != nil {
		pool.Close()
		return nil, func() {}, err
	}
	return &httpapi.App{
		Service: service,
		DBPool:  pool,
	}, pool.Close, nil
}

// OpenJobOperatorStore opens PostgreSQL for read-mostly operator job commands.
func OpenJobOperatorStore(ctx context.Context, cfg config.Config) (JobOperatorStore, func(), error) {
	pool, err := db.NewPool(ctx, cfg)
	if err != nil {
		return nil, func() {}, err
	}
	return postgres.NewStore(pool), pool.Close, nil
}

// NewServiceFromPool composes the shared service over a PostgreSQL pool.
func NewServiceFromPool(pool *pgxpool.Pool) (core.VibeGravityService, error) {
	pgStore := postgres.NewStore(pool)
	ingestService, err := ingest.NewService(ingest.Dependencies{
		RawEvents: pgStore,
		Jobs:      pgStore,
	})
	if err != nil {
		return nil, err
	}
	recallAssembler := recall.NewAssembler(recall.Dependencies{
		Notes:     pgStore,
		Plans:     pgStore,
		Memories:  pgStore,
		Documents: pgStore,
		Profiles:  pgStore,
		Summaries: pgStore,
		Groups:    pgStore,
		Freshness: recall.BacklogFreshnessProvider{Jobs: pgStore},
	})
	return kernel.NewService(kernel.Dependencies{
		Ingest:      ingestService,
		Recall:      recallAssembler,
		Notes:       pgStore,
		Plans:       pgStore,
		Memories:    pgStore,
		Corrections: pgStore,
		Jobs:        pgStore,
		Timeline:    pgStore,
		Documents:   pgStore,
	})
}
