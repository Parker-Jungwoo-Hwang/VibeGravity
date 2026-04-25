# 05 Gpt Pro Postgres Store And Migrations

Generated: 2026-04-25

This file is part of the GPT-Pro review material bundle for VibeGravity.

## Included Sources

- `internal/store/postgres/concurrency_integration_test.go`
- `internal/store/postgres/corrections.go`
- `internal/store/postgres/corrections_test.go`
- `internal/store/postgres/documents.go`
- `internal/store/postgres/documents_test.go`
- `internal/store/postgres/dreaming.go`
- `internal/store/postgres/dreaming_test.go`
- `internal/store/postgres/groups.go`
- `internal/store/postgres/helpers.go`
- `internal/store/postgres/jobs.go`
- `internal/store/postgres/jobs_test.go`
- `internal/store/postgres/memories.go`
- `internal/store/postgres/memories_test.go`
- `internal/store/postgres/notes_plans.go`
- `internal/store/postgres/notes_plans_test.go`
- `internal/store/postgres/profiles_summaries.go`
- `internal/store/postgres/raw_events.go`
- `internal/store/postgres/search.go`
- `internal/store/postgres/search_test.go`
- `internal/store/postgres/store.go`
- `internal/store/postgres/timeline.go`
- `internal/store/postgres/timeline_test.go`
- `internal/store/store.go`
- `migrations/000001_create_pgvector_extension.down.sql`
- `migrations/000001_create_pgvector_extension.up.sql`
- `migrations/000002_create_core_tables.down.sql`
- `migrations/000002_create_core_tables.up.sql`
- `migrations/000003_add_vector_columns.down.sql`
- `migrations/000003_add_vector_columns.up.sql`

## Source Contents


<!-- Source: internal/store/postgres/concurrency_integration_test.go | bytes=8877 | lines=263 | sha16=91a375d4ebc66513 -->

```go
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

```



<!-- Source: internal/store/postgres/corrections.go | bytes=4579 | lines=122 | sha16=00e0504c3c1904b2 -->

```go
// ============================================================
// FILE     : internal/store/postgres/corrections.go
// PURPOSE  : Implements PostgreSQL persistence for human memory corrections.
// LAYER    : infra
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : RecordMemoryCorrection
// DEPENDS  : internal/core, github.com/jackc/pgx/v5
// USED_BY  : kernel CorrectMemory path
// ------------------------------------------------------------
// AGENT_NOTE: Corrections are append-safe operator artifacts; do not mutate memory_trace or latest_flag here.
// ============================================================

package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

// RecordMemoryCorrection writes a raw correction event and operator-visible artifact idempotently.
func (s *Store) RecordMemoryCorrection(ctx context.Context, event *core.RawEvent, correction *core.MemoryCorrection) (*core.MemoryCorrection, error) {
	if event == nil {
		return nil, fmt.Errorf("%w: raw correction event is required", core.ErrInvalidArgument)
	}
	if correction == nil {
		return nil, fmt.Errorf("%w: memory correction is required", core.ErrInvalidArgument)
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin memory correction transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	rawEventID, err := insertCorrectionRawEvent(ctx, tx, event)
	if err != nil {
		return nil, err
	}
	correction.RawEventID = rawEventID
	recorded, err := insertMemoryCorrection(ctx, tx, correction)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit memory correction transaction: %w", err)
	}
	return recorded, nil
}

func insertCorrectionRawEvent(ctx context.Context, tx pgx.Tx, event *core.RawEvent) (string, error) {
	id := event.ID
	var err error
	if id == "" {
		id, err = newID("evt")
		if err != nil {
			return "", err
		}
	}
	var rawEventID string
	err = tx.QueryRow(ctx, `
		INSERT INTO raw_events (
			id, tenant_id, workspace_id, session_id, actor_id, event_kind,
			source, idempotency_key, fingerprint, occurred_at, payload_json, created_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		ON CONFLICT (tenant_id, source, idempotency_key) DO UPDATE
		SET idempotency_key = EXCLUDED.idempotency_key
		WHERE raw_events.workspace_id = EXCLUDED.workspace_id
		RETURNING id
	`, id, event.TenantID, event.WorkspaceID, event.SessionID, event.ActorID,
		event.EventKind, event.Source, event.IdempotencyKey, event.Fingerprint,
		timeOrNow(event.OccurredAt), rawJSONOrEmpty(event.PayloadJSON), timeOrNow(event.CreatedAt)).Scan(&rawEventID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", fmt.Errorf("%w: correction idempotency key belongs to a different workspace", core.ErrConflict)
	}
	if err != nil {
		return "", fmt.Errorf("insert correction raw event: %w", err)
	}
	return rawEventID, nil
}

func insertMemoryCorrection(ctx context.Context, tx pgx.Tx, correction *core.MemoryCorrection) (*core.MemoryCorrection, error) {
	id := correction.ID
	var err error
	if id == "" {
		id, err = newID("corr")
		if err != nil {
			return nil, err
		}
	}
	recorded := &core.MemoryCorrection{}
	err = tx.QueryRow(ctx, `
		INSERT INTO memory_corrections (
			id, tenant_id, workspace_id, memory_id, operator_id, raw_event_id,
			idempotency_key, correction_text, evidence_json, status, created_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (tenant_id, workspace_id, idempotency_key) DO UPDATE
		SET idempotency_key = EXCLUDED.idempotency_key
		RETURNING id, tenant_id, workspace_id, memory_id, operator_id, raw_event_id,
		          idempotency_key, correction_text, evidence_json, status, created_at
	`, id, correction.TenantID, correction.WorkspaceID, correction.MemoryID, correction.OperatorID,
		correction.RawEventID, correction.IdempotencyKey, correction.CorrectionText,
		rawJSONOrEmpty(correction.EvidenceJSON), valueOr(correction.Status, "recorded"),
		timeOrNow(correction.CreatedAt)).Scan(&recorded.ID, &recorded.TenantID,
		&recorded.WorkspaceID, &recorded.MemoryID, &recorded.OperatorID,
		&recorded.RawEventID, &recorded.IdempotencyKey, &recorded.CorrectionText,
		&recorded.EvidenceJSON, &recorded.Status, &recorded.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, core.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("insert memory correction: %w", err)
	}
	return recorded, nil
}

```



<!-- Source: internal/store/postgres/corrections_test.go | bytes=2869 | lines=79 | sha16=7ca2ef4d479242e1 -->

```go
// ============================================================
// FILE     : internal/store/postgres/corrections_test.go
// PURPOSE  : Verifies correction persistence SQL without requiring a live database.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : postgres correction source tests
// DEPENDS  : strings, testing
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: Lock the narrow correction intake boundary without testing supersession semantics here.
// ============================================================

package postgres

import (
	"strings"
	"testing"
)

func TestRecordMemoryCorrectionUsesSingleTransaction(t *testing.T) {
	t.Parallel()

	source := readPostgresSourceFile(t, "corrections.go")
	recordSource := extractPostgresSourceBetween(t, source, "func (s *Store) RecordMemoryCorrection", "func insertCorrectionRawEvent")

	for _, want := range []string{
		"s.pool.BeginTx(ctx, pgx.TxOptions{})",
		"defer func() { _ = tx.Rollback(ctx) }()",
		"insertCorrectionRawEvent(ctx, tx, event)",
		"insertMemoryCorrection(ctx, tx, correction)",
		"tx.Commit(ctx)",
	} {
		if !strings.Contains(recordSource, want) {
			t.Fatalf("RecordMemoryCorrection must preserve %q, got:\n%s", want, recordSource)
		}
	}
}

func TestCorrectionRawEventReturnsStableIDOnRetry(t *testing.T) {
	t.Parallel()

	source := readPostgresSourceFile(t, "corrections.go")
	insertSource := extractPostgresSourceBetween(t, source, "func insertCorrectionRawEvent", "func insertMemoryCorrection")

	for _, want := range []string{
		"ON CONFLICT (tenant_id, source, idempotency_key) DO UPDATE",
		"WHERE raw_events.workspace_id = EXCLUDED.workspace_id",
		"RETURNING id",
	} {
		if !strings.Contains(insertSource, want) {
			t.Fatalf("correction raw event insert must preserve stable retry IDs; missing %q in:\n%s", want, insertSource)
		}
	}
}

func TestMemoryCorrectionArtifactIsAppendSafe(t *testing.T) {
	t.Parallel()

	source := readPostgresSourceFile(t, "corrections.go")
	start := strings.Index(source, "func insertMemoryCorrection")
	if start < 0 {
		t.Fatal("missing insertMemoryCorrection")
	}
	insertSource := source[start:]
	if strings.Contains(insertSource, "latest_flag") || strings.Contains(insertSource, "memory_trace") || strings.Contains(insertSource, "memory_edges") {
		t.Fatalf("correction intake must not mutate graph apply/provenance tables, got:\n%s", insertSource)
	}
	for _, want := range []string{
		"INSERT INTO memory_corrections",
		"ON CONFLICT (tenant_id, workspace_id, idempotency_key) DO UPDATE",
		"RETURNING id, tenant_id, workspace_id, memory_id, operator_id, raw_event_id",
	} {
		if !strings.Contains(insertSource, want) {
			t.Fatalf("memory correction insert must preserve %q, got:\n%s", want, insertSource)
		}
	}
}

```



<!-- Source: internal/store/postgres/documents.go | bytes=5611 | lines=166 | sha16=54a5d2e39304c3f6 -->

```go
// ============================================================
// FILE     : internal/store/postgres/documents.go
// PURPOSE  : Implements PostgreSQL persistence for documents and document chunks.
// LAYER    : infra
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : AddDocumentWithChunks, AddDocument, AddDocumentChunks
// DEPENDS  : internal/core, github.com/jackc/pgx/v5
// USED_BY  : document API, recall document search
// ------------------------------------------------------------
// AGENT_NOTE: Keep document storage separate from derived memory storage.
// ============================================================

package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

// AddDocumentWithChunks writes a document and replaces its chunks in one transaction.
func (s *Store) AddDocumentWithChunks(ctx context.Context, document *core.Document, chunks []*core.DocumentChunk) error {
	if document == nil {
		return fmt.Errorf("%w: document is required", core.ErrInvalidArgument)
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin add document with chunks: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := addDocumentInTx(ctx, tx, document); err != nil {
		return err
	}
	if err := addDocumentChunksInTx(ctx, tx, document.ID, chunks); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit add document with chunks: %w", err)
	}
	return nil
}

// AddDocument writes a document.
func (s *Store) AddDocument(ctx context.Context, document *core.Document) error {
	if document == nil {
		return fmt.Errorf("%w: document is required", core.ErrInvalidArgument)
	}
	return addDocumentInTx(ctx, s.pool, document)
}

// AddDocumentChunks writes retrieval chunks for a document.
func (s *Store) AddDocumentChunks(ctx context.Context, chunks []*core.DocumentChunk) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin add document chunks: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	replacedDocuments := make(map[string]struct{})
	for _, chunk := range chunks {
		if chunk == nil {
			continue
		}
		if _, ok := replacedDocuments[chunk.DocumentID]; !ok {
			if err := replaceDocumentChunksInTx(ctx, tx, chunk.DocumentID, []*core.DocumentChunk{chunk}); err != nil {
				return err
			}
			replacedDocuments[chunk.DocumentID] = struct{}{}
			continue
		}
		if err := insertDocumentChunkInTx(ctx, tx, chunk.DocumentID, chunk); err != nil {
			return err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit add document chunks: %w", err)
	}
	return nil
}

type documentExecutor interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func addDocumentInTx(ctx context.Context, exec documentExecutor, document *core.Document) error {
	var err error
	if document.ID == "" {
		document.ID, err = newID("doc")
		if err != nil {
			return err
		}
	}
	err = exec.QueryRow(ctx, `
		INSERT INTO documents (
			id, tenant_id, workspace_id, source, title, fingerprint,
			metadata_json, version_hint, created_at, updated_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		ON CONFLICT (tenant_id, workspace_id, fingerprint) DO UPDATE
		SET source = EXCLUDED.source,
		    title = EXCLUDED.title,
		    metadata_json = EXCLUDED.metadata_json,
		    version_hint = EXCLUDED.version_hint,
		    updated_at = EXCLUDED.updated_at
		RETURNING id
	`, document.ID, document.TenantID, document.WorkspaceID, document.Source,
		document.Title, document.Fingerprint, rawJSONOrEmpty(document.MetadataJSON),
		document.VersionHint, timeOrNow(document.CreatedAt), timeOrNow(document.UpdatedAt)).Scan(&document.ID)
	if err != nil {
		return fmt.Errorf("upsert document: %w", err)
	}
	return nil
}

func addDocumentChunksInTx(ctx context.Context, exec documentExecutor, documentID string, chunks []*core.DocumentChunk) error {
	return replaceDocumentChunksInTx(ctx, exec, documentID, chunks)
}

func replaceDocumentChunksInTx(ctx context.Context, exec documentExecutor, documentID string, chunks []*core.DocumentChunk) error {
	if _, err := exec.Exec(ctx, `DELETE FROM document_chunks WHERE document_id = $1`, documentID); err != nil {
		return fmt.Errorf("delete existing document chunks: %w", err)
	}
	for _, chunk := range chunks {
		if chunk == nil {
			continue
		}
		if err := insertDocumentChunkInTx(ctx, exec, documentID, chunk); err != nil {
			return err
		}
	}
	return nil
}

func insertDocumentChunkInTx(ctx context.Context, exec documentExecutor, documentID string, chunk *core.DocumentChunk) error {
	var err error
	if chunk.ID == "" {
		chunk.ID, err = newID("chunk")
		if err != nil {
			return err
		}
	}
	chunk.DocumentID = documentID
	_, err = exec.Exec(ctx, `
			INSERT INTO document_chunks (
				id, document_id, chunk_index, text, heading_path, metadata_json,
				neighbor_chunk_ids, embedding_model, embedding_dims, embedding_updated_at,
				created_at, updated_at
			)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		`, chunk.ID, chunk.DocumentID, chunk.ChunkIndex, chunk.Text, chunk.HeadingPath,
		rawJSONOrEmpty(chunk.MetadataJSON), chunk.NeighborChunkIDs,
		valueOr(chunk.EmbeddingModel, "pending"), chunk.EmbeddingDims, chunk.EmbeddingUpdatedAt,
		timeOrNow(chunk.CreatedAt), timeOrNow(chunk.UpdatedAt))
	if err != nil {
		return fmt.Errorf("insert document chunk: %w", err)
	}
	return nil
}

```



<!-- Source: internal/store/postgres/documents_test.go | bytes=3155 | lines=89 | sha16=a009591bdff366eb -->

```go
// ============================================================
// FILE     : internal/store/postgres/documents_test.go
// PURPOSE  : Verifies document storage source preserves idempotent chunk replacement.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : postgres document source tests
// DEPENDS  : os, path/filepath, runtime, strings, testing
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: These tests lock storage contracts without requiring a live database.
// ============================================================

package postgres

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestAddDocumentChunksReplacesExistingChunksForDocument(t *testing.T) {
	t.Parallel()

	source := readPostgresSourceFile(t, "documents.go")
	addChunksSource := extractPostgresSourceBetween(t, source, "func (s *Store) AddDocumentChunks", "if err := tx.Commit")
	replaceChunksSource := extractPostgresSourceBetween(t, source, "func replaceDocumentChunksInTx", "func insertDocumentChunkInTx")

	if !strings.Contains(replaceChunksSource, "DELETE FROM document_chunks WHERE document_id = $1") {
		t.Fatalf("document chunk replacement must delete old chunks for idempotent document upserts, got:\n%s", replaceChunksSource)
	}
	if !strings.Contains(addChunksSource, "replacedDocuments") {
		t.Fatalf("AddDocumentChunks should delete once per document, got:\n%s", addChunksSource)
	}
	if !strings.Contains(addChunksSource, "replaceDocumentChunksInTx(ctx, tx, chunk.DocumentID") {
		t.Fatalf("AddDocumentChunks should route first chunk per document through replacement helper, got:\n%s", addChunksSource)
	}
}

func TestAddDocumentWithChunksUsesSingleTransaction(t *testing.T) {
	t.Parallel()

	source := readPostgresSourceFile(t, "documents.go")
	atomicSource := extractPostgresSourceBetween(t, source, "func (s *Store) AddDocumentWithChunks", "func (s *Store) AddDocument(")

	for _, required := range []string{
		"s.pool.Begin(ctx)",
		"defer func() { _ = tx.Rollback(ctx) }()",
		"addDocumentInTx(ctx, tx, document)",
		"addDocumentChunksInTx(ctx, tx, document.ID, chunks)",
		"tx.Commit(ctx)",
	} {
		if !strings.Contains(atomicSource, required) {
			t.Fatalf("AddDocumentWithChunks must keep document and chunks in one transaction; missing %q in:\n%s", required, atomicSource)
		}
	}
}

func readPostgresSourceFile(t *testing.T, name string) string {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not locate current test file")
	}
	data, err := os.ReadFile(filepath.Join(filepath.Dir(currentFile), name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return string(data)
}

func extractPostgresSourceBetween(t *testing.T, source, startMarker, endMarker string) string {
	t.Helper()

	start := strings.Index(source, startMarker)
	if start < 0 {
		t.Fatalf("missing start marker %q", startMarker)
	}
	remainder := source[start:]
	end := strings.Index(remainder, endMarker)
	if end < 0 {
		t.Fatalf("missing end marker %q", endMarker)
	}
	return remainder[:end]
}

```



<!-- Source: internal/store/postgres/dreaming.go | bytes=8159 | lines=241 | sha16=e1e683beb21ad07e -->

```go
// ============================================================
// FILE     : internal/store/postgres/dreaming.go
// PURPOSE  : Implements PostgreSQL queries for background dreaming consolidation and tier promotion.
// LAYER    : infra
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : LoadDreamingSessionInput, PromoteMemories
// DEPENDS  : context, fmt, strings, time, internal/core, github.com/jackc/pgx/v5
// USED_BY  : graph dreaming service, worker
// ------------------------------------------------------------
// AGENT_NOTE: Dreaming metadata updates must never change memory scope, owner, or provenance.
// ============================================================

package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

const dreamingPromotionLimit = 100

// LoadDreamingSessionInput loads raw event IDs and derived memories for one session.
func (s *Store) LoadDreamingSessionInput(ctx context.Context, req *core.DreamSessionRequest) (*core.DreamingSessionInput, error) {
	if req == nil {
		return nil, fmt.Errorf("%w: dream_session request is required", core.ErrInvalidArgument)
	}
	if strings.TrimSpace(req.TenantID) == "" || strings.TrimSpace(req.WorkspaceID) == "" || strings.TrimSpace(req.SessionID) == "" {
		return nil, fmt.Errorf("%w: dream_session tenant_id, workspace_id, and session_id are required", core.ErrInvalidArgument)
	}

	rawEventIDs, err := s.dreamingRawEventIDs(ctx, req)
	if err != nil {
		return nil, err
	}
	memories, err := s.dreamingMemoriesForRawEvents(ctx, req, rawEventIDs)
	if err != nil {
		return nil, err
	}
	return &core.DreamingSessionInput{RawEventIDs: rawEventIDs, Memories: memories}, nil
}

// PromoteMemories marks existing memories with a deeper dreaming tier in metadata_json.
func (s *Store) PromoteMemories(ctx context.Context, req *core.DreamingPromotionRequest) (*core.DreamingPromotionResult, error) {
	if err := validateDreamingPromotionRequest(req); err != nil {
		return nil, err
	}
	sql, args := dreamingPromotionStatement(req)
	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("promote dreaming memories: %w", err)
	}
	defer rows.Close()

	result := &core.DreamingPromotionResult{MemoryIDs: []string{}}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan promoted memory id: %w", err)
		}
		result.MemoryIDs = append(result.MemoryIDs, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("scan promoted memories: %w", err)
	}
	result.PromotedCount = len(result.MemoryIDs)
	return result, nil
}

func validateDreamingPromotionRequest(req *core.DreamingPromotionRequest) error {
	if req == nil {
		return fmt.Errorf("%w: dreaming promotion request is required", core.ErrInvalidArgument)
	}
	if strings.TrimSpace(req.TenantID) == "" {
		return fmt.Errorf("%w: dreaming promotion tenant_id is required", core.ErrInvalidArgument)
	}
	if strings.TrimSpace(req.WorkspaceID) == "" {
		return fmt.Errorf("%w: dreaming promotion workspace_id is required", core.ErrInvalidArgument)
	}
	if strings.TrimSpace(req.JobID) == "" {
		return fmt.Errorf("%w: dreaming promotion job_id is required", core.ErrInvalidArgument)
	}
	switch req.Tier {
	case core.DreamingTierMidTerm, core.DreamingTierLongTerm, core.DreamingTierUltraLongTerm:
	case "":
		return fmt.Errorf("%w: dreaming promotion tier is required", core.ErrInvalidArgument)
	default:
		return fmt.Errorf("%w: unsupported dreaming promotion tier %q", core.ErrInvalidArgument, req.Tier)
	}
	if req.MinConfidence < 0 || req.MinConfidence > 1 {
		return fmt.Errorf("%w: dreaming promotion min_confidence must be between 0 and 1", core.ErrInvalidArgument)
	}
	return nil
}

func (s *Store) dreamingRawEventIDs(ctx context.Context, req *core.DreamSessionRequest) ([]string, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id
		FROM raw_events
		WHERE tenant_id = $1
		  AND workspace_id = $2
		  AND session_id = $3
		ORDER BY occurred_at ASC, created_at ASC, id ASC
	`, req.TenantID, req.WorkspaceID, req.SessionID)
	if err != nil {
		return nil, fmt.Errorf("load dream_session raw events: %w", err)
	}
	defer rows.Close()

	ids := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan dream_session raw event id: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("scan dream_session raw event ids: %w", err)
	}
	return ids, nil
}

