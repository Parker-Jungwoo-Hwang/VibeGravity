// ============================================================
// FILE     : internal/store/postgres/jobs.go
// PURPOSE  : Implements PostgreSQL-backed ingest job queue operations.
// LAYER    : infra
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : EnqueueJobs, EnsureCorrectionApplyJob, ClaimJobs, CompleteJob, FailJob, BlockJob, GetJobBacklogMetrics, ListBlockedJobs, RequeueBlockedJob
// DEPENDS  : internal/core, github.com/jackc/pgx/v5/pgconn
// USED_BY  : internal/ingest, cmd/worker, cmd/cli
// ------------------------------------------------------------
// AGENT_NOTE: Job claiming must use database locking to avoid duplicate worker ownership.
// ============================================================

package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

const defaultBlockedJobListLimit = 20
const defaultJobMetricsDrainWindow = 15 * time.Minute
const maxJobMetricsDrainWindow = 24 * time.Hour

type jobExecutor interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

type ingestJobRows interface {
	Close()
	Next() bool
	Scan(dest ...any) error
	Err() error
}

type jobMetricsScanner interface {
	Scan(dest ...any) error
}

// EnsureCorrectionApplyJob creates or reuses the completed job row used by correction supersession provenance.
func (s *Store) EnsureCorrectionApplyJob(ctx context.Context, job *core.IngestJob) (string, error) {
	sql, args, err := correctionApplyJobStatement(job)
	if err != nil {
		return "", err
	}
	var id string
	if err := s.pool.QueryRow(ctx, sql, args...).Scan(&id); err != nil {
		return "", fmt.Errorf("ensure correction apply job: %w", err)
	}
	return id, nil
}

func correctionApplyJobStatement(job *core.IngestJob) (string, []any, error) {
	if job == nil {
		return "", nil, fmt.Errorf("%w: correction apply job is required", core.ErrInvalidArgument)
	}
	if strings.TrimSpace(job.ID) == "" {
		return "", nil, fmt.Errorf("%w: correction apply job id is required", core.ErrInvalidArgument)
	}
	if job.JobKind != core.JobKindCorrectionApply {
		return "", nil, fmt.Errorf("%w: correction apply job kind is required", core.ErrInvalidArgument)
	}
	return `
		INSERT INTO ingest_jobs (
			id, tenant_id, workspace_id, job_kind, status, raw_event_ids,
			payload_json, attempts, available_at, locked_by, locked_at,
			last_error, created_at, updated_at
		)
		VALUES ($1,$2,$3,$4,'complete',$5,$6,0,$7,NULL,NULL,NULL,$8,$9)
		ON CONFLICT (id) DO UPDATE
		SET status = 'complete',
		    raw_event_ids = EXCLUDED.raw_event_ids,
		    payload_json = EXCLUDED.payload_json,
		    locked_by = NULL,
		    locked_at = NULL,
		    last_error = NULL,
		    updated_at = EXCLUDED.updated_at
		RETURNING id
	`, []any{
			job.ID,
			job.TenantID,
			job.WorkspaceID,
			job.JobKind,
			job.RawEventIDs,
			rawJSONOrEmpty(job.PayloadJSON),
			timeOrNow(job.AvailableAt),
			timeOrNow(job.CreatedAt),
			timeOrNow(job.UpdatedAt),
		}, nil
}

// EnqueueJobs inserts queued jobs and returns their IDs.
func (s *Store) EnqueueJobs(ctx context.Context, jobs []*core.IngestJob) ([]string, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin job enqueue: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	ids := make([]string, 0, len(jobs))
	for _, job := range jobs {
		if job == nil {
			continue
		}
		id := job.ID
		if id == "" {
			id, err = newID("job")
			if err != nil {
				return nil, err
			}
		}
		row := tx.QueryRow(ctx, `
			INSERT INTO ingest_jobs (
				id, tenant_id, workspace_id, job_kind, status, raw_event_ids,
				payload_json, attempts, available_at, locked_by, locked_at,
				last_error, created_at, updated_at
			)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
			RETURNING id
		`, id, job.TenantID, job.WorkspaceID, job.JobKind, valueOr(job.Status, "queued"),
			job.RawEventIDs, rawJSONOrEmpty(job.PayloadJSON), job.Attempts, timeOrNow(job.AvailableAt),
			job.LockedBy, job.LockedAt, job.LastError, timeOrNow(job.CreatedAt), timeOrNow(job.UpdatedAt))
		var insertedID string
		if err := row.Scan(&insertedID); err != nil {
			return nil, fmt.Errorf("insert ingest job: %w", err)
		}
		ids = append(ids, insertedID)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit job enqueue: %w", err)
	}
	return ids, nil
}

