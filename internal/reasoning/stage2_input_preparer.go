// ============================================================
// FILE     : internal/reasoning/stage2_input_preparer.go
// PURPOSE  : Assembles the schema-first Stage 2 input envelope from retrieval context.
// LAYER    : application
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : Stage2ResolveOutputSchemaV0, Stage2InputRequest, Stage2InputSources, Stage2InputPreparer
// DEPENDS  : context, fmt, internal/core
// USED_BY  : internal/worker, internal/reasoning tests
// ------------------------------------------------------------
// AGENT_NOTE: This layer only prepares retrieval context; it must not extract text or call Codex.
// ============================================================

package reasoning

import (
	"context"
	"fmt"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

// Stage2ResolveOutputSchemaV0 marks the required structured Stage 2 output contract.
const Stage2ResolveOutputSchemaV0 = "stage2.resolve.output.v0"

// Stage2InputRequest carries current job data and Stage 1 candidates into context preparation.
type Stage2InputRequest struct {
	JobID       string
	TenantID    string
	WorkspaceID string
	RawEvents   []*core.RawEvent
	Stage1      Stage1Output
}

// Stage2ProfileSource loads the current rebuildable profile snapshot for Stage 2.
type Stage2ProfileSource interface {
	LoadStage2Profile(ctx context.Context, req Stage2InputRequest) (*core.Profile, error)
}

// Stage2MemorySource loads relevant derived memories for Stage 2 conflict resolution.
type Stage2MemorySource interface {
	LoadStage2Memories(ctx context.Context, req Stage2InputRequest) ([]core.MemoryResult, error)
}

// Stage2DocumentSource loads relevant document chunks for Stage 2 grounding.
type Stage2DocumentSource interface {
	LoadStage2Documents(ctx context.Context, req Stage2InputRequest) ([]core.DocumentChunkResult, error)
}

// Stage2PlanSource loads active plans that may affect Stage 2 plan deltas.
type Stage2PlanSource interface {
	LoadStage2ActivePlans(ctx context.Context, req Stage2InputRequest) ([]*core.Plan, error)
}

// Stage2NoteSource loads pinned notes that must be visible to Stage 2 reasoning.
type Stage2NoteSource interface {
	LoadStage2PinnedNotes(ctx context.Context, req Stage2InputRequest) ([]*core.Note, error)
}

// Stage2InputSources groups retrieval/context dependencies for Stage 2 preparation.
type Stage2InputSources struct {
	Profiles  Stage2ProfileSource
	Memories  Stage2MemorySource
	Documents Stage2DocumentSource
	Plans     Stage2PlanSource
	Notes     Stage2NoteSource
}

// Stage2InputPreparer assembles Stage2Input without performing extraction or reasoning.
type Stage2InputPreparer struct {
	sources Stage2InputSources
}

// NewStage2InputPreparer creates an interface-driven Stage 2 input preparer.
func NewStage2InputPreparer(sources Stage2InputSources) *Stage2InputPreparer {
	return &Stage2InputPreparer{sources: sources}
}

// Prepare assembles a structured Stage2Input from current events, Stage 1 candidates, and retrieval sources.
func (p *Stage2InputPreparer) Prepare(ctx context.Context, req Stage2InputRequest) (Stage2Input, error) {
	if err := validateStage2InputRequest(req); err != nil {
		return Stage2Input{}, err
	}

	profile, err := p.loadProfile(ctx, req)
	if err != nil {
		return Stage2Input{}, err
	}
	memories, err := p.loadMemories(ctx, req)
	if err != nil {
		return Stage2Input{}, err
	}
	documents, err := p.loadDocuments(ctx, req)
	if err != nil {
		return Stage2Input{}, err
	}
	plans, err := p.loadPlans(ctx, req)
	if err != nil {
		return Stage2Input{}, err
	}
	notes, err := p.loadNotes(ctx, req)
	if err != nil {
		return Stage2Input{}, err
	}

	return Stage2Input{
		JobID:                req.JobID,
		TenantID:             req.TenantID,
		WorkspaceID:          req.WorkspaceID,
		RawEvents:            append([]*core.RawEvent(nil), req.RawEvents...),
		Stage1:               req.Stage1,
		ExistingProfile:      profile,
		RelevantMemories:     cloneMemoryResults(memories),
		RelevantDocuments:    cloneDocumentChunkResults(documents),
		ActivePlans:          clonePlans(plans),
		PinnedNotes:          cloneNotes(notes),
		RequiredOutputName:   StageNameResolve,
		RequiredOutputSchema: Stage2ResolveOutputSchemaV0,
	}, nil
}

func validateStage2InputRequest(req Stage2InputRequest) error {
	if req.JobID == "" {
		return fmt.Errorf("%w: stage2 input job_id is required", core.ErrInvalidArgument)
	}
	if req.TenantID == "" {
		return fmt.Errorf("%w: stage2 input tenant_id is required", core.ErrInvalidArgument)
	}
	if req.WorkspaceID == "" {
		return fmt.Errorf("%w: stage2 input workspace_id is required", core.ErrInvalidArgument)
	}
	if len(req.RawEvents) == 0 {
		return fmt.Errorf("%w: stage2 input raw events are required", core.ErrInvalidArgument)
	}
	return nil
}

func cloneMemoryResults(memories []core.MemoryResult) []core.MemoryResult {
	if len(memories) == 0 {
		return []core.MemoryResult{}
	}
	return append([]core.MemoryResult(nil), memories...)
}

func cloneDocumentChunkResults(documents []core.DocumentChunkResult) []core.DocumentChunkResult {
	if len(documents) == 0 {
		return []core.DocumentChunkResult{}
	}
	return append([]core.DocumentChunkResult(nil), documents...)
}

func clonePlans(plans []*core.Plan) []*core.Plan {
	if len(plans) == 0 {
		return []*core.Plan{}
	}
	return append([]*core.Plan(nil), plans...)
}

func cloneNotes(notes []*core.Note) []*core.Note {
	if len(notes) == 0 {
		return []*core.Note{}
	}
	return append([]*core.Note(nil), notes...)
}

func (p *Stage2InputPreparer) loadProfile(ctx context.Context, req Stage2InputRequest) (*core.Profile, error) {
	if p == nil || p.sources.Profiles == nil {
		return nil, nil
	}
	profile, err := p.sources.Profiles.LoadStage2Profile(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("load stage2 profile: %w", err)
	}
	return profile, nil
}

func (p *Stage2InputPreparer) loadMemories(ctx context.Context, req Stage2InputRequest) ([]core.MemoryResult, error) {
	if p == nil || p.sources.Memories == nil {
		return []core.MemoryResult{}, nil
	}
	memories, err := p.sources.Memories.LoadStage2Memories(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("load stage2 memories: %w", err)
	}
	return memories, nil
}

func (p *Stage2InputPreparer) loadDocuments(ctx context.Context, req Stage2InputRequest) ([]core.DocumentChunkResult, error) {
	if p == nil || p.sources.Documents == nil {
		return []core.DocumentChunkResult{}, nil
	}
	documents, err := p.sources.Documents.LoadStage2Documents(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("load stage2 documents: %w", err)
	}
	return documents, nil
}

func (p *Stage2InputPreparer) loadPlans(ctx context.Context, req Stage2InputRequest) ([]*core.Plan, error) {
	if p == nil || p.sources.Plans == nil {
		return []*core.Plan{}, nil
	}
	plans, err := p.sources.Plans.LoadStage2ActivePlans(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("load stage2 active plans: %w", err)
	}
	return plans, nil
}

func (p *Stage2InputPreparer) loadNotes(ctx context.Context, req Stage2InputRequest) ([]*core.Note, error) {
	if p == nil || p.sources.Notes == nil {
		return []*core.Note{}, nil
	}
	notes, err := p.sources.Notes.LoadStage2PinnedNotes(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("load stage2 pinned notes: %w", err)
	}
	return notes, nil
}
