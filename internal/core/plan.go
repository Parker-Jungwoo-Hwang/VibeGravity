package core

import (
	"encoding/json"
	"time"
)

// Plan is a structured operator or agent plan.
type Plan struct {
	ID            string          `json:"id"`
	TenantID      string          `json:"tenant_id"`
	WorkspaceID   string          `json:"workspace_id"`
	Title         string          `json:"title"`
	Status        string          `json:"status"`
	Scope         MemoryScope     `json:"scope"`
	OwnerEntityID string          `json:"owner_entity_id"`
	EvidenceJSON  json.RawMessage `json:"evidence_json"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

// PlanItem is a task within a structured plan.
type PlanItem struct {
	ID           string          `json:"id"`
	PlanID       string          `json:"plan_id"`
	Title        string          `json:"title"`
	Status       string          `json:"status"`
	EvidenceJSON json.RawMessage `json:"evidence_json"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}
