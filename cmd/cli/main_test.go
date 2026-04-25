// ============================================================
// FILE     : cmd/cli/main_test.go
// PURPOSE  : Verifies operator CLI commands for blocked job inspection, MCP serving, and Hermes bootstrap.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : CLI tests
// DEPENDS  : bytes, context, strings, testing, time, internal/core
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: CLI tests must not open a real database or call external services.
// ============================================================

package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

func TestMaskPasswordRedactsURLCredentials(t *testing.T) {
	t.Parallel()

	got := maskPassword("postgres://vibe:super-secret@localhost:5432/vibegravity?sslmode=disable")
	if strings.Contains(got, "super-secret") {
		t.Fatalf("database URL password was not redacted: %s", got)
	}
	if !strings.Contains(got, "postgres://vibe:xxxxx@localhost:5432/vibegravity") {
		t.Fatalf("database URL should preserve non-secret connection shape, got %s", got)
	}
}

func TestMaskPasswordRedactsKeywordDSNPassword(t *testing.T) {
	t.Parallel()

	got := maskPassword("host=localhost user=vibe password=super-secret dbname=vibegravity")
	if strings.Contains(got, "super-secret") {
		t.Fatalf("keyword DSN password was not redacted: %s", got)
	}
	if !strings.Contains(got, "password=xxxxx") {
		t.Fatalf("keyword DSN password placeholder missing: %s", got)
	}
}

func TestRunCLIListsBlockedJobs(t *testing.T) {
	t.Parallel()

	lastError := "not implemented: update_memory"
	store := &fakeBlockedJobStore{jobs: []*core.IngestJob{
		{
			ID:          "job_blocked_1",
			TenantID:    "tenant_1",
			WorkspaceID: "workspace_1",
			JobKind:     core.JobKindProcessTurnEvent,
			Status:      "blocked",
			Attempts:    4,
			LastError:   &lastError,
			UpdatedAt:   time.Date(2026, time.April, 24, 8, 0, 0, 0, time.UTC),
		},
	}}
	var out bytes.Buffer

	code := runCLI(context.Background(), []string{"jobs", "blocked", "--limit", "3"}, nil, &out, fakeStoreFactory(store), fakeServiceFactory(&fakeCLIService{}))

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; output: %s", code, out.String())
	}
	if store.listLimit != 3 {
		t.Fatalf("expected list limit 3, got %d", store.listLimit)
	}
	output := out.String()
	if !strings.Contains(output, "job_blocked_1") || !strings.Contains(output, "not implemented: update_memory") {
		t.Fatalf("expected blocked job details in output, got: %s", output)
	}
}

func TestRunCLIPrintsJobMetrics(t *testing.T) {
	t.Parallel()

	drainRate := 2.5
	recoveryETA := int64(240)
	oldestAge := int64(125)
	oldestRunningAge := int64(360)
	oldestAt := time.Date(2026, time.April, 24, 8, 10, 0, 0, time.UTC)
	oldestRunningAt := time.Date(2026, time.April, 24, 8, 6, 5, 0, time.UTC)
	store := &fakeBlockedJobStore{metrics: &core.JobBacklogMetrics{
		Counts: core.JobStatusCounts{
			Queued:      10,
			ReadyQueued: 7,
			Running:     2,
			Failed:      0,
			Blocked:     3,
			Complete:    50,
		},
		OldestQueuedAt:          &oldestAt,
		OldestQueuedAgeSeconds:  &oldestAge,
		OldestRunningAt:         &oldestRunningAt,
		OldestRunningAgeSeconds: &oldestRunningAge,
		DrainWindowSeconds:      600,
		CompletedInWindow:       25,
		DrainRateJobsPerMinute:  &drainRate,
		RecoveryETASeconds:      &recoveryETA,
		RetryableQueuedAttempts: 4,
		GeneratedAt:             time.Date(2026, time.April, 24, 8, 12, 5, 0, time.UTC),
	}}
	var out bytes.Buffer

	code := runCLI(context.Background(), []string{"jobs", "metrics", "--window", "10m", "--tenant", "tenant_1", "--workspace", "workspace_1"}, nil, &out, fakeStoreFactory(store), fakeServiceFactory(&fakeCLIService{}))

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; output: %s", code, out.String())
	}
	if store.metricsReq == nil {
		t.Fatalf("expected metrics request")
	}
	if store.metricsReq.DrainWindow != 10*time.Minute || store.metricsReq.TenantID != "tenant_1" || store.metricsReq.WorkspaceID != "workspace_1" {
		t.Fatalf("unexpected metrics request: %#v", store.metricsReq)
	}
	if store.listLimit != 0 || store.requeuedJobID != "" {
		t.Fatalf("jobs metrics must be read-only, got listLimit=%d requeued=%q", store.listLimit, store.requeuedJobID)
	}
	output := out.String()
	for _, want := range []string{
		"JOB BACKLOG",
		"queued:       10",
		"ready queued: 7",
		"running:      2",
		"blocked:      3",
		"retryable queued attempts: 4",
		"oldest running age: 6m0s",
		"drain rate:        2.50 jobs/min",
		"recovery ETA:      4m0s",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected output to contain %q, got: %s", want, output)
		}
	}
}

