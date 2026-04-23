// ============================================================
// FILE     : internal/core/plan.go
// PURPOSE  : Defines structured plan records and their task items.
// LAYER    : domain
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : Plan, PlanItem
// DEPENDS  : encoding/json, time, internal/core/scope.go
// USED_BY  : internal/store, recall assembler, plan API
// ------------------------------------------------------------
// AGENT_NOTE: Active plans get recall priority, so preserve scope and evidence.
// ============================================================

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
