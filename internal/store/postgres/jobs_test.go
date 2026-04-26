// ============================================================
// FILE     : internal/store/postgres/jobs_test.go
// PURPOSE  : Verifies PostgreSQL job status update contracts without requiring a live database.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : postgres job status, inspection, and manual requeue tests
// DEPENDS  : context, encoding/json, errors, strings, testing, time, internal/core, github.com/jackc/pgx/v5/pgconn
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: These tests lock SQL status semantics; keep deterministic blocks out of the retry queue.
// ============================================================

package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

func TestListBlockedJobsStatementOnlySelectsBlockedStatus(t *testing.T) {
	t.Parallel()

	sql, args := listBlockedJobsStatement(7)

	if !strings.Contains(sql, "WHERE status = 'blocked'") {
		t.Fatalf("expected blocked status predicate, got: %s", sql)
	}
	if !strings.Contains(sql, "ORDER BY updated_at DESC, created_at DESC") {
		t.Fatalf("expected deterministic newest-first order, got: %s", sql)
	}
	if strings.Contains(sql, "status = 'queued'") {
		t.Fatalf("blocked inspection must not inspect retry queue, got: %s", sql)
	}
	if len(args) != 1 || args[0] != 7 {
		t.Fatalf("unexpected list blocked args: %#v", args)
	}
}

func TestJobBacklogMetricsStatementIsReadOnlyAndSeparatesStatuses(t *testing.T) {
	t.Parallel()

	sql, args := jobBacklogMetricsStatement(&core.JobBacklogMetricsRequest{
		TenantID:     "tenant_1",
		WorkspaceID:  "workspace_1",
		DrainWindow:  15 * time.Minute,
		GeneratedNow: time.Date(2026, time.April, 24, 8, 0, 0, 0, time.UTC),
	})
	upperSQL := strings.ToUpper(sql)
	for _, forbidden := range []string{"UPDATE ", "INSERT ", "DELETE ", "FOR UPDATE", "SKIP LOCKED"} {
		if strings.Contains(upperSQL, forbidden) {
			t.Fatalf("metrics statement must be read-only; found %q in:\n%s", forbidden, sql)
		}
	}
	for _, want := range []string{
		"status = 'queued'",
		"status = 'queued' AND available_at <=",
		"status = 'running'",
		"status = 'failed'",
		"status = 'blocked'",
		"status = 'complete'",
		"available_at <= snapshot.generated_at",
		"status = 'queued' AND attempts > 0",
		"status = 'running'",
		"COALESCE(locked_at, updated_at)",
	} {
		if !strings.Contains(sql, want) {
			t.Fatalf("expected metrics SQL to contain %q, got:\n%s", want, sql)
		}
	}
	if len(args) != 4 || args[0] != "tenant_1" || args[1] != "workspace_1" || args[2] != int64(900) {
		t.Fatalf("unexpected metrics args: %#v", args)
	}
}

func TestNormalizeJobBacklogMetricsRequestDefaultsAndRejectsInvalid(t *testing.T) {
	t.Parallel()

	req, err := normalizeJobBacklogMetricsRequest(nil)
	if err != nil {
		t.Fatalf("normalizeJobBacklogMetricsRequest returned error: %v", err)
	}
	if req.DrainWindow != defaultJobMetricsDrainWindow {
		t.Fatalf("expected default window %s, got %s", defaultJobMetricsDrainWindow, req.DrainWindow)
	}

	_, err = normalizeJobBacklogMetricsRequest(&core.JobBacklogMetricsRequest{DrainWindow: -time.Second})
	if !errors.Is(err, core.ErrInvalidArgument) {
		t.Fatalf("expected invalid argument for negative window, got %v", err)
	}

	_, err = normalizeJobBacklogMetricsRequest(&core.JobBacklogMetricsRequest{DrainWindow: time.Nanosecond})
	if !errors.Is(err, core.ErrInvalidArgument) {
		t.Fatalf("expected invalid argument for sub-second window, got %v", err)
	}

	_, err = normalizeJobBacklogMetricsRequest(&core.JobBacklogMetricsRequest{DrainWindow: 25 * time.Hour})
	if !errors.Is(err, core.ErrInvalidArgument) {
		t.Fatalf("expected invalid argument for oversized window, got %v", err)
	}
}