// ClaimJobs claims available jobs for one worker using FOR UPDATE SKIP LOCKED.
func (s *Store) ClaimJobs(ctx context.Context, workerID string, limit int) ([]*core.IngestJob, error) {
	if limit <= 0 {
		limit = 1
	}
	rows, err := s.pool.Query(ctx, `
		WITH claimed AS (
			SELECT id
			FROM ingest_jobs
			WHERE status = 'queued' AND available_at <= now()
			ORDER BY available_at ASC, created_at ASC
			LIMIT $2
			FOR UPDATE SKIP LOCKED
		)
		UPDATE ingest_jobs
		SET status = 'running',
		    locked_by = $1,
		    locked_at = now(),
		    updated_at = now()
		WHERE id IN (SELECT id FROM claimed)
		RETURNING id, tenant_id, workspace_id, job_kind, status, raw_event_ids,
		          payload_json, attempts, available_at, locked_by, locked_at,
		          last_error, created_at, updated_at
	`, workerID, limit)
	if err != nil {
		return nil, fmt.Errorf("claim jobs: %w", err)
	}
	jobs, err := scanIngestJobRows(rows, limit)
	if err != nil {
		return nil, err
	}
	return jobs, nil
}

// ListBlockedJobs lists jobs blocked by deterministic unsupported work for operator inspection.
func (s *Store) ListBlockedJobs(ctx context.Context, limit int) ([]*core.IngestJob, error) {
	limit = normalizeBlockedJobListLimit(limit)
	sql, args := listBlockedJobsStatement(limit)
	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("list blocked jobs: %w", err)
	}
	return scanIngestJobRows(rows, limit)
}

func listBlockedJobsStatement(limit int) (string, []any) {
	limit = normalizeBlockedJobListLimit(limit)
	return `
		SELECT id, tenant_id, workspace_id, job_kind, status, raw_event_ids,
		       payload_json, attempts, available_at, locked_by, locked_at,
		       last_error, created_at, updated_at
		FROM ingest_jobs
		WHERE status = 'blocked'
		ORDER BY updated_at DESC, created_at DESC
		LIMIT $1
	`, []any{limit}
}

func normalizeBlockedJobListLimit(limit int) int {
	if limit <= 0 {
		return defaultBlockedJobListLimit
	}
	return limit
}

// GetJobBacklogMetrics returns read-only worker queue status and recovery estimates.
func (s *Store) GetJobBacklogMetrics(ctx context.Context, req *core.JobBacklogMetricsRequest) (*core.JobBacklogMetrics, error) {
	normalized, err := normalizeJobBacklogMetricsRequest(req)
	if err != nil {
		return nil, err
	}
	sql, args := jobBacklogMetricsStatement(normalized)
	metrics, err := scanJobBacklogMetrics(s.pool.QueryRow(ctx, sql, args...), normalized.DrainWindow)
	if err != nil {
		return nil, fmt.Errorf("get job backlog metrics: %w", err)
	}
	return metrics, nil
}

func normalizeJobBacklogMetricsRequest(req *core.JobBacklogMetricsRequest) (*core.JobBacklogMetricsRequest, error) {
	normalized := &core.JobBacklogMetricsRequest{}
	if req != nil {
		*normalized = *req
	}
	normalized.TenantID = strings.TrimSpace(normalized.TenantID)
	normalized.WorkspaceID = strings.TrimSpace(normalized.WorkspaceID)
	if normalized.DrainWindow == 0 {
		normalized.DrainWindow = defaultJobMetricsDrainWindow
	}
	if normalized.DrainWindow < time.Second {
		return nil, fmt.Errorf("%w: job metrics drain window must be at least 1s", core.ErrInvalidArgument)
	}
	if normalized.DrainWindow > maxJobMetricsDrainWindow {
		return nil, fmt.Errorf("%w: job metrics drain window must be at most 24h", core.ErrInvalidArgument)
	}
	if normalized.GeneratedNow.IsZero() {
		normalized.GeneratedNow = time.Now().UTC()
	}
	return normalized, nil
}

