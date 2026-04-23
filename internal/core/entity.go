// ============================================================
// FILE     : internal/core/entity.go
// PURPOSE  : Defines entity records for users, agents, workspaces, projects, and groups.
// LAYER    : domain
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : Entity
// DEPENDS  : encoding/json, time
// USED_BY  : group membership, profile, and scope-aware storage paths
// ------------------------------------------------------------
// AGENT_NOTE: Preserve tenant and workspace fields on every persisted entity.
// ============================================================

package core

import (
	"encoding/json"
	"time"
)

// Entity represents a user, agent, workspace, project, or group.
type Entity struct {
	ID           string          `json:"id"`
	TenantID     string          `json:"tenant_id"`
	WorkspaceID  string          `json:"workspace_id"`
	EntityKind   string          `json:"entity_kind"`
	DisplayName  string          `json:"display_name"`
	MetadataJSON json.RawMessage `json:"metadata_json"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}
