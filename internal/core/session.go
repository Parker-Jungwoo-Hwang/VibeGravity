// ============================================================
// FILE     : internal/core/session.go
// PURPOSE  : Defines rebuildable summaries for session-level memory consolidation.
// LAYER    : domain
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : SessionSummary
// DEPENDS  : time
// USED_BY  : dreaming jobs, recall assembler, storage
// ------------------------------------------------------------
// AGENT_NOTE: Session summaries are derived artifacts and must keep source IDs.
// ============================================================

package core

import "time"

// SessionSummary is a rebuildable summary for one session.
type SessionSummary struct {
	ID              string    `json:"id"`
	TenantID        string    `json:"tenant_id"`
	WorkspaceID     string    `json:"workspace_id"`
	SessionID       string    `json:"session_id"`
	SummaryText     string    `json:"summary_text"`
	SourceEventIDs  []string  `json:"source_event_ids"`
	SourceMemoryIDs []string  `json:"source_memory_ids"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
