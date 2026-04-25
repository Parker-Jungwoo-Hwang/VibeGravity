# 07 Gpt Pro Tests And Golden Scenarios

Generated: 2026-04-25

This file is part of the GPT-Pro review material bundle for VibeGravity.

## Included Sources

- `cmd/cli/main_test.go`
- `internal/core/service_test.go`
- `internal/eval/demo_test.go`
- `internal/eval/golden_test.go`
- `internal/graph/apply_test.go`
- `internal/graph/dreaming_test.go`
- `internal/graph/store_apply_test.go`
- `internal/hermes/provider_test.go`
- `internal/httpapi/router_test.go`
- `internal/ingest/service_test.go`
- `internal/kernel/service_test.go`
- `internal/mcp/protocol_test.go`
- `internal/mcp/surface_test.go`
- `internal/reasoning/codex_bridge_test.go`
- `internal/reasoning/orchestrator_test.go`
- `internal/reasoning/stage2_input_preparer_test.go`
- `internal/recall/assembler_test.go`
- `internal/store/postgres/concurrency_integration_test.go`
- `internal/store/postgres/corrections_test.go`
- `internal/store/postgres/documents_test.go`
- `internal/store/postgres/dreaming_test.go`
- `internal/store/postgres/jobs_test.go`
- `internal/store/postgres/memories_test.go`
- `internal/store/postgres/notes_plans_test.go`
- `internal/store/postgres/search_test.go`
- `internal/store/postgres/timeline_test.go`
- `internal/worker/processor_test.go`
- `internal/worker/stage2_sources_test.go`
- `tests/baseline_test.go`
- `tests/golden/replay_eval.json`
- `tests/migration_contract_test.go`
- `tools/headercheck/main.go`

## Source Contents


<!-- Source: cmd/cli/main_test.go | bytes=12120 | lines=350 | sha16=3f2f3715ff40be13 -->

```go
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

```



<!-- Source: internal/core/service_test.go | bytes=3659 | lines=120 | sha16=acc8d58cea98b1e5 -->

```go
// ============================================================
// FILE     : internal/core/service_test.go
// PURPOSE  : Verifies the core service interface and domain records compile together.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : TestVibeGravityService_Baseline, TestDomainTypes_Compile
// DEPENDS  : internal/core
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: Keep this as a fast contract smoke test for domain changes.
// ============================================================

package core

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestVibeGravityService_Baseline(t *testing.T) {
	t.Helper()
	var _ VibeGravityService = (*contractService)(nil)
}

func TestDomainTypes_Compile(t *testing.T) {
	now := time.Date(2026, time.April, 24, 0, 0, 0, 0, time.UTC)
	payload := json.RawMessage(`{"text":"hello"}`)

	memory := Memory{
		ID:             "mem_1",
		TenantID:       "tenant_1",
		WorkspaceID:    "workspace_1",
		Scope:          MemoryScopeWorkspaceShared,
		OwnerEntityID:  "agent:hermes-main",
		Kind:           MemoryKindDecision,
		ArtifactClass:  ArtifactClassKnowledge,
		Text:           "VibeGravity is Hermes-first.",
		Fingerprint:    "fp_1",
		Confidence:     0.99,
		Status:         MemoryStatusActive,
		ValidFrom:      now,
		LatestFlag:     true,
		MetadataJSON:   payload,
		EmbeddingModel: "pending",
		EmbeddingDims:  0,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if memory.Scope != MemoryScopeWorkspaceShared {
		t.Fatalf("unexpected scope: %s", memory.Scope)
	}
	if memory.ArtifactClass != ArtifactClassKnowledge {
		t.Fatalf("unexpected artifact class: %s", memory.ArtifactClass)
	}

	trace := MemoryTrace{
		MemoryID:               memory.ID,
		RawEventIDs:            []string{"evt_1"},
		ReasoningJobID:         "job_1",
		ReasoningStage:         "resolve",
		CandidateSnapshotJSON:  payload,
		AppliedOperationsJSON:  payload,
		OperatorCorrectionFlag: false,
		RelatedDocumentIDs:     []string{"doc_1"},
		CreatedAt:              now,
	}
	if trace.RawEventIDs[0] != "evt_1" {
		t.Fatalf("unexpected trace event id: %s", trace.RawEventIDs[0])
	}
}

type contractService struct{}

func (contractService) Prefetch(context.Context, *PrefetchRequest) (*PrefetchResponse, error) {
	return nil, nil
}

func (contractService) SyncTurn(context.Context, *SyncTurnRequest) (*SyncTurnResponse, error) {
	return nil, nil
}

func (contractService) AddDocument(context.Context, *AddDocumentRequest) (*AddDocumentResponse, error) {
	return nil, nil
}

func (contractService) SearchMemories(context.Context, *SearchMemoriesRequest) (*SearchMemoriesResponse, error) {
	return nil, nil
}

func (contractService) SearchDocuments(context.Context, *SearchDocumentsRequest) (*SearchDocumentsResponse, error) {
	return nil, nil
}

func (contractService) AddNote(context.Context, *AddNoteRequest) (*AddNoteResponse, error) {
	return nil, nil
}

func (contractService) CreatePlan(context.Context, *CreatePlanRequest) (*CreatePlanResponse, error) {
	return nil, nil
}

func (contractService) UpdatePlan(context.Context, *UpdatePlanRequest) (*UpdatePlanResponse, error) {
	return nil, nil
}

func (contractService) CorrectMemory(context.Context, *CorrectMemoryRequest) (*CorrectMemoryResponse, error) {
	return nil, nil
}

func (contractService) GetTimeline(context.Context, *GetTimelineRequest) (*GetTimelineResponse, error) {
	return nil, nil
}

func (contractService) ExplainMemory(context.Context, *ExplainMemoryRequest) (*ExplainMemoryResponse, error) {
	return nil, nil
}

```



<!-- Source: internal/eval/demo_test.go | bytes=1350 | lines=45 | sha16=f4c09398d1a1731b -->

```go
// ============================================================
// FILE     : internal/eval/demo_test.go
// PURPOSE  : Verifies the local Hermes Memory trust-loop demo eval.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : Hermes Memory demo eval tests
// DEPENDS  : context, testing
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: Demo eval tests must stay local-only and deterministic.
// ============================================================

package eval

import (
	"context"
	"testing"
)

func TestRunHermesMemoryDemoPassesTrustLoop(t *testing.T) {
	t.Parallel()

	summary := RunHermesMemoryDemo(context.Background())
	if summary == nil || !summary.Passed {
		t.Fatalf("expected demo eval to pass, got %#v", summary)
	}

	names := make(map[string]bool, len(summary.Results))
	for _, result := range summary.Results {
		names[result.Scenario] = true
	}
	for _, want := range []string{
		"demo initial recall shows rule plan and trust metadata",
		"demo explain shows recalled memory provenance",
		"demo correction writes supersession",
		"demo next recall uses correction",
		"demo private scope separation",
	} {
		if !names[want] {
			t.Fatalf("expected demo step %q in %#v", want, summary.Results)
		}
	}
}

```



<!-- Source: internal/eval/golden_test.go | bytes=4463 | lines=151 | sha16=edd99763b39b0c64 -->

```go
// ============================================================
// FILE     : internal/eval/golden_test.go
// PURPOSE  : Verifies the deterministic golden scenario runner.
// LAYER    : test
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : TestRunFilePassesGoldenFixture, TestRunScenariosReportsRegression
// DEPENDS  : context, path/filepath, testing, internal/eval
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: Eval tests should fail loudly when fixture expectations drift.
// ============================================================

package eval

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

func TestRunFilePassesGoldenFixture(t *testing.T) {
	t.Parallel()

	summary, err := RunFile(context.Background(), filepath.Join("..", "..", "tests", "golden", "replay_eval.json"))
	if err != nil {
		t.Fatalf("RunFile returned error: %v", err)
	}
	if !summary.Passed {
		t.Fatalf("expected golden fixture to pass, got %#v", summary.Results)
	}
	if len(summary.Results) < 4 {
		t.Fatalf("expected multiple golden scenarios, got %d", len(summary.Results))
	}
}

func TestRunFileCoversGraphReplayGates(t *testing.T) {
	t.Parallel()

	summary, err := RunFile(context.Background(), filepath.Join("..", "..", "tests", "golden", "replay_eval.json"))
	if err != nil {
		t.Fatalf("RunFile returned error: %v", err)
	}
	names := make(map[string]bool, len(summary.Results))
	for _, result := range summary.Results {
		names[result.Scenario] = true
	}
	for _, want := range []string{
		"update memory replay suppresses prior fact",
		"correction replay changes later recall",
		"group shared graph write remains rejected",
	} {
		if !names[want] {
			t.Fatalf("expected graph replay scenario %q", want)
		}
	}
}

func TestRunFileCoversWorkerBacklogGates(t *testing.T) {
	t.Parallel()

	summary, err := RunFile(context.Background(), filepath.Join("..", "..", "tests", "golden", "replay_eval.json"))
	if err != nil {
		t.Fatalf("RunFile returned error: %v", err)
	}
	names := make(map[string]bool, len(summary.Results))
	for _, result := range summary.Results {
		names[result.Scenario] = true
	}
	for _, want := range []string{
		"stage1 outage retries without graph writes",
		"stage2 outage recovery replay is idempotent",
		"unsupported apply work becomes blocked",
	} {
		if !names[want] {
			t.Fatalf("expected worker backlog scenario %q", want)
		}
	}
}

func TestRunScenariosReportsRegression(t *testing.T) {
	t.Parallel()

	summary := RunScenarios(context.Background(), []Scenario{{
		Name: "missing pinned note is visible",
		Prefetch: core.PrefetchRequest{
			TenantID:     "tenant_1",
			WorkspaceID:  "workspace_1",
			SessionID:    "session_1",
			ActorID:      "agent:hermes-main",
			Query:        "plan",
			BudgetTokens: 100,
		},
		Expect: Expectation{Contains: []string{"must appear"}},
	}})

	if summary.Passed {
		t.Fatalf("expected scenario to fail")
	}
	if len(summary.Results) != 1 || len(summary.Results[0].Errors) == 0 {
		t.Fatalf("expected failure details, got %#v", summary.Results)
	}
}

func TestRunScenariosReportsBlockMetadataRegression(t *testing.T) {
	t.Parallel()

	summary := RunScenarios(context.Background(), []Scenario{{
		Name: "trust metadata mismatch is visible",
		Prefetch: core.PrefetchRequest{
			TenantID:     "tenant_1",
			WorkspaceID:  "workspace_1",
			SessionID:    "session_1",
			ActorID:      "agent:hermes-main",
			Query:        "Hermes",
			BudgetTokens: 100,
		},
		Fixtures: Fixtures{
			Memories: []core.MemoryResult{{
				MemoryID:      "mem_1",
				Kind:          core.MemoryKindFact,
				ArtifactClass: core.ArtifactClassKnowledge,
				Text:          "Hermes remembers scoped project context.",
				Confidence:    0.9,
				Scope:         core.MemoryScopeWorkspaceShared,
				LatestFlag:    true,
			}},
		},
		Expect: Expectation{
			BlockMetadata: []BlockMetadataExpectation{{
				Kind:      "memory",
				Scope:     core.MemoryScopeAgentPrivate,
				Source:    "memories",
				SourceID:  "mem_1",
				Status:    "active",
				Freshness: "stored",
			}},
		},
	}})

	if summary.Passed {
		t.Fatalf("expected scenario to fail")
	}
	if got := summary.Results[0].Errors; len(got) == 0 || !strings.Contains(strings.Join(got, "\n"), "scope got") {
		t.Fatalf("expected metadata failure details, got %#v", summary.Results)
	}
}

```



<!-- Source: internal/graph/apply_test.go | bytes=7743 | lines=260 | sha16=dc45349205816ceb -->

```go
// ============================================================
// FILE     : internal/graph/apply_test.go
// PURPOSE  : Verifies no-op apply validation for structured Stage 2 graph operations.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : TestNoopApplyEngine_ValidatesStructuredOperations, TestNoopApplyEngine_RejectsInvalidOperations
// DEPENDS  : context, encoding/json, errors, strings, testing, internal/core, internal/reasoning
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: These tests lock the apply boundary only; they must not assert memory quality or extraction behavior.
// ============================================================

package graph

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/reasoning"
)

func TestNoopApplyEngine_ValidatesStructuredOperations(t *testing.T) {
	t.Parallel()

	engine := NewNoopApplyEngine()
	result, err := engine.Apply(context.Background(), validApplyRequest(validCreateOperation()))
	if err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}

	if result.AppliedOperationCount != 0 {
		t.Fatalf("expected no-op engine to commit no operations, got %d", result.AppliedOperationCount)
	}
	if result.TraceWritten {
		t.Fatalf("expected no-op engine to write no trace")
	}
	if len(result.MemoryIDs) != 0 {
		t.Fatalf("expected no memory IDs, got %v", result.MemoryIDs)
	}
}

func TestNoopApplyEngine_RejectsInvalidOperations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		operation reasoning.GraphOperation
		wantError string
	}{
		{
			name: "empty operation kind",
			operation: func() reasoning.GraphOperation {
				operation := validCreateOperation()
				operation.Kind = ""
				return operation
			}(),
			wantError: "kind is required",
		},
		{
			name: "unsupported operation kind",
			operation: func() reasoning.GraphOperation {
				operation := validCreateOperation()
				operation.Kind = reasoning.OperationKind("merge_memory")
				return operation
			}(),
			wantError: "kind is unsupported",
		},
		{
			name: "missing scope",
			operation: func() reasoning.GraphOperation {
				operation := validCreateOperation()
				operation.Memory.Scope = ""
				return operation
			}(),
			wantError: "memory.scope is required",
		},
		{
			name: "invalid confidence",
			operation: func() reasoning.GraphOperation {
				operation := validCreateOperation()
				operation.Memory.Confidence = 1.01
				return operation
			}(),
			wantError: "memory.confidence",
		},
		{
			name: "update without target",
			operation: func() reasoning.GraphOperation {
				operation := validUpdateOperation()
				operation.Memory.TargetID = ""
				operation.Edge.ToMemoryID = ""
				return operation
			}(),
			wantError: "memory.target_id is required",
		},
		{
			name: "extend without target",
			operation: func() reasoning.GraphOperation {
				operation := validExtendOperation()
				operation.Memory.TargetID = ""
				operation.Edge.ToMemoryID = ""
				return operation
			}(),
			wantError: "memory.target_id is required",
		},
		{
			name: "unstructured operation metadata",
			operation: func() reasoning.GraphOperation {
				operation := validCreateOperation()
				operation.Metadata = json.RawMessage(`"free-form explanation"`)
				return operation
			}(),
			wantError: "metadata must be a JSON object",
		},
		{
			name: "update edge must use updates",
			operation: func() reasoning.GraphOperation {
				operation := validUpdateOperation()
				operation.Edge.EdgeKind = core.EdgeKindExtends
				return operation
			}(),
			wantError: `edge.edge_kind must be "updates"`,
		},
		{
			name: "extend edge must use extends",
			operation: func() reasoning.GraphOperation {
				operation := validExtendOperation()
				operation.Edge.EdgeKind = core.EdgeKindUpdates
				return operation
			}(),
			wantError: `edge.edge_kind must be "extends"`,
		},
		{
			name: "archive without target",
			operation: func() reasoning.GraphOperation {
				operation := validArchiveOperation()
				operation.Memory.TargetID = ""
				operation.Memory.MemoryID = ""
				return operation
			}(),
			wantError: "memory target is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewNoopApplyEngine().Apply(context.Background(), validApplyRequest(tt.operation))
			if err == nil {
				t.Fatalf("expected Apply to reject operation")
			}
			if !errors.Is(err, core.ErrInvalidArgument) {
				t.Fatalf("expected ErrInvalidArgument, got %v", err)
			}
			if !strings.Contains(err.Error(), tt.wantError) {
				t.Fatalf("expected error to contain %q, got %v", tt.wantError, err)
			}
		})
	}
}

func validApplyRequest(operation reasoning.GraphOperation) *ApplyRequest {
	return &ApplyRequest{
		JobID:       "job_1",
		TenantID:    "tenant_1",
		WorkspaceID: "workspace_1",
		RawEventIDs: []string{"evt_1"},
		Reasoning: &reasoning.ProcessTurnResult{
			Stage1: reasoning.Stage1Output{
				CandidateEntities: []reasoning.CandidateEntity{},
				CandidateMemories: []reasoning.CandidateMemory{},
			},
			Stage2: reasoning.Stage2Output{
				Operations:     []reasoning.GraphOperation{operation},
				ProfileDelta:   json.RawMessage(`{}`),
				SessionSummary: "",
				PlanDelta:      json.RawMessage(`{}`),
				Trace: reasoning.Trace{
					SchemaVersion: "v0",
					Stage:         reasoning.StageNameResolve,
					Codes:         []string{"test"},
					MetadataJSON:  json.RawMessage(`{}`),
				},
			},
		},
	}
}

func validCreateOperation() reasoning.GraphOperation {
	return reasoning.GraphOperation{
		OperationID: "op_1",
		Kind:        reasoning.OperationKindCreateMemory,
		Memory:      validMemoryMutation(),
		RawEventIDs: []string{"evt_1"},
		Metadata:    json.RawMessage(`{"source":"test"}`),
	}
}

func validUpdateOperation() reasoning.GraphOperation {
	operation := validCreateOperation()
	operation.OperationID = "op_update"
	operation.Kind = reasoning.OperationKindUpdateMemory
	operation.Memory.MemoryID = "mem_new"
	operation.Memory.TargetID = "mem_old"
	operation.Edge = &reasoning.EdgeMutation{
		FromMemoryID: "mem_new",
		ToMemoryID:   "mem_old",
		EdgeKind:     core.EdgeKindUpdates,
		Confidence:   0.8,
	}
	return operation
}

func validExtendOperation() reasoning.GraphOperation {
	operation := validCreateOperation()
	operation.OperationID = "op_extend"
	operation.Kind = reasoning.OperationKindExtendMemory
	operation.Memory.MemoryID = "mem_detail"
	operation.Memory.TargetID = "mem_base"
	operation.Edge = &reasoning.EdgeMutation{
		FromMemoryID: "mem_detail",
		ToMemoryID:   "mem_base",
		EdgeKind:     core.EdgeKindExtends,
		Confidence:   0.8,
	}
	return operation
}

func validArchiveOperation() reasoning.GraphOperation {
	operation := validCreateOperation()
	operation.OperationID = "op_archive"
	operation.Kind = reasoning.OperationKindArchiveMemory
	operation.Memory = &reasoning.MemoryMutation{
		TargetID:      "mem_old",
		Scope:         core.MemoryScopeWorkspaceShared,
		OwnerEntityID: "agent:hermes-main",
		Confidence:    0.8,
		MetadataJSON:  json.RawMessage(`{"reason":"stale"}`),
	}
	return operation
}

func validMemoryMutation() *reasoning.MemoryMutation {
	return &reasoning.MemoryMutation{
		Kind:          core.MemoryKindDecision,
		ArtifactClass: core.ArtifactClassKnowledge,
		Scope:         core.MemoryScopeWorkspaceShared,
		OwnerEntityID: "agent:hermes-main",
		Text:          "VibeGravity keeps Stage 2 operations structured.",
		Confidence:    0.8,
		MetadataJSON:  json.RawMessage(`{"source":"test"}`),
	}
}

```



<!-- Source: internal/graph/dreaming_test.go | bytes=4929 | lines=150 | sha16=561304430d3453fd -->

```go
// ============================================================
// FILE     : internal/graph/dreaming_test.go
// PURPOSE  : Verifies dreaming session summaries and tier-promotion orchestration.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : graph dreaming tests
// DEPENDS  : context, testing, time, internal/core
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: Dreaming tests should prove no duplicate memory creation is required.
// ============================================================

package graph

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

func TestDreamSessionWritesSummaryAndPromotesExistingMemories(t *testing.T) {
	t.Parallel()

	store := &fakeGraphDreamingStore{
		input: &core.DreamingSessionInput{
			RawEventIDs: []string{"evt_1", "evt_2"},
			Memories: []*core.Memory{
				{ID: "mem_1", Text: "User prefers narrow work-pack slices."},
				{ID: "mem_2", Text: "Workspace decisions are kept file-backed."},
			},
		},
		promotions: map[core.DreamingTier]*core.DreamingPromotionResult{
			core.DreamingTierMidTerm:       {PromotedCount: 2, MemoryIDs: []string{"mem_1", "mem_2"}},
			core.DreamingTierLongTerm:      {PromotedCount: 1, MemoryIDs: []string{"mem_2"}},
			core.DreamingTierUltraLongTerm: {PromotedCount: 0},
		},
	}
	service := newTestGraphDreamingService(t, store)

	result, err := service.DreamSession(context.Background(), &core.DreamSessionRequest{
		JobID:       "job_dream_1",
		TenantID:    "tenant_1",
		WorkspaceID: "workspace_1",
		SessionID:   "session_1",
	})
	if err != nil {
		t.Fatalf("DreamSession returned error: %v", err)
	}

	if !result.SessionSummaryWritten || result.MidTermPromoted != 2 || result.LongTermPromoted != 1 {
		t.Fatalf("unexpected dreaming result: %#v", result)
	}
	if store.summary == nil {
		t.Fatalf("expected session summary to be written")
	}
	if !strings.Contains(store.summary.SummaryText, "2 raw events and 2 derived memories") {
		t.Fatalf("unexpected summary text: %s", store.summary.SummaryText)
	}
	if len(store.requests) != 3 {
		t.Fatalf("expected mid, long, and ultra promotion requests, got %#v", store.requests)
	}
	if store.requests[0].Tier != core.DreamingTierMidTerm || len(store.requests[0].MemoryIDs) != 2 {
		t.Fatalf("unexpected mid-term request: %#v", store.requests[0])
	}
}

func TestDreamWorkspacePromotesStableTiersOnly(t *testing.T) {
	t.Parallel()

	store := &fakeGraphDreamingStore{
		promotions: map[core.DreamingTier]*core.DreamingPromotionResult{
			core.DreamingTierLongTerm:      {PromotedCount: 3},
			core.DreamingTierUltraLongTerm: {PromotedCount: 1},
		},
	}
	service := newTestGraphDreamingService(t, store)

	result, err := service.DreamWorkspace(context.Background(), &core.DreamWorkspaceRequest{
		JobID:       "job_dream_workspace",
		TenantID:    "tenant_1",
		WorkspaceID: "workspace_1",
	})
	if err != nil {
		t.Fatalf("DreamWorkspace returned error: %v", err)
	}

	if result.LongTermPromoted != 3 || result.UltraLongTermPromoted != 1 {
		t.Fatalf("unexpected workspace dreaming result: %#v", result)
	}
	if len(store.requests) != 2 {
		t.Fatalf("expected two promotion requests, got %#v", store.requests)
	}
	for _, req := range store.requests {
		if !req.RequireStableKind {
			t.Fatalf("workspace promotion must require stable kind: %#v", req)
		}
	}
}

func newTestGraphDreamingService(t *testing.T, store *fakeGraphDreamingStore) *DreamingService {
	t.Helper()
	service, err := NewDreamingService(DreamingDependencies{
		Store: store,
		Clock: func() time.Time {
			return time.Date(2026, time.April, 24, 2, 0, 0, 0, time.UTC)
		},
	})
	if err != nil {
		t.Fatalf("NewDreamingService returned error: %v", err)
	}
	return service
}

type fakeGraphDreamingStore struct {
	input      *core.DreamingSessionInput
	summary    *core.SessionSummary
	promotions map[core.DreamingTier]*core.DreamingPromotionResult
	requests   []*core.DreamingPromotionRequest
}

func (s *fakeGraphDreamingStore) LoadDreamingSessionInput(context.Context, *core.DreamSessionRequest) (*core.DreamingSessionInput, error) {
	if s.input != nil {
		return s.input, nil
	}
	return &core.DreamingSessionInput{}, nil
}

func (s *fakeGraphDreamingStore) PromoteMemories(_ context.Context, req *core.DreamingPromotionRequest) (*core.DreamingPromotionResult, error) {
	s.requests = append(s.requests, req)
	if s.promotions != nil {
		if result, ok := s.promotions[req.Tier]; ok {
			return result, nil
		}
	}
	return &core.DreamingPromotionResult{}, nil
}

func (s *fakeGraphDreamingStore) UpsertSessionSummary(_ context.Context, summary *core.SessionSummary) error {
	s.summary = summary
	return nil
}

func (s *fakeGraphDreamingStore) GetSessionSummary(context.Context, string) (*core.SessionSummary, error) {
	return nil, core.ErrNotFound
}

```



<!-- Source: internal/graph/store_apply_test.go | bytes=14801 | lines=382 | sha16=55a838bbb63fe671 -->

```go
// ============================================================
// FILE     : internal/graph/store_apply_test.go
// PURPOSE  : Verifies store-backed apply writes safe memory graph rows with mandatory trace provenance.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : TestStoreBackedApplyEngine_WritesCreateMemoryWithTrace, TestStoreBackedApplyEngine_WritesExtendMemoryWithTraceAndEdge, TestStoreBackedApplyEngine_WritesUpdateMemoryWithTraceAndSupersessionEdge, TestStoreBackedApplyEngine_RejectsUnsupportedWrites
// DEPENDS  : context, encoding/json, errors, strings, testing, time, internal/core, internal/reasoning
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: These tests prove provenance is part of the success boundary for derived memory writes.
// ============================================================

package graph

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/reasoning"
)

func TestStoreBackedApplyEngine_WritesCreateMemoryWithTrace(t *testing.T) {
	t.Parallel()

	memories := &fakeMemoryTraceCreator{}
	engine := newTestStoreBackedApplyEngine(t, memories)

	result, err := engine.Apply(context.Background(), validApplyRequest(validCreateOperation()))
	if err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}

	if result.AppliedOperationCount != 1 {
		t.Fatalf("unexpected applied count: got %d want 1", result.AppliedOperationCount)
	}
	if !result.TraceWritten {
		t.Fatalf("expected trace to be written")
	}
	if len(result.MemoryIDs) != 1 || !strings.HasPrefix(result.MemoryIDs[0], "mem_") {
		t.Fatalf("unexpected memory IDs: %v", result.MemoryIDs)
	}
	if len(memories.memories) != 1 || len(memories.traces) != 1 {
		t.Fatalf("expected one memory and one trace, got memories=%d traces=%d", len(memories.memories), len(memories.traces))
	}

	memory := memories.memories[0]
	if memory.ID != result.MemoryIDs[0] {
		t.Fatalf("result memory ID did not match written memory: got %q want %q", result.MemoryIDs[0], memory.ID)
	}
	if memory.TenantID != "tenant_1" || memory.WorkspaceID != "workspace_1" {
		t.Fatalf("memory tenant/workspace not copied from apply request: %#v", memory)
	}
	if memory.Scope != core.MemoryScopeWorkspaceShared {
		t.Fatalf("memory scope not preserved: %q", memory.Scope)
	}
	if memory.Status != core.MemoryStatusActive || !memory.LatestFlag {
		t.Fatalf("memory should start active latest, got status=%q latest=%v", memory.Status, memory.LatestFlag)
	}
	if memory.Fingerprint == "" || !strings.HasPrefix(memory.Fingerprint, "fp_") {
		t.Fatalf("expected stable fingerprint, got %q", memory.Fingerprint)
	}
	if !memory.CreatedAt.Equal(fixedApplyTime) || !memory.UpdatedAt.Equal(fixedApplyTime) || !memory.ValidFrom.Equal(fixedApplyTime) {
		t.Fatalf("memory timestamps should use apply clock: %#v", memory)
	}

	trace := memories.traces[0]
	if trace.MemoryID != memory.ID {
		t.Fatalf("trace memory ID mismatch: got %q want %q", trace.MemoryID, memory.ID)
	}
	if trace.ReasoningJobID != "job_1" || trace.ReasoningStage != string(reasoning.StageNameResolve) {
		t.Fatalf("trace reasoning provenance mismatch: %#v", trace)
	}
	if len(trace.RawEventIDs) != 1 || trace.RawEventIDs[0] != "evt_1" {
		t.Fatalf("trace raw event provenance mismatch: %#v", trace.RawEventIDs)
	}
	assertJSONContains(t, trace.CandidateSnapshotJSON, "candidate_memories")
	assertJSONContains(t, trace.AppliedOperationsJSON, "op_1")
}

func TestStoreBackedApplyEngine_WritesExtendMemoryWithTraceAndEdge(t *testing.T) {
	t.Parallel()

	memories := &fakeMemoryTraceCreator{}
	engine := newTestStoreBackedApplyEngine(t, memories)
	operation := validExtendOperation()

	result, err := engine.Apply(context.Background(), validApplyRequest(operation))
	if err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}

	if result.AppliedOperationCount != 1 {
		t.Fatalf("unexpected applied count: got %d want 1", result.AppliedOperationCount)
	}
	if !result.TraceWritten {
		t.Fatalf("expected trace to be written")
	}
	if len(memories.memories) != 1 || len(memories.traces) != 1 || len(memories.edges) != 1 {
		t.Fatalf("expected one memory, trace, and edge; got memories=%d traces=%d edges=%d", len(memories.memories), len(memories.traces), len(memories.edges))
	}

	memory := memories.memories[0]
	if memory.ID == "" || !strings.HasPrefix(memory.ID, "mem_") {
		t.Fatalf("expected deterministic memory ID, got %q", memory.ID)
	}
	if memory.Status != core.MemoryStatusActive || !memory.LatestFlag {
		t.Fatalf("extension memory should start active latest without changing the target, got status=%q latest=%v", memory.Status, memory.LatestFlag)
	}
	if memory.Text != operation.Memory.Text || memory.Scope != operation.Memory.Scope || memory.OwnerEntityID != operation.Memory.OwnerEntityID {
		t.Fatalf("extension memory did not preserve mutation fields: %#v", memory)
	}

	trace := memories.traces[0]
	if trace.MemoryID != memory.ID {
		t.Fatalf("trace memory ID mismatch: got %q want %q", trace.MemoryID, memory.ID)
	}
	if len(trace.RawEventIDs) != 1 || trace.RawEventIDs[0] != "evt_1" {
		t.Fatalf("trace raw event provenance mismatch: %#v", trace.RawEventIDs)
	}
	assertJSONContains(t, trace.AppliedOperationsJSON, "op_extend")

	edge := memories.edges[0]
	if edge.FromMemoryID != memory.ID {
		t.Fatalf("extension edge must originate from the written memory: got %q want %q", edge.FromMemoryID, memory.ID)
	}
	if edge.ToMemoryID != operation.Memory.TargetID {
		t.Fatalf("extension edge target mismatch: got %q want %q", edge.ToMemoryID, operation.Memory.TargetID)
	}
	if edge.EdgeKind != core.EdgeKindExtends {
		t.Fatalf("extension edge kind mismatch: got %q", edge.EdgeKind)
	}
	if edge.Confidence != operation.Edge.Confidence || edge.CreatedByJobID != "job_1" || !edge.CreatedAt.Equal(fixedApplyTime) {
		t.Fatalf("extension edge metadata mismatch: %#v", edge)
	}
}

func TestStoreBackedApplyEngine_WritesUpdateMemoryWithTraceAndSupersessionEdge(t *testing.T) {
	t.Parallel()

	memories := &fakeMemoryTraceCreator{}
	engine := newTestStoreBackedApplyEngine(t, memories)
	operation := validUpdateOperation()

	result, err := engine.Apply(context.Background(), validApplyRequest(operation))
	if err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}

	if result.AppliedOperationCount != 1 {
		t.Fatalf("unexpected applied count: got %d want 1", result.AppliedOperationCount)
	}
	if !result.TraceWritten {
		t.Fatalf("expected trace to be written")
	}
	if len(memories.updateMemories) != 1 || len(memories.updateTraces) != 1 || len(memories.updateEdges) != 1 {
		t.Fatalf("expected one update memory, trace, and edge; got memories=%d traces=%d edges=%d", len(memories.updateMemories), len(memories.updateTraces), len(memories.updateEdges))
	}

	memory := memories.updateMemories[0]
	if memory.ID == "" || !strings.HasPrefix(memory.ID, "mem_") {
		t.Fatalf("expected deterministic memory ID, got %q", memory.ID)
	}
	if memory.Status != core.MemoryStatusActive || !memory.LatestFlag {
		t.Fatalf("update memory should become the active latest memory, got status=%q latest=%v", memory.Status, memory.LatestFlag)
	}
	if memory.Scope != operation.Memory.Scope || memory.OwnerEntityID != operation.Memory.OwnerEntityID {
		t.Fatalf("update memory did not preserve mutation scope/owner: %#v", memory)
	}

	trace := memories.updateTraces[0]
	if trace.MemoryID != memory.ID {
		t.Fatalf("trace memory ID mismatch: got %q want %q", trace.MemoryID, memory.ID)
	}
	assertJSONContains(t, trace.AppliedOperationsJSON, "op_update")

	edge := memories.updateEdges[0]
	if edge.FromMemoryID != memory.ID {
		t.Fatalf("updates edge must originate from the written memory: got %q want %q", edge.FromMemoryID, memory.ID)
	}
	if edge.ToMemoryID != operation.Memory.TargetID {
		t.Fatalf("updates edge target mismatch: got %q want %q", edge.ToMemoryID, operation.Memory.TargetID)
	}
	if edge.EdgeKind != core.EdgeKindUpdates {
		t.Fatalf("updates edge kind mismatch: got %q", edge.EdgeKind)
	}
}

func TestStoreBackedApplyEngine_RejectsUnsupportedWrites(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		mutateReq func(*ApplyRequest)
		wantError string
	}{
		{
			name: "archive remains not implemented",
			mutateReq: func(req *ApplyRequest) {
				req.Reasoning.Stage2.Operations = []reasoning.GraphOperation{validArchiveOperation()}
			},
			wantError: "archive status handling",
		},
		{
			name: "group shared waits for membership validation",
			mutateReq: func(req *ApplyRequest) {
				groupID := "group_1"
				req.Reasoning.Stage2.Operations[0].Memory.Scope = core.MemoryScopeGroupShared
				req.Reasoning.Stage2.Operations[0].Memory.GroupID = &groupID
			},
			wantError: "membership validation",
		},
		{
			name: "profile delta waits for merge implementation",
			mutateReq: func(req *ApplyRequest) {
				req.Reasoning.Stage2.ProfileDelta = json.RawMessage(`{"static":{"tone":"brief"}}`)
			},
			wantError: "profile_delta writes are not implemented",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			memories := &fakeMemoryTraceCreator{}
			engine := newTestStoreBackedApplyEngine(t, memories)
			req := validApplyRequest(validCreateOperation())
			tt.mutateReq(req)

			_, err := engine.Apply(context.Background(), req)
			if err == nil {
				t.Fatalf("expected Apply to reject unsupported write")
			}
			if !errors.Is(err, core.ErrNotImplemented) {
				t.Fatalf("expected ErrNotImplemented, got %v", err)
			}
			if !strings.Contains(err.Error(), tt.wantError) {
				t.Fatalf("expected error to contain %q, got %v", tt.wantError, err)
			}
			if len(memories.memories) != 0 || len(memories.traces) != 0 || len(memories.edges) != 0 {
				t.Fatalf("unsupported write should not touch storage: memories=%d traces=%d edges=%d", len(memories.memories), len(memories.traces), len(memories.edges))
			}
		})
	}
}

func TestStoreBackedApplyEngine_TraceFailureDoesNotReportSuccessfulApply(t *testing.T) {
	t.Parallel()

	memories := &fakeMemoryTraceCreator{err: errors.New("trace insert failed")}
	engine := newTestStoreBackedApplyEngine(t, memories)

	result, err := engine.Apply(context.Background(), validApplyRequest(validCreateOperation()))
	if err == nil {
		t.Fatalf("expected Apply to fail when memory trace persistence fails")
	}
	if result != nil {
		t.Fatalf("failed trace persistence must not return a successful apply result: %#v", result)
	}
	if len(memories.memories) != 0 || len(memories.traces) != 0 || len(memories.edges) != 0 {
		t.Fatalf("fake atomic store should record no successful writes on trace failure: memories=%d traces=%d edges=%d", len(memories.memories), len(memories.traces), len(memories.edges))
	}
}

func TestStoreBackedApplyEngine_ExtendEdgeFailureDoesNotReportSuccessfulApply(t *testing.T) {
	t.Parallel()

	memories := &fakeMemoryTraceCreator{edgeErr: errors.New("edge insert failed")}
	engine := newTestStoreBackedApplyEngine(t, memories)

	result, err := engine.Apply(context.Background(), validApplyRequest(validExtendOperation()))
	if err == nil {
		t.Fatalf("expected Apply to fail when extension edge persistence fails")
	}
	if !strings.Contains(err.Error(), "edge insert failed") {
		t.Fatalf("expected edge persistence error, got %v", err)
	}
	if result != nil {
		t.Fatalf("failed edge persistence must not return a successful apply result: %#v", result)
	}
	if len(memories.memories) != 0 || len(memories.traces) != 0 || len(memories.edges) != 0 {
		t.Fatalf("fake atomic store should record no successful writes on edge failure: memories=%d traces=%d edges=%d", len(memories.memories), len(memories.traces), len(memories.edges))
	}
}

func TestStoreBackedApplyEngine_UpdateFailureDoesNotReportSuccessfulApply(t *testing.T) {
	t.Parallel()

	memories := &fakeMemoryTraceCreator{updateErr: errors.New("target not latest")}
	engine := newTestStoreBackedApplyEngine(t, memories)

	result, err := engine.Apply(context.Background(), validApplyRequest(validUpdateOperation()))
	if err == nil {
		t.Fatalf("expected Apply to fail when update persistence fails")
	}
	if !strings.Contains(err.Error(), "target not latest") {
		t.Fatalf("expected update persistence error, got %v", err)
	}
	if result != nil {
		t.Fatalf("failed update persistence must not return a successful apply result: %#v", result)
	}
	if len(memories.updateMemories) != 0 || len(memories.updateTraces) != 0 || len(memories.updateEdges) != 0 {
		t.Fatalf("fake atomic store should record no successful update writes: memories=%d traces=%d edges=%d", len(memories.updateMemories), len(memories.updateTraces), len(memories.updateEdges))
	}
}

var fixedApplyTime = time.Date(2026, 4, 24, 12, 0, 0, 0, time.UTC)

func newTestStoreBackedApplyEngine(t *testing.T, memories *fakeMemoryTraceCreator) *StoreBackedApplyEngine {
	t.Helper()

	engine, err := NewStoreBackedApplyEngine(memories)
	if err != nil {
		t.Fatalf("NewStoreBackedApplyEngine returned error: %v", err)
	}
	engine.now = func() time.Time {
		return fixedApplyTime
	}
	return engine
}

func assertJSONContains(t *testing.T, raw json.RawMessage, want string) {
	t.Helper()

	if !json.Valid(raw) {
		t.Fatalf("expected valid JSON, got %s", string(raw))
	}
	if !strings.Contains(string(raw), want) {
		t.Fatalf("expected JSON to contain %q, got %s", want, string(raw))
	}
}

type fakeMemoryTraceCreator struct {
	memories       []*core.Memory
	traces         []*core.MemoryTrace
	edges          []*core.MemoryEdge
	updateMemories []*core.Memory
	updateTraces   []*core.MemoryTrace
	updateEdges    []*core.MemoryEdge
	err            error
	edgeErr        error
	updateErr      error
}

func (s *fakeMemoryTraceCreator) CreateMemoryWithTrace(_ context.Context, memory *core.Memory, trace *core.MemoryTrace) error {
	if s.err != nil {
		return s.err
	}
	s.memories = append(s.memories, memory)
	s.traces = append(s.traces, trace)
	return nil
}

func (s *fakeMemoryTraceCreator) CreateMemoryWithTraceAndEdge(_ context.Context, memory *core.Memory, trace *core.MemoryTrace, edge *core.MemoryEdge) error {
	if s.err != nil {
		return s.err
	}
	if s.edgeErr != nil {
		return s.edgeErr
	}
	s.memories = append(s.memories, memory)
	s.traces = append(s.traces, trace)
	s.edges = append(s.edges, edge)
	return nil
}

func (s *fakeMemoryTraceCreator) CreateMemoryWithTraceAndUpdateEdge(_ context.Context, memory *core.Memory, trace *core.MemoryTrace, edge *core.MemoryEdge) error {
	if s.updateErr != nil {
		return s.updateErr
	}
	s.updateMemories = append(s.updateMemories, memory)
	s.updateTraces = append(s.updateTraces, trace)
	s.updateEdges = append(s.updateEdges, edge)
	return nil
}

```



