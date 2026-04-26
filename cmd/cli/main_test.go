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
	"errors"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/config"
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

	code := runCLI(context.Background(), []string{"jobs", "requeue-blocked", "job_blocked_1", "--reason", "apply support landed", "--yes"}, nil, &out, fakeStoreFactory(store), fakeServiceFactory(&fakeCLIService{}))

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; output: %s", code, out.String())
	}
	if store.requeuedJobID != "job_blocked_1" {
		t.Fatalf("expected job_blocked_1 to be requeued, got %q", store.requeuedJobID)
	}
	if store.requeueReason != "apply support landed" {
		t.Fatalf("expected requeue reason to be recorded, got %q", store.requeueReason)
	}
	if !strings.Contains(out.String(), "requeued blocked job job_blocked_1") {
		t.Fatalf("unexpected output: %s", out.String())
	}
}

func TestRunCLIRequeueBlockedDryRunDoesNotOpenStore(t *testing.T) {
	t.Parallel()

	store := &fakeBlockedJobStore{}
	var out bytes.Buffer

	code := runCLI(context.Background(), []string{"jobs", "requeue-blocked", "job_blocked_1", "--reason", "checking recovery", "--dry-run"}, nil, &out, fakeStoreFactory(store), fakeServiceFactory(&fakeCLIService{}))

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; output: %s", code, out.String())
	}
	if store.requeuedJobID != "" {
		t.Fatalf("dry-run must not requeue, got %q", store.requeuedJobID)
	}
	output := out.String()
	if !strings.Contains(output, "dry run: would requeue blocked job job_blocked_1") || !strings.Contains(output, "reason: checking recovery") {
		t.Fatalf("unexpected dry-run output: %s", output)
	}
}

func TestRunCLIRequeueBlockedAcceptsInteractiveConfirmation(t *testing.T) {
	t.Parallel()

	store := &fakeBlockedJobStore{}
	var out bytes.Buffer

	code := runCLI(context.Background(), []string{"jobs", "requeue-blocked", "job_blocked_1"}, strings.NewReader("job_blocked_1\n"), &out, fakeStoreFactory(store), fakeServiceFactory(&fakeCLIService{}))

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; output: %s", code, out.String())
	}
	if store.requeuedJobID != "job_blocked_1" {
		t.Fatalf("expected confirmed job to be requeued, got %q", store.requeuedJobID)
	}
}

