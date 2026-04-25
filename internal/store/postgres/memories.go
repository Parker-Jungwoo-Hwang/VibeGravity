// ============================================================
// FILE     : internal/store/postgres/memories.go
// PURPOSE  : Implements PostgreSQL persistence for memories, edges, and traces.
// LAYER    : infra
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : CreateMemoryWithTrace, CreateMemoryWithTraceAndEdge, CreateMemoryWithTraceAndUpdateEdge, UpsertMemory, GetMemory, UpsertMemoryEdge, WriteMemoryTrace, ExplainMemory
// DEPENDS  : bytes, context, encoding/json, errors, fmt, math, time, internal/core, pgx
// USED_BY  : graph apply engine, recall assembler, provenance APIs
// ------------------------------------------------------------
// AGENT_NOTE: Memory writes must preserve explicit scope and mandatory provenance.
// ============================================================

package postgres

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

type memoryExecutor interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

type memoryReadWriter interface {
	memoryExecutor
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
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

	if err := insertMemoryOrValidateReplay(ctx, tx, memory); err != nil {
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

	if err := insertMemoryOrValidateReplay(ctx, tx, memory); err != nil {
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
	if err := insertMemoryEdgeOrValidateReplay(ctx, tx, edge); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit memory trace edge transaction: %w", err)
	}
	return nil
}

// CreateMemoryWithTraceAndUpdateEdge writes a new memory, provenance, updates edge, and target supersession in one transaction.
func (s *Store) CreateMemoryWithTraceAndUpdateEdge(ctx context.Context, memory *core.Memory, trace *core.MemoryTrace, edge *core.MemoryEdge) error {
	return s.createMemoryWithTraceAndUpdateEdge(ctx, memory, trace, edge, "")
}

// CreateCorrectionSupersession writes a correction replacement and marks its correction artifact applied atomically.
func (s *Store) CreateCorrectionSupersession(ctx context.Context, memory *core.Memory, trace *core.MemoryTrace, edge *core.MemoryEdge, correctionID string) error {
	if correctionID == "" {
		return fmt.Errorf("%w: correction id is required", core.ErrInvalidArgument)
	}
	return s.createMemoryWithTraceAndUpdateEdge(ctx, memory, trace, edge, correctionID)
}

func (s *Store) createMemoryWithTraceAndUpdateEdge(ctx context.Context, memory *core.Memory, trace *core.MemoryTrace, edge *core.MemoryEdge, correctionID string) error {
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
	completed, err := completedUpdateAlreadyApplied(ctx, tx, memory, trace, edge)
	if err != nil {
		return err
	}
	if completed {
		if err := markMemoryCorrectionApplied(ctx, tx, correctionID); err != nil {
			return err
		}
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
	if err := insertMemoryOrValidateReplay(ctx, tx, memory); err != nil {
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
	if err := insertMemoryEdgeOrValidateReplay(ctx, tx, edge); err != nil {
		return err
	}
	if err := supersedeMemoryTarget(ctx, tx, edge.ToMemoryID, timeOrNow(memory.ValidFrom)); err != nil {
		return err
	}
	if err := markMemoryCorrectionApplied(ctx, tx, correctionID); err != nil {
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

func completedUpdateAlreadyApplied(ctx context.Context, tx pgx.Tx, memory *core.Memory, trace *core.MemoryTrace, edge *core.MemoryEdge) (bool, error) {
	var existing struct {
		TraceMemoryID *string
		EdgeFromID    *string
	}
	err := tx.QueryRow(ctx, `
		SELECT mt.memory_id, me.from_memory_id
		FROM memories m
		LEFT JOIN memory_trace mt ON mt.memory_id = m.id
		LEFT JOIN memory_edges me
		  ON me.from_memory_id = m.id
		 AND me.to_memory_id = $2
		 AND me.edge_kind = $3
		WHERE m.id = $1
		FOR UPDATE
	`, memory.ID, edge.ToMemoryID, core.EdgeKindUpdates).Scan(&existing.TraceMemoryID, &existing.EdgeFromID)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("check idempotent update memory: %w", err)
	}
	if existing.TraceMemoryID == nil || existing.EdgeFromID == nil {
		return false, fmt.Errorf("%w: existing update memory is incomplete", core.ErrConflict)
	}
	if err := validateReplayEvidence(ctx, tx, memory, trace, edge); err != nil {
		return false, err
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

func markMemoryCorrectionApplied(ctx context.Context, exec memoryExecutor, correctionID string) error {
	if correctionID == "" {
		return nil
	}
	tag, err := exec.Exec(ctx, `
		UPDATE memory_corrections
		SET status = 'applied'
		WHERE id = $1
		  AND status IN ('recorded', 'applied')
	`, correctionID)
	if err != nil {
		return fmt.Errorf("mark memory correction applied: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%w: correction artifact must exist before apply", core.ErrConflict)
	}
	return nil
}

func sameOptionalString(left *string, right *string) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func insertMemoryOrValidateReplay(ctx context.Context, exec memoryReadWriter, memory *core.Memory) error {
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
	tag, err := exec.Exec(ctx, `
		INSERT INTO memories (
			id, tenant_id, workspace_id, scope, group_id, owner_entity_id,
			kind, artifact_class, text, fingerprint, confidence, status,
			valid_from, valid_to, latest_flag, metadata_json,
			embedding_model, embedding_dims, embedding_updated_at, created_at, updated_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21)
		ON CONFLICT (id) DO NOTHING
	`, memory.ID, memory.TenantID, memory.WorkspaceID, memory.Scope, memory.GroupID,
		memory.OwnerEntityID, memory.Kind, memory.ArtifactClass, memory.Text, memory.Fingerprint,
		memory.Confidence, valueOr(string(memory.Status), string(core.MemoryStatusActive)),
		timeOrNow(memory.ValidFrom), memory.ValidTo, memory.LatestFlag, rawJSONOrEmpty(memory.MetadataJSON),
		valueOr(memory.EmbeddingModel, "pending"), memory.EmbeddingDims, memory.EmbeddingUpdatedAt,
		timeOrNow(memory.CreatedAt), timeOrNow(memory.UpdatedAt))
	if err != nil {
		return fmt.Errorf("insert memory: %w", err)
	}
	if tag.RowsAffected() == 1 {
		return nil
	}
	if err := validateExistingMemoryEvidence(ctx, exec, memory); err != nil {
		return err
	}
	return nil
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

func insertMemoryEdgeOrValidateReplay(ctx context.Context, exec memoryReadWriter, edge *core.MemoryEdge) error {
	if edge == nil {
		return fmt.Errorf("%w: memory edge is required", core.ErrInvalidArgument)
	}
	tag, err := exec.Exec(ctx, `
		INSERT INTO memory_edges (
			from_memory_id, to_memory_id, edge_kind, confidence, created_by_job_id, created_at
		)
		VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (from_memory_id, to_memory_id, edge_kind) DO NOTHING
	`, edge.FromMemoryID, edge.ToMemoryID, edge.EdgeKind, edge.Confidence,
		edge.CreatedByJobID, timeOrNow(edge.CreatedAt))
	if err != nil {
		return fmt.Errorf("insert memory edge: %w", err)
	}
	if tag.RowsAffected() == 1 {
		return nil
	}
	return validateExistingMemoryEdgeEvidence(ctx, exec, edge)
}

// WriteMemoryTrace writes mandatory provenance for a memory.
func (s *Store) WriteMemoryTrace(ctx context.Context, trace *core.MemoryTrace) error {
	return writeMemoryTrace(ctx, s.pool, trace)
}

func writeMemoryTrace(ctx context.Context, exec memoryReadWriter, trace *core.MemoryTrace) error {
	if trace == nil {
		return fmt.Errorf("%w: memory trace is required", core.ErrInvalidArgument)
	}
	tag, err := exec.Exec(ctx, `
		INSERT INTO memory_trace (
			memory_id, raw_event_ids, reasoning_job_id, reasoning_stage,
			candidate_snapshot_json, applied_operations_json,
			operator_correction_flag, related_document_ids, created_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (memory_id) DO NOTHING
	`, trace.MemoryID, trace.RawEventIDs, nullIfEmpty(trace.ReasoningJobID), trace.ReasoningStage,
		rawJSONOrEmpty(trace.CandidateSnapshotJSON), rawJSONOrEmpty(trace.AppliedOperationsJSON),
		trace.OperatorCorrectionFlag, trace.RelatedDocumentIDs, timeOrNow(trace.CreatedAt))
	if err != nil {
		return fmt.Errorf("write memory trace: %w", err)
	}
	if tag.RowsAffected() == 1 {
		return nil
	}
	if err := validateExistingTraceEvidence(ctx, exec, trace); err != nil {
		return err
	}
	return nil
}

func validateReplayEvidence(ctx context.Context, exec memoryReadWriter, memory *core.Memory, trace *core.MemoryTrace, edge *core.MemoryEdge) error {
	if err := validateExistingMemoryEvidence(ctx, exec, memory); err != nil {
		return err
	}
	if err := validateExistingTraceEvidence(ctx, exec, trace); err != nil {
		return err
	}
	if err := validateExistingMemoryEdgeEvidence(ctx, exec, edge); err != nil {
		return err
	}
	return nil
}

func validateExistingMemoryEvidence(ctx context.Context, exec memoryReadWriter, memory *core.Memory) error {
	var existing core.Memory
	err := exec.QueryRow(ctx, `
		SELECT tenant_id, workspace_id, scope, group_id, owner_entity_id,
		       kind, artifact_class, text, fingerprint, confidence, status,
		       latest_flag, metadata_json
		FROM memories
		WHERE id = $1
		FOR UPDATE
	`, memory.ID).Scan(&existing.TenantID, &existing.WorkspaceID, &existing.Scope,
		&existing.GroupID, &existing.OwnerEntityID, &existing.Kind, &existing.ArtifactClass,
		&existing.Text, &existing.Fingerprint, &existing.Confidence, &existing.Status,
		&existing.LatestFlag, &existing.MetadataJSON)
	if errors.Is(err, pgx.ErrNoRows) {
		return core.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("load existing memory replay evidence: %w", err)
	}
	if existing.TenantID != memory.TenantID ||
		existing.WorkspaceID != memory.WorkspaceID ||
		existing.Scope != memory.Scope ||
		!sameOptionalString(existing.GroupID, memory.GroupID) ||
		existing.OwnerEntityID != memory.OwnerEntityID ||
		existing.Kind != memory.Kind ||
		existing.ArtifactClass != memory.ArtifactClass ||
		existing.Text != memory.Text ||
		existing.Fingerprint != memory.Fingerprint ||
		!sameFloat(existing.Confidence, memory.Confidence) ||
		existing.Status != valueMemoryStatus(memory.Status) ||
		existing.LatestFlag != memory.LatestFlag ||
		!sameJSON(existing.MetadataJSON, rawJSONOrEmpty(memory.MetadataJSON)) {
		return fmt.Errorf("%w: existing memory replay evidence differs", core.ErrConflict)
	}
	return nil
}

func validateExistingTraceEvidence(ctx context.Context, exec memoryReadWriter, trace *core.MemoryTrace) error {
	var existing core.MemoryTrace
	err := exec.QueryRow(ctx, `
		SELECT raw_event_ids, COALESCE(reasoning_job_id, ''), reasoning_stage,
		       candidate_snapshot_json, applied_operations_json,
		       operator_correction_flag, related_document_ids
		FROM memory_trace
		WHERE memory_id = $1
		FOR UPDATE
	`, trace.MemoryID).Scan(&existing.RawEventIDs, &existing.ReasoningJobID,
		&existing.ReasoningStage, &existing.CandidateSnapshotJSON,
		&existing.AppliedOperationsJSON, &existing.OperatorCorrectionFlag,
		&existing.RelatedDocumentIDs)
	if errors.Is(err, pgx.ErrNoRows) {
		return core.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("load existing trace replay evidence: %w", err)
	}
	if !sameStringSlice(existing.RawEventIDs, trace.RawEventIDs) ||
		existing.ReasoningJobID != trace.ReasoningJobID ||
		existing.ReasoningStage != trace.ReasoningStage ||
		!sameJSON(existing.CandidateSnapshotJSON, rawJSONOrEmpty(trace.CandidateSnapshotJSON)) ||
		!sameJSON(existing.AppliedOperationsJSON, rawJSONOrEmpty(trace.AppliedOperationsJSON)) ||
		!sameOperationIDs(existing.AppliedOperationsJSON, trace.AppliedOperationsJSON) ||
		existing.OperatorCorrectionFlag != trace.OperatorCorrectionFlag ||
		!sameStringSlice(existing.RelatedDocumentIDs, trace.RelatedDocumentIDs) {
		return fmt.Errorf("%w: existing memory trace replay evidence differs", core.ErrConflict)
	}
	return nil
}

func validateExistingMemoryEdgeEvidence(ctx context.Context, exec memoryReadWriter, edge *core.MemoryEdge) error {
	var existing core.MemoryEdge
	err := exec.QueryRow(ctx, `
		SELECT confidence, created_by_job_id
		FROM memory_edges
		WHERE from_memory_id = $1
		  AND to_memory_id = $2
		  AND edge_kind = $3
	`, edge.FromMemoryID, edge.ToMemoryID, edge.EdgeKind).Scan(&existing.Confidence, &existing.CreatedByJobID)
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%w: replay memory edge evidence is missing", core.ErrConflict)
	}
	if err != nil {
		return fmt.Errorf("load existing edge replay evidence: %w", err)
	}
	if !sameFloat(existing.Confidence, edge.Confidence) || existing.CreatedByJobID != edge.CreatedByJobID {
		return fmt.Errorf("%w: existing memory edge replay evidence differs", core.ErrConflict)
	}
	return nil
}

func valueMemoryStatus(status core.MemoryStatus) core.MemoryStatus {
	if status == "" {
		return core.MemoryStatusActive
	}
	return status
}

func sameFloat(left float64, right float64) bool {
	return math.Abs(left-right) < 0.000000001
}

func sameStringSlice(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func sameJSON(left []byte, right []byte) bool {
	leftCanonical, leftOK := canonicalJSON(left)
	rightCanonical, rightOK := canonicalJSON(right)
	if !leftOK || !rightOK {
		return bytes.Equal(bytes.TrimSpace(left), bytes.TrimSpace(right))
	}
	return bytes.Equal(leftCanonical, rightCanonical)
}

func canonicalJSON(data []byte) ([]byte, bool) {
	if len(bytes.TrimSpace(data)) == 0 {
		data = []byte(`{}`)
	}
	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		return nil, false
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, false
	}
	return encoded, true
}

func sameOperationIDs(left []byte, right []byte) bool {
	leftIDs, leftOK := operationIDs(left)
	rightIDs, rightOK := operationIDs(right)
	if !leftOK || !rightOK {
		return false
	}
	return sameStringSlice(leftIDs, rightIDs)
}

func operationIDs(data []byte) ([]string, bool) {
	var operations []struct {
		OperationID string `json:"operation_id"`
	}
	if err := json.Unmarshal(data, &operations); err != nil {
		return nil, false
	}
	ids := make([]string, 0, len(operations))
	for _, operation := range operations {
		if operation.OperationID == "" {
			return nil, false
		}
		ids = append(ids, operation.OperationID)
	}
	return ids, true
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