func (s *Store) dreamingMemoriesForRawEvents(ctx context.Context, req *core.DreamSessionRequest, rawEventIDs []string) ([]*core.Memory, error) {
	if len(rawEventIDs) == 0 {
		return []*core.Memory{}, nil
	}
	rows, err := s.pool.Query(ctx, `
		SELECT DISTINCT m.id, m.tenant_id, m.workspace_id, m.scope, m.group_id,
		       m.owner_entity_id, m.kind, m.artifact_class, m.text, m.fingerprint,
		       m.confidence, m.status, m.valid_from, m.valid_to, m.latest_flag,
		       m.metadata_json, m.embedding_model, m.embedding_dims,
		       m.embedding_updated_at, m.created_at, m.updated_at
		FROM memories m
		JOIN memory_trace mt ON mt.memory_id = m.id
		WHERE m.tenant_id = $1
		  AND m.workspace_id = $2
		  AND m.status = $3
		  AND m.latest_flag = true
		  AND mt.raw_event_ids && $4::text[]
		ORDER BY m.confidence DESC, m.updated_at DESC, m.id ASC
	`, req.TenantID, req.WorkspaceID, core.MemoryStatusActive, rawEventIDs)
	if err != nil {
		return nil, fmt.Errorf("load dream_session memories: %w", err)
	}
	defer rows.Close()
	return scanDreamingMemories(rows)
}

func scanDreamingMemories(rows pgx.Rows) ([]*core.Memory, error) {
	memories := []*core.Memory{}
	for rows.Next() {
		memory := &core.Memory{}
		if err := rows.Scan(
			&memory.ID, &memory.TenantID, &memory.WorkspaceID, &memory.Scope,
			&memory.GroupID, &memory.OwnerEntityID, &memory.Kind, &memory.ArtifactClass,
			&memory.Text, &memory.Fingerprint, &memory.Confidence, &memory.Status,
			&memory.ValidFrom, &memory.ValidTo, &memory.LatestFlag, &memory.MetadataJSON,
			&memory.EmbeddingModel, &memory.EmbeddingDims, &memory.EmbeddingUpdatedAt,
			&memory.CreatedAt, &memory.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan dream_session memory: %w", err)
		}
		memories = append(memories, memory)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("scan dream_session memories: %w", err)
	}
	return memories, nil
}

func dreamingPromotionStatement(req *core.DreamingPromotionRequest) (string, []any) {
	now := req.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
	args := []any{
		req.TenantID,
		req.WorkspaceID,
		req.Tier,
		req.JobID,
		now.UTC(),
		req.MinConfidence,
		nullableStringSlice(req.MemoryIDs),
	}
	filters := []string{
		"tenant_id = $1",
		"workspace_id = $2",
		"status = 'active'",
		"latest_flag = true",
		"confidence >= $6",
		"($7::text[] IS NULL OR id = ANY($7::text[]))",
		"(metadata_json #>> '{dreaming,tier}' IS NULL OR metadata_json #>> '{dreaming,tier}' <> $3)",
	}
	if req.RequireStableKind {
		filters = append(filters, "kind IN ('fact','preference','trait','goal','constraint','decision','procedure')")
		filters = append(filters, "scope <> 'session_scratch'")
	}
	where := strings.Join(filters, "\n		  AND ")
	sql := fmt.Sprintf(`
		WITH selected AS (
			SELECT id
			FROM memories
			WHERE %s
			ORDER BY confidence DESC, updated_at DESC, id ASC
			LIMIT %d
			FOR UPDATE SKIP LOCKED
		)
		UPDATE memories m
		SET metadata_json = jsonb_set(
		        m.metadata_json,
		        '{dreaming}',
		        COALESCE(m.metadata_json->'dreaming', '{}'::jsonb)
		            || jsonb_build_object(
		                'tier', $3,
		                'last_dream_job_id', $4,
		                'promoted_at', to_jsonb($5::timestamptz)
		            ),
		        true
		    ),
		    updated_at = $5
		FROM selected
		WHERE m.id = selected.id
		RETURNING m.id
	`, where, dreamingPromotionLimit)
	return sql, args
}

func nullableStringSlice(values []string) any {
	if len(values) == 0 {
		return nil
	}
	return values
}

```



<!-- Source: internal/store/postgres/dreaming_test.go | bytes=2957 | lines=87 | sha16=4f07fc03d07611cb -->

```go
// ============================================================
// FILE     : internal/store/postgres/dreaming_test.go
// PURPOSE  : Verifies PostgreSQL dreaming promotion query contracts without a live database.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : postgres dreaming helper tests
// DEPENDS  : strings, testing, time, internal/core
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: Promotion SQL must update metadata only, never scope or provenance fields.
// ============================================================

package postgres

import (
	"strings"
	"testing"
	"time"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

func TestDreamingPromotionStatementMarksMetadataOnly(t *testing.T) {
	t.Parallel()

	sql, args := dreamingPromotionStatement(&core.DreamingPromotionRequest{
		JobID:         "job_dream_1",
		TenantID:      "tenant_1",
		WorkspaceID:   "workspace_1",
		MemoryIDs:     []string{"mem_1", "mem_2"},
		Tier:          core.DreamingTierMidTerm,
		Now:           time.Date(2026, time.April, 24, 3, 0, 0, 0, time.UTC),
		MinConfidence: 0.5,
	})

	if !strings.Contains(sql, "jsonb_set") || !strings.Contains(sql, "'{dreaming}'") {
		t.Fatalf("expected dreaming metadata update, got: %s", sql)
	}
	if strings.Contains(sql, "scope =") || strings.Contains(sql, "owner_entity_id =") || strings.Contains(sql, "memory_trace") {
		t.Fatalf("promotion SQL must not mutate scope, owner, or trace: %s", sql)
	}
	if !strings.Contains(sql, "id = ANY($7::text[])") {
		t.Fatalf("expected explicit memory id filter, got: %s", sql)
	}
	if len(args) != 7 {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestDreamingPromotionStatementStableWorkspaceFilters(t *testing.T) {
	t.Parallel()

	sql, _ := dreamingPromotionStatement(&core.DreamingPromotionRequest{
		JobID:             "job_dream_workspace",
		TenantID:          "tenant_1",
		WorkspaceID:       "workspace_1",
		Tier:              core.DreamingTierLongTerm,
		MinConfidence:     0.85,
		RequireStableKind: true,
	})

	if !strings.Contains(sql, "kind IN ('fact','preference','trait','goal','constraint','decision','procedure')") {
		t.Fatalf("expected stable-kind filter, got: %s", sql)
	}
	if !strings.Contains(sql, "scope <> 'session_scratch'") {
		t.Fatalf("expected session scratch exclusion, got: %s", sql)
	}
	if !strings.Contains(sql, "FOR UPDATE SKIP LOCKED") {
		t.Fatalf("expected safe concurrent promotion locking, got: %s", sql)
	}
}

func TestValidateDreamingPromotionRequestRejectsUnsupportedTier(t *testing.T) {
	t.Parallel()

	err := validateDreamingPromotionRequest(&core.DreamingPromotionRequest{
		JobID:       "job_dream_1",
		TenantID:    "tenant_1",
		WorkspaceID: "workspace_1",
		Tier:        core.DreamingTierShortTerm,
	})
	if err == nil {
		t.Fatalf("expected unsupported short-term promotion to fail")
	}
}

```



<!-- Source: internal/store/postgres/groups.go | bytes=4009 | lines=114 | sha16=e1afbfa43581d7ee -->

```go
// ============================================================
// FILE     : internal/store/postgres/groups.go
// PURPOSE  : Implements PostgreSQL persistence for memory groups and memberships.
// LAYER    : infra
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : CreateMemoryGroup, AddMembership, ListMemberships, ListMembershipsForEntity
// DEPENDS  : context, fmt, internal/core
// USED_BY  : recall assembler, Stage 2 source adapters, future group APIs
// ------------------------------------------------------------
// AGENT_NOTE: Group visibility must be membership-backed before group_shared memory is returned.
// ============================================================

package postgres

import (
	"context"
	"fmt"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

// CreateMemoryGroup writes a memory group.
func (s *Store) CreateMemoryGroup(ctx context.Context, group *core.MemoryGroup) error {
	if group == nil {
		return fmt.Errorf("%w: memory group is required", core.ErrInvalidArgument)
	}
	if group.ID == "" {
		id, err := newID("group")
		if err != nil {
			return err
		}
		group.ID = id
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO memory_groups (id, tenant_id, workspace_id, name, description, created_at)
		VALUES ($1,$2,$3,$4,$5,$6)
	`, group.ID, group.TenantID, group.WorkspaceID, group.Name, group.Description, timeOrNow(group.CreatedAt))
	if err != nil {
		return fmt.Errorf("insert memory group: %w", err)
	}
	return nil
}

// AddMembership adds an entity to a memory group.
func (s *Store) AddMembership(ctx context.Context, membership *core.MemoryGroupMembership) error {
	if membership == nil {
		return fmt.Errorf("%w: memory group membership is required", core.ErrInvalidArgument)
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO memory_group_memberships (group_id, entity_id, created_at)
		VALUES ($1,$2,$3)
		ON CONFLICT (group_id, entity_id) DO NOTHING
	`, membership.GroupID, membership.EntityID, timeOrNow(membership.CreatedAt))
	if err != nil {
		return fmt.Errorf("insert memory group membership: %w", err)
	}
	return nil
}

// ListMemberships loads memberships for a memory group.
func (s *Store) ListMemberships(ctx context.Context, groupID string) ([]*core.MemoryGroupMembership, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT group_id, entity_id, created_at
		FROM memory_group_memberships
		WHERE group_id = $1
		ORDER BY created_at ASC
	`, groupID)
	if err != nil {
		return nil, fmt.Errorf("query memory group memberships: %w", err)
	}
	defer rows.Close()
	return scanMemoryGroupMemberships(rows)
}

// ListMembershipsForEntity loads groups visible to an entity in one workspace.
func (s *Store) ListMembershipsForEntity(ctx context.Context, tenantID string, workspaceID string, entityID string) ([]*core.MemoryGroupMembership, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT gm.group_id, gm.entity_id, gm.created_at
		FROM memory_group_memberships gm
		JOIN memory_groups g ON g.id = gm.group_id
		WHERE g.tenant_id = $1
		  AND g.workspace_id = $2
		  AND gm.entity_id = $3
		ORDER BY gm.created_at ASC
	`, tenantID, workspaceID, entityID)
	if err != nil {
		return nil, fmt.Errorf("query entity memory group memberships: %w", err)
	}
	defer rows.Close()
	return scanMemoryGroupMemberships(rows)
}

type memoryGroupMembershipRows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}

func scanMemoryGroupMemberships(rows memoryGroupMembershipRows) ([]*core.MemoryGroupMembership, error) {
	memberships := make([]*core.MemoryGroupMembership, 0, 8)
	for rows.Next() {
		membership := &core.MemoryGroupMembership{}
		if err := rows.Scan(&membership.GroupID, &membership.EntityID, &membership.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan memory group membership: %w", err)
		}
		memberships = append(memberships, membership)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate memory group memberships: %w", err)
	}
	return memberships, nil
}

```



<!-- Source: internal/store/postgres/helpers.go | bytes=1304 | lines=52 | sha16=a78a500ecf85ea17 -->

```go
// ============================================================
// FILE     : internal/store/postgres/helpers.go
// PURPOSE  : Provides shared PostgreSQL store helpers for IDs, JSON, and row scanning.
// LAYER    : infra
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : none
// DEPENDS  : crypto/rand, encoding/hex, encoding/json, time
// USED_BY  : internal/store/postgres implementations
// ------------------------------------------------------------
// AGENT_NOTE: Keep helper behavior deterministic where callers already provide stable IDs.
// ============================================================

package postgres

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

func newID(prefix string) (string, error) {
	var b [12]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("generate id: %w", err)
	}
	return prefix + "_" + hex.EncodeToString(b[:]), nil
}

func rawJSONOrEmpty(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 {
		return json.RawMessage(`{}`)
	}
	return raw
}

func rawJSONOrNil(raw json.RawMessage) any {
	if len(raw) == 0 {
		return nil
	}
	return raw
}

func timeOrNow(value time.Time) time.Time {
	if value.IsZero() {
		return time.Now().UTC()
	}
	return value.UTC()
}

```



<!-- Source: internal/store/postgres/jobs.go | bytes=14440 | lines=467 | sha16=b518243b8ea5aa24 -->

```go
// ============================================================
// FILE     : internal/store/postgres/jobs.go
// PURPOSE  : Implements PostgreSQL-backed ingest job queue operations.
// LAYER    : infra
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : EnqueueJobs, ClaimJobs, CompleteJob, FailJob, BlockJob, GetJobBacklogMetrics, ListBlockedJobs, RequeueBlockedJob
// DEPENDS  : internal/core, github.com/jackc/pgx/v5/pgconn
// USED_BY  : internal/ingest, cmd/worker, cmd/cli
// ------------------------------------------------------------
// AGENT_NOTE: Job claiming must use database locking to avoid duplicate worker ownership.
// ============================================================

package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

const defaultBlockedJobListLimit = 20
const defaultJobMetricsDrainWindow = 15 * time.Minute
const maxJobMetricsDrainWindow = 24 * time.Hour

type jobExecutor interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

type ingestJobRows interface {
	Close()
	Next() bool
	Scan(dest ...any) error
	Err() error
}

type jobMetricsScanner interface {
	Scan(dest ...any) error
}

// EnqueueJobs inserts queued jobs and returns their IDs.
func (s *Store) EnqueueJobs(ctx context.Context, jobs []*core.IngestJob) ([]string, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin job enqueue: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	ids := make([]string, 0, len(jobs))
	for _, job := range jobs {
		if job == nil {
			continue
		}
		id := job.ID
		if id == "" {
			id, err = newID("job")
			if err != nil {
				return nil, err
			}
		}
		row := tx.QueryRow(ctx, `
			INSERT INTO ingest_jobs (
				id, tenant_id, workspace_id, job_kind, status, raw_event_ids,
				payload_json, attempts, available_at, locked_by, locked_at,
				last_error, created_at, updated_at
			)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
			RETURNING id
		`, id, job.TenantID, job.WorkspaceID, job.JobKind, valueOr(job.Status, "queued"),
			job.RawEventIDs, rawJSONOrEmpty(job.PayloadJSON), job.Attempts, timeOrNow(job.AvailableAt),
			job.LockedBy, job.LockedAt, job.LastError, timeOrNow(job.CreatedAt), timeOrNow(job.UpdatedAt))
		var insertedID string
		if err := row.Scan(&insertedID); err != nil {
			return nil, fmt.Errorf("insert ingest job: %w", err)
		}
		ids = append(ids, insertedID)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit job enqueue: %w", err)
	}
	return ids, nil
}

// ClaimJobs claims available jobs for one worker using FOR UPDATE SKIP LOCKED.
func (s *Store) ClaimJobs(ctx context.Context, workerID string, limit int) ([]*core.IngestJob, error) {
	if limit <= 0 {
		limit = 1
	}
	rows, err := s.pool.Query(ctx, `
		WITH claimed AS (
			SELECT id
			FROM ingest_jobs
			WHERE status = 'queued' AND available_at <= now()
			ORDER BY available_at ASC, created_at ASC
			LIMIT $2
			FOR UPDATE SKIP LOCKED
		)
		UPDATE ingest_jobs
		SET status = 'running',
		    locked_by = $1,
		    locked_at = now(),
		    updated_at = now()
		WHERE id IN (SELECT id FROM claimed)
		RETURNING id, tenant_id, workspace_id, job_kind, status, raw_event_ids,
		          payload_json, attempts, available_at, locked_by, locked_at,
		          last_error, created_at, updated_at
	`, workerID, limit)
	if err != nil {
		return nil, fmt.Errorf("claim jobs: %w", err)
	}
	jobs, err := scanIngestJobRows(rows, limit)
	if err != nil {
		return nil, err
	}
	return jobs, nil
}

// ListBlockedJobs lists jobs blocked by deterministic unsupported work for operator inspection.
func (s *Store) ListBlockedJobs(ctx context.Context, limit int) ([]*core.IngestJob, error) {
	limit = normalizeBlockedJobListLimit(limit)
	sql, args := listBlockedJobsStatement(limit)
	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("list blocked jobs: %w", err)
	}
	return scanIngestJobRows(rows, limit)
}

