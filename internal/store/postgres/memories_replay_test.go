// ============================================================
// FILE     : internal/store/postgres/memories_replay_test.go
// PURPOSE  : Locks evidence-safe replay idempotency contracts for memory graph writes.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : TestPostgresUpdateMemoryReplayRequiresIdenticalEvidence, TestPostgresCreateMemoryReplayRejectsTraceEvidenceOverwrite, TestPostgresReplaySourceContractsRequireFullEvidenceComparison
// DEPENDS  : context, errors, os, strings, testing, time, internal/core, github.com/jackc/pgx/v5/pgxpool
// USED_BY  : go test ./internal/store/postgres
// ------------------------------------------------------------
// AGENT_NOTE: Same reasoning_job_id and operation_id retries must be idempotent only when all evidence is identical.
// ============================================================

package postgres

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

const (
	replayWorkspaceID = "workspace_replay"
	replayOwnerID     = "agent:hermes-main"
	replayJobID       = "job_replay_update"
	replayUpdateID    = "mem_replay_update"
)

func TestPostgresUpdateMemoryReplayRequiresIdenticalEvidence(t *testing.T) {
	dbURL := os.Getenv("VIBEGRAVITY_DB_URL")
	if dbURL == "" {
		t.Skip("Skipping Postgres replay integration test because VIBEGRAVITY_DB_URL is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("connect test database: %v", err)
	}
	defer pool.Close()

	store := NewStore(pool)
	tenantID := fmt.Sprintf("tenant_replay_%d", time.Now().UnixNano())
	workspaceID := replayWorkspaceID
	ownerID := replayOwnerID
	targetID := "mem_replay_target"
	updateID := replayUpdateID
	reasoningJobID := replayJobID
	createdAt := time.Now().UTC()

	cleanupPostgresReplayRows(ctx, t, pool, tenantID)
	defer cleanupPostgresReplayRows(context.Background(), t, pool, tenantID)

	mustSeedJob(ctx, t, pool, tenantID, workspaceID, "job_replay_seed")
	mustSeedJob(ctx, t, pool, tenantID, workspaceID, reasoningJobID)
	seedReplayTargetMemory(ctx, t, store, tenantID, workspaceID, ownerID, targetID, createdAt)

	memory := replayMemory(tenantID, updateID, "Hermes should preserve evidence-safe replay.", "fp_replay_update", createdAt)
	trace := replayTrace(updateID, reasoningJobID, []string{"evt_replay_1"}, `[{"operation_id":"op_replay","kind":"update_memory","memory":{"text":"Hermes should preserve evidence-safe replay."}}]`, createdAt)
	edge := replayUpdateEdge(targetID, createdAt)

	if err := store.CreateMemoryWithTraceAndUpdateEdge(ctx, memory, trace, edge); err != nil {
		t.Fatalf("initial update apply failed: %v", err)
	}
	if err := store.CreateMemoryWithTraceAndUpdateEdge(ctx, memory, trace, edge); err != nil {
		t.Fatalf("identical update replay should be idempotent success: %v", err)
	}

	tests := []struct {
		name   string
		memory *core.Memory
		trace  *core.MemoryTrace
		edge   *core.MemoryEdge
	}{
		{
			name:   "changed replacement text",
			memory: replayMemory(tenantID, updateID, "Hermes should accept changed replay text.", "fp_replay_update_changed", createdAt),
			trace:  replayTrace(updateID, reasoningJobID, []string{"evt_replay_1"}, `[{"operation_id":"op_replay","kind":"update_memory","memory":{"text":"Hermes should accept changed replay text."}}]`, createdAt),
			edge:   replayUpdateEdge(targetID, createdAt),
		},
		{
			name:   "changed raw event ids",
			memory: replayMemory(tenantID, updateID, "Hermes should preserve evidence-safe replay.", "fp_replay_update", createdAt),
			trace:  replayTrace(updateID, reasoningJobID, []string{"evt_replay_2"}, `[{"operation_id":"op_replay","kind":"update_memory","memory":{"text":"Hermes should preserve evidence-safe replay."}}]`, createdAt),
			edge:   replayUpdateEdge(targetID, createdAt),
		},
		{
			name:   "changed target memory id",
			memory: replayMemory(tenantID, updateID, "Hermes should preserve evidence-safe replay.", "fp_replay_update", createdAt),
			trace:  replayTrace(updateID, reasoningJobID, []string{"evt_replay_1"}, `[{"operation_id":"op_replay","kind":"update_memory","memory":{"text":"Hermes should preserve evidence-safe replay."}}]`, createdAt),
			edge:   replayUpdateEdge("mem_replay_other_target", createdAt),
		},
		{
			name:   "changed edge kind",
			memory: replayMemory(tenantID, updateID, "Hermes should preserve evidence-safe replay.", "fp_replay_update", createdAt),
			trace:  replayTrace(updateID, reasoningJobID, []string{"evt_replay_1"}, `[{"operation_id":"op_replay","kind":"update_memory","memory":{"text":"Hermes should preserve evidence-safe replay."}}]`, createdAt),
			edge: &core.MemoryEdge{
				FromMemoryID:   updateID,
				ToMemoryID:     targetID,
				EdgeKind:       core.EdgeKindExtends,
				Confidence:     0.91,
				CreatedByJobID: reasoningJobID,
				CreatedAt:      createdAt,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.CreateMemoryWithTraceAndUpdateEdge(ctx, tt.memory, tt.trace, tt.edge)
			if !errors.Is(err, core.ErrConflict) {
				t.Fatalf("changed replay evidence must return ErrConflict, got %v", err)
			}
		})
	}
}

func TestPostgresCreateMemoryReplayRejectsTraceEvidenceOverwrite(t *testing.T) {
	dbURL := os.Getenv("VIBEGRAVITY_DB_URL")
	if dbURL == "" {
		t.Skip("Skipping Postgres replay integration test because VIBEGRAVITY_DB_URL is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("connect test database: %v", err)
	}
	defer pool.Close()

	store := NewStore(pool)
	tenantID := fmt.Sprintf("tenant_replay_create_%d", time.Now().UnixNano())
	workspaceID := replayWorkspaceID
	memoryID := "mem_replay_create"
	reasoningJobID := "job_replay_create"
	createdAt := time.Now().UTC()

	cleanupPostgresReplayRows(ctx, t, pool, tenantID)
	defer cleanupPostgresReplayRows(context.Background(), t, pool, tenantID)

	mustSeedJob(ctx, t, pool, tenantID, workspaceID, reasoningJobID)
	memory := replayMemory(tenantID, memoryID, "Create retry keeps its original trace evidence.", "fp_replay_create", createdAt)
	trace := replayTrace(memoryID, reasoningJobID, []string{"evt_create_1"}, `[{"operation_id":"op_create","kind":"create_memory"}]`, createdAt)

	if err := store.CreateMemoryWithTrace(ctx, memory, trace); err != nil {
		t.Fatalf("initial create apply failed: %v", err)
	}
	changedTrace := replayTrace(memoryID, reasoningJobID, []string{"evt_create_2"}, `[{"operation_id":"op_create","kind":"create_memory","memory":{"text":"changed"}}]`, createdAt)
	err = store.CreateMemoryWithTrace(ctx, memory, changedTrace)
	if !errors.Is(err, core.ErrConflict) {
		t.Fatalf("changed create replay trace evidence must return ErrConflict, got %v", err)
	}
}

func TestPostgresReplaySourceContractsRequireFullEvidenceComparison(t *testing.T) {
	t.Parallel()

	source := readPostgresSourceFile(t, "memories.go")
	completedSource := extractPostgresSourceBetween(t, source, "func completedUpdateAlreadyApplied", "type updateTargetMemory")
	traceSource := extractPostgresSourceBetween(t, source, "func writeMemoryTrace", "func (s *Store) ExplainMemory")

	for _, want := range []string{
		"validateReplayEvidence(ctx, tx, memory, trace, edge)",
		"validateExistingMemoryEvidence(ctx, exec, memory)",
		"validateExistingTraceEvidence(ctx, exec, trace)",
		"validateExistingMemoryEdgeEvidence(ctx, exec, edge)",
		"existing.Text != memory.Text",
		"existing.Fingerprint != memory.Fingerprint",
		"!sameFloat(existing.Confidence, memory.Confidence)",
		"!sameStringSlice(existing.RawEventIDs, trace.RawEventIDs)",
		"existing.ReasoningJobID != trace.ReasoningJobID",
		"!sameJSON(existing.AppliedOperationsJSON, rawJSONOrEmpty(trace.AppliedOperationsJSON))",
		"!sameOperationIDs(existing.AppliedOperationsJSON, trace.AppliedOperationsJSON)",
		"existing.CreatedByJobID != edge.CreatedByJobID",
	} {
		if !strings.Contains(source, want) {
			t.Errorf("replay source must preserve %q before idempotent success", want)
		}
	}
	if strings.Contains(traceSource, "ON CONFLICT (memory_id) DO UPDATE") {
		t.Errorf("memory_trace replay must not silently overwrite provenance on memory_id conflict")
	}
	if t.Failed() {
		t.Logf("completedUpdateAlreadyApplied source:\n%s", completedSource)
		t.Logf("writeMemoryTrace source:\n%s", traceSource)
	}
}

func seedReplayTargetMemory(ctx context.Context, t *testing.T, store *Store, tenantID string, workspaceID string, ownerID string, targetID string, createdAt time.Time) {
	t.Helper()

	if err := store.CreateMemoryWithTrace(ctx, &core.Memory{
		ID:            targetID,
		TenantID:      tenantID,
		WorkspaceID:   workspaceID,
		Scope:         core.MemoryScopeWorkspaceShared,
		OwnerEntityID: ownerID,
		Kind:          core.MemoryKindFact,
		ArtifactClass: core.ArtifactClassKnowledge,
		Text:          "Original replay target memory.",
		Fingerprint:   "fp_replay_target",
		Confidence:    0.72,
		Status:        core.MemoryStatusActive,
		ValidFrom:     createdAt,
		LatestFlag:    true,
		MetadataJSON:  []byte(`{}`),
		CreatedAt:     createdAt,
		UpdatedAt:     createdAt,
	}, &core.MemoryTrace{
		MemoryID:              targetID,
		RawEventIDs:           []string{"evt_seed"},
		ReasoningJobID:        "job_replay_seed",
		ReasoningStage:        "resolve",
		CandidateSnapshotJSON: []byte(`{"seed":true}`),
		AppliedOperationsJSON: []byte(`[{"operation_id":"seed"}]`),
		RelatedDocumentIDs:    []string{},
		CreatedAt:             createdAt,
	}); err != nil {
		t.Fatalf("seed replay target memory: %v", err)
	}
}

func replayMemory(tenantID string, memoryID string, text string, fingerprint string, createdAt time.Time) *core.Memory {
	return &core.Memory{
		ID:            memoryID,
		TenantID:      tenantID,
		WorkspaceID:   replayWorkspaceID,
		Scope:         core.MemoryScopeWorkspaceShared,
		OwnerEntityID: replayOwnerID,
		Kind:          core.MemoryKindFact,
		ArtifactClass: core.ArtifactClassKnowledge,
		Text:          text,
		Fingerprint:   fingerprint,
		Confidence:    0.91,
		Status:        core.MemoryStatusActive,
		ValidFrom:     createdAt,
		LatestFlag:    true,
		MetadataJSON:  []byte(`{}`),
		CreatedAt:     createdAt,
		UpdatedAt:     createdAt,
	}
}

func replayTrace(memoryID string, reasoningJobID string, rawEventIDs []string, appliedOperations string, createdAt time.Time) *core.MemoryTrace {
	return &core.MemoryTrace{
		MemoryID:              memoryID,
		RawEventIDs:           rawEventIDs,
		ReasoningJobID:        reasoningJobID,
		ReasoningStage:        "resolve",
		CandidateSnapshotJSON: []byte(`{"candidate_memories":[]}`),
		AppliedOperationsJSON: []byte(appliedOperations),
		RelatedDocumentIDs:    []string{},
		CreatedAt:             createdAt,
	}
}

func replayUpdateEdge(targetID string, createdAt time.Time) *core.MemoryEdge {
	return &core.MemoryEdge{
		FromMemoryID:   replayUpdateID,
		ToMemoryID:     targetID,
		EdgeKind:       core.EdgeKindUpdates,
		Confidence:     0.91,
		CreatedByJobID: replayJobID,
		CreatedAt:      createdAt,
	}
}

func cleanupPostgresReplayRows(ctx context.Context, t testing.TB, pool *pgxpool.Pool, tenantID string) {
	t.Helper()

	if _, err := pool.Exec(ctx, `
		DELETE FROM memories
		WHERE tenant_id = $1 AND workspace_id = $2
	`, tenantID, replayWorkspaceID); err != nil {
		t.Fatalf("cleanup replay memories: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		DELETE FROM ingest_jobs
		WHERE tenant_id = $1 AND workspace_id = $2
	`, tenantID, replayWorkspaceID); err != nil {
		t.Fatalf("cleanup replay ingest jobs: %v", err)
	}
}