<!-- Source: internal/hermes/provider_test.go | bytes=10507 | lines=274 | sha16=315a530430c500d8 -->

```go
// ============================================================
// FILE     : internal/hermes/provider_test.go
// PURPOSE  : Verifies Hermes provider lifecycle hooks delegate to VibeGravity core semantics.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : Hermes provider adapter tests
// DEPENDS  : context, errors, strings, testing, internal/core
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: These tests use a fake core service; they do not call a real Hermes runtime.
// ============================================================

package hermes

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

func TestProviderDelegatesPrefetchAndRendersContext(t *testing.T) {
	t.Parallel()

	service := &fakeService{
		prefetchResp: &core.PrefetchResponse{
			Blocks: []core.RecallBlock{
				{Kind: "pinned_note", Priority: 100, Text: "Keep Hermes first.", Scope: core.MemoryScopeWorkspaceShared, Source: "notes", Freshness: "stored"},
				{Kind: "active_plan", Priority: 95, Text: "Finish V1 core semantics."},
			},
			Meta: core.RecallMeta{EstimatedTokens: 12, Sources: []string{"notes", "plans"}},
		},
	}
	provider := newTestProvider(t, service)

	resp, err := provider.Prefetch(context.Background(), &core.PrefetchRequest{
		TenantID: "tenant_1", WorkspaceID: "workspace_1", SessionID: "session_1", ActorID: "agent:hermes-main",
	})
	if err != nil {
		t.Fatalf("Prefetch returned error: %v", err)
	}
	if service.prefetchCalls != 1 {
		t.Fatalf("expected one prefetch call, got %d", service.prefetchCalls)
	}
	rendered := provider.RenderContext(resp)
	if !strings.Contains(rendered, "[pinned_note:100:workspace_shared:notes:stored] Keep Hermes first.") {
		t.Fatalf("rendered context lost pinned note: %q", rendered)
	}
	if !strings.Contains(rendered, "[active_plan:95] Finish V1 core semantics.") {
		t.Fatalf("rendered context lost active plan: %q", rendered)
	}
}

func TestProviderDelegatesSyncTurn(t *testing.T) {
	t.Parallel()

	service := &fakeService{syncResp: &core.SyncTurnResponse{Status: "accepted", EventIDs: []string{"evt_1"}, JobIDs: []string{"job_1"}}}
	provider := newTestProvider(t, service)

	resp, err := provider.SyncTurn(context.Background(), &core.SyncTurnRequest{
		TenantID: "tenant_1", WorkspaceID: "workspace_1", SessionID: "session_1", ActorID: "agent:hermes-main",
	})
	if err != nil {
		t.Fatalf("SyncTurn returned error: %v", err)
	}
	if service.syncCalls != 1 {
		t.Fatalf("expected one sync call, got %d", service.syncCalls)
	}
	if resp.Status != "accepted" || len(resp.JobIDs) != 1 {
		t.Fatalf("unexpected sync response: %#v", resp)
	}
}

func TestProviderAvailabilityReflectsPrefetchHealth(t *testing.T) {
	t.Parallel()

	okProvider := newTestProvider(t, &fakeService{prefetchResp: &core.PrefetchResponse{}})
	if !okProvider.IsAvailable(context.Background(), &core.PrefetchRequest{TenantID: "tenant_1"}) {
		t.Fatalf("expected provider to be available when prefetch succeeds")
	}

	failingProvider := newTestProvider(t, &fakeService{prefetchErr: errors.New("database unavailable")})
	if failingProvider.IsAvailable(context.Background(), &core.PrefetchRequest{TenantID: "tenant_1"}) {
		t.Fatalf("expected provider to be unavailable when prefetch fails")
	}
}

func TestProviderToolsExposeV1HermesSurface(t *testing.T) {
	t.Parallel()

	provider := newTestProvider(t, &fakeService{})
	tools := provider.GetTools()
	names := make(map[string]bool, len(tools))
	for _, tool := range tools {
		names[tool.Name] = true
	}
	for _, want := range []string{"recall_preview", "search_memory", "add_note", "show_plan", "explain_memory", "correct_memory", "view_timeline", "degraded_status"} {
		if !names[want] {
			t.Fatalf("expected provider tool %q in %#v", want, tools)
		}
	}
}

func TestProviderCallToolDelegatesRecallPreview(t *testing.T) {
	t.Parallel()

	service := &fakeService{prefetchResp: &core.PrefetchResponse{Blocks: []core.RecallBlock{{Text: "Use scoped recall."}}}}
	provider := newTestProvider(t, service)

	raw, err := provider.CallTool(context.Background(), "recall_preview", json.RawMessage(`{"tenant_id":"tenant_1","workspace_id":"workspace_1"}`))
	if err != nil {
		t.Fatalf("CallTool returned error: %v", err)
	}
	if service.prefetchCalls != 1 {
		t.Fatalf("expected one prefetch call, got %d", service.prefetchCalls)
	}
	if !strings.Contains(string(raw), "Use scoped recall.") {
		t.Fatalf("expected encoded recall preview output, got %s", string(raw))
	}
}

func TestProviderCallToolDelegatesTrustLoopTools(t *testing.T) {
	t.Parallel()

	service := &fakeService{
		searchResp:   &core.SearchMemoriesResponse{Memories: []core.MemoryResult{{MemoryID: "mem_1", Text: "Project rule"}}},
		addNoteResp:  &core.AddNoteResponse{NoteID: "note_1", Status: "created"},
		explainResp:  &core.ExplainMemoryResponse{MemoryID: "mem_1", Trace: core.MemoryTraceResult{ReasoningJobID: "job_1"}},
		correctResp:  &core.CorrectMemoryResponse{MemoryID: "mem_1", CorrectionRecorded: true},
		timelineResp: &core.GetTimelineResponse{Items: []core.TimelineItem{{ID: "item_1", Text: "Correction recorded"}}},
	}
	provider := newTestProvider(t, service)

	cases := []struct {
		name string
		raw  json.RawMessage
		want string
	}{
		{name: "search_memory", raw: json.RawMessage(`{"tenant_id":"tenant_1","workspace_id":"workspace_1","query":"rule"}`), want: "Project rule"},
		{name: "add_note", raw: json.RawMessage(`{"tenant_id":"tenant_1","workspace_id":"workspace_1","text":"Pin this"}`), want: "note_1"},
		{name: "explain_memory", raw: json.RawMessage(`{"tenant_id":"tenant_1","workspace_id":"workspace_1","memory_id":"mem_1"}`), want: "job_1"},
		{name: "correct_memory", raw: json.RawMessage(`{"tenant_id":"tenant_1","workspace_id":"workspace_1","memory_id":"mem_1"}`), want: "correction_recorded"},
		{name: "view_timeline", raw: json.RawMessage(`{"tenant_id":"tenant_1","workspace_id":"workspace_1","entity_id":"agent:hermes-main"}`), want: "Correction recorded"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := provider.CallTool(context.Background(), tc.name, tc.raw)
			if err != nil {
				t.Fatalf("CallTool returned error: %v", err)
			}
			if !strings.Contains(string(raw), tc.want) {
				t.Fatalf("expected output to contain %q, got %s", tc.want, string(raw))
			}
		})
	}
}

func TestProviderCallToolReturnsDegradedStatusFromPrefetchMeta(t *testing.T) {
	t.Parallel()

	service := &fakeService{prefetchResp: &core.PrefetchResponse{Meta: core.RecallMeta{Freshness: "stale", Degraded: true, DegradedReasons: []string{"worker backlog"}}}}
	provider := newTestProvider(t, service)

	raw, err := provider.CallTool(context.Background(), "degraded_status", json.RawMessage(`{"tenant_id":"tenant_1","workspace_id":"workspace_1"}`))
	if err != nil {
		t.Fatalf("CallTool returned error: %v", err)
	}
	if service.prefetchCalls != 1 {
		t.Fatalf("expected one prefetch call, got %d", service.prefetchCalls)
	}
	if !strings.Contains(string(raw), `"freshness":"stale"`) || !strings.Contains(string(raw), "worker backlog") {
		t.Fatalf("expected encoded recall meta, got %s", string(raw))
	}
}

func TestProviderCallToolRejectsUnknownInvalidAndUnbackedTools(t *testing.T) {
	t.Parallel()

	provider := newTestProvider(t, &fakeService{})
	if _, err := provider.CallTool(context.Background(), "unknown", json.RawMessage(`{}`)); !errors.Is(err, core.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument for unknown tool, got %v", err)
	}
	if _, err := provider.CallTool(context.Background(), "recall_preview", json.RawMessage(`{`)); !errors.Is(err, core.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument for invalid JSON, got %v", err)
	}
	if _, err := provider.CallTool(context.Background(), "show_plan", json.RawMessage(`{}`)); !errors.Is(err, core.ErrNotImplemented) {
		t.Fatalf("expected ErrNotImplemented for show_plan, got %v", err)
	}
}

func newTestProvider(t *testing.T, service core.VibeGravityService) *Provider {
	t.Helper()

	provider, err := NewProvider(service)
	if err != nil {
		t.Fatalf("NewProvider returned error: %v", err)
	}
	return provider
}

type fakeService struct {
	prefetchCalls int
	syncCalls     int
	searchCalls   int
	addNoteCalls  int
	explainCalls  int
	correctCalls  int
	timelineCalls int
	prefetchResp  *core.PrefetchResponse
	prefetchErr   error
	syncResp      *core.SyncTurnResponse
	syncErr       error
	searchResp    *core.SearchMemoriesResponse
	addNoteResp   *core.AddNoteResponse
	explainResp   *core.ExplainMemoryResponse
	correctResp   *core.CorrectMemoryResponse
	timelineResp  *core.GetTimelineResponse
}

func (s *fakeService) Prefetch(context.Context, *core.PrefetchRequest) (*core.PrefetchResponse, error) {
	s.prefetchCalls++
	return s.prefetchResp, s.prefetchErr
}

func (s *fakeService) SyncTurn(context.Context, *core.SyncTurnRequest) (*core.SyncTurnResponse, error) {
	s.syncCalls++
	return s.syncResp, s.syncErr
}

func (s *fakeService) AddDocument(context.Context, *core.AddDocumentRequest) (*core.AddDocumentResponse, error) {
	return nil, core.ErrNotImplemented
}

func (s *fakeService) SearchMemories(context.Context, *core.SearchMemoriesRequest) (*core.SearchMemoriesResponse, error) {
	s.searchCalls++
	return s.searchResp, nil
}

func (s *fakeService) SearchDocuments(context.Context, *core.SearchDocumentsRequest) (*core.SearchDocumentsResponse, error) {
	return nil, core.ErrNotImplemented
}

func (s *fakeService) AddNote(context.Context, *core.AddNoteRequest) (*core.AddNoteResponse, error) {
	s.addNoteCalls++
	return s.addNoteResp, nil
}

func (s *fakeService) CreatePlan(context.Context, *core.CreatePlanRequest) (*core.CreatePlanResponse, error) {
	return nil, core.ErrNotImplemented
}

func (s *fakeService) UpdatePlan(context.Context, *core.UpdatePlanRequest) (*core.UpdatePlanResponse, error) {
	return nil, core.ErrNotImplemented
}

func (s *fakeService) CorrectMemory(context.Context, *core.CorrectMemoryRequest) (*core.CorrectMemoryResponse, error) {
	s.correctCalls++
	return s.correctResp, nil
}

func (s *fakeService) GetTimeline(context.Context, *core.GetTimelineRequest) (*core.GetTimelineResponse, error) {
	s.timelineCalls++
	return s.timelineResp, nil
}

func (s *fakeService) ExplainMemory(context.Context, *core.ExplainMemoryRequest) (*core.ExplainMemoryResponse, error) {
	s.explainCalls++
	return s.explainResp, nil
}

```



<!-- Source: internal/httpapi/router_test.go | bytes=11633 | lines=304 | sha16=89f5d8be3435d63b -->

```go
// ============================================================
// FILE     : internal/httpapi/router_test.go
// PURPOSE  : Verifies HTTP transport handlers delegate to the core service contract.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : TestPrefetchHandler_CallsService, TestSyncTurnHandler_CallsService
// DEPENDS  : internal/httpapi, internal/core
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: Keep handler tests about transport behavior; product rules belong in service tests.
// ============================================================

package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

func TestPrefetchHandler_CallsService(t *testing.T) {
	t.Parallel()

	service := &fakeVibeGravityService{}
	router := NewRouter(&App{Service: service})
	body := `{"tenant_id":"tenant_1","workspace_id":"workspace_1","session_id":"session_1","actor_id":"agent:hermes-main","query":"next","budget_tokens":2200,"mode":"default"}`

	req := httptest.NewRequest(http.MethodPost, "/v1/prefetch", strings.NewReader(body))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if service.prefetchReq == nil || service.prefetchReq.ActorID != "agent:hermes-main" {
		t.Fatalf("service did not receive prefetch request: %#v", service.prefetchReq)
	}
	if !strings.Contains(rr.Body.String(), "pinned_note") {
		t.Fatalf("response body did not include recall block: %s", rr.Body.String())
	}
}

func TestSyncTurnHandler_CallsService(t *testing.T) {
	t.Parallel()

	service := &fakeVibeGravityService{}
	router := NewRouter(&App{Service: service})
	body := `{"tenant_id":"tenant_1","workspace_id":"workspace_1","session_id":"session_1","actor_id":"agent:hermes-main","idempotency_key":"turn_1","turn_events":[{"event_kind":"user_message","source":"hermes","fingerprint":"fp_1","occurred_at":"2026-04-24T00:00:00Z","payload_json":{"text":"hello"}}]}`

	req := httptest.NewRequest(http.MethodPost, "/v1/sync-turn", strings.NewReader(body))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d: %s", rr.Code, rr.Body.String())
	}
	if service.syncReq == nil || service.syncReq.IdempotencyKey != "turn_1" {
		t.Fatalf("service did not receive sync request: %#v", service.syncReq)
	}
	if !strings.Contains(rr.Body.String(), "accepted") {
		t.Fatalf("response body did not include accepted status: %s", rr.Body.String())
	}
}

func TestHealthz_ReturnsUnavailableWithoutDBPool(t *testing.T) {
	t.Parallel()

	router := NewRouter(&App{})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d: %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "database unavailable") {
		t.Fatalf("response body did not explain database unavailability: %s", rr.Body.String())
	}
}

func TestAddDocumentHandler_CallsService(t *testing.T) {
	t.Parallel()

	service := &fakeVibeGravityService{}
	router := NewRouter(&App{Service: service})
	body := `{"tenant_id":"tenant_1","workspace_id":"workspace_1","source":"operator_upload","title":"Runtime Notes","content":"important context"}`

	req := httptest.NewRequest(http.MethodPost, "/v1/documents", strings.NewReader(body))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
	if service.addDocumentReq == nil || service.addDocumentReq.Title != "Runtime Notes" {
		t.Fatalf("service did not receive add document request: %#v", service.addDocumentReq)
	}
	if !strings.Contains(rr.Body.String(), "doc_1") {
		t.Fatalf("response body did not include document id: %s", rr.Body.String())
	}
}

func TestUpdatePlanHandler_UsesPathID(t *testing.T) {
	t.Parallel()

	service := &fakeVibeGravityService{}
	router := NewRouter(&App{Service: service})
	body := `{"tenant_id":"tenant_1","workspace_id":"workspace_1","plan_id":"body_value","status":"active"}`

	req := httptest.NewRequest(http.MethodPatch, "/v1/plans/plan_path", strings.NewReader(body))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if service.updatePlanReq == nil || service.updatePlanReq.PlanID != "plan_path" {
		t.Fatalf("service did not receive path plan id: %#v", service.updatePlanReq)
	}
	if !strings.Contains(rr.Body.String(), "updated") {
		t.Fatalf("response body did not include updated status: %s", rr.Body.String())
	}
}

func TestCorrectMemoryHandler_CallsService(t *testing.T) {
	t.Parallel()

	service := &fakeVibeGravityService{}
	router := NewRouter(&App{Service: service})
	body := `{"tenant_id":"tenant_1","workspace_id":"workspace_1","memory_id":"mem_1","operator_id":"operator_1","idempotency_key":"correction_1","correction_text":"Use the newer fact."}`

	req := httptest.NewRequest(http.MethodPost, "/v1/memory/correct", strings.NewReader(body))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if service.correctMemoryReq == nil || service.correctMemoryReq.MemoryID != "mem_1" {
		t.Fatalf("service did not receive correction request: %#v", service.correctMemoryReq)
	}
}

func TestGetTimelineHandler_CallsService(t *testing.T) {
	t.Parallel()

	service := &fakeVibeGravityService{}
	router := NewRouter(&App{Service: service})

	req := httptest.NewRequest(http.MethodGet, "/v1/timeline?tenant_id=tenant_1&workspace_id=workspace_1&entity_id=agent:hermes-main&scopes=agent_private,workspace_shared&from=2026-04-24T00:00:00Z&to=2026-04-25T00:00:00Z&limit=25", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if service.timelineReq == nil || service.timelineReq.EntityID != "agent:hermes-main" {
		t.Fatalf("service did not receive timeline request: %#v", service.timelineReq)
	}
	if got := service.timelineReq.Scopes; len(got) != 2 || got[0] != core.MemoryScopeAgentPrivate || got[1] != core.MemoryScopeWorkspaceShared {
		t.Fatalf("handler did not parse scopes: %#v", got)
	}
	if service.timelineReq.From == nil || service.timelineReq.To == nil || service.timelineReq.Limit != 25 {
		t.Fatalf("handler did not parse from/to/limit: %#v", service.timelineReq)
	}
	if !service.timelineReq.From.Equal(time.Date(2026, 4, 24, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected from time: %v", service.timelineReq.From)
	}
}

func TestGetTimelineHandler_RejectsInvalidQuery(t *testing.T) {
	t.Parallel()

	service := &fakeVibeGravityService{}
	router := NewRouter(&App{Service: service})

	req := httptest.NewRequest(http.MethodGet, "/v1/timeline?tenant_id=tenant_1&workspace_id=workspace_1&entity_id=agent:hermes-main&limit=not-a-number", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
	if service.timelineReq != nil {
		t.Fatalf("service should not receive invalid timeline request: %#v", service.timelineReq)
	}
}

func TestExplainMemoryHandler_CallsServiceWithVisibility(t *testing.T) {
	t.Parallel()

	service := &fakeVibeGravityService{}
	router := NewRouter(&App{Service: service})

	req := httptest.NewRequest(http.MethodGet, "/v1/memory/mem_1/explain?tenant_id=tenant_1&workspace_id=workspace_1&entity_id=agent:hermes-main&visible_group_ids=group_design,group_ops", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if service.explainReq == nil || service.explainReq.MemoryID != "mem_1" || service.explainReq.EntityID != "agent:hermes-main" {
		t.Fatalf("service did not receive explain visibility request: %#v", service.explainReq)
	}
	if got := service.explainReq.VisibleGroupIDs; len(got) != 2 || got[0] != "group_design" || got[1] != "group_ops" {
		t.Fatalf("handler did not parse visible group ids: %#v", service.explainReq)
	}
}

type fakeVibeGravityService struct {
	prefetchReq      *core.PrefetchRequest
	syncReq          *core.SyncTurnRequest
	addDocumentReq   *core.AddDocumentRequest
	updatePlanReq    *core.UpdatePlanRequest
	correctMemoryReq *core.CorrectMemoryRequest
	timelineReq      *core.GetTimelineRequest
	explainReq       *core.ExplainMemoryRequest
}

func (s *fakeVibeGravityService) Prefetch(_ context.Context, req *core.PrefetchRequest) (*core.PrefetchResponse, error) {
	s.prefetchReq = req
	return &core.PrefetchResponse{
		Blocks: []core.RecallBlock{{
			Kind:     "pinned_note",
			Priority: 100,
			Text:     "Keep the Hermes-first plan visible.",
		}},
		Meta: core.RecallMeta{
			EstimatedTokens: 8,
			Sources:         []string{"notes"},
		},
	}, nil
}

func (s *fakeVibeGravityService) SyncTurn(_ context.Context, req *core.SyncTurnRequest) (*core.SyncTurnResponse, error) {
	s.syncReq = req
	return &core.SyncTurnResponse{
		Status:         "accepted",
		SessionID:      req.SessionID,
		EventIDs:       []string{"evt_1"},
		JobIDs:         []string{"job_1"},
		DuplicateCount: 0,
	}, nil
}

func (s *fakeVibeGravityService) AddDocument(_ context.Context, req *core.AddDocumentRequest) (*core.AddDocumentResponse, error) {
	s.addDocumentReq = req
	return &core.AddDocumentResponse{
		DocumentID: "doc_1",
		ChunkIDs:   []string{"chunk_1"},
		Status:     "created",
	}, nil
}

func (s *fakeVibeGravityService) SearchMemories(context.Context, *core.SearchMemoriesRequest) (*core.SearchMemoriesResponse, error) {
	return nil, core.ErrNotImplemented
}

func (s *fakeVibeGravityService) SearchDocuments(context.Context, *core.SearchDocumentsRequest) (*core.SearchDocumentsResponse, error) {
	return nil, core.ErrNotImplemented
}

func (s *fakeVibeGravityService) AddNote(context.Context, *core.AddNoteRequest) (*core.AddNoteResponse, error) {
	return nil, core.ErrNotImplemented
}

func (s *fakeVibeGravityService) CreatePlan(context.Context, *core.CreatePlanRequest) (*core.CreatePlanResponse, error) {
	return nil, core.ErrNotImplemented
}

func (s *fakeVibeGravityService) UpdatePlan(_ context.Context, req *core.UpdatePlanRequest) (*core.UpdatePlanResponse, error) {
	s.updatePlanReq = req
	return &core.UpdatePlanResponse{PlanID: req.PlanID, Status: "updated"}, nil
}

func (s *fakeVibeGravityService) CorrectMemory(_ context.Context, req *core.CorrectMemoryRequest) (*core.CorrectMemoryResponse, error) {
	s.correctMemoryReq = req
	return &core.CorrectMemoryResponse{
		MemoryID:           req.MemoryID,
		RawEventID:         "evt_correction",
		CorrectionID:       "corr_1",
		CorrectionRecorded: true,
		TraceWritten:       false,
		Status:             "accepted",
	}, nil
}

func (s *fakeVibeGravityService) GetTimeline(_ context.Context, req *core.GetTimelineRequest) (*core.GetTimelineResponse, error) {
	s.timelineReq = req
	return &core.GetTimelineResponse{Items: []core.TimelineItem{}}, nil
}

func (s *fakeVibeGravityService) ExplainMemory(_ context.Context, req *core.ExplainMemoryRequest) (*core.ExplainMemoryResponse, error) {
	s.explainReq = req
	return &core.ExplainMemoryResponse{MemoryID: req.MemoryID}, nil
}

```



<!-- Source: internal/ingest/service_test.go | bytes=5981 | lines=205 | sha16=77803e93b205356b -->

```go
// ============================================================
// FILE     : internal/ingest/service_test.go
// PURPOSE  : Verifies sync_turn idempotency, validation, and job enqueue behavior.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : TestServiceSyncTurn_EnqueuesJobForNewEvents, TestServiceSyncTurn_ReplayedTurnIsDuplicate
// DEPENDS  : internal/ingest, internal/core
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: Keep these tests focused on the API hot path; reasoning belongs to worker tests.
// ============================================================

package ingest

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

func TestServiceSyncTurn_EnqueuesJobForNewEvents(t *testing.T) {
	t.Parallel()

	rawEvents := newFakeRawEventStore()
	jobs := &fakeJobStore{}
	service := newTestService(t, rawEvents, jobs)

	resp, err := service.SyncTurn(context.Background(), testSyncTurnRequest())
	if err != nil {
		t.Fatalf("SyncTurn returned error: %v", err)
	}

	if resp.Status != "accepted" {
		t.Fatalf("unexpected status: %s", resp.Status)
	}
	if len(resp.EventIDs) != 2 {
		t.Fatalf("expected 2 event IDs, got %d", len(resp.EventIDs))
	}
	if len(resp.JobIDs) != 1 {
		t.Fatalf("expected 1 job ID, got %d", len(resp.JobIDs))
	}
	if resp.DuplicateCount != 0 {
		t.Fatalf("expected duplicate count 0, got %d", resp.DuplicateCount)
	}
	if len(rawEvents.events) != 2 {
		t.Fatalf("expected 2 stored events, got %d", len(rawEvents.events))
	}
	if len(jobs.jobs) != 1 {
		t.Fatalf("expected 1 stored job, got %d", len(jobs.jobs))
	}
	if jobs.jobs[0].JobKind != core.JobKindProcessTurnEvent {
		t.Fatalf("unexpected job kind: %s", jobs.jobs[0].JobKind)
	}
	if len(jobs.jobs[0].RawEventIDs) != 2 {
		t.Fatalf("expected job to reference 2 events, got %d", len(jobs.jobs[0].RawEventIDs))
	}
}

func TestServiceSyncTurn_ReplayedTurnIsDuplicate(t *testing.T) {
	t.Parallel()

	rawEvents := newFakeRawEventStore()
	jobs := &fakeJobStore{}
	service := newTestService(t, rawEvents, jobs)
	req := testSyncTurnRequest()

	if _, err := service.SyncTurn(context.Background(), req); err != nil {
		t.Fatalf("first SyncTurn returned error: %v", err)
	}
	resp, err := service.SyncTurn(context.Background(), req)
	if err != nil {
		t.Fatalf("second SyncTurn returned error: %v", err)
	}

	if len(resp.EventIDs) != 0 {
		t.Fatalf("expected duplicate replay to return no new events, got %d", len(resp.EventIDs))
	}
	if len(resp.JobIDs) != 0 {
		t.Fatalf("expected duplicate replay to enqueue no jobs, got %d", len(resp.JobIDs))
	}
	if resp.DuplicateCount != 2 {
		t.Fatalf("expected duplicate count 2, got %d", resp.DuplicateCount)
	}
	if len(jobs.jobs) != 1 {
		t.Fatalf("expected only the first call to enqueue a job, got %d jobs", len(jobs.jobs))
	}
}

func TestServiceSyncTurn_ValidatesRequiredFields(t *testing.T) {
	t.Parallel()

	service := newTestService(t, newFakeRawEventStore(), &fakeJobStore{})
	req := testSyncTurnRequest()
	req.TenantID = ""

	_, err := service.SyncTurn(context.Background(), req)
	if !errors.Is(err, core.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}

func newTestService(t *testing.T, rawEvents *fakeRawEventStore, jobs *fakeJobStore) *Service {
	t.Helper()
	service, err := NewService(Dependencies{
		RawEvents: rawEvents,
		Jobs:      jobs,
		Clock: func() time.Time {
			return time.Date(2026, time.April, 24, 0, 0, 0, 0, time.UTC)
		},
	})
	if err != nil {
		t.Fatalf("NewService returned error: %v", err)
	}
	return service
}

func testSyncTurnRequest() *core.SyncTurnRequest {
	return &core.SyncTurnRequest{
		TenantID:       "tenant_1",
		WorkspaceID:    "workspace_1",
		SessionID:      "session_1",
		ActorID:        "agent:hermes-main",
		IdempotencyKey: "turn_1",
		TurnEvents: []core.RawEventPayload{
			{
				EventKind:   "user_message",
				Source:      "hermes",
				Fingerprint: "fp_user",
				OccurredAt:  time.Date(2026, time.April, 24, 0, 1, 0, 0, time.UTC),
				PayloadJSON: json.RawMessage(`{"text":"Remember the plan."}`),
			},
			{
				EventKind:   "assistant_message",
				Source:      "hermes",
				Fingerprint: "fp_assistant",
				OccurredAt:  time.Date(2026, time.April, 24, 0, 2, 0, 0, time.UTC),
				PayloadJSON: json.RawMessage(`{"text":"Acknowledged."}`),
			},
		},
	}
}

type fakeRawEventStore struct {
	seen   map[string]struct{}
	events []*core.RawEvent
}

func newFakeRawEventStore() *fakeRawEventStore {
	return &fakeRawEventStore{
		seen: make(map[string]struct{}),
	}
}

func (s *fakeRawEventStore) AppendRawEvents(_ context.Context, events []*core.RawEvent) ([]string, error) {
	ids := make([]string, 0, len(events))
	for _, event := range events {
		key := event.TenantID + "\x00" + event.Source + "\x00" + event.IdempotencyKey
		if _, ok := s.seen[key]; ok {
			continue
		}
		s.seen[key] = struct{}{}
		s.events = append(s.events, event)
		ids = append(ids, event.ID)
	}
	return ids, nil
}

func (s *fakeRawEventStore) GetRawEvents(_ context.Context, _ []string) ([]*core.RawEvent, error) {
	return nil, core.ErrNotImplemented
}

type fakeJobStore struct {
	jobs []*core.IngestJob
}

func (s *fakeJobStore) EnqueueJobs(_ context.Context, jobs []*core.IngestJob) ([]string, error) {
	ids := make([]string, 0, len(jobs))
	for _, job := range jobs {
		s.jobs = append(s.jobs, job)
		ids = append(ids, job.ID)
	}
	return ids, nil
}

func (s *fakeJobStore) ClaimJobs(context.Context, string, int) ([]*core.IngestJob, error) {
	return nil, core.ErrNotImplemented
}

func (s *fakeJobStore) CompleteJob(context.Context, string) error {
	return core.ErrNotImplemented
}

func (s *fakeJobStore) FailJob(context.Context, string, error) error {
	return core.ErrNotImplemented
}

func (s *fakeJobStore) BlockJob(context.Context, string, error) error {
	return core.ErrNotImplemented
}

```



<!-- Source: internal/kernel/service_test.go | bytes=18655 | lines=560 | sha16=5bb741f0b98b04c5 -->

```go
// ============================================================
// FILE     : internal/kernel/service_test.go
// PURPOSE  : Verifies kernel-level document and plan API behavior.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : kernel service tests
// DEPENDS  : context, testing, internal/core
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: Keep these tests focused on service composition, not PostgreSQL details.
// ============================================================

package kernel

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

func TestAddDocumentStoresDocumentAndChunks(t *testing.T) {
	t.Parallel()

	documents := &fakeDocumentStore{}
	service := &Service{documents: documents}

	resp, err := service.AddDocument(context.Background(), &core.AddDocumentRequest{
		TenantID:    "tenant_1",
		WorkspaceID: "workspace_1",
		Source:      "operator_upload",
		Title:       "Runtime Notes",
		Content:     strings.Repeat("A", documentChunkMaxRunes+5),
	})
	if err != nil {
		t.Fatalf("AddDocument returned error: %v", err)
	}
	if resp.DocumentID != "doc_test" || resp.Status != "created" {
		t.Fatalf("unexpected response: %#v", resp)
	}
	if documents.document == nil || documents.document.Fingerprint == "" {
		t.Fatalf("document was not stored with a fingerprint: %#v", documents.document)
	}
	if len(documents.chunks) != 2 || len(resp.ChunkIDs) != 2 {
		t.Fatalf("expected long content to become two chunks, chunks=%d resp=%#v", len(documents.chunks), resp)
	}
	if documents.chunks[0].DocumentID != "doc_test" || documents.chunks[0].ChunkIndex != 0 {
		t.Fatalf("first chunk not linked/indexed correctly: %#v", documents.chunks[0])
	}
	if documents.atomicWrites != 1 || documents.separateDocumentWrites != 0 || documents.separateChunkWrites != 0 {
		t.Fatalf("document ingestion must use one atomic store call: %#v", documents)
	}
}

func TestAddDocumentDoesNotReportSuccessWhenAtomicStoreFails(t *testing.T) {
	t.Parallel()

	storeErr := errors.New("chunk insert failed")
	documents := &fakeDocumentStore{atomicErr: storeErr}
	service := &Service{documents: documents}

	resp, err := service.AddDocument(context.Background(), &core.AddDocumentRequest{
		TenantID:    "tenant_1",
		WorkspaceID: "workspace_1",
		Source:      "operator_upload",
		Title:       "Runtime Notes",
		Content:     "chunk body",
	})
	if !errors.Is(err, storeErr) {
		t.Fatalf("AddDocument error = %v, want %v", err, storeErr)
	}
	if resp != nil {
		t.Fatalf("AddDocument returned response on failed atomic write: %#v", resp)
	}
	if documents.document != nil || len(documents.chunks) != 0 {
		t.Fatalf("failed atomic write must not be treated as committed, document=%#v chunks=%#v", documents.document, documents.chunks)
	}
}

func TestUpdatePlanDelegatesPatchAndItems(t *testing.T) {
	t.Parallel()

	plans := &fakePlanStore{}
	service := &Service{plans: plans}
	title := "Ship Work Pack 03"
	status := "active"

	resp, err := service.UpdatePlan(context.Background(), &core.UpdatePlanRequest{
		TenantID:    "tenant_1",
		WorkspaceID: "workspace_1",
		PlanID:      "plan_1",
		Title:       &title,
		Status:      &status,
		Items: []core.PlanItemInput{{
			Title: "Wire document API",
		}},
	})
	if err != nil {
		t.Fatalf("UpdatePlan returned error: %v", err)
	}
	if resp.PlanID != "plan_1" || resp.Status != "updated" {
		t.Fatalf("unexpected response: %#v", resp)
	}
	if plans.updatedPlan == nil || plans.updatedPlan.Title != title || plans.updatedPlan.Status != status {
		t.Fatalf("plan update was not delegated: %#v", plans.updatedPlan)
	}
	if len(plans.updatedItems) != 1 || plans.updatedItems[0].Status != "open" {
		t.Fatalf("plan items were not normalized/delegated: %#v", plans.updatedItems)
	}
}

func TestCorrectMemoryValidatesRequiredFields(t *testing.T) {
	t.Parallel()

	service := &Service{
		memories:    &fakeKernelMemoryStore{},
		corrections: &fakeCorrectionStore{},
	}

	_, err := service.CorrectMemory(context.Background(), &core.CorrectMemoryRequest{
		TenantID:       "tenant_1",
		WorkspaceID:    "workspace_1",
		MemoryID:       "mem_1",
		OperatorID:     "operator_1",
		IdempotencyKey: "correction_1",
		CorrectionText: "   ",
	})
	if !errors.Is(err, core.ErrInvalidArgument) {
		t.Fatalf("expected invalid argument for blank correction text, got %v", err)
	}
}

func TestCorrectMemoryReturnsNotFoundForMissingMemory(t *testing.T) {
	t.Parallel()

	service := &Service{
		memories:    &fakeKernelMemoryStore{err: core.ErrNotFound},
		corrections: &fakeCorrectionStore{},
	}

	_, err := service.CorrectMemory(context.Background(), validCorrectionRequest())
	if !errors.Is(err, core.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestCorrectMemoryRecordsRawEventAndCorrection(t *testing.T) {
	t.Parallel()

	corrections := &fakeCorrectionStore{}
	service := &Service{
		memories:    &fakeKernelMemoryStore{memory: validCorrectionTargetMemory()},
		corrections: corrections,
	}

	resp, err := service.CorrectMemory(context.Background(), validCorrectionRequest())
	if err != nil {
		t.Fatalf("CorrectMemory returned error: %v", err)
	}
	if resp.MemoryID != "mem_1" || resp.RawEventID != "evt_correction" || resp.CorrectionID != "corr_1" || !resp.CorrectionRecorded || !resp.TraceWritten || resp.Status != "applied" {
		t.Fatalf("unexpected correction response: %#v", resp)
	}
	if corrections.event == nil || corrections.event.EventKind != "memory_correction" {
		t.Fatalf("raw correction event was not recorded: %#v", corrections.event)
	}
	if corrections.event.Source != "operator_correction" || corrections.event.ActorID != "operator_1" {
		t.Fatalf("raw correction event source/actor mismatch: %#v", corrections.event)
	}
	if corrections.correction == nil || corrections.correction.CorrectionText != "Use the newer fact." {
		t.Fatalf("operator-visible correction artifact was not recorded: %#v", corrections.correction)
	}
	var payload map[string]any
	if err := json.Unmarshal(corrections.event.PayloadJSON, &payload); err != nil {
		t.Fatalf("correction event payload is not JSON: %v", err)
	}
	if payload["memory_id"] != "mem_1" || payload["correction_text"] != "Use the newer fact." {
		t.Fatalf("correction payload lost intent: %#v", payload)
	}
	if corrections.correction == nil || corrections.correction.Status != "recorded" {
		t.Fatalf("correction artifact should be recorded before supersession: %#v", corrections.correction)
	}
	memories := service.memories.(*fakeKernelMemoryStore)
	if memories.updateMemory == nil || memories.updateTrace == nil || memories.updateEdge == nil {
		t.Fatalf("correction did not apply graph supersession: memory=%#v trace=%#v edge=%#v", memories.updateMemory, memories.updateTrace, memories.updateEdge)
	}
	if memories.updateMemory.Text != "Use the newer fact." || memories.updateMemory.Scope != core.MemoryScopeWorkspaceShared || memories.updateMemory.OwnerEntityID != "agent:hermes-main" {
		t.Fatalf("replacement memory did not preserve target boundary and corrected text: %#v", memories.updateMemory)
	}
	if memories.updateTrace.RawEventIDs[0] != "evt_correction" || !memories.updateTrace.OperatorCorrectionFlag || memories.updateTrace.ReasoningStage != "operator_correction" {
		t.Fatalf("correction trace did not preserve operator provenance: %#v", memories.updateTrace)
	}
	if memories.updateEdge.FromMemoryID != memories.updateMemory.ID || memories.updateEdge.ToMemoryID != "mem_1" || memories.updateEdge.EdgeKind != core.EdgeKindUpdates {
		t.Fatalf("correction updates edge mismatch: %#v", memories.updateEdge)
	}
}

func TestCorrectMemoryIdempotentRetryReturnsRecordedArtifact(t *testing.T) {
	t.Parallel()

	corrections := &fakeCorrectionStore{
		recorded: &core.MemoryCorrection{
			ID:             "corr_existing",
			MemoryID:       "mem_1",
			RawEventID:     "evt_existing",
			IdempotencyKey: "correction_1",
			CorrectionText: "Use the newer fact.",
			OperatorID:     "operator_1",
			Status:         "recorded",
		},
	}
	service := &Service{
		memories:    &fakeKernelMemoryStore{memory: validCorrectionTargetMemory()},
		corrections: corrections,
	}

	resp, err := service.CorrectMemory(context.Background(), validCorrectionRequest())
	if err != nil {
		t.Fatalf("CorrectMemory retry returned error: %v", err)
	}
	if resp.RawEventID != "evt_existing" || resp.CorrectionID != "corr_existing" || resp.Status != "applied" || !resp.TraceWritten {
		t.Fatalf("idempotent retry did not return existing correction artifact: %#v", resp)
	}
}

func TestCorrectMemoryRejectsNonLatestTargetBeforeRecordingCorrection(t *testing.T) {
	t.Parallel()

	target := validCorrectionTargetMemory()
	target.Status = core.MemoryStatusSuperseded
	target.LatestFlag = false
	corrections := &fakeCorrectionStore{}
	service := &Service{
		memories:    &fakeKernelMemoryStore{memory: target},
		corrections: corrections,
	}

	_, err := service.CorrectMemory(context.Background(), validCorrectionRequest())
	if !errors.Is(err, core.ErrConflict) {
		t.Fatalf("expected ErrConflict for non-latest correction target, got %v", err)
	}
	if corrections.event != nil || corrections.correction != nil {
		t.Fatalf("non-latest target must not record correction side effects: event=%#v correction=%#v", corrections.event, corrections.correction)
	}
}

func TestCorrectMemoryDoesNotReportSuccessWhenSupersessionFails(t *testing.T) {
	t.Parallel()

	storeErr := errors.New("supersession failed")
	memories := &fakeKernelMemoryStore{memory: validCorrectionTargetMemory(), updateErr: storeErr}
	service := &Service{
		memories:    memories,
		corrections: &fakeCorrectionStore{},
	}

	resp, err := service.CorrectMemory(context.Background(), validCorrectionRequest())
	if !errors.Is(err, storeErr) {
		t.Fatalf("CorrectMemory error = %v, want %v", err, storeErr)
	}
	if resp != nil {
		t.Fatalf("failed supersession must not return success: %#v", resp)
	}
}

func TestGetTimelineDefaultsScopesLimitAndDelegates(t *testing.T) {
	t.Parallel()

	timeline := &fakeTimelineStore{}
	service := &Service{timeline: timeline}

	resp, err := service.GetTimeline(context.Background(), &core.GetTimelineRequest{
		TenantID:    "tenant_1",
		WorkspaceID: "workspace_1",
		EntityID:    "agent:hermes-main",
	})
	if err != nil {
		t.Fatalf("GetTimeline returned error: %v", err)
	}
	if resp == nil || len(resp.Items) != 1 || resp.Items[0].ID != "tl_1" {
		t.Fatalf("unexpected timeline response: %#v", resp)
	}
	if timeline.req == nil || timeline.req.Limit != timelineDefaultLimit {
		t.Fatalf("timeline request was not defaulted: %#v", timeline.req)
	}
	wantScopes := []core.MemoryScope{
		core.MemoryScopeAgentPrivate,
		core.MemoryScopeWorkspaceShared,
		core.MemoryScopeSessionScratch,
	}
	if !sameMemoryScopes(timeline.req.Scopes, wantScopes) {
		t.Fatalf("timeline scopes = %#v, want %#v", timeline.req.Scopes, wantScopes)
	}
}

func TestGetTimelineRejectsInvalidScopeAndLimit(t *testing.T) {
	t.Parallel()

	service := &Service{timeline: &fakeTimelineStore{}}

	_, err := service.GetTimeline(context.Background(), &core.GetTimelineRequest{
		TenantID:    "tenant_1",
		WorkspaceID: "workspace_1",
		EntityID:    "agent:hermes-main",
		Scopes:      []core.MemoryScope{"public"},
	})
	if !errors.Is(err, core.ErrInvalidArgument) {
		t.Fatalf("expected invalid scope error, got %v", err)
	}

	_, err = service.GetTimeline(context.Background(), &core.GetTimelineRequest{
		TenantID:    "tenant_1",
		WorkspaceID: "workspace_1",
		EntityID:    "agent:hermes-main",
		Limit:       timelineMaxLimit + 1,
	})
	if !errors.Is(err, core.ErrInvalidArgument) {
		t.Fatalf("expected invalid limit error, got %v", err)
	}
}

func TestGetTimelineExcludesGroupSharedUntilMembershipFiltering(t *testing.T) {
	t.Parallel()

	timeline := &fakeTimelineStore{}
	service := &Service{timeline: timeline}

	_, err := service.GetTimeline(context.Background(), &core.GetTimelineRequest{
		TenantID:    "tenant_1",
		WorkspaceID: "workspace_1",
		EntityID:    "agent:hermes-main",
		Scopes: []core.MemoryScope{
			core.MemoryScopeGroupShared,
			core.MemoryScopeWorkspaceShared,
		},
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("GetTimeline returned error: %v", err)
	}
	if !sameMemoryScopes(timeline.req.Scopes, []core.MemoryScope{core.MemoryScopeWorkspaceShared}) {
		t.Fatalf("group_shared should be excluded from timeline scopes, got %#v", timeline.req.Scopes)
	}
}

func TestExplainMemoryDelegatesVisibilityFields(t *testing.T) {
	t.Parallel()

	memories := &fakeKernelMemoryStore{}
	service := &Service{memories: memories}

	_, err := service.ExplainMemory(context.Background(), &core.ExplainMemoryRequest{
		TenantID:        "tenant_1",
		WorkspaceID:     "workspace_1",
		MemoryID:        "mem_1",
		EntityID:        "agent:hermes-main",
		VisibleGroupIDs: []string{"group_design"},
	})
	if err != nil {
		t.Fatalf("ExplainMemory returned error: %v", err)
	}
	if memories.explainReq == nil || memories.explainReq.EntityID != "agent:hermes-main" {
		t.Fatalf("explain visibility fields were not delegated: %#v", memories.explainReq)
	}
	if got := memories.explainReq.VisibleGroupIDs; len(got) != 1 || got[0] != "group_design" {
		t.Fatalf("visible group ids were not delegated: %#v", memories.explainReq)
	}
}

func validCorrectionRequest() *core.CorrectMemoryRequest {
	return &core.CorrectMemoryRequest{
		TenantID:       "tenant_1",
		WorkspaceID:    "workspace_1",
		MemoryID:       "mem_1",
		OperatorID:     "operator_1",
		IdempotencyKey: "correction_1",
		CorrectionText: "Use the newer fact.",
	}
}

func validCorrectionTargetMemory() *core.Memory {
	return &core.Memory{
		ID:            "mem_1",
		TenantID:      "tenant_1",
		WorkspaceID:   "workspace_1",
		Scope:         core.MemoryScopeWorkspaceShared,
		OwnerEntityID: "agent:hermes-main",
		Kind:          core.MemoryKindFact,
		ArtifactClass: core.ArtifactClassKnowledge,
		Text:          "Old fact.",
		Confidence:    0.7,
		Status:        core.MemoryStatusActive,
		LatestFlag:    true,
	}
}

func sameMemoryScopes(got, want []core.MemoryScope) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

type fakeDocumentStore struct {
	document               *core.Document
	chunks                 []*core.DocumentChunk
	atomicErr              error
	atomicWrites           int
	separateDocumentWrites int
	separateChunkWrites    int
}

func (s *fakeDocumentStore) AddDocumentWithChunks(_ context.Context, document *core.Document, chunks []*core.DocumentChunk) error {
	s.atomicWrites++
	if s.atomicErr != nil {
		return s.atomicErr
	}
	document.ID = "doc_test"
	for i, chunk := range chunks {
		chunk.DocumentID = document.ID
		chunk.ID = "chunk_test_" + string(rune('a'+i))
	}
	s.document = document
	s.chunks = chunks
	return nil
}

func (s *fakeDocumentStore) AddDocument(_ context.Context, document *core.Document) error {
	s.separateDocumentWrites++
	document.ID = "doc_test"
	s.document = document
	return nil
}

func (s *fakeDocumentStore) AddDocumentChunks(_ context.Context, chunks []*core.DocumentChunk) error {
	s.separateChunkWrites++
	for i, chunk := range chunks {
		chunk.ID = "chunk_test_" + string(rune('a'+i))
	}
	s.chunks = chunks
	return nil
}

func (s *fakeDocumentStore) SearchDocuments(context.Context, *core.SearchDocumentsRequest) (*core.SearchDocumentsResponse, error) {
	return nil, core.ErrNotImplemented
}

type fakePlanStore struct {
	updatedPlan  *core.Plan
	updatedItems []*core.PlanItem
}

func (s *fakePlanStore) CreatePlan(context.Context, *core.Plan, []*core.PlanItem) error {
	return core.ErrNotImplemented
}

func (s *fakePlanStore) UpdatePlan(_ context.Context, plan *core.Plan, items []*core.PlanItem) error {
	s.updatedPlan = plan
	s.updatedItems = items
	return nil
}

func (s *fakePlanStore) GetActivePlans(context.Context, *core.GetActivePlansRequest) ([]*core.Plan, error) {
	return nil, core.ErrNotImplemented
}

type fakeKernelMemoryStore struct {
	memory       *core.Memory
	updateMemory *core.Memory
	updateTrace  *core.MemoryTrace
	updateEdge   *core.MemoryEdge
	explainReq   *core.ExplainMemoryRequest
	err          error
	updateErr    error
}

func (s *fakeKernelMemoryStore) UpsertMemory(context.Context, *core.Memory) error {
	return core.ErrNotImplemented
}

func (s *fakeKernelMemoryStore) GetMemory(context.Context, string) (*core.Memory, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.memory, nil
}

func (s *fakeKernelMemoryStore) SearchMemories(context.Context, *core.SearchMemoriesRequest) (*core.SearchMemoriesResponse, error) {
	return nil, core.ErrNotImplemented
}

func (s *fakeKernelMemoryStore) UpsertMemoryEdge(context.Context, *core.MemoryEdge) error {
	return core.ErrNotImplemented
}

func (s *fakeKernelMemoryStore) WriteMemoryTrace(context.Context, *core.MemoryTrace) error {
	return core.ErrNotImplemented
}

func (s *fakeKernelMemoryStore) ExplainMemory(_ context.Context, req *core.ExplainMemoryRequest) (*core.ExplainMemoryResponse, error) {
	s.explainReq = req
	return &core.ExplainMemoryResponse{MemoryID: req.MemoryID}, nil
}

func (s *fakeKernelMemoryStore) CreateMemoryWithTraceAndUpdateEdge(_ context.Context, memory *core.Memory, trace *core.MemoryTrace, edge *core.MemoryEdge) error {
	if s.updateErr != nil {
		return s.updateErr
	}
	s.updateMemory = memory
	s.updateTrace = trace
	s.updateEdge = edge
	return nil
}

type fakeCorrectionStore struct {
	event      *core.RawEvent
	correction *core.MemoryCorrection
	recorded   *core.MemoryCorrection
}

type fakeTimelineStore struct {
	req *core.GetTimelineRequest
}

func (s *fakeTimelineStore) GetTimeline(_ context.Context, req *core.GetTimelineRequest) (*core.GetTimelineResponse, error) {
	reqCopy := *req
	reqCopy.Scopes = append([]core.MemoryScope(nil), req.Scopes...)
	s.req = &reqCopy
	return &core.GetTimelineResponse{
		Items: []core.TimelineItem{{
			ID:            "tl_1",
			Kind:          core.MemoryKindCorrection,
			ArtifactClass: core.ArtifactClassTimeline,
			Text:          "Correction for memory mem_1: Use the newer fact.",
			MemoryID:      "mem_1",
			RawEventID:    "evt_1",
		}},
	}, nil
}

func (s *fakeCorrectionStore) RecordMemoryCorrection(_ context.Context, event *core.RawEvent, correction *core.MemoryCorrection) (*core.MemoryCorrection, error) {
	s.event = event
	s.correction = correction
	if s.recorded != nil {
		return s.recorded, nil
	}
	correction.ID = "corr_1"
	correction.RawEventID = "evt_correction"
	correction.Status = "recorded"
	return correction, nil
}

```



<!-- Source: internal/mcp/protocol_test.go | bytes=7232 | lines=180 | sha16=c284eed301d1b8bd -->

```go
// ============================================================
// FILE     : internal/mcp/protocol_test.go
// PURPOSE  : Verifies MCP JSON-RPC lifecycle, tool listing, and tool call roundtrips.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : MCP protocol server tests
// DEPENDS  : internal/mcp/protocol.go, internal/mcp/surface_test.go
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: Protocol tests should prove real JSON-RPC shape, not only adapter delegation.
// ============================================================

package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

func TestServerHandlesInitializeAndToolRoundtrip(t *testing.T) {
	t.Parallel()

	surface := newTestSurface(t, &fakeService{
		prefetchResp: &core.PrefetchResponse{
			Blocks: []core.RecallBlock{{Kind: "pinned_note", Priority: 100, Text: "Keep Hermes first."}},
			Meta:   core.RecallMeta{EstimatedTokens: 4, Sources: []string{"notes"}},
		},
	})
	server := newProtocolServer(t, surface)

	initResp, respond := server.HandleMessage(context.Background(), json.RawMessage(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"test","version":"0"}}}`))
	if !respond {
		t.Fatalf("initialize did not produce a response")
	}
	var initEnvelope map[string]any
	decodeJSONMessage(t, initResp, &initEnvelope)
	result := initEnvelope["result"].(map[string]any)
	if result["protocolVersion"] != ProtocolVersion {
		t.Fatalf("unexpected protocol version: %#v", result["protocolVersion"])
	}
	if _, ok := result["capabilities"].(map[string]any)["tools"]; !ok {
		t.Fatalf("initialize response did not advertise tools: %s", string(initResp))
	}

	listResp, respond := server.HandleMessage(context.Background(), json.RawMessage(`{"jsonrpc":"2.0","id":"tools","method":"tools/list"}`))
	if !respond {
		t.Fatalf("tools/list did not produce a response")
	}
	if !strings.Contains(string(listResp), `"prefetch"`) || !strings.Contains(string(listResp), `"inputSchema"`) {
		t.Fatalf("tools/list missing expected tool schema: %s", string(listResp))
	}
	if !strings.Contains(string(listResp), `"required":["tenant_id","workspace_id"]`) || !strings.Contains(string(listResp), `"recall_preview"`) {
		t.Fatalf("tools/list did not expose recall preview required inputs: %s", string(listResp))
	}

	callResp, respond := server.HandleMessage(context.Background(), json.RawMessage(`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"prefetch","arguments":{"tenant_id":"tenant_1","workspace_id":"workspace_1","session_id":"session_1","actor_id":"agent:hermes-main"}}}`))
	if !respond {
		t.Fatalf("tools/call did not produce a response")
	}
	if !strings.Contains(string(callResp), `"structuredContent"`) || !strings.Contains(string(callResp), "Keep Hermes first.") {
		t.Fatalf("tools/call did not return structured prefetch output: %s", string(callResp))
	}
}