func listBlockedJobsStatement(limit int) (string, []any) {
	limit = normalizeBlockedJobListLimit(limit)
	return `
		SELECT id, tenant_id, workspace_id, job_kind, status, raw_event_ids,
		       payload_json, attempts, available_at, locked_by, locked_at,
		       last_error, created_at, updated_at
		FROM ingest_jobs
		WHERE status = 'blocked'
		ORDER BY updated_at DESC, created_at DESC
		LIMIT $1
	`, []any{limit}
}

func normalizeBlockedJobListLimit(limit int) int {
	if limit <= 0 {
		return defaultBlockedJobListLimit
	}
	return limit
}

// GetJobBacklogMetrics returns read-only worker queue status and recovery estimates.
func (s *Store) GetJobBacklogMetrics(ctx context.Context, req *core.JobBacklogMetricsRequest) (*core.JobBacklogMetrics, error) {
	normalized, err := normalizeJobBacklogMetricsRequest(req)
	if err != nil {
		return nil, err
	}
	sql, args := jobBacklogMetricsStatement(normalized)
	metrics, err := scanJobBacklogMetrics(s.pool.QueryRow(ctx, sql, args...), normalized.DrainWindow)
	if err != nil {
		return nil, fmt.Errorf("get job backlog metrics: %w", err)
	}
	return metrics, nil
}

func normalizeJobBacklogMetricsRequest(req *core.JobBacklogMetricsRequest) (*core.JobBacklogMetricsRequest, error) {
	normalized := &core.JobBacklogMetricsRequest{}
	if req != nil {
		*normalized = *req
	}
	normalized.TenantID = strings.TrimSpace(normalized.TenantID)
	normalized.WorkspaceID = strings.TrimSpace(normalized.WorkspaceID)
	if normalized.DrainWindow == 0 {
		normalized.DrainWindow = defaultJobMetricsDrainWindow
	}
	if normalized.DrainWindow < time.Second {
		return nil, fmt.Errorf("%w: job metrics drain window must be at least 1s", core.ErrInvalidArgument)
	}
	if normalized.DrainWindow > maxJobMetricsDrainWindow {
		return nil, fmt.Errorf("%w: job metrics drain window must be at most 24h", core.ErrInvalidArgument)
	}
	if normalized.GeneratedNow.IsZero() {
		normalized.GeneratedNow = time.Now().UTC()
	}
	return normalized, nil
}

func jobBacklogMetricsStatement(req *core.JobBacklogMetricsRequest) (string, []any) {
	windowSeconds := int64(req.DrainWindow.Seconds())
	return `
		WITH snapshot AS (
			SELECT $4::timestamptz AS generated_at
		),
		filtered AS (
			SELECT *
			FROM ingest_jobs
			WHERE ($1 = '' OR tenant_id = $1)
			  AND ($2 = '' OR workspace_id = $2)
		),
		counts AS (
			SELECT
				count(*) FILTER (WHERE status = 'queued')::int AS queued,
				count(*) FILTER (WHERE status = 'queued' AND available_at <= (SELECT generated_at FROM snapshot))::int AS ready_queued,
				count(*) FILTER (WHERE status = 'running')::int AS running,
				count(*) FILTER (WHERE status = 'failed')::int AS failed,
				count(*) FILTER (WHERE status = 'blocked')::int AS blocked,
				count(*) FILTER (WHERE status = 'complete')::int AS complete,
				count(*) FILTER (WHERE status = 'queued' AND attempts > 0)::int AS retryable_queued_attempts
			FROM filtered
		),
		oldest AS (
			SELECT min(available_at) AS oldest_queued_at
			FROM filtered, snapshot
			WHERE status = 'queued'
			  AND available_at <= snapshot.generated_at
		),
		oldest_running AS (
			SELECT min(COALESCE(locked_at, updated_at)) AS oldest_running_at
			FROM filtered
			WHERE status = 'running'
		),
		drain AS (
			SELECT count(*)::int AS completed_in_window
			FROM filtered, snapshot
			WHERE status = 'complete'
			  AND updated_at >= snapshot.generated_at - ($3::bigint * interval '1 second')
		)
		SELECT
			counts.queued,
			counts.ready_queued,
			counts.running,
			counts.failed,
			counts.blocked,
			counts.complete,
			counts.retryable_queued_attempts,
			COALESCE(oldest.oldest_queued_at, snapshot.generated_at) AS oldest_queued_at,
			oldest.oldest_queued_at IS NOT NULL AS has_oldest_queued_at,
			CASE
				WHEN oldest.oldest_queued_at IS NULL THEN 0
				ELSE GREATEST(0, extract(epoch FROM (snapshot.generated_at - oldest.oldest_queued_at)))::bigint
			END AS oldest_queued_age_seconds,
			COALESCE(oldest_running.oldest_running_at, snapshot.generated_at) AS oldest_running_at,
			oldest_running.oldest_running_at IS NOT NULL AS has_oldest_running_at,
			CASE
				WHEN oldest_running.oldest_running_at IS NULL THEN 0
				ELSE GREATEST(0, extract(epoch FROM (snapshot.generated_at - oldest_running.oldest_running_at)))::bigint
			END AS oldest_running_age_seconds,
			$3::bigint AS drain_window_seconds,
			drain.completed_in_window,
			CASE
				WHEN drain.completed_in_window = 0 THEN 0
				ELSE drain.completed_in_window::double precision / ($3::double precision / 60.0)
			END AS drain_rate_jobs_per_minute,
			drain.completed_in_window > 0 AS has_drain_rate,
			CASE
				WHEN counts.queued = 0 THEN 0
				WHEN drain.completed_in_window = 0 THEN 0
				ELSE ceil(counts.queued::double precision / (drain.completed_in_window::double precision / $3::double precision))::bigint
			END AS recovery_eta_seconds,
			(counts.queued = 0 OR drain.completed_in_window > 0) AS has_recovery_eta,
			snapshot.generated_at
		FROM snapshot, counts, oldest, oldest_running, drain
	`, []any{req.TenantID, req.WorkspaceID, windowSeconds, req.GeneratedNow.UTC()}
}

func scanJobBacklogMetrics(row jobMetricsScanner, _ time.Duration) (*core.JobBacklogMetrics, error) {
	metrics := &core.JobBacklogMetrics{}
	var oldestQueuedAt time.Time
	var hasOldestQueuedAt bool
	var oldestQueuedAgeSeconds int64
	var oldestRunningAt time.Time
	var hasOldestRunningAt bool
	var oldestRunningAgeSeconds int64
	var drainRateJobsPerMinute float64
	var hasDrainRate bool
	var recoveryETASeconds int64
	var hasRecoveryETA bool
	err := row.Scan(
		&metrics.Counts.Queued,
		&metrics.Counts.ReadyQueued,
		&metrics.Counts.Running,
		&metrics.Counts.Failed,
		&metrics.Counts.Blocked,
		&metrics.Counts.Complete,
		&metrics.RetryableQueuedAttempts,
		&oldestQueuedAt,
		&hasOldestQueuedAt,
		&oldestQueuedAgeSeconds,
		&oldestRunningAt,
		&hasOldestRunningAt,
		&oldestRunningAgeSeconds,
		&metrics.DrainWindowSeconds,
		&metrics.CompletedInWindow,
		&drainRateJobsPerMinute,
		&hasDrainRate,
		&recoveryETASeconds,
		&hasRecoveryETA,
		&metrics.GeneratedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, core.ErrNotFound
		}
		return nil, err
	}
	if hasOldestQueuedAt {
		metrics.OldestQueuedAt = &oldestQueuedAt
		metrics.OldestQueuedAgeSeconds = &oldestQueuedAgeSeconds
	}
	if hasOldestRunningAt {
		metrics.OldestRunningAt = &oldestRunningAt
		metrics.OldestRunningAgeSeconds = &oldestRunningAgeSeconds
	}
	if hasDrainRate {
		metrics.DrainRateJobsPerMinute = &drainRateJobsPerMinute
	}
	if hasRecoveryETA {
		metrics.RecoveryETASeconds = &recoveryETASeconds
	}
	return metrics, nil
}

