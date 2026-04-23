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