func TestServerToolSchemasExposeTrustLoopInputs(t *testing.T) {
	t.Parallel()

	server := newProtocolServer(t, newTestSurface(t, &fakeService{}))
	tools := server.protocolTools()

	byName := make(map[string]protocolTool, len(tools))
	for _, tool := range tools {
		byName[tool.Name] = tool
	}
	assertRequiredInputs(t, byName["recall_preview"], "tenant_id", "workspace_id")
	assertRequiredInputs(t, byName["correct_memory"], "tenant_id", "workspace_id", "memory_id", "operator_id", "correction_text")
	assertRequiredInputs(t, byName["view_timeline"], "tenant_id", "workspace_id", "entity_id")
	assertRequiredInputs(t, byName["explain_memory"], "tenant_id", "workspace_id", "memory_id")

	correctionProps := byName["correct_memory"].InputSchema["properties"].(map[string]any)
	if _, ok := correctionProps["evidence_json"]; !ok {
		t.Fatalf("correct_memory schema should expose evidence_json for provenance")
	}
	timelineProps := byName["view_timeline"].InputSchema["properties"].(map[string]any)
	if _, ok := timelineProps["scopes"]; !ok {
		t.Fatalf("view_timeline schema should expose scopes for visibility review")
	}
	explainProps := byName["explain_memory"].InputSchema["properties"].(map[string]any)
	if _, ok := explainProps["entity_id"]; !ok {
		t.Fatalf("explain_memory schema should expose entity_id for private visibility")
	}
	if _, ok := explainProps["visible_group_ids"]; !ok {
		t.Fatalf("explain_memory schema should expose visible_group_ids for group visibility")
	}
}

func TestServerServeStdioRoundtrip(t *testing.T) {
	t.Parallel()

	surface := newTestSurface(t, &fakeService{prefetchResp: &core.PrefetchResponse{Blocks: []core.RecallBlock{{Kind: "note", Text: "stdio ok"}}}})
	server := newProtocolServer(t, surface)
	input := strings.Join([]string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"test","version":"0"}}}`,
		`{"jsonrpc":"2.0","method":"notifications/initialized"}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"prefetch","arguments":{"tenant_id":"tenant_1","workspace_id":"workspace_1"}}}`,
		"",
	}, "\n")
	var out bytes.Buffer

	if err := server.ServeStdio(context.Background(), strings.NewReader(input), &out); err != nil {
		t.Fatalf("ServeStdio returned error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected two JSON-RPC responses, got %d: %q", len(lines), out.String())
	}
	if !strings.Contains(lines[0], `"protocolVersion":"2025-11-25"`) {
		t.Fatalf("first response was not initialize: %s", lines[0])
	}
	if !strings.Contains(lines[1], "stdio ok") {
		t.Fatalf("second response was not tool output: %s", lines[1])
	}
}

func TestServerReturnsProtocolErrorForUnknownMethod(t *testing.T) {
	t.Parallel()

	server := newProtocolServer(t, newTestSurface(t, &fakeService{}))
	raw, respond := server.HandleMessage(context.Background(), json.RawMessage(`{"jsonrpc":"2.0","id":1,"method":"missing"}`))
	if !respond {
		t.Fatalf("unknown method did not produce a response")
	}
	if !strings.Contains(string(raw), `"code":-32601`) {
		t.Fatalf("expected method-not-found error, got %s", string(raw))
	}
}

func newProtocolServer(t *testing.T, surface *Surface) *Server {
	t.Helper()

	server, err := NewServer(surface)
	if err != nil {
		t.Fatalf("NewServer returned error: %v", err)
	}
	return server
}

func decodeJSONMessage(t *testing.T, raw json.RawMessage, out any) {
	t.Helper()

	if err := json.Unmarshal(raw, out); err != nil {
		t.Fatalf("decode JSON message: %v; raw=%s", err, string(raw))
	}
}

func assertRequiredInputs(t *testing.T, tool protocolTool, want ...string) {
	t.Helper()

	required, ok := tool.InputSchema["required"].([]string)
	if !ok {
		t.Fatalf("%s schema missing required input list: %#v", tool.Name, tool.InputSchema)
	}
	got := make(map[string]bool, len(required))
	for _, field := range required {
		got[field] = true
	}
	for _, field := range want {
		if !got[field] {
			t.Fatalf("%s schema missing required field %q in %#v", tool.Name, field, required)
		}
	}
}

```



<!-- Source: internal/mcp/surface_test.go | bytes=8166 | lines=220 | sha16=9bcfbc5d30cdf32b -->

