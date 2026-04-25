// ============================================================
// FILE     : internal/store/postgres/concurrency_integration_test.go
// PURPOSE  : Stress-tests PostgreSQL graph update concurrency and provenance integrity.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : TestPostgresConcurrentUpdateMemoryAllowsOneWinnerNoDanglingWrites
// DEPENDS  : context, os, sync, testing, time, internal/core, github.com/jackc/pgx/v5/pgxpool
// USED_BY  : go test ./internal/store/postgres
// ------------------------------------------------------------
// AGENT_NOTE: Keep this test skippable when VIBEGRAVITY_DB_URL is unset; it verifies real row-lock behavior.
// ============================================================

package postgres

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

func TestPostgresConcurrentUpdateMemoryAllowsOneWinnerNoDanglingWrites(t *testing.T) {
	dbURL := os.Getenv("VIBEGRAVITY_DB_URL")
	if dbURL == "" {
		t.Skip("Skipping Postgres concurrency integration test because VIBEGRAVITY_DB_URL is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("connect test database: %v", err)
	}
	defer pool.Close()

	store := NewStore(pool)
	tenantID := fmt.Sprintf("tenant_concurrency_%d", time.Now().UnixNano())
	workspaceID := "workspace_concurrency"
	ownerID := "agent:hermes-main"
	targetID := "mem_concurrency_target"
	startedAt := time.Now().UTC()

	cleanupPostgresConcurrencyRows(ctx, t, pool, tenantID, workspaceID)
	defer cleanupPostgresConcurrencyRows(context.Background(), t, pool, tenantID, workspaceID)

	mustSeedJob(ctx, t, pool, tenantID, workspaceID, "job_seed")
	if err := store.CreateMemoryWithTrace(ctx, &core.Memory{
		ID:            targetID,
		TenantID:      tenantID,
		WorkspaceID:   workspaceID,
		Scope:         core.MemoryScopeWorkspaceShared,
		OwnerEntityID: ownerID,
		Kind:          core.MemoryKindFact,
		ArtifactClass: core.ArtifactClassKnowledge,
		Text:          "Original memory before concurrent update.",
		Fingerprint:   "fp_concurrency_target",
		Confidence:    0.7,
		Status:        core.MemoryStatusActive,
		ValidFrom:     startedAt,
		LatestFlag:    true,
		MetadataJSON:  []byte(`{}`),
		CreatedAt:     startedAt,
		UpdatedAt:     startedAt,
	}, &core.MemoryTrace{
		MemoryID:              targetID,
		RawEventIDs:           []string{"evt_seed"},
		ReasoningJobID:        "job_seed",
		ReasoningStage:        "resolve",
		CandidateSnapshotJSON: []byte(`{"seed":true}`),
		AppliedOperationsJSON: []byte(`[{"operation_id":"seed"}]`),
		RelatedDocumentIDs:    []string{},
		CreatedAt:             startedAt,
	}); err != nil {
		t.Fatalf("seed target memory with trace: %v", err)
	}

	const workers = 16
	var ready sync.WaitGroup
	ready.Add(workers)
	start := make(chan struct{})
	errs := make(chan error, workers)
	var successes atomic.Int32

	for i := 0; i < workers; i++ {
		i := i
		go func() {
			ready.Done()
			<-start

			jobID := fmt.Sprintf("job_concurrency_%02d", i)
			if err := insertSeedJob(context.Background(), pool, tenantID, workspaceID, jobID); err != nil {
				errs <- err
				return
			}
			workerCtx, workerCancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer workerCancel()

			memoryID := fmt.Sprintf("mem_concurrency_update_%02d", i)
			err := store.CreateMemoryWithTraceAndUpdateEdge(workerCtx,
				&core.Memory{
					ID:            memoryID,
					TenantID:      tenantID,
					WorkspaceID:   workspaceID,
					Scope:         core.MemoryScopeWorkspaceShared,
					OwnerEntityID: ownerID,
					Kind:          core.MemoryKindFact,
					ArtifactClass: core.ArtifactClassKnowledge,
					Text:          fmt.Sprintf("Concurrent update winner candidate %02d.", i),
					Fingerprint:   fmt.Sprintf("fp_concurrency_update_%02d", i),
					Confidence:    0.8,
					Status:        core.MemoryStatusActive,
					ValidFrom:     startedAt.Add(time.Duration(i+1) * time.Millisecond),
					LatestFlag:    true,
					MetadataJSON:  []byte(`{}`),
					CreatedAt:     startedAt.Add(time.Duration(i+1) * time.Millisecond),
					UpdatedAt:     startedAt.Add(time.Duration(i+1) * time.Millisecond),
				},
				&core.MemoryTrace{
					MemoryID:              memoryID,
					RawEventIDs:           []string{fmt.Sprintf("evt_concurrency_%02d", i)},
					ReasoningJobID:        jobID,
					ReasoningStage:        "resolve",
					CandidateSnapshotJSON: []byte(`{"candidate_memories":[]}`),
					AppliedOperationsJSON: []byte(fmt.Sprintf(`[{"operation_id":"op_update_%02d"}]`, i)),
					RelatedDocumentIDs:    []string{},
					CreatedAt:             time.Now().UTC(),
				},
				&core.MemoryEdge{
					FromMemoryID:   memoryID,
					ToMemoryID:     targetID,
					EdgeKind:       core.EdgeKindUpdates,
					Confidence:     0.8,
					CreatedByJobID: jobID,
					CreatedAt:      time.Now().UTC(),
				})
			if err == nil {
				successes.Add(1)
				errs <- nil
				return
			}
			if errors.Is(err, core.ErrConflict) {
				errs <- nil
				return
			}
			errs <- err
		}()
	}

	ready.Wait()
	close(start)
	for i := 0; i < workers; i++ {
		if err := <-errs; err != nil {
			t.Fatalf("concurrent update returned unexpected error: %v", err)
		}
	}

	if got := successes.Load(); got != 1 {
		t.Fatalf("expected exactly one concurrent update winner, got %d", got)
	}
	assertPostgresConcurrencyGraphIntegrity(ctx, t, pool, tenantID, workspaceID, targetID)
}

func mustSeedJob(ctx context.Context, t testing.TB, pool *pgxpool.Pool, tenantID string, workspaceID string, jobID string) {
	t.Helper()
	if err := insertSeedJob(ctx, pool, tenantID, workspaceID, jobID); err != nil {
		t.Fatalf("seed job %q: %v", jobID, err)
	}
}

func insertSeedJob(ctx context.Context, pool *pgxpool.Pool, tenantID string, workspaceID string, jobID string) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO ingest_jobs (id, tenant_id, workspace_id, job_kind, status, raw_event_ids, payload_json)
		VALUES ($1, $2, $3, 'process_turn_event', 'completed', '{}', '{}'::jsonb)
		ON CONFLICT (id) DO NOTHING
	`, jobID, tenantID, workspaceID)
	if err != nil {
		return fmt.Errorf("insert ingest job %q: %w", jobID, err)
	}
	return nil
}

func assertPostgresConcurrencyGraphIntegrity(ctx context.Context, t *testing.T, pool *pgxpool.Pool, tenantID string, workspaceID string, targetID string) {
	t.Helper()

	var winnerCount int
	if err := pool.QueryRow(ctx, `
		SELECT count(*)
		FROM memory_edges me
		JOIN memories m ON m.id = me.from_memory_id
		JOIN memory_trace mt ON mt.memory_id = m.id
		WHERE m.tenant_id = $1
		  AND m.workspace_id = $2
		  AND me.to_memory_id = $3
		  AND me.edge_kind = 'updates'
		  AND m.status = 'active'
		  AND m.latest_flag = true
	`, tenantID, workspaceID, targetID).Scan(&winnerCount); err != nil {
		t.Fatalf("count committed update winner: %v", err)
	}
	if winnerCount != 1 {
		t.Fatalf("expected one active latest update winner with trace and edge, got %d", winnerCount)
	}

	var targetStatus core.MemoryStatus
	var targetLatest bool
	if err := pool.QueryRow(ctx, `
		SELECT status, latest_flag
		FROM memories
		WHERE id = $1 AND tenant_id = $2 AND workspace_id = $3
	`, targetID, tenantID, workspaceID).Scan(&targetStatus, &targetLatest); err != nil {
		t.Fatalf("load update target status: %v", err)
	}
	if targetStatus != core.MemoryStatusSuperseded || targetLatest {
		t.Fatalf("target should be superseded and non-latest, got status=%q latest=%v", targetStatus, targetLatest)
	}

	var danglingCount int
	if err := pool.QueryRow(ctx, `
		SELECT count(*)
		FROM memories m
		LEFT JOIN memory_trace mt ON mt.memory_id = m.id
		LEFT JOIN memory_edges me
		  ON me.from_memory_id = m.id
		 AND me.to_memory_id = $3
		 AND me.edge_kind = 'updates'
		WHERE m.tenant_id = $1
		  AND m.workspace_id = $2
		  AND m.id LIKE 'mem_concurrency_update_%'
		  AND (mt.memory_id IS NULL OR me.from_memory_id IS NULL)
	`, tenantID, workspaceID, targetID).Scan(&danglingCount); err != nil {
		t.Fatalf("count dangling update writes: %v", err)
	}
	if danglingCount != 0 {
		t.Fatalf("expected no dangling memory/trace rows from losing workers, got %d", danglingCount)
	}
}

func cleanupPostgresConcurrencyRows(ctx context.Context, t testing.TB, pool *pgxpool.Pool, tenantID string, workspaceID string) {
	t.Helper()
	if _, err := pool.Exec(ctx, `
		DELETE FROM memories
		WHERE tenant_id = $1 AND workspace_id = $2
	`, tenantID, workspaceID); err != nil {
		t.Fatalf("cleanup memories: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		DELETE FROM ingest_jobs
		WHERE tenant_id = $1 AND workspace_id = $2
	`, tenantID, workspaceID); err != nil {
		t.Fatalf("cleanup ingest jobs: %v", err)
	}
}
