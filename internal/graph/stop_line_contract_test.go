// ============================================================
// FILE     : internal/graph/stop_line_contract_test.go
// PURPOSE  : Guards graph write stop-lines that must remain closed until validation exists.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : graph stop-line tests
// DEPENDS  : context, encoding/json, errors, strings, testing, internal/core, internal/graph, internal/reasoning
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: Group-shared graph writes must stay blocked until membership validation is part of the write transaction.
// ============================================================

package graph_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/graph"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/reasoning"
)

func TestStoreBackedApplyRejectsGroupSharedWritesUntilMembershipValidation(t *testing.T) {
	t.Parallel()

	memories := &stopLineMemoryStore{}
	engine, err := graph.NewStoreBackedApplyEngine(memories)
	if err != nil {
		t.Fatalf("NewStoreBackedApplyEngine returned error: %v", err)
	}

	_, err = engine.Apply(context.Background(), groupSharedApplyRequest())
	if err == nil {
		t.Fatalf("expected group_shared graph write to be rejected")
	}
	if !errors.Is(err, core.ErrNotImplemented) {
		t.Fatalf("expected ErrNotImplemented until membership validation exists, got %v", err)
	}
	if !strings.Contains(err.Error(), "membership validation") {
		t.Fatalf("expected membership validation stop-line error, got %v", err)
	}
	if memories.writeCount != 0 {
		t.Fatalf("group_shared rejection must happen before storage writes, got %d writes", memories.writeCount)
	}
}

func groupSharedApplyRequest() *graph.ApplyRequest {
	groupID := "group_1"
	return &graph.ApplyRequest{
		JobID:       "job_stop_line",
		TenantID:    "tenant_1",
		WorkspaceID: "workspace_1",
		RawEventIDs: []string{"evt_stop_line_1"},
		Reasoning: &reasoning.ProcessTurnResult{
			Stage1: reasoning.Stage1Output{
				CandidateEntities: []reasoning.CandidateEntity{},
				CandidateMemories: []reasoning.CandidateMemory{},
			},
			Stage2: reasoning.Stage2Output{
				Operations: []reasoning.GraphOperation{{
					OperationID: "op_group_shared_create",
					Kind:        reasoning.OperationKindCreateMemory,
					RawEventIDs: []string{"evt_stop_line_1"},
					Metadata:    json.RawMessage(`{}`),
					Memory: &reasoning.MemoryMutation{
						Kind:          core.MemoryKindConstraint,
						ArtifactClass: core.ArtifactClassKnowledge,
						Scope:         core.MemoryScopeGroupShared,
						GroupID:       &groupID,
						OwnerEntityID: "agent:hermes-main",
						Text:          "Group-shared writes need membership validation.",
						Confidence:    0.9,
						MetadataJSON:  json.RawMessage(`{}`),
					},
				}},
				ProfileDelta:   json.RawMessage(`{}`),
				SessionSummary: "",
				PlanDelta:      json.RawMessage(`{}`),
				Trace: reasoning.Trace{
					SchemaVersion: "v0",
					Stage:         reasoning.StageNameResolve,
					Codes:         []string{"stop_line_group_shared_write"},
					MetadataJSON:  json.RawMessage(`{}`),
				},
			},
		},
	}
}

type stopLineMemoryStore struct {
	writeCount int
}

func (s *stopLineMemoryStore) CreateMemoryWithTrace(context.Context, *core.Memory, *core.MemoryTrace) error {
	s.writeCount++
	return nil
}

func (s *stopLineMemoryStore) CreateMemoryWithTraceAndEdge(context.Context, *core.Memory, *core.MemoryTrace, *core.MemoryEdge) error {
	s.writeCount++
	return nil
}

func (s *stopLineMemoryStore) CreateMemoryWithTraceAndUpdateEdge(context.Context, *core.Memory, *core.MemoryTrace, *core.MemoryEdge) error {
	s.writeCount++
	return nil
}