```go
// ============================================================
// FILE     : internal/mcp/surface_test.go
// PURPOSE  : Verifies the MCP-style tool surface delegates to shared core semantics.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : MCP surface adapter tests
// DEPENDS  : context, encoding/json, errors, strings, testing, internal/core
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: These tests do not start a protocol server; they lock tool-to-core mapping.
// ============================================================

package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

func TestSurfaceListsV1Tools(t *testing.T) {
	t.Parallel()

	surface := newTestSurface(t, &fakeService{})
	tools := surface.Tools()
	names := make(map[string]bool, len(tools))
	for _, tool := range tools {
		names[tool.Name] = true
	}
	for _, want := range []string{"prefetch", "recall_preview", "sync_turn", "search_memory", "add_note", "correct_memory", "view_timeline", "explain_memory"} {
		if !names[want] {
			t.Fatalf("expected tool %q in %#v", want, tools)
		}
	}
}

func TestSurfaceCallsRecallPreviewAlias(t *testing.T) {
	t.Parallel()

	service := &fakeService{prefetchResp: &core.PrefetchResponse{Blocks: []core.RecallBlock{{Kind: "memory", Priority: 90, Text: "Preview"}}}}
	surface := newTestSurface(t, service)

	raw, err := surface.Call(context.Background(), "recall_preview", json.RawMessage(`{"tenant_id":"tenant_1","workspace_id":"workspace_1"}`))
	if err != nil {
		t.Fatalf("Call returned error: %v", err)
	}
	if service.prefetchCalls != 1 {
		t.Fatalf("expected recall_preview to delegate to prefetch, got %d calls", service.prefetchCalls)
	}
	if !strings.Contains(string(raw), "Preview") {
		t.Fatalf("expected encoded recall preview output, got %s", string(raw))
	}
}

func TestSurfaceCallsPrefetch(t *testing.T) {
	t.Parallel()

	service := &fakeService{prefetchResp: &core.PrefetchResponse{Blocks: []core.RecallBlock{{Kind: "note", Priority: 100, Text: "Pinned"}}}}
	surface := newTestSurface(t, service)

	raw, err := surface.Call(context.Background(), "prefetch", json.RawMessage(`{"tenant_id":"tenant_1","workspace_id":"workspace_1"}`))
	if err != nil {
		t.Fatalf("Call returned error: %v", err)
	}
	if service.prefetchCalls != 1 {
		t.Fatalf("expected one prefetch call, got %d", service.prefetchCalls)
	}
	if !strings.Contains(string(raw), "Pinned") {
		t.Fatalf("expected encoded prefetch output, got %s", string(raw))
	}
}

func TestSurfaceCallsCorrectMemory(t *testing.T) {
	t.Parallel()

	service := &fakeService{correctResp: &core.CorrectMemoryResponse{MemoryID: "mem_1", CorrectionRecorded: true, Status: "recorded"}}
	surface := newTestSurface(t, service)

	raw, err := surface.Call(context.Background(), "correct_memory", json.RawMessage(`{"tenant_id":"tenant_1","workspace_id":"workspace_1","memory_id":"mem_1"}`))
	if err != nil {
		t.Fatalf("Call returned error: %v", err)
	}
	if service.correctCalls != 1 {
		t.Fatalf("expected one correction call, got %d", service.correctCalls)
	}
	if !strings.Contains(string(raw), `"correction_recorded":true`) {
		t.Fatalf("expected encoded correction output, got %s", string(raw))
	}
}

func TestSurfaceCallsViewTimeline(t *testing.T) {
	t.Parallel()

	service := &fakeService{timelineResp: &core.GetTimelineResponse{Items: []core.TimelineItem{{ID: "item_1", MemoryID: "mem_1", Text: "Corrected project rule"}}}}
	surface := newTestSurface(t, service)

	raw, err := surface.Call(context.Background(), "view_timeline", json.RawMessage(`{"tenant_id":"tenant_1","workspace_id":"workspace_1","entity_id":"agent:hermes-main"}`))
	if err != nil {
		t.Fatalf("Call returned error: %v", err)
	}
	if service.timelineCalls != 1 {
		t.Fatalf("expected one timeline call, got %d", service.timelineCalls)
	}
	if !strings.Contains(string(raw), "Corrected project rule") {
		t.Fatalf("expected encoded timeline output, got %s", string(raw))
	}
}

func TestSurfaceCallsExplainMemory(t *testing.T) {
	t.Parallel()

	service := &fakeService{explainResp: &core.ExplainMemoryResponse{MemoryID: "mem_1", Trace: core.MemoryTraceResult{ReasoningJobID: "job_1"}}}
	surface := newTestSurface(t, service)

	raw, err := surface.Call(context.Background(), "explain_memory", json.RawMessage(`{"tenant_id":"tenant_1","workspace_id":"workspace_1","memory_id":"mem_1","entity_id":"agent:hermes-main","visible_group_ids":["group_design"]}`))
	if err != nil {
		t.Fatalf("Call returned error: %v", err)
	}
	if service.explainCalls != 1 {
		t.Fatalf("expected one explain call, got %d", service.explainCalls)
	}
	if service.explainReq == nil || service.explainReq.EntityID != "agent:hermes-main" {
		t.Fatalf("expected explain visibility request, got %#v", service.explainReq)
	}
	if got := service.explainReq.VisibleGroupIDs; len(got) != 1 || got[0] != "group_design" {
		t.Fatalf("expected visible group ids, got %#v", service.explainReq)
	}
	if !strings.Contains(string(raw), `"reasoning_job_id":"job_1"`) {
		t.Fatalf("expected encoded explain output, got %s", string(raw))
	}
}

func TestSurfaceRejectsUnknownToolAndInvalidJSON(t *testing.T) {
	t.Parallel()

	surface := newTestSurface(t, &fakeService{})
	if _, err := surface.Call(context.Background(), "unknown", json.RawMessage(`{}`)); !errors.Is(err, core.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument for unknown tool, got %v", err)
	}
	if _, err := surface.Call(context.Background(), "prefetch", json.RawMessage(`{`)); !errors.Is(err, core.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument for invalid JSON, got %v", err)
	}
}

func newTestSurface(t *testing.T, service core.VibeGravityService) *Surface {
	t.Helper()

	surface, err := NewSurface(service)
	if err != nil {
		t.Fatalf("NewSurface returned error: %v", err)
	}
	return surface
}

type fakeService struct {
	prefetchCalls int
	correctCalls  int
	timelineCalls int
	explainCalls  int
	prefetchResp  *core.PrefetchResponse
	correctResp   *core.CorrectMemoryResponse
	timelineResp  *core.GetTimelineResponse
	explainResp   *core.ExplainMemoryResponse
	explainReq    *core.ExplainMemoryRequest
}

func (s *fakeService) Prefetch(context.Context, *core.PrefetchRequest) (*core.PrefetchResponse, error) {
	s.prefetchCalls++
	return s.prefetchResp, nil
}

func (s *fakeService) SyncTurn(context.Context, *core.SyncTurnRequest) (*core.SyncTurnResponse, error) {
	return &core.SyncTurnResponse{Status: "accepted"}, nil
}

func (s *fakeService) AddDocument(context.Context, *core.AddDocumentRequest) (*core.AddDocumentResponse, error) {
	return &core.AddDocumentResponse{Status: "created"}, nil
}

func (s *fakeService) SearchMemories(context.Context, *core.SearchMemoriesRequest) (*core.SearchMemoriesResponse, error) {
	return &core.SearchMemoriesResponse{}, nil
}

func (s *fakeService) SearchDocuments(context.Context, *core.SearchDocumentsRequest) (*core.SearchDocumentsResponse, error) {
	return &core.SearchDocumentsResponse{}, nil
}

func (s *fakeService) AddNote(context.Context, *core.AddNoteRequest) (*core.AddNoteResponse, error) {
	return &core.AddNoteResponse{Status: "created"}, nil
}

func (s *fakeService) CreatePlan(context.Context, *core.CreatePlanRequest) (*core.CreatePlanResponse, error) {
	return &core.CreatePlanResponse{Status: "created"}, nil
}

func (s *fakeService) UpdatePlan(context.Context, *core.UpdatePlanRequest) (*core.UpdatePlanResponse, error) {
	return &core.UpdatePlanResponse{Status: "updated"}, nil
}

func (s *fakeService) CorrectMemory(context.Context, *core.CorrectMemoryRequest) (*core.CorrectMemoryResponse, error) {
	s.correctCalls++
	return s.correctResp, nil
}

func (s *fakeService) GetTimeline(context.Context, *core.GetTimelineRequest) (*core.GetTimelineResponse, error) {
	s.timelineCalls++
	return s.timelineResp, nil
}

func (s *fakeService) ExplainMemory(_ context.Context, req *core.ExplainMemoryRequest) (*core.ExplainMemoryResponse, error) {
	s.explainCalls++
	s.explainReq = req
	return s.explainResp, nil
}

```



<!-- Source: internal/reasoning/codex_bridge_test.go | bytes=8015 | lines=262 | sha16=36dd4168198b0b7c -->

```go
// ============================================================
// FILE     : internal/reasoning/codex_bridge_test.go
// PURPOSE  : Guards the disabled Codex bridge boundary and strict JSON decoding.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : codex bridge tests
// DEPENDS  : context, encoding/json, errors, strings, testing, internal/core
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: These tests use a fake client only; do not call real Codex here.
// ============================================================

package reasoning

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

func TestCodexStage1ExtractorDisabledByDefault(t *testing.T) {
	t.Parallel()

	extractor, err := NewCodexStage1Extractor(CodexBridgeConfig{}, nil)
	if err != nil {
		t.Fatalf("NewCodexStage1Extractor returned error: %v", err)
	}

	output, err := extractor.Extract(context.Background(), testProcessTurnEnvelope().Stage1)
	if err == nil {
		t.Fatalf("expected disabled Codex Stage 1 to fail")
	}
	if !errors.Is(err, core.ErrNotImplemented) {
		t.Fatalf("expected ErrNotImplemented, got %v", err)
	}
	if output.CandidateEntities != nil || output.CandidateMemories != nil {
		t.Fatalf("expected no Stage 1 output when disabled, got %#v", output)
	}
}

func TestCodexStage1ExtractorRejectsUnknownOutputFields(t *testing.T) {
	t.Parallel()

	extractor, err := NewCodexStage1Extractor(CodexBridgeConfig{Enabled: true}, fakeCodexJSONClient{
		resp: CodexResponse{OutputJSON: json.RawMessage(`{
			"candidate_entities": [],
			"candidate_memories": [],
			"summary_hint": "",
			"task_hint": "",
			"freeform": "not allowed"
		}`)},
	})
	if err != nil {
		t.Fatalf("NewCodexStage1Extractor returned error: %v", err)
	}

	_, err = extractor.Extract(context.Background(), testProcessTurnEnvelope().Stage1)
	if err == nil {
		t.Fatalf("expected unknown Stage 1 field to be rejected")
	}
	if !errors.Is(err, core.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
	if !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("expected strict JSON unknown-field error, got %v", err)
	}
}

func TestCodexStage1ExtractorSendsSchemaMarkedRequest(t *testing.T) {
	t.Parallel()

	client := &capturingCodexJSONClient{
		resp: CodexResponse{OutputJSON: json.RawMessage(`{
			"candidate_entities": [{
				"entity_kind": "agent",
				"display_name": "Hermes",
				"confidence": 0.91,
				"metadata_json": {},
				"source_event_id": "evt_contract_1"
			}],
			"candidate_memories": [{
				"kind": "constraint",
				"artifact_class": "knowledge",
				"scope": "workspace_shared",
				"text": "Reasoning output must stay structured.",
				"confidence": 0.92,
				"raw_event_ids": ["evt_contract_1"]
			}],
			"summary_hint": "structured",
			"task_hint": "resolve graph operations"
		}`)},
	}
	extractor, err := NewCodexStage1Extractor(CodexBridgeConfig{Enabled: true}, client)
	if err != nil {
		t.Fatalf("NewCodexStage1Extractor returned error: %v", err)
	}

	output, err := extractor.Extract(context.Background(), testProcessTurnEnvelope().Stage1)
	if err != nil {
		t.Fatalf("Extract returned error: %v", err)
	}
	if client.req.Stage != StageNameExtract {
		t.Fatalf("expected extract stage request, got %#v", client.req)
	}
	if client.req.RequiredOutputSchema != Stage1ExtractOutputSchemaV0 {
		t.Fatalf("expected Stage 1 schema marker, got %q", client.req.RequiredOutputSchema)
	}
	if len(output.CandidateMemories) != 1 {
		t.Fatalf("expected decoded candidate memory, got %#v", output)
	}
}

func TestCodexStage2ResolverPreservesRequiredOutputSchema(t *testing.T) {
	t.Parallel()

	client := &capturingCodexJSONClient{
		resp: CodexResponse{OutputJSON: json.RawMessage(`{
			"operations": [],
			"profile_delta": {},
			"session_summary": "",
			"plan_delta": {},
			"trace": {
				"schema_version": "v0",
				"stage": "resolve",
				"codes": ["test_codex_stage2"],
				"metadata_json": {}
			}
		}`)},
	}
	resolver, err := NewCodexStage2Resolver(CodexBridgeConfig{Enabled: true}, client)
	if err != nil {
		t.Fatalf("NewCodexStage2Resolver returned error: %v", err)
	}

	output, err := resolver.Resolve(context.Background(), testProcessTurnEnvelope().Stage2)
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	if client.req.Stage != StageNameResolve {
		t.Fatalf("expected resolve stage request, got %#v", client.req)
	}
	if client.req.RequiredOutputSchema != Stage2ResolveOutputSchemaV0 {
		t.Fatalf("expected Stage 2 schema marker to be preserved, got %q", client.req.RequiredOutputSchema)
	}
	if output.Trace.Stage != StageNameResolve {
		t.Fatalf("expected resolved trace output, got %#v", output.Trace)
	}
}

func TestCodexStage2ResolverRejectsInvalidJSONObjects(t *testing.T) {
	t.Parallel()

	resolver, err := NewCodexStage2Resolver(CodexBridgeConfig{Enabled: true}, fakeCodexJSONClient{
		resp: CodexResponse{OutputJSON: json.RawMessage(`{
			"operations": [],
			"profile_delta": [],
			"session_summary": "",
			"plan_delta": {},
			"trace": {
				"schema_version": "v0",
				"stage": "resolve",
				"codes": [],
				"metadata_json": {}
			}
		}`)},
	})
	if err != nil {
		t.Fatalf("NewCodexStage2Resolver returned error: %v", err)
	}

	_, err = resolver.Resolve(context.Background(), testProcessTurnEnvelope().Stage2)
	if err == nil {
		t.Fatalf("expected non-object profile_delta to be rejected")
	}
	if !errors.Is(err, core.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
	if !strings.Contains(err.Error(), "profile_delta") {
		t.Fatalf("expected profile_delta error, got %v", err)
	}
}

func TestCodexStage2ResolverRejectsMismatchedRequiredOutputSchema(t *testing.T) {
	t.Parallel()

	resolver, err := NewCodexStage2Resolver(CodexBridgeConfig{Enabled: true}, fakeCodexJSONClient{})
	if err != nil {
		t.Fatalf("NewCodexStage2Resolver returned error: %v", err)
	}
	input := testProcessTurnEnvelope().Stage2
	input.RequiredOutputSchema = "wrong.schema"

	_, err = resolver.Resolve(context.Background(), input)
	if err == nil {
		t.Fatalf("expected mismatched Stage 2 schema marker to be rejected")
	}
	if !errors.Is(err, core.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}

func TestMockCodexJSONClientRunsThroughBridgeRunners(t *testing.T) {
	t.Parallel()

	mockClient := NewMockCodexJSONClient()
	cfg := CodexBridgeConfig{Enabled: true}
	stage1, err := NewCodexStage1Extractor(cfg, mockClient)
	if err != nil {
		t.Fatalf("NewCodexStage1Extractor returned error: %v", err)
	}
	stage2, err := NewCodexStage2Resolver(cfg, mockClient)
	if err != nil {
		t.Fatalf("NewCodexStage2Resolver returned error: %v", err)
	}
	orchestrator, err := NewPipelineOrchestrator(stage1, stage2, nil)
	if err != nil {
		t.Fatalf("NewPipelineOrchestrator returned error: %v", err)
	}

	result, err := orchestrator.ProcessTurn(context.Background(), testProcessTurnEnvelope())
	if err != nil {
		t.Fatalf("ProcessTurn returned error: %v", err)
	}
	if result.Stage1.SummaryHint != "mock_codex_stage1_no_candidates" {
		t.Fatalf("expected mocked Stage 1 output, got %#v", result.Stage1)
	}
	if len(result.Stage2.Trace.Codes) != 1 || result.Stage2.Trace.Codes[0] != "mock_codex_bridge_no_operations" {
		t.Fatalf("expected mocked Stage 2 bridge trace, got %#v", result.Stage2.Trace)
	}
}

type fakeCodexJSONClient struct {
	resp CodexResponse
	err  error
}

func (c fakeCodexJSONClient) CompleteJSON(context.Context, CodexRequest) (CodexResponse, error) {
	if c.err != nil {
		return CodexResponse{}, c.err
	}
	return c.resp, nil
}

type capturingCodexJSONClient struct {
	req  CodexRequest
	resp CodexResponse
	err  error
}

func (c *capturingCodexJSONClient) CompleteJSON(_ context.Context, req CodexRequest) (CodexResponse, error) {
	c.req = req
	if c.err != nil {
		return CodexResponse{}, c.err
	}
	return c.resp, nil
}

```



<!-- Source: internal/reasoning/orchestrator_test.go | bytes=7724 | lines=242 | sha16=820b735f1b6b22d5 -->

```go
// ============================================================
// FILE     : internal/reasoning/orchestrator_test.go
// PURPOSE  : Guards schema-first reasoning orchestration and stub requirements.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : reasoning orchestrator tests
// DEPENDS  : context, encoding/json, errors, strings, testing, time, internal/core
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: These tests must not introduce local extraction or real Codex calls.
// ============================================================

package reasoning

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

func TestStubOrchestratorRejectsMissingStage2OutputSchema(t *testing.T) {
	t.Parallel()

	envelope := testProcessTurnEnvelope()
	envelope.Stage2.RequiredOutputSchema = ""

	result, err := NewStubOrchestrator().ProcessTurn(context.Background(), envelope)
	if err == nil {
		t.Fatalf("expected missing required output schema to be rejected")
	}
	if result != nil {
		t.Fatalf("expected no reasoning result for invalid envelope, got %#v", result)
	}
	if !errors.Is(err, core.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
	if !strings.Contains(err.Error(), "required output schema") {
		t.Fatalf("expected required output schema error, got %v", err)
	}
}

func TestStubOrchestratorAcceptsPreparedStage2Envelope(t *testing.T) {
	t.Parallel()

	result, err := NewStubOrchestrator().ProcessTurn(context.Background(), testProcessTurnEnvelope())
	if err != nil {
		t.Fatalf("ProcessTurn returned error: %v", err)
	}
	if result == nil {
		t.Fatalf("expected structured stub result")
	}
	if result.Stage2.Trace.Stage != StageNameResolve {
		t.Fatalf("expected resolve trace, got %#v", result.Stage2.Trace)
	}
	if result.Stage2.Trace.SchemaVersion == "" {
		t.Fatalf("expected schema-shaped trace result, got %#v", result.Stage2.Trace)
	}
}

func TestPipelineOrchestratorPreparesStage2AfterStage1(t *testing.T) {
	t.Parallel()

	stage1Output := Stage1Output{
		CandidateEntities: []CandidateEntity{{
			EntityKind:    "agent",
			DisplayName:   "Hermes",
			Confidence:    0.91,
			MetadataJSON:  json.RawMessage(`{}`),
			SourceEventID: "evt_contract_1",
		}},
		CandidateMemories: []CandidateMemory{{
			Kind:          core.MemoryKindConstraint,
			ArtifactClass: core.ArtifactClassKnowledge,
			Scope:         core.MemoryScopeWorkspaceShared,
			Text:          "Use schema-first reasoning only.",
			Confidence:    0.93,
			RawEventIDs:   []string{"evt_contract_1"},
		}},
		SummaryHint: "schema-first",
		TaskHint:    "prepare stage 2 with stage 1 candidates",
	}
	memorySource := &capturingStage2MemorySource{
		memories: []core.MemoryResult{{
			MemoryID:      "mem_stage2_source",
			Kind:          core.MemoryKindConstraint,
			ArtifactClass: core.ArtifactClassKnowledge,
			Text:          "Existing source loaded after Stage 1.",
			Confidence:    0.88,
			Scope:         core.MemoryScopeWorkspaceShared,
			LatestFlag:    true,
		}},
	}
	stage2 := &capturingStage2Resolver{}
	orchestrator, err := NewPipelineOrchestrator(
		&fixedStage1Extractor{output: stage1Output},
		stage2,
		NewStage2InputPreparer(Stage2InputSources{Memories: memorySource}),
	)
	if err != nil {
		t.Fatalf("NewPipelineOrchestrator returned error: %v", err)
	}

	result, err := orchestrator.ProcessTurn(context.Background(), testProcessTurnEnvelope())
	if err != nil {
		t.Fatalf("ProcessTurn returned error: %v", err)
	}

	if memorySource.received.Stage1.TaskHint != stage1Output.TaskHint {
		t.Fatalf("expected Stage 2 source to receive Stage 1 output, got %#v", memorySource.received.Stage1)
	}
	if stage2.received.RequiredOutputSchema != Stage2ResolveOutputSchemaV0 {
		t.Fatalf("expected resolver to receive required output schema, got %q", stage2.received.RequiredOutputSchema)
	}
	if len(stage2.received.RelevantMemories) != 1 || stage2.received.RelevantMemories[0].MemoryID != "mem_stage2_source" {
		t.Fatalf("expected resolver to receive prepared memory source, got %#v", stage2.received.RelevantMemories)
	}
	if result.Stage1.TaskHint != stage1Output.TaskHint {
		t.Fatalf("expected result to preserve Stage 1 output, got %#v", result.Stage1)
	}
	if result.Stage2.Trace.Stage != StageNameResolve {
		t.Fatalf("expected resolve trace, got %#v", result.Stage2.Trace)
	}
}

func TestPipelineOrchestratorStopsBeforeStage2WhenStage1Fails(t *testing.T) {
	t.Parallel()

	memorySource := &capturingStage2MemorySource{}
	stage2 := &capturingStage2Resolver{}
	orchestrator, err := NewPipelineOrchestrator(
		&fixedStage1Extractor{err: errors.New("codex extract unavailable")},
		stage2,
		NewStage2InputPreparer(Stage2InputSources{Memories: memorySource}),
	)
	if err != nil {
		t.Fatalf("NewPipelineOrchestrator returned error: %v", err)
	}

	result, err := orchestrator.ProcessTurn(context.Background(), testProcessTurnEnvelope())
	if err == nil {
		t.Fatalf("expected Stage 1 error")
	}
	if result != nil {
		t.Fatalf("expected no result after Stage 1 failure, got %#v", result)
	}
	if memorySource.called {
		t.Fatalf("Stage 2 sources should not load after Stage 1 failure")
	}
	if stage2.called {
		t.Fatalf("Stage 2 resolver should not run after Stage 1 failure")
	}
	if !strings.Contains(err.Error(), "stage1 extract") {
		t.Fatalf("expected wrapped Stage 1 error, got %v", err)
	}
}

func testProcessTurnEnvelope() *ProcessTurnEnvelope {
	rawEvents := []*core.RawEvent{
		{
			ID:          "evt_contract_1",
			TenantID:    "tenant_1",
			WorkspaceID: "workspace_1",
			SessionID:   "session_1",
			ActorID:     "agent:hermes-main",
			EventKind:   "message",
			Source:      "hermes",
			Fingerprint: "fp_evt_contract_1",
			OccurredAt:  time.Date(2026, time.April, 24, 0, 0, 0, 0, time.UTC),
			PayloadJSON: json.RawMessage(`{"text":"contract gate"}`),
			CreatedAt:   time.Date(2026, time.April, 24, 0, 0, 0, 0, time.UTC),
		},
	}
	stage1 := Stage1Input{
		JobID:       "job_contract",
		TenantID:    "tenant_1",
		WorkspaceID: "workspace_1",
		RawEvents:   rawEvents,
	}
	stage2 := Stage2Input{
		JobID:                "job_contract",
		TenantID:             "tenant_1",
		WorkspaceID:          "workspace_1",
		RawEvents:            rawEvents,
		RelevantMemories:     []core.MemoryResult{},
		RelevantDocuments:    []core.DocumentChunkResult{},
		ActivePlans:          []*core.Plan{},
		PinnedNotes:          []*core.Note{},
		RequiredOutputName:   StageNameResolve,
		RequiredOutputSchema: Stage2ResolveOutputSchemaV0,
	}
	return &ProcessTurnEnvelope{
		JobID:       "job_contract",
		TenantID:    "tenant_1",
		WorkspaceID: "workspace_1",
		RawEventIDs: []string{"evt_contract_1"},
		RawEvents:   rawEvents,
		Stage1:      stage1,
		Stage2:      stage2,
	}
}

type fixedStage1Extractor struct {
	output Stage1Output
	err    error
}

func (e *fixedStage1Extractor) Extract(context.Context, Stage1Input) (Stage1Output, error) {
	if e.err != nil {
		return Stage1Output{}, e.err
	}
	return e.output, nil
}

type capturingStage2MemorySource struct {
	called   bool
	received Stage2InputRequest
	memories []core.MemoryResult
}

func (s *capturingStage2MemorySource) LoadStage2Memories(_ context.Context, req Stage2InputRequest) ([]core.MemoryResult, error) {
	s.called = true
	s.received = req
	return s.memories, nil
}

type capturingStage2Resolver struct {
	called   bool
	received Stage2Input
}

func (r *capturingStage2Resolver) Resolve(_ context.Context, input Stage2Input) (Stage2Output, error) {
	r.called = true
	r.received = input
	return emptyStage2Output(), nil
}

```



<!-- Source: internal/reasoning/stage2_input_preparer_test.go | bytes=7663 | lines=212 | sha16=437bf9fa6704fefc -->

```go
// ============================================================
// FILE     : internal/reasoning/stage2_input_preparer_test.go
// PURPOSE  : Verifies Stage 2 reasoning input envelope preparation.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : TestStage2InputPreparerPrepare_AssemblesRetrievedContext, TestStage2InputPreparerPrepare_AllowsMissingSources
// DEPENDS  : context, encoding/json, testing, time, internal/core, internal/reasoning
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: These tests must not introduce local extraction or real Codex calls.
// ============================================================

package reasoning

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

func TestStage2InputPreparerPrepare_AssemblesRetrievedContext(t *testing.T) {
	t.Parallel()

	events := []*core.RawEvent{testStage2RawEvent("evt_1")}
	stage1 := Stage1Output{
		CandidateEntities: []CandidateEntity{
			{EntityKind: "project", DisplayName: "VibeGravity", Confidence: 0.91, SourceEventID: "evt_1"},
		},
		CandidateMemories: []CandidateMemory{
			{
				Kind:          core.MemoryKindFact,
				ArtifactClass: core.ArtifactClassKnowledge,
				Scope:         core.MemoryScopeWorkspaceShared,
				Text:          "VibeGravity keeps local runtime embedding-only.",
				Confidence:    0.87,
				RawEventIDs:   []string{"evt_1"},
			},
		},
		SummaryHint: "reasoning envelope work",
		TaskHint:    "prepare Stage 2 input",
	}
	profile := &core.Profile{
		EntityID:    "agent:hermes-main",
		Scope:       core.MemoryScopeAgentPrivate,
		StaticJSON:  json.RawMessage(`{"name":"Hermes"}`),
		DynamicJSON: json.RawMessage(`{"project":"VibeGravity"}`),
		Version:     4,
	}
	memory := core.MemoryResult{
		MemoryID:      "mem_1",
		Kind:          core.MemoryKindConstraint,
		ArtifactClass: core.ArtifactClassKnowledge,
		Text:          "Do not reintroduce a local extractor.",
		Confidence:    0.98,
		Scope:         core.MemoryScopeWorkspaceShared,
		LatestFlag:    true,
	}
	document := core.DocumentChunkResult{ChunkID: "chunk_1", DocumentID: "doc_1", Text: "Stage 2 consumes related documents.", Score: 0.77}
	plan := &core.Plan{ID: "plan_1", TenantID: "tenant_1", WorkspaceID: "workspace_1", Title: "Ship reasoning envelope", Status: "active"}
	note := &core.Note{ID: "note_1", TenantID: "tenant_1", WorkspaceID: "workspace_1", Text: "Pinned constraint", Pinned: true}

	preparer := NewStage2InputPreparer(Stage2InputSources{
		Profiles:  staticProfileSource{profile: profile},
		Memories:  staticMemorySource{memories: []core.MemoryResult{memory}},
		Documents: staticDocumentSource{documents: []core.DocumentChunkResult{document}},
		Plans:     staticPlanSource{plans: []*core.Plan{plan}},
		Notes:     staticNoteSource{notes: []*core.Note{note}},
	})

	input, err := preparer.Prepare(context.Background(), Stage2InputRequest{
		JobID:       "job_1",
		TenantID:    "tenant_1",
		WorkspaceID: "workspace_1",
		RawEvents:   events,
		Stage1:      stage1,
	})
	if err != nil {
		t.Fatalf("Prepare returned error: %v", err)
	}

	if input.JobID != "job_1" || input.TenantID != "tenant_1" || input.WorkspaceID != "workspace_1" {
		t.Fatalf("unexpected identity fields: %#v", input)
	}
	if len(input.RawEvents) != 1 || input.RawEvents[0].ID != "evt_1" {
		t.Fatalf("raw events were not copied into Stage 2 input: %#v", input.RawEvents)
	}
	if got := input.Stage1.CandidateMemories[0].Text; got != stage1.CandidateMemories[0].Text {
		t.Fatalf("stage 1 candidates not preserved: %q", got)
	}
	if input.ExistingProfile == nil || input.ExistingProfile.EntityID != profile.EntityID {
		t.Fatalf("expected existing profile, got %#v", input.ExistingProfile)
	}
	if len(input.RelevantMemories) != 1 || input.RelevantMemories[0].MemoryID != memory.MemoryID {
		t.Fatalf("expected relevant memories, got %#v", input.RelevantMemories)
	}
	if len(input.RelevantDocuments) != 1 || input.RelevantDocuments[0].ChunkID != document.ChunkID {
		t.Fatalf("expected relevant documents, got %#v", input.RelevantDocuments)
	}
	if len(input.ActivePlans) != 1 || input.ActivePlans[0].ID != plan.ID {
		t.Fatalf("expected active plans, got %#v", input.ActivePlans)
	}
	if len(input.PinnedNotes) != 1 || input.PinnedNotes[0].ID != note.ID {
		t.Fatalf("expected pinned notes, got %#v", input.PinnedNotes)
	}
	if input.RequiredOutputName != StageNameResolve {
		t.Fatalf("unexpected required output name: %s", input.RequiredOutputName)
	}
	if input.RequiredOutputSchema != Stage2ResolveOutputSchemaV0 {
		t.Fatalf("unexpected required output schema marker: %s", input.RequiredOutputSchema)
	}
}

func TestStage2InputPreparerPrepare_AllowsMissingSources(t *testing.T) {
	t.Parallel()

	preparer := NewStage2InputPreparer(Stage2InputSources{})
	input, err := preparer.Prepare(context.Background(), Stage2InputRequest{
		JobID:       "job_1",
		TenantID:    "tenant_1",
		WorkspaceID: "workspace_1",
		RawEvents:   []*core.RawEvent{testStage2RawEvent("evt_1")},
		Stage1:      Stage1Output{},
	})
	if err != nil {
		t.Fatalf("Prepare returned error: %v", err)
	}

	if input.ExistingProfile != nil {
		t.Fatalf("expected nil profile when no profile source is configured")
	}
	if input.RelevantMemories == nil || input.RelevantDocuments == nil || input.ActivePlans == nil || input.PinnedNotes == nil {
		t.Fatalf("expected empty non-nil context slices: %#v", input)
	}
	if len(input.RelevantMemories) != 0 || len(input.RelevantDocuments) != 0 || len(input.ActivePlans) != 0 || len(input.PinnedNotes) != 0 {
		t.Fatalf("expected empty context slices: %#v", input)
	}
}

func TestStage2InputPreparerPrepare_RejectsMissingIdentity(t *testing.T) {
	t.Parallel()

	preparer := NewStage2InputPreparer(Stage2InputSources{})
	_, err := preparer.Prepare(context.Background(), Stage2InputRequest{
		TenantID:    "tenant_1",
		WorkspaceID: "workspace_1",
		RawEvents:   []*core.RawEvent{testStage2RawEvent("evt_1")},
	})
	if err == nil {
		t.Fatalf("expected missing job_id to be rejected")
	}
}

func testStage2RawEvent(id string) *core.RawEvent {
	return &core.RawEvent{
		ID:          id,
		TenantID:    "tenant_1",
		WorkspaceID: "workspace_1",
		SessionID:   "session_1",
		ActorID:     "agent:hermes-main",
		EventKind:   "message",
		Source:      "hermes",
		Fingerprint: "fp_" + id,
		OccurredAt:  time.Date(2026, time.April, 24, 0, 0, 0, 0, time.UTC),
		PayloadJSON: json.RawMessage(`{"text":"prepare Stage 2 input"}`),
		CreatedAt:   time.Date(2026, time.April, 24, 0, 0, 0, 0, time.UTC),
	}
}

type staticProfileSource struct {
	profile *core.Profile
}

func (s staticProfileSource) LoadStage2Profile(context.Context, Stage2InputRequest) (*core.Profile, error) {
	return s.profile, nil
}

type staticMemorySource struct {
	memories []core.MemoryResult
}

func (s staticMemorySource) LoadStage2Memories(context.Context, Stage2InputRequest) ([]core.MemoryResult, error) {
	return s.memories, nil
}

type staticDocumentSource struct {
	documents []core.DocumentChunkResult
}

func (s staticDocumentSource) LoadStage2Documents(context.Context, Stage2InputRequest) ([]core.DocumentChunkResult, error) {
	return s.documents, nil
}

type staticPlanSource struct {
	plans []*core.Plan
}

func (s staticPlanSource) LoadStage2ActivePlans(context.Context, Stage2InputRequest) ([]*core.Plan, error) {
	return s.plans, nil
}

type staticNoteSource struct {
	notes []*core.Note
}

func (s staticNoteSource) LoadStage2PinnedNotes(context.Context, Stage2InputRequest) ([]*core.Note, error) {
	return s.notes, nil
}

```



<!-- Source: internal/recall/assembler_test.go | bytes=18185 | lines=552 | sha16=3f556040623b812a -->

```go
// ============================================================
// FILE     : internal/recall/assembler_test.go
// PURPOSE  : Verifies prefetch typed block assembly, priority, scopes, and budget behavior.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : recall assembler behavior tests
// DEPENDS  : internal/recall, internal/core
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: Recall tests should assert typed blocks before any Hermes text rendering exists.
// ============================================================

package recall

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

func TestAssemblerPrefetch_PrioritizesManualControls(t *testing.T) {
	t.Parallel()

	notes := &fakeNoteStore{notes: []*core.Note{{
		ID:            "note_1",
		Text:          "Always prefer the Go-first plan.",
		Scope:         core.MemoryScopeWorkspaceShared,
		OwnerEntityID: "workspace:workspace_1",
		Pinned:        true,
	}}}
	plans := &fakePlanStore{plans: []*core.Plan{{
		ID:            "plan_1",
		Title:         "Implement sync_turn before worker reasoning.",
		Status:        "active",
		Scope:         core.MemoryScopeAgentPrivate,
		OwnerEntityID: "agent:hermes-main",
	}}}
	assembler := NewAssembler(Dependencies{
		Notes: notes,
		Plans: plans,
		Profiles: &fakeProfileStore{profiles: map[string]*core.Profile{
			"agent:hermes-main|agent_private": {
				EntityID:   "agent:hermes-main",
				Scope:      core.MemoryScopeAgentPrivate,
				StaticJSON: json.RawMessage(`{"style":"brief"}`),
			},
		}},
	})

	resp, err := assembler.Prefetch(context.Background(), testPrefetchRequest())
	if err != nil {
		t.Fatalf("Prefetch returned error: %v", err)
	}

	gotKinds := recallKinds(resp.Blocks)
	wantKinds := []string{"pinned_note", "active_plan", "profile_static"}
	if !reflect.DeepEqual(gotKinds, wantKinds) {
		t.Fatalf("unexpected block kinds: got %v want %v", gotKinds, wantKinds)
	}
	if resp.Meta.EstimatedTokens <= 0 {
		t.Fatalf("expected positive token estimate")
	}
	if !reflect.DeepEqual(resp.Meta.Sources, []string{"notes", "plans", "profile"}) {
		t.Fatalf("unexpected sources: %v", resp.Meta.Sources)
	}
	if resp.Blocks[0].Scope != core.MemoryScopeWorkspaceShared || resp.Blocks[0].Source != "notes" || resp.Blocks[0].SourceID != "note_1" || resp.Blocks[0].Status != "pinned" || resp.Blocks[0].Freshness != "stored" {
		t.Fatalf("pinned note block did not expose trust metadata: %#v", resp.Blocks[0])
	}
	if resp.Blocks[1].Scope != core.MemoryScopeAgentPrivate || resp.Blocks[1].Source != "plans" || resp.Blocks[1].SourceID != "plan_1" || resp.Blocks[1].Status != "active" {
		t.Fatalf("active plan block did not expose trust metadata: %#v", resp.Blocks[1])
	}
	if notes.lastReq.OwnerEntityID != "agent:hermes-main" || notes.lastReq.TenantID != "tenant_1" {
		t.Fatalf("pinned notes request was not actor scoped: %#v", notes.lastReq)
	}
	if plans.lastReq.OwnerEntityID != "agent:hermes-main" || plans.lastReq.TenantID != "tenant_1" {
		t.Fatalf("active plans request was not actor scoped: %#v", plans.lastReq)
	}
}

func TestAssemblerPrefetch_UsesScopeAwareMemorySearch(t *testing.T) {
	t.Parallel()

	memories := &fakeMemoryStore{resp: &core.SearchMemoriesResponse{
		Memories: []core.MemoryResult{{
			MemoryID:      "mem_1",
			Text:          "VibeGravity keeps private and shared memory separate.",
			Scope:         core.MemoryScopeWorkspaceShared,
			OwnerEntityID: "workspace:workspace_1",
			LatestFlag:    true,
		}},
	}}
	assembler := NewAssembler(Dependencies{
		Memories: memories,
	})

	resp, err := assembler.Prefetch(context.Background(), testPrefetchRequest())
	if err != nil {
		t.Fatalf("Prefetch returned error: %v", err)
	}

	if len(resp.Blocks) != 1 || resp.Blocks[0].Kind != "memory" {
		t.Fatalf("expected one memory block, got %#v", resp.Blocks)
	}
	wantScopes := []core.MemoryScope{
		core.MemoryScopeAgentPrivate,
		core.MemoryScopeWorkspaceShared,
		core.MemoryScopeSessionScratch,
	}
	if !reflect.DeepEqual(memories.lastReq.Scopes, wantScopes) {
		t.Fatalf("unexpected scopes: got %v want %v", memories.lastReq.Scopes, wantScopes)
	}
	if memories.lastReq.OwnerEntityID != "agent:hermes-main" {
		t.Fatalf("memory search request was not actor scoped: %#v", memories.lastReq)
	}
	if resp.Blocks[0].Source != "memories" || resp.Blocks[0].SourceID != "mem_1" || resp.Blocks[0].Scope != core.MemoryScopeWorkspaceShared || resp.Blocks[0].Status != "active" {
		t.Fatalf("memory block did not expose source and scope metadata: %#v", resp.Blocks[0])
	}
}

func TestAssemblerPrefetch_IncludesGroupSharedMemoriesForMemberActor(t *testing.T) {
	t.Parallel()

	groupID := "group_design"
	memories := &fakeMemoryStore{resp: &core.SearchMemoriesResponse{
		Memories: []core.MemoryResult{{
			MemoryID:   "mem_group",
			Text:       "Design group agreed to keep MCP as the first external protocol.",
			Scope:      core.MemoryScopeGroupShared,
			GroupID:    &groupID,
			LatestFlag: true,
		}},
	}}
	assembler := NewAssembler(Dependencies{
		Memories: memories,
		Groups: &fakeGroupStore{memberships: []*core.MemoryGroupMembership{{
			GroupID:  "group_design",
			EntityID: "agent:hermes-main",
		}}},
	})

	resp, err := assembler.Prefetch(context.Background(), testPrefetchRequest())
	if err != nil {
		t.Fatalf("Prefetch returned error: %v", err)
	}

	if len(resp.Blocks) != 1 || resp.Blocks[0].Kind != "memory" {
		t.Fatalf("expected group memory block, got %#v", resp.Blocks)
	}
	wantScopes := []core.MemoryScope{
		core.MemoryScopeAgentPrivate,
		core.MemoryScopeWorkspaceShared,
		core.MemoryScopeSessionScratch,
		core.MemoryScopeGroupShared,
	}
	if !reflect.DeepEqual(memories.lastReq.Scopes, wantScopes) {
		t.Fatalf("unexpected scopes: got %v want %v", memories.lastReq.Scopes, wantScopes)
	}
	if !reflect.DeepEqual(memories.lastReq.VisibleGroupIDs, []string{"group_design"}) {
		t.Fatalf("expected visible group ids, got %#v", memories.lastReq.VisibleGroupIDs)
	}
	if resp.Blocks[0].Scope != core.MemoryScopeGroupShared {
		t.Fatalf("group memory block should expose group scope: %#v", resp.Blocks[0])
	}
}

func TestAssemblerPrefetch_MarksMissingStoresAsDegraded(t *testing.T) {
	t.Parallel()

	assembler := NewAssembler(Dependencies{
		Notes: &fakeNoteStore{notes: []*core.Note{{
			ID:            "note_1",
			Text:          "Keep Hermes Memory trust loop visible.",
			Scope:         core.MemoryScopeWorkspaceShared,
			OwnerEntityID: "workspace:workspace_1",
			Pinned:        true,
		}}},
	})

	resp, err := assembler.Prefetch(context.Background(), testPrefetchRequest())
	if err != nil {
		t.Fatalf("Prefetch returned error: %v", err)
	}

	if !resp.Meta.Degraded {
		t.Fatalf("expected degraded metadata when stores are unavailable: %#v", resp.Meta)
	}
	if !containsString(resp.Meta.DegradedReasons, "memories_unavailable") {
		t.Fatalf("expected memories_unavailable reason, got %#v", resp.Meta.DegradedReasons)
	}
}

func TestAssemblerPrefetch_MarksDerivedRecallStaleFromBacklogFreshness(t *testing.T) {
	t.Parallel()

	lagSeconds := int64(120)
	assembler := NewAssembler(Dependencies{
		Notes: &fakeNoteStore{notes: []*core.Note{{
			ID:            "note_1",
			Text:          "Manual operator guardrail stays current.",
			Scope:         core.MemoryScopeWorkspaceShared,
			OwnerEntityID: "workspace:workspace_1",
			Pinned:        true,
		}}},
		Memories: &fakeMemoryStore{resp: &core.SearchMemoriesResponse{
			Memories: []core.MemoryResult{{
				MemoryID:      "mem_1",
				Text:          "Worker-derived memory may lag during Codex outage.",
				Scope:         core.MemoryScopeAgentPrivate,
				OwnerEntityID: "agent:hermes-main",
				LatestFlag:    true,
			}},
		}},
		Freshness: fakeFreshnessProvider{state: Freshness{
			Freshness:       "stale",
			LagSeconds:      &lagSeconds,
			Reasons:         []string{"worker_backlog_stale", "codex_or_worker_retry_backlog"},
			AffectedSources: []string{"memories"},
		}},
	})

	resp, err := assembler.Prefetch(context.Background(), testPrefetchRequest())
	if err != nil {
		t.Fatalf("Prefetch returned error: %v", err)
	}

	if !resp.Meta.Degraded {
		t.Fatalf("expected degraded meta from freshness provider: %#v", resp.Meta)
	}
	if resp.Meta.Freshness != "stale" || resp.Meta.FreshnessLagSeconds == nil || *resp.Meta.FreshnessLagSeconds != lagSeconds {
		t.Fatalf("unexpected freshness metadata: %#v", resp.Meta)
	}
	if !containsString(resp.Meta.DegradedReasons, "worker_backlog_stale") || !containsString(resp.Meta.DegradedReasons, "codex_or_worker_retry_backlog") {
		t.Fatalf("missing freshness degraded reasons: %#v", resp.Meta.DegradedReasons)
	}
	if resp.Blocks[0].Kind != "pinned_note" || resp.Blocks[0].Freshness != "stored" {
		t.Fatalf("manual note freshness should remain stored: %#v", resp.Blocks)
	}
	if resp.Blocks[1].Kind != "memory" || resp.Blocks[1].Freshness != "stale" {
		t.Fatalf("derived memory freshness should be stale: %#v", resp.Blocks)
	}
}

func TestBacklogFreshnessProvider_MarksRetryBacklogStale(t *testing.T) {
	t.Parallel()

	lagSeconds := int64(90)
	provider := BacklogFreshnessProvider{
		Jobs: fakeJobMetricsStore{metrics: &core.JobBacklogMetrics{
			Counts: core.JobStatusCounts{
				Queued:      1,
				ReadyQueued: 1,
			},
			OldestQueuedAgeSeconds:  &lagSeconds,
			RetryableQueuedAttempts: 1,
		}},
		StaleAfter: time.Minute,
	}

	state, err := provider.RecallFreshness(context.Background(), testPrefetchRequest())
	if err != nil {
		t.Fatalf("RecallFreshness returned error: %v", err)
	}
	if state.Freshness != "stale" || state.LagSeconds == nil || *state.LagSeconds != lagSeconds {
		t.Fatalf("unexpected backlog freshness state: %#v", state)
	}
	if !containsString(state.Reasons, "worker_backlog_stale") || !containsString(state.Reasons, "codex_or_worker_retry_backlog") {
		t.Fatalf("unexpected backlog freshness reasons: %#v", state.Reasons)
	}
}

func TestBacklogFreshnessProvider_MarksLongRunningJobsStale(t *testing.T) {
	t.Parallel()

	queuedLagSeconds := int64(45)
	runningLagSeconds := int64(180)
	provider := BacklogFreshnessProvider{
		Jobs: fakeJobMetricsStore{metrics: &core.JobBacklogMetrics{
			Counts: core.JobStatusCounts{
				Queued:      1,
				ReadyQueued: 1,
				Running:     1,
			},
			OldestQueuedAgeSeconds:  &queuedLagSeconds,
			OldestRunningAgeSeconds: &runningLagSeconds,
		}},
		StaleAfter: time.Minute,
	}

	state, err := provider.RecallFreshness(context.Background(), testPrefetchRequest())
	if err != nil {
		t.Fatalf("RecallFreshness returned error: %v", err)
	}
	if state.Freshness != "stale" || state.LagSeconds == nil || *state.LagSeconds != runningLagSeconds {
		t.Fatalf("expected running job lag to drive stale freshness: %#v", state)
	}
	if containsString(state.Reasons, "worker_backlog_stale") {
		t.Fatalf("queued job younger than threshold should not mark backlog stale: %#v", state.Reasons)
	}
	if !containsString(state.Reasons, "worker_running_stale") {
		t.Fatalf("missing running stale reason: %#v", state.Reasons)
	}
}

func TestAssemblerPrefetch_TruncatesToBudget(t *testing.T) {
	t.Parallel()

	assembler := NewAssembler(Dependencies{
		Notes: &fakeNoteStore{notes: []*core.Note{{
			Text:   "one two three four five six seven eight nine ten eleven twelve",
			Pinned: true,
		}}},
	})
	req := testPrefetchRequest()
	req.BudgetTokens = 4

	resp, err := assembler.Prefetch(context.Background(), req)
	if err != nil {
		t.Fatalf("Prefetch returned error: %v", err)
	}
	if len(resp.Blocks) != 1 {
		t.Fatalf("expected one truncated block, got %d", len(resp.Blocks))
	}
	if resp.Meta.EstimatedTokens > req.BudgetTokens {
		t.Fatalf("estimated tokens exceeded budget: got %d budget %d", resp.Meta.EstimatedTokens, req.BudgetTokens)
	}
}

func TestAssemblerPrefetch_PreservesPlanWhenPinnedNoteIsLong(t *testing.T) {
	t.Parallel()

	assembler := NewAssembler(Dependencies{
		Notes: &fakeNoteStore{notes: []*core.Note{{
			Text:   strings.Repeat("manual guardrail ", 120),
			Pinned: true,
		}}},
		Plans: &fakePlanStore{plans: []*core.Plan{{
			Title:  "Finish recall token budgeting before graph quality work.",
			Status: "active",
		}}},
	})
	req := testPrefetchRequest()
	req.BudgetTokens = 40

	resp, err := assembler.Prefetch(context.Background(), req)
	if err != nil {
		t.Fatalf("Prefetch returned error: %v", err)
	}

	gotKinds := recallKinds(resp.Blocks)
	wantKinds := []string{"pinned_note", "active_plan"}
	if !reflect.DeepEqual(gotKinds, wantKinds) {
		t.Fatalf("unexpected block kinds: got %v want %v", gotKinds, wantKinds)
	}
	if resp.Meta.EstimatedTokens > req.BudgetTokens {
		t.Fatalf("estimated tokens exceeded budget: got %d budget %d", resp.Meta.EstimatedTokens, req.BudgetTokens)
	}
}

func TestAssemblerPrefetch_RanksMemoryByRelevanceConfidenceAndRecency(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 24, 0, 0, 0, 0, time.UTC)
	memories := &fakeMemoryStore{resp: &core.SearchMemoriesResponse{
		Memories: []core.MemoryResult{
			{
				MemoryID:   "old_generic",
				Text:       "General workspace context for prior implementation work.",
				Confidence: 0.95,
				ValidFrom:  now.Add(-72 * time.Hour),
				LatestFlag: true,
			},
			{
				MemoryID:   "recent_recall",
				Text:       "Recall token budgeting must preserve active plans and suppress noisy context.",
				Confidence: 0.80,
				ValidFrom:  now.Add(-1 * time.Hour),
				LatestFlag: true,
			},
		},
	}}
	assembler := NewAssembler(Dependencies{
		Memories: memories,
		Clock:    func() time.Time { return now },
	})
	req := testPrefetchRequest()
	req.Query = "recall token budgeting quality"

	resp, err := assembler.Prefetch(context.Background(), req)
	if err != nil {
		t.Fatalf("Prefetch returned error: %v", err)
	}
	if len(resp.Blocks) < 2 {
		t.Fatalf("expected both memory candidates, got %#v", resp.Blocks)
	}
	if resp.Blocks[0].Text != "Recall token budgeting must preserve active plans and suppress noisy context." {
		t.Fatalf("expected relevant recent memory first, got %#v", resp.Blocks)
	}
}

func testPrefetchRequest() *core.PrefetchRequest {
	return &core.PrefetchRequest{
		TenantID:     "tenant_1",
		WorkspaceID:  "workspace_1",
		SessionID:    "session_1",
		ActorID:      "agent:hermes-main",
		Query:        "What should Hermes remember next?",
		BudgetTokens: 2200,
		Mode:         "default",
	}
}

func recallKinds(blocks []core.RecallBlock) []string {
	kinds := make([]string, 0, len(blocks))
	for _, block := range blocks {
		kinds = append(kinds, block.Kind)
	}
	return kinds
}

func containsString(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

type fakeNoteStore struct {
	lastReq *core.ListPinnedNotesRequest
	notes   []*core.Note
}

func (s *fakeNoteStore) AddNote(context.Context, *core.Note) error {
	return core.ErrNotImplemented
}

func (s *fakeNoteStore) ListPinnedNotes(_ context.Context, req *core.ListPinnedNotesRequest) ([]*core.Note, error) {
	reqCopy := *req
	reqCopy.Scopes = append([]core.MemoryScope(nil), req.Scopes...)
	s.lastReq = &reqCopy
	return s.notes, nil
}

type fakePlanStore struct {
	lastReq *core.GetActivePlansRequest
	plans   []*core.Plan
}

func (s *fakePlanStore) CreatePlan(context.Context, *core.Plan, []*core.PlanItem) error {
	return core.ErrNotImplemented
}

func (s *fakePlanStore) UpdatePlan(context.Context, *core.Plan, []*core.PlanItem) error {
	return core.ErrNotImplemented
}

func (s *fakePlanStore) GetActivePlans(_ context.Context, req *core.GetActivePlansRequest) ([]*core.Plan, error) {
	reqCopy := *req
	reqCopy.Scopes = append([]core.MemoryScope(nil), req.Scopes...)
	s.lastReq = &reqCopy
	return s.plans, nil
}

type fakeProfileStore struct {
	profiles map[string]*core.Profile
}

func (s *fakeProfileStore) GetProfile(_ context.Context, entityID string, scope core.MemoryScope) (*core.Profile, error) {
	profile, ok := s.profiles[entityID+"|"+string(scope)]
	if !ok {
		return nil, core.ErrNotFound
	}
	return profile, nil
}

func (s *fakeProfileStore) UpsertProfile(context.Context, *core.Profile) error {
	return core.ErrNotImplemented
}

type fakeMemoryStore struct {
	lastReq *core.SearchMemoriesRequest
	resp    *core.SearchMemoriesResponse
}

func (s *fakeMemoryStore) UpsertMemory(context.Context, *core.Memory) error {
	return core.ErrNotImplemented
}

func (s *fakeMemoryStore) GetMemory(context.Context, string) (*core.Memory, error) {
	return nil, core.ErrNotImplemented
}

func (s *fakeMemoryStore) SearchMemories(_ context.Context, req *core.SearchMemoriesRequest) (*core.SearchMemoriesResponse, error) {
	s.lastReq = req
	return s.resp, nil
}

func (s *fakeMemoryStore) UpsertMemoryEdge(context.Context, *core.MemoryEdge) error {
	return core.ErrNotImplemented
}

func (s *fakeMemoryStore) WriteMemoryTrace(context.Context, *core.MemoryTrace) error {
	return core.ErrNotImplemented
}

func (s *fakeMemoryStore) ExplainMemory(context.Context, *core.ExplainMemoryRequest) (*core.ExplainMemoryResponse, error) {
	return nil, core.ErrNotImplemented
}

type fakeGroupStore struct {
	memberships []*core.MemoryGroupMembership
}

func (s *fakeGroupStore) CreateMemoryGroup(context.Context, *core.MemoryGroup) error {
	return core.ErrNotImplemented
}

func (s *fakeGroupStore) AddMembership(context.Context, *core.MemoryGroupMembership) error {
	return core.ErrNotImplemented
}

func (s *fakeGroupStore) ListMemberships(context.Context, string) ([]*core.MemoryGroupMembership, error) {
	return nil, core.ErrNotImplemented
}

func (s *fakeGroupStore) ListMembershipsForEntity(context.Context, string, string, string) ([]*core.MemoryGroupMembership, error) {
	return s.memberships, nil
}

type fakeFreshnessProvider struct {
	state Freshness
}

func (p fakeFreshnessProvider) RecallFreshness(context.Context, *core.PrefetchRequest) (Freshness, error) {
	return p.state, nil
}

type fakeJobMetricsStore struct {
	metrics *core.JobBacklogMetrics
}

func (s fakeJobMetricsStore) GetJobBacklogMetrics(context.Context, *core.JobBacklogMetricsRequest) (*core.JobBacklogMetrics, error) {
	return s.metrics, nil
}

```



<!-- Source: internal/store/postgres/concurrency_integration_test.go | bytes=8877 | lines=263 | sha16=91a375d4ebc66513 -->

```go
// ============================================================
// FILE     : internal/store/postgres/concurrency_integration_test.go
// PURPOSE  : Stress-tests PostgreSQL graph update concurrency and provenance integrity.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : TestPostgresConcurrentUpdateMemoryAllowsOneWinnerNoDanglingWrites
// DEPENDS  : context, os, sync, testing, time, internal/core, github.com/jackc/pgx/v5/pgxpool
// USED_BY  : go test ./internal/store/postgres
// ------------------------------------------------------------
// AGENT_NOTE: Keep this test skippable when VIBEGRAVITY_DB_URL is unset; it verifies real row-lock behavior.
// ============================================================

package postgres

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

func TestPostgresConcurrentUpdateMemoryAllowsOneWinnerNoDanglingWrites(t *testing.T) {
	dbURL := os.Getenv("VIBEGRAVITY_DB_URL")
	if dbURL == "" {
		t.Skip("Skipping Postgres concurrency integration test because VIBEGRAVITY_DB_URL is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("connect test database: %v", err)
	}
	defer pool.Close()

	store := NewStore(pool)
	tenantID := fmt.Sprintf("tenant_concurrency_%d", time.Now().UnixNano())
	workspaceID := "workspace_concurrency"
	ownerID := "agent:hermes-main"
	targetID := "mem_concurrency_target"
	startedAt := time.Now().UTC()

	cleanupPostgresConcurrencyRows(ctx, t, pool, tenantID, workspaceID)
	defer cleanupPostgresConcurrencyRows(context.Background(), t, pool, tenantID, workspaceID)

	mustSeedJob(ctx, t, pool, tenantID, workspaceID, "job_seed")
	if err := store.CreateMemoryWithTrace(ctx, &core.Memory{
		ID:            targetID,
		TenantID:      tenantID,
		WorkspaceID:   workspaceID,
		Scope:         core.MemoryScopeWorkspaceShared,
		OwnerEntityID: ownerID,
		Kind:          core.MemoryKindFact,
		ArtifactClass: core.ArtifactClassKnowledge,
		Text:          "Original memory before concurrent update.",
		Fingerprint:   "fp_concurrency_target",
		Confidence:    0.7,
		Status:        core.MemoryStatusActive,
		ValidFrom:     startedAt,
		LatestFlag:    true,
		MetadataJSON:  []byte(`{}`),
		CreatedAt:     startedAt,
		UpdatedAt:     startedAt,
	}, &core.MemoryTrace{
		MemoryID:              targetID,
		RawEventIDs:           []string{"evt_seed"},
		ReasoningJobID:        "job_seed",
		ReasoningStage:        "resolve",
		CandidateSnapshotJSON: []byte(`{"seed":true}`),
		AppliedOperationsJSON: []byte(`[{"operation_id":"seed"}]`),
		RelatedDocumentIDs:    []string{},
		CreatedAt:             startedAt,
	}); err != nil {
		t.Fatalf("seed target memory with trace: %v", err)
	}

	const workers = 16
	var ready sync.WaitGroup
	ready.Add(workers)
	start := make(chan struct{})
	errs := make(chan error, workers)
	var successes atomic.Int32

	for i := 0; i < workers; i++ {
		i := i
		go func() {
			ready.Done()
			<-start

			jobID := fmt.Sprintf("job_concurrency_%02d", i)
			if err := insertSeedJob(context.Background(), pool, tenantID, workspaceID, jobID); err != nil {
				errs <- err
				return
			}
			workerCtx, workerCancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer workerCancel()

			memoryID := fmt.Sprintf("mem_concurrency_update_%02d", i)
			err := store.CreateMemoryWithTraceAndUpdateEdge(workerCtx,
				&core.Memory{
					ID:            memoryID,
					TenantID:      tenantID,
					WorkspaceID:   workspaceID,
					Scope:         core.MemoryScopeWorkspaceShared,
					OwnerEntityID: ownerID,
					Kind:          core.MemoryKindFact,
					ArtifactClass: core.ArtifactClassKnowledge,
					Text:          fmt.Sprintf("Concurrent update winner candidate %02d.", i),
					Fingerprint:   fmt.Sprintf("fp_concurrency_update_%02d", i),
					Confidence:    0.8,
					Status:        core.MemoryStatusActive,
					ValidFrom:     startedAt.Add(time.Duration(i+1) * time.Millisecond),
					LatestFlag:    true,
					MetadataJSON:  []byte(`{}`),
					CreatedAt:     startedAt.Add(time.Duration(i+1) * time.Millisecond),
					UpdatedAt:     startedAt.Add(time.Duration(i+1) * time.Millisecond),
				},
				&core.MemoryTrace{
					MemoryID:              memoryID,
					RawEventIDs:           []string{fmt.Sprintf("evt_concurrency_%02d", i)},
					ReasoningJobID:        jobID,
					ReasoningStage:        "resolve",
					CandidateSnapshotJSON: []byte(`{"candidate_memories":[]}`),
					AppliedOperationsJSON: []byte(fmt.Sprintf(`[{"operation_id":"op_update_%02d"}]`, i)),
					RelatedDocumentIDs:    []string{},
					CreatedAt:             time.Now().UTC(),
				},
				&core.MemoryEdge{
					FromMemoryID:   memoryID,
					ToMemoryID:     targetID,
					EdgeKind:       core.EdgeKindUpdates,
					Confidence:     0.8,
					CreatedByJobID: jobID,
					CreatedAt:      time.Now().UTC(),
				})
			if err == nil {
				successes.Add(1)
				errs <- nil
				return
			}
			if errors.Is(err, core.ErrConflict) {
				errs <- nil
				return
			}
			errs <- err
		}()
	}

	ready.Wait()
	close(start)
	for i := 0; i < workers; i++ {
		if err := <-errs; err != nil {
			t.Fatalf("concurrent update returned unexpected error: %v", err)
		}
	}

	if got := successes.Load(); got != 1 {
		t.Fatalf("expected exactly one concurrent update winner, got %d", got)
	}
	assertPostgresConcurrencyGraphIntegrity(ctx, t, pool, tenantID, workspaceID, targetID)
}

func mustSeedJob(ctx context.Context, t testing.TB, pool *pgxpool.Pool, tenantID string, workspaceID string, jobID string) {
	t.Helper()
	if err := insertSeedJob(ctx, pool, tenantID, workspaceID, jobID); err != nil {
		t.Fatalf("seed job %q: %v", jobID, err)
	}
}

func insertSeedJob(ctx context.Context, pool *pgxpool.Pool, tenantID string, workspaceID string, jobID string) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO ingest_jobs (id, tenant_id, workspace_id, job_kind, status, raw_event_ids, payload_json)
		VALUES ($1, $2, $3, 'process_turn_event', 'completed', '{}', '{}'::jsonb)
		ON CONFLICT (id) DO NOTHING
	`, jobID, tenantID, workspaceID)
	if err != nil {
		return fmt.Errorf("insert ingest job %q: %w", jobID, err)
	}
	return nil
}

