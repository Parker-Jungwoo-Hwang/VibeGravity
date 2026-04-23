// ============================================================
// FILE     : internal/core/note.go
// PURPOSE  : Defines human-authored note records that can influence recall.
// LAYER    : domain
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : Note
// DEPENDS  : time, internal/core/scope.go
// USED_BY  : internal/store, recall assembler, note API
// ------------------------------------------------------------
// AGENT_NOTE: Notes are operator intent and must stay distinct from memories.
// ============================================================

package core

import "time"

// Note is a human-authored memory control artifact.
type Note struct {
	ID            string      `json:"id"`
	TenantID      string      `json:"tenant_id"`
	WorkspaceID   string      `json:"workspace_id"`
	NoteKind      string      `json:"note_kind"`
	Scope         MemoryScope `json:"scope"`
	OwnerEntityID string      `json:"owner_entity_id"`
	Text          string      `json:"text"`
	Pinned        bool        `json:"pinned"`
	ExpiresAt     *time.Time  `json:"expires_at,omitempty"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}