func TestCorrectionApplyJobStatementCreatesCompletedDeterministicJob(t *testing.T) {
	t.Parallel()

	appliedAt := time.Date(2026, time.April, 25, 7, 0, 0, 0, time.UTC)
	job := core.NewCorrectionApplyJob(core.CorrectionApplyJobInput{
		TenantID:       "tenant_1",
		WorkspaceID:    "workspace_1",
		CorrectionID:   "corr_1",
		TargetMemoryID: "mem_1",
		IdempotencyKey: "correction_1",
		RawEventID:     "evt_correction",
		OperatorID:     "operator_1",
		AppliedAt:      appliedAt,
	})

	sql, args, err := correctionApplyJobStatement(job)
	if err != nil {
		t.Fatalf("correctionApplyJobStatement returned error: %v", err)
	}
	upperSQL := strings.ToUpper(sql)
	for _, want := range []string{
		"INSERT INTO INGEST_JOBS",
		"'COMPLETE'",
		"ON CONFLICT (ID) DO UPDATE",
		"RETURNING ID",
	} {
		if !strings.Contains(upperSQL, want) {
			t.Fatalf("expected correction apply SQL to contain %q, got:\n%s", want, sql)
		}
	}
	if strings.Contains(upperSQL, "DELETE ") {
		t.Fatalf("correction apply job upsert must not delete rows:\n%s", sql)
	}
	wantID := core.CorrectionApplyJobID("tenant_1", "workspace_1", "corr_1", "mem_1", "correction_1")
	if job.ID != wantID {
		t.Fatalf("job id = %q, want %q", job.ID, wantID)
	}
	if len(args) != 9 || args[0] != wantID || args[3] != core.JobKindCorrectionApply {
		t.Fatalf("unexpected correction apply job args: %#v", args)
	}
	rawIDs, ok := args[4].([]string)
	if !ok || len(rawIDs) != 1 || rawIDs[0] != "evt_correction" {
		t.Fatalf("raw event ids not preserved in args: %#v", args[4])
	}
}

func TestCorrectionApplyJobIDUsesRequiredStableInputs(t *testing.T) {
	t.Parallel()

	first := core.CorrectionApplyJobID("tenant_1", "workspace_1", "corr_1", "mem_1", "correction_1")
	second := core.CorrectionApplyJobID("tenant_1", "workspace_1", "corr_1", "mem_1", "correction_1")
	otherCorrection := core.CorrectionApplyJobID("tenant_1", "workspace_1", "corr_2", "mem_1", "correction_1")
	otherTarget := core.CorrectionApplyJobID("tenant_1", "workspace_1", "corr_1", "mem_2", "correction_1")

	if first == "" || first != second {
		t.Fatalf("correction apply job id must be stable, first=%q second=%q", first, second)
	}
	if first == otherCorrection || first == otherTarget {
		t.Fatalf("correction apply job id must include correction and target inputs: base=%q otherCorrection=%q otherTarget=%q", first, otherCorrection, otherTarget)
	}
}

func TestScanJobBacklogMetricsHandlesUnavailableDrainRateAndETA(t *testing.T) {
	t.Parallel()

	generatedAt := time.Date(2026, time.April, 24, 8, 0, 0, 0, time.UTC)
	row := fakeJobMetricsRow{
		values: []any{
			3,
			1,
			1,
			0,
			2,
			10,
			1,
			generatedAt,
			false,
			int64(0),
			generatedAt,
			false,
			int64(0),
			int64(900),
			0,
			float64(0),
			false,
			int64(0),
			false,
			generatedAt,
		},
	}

	metrics, err := scanJobBacklogMetrics(row, 15*time.Minute)
	if err != nil {
		t.Fatalf("scanJobBacklogMetrics returned error: %v", err)
	}
	if metrics.Counts.Queued != 3 || metrics.Counts.ReadyQueued != 1 || metrics.Counts.Blocked != 2 || metrics.RetryableQueuedAttempts != 1 {
		t.Fatalf("unexpected metrics counts: %#v", metrics)
	}
	if metrics.OldestQueuedAt != nil || metrics.OldestQueuedAgeSeconds != nil {
		t.Fatalf("expected unavailable oldest queued fields, got %#v", metrics)
	}
	if metrics.OldestRunningAt != nil || metrics.OldestRunningAgeSeconds != nil {
		t.Fatalf("expected unavailable oldest running fields, got %#v", metrics)
	}
	if metrics.DrainRateJobsPerMinute != nil || metrics.RecoveryETASeconds != nil {
		t.Fatalf("expected unavailable drain rate and ETA, got %#v", metrics)
	}
}