func assertPostgresConcurrencyGraphIntegrity(ctx context.Context, t *testing.T, pool *pgxpool.Pool, tenantID string, workspaceID string, targetID string) {
	t.Helper()

	var winnerCount int
	if err := pool.QueryRow(ctx, `
		SELECT count(*)
		FROM memory_edges me
		JOIN memories m ON m.id = me.from_memory_id
		JOIN memory_trace mt ON mt.memory_id = m.id
		WHERE m.tenant_id = $1
		  AND m.workspace_id = $2
		  AND me.to_memory_id = $3
		  AND me.edge_kind = 'updates'
		  AND m.status = 'active'
		  AND m.latest_flag = true
	`, tenantID, workspaceID, targetID).Scan(&winnerCount); err != nil {
		t.Fatalf("count committed update winner: %v", err)
	}
	if winnerCount != 1 {
		t.Fatalf("expected one active latest update winner with trace and edge, got %d", winnerCount)
	}

	var targetStatus core.MemoryStatus
	var targetLatest bool
	if err := pool.QueryRow(ctx, `
		SELECT status, latest_flag
		FROM memories
		WHERE id = $1 AND tenant_id = $2 AND workspace_id = $3
	`, targetID, tenantID, workspaceID).Scan(&targetStatus, &targetLatest); err != nil {
		t.Fatalf("load update target status: %v", err)
	}
	if targetStatus != core.MemoryStatusSuperseded || targetLatest {
		t.Fatalf("target should be superseded and non-latest, got status=%q latest=%v", targetStatus, targetLatest)
	}

	var danglingCount int
	if err := pool.QueryRow(ctx, `
		SELECT count(*)
		FROM memories m
		LEFT JOIN memory_trace mt ON mt.memory_id = m.id
		LEFT JOIN memory_edges me
		  ON me.from_memory_id = m.id
		 AND me.to_memory_id = $3
		 AND me.edge_kind = 'updates'
		WHERE m.tenant_id = $1
		  AND m.workspace_id = $2
		  AND m.id LIKE 'mem_concurrency_update_%'
		  AND (mt.memory_id IS NULL OR me.from_memory_id IS NULL)
	`, tenantID, workspaceID, targetID).Scan(&danglingCount); err != nil {
		t.Fatalf("count dangling update writes: %v", err)
	}
	if danglingCount != 0 {
		t.Fatalf("expected no dangling memory/trace rows from losing workers, got %d", danglingCount)
	}
}

