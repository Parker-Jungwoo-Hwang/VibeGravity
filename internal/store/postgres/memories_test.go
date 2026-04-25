// ============================================================
// FILE     : internal/store/postgres/memories_test.go
// PURPOSE  : Verifies PostgreSQL memory graph helper contracts without a live database.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : postgres memory helper tests
// DEPENDS  : context, errors, reflect, strings, testing, time, internal/core, pgx
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: These tests lock update_memory supersession guards before live DB integration tests exist.
// ============================================================

package postgres

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
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

func TestMarkMemoryCorrectionAppliedOnlyUpdatesRecordedOrAppliedRows(t *testing.T) {
	t.Parallel()

	exec := &recordingMemoryExecutor{tag: pgconn.NewCommandTag("UPDATE 1")}

	if err := markMemoryCorrectionApplied(context.Background(), exec, "corr_1"); err != nil {
		t.Fatalf("markMemoryCorrectionApplied returned error: %v", err)
	}
	if !strings.Contains(exec.sql, "SET status = 'applied'") || !strings.Contains(exec.sql, "status IN ('recorded', 'applied')") {
		t.Fatalf("expected correction applied status guard, got: %s", exec.sql)
	}
	if len(exec.args) != 1 || exec.args[0] != "corr_1" {
		t.Fatalf("unexpected correction status args: %#v", exec.args)
	}
}

func TestMarkMemoryCorrectionAppliedReturnsConflictWhenNoArtifactChanged(t *testing.T) {
	t.Parallel()

	exec := &recordingMemoryExecutor{tag: pgconn.NewCommandTag("UPDATE 0")}

	err := markMemoryCorrectionApplied(context.Background(), exec, "corr_missing")
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

func TestValidateReplayEvidenceAcceptsIdenticalMemoryTraceAndEdge(t *testing.T) {
	t.Parallel()

	memory, trace, edge := replayEvidenceFixture()
	exec := replayEvidenceExecutorFor(memory, trace, edge)

	if err := validateReplayEvidence(context.Background(), exec, memory, trace, edge); err != nil {
		t.Fatalf("validateReplayEvidence returned error: %v", err)
	}
}

func TestValidateReplayEvidenceRejectsChangedSemanticEvidence(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		mutate     func(memory *core.Memory, trace *core.MemoryTrace, edge *core.MemoryEdge)
		missingRow int
	}{
		{
			name: "replacement text",
			mutate: func(memory *core.Memory, _ *core.MemoryTrace, _ *core.MemoryEdge) {
				memory.Text = "Changed replacement text."
			},
		},
		{
			name: "fingerprint",
			mutate: func(memory *core.Memory, _ *core.MemoryTrace, _ *core.MemoryEdge) {
				memory.Fingerprint = "fp_changed"
			},
		},
		{
			name: "confidence",
			mutate: func(memory *core.Memory, _ *core.MemoryTrace, _ *core.MemoryEdge) {
				memory.Confidence = 0.91
			},
		},
		{
			name: "reasoning job",
			mutate: func(_ *core.Memory, trace *core.MemoryTrace, _ *core.MemoryEdge) {
				trace.ReasoningJobID = "job_changed"
			},
		},
		{
			name: "raw event ids",
			mutate: func(_ *core.Memory, trace *core.MemoryTrace, _ *core.MemoryEdge) {
				trace.RawEventIDs = []string{"evt_1", "evt_changed"}
			},
		},
		{
			name: "applied operation json",
			mutate: func(_ *core.Memory, trace *core.MemoryTrace, _ *core.MemoryEdge) {
				trace.AppliedOperationsJSON = []byte(`[{"operation_id":"op_update","memory":{"text":"changed"}}]`)
			},
		},
		{
			name: "operation id",
			mutate: func(_ *core.Memory, trace *core.MemoryTrace, _ *core.MemoryEdge) {
				trace.AppliedOperationsJSON = []byte(`[{"operation_id":"op_changed","memory":{"text":"Replacement memory."}}]`)
			},
		},
		{
			name: "target memory id",
			mutate: func(_ *core.Memory, _ *core.MemoryTrace, edge *core.MemoryEdge) {
				edge.ToMemoryID = "mem_other_target"
			},
			missingRow: 2,
		},
		{
			name: "edge kind",
			mutate: func(_ *core.Memory, _ *core.MemoryTrace, edge *core.MemoryEdge) {
				edge.EdgeKind = core.EdgeKindExtends
			},
			missingRow: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			existingMemory, existingTrace, existingEdge := replayEvidenceFixture()
			attemptMemory, attemptTrace, attemptEdge := replayEvidenceFixture()
			tt.mutate(attemptMemory, attemptTrace, attemptEdge)
			exec := replayEvidenceExecutorFor(existingMemory, existingTrace, existingEdge)
			if tt.missingRow > 0 {
				exec.rows[tt.missingRow] = fakeReplayRow{err: pgx.ErrNoRows}
			}

			err := validateReplayEvidence(context.Background(), exec, attemptMemory, attemptTrace, attemptEdge)
			if !errors.Is(err, core.ErrConflict) {
				t.Fatalf("expected ErrConflict, got %v", err)
			}
		})
	}
}

