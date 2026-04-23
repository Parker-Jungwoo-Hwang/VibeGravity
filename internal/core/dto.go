// ============================================================
// FILE     : internal/core/dto.go
// PURPOSE  : Defines v1 request and response DTOs shared by runtime surfaces.
// LAYER    : domain
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : PrefetchRequest, SyncTurnRequest, search/note/plan/memory DTOs
// DEPENDS  : encoding/json, time, plans/05_runtime-contracts_ingest-recall-apply.md
// USED_BY  : internal/core/service.go, internal/httpapi, tests
// ------------------------------------------------------------
// AGENT_NOTE: Keep DTO changes synchronized with runtime contract docs.
// ============================================================

package core

import (
	"encoding/json"
	"time"
)

// PrefetchRequest asks the recall assembler for a typed next-turn recall pack.
type PrefetchRequest struct {
	TenantID     string `json:"tenant_id"`
	WorkspaceID  string `json:"workspace_id"`
	SessionID    string `json:"session_id"`
	ActorID      string `json:"actor_id"`
	Query        string `json:"query"`
	BudgetTokens int    `json:"budget_tokens"`
	Mode         string `json:"mode"`
}

// PrefetchResponse returns budget-aware typed recall blocks.
type PrefetchResponse struct {
	Blocks []RecallBlock `json:"blocks"`
	Meta   RecallMeta    `json:"meta"`
}

// RecallBlock is one typed item in a recall pack.
type RecallBlock struct {
	Kind     string `json:"kind"`
	Priority int    `json:"priority"`
	Text     string `json:"text"`
}

// RecallMeta describes recall assembly and token budget metadata.
type RecallMeta struct {
	EstimatedTokens int      `json:"estimated_tokens"`
	Sources         []string `json:"sources"`
}

// SyncTurnRequest records a complete turn through the hot ingest path.
type SyncTurnRequest struct {
	TenantID       string            `json:"tenant_id"`
	WorkspaceID    string            `json:"workspace_id"`
	SessionID      string            `json:"session_id"`
	ActorID        string            `json:"actor_id"`
	IdempotencyKey string            `json:"idempotency_key"`
	TurnEvents     []RawEventPayload `json:"turn_events"`
}

// RawEventPayload is the event payload accepted by SyncTurn.
type RawEventPayload struct {
	EventKind   string          `json:"event_kind"`
	Source      string          `json:"source"`
	Fingerprint string          `json:"fingerprint"`
	OccurredAt  time.Time       `json:"occurred_at"`
	PayloadJSON json.RawMessage `json:"payload_json"`
}

// SyncTurnResponse acknowledges accepted raw events and queued jobs.
type SyncTurnResponse struct {
	Status         string   `json:"status"`
	SessionID      string   `json:"session_id"`
	EventIDs       []string `json:"event_ids"`
	JobIDs         []string `json:"job_ids"`
	DuplicateCount int      `json:"duplicate_count"`
}

// AddDocumentRequest adds a document for chunking and document search.
type AddDocumentRequest struct {
	TenantID     string          `json:"tenant_id"`
	WorkspaceID  string          `json:"workspace_id"`
	Source       string          `json:"source"`
	Title        string          `json:"title"`
	Content      string          `json:"content"`
	Fingerprint  string          `json:"fingerprint"`
	MetadataJSON json.RawMessage `json:"metadata_json"`
	VersionHint  string          `json:"version_hint"`
}

// AddDocumentResponse reports the created document and chunk identifiers.
type AddDocumentResponse struct {
	DocumentID string   `json:"document_id"`
	ChunkIDs   []string `json:"chunk_ids"`
	Status     string   `json:"status"`
}

// SearchMemoriesRequest searches memories with scope and artifact-class filters.
type SearchMemoriesRequest struct {
	TenantID        string          `json:"tenant_id"`
	WorkspaceID     string          `json:"workspace_id"`
	Query           string          `json:"query"`
	Scopes          []MemoryScope   `json:"scopes"`
	ArtifactClasses []ArtifactClass `json:"artifact_classes"`
}