func cleanupPostgresConcurrencyRows(ctx context.Context, t testing.TB, pool *pgxpool.Pool, tenantID string, workspaceID string) {
	t.Helper()
	if _, err := pool.Exec(ctx, `
		DELETE FROM memories
		WHERE tenant_id = $1 AND workspace_id = $2
	`, tenantID, workspaceID); err != nil {
		t.Fatalf("cleanup memories: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		DELETE FROM ingest_jobs
		WHERE tenant_id = $1 AND workspace_id = $2
	`, tenantID, workspaceID); err != nil {
		t.Fatalf("cleanup ingest jobs: %v", err)
	}
}

```



<!-- Source: internal/store/postgres/corrections_test.go | bytes=2869 | lines=79 | sha16=7ca2ef4d479242e1 -->

```go
// ============================================================
// FILE     : internal/store/postgres/corrections_test.go
// PURPOSE  : Verifies correction persistence SQL without requiring a live database.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : postgres correction source tests
// DEPENDS  : strings, testing
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: Lock the narrow correction intake boundary without testing supersession semantics here.
// ============================================================

package postgres

import (
	"strings"
	"testing"
)

func TestRecordMemoryCorrectionUsesSingleTransaction(t *testing.T) {
	t.Parallel()

	source := readPostgresSourceFile(t, "corrections.go")
	recordSource := extractPostgresSourceBetween(t, source, "func (s *Store) RecordMemoryCorrection", "func insertCorrectionRawEvent")

	for _, want := range []string{
		"s.pool.BeginTx(ctx, pgx.TxOptions{})",
		"defer func() { _ = tx.Rollback(ctx) }()",
		"insertCorrectionRawEvent(ctx, tx, event)",
		"insertMemoryCorrection(ctx, tx, correction)",
		"tx.Commit(ctx)",
	} {
		if !strings.Contains(recordSource, want) {
			t.Fatalf("RecordMemoryCorrection must preserve %q, got:\n%s", want, recordSource)
		}
	}
}

func TestCorrectionRawEventReturnsStableIDOnRetry(t *testing.T) {
	t.Parallel()

	source := readPostgresSourceFile(t, "corrections.go")
	insertSource := extractPostgresSourceBetween(t, source, "func insertCorrectionRawEvent", "func insertMemoryCorrection")

	for _, want := range []string{
		"ON CONFLICT (tenant_id, source, idempotency_key) DO UPDATE",
		"WHERE raw_events.workspace_id = EXCLUDED.workspace_id",
		"RETURNING id",
	} {
		if !strings.Contains(insertSource, want) {
			t.Fatalf("correction raw event insert must preserve stable retry IDs; missing %q in:\n%s", want, insertSource)
		}
	}
}

func TestMemoryCorrectionArtifactIsAppendSafe(t *testing.T) {
	t.Parallel()

	source := readPostgresSourceFile(t, "corrections.go")
	start := strings.Index(source, "func insertMemoryCorrection")
	if start < 0 {
		t.Fatal("missing insertMemoryCorrection")
	}
	insertSource := source[start:]
	if strings.Contains(insertSource, "latest_flag") || strings.Contains(insertSource, "memory_trace") || strings.Contains(insertSource, "memory_edges") {
		t.Fatalf("correction intake must not mutate graph apply/provenance tables, got:\n%s", insertSource)
	}
	for _, want := range []string{
		"INSERT INTO memory_corrections",
		"ON CONFLICT (tenant_id, workspace_id, idempotency_key) DO UPDATE",
		"RETURNING id, tenant_id, workspace_id, memory_id, operator_id, raw_event_id",
	} {
		if !strings.Contains(insertSource, want) {
			t.Fatalf("memory correction insert must preserve %q, got:\n%s", want, insertSource)
		}
	}
}

```



<!-- Source: internal/store/postgres/documents_test.go | bytes=3155 | lines=89 | sha16=a009591bdff366eb -->

```go
// ============================================================
// FILE     : internal/store/postgres/documents_test.go
// PURPOSE  : Verifies document storage source preserves idempotent chunk replacement.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : postgres document source tests
// DEPENDS  : os, path/filepath, runtime, strings, testing
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: These tests lock storage contracts without requiring a live database.
// ============================================================

package postgres

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestAddDocumentChunksReplacesExistingChunksForDocument(t *testing.T) {
	t.Parallel()

	source := readPostgresSourceFile(t, "documents.go")
	addChunksSource := extractPostgresSourceBetween(t, source, "func (s *Store) AddDocumentChunks", "if err := tx.Commit")
	replaceChunksSource := extractPostgresSourceBetween(t, source, "func replaceDocumentChunksInTx", "func insertDocumentChunkInTx")

	if !strings.Contains(replaceChunksSource, "DELETE FROM document_chunks WHERE document_id = $1") {
		t.Fatalf("document chunk replacement must delete old chunks for idempotent document upserts, got:\n%s", replaceChunksSource)
	}
	if !strings.Contains(addChunksSource, "replacedDocuments") {
		t.Fatalf("AddDocumentChunks should delete once per document, got:\n%s", addChunksSource)
	}
	if !strings.Contains(addChunksSource, "replaceDocumentChunksInTx(ctx, tx, chunk.DocumentID") {
		t.Fatalf("AddDocumentChunks should route first chunk per document through replacement helper, got:\n%s", addChunksSource)
	}
}

func TestAddDocumentWithChunksUsesSingleTransaction(t *testing.T) {
	t.Parallel()

	source := readPostgresSourceFile(t, "documents.go")
	atomicSource := extractPostgresSourceBetween(t, source, "func (s *Store) AddDocumentWithChunks", "func (s *Store) AddDocument(")

	for _, required := range []string{
		"s.pool.Begin(ctx)",
		"defer func() { _ = tx.Rollback(ctx) }()",
		"addDocumentInTx(ctx, tx, document)",
		"addDocumentChunksInTx(ctx, tx, document.ID, chunks)",
		"tx.Commit(ctx)",
	} {
		if !strings.Contains(atomicSource, required) {
			t.Fatalf("AddDocumentWithChunks must keep document and chunks in one transaction; missing %q in:\n%s", required, atomicSource)
		}
	}
}

func readPostgresSourceFile(t *testing.T, name string) string {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not locate current test file")
	}
	data, err := os.ReadFile(filepath.Join(filepath.Dir(currentFile), name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return string(data)
}

func extractPostgresSourceBetween(t *testing.T, source, startMarker, endMarker string) string {
	t.Helper()

	start := strings.Index(source, startMarker)
	if start < 0 {
		t.Fatalf("missing start marker %q", startMarker)
	}
	remainder := source[start:]
	end := strings.Index(remainder, endMarker)
	if end < 0 {
		t.Fatalf("missing end marker %q", endMarker)
	}
	return remainder[:end]
}

```



<!-- Source: internal/store/postgres/dreaming_test.go | bytes=2957 | lines=87 | sha16=4f07fc03d07611cb -->

```go
// ============================================================
// FILE     : internal/store/postgres/dreaming_test.go
// PURPOSE  : Verifies PostgreSQL dreaming promotion query contracts without a live database.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : postgres dreaming helper tests
// DEPENDS  : strings, testing, time, internal/core
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: Promotion SQL must update metadata only, never scope or provenance fields.
// ============================================================

package postgres

import (
	"strings"
	"testing"
	"time"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

func TestDreamingPromotionStatementMarksMetadataOnly(t *testing.T) {
	t.Parallel()

	sql, args := dreamingPromotionStatement(&core.DreamingPromotionRequest{
		JobID:         "job_dream_1",
		TenantID:      "tenant_1",
		WorkspaceID:   "workspace_1",
		MemoryIDs:     []string{"mem_1", "mem_2"},
		Tier:          core.DreamingTierMidTerm,
		Now:           time.Date(2026, time.April, 24, 3, 0, 0, 0, time.UTC),
		MinConfidence: 0.5,
	})

	if !strings.Contains(sql, "jsonb_set") || !strings.Contains(sql, "'{dreaming}'") {
		t.Fatalf("expected dreaming metadata update, got: %s", sql)
	}
	if strings.Contains(sql, "scope =") || strings.Contains(sql, "owner_entity_id =") || strings.Contains(sql, "memory_trace") {
		t.Fatalf("promotion SQL must not mutate scope, owner, or trace: %s", sql)
	}
	if !strings.Contains(sql, "id = ANY($7::text[])") {
		t.Fatalf("expected explicit memory id filter, got: %s", sql)
	}
	if len(args) != 7 {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestDreamingPromotionStatementStableWorkspaceFilters(t *testing.T) {
	t.Parallel()

	sql, _ := dreamingPromotionStatement(&core.DreamingPromotionRequest{
		JobID:             "job_dream_workspace",
		TenantID:          "tenant_1",
		WorkspaceID:       "workspace_1",
		Tier:              core.DreamingTierLongTerm,
		MinConfidence:     0.85,
		RequireStableKind: true,
	})

	if !strings.Contains(sql, "kind IN ('fact','preference','trait','goal','constraint','decision','procedure')") {
		t.Fatalf("expected stable-kind filter, got: %s", sql)
	}
	if !strings.Contains(sql, "scope <> 'session_scratch'") {
		t.Fatalf("expected session scratch exclusion, got: %s", sql)
	}
	if !strings.Contains(sql, "FOR UPDATE SKIP LOCKED") {
		t.Fatalf("expected safe concurrent promotion locking, got: %s", sql)
	}
}

func TestValidateDreamingPromotionRequestRejectsUnsupportedTier(t *testing.T) {
	t.Parallel()

	err := validateDreamingPromotionRequest(&core.DreamingPromotionRequest{
		JobID:       "job_dream_1",
		TenantID:    "tenant_1",
		WorkspaceID: "workspace_1",
		Tier:        core.DreamingTierShortTerm,
	})
	if err == nil {
		t.Fatalf("expected unsupported short-term promotion to fail")
	}
}

```



<!-- Source: internal/store/postgres/jobs_test.go | bytes=13185 | lines=435 | sha16=35302f2b62beda48 -->

```go
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

	if err := requeueBlockedJob(context.Background(), exec, "job_blocked_1"); err != nil {
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
	if len(exec.args) != 1 || exec.args[0] != "job_blocked_1" {
		t.Fatalf("unexpected requeue args: %#v", exec.args)
	}
}

func TestRequeueBlockedJobReturnsNotFoundWhenJobIsNotBlocked(t *testing.T) {
	t.Parallel()

	exec := &recordingJobExecutor{tag: pgconn.NewCommandTag("UPDATE 0")}

	err := requeueBlockedJob(context.Background(), exec, "job_not_blocked")
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

```



<!-- Source: internal/store/postgres/memories_test.go | bytes=6018 | lines=207 | sha16=0dbecbace6b9a184 -->

```go
// ============================================================
// FILE     : internal/store/postgres/memories_test.go
// PURPOSE  : Verifies PostgreSQL memory graph helper contracts without a live database.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : postgres memory helper tests
// DEPENDS  : context, errors, strings, testing, time, internal/core, github.com/jackc/pgx/v5/pgconn
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: These tests lock update_memory supersession guards before live DB integration tests exist.
// ============================================================

package postgres

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

func TestValidateUpdateTargetRequiresSameScopeGroupAndOwner(t *testing.T) {
	t.Parallel()

	groupID := "group_1"
	tests := []struct {
		name   string
		memory *core.Memory
		target *updateTargetMemory
	}{
		{
			name: "scope mismatch",
			memory: &core.Memory{
				Scope:         core.MemoryScopeWorkspaceShared,
				OwnerEntityID: "agent:hermes-main",
			},
			target: &updateTargetMemory{
				Scope:         core.MemoryScopeAgentPrivate,
				OwnerEntityID: "agent:hermes-main",
			},
		},
		{
			name: "group mismatch",
			memory: &core.Memory{
				Scope:         core.MemoryScopeGroupShared,
				GroupID:       &groupID,
				OwnerEntityID: "agent:hermes-main",
			},
			target: &updateTargetMemory{
				Scope:         core.MemoryScopeGroupShared,
				OwnerEntityID: "agent:hermes-main",
			},
		},
		{
			name: "owner mismatch",
			memory: &core.Memory{
				Scope:         core.MemoryScopeAgentPrivate,
				OwnerEntityID: "agent:hermes-main",
			},
			target: &updateTargetMemory{
				Scope:         core.MemoryScopeAgentPrivate,
				OwnerEntityID: "agent:other",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateUpdateTarget(tt.memory, tt.target)
			if !errors.Is(err, core.ErrInvalidArgument) {
				t.Fatalf("expected ErrInvalidArgument, got %v", err)
			}
		})
	}
}

func TestValidateUpdateTargetAllowsSameScopeGroupAndOwner(t *testing.T) {
	t.Parallel()

	groupID := "group_1"
	memory := &core.Memory{
		Scope:         core.MemoryScopeGroupShared,
		GroupID:       &groupID,
		OwnerEntityID: "agent:hermes-main",
	}
	target := &updateTargetMemory{
		Scope:         core.MemoryScopeGroupShared,
		GroupID:       &groupID,
		OwnerEntityID: "agent:hermes-main",
	}

	if err := validateUpdateTarget(memory, target); err != nil {
		t.Fatalf("validateUpdateTarget returned error: %v", err)
	}
}

func TestSupersedeMemoryTargetOnlyUpdatesActiveLatestRows(t *testing.T) {
	t.Parallel()

	exec := &recordingMemoryExecutor{tag: pgconn.NewCommandTag("UPDATE 1")}
	supersededAt := time.Date(2026, time.April, 24, 13, 30, 0, 0, time.UTC)

	if err := supersedeMemoryTarget(context.Background(), exec, "mem_old", supersededAt); err != nil {
		t.Fatalf("supersedeMemoryTarget returned error: %v", err)
	}

	if !strings.Contains(exec.sql, "status = $2") || !strings.Contains(exec.sql, "latest_flag = false") {
		t.Fatalf("expected supersession update fields, got: %s", exec.sql)
	}
	if !strings.Contains(exec.sql, "AND status = $4") || !strings.Contains(exec.sql, "AND latest_flag = true") {
		t.Fatalf("expected active/latest guard, got: %s", exec.sql)
	}
	if len(exec.args) != 4 {
		t.Fatalf("unexpected supersede args: %#v", exec.args)
	}
	if exec.args[0] != "mem_old" || exec.args[1] != core.MemoryStatusSuperseded || exec.args[3] != core.MemoryStatusActive {
		t.Fatalf("unexpected supersede args: %#v", exec.args)
	}
	if got := exec.args[2].(time.Time); !got.Equal(supersededAt) {
		t.Fatalf("unexpected superseded timestamp: got %s want %s", got, supersededAt)
	}
}

func TestSupersedeMemoryTargetReturnsConflictWhenNoLatestRowChanged(t *testing.T) {
	t.Parallel()

	exec := &recordingMemoryExecutor{tag: pgconn.NewCommandTag("UPDATE 0")}

	err := supersedeMemoryTarget(context.Background(), exec, "mem_old", time.Now().UTC())
	if !errors.Is(err, core.ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

func TestExplainMemoryTraceStatementScopesMemoryToTenantWorkspace(t *testing.T) {
	t.Parallel()

	sql := explainMemoryTraceStatement()

	for _, want := range []string{
		"FROM memory_trace mt",
		"JOIN memories m ON m.id = mt.memory_id",
		"mt.memory_id = $1",
		"m.tenant_id = $2",
		"m.workspace_id = $3",
		"m.scope <> 'agent_private'",
		"m.owner_entity_id = $4",
		"m.scope <> 'group_shared'",
		"m.group_id = ANY($5)",
	} {
		if !strings.Contains(sql, want) {
			t.Fatalf("explain memory trace query must preserve %q, got:\n%s", want, sql)
		}
	}
}

func TestExplainMemoryProvenanceQueriesScopeEvidenceToTenantWorkspace(t *testing.T) {
	t.Parallel()

	source := readPostgresSourceFile(t, "memories.go")
	eventSource := extractPostgresSourceBetween(t, source, "func (s *Store) provenanceEvents", "func (s *Store) provenanceDocuments")
	documentSource := extractPostgresSourceBetween(t, source, "func (s *Store) provenanceDocuments", "func nullIfEmpty")

	for _, tt := range []struct {
		name   string
		source string
	}{
		{name: "events", source: eventSource},
		{name: "documents", source: documentSource},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			for _, want := range []string{
				"tenant_id = $1",
				"workspace_id = $2",
				"id = ANY($3)",
			} {
				if !strings.Contains(tt.source, want) {
					t.Fatalf("provenance %s query must preserve %q, got:\n%s", tt.name, want, tt.source)
				}
			}
		})
	}
}

type recordingMemoryExecutor struct {
	sql  string
	args []any
	tag  pgconn.CommandTag
	err  error
}

func (e *recordingMemoryExecutor) Exec(_ context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	e.sql = sql
	e.args = append([]any(nil), arguments...)
	return e.tag, e.err
}

```



<!-- Source: internal/store/postgres/notes_plans_test.go | bytes=3881 | lines=122 | sha16=947b68124da0336d -->

```go
// ============================================================
// FILE     : internal/store/postgres/notes_plans_test.go
// PURPOSE  : Verifies note and plan lookup SQL preserves actor-private visibility boundaries.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : postgres note and plan statement tests
// DEPENDS  : strings, testing, internal/core
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: Keep pinned note and active plan retrieval tenant- and actor-scoped.
// ============================================================

package postgres

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

func TestListPinnedNotesStatementScopesAgentPrivateToOwner(t *testing.T) {
	t.Parallel()

	sql, args := listPinnedNotesStatement(&core.ListPinnedNotesRequest{
		TenantID:      "tenant_1",
		WorkspaceID:   "workspace_1",
		OwnerEntityID: "agent:hermes-main",
		Scopes:        []core.MemoryScope{core.MemoryScopeAgentPrivate, core.MemoryScopeWorkspaceShared},
	})

	if !strings.Contains(sql, "tenant_id = $1") {
		t.Fatalf("expected tenant predicate for pinned notes, got:\n%s", sql)
	}
	if !strings.Contains(sql, "scope <> 'agent_private'") {
		t.Fatalf("expected non-private note scopes to bypass owner filter, got:\n%s", sql)
	}
	if !strings.Contains(sql, "owner_entity_id = $4") {
		t.Fatalf("expected agent_private note owner predicate, got:\n%s", sql)
	}
	if len(args) != 4 || args[3] != "agent:hermes-main" {
		t.Fatalf("unexpected pinned notes args: %#v", args)
	}
}

func readCurrentFile(t *testing.T, name string) string {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not locate current test file")
	}
	data, err := os.ReadFile(filepath.Join(filepath.Dir(currentFile), name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return string(data)
}

func extractSourceBetween(t *testing.T, source, startMarker, endMarker string) string {
	t.Helper()

	start := strings.Index(source, startMarker)
	if start < 0 {
		t.Fatalf("missing start marker %q", startMarker)
	}
	remainder := source[start:]
	end := strings.Index(remainder, endMarker)
	if end < 0 {
		t.Fatalf("missing end marker %q", endMarker)
	}
	return remainder[:end]
}

func TestGetActivePlansStatementScopesAgentPrivateToOwner(t *testing.T) {
	t.Parallel()

	sql, args := getActivePlansStatement(&core.GetActivePlansRequest{
		TenantID:      "tenant_1",
		WorkspaceID:   "workspace_1",
		OwnerEntityID: "agent:hermes-main",
		Scopes:        []core.MemoryScope{core.MemoryScopeAgentPrivate, core.MemoryScopeWorkspaceShared},
	})

	if !strings.Contains(sql, "tenant_id = $1") {
		t.Fatalf("expected tenant predicate for active plans, got:\n%s", sql)
	}
	if !strings.Contains(sql, "scope <> 'agent_private'") {
		t.Fatalf("expected non-private plan scopes to bypass owner filter, got:\n%s", sql)
	}
	if !strings.Contains(sql, "owner_entity_id = $4") {
		t.Fatalf("expected agent_private plan owner predicate, got:\n%s", sql)
	}
	if len(args) != 4 || args[3] != "agent:hermes-main" {
		t.Fatalf("unexpected active plans args: %#v", args)
	}
}

func TestUpdatePlanSourceUsesTenantWorkspaceAndPatchSemantics(t *testing.T) {
	t.Parallel()

	source := readCurrentFile(t, "notes_plans.go")
	updateSource := extractSourceBetween(t, source, "func (s *Store) UpdatePlan", "// GetActivePlans loads active plans for recall.")

	for _, want := range []string{
		"COALESCE(NULLIF($2, ''), title)",
		"COALESCE(NULLIF($3, ''), status)",
		"COALESCE($4, evidence_json)",
		"AND tenant_id = $5",
		"AND workspace_id = $6",
		"if items != nil",
	} {
		if !strings.Contains(updateSource, want) {
			t.Fatalf("UpdatePlan must preserve %q, got:\n%s", want, updateSource)
		}
	}
}

```



<!-- Source: internal/store/postgres/search_test.go | bytes=2764 | lines=76 | sha16=fa35492f854bd3eb -->

```go
// ============================================================
// FILE     : internal/store/postgres/search_test.go
// PURPOSE  : Verifies PostgreSQL search SQL preserves actor-private visibility boundaries.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : postgres search statement tests
// DEPENDS  : strings, testing, internal/core
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: These tests lock privacy predicates without requiring a live database.
// ============================================================

package postgres

import (
	"strings"
	"testing"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

func TestSearchMemoriesStatementScopesAgentPrivateToOwner(t *testing.T) {
	t.Parallel()

	sql, args := searchMemoriesStatement(&core.SearchMemoriesRequest{
		TenantID:        "tenant_1",
		WorkspaceID:     "workspace_1",
		OwnerEntityID:   "agent:hermes-main",
		Query:           "stage 2",
		Scopes:          []core.MemoryScope{core.MemoryScopeAgentPrivate, core.MemoryScopeWorkspaceShared},
		ArtifactClasses: []core.ArtifactClass{core.ArtifactClassKnowledge},
	})

	if !strings.Contains(sql, "scope <> 'agent_private'") {
		t.Fatalf("expected non-private scopes to bypass owner filter, got:\n%s", sql)
	}
	if !strings.Contains(sql, "owner_entity_id = $6") {
		t.Fatalf("expected agent_private owner predicate, got:\n%s", sql)
	}
	if !strings.Contains(sql, "group_id, owner_entity_id, valid_from") {
		t.Fatalf("memory search must return owner_entity_id for caller-side visibility checks, got:\n%s", sql)
	}
	if len(args) != 7 || args[5] != "agent:hermes-main" {
		t.Fatalf("unexpected memory search args: %#v", args)
	}
}

func TestSearchMemoriesStatementScopesGroupSharedToMemberships(t *testing.T) {
	t.Parallel()

	sql, args := searchMemoriesStatement(&core.SearchMemoriesRequest{
		TenantID:        "tenant_1",
		WorkspaceID:     "workspace_1",
		OwnerEntityID:   "agent:hermes-main",
		VisibleGroupIDs: []string{"group_design"},
		Query:           "stage 2",
		Scopes:          []core.MemoryScope{core.MemoryScopeGroupShared},
		ArtifactClasses: []core.ArtifactClass{core.ArtifactClassKnowledge},
	})

	if !strings.Contains(sql, "scope <> 'group_shared'") {
		t.Fatalf("expected non-group scopes to bypass group filter, got:\n%s", sql)
	}
	if !strings.Contains(sql, "group_id = ANY($7)") {
		t.Fatalf("expected group_shared membership predicate, got:\n%s", sql)
	}
	if len(args) != 7 {
		t.Fatalf("unexpected arg count: %#v", args)
	}
	gotGroups, ok := args[6].([]string)
	if !ok || len(gotGroups) != 1 || gotGroups[0] != "group_design" {
		t.Fatalf("unexpected group args: %#v", args[6])
	}
}

```



<!-- Source: internal/store/postgres/timeline_test.go | bytes=2386 | lines=79 | sha16=6a9c29e82c10aded -->

```go
// ============================================================
// FILE     : internal/store/postgres/timeline_test.go
// PURPOSE  : Verifies timeline query source preserves read-only scope filtering.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : postgres timeline source tests
// DEPENDS  : strings, testing
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: Timeline tests lock read-only query shape without requiring a live database.
// ============================================================

package postgres

import (
	"strings"
	"testing"
)

func TestTimelineStatementReadsMemoriesAndCorrections(t *testing.T) {
	t.Parallel()

	source := readPostgresSourceFile(t, "timeline.go")
	statementSource := extractPostgresSourceBetween(t, source, "func timelineStatement", "`, []any")

	for _, want := range []string{
		"FROM memories m",
		"LEFT JOIN memory_trace mt ON mt.memory_id = m.id",
		"FROM memory_corrections c",
		"JOIN memories m ON m.id = c.memory_id",
		"UNION ALL",
		"ORDER BY occurred_at DESC, id DESC",
	} {
		if !strings.Contains(statementSource, want) {
			t.Fatalf("timeline statement must preserve %q, got:\n%s", want, statementSource)
		}
	}
}

func TestTimelineStatementPreservesScopeAndOwnerFiltering(t *testing.T) {
	t.Parallel()

	source := readPostgresSourceFile(t, "timeline.go")
	statementSource := extractPostgresSourceBetween(t, source, "func timelineStatement", "`, []any")

	for _, want := range []string{
		"m.tenant_id = $1",
		"m.workspace_id = $2",
		"m.scope = ANY($3)",
		"m.scope <> 'group_shared'",
		"m.scope <> 'agent_private'",
		"m.owner_entity_id = $6",
	} {
		if !strings.Contains(statementSource, want) {
			t.Fatalf("timeline statement must preserve %q, got:\n%s", want, statementSource)
		}
	}
}

func TestTimelineStatementDoesNotMutateGraphState(t *testing.T) {
	t.Parallel()

	source := readPostgresSourceFile(t, "timeline.go")
	statementSource := extractPostgresSourceBetween(t, source, "func timelineStatement", "`, []any")

	for _, blocked := range []string{
		"UPDATE ",
		"INSERT ",
		"DELETE ",
		"latest_flag =",
		"memory_edges",
	} {
		if strings.Contains(statementSource, blocked) {
			t.Fatalf("timeline statement must stay read-only; found %q in:\n%s", blocked, statementSource)
		}
	}
}

```



<!-- Source: internal/worker/processor_test.go | bytes=24166 | lines=749 | sha16=5f81fa4269df391a -->

```go
// ============================================================
// FILE     : internal/worker/processor_test.go
// PURPOSE  : Verifies worker job claiming, dispatch, completion, and retry handoff.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : TestProcessorRunOnce_ClaimsOneJob, TestProcessorRunOnce_UnknownJobKindFailsSafely, worker processor tests
// DEPENDS  : internal/worker, internal/core, internal/graph, internal/reasoning
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: These tests lock the pipeline skeleton only; they must not assert memory extraction quality.
// ============================================================

package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/graph"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/reasoning"
)

func TestProcessorRunOnce_ClaimsOneJob(t *testing.T) {
	t.Parallel()

	job := testProcessTurnJob()
	jobs := &fakeWorkerJobStore{claimed: []*core.IngestJob{job}}
	rawEvents := fakeRawEventsForJob(job)
	processor := newTestProcessor(t, jobs, rawEvents, &fakeReasoner{}, &fakeApplyEngine{})

	result, err := processor.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce returned error: %v", err)
	}

	if jobs.claimWorkerID != "worker:test" {
		t.Fatalf("unexpected worker ID: %s", jobs.claimWorkerID)
	}
	if jobs.claimLimit != 1 {
		t.Fatalf("expected default claim limit 1, got %d", jobs.claimLimit)
	}
	if result.Claimed != 1 || result.Completed != 1 || result.Failed != 0 {
		t.Fatalf("unexpected run result: %#v", result)
	}
}

func TestProcessorRunOnce_UnknownJobKindFailsSafely(t *testing.T) {
	t.Parallel()

	job := testProcessTurnJob()
	job.JobKind = core.JobKind("unknown_job")
	jobs := &fakeWorkerJobStore{claimed: []*core.IngestJob{job}}
	rawEvents := fakeRawEventsForJob(job)
	processor := newTestProcessor(t, jobs, rawEvents, &fakeReasoner{}, &fakeApplyEngine{})

	result, err := processor.RunOnce(context.Background())
	if err == nil {
		t.Fatalf("expected RunOnce to report unsupported job kind")
	}

	if result.Claimed != 1 || result.Completed != 0 || result.Failed != 1 {
		t.Fatalf("unexpected run result: %#v", result)
	}
	if len(jobs.completed) != 0 {
		t.Fatalf("expected no completed jobs, got %v", jobs.completed)
	}
	if len(jobs.failed) != 1 {
		t.Fatalf("expected one failed job, got %d", len(jobs.failed))
	}
	if !strings.Contains(jobs.failed[0].err.Error(), "unsupported job kind") {
		t.Fatalf("unexpected failure error: %v", jobs.failed[0].err)
	}
	if len(rawEvents.requestedIDs) != 0 {
		t.Fatalf("unknown job kind should not load raw events, got %v", rawEvents.requestedIDs)
	}
}

func TestProcessorProcessTurnEvent_LoadsRawEvents(t *testing.T) {
	t.Parallel()

	job := testProcessTurnJob()
	jobs := &fakeWorkerJobStore{claimed: []*core.IngestJob{job}}
	rawEvents := fakeRawEventsForJob(job)
	reasoner := &fakeReasoner{}
	processor := newTestProcessor(t, jobs, rawEvents, reasoner, &fakeApplyEngine{})

	if _, err := processor.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce returned error: %v", err)
	}

	if !reflect.DeepEqual(rawEvents.requestedIDs, job.RawEventIDs) {
		t.Fatalf("unexpected raw event IDs: got %v want %v", rawEvents.requestedIDs, job.RawEventIDs)
	}
	if reasoner.received == nil {
		t.Fatalf("expected reasoner to receive an envelope")
	}
	if len(reasoner.received.RawEvents) != len(job.RawEventIDs) {
		t.Fatalf("expected %d raw events, got %d", len(job.RawEventIDs), len(reasoner.received.RawEvents))
	}
	if reasoner.received.Stage2.RequiredOutputName != reasoning.StageNameResolve {
		t.Fatalf("unexpected stage 2 output contract: %s", reasoner.received.Stage2.RequiredOutputName)
	}
	if reasoner.received.Stage2.RequiredOutputSchema != reasoning.Stage2ResolveOutputSchemaV0 {
		t.Fatalf("unexpected stage 2 output schema: %s", reasoner.received.Stage2.RequiredOutputSchema)
	}
}

func TestProcessorProcessTurnEvent_PassesStageEnvelopeToReasoner(t *testing.T) {
	t.Parallel()

	job := testProcessTurnJob()
	jobs := &fakeWorkerJobStore{claimed: []*core.IngestJob{job}}
	rawEvents := fakeRawEventsForJob(job)
	reasoner := &fakeReasoner{}
	processor := newTestProcessor(t, jobs, rawEvents, reasoner, &fakeApplyEngine{})

	if _, err := processor.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce returned error: %v", err)
	}

	if reasoner.received == nil {
		t.Fatalf("expected reasoner to receive an envelope")
	}
	if reasoner.received.Stage1.JobID != job.ID {
		t.Fatalf("expected Stage 1 envelope for job %s, got %#v", job.ID, reasoner.received.Stage1)
	}
	stage2 := reasoner.received.Stage2
	if stage2.RequiredOutputSchema != reasoning.Stage2ResolveOutputSchemaV0 {
		t.Fatalf("expected required output schema %q, got %q", reasoning.Stage2ResolveOutputSchemaV0, stage2.RequiredOutputSchema)
	}
	if len(stage2.RelevantMemories) != 0 {
		t.Fatalf("worker should leave retrieved Stage 2 context to the reasoner, got %#v", stage2.RelevantMemories)
	}
}

func TestProcessorRunOnce_CompletesJobOnSuccess(t *testing.T) {
	t.Parallel()

	job := testProcessTurnJob()
	jobs := &fakeWorkerJobStore{claimed: []*core.IngestJob{job}}
	applyEngine := &fakeApplyEngine{}
	processor := newTestProcessor(t, jobs, fakeRawEventsForJob(job), &fakeReasoner{}, applyEngine)

	result, err := processor.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce returned error: %v", err)
	}

	if !reflect.DeepEqual(jobs.completed, []string{job.ID}) {
		t.Fatalf("unexpected completed jobs: %v", jobs.completed)
	}
	if len(jobs.failed) != 0 {
		t.Fatalf("expected no failed jobs, got %d", len(jobs.failed))
	}
	if applyEngine.received == nil || applyEngine.received.JobID != job.ID {
		t.Fatalf("expected apply request for job %s, got %#v", job.ID, applyEngine.received)
	}
	if result.Completed != 1 {
		t.Fatalf("expected one completed job, got %#v", result)
	}
}

func TestProcessorRunOnce_FailsJobOnReasoningOrApplyError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		reasoner  reasoning.Orchestrator
		apply     graph.ApplyEngine
		wantError string
	}{
		{
			name:      "reasoning",
			reasoner:  &fakeReasoner{err: errors.New("codex bridge unavailable")},
			apply:     &fakeApplyEngine{},
			wantError: "reason process_turn_event",
		},
		{
			name:      "apply",
			reasoner:  &fakeReasoner{},
			apply:     &fakeApplyEngine{err: errors.New("apply validation failed")},
			wantError: "apply process_turn_event",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			job := testProcessTurnJob()
			jobs := &fakeWorkerJobStore{claimed: []*core.IngestJob{job}}
			processor := newTestProcessor(t, jobs, fakeRawEventsForJob(job), tt.reasoner, tt.apply)

			result, err := processor.RunOnce(context.Background())
			if err == nil {
				t.Fatalf("expected RunOnce to return an error")
			}

			if !strings.Contains(err.Error(), tt.wantError) {
				t.Fatalf("expected error to contain %q, got %v", tt.wantError, err)
			}
			if result.Claimed != 1 || result.Completed != 0 || result.Failed != 1 {
				t.Fatalf("unexpected run result: %#v", result)
			}
			if len(jobs.completed) != 0 {
				t.Fatalf("expected no completed jobs, got %v", jobs.completed)
			}
			if len(jobs.failed) != 1 || jobs.failed[0].jobID != job.ID {
				t.Fatalf("unexpected failed jobs: %#v", jobs.failed)
			}
		})
	}
}

func TestProcessorRunOnce_BlocksJobWhenApplyNotImplemented(t *testing.T) {
	t.Parallel()

	job := testProcessTurnJob()
	jobs := &fakeWorkerJobStore{claimed: []*core.IngestJob{job}}
	applyErr := fmt.Errorf("%w: operations[0].kind %q is validation-only in store-backed apply", core.ErrNotImplemented, reasoning.OperationKindUpdateMemory)
	applyEngine := &fakeApplyEngine{err: applyErr}
	processor := newTestProcessor(t, jobs, fakeRawEventsForJob(job), &fakeReasoner{}, applyEngine)

	result, err := processor.RunOnce(context.Background())
	if err == nil {
		t.Fatalf("expected RunOnce to return an error")
	}
	if !errors.Is(err, core.ErrNotImplemented) {
		t.Fatalf("expected RunOnce error to wrap ErrNotImplemented, got %v", err)
	}
	if len(jobs.completed) != 0 {
		t.Fatalf("expected unsupported apply work not to complete the job, got %v", jobs.completed)
	}
	if len(jobs.failed) != 0 {
		t.Fatalf("expected unsupported apply work not to use retry failure path, got %#v", jobs.failed)
	}
	if len(jobs.blocked) != 1 || !errors.Is(jobs.blocked[0].err, core.ErrNotImplemented) {
		t.Fatalf("expected blocked job to record ErrNotImplemented, got %#v", jobs.blocked)
	}
	if !strings.Contains(jobs.blocked[0].err.Error(), "unsupported apply work") {
		t.Fatalf("expected blocked record to explain unsupported apply work, got %v", jobs.blocked[0].err)
	}
	if result.Claimed != 1 || result.Completed != 0 || result.Failed != 0 || result.Blocked != 1 {
		t.Fatalf("unexpected run result: %#v", result)
	}
	if len(result.Failures) != 1 || result.Failures[0].JobID != job.ID {
		t.Fatalf("expected per-job failure report for %s, got %#v", job.ID, result.Failures)
	}
	if !strings.Contains(result.Failures[0].Error, "unsupported apply work") {
		t.Fatalf("expected failure report to include unsupported apply work, got %#v", result.Failures[0])
	}
}

func TestProcessorRunOnce_RejectsIncompleteRawEventBundleBeforeReasoning(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		mutate      func(*fakeWorkerRawEventStore)
		wantWrapped error
	}{
		{
			name: "missing requested event",
			mutate: func(rawEvents *fakeWorkerRawEventStore) {
				delete(rawEvents.events, "evt_2")
			},
			wantWrapped: core.ErrNotFound,
		},
		{
			name: "returned event id mismatch",
			mutate: func(rawEvents *fakeWorkerRawEventStore) {
				rawEvents.events["evt_2"].ID = "evt_unrequested"
			},
			wantWrapped: core.ErrConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			job := testProcessTurnJob()
			jobs := &fakeWorkerJobStore{claimed: []*core.IngestJob{job}}
			rawEvents := fakeRawEventsForJob(job)
			tt.mutate(rawEvents)
			reasoner := &fakeReasoner{}
			applyEngine := &fakeApplyEngine{}
			processor := newTestProcessor(t, jobs, rawEvents, reasoner, applyEngine)

			result, err := processor.RunOnce(context.Background())
			if err == nil {
				t.Fatalf("expected RunOnce to return an error")
			}
			if !errors.Is(err, tt.wantWrapped) {
				t.Fatalf("expected error to wrap %v, got %v", tt.wantWrapped, err)
			}
			if !strings.Contains(err.Error(), "raw event bundle") {
				t.Fatalf("expected raw event bundle error, got %v", err)
			}
			if reasoner.received != nil {
				t.Fatalf("reasoner should not run for incomplete raw event bundle")
			}
			if applyEngine.received != nil {
				t.Fatalf("apply should not run for incomplete raw event bundle")
			}
			if len(jobs.completed) != 0 {
				t.Fatalf("expected incomplete raw event bundle not to complete the job, got %v", jobs.completed)
			}
			if len(jobs.failed) != 1 {
				t.Fatalf("expected one failed job, got %#v", jobs.failed)
			}
			if result.Claimed != 1 || result.Completed != 0 || result.Failed != 1 {
				t.Fatalf("unexpected run result: %#v", result)
			}
		})
	}
}

