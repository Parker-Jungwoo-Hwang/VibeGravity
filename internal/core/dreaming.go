// ============================================================
// FILE     : internal/core/dreaming.go
// PURPOSE  : Defines background dreaming requests, inputs, and promotion results.
// LAYER    : domain
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : DreamingTier, DreamSessionRequest, DreamWorkspaceRequest, DreamingSessionInput, DreamingPromotionRequest, DreamingPromotionResult, DreamingResult
// DEPENDS  : time, internal/core/memory.go
// USED_BY  : graph dreaming service, worker, storage
// ------------------------------------------------------------
// AGENT_NOTE: Dreaming changes consolidation metadata and summaries without blurring scope boundaries.
// ============================================================

package core

import "time"

// DreamingTier describes consolidation depth for a memory.
type DreamingTier string

const (
	// DreamingTierShortTerm is recent scratch or raw-tail material.
	DreamingTierShortTerm DreamingTier = "short-term"
	// DreamingTierMidTerm is session or active-topic consolidation.
	DreamingTierMidTerm DreamingTier = "mid-term"
	// DreamingTierLongTerm is stable reusable memory.
	DreamingTierLongTerm DreamingTier = "long-term"
	// DreamingTierUltraLongTerm is canonical, repeatedly confirmed memory.
	DreamingTierUltraLongTerm DreamingTier = "ultra-long-term"
)

// DreamSessionRequest asks dreaming to consolidate one session.
type DreamSessionRequest struct {
	JobID       string    `json:"job_id"`
	TenantID    string    `json:"tenant_id"`
	WorkspaceID string    `json:"workspace_id"`
	SessionID   string    `json:"session_id"`
	Now         time.Time `json:"now"`
}

// DreamWorkspaceRequest asks dreaming to promote stable workspace memories.
type DreamWorkspaceRequest struct {
	JobID       string    `json:"job_id"`
	TenantID    string    `json:"tenant_id"`
	WorkspaceID string    `json:"workspace_id"`
	Now         time.Time `json:"now"`
}

// DreamingSessionInput is the source material for session consolidation.
type DreamingSessionInput struct {
	RawEventIDs []string  `json:"raw_event_ids"`
	Memories    []*Memory `json:"memories"`
}

// DreamingPromotionRequest selects existing memories for tier promotion.
type DreamingPromotionRequest struct {
	JobID             string       `json:"job_id"`
	TenantID          string       `json:"tenant_id"`
	WorkspaceID       string       `json:"workspace_id"`
	SessionID         string       `json:"session_id,omitempty"`
	MemoryIDs         []string     `json:"memory_ids,omitempty"`
	Tier              DreamingTier `json:"tier"`
	MinConfidence     float64      `json:"min_confidence"`
	RequireStableKind bool         `json:"require_stable_kind"`
	Now               time.Time    `json:"now"`
}

// DreamingPromotionResult reports metadata-only memory promotions.
type DreamingPromotionResult struct {
	PromotedCount int      `json:"promoted_count"`
	MemoryIDs     []string `json:"memory_ids"`
}

// DreamingResult reports one background dreaming job outcome.
type DreamingResult struct {
	SessionSummaryWritten bool `json:"session_summary_written"`
	MidTermPromoted       int  `json:"mid_term_promoted"`
	LongTermPromoted      int  `json:"long_term_promoted"`
	UltraLongTermPromoted int  `json:"ultra_long_term_promoted"`
}