// SearchMemoriesResponse returns matching memory records.
type SearchMemoriesResponse struct {
	Memories []MemoryResult `json:"memories"`
}

// MemoryResult is a recall-safe search result for a memory.
type MemoryResult struct {
	MemoryID      string        `json:"memory_id"`
	Kind          MemoryKind    `json:"kind"`
	ArtifactClass ArtifactClass `json:"artifact_class"`
	Text          string        `json:"text"`
	Confidence    float64       `json:"confidence"`
	Scope         MemoryScope   `json:"scope"`
	ValidFrom     time.Time     `json:"valid_from"`
	LatestFlag    bool          `json:"latest_flag"`
}

// SearchDocumentsRequest searches document chunks.
type SearchDocumentsRequest struct {
	TenantID    string `json:"tenant_id"`
	WorkspaceID string `json:"workspace_id"`
	Query       string `json:"query"`
}

// SearchDocumentsResponse returns matching document chunks.
type SearchDocumentsResponse struct {
	Chunks []DocumentChunkResult `json:"chunks"`
}

// DocumentChunkResult is a search result for a document retrieval unit.
type DocumentChunkResult struct {
	ChunkID    string  `json:"chunk_id"`
	DocumentID string  `json:"document_id"`
	Text       string  `json:"text"`
	Score      float64 `json:"score"`
}

// AddNoteRequest creates a human-authored note.
type AddNoteRequest struct {
	TenantID      string      `json:"tenant_id"`
	WorkspaceID   string      `json:"workspace_id"`
	NoteKind      string      `json:"note_kind"`
	Scope         MemoryScope `json:"scope"`
	OwnerEntityID string      `json:"owner_entity_id"`
	Text          string      `json:"text"`
	Pinned        bool        `json:"pinned"`
	ExpiresAt     *time.Time  `json:"expires_at,omitempty"`
}

// AddNoteResponse reports the created note.
type AddNoteResponse struct {
	NoteID string `json:"note_id"`
	Status string `json:"status"`
}

// PlanItemInput is an item payload used when creating or updating a plan.
type PlanItemInput struct {
	ID           string          `json:"id,omitempty"`
	Title        string          `json:"title"`
	Status       string          `json:"status"`
	EvidenceJSON json.RawMessage `json:"evidence_json"`
}

// CreatePlanRequest creates a structured plan.
type CreatePlanRequest struct {
	TenantID      string          `json:"tenant_id"`
	WorkspaceID   string          `json:"workspace_id"`
	Title         string          `json:"title"`
	Status        string          `json:"status"`
	Scope         MemoryScope     `json:"scope"`
	OwnerEntityID string          `json:"owner_entity_id"`
	EvidenceJSON  json.RawMessage `json:"evidence_json"`
	Items         []PlanItemInput `json:"items"`
}

// CreatePlanResponse reports the created plan and item identifiers.
type CreatePlanResponse struct {
	PlanID  string   `json:"plan_id"`
	ItemIDs []string `json:"item_ids"`
	Status  string   `json:"status"`
}

// UpdatePlanRequest updates a structured plan.
type UpdatePlanRequest struct {
	TenantID     string          `json:"tenant_id"`
	WorkspaceID  string          `json:"workspace_id"`
	PlanID       string          `json:"plan_id"`
	Title        *string         `json:"title,omitempty"`
	Status       *string         `json:"status,omitempty"`
	EvidenceJSON json.RawMessage `json:"evidence_json"`
	Items        []PlanItemInput `json:"items"`
}

// UpdatePlanResponse reports the updated plan.
type UpdatePlanResponse struct {
	PlanID string `json:"plan_id"`
	Status string `json:"status"`
}

// CorrectMemoryRequest records a human correction for a memory.
type CorrectMemoryRequest struct {
	TenantID       string          `json:"tenant_id"`
	WorkspaceID    string          `json:"workspace_id"`
	MemoryID       string          `json:"memory_id"`
	OperatorID     string          `json:"operator_id"`
	IdempotencyKey string          `json:"idempotency_key"`
	CorrectionText string          `json:"correction_text"`
	EvidenceJSON   json.RawMessage `json:"evidence_json"`
}