func TestProcessorRunOnce_RejectsAmbiguousRawEventActorBeforeStage2Sources(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		mutate      func(*fakeWorkerRawEventStore)
		wantWrapped error
		wantError   string
	}{
		{
			name: "empty actor",
			mutate: func(rawEvents *fakeWorkerRawEventStore) {
				rawEvents.events["evt_2"].ActorID = ""
			},
			wantWrapped: core.ErrInvalidArgument,
			wantError:   "empty actor_id",
		},
		{
			name: "mixed actors",
			mutate: func(rawEvents *fakeWorkerRawEventStore) {
				rawEvents.events["evt_2"].ActorID = "agent:other"
			},
			wantWrapped: core.ErrConflict,
			wantError:   "mixed actor_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			job := testProcessTurnJob()
			jobs := &fakeWorkerJobStore{claimed: []*core.IngestJob{job}}
			rawEvents := fakeRawEventsForJob(job)
			tt.mutate(rawEvents)
			reasoner := &fakeReasoner{}
			applyEngine := &fakeApplyEngine{}
			processor := newTestProcessor(t, jobs, rawEvents, reasoner, applyEngine)

			result, err := processor.RunOnce(context.Background())
			if err == nil {
				t.Fatalf("expected RunOnce to reject ambiguous raw event actors")
			}
			if !errors.Is(err, tt.wantWrapped) {
				t.Fatalf("expected error to wrap %v, got %v", tt.wantWrapped, err)
			}
			if !strings.Contains(err.Error(), tt.wantError) {
				t.Fatalf("expected error to contain %q, got %v", tt.wantError, err)
			}
			if reasoner.received != nil {
				t.Fatalf("reasoner should not run for ambiguous raw event actors")
			}
			if applyEngine.received != nil {
				t.Fatalf("apply should not run for ambiguous raw event actors")
			}
			if len(jobs.completed) != 0 {
				t.Fatalf("expected ambiguous actor bundle not to complete the job, got %v", jobs.completed)
			}
			if len(jobs.failed) != 1 {
				t.Fatalf("expected one failed job, got %#v", jobs.failed)
			}
			if result.Claimed != 1 || result.Completed != 0 || result.Failed != 1 {
				t.Fatalf("unexpected run result: %#v", result)
			}
		})
	}
}

func TestProcessorRunOnce_ReportsAppliedOperationCounts(t *testing.T) {
	t.Parallel()

	job := testProcessTurnJob()
	jobs := &fakeWorkerJobStore{claimed: []*core.IngestJob{job}}
	applyEngine := &fakeApplyEngine{result: &graph.ApplyResult{
		AppliedOperationCount: 2,
		MemoryIDs:             []string{"mem_1", "mem_2"},
		TraceWritten:          true,
	}}
	processor := newTestProcessor(t, jobs, fakeRawEventsForJob(job), &fakeReasoner{}, applyEngine)

	result, err := processor.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce returned error: %v", err)
	}

	if result.AppliedOperationCount != 2 {
		t.Fatalf("unexpected applied operation count: got %d want 2", result.AppliedOperationCount)
	}
	if result.MemoryIDCount != 2 {
		t.Fatalf("unexpected memory id count: got %d want 2", result.MemoryIDCount)
	}
	if result.TraceWrittenCount != 1 {
		t.Fatalf("unexpected trace written count: got %d want 1", result.TraceWrittenCount)
	}
}

func TestProcessorRunOnce_ProcessesDreamSessionJob(t *testing.T) {
	t.Parallel()

	job := testProcessTurnJob()
	job.JobKind = core.JobKindDreamSession
	job.RawEventIDs = nil
	job.PayloadJSON = json.RawMessage(`{"session_id":"session_1"}`)
	jobs := &fakeWorkerJobStore{claimed: []*core.IngestJob{job}}
	dreamingStore := &fakeDreamingStore{
		input: &core.DreamingSessionInput{
			RawEventIDs: []string{"evt_1"},
			Memories: []*core.Memory{{
				ID:         "mem_1",
				Text:       "User prefers contract-first implementation.",
				Confidence: 0.9,
			}},
		},
		promotions: map[core.DreamingTier]*core.DreamingPromotionResult{
			core.DreamingTierMidTerm:       {PromotedCount: 1, MemoryIDs: []string{"mem_1"}},
			core.DreamingTierLongTerm:      {PromotedCount: 1, MemoryIDs: []string{"mem_1"}},
			core.DreamingTierUltraLongTerm: {PromotedCount: 0},
		},
	}
	processor := newTestProcessorWithDreaming(t, jobs, fakeRawEventsForJob(testProcessTurnJob()), &fakeReasoner{}, &fakeApplyEngine{}, dreamingStore)

	result, err := processor.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce returned error: %v", err)
	}

	if result.Completed != 1 || result.SessionDreamCount != 2 {
		t.Fatalf("unexpected dream session result: %#v", result)
	}
	if dreamingStore.summary == nil || dreamingStore.summary.SessionID != "session_1" {
		t.Fatalf("expected session summary to be written, got %#v", dreamingStore.summary)
	}
	if len(dreamingStore.requests) == 0 || dreamingStore.requests[0].Tier != core.DreamingTierMidTerm {
		t.Fatalf("expected mid-term promotion request, got %#v", dreamingStore.requests)
	}
}

func TestProcessorRunOnce_ProcessesDreamWorkspaceJob(t *testing.T) {
	t.Parallel()

	job := testProcessTurnJob()
	job.JobKind = core.JobKindDreamWorkspace
	job.RawEventIDs = nil
	job.PayloadJSON = json.RawMessage(`{}`)
	jobs := &fakeWorkerJobStore{claimed: []*core.IngestJob{job}}
	dreamingStore := &fakeDreamingStore{
		promotions: map[core.DreamingTier]*core.DreamingPromotionResult{
			core.DreamingTierLongTerm:      {PromotedCount: 2, MemoryIDs: []string{"mem_1", "mem_2"}},
			core.DreamingTierUltraLongTerm: {PromotedCount: 1, MemoryIDs: []string{"mem_3"}},
		},
	}
	processor := newTestProcessorWithDreaming(t, jobs, fakeRawEventsForJob(testProcessTurnJob()), &fakeReasoner{}, &fakeApplyEngine{}, dreamingStore)

	result, err := processor.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce returned error: %v", err)
	}

	if result.Completed != 1 || result.WorkspaceDreamCount != 3 {
		t.Fatalf("unexpected dream workspace result: %#v", result)
	}
	if len(dreamingStore.requests) != 2 {
		t.Fatalf("expected long and ultra-long promotion requests, got %#v", dreamingStore.requests)
	}
}

func newTestProcessor(
	t *testing.T,
	jobs *fakeWorkerJobStore,
	rawEvents *fakeWorkerRawEventStore,
	reasoner reasoning.Orchestrator,
	applyEngine graph.ApplyEngine,
) *Processor {
	t.Helper()
	processor, err := NewProcessor(Dependencies{
		WorkerID:    "worker:test",
		Jobs:        jobs,
		RawEvents:   rawEvents,
		Reasoner:    reasoner,
		ApplyEngine: applyEngine,
		Dreaming:    newTestDreamingService(t, &fakeDreamingStore{}),
	})
	if err != nil {
		t.Fatalf("NewProcessor returned error: %v", err)
	}
	return processor
}

func newTestProcessorWithDreaming(
	t *testing.T,
	jobs *fakeWorkerJobStore,
	rawEvents *fakeWorkerRawEventStore,
	reasoner reasoning.Orchestrator,
	applyEngine graph.ApplyEngine,
	dreamingStore *fakeDreamingStore,
) *Processor {
	t.Helper()
	processor, err := NewProcessor(Dependencies{
		WorkerID:    "worker:test",
		Jobs:        jobs,
		RawEvents:   rawEvents,
		Reasoner:    reasoner,
		ApplyEngine: applyEngine,
		Dreaming:    newTestDreamingService(t, dreamingStore),
		Clock: func() time.Time {
			return time.Date(2026, time.April, 24, 1, 0, 0, 0, time.UTC)
		},
	})
	if err != nil {
		t.Fatalf("NewProcessor returned error: %v", err)
	}
	return processor
}

func newTestDreamingService(t *testing.T, store *fakeDreamingStore) *graph.DreamingService {
	t.Helper()
	service, err := graph.NewDreamingService(graph.DreamingDependencies{
		Store: store,
		Clock: func() time.Time {
			return time.Date(2026, time.April, 24, 1, 0, 0, 0, time.UTC)
		},
	})
	if err != nil {
		t.Fatalf("NewDreamingService returned error: %v", err)
	}
	return service
}

func testProcessTurnJob() *core.IngestJob {
	return &core.IngestJob{
		ID:          "job_1",
		TenantID:    "tenant_1",
		WorkspaceID: "workspace_1",
		JobKind:     core.JobKindProcessTurnEvent,
		Status:      "running",
		RawEventIDs: []string{"evt_1", "evt_2"},
		PayloadJSON: json.RawMessage(`{"session_id":"session_1"}`),
		CreatedAt:   time.Date(2026, time.April, 24, 0, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, time.April, 24, 0, 0, 0, 0, time.UTC),
	}
}

func fakeRawEventsForJob(job *core.IngestJob) *fakeWorkerRawEventStore {
	events := make(map[string]*core.RawEvent, len(job.RawEventIDs))
	for i, eventID := range job.RawEventIDs {
		events[eventID] = &core.RawEvent{
			ID:          eventID,
			TenantID:    job.TenantID,
			WorkspaceID: job.WorkspaceID,
			SessionID:   "session_1",
			ActorID:     "agent:hermes-main",
			EventKind:   "message",
			Source:      "hermes",
			Fingerprint: "fp_" + eventID,
			OccurredAt:  time.Date(2026, time.April, 24, 0, i, 0, 0, time.UTC),
			PayloadJSON: json.RawMessage(`{"text":"hello"}`),
			CreatedAt:   time.Date(2026, time.April, 24, 0, i, 0, 0, time.UTC),
		}
	}
	return &fakeWorkerRawEventStore{events: events}
}

type failedJob struct {
	jobID string
	err   error
}

type fakeWorkerJobStore struct {
	claimed       []*core.IngestJob
	claimWorkerID string
	claimLimit    int
	completed     []string
	failed        []failedJob
	blocked       []failedJob
}

func (s *fakeWorkerJobStore) EnqueueJobs(context.Context, []*core.IngestJob) ([]string, error) {
	return nil, core.ErrNotImplemented
}

func (s *fakeWorkerJobStore) ClaimJobs(_ context.Context, workerID string, limit int) ([]*core.IngestJob, error) {
	s.claimWorkerID = workerID
	s.claimLimit = limit
	return s.claimed, nil
}

func (s *fakeWorkerJobStore) CompleteJob(_ context.Context, jobID string) error {
	s.completed = append(s.completed, jobID)
	return nil
}

func (s *fakeWorkerJobStore) FailJob(_ context.Context, jobID string, err error) error {
	s.failed = append(s.failed, failedJob{
		jobID: jobID,
		err:   err,
	})
	return nil
}

func (s *fakeWorkerJobStore) BlockJob(_ context.Context, jobID string, err error) error {
	s.blocked = append(s.blocked, failedJob{
		jobID: jobID,
		err:   err,
	})
	return nil
}

type fakeWorkerRawEventStore struct {
	events       map[string]*core.RawEvent
	requestedIDs []string
	err          error
}

func (s *fakeWorkerRawEventStore) AppendRawEvents(context.Context, []*core.RawEvent) ([]string, error) {
	return nil, core.ErrNotImplemented
}

func (s *fakeWorkerRawEventStore) GetRawEvents(_ context.Context, ids []string) ([]*core.RawEvent, error) {
	s.requestedIDs = append([]string(nil), ids...)
	if s.err != nil {
		return nil, s.err
	}
	events := make([]*core.RawEvent, 0, len(ids))
	for _, id := range ids {
		if event, ok := s.events[id]; ok {
			events = append(events, event)
		}
	}
	return events, nil
}

type fakeDreamingStore struct {
	input      *core.DreamingSessionInput
	summary    *core.SessionSummary
	promotions map[core.DreamingTier]*core.DreamingPromotionResult
	requests   []*core.DreamingPromotionRequest
}

func (s *fakeDreamingStore) LoadDreamingSessionInput(context.Context, *core.DreamSessionRequest) (*core.DreamingSessionInput, error) {
	if s.input != nil {
		return s.input, nil
	}
	return &core.DreamingSessionInput{}, nil
}

func (s *fakeDreamingStore) PromoteMemories(_ context.Context, req *core.DreamingPromotionRequest) (*core.DreamingPromotionResult, error) {
	s.requests = append(s.requests, req)
	if s.promotions != nil {
		if result, ok := s.promotions[req.Tier]; ok {
			return result, nil
		}
	}
	return &core.DreamingPromotionResult{}, nil
}

func (s *fakeDreamingStore) UpsertSessionSummary(_ context.Context, summary *core.SessionSummary) error {
	s.summary = summary
	return nil
}

func (s *fakeDreamingStore) GetSessionSummary(context.Context, string) (*core.SessionSummary, error) {
	return nil, core.ErrNotFound
}

type fakeReasoner struct {
	received *reasoning.ProcessTurnEnvelope
	err      error
	result   *reasoning.ProcessTurnResult
}

func (s *fakeReasoner) ProcessTurn(_ context.Context, envelope *reasoning.ProcessTurnEnvelope) (*reasoning.ProcessTurnResult, error) {
	s.received = envelope
	if s.err != nil {
		return nil, s.err
	}
	if s.result != nil {
		return s.result, nil
	}
	return successfulReasoningResult(), nil
}

type fakeApplyEngine struct {
	received *graph.ApplyRequest
	err      error
	result   *graph.ApplyResult
}

func (s *fakeApplyEngine) Apply(_ context.Context, req *graph.ApplyRequest) (*graph.ApplyResult, error) {
	s.received = req
	if s.err != nil {
		return nil, s.err
	}
	if s.result != nil {
		return s.result, nil
	}
	return &graph.ApplyResult{
		AppliedOperationCount: 0,
		MemoryIDs:             []string{},
		TraceWritten:          false,
	}, nil
}

func successfulReasoningResult() *reasoning.ProcessTurnResult {
	return &reasoning.ProcessTurnResult{
		Stage1: reasoning.Stage1Output{
			CandidateEntities: []reasoning.CandidateEntity{},
			CandidateMemories: []reasoning.CandidateMemory{},
		},
		Stage2: reasoning.Stage2Output{
			Operations:     []reasoning.GraphOperation{},
			ProfileDelta:   json.RawMessage(`{}`),
			SessionSummary: "",
			PlanDelta:      json.RawMessage(`{}`),
			Trace: reasoning.Trace{
				SchemaVersion: "v0",
				Stage:         reasoning.StageNameResolve,
				Codes:         []string{"test"},
				MetadataJSON:  json.RawMessage(`{}`),
			},
		},
	}
}

```



<!-- Source: internal/worker/stage2_sources_test.go | bytes=19581 | lines=516 | sha16=793cbc97a9123b29 -->

```go
// ============================================================
// FILE     : internal/worker/stage2_sources_test.go
// PURPOSE  : Verifies store-backed Stage 2 source adapters feed the reasoning envelope.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : TestStoreBackedStage2InputPreparer_LoadsContextFromStores, TestStoreBackedStage2InputPreparer_UsesWorkspaceProfileFallback, TestStoreBackedStage2InputPreparer_PropagatesStoreErrors
// DEPENDS  : context, encoding/json, errors, reflect, strings, testing, internal/core, internal/reasoning
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: These tests must not introduce local extraction or real Codex calls.
// ============================================================

package worker

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/reasoning"
)

func TestStoreBackedStage2InputPreparer_LoadsContextFromStores(t *testing.T) {
	t.Parallel()

	profileStore := &fakeStage2ProfileStore{profiles: map[string]*core.Profile{
		"agent:hermes-main|agent_private": {
			EntityID:   "agent:hermes-main",
			Scope:      core.MemoryScopeAgentPrivate,
			StaticJSON: json.RawMessage(`{"style":"brief"}`),
		},
	}}
	memoryStore := &fakeStage2MemoryStore{resp: &core.SearchMemoriesResponse{Memories: []core.MemoryResult{{
		MemoryID:      "mem_context",
		Kind:          core.MemoryKindConstraint,
		ArtifactClass: core.ArtifactClassKnowledge,
		Text:          "Stage 2 uses store-backed memory context.",
		Confidence:    0.97,
		Scope:         core.MemoryScopeWorkspaceShared,
		OwnerEntityID: "workspace:workspace_1",
		LatestFlag:    true,
	}}}}
	documentStore := &fakeStage2DocumentStore{resp: &core.SearchDocumentsResponse{Chunks: []core.DocumentChunkResult{{
		ChunkID:    "chunk_1",
		DocumentID: "doc_1",
		Text:       "Stage 2 uses store-backed document context.",
		Score:      1,
	}}}}
	planStore := &fakeStage2PlanStore{plans: []*core.Plan{{ID: "plan_1", WorkspaceID: "workspace_1", Status: "active", Scope: core.MemoryScopeWorkspaceShared, OwnerEntityID: "workspace:workspace_1", Title: "Ship Stage 2 sources"}}}
	noteStore := &fakeStage2NoteStore{notes: []*core.Note{{ID: "note_1", WorkspaceID: "workspace_1", Scope: core.MemoryScopeWorkspaceShared, OwnerEntityID: "workspace:workspace_1", Pinned: true, Text: "Pinned Stage 2 constraint"}}}
	groupStore := &fakeStage2GroupStore{}

	preparer := NewStoreBackedStage2InputPreparer(Stage2SourceStores{
		Profiles:  profileStore,
		Memories:  memoryStore,
		Documents: documentStore,
		Plans:     planStore,
		Notes:     noteStore,
		Groups:    groupStore,
	})

	input, err := preparer.Prepare(context.Background(), reasoning.Stage2InputRequest{
		JobID:       "job_1",
		TenantID:    "tenant_1",
		WorkspaceID: "workspace_1",
		RawEvents: []*core.RawEvent{{
			ID:          "evt_1",
			TenantID:    "tenant_1",
			WorkspaceID: "workspace_1",
			ActorID:     "agent:hermes-main",
			PayloadJSON: json.RawMessage(`{"text":"raw local extraction bait"}`),
		}},
		Stage1: reasoning.Stage1Output{
			SummaryHint: "store-backed Stage 2 context",
			TaskHint:    "wire source adapters",
			CandidateMemories: []reasoning.CandidateMemory{{
				Text: "candidate memory text",
			}},
		},
	})
	if err != nil {
		t.Fatalf("Prepare returned error: %v", err)
	}

	if input.RequiredOutputSchema != reasoning.Stage2ResolveOutputSchemaV0 {
		t.Fatalf("expected Stage 2 output schema %q, got %q", reasoning.Stage2ResolveOutputSchemaV0, input.RequiredOutputSchema)
	}
	if input.ExistingProfile == nil || input.ExistingProfile.EntityID != "agent:hermes-main" {
		t.Fatalf("expected agent private profile, got %#v", input.ExistingProfile)
	}
	if len(input.RelevantMemories) != 1 || input.RelevantMemories[0].MemoryID != "mem_context" {
		t.Fatalf("expected store-backed memory context, got %#v", input.RelevantMemories)
	}
	if len(input.RelevantDocuments) != 1 || input.RelevantDocuments[0].ChunkID != "chunk_1" {
		t.Fatalf("expected store-backed document context, got %#v", input.RelevantDocuments)
	}
	if len(input.ActivePlans) != 1 || input.ActivePlans[0].ID != "plan_1" {
		t.Fatalf("expected active plans, got %#v", input.ActivePlans)
	}
	if len(input.PinnedNotes) != 1 || input.PinnedNotes[0].ID != "note_1" {
		t.Fatalf("expected pinned notes, got %#v", input.PinnedNotes)
	}

	if got, want := profileStore.calls, []profileCall{{entityID: "agent:hermes-main", scope: core.MemoryScopeAgentPrivate}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected profile calls: got %#v want %#v", got, want)
	}
	wantScopes := []core.MemoryScope{core.MemoryScopeAgentPrivate, core.MemoryScopeWorkspaceShared, core.MemoryScopeSessionScratch}
	if !reflect.DeepEqual(memoryStore.lastReq.Scopes, wantScopes) {
		t.Fatalf("unexpected memory scopes: got %v want %v", memoryStore.lastReq.Scopes, wantScopes)
	}
	wantArtifactClasses := []core.ArtifactClass{core.ArtifactClassContext, core.ArtifactClassKnowledge, core.ArtifactClassTimeline, core.ArtifactClassPlan}
	if !reflect.DeepEqual(memoryStore.lastReq.ArtifactClasses, wantArtifactClasses) {
		t.Fatalf("unexpected artifact classes: got %v want %v", memoryStore.lastReq.ArtifactClasses, wantArtifactClasses)
	}
	if memoryStore.lastReq.TenantID != "tenant_1" || memoryStore.lastReq.WorkspaceID != "workspace_1" {
		t.Fatalf("unexpected memory search identity: %#v", memoryStore.lastReq)
	}
	if memoryStore.lastReq.OwnerEntityID != "agent:hermes-main" {
		t.Fatalf("expected memory search to be actor-scoped, got owner %q", memoryStore.lastReq.OwnerEntityID)
	}
	if len(memoryStore.lastReq.VisibleGroupIDs) != 0 {
		t.Fatalf("unexpected visible group ids without membership: %#v", memoryStore.lastReq.VisibleGroupIDs)
	}
	if !strings.Contains(memoryStore.lastReq.Query, "store-backed Stage 2 context") || !strings.Contains(memoryStore.lastReq.Query, "candidate memory text") {
		t.Fatalf("expected search query to use structured Stage 1 hints, got %q", memoryStore.lastReq.Query)
	}
	if strings.Contains(memoryStore.lastReq.Query, "raw local extraction bait") {
		t.Fatalf("stage2 source query must not parse raw event payload text, got %q", memoryStore.lastReq.Query)
	}
	if documentStore.lastReq.Query != memoryStore.lastReq.Query {
		t.Fatalf("expected document and memory stores to receive same structured query, got documents=%q memories=%q", documentStore.lastReq.Query, memoryStore.lastReq.Query)
	}
	if planStore.lastReq.WorkspaceID != "workspace_1" || planStore.lastReq.OwnerEntityID != "agent:hermes-main" || !reflect.DeepEqual(planStore.lastReq.Scopes, wantScopes) {
		t.Fatalf("unexpected plan request: %#v", planStore.lastReq)
	}
	if noteStore.lastReq.WorkspaceID != "workspace_1" || noteStore.lastReq.OwnerEntityID != "agent:hermes-main" || !reflect.DeepEqual(noteStore.lastReq.Scopes, wantScopes) {
		t.Fatalf("unexpected note request: %#v", noteStore.lastReq)
	}
}

func TestStoreBackedStage2InputPreparer_IncludesGroupMemoriesForMemberActor(t *testing.T) {
	t.Parallel()

	groupID := "group_design"
	memoryStore := &fakeStage2MemoryStore{resp: &core.SearchMemoriesResponse{Memories: []core.MemoryResult{{
		MemoryID:      "mem_group",
		Kind:          core.MemoryKindDecision,
		ArtifactClass: core.ArtifactClassKnowledge,
		Text:          "Design group chose stdio MCP first.",
		Confidence:    0.91,
		Scope:         core.MemoryScopeGroupShared,
		GroupID:       &groupID,
		OwnerEntityID: "group:group_design",
		LatestFlag:    true,
	}}}}
	groupStore := &fakeStage2GroupStore{memberships: []*core.MemoryGroupMembership{{
		GroupID:  "group_design",
		EntityID: "agent:hermes-main",
	}}}
	preparer := NewStoreBackedStage2InputPreparer(Stage2SourceStores{
		Memories: memoryStore,
		Groups:   groupStore,
	})

	input, err := preparer.Prepare(context.Background(), reasoning.Stage2InputRequest{
		JobID:       "job_1",
		TenantID:    "tenant_1",
		WorkspaceID: "workspace_1",
		RawEvents: []*core.RawEvent{{
			ID:          "evt_1",
			TenantID:    "tenant_1",
			WorkspaceID: "workspace_1",
			ActorID:     "agent:hermes-main",
			PayloadJSON: json.RawMessage(`{"text":"ignored"}`),
		}},
		Stage1: reasoning.Stage1Output{SummaryHint: "group decision"},
	})
	if err != nil {
		t.Fatalf("Prepare returned error: %v", err)
	}

	if len(input.RelevantMemories) != 1 || input.RelevantMemories[0].MemoryID != "mem_group" {
		t.Fatalf("expected group memory context, got %#v", input.RelevantMemories)
	}
	wantScopes := []core.MemoryScope{
		core.MemoryScopeAgentPrivate,
		core.MemoryScopeWorkspaceShared,
		core.MemoryScopeSessionScratch,
		core.MemoryScopeGroupShared,
	}
	if !reflect.DeepEqual(memoryStore.lastReq.Scopes, wantScopes) {
		t.Fatalf("unexpected memory scopes: got %v want %v", memoryStore.lastReq.Scopes, wantScopes)
	}
	if !reflect.DeepEqual(memoryStore.lastReq.VisibleGroupIDs, []string{"group_design"}) {
		t.Fatalf("expected visible group ids, got %#v", memoryStore.lastReq.VisibleGroupIDs)
	}
}

func TestStoreBackedStage2InputPreparer_FiltersPrivateSourcesToRawEventActor(t *testing.T) {
	t.Parallel()

	preparer := NewStoreBackedStage2InputPreparer(Stage2SourceStores{
		Memories: &fakeStage2MemoryStore{resp: &core.SearchMemoriesResponse{Memories: []core.MemoryResult{
			{
				MemoryID:      "mem_actor_private",
				Scope:         core.MemoryScopeAgentPrivate,
				OwnerEntityID: "agent:hermes-main",
				Text:          "visible private memory",
				LatestFlag:    true,
			},
			{
				MemoryID:      "mem_other_private",
				Scope:         core.MemoryScopeAgentPrivate,
				OwnerEntityID: "agent:other",
				Text:          "other agent private memory",
				LatestFlag:    true,
			},
			{
				MemoryID:      "mem_workspace",
				Scope:         core.MemoryScopeWorkspaceShared,
				OwnerEntityID: "workspace:workspace_1",
				Text:          "workspace memory",
				LatestFlag:    true,
			},
			{
				MemoryID:   "mem_scratch",
				Scope:      core.MemoryScopeSessionScratch,
				Text:       "session scratch memory",
				LatestFlag: true,
			},
			{
				MemoryID:      "mem_group",
				Scope:         core.MemoryScopeGroupShared,
				OwnerEntityID: "group:alpha",
				Text:          "group memory without membership filtering",
				LatestFlag:    true,
			},
		}}},
		Plans: &fakeStage2PlanStore{plans: []*core.Plan{
			{ID: "plan_actor_private", Scope: core.MemoryScopeAgentPrivate, OwnerEntityID: "agent:hermes-main", Title: "visible private plan"},
			{ID: "plan_other_private", Scope: core.MemoryScopeAgentPrivate, OwnerEntityID: "agent:other", Title: "other private plan"},
			{ID: "plan_workspace", Scope: core.MemoryScopeWorkspaceShared, OwnerEntityID: "workspace:workspace_1", Title: "workspace plan"},
			{ID: "plan_scratch", Scope: core.MemoryScopeSessionScratch, Title: "session scratch plan"},
			{ID: "plan_group", Scope: core.MemoryScopeGroupShared, OwnerEntityID: "group:alpha", Title: "group plan"},
		}},
		Notes: &fakeStage2NoteStore{notes: []*core.Note{
			{ID: "note_actor_private", Scope: core.MemoryScopeAgentPrivate, OwnerEntityID: "agent:hermes-main", Pinned: true, Text: "visible private note"},
			{ID: "note_other_private", Scope: core.MemoryScopeAgentPrivate, OwnerEntityID: "agent:other", Pinned: true, Text: "other private note"},
			{ID: "note_workspace", Scope: core.MemoryScopeWorkspaceShared, OwnerEntityID: "workspace:workspace_1", Pinned: true, Text: "workspace note"},
			{ID: "note_scratch", Scope: core.MemoryScopeSessionScratch, Pinned: true, Text: "session scratch note"},
			{ID: "note_group", Scope: core.MemoryScopeGroupShared, OwnerEntityID: "group:alpha", Pinned: true, Text: "group note"},
		}},
	})

	input, err := preparer.Prepare(context.Background(), minimalStage2InputRequest())
	if err != nil {
		t.Fatalf("Prepare returned error: %v", err)
	}

	if got, want := memoryResultIDs(input.RelevantMemories), []string{"mem_actor_private", "mem_workspace", "mem_scratch"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected memory ids: got %v want %v", got, want)
	}
	if got, want := planIDs(input.ActivePlans), []string{"plan_actor_private", "plan_workspace", "plan_scratch"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected plan ids: got %v want %v", got, want)
	}
	if got, want := noteIDs(input.PinnedNotes), []string{"note_actor_private", "note_workspace", "note_scratch"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected note ids: got %v want %v", got, want)
	}
}

func TestStoreBackedStage2InputPreparer_UsesWorkspaceProfileFallback(t *testing.T) {
	t.Parallel()

	profileStore := &fakeStage2ProfileStore{profiles: map[string]*core.Profile{
		"workspace:workspace_1|workspace_shared": {
			EntityID:   "workspace:workspace_1",
			Scope:      core.MemoryScopeWorkspaceShared,
			StaticJSON: json.RawMessage(`{"project":"VibeGravity"}`),
		},
	}}
	preparer := NewStoreBackedStage2InputPreparer(Stage2SourceStores{Profiles: profileStore})

	input, err := preparer.Prepare(context.Background(), minimalStage2InputRequest())
	if err != nil {
		t.Fatalf("Prepare returned error: %v", err)
	}
	if input.ExistingProfile == nil || input.ExistingProfile.EntityID != "workspace:workspace_1" {
		t.Fatalf("expected workspace profile fallback, got %#v", input.ExistingProfile)
	}
	wantCalls := []profileCall{
		{entityID: "agent:hermes-main", scope: core.MemoryScopeAgentPrivate},
		{entityID: "workspace:workspace_1", scope: core.MemoryScopeWorkspaceShared},
	}
	if !reflect.DeepEqual(profileStore.calls, wantCalls) {
		t.Fatalf("unexpected profile lookup order: got %#v want %#v", profileStore.calls, wantCalls)
	}
}

func TestStoreBackedStage2InputPreparer_PropagatesStoreErrors(t *testing.T) {
	t.Parallel()

	preparer := NewStoreBackedStage2InputPreparer(Stage2SourceStores{
		Memories: &fakeStage2MemoryStore{err: errors.New("memory search unavailable")},
	})

	_, err := preparer.Prepare(context.Background(), minimalStage2InputRequest())
	if err == nil {
		t.Fatalf("expected store error")
	}
	if !strings.Contains(err.Error(), "load stage2 memories") || !strings.Contains(err.Error(), "memory search unavailable") {
		t.Fatalf("expected wrapped memory source error, got %v", err)
	}
}

func minimalStage2InputRequest() reasoning.Stage2InputRequest {
	return reasoning.Stage2InputRequest{
		JobID:       "job_1",
		TenantID:    "tenant_1",
		WorkspaceID: "workspace_1",
		RawEvents: []*core.RawEvent{{
			ID:          "evt_1",
			TenantID:    "tenant_1",
			WorkspaceID: "workspace_1",
			ActorID:     "agent:hermes-main",
		}},
		Stage1: reasoning.Stage1Output{},
	}
}

func memoryResultIDs(memories []core.MemoryResult) []string {
	ids := make([]string, 0, len(memories))
	for _, memory := range memories {
		ids = append(ids, memory.MemoryID)
	}
	return ids
}

func planIDs(plans []*core.Plan) []string {
	ids := make([]string, 0, len(plans))
	for _, plan := range plans {
		if plan == nil {
			continue
		}
		ids = append(ids, plan.ID)
	}
	return ids
}

func noteIDs(notes []*core.Note) []string {
	ids := make([]string, 0, len(notes))
	for _, note := range notes {
		if note == nil {
			continue
		}
		ids = append(ids, note.ID)
	}
	return ids
}

type profileCall struct {
	entityID string
	scope    core.MemoryScope
}

type fakeStage2ProfileStore struct {
	profiles map[string]*core.Profile
	calls    []profileCall
	err      error
}

func (s *fakeStage2ProfileStore) GetProfile(_ context.Context, entityID string, scope core.MemoryScope) (*core.Profile, error) {
	s.calls = append(s.calls, profileCall{entityID: entityID, scope: scope})
	if s.err != nil {
		return nil, s.err
	}
	profile, ok := s.profiles[entityID+"|"+string(scope)]
	if !ok {
		return nil, core.ErrNotFound
	}
	return profile, nil
}