func jobBacklogMetricsStatement(req *core.JobBacklogMetricsRequest) (string, []any) {
	windowSeconds := int64(req.DrainWindow.Seconds())
	return `
		WITH snapshot AS (
			SELECT $4::timestamptz AS generated_at
		),
		filtered AS (
			SELECT *
			FROM ingest_jobs
			WHERE ($1 = '' OR tenant_id = $1)
			  AND ($2 = '' OR workspace_id = $2)
		),
		counts AS (
			SELECT
				count(*) FILTER (WHERE status = 'queued')::int AS queued,
				count(*) FILTER (WHERE status = 'queued' AND available_at <= (SELECT generated_at FROM snapshot))::int AS ready_queued,
				count(*) FILTER (WHERE status = 'running')::int AS running,
				count(*) FILTER (WHERE status = 'failed')::int AS failed,
				count(*) FILTER (WHERE status = 'blocked')::int AS blocked,
				count(*) FILTER (WHERE status = 'complete')::int AS complete,
				count(*) FILTER (WHERE status = 'queued' AND attempts > 0)::int AS retryable_queued_attempts
			FROM filtered
		),
		oldest AS (
			SELECT min(available_at) AS oldest_queued_at
			FROM filtered, snapshot
			WHERE status = 'queued'
			  AND available_at <= snapshot.generated_at
		),
		oldest_running AS (
			SELECT min(COALESCE(locked_at, updated_at)) AS oldest_running_at
			FROM filtered
			WHERE status = 'running'
		),
		drain AS (
			SELECT count(*)::int AS completed_in_window
			FROM filtered, snapshot
			WHERE status = 'complete'
			  AND updated_at >= snapshot.generated_at - ($3::bigint * interval '1 second')
		)
		SELECT
			counts.queued,
			counts.ready_queued,
			counts.running,
			counts.failed,
			counts.blocked,
			counts.complete,
			counts.retryable_queued_attempts,
			COALESCE(oldest.oldest_queued_at, snapshot.generated_at) AS oldest_queued_at,
			oldest.oldest_queued_at IS NOT NULL AS has_oldest_queued_at,
			CASE
				WHEN oldest.oldest_queued_at IS NULL THEN 0
				ELSE GREATEST(0, extract(epoch FROM (snapshot.generated_at - oldest.oldest_queued_at)))::bigint
			END AS oldest_queued_age_seconds,
			COALESCE(oldest_running.oldest_running_at, snapshot.generated_at) AS oldest_running_at,
			oldest_running.oldest_running_at IS NOT NULL AS has_oldest_running_at,
			CASE
				WHEN oldest_running.oldest_running_at IS NULL THEN 0
				ELSE GREATEST(0, extract(epoch FROM (snapshot.generated_at - oldest_running.oldest_running_at)))::bigint
			END AS oldest_running_age_seconds,
			$3::bigint AS drain_window_seconds,
			drain.completed_in_window,
			CASE
				WHEN drain.completed_in_window = 0 THEN 0
				ELSE drain.completed_in_window::double precision / ($3::double precision / 60.0)
			END AS drain_rate_jobs_per_minute,
			drain.completed_in_window > 0 AS has_drain_rate,
			CASE
				WHEN counts.queued = 0 THEN 0
				WHEN drain.completed_in_window = 0 THEN 0
				ELSE ceil(counts.queued::double precision / (drain.completed_in_window::double precision / $3::double precision))::bigint
			END AS recovery_eta_seconds,
			(counts.queued = 0 OR drain.completed_in_window > 0) AS has_recovery_eta,
			snapshot.generated_at
		FROM snapshot, counts, oldest, oldest_running, drain
	`, []any{req.TenantID, req.WorkspaceID, windowSeconds, req.GeneratedNow.UTC()}
}

func scanJobBacklogMetrics(row jobMetricsScanner, _ time.Duration) (*core.JobBacklogMetrics, error) {
	metrics := &core.JobBacklogMetrics{}
	var oldestQueuedAt time.Time
	var hasOldestQueuedAt bool
	var oldestQueuedAgeSeconds int64
	var oldestRunningAt time.Time
	var hasOldestRunningAt bool
	var oldestRunningAgeSeconds int64
	var drainRateJobsPerMinute float64
	var hasDrainRate bool
	var recoveryETASeconds int64
	var hasRecoveryETA bool
	err := row.Scan(
		&metrics.Counts.Queued,
		&metrics.Counts.ReadyQueued,
		&metrics.Counts.Running,
		&metrics.Counts.Failed,
		&metrics.Counts.Blocked,
		&metrics.Counts.Complete,
		&metrics.RetryableQueuedAttempts,
		&oldestQueuedAt,
		&hasOldestQueuedAt,
		&oldestQueuedAgeSeconds,
		&oldestRunningAt,
		&hasOldestRunningAt,
		&oldestRunningAgeSeconds,
		&metrics.DrainWindowSeconds,
		&metrics.CompletedInWindow,
		&drainRateJobsPerMinute,
		&hasDrainRate,
		&recoveryETASeconds,
		&hasRecoveryETA,
		&metrics.GeneratedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, core.ErrNotFound
		}
		return nil, err
	}
	if hasOldestQueuedAt {
		metrics.OldestQueuedAt = &oldestQueuedAt
		metrics.OldestQueuedAgeSeconds = &oldestQueuedAgeSeconds
	}
	if hasOldestRunningAt {
		metrics.OldestRunningAt = &oldestRunningAt
		metrics.OldestRunningAgeSeconds = &oldestRunningAgeSeconds
	}
	if hasDrainRate {
		metrics.DrainRateJobsPerMinute = &drainRateJobsPerMinute
	}
	if hasRecoveryETA {
		metrics.RecoveryETASeconds = &recoveryETASeconds
	}
	return metrics, nil
}