func TestRunCLIRequeueBlockedRejectsMissingConfirmation(t *testing.T) {
	t.Parallel()

	store := &fakeBlockedJobStore{}
	var out bytes.Buffer

	code := runCLI(context.Background(), []string{"jobs", "requeue-blocked", "job_blocked_1"}, strings.NewReader("nope\n"), &out, fakeStoreFactory(store), fakeServiceFactory(&fakeCLIService{}))

	if code == 0 {
		t.Fatalf("expected non-zero exit without confirmation")
	}
	if store.requeuedJobID != "" {
		t.Fatalf("unconfirmed requeue must not mutate, got %q", store.requeuedJobID)
	}
	if !strings.Contains(out.String(), "requeue canceled") {
		t.Fatalf("expected cancellation message, got: %s", out.String())
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

	code := runCLI(context.Background(), []string{"jobs", "requeue-blocked", "missing_job", "--yes"}, nil, &out, fakeStoreFactory(store), fakeServiceFactory(&fakeCLIService{}))

	if code == 0 {
		t.Fatalf("expected non-zero exit for missing blocked job")
	}
	if !strings.Contains(out.String(), "blocked job not found") {
		t.Fatalf("expected not found message, got: %s", out.String())
	}
}

func TestRunCLIDoctorStrictJSONSeparatesRequiredAndEmbeddingFailures(t *testing.T) {
	t.Setenv("VIBEGRAVITY_DB_URL", "postgres://%")
	t.Setenv("VIBEGRAVITY_MIGRATION_PATH", t.TempDir())
	t.Setenv("VIBEGRAVITY_EMBEDDING_ENDPOINT", "http://127.0.0.1:1")
	t.Setenv("VIBEGRAVITY_EMBEDDING_MODEL", "pending")
	t.Setenv("VIBEGRAVITY_EMBEDDING_DIMS", "0")
	var out bytes.Buffer

	code := runCLI(context.Background(), []string{"doctor", "--strict", "--json"}, strings.NewReader(""), &out, fakeStoreFactory(&fakeBlockedJobStore{}), fakeServiceFactory(&fakeCLIService{}))

	if code != 1 {
		t.Fatalf("expected strict doctor to fail, got %d; output: %s", code, out.String())
	}
	output := out.String()
	for _, want := range []string{
		`"strict": true`,
		`"name": "database"`,
		`"name": "embedding_endpoint"`,
		`"name": "embedding_model"`,
		`"name": "embedding_dims"`,
		`"status": "fail"`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected doctor JSON to contain %q, got: %s", want, output)
		}
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

func TestRunCLIUsageStartsWithQuickstart(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	code := runCLI(context.Background(), nil, strings.NewReader(""), &out, fakeStoreFactory(&fakeBlockedJobStore{}), fakeServiceFactory(&fakeCLIService{}))

	if code == 0 {
		t.Fatalf("expected usage exit code")
	}
	if !strings.HasPrefix(out.String(), "처음이면 quickstart 실행: vibegravity quickstart\n") {
		t.Fatalf("usage should start with quickstart guidance, got: %s", out.String())
	}
}

func TestRunCLIRejectsUnknownCommand(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	code := runCLI(context.Background(), []string{"unknown-command"}, strings.NewReader(""), &out, fakeStoreFactory(&fakeBlockedJobStore{}), fakeServiceFactory(&fakeCLIService{}))

	if code == 0 {
		t.Fatalf("expected non-zero exit code for unknown command")
	}
	if !strings.Contains(out.String(), "Unknown command: unknown-command") {
		t.Fatalf("expected unknown command output, got: %s", out.String())
	}
}

func TestRunDoctorReturnsFailureWhenDatabaseCheckFails(t *testing.T) {
	restore := overrideDoctorChecks(t, errors.New("database refused connection"), nil)
	defer restore()
	setDoctorEnv(t, t.TempDir(), "http://embedding.test/health", "local-embedding", "384")

	var out bytes.Buffer
	code := runDoctor(context.Background(), nil, &out)

	if code != 1 {
		t.Fatalf("expected doctor exit code 1 for required database failure, got %d; output: %s", code, out.String())
	}
	if !strings.Contains(out.String(), "[FAIL] database (required)") || !strings.Contains(out.String(), "database refused connection") {
		t.Fatalf("expected database failure details, got: %s", out.String())
	}
}

func TestRunDoctorStrictReturnsFailureWhenEmbeddingCheckFails(t *testing.T) {
	restore := overrideDoctorChecks(t, nil, errors.New("embedding endpoint unavailable"))
	defer restore()
	setDoctorEnv(t, t.TempDir(), "http://embedding.test/health", "local-embedding", "384")

	var out bytes.Buffer
	code := runDoctor(context.Background(), []string{"--strict"}, &out)

	if code != 1 {
		t.Fatalf("expected doctor strict exit code 1 for embedding failure, got %d; output: %s", code, out.String())
	}
	output := out.String()
	if !strings.Contains(output, "[FAIL] embedding_endpoint (required)") || !strings.Contains(output, "embedding endpoint unavailable") {
		t.Fatalf("expected strict embedding failure details, got: %s", output)
	}
}

func TestRunCLIGoldenEvalFailureReturnsNonZero(t *testing.T) {
	t.Parallel()

	fixture := filepath.Join(t.TempDir(), "failing_golden.json")
	if err := os.WriteFile(fixture, []byte(`{
  "scenarios": [
    {
      "name": "intentional golden failure",
      "prefetch": {
        "tenant_id": "tenant_1",
        "workspace_id": "workspace_1",
        "session_id": "session_1",
        "actor_id": "agent:hermes-main",
        "query": "missing",
        "budget_tokens": 32
      },
      "expect": {"contains": ["this text will not be recalled"]}
    }
  ]
}`), 0o600); err != nil {
		t.Fatalf("write failing fixture: %v", err)
	}
	var out bytes.Buffer

	code := runCLI(context.Background(), []string{"eval", "golden", "--path", fixture}, strings.NewReader(""), &out, fakeStoreFactory(&fakeBlockedJobStore{}), fakeServiceFactory(&fakeCLIService{}))

	if code != 1 {
		t.Fatalf("expected failing golden eval exit code 1, got %d; output: %s", code, out.String())
	}
	output := out.String()
	if !strings.Contains(output, "FAIL\tintentional golden failure") || !strings.Contains(output, "Golden eval failed.") {
		t.Fatalf("expected golden failure output, got: %s", output)
	}
}

func TestBuiltCLIBinarySmoke(t *testing.T) {
	t.Parallel()

	root := filepath.Clean(filepath.Join("..", ".."))
	bin := filepath.Join(t.TempDir(), "cli")
	build := exec.Command("go", "build", "-ldflags", "-X main.version=test-smoke", "-o", bin, "./cmd/cli")
	build.Dir = root
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build cli binary: %v\n%s", err, output)
	}

	run := exec.Command(bin, "version")
	output, err := run.CombinedOutput()
	if err != nil {
		t.Fatalf("run built binary smoke: %v\n%s", err, output)
	}
	if !strings.Contains(string(output), "vibegravity test-smoke") {
		t.Fatalf("expected built binary version output, got: %s", output)
	}
}

type fakeDoctorPool struct{}

func (fakeDoctorPool) Close() {}

type fakeDoctorHTTPClient struct {
	err error
}

func (c fakeDoctorHTTPClient) Get(string) (*http.Response, error) {
	if c.err != nil {
		return nil, c.err
	}
	return &http.Response{Status: "200 OK", Body: http.NoBody}, nil
}

func overrideDoctorChecks(t *testing.T, dbErr error, embeddingErr error) func() {
	t.Helper()
	oldPool := newDoctorPool
	oldHTTPClient := newDoctorHTTPClient
	newDoctorPool = func(context.Context, config.Config) (doctorPool, error) {
		if dbErr != nil {
			return nil, dbErr
		}
		return fakeDoctorPool{}, nil
	}
	newDoctorHTTPClient = func() doctorHTTPGetter {
		return fakeDoctorHTTPClient{err: embeddingErr}
	}
	return func() {
		newDoctorPool = oldPool
		newDoctorHTTPClient = oldHTTPClient
	}
}

func setDoctorEnv(t *testing.T, migrationPath string, embeddingEndpoint string, embeddingModel string, embeddingDims string) {
	t.Helper()
	t.Setenv("VIBEGRAVITY_DB_URL", "postgres://doctor:secret@localhost:5432/vibegravity?sslmode=disable")
	t.Setenv("VIBEGRAVITY_MIGRATION_PATH", migrationPath)
	t.Setenv("VIBEGRAVITY_EMBEDDING_ENDPOINT", embeddingEndpoint)
	t.Setenv("VIBEGRAVITY_EMBEDDING_MODEL", embeddingModel)
	t.Setenv("VIBEGRAVITY_EMBEDDING_DIMS", embeddingDims)
}

type fakeBlockedJobStore struct {
	jobs          []*core.IngestJob
	listLimit     int
	listErr       error
	requeuedJobID string
	requeueReason string
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

func (s *fakeBlockedJobStore) RequeueBlockedJob(_ context.Context, jobID string, reason string) error {
	s.requeuedJobID = jobID
	s.requeueReason = reason
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
