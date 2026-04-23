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
