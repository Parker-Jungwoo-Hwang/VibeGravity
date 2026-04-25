// ============================================================
// FILE     : internal/core/job.go
// PURPOSE  : Defines PostgreSQL-backed worker queue job records.
// LAYER    : domain
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : IngestJob, JobBacklogMetricsRequest, JobStatusCounts, JobBacklogMetrics
// DEPENDS  : encoding/json, time, internal/core/kind.go
// USED_BY  : internal/store, cmd/worker, ingest pipeline
// ------------------------------------------------------------
// AGENT_NOTE: Jobs must support retry without duplicate apply side effects.
// ============================================================

package core

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
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

// CorrectionApplyJobInput carries the stable inputs for synchronous correction provenance.
type CorrectionApplyJobInput struct {
	TenantID       string
	WorkspaceID    string
	CorrectionID   string
	TargetMemoryID string
	IdempotencyKey string
	RawEventID     string
	OperatorID     string
	AppliedAt      time.Time
}

// NewCorrectionApplyJob builds the completed ingest_jobs row that backs correction supersession provenance.
func NewCorrectionApplyJob(input CorrectionApplyJobInput) *IngestJob {
	appliedAt := input.AppliedAt.UTC()
	if appliedAt.IsZero() {
		appliedAt = time.Now().UTC()
	}
	payload, err := json.Marshal(struct {
		Source         string `json:"source"`
		CorrectionID   string `json:"correction_id"`
		TargetMemoryID string `json:"target_memory_id"`
		IdempotencyKey string `json:"idempotency_key"`
		OperatorID     string `json:"operator_id"`
	}{
		Source:         "operator_correction",
		CorrectionID:   input.CorrectionID,
		TargetMemoryID: input.TargetMemoryID,
		IdempotencyKey: input.IdempotencyKey,
		OperatorID:     input.OperatorID,
	})
	if err != nil {
		payload = []byte(`{"source":"operator_correction"}`)
	}
	rawEventIDs := []string{}
	if strings.TrimSpace(input.RawEventID) != "" {
		rawEventIDs = []string{input.RawEventID}
	}
	return &IngestJob{
		ID:          CorrectionApplyJobID(input.TenantID, input.WorkspaceID, input.CorrectionID, input.TargetMemoryID, input.IdempotencyKey),
		TenantID:    input.TenantID,
		WorkspaceID: input.WorkspaceID,
		JobKind:     JobKindCorrectionApply,
		Status:      "complete",
		RawEventIDs: rawEventIDs,
		PayloadJSON: payload,
		AvailableAt: appliedAt,
		CreatedAt:   appliedAt,
		UpdatedAt:   appliedAt,
	}
}

// CorrectionApplyJobID returns the deterministic ingest job ID for one correction apply attempt.
func CorrectionApplyJobID(tenantID, workspaceID, correctionID, targetMemoryID, idempotencyKey string) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{
		strings.TrimSpace(tenantID),
		strings.TrimSpace(workspaceID),
		strings.TrimSpace(correctionID),
		strings.TrimSpace(targetMemoryID),
		strings.TrimSpace(idempotencyKey),
	}, "\x00")))
	return fmt.Sprintf("job_corr_apply_%x", sum[:12])
}

// JobBacklogMetricsRequest scopes read-only worker backlog metrics for operators.
type JobBacklogMetricsRequest struct {
	TenantID     string        `json:"tenant_id,omitempty"`
	WorkspaceID  string        `json:"workspace_id,omitempty"`
	DrainWindow  time.Duration `json:"drain_window,omitempty"`
	GeneratedNow time.Time     `json:"generated_now,omitempty"`
}

// JobStatusCounts summarizes ingest job rows by durable queue status.
type JobStatusCounts struct {
	Queued      int `json:"queued"`
	ReadyQueued int `json:"ready_queued"`
	Running     int `json:"running"`
	Failed      int `json:"failed"`
	Blocked     int `json:"blocked"`
	Complete    int `json:"complete"`
}

// JobBacklogMetrics is a read-only operator view of worker queue health.
type JobBacklogMetrics struct {
	Counts                  JobStatusCounts `json:"counts"`
	OldestQueuedAt          *time.Time      `json:"oldest_queued_at,omitempty"`
	OldestQueuedAgeSeconds  *int64          `json:"oldest_queued_age_seconds,omitempty"`
	OldestRunningAt         *time.Time      `json:"oldest_running_at,omitempty"`
	OldestRunningAgeSeconds *int64          `json:"oldest_running_age_seconds,omitempty"`
	DrainWindowSeconds      int64           `json:"drain_window_seconds"`
	CompletedInWindow       int             `json:"completed_in_window"`
	DrainRateJobsPerMinute  *float64        `json:"drain_rate_jobs_per_minute,omitempty"`
	RecoveryETASeconds      *int64          `json:"recovery_eta_seconds,omitempty"`
	RetryableQueuedAttempts int             `json:"retryable_queued_attempts"`
	GeneratedAt             time.Time       `json:"generated_at"`
}
