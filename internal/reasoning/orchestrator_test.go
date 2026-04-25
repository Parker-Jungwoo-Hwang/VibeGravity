// ============================================================
// FILE     : internal/reasoning/orchestrator_test.go
// PURPOSE  : Guards schema-first reasoning orchestration and stub requirements.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : reasoning orchestrator tests
// DEPENDS  : context, encoding/json, errors, strings, testing, time, internal/core
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: These tests must not introduce local extraction or real Codex calls.
// ============================================================

package reasoning

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

func TestStubOrchestratorRejectsMissingStage2OutputSchema(t *testing.T) {
	t.Parallel()

	envelope := testProcessTurnEnvelope()
	envelope.Stage2.RequiredOutputSchema = ""

	result, err := NewStubOrchestrator().ProcessTurn(context.Background(), envelope)
	if err == nil {
		t.Fatalf("expected missing required output schema to be rejected")
	}
	if result != nil {
		t.Fatalf("expected no reasoning result for invalid envelope, got %#v", result)
	}
	if !errors.Is(err, core.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
	if !strings.Contains(err.Error(), "required output schema") {
		t.Fatalf("expected required output schema error, got %v", err)
	}
}

func TestStubOrchestratorAcceptsPreparedStage2Envelope(t *testing.T) {
	t.Parallel()

	result, err := NewStubOrchestrator().ProcessTurn(context.Background(), testProcessTurnEnvelope())
	if err != nil {
		t.Fatalf("ProcessTurn returned error: %v", err)
	}
	if result == nil {
		t.Fatalf("expected structured stub result")
	}
	if result.Stage2.Trace.Stage != StageNameResolve {
		t.Fatalf("expected resolve trace, got %#v", result.Stage2.Trace)
	}
	if result.Stage2.Trace.SchemaVersion == "" {
		t.Fatalf("expected schema-shaped trace result, got %#v", result.Stage2.Trace)
	}
}

func TestPipelineOrchestratorPreparesStage2AfterStage1(t *testing.T) {
	t.Parallel()

	stage1Output := Stage1Output{
		CandidateEntities: []CandidateEntity{{
			EntityKind:    "agent",
			DisplayName:   "Hermes",
			Confidence:    0.91,
			MetadataJSON:  json.RawMessage(`{}`),
			SourceEventID: "evt_contract_1",
		}},
		CandidateMemories: []CandidateMemory{{
			Kind:          core.MemoryKindConstraint,
			ArtifactClass: core.ArtifactClassKnowledge,
			Scope:         core.MemoryScopeWorkspaceShared,
			Text:          "Use schema-first reasoning only.",
			Confidence:    0.93,
			RawEventIDs:   []string{"evt_contract_1"},
		}},
		SummaryHint: "schema-first",
		TaskHint:    "prepare stage 2 with stage 1 candidates",
	}
	memorySource := &capturingStage2MemorySource{
		memories: []core.MemoryResult{{
			MemoryID:      "mem_stage2_source",
			Kind:          core.MemoryKindConstraint,
			ArtifactClass: core.ArtifactClassKnowledge,
			Text:          "Existing source loaded after Stage 1.",
			Confidence:    0.88,
			Scope:         core.MemoryScopeWorkspaceShared,
			LatestFlag:    true,
		}},
	}
	stage2 := &capturingStage2Resolver{}
	orchestrator, err := NewPipelineOrchestrator(
		&fixedStage1Extractor{output: stage1Output},
		stage2,
		NewStage2InputPreparer(Stage2InputSources{Memories: memorySource}),
	)
	if err != nil {
		t.Fatalf("NewPipelineOrchestrator returned error: %v", err)
	}

	result, err := orchestrator.ProcessTurn(context.Background(), testProcessTurnEnvelope())
	if err != nil {
		t.Fatalf("ProcessTurn returned error: %v", err)
	}

	if memorySource.received.Stage1.TaskHint != stage1Output.TaskHint {
		t.Fatalf("expected Stage 2 source to receive Stage 1 output, got %#v", memorySource.received.Stage1)
	}
	if stage2.received.RequiredOutputSchema != Stage2ResolveOutputSchemaV0 {
		t.Fatalf("expected resolver to receive required output schema, got %q", stage2.received.RequiredOutputSchema)
	}
	if len(stage2.received.RelevantMemories) != 1 || stage2.received.RelevantMemories[0].MemoryID != "mem_stage2_source" {
		t.Fatalf("expected resolver to receive prepared memory source, got %#v", stage2.received.RelevantMemories)
	}
	if result.Stage1.TaskHint != stage1Output.TaskHint {
		t.Fatalf("expected result to preserve Stage 1 output, got %#v", result.Stage1)
	}
	if result.Stage2.Trace.Stage != StageNameResolve {
		t.Fatalf("expected resolve trace, got %#v", result.Stage2.Trace)
	}
}

func TestPipelineOrchestratorStopsBeforeStage2WhenStage1Fails(t *testing.T) {
	t.Parallel()

	memorySource := &capturingStage2MemorySource{}
	stage2 := &capturingStage2Resolver{}
	orchestrator, err := NewPipelineOrchestrator(
		&fixedStage1Extractor{err: errors.New("codex extract unavailable")},
		stage2,
		NewStage2InputPreparer(Stage2InputSources{Memories: memorySource}),
	)
	if err != nil {
		t.Fatalf("NewPipelineOrchestrator returned error: %v", err)
	}

	result, err := orchestrator.ProcessTurn(context.Background(), testProcessTurnEnvelope())
	if err == nil {
		t.Fatalf("expected Stage 1 error")
	}
	if result != nil {
		t.Fatalf("expected no result after Stage 1 failure, got %#v", result)
	}
	if memorySource.called {
		t.Fatalf("Stage 2 sources should not load after Stage 1 failure")
	}
	if stage2.called {
		t.Fatalf("Stage 2 resolver should not run after Stage 1 failure")
	}
	if !strings.Contains(err.Error(), "stage1 extract") {
		t.Fatalf("expected wrapped Stage 1 error, got %v", err)
	}
}

func testProcessTurnEnvelope() *ProcessTurnEnvelope {
	rawEvents := []*core.RawEvent{
		{
			ID:          "evt_contract_1",
			TenantID:    "tenant_1",
			WorkspaceID: "workspace_1",
			SessionID:   "session_1",
			ActorID:     "agent:hermes-main",
			EventKind:   "message",
			Source:      "hermes",
			Fingerprint: "fp_evt_contract_1",
			OccurredAt:  time.Date(2026, time.April, 24, 0, 0, 0, 0, time.UTC),
			PayloadJSON: json.RawMessage(`{"text":"contract gate"}`),
			CreatedAt:   time.Date(2026, time.April, 24, 0, 0, 0, 0, time.UTC),
		},
	}
	stage1 := Stage1Input{
		JobID:       "job_contract",
		TenantID:    "tenant_1",
		WorkspaceID: "workspace_1",
		RawEvents:   rawEvents,
	}
	stage2 := Stage2Input{
		JobID:                "job_contract",
		TenantID:             "tenant_1",
		WorkspaceID:          "workspace_1",
		RawEvents:            rawEvents,
		RelevantMemories:     []core.MemoryResult{},
		RelevantDocuments:    []core.DocumentChunkResult{},
		ActivePlans:          []*core.Plan{},
		PinnedNotes:          []*core.Note{},
		RequiredOutputName:   StageNameResolve,
		RequiredOutputSchema: Stage2ResolveOutputSchemaV0,
	}
	return &ProcessTurnEnvelope{
		JobID:       "job_contract",
		TenantID:    "tenant_1",
		WorkspaceID: "workspace_1",
		RawEventIDs: []string{"evt_contract_1"},
		RawEvents:   rawEvents,
		Stage1:      stage1,
		Stage2:      stage2,
	}
}

type fixedStage1Extractor struct {
	output Stage1Output
	err    error
}

func (e *fixedStage1Extractor) Extract(context.Context, Stage1Input) (Stage1Output, error) {
	if e.err != nil {
		return Stage1Output{}, e.err
	}
	return e.output, nil
}

type capturingStage2MemorySource struct {
	called   bool
	received Stage2InputRequest
	memories []core.MemoryResult
}

func (s *capturingStage2MemorySource) LoadStage2Memories(_ context.Context, req Stage2InputRequest) ([]core.MemoryResult, error) {
	s.called = true
	s.received = req
	return s.memories, nil
}

type capturingStage2Resolver struct {
	called   bool
	received Stage2Input
}

func (r *capturingStage2Resolver) Resolve(_ context.Context, input Stage2Input) (Stage2Output, error) {
	r.called = true
	r.received = input
	return emptyStage2Output(), nil
}