func TestWriteMemoryTracePreservesExistingTraceOnConflict(t *testing.T) {
	t.Parallel()

	source := readPostgresSourceFile(t, "memories.go")
	traceSource := extractPostgresSourceBetween(t, source, "func writeMemoryTrace", "func validateReplayEvidence")

	for _, want := range []string{
		"ON CONFLICT (memory_id) DO NOTHING",
		"validateExistingTraceEvidence(ctx, exec, trace)",
	} {
		if !strings.Contains(traceSource, want) {
			t.Fatalf("writeMemoryTrace must preserve existing evidence on replay; missing %q in:\n%s", want, traceSource)
		}
	}
	if strings.Contains(traceSource, "DO UPDATE") {
		t.Fatalf("writeMemoryTrace must not overwrite existing trace evidence, got:\n%s", traceSource)
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

type replayEvidenceExecutor struct {
	rows []fakeReplayRow
}

func replayEvidenceExecutorFor(memory *core.Memory, trace *core.MemoryTrace, edge *core.MemoryEdge) *replayEvidenceExecutor {
	return &replayEvidenceExecutor{
		rows: []fakeReplayRow{
			{values: []any{
				memory.TenantID,
				memory.WorkspaceID,
				memory.Scope,
				memory.GroupID,
				memory.OwnerEntityID,
				memory.Kind,
				memory.ArtifactClass,
				memory.Text,
				memory.Fingerprint,
				memory.Confidence,
				memory.Status,
				memory.LatestFlag,
				memory.MetadataJSON,
			}},
			{values: []any{
				trace.RawEventIDs,
				trace.ReasoningJobID,
				trace.ReasoningStage,
				trace.CandidateSnapshotJSON,
				trace.AppliedOperationsJSON,
				trace.OperatorCorrectionFlag,
				trace.RelatedDocumentIDs,
			}},
			{values: []any{
				edge.Confidence,
				edge.CreatedByJobID,
			}},
		},
	}
}

func (e *replayEvidenceExecutor) Exec(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
	return pgconn.NewCommandTag("INSERT 0 0"), nil
}

func (e *replayEvidenceExecutor) QueryRow(_ context.Context, _ string, _ ...any) pgx.Row {
	if len(e.rows) == 0 {
		return fakeReplayRow{err: pgx.ErrNoRows}
	}
	row := e.rows[0]
	e.rows = e.rows[1:]
	return row
}

type fakeReplayRow struct {
	values []any
	err    error
}

func (r fakeReplayRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	if len(dest) != len(r.values) {
		return errors.New("fake row destination count mismatch")
	}
	for i := range dest {
		target := reflect.ValueOf(dest[i])
		if target.Kind() != reflect.Pointer || target.IsNil() {
			return errors.New("fake row destination must be pointer")
		}
		slot := target.Elem()
		if r.values[i] == nil {
			slot.Set(reflect.Zero(slot.Type()))
			continue
		}
		value := reflect.ValueOf(r.values[i])
		if value.Type().AssignableTo(slot.Type()) {
			slot.Set(value)
			continue
		}
		if value.Type().ConvertibleTo(slot.Type()) {
			slot.Set(value.Convert(slot.Type()))
			continue
		}
		return errors.New("fake row value type mismatch")
	}
	return nil
}

func replayEvidenceFixture() (*core.Memory, *core.MemoryTrace, *core.MemoryEdge) {
	memory := &core.Memory{
		ID:            "mem_replay_update",
		TenantID:      "tenant_1",
		WorkspaceID:   "workspace_1",
		Scope:         core.MemoryScopeWorkspaceShared,
		OwnerEntityID: "agent:hermes-main",
		Kind:          core.MemoryKindFact,
		ArtifactClass: core.ArtifactClassKnowledge,
		Text:          "Replacement memory.",
		Fingerprint:   "fp_replay_update",
		Confidence:    0.8,
		Status:        core.MemoryStatusActive,
		LatestFlag:    true,
		MetadataJSON:  []byte(`{"source":"test"}`),
	}
	trace := &core.MemoryTrace{
		MemoryID:              memory.ID,
		RawEventIDs:           []string{"evt_1", "evt_2"},
		ReasoningJobID:        "job_replay",
		ReasoningStage:        "resolve",
		CandidateSnapshotJSON: []byte(`{"candidate_memories":[]}`),
		AppliedOperationsJSON: []byte(`[{"operation_id":"op_update","memory":{"text":"Replacement memory."}}]`),
		RelatedDocumentIDs:    []string{"doc_1"},
	}
	edge := &core.MemoryEdge{
		FromMemoryID:   memory.ID,
		ToMemoryID:     "mem_target",
		EdgeKind:       core.EdgeKindUpdates,
		Confidence:     0.8,
		CreatedByJobID: "job_replay",
	}
	return memory, trace, edge
}
