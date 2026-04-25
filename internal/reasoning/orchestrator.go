// ============================================================
// FILE     : internal/reasoning/orchestrator.go
// PURPOSE  : Provides schema-first reasoning orchestration interfaces and safe stubs.
// LAYER    : application
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : Orchestrator, Stage1Extractor, Stage2Resolver, StubOrchestrator, stub runners, PipelineOrchestrator
// DEPENDS  : context, encoding/json, fmt, internal/core
// USED_BY  : internal/worker, cmd/worker, tests
// ------------------------------------------------------------
// AGENT_NOTE: Runners may call Codex later, but this package must not add local extraction.
// ============================================================

package reasoning

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

var emptyJSONObject = json.RawMessage(`{}`)

// Orchestrator runs schema-first Stage 1 extract and Stage 2 resolve.
type Orchestrator interface {
	ProcessTurn(ctx context.Context, envelope *ProcessTurnEnvelope) (*ProcessTurnResult, error)
}

// Stage1Extractor runs the schema-first extract pass.
type Stage1Extractor interface {
	Extract(ctx context.Context, input Stage1Input) (Stage1Output, error)
}

// Stage2Resolver runs the schema-first resolve pass.
type Stage2Resolver interface {
	Resolve(ctx context.Context, input Stage2Input) (Stage2Output, error)
}

// StubStage1Extractor is a contract-preserving placeholder until the Codex extract bridge lands.
type StubStage1Extractor struct{}

// NewStubStage1Extractor creates a Stage 1 extractor that returns empty candidates.
func NewStubStage1Extractor() *StubStage1Extractor {
	return &StubStage1Extractor{}
}

// Extract validates Stage 1 input and returns empty structured candidate output.
func (e *StubStage1Extractor) Extract(_ context.Context, input Stage1Input) (Stage1Output, error) {
	if input.JobID == "" {
		return Stage1Output{}, fmt.Errorf("%w: stage1 job_id is required", core.ErrInvalidArgument)
	}
	if input.TenantID == "" {
		return Stage1Output{}, fmt.Errorf("%w: stage1 tenant_id is required", core.ErrInvalidArgument)
	}
	if input.WorkspaceID == "" {
		return Stage1Output{}, fmt.Errorf("%w: stage1 workspace_id is required", core.ErrInvalidArgument)
	}
	if len(input.RawEvents) == 0 {
		return Stage1Output{}, fmt.Errorf("%w: stage1 raw event bundle is required", core.ErrInvalidArgument)
	}
	return Stage1Output{
		CandidateEntities: []CandidateEntity{},
		CandidateMemories: []CandidateMemory{},
	}, nil
}

// StubStage2Resolver is a contract-preserving placeholder until the Codex resolve bridge lands.
type StubStage2Resolver struct{}

// NewStubStage2Resolver creates a Stage 2 resolver that returns no graph operations.
func NewStubStage2Resolver() *StubStage2Resolver {
	return &StubStage2Resolver{}
}

// Resolve validates the prepared Stage 2 input and returns empty structured output.
func (r *StubStage2Resolver) Resolve(_ context.Context, input Stage2Input) (Stage2Output, error) {
	if err := validatePreparedStage2Input(input); err != nil {
		return Stage2Output{}, err
	}
	return emptyStage2Output(), nil
}

// PipelineOrchestrator runs Stage 1, prepares Stage 2 from the Stage 1 output, then runs Stage 2.
type PipelineOrchestrator struct {
	stage1              Stage1Extractor
	stage2              Stage2Resolver
	stage2InputPreparer *Stage2InputPreparer
}

// NewPipelineOrchestrator creates a mockable two-stage reasoning orchestrator.
func NewPipelineOrchestrator(stage1 Stage1Extractor, stage2 Stage2Resolver, preparer *Stage2InputPreparer) (*PipelineOrchestrator, error) {
	if stage1 == nil {
		return nil, fmt.Errorf("%w: stage1 extractor is required", core.ErrInvalidArgument)
	}
	if stage2 == nil {
		return nil, fmt.Errorf("%w: stage2 resolver is required", core.ErrInvalidArgument)
	}
	if preparer == nil {
		preparer = NewStage2InputPreparer(Stage2InputSources{})
	}
	return &PipelineOrchestrator{
		stage1:              stage1,
		stage2:              stage2,
		stage2InputPreparer: preparer,
	}, nil
}

