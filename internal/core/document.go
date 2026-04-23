// ============================================================
// FILE     : internal/core/document.go
// PURPOSE  : Defines document and chunk records used by document retrieval.
// LAYER    : domain
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : Document, DocumentChunk
// DEPENDS  : encoding/json, time
// USED_BY  : internal/store, internal/core/dto.go, recall and search paths
// ------------------------------------------------------------
// AGENT_NOTE: Keep documents separate from derived memories and memory trace.
// ============================================================

package core

import (
	"encoding/json"
	"time"
)

// Document is a deduplicated document-level source artifact.
type Document struct {
	ID           string          `json:"id"`
	TenantID     string          `json:"tenant_id"`
	WorkspaceID  string          `json:"workspace_id"`
	Source       string          `json:"source"`
	Title        string          `json:"title"`
	Fingerprint  string          `json:"fingerprint"`
	MetadataJSON json.RawMessage `json:"metadata_json"`
	VersionHint  string          `json:"version_hint"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// DocumentChunk is a searchable retrieval unit for a document.
type DocumentChunk struct {
	ID                 string          `json:"id"`
	DocumentID         string          `json:"document_id"`
	ChunkIndex         int             `json:"chunk_index"`
	Text               string          `json:"text"`
	HeadingPath        string          `json:"heading_path"`
	MetadataJSON       json.RawMessage `json:"metadata_json"`
	NeighborChunkIDs   []string        `json:"neighbor_chunk_ids"`
	EmbeddingModel     string          `json:"embedding_model"`
	EmbeddingDims      int             `json:"embedding_dims"`
	EmbeddingUpdatedAt *time.Time      `json:"embedding_updated_at,omitempty"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}
