// ============================================================
// FILE     : internal/kernel/correction_trust_loop_integration_test.go
// PURPOSE  : Verifies the full CorrectMemory trust loop against live PostgreSQL.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : TestPostgresCorrectMemoryTrustLoop
// DEPENDS  : context, errors, os, testing, time, internal/core, internal/ingest, internal/recall, internal/store/postgres, pgxpool
// USED_BY  : make integration-postgres, go test ./internal/kernel
// ------------------------------------------------------------
// AGENT_NOTE: Keep this test opt-in through VIBEGRAVITY_DB_URL; it must exercise real migrations, FKs, transactions, and store behavior.
// ============================================================

package kernel

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/ingest"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/recall"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/store/postgres"
)

func TestPostgresCorrectMemoryTrustLoop(t *testing.T) {
	dbURL := os.Getenv("VIBEGRAVITY_DB_URL")
	if dbURL == "" {
		t.Skip("Skipping live Postgres CorrectMemory trust-loop test because VIBEGRAVITY_DB_URL is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("connect live Postgres test database: %v", err)
	}
	defer pool.Close()

	store := postgres.NewStore(pool)
	ingestService, err := ingest.NewService(ingest.Dependencies{
		RawEvents: store,
		Jobs:      store,
	})
	if err != nil {
		t.Fatalf("build ingest service: %v", err)
	}
	service, err := NewService(Dependencies{
		Ingest:      ingestService,
		Recall:      recall.NewAssembler(recall.Dependencies{Memories: store}),
		Memories:    store,
		Corrections: store,
		Jobs:        store,
		Timeline:    store,
	})
	if err != nil {
		t.Fatalf("build kernel service: %v", err)
	}

	tenantID := fmt.Sprintf("tenant_correction_trust_%d", time.Now().UnixNano())
	workspaceID := "workspace_correction_trust"
	ownerID := "agent:hermes-main"
	operatorID := "operator:trust-loop"
	targetID := "mem_correction_trust_target"
	seedJobID := "job_correction_trust_seed"
	startedAt := time.Now().UTC().Add(-1 * time.Minute)

	cleanupPostgresCorrectionTrustRows(ctx, t, pool, tenantID, workspaceID)
	defer cleanupPostgresCorrectionTrustRows(context.Background(), t, pool, tenantID, workspaceID)

	insertPostgresCorrectionTrustJob(ctx, t, pool, tenantID, workspaceID, seedJobID, []string{"evt_correction_trust_seed"})
	if err := store.CreateMemoryWithTrace(ctx, &core.Memory{
		ID:            targetID,
		TenantID:      tenantID,
		WorkspaceID:   workspaceID,
		Scope:         core.MemoryScopeWorkspaceShared,
		OwnerEntityID: ownerID,
		Kind:          core.MemoryKindFact,
		ArtifactClass: core.ArtifactClassKnowledge,
		Text:          "Hermes should use the legacy deployment key for VibeGravity.",
		Fingerprint:   "fp_correction_trust_target",
		Confidence:    0.64,
		Status:        core.MemoryStatusActive,
		ValidFrom:     startedAt,
		LatestFlag:    true,
		MetadataJSON:  json.RawMessage(`{"seed":true}`),
		CreatedAt:     startedAt,
		UpdatedAt:     startedAt,
	}, &core.MemoryTrace{
		MemoryID:              targetID,
		RawEventIDs:           []string{"evt_correction_trust_seed"},
		ReasoningJobID:        seedJobID,
		ReasoningStage:        "resolve",
		CandidateSnapshotJSON: json.RawMessage(`{"seed":true}`),
		AppliedOperationsJSON: json.RawMessage(`[{"operation_id":"seed-correction-trust-target"}]`),
		RelatedDocumentIDs:    []string{},
		CreatedAt:             startedAt,
	}); err != nil {
		t.Fatalf("seed target memory and trace through real store: %v", err)
	}

	req := &core.CorrectMemoryRequest{
		TenantID:       tenantID,
		WorkspaceID:    workspaceID,
		MemoryID:       targetID,
		OperatorID:     operatorID,
		IdempotencyKey: "correct-vibegravity-deployment-key",
		CorrectionText: "Hermes should use the rotated VibeGravity deployment key.",
		EvidenceJSON:   json.RawMessage(`{"source":"operator_live_postgres_test"}`),
	}
	resp, err := service.CorrectMemory(ctx, req)
	if err != nil {
		t.Fatalf("CorrectMemory live Postgres call failed: %v", err)
	}
	if resp.Status != "applied" || !resp.CorrectionRecorded || !resp.TraceWritten {
		t.Fatalf("CorrectMemory did not report applied trust-loop side effects: %#v", resp)
	}

	replacementID := assertPostgresCorrectionGraph(ctx, t, pool, tenantID, workspaceID, targetID, req.CorrectionText)
	assertPostgresCorrectionFKs(ctx, t, pool, replacementID, targetID)
	assertPostgresCorrectionExplain(ctx, t, service, tenantID, workspaceID, ownerID, replacementID, targetID, resp.RawEventID)
	assertPostgresCorrectionTimeline(ctx, t, service, tenantID, workspaceID, ownerID, replacementID, targetID, resp.RawEventID, req.CorrectionText)
	assertPostgresCorrectionRecall(ctx, t, service, tenantID, workspaceID, ownerID, replacementID, targetID, req.CorrectionText)

	beforeRetry := countPostgresCorrectionRows(ctx, t, pool, tenantID, workspaceID)
	retryResp, err := service.CorrectMemory(ctx, req)
	if err != nil {
		t.Fatalf("CorrectMemory idempotent live retry failed: %v", err)
	}
	if retryResp.RawEventID != resp.RawEventID || retryResp.CorrectionID != resp.CorrectionID || retryResp.Status != "applied" {
		t.Fatalf("CorrectMemory retry did not return the same correction artifact: first=%#v retry=%#v", resp, retryResp)
	}
	afterRetry := countPostgresCorrectionRows(ctx, t, pool, tenantID, workspaceID)
	if beforeRetry != afterRetry {
		t.Fatalf("CorrectMemory retry duplicated rows: before=%+v after=%+v", beforeRetry, afterRetry)
	}

	changed := *req
	changed.CorrectionText = "Hermes should use a different text under the same idempotency key."
	_, err = service.CorrectMemory(ctx, &changed)
	if !errors.Is(err, core.ErrConflict) {
		t.Fatalf("same idempotency key with changed correction text should conflict, got %v", err)
	}
	afterConflict := countPostgresCorrectionRows(ctx, t, pool, tenantID, workspaceID)
	if afterRetry != afterConflict {
		t.Fatalf("conflicting correction retry changed row counts: before=%+v after=%+v", afterRetry, afterConflict)
	}
}

type correctionTrustCounts struct {
	Memories    int
	Traces      int
	Edges       int
	Corrections int
	RawEvents   int
	Jobs        int
}

func assertPostgresCorrectionGraph(ctx context.Context, t *testing.T, pool *pgxpool.Pool, tenantID, workspaceID, targetID, correctedText string) string {
	t.Helper()

	var replacementID string
	var replacementText string
	var replacementStatus core.MemoryStatus
	var replacementLatest bool
	var targetStatus core.MemoryStatus
	var targetLatest bool
	if err := pool.QueryRow(ctx, `
		SELECT replacement.id, replacement.text, replacement.status, replacement.latest_flag,
		       target.status, target.latest_flag
		FROM memory_edges edge
		JOIN memories replacement ON replacement.id = edge.from_memory_id
		JOIN memories target ON target.id = edge.to_memory_id
		WHERE target.tenant_id = $1
		  AND target.workspace_id = $2
		  AND edge.to_memory_id = $3
		  AND edge.edge_kind = 'updates'
	`, tenantID, workspaceID, targetID).Scan(&replacementID, &replacementText, &replacementStatus,
		&replacementLatest, &targetStatus, &targetLatest); err != nil {
		t.Fatalf("load correction replacement graph: %v", err)
	}
	if replacementText != correctedText || replacementStatus != core.MemoryStatusActive || !replacementLatest {
		t.Fatalf("replacement memory is not the active corrected memory: id=%s text=%q status=%q latest=%v", replacementID, replacementText, replacementStatus, replacementLatest)
	}
	if targetStatus != core.MemoryStatusSuperseded || targetLatest {
		t.Fatalf("target memory should be superseded and non-latest, got status=%q latest=%v", targetStatus, targetLatest)
	}

	var dangling int
	if err := pool.QueryRow(ctx, `
		SELECT count(*)
		FROM memories m
		LEFT JOIN memory_trace mt ON mt.memory_id = m.id
		WHERE m.tenant_id = $1
		  AND m.workspace_id = $2
		  AND m.id = $3
		  AND mt.memory_id IS NULL
	`, tenantID, workspaceID, replacementID).Scan(&dangling); err != nil {
		t.Fatalf("check replacement trace presence: %v", err)
	}
	if dangling != 0 {
		t.Fatalf("replacement memory has no trace")
	}
	return replacementID
}

func assertPostgresCorrectionFKs(ctx context.Context, t *testing.T, pool *pgxpool.Pool, replacementID, targetID string) {
	t.Helper()

	var traceJobID string
	var edgeJobID string
	var traceJobExists bool
	var edgeJobExists bool
	if err := pool.QueryRow(ctx, `
		SELECT mt.reasoning_job_id,
		       me.created_by_job_id,
		       trace_job.id IS NOT NULL,
		       edge_job.id IS NOT NULL
		FROM memory_trace mt
		JOIN memory_edges me ON me.from_memory_id = mt.memory_id
		LEFT JOIN ingest_jobs trace_job ON trace_job.id = mt.reasoning_job_id
		LEFT JOIN ingest_jobs edge_job ON edge_job.id = me.created_by_job_id
		WHERE mt.memory_id = $1
		  AND me.to_memory_id = $2
		  AND me.edge_kind = 'updates'
	`, replacementID, targetID).Scan(&traceJobID, &edgeJobID, &traceJobExists, &edgeJobExists); err != nil {
		t.Fatalf("load correction FK proof: %v", err)
	}
	if traceJobID == "" || edgeJobID == "" || traceJobID != edgeJobID || !traceJobExists || !edgeJobExists {
		t.Fatalf("trace and edge must share an existing ingest_jobs row: trace=%q exists=%v edge=%q exists=%v", traceJobID, traceJobExists, edgeJobID, edgeJobExists)
	}
}

func assertPostgresCorrectionExplain(ctx context.Context, t *testing.T, service *Service, tenantID, workspaceID, ownerID, replacementID, targetID, rawEventID string) {
	t.Helper()

	explain, err := service.ExplainMemory(ctx, &core.ExplainMemoryRequest{
		TenantID:    tenantID,
		WorkspaceID: workspaceID,
		MemoryID:    replacementID,
		EntityID:    ownerID,
	})
	if err != nil {
		t.Fatalf("ExplainMemory for replacement failed: %v", err)
	}
	if !explain.Trace.OperatorCorrectionFlag || explain.Trace.ReasoningJobID == "" {
		t.Fatalf("ExplainMemory lost replacement correction trace: %#v", explain.Trace)
	}
	if len(explain.Trace.RawEventIDs) != 1 || explain.Trace.RawEventIDs[0] != rawEventID {
		t.Fatalf("ExplainMemory trace should point at correction raw event %q: %#v", rawEventID, explain.Trace.RawEventIDs)
	}
	if !hasUpdatesEdge(explain.Edges, replacementID, targetID) {
		t.Fatalf("ExplainMemory should show replacement updates edge: %#v", explain.Edges)
	}
	if len(explain.SourceEvents) != 1 || explain.SourceEvents[0].EventID != rawEventID || explain.SourceEvents[0].EventKind != "memory_correction" {
		t.Fatalf("ExplainMemory should include correction source event: %#v", explain.SourceEvents)
	}
}

func assertPostgresCorrectionTimeline(ctx context.Context, t *testing.T, service *Service, tenantID, workspaceID, ownerID, replacementID, targetID, rawEventID, correctedText string) {
	t.Helper()

	timeline, err := service.GetTimeline(ctx, &core.GetTimelineRequest{
		TenantID:    tenantID,
		WorkspaceID: workspaceID,
		EntityID:    ownerID,
		Scopes:      []core.MemoryScope{core.MemoryScopeWorkspaceShared},
		Limit:       20,
	})
	if err != nil {
		t.Fatalf("GetTimeline failed: %v", err)
	}
	var sawReplacement bool
	var sawCorrection bool
	for _, item := range timeline.Items {
		if item.MemoryID == replacementID && item.Text == correctedText {
			sawReplacement = true
		}
		if item.Kind == core.MemoryKindCorrection && item.MemoryID == targetID && item.RawEventID == rawEventID && strings.Contains(item.Text, correctedText) {
			sawCorrection = true
		}
	}
	if !sawReplacement || !sawCorrection {
		t.Fatalf("timeline should include replacement and correction artifact: replacement=%v correction=%v items=%#v", sawReplacement, sawCorrection, timeline.Items)
	}
}

func assertPostgresCorrectionRecall(ctx context.Context, t *testing.T, service *Service, tenantID, workspaceID, ownerID, replacementID, targetID, correctedText string) {
	t.Helper()

	search, err := service.SearchMemories(ctx, &core.SearchMemoriesRequest{
		TenantID:        tenantID,
		WorkspaceID:     workspaceID,
		OwnerEntityID:   ownerID,
		Query:           "VibeGravity deployment key",
		Scopes:          []core.MemoryScope{core.MemoryScopeWorkspaceShared},
		ArtifactClasses: []core.ArtifactClass{core.ArtifactClassKnowledge},
	})
	if err != nil {
		t.Fatalf("SearchMemories failed: %v", err)
	}
	if !hasMemoryResult(search.Memories, replacementID, correctedText) || hasMemoryID(search.Memories, targetID) {
		t.Fatalf("SearchMemories should return corrected latest memory and suppress target: results=%#v", search.Memories)
	}

	prefetch, err := service.Prefetch(ctx, &core.PrefetchRequest{
		TenantID:    tenantID,
		WorkspaceID: workspaceID,
		SessionID:   "session-correction-trust",
		ActorID:     ownerID,
		Query:       "VibeGravity deployment key",
	})
	if err != nil {
		t.Fatalf("Prefetch failed: %v", err)
	}
	if !hasRecallBlock(prefetch.Blocks, replacementID, correctedText) || hasRecallBlockID(prefetch.Blocks, targetID) {
		t.Fatalf("Prefetch should return corrected latest memory and suppress target: blocks=%#v", prefetch.Blocks)
	}
}

func countPostgresCorrectionRows(ctx context.Context, t *testing.T, pool *pgxpool.Pool, tenantID, workspaceID string) correctionTrustCounts {
	t.Helper()

	var counts correctionTrustCounts
	if err := pool.QueryRow(ctx, `
		SELECT
			(SELECT count(*) FROM memories WHERE tenant_id = $1 AND workspace_id = $2),
			(SELECT count(*) FROM memory_trace mt JOIN memories m ON m.id = mt.memory_id WHERE m.tenant_id = $1 AND m.workspace_id = $2),
			(SELECT count(*) FROM memory_edges me JOIN memories m ON m.id = me.from_memory_id WHERE m.tenant_id = $1 AND m.workspace_id = $2),
			(SELECT count(*) FROM memory_corrections WHERE tenant_id = $1 AND workspace_id = $2),
			(SELECT count(*) FROM raw_events WHERE tenant_id = $1 AND workspace_id = $2),
			(SELECT count(*) FROM ingest_jobs WHERE tenant_id = $1 AND workspace_id = $2)
	`, tenantID, workspaceID).Scan(&counts.Memories, &counts.Traces, &counts.Edges,
		&counts.Corrections, &counts.RawEvents, &counts.Jobs); err != nil {
		t.Fatalf("count correction trust-loop rows: %v", err)
	}
	return counts
}

func insertPostgresCorrectionTrustJob(ctx context.Context, t *testing.T, pool *pgxpool.Pool, tenantID, workspaceID, jobID string, rawEventIDs []string) {
	t.Helper()

	_, err := pool.Exec(ctx, `
		INSERT INTO ingest_jobs (id, tenant_id, workspace_id, job_kind, status, raw_event_ids, payload_json)
		VALUES ($1, $2, $3, 'process_turn_event', 'complete', $4, '{}'::jsonb)
		ON CONFLICT (id) DO NOTHING
	`, jobID, tenantID, workspaceID, rawEventIDs)
	if err != nil {
		t.Fatalf("seed ingest job %q: %v", jobID, err)
	}
}

func cleanupPostgresCorrectionTrustRows(ctx context.Context, t testing.TB, pool *pgxpool.Pool, tenantID, workspaceID string) {
	t.Helper()

	if _, err := pool.Exec(ctx, `
		DELETE FROM memories
		WHERE tenant_id = $1 AND workspace_id = $2
	`, tenantID, workspaceID); err != nil {
		t.Fatalf("cleanup memories: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		DELETE FROM raw_events
		WHERE tenant_id = $1 AND workspace_id = $2
	`, tenantID, workspaceID); err != nil {
		t.Fatalf("cleanup raw events: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		DELETE FROM ingest_jobs
		WHERE tenant_id = $1 AND workspace_id = $2
	`, tenantID, workspaceID); err != nil {
		t.Fatalf("cleanup ingest jobs: %v", err)
	}
}

func hasUpdatesEdge(edges []core.MemoryEdgeResult, fromID, toID string) bool {
	for _, edge := range edges {
		if edge.FromMemoryID == fromID && edge.ToMemoryID == toID && edge.EdgeKind == core.EdgeKindUpdates {
			return true
		}
	}
	return false
}

func hasMemoryResult(results []core.MemoryResult, memoryID, text string) bool {
	for _, result := range results {
		if result.MemoryID == memoryID && result.Text == text {
			return true
		}
	}
	return false
}

func hasMemoryID(results []core.MemoryResult, memoryID string) bool {
	for _, result := range results {
		if result.MemoryID == memoryID {
			return true
		}
	}
	return false
}

func hasRecallBlock(blocks []core.RecallBlock, sourceID, text string) bool {
	for _, block := range blocks {
		if block.SourceID == sourceID && block.Text == text {
			return true
		}
	}
	return false
}

func hasRecallBlockID(blocks []core.RecallBlock, sourceID string) bool {
	for _, block := range blocks {
		if block.SourceID == sourceID {
			return true
		}
	}
	return false
}