// CompleteJob marks a job complete.
func (s *Store) CompleteJob(ctx context.Context, jobID string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE ingest_jobs
		SET status = 'complete',
		    locked_by = NULL,
		    locked_at = NULL,
		    updated_at = now()
		WHERE id = $1
	`, jobID)
	if err != nil {
		return fmt.Errorf("complete job: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return core.ErrNotFound
	}
	return nil
}

// FailJob records a failed attempt and schedules retry.
func (s *Store) FailJob(ctx context.Context, jobID string, jobErr error) error {
	return failJob(ctx, s.pool, jobID, jobErr)
}

func failJob(ctx context.Context, exec jobExecutor, jobID string, jobErr error) error {
	tag, err := exec.Exec(ctx, `
		UPDATE ingest_jobs
		SET status = 'queued',
		    attempts = attempts + 1,
		    available_at = now() + interval '30 seconds',
		    locked_by = NULL,
		    locked_at = NULL,
		    last_error = $2,
		    updated_at = now()
		WHERE id = $1
	`, jobID, jobErrorString(jobErr))
	if err != nil {
		return fmt.Errorf("fail job: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return core.ErrNotFound
	}
	return nil
}

// BlockJob records deterministic unsupported work without scheduling retry.
func (s *Store) BlockJob(ctx context.Context, jobID string, jobErr error) error {
	return blockJob(ctx, s.pool, jobID, jobErr)
}

func blockJob(ctx context.Context, exec jobExecutor, jobID string, jobErr error) error {
	tag, err := exec.Exec(ctx, `
		UPDATE ingest_jobs
		SET status = 'blocked',
		    attempts = attempts + 1,
		    available_at = now(),
		    locked_by = NULL,
		    locked_at = NULL,
		    last_error = $2,
		    updated_at = now()
		WHERE id = $1
	`, jobID, jobErrorString(jobErr))
	if err != nil {
		return fmt.Errorf("block job: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return core.ErrNotFound
	}
	return nil
}

func jobErrorString(jobErr error) string {
	if jobErr == nil {
		return ""
	}
	return jobErr.Error()
}

// RequeueBlockedJob manually returns one blocked job to the queued worker pool.
func (s *Store) RequeueBlockedJob(ctx context.Context, jobID string) error {
	return requeueBlockedJob(ctx, s.pool, jobID)
}

func requeueBlockedJob(ctx context.Context, exec jobExecutor, jobID string) error {
	if jobID == "" {
		return fmt.Errorf("%w: blocked job id is required", core.ErrInvalidArgument)
	}
	tag, err := exec.Exec(ctx, `
		UPDATE ingest_jobs
		SET status = 'queued',
		    available_at = now(),
		    locked_by = NULL,
		    locked_at = NULL,
		    updated_at = now()
		WHERE id = $1 AND status = 'blocked'
	`, jobID)
	if err != nil {
		return fmt.Errorf("requeue blocked job: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return core.ErrNotFound
	}
	return nil
}

type ingestJobScanner interface {
	Scan(dest ...any) error
}

func scanIngestJobRows(rows ingestJobRows, capacity int) ([]*core.IngestJob, error) {
	defer rows.Close()
	jobs := make([]*core.IngestJob, 0, capacity)
	for rows.Next() {
		job, err := scanIngestJob(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate ingest jobs: %w", err)
	}
	return jobs, nil
}

func scanIngestJob(row ingestJobScanner) (*core.IngestJob, error) {
	job := &core.IngestJob{}
	if err := row.Scan(&job.ID, &job.TenantID, &job.WorkspaceID, &job.JobKind, &job.Status,
		&job.RawEventIDs, &job.PayloadJSON, &job.Attempts, &job.AvailableAt, &job.LockedBy,
		&job.LockedAt, &job.LastError, &job.CreatedAt, &job.UpdatedAt); err != nil {
		return nil, fmt.Errorf("scan ingest job: %w", err)
	}
	return job, nil
}

func valueOr(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

```



<!-- Source: internal/store/postgres/jobs_test.go | bytes=13185 | lines=435 | sha16=35302f2b62beda48 -->

```go
// ============================================================
// FILE     : internal/store/postgres/jobs_test.go
// PURPOSE  : Verifies PostgreSQL job status update contracts without requiring a live database.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : postgres job status, inspection, and manual requeue tests
// DEPENDS  : context, encoding/json, errors, strings, testing, time, internal/core, github.com/jackc/pgx/v5/pgconn
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: These tests lock SQL status semantics; keep deterministic blocks out of the retry queue.
// ============================================================

package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

func TestListBlockedJobsStatementOnlySelectsBlockedStatus(t *testing.T) {
	t.Parallel()

	sql, args := listBlockedJobsStatement(7)

	if !strings.Contains(sql, "WHERE status = 'blocked'") {
		t.Fatalf("expected blocked status predicate, got: %s", sql)
	}
	if !strings.Contains(sql, "ORDER BY updated_at DESC, created_at DESC") {
		t.Fatalf("expected deterministic newest-first order, got: %s", sql)
	}
	if strings.Contains(sql, "status = 'queued'") {
		t.Fatalf("blocked inspection must not inspect retry queue, got: %s", sql)
	}
	if len(args) != 1 || args[0] != 7 {
		t.Fatalf("unexpected list blocked args: %#v", args)
	}
}

func TestJobBacklogMetricsStatementIsReadOnlyAndSeparatesStatuses(t *testing.T) {
	t.Parallel()

	sql, args := jobBacklogMetricsStatement(&core.JobBacklogMetricsRequest{
		TenantID:     "tenant_1",
		WorkspaceID:  "workspace_1",
		DrainWindow:  15 * time.Minute,
		GeneratedNow: time.Date(2026, time.April, 24, 8, 0, 0, 0, time.UTC),
	})
	upperSQL := strings.ToUpper(sql)
	for _, forbidden := range []string{"UPDATE ", "INSERT ", "DELETE ", "FOR UPDATE", "SKIP LOCKED"} {
		if strings.Contains(upperSQL, forbidden) {
			t.Fatalf("metrics statement must be read-only; found %q in:\n%s", forbidden, sql)
		}
	}
	for _, want := range []string{
		"status = 'queued'",
		"status = 'queued' AND available_at <=",
		"status = 'running'",
		"status = 'failed'",
		"status = 'blocked'",
		"status = 'complete'",
		"available_at <= snapshot.generated_at",
		"status = 'queued' AND attempts > 0",
		"status = 'running'",
		"COALESCE(locked_at, updated_at)",
	} {
		if !strings.Contains(sql, want) {
			t.Fatalf("expected metrics SQL to contain %q, got:\n%s", want, sql)
		}
	}
	if len(args) != 4 || args[0] != "tenant_1" || args[1] != "workspace_1" || args[2] != int64(900) {
		t.Fatalf("unexpected metrics args: %#v", args)
	}
}

func TestNormalizeJobBacklogMetricsRequestDefaultsAndRejectsInvalid(t *testing.T) {
	t.Parallel()

	req, err := normalizeJobBacklogMetricsRequest(nil)
	if err != nil {
		t.Fatalf("normalizeJobBacklogMetricsRequest returned error: %v", err)
	}
	if req.DrainWindow != defaultJobMetricsDrainWindow {
		t.Fatalf("expected default window %s, got %s", defaultJobMetricsDrainWindow, req.DrainWindow)
	}

	_, err = normalizeJobBacklogMetricsRequest(&core.JobBacklogMetricsRequest{DrainWindow: -time.Second})
	if !errors.Is(err, core.ErrInvalidArgument) {
		t.Fatalf("expected invalid argument for negative window, got %v", err)
	}

	_, err = normalizeJobBacklogMetricsRequest(&core.JobBacklogMetricsRequest{DrainWindow: time.Nanosecond})
	if !errors.Is(err, core.ErrInvalidArgument) {
		t.Fatalf("expected invalid argument for sub-second window, got %v", err)
	}

	_, err = normalizeJobBacklogMetricsRequest(&core.JobBacklogMetricsRequest{DrainWindow: 25 * time.Hour})
	if !errors.Is(err, core.ErrInvalidArgument) {
		t.Fatalf("expected invalid argument for oversized window, got %v", err)
	}
}

func TestScanJobBacklogMetricsHandlesUnavailableDrainRateAndETA(t *testing.T) {
	t.Parallel()

	generatedAt := time.Date(2026, time.April, 24, 8, 0, 0, 0, time.UTC)
	row := fakeJobMetricsRow{
		values: []any{
			3,
			1,
			1,
			0,
			2,
			10,
			1,
			generatedAt,
			false,
			int64(0),
			generatedAt,
			false,
			int64(0),
			int64(900),
			0,
			float64(0),
			false,
			int64(0),
			false,
			generatedAt,
		},
	}

	metrics, err := scanJobBacklogMetrics(row, 15*time.Minute)
	if err != nil {
		t.Fatalf("scanJobBacklogMetrics returned error: %v", err)
	}
	if metrics.Counts.Queued != 3 || metrics.Counts.ReadyQueued != 1 || metrics.Counts.Blocked != 2 || metrics.RetryableQueuedAttempts != 1 {
		t.Fatalf("unexpected metrics counts: %#v", metrics)
	}
	if metrics.OldestQueuedAt != nil || metrics.OldestQueuedAgeSeconds != nil {
		t.Fatalf("expected unavailable oldest queued fields, got %#v", metrics)
	}
	if metrics.OldestRunningAt != nil || metrics.OldestRunningAgeSeconds != nil {
		t.Fatalf("expected unavailable oldest running fields, got %#v", metrics)
	}
	if metrics.DrainRateJobsPerMinute != nil || metrics.RecoveryETASeconds != nil {
		t.Fatalf("expected unavailable drain rate and ETA, got %#v", metrics)
	}
}

func TestScanJobBacklogMetricsReturnsAvailableRecoveryFields(t *testing.T) {
	t.Parallel()

	generatedAt := time.Date(2026, time.April, 24, 8, 0, 0, 0, time.UTC)
	oldestAt := generatedAt.Add(-5 * time.Minute)
	oldestRunningAt := generatedAt.Add(-7 * time.Minute)
	row := fakeJobMetricsRow{
		values: []any{
			4,
			4,
			1,
			0,
			1,
			20,
			2,
			oldestAt,
			true,
			int64(300),
			oldestRunningAt,
			true,
			int64(420),
			int64(600),
			12,
			float64(1.2),
			true,
			int64(200),
			true,
			generatedAt,
		},
	}

	metrics, err := scanJobBacklogMetrics(row, 10*time.Minute)
	if err != nil {
		t.Fatalf("scanJobBacklogMetrics returned error: %v", err)
	}
	if metrics.OldestQueuedAt == nil || !metrics.OldestQueuedAt.Equal(oldestAt) {
		t.Fatalf("unexpected oldest queued at: %#v", metrics.OldestQueuedAt)
	}
	if metrics.OldestQueuedAgeSeconds == nil || *metrics.OldestQueuedAgeSeconds != 300 {
		t.Fatalf("unexpected oldest queued age: %#v", metrics.OldestQueuedAgeSeconds)
	}
	if metrics.OldestRunningAt == nil || !metrics.OldestRunningAt.Equal(oldestRunningAt) {
		t.Fatalf("unexpected oldest running at: %#v", metrics.OldestRunningAt)
	}
	if metrics.OldestRunningAgeSeconds == nil || *metrics.OldestRunningAgeSeconds != 420 {
		t.Fatalf("unexpected oldest running age: %#v", metrics.OldestRunningAgeSeconds)
	}
	if metrics.DrainRateJobsPerMinute == nil || *metrics.DrainRateJobsPerMinute != 1.2 {
		t.Fatalf("unexpected drain rate: %#v", metrics.DrainRateJobsPerMinute)
	}
	if metrics.RecoveryETASeconds == nil || *metrics.RecoveryETASeconds != 200 {
		t.Fatalf("unexpected recovery ETA: %#v", metrics.RecoveryETASeconds)
	}
}

func TestScanIngestJobRowsReturnsBlockedJobs(t *testing.T) {
	t.Parallel()

	updatedAt := time.Date(2026, time.April, 24, 8, 0, 0, 0, time.UTC)
	lastError := "not implemented: update_memory"
	rows := &fakeIngestJobRows{jobs: []*core.IngestJob{
		{
			ID:          "job_blocked_1",
			TenantID:    "tenant_1",
			WorkspaceID: "workspace_1",
			JobKind:     core.JobKindProcessTurnEvent,
			Status:      "blocked",
			RawEventIDs: []string{"evt_1"},
			PayloadJSON: json.RawMessage(`{"session_id":"session_1"}`),
			Attempts:    3,
			AvailableAt: updatedAt,
			LastError:   &lastError,
			CreatedAt:   updatedAt.Add(-time.Hour),
			UpdatedAt:   updatedAt,
		},
	}}

	jobs, err := scanIngestJobRows(rows, 1)
	if err != nil {
		t.Fatalf("scanIngestJobRows returned error: %v", err)
	}
	if len(jobs) != 1 {
		t.Fatalf("expected one blocked job, got %d", len(jobs))
	}
	if jobs[0].ID != "job_blocked_1" || jobs[0].Status != "blocked" {
		t.Fatalf("unexpected blocked job: %#v", jobs[0])
	}
	if jobs[0].LastError == nil || *jobs[0].LastError != lastError {
		t.Fatalf("expected last_error to survive inspection, got %#v", jobs[0].LastError)
	}
	if !rows.closed {
		t.Fatalf("expected rows to be closed after scanning")
	}
}

func TestRequeueBlockedJobRequiresBlockedStatus(t *testing.T) {
	t.Parallel()

	exec := &recordingJobExecutor{tag: pgconn.NewCommandTag("UPDATE 1")}

	if err := requeueBlockedJob(context.Background(), exec, "job_blocked_1"); err != nil {
		t.Fatalf("requeueBlockedJob returned error: %v", err)
	}

	if !strings.Contains(exec.sql, "status = 'queued'") {
		t.Fatalf("expected requeue to set queued status, got: %s", exec.sql)
	}
	if !strings.Contains(exec.sql, "WHERE id = $1 AND status = 'blocked'") {
		t.Fatalf("expected requeue to require currently blocked job, got: %s", exec.sql)
	}
	if strings.Contains(exec.sql, "attempts = attempts + 1") {
		t.Fatalf("manual requeue must not increment attempts, got: %s", exec.sql)
	}
	if strings.Contains(exec.sql, "interval '30 seconds'") {
		t.Fatalf("manual requeue must not schedule retry interval, got: %s", exec.sql)
	}
	if len(exec.args) != 1 || exec.args[0] != "job_blocked_1" {
		t.Fatalf("unexpected requeue args: %#v", exec.args)
	}
}

func TestRequeueBlockedJobReturnsNotFoundWhenJobIsNotBlocked(t *testing.T) {
	t.Parallel()

	exec := &recordingJobExecutor{tag: pgconn.NewCommandTag("UPDATE 0")}

	err := requeueBlockedJob(context.Background(), exec, "job_not_blocked")
	if !errors.Is(err, core.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestBlockJobMarksJobBlockedWithoutRetry(t *testing.T) {
	t.Parallel()

	exec := &recordingJobExecutor{tag: pgconn.NewCommandTag("UPDATE 1")}
	jobErr := errors.New("unsupported apply work")

	if err := blockJob(context.Background(), exec, "job_1", jobErr); err != nil {
		t.Fatalf("blockJob returned error: %v", err)
	}

	if !strings.Contains(exec.sql, "status = 'blocked'") {
		t.Fatalf("expected blocked status update SQL, got: %s", exec.sql)
	}
	if strings.Contains(exec.sql, "interval '30 seconds'") {
		t.Fatalf("blocked jobs must not schedule the retry interval, got: %s", exec.sql)
	}
	if len(exec.args) != 2 || exec.args[0] != "job_1" || exec.args[1] != jobErr.Error() {
		t.Fatalf("unexpected blockJob args: %#v", exec.args)
	}
}

func TestFailJobSchedulesRetry(t *testing.T) {
	t.Parallel()

	exec := &recordingJobExecutor{tag: pgconn.NewCommandTag("UPDATE 1")}

	if err := failJob(context.Background(), exec, "job_1", errors.New("transient apply failure")); err != nil {
		t.Fatalf("failJob returned error: %v", err)
	}

	if !strings.Contains(exec.sql, "status = 'queued'") {
		t.Fatalf("expected queued retry SQL, got: %s", exec.sql)
	}
	if !strings.Contains(exec.sql, "interval '30 seconds'") {
		t.Fatalf("expected retry interval SQL, got: %s", exec.sql)
	}
}

func TestBlockJobReturnsNotFoundWhenNoRowsUpdate(t *testing.T) {
	t.Parallel()

	exec := &recordingJobExecutor{tag: pgconn.NewCommandTag("UPDATE 0")}

	err := blockJob(context.Background(), exec, "missing_job", errors.New("unsupported apply work"))
	if !errors.Is(err, core.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

type fakeIngestJobRows struct {
	jobs   []*core.IngestJob
	idx    int
	closed bool
	err    error
}

func (r *fakeIngestJobRows) Close() {
	r.closed = true
}

func (r *fakeIngestJobRows) Next() bool {
	if r.idx >= len(r.jobs) {
		return false
	}
	r.idx++
	return true
}

func (r *fakeIngestJobRows) Scan(dest ...any) error {
	job := r.jobs[r.idx-1]
	assignScannedJobValue(dest[0], job.ID)
	assignScannedJobValue(dest[1], job.TenantID)
	assignScannedJobValue(dest[2], job.WorkspaceID)
	assignScannedJobValue(dest[3], job.JobKind)
	assignScannedJobValue(dest[4], job.Status)
	assignScannedJobValue(dest[5], job.RawEventIDs)
	assignScannedJobValue(dest[6], job.PayloadJSON)
	assignScannedJobValue(dest[7], job.Attempts)
	assignScannedJobValue(dest[8], job.AvailableAt)
	assignScannedJobValue(dest[9], job.LockedBy)
	assignScannedJobValue(dest[10], job.LockedAt)
	assignScannedJobValue(dest[11], job.LastError)
	assignScannedJobValue(dest[12], job.CreatedAt)
	assignScannedJobValue(dest[13], job.UpdatedAt)
	return nil
}

func (r *fakeIngestJobRows) Err() error {
	return r.err
}

func assignScannedJobValue(dest any, value any) {
	switch target := dest.(type) {
	case *string:
		*target = value.(string)
	case *core.JobKind:
		*target = value.(core.JobKind)
	case *[]string:
		*target = append([]string(nil), value.([]string)...)
	case *json.RawMessage:
		*target = append(json.RawMessage(nil), value.(json.RawMessage)...)
	case *int:
		*target = value.(int)
	case *time.Time:
		*target = value.(time.Time)
	case *bool:
		*target = value.(bool)
	case *int64:
		*target = value.(int64)
	case *float64:
		*target = value.(float64)
	case **string:
		*target = value.(*string)
	case **time.Time:
		*target = value.(*time.Time)
	}
}

type fakeJobMetricsRow struct {
	values []any
	err    error
}

func (r fakeJobMetricsRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	for i := range dest {
		assignScannedJobValue(dest[i], r.values[i])
	}
	return nil
}

type recordingJobExecutor struct {
	sql  string
	args []any
	tag  pgconn.CommandTag
	err  error
}

func (e *recordingJobExecutor) Exec(_ context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	e.sql = sql
	e.args = append([]any(nil), arguments...)
	return e.tag, e.err
}

```



<!-- Source: internal/store/postgres/memories.go | bytes=20855 | lines=604 | sha16=2160722111596dc8 -->

```go
// ============================================================
// FILE     : internal/store/postgres/memories.go
// PURPOSE  : Implements PostgreSQL persistence for memories, edges, and traces.
// LAYER    : infra
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : CreateMemoryWithTrace, CreateMemoryWithTraceAndEdge, CreateMemoryWithTraceAndUpdateEdge, UpsertMemory, GetMemory, UpsertMemoryEdge, WriteMemoryTrace, ExplainMemory
// DEPENDS  : internal/core, github.com/jackc/pgx/v5, github.com/jackc/pgx/v5/pgconn
// USED_BY  : graph apply engine, recall assembler, provenance APIs
// ------------------------------------------------------------
// AGENT_NOTE: Memory writes must preserve explicit scope and mandatory provenance.
// ============================================================

package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

type memoryExecutor interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

// CreateMemoryWithTrace writes a derived memory and its mandatory provenance in one transaction.
func (s *Store) CreateMemoryWithTrace(ctx context.Context, memory *core.Memory, trace *core.MemoryTrace) error {
	if memory == nil {
		return fmt.Errorf("%w: memory is required", core.ErrInvalidArgument)
	}
	if trace == nil {
		return fmt.Errorf("%w: memory trace is required", core.ErrInvalidArgument)
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin memory trace transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := upsertMemory(ctx, tx, memory); err != nil {
		return err
	}
	if trace.MemoryID == "" {
		trace.MemoryID = memory.ID
	}
	if trace.MemoryID != memory.ID {
		return fmt.Errorf("%w: memory trace memory_id must match memory id", core.ErrInvalidArgument)
	}
	if err := writeMemoryTrace(ctx, tx, trace); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit memory trace transaction: %w", err)
	}
	return nil
}

// CreateMemoryWithTraceAndEdge writes a derived memory, provenance, and lineage edge in one transaction.
func (s *Store) CreateMemoryWithTraceAndEdge(ctx context.Context, memory *core.Memory, trace *core.MemoryTrace, edge *core.MemoryEdge) error {
	if memory == nil {
		return fmt.Errorf("%w: memory is required", core.ErrInvalidArgument)
	}
	if trace == nil {
		return fmt.Errorf("%w: memory trace is required", core.ErrInvalidArgument)
	}
	if edge == nil {
		return fmt.Errorf("%w: memory edge is required", core.ErrInvalidArgument)
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin memory trace edge transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := upsertMemory(ctx, tx, memory); err != nil {
		return err
	}
	if trace.MemoryID == "" {
		trace.MemoryID = memory.ID
	}
	if trace.MemoryID != memory.ID {
		return fmt.Errorf("%w: memory trace memory_id must match memory id", core.ErrInvalidArgument)
	}
	if err := writeMemoryTrace(ctx, tx, trace); err != nil {
		return err
	}
	if edge.FromMemoryID == "" {
		edge.FromMemoryID = memory.ID
	}
	if edge.FromMemoryID != memory.ID {
		return fmt.Errorf("%w: memory edge from_memory_id must match memory id", core.ErrInvalidArgument)
	}
	if err := upsertMemoryEdge(ctx, tx, edge); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit memory trace edge transaction: %w", err)
	}
	return nil
}

// CreateMemoryWithTraceAndUpdateEdge writes a new memory, provenance, updates edge, and target supersession in one transaction.
func (s *Store) CreateMemoryWithTraceAndUpdateEdge(ctx context.Context, memory *core.Memory, trace *core.MemoryTrace, edge *core.MemoryEdge) error {
	if memory == nil {
		return fmt.Errorf("%w: memory is required", core.ErrInvalidArgument)
	}
	if trace == nil {
		return fmt.Errorf("%w: memory trace is required", core.ErrInvalidArgument)
	}
	if edge == nil {
		return fmt.Errorf("%w: memory edge is required", core.ErrInvalidArgument)
	}
	if memory.ID == "" {
		return fmt.Errorf("%w: update memory id is required", core.ErrInvalidArgument)
	}
	if edge.EdgeKind != core.EdgeKindUpdates {
		return fmt.Errorf("%w: memory edge kind must be updates", core.ErrInvalidArgument)
	}
	if edge.ToMemoryID == "" {
		return fmt.Errorf("%w: update target memory id is required", core.ErrInvalidArgument)
	}
	if edge.ToMemoryID == memory.ID {
		return fmt.Errorf("%w: update memory edge cannot target itself", core.ErrInvalidArgument)
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin memory update transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if edge.FromMemoryID == "" {
		edge.FromMemoryID = memory.ID
	}
	if edge.FromMemoryID != memory.ID {
		return fmt.Errorf("%w: memory edge from_memory_id must match memory id", core.ErrInvalidArgument)
	}
	completed, err := completedUpdateAlreadyApplied(ctx, tx, memory, edge)
	if err != nil {
		return err
	}
	if completed {
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit idempotent memory update transaction: %w", err)
		}
		return nil
	}

	target, err := lockLatestMemoryTarget(ctx, tx, memory, edge.ToMemoryID)
	if err != nil {
		return err
	}
	if err := validateUpdateTarget(memory, target); err != nil {
		return err
	}
	if err := upsertMemory(ctx, tx, memory); err != nil {
		return err
	}
	if trace.MemoryID == "" {
		trace.MemoryID = memory.ID
	}
	if trace.MemoryID != memory.ID {
		return fmt.Errorf("%w: memory trace memory_id must match memory id", core.ErrInvalidArgument)
	}
	if err := writeMemoryTrace(ctx, tx, trace); err != nil {
		return err
	}
	if err := upsertMemoryEdge(ctx, tx, edge); err != nil {
		return err
	}
	if err := supersedeMemoryTarget(ctx, tx, edge.ToMemoryID, timeOrNow(memory.ValidFrom)); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit memory update transaction: %w", err)
	}
	return nil
}

// UpsertMemory writes a memory after apply validation.
func (s *Store) UpsertMemory(ctx context.Context, memory *core.Memory) error {
	return upsertMemory(ctx, s.pool, memory)
}

func completedUpdateAlreadyApplied(ctx context.Context, tx pgx.Tx, memory *core.Memory, edge *core.MemoryEdge) (bool, error) {
	var existing struct {
		TenantID      string
		WorkspaceID   string
		Scope         core.MemoryScope
		GroupID       *string
		OwnerEntityID string
		Status        core.MemoryStatus
		LatestFlag    bool
		HasTrace      bool
		HasEdge       bool
	}
	err := tx.QueryRow(ctx, `
		SELECT m.tenant_id, m.workspace_id, m.scope, m.group_id, m.owner_entity_id,
		       m.status, m.latest_flag,
		       EXISTS (
		           SELECT 1 FROM memory_trace mt
		           WHERE mt.memory_id = m.id
		       ) AS has_trace,
		       EXISTS (
		           SELECT 1 FROM memory_edges me
		           WHERE me.from_memory_id = m.id
		             AND me.to_memory_id = $2
		             AND me.edge_kind = $3
		       ) AS has_edge
		FROM memories m
		WHERE m.id = $1
		FOR UPDATE
	`, memory.ID, edge.ToMemoryID, core.EdgeKindUpdates).Scan(&existing.TenantID,
		&existing.WorkspaceID, &existing.Scope, &existing.GroupID, &existing.OwnerEntityID,
		&existing.Status, &existing.LatestFlag, &existing.HasTrace, &existing.HasEdge)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("check idempotent update memory: %w", err)
	}
	if existing.TenantID != memory.TenantID ||
		existing.WorkspaceID != memory.WorkspaceID ||
		existing.Scope != memory.Scope ||
		!sameOptionalString(existing.GroupID, memory.GroupID) ||
		existing.OwnerEntityID != memory.OwnerEntityID ||
		existing.Status != core.MemoryStatusActive ||
		!existing.LatestFlag ||
		!existing.HasTrace ||
		!existing.HasEdge {
		return false, fmt.Errorf("%w: existing update memory is incomplete or mismatched", core.ErrConflict)
	}
	return true, nil
}

type updateTargetMemory struct {
	Scope         core.MemoryScope
	GroupID       *string
	OwnerEntityID string
	Status        core.MemoryStatus
	LatestFlag    bool
}

func lockLatestMemoryTarget(ctx context.Context, tx pgx.Tx, memory *core.Memory, targetID string) (*updateTargetMemory, error) {
	target := &updateTargetMemory{}
	err := tx.QueryRow(ctx, `
		SELECT scope, group_id, owner_entity_id, status, latest_flag
		FROM memories
		WHERE id = $1 AND tenant_id = $2 AND workspace_id = $3
		FOR UPDATE
	`, targetID, memory.TenantID, memory.WorkspaceID).Scan(&target.Scope, &target.GroupID,
		&target.OwnerEntityID, &target.Status, &target.LatestFlag)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, core.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("lock update target memory: %w", err)
	}
	if target.Status != core.MemoryStatusActive || !target.LatestFlag {
		return nil, fmt.Errorf("%w: update target memory must be active latest", core.ErrConflict)
	}
	return target, nil
}

func validateUpdateTarget(memory *core.Memory, target *updateTargetMemory) error {
	if memory.Scope != target.Scope {
		return fmt.Errorf("%w: update memory scope must match target scope", core.ErrInvalidArgument)
	}
	if !sameOptionalString(memory.GroupID, target.GroupID) {
		return fmt.Errorf("%w: update memory group_id must match target group_id", core.ErrInvalidArgument)
	}
	if memory.OwnerEntityID != target.OwnerEntityID {
		return fmt.Errorf("%w: update memory owner_entity_id must match target owner_entity_id", core.ErrInvalidArgument)
	}
	return nil
}

func supersedeMemoryTarget(ctx context.Context, exec memoryExecutor, targetID string, supersededAt time.Time) error {
	tag, err := exec.Exec(ctx, `
		UPDATE memories
		SET status = $2,
		    latest_flag = false,
		    valid_to = $3,
		    updated_at = $3
		WHERE id = $1
		  AND status = $4
		  AND latest_flag = true
	`, targetID, core.MemoryStatusSuperseded, supersededAt.UTC(), core.MemoryStatusActive)
	if err != nil {
		return fmt.Errorf("supersede update target memory: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return core.ErrConflict
	}
	return nil
}

func sameOptionalString(left *string, right *string) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func upsertMemory(ctx context.Context, exec memoryExecutor, memory *core.Memory) error {
	if memory == nil {
		return fmt.Errorf("%w: memory is required", core.ErrInvalidArgument)
	}
	if memory.ID == "" {
		id, err := newID("mem")
		if err != nil {
			return err
		}
		memory.ID = id
	}
	_, err := exec.Exec(ctx, `
		INSERT INTO memories (
			id, tenant_id, workspace_id, scope, group_id, owner_entity_id,
			kind, artifact_class, text, fingerprint, confidence, status,
			valid_from, valid_to, latest_flag, metadata_json,
			embedding_model, embedding_dims, embedding_updated_at, created_at, updated_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21)
		ON CONFLICT (id) DO UPDATE
		SET scope = EXCLUDED.scope,
		    group_id = EXCLUDED.group_id,
		    owner_entity_id = EXCLUDED.owner_entity_id,
		    kind = EXCLUDED.kind,
		    artifact_class = EXCLUDED.artifact_class,
		    text = EXCLUDED.text,
		    fingerprint = EXCLUDED.fingerprint,
		    confidence = EXCLUDED.confidence,
		    status = EXCLUDED.status,
		    valid_from = EXCLUDED.valid_from,
		    valid_to = EXCLUDED.valid_to,
		    latest_flag = EXCLUDED.latest_flag,
		    metadata_json = EXCLUDED.metadata_json,
		    embedding_model = EXCLUDED.embedding_model,
		    embedding_dims = EXCLUDED.embedding_dims,
		    embedding_updated_at = EXCLUDED.embedding_updated_at,
		    updated_at = EXCLUDED.updated_at
	`, memory.ID, memory.TenantID, memory.WorkspaceID, memory.Scope, memory.GroupID,
		memory.OwnerEntityID, memory.Kind, memory.ArtifactClass, memory.Text, memory.Fingerprint,
		memory.Confidence, valueOr(string(memory.Status), string(core.MemoryStatusActive)),
		timeOrNow(memory.ValidFrom), memory.ValidTo, memory.LatestFlag, rawJSONOrEmpty(memory.MetadataJSON),
		valueOr(memory.EmbeddingModel, "pending"), memory.EmbeddingDims, memory.EmbeddingUpdatedAt,
		timeOrNow(memory.CreatedAt), timeOrNow(memory.UpdatedAt))
	if err != nil {
		return fmt.Errorf("upsert memory: %w", err)
	}
	return nil
}

// GetMemory loads one memory by ID.
func (s *Store) GetMemory(ctx context.Context, memoryID string) (*core.Memory, error) {
	memory := &core.Memory{}
	err := s.pool.QueryRow(ctx, `
		SELECT id, tenant_id, workspace_id, scope, group_id, owner_entity_id,
		       kind, artifact_class, text, fingerprint, confidence, status,
		       valid_from, valid_to, latest_flag, metadata_json,
		       embedding_model, embedding_dims, embedding_updated_at, created_at, updated_at
		FROM memories
		WHERE id = $1
	`, memoryID).Scan(&memory.ID, &memory.TenantID, &memory.WorkspaceID, &memory.Scope,
		&memory.GroupID, &memory.OwnerEntityID, &memory.Kind, &memory.ArtifactClass,
		&memory.Text, &memory.Fingerprint, &memory.Confidence, &memory.Status,
		&memory.ValidFrom, &memory.ValidTo, &memory.LatestFlag, &memory.MetadataJSON,
		&memory.EmbeddingModel, &memory.EmbeddingDims, &memory.EmbeddingUpdatedAt,
		&memory.CreatedAt, &memory.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, core.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get memory: %w", err)
	}
	return memory, nil
}

// UpsertMemoryEdge writes a graph edge.
func (s *Store) UpsertMemoryEdge(ctx context.Context, edge *core.MemoryEdge) error {
	return upsertMemoryEdge(ctx, s.pool, edge)
}

func upsertMemoryEdge(ctx context.Context, exec memoryExecutor, edge *core.MemoryEdge) error {
	if edge == nil {
		return fmt.Errorf("%w: memory edge is required", core.ErrInvalidArgument)
	}
	_, err := exec.Exec(ctx, `
		INSERT INTO memory_edges (
			from_memory_id, to_memory_id, edge_kind, confidence, created_by_job_id, created_at
		)
		VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (from_memory_id, to_memory_id, edge_kind) DO UPDATE
		SET confidence = EXCLUDED.confidence,
		    created_by_job_id = EXCLUDED.created_by_job_id,
		    created_at = EXCLUDED.created_at
	`, edge.FromMemoryID, edge.ToMemoryID, edge.EdgeKind, edge.Confidence,
		edge.CreatedByJobID, timeOrNow(edge.CreatedAt))
	if err != nil {
		return fmt.Errorf("upsert memory edge: %w", err)
	}
	return nil
}

// WriteMemoryTrace writes mandatory provenance for a memory.
func (s *Store) WriteMemoryTrace(ctx context.Context, trace *core.MemoryTrace) error {
	return writeMemoryTrace(ctx, s.pool, trace)
}

func writeMemoryTrace(ctx context.Context, exec memoryExecutor, trace *core.MemoryTrace) error {
	if trace == nil {
		return fmt.Errorf("%w: memory trace is required", core.ErrInvalidArgument)
	}
	_, err := exec.Exec(ctx, `
		INSERT INTO memory_trace (
			memory_id, raw_event_ids, reasoning_job_id, reasoning_stage,
			candidate_snapshot_json, applied_operations_json,
			operator_correction_flag, related_document_ids, created_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (memory_id) DO UPDATE
		SET raw_event_ids = EXCLUDED.raw_event_ids,
		    reasoning_job_id = EXCLUDED.reasoning_job_id,
		    reasoning_stage = EXCLUDED.reasoning_stage,
		    candidate_snapshot_json = EXCLUDED.candidate_snapshot_json,
		    applied_operations_json = EXCLUDED.applied_operations_json,
		    operator_correction_flag = EXCLUDED.operator_correction_flag,
		    related_document_ids = EXCLUDED.related_document_ids,
		    created_at = EXCLUDED.created_at
	`, trace.MemoryID, trace.RawEventIDs, nullIfEmpty(trace.ReasoningJobID), trace.ReasoningStage,
		rawJSONOrEmpty(trace.CandidateSnapshotJSON), rawJSONOrEmpty(trace.AppliedOperationsJSON),
		trace.OperatorCorrectionFlag, trace.RelatedDocumentIDs, timeOrNow(trace.CreatedAt))
	if err != nil {
		return fmt.Errorf("write memory trace: %w", err)
	}
	return nil
}

// ExplainMemory loads provenance for one memory.
func (s *Store) ExplainMemory(ctx context.Context, req *core.ExplainMemoryRequest) (*core.ExplainMemoryResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("%w: explain memory request is required", core.ErrInvalidArgument)
	}
	resp := &core.ExplainMemoryResponse{MemoryID: req.MemoryID}
	err := s.pool.QueryRow(ctx, explainMemoryTraceStatement(), req.MemoryID, req.TenantID, req.WorkspaceID, req.EntityID, req.VisibleGroupIDs).Scan(
		&resp.Trace.RawEventIDs, &resp.Trace.ReasoningJobID,
		&resp.Trace.ReasoningStage, &resp.Trace.CandidateSnapshotJSON,
		&resp.Trace.AppliedOperationsJSON, &resp.Trace.OperatorCorrectionFlag,
		&resp.Trace.RelatedDocumentIDs, &resp.Trace.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, core.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query memory trace: %w", err)
	}
	edges, err := s.memoryEdges(ctx, req.MemoryID)
	if err != nil {
		return nil, err
	}
	resp.Edges = edges

	events, err := s.provenanceEvents(ctx, req.TenantID, req.WorkspaceID, resp.Trace.RawEventIDs)
	if err != nil {
		return nil, err
	}
	resp.SourceEvents = events

	documents, err := s.provenanceDocuments(ctx, req.TenantID, req.WorkspaceID, resp.Trace.RelatedDocumentIDs)
	if err != nil {
		return nil, err
	}
	resp.Documents = documents
	return resp, nil
}

func explainMemoryTraceStatement() string {
	return `
		SELECT raw_event_ids, COALESCE(reasoning_job_id, ''), reasoning_stage,
		       candidate_snapshot_json, applied_operations_json,
		       operator_correction_flag, related_document_ids, mt.created_at
		FROM memory_trace mt
		JOIN memories m ON m.id = mt.memory_id
		WHERE mt.memory_id = $1
		  AND m.tenant_id = $2
		  AND m.workspace_id = $3
		  AND (
		    m.scope <> 'agent_private'
		    OR ($4 <> '' AND m.owner_entity_id = $4)
		  )
		  AND (
		    m.scope <> 'group_shared'
		    OR (m.group_id IS NOT NULL AND m.group_id = ANY($5))
		  )
	`
}

func (s *Store) memoryEdges(ctx context.Context, memoryID string) ([]core.MemoryEdgeResult, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT from_memory_id, to_memory_id, edge_kind, confidence, created_at
		FROM memory_edges
		WHERE from_memory_id = $1 OR to_memory_id = $1
		ORDER BY created_at DESC
	`, memoryID)
	if err != nil {
		return nil, fmt.Errorf("query memory edges: %w", err)
	}
	defer rows.Close()

	edges := make([]core.MemoryEdgeResult, 0)
	for rows.Next() {
		var edge core.MemoryEdgeResult
		if err := rows.Scan(&edge.FromMemoryID, &edge.ToMemoryID, &edge.EdgeKind,
			&edge.Confidence, &edge.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan memory edge: %w", err)
		}
		edges = append(edges, edge)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate memory edges: %w", err)
	}
	return edges, nil
}