func TestScanJobBacklogMetricsReturnsAvailableRecoveryFields(t *testing.T) {
	t.Parallel()

	generatedAt := time.Date(2026, time.April, 24, 8, 0, 0, 0, time.UTC)
	oldestAt := generatedAt.Add(-5 * time.Minute)
	oldestRunningAt := generatedAt.Add(-7 * time.Minute)
	row := fakeJobMetricsRow{
		values: []any{
			4,
			4,
			1,
			0,
			1,
			20,
			2,
			oldestAt,
			true,
			int64(300),
			oldestRunningAt,
			true,
			int64(420),
			int64(600),
			12,
			float64(1.2),
			true,
			int64(200),
			true,
			generatedAt,
		},
	}

	metrics, err := scanJobBacklogMetrics(row, 10*time.Minute)
	if err != nil {
		t.Fatalf("scanJobBacklogMetrics returned error: %v", err)
	}
	if metrics.OldestQueuedAt == nil || !metrics.OldestQueuedAt.Equal(oldestAt) {
		t.Fatalf("unexpected oldest queued at: %#v", metrics.OldestQueuedAt)
	}
	if metrics.OldestQueuedAgeSeconds == nil || *metrics.OldestQueuedAgeSeconds != 300 {
		t.Fatalf("unexpected oldest queued age: %#v", metrics.OldestQueuedAgeSeconds)
	}
	if metrics.OldestRunningAt == nil || !metrics.OldestRunningAt.Equal(oldestRunningAt) {
		t.Fatalf("unexpected oldest running at: %#v", metrics.OldestRunningAt)
	}
	if metrics.OldestRunningAgeSeconds == nil || *metrics.OldestRunningAgeSeconds != 420 {
		t.Fatalf("unexpected oldest running age: %#v", metrics.OldestRunningAgeSeconds)
	}
	if metrics.DrainRateJobsPerMinute == nil || *metrics.DrainRateJobsPerMinute != 1.2 {
		t.Fatalf("unexpected drain rate: %#v", metrics.DrainRateJobsPerMinute)
	}
	if metrics.RecoveryETASeconds == nil || *metrics.RecoveryETASeconds != 200 {
		t.Fatalf("unexpected recovery ETA: %#v", metrics.RecoveryETASeconds)
	}
}

func TestScanIngestJobRowsReturnsBlockedJobs(t *testing.T) {
	t.Parallel()

	updatedAt := time.Date(2026, time.April, 24, 8, 0, 0, 0, time.UTC)
	lastError := "not implemented: update_memory"
	rows := &fakeIngestJobRows{jobs: []*core.IngestJob{
		{
			ID:          "job_blocked_1",
			TenantID:    "tenant_1",
			WorkspaceID: "workspace_1",
			JobKind:     core.JobKindProcessTurnEvent,
			Status:      "blocked",
			RawEventIDs: []string{"evt_1"},
			PayloadJSON: json.RawMessage(`{"session_id":"session_1"}`),
			Attempts:    3,
			AvailableAt: updatedAt,
			LastError:   &lastError,
			CreatedAt:   updatedAt.Add(-time.Hour),
			UpdatedAt:   updatedAt,
		},
	}}

	jobs, err := scanIngestJobRows(rows, 1)
	if err != nil {
		t.Fatalf("scanIngestJobRows returned error: %v", err)
	}
	if len(jobs) != 1 {
		t.Fatalf("expected one blocked job, got %d", len(jobs))
	}
	if jobs[0].ID != "job_blocked_1" || jobs[0].Status != "blocked" {
		t.Fatalf("unexpected blocked job: %#v", jobs[0])
	}
	if jobs[0].LastError == nil || *jobs[0].LastError != lastError {
		t.Fatalf("expected last_error to survive inspection, got %#v", jobs[0].LastError)
	}
	if !rows.closed {
		t.Fatalf("expected rows to be closed after scanning")
	}
}

func TestRequeueBlockedJobRequiresBlockedStatus(t *testing.T) {
	t.Parallel()

	exec := &recordingJobExecutor{tag: pgconn.NewCommandTag("UPDATE 1")}

	if err := requeueBlockedJob(context.Background(), exec, "job_blocked_1", "apply support landed"); err != nil {
		t.Fatalf("requeueBlockedJob returned error: %v", err)
	}

	if !strings.Contains(exec.sql, "status = 'queued'") {
		t.Fatalf("expected requeue to set queued status, got: %s", exec.sql)
	}
	if !strings.Contains(exec.sql, "WHERE id = $1 AND status = 'blocked'") {
		t.Fatalf("expected requeue to require currently blocked job, got: %s", exec.sql)
	}
	if strings.Contains(exec.sql, "attempts = attempts + 1") {
		t.Fatalf("manual requeue must not increment attempts, got: %s", exec.sql)
	}
	if strings.Contains(exec.sql, "interval '30 seconds'") {
		t.Fatalf("manual requeue must not schedule retry interval, got: %s", exec.sql)
	}
	if !strings.Contains(exec.sql, "operator requeue reason") {
		t.Fatalf("expected requeue reason to be preserved in last_error, got: %s", exec.sql)
	}
	if len(exec.args) != 2 || exec.args[0] != "job_blocked_1" || exec.args[1] != "apply support landed" {
		t.Fatalf("unexpected requeue args: %#v", exec.args)
	}
}

