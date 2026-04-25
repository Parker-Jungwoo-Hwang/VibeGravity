// ============================================================
// FILE     : internal/graph/store_apply_test.go
// PURPOSE  : Verifies store-backed apply writes safe memory graph rows with mandatory trace provenance.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : TestStoreBackedApplyEngine_WritesCreateMemoryWithTrace, TestStoreBackedApplyEngine_WritesExtendMemoryWithTraceAndEdge, TestStoreBackedApplyEngine_WritesUpdateMemoryWithTraceAndSupersessionEdge, TestStoreBackedApplyEngine_RejectsUnsupportedWrites
// DEPENDS  : context, encoding/json, errors, strings, testing, time, internal/core, internal/reasoning
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: These tests prove provenance is part of the success boundary for derived memory writes.
// ============================================================

package graph

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/reasoning"
)

func TestStoreBackedApplyEngine_WritesCreateMemoryWithTrace(t *testing.T) {
	t.Parallel()

	memories := &fakeMemoryTraceCreator{}
	engine := newTestStoreBackedApplyEngine(t, memories)

	result, err := engine.Apply(context.Background(), validApplyRequest(validCreateOperation()))
	if err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}

	if result.AppliedOperationCount != 1 {
		t.Fatalf("unexpected applied count: got %d want 1", result.AppliedOperationCount)
	}
	if !result.TraceWritten {
		t.Fatalf("expected trace to be written")
	}
	if len(result.MemoryIDs) != 1 || !strings.HasPrefix(result.MemoryIDs[0], "mem_") {
		t.Fatalf("unexpected memory IDs: %v", result.MemoryIDs)
	}
	if len(memories.memories) != 1 || len(memories.traces) != 1 {
		t.Fatalf("expected one memory and one trace, got memories=%d traces=%d", len(memories.memories), len(memories.traces))
	}

	memory := memories.memories[0]
	if memory.ID != result.MemoryIDs[0] {
		t.Fatalf("result memory ID did not match written memory: got %q want %q", result.MemoryIDs[0], memory.ID)
	}
	if memory.TenantID != "tenant_1" || memory.WorkspaceID != "workspace_1" {
		t.Fatalf("memory tenant/workspace not copied from apply request: %#v", memory)
	}
	if memory.Scope != core.MemoryScopeWorkspaceShared {
		t.Fatalf("memory scope not preserved: %q", memory.Scope)
	}
	if memory.Status != core.MemoryStatusActive || !memory.LatestFlag {
		t.Fatalf("memory should start active latest, got status=%q latest=%v", memory.Status, memory.LatestFlag)
	}
	if memory.Fingerprint == "" || !strings.HasPrefix(memory.Fingerprint, "fp_") {
		t.Fatalf("expected stable fingerprint, got %q", memory.Fingerprint)
	}
	if !memory.CreatedAt.Equal(fixedApplyTime) || !memory.UpdatedAt.Equal(fixedApplyTime) || !memory.ValidFrom.Equal(fixedApplyTime) {
		t.Fatalf("memory timestamps should use apply clock: %#v", memory)
	}

	trace := memories.traces[0]
	if trace.MemoryID != memory.ID {
		t.Fatalf("trace memory ID mismatch: got %q want %q", trace.MemoryID, memory.ID)
	}
	if trace.ReasoningJobID != "job_1" || trace.ReasoningStage != string(reasoning.StageNameResolve) {
		t.Fatalf("trace reasoning provenance mismatch: %#v", trace)
	}
	if len(trace.RawEventIDs) != 1 || trace.RawEventIDs[0] != "evt_1" {
		t.Fatalf("trace raw event provenance mismatch: %#v", trace.RawEventIDs)
	}
	assertJSONContains(t, trace.CandidateSnapshotJSON, "candidate_memories")
	assertJSONContains(t, trace.AppliedOperationsJSON, "op_1")
}