func (s *Store) provenanceEvents(ctx context.Context, tenantID, workspaceID string, ids []string) ([]core.ProvenanceEventResult, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id, event_kind, source, fingerprint, occurred_at, payload_json
		FROM raw_events
		WHERE tenant_id = $1
		  AND workspace_id = $2
		  AND id = ANY($3)
		ORDER BY occurred_at ASC
	`, tenantID, workspaceID, ids)
	if err != nil {
		return nil, fmt.Errorf("query provenance events: %w", err)
	}
	defer rows.Close()

	events := make([]core.ProvenanceEventResult, 0, len(ids))
	for rows.Next() {
		var event core.ProvenanceEventResult
		if err := rows.Scan(&event.EventID, &event.EventKind, &event.Source,
			&event.Fingerprint, &event.OccurredAt, &event.PayloadJSON); err != nil {
			return nil, fmt.Errorf("scan provenance event: %w", err)
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate provenance events: %w", err)
	}
	return events, nil
}

func (s *Store) provenanceDocuments(ctx context.Context, tenantID, workspaceID string, ids []string) ([]core.ProvenanceDocumentLink, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id, title
		FROM documents
		WHERE tenant_id = $1
		  AND workspace_id = $2
		  AND id = ANY($3)
		ORDER BY updated_at DESC
	`, tenantID, workspaceID, ids)
	if err != nil {
		return nil, fmt.Errorf("query provenance documents: %w", err)
	}
	defer rows.Close()

	documents := make([]core.ProvenanceDocumentLink, 0, len(ids))
	for rows.Next() {
		var document core.ProvenanceDocumentLink
		if err := rows.Scan(&document.DocumentID, &document.Title); err != nil {
			return nil, fmt.Errorf("scan provenance document: %w", err)
		}
		documents = append(documents, document)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate provenance documents: %w", err)
	}
	return documents, nil
}