func TestRunCLIJobMetricsDefaultsWindow(t *testing.T) {
	t.Parallel()

	store := &fakeBlockedJobStore{metrics: &core.JobBacklogMetrics{}}
	var out bytes.Buffer

	code := runCLI(context.Background(), []string{"jobs", "metrics"}, nil, &out, fakeStoreFactory(store), fakeServiceFactory(&fakeCLIService{}))

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; output: %s", code, out.String())
	}
	if store.metricsReq == nil || store.metricsReq.DrainWindow != 15*time.Minute {
		t.Fatalf("expected default 15m metrics window, got %#v", store.metricsReq)
	}
}

func TestRunCLIRejectsInvalidMetricsWindow(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	code := runCLI(context.Background(), []string{"jobs", "metrics", "--window", "0s"}, nil, &out, fakeStoreFactory(&fakeBlockedJobStore{}), fakeServiceFactory(&fakeCLIService{}))

	if code == 0 {
		t.Fatalf("expected non-zero exit for invalid metrics window")
	}
	if !strings.Contains(out.String(), "window must be at least 1s") {
		t.Fatalf("expected invalid window message, got: %s", out.String())
	}
}

func TestRunCLIRequeuesBlockedJob(t *testing.T) {
	t.Parallel()

	store := &fakeBlockedJobStore{}
	var out bytes.Buffer

	code := runCLI(context.Background(), []string{"jobs", "requeue-blocked", "job_blocked_1"}, nil, &out, fakeStoreFactory(store), fakeServiceFactory(&fakeCLIService{}))

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; output: %s", code, out.String())
	}
	if store.requeuedJobID != "job_blocked_1" {
		t.Fatalf("expected job_blocked_1 to be requeued, got %q", store.requeuedJobID)
	}
	if !strings.Contains(out.String(), "requeued blocked job job_blocked_1") {
		t.Fatalf("unexpected output: %s", out.String())
	}
}

func TestRunCLIRejectsInvalidBlockedJobLimit(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	code := runCLI(context.Background(), []string{"jobs", "blocked", "--limit", "0"}, nil, &out, fakeStoreFactory(&fakeBlockedJobStore{}), fakeServiceFactory(&fakeCLIService{}))

	if code == 0 {
		t.Fatalf("expected non-zero exit for invalid limit")
	}
	if !strings.Contains(out.String(), "limit must be greater than 0") {
		t.Fatalf("expected invalid limit message, got: %s", out.String())
	}
}

func TestRunCLIReportsRequeueStoreError(t *testing.T) {
	t.Parallel()

	store := &fakeBlockedJobStore{requeueErr: core.ErrNotFound}
	var out bytes.Buffer

	code := runCLI(context.Background(), []string{"jobs", "requeue-blocked", "missing_job"}, nil, &out, fakeStoreFactory(store), fakeServiceFactory(&fakeCLIService{}))

	if code == 0 {
		t.Fatalf("expected non-zero exit for missing blocked job")
	}
	if !strings.Contains(out.String(), "blocked job not found") {
		t.Fatalf("expected not found message, got: %s", out.String())
	}
}

