// ============================================================
// FILE     : internal/core/memory.go
// PURPOSE  : Defines derived memory, graph edge, and provenance trace records.
// LAYER    : domain
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : Memory, MemoryEdge, MemoryTrace, MemoryCorrection
// DEPENDS  : encoding/json, time, internal/core/kind.go, internal/core/scope.go
// USED_BY  : apply engine, recall, storage, explain-memory path
// ------------------------------------------------------------
// AGENT_NOTE: Never blur raw events, derived memories, and memory_trace.
// ============================================================

package core

import (
	"encoding/json"
	"time"
)

// Memory is a derived structured object used by recall and graph operations.
type Memory struct {
	ID                 string          `json:"id"`
	TenantID           string          `json:"tenant_id"`
	WorkspaceID        string          `json:"workspace_id"`
	Scope              MemoryScope     `json:"scope"`
	GroupID            *string         `json:"group_id,omitempty"`
	OwnerEntityID      string          `json:"owner_entity_id"`
	Kind               MemoryKind      `json:"kind"`
	ArtifactClass      ArtifactClass   `json:"artifact_class"`
	Text               string          `json:"text"`
	Fingerprint        string          `json:"fingerprint"`
	Confidence         float64         `json:"confidence"`
	Status             MemoryStatus    `json:"status"`
	ValidFrom          time.Time       `json:"valid_from"`
	ValidTo            *time.Time      `json:"valid_to,omitempty"`
	LatestFlag         bool            `json:"latest_flag"`
	MetadataJSON       json.RawMessage `json:"metadata_json"`
	EmbeddingModel     string          `json:"embedding_model"`
	EmbeddingDims      int             `json:"embedding_dims"`
	EmbeddingUpdatedAt *time.Time      `json:"embedding_updated_at,omitempty"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}

// MemoryEdge records a relationship between two memory objects.
type MemoryEdge struct {
	FromMemoryID   string    `json:"from_memory_id"`
	ToMemoryID     string    `json:"to_memory_id"`
	EdgeKind       EdgeKind  `json:"edge_kind"`
	Confidence     float64   `json:"confidence"`
	CreatedByJobID string    `json:"created_by_job_id"`
	CreatedAt      time.Time `json:"created_at"`
}

// MemoryTrace stores provenance and reasoning evidence for one memory.
type MemoryTrace struct {
	MemoryID               string          `json:"memory_id"`
	RawEventIDs            []string        `json:"raw_event_ids"`
	ReasoningJobID         string          `json:"reasoning_job_id"`
	ReasoningStage         string          `json:"reasoning_stage"`
	CandidateSnapshotJSON  json.RawMessage `json:"candidate_snapshot_json"`
	AppliedOperationsJSON  json.RawMessage `json:"applied_operations_json"`
	OperatorCorrectionFlag bool            `json:"operator_correction_flag"`
	RelatedDocumentIDs     []string        `json:"related_document_ids"`
	CreatedAt              time.Time       `json:"created_at"`
}

// MemoryCorrection records a human correction intent without superseding a memory.
type MemoryCorrection struct {
	ID             string          `json:"id"`
	TenantID       string          `json:"tenant_id"`
	WorkspaceID    string          `json:"workspace_id"`
	MemoryID       string          `json:"memory_id"`
	OperatorID     string          `json:"operator_id"`
	RawEventID     string          `json:"raw_event_id"`
	IdempotencyKey string          `json:"idempotency_key"`
	CorrectionText string          `json:"correction_text"`
	EvidenceJSON   json.RawMessage `json:"evidence_json"`
	Status         string          `json:"status"`
	CreatedAt      time.Time       `json:"created_at"`
}