func TestRequeueBlockedJobReturnsNotFoundWhenJobIsNotBlocked(t *testing.T) {
	t.Parallel()

	exec := &recordingJobExecutor{tag: pgconn.NewCommandTag("UPDATE 0")}

	err := requeueBlockedJob(context.Background(), exec, "job_not_blocked", "")
	if !errors.Is(err, core.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestBlockJobMarksJobBlockedWithoutRetry(t *testing.T) {
	t.Parallel()

	exec := &recordingJobExecutor{tag: pgconn.NewCommandTag("UPDATE 1")}
	jobErr := errors.New("unsupported apply work")

	if err := blockJob(context.Background(), exec, "job_1", jobErr); err != nil {
		t.Fatalf("blockJob returned error: %v", err)
	}

	if !strings.Contains(exec.sql, "status = 'blocked'") {
		t.Fatalf("expected blocked status update SQL, got: %s", exec.sql)
	}
	if strings.Contains(exec.sql, "interval '30 seconds'") {
		t.Fatalf("blocked jobs must not schedule the retry interval, got: %s", exec.sql)
	}
	if len(exec.args) != 2 || exec.args[0] != "job_1" || exec.args[1] != jobErr.Error() {
		t.Fatalf("unexpected blockJob args: %#v", exec.args)
	}
}

func TestFailJobSchedulesRetry(t *testing.T) {
	t.Parallel()

	exec := &recordingJobExecutor{tag: pgconn.NewCommandTag("UPDATE 1")}

	if err := failJob(context.Background(), exec, "job_1", errors.New("transient apply failure")); err != nil {
		t.Fatalf("failJob returned error: %v", err)
	}

	if !strings.Contains(exec.sql, "status = 'queued'") {
		t.Fatalf("expected queued retry SQL, got: %s", exec.sql)
	}
	if !strings.Contains(exec.sql, "interval '30 seconds'") {
		t.Fatalf("expected retry interval SQL, got: %s", exec.sql)
	}
}

func TestBlockJobReturnsNotFoundWhenNoRowsUpdate(t *testing.T) {
	t.Parallel()

	exec := &recordingJobExecutor{tag: pgconn.NewCommandTag("UPDATE 0")}

	err := blockJob(context.Background(), exec, "missing_job", errors.New("unsupported apply work"))
	if !errors.Is(err, core.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

type fakeIngestJobRows struct {
	jobs   []*core.IngestJob
	idx    int
	closed bool
	err    error
}

func (r *fakeIngestJobRows) Close() {
	r.closed = true
}

func (r *fakeIngestJobRows) Next() bool {
	if r.idx >= len(r.jobs) {
		return false
	}
	r.idx++
	return true
}

func (r *fakeIngestJobRows) Scan(dest ...any) error {
	job := r.jobs[r.idx-1]
	assignScannedJobValue(dest[0], job.ID)
	assignScannedJobValue(dest[1], job.TenantID)
	assignScannedJobValue(dest[2], job.WorkspaceID)
	assignScannedJobValue(dest[3], job.JobKind)
	assignScannedJobValue(dest[4], job.Status)
	assignScannedJobValue(dest[5], job.RawEventIDs)
	assignScannedJobValue(dest[6], job.PayloadJSON)
	assignScannedJobValue(dest[7], job.Attempts)
	assignScannedJobValue(dest[8], job.AvailableAt)
	assignScannedJobValue(dest[9], job.LockedBy)
	assignScannedJobValue(dest[10], job.LockedAt)
	assignScannedJobValue(dest[11], job.LastError)
	assignScannedJobValue(dest[12], job.CreatedAt)
	assignScannedJobValue(dest[13], job.UpdatedAt)
	return nil
}

func (r *fakeIngestJobRows) Err() error {
	return r.err
}

func assignScannedJobValue(dest any, value any) {
	switch target := dest.(type) {
	case *string:
		*target = value.(string)
	case *core.JobKind:
		*target = value.(core.JobKind)
	case *[]string:
		*target = append([]string(nil), value.([]string)...)
	case *json.RawMessage:
		*target = append(json.RawMessage(nil), value.(json.RawMessage)...)
	case *int:
		*target = value.(int)
	case *time.Time:
		*target = value.(time.Time)
	case *bool:
		*target = value.(bool)
	case *int64:
		*target = value.(int64)
	case *float64:
		*target = value.(float64)
	case **string:
		*target = value.(*string)
	case **time.Time:
		*target = value.(*time.Time)
	}
}

type fakeJobMetricsRow struct {
	values []any
	err    error
}

func (r fakeJobMetricsRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	for i := range dest {
		assignScannedJobValue(dest[i], r.values[i])
	}
	return nil
}

type recordingJobExecutor struct {
	sql  string
	args []any
	tag  pgconn.CommandTag
	err  error
}

func (e *recordingJobExecutor) Exec(_ context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	e.sql = sql
	e.args = append([]any(nil), arguments...)
	return e.tag, e.err
}