func TestRunCLIServesMCPStdio(t *testing.T) {
	t.Parallel()

	input := strings.NewReader(strings.Join([]string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"test","version":"0"}}}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`,
		"",
	}, "\n"))
	var out bytes.Buffer

	code := runCLI(context.Background(), []string{"mcp", "serve", "--stdio"}, input, &out, fakeStoreFactory(&fakeBlockedJobStore{}), fakeServiceFactory(&fakeCLIService{}))

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; output: %s", code, out.String())
	}
	if !strings.Contains(out.String(), `"protocolVersion":"2025-11-25"`) || !strings.Contains(out.String(), `"tools"`) {
		t.Fatalf("expected MCP initialize and tools/list responses, got: %s", out.String())
	}
}

func TestRunCLIRunsHermesMemoryDemoEval(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	code := runCLI(context.Background(), []string{"eval", "demo"}, strings.NewReader(""), &out, fakeStoreFactory(&fakeBlockedJobStore{}), fakeServiceFactory(&fakeCLIService{}))

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; output: %s", code, out.String())
	}
	output := out.String()
	for _, want := range []string{
		"demo initial recall shows rule plan and trust metadata",
		"demo explain shows recalled memory provenance",
		"demo next recall uses correction",
		"demo private scope separation",
		"Hermes Memory demo eval passed.",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected demo eval output to contain %q, got: %s", want, output)
		}
	}
}

func TestRunCLIHermesBootstrapPrintsRegistrationCommand(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	code := runCLI(context.Background(), []string{"hermes", "bootstrap", "--name", "vg", "--command", "/tmp/vibe cli"}, strings.NewReader(""), &out, fakeStoreFactory(&fakeBlockedJobStore{}), fakeServiceFactory(&fakeCLIService{}))

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; output: %s", code, out.String())
	}
	output := out.String()
	if !strings.Contains(output, "hermes mcp add vg --command '/tmp/vibe cli' --args mcp serve --stdio") {
		t.Fatalf("expected Hermes registration command, got: %s", output)
	}
	if !strings.Contains(output, "hermes mcp test vg") {
		t.Fatalf("expected verification command, got: %s", output)
	}
}

type fakeBlockedJobStore struct {
	jobs          []*core.IngestJob
	listLimit     int
	listErr       error
	requeuedJobID string
	requeueErr    error
	metrics       *core.JobBacklogMetrics
	metricsReq    *core.JobBacklogMetricsRequest
	metricsErr    error
}

func (s *fakeBlockedJobStore) ListBlockedJobs(_ context.Context, limit int) ([]*core.IngestJob, error) {
	s.listLimit = limit
	if s.listErr != nil {
		return nil, s.listErr
	}
	return s.jobs, nil
}

func (s *fakeBlockedJobStore) RequeueBlockedJob(_ context.Context, jobID string) error {
	s.requeuedJobID = jobID
	return s.requeueErr
}

func (s *fakeBlockedJobStore) GetJobBacklogMetrics(_ context.Context, req *core.JobBacklogMetricsRequest) (*core.JobBacklogMetrics, error) {
	s.metricsReq = req
	if s.metricsErr != nil {
		return nil, s.metricsErr
	}
	if s.metrics != nil {
		return s.metrics, nil
	}
	return &core.JobBacklogMetrics{}, nil
}

func fakeStoreFactory(store *fakeBlockedJobStore) jobOperatorStoreFactory {
	return func(context.Context) (jobOperatorStore, func(), error) {
		return store, func() {}, nil
	}
}

type fakeCLIService struct{}

func (s *fakeCLIService) Prefetch(context.Context, *core.PrefetchRequest) (*core.PrefetchResponse, error) {
	return &core.PrefetchResponse{Blocks: []core.RecallBlock{{Kind: "note", Text: "mcp cli ok"}}}, nil
}

func (s *fakeCLIService) SyncTurn(context.Context, *core.SyncTurnRequest) (*core.SyncTurnResponse, error) {
	return &core.SyncTurnResponse{Status: "accepted"}, nil
}

func (s *fakeCLIService) AddDocument(context.Context, *core.AddDocumentRequest) (*core.AddDocumentResponse, error) {
	return &core.AddDocumentResponse{Status: "created"}, nil
}

func (s *fakeCLIService) SearchMemories(context.Context, *core.SearchMemoriesRequest) (*core.SearchMemoriesResponse, error) {
	return &core.SearchMemoriesResponse{}, nil
}

func (s *fakeCLIService) SearchDocuments(context.Context, *core.SearchDocumentsRequest) (*core.SearchDocumentsResponse, error) {
	return &core.SearchDocumentsResponse{}, nil
}

func (s *fakeCLIService) AddNote(context.Context, *core.AddNoteRequest) (*core.AddNoteResponse, error) {
	return &core.AddNoteResponse{Status: "created"}, nil
}

func (s *fakeCLIService) CreatePlan(context.Context, *core.CreatePlanRequest) (*core.CreatePlanResponse, error) {
	return &core.CreatePlanResponse{Status: "created"}, nil
}

func (s *fakeCLIService) UpdatePlan(context.Context, *core.UpdatePlanRequest) (*core.UpdatePlanResponse, error) {
	return &core.UpdatePlanResponse{Status: "updated"}, nil
}

func (s *fakeCLIService) CorrectMemory(context.Context, *core.CorrectMemoryRequest) (*core.CorrectMemoryResponse, error) {
	return &core.CorrectMemoryResponse{Status: "recorded"}, nil
}

func (s *fakeCLIService) GetTimeline(context.Context, *core.GetTimelineRequest) (*core.GetTimelineResponse, error) {
	return &core.GetTimelineResponse{}, nil
}

func (s *fakeCLIService) ExplainMemory(context.Context, *core.ExplainMemoryRequest) (*core.ExplainMemoryResponse, error) {
	return &core.ExplainMemoryResponse{}, nil
}

func fakeServiceFactory(service core.VibeGravityService) serviceFactory {
	return func(context.Context) (core.VibeGravityService, func(), error) {
		return service, func() {}, nil
	}
}
