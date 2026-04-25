// ============================================================
// FILE     : internal/reasoning/contracts.go
// PURPOSE  : Defines schema-first contracts for Codex extract and resolve stages.
// LAYER    : application
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : StageName, process turn contracts, stage contracts, graph operation DTOs, Trace
// DEPENDS  : encoding/json, internal/core
// USED_BY  : internal/worker, internal/graph, tests
// ------------------------------------------------------------
// AGENT_NOTE: Stage outputs must be structured data only; free-form reasoning must not cross apply.
// ============================================================

package reasoning

import (
	"encoding/json"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

// StageName identifies a reasoning stage in memory trace records.
type StageName string

const (
	// StageNameExtract is the first Codex pass that proposes candidates only.
	StageNameExtract StageName = "extract"
	// StageNameResolve is the second Codex pass that produces apply operations.
	StageNameResolve StageName = "resolve"
)

// ProcessTurnEnvelope is the worker-built input bundle for a process_turn_event job.
type ProcessTurnEnvelope struct {
	JobID       string           `json:"job_id"`
	TenantID    string           `json:"tenant_id"`
	WorkspaceID string           `json:"workspace_id"`
	RawEventIDs []string         `json:"raw_event_ids"`
	RawEvents   []*core.RawEvent `json:"raw_events"`
	Stage1      Stage1Input      `json:"stage_1"`
	Stage2      Stage2Input      `json:"stage_2"`
}

// ProcessTurnResult contains the structured outputs from both reasoning stages.
type ProcessTurnResult struct {
	Stage1 Stage1Output `json:"stage_1"`
	Stage2 Stage2Output `json:"stage_2"`
}

// Stage1Input is the schema for candidate extraction.
type Stage1Input struct {
	JobID       string           `json:"job_id"`
	TenantID    string           `json:"tenant_id"`
	WorkspaceID string           `json:"workspace_id"`
	RawEvents   []*core.RawEvent `json:"raw_events"`
}

// Stage1Output is candidate-only output; it is not an apply contract.
type Stage1Output struct {
	CandidateEntities []CandidateEntity `json:"candidate_entities"`
	CandidateMemories []CandidateMemory `json:"candidate_memories"`
	SummaryHint       string            `json:"summary_hint"`
	TaskHint          string            `json:"task_hint"`
}

// CandidateEntity describes an entity mention proposed by Stage 1.
type CandidateEntity struct {
	EntityID      string          `json:"entity_id,omitempty"`
	EntityKind    string          `json:"entity_kind"`
	DisplayName   string          `json:"display_name"`
	Confidence    float64         `json:"confidence"`
	MetadataJSON  json.RawMessage `json:"metadata_json"`
	SourceEventID string          `json:"source_event_id"`
}

// CandidateMemory describes a possible memory proposed by Stage 1.
type CandidateMemory struct {
	Kind          core.MemoryKind    `json:"kind"`
	ArtifactClass core.ArtifactClass `json:"artifact_class"`
	Scope         core.MemoryScope   `json:"scope"`
	Text          string             `json:"text"`
	Confidence    float64            `json:"confidence"`
	RawEventIDs   []string           `json:"raw_event_ids"`
}

// Stage2Input is the schema for conflict resolution and operation planning.
type Stage2Input struct {
	JobID                string                     `json:"job_id"`
	TenantID             string                     `json:"tenant_id"`
	WorkspaceID          string                     `json:"workspace_id"`
	RawEvents            []*core.RawEvent           `json:"raw_events"`
	Stage1               Stage1Output               `json:"stage_1"`
	ExistingProfile      *core.Profile              `json:"existing_profile,omitempty"`
	RelevantMemories     []core.MemoryResult        `json:"relevant_memories"`
	RelevantDocuments    []core.DocumentChunkResult `json:"relevant_documents"`
	ActivePlans          []*core.Plan               `json:"active_plans"`
	PinnedNotes          []*core.Note               `json:"pinned_notes"`
	RequiredOutputName   StageName                  `json:"required_output_name"`
	RequiredOutputSchema string                     `json:"required_output_schema"`
}

// Stage2Output is the only reasoning output that may cross into the apply engine.
type Stage2Output struct {
	Operations     []GraphOperation `json:"operations"`
	ProfileDelta   json.RawMessage  `json:"profile_delta"`
	SessionSummary string           `json:"session_summary"`
	PlanDelta      json.RawMessage  `json:"plan_delta"`
	Trace          Trace            `json:"trace"`
}

// GraphOperation is a structured memory graph operation produced by Stage 2.
type GraphOperation struct {
	OperationID string          `json:"operation_id"`
	Kind        OperationKind   `json:"kind"`
	Memory      *MemoryMutation `json:"memory,omitempty"`
	Edge        *EdgeMutation   `json:"edge,omitempty"`
	RawEventIDs []string        `json:"raw_event_ids"`
	Metadata    json.RawMessage `json:"metadata"`
}

// OperationKind identifies the apply action requested by Stage 2.
type OperationKind string

const (
	// OperationKindCreateMemory requests a new derived memory.
	OperationKindCreateMemory OperationKind = "create_memory"
	// OperationKindUpdateMemory requests an updates edge and latest resolution.
	OperationKindUpdateMemory OperationKind = "update_memory"
	// OperationKindExtendMemory requests an extends edge while keeping prior memory alive.
	OperationKindExtendMemory OperationKind = "extend_memory"
	// OperationKindArchiveMemory requests recall suppression for an existing memory.
	OperationKindArchiveMemory OperationKind = "archive_memory"
)

// MemoryMutation is the structured memory payload for create/update operations.
type MemoryMutation struct {
	MemoryID      string             `json:"memory_id,omitempty"`
	TargetID      string             `json:"target_id,omitempty"`
	Kind          core.MemoryKind    `json:"kind"`
	ArtifactClass core.ArtifactClass `json:"artifact_class"`
	Scope         core.MemoryScope   `json:"scope"`
	GroupID       *string            `json:"group_id,omitempty"`
	OwnerEntityID string             `json:"owner_entity_id"`
	Text          string             `json:"text"`
	Confidence    float64            `json:"confidence"`
	MetadataJSON  json.RawMessage    `json:"metadata_json"`
}

// EdgeMutation is the structured edge payload for graph operations.
type EdgeMutation struct {
	FromMemoryID string        `json:"from_memory_id"`
	ToMemoryID   string        `json:"to_memory_id"`
	EdgeKind     core.EdgeKind `json:"edge_kind"`
	Confidence   float64       `json:"confidence"`
}

// Trace stores structured debugging evidence for the reasoning run.
type Trace struct {
	SchemaVersion string          `json:"schema_version"`
	Stage         StageName       `json:"stage"`
	Codes         []string        `json:"codes"`
	MetadataJSON  json.RawMessage `json:"metadata_json"`
}