// CorrectMemoryResponse reports the correction side effects.
type CorrectMemoryResponse struct {
	MemoryID     string `json:"memory_id"`
	RawEventID   string `json:"raw_event_id"`
	TraceWritten bool   `json:"trace_written"`
	Status       string `json:"status"`
}

// GetTimelineRequest asks for a timeline view assembled from existing artifacts.
type GetTimelineRequest struct {
	TenantID    string        `json:"tenant_id"`
	WorkspaceID string        `json:"workspace_id"`
	Scopes      []MemoryScope `json:"scopes"`
	EntityID    string        `json:"entity_id"`
	From        *time.Time    `json:"from,omitempty"`
	To          *time.Time    `json:"to,omitempty"`
	Limit       int           `json:"limit"`
}

// GetTimelineResponse returns timeline view items without requiring a cache table.
type GetTimelineResponse struct {
	Items []TimelineItem `json:"items"`
}

// TimelineItem is one row in the timeline view.
type TimelineItem struct {
	ID            string        `json:"id"`
	Kind          MemoryKind    `json:"kind"`
	ArtifactClass ArtifactClass `json:"artifact_class"`
	Text          string        `json:"text"`
	OccurredAt    time.Time     `json:"occurred_at"`
	MemoryID      string        `json:"memory_id"`
	RawEventID    string        `json:"raw_event_id"`
}

// ExplainMemoryRequest asks for provenance for one memory.
type ExplainMemoryRequest struct {
	TenantID    string `json:"tenant_id"`
	WorkspaceID string `json:"workspace_id"`
	MemoryID    string `json:"memory_id"`
}

// ExplainMemoryResponse returns trace, edges, and source event evidence.
type ExplainMemoryResponse struct {
	MemoryID     string                   `json:"memory_id"`
	Trace        MemoryTraceResult        `json:"trace"`
	Edges        []MemoryEdgeResult       `json:"edges"`
	SourceEvents []ProvenanceEventResult  `json:"source_events"`
	Documents    []ProvenanceDocumentLink `json:"documents"`
}

// MemoryTraceResult is the DTO shape for memory provenance.
type MemoryTraceResult struct {
	RawEventIDs            []string        `json:"raw_event_ids"`
	ReasoningJobID         string          `json:"reasoning_job_id"`
	ReasoningStage         string          `json:"reasoning_stage"`
	CandidateSnapshotJSON  json.RawMessage `json:"candidate_snapshot_json"`
	AppliedOperationsJSON  json.RawMessage `json:"applied_operations_json"`
	OperatorCorrectionFlag bool            `json:"operator_correction_flag"`
	RelatedDocumentIDs     []string        `json:"related_document_ids"`
	CreatedAt              time.Time       `json:"created_at"`
}

// MemoryEdgeResult is the DTO shape for one memory edge.
type MemoryEdgeResult struct {
	FromMemoryID string    `json:"from_memory_id"`
	ToMemoryID   string    `json:"to_memory_id"`
	EdgeKind     EdgeKind  `json:"edge_kind"`
	Confidence   float64   `json:"confidence"`
	CreatedAt    time.Time `json:"created_at"`
}

// ProvenanceEventResult summarizes a source raw event for provenance display.
type ProvenanceEventResult struct {
	EventID     string          `json:"event_id"`
	EventKind   string          `json:"event_kind"`
	Source      string          `json:"source"`
	Fingerprint string          `json:"fingerprint"`
	OccurredAt  time.Time       `json:"occurred_at"`
	PayloadJSON json.RawMessage `json:"payload_json"`
}

// ProvenanceDocumentLink summarizes a document used as memory evidence.
type ProvenanceDocumentLink struct {
	DocumentID string `json:"document_id"`
	Title      string `json:"title"`
}