func (s *fakeStage2ProfileStore) UpsertProfile(context.Context, *core.Profile) error {
	return core.ErrNotImplemented
}

type fakeStage2MemoryStore struct {
	lastReq *core.SearchMemoriesRequest
	resp    *core.SearchMemoriesResponse
	err     error
}

func (s *fakeStage2MemoryStore) UpsertMemory(context.Context, *core.Memory) error {
	return core.ErrNotImplemented
}

func (s *fakeStage2MemoryStore) GetMemory(context.Context, string) (*core.Memory, error) {
	return nil, core.ErrNotImplemented
}

func (s *fakeStage2MemoryStore) SearchMemories(_ context.Context, req *core.SearchMemoriesRequest) (*core.SearchMemoriesResponse, error) {
	s.lastReq = req
	if s.err != nil {
		return nil, s.err
	}
	return s.resp, nil
}

func (s *fakeStage2MemoryStore) UpsertMemoryEdge(context.Context, *core.MemoryEdge) error {
	return core.ErrNotImplemented
}

func (s *fakeStage2MemoryStore) WriteMemoryTrace(context.Context, *core.MemoryTrace) error {
	return core.ErrNotImplemented
}

func (s *fakeStage2MemoryStore) ExplainMemory(context.Context, *core.ExplainMemoryRequest) (*core.ExplainMemoryResponse, error) {
	return nil, core.ErrNotImplemented
}

type fakeStage2DocumentStore struct {
	lastReq *core.SearchDocumentsRequest
	resp    *core.SearchDocumentsResponse
	err     error
}

func (s *fakeStage2DocumentStore) AddDocumentWithChunks(context.Context, *core.Document, []*core.DocumentChunk) error {
	return core.ErrNotImplemented
}

func (s *fakeStage2DocumentStore) AddDocument(context.Context, *core.Document) error {
	return core.ErrNotImplemented
}

func (s *fakeStage2DocumentStore) AddDocumentChunks(context.Context, []*core.DocumentChunk) error {
	return core.ErrNotImplemented
}

func (s *fakeStage2DocumentStore) SearchDocuments(_ context.Context, req *core.SearchDocumentsRequest) (*core.SearchDocumentsResponse, error) {
	s.lastReq = req
	if s.err != nil {
		return nil, s.err
	}
	return s.resp, nil
}

type fakeStage2PlanStore struct {
	lastReq *core.GetActivePlansRequest
	plans   []*core.Plan
	err     error
}

func (s *fakeStage2PlanStore) CreatePlan(context.Context, *core.Plan, []*core.PlanItem) error {
	return core.ErrNotImplemented
}

func (s *fakeStage2PlanStore) UpdatePlan(context.Context, *core.Plan, []*core.PlanItem) error {
	return core.ErrNotImplemented
}

func (s *fakeStage2PlanStore) GetActivePlans(_ context.Context, req *core.GetActivePlansRequest) ([]*core.Plan, error) {
	reqCopy := *req
	reqCopy.Scopes = append([]core.MemoryScope(nil), req.Scopes...)
	s.lastReq = &reqCopy
	if s.err != nil {
		return nil, s.err
	}
	return s.plans, nil
}

type fakeStage2NoteStore struct {
	lastReq *core.ListPinnedNotesRequest
	notes   []*core.Note
	err     error
}

func (s *fakeStage2NoteStore) AddNote(context.Context, *core.Note) error {
	return core.ErrNotImplemented
}

func (s *fakeStage2NoteStore) ListPinnedNotes(_ context.Context, req *core.ListPinnedNotesRequest) ([]*core.Note, error) {
	reqCopy := *req
	reqCopy.Scopes = append([]core.MemoryScope(nil), req.Scopes...)
	s.lastReq = &reqCopy
	if s.err != nil {
		return nil, s.err
	}
	return s.notes, nil
}

type fakeStage2GroupStore struct {
	memberships []*core.MemoryGroupMembership
	err         error
}

func (s *fakeStage2GroupStore) CreateMemoryGroup(context.Context, *core.MemoryGroup) error {
	return core.ErrNotImplemented
}

func (s *fakeStage2GroupStore) AddMembership(context.Context, *core.MemoryGroupMembership) error {
	return core.ErrNotImplemented
}

func (s *fakeStage2GroupStore) ListMemberships(context.Context, string) ([]*core.MemoryGroupMembership, error) {
	return nil, core.ErrNotImplemented
}

func (s *fakeStage2GroupStore) ListMembershipsForEntity(context.Context, string, string, string) ([]*core.MemoryGroupMembership, error) {
	return s.memberships, s.err
}

```



<!-- Source: tests/baseline_test.go | bytes=2263 | lines=80 | sha16=e273f196c7b9bb48 -->

```go
// ============================================================
// FILE     : tests/baseline_test.go
// PURPOSE  : Provides baseline integration smoke tests for config and health checks.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : TestConfigLoad, TestHealthzEndpoint
// DEPENDS  : internal/config, internal/db, internal/httpapi
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: Keep database-dependent tests skippable when VIBEGRAVITY_DB_URL is unset.
// ============================================================

package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/config"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/db"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/httpapi"
)

func TestConfigLoad(t *testing.T) {
	_ = os.Setenv("VIBEGRAVITY_EMBEDDING_MODEL", "test-model-xyz")
	defer func() { _ = os.Unsetenv("VIBEGRAVITY_EMBEDDING_MODEL") }()

	cfg := config.LoadConfig()
	if cfg.EmbeddingModel != "test-model-xyz" {
		t.Errorf("Expected embedding model to be test-model-xyz, got %s", cfg.EmbeddingModel)
	}
}

func TestHealthzEndpoint(t *testing.T) {
	// Skip this test if no database is available
	dbURL := os.Getenv("VIBEGRAVITY_DB_URL")
	if dbURL == "" {
		t.Skip("Skipping TestHealthzEndpoint because VIBEGRAVITY_DB_URL is not set")
	}

	cfg := config.LoadConfig()
	ctx := context.Background()

	pool, err := db.NewPool(ctx, cfg)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	app := &httpapi.App{
		DBPool: pool,
	}

	router := httpapi.NewRouter(app)

	req, err := http.NewRequest("GET", "/healthz", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// We expect a JSON response with status "ok"
	expected := `{"status":"ok"}` + "\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

```



<!-- Source: tests/golden/replay_eval.json | bytes=20565 | lines=609 | sha16=397e4c1ae10655e0 -->

```json
{
  "scenarios": [
    {
      "name": "pinned note and active plan outrank memory",
      "description": "Manual controls stay above learned memory in the recall pack.",
      "prefetch": {
        "tenant_id": "tenant_1",
        "workspace_id": "workspace_1",
        "session_id": "session_1",
        "actor_id": "agent:hermes-main",
        "query": "VibeGravity plan",
        "budget_tokens": 2200,
        "mode": "default"
      },
      "fixtures": {
        "notes": [
          {
            "id": "note_hermes_priority",
            "tenant_id": "tenant_1",
            "workspace_id": "workspace_1",
            "scope": "workspace_shared",
            "owner_entity_id": "workspace:workspace_1",
            "text": "Pinned note: keep Hermes-first delivery as the current priority.",
            "pinned": true
          }
        ],
        "plans": [
          {
            "id": "plan_replay_eval_gates",
            "tenant_id": "tenant_1",
            "workspace_id": "workspace_1",
            "title": "Active plan: finish replay and eval gates before broad packaging.",
            "status": "active",
            "scope": "workspace_shared",
            "owner_entity_id": "workspace:workspace_1"
          }
        ],
        "memories": [
          {
            "memory_id": "mem_1",
            "kind": "procedure",
            "artifact_class": "knowledge",
            "text": "VibeGravity uses Go-first contracts for worker and recall implementation.",
            "confidence": 0.91,
            "scope": "workspace_shared",
            "latest_flag": true
          }
        ]
      },
      "expect": {
        "block_kinds": ["pinned_note", "active_plan", "memory"],
        "block_metadata": [
          {
            "kind": "pinned_note",
            "scope": "workspace_shared",
            "source": "notes",
            "source_id": "note_hermes_priority",
            "status": "pinned",
            "freshness": "stored",
            "owner_entity_id": "workspace:workspace_1"
          },
          {
            "kind": "active_plan",
            "scope": "workspace_shared",
            "source": "plans",
            "source_id": "plan_replay_eval_gates",
            "status": "active",
            "freshness": "stored",
            "owner_entity_id": "workspace:workspace_1"
          },
          {
            "kind": "memory",
            "scope": "workspace_shared",
            "source": "memories",
            "source_id": "mem_1",
            "status": "active",
            "freshness": "stored"
          }
        ],
        "contains": ["Hermes-first", "replay and eval gates", "Go-first contracts"],
        "sources": ["notes", "plans", "memories"],
        "max_tokens": 2200
      }
    },
    {
      "name": "agent private memory stays owner scoped",
      "description": "Private memory belonging to another agent must not leak into Hermes recall.",
      "prefetch": {
        "tenant_id": "tenant_1",
        "workspace_id": "workspace_1",
        "session_id": "session_1",
        "actor_id": "agent:hermes-main",
        "query": "private preference",
        "budget_tokens": 2200,
        "mode": "default"
      },
      "fixtures": {
        "memories": [
          {
            "memory_id": "mem_private_visible",
            "kind": "preference",
            "artifact_class": "knowledge",
            "text": "Hermes private preference: use compact recall wording.",
            "confidence": 0.88,
            "scope": "agent_private",
            "owner_entity_id": "agent:hermes-main",
            "latest_flag": true
          },
          {
            "memory_id": "mem_private_other",
            "kind": "preference",
            "artifact_class": "knowledge",
            "text": "Other agent private preference: never show this in Hermes recall.",
            "confidence": 0.88,
            "scope": "agent_private",
            "owner_entity_id": "agent:claude",
            "latest_flag": true
          }
        ]
      },
      "expect": {
        "block_kinds": ["memory"],
        "contains": ["Hermes private preference"],
        "not_contains": ["Other agent private preference"],
        "sources": ["memories"],
        "max_tokens": 2200
      }
    },
    {
      "name": "superseded memory is suppressed",
      "description": "Recall should keep latest memory and hide superseded facts.",
      "prefetch": {
        "tenant_id": "tenant_1",
        "workspace_id": "workspace_1",
        "session_id": "session_1",
        "actor_id": "agent:hermes-main",
        "query": "database store",
        "budget_tokens": 2200,
        "mode": "default"
      },
      "fixtures": {
        "memories": [
          {
            "memory_id": "mem_old",
            "kind": "fact",
            "artifact_class": "knowledge",
            "text": "Old fact: SQLite is the canonical store.",
            "confidence": 0.7,
            "scope": "workspace_shared",
            "latest_flag": false
          },
          {
            "memory_id": "mem_new",
            "kind": "fact",
            "artifact_class": "knowledge",
            "text": "Current fact: PostgreSQL is the canonical store.",
            "confidence": 0.95,
            "scope": "workspace_shared",
            "latest_flag": true
          }
        ]
      },
      "expect": {
        "block_kinds": ["memory"],
        "contains": ["PostgreSQL is the canonical store"],
        "not_contains": ["SQLite is the canonical store"],
        "sources": ["memories"],
        "max_tokens": 2200
      }
    },
    {
      "name": "degraded recall still returns profile and summary",
      "description": "When memory search is unavailable or query is empty, existing profile and session summary still provide useful context.",
      "prefetch": {
        "tenant_id": "tenant_1",
        "workspace_id": "workspace_1",
        "session_id": "session_1",
        "actor_id": "agent:hermes-main",
        "query": "",
        "budget_tokens": 120,
        "mode": "small"
      },
      "fixtures": {
        "profiles": [
          {
            "entity_id": "agent:hermes-main",
            "scope": "agent_private",
            "static_json": {"style": "brief", "runtime": "Hermes-first"}
          }
        ],
        "session_summaries": [
          {
            "tenant_id": "tenant_1",
            "workspace_id": "workspace_1",
            "session_id": "session_1",
            "summary_text": "Session summary: eval work is focused on deterministic replay gates."
          }
        ]
      },
      "expect": {
        "block_kinds": ["profile_static", "session_summary"],
        "contains": ["Hermes-first", "deterministic replay gates"],
        "sources": ["profile", "session_summaries"],
        "max_tokens": 120
      }
    },
    {
      "name": "budget truncates low priority context",
      "description": "A small token budget must keep the pack within the requested ceiling.",
      "prefetch": {
        "tenant_id": "tenant_1",
        "workspace_id": "workspace_1",
        "session_id": "session_1",
        "actor_id": "agent:hermes-main",
        "query": "quality",
        "budget_tokens": 8,
        "mode": "small"
      },
      "fixtures": {
        "notes": [
          {
            "tenant_id": "tenant_1",
            "workspace_id": "workspace_1",
            "scope": "workspace_shared",
            "owner_entity_id": "workspace:workspace_1",
            "text": "Quality gate note: golden scenarios should remain compact deterministic and easy to review.",
            "pinned": true
          }
        ],
        "memories": [
          {
            "memory_id": "mem_quality",
            "kind": "procedure",
            "artifact_class": "knowledge",
            "text": "Quality memory: use replay before prompt changes.",
            "confidence": 0.9,
            "scope": "workspace_shared",
            "latest_flag": true
          }
        ]
      },
      "expect": {
        "block_kinds": ["pinned_note"],
        "contains": ["Quality gate note"],
        "not_contains": ["Quality memory"],
        "sources": ["notes"],
        "max_tokens": 8
      }
    }
  ],
  "graph_replay_scenarios": [
    {
      "name": "update memory replay suppresses prior fact",
      "description": "Replaying the same update_memory operation must keep one replacement, one trace, one updates edge, and hide the superseded fact from recall.",
      "tenant_id": "tenant_1",
      "workspace_id": "workspace_1",
      "job_id": "job_graph_update_store",
      "raw_event_ids": ["evt_graph_update_store"],
      "retry_count": 2,
      "initial_memories": [
        {
          "memory_id": "mem_graph_store_old",
          "kind": "fact",
          "artifact_class": "knowledge",
          "text": "Old fact: SQLite is the canonical store.",
          "confidence": 0.7,
          "scope": "workspace_shared",
          "owner_entity_id": "workspace:workspace_1",
          "latest_flag": true
        }
      ],
      "operations": [
        {
          "operation_id": "op_graph_update_store",
          "kind": "update_memory",
          "memory": {
            "target_id": "mem_graph_store_old",
            "kind": "fact",
            "artifact_class": "knowledge",
            "scope": "workspace_shared",
            "owner_entity_id": "workspace:workspace_1",
            "text": "Current fact: PostgreSQL is the canonical store.",
            "confidence": 0.97,
            "metadata_json": {}
          },
          "edge": {
            "to_memory_id": "mem_graph_store_old",
            "edge_kind": "updates",
            "confidence": 0.99
          },
          "raw_event_ids": ["evt_graph_update_store"],
          "metadata": {}
        }
      ],
      "prefetch": {
        "tenant_id": "tenant_1",
        "workspace_id": "workspace_1",
        "session_id": "session_1",
        "actor_id": "agent:hermes-main",
        "query": "canonical store",
        "budget_tokens": 2200,
        "mode": "default"
      },
      "expect": {
        "applied_operation_count": 1,
        "memory_count": 2,
        "trace_count": 1,
        "edge_count": 1,
        "superseded_memory_ids": ["mem_graph_store_old"],
        "block_kinds": ["memory"],
        "contains": ["PostgreSQL is the canonical store"],
        "not_contains": ["SQLite is the canonical store"],
        "sources": ["memories"],
        "max_tokens": 2200
      }
    },
    {
      "name": "correction replay changes later recall",
      "description": "A correction-shaped update_memory replay should make the corrected replacement visible and keep the prior preference suppressed.",
      "tenant_id": "tenant_1",
      "workspace_id": "workspace_1",
      "job_id": "job_graph_correction_replay",
      "raw_event_ids": ["evt_graph_correction_replay"],
      "retry_count": 2,
      "initial_memories": [
        {
          "memory_id": "mem_graph_preference_old",
          "kind": "preference",
          "artifact_class": "knowledge",
          "text": "Old preference: use verbose recall summaries.",
          "confidence": 0.75,
          "scope": "agent_private",
          "owner_entity_id": "agent:hermes-main",
          "latest_flag": true
        }
      ],
      "operations": [
        {
          "operation_id": "op_graph_correction_replay",
          "kind": "update_memory",
          "memory": {
            "target_id": "mem_graph_preference_old",
            "kind": "preference",
            "artifact_class": "knowledge",
            "scope": "agent_private",
            "owner_entity_id": "agent:hermes-main",
            "text": "Corrected preference: use compact recall summaries.",
            "confidence": 0.96,
            "metadata_json": {"source": "operator_correction"}
          },
          "edge": {
            "to_memory_id": "mem_graph_preference_old",
            "edge_kind": "updates",
            "confidence": 0.99
          },
          "raw_event_ids": ["evt_graph_correction_replay"],
          "metadata": {"source": "operator_correction"}
        }
      ],
      "prefetch": {
        "tenant_id": "tenant_1",
        "workspace_id": "workspace_1",
        "session_id": "session_1",
        "actor_id": "agent:hermes-main",
        "query": "recall summaries",
        "budget_tokens": 2200,
        "mode": "default"
      },
      "expect": {
        "applied_operation_count": 1,
        "memory_count": 2,
        "trace_count": 1,
        "edge_count": 1,
        "superseded_memory_ids": ["mem_graph_preference_old"],
        "trace_contains": ["operator_correction"],
        "block_kinds": ["memory"],
        "contains": ["Corrected preference: use compact recall summaries"],
        "not_contains": ["verbose recall summaries"],
        "sources": ["memories"],
        "max_tokens": 2200
      }
    },
    {
      "name": "group shared graph write remains rejected",
      "description": "The eval fixture keeps the group_shared stop-line visible until membership validation exists.",
      "tenant_id": "tenant_1",
      "workspace_id": "workspace_1",
      "job_id": "job_graph_group_rejected",
      "raw_event_ids": ["evt_graph_group_rejected"],
      "initial_memories": [
        {
          "memory_id": "mem_group_old",
          "kind": "fact",
          "artifact_class": "knowledge",
          "text": "Group fact: private launch room decision.",
          "confidence": 0.8,
          "scope": "group_shared",
          "group_id": "group_design",
          "owner_entity_id": "group:group_design",
          "latest_flag": true
        }
      ],
      "operations": [
        {
          "operation_id": "op_graph_group_rejected",
          "kind": "update_memory",
          "memory": {
            "target_id": "mem_group_old",
            "kind": "fact",
            "artifact_class": "knowledge",
            "scope": "group_shared",
            "group_id": "group_design",
            "owner_entity_id": "group:group_design",
            "text": "Group fact: launch room decision changed.",
            "confidence": 0.9,
            "metadata_json": {}
          },
          "edge": {
            "to_memory_id": "mem_group_old",
            "edge_kind": "updates",
            "confidence": 0.9
          },
          "raw_event_ids": ["evt_graph_group_rejected"],
          "metadata": {}
        }
      ],
      "expect": {
        "rejected": true,
        "error_contains": "requires membership validation before writes",
        "applied_operation_count": 0,
        "memory_count": 1,
        "trace_count": 0,
        "edge_count": 0
      }
    }
  ],
  "worker_backlog_scenarios": [
    {
      "name": "stage1 outage retries without graph writes",
      "description": "A deterministic mocked Stage 1 outage uses the retry path, then recovery writes one derived memory with trace only after reasoning succeeds.",
      "tenant_id": "tenant_1",
      "workspace_id": "workspace_1",
      "job_id": "job_worker_stage1_outage",
      "stage_1_outage_attempts": 1,
      "max_worker_passes": 2,
      "raw_events": [
        {
          "event_id": "evt_worker_stage1_outage",
          "session_id": "session_1",
          "actor_id": "agent:hermes-main",
          "payload_json": {"text": "Remember that eval gates should simulate Codex outage locally."}
        }
      ],
      "operations": [
        {
          "operation_id": "op_worker_stage1_create",
          "kind": "create_memory",
          "memory": {
            "kind": "procedure",
            "artifact_class": "knowledge",
            "scope": "workspace_shared",
            "owner_entity_id": "workspace:workspace_1",
            "text": "Eval gates simulate Codex outage locally before real Codex is enabled.",
            "confidence": 0.92,
            "metadata_json": {}
          },
          "raw_event_ids": ["evt_worker_stage1_outage"],
          "metadata": {}
        }
      ],
      "expect": {
        "completed_jobs": 1,
        "failed_attempts": 1,
        "blocked_jobs": 0,
        "queued_jobs": 0,
        "applied_operation_count": 1,
        "memory_count": 1,
        "trace_count": 1,
        "edge_count": 0,
        "no_graph_writes_before_success": true,
        "error_contains": "mock codex stage1 outage"
      }
    },
    {
      "name": "stage2 outage recovery replay is idempotent",
      "description": "A deterministic mocked Stage 2 outage retries, then update_memory recovery and replay keep one replacement, one trace, and one updates edge.",
      "tenant_id": "tenant_1",
      "workspace_id": "workspace_1",
      "job_id": "job_worker_stage2_outage",
      "stage_2_outage_attempts": 1,
      "replay_after_success_count": 1,
      "max_worker_passes": 2,
      "raw_events": [
        {
          "event_id": "evt_worker_stage2_outage",
          "session_id": "session_1",
          "actor_id": "agent:hermes-main",
          "payload_json": {"text": "Correction: worker backlog recovery must not duplicate graph rows."}
        }
      ],
      "initial_memories": [
        {
          "memory_id": "mem_worker_replay_old",
          "kind": "fact",
          "artifact_class": "knowledge",
          "text": "Old fact: worker replay can duplicate graph rows.",
          "confidence": 0.7,
          "scope": "workspace_shared",
          "owner_entity_id": "workspace:workspace_1",
          "latest_flag": true
        }
      ],
      "operations": [
        {
          "operation_id": "op_worker_stage2_update",
          "kind": "update_memory",
          "memory": {
            "target_id": "mem_worker_replay_old",
            "kind": "fact",
            "artifact_class": "knowledge",
            "scope": "workspace_shared",
            "owner_entity_id": "workspace:workspace_1",
            "text": "Current fact: worker recovery replay stays idempotent for memory, trace, and edge rows.",
            "confidence": 0.96,
            "metadata_json": {}
          },
          "edge": {
            "to_memory_id": "mem_worker_replay_old",
            "edge_kind": "updates",
            "confidence": 0.99
          },
          "raw_event_ids": ["evt_worker_stage2_outage"],
          "metadata": {}
        }
      ],
      "expect": {
        "completed_jobs": 1,
        "failed_attempts": 1,
        "blocked_jobs": 0,
        "queued_jobs": 0,
        "applied_operation_count": 1,
        "memory_count": 2,
        "trace_count": 1,
        "edge_count": 1,
        "no_graph_writes_before_success": true,
        "error_contains": "mock codex stage2 outage"
      }
    },
    {
      "name": "unsupported apply work becomes blocked",
      "description": "Deterministic archive_memory work is observable as blocked and does not loop through transient retry.",
      "tenant_id": "tenant_1",
      "workspace_id": "workspace_1",
      "job_id": "job_worker_archive_blocked",
      "max_worker_passes": 2,
      "raw_events": [
        {
          "event_id": "evt_worker_archive_blocked",
          "session_id": "session_1",
          "actor_id": "agent:hermes-main",
          "payload_json": {"text": "Archive the outdated memory."}
        }
      ],
      "initial_memories": [
        {
          "memory_id": "mem_worker_archive_target",
          "kind": "fact",
          "artifact_class": "knowledge",
          "text": "Fact targeted by unsupported archive work.",
          "confidence": 0.8,
          "scope": "workspace_shared",
          "owner_entity_id": "workspace:workspace_1",
          "latest_flag": true
        }
      ],
      "operations": [
        {
          "operation_id": "op_worker_archive_blocked",
          "kind": "archive_memory",
          "memory": {
            "target_id": "mem_worker_archive_target",
            "kind": "fact",
            "artifact_class": "knowledge",
            "scope": "workspace_shared",
            "owner_entity_id": "workspace:workspace_1",
            "text": "Archive target placeholder.",
            "confidence": 0.8,
            "metadata_json": {}
          },
          "raw_event_ids": ["evt_worker_archive_blocked"],
          "metadata": {}
        }
      ],
      "expect": {
        "completed_jobs": 0,
        "failed_attempts": 0,
        "blocked_jobs": 1,
        "queued_jobs": 0,
        "applied_operation_count": 0,
        "memory_count": 1,
        "trace_count": 0,
        "edge_count": 0,
        "no_graph_writes_before_success": true,
        "error_contains": "unsupported apply work"
      }
    }
  ]
}

```



<!-- Source: tests/migration_contract_test.go | bytes=4572 | lines=121 | sha16=20da94220156d03a -->

```go
// ============================================================
// FILE     : tests/migration_contract_test.go
// PURPOSE  : Guards migration and job-state contracts that must not regress during parallel Work Pack 03 fixes.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : TestMigrationContractUpdatesUniquenessTargetsPriorMemory, TestJobFailureContractSeparatesBlockedFromRetryable
// DEPENDS  : os, path/filepath, runtime, strings, testing
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: Keep these tests contract-focused; do not turn them into feature implementation tests.
// ============================================================

package tests

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestMigrationContractUpdatesUniquenessTargetsPriorMemory(t *testing.T) {
	t.Parallel()

	sql := readRepoFile(t, "migrations", "000002_create_core_tables.up.sql")
	indexSQL := extractBetween(t, sql, "CREATE UNIQUE INDEX memory_edges_single_updates_target_idx", ";")

	if !strings.Contains(indexSQL, "ON memory_edges (to_memory_id)") {
		t.Fatalf("updates uniqueness must guard the prior memory target, got:\n%s", indexSQL)
	}
	if strings.Contains(indexSQL, "ON memory_edges (from_memory_id)") {
		t.Fatalf("updates uniqueness must not be keyed by from_memory_id, got:\n%s", indexSQL)
	}
	if !strings.Contains(indexSQL, "WHERE edge_kind = 'updates'") {
		t.Fatalf("updates uniqueness must remain a partial updates-only index, got:\n%s", indexSQL)
	}
}

func TestJobFailureContractSeparatesBlockedFromRetryable(t *testing.T) {
	t.Parallel()

	source := readRepoFile(t, "internal", "store", "postgres", "jobs.go")
	failJob := extractBetween(t, source, "func failJob", "// BlockJob records deterministic unsupported work")
	blockJob := extractBetween(t, source, "func blockJob", "func jobErrorString")

	if !strings.Contains(failJob, "status = 'queued'") {
		t.Fatalf("FailJob must return transient failures to the queued retry state, got:\n%s", failJob)
	}
	if !strings.Contains(failJob, "interval '30 seconds'") {
		t.Fatalf("FailJob must preserve retry scheduling, got:\n%s", failJob)
	}
	if strings.Contains(failJob, "status = 'blocked'") {
		t.Fatalf("FailJob must not use the permanent blocked state, got:\n%s", failJob)
	}

	if !strings.Contains(blockJob, "status = 'blocked'") {
		t.Fatalf("BlockJob must use the permanent blocked state, got:\n%s", blockJob)
	}
	if strings.Contains(blockJob, "interval '30 seconds'") {
		t.Fatalf("BlockJob must not schedule automatic retry, got:\n%s", blockJob)
	}
	if strings.Contains(blockJob, "status = 'queued'") {
		t.Fatalf("BlockJob must not return deterministic unsupported work to queued, got:\n%s", blockJob)
	}
}

func TestMigrationContractAddsAppendSafeMemoryCorrections(t *testing.T) {
	t.Parallel()

	sql := readRepoFile(t, "migrations", "000002_create_core_tables.up.sql")
	tableSQL := extractBetween(t, sql, "CREATE TABLE memory_corrections", ");")
	indexSQL := extractBetween(t, sql, "CREATE UNIQUE INDEX memory_corrections_tenant_workspace_idempotency_key_idx", ";")

	for _, want := range []string{
		"memory_id TEXT NOT NULL REFERENCES memories (id)",
		"raw_event_id TEXT NOT NULL REFERENCES raw_events (id)",
		"correction_text TEXT NOT NULL",
		"status TEXT NOT NULL DEFAULT 'recorded'",
	} {
		if !strings.Contains(tableSQL, want) {
			t.Fatalf("memory_corrections must preserve %q, got:\n%s", want, tableSQL)
		}
	}
	if !strings.Contains(indexSQL, "ON memory_corrections (tenant_id, workspace_id, idempotency_key)") {
		t.Fatalf("memory correction idempotency must be tenant/workspace scoped, got:\n%s", indexSQL)
	}
}

func readRepoFile(t *testing.T, parts ...string) string {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not locate current test file")
	}
	repoRoot := filepath.Dir(filepath.Dir(currentFile))
	pathParts := append([]string{repoRoot}, parts...)
	data, err := os.ReadFile(filepath.Join(pathParts...))
	if err != nil {
		t.Fatalf("read repo file %v: %v", parts, err)
	}
	return string(data)
}

func extractBetween(t *testing.T, text, startMarker, endMarker string) string {
	t.Helper()

	start := strings.Index(text, startMarker)
	if start < 0 {
		t.Fatalf("missing start marker %q", startMarker)
	}
	remainder := text[start:]
	end := strings.Index(remainder, endMarker)
	if end < 0 {
		t.Fatalf("missing end marker %q after %q", endMarker, startMarker)
	}
	return remainder[:end+len(endMarker)]
}

```



<!-- Source: tools/headercheck/main.go | bytes=4298 | lines=180 | sha16=1c1c9f51abff5a0f -->

```go
// ============================================================
// FILE     : tools/headercheck/main.go
// PURPOSE  : Validates machine-readable headers on Go source files.
// LAYER    : util
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : main
// DEPENDS  : docs/code-header-policy.md
// USED_BY  : Makefile, agent handoff checks
// ------------------------------------------------------------
// AGENT_NOTE: Keep this dependency-free so header checks run in a fresh clone.
// ============================================================

// Package main implements the VibeGravity source header checker.
package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var (
	requiredFields = []string{
		"FILE",
		"PURPOSE",
		"LAYER",
		"STATUS",
		"EXPORTS",
		"DEPENDS",
		"USED_BY",
		"AGENT_NOTE",
	}
	validLayers = map[string]bool{
		"domain":      true,
		"application": true,
		"interface":   true,
		"infra":       true,
		"util":        true,
		"test":        true,
	}
	validStatuses = map[string]bool{
		"draft":        true,
		"active":       true,
		"experimental": true,
		"deprecated":   true,
	}
	fieldPattern = regexp.MustCompile(`^//\s*([A-Z_]+)\s*:\s*(.*)$`)
)

func main() {
	root := flag.String("root", ".", "repository root to scan")
	flag.Parse()

	absRoot, err := filepath.Abs(*root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve root: %v\n", err)
		os.Exit(2)
	}

	var failures []string
	err = filepath.WalkDir(absRoot, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", path, err))
			return nil
		}
		if entry.IsDir() {
			if shouldSkipDir(entry.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" {
			return nil
		}

		rel, err := filepath.Rel(absRoot, path)
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", path, err))
			return nil
		}
		rel = filepath.ToSlash(rel)
		failures = append(failures, validateFile(path, rel)...)
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "scan files: %v\n", err)
		os.Exit(2)
	}

	sort.Strings(failures)
	if len(failures) > 0 {
		fmt.Println("code header check failed:")
		for _, failure := range failures {
			fmt.Printf("  - %s\n", failure)
		}
		os.Exit(1)
	}

	fmt.Println("code header check passed")
}

func shouldSkipDir(name string) bool {
	switch name {
	case ".git", "bin", "dist", "node_modules", "vendor":
		return true
	default:
		return false
	}
}

func validateFile(path, rel string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return []string{fmt.Sprintf("%s: read: %v", rel, err)}
	}
	text := string(data)
	if strings.HasPrefix(text, "// Code generated ") {
		return nil
	}

	packageOffset := packageOffset(text)
	if packageOffset < 0 {
		return []string{fmt.Sprintf("%s: missing package declaration", rel)}
	}

	fields := parseHeaderFields(text[:packageOffset])
	var failures []string
	for _, field := range requiredFields {
		value, ok := fields[field]
		if !ok {
			failures = append(failures, fmt.Sprintf("%s: missing %s header field", rel, field))
			continue
		}
		if strings.TrimSpace(value) == "" {
			failures = append(failures, fmt.Sprintf("%s: empty %s header field", rel, field))
		}
	}

	if got := fields["FILE"]; got != "" && got != rel {
		failures = append(failures, fmt.Sprintf("%s: FILE header is %q", rel, got))
	}
	if got := fields["LAYER"]; got != "" && !validLayers[got] {
		failures = append(failures, fmt.Sprintf("%s: invalid LAYER %q", rel, got))
	}
	if got := fields["STATUS"]; got != "" && !validStatuses[got] {
		failures = append(failures, fmt.Sprintf("%s: invalid STATUS %q", rel, got))
	}

	return failures
}

func parseHeaderFields(prefix string) map[string]string {
	fields := make(map[string]string)
	for _, line := range strings.Split(prefix, "\n") {
		match := fieldPattern.FindStringSubmatch(line)
		if len(match) != 3 {
			continue
		}
		fields[match[1]] = strings.TrimSpace(match[2])
	}
	return fields
}

func packageOffset(text string) int {
	offset := 0
	for _, line := range strings.SplitAfter(text, "\n") {
		if strings.HasPrefix(line, "package ") {
			return offset
		}
		offset += len(line)
	}
	return -1
}

```
