// ============================================================
// FILE     : internal/graph/apply_test.go
// PURPOSE  : Verifies no-op apply validation for structured Stage 2 graph operations.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : TestNoopApplyEngine_ValidatesStructuredOperations, TestNoopApplyEngine_RejectsInvalidOperations
// DEPENDS  : context, encoding/json, errors, strings, testing, internal/core, internal/reasoning
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: These tests lock the apply boundary only; they must not assert memory quality or extraction behavior.
// ============================================================

package graph

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/reasoning"
)

func TestNoopApplyEngine_ValidatesStructuredOperations(t *testing.T) {
	t.Parallel()

	engine := NewNoopApplyEngine()
	result, err := engine.Apply(context.Background(), validApplyRequest(validCreateOperation()))
	if err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}

	if result.AppliedOperationCount != 0 {
		t.Fatalf("expected no-op engine to commit no operations, got %d", result.AppliedOperationCount)
	}
	if result.TraceWritten {
		t.Fatalf("expected no-op engine to write no trace")
	}
	if len(result.MemoryIDs) != 0 {
		t.Fatalf("expected no memory IDs, got %v", result.MemoryIDs)
	}
}

func TestNoopApplyEngine_RejectsInvalidOperations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		operation reasoning.GraphOperation
		wantError string
	}{
		{
			name: "empty operation kind",
			operation: func() reasoning.GraphOperation {
				operation := validCreateOperation()
				operation.Kind = ""
				return operation
			}(),
			wantError: "kind is required",
		},
		{
			name: "unsupported operation kind",
			operation: func() reasoning.GraphOperation {
				operation := validCreateOperation()
				operation.Kind = reasoning.OperationKind("merge_memory")
				return operation
			}(),
			wantError: "kind is unsupported",
		},
		{
			name: "missing scope",
			operation: func() reasoning.GraphOperation {
				operation := validCreateOperation()
				operation.Memory.Scope = ""
				return operation
			}(),
			wantError: "memory.scope is required",
		},
		{
			name: "invalid confidence",
			operation: func() reasoning.GraphOperation {
				operation := validCreateOperation()
				operation.Memory.Confidence = 1.01
				return operation
			}(),
			wantError: "memory.confidence",
		},
		{
			name: "update without target",
			operation: func() reasoning.GraphOperation {
				operation := validUpdateOperation()
				operation.Memory.TargetID = ""
				operation.Edge.ToMemoryID = ""
				return operation
			}(),
			wantError: "memory.target_id is required",
		},
		{
			name: "extend without target",
			operation: func() reasoning.GraphOperation {
				operation := validExtendOperation()
				operation.Memory.TargetID = ""
				operation.Edge.ToMemoryID = ""
				return operation
			}(),
			wantError: "memory.target_id is required",
		},
		{
			name: "unstructured operation metadata",
			operation: func() reasoning.GraphOperation {
				operation := validCreateOperation()
				operation.Metadata = json.RawMessage(`"free-form explanation"`)
				return operation
			}(),
			wantError: "metadata must be a JSON object",
		},
		{
			name: "update edge must use updates",
			operation: func() reasoning.GraphOperation {
				operation := validUpdateOperation()
				operation.Edge.EdgeKind = core.EdgeKindExtends
				return operation
			}(),
			wantError: `edge.edge_kind must be "updates"`,
		},
		{
			name: "extend edge must use extends",
			operation: func() reasoning.GraphOperation {
				operation := validExtendOperation()
				operation.Edge.EdgeKind = core.EdgeKindUpdates
				return operation
			}(),
			wantError: `edge.edge_kind must be "extends"`,
		},
		{
			name: "archive without target",
			operation: func() reasoning.GraphOperation {
				operation := validArchiveOperation()
				operation.Memory.TargetID = ""
				operation.Memory.MemoryID = ""
				return operation
			}(),
			wantError: "memory target is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewNoopApplyEngine().Apply(context.Background(), validApplyRequest(tt.operation))
			if err == nil {
				t.Fatalf("expected Apply to reject operation")
			}
			if !errors.Is(err, core.ErrInvalidArgument) {
				t.Fatalf("expected ErrInvalidArgument, got %v", err)
			}
			if !strings.Contains(err.Error(), tt.wantError) {
				t.Fatalf("expected error to contain %q, got %v", tt.wantError, err)
			}
		})
	}
}