func TestStoreBackedApplyEngine_WritesExtendMemoryWithTraceAndEdge(t *testing.T) {
	t.Parallel()

	memories := &fakeMemoryTraceCreator{}
	engine := newTestStoreBackedApplyEngine(t, memories)
	operation := validExtendOperation()

	result, err := engine.Apply(context.Background(), validApplyRequest(operation))
	if err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}

	if result.AppliedOperationCount != 1 {
		t.Fatalf("unexpected applied count: got %d want 1", result.AppliedOperationCount)
	}
	if !result.TraceWritten {
		t.Fatalf("expected trace to be written")
	}
	if len(memories.memories) != 1 || len(memories.traces) != 1 || len(memories.edges) != 1 {
		t.Fatalf("expected one memory, trace, and edge; got memories=%d traces=%d edges=%d", len(memories.memories), len(memories.traces), len(memories.edges))
	}

	memory := memories.memories[0]
	if memory.ID == "" || !strings.HasPrefix(memory.ID, "mem_") {
		t.Fatalf("expected deterministic memory ID, got %q", memory.ID)
	}
	if memory.Status != core.MemoryStatusActive || !memory.LatestFlag {
		t.Fatalf("extension memory should start active latest without changing the target, got status=%q latest=%v", memory.Status, memory.LatestFlag)
	}
	if memory.Text != operation.Memory.Text || memory.Scope != operation.Memory.Scope || memory.OwnerEntityID != operation.Memory.OwnerEntityID {
		t.Fatalf("extension memory did not preserve mutation fields: %#v", memory)
	}

	trace := memories.traces[0]
	if trace.MemoryID != memory.ID {
		t.Fatalf("trace memory ID mismatch: got %q want %q", trace.MemoryID, memory.ID)
	}
	if len(trace.RawEventIDs) != 1 || trace.RawEventIDs[0] != "evt_1" {
		t.Fatalf("trace raw event provenance mismatch: %#v", trace.RawEventIDs)
	}
	assertJSONContains(t, trace.AppliedOperationsJSON, "op_extend")

	edge := memories.edges[0]
	if edge.FromMemoryID != memory.ID {
		t.Fatalf("extension edge must originate from the written memory: got %q want %q", edge.FromMemoryID, memory.ID)
	}
	if edge.ToMemoryID != operation.Memory.TargetID {
		t.Fatalf("extension edge target mismatch: got %q want %q", edge.ToMemoryID, operation.Memory.TargetID)
	}
	if edge.EdgeKind != core.EdgeKindExtends {
		t.Fatalf("extension edge kind mismatch: got %q", edge.EdgeKind)
	}
	if edge.Confidence != operation.Edge.Confidence || edge.CreatedByJobID != "job_1" || !edge.CreatedAt.Equal(fixedApplyTime) {
		t.Fatalf("extension edge metadata mismatch: %#v", edge)
	}
}

func TestStoreBackedApplyEngine_WritesUpdateMemoryWithTraceAndSupersessionEdge(t *testing.T) {
	t.Parallel()

	memories := &fakeMemoryTraceCreator{}
	engine := newTestStoreBackedApplyEngine(t, memories)
	operation := validUpdateOperation()

	result, err := engine.Apply(context.Background(), validApplyRequest(operation))
	if err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}

	if result.AppliedOperationCount != 1 {
		t.Fatalf("unexpected applied count: got %d want 1", result.AppliedOperationCount)
	}
	if !result.TraceWritten {
		t.Fatalf("expected trace to be written")
	}
	if len(memories.updateMemories) != 1 || len(memories.updateTraces) != 1 || len(memories.updateEdges) != 1 {
		t.Fatalf("expected one update memory, trace, and edge; got memories=%d traces=%d edges=%d", len(memories.updateMemories), len(memories.updateTraces), len(memories.updateEdges))
	}

	memory := memories.updateMemories[0]
	if memory.ID == "" || !strings.HasPrefix(memory.ID, "mem_") {
		t.Fatalf("expected deterministic memory ID, got %q", memory.ID)
	}
	if memory.Status != core.MemoryStatusActive || !memory.LatestFlag {
		t.Fatalf("update memory should become the active latest memory, got status=%q latest=%v", memory.Status, memory.LatestFlag)
	}
	if memory.Scope != operation.Memory.Scope || memory.OwnerEntityID != operation.Memory.OwnerEntityID {
		t.Fatalf("update memory did not preserve mutation scope/owner: %#v", memory)
	}

	trace := memories.updateTraces[0]
	if trace.MemoryID != memory.ID {
		t.Fatalf("trace memory ID mismatch: got %q want %q", trace.MemoryID, memory.ID)
	}
	assertJSONContains(t, trace.AppliedOperationsJSON, "op_update")

	edge := memories.updateEdges[0]
	if edge.FromMemoryID != memory.ID {
		t.Fatalf("updates edge must originate from the written memory: got %q want %q", edge.FromMemoryID, memory.ID)
	}
	if edge.ToMemoryID != operation.Memory.TargetID {
		t.Fatalf("updates edge target mismatch: got %q want %q", edge.ToMemoryID, operation.Memory.TargetID)
	}
	if edge.EdgeKind != core.EdgeKindUpdates {
		t.Fatalf("updates edge kind mismatch: got %q", edge.EdgeKind)
	}
}