func nullIfEmpty(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

```



<!-- Source: internal/store/postgres/memories_test.go | bytes=6018 | lines=207 | sha16=0dbecbace6b9a184 -->

```go
// ============================================================
// FILE     : internal/store/postgres/memories_test.go
// PURPOSE  : Verifies PostgreSQL memory graph helper contracts without a live database.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : postgres memory helper tests
// DEPENDS  : context, errors, strings, testing, time, internal/core, github.com/jackc/pgx/v5/pgconn
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: These tests lock update_memory supersession guards before live DB integration tests exist.
// ============================================================

package postgres

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

func TestValidateUpdateTargetRequiresSameScopeGroupAndOwner(t *testing.T) {
	t.Parallel()

	groupID := "group_1"
	tests := []struct {
		name   string
		memory *core.Memory
		target *updateTargetMemory
	}{
		{
			name: "scope mismatch",
			memory: &core.Memory{
				Scope:         core.MemoryScopeWorkspaceShared,
				OwnerEntityID: "agent:hermes-main",
			},
			target: &updateTargetMemory{
				Scope:         core.MemoryScopeAgentPrivate,
				OwnerEntityID: "agent:hermes-main",
			},
		},
		{
			name: "group mismatch",
			memory: &core.Memory{
				Scope:         core.MemoryScopeGroupShared,
				GroupID:       &groupID,
				OwnerEntityID: "agent:hermes-main",
			},
			target: &updateTargetMemory{
				Scope:         core.MemoryScopeGroupShared,
				OwnerEntityID: "agent:hermes-main",
			},
		},
		{
			name: "owner mismatch",
			memory: &core.Memory{
				Scope:         core.MemoryScopeAgentPrivate,
				OwnerEntityID: "agent:hermes-main",
			},
			target: &updateTargetMemory{
				Scope:         core.MemoryScopeAgentPrivate,
				OwnerEntityID: "agent:other",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateUpdateTarget(tt.memory, tt.target)
			if !errors.Is(err, core.ErrInvalidArgument) {
				t.Fatalf("expected ErrInvalidArgument, got %v", err)
			}
		})
	}
}

func TestValidateUpdateTargetAllowsSameScopeGroupAndOwner(t *testing.T) {
	t.Parallel()

	groupID := "group_1"
	memory := &core.Memory{
		Scope:         core.MemoryScopeGroupShared,
		GroupID:       &groupID,
		OwnerEntityID: "agent:hermes-main",
	}
	target := &updateTargetMemory{
		Scope:         core.MemoryScopeGroupShared,
		GroupID:       &groupID,
		OwnerEntityID: "agent:hermes-main",
	}

	if err := validateUpdateTarget(memory, target); err != nil {
		t.Fatalf("validateUpdateTarget returned error: %v", err)
	}
}

func TestSupersedeMemoryTargetOnlyUpdatesActiveLatestRows(t *testing.T) {
	t.Parallel()

	exec := &recordingMemoryExecutor{tag: pgconn.NewCommandTag("UPDATE 1")}
	supersededAt := time.Date(2026, time.April, 24, 13, 30, 0, 0, time.UTC)

	if err := supersedeMemoryTarget(context.Background(), exec, "mem_old", supersededAt); err != nil {
		t.Fatalf("supersedeMemoryTarget returned error: %v", err)
	}

	if !strings.Contains(exec.sql, "status = $2") || !strings.Contains(exec.sql, "latest_flag = false") {
		t.Fatalf("expected supersession update fields, got: %s", exec.sql)
	}
	if !strings.Contains(exec.sql, "AND status = $4") || !strings.Contains(exec.sql, "AND latest_flag = true") {
		t.Fatalf("expected active/latest guard, got: %s", exec.sql)
	}
	if len(exec.args) != 4 {
		t.Fatalf("unexpected supersede args: %#v", exec.args)
	}
	if exec.args[0] != "mem_old" || exec.args[1] != core.MemoryStatusSuperseded || exec.args[3] != core.MemoryStatusActive {
		t.Fatalf("unexpected supersede args: %#v", exec.args)
	}
	if got := exec.args[2].(time.Time); !got.Equal(supersededAt) {
		t.Fatalf("unexpected superseded timestamp: got %s want %s", got, supersededAt)
	}
}

func TestSupersedeMemoryTargetReturnsConflictWhenNoLatestRowChanged(t *testing.T) {
	t.Parallel()

	exec := &recordingMemoryExecutor{tag: pgconn.NewCommandTag("UPDATE 0")}

	err := supersedeMemoryTarget(context.Background(), exec, "mem_old", time.Now().UTC())
	if !errors.Is(err, core.ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

func TestExplainMemoryTraceStatementScopesMemoryToTenantWorkspace(t *testing.T) {
	t.Parallel()

	sql := explainMemoryTraceStatement()

	for _, want := range []string{
		"FROM memory_trace mt",
		"JOIN memories m ON m.id = mt.memory_id",
		"mt.memory_id = $1",
		"m.tenant_id = $2",
		"m.workspace_id = $3",
		"m.scope <> 'agent_private'",
		"m.owner_entity_id = $4",
		"m.scope <> 'group_shared'",
		"m.group_id = ANY($5)",
	} {
		if !strings.Contains(sql, want) {
			t.Fatalf("explain memory trace query must preserve %q, got:\n%s", want, sql)
		}
	}
}

func TestExplainMemoryProvenanceQueriesScopeEvidenceToTenantWorkspace(t *testing.T) {
	t.Parallel()

	source := readPostgresSourceFile(t, "memories.go")
	eventSource := extractPostgresSourceBetween(t, source, "func (s *Store) provenanceEvents", "func (s *Store) provenanceDocuments")
	documentSource := extractPostgresSourceBetween(t, source, "func (s *Store) provenanceDocuments", "func nullIfEmpty")

	for _, tt := range []struct {
		name   string
		source string
	}{
		{name: "events", source: eventSource},
		{name: "documents", source: documentSource},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			for _, want := range []string{
				"tenant_id = $1",
				"workspace_id = $2",
				"id = ANY($3)",
			} {
				if !strings.Contains(tt.source, want) {
					t.Fatalf("provenance %s query must preserve %q, got:\n%s", tt.name, want, tt.source)
				}
			}
		})
	}
}

type recordingMemoryExecutor struct {
	sql  string
	args []any
	tag  pgconn.CommandTag
	err  error
}

func (e *recordingMemoryExecutor) Exec(_ context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	e.sql = sql
	e.args = append([]any(nil), arguments...)
	return e.tag, e.err
}

```



<!-- Source: internal/store/postgres/notes_plans.go | bytes=7979 | lines=253 | sha16=94abac90df49db52 -->

```go
// ============================================================
// FILE     : internal/store/postgres/notes_plans.go
// PURPOSE  : Implements PostgreSQL persistence for notes and structured plans.
// LAYER    : infra
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : AddNote, ListPinnedNotes, CreatePlan, UpdatePlan, GetActivePlans
// DEPENDS  : internal/core
// USED_BY  : recall assembler, future notes and plans APIs
// ------------------------------------------------------------
// AGENT_NOTE: Notes and plans are operator-intent artifacts; keep them separate from memories.
// ============================================================

package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

// AddNote writes a note.
func (s *Store) AddNote(ctx context.Context, note *core.Note) error {
	if note == nil {
		return fmt.Errorf("%w: note is required", core.ErrInvalidArgument)
	}
	id := note.ID
	var err error
	if id == "" {
		id, err = newID("note")
		if err != nil {
			return err
		}
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO notes (
			id, tenant_id, workspace_id, note_kind, scope, owner_entity_id,
			text, pinned, expires_at, created_at, updated_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
	`, id, note.TenantID, note.WorkspaceID, note.NoteKind, note.Scope, note.OwnerEntityID,
		note.Text, note.Pinned, note.ExpiresAt, timeOrNow(note.CreatedAt), timeOrNow(note.UpdatedAt))
	if err != nil {
		return fmt.Errorf("insert note: %w", err)
	}
	note.ID = id
	return nil
}

// ListPinnedNotes loads pinned notes for recall.
func (s *Store) ListPinnedNotes(ctx context.Context, req *core.ListPinnedNotesRequest) ([]*core.Note, error) {
	if req == nil {
		return nil, fmt.Errorf("%w: list pinned notes request is required", core.ErrInvalidArgument)
	}
	sql, args := listPinnedNotesStatement(req)
	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query pinned notes: %w", err)
	}
	defer rows.Close()

	notes := make([]*core.Note, 0, 20)
	for rows.Next() {
		note := &core.Note{}
		if err := rows.Scan(&note.ID, &note.TenantID, &note.WorkspaceID, &note.NoteKind,
			&note.Scope, &note.OwnerEntityID, &note.Text, &note.Pinned, &note.ExpiresAt,
			&note.CreatedAt, &note.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan note: %w", err)
		}
		notes = append(notes, note)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate pinned notes: %w", err)
	}
	return notes, nil
}

func listPinnedNotesStatement(req *core.ListPinnedNotesRequest) (string, []any) {
	return `
		SELECT id, tenant_id, workspace_id, note_kind, scope, owner_entity_id,
		       text, pinned, expires_at, created_at, updated_at
		FROM notes
		WHERE tenant_id = $1
		  AND workspace_id = $2
		  AND scope = ANY($3)
		  AND pinned = true
		  AND (expires_at IS NULL OR expires_at > now())
		  AND (
		    scope <> 'agent_private'
		    OR ($4 <> '' AND owner_entity_id = $4)
		  )
		ORDER BY updated_at DESC, created_at DESC
		LIMIT 20
	`, []any{req.TenantID, req.WorkspaceID, req.Scopes, req.OwnerEntityID}
}

// CreatePlan writes a plan and its initial items.
func (s *Store) CreatePlan(ctx context.Context, plan *core.Plan, items []*core.PlanItem) error {
	if plan == nil {
		return fmt.Errorf("%w: plan is required", core.ErrInvalidArgument)
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin create plan: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if plan.ID == "" {
		plan.ID, err = newID("plan")
		if err != nil {
			return err
		}
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO plans (
			id, tenant_id, workspace_id, title, status, scope,
			owner_entity_id, evidence_json, created_at, updated_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
	`, plan.ID, plan.TenantID, plan.WorkspaceID, plan.Title, plan.Status, plan.Scope,
		plan.OwnerEntityID, rawJSONOrEmpty(plan.EvidenceJSON), timeOrNow(plan.CreatedAt), timeOrNow(plan.UpdatedAt))
	if err != nil {
		return fmt.Errorf("insert plan: %w", err)
	}
	if err := insertPlanItems(ctx, tx, plan.ID, items); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit create plan: %w", err)
	}
	return nil
}

// UpdatePlan updates a plan and replaces its provided items when items is non-nil.
func (s *Store) UpdatePlan(ctx context.Context, plan *core.Plan, items []*core.PlanItem) error {
	if plan == nil || plan.ID == "" {
		return fmt.Errorf("%w: plan id is required", core.ErrInvalidArgument)
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin update plan: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	tag, err := tx.Exec(ctx, `
		UPDATE plans
		SET title = COALESCE(NULLIF($2, ''), title),
		    status = COALESCE(NULLIF($3, ''), status),
		    evidence_json = COALESCE($4, evidence_json),
		    updated_at = now()
		WHERE id = $1
		  AND tenant_id = $5
		  AND workspace_id = $6
	`, plan.ID, plan.Title, plan.Status, rawJSONOrNil(plan.EvidenceJSON), plan.TenantID, plan.WorkspaceID)
	if err != nil {
		return fmt.Errorf("update plan: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return core.ErrNotFound
	}
	if items != nil {
		if _, err := tx.Exec(ctx, `DELETE FROM plan_items WHERE plan_id = $1`, plan.ID); err != nil {
			return fmt.Errorf("delete plan items: %w", err)
		}
		if err := insertPlanItems(ctx, tx, plan.ID, items); err != nil {
			return err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit update plan: %w", err)
	}
	return nil
}

// GetActivePlans loads active plans for recall.
func (s *Store) GetActivePlans(ctx context.Context, req *core.GetActivePlansRequest) ([]*core.Plan, error) {
	if req == nil {
		return nil, fmt.Errorf("%w: get active plans request is required", core.ErrInvalidArgument)
	}
	sql, args := getActivePlansStatement(req)
	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query active plans: %w", err)
	}
	defer rows.Close()

	plans := make([]*core.Plan, 0, 20)
	for rows.Next() {
		plan := &core.Plan{}
		if err := rows.Scan(&plan.ID, &plan.TenantID, &plan.WorkspaceID, &plan.Title,
			&plan.Status, &plan.Scope, &plan.OwnerEntityID, &plan.EvidenceJSON,
			&plan.CreatedAt, &plan.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan plan: %w", err)
		}
		plans = append(plans, plan)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate active plans: %w", err)
	}
	return plans, nil
}

func getActivePlansStatement(req *core.GetActivePlansRequest) (string, []any) {
	return `
		SELECT id, tenant_id, workspace_id, title, status, scope,
		       owner_entity_id, evidence_json, created_at, updated_at
		FROM plans
		WHERE tenant_id = $1
		  AND workspace_id = $2
		  AND scope = ANY($3)
		  AND lower(status) NOT IN ('done', 'completed', 'archived', 'deleted', 'cancelled')
		  AND (
		    scope <> 'agent_private'
		    OR ($4 <> '' AND owner_entity_id = $4)
		  )
		ORDER BY updated_at DESC, created_at DESC
		LIMIT 20
	`, []any{req.TenantID, req.WorkspaceID, req.Scopes, req.OwnerEntityID}
}

type planItemExecutor interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

func insertPlanItems(ctx context.Context, tx planItemExecutor, planID string, items []*core.PlanItem) error {
	for _, item := range items {
		if item == nil {
			continue
		}
		var err error
		if item.ID == "" {
			item.ID, err = newID("plan_item")
			if err != nil {
				return err
			}
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO plan_items (
				id, plan_id, title, status, evidence_json, created_at, updated_at
			)
			VALUES ($1,$2,$3,$4,$5,$6,$7)
		`, item.ID, planID, item.Title, item.Status, rawJSONOrEmpty(item.EvidenceJSON),
			timeOrNow(item.CreatedAt), timeOrNow(item.UpdatedAt))
		if err != nil {
			return fmt.Errorf("insert plan item: %w", err)
		}
	}
	return nil
}

```



<!-- Source: internal/store/postgres/notes_plans_test.go | bytes=3881 | lines=122 | sha16=947b68124da0336d -->

```go
// ============================================================
// FILE     : internal/store/postgres/notes_plans_test.go
// PURPOSE  : Verifies note and plan lookup SQL preserves actor-private visibility boundaries.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : postgres note and plan statement tests
// DEPENDS  : strings, testing, internal/core
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: Keep pinned note and active plan retrieval tenant- and actor-scoped.
// ============================================================

package postgres

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

func TestListPinnedNotesStatementScopesAgentPrivateToOwner(t *testing.T) {
	t.Parallel()

	sql, args := listPinnedNotesStatement(&core.ListPinnedNotesRequest{
		TenantID:      "tenant_1",
		WorkspaceID:   "workspace_1",
		OwnerEntityID: "agent:hermes-main",
		Scopes:        []core.MemoryScope{core.MemoryScopeAgentPrivate, core.MemoryScopeWorkspaceShared},
	})

	if !strings.Contains(sql, "tenant_id = $1") {
		t.Fatalf("expected tenant predicate for pinned notes, got:\n%s", sql)
	}
	if !strings.Contains(sql, "scope <> 'agent_private'") {
		t.Fatalf("expected non-private note scopes to bypass owner filter, got:\n%s", sql)
	}
	if !strings.Contains(sql, "owner_entity_id = $4") {
		t.Fatalf("expected agent_private note owner predicate, got:\n%s", sql)
	}
	if len(args) != 4 || args[3] != "agent:hermes-main" {
		t.Fatalf("unexpected pinned notes args: %#v", args)
	}
}

func readCurrentFile(t *testing.T, name string) string {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not locate current test file")
	}
	data, err := os.ReadFile(filepath.Join(filepath.Dir(currentFile), name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return string(data)
}

func extractSourceBetween(t *testing.T, source, startMarker, endMarker string) string {
	t.Helper()

	start := strings.Index(source, startMarker)
	if start < 0 {
		t.Fatalf("missing start marker %q", startMarker)
	}
	remainder := source[start:]
	end := strings.Index(remainder, endMarker)
	if end < 0 {
		t.Fatalf("missing end marker %q", endMarker)
	}
	return remainder[:end]
}

func TestGetActivePlansStatementScopesAgentPrivateToOwner(t *testing.T) {
	t.Parallel()

	sql, args := getActivePlansStatement(&core.GetActivePlansRequest{
		TenantID:      "tenant_1",
		WorkspaceID:   "workspace_1",
		OwnerEntityID: "agent:hermes-main",
		Scopes:        []core.MemoryScope{core.MemoryScopeAgentPrivate, core.MemoryScopeWorkspaceShared},
	})

	if !strings.Contains(sql, "tenant_id = $1") {
		t.Fatalf("expected tenant predicate for active plans, got:\n%s", sql)
	}
	if !strings.Contains(sql, "scope <> 'agent_private'") {
		t.Fatalf("expected non-private plan scopes to bypass owner filter, got:\n%s", sql)
	}
	if !strings.Contains(sql, "owner_entity_id = $4") {
		t.Fatalf("expected agent_private plan owner predicate, got:\n%s", sql)
	}
	if len(args) != 4 || args[3] != "agent:hermes-main" {
		t.Fatalf("unexpected active plans args: %#v", args)
	}
}

func TestUpdatePlanSourceUsesTenantWorkspaceAndPatchSemantics(t *testing.T) {
	t.Parallel()

	source := readCurrentFile(t, "notes_plans.go")
	updateSource := extractSourceBetween(t, source, "func (s *Store) UpdatePlan", "// GetActivePlans loads active plans for recall.")

	for _, want := range []string{
		"COALESCE(NULLIF($2, ''), title)",
		"COALESCE(NULLIF($3, ''), status)",
		"COALESCE($4, evidence_json)",
		"AND tenant_id = $5",
		"AND workspace_id = $6",
		"if items != nil",
	} {
		if !strings.Contains(updateSource, want) {
			t.Fatalf("UpdatePlan must preserve %q, got:\n%s", want, updateSource)
		}
	}
}

```



<!-- Source: internal/store/postgres/profiles_summaries.go | bytes=4418 | lines=120 | sha16=6f4646bf46791812 -->

```go
// ============================================================
// FILE     : internal/store/postgres/profiles_summaries.go
// PURPOSE  : Implements PostgreSQL persistence for profiles and session summaries.
// LAYER    : infra
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : GetProfile, UpsertProfile, UpsertSessionSummary, GetSessionSummary
// DEPENDS  : internal/core, github.com/jackc/pgx/v5
// USED_BY  : recall assembler, graph and dreaming slices
// ------------------------------------------------------------
// AGENT_NOTE: Profiles and summaries must remain rebuildable from source artifacts.
// ============================================================

package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

// GetProfile loads a profile snapshot.
func (s *Store) GetProfile(ctx context.Context, entityID string, scope core.MemoryScope) (*core.Profile, error) {
	profile := &core.Profile{}
	err := s.pool.QueryRow(ctx, `
		SELECT entity_id, scope, static_json, dynamic_json, source_memory_ids, updated_at, version
		FROM profiles
		WHERE entity_id = $1 AND scope = $2
	`, entityID, scope).Scan(&profile.EntityID, &profile.Scope, &profile.StaticJSON,
		&profile.DynamicJSON, &profile.SourceMemoryIDs, &profile.UpdatedAt, &profile.Version)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, core.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get profile: %w", err)
	}
	return profile, nil
}

// UpsertProfile writes a profile snapshot.
func (s *Store) UpsertProfile(ctx context.Context, profile *core.Profile) error {
	if profile == nil {
		return fmt.Errorf("%w: profile is required", core.ErrInvalidArgument)
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO profiles (
			entity_id, scope, static_json, dynamic_json, source_memory_ids, updated_at, version
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (entity_id, scope) DO UPDATE
		SET static_json = EXCLUDED.static_json,
		    dynamic_json = EXCLUDED.dynamic_json,
		    source_memory_ids = EXCLUDED.source_memory_ids,
		    updated_at = EXCLUDED.updated_at,
		    version = profiles.version + 1
	`, profile.EntityID, profile.Scope, rawJSONOrEmpty(profile.StaticJSON), rawJSONOrEmpty(profile.DynamicJSON),
		profile.SourceMemoryIDs, timeOrNow(profile.UpdatedAt), profile.Version)
	if err != nil {
		return fmt.Errorf("upsert profile: %w", err)
	}
	return nil
}

// UpsertSessionSummary writes a session summary.
func (s *Store) UpsertSessionSummary(ctx context.Context, summary *core.SessionSummary) error {
	if summary == nil {
		return fmt.Errorf("%w: session summary is required", core.ErrInvalidArgument)
	}
	var err error
	if summary.ID == "" {
		summary.ID, err = newID("sum")
		if err != nil {
			return err
		}
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO session_summaries (
			id, tenant_id, workspace_id, session_id, summary_text,
			source_event_ids, source_memory_ids, created_at, updated_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (id) DO UPDATE
		SET summary_text = EXCLUDED.summary_text,
		    source_event_ids = EXCLUDED.source_event_ids,
		    source_memory_ids = EXCLUDED.source_memory_ids,
		    updated_at = EXCLUDED.updated_at
	`, summary.ID, summary.TenantID, summary.WorkspaceID, summary.SessionID, summary.SummaryText,
		summary.SourceEventIDs, summary.SourceMemoryIDs, timeOrNow(summary.CreatedAt), timeOrNow(summary.UpdatedAt))
	if err != nil {
		return fmt.Errorf("upsert session summary: %w", err)
	}
	return nil
}

// GetSessionSummary loads the current summary for a session.
func (s *Store) GetSessionSummary(ctx context.Context, sessionID string) (*core.SessionSummary, error) {
	summary := &core.SessionSummary{}
	err := s.pool.QueryRow(ctx, `
		SELECT id, tenant_id, workspace_id, session_id, summary_text,
		       source_event_ids, source_memory_ids, created_at, updated_at
		FROM session_summaries
		WHERE session_id = $1
		ORDER BY updated_at DESC
		LIMIT 1
	`, sessionID).Scan(&summary.ID, &summary.TenantID, &summary.WorkspaceID,
		&summary.SessionID, &summary.SummaryText, &summary.SourceEventIDs,
		&summary.SourceMemoryIDs, &summary.CreatedAt, &summary.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, core.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get session summary: %w", err)
	}
	return summary, nil
}

```



<!-- Source: internal/store/postgres/raw_events.go | bytes=3361 | lines=101 | sha16=9b50ae7943a247da -->

```go
// ============================================================
// FILE     : internal/store/postgres/raw_events.go
// PURPOSE  : Implements PostgreSQL persistence for immutable raw events.
// LAYER    : infra
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : AppendRawEvents, GetRawEvents
// DEPENDS  : internal/core, github.com/jackc/pgx/v5
// USED_BY  : internal/ingest, worker pipeline
// ------------------------------------------------------------
// AGENT_NOTE: Preserve raw_events as immutable source records and use idempotency for duplicates.
// ============================================================

package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

// AppendRawEvents inserts raw events idempotently and returns newly inserted IDs.
func (s *Store) AppendRawEvents(ctx context.Context, events []*core.RawEvent) ([]string, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin raw event insert: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	ids := make([]string, 0, len(events))
	for _, event := range events {
		if event == nil {
			continue
		}
		id := event.ID
		if id == "" {
			id, err = newID("evt")
			if err != nil {
				return nil, err
			}
		}
		row := tx.QueryRow(ctx, `
			INSERT INTO raw_events (
				id, tenant_id, workspace_id, session_id, actor_id, event_kind,
				source, idempotency_key, fingerprint, occurred_at, payload_json, created_at
			)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
			ON CONFLICT (tenant_id, source, idempotency_key) DO NOTHING
			RETURNING id
		`, id, event.TenantID, event.WorkspaceID, event.SessionID, event.ActorID,
			event.EventKind, event.Source, event.IdempotencyKey, event.Fingerprint,
			timeOrNow(event.OccurredAt), rawJSONOrEmpty(event.PayloadJSON), timeOrNow(event.CreatedAt))
		var insertedID string
		if err := row.Scan(&insertedID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				continue
			}
			return nil, fmt.Errorf("insert raw event: %w", err)
		}
		ids = append(ids, insertedID)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit raw event insert: %w", err)
	}
	return ids, nil
}

// GetRawEvents loads raw events by ID.
func (s *Store) GetRawEvents(ctx context.Context, ids []string) ([]*core.RawEvent, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, tenant_id, workspace_id, session_id, actor_id, event_kind,
		       source, idempotency_key, fingerprint, occurred_at, payload_json, created_at
		FROM raw_events
		WHERE id = ANY($1)
		ORDER BY occurred_at ASC, created_at ASC
	`, ids)
	if err != nil {
		return nil, fmt.Errorf("query raw events: %w", err)
	}
	defer rows.Close()

	events := make([]*core.RawEvent, 0, len(ids))
	for rows.Next() {
		event := &core.RawEvent{}
		if err := rows.Scan(&event.ID, &event.TenantID, &event.WorkspaceID, &event.SessionID,
			&event.ActorID, &event.EventKind, &event.Source, &event.IdempotencyKey,
			&event.Fingerprint, &event.OccurredAt, &event.PayloadJSON, &event.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan raw event: %w", err)
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate raw events: %w", err)
	}
	return events, nil
}

```



<!-- Source: internal/store/postgres/search.go | bytes=3914 | lines=107 | sha16=5d3a404273b22cfe -->

```go
// ============================================================
// FILE     : internal/store/postgres/search.go
// PURPOSE  : Implements minimal lexical memory and document search for degraded recall.
// LAYER    : infra
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : SearchMemories, SearchDocuments
// DEPENDS  : internal/core
// USED_BY  : recall assembler, future search APIs
// ------------------------------------------------------------
// AGENT_NOTE: Keep this as lexical fallback; vector ranking belongs behind explicit embedding paths.
// ============================================================

package postgres

import (
	"context"
	"fmt"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

// SearchMemories searches active latest memories with lexical fallback.
func (s *Store) SearchMemories(ctx context.Context, req *core.SearchMemoriesRequest) (*core.SearchMemoriesResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("%w: search memories request is required", core.ErrInvalidArgument)
	}
	sql, args := searchMemoriesStatement(req)
	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query memories: %w", err)
	}
	defer rows.Close()

	results := make([]core.MemoryResult, 0, 20)
	for rows.Next() {
		var result core.MemoryResult
		if err := rows.Scan(&result.MemoryID, &result.Kind, &result.ArtifactClass, &result.Text,
			&result.Confidence, &result.Scope, &result.GroupID, &result.OwnerEntityID, &result.ValidFrom, &result.LatestFlag); err != nil {
			return nil, fmt.Errorf("scan memory result: %w", err)
		}
		results = append(results, result)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate memory results: %w", err)
	}
	return &core.SearchMemoriesResponse{Memories: results}, nil
}

func searchMemoriesStatement(req *core.SearchMemoriesRequest) (string, []any) {
	return `
		SELECT id, kind, artifact_class, text, confidence, scope, group_id, owner_entity_id, valid_from, latest_flag
		FROM memories
		WHERE tenant_id = $1
		  AND workspace_id = $2
		  AND scope = ANY($3)
		  AND artifact_class = ANY($4)
		  AND status = 'active'
		  AND latest_flag = true
		  AND ($5 = '' OR text ILIKE '%' || $5 || '%')
		  AND (
		    scope <> 'agent_private'
		    OR ($6 <> '' AND owner_entity_id = $6)
		  )
		  AND (
		    scope <> 'group_shared'
		    OR ($7::text[] IS NOT NULL AND group_id = ANY($7))
		  )
		ORDER BY valid_from DESC, updated_at DESC
		LIMIT 20
	`, []any{req.TenantID, req.WorkspaceID, req.Scopes, req.ArtifactClasses, req.Query, req.OwnerEntityID, req.VisibleGroupIDs}
}

// SearchDocuments searches document chunks with lexical fallback.
func (s *Store) SearchDocuments(ctx context.Context, req *core.SearchDocumentsRequest) (*core.SearchDocumentsResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("%w: search documents request is required", core.ErrInvalidArgument)
	}
	rows, err := s.pool.Query(ctx, `
		SELECT c.id, c.document_id, c.text, 1.0::double precision AS score
		FROM document_chunks c
		JOIN documents d ON d.id = c.document_id
		WHERE d.tenant_id = $1
		  AND d.workspace_id = $2
		  AND ($3 = '' OR c.text ILIKE '%' || $3 || '%')
		ORDER BY d.updated_at DESC, c.chunk_index ASC
		LIMIT 20
	`, req.TenantID, req.WorkspaceID, req.Query)
	if err != nil {
		return nil, fmt.Errorf("query document chunks: %w", err)
	}
	defer rows.Close()

	chunks := make([]core.DocumentChunkResult, 0, 20)
	for rows.Next() {
		var chunk core.DocumentChunkResult
		if err := rows.Scan(&chunk.ChunkID, &chunk.DocumentID, &chunk.Text, &chunk.Score); err != nil {
			return nil, fmt.Errorf("scan document chunk result: %w", err)
		}
		chunks = append(chunks, chunk)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate document chunk results: %w", err)
	}
	return &core.SearchDocumentsResponse{Chunks: chunks}, nil
}

```



<!-- Source: internal/store/postgres/search_test.go | bytes=2764 | lines=76 | sha16=fa35492f854bd3eb -->

```go
// ============================================================
// FILE     : internal/store/postgres/search_test.go
// PURPOSE  : Verifies PostgreSQL search SQL preserves actor-private visibility boundaries.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : postgres search statement tests
// DEPENDS  : strings, testing, internal/core
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: These tests lock privacy predicates without requiring a live database.
// ============================================================

package postgres

import (
	"strings"
	"testing"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

func TestSearchMemoriesStatementScopesAgentPrivateToOwner(t *testing.T) {
	t.Parallel()

	sql, args := searchMemoriesStatement(&core.SearchMemoriesRequest{
		TenantID:        "tenant_1",
		WorkspaceID:     "workspace_1",
		OwnerEntityID:   "agent:hermes-main",
		Query:           "stage 2",
		Scopes:          []core.MemoryScope{core.MemoryScopeAgentPrivate, core.MemoryScopeWorkspaceShared},
		ArtifactClasses: []core.ArtifactClass{core.ArtifactClassKnowledge},
	})

	if !strings.Contains(sql, "scope <> 'agent_private'") {
		t.Fatalf("expected non-private scopes to bypass owner filter, got:\n%s", sql)
	}
	if !strings.Contains(sql, "owner_entity_id = $6") {
		t.Fatalf("expected agent_private owner predicate, got:\n%s", sql)
	}
	if !strings.Contains(sql, "group_id, owner_entity_id, valid_from") {
		t.Fatalf("memory search must return owner_entity_id for caller-side visibility checks, got:\n%s", sql)
	}
	if len(args) != 7 || args[5] != "agent:hermes-main" {
		t.Fatalf("unexpected memory search args: %#v", args)
	}
}

func TestSearchMemoriesStatementScopesGroupSharedToMemberships(t *testing.T) {
	t.Parallel()

	sql, args := searchMemoriesStatement(&core.SearchMemoriesRequest{
		TenantID:        "tenant_1",
		WorkspaceID:     "workspace_1",
		OwnerEntityID:   "agent:hermes-main",
		VisibleGroupIDs: []string{"group_design"},
		Query:           "stage 2",
		Scopes:          []core.MemoryScope{core.MemoryScopeGroupShared},
		ArtifactClasses: []core.ArtifactClass{core.ArtifactClassKnowledge},
	})

	if !strings.Contains(sql, "scope <> 'group_shared'") {
		t.Fatalf("expected non-group scopes to bypass group filter, got:\n%s", sql)
	}
	if !strings.Contains(sql, "group_id = ANY($7)") {
		t.Fatalf("expected group_shared membership predicate, got:\n%s", sql)
	}
	if len(args) != 7 {
		t.Fatalf("unexpected arg count: %#v", args)
	}
	gotGroups, ok := args[6].([]string)
	if !ok || len(gotGroups) != 1 || gotGroups[0] != "group_design" {
		t.Fatalf("unexpected group args: %#v", args[6])
	}
}

```



<!-- Source: internal/store/postgres/store.go | bytes=1624 | lines=48 | sha16=498264d80b8d1302 -->

```go
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

```



<!-- Source: internal/store/postgres/timeline.go | bytes=3649 | lines=108 | sha16=3afa25ff6b455ce3 -->

```go
// ============================================================
// FILE     : internal/store/postgres/timeline.go
// PURPOSE  : Implements PostgreSQL read paths for operator timeline views.
// LAYER    : infra
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : GetTimeline
// DEPENDS  : internal/core
// USED_BY  : kernel GetTimeline path
// ------------------------------------------------------------
// AGENT_NOTE: Timeline is read-only here; do not mutate graph, trace, or recall state.
// ============================================================

package postgres

import (
	"context"
	"fmt"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

// GetTimeline loads a read-only timeline view over memories and correction artifacts.
func (s *Store) GetTimeline(ctx context.Context, req *core.GetTimelineRequest) (*core.GetTimelineResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("%w: get timeline request is required", core.ErrInvalidArgument)
	}
	sql, args := timelineStatement(req)
	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query timeline: %w", err)
	}
	defer rows.Close()

	items := make([]core.TimelineItem, 0, req.Limit)
	for rows.Next() {
		var item core.TimelineItem
		if err := rows.Scan(&item.ID, &item.Kind, &item.ArtifactClass, &item.Text,
			&item.OccurredAt, &item.MemoryID, &item.RawEventID); err != nil {
			return nil, fmt.Errorf("scan timeline item: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate timeline items: %w", err)
	}
	return &core.GetTimelineResponse{Items: items}, nil
}

func timelineStatement(req *core.GetTimelineRequest) (string, []any) {
	return `
		WITH memory_items AS (
			SELECT
				m.id AS id,
				m.kind AS kind,
				m.artifact_class AS artifact_class,
				m.text AS text,
				COALESCE(mt.created_at, m.valid_from, m.created_at) AS occurred_at,
				m.id AS memory_id,
				'' AS raw_event_id
			FROM memories m
			LEFT JOIN memory_trace mt ON mt.memory_id = m.id
			WHERE m.tenant_id = $1
			  AND m.workspace_id = $2
			  AND m.scope = ANY($3)
			  AND m.scope <> 'group_shared'
			  AND m.status <> 'deleted'
			  AND ($4::timestamptz IS NULL OR COALESCE(mt.created_at, m.valid_from, m.created_at) >= $4)
			  AND ($5::timestamptz IS NULL OR COALESCE(mt.created_at, m.valid_from, m.created_at) <= $5)
			  AND (
			    m.scope <> 'agent_private'
			    OR ($6 <> '' AND m.owner_entity_id = $6)
			  )
		),
		correction_items AS (
			SELECT
				c.id AS id,
				'correction' AS kind,
				'timeline' AS artifact_class,
				'Correction for memory ' || c.memory_id || ': ' || c.correction_text AS text,
				c.created_at AS occurred_at,
				c.memory_id AS memory_id,
				c.raw_event_id AS raw_event_id
			FROM memory_corrections c
			JOIN memories m ON m.id = c.memory_id
			WHERE c.tenant_id = $1
			  AND c.workspace_id = $2
			  AND m.tenant_id = $1
			  AND m.workspace_id = $2
			  AND m.scope = ANY($3)
			  AND m.scope <> 'group_shared'
			  AND ($4::timestamptz IS NULL OR c.created_at >= $4)
			  AND ($5::timestamptz IS NULL OR c.created_at <= $5)
			  AND (
			    m.scope <> 'agent_private'
			    OR ($6 <> '' AND m.owner_entity_id = $6)
			  )
		)
		SELECT id, kind, artifact_class, text, occurred_at, memory_id, raw_event_id
		FROM memory_items
		UNION ALL
		SELECT id, kind, artifact_class, text, occurred_at, memory_id, raw_event_id
		FROM correction_items
		ORDER BY occurred_at DESC, id DESC
		LIMIT $7
	`, []any{req.TenantID, req.WorkspaceID, req.Scopes, req.From, req.To, req.EntityID, req.Limit}
}

```



<!-- Source: internal/store/postgres/timeline_test.go | bytes=2386 | lines=79 | sha16=6a9c29e82c10aded -->

```go
// ============================================================
// FILE     : internal/store/postgres/timeline_test.go
// PURPOSE  : Verifies timeline query source preserves read-only scope filtering.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : postgres timeline source tests
// DEPENDS  : strings, testing
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: Timeline tests lock read-only query shape without requiring a live database.
// ============================================================

package postgres

import (
	"strings"
	"testing"
)

func TestTimelineStatementReadsMemoriesAndCorrections(t *testing.T) {
	t.Parallel()

	source := readPostgresSourceFile(t, "timeline.go")
	statementSource := extractPostgresSourceBetween(t, source, "func timelineStatement", "`, []any")

	for _, want := range []string{
		"FROM memories m",
		"LEFT JOIN memory_trace mt ON mt.memory_id = m.id",
		"FROM memory_corrections c",
		"JOIN memories m ON m.id = c.memory_id",
		"UNION ALL",
		"ORDER BY occurred_at DESC, id DESC",
	} {
		if !strings.Contains(statementSource, want) {
			t.Fatalf("timeline statement must preserve %q, got:\n%s", want, statementSource)
		}
	}
}

func TestTimelineStatementPreservesScopeAndOwnerFiltering(t *testing.T) {
	t.Parallel()

	source := readPostgresSourceFile(t, "timeline.go")
	statementSource := extractPostgresSourceBetween(t, source, "func timelineStatement", "`, []any")

	for _, want := range []string{
		"m.tenant_id = $1",
		"m.workspace_id = $2",
		"m.scope = ANY($3)",
		"m.scope <> 'group_shared'",
		"m.scope <> 'agent_private'",
		"m.owner_entity_id = $6",
	} {
		if !strings.Contains(statementSource, want) {
			t.Fatalf("timeline statement must preserve %q, got:\n%s", want, statementSource)
		}
	}
}

func TestTimelineStatementDoesNotMutateGraphState(t *testing.T) {
	t.Parallel()

	source := readPostgresSourceFile(t, "timeline.go")
	statementSource := extractPostgresSourceBetween(t, source, "func timelineStatement", "`, []any")

	for _, blocked := range []string{
		"UPDATE ",
		"INSERT ",
		"DELETE ",
		"latest_flag =",
		"memory_edges",
	} {
		if strings.Contains(statementSource, blocked) {
			t.Fatalf("timeline statement must stay read-only; found %q in:\n%s", blocked, statementSource)
		}
	}
}

```



<!-- Source: internal/store/store.go | bytes=7299 | lines=144 | sha16=1ff7387f4b7b2c16 -->

```go
// ============================================================
// FILE     : internal/store/store.go
// PURPOSE  : Defines persistence interfaces for raw, memory, job, profile, note, plan, document, group, and dreaming stores.
// LAYER    : infra
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : RawEventStore, MemoryStore, CorrectionStore, TimelineStore, JobStore, JobMetricsStore, ProfileStore, NoteStore, PlanStore, DocumentStore, SessionSummaryStore, GroupStore, DreamingStore
// DEPENDS  : context, internal/core
// USED_BY  : internal/store/postgres, service implementations
// ------------------------------------------------------------
// AGENT_NOTE: Store contracts must preserve idempotency, provenance, and scope separation.
// ============================================================

// Package store defines storage contracts for VibeGravity persistence.
package store

import (
	"context"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

// RawEventStore persists immutable raw events.
type RawEventStore interface {
	// AppendRawEvents inserts raw events idempotently and returns their IDs.
	AppendRawEvents(ctx context.Context, events []*core.RawEvent) ([]string, error)
	// GetRawEvents loads raw events by ID.
	GetRawEvents(ctx context.Context, ids []string) ([]*core.RawEvent, error)
}

// MemoryStore persists derived memories, edges, and traces.
type MemoryStore interface {
	// UpsertMemory writes a memory after apply validation.
	UpsertMemory(ctx context.Context, memory *core.Memory) error
	// GetMemory loads one memory by ID.
	GetMemory(ctx context.Context, memoryID string) (*core.Memory, error)
	// SearchMemories searches memories with the core search request contract.
	SearchMemories(ctx context.Context, req *core.SearchMemoriesRequest) (*core.SearchMemoriesResponse, error)
	// UpsertMemoryEdge writes a graph edge.
	UpsertMemoryEdge(ctx context.Context, edge *core.MemoryEdge) error
	// WriteMemoryTrace writes mandatory provenance for a memory.
	WriteMemoryTrace(ctx context.Context, trace *core.MemoryTrace) error
	// ExplainMemory loads provenance for one memory.
	ExplainMemory(ctx context.Context, req *core.ExplainMemoryRequest) (*core.ExplainMemoryResponse, error)
}

// CorrectionStore persists human correction intent beside immutable raw events.
type CorrectionStore interface {
	// RecordMemoryCorrection writes the raw correction event and operator-visible artifact idempotently.
	RecordMemoryCorrection(ctx context.Context, event *core.RawEvent, correction *core.MemoryCorrection) (*core.MemoryCorrection, error)
}

// TimelineStore reads operator-visible memory activity.
type TimelineStore interface {
	// GetTimeline loads a read-only timeline view.
	GetTimeline(ctx context.Context, req *core.GetTimelineRequest) (*core.GetTimelineResponse, error)
}

// JobStore persists and claims worker jobs.
type JobStore interface {
	// EnqueueJobs inserts jobs created by the hot ingest path.
	EnqueueJobs(ctx context.Context, jobs []*core.IngestJob) ([]string, error)
	// ClaimJobs claims available jobs for a worker.
	ClaimJobs(ctx context.Context, workerID string, limit int) ([]*core.IngestJob, error)
	// CompleteJob marks a job complete.
	CompleteJob(ctx context.Context, jobID string) error
	// FailJob records a failed attempt and retry state.
	FailJob(ctx context.Context, jobID string, err error) error
	// BlockJob records deterministic unsupported work without scheduling retry.
	BlockJob(ctx context.Context, jobID string, err error) error
}

// JobMetricsStore reads worker queue health without mutating job state.
type JobMetricsStore interface {
	// GetJobBacklogMetrics returns operator-visible backlog counts and recovery estimates.
	GetJobBacklogMetrics(ctx context.Context, req *core.JobBacklogMetricsRequest) (*core.JobBacklogMetrics, error)
}

// ProfileStore persists rebuildable profile snapshots.
type ProfileStore interface {
	// GetProfile loads a profile snapshot.
	GetProfile(ctx context.Context, entityID string, scope core.MemoryScope) (*core.Profile, error)
	// UpsertProfile writes a profile snapshot.
	UpsertProfile(ctx context.Context, profile *core.Profile) error
}

// NoteStore persists human-authored notes.
type NoteStore interface {
	// AddNote writes a note.
	AddNote(ctx context.Context, note *core.Note) error
	// ListPinnedNotes loads pinned notes for recall.
	ListPinnedNotes(ctx context.Context, req *core.ListPinnedNotesRequest) ([]*core.Note, error)
}

// PlanStore persists structured plans and plan items.
type PlanStore interface {
	// CreatePlan writes a plan and its initial items.
	CreatePlan(ctx context.Context, plan *core.Plan, items []*core.PlanItem) error
	// UpdatePlan updates a plan and its items.
	UpdatePlan(ctx context.Context, plan *core.Plan, items []*core.PlanItem) error
	// GetActivePlans loads active plans for recall.
	GetActivePlans(ctx context.Context, req *core.GetActivePlansRequest) ([]*core.Plan, error)
}

// DocumentStore persists documents and searchable chunks.
type DocumentStore interface {
	// AddDocumentWithChunks writes a document and replaces its chunks atomically.
	AddDocumentWithChunks(ctx context.Context, document *core.Document, chunks []*core.DocumentChunk) error
	// AddDocument writes a document.
	AddDocument(ctx context.Context, document *core.Document) error
	// AddDocumentChunks writes retrieval chunks for a document.
	AddDocumentChunks(ctx context.Context, chunks []*core.DocumentChunk) error
	// SearchDocuments searches document chunks with the core search contract.
	SearchDocuments(ctx context.Context, req *core.SearchDocumentsRequest) (*core.SearchDocumentsResponse, error)
}

// SessionSummaryStore persists per-session summaries.
type SessionSummaryStore interface {
	// UpsertSessionSummary writes a session summary.
	UpsertSessionSummary(ctx context.Context, summary *core.SessionSummary) error
	// GetSessionSummary loads the current summary for a session.
	GetSessionSummary(ctx context.Context, sessionID string) (*core.SessionSummary, error)
}

// DreamingStore persists and loads background consolidation state.
type DreamingStore interface {
	// LoadDreamingSessionInput loads raw event and derived memory inputs for one session.
	LoadDreamingSessionInput(ctx context.Context, req *core.DreamSessionRequest) (*core.DreamingSessionInput, error)
	// PromoteMemories marks existing memories with a deeper dreaming tier without changing scope.
	PromoteMemories(ctx context.Context, req *core.DreamingPromotionRequest) (*core.DreamingPromotionResult, error)
}

// GroupStore persists memory groups and memberships.
type GroupStore interface {
	// CreateMemoryGroup writes a memory group.
	CreateMemoryGroup(ctx context.Context, group *core.MemoryGroup) error
	// AddMembership adds an entity to a memory group.
	AddMembership(ctx context.Context, membership *core.MemoryGroupMembership) error
	// ListMemberships loads memberships for a memory group.
	ListMemberships(ctx context.Context, groupID string) ([]*core.MemoryGroupMembership, error)
	// ListMembershipsForEntity loads groups visible to an entity in one workspace.
	ListMembershipsForEntity(ctx context.Context, tenantID string, workspaceID string, entityID string) ([]*core.MemoryGroupMembership, error)
}

```



<!-- Source: migrations/000001_create_pgvector_extension.down.sql | bytes=33 | lines=2 | sha16=92a6c13145ed5919 -->

```sql
DROP EXTENSION IF EXISTS vector;

```



<!-- Source: migrations/000001_create_pgvector_extension.up.sql | bytes=39 | lines=2 | sha16=9e9b2cfec47519f4 -->

```sql
CREATE EXTENSION IF NOT EXISTS vector;

```



<!-- Source: migrations/000002_create_core_tables.down.sql | bytes=553 | lines=17 | sha16=63920f9c5c460dd5 -->

```sql
DROP TABLE IF EXISTS memory_group_memberships;
DROP TABLE IF EXISTS document_chunks;
DROP TABLE IF EXISTS documents;
DROP TABLE IF EXISTS plan_items;
DROP TABLE IF EXISTS plans;
DROP TABLE IF EXISTS notes;
DROP TABLE IF EXISTS session_summaries;
DROP TABLE IF EXISTS profiles;
DROP TABLE IF EXISTS memory_corrections;
DROP TABLE IF EXISTS memory_trace;
DROP TABLE IF EXISTS memory_edges;
DROP TABLE IF EXISTS memories;
DROP TABLE IF EXISTS memory_groups;
DROP TABLE IF EXISTS entities;
DROP TABLE IF EXISTS ingest_jobs;
DROP TABLE IF EXISTS raw_events;

```



<!-- Source: migrations/000002_create_core_tables.up.sql | bytes=9907 | lines=307 | sha16=89c179f9dc95742b -->

```sql
CREATE TABLE raw_events (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    session_id TEXT NOT NULL,
    actor_id TEXT NOT NULL,
    event_kind TEXT NOT NULL,
    source TEXT NOT NULL,
    idempotency_key TEXT NOT NULL,
    fingerprint TEXT NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL,
    payload_json JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX raw_events_tenant_source_idempotency_key_idx
    ON raw_events (tenant_id, source, idempotency_key);
CREATE INDEX raw_events_tenant_workspace_session_created_at_idx
    ON raw_events (tenant_id, workspace_id, session_id, created_at DESC);

CREATE TABLE ingest_jobs (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    job_kind TEXT NOT NULL CHECK (job_kind IN (
        'process_turn_event',
        'embed_document_chunks',
        'dream_session',
        'dream_workspace',
        'rebuild_profile',
        'maintenance'
    )),
    status TEXT NOT NULL DEFAULT 'queued',
    raw_event_ids TEXT[] NOT NULL DEFAULT '{}',
    payload_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    attempts INTEGER NOT NULL DEFAULT 0 CHECK (attempts >= 0),
    available_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    locked_by TEXT,
    locked_at TIMESTAMPTZ,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX ingest_jobs_kind_status_available_at_idx
    ON ingest_jobs (job_kind, status, available_at);

CREATE TABLE entities (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    entity_kind TEXT NOT NULL,
    display_name TEXT NOT NULL,
    metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE memory_groups (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE memories (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    scope TEXT NOT NULL CHECK (scope IN (
        'agent_private',
        'workspace_shared',
        'group_shared',
        'session_scratch'
    )),
    group_id TEXT,
    owner_entity_id TEXT NOT NULL,
    kind TEXT NOT NULL CHECK (kind IN (
        'fact',
        'preference',
        'trait',
        'goal',
        'constraint',
        'relationship',
        'decision',
        'procedure',
        'task_state',
        'doc_fact',
        'summary',
        'hypothesis'
    )),
    artifact_class TEXT NOT NULL DEFAULT 'knowledge' CHECK (artifact_class IN (
        'context',
        'knowledge',
        'timeline',
        'plan'
    )),
    text TEXT NOT NULL,
    fingerprint TEXT NOT NULL,
    confidence DOUBLE PRECISION NOT NULL CHECK (confidence >= 0 AND confidence <= 1),
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN (
        'active',
        'superseded',
        'archived',
        'deleted'
    )),
    valid_from TIMESTAMPTZ NOT NULL,
    valid_to TIMESTAMPTZ,
    latest_flag BOOLEAN NOT NULL DEFAULT true,
    metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX memories_tenant_workspace_scope_status_idx
    ON memories (tenant_id, workspace_id, scope, status);
CREATE INDEX memories_fingerprint_idx
    ON memories (fingerprint);

CREATE TABLE memory_edges (
    from_memory_id TEXT NOT NULL REFERENCES memories (id) ON DELETE CASCADE,
    to_memory_id TEXT NOT NULL REFERENCES memories (id) ON DELETE CASCADE,
    edge_kind TEXT NOT NULL CHECK (edge_kind IN (
        'updates',
        'extends',
        'supports',
        'contradicts',
        'derived_from',
        'references_doc',
        'belongs_to',
        'corrected_by'
    )),
    confidence DOUBLE PRECISION NOT NULL CHECK (confidence >= 0 AND confidence <= 1),
    created_by_job_id TEXT REFERENCES ingest_jobs (id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (from_memory_id, to_memory_id, edge_kind)
);

CREATE INDEX memory_edges_from_memory_edge_kind_idx
    ON memory_edges (from_memory_id, edge_kind);
CREATE INDEX memory_edges_to_memory_edge_kind_idx
    ON memory_edges (to_memory_id, edge_kind);
CREATE UNIQUE INDEX memory_edges_single_updates_target_idx
    ON memory_edges (to_memory_id)
    WHERE edge_kind = 'updates';

CREATE TABLE memory_trace (
    memory_id TEXT PRIMARY KEY REFERENCES memories (id) ON DELETE CASCADE,
    raw_event_ids TEXT[] NOT NULL,
    reasoning_job_id TEXT REFERENCES ingest_jobs (id),
    reasoning_stage TEXT NOT NULL,
    candidate_snapshot_json JSONB NOT NULL,
    applied_operations_json JSONB NOT NULL,
    operator_correction_flag BOOLEAN NOT NULL DEFAULT false,
    related_document_ids TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE memory_corrections (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    memory_id TEXT NOT NULL REFERENCES memories (id) ON DELETE CASCADE,
    operator_id TEXT NOT NULL,
    raw_event_id TEXT NOT NULL REFERENCES raw_events (id) ON DELETE CASCADE,
    idempotency_key TEXT NOT NULL,
    correction_text TEXT NOT NULL,
    evidence_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    status TEXT NOT NULL DEFAULT 'recorded' CHECK (status IN (
        'recorded',
        'applied',
        'dismissed'
    )),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX memory_corrections_tenant_workspace_idempotency_key_idx
    ON memory_corrections (tenant_id, workspace_id, idempotency_key);
CREATE UNIQUE INDEX memory_corrections_raw_event_id_idx
    ON memory_corrections (raw_event_id);
CREATE INDEX memory_corrections_memory_created_at_idx
    ON memory_corrections (memory_id, created_at DESC);

CREATE TABLE profiles (
    entity_id TEXT NOT NULL,
    scope TEXT NOT NULL CHECK (scope IN (
        'agent_private',
        'workspace_shared',
        'group_shared',
        'session_scratch'
    )),
    static_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    dynamic_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    source_memory_ids TEXT[] NOT NULL DEFAULT '{}',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    version BIGINT NOT NULL DEFAULT 1,
    PRIMARY KEY (entity_id, scope)
);

CREATE INDEX profiles_entity_updated_at_idx
    ON profiles (entity_id, updated_at DESC);

CREATE TABLE session_summaries (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    session_id TEXT NOT NULL,
    summary_text TEXT NOT NULL,
    source_event_ids TEXT[] NOT NULL DEFAULT '{}',
    source_memory_ids TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE notes (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    note_kind TEXT NOT NULL,
    scope TEXT NOT NULL CHECK (scope IN (
        'agent_private',
        'workspace_shared',
        'group_shared',
        'session_scratch'
    )),
    owner_entity_id TEXT NOT NULL,
    text TEXT NOT NULL,
    pinned BOOLEAN NOT NULL DEFAULT false,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX notes_workspace_pinned_expires_at_idx
    ON notes (workspace_id, pinned, expires_at);

CREATE TABLE plans (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    title TEXT NOT NULL,
    status TEXT NOT NULL,
    scope TEXT NOT NULL CHECK (scope IN (
        'agent_private',
        'workspace_shared',
        'group_shared',
        'session_scratch'
    )),
    owner_entity_id TEXT NOT NULL,
    evidence_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX plans_workspace_status_idx
    ON plans (workspace_id, status);

CREATE TABLE plan_items (
    id TEXT PRIMARY KEY,
    plan_id TEXT NOT NULL REFERENCES plans (id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    status TEXT NOT NULL,
    evidence_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE documents (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    source TEXT NOT NULL,
    title TEXT NOT NULL,
    fingerprint TEXT NOT NULL,
    metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    version_hint TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX documents_tenant_workspace_fingerprint_idx
    ON documents (tenant_id, workspace_id, fingerprint);

CREATE TABLE document_chunks (
    id TEXT PRIMARY KEY,
    document_id TEXT NOT NULL REFERENCES documents (id) ON DELETE CASCADE,
    chunk_index INTEGER NOT NULL CHECK (chunk_index >= 0),
    text TEXT NOT NULL,
    heading_path TEXT NOT NULL DEFAULT '',
    metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    neighbor_chunk_ids TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX document_chunks_document_chunk_index_idx
    ON document_chunks (document_id, chunk_index);

CREATE TABLE memory_group_memberships (
    group_id TEXT NOT NULL REFERENCES memory_groups (id) ON DELETE CASCADE,
    entity_id TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (group_id, entity_id)
);

```



<!-- Source: migrations/000003_add_vector_columns.down.sql | bytes=390 | lines=12 | sha16=49c02fb4bb2e4275 -->

```sql
ALTER TABLE document_chunks
    DROP COLUMN IF EXISTS embedding_updated_at,
    DROP COLUMN IF EXISTS embedding_dims,
    DROP COLUMN IF EXISTS embedding_model,
    DROP COLUMN IF EXISTS embedding;

ALTER TABLE memories
    DROP COLUMN IF EXISTS embedding_updated_at,
    DROP COLUMN IF EXISTS embedding_dims,
    DROP COLUMN IF EXISTS embedding_model,
    DROP COLUMN IF EXISTS embedding;

```



<!-- Source: migrations/000003_add_vector_columns.up.sql | bytes=813 | lines=16 | sha16=e3dd3c20d81b2074 -->

```sql
-- ADR-002 forbids hardcoding embedding dimensions before model selection.
-- The pgvector type is intentionally dimensionless here; when ADR-002b fixes
-- the model, regenerate this migration or add a follow-up migration that uses
-- vector(<embedding_dims>) from config and backfills rows safely.
ALTER TABLE memories
    ADD COLUMN embedding vector,
    ADD COLUMN embedding_model TEXT NOT NULL DEFAULT 'pending',
    ADD COLUMN embedding_dims INTEGER NOT NULL DEFAULT 0 CHECK (embedding_dims >= 0),
    ADD COLUMN embedding_updated_at TIMESTAMPTZ;

ALTER TABLE document_chunks
    ADD COLUMN embedding vector,
    ADD COLUMN embedding_model TEXT NOT NULL DEFAULT 'pending',
    ADD COLUMN embedding_dims INTEGER NOT NULL DEFAULT 0 CHECK (embedding_dims >= 0),
    ADD COLUMN embedding_updated_at TIMESTAMPTZ;

```
