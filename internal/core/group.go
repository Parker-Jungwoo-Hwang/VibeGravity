package core

import "time"

// MemoryGroup defines a named group for group-shared memory.
type MemoryGroup struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	WorkspaceID string    `json:"workspace_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// MemoryGroupMembership links an entity to a memory group.
type MemoryGroupMembership struct {
	GroupID   string    `json:"group_id"`
	EntityID  string    `json:"entity_id"`
	CreatedAt time.Time `json:"created_at"`
}