func TestStoreBackedApplyEngine_RejectsUnsupportedWrites(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		mutateReq func(*ApplyRequest)
		wantError string
	}{
		{
			name: "archive remains not implemented",
			mutateReq: func(req *ApplyRequest) {
				req.Reasoning.Stage2.Operations = []reasoning.GraphOperation{validArchiveOperation()}
			},
			wantError: "archive status handling",
		},
		{
			name: "group shared waits for membership validation",
			mutateReq: func(req *ApplyRequest) {
				groupID := "group_1"
				req.Reasoning.Stage2.Operations[0].Memory.Scope = core.MemoryScopeGroupShared
				req.Reasoning.Stage2.Operations[0].Memory.GroupID = &groupID
			},
			wantError: "membership validation",
		},
		{
			name: "profile delta waits for merge implementation",
			mutateReq: func(req *ApplyRequest) {
				req.Reasoning.Stage2.ProfileDelta = json.RawMessage(`{"static":{"tone":"brief"}}`)
			},
			wantError: "profile_delta writes are not implemented",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			memories := &fakeMemoryTraceCreator{}
			engine := newTestStoreBackedApplyEngine(t, memories)
			req := validApplyRequest(validCreateOperation())
			tt.mutateReq(req)

			_, err := engine.Apply(context.Background(), req)
			if err == nil {
				t.Fatalf("expected Apply to reject unsupported write")
			}
			if !errors.Is(err, core.ErrNotImplemented) {
				t.Fatalf("expected ErrNotImplemented, got %v", err)
			}
			if !strings.Contains(err.Error(), tt.wantError) {
				t.Fatalf("expected error to contain %q, got %v", tt.wantError, err)
			}
			if len(memories.memories) != 0 || len(memories.traces) != 0 || len(memories.edges) != 0 {
				t.Fatalf("unsupported write should not touch storage: memories=%d traces=%d edges=%d", len(memories.memories), len(memories.traces), len(memories.edges))
			}
		})
	}
}

func TestStoreBackedApplyEngine_RejectsGroupSharedWritesForAllWriteKinds(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		operation reasoning.GraphOperation
	}{
		{name: "create", operation: validCreateOperation()},
		{name: "extend", operation: validExtendOperation()},
		{name: "update", operation: validUpdateOperation()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			memories := &fakeMemoryTraceCreator{}
			engine := newTestStoreBackedApplyEngine(t, memories)
			req := validApplyRequest(asGroupSharedOperation(tt.operation))

			_, err := engine.Apply(context.Background(), req)
			if err == nil {
				t.Fatalf("expected group_shared %s write to be rejected", tt.name)
			}
			if !errors.Is(err, core.ErrNotImplemented) {
				t.Fatalf("expected ErrNotImplemented until membership validation exists, got %v", err)
			}
			if !strings.Contains(err.Error(), "membership validation") {
				t.Fatalf("expected membership validation stop-line error, got %v", err)
			}
			if len(memories.memories) != 0 || len(memories.traces) != 0 || len(memories.edges) != 0 ||
				len(memories.updateMemories) != 0 || len(memories.updateTraces) != 0 || len(memories.updateEdges) != 0 {
				t.Fatalf("group_shared rejection must happen before storage writes: %#v", memories)
			}
		})
	}
}

func TestStoreBackedApplyEngine_TraceFailureDoesNotReportSuccessfulApply(t *testing.T) {
	t.Parallel()

	memories := &fakeMemoryTraceCreator{err: errors.New("trace insert failed")}
	engine := newTestStoreBackedApplyEngine(t, memories)

	result, err := engine.Apply(context.Background(), validApplyRequest(validCreateOperation()))
	if err == nil {
		t.Fatalf("expected Apply to fail when memory trace persistence fails")
	}
	if result != nil {
		t.Fatalf("failed trace persistence must not return a successful apply result: %#v", result)
	}
	if len(memories.memories) != 0 || len(memories.traces) != 0 || len(memories.edges) != 0 {
		t.Fatalf("fake atomic store should record no successful writes on trace failure: memories=%d traces=%d edges=%d", len(memories.memories), len(memories.traces), len(memories.edges))
	}
}

