package core

import (
	"encoding/json"
	"time"
)

// RawEvent is an immutable ingest record before memory derivation.
type RawEvent struct {
	ID             string          `json:"id"`
	TenantID       string          `json:"tenant_id"`
	WorkspaceID    string          `json:"workspace_id"`
	SessionID      string          `json:"session_id"`
	ActorID        string          `json:"actor_id"`
	EventKind      string          `json:"event_kind"`
	Source         string          `json:"source"`
	IdempotencyKey string          `json:"idempotency_key"`
	Fingerprint    string          `json:"fingerprint"`
	OccurredAt     time.Time       `json:"occurred_at"`
	PayloadJSON    json.RawMessage `json:"payload_json"`
	CreatedAt      time.Time       `json:"created_at"`
}
