// ============================================================
// FILE     : internal/core/job.go
// PURPOSE  : Defines PostgreSQL-backed worker queue job records.
// LAYER    : domain
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : IngestJob
// DEPENDS  : encoding/json, time, internal/core/kind.go
// USED_BY  : internal/store, cmd/worker, ingest pipeline
// ------------------------------------------------------------
// AGENT_NOTE: Jobs must support retry without duplicate apply side effects.
// ============================================================

package core

import (
	"encoding/json"
	"time"
)

// IngestJob is a PostgreSQL-backed worker queue item.
type IngestJob struct {
	ID          string          `json:"id"`
	TenantID    string          `json:"tenant_id"`
	WorkspaceID string          `json:"workspace_id"`
	JobKind     JobKind         `json:"job_kind"`
	Status      string          `json:"status"`
	RawEventIDs []string        `json:"raw_event_ids"`
	PayloadJSON json.RawMessage `json:"payload_json"`
	Attempts    int             `json:"attempts"`
	AvailableAt time.Time       `json:"available_at"`
	LockedBy    *string         `json:"locked_by,omitempty"`
	LockedAt    *time.Time      `json:"locked_at,omitempty"`
	LastError   *string         `json:"last_error,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}