func validApplyRequest(operation reasoning.GraphOperation) *ApplyRequest {
	return &ApplyRequest{
		JobID:       "job_1",
		TenantID:    "tenant_1",
		WorkspaceID: "workspace_1",
		RawEventIDs: []string{"evt_1"},
		Reasoning: &reasoning.ProcessTurnResult{
			Stage1: reasoning.Stage1Output{
				CandidateEntities: []reasoning.CandidateEntity{},
				CandidateMemories: []reasoning.CandidateMemory{},
			},
			Stage2: reasoning.Stage2Output{
				Operations:     []reasoning.GraphOperation{operation},
				ProfileDelta:   json.RawMessage(`{}`),
				SessionSummary: "",
				PlanDelta:      json.RawMessage(`{}`),
				Trace: reasoning.Trace{
					SchemaVersion: "v0",
					Stage:         reasoning.StageNameResolve,
					Codes:         []string{"test"},
					MetadataJSON:  json.RawMessage(`{}`),
				},
			},
		},
	}
}

func validCreateOperation() reasoning.GraphOperation {
	return reasoning.GraphOperation{
		OperationID: "op_1",
		Kind:        reasoning.OperationKindCreateMemory,
		Memory:      validMemoryMutation(),
		RawEventIDs: []string{"evt_1"},
		Metadata:    json.RawMessage(`{"source":"test"}`),
	}
}

func validUpdateOperation() reasoning.GraphOperation {
	operation := validCreateOperation()
	operation.OperationID = "op_update"
	operation.Kind = reasoning.OperationKindUpdateMemory
	operation.Memory.MemoryID = "mem_new"
	operation.Memory.TargetID = "mem_old"
	operation.Edge = &reasoning.EdgeMutation{
		FromMemoryID: "mem_new",
		ToMemoryID:   "mem_old",
		EdgeKind:     core.EdgeKindUpdates,
		Confidence:   0.8,
	}
	return operation
}

func validExtendOperation() reasoning.GraphOperation {
	operation := validCreateOperation()
	operation.OperationID = "op_extend"
	operation.Kind = reasoning.OperationKindExtendMemory
	operation.Memory.MemoryID = "mem_detail"
	operation.Memory.TargetID = "mem_base"
	operation.Edge = &reasoning.EdgeMutation{
		FromMemoryID: "mem_detail",
		ToMemoryID:   "mem_base",
		EdgeKind:     core.EdgeKindExtends,
		Confidence:   0.8,
	}
	return operation
}

func validArchiveOperation() reasoning.GraphOperation {
	operation := validCreateOperation()
	operation.OperationID = "op_archive"
	operation.Kind = reasoning.OperationKindArchiveMemory
	operation.Memory = &reasoning.MemoryMutation{
		TargetID:      "mem_old",
		Scope:         core.MemoryScopeWorkspaceShared,
		OwnerEntityID: "agent:hermes-main",
		Confidence:    0.8,
		MetadataJSON:  json.RawMessage(`{"reason":"stale"}`),
	}
	return operation
}

func validMemoryMutation() *reasoning.MemoryMutation {
	return &reasoning.MemoryMutation{
		Kind:          core.MemoryKindDecision,
		ArtifactClass: core.ArtifactClassKnowledge,
		Scope:         core.MemoryScopeWorkspaceShared,
		OwnerEntityID: "agent:hermes-main",
		Text:          "VibeGravity keeps Stage 2 operations structured.",
		Confidence:    0.8,
		MetadataJSON:  json.RawMessage(`{"source":"test"}`),
	}
}