// CompleteJob marks a job complete.
func (s *Store) CompleteJob(ctx context.Context, jobID string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE ingest_jobs
		SET status = 'complete',
		    locked_by = NULL,
		    locked_at = NULL,
		    updated_at = now()
		WHERE id = $1
	`, jobID)
	if err != nil {
		return fmt.Errorf("complete job: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return core.ErrNotFound
	}
	return nil
}

// FailJob records a failed attempt and schedules retry.
func (s *Store) FailJob(ctx context.Context, jobID string, jobErr error) error {
	return failJob(ctx, s.pool, jobID, jobErr)
}

func failJob(ctx context.Context, exec jobExecutor, jobID string, jobErr error) error {
	tag, err := exec.Exec(ctx, `
		UPDATE ingest_jobs
		SET status = 'queued',
		    attempts = attempts + 1,
		    available_at = now() + interval '30 seconds',
		    locked_by = NULL,
		    locked_at = NULL,
		    last_error = $2,
		    updated_at = now()
		WHERE id = $1
	`, jobID, jobErrorString(jobErr))
	if err != nil {
		return fmt.Errorf("fail job: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return core.ErrNotFound
	}
	return nil
}

// BlockJob records deterministic unsupported work without scheduling retry.
func (s *Store) BlockJob(ctx context.Context, jobID string, jobErr error) error {
	return blockJob(ctx, s.pool, jobID, jobErr)
}

func blockJob(ctx context.Context, exec jobExecutor, jobID string, jobErr error) error {
	tag, err := exec.Exec(ctx, `
		UPDATE ingest_jobs
		SET status = 'blocked',
		    attempts = attempts + 1,
		    available_at = now(),
		    locked_by = NULL,
		    locked_at = NULL,
		    last_error = $2,
		    updated_at = now()
		WHERE id = $1
	`, jobID, jobErrorString(jobErr))
	if err != nil {
		return fmt.Errorf("block job: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return core.ErrNotFound
	}
	return nil
}

func jobErrorString(jobErr error) string {
	if jobErr == nil {
		return ""
	}
	return jobErr.Error()
}

// RequeueBlockedJob manually returns one blocked job to the queued worker pool.
func (s *Store) RequeueBlockedJob(ctx context.Context, jobID string, reason string) error {
	return requeueBlockedJob(ctx, s.pool, jobID, reason)
}

func requeueBlockedJob(ctx context.Context, exec jobExecutor, jobID string, reason string) error {
	if jobID == "" {
		return fmt.Errorf("%w: blocked job id is required", core.ErrInvalidArgument)
	}
	reason = strings.TrimSpace(reason)
	tag, err := exec.Exec(ctx, `
		UPDATE ingest_jobs
		SET status = 'queued',
		    available_at = now(),
		    locked_by = NULL,
		    locked_at = NULL,
		    last_error = CASE
		        WHEN $2 = '' THEN last_error
		        WHEN last_error IS NULL OR last_error = '' THEN 'operator requeue reason: ' || $2
		        ELSE last_error || E'\noperator requeue reason: ' || $2
		    END,
		    updated_at = now()
		WHERE id = $1 AND status = 'blocked'
	`, jobID, reason)
	if err != nil {
		return fmt.Errorf("requeue blocked job: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return core.ErrNotFound
	}
	return nil
}

type ingestJobScanner interface {
	Scan(dest ...any) error
}

func scanIngestJobRows(rows ingestJobRows, capacity int) ([]*core.IngestJob, error) {
	defer rows.Close()
	jobs := make([]*core.IngestJob, 0, capacity)
	for rows.Next() {
		job, err := scanIngestJob(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate ingest jobs: %w", err)
	}
	return jobs, nil
}

func scanIngestJob(row ingestJobScanner) (*core.IngestJob, error) {
	job := &core.IngestJob{}
	if err := row.Scan(&job.ID, &job.TenantID, &job.WorkspaceID, &job.JobKind, &job.Status,
		&job.RawEventIDs, &job.PayloadJSON, &job.Attempts, &job.AvailableAt, &job.LockedBy,
		&job.LockedAt, &job.LastError, &job.CreatedAt, &job.UpdatedAt); err != nil {
		return nil, fmt.Errorf("scan ingest job: %w", err)
	}
	return job, nil
}

func valueOr(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