func TestStoreBackedApplyEngine_ExtendEdgeFailureDoesNotReportSuccessfulApply(t *testing.T) {
	t.Parallel()

	memories := &fakeMemoryTraceCreator{edgeErr: errors.New("edge insert failed")}
	engine := newTestStoreBackedApplyEngine(t, memories)

	result, err := engine.Apply(context.Background(), validApplyRequest(validExtendOperation()))
	if err == nil {
		t.Fatalf("expected Apply to fail when extension edge persistence fails")
	}
	if !strings.Contains(err.Error(), "edge insert failed") {
		t.Fatalf("expected edge persistence error, got %v", err)
	}
	if result != nil {
		t.Fatalf("failed edge persistence must not return a successful apply result: %#v", result)
	}
	if len(memories.memories) != 0 || len(memories.traces) != 0 || len(memories.edges) != 0 {
		t.Fatalf("fake atomic store should record no successful writes on edge failure: memories=%d traces=%d edges=%d", len(memories.memories), len(memories.traces), len(memories.edges))
	}
}

func TestStoreBackedApplyEngine_UpdateFailureDoesNotReportSuccessfulApply(t *testing.T) {
	t.Parallel()

	memories := &fakeMemoryTraceCreator{updateErr: errors.New("target not latest")}
	engine := newTestStoreBackedApplyEngine(t, memories)

	result, err := engine.Apply(context.Background(), validApplyRequest(validUpdateOperation()))
	if err == nil {
		t.Fatalf("expected Apply to fail when update persistence fails")
	}
	if !strings.Contains(err.Error(), "target not latest") {
		t.Fatalf("expected update persistence error, got %v", err)
	}
	if result != nil {
		t.Fatalf("failed update persistence must not return a successful apply result: %#v", result)
	}
	if len(memories.updateMemories) != 0 || len(memories.updateTraces) != 0 || len(memories.updateEdges) != 0 {
		t.Fatalf("fake atomic store should record no successful update writes: memories=%d traces=%d edges=%d", len(memories.updateMemories), len(memories.updateTraces), len(memories.updateEdges))
	}
}

func asGroupSharedOperation(operation reasoning.GraphOperation) reasoning.GraphOperation {
	groupID := "group_design"
	operation.Memory.Scope = core.MemoryScopeGroupShared
	operation.Memory.GroupID = &groupID
	return operation
}

var fixedApplyTime = time.Date(2026, 4, 24, 12, 0, 0, 0, time.UTC)

func newTestStoreBackedApplyEngine(t *testing.T, memories *fakeMemoryTraceCreator) *StoreBackedApplyEngine {
	t.Helper()

	engine, err := NewStoreBackedApplyEngine(memories)
	if err != nil {
		t.Fatalf("NewStoreBackedApplyEngine returned error: %v", err)
	}
	engine.now = func() time.Time {
		return fixedApplyTime
	}
	return engine
}

func assertJSONContains(t *testing.T, raw json.RawMessage, want string) {
	t.Helper()

	if !json.Valid(raw) {
		t.Fatalf("expected valid JSON, got %s", string(raw))
	}
	if !strings.Contains(string(raw), want) {
		t.Fatalf("expected JSON to contain %q, got %s", want, string(raw))
	}
}

type fakeMemoryTraceCreator struct {
	memories       []*core.Memory
	traces         []*core.MemoryTrace
	edges          []*core.MemoryEdge
	updateMemories []*core.Memory
	updateTraces   []*core.MemoryTrace
	updateEdges    []*core.MemoryEdge
	err            error
	edgeErr        error
	updateErr      error
}

func (s *fakeMemoryTraceCreator) CreateMemoryWithTrace(_ context.Context, memory *core.Memory, trace *core.MemoryTrace) error {
	if s.err != nil {
		return s.err
	}
	s.memories = append(s.memories, memory)
	s.traces = append(s.traces, trace)
	return nil
}

func (s *fakeMemoryTraceCreator) CreateMemoryWithTraceAndEdge(_ context.Context, memory *core.Memory, trace *core.MemoryTrace, edge *core.MemoryEdge) error {
	if s.err != nil {
		return s.err
	}
	if s.edgeErr != nil {
		return s.edgeErr
	}
	s.memories = append(s.memories, memory)
	s.traces = append(s.traces, trace)
	s.edges = append(s.edges, edge)
	return nil
}

func (s *fakeMemoryTraceCreator) CreateMemoryWithTraceAndUpdateEdge(_ context.Context, memory *core.Memory, trace *core.MemoryTrace, edge *core.MemoryEdge) error {
	if s.updateErr != nil {
		return s.updateErr
	}
	s.updateMemories = append(s.updateMemories, memory)
	s.updateTraces = append(s.updateTraces, trace)
	s.updateEdges = append(s.updateEdges, edge)
	return nil
}