// ProcessTurn executes the schema-first two-stage reasoning chain.
func (o *PipelineOrchestrator) ProcessTurn(ctx context.Context, envelope *ProcessTurnEnvelope) (*ProcessTurnResult, error) {
	if err := validateProcessTurnEnvelope(envelope); err != nil {
		return nil, err
	}
	stage1Output, err := o.stage1.Extract(ctx, envelope.Stage1)
	if err != nil {
		return nil, fmt.Errorf("stage1 extract: %w", err)
	}
	stage2Input, err := o.stage2InputPreparer.Prepare(ctx, Stage2InputRequest{
		JobID:       envelope.JobID,
		TenantID:    envelope.TenantID,
		WorkspaceID: envelope.WorkspaceID,
		RawEvents:   envelope.RawEvents,
		Stage1:      stage1Output,
	})
	if err != nil {
		return nil, fmt.Errorf("prepare stage2 input: %w", err)
	}
	stage2Output, err := o.stage2.Resolve(ctx, stage2Input)
	if err != nil {
		return nil, fmt.Errorf("stage2 resolve: %w", err)
	}
	if err := validateStage2Output(stage2Output); err != nil {
		return nil, err
	}
	return &ProcessTurnResult{
		Stage1: stage1Output,
		Stage2: stage2Output,
	}, nil
}

// StubOrchestrator is a compatibility wrapper around the safe two-stage stub pipeline.
type StubOrchestrator struct {
	pipeline *PipelineOrchestrator
}

// NewStubOrchestrator creates a reasoning orchestrator that returns empty structured output.
func NewStubOrchestrator() *StubOrchestrator {
	pipeline, err := NewPipelineOrchestrator(NewStubStage1Extractor(), NewStubStage2Resolver(), nil)
	if err != nil {
		panic(err)
	}
	return &StubOrchestrator{pipeline: pipeline}
}

// ProcessTurn validates the envelope and returns schema-shaped empty Stage 1 and Stage 2 output through the pipeline.
func (o *StubOrchestrator) ProcessTurn(ctx context.Context, envelope *ProcessTurnEnvelope) (*ProcessTurnResult, error) {
	return o.pipeline.ProcessTurn(ctx, envelope)
}

func validateProcessTurnEnvelope(envelope *ProcessTurnEnvelope) error {
	if envelope == nil {
		return fmt.Errorf("%w: reasoning envelope is required", core.ErrInvalidArgument)
	}
	if envelope.JobID == "" {
		return fmt.Errorf("%w: reasoning job_id is required", core.ErrInvalidArgument)
	}
	if envelope.TenantID == "" {
		return fmt.Errorf("%w: reasoning tenant_id is required", core.ErrInvalidArgument)
	}
	if envelope.WorkspaceID == "" {
		return fmt.Errorf("%w: reasoning workspace_id is required", core.ErrInvalidArgument)
	}
	if len(envelope.RawEvents) == 0 {
		return fmt.Errorf("%w: reasoning raw event bundle is required", core.ErrInvalidArgument)
	}
	if envelope.Stage2.RequiredOutputName != StageNameResolve {
		return fmt.Errorf("%w: reasoning stage 2 required output name must be resolve", core.ErrInvalidArgument)
	}
	if envelope.Stage2.RequiredOutputSchema == "" {
		return fmt.Errorf("%w: reasoning stage 2 required output schema is required", core.ErrInvalidArgument)
	}
	return nil
}

func validatePreparedStage2Input(input Stage2Input) error {
	if input.JobID == "" {
		return fmt.Errorf("%w: stage2 job_id is required", core.ErrInvalidArgument)
	}
	if input.TenantID == "" {
		return fmt.Errorf("%w: stage2 tenant_id is required", core.ErrInvalidArgument)
	}
	if input.WorkspaceID == "" {
		return fmt.Errorf("%w: stage2 workspace_id is required", core.ErrInvalidArgument)
	}
	if len(input.RawEvents) == 0 {
		return fmt.Errorf("%w: stage2 raw event bundle is required", core.ErrInvalidArgument)
	}
	if input.RequiredOutputName != StageNameResolve {
		return fmt.Errorf("%w: stage2 required output name must be resolve", core.ErrInvalidArgument)
	}
	if input.RequiredOutputSchema == "" {
		return fmt.Errorf("%w: stage2 required output schema is required", core.ErrInvalidArgument)
	}
	return nil
}

func validateStage2Output(output Stage2Output) error {
	if output.Trace.SchemaVersion == "" {
		return fmt.Errorf("%w: stage2 trace schema_version is required", core.ErrInvalidArgument)
	}
	if output.Trace.Stage != StageNameResolve {
		return fmt.Errorf("%w: stage2 trace must be resolve", core.ErrInvalidArgument)
	}
	return nil
}

func emptyStage2Output() Stage2Output {
	return Stage2Output{
		Operations:     []GraphOperation{},
		ProfileDelta:   cloneRawJSON(emptyJSONObject),
		SessionSummary: "",
		PlanDelta:      cloneRawJSON(emptyJSONObject),
		Trace: Trace{
			SchemaVersion: "v0",
			Stage:         StageNameResolve,
			Codes:         []string{"stub_reasoning_no_operations"},
			MetadataJSON:  cloneRawJSON(emptyJSONObject),
		},
	}
}

func cloneRawJSON(raw json.RawMessage) json.RawMessage {
	return append(json.RawMessage(nil), raw...)
}
