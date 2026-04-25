// ============================================================
// FILE     : internal/eval/worker_backlog.go
// PURPOSE  : Runs deterministic worker outage and backlog recovery eval scenarios.
// LAYER    : application
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : WorkerBacklogScenario, WorkerBacklogExpectation, RunWorkerBacklogScenarios
// DEPENDS  : context, encoding/json, fmt, sort, strings, time, internal/core, internal/graph, internal/reasoning, internal/worker
// USED_BY  : internal/eval golden runner, tests/golden/replay_eval.json
// ------------------------------------------------------------
// AGENT_NOTE: These evals use mocked Stage 1/Stage 2 runners only; they must not call real Codex.
// ============================================================

package eval

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/graph"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/reasoning"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/worker"
)

// WorkerBacklogScenario replays worker passes with mocked Stage 1/Stage 2 outage controls.
type WorkerBacklogScenario struct {
	Name                    string                     `json:"name"`
	Description             string                     `json:"description"`
	TenantID                string                     `json:"tenant_id"`
	WorkspaceID             string                     `json:"workspace_id"`
	JobID                   string                     `json:"job_id"`
	RawEvents               []WorkerRawEventFixture    `json:"raw_events"`
	InitialMemories         []GraphMemoryFixture       `json:"initial_memories"`
	Stage1OutageAttempts    int                        `json:"stage_1_outage_attempts"`
	Stage2OutageAttempts    int                        `json:"stage_2_outage_attempts"`
	Operations              []reasoning.GraphOperation `json:"operations"`
	ReplayAfterSuccessCount int                        `json:"replay_after_success_count"`
	MaxWorkerPasses         int                        `json:"max_worker_passes"`
	Expect                  WorkerBacklogExpectation   `json:"expect"`
}

// WorkerRawEventFixture is the compact raw event row used by worker backlog fixtures.
type WorkerRawEventFixture struct {
	EventID     string          `json:"event_id"`
	SessionID   string          `json:"session_id"`
	ActorID     string          `json:"actor_id"`
	EventKind   string          `json:"event_kind"`
	Source      string          `json:"source"`
	PayloadJSON json.RawMessage `json:"payload_json"`
}

// WorkerBacklogExpectation captures worker status and graph side-effect expectations.
type WorkerBacklogExpectation struct {
	CompletedJobs              int    `json:"completed_jobs"`
	FailedAttempts             int    `json:"failed_attempts"`
	BlockedJobs                int    `json:"blocked_jobs"`
	QueuedJobs                 int    `json:"queued_jobs"`
	AppliedOperationCount      int    `json:"applied_operation_count"`
	MemoryCount                int    `json:"memory_count"`
	TraceCount                 int    `json:"trace_count"`
	EdgeCount                  int    `json:"edge_count"`
	NoGraphWritesBeforeSuccess bool   `json:"no_graph_writes_before_success"`
	ErrorContains              string `json:"error_contains"`
}

type workerBacklogObserved struct {
	completedJobs         int
	failedAttempts        int
	blockedJobs           int
	queuedJobs            int
	appliedOperationCount int
	memoryCount           int
	traceCount            int
	edgeCount             int
	firstSuccessPass      int
	firstWritePass        int
	errors                []string
}

// RunWorkerBacklogScenarios executes deterministic worker outage and recovery scenarios.
func RunWorkerBacklogScenarios(ctx context.Context, scenarios []WorkerBacklogScenario) *Summary {
	results := make([]Result, 0, len(scenarios))
	passed := true
	for _, scenario := range scenarios {
		result := runWorkerBacklogScenario(ctx, scenario)
		if !result.Passed {
			passed = false
		}
		results = append(results, result)
	}
	return &Summary{Passed: passed, Results: results}
}

func runWorkerBacklogScenario(ctx context.Context, scenario WorkerBacklogScenario) Result {
	errs := validateWorkerBacklogScenario(scenario)
	store := newGraphReplayMemoryStore()
	for _, fixture := range scenario.InitialMemories {
		memory, err := fixture.toMemory(GraphReplayScenario{
			TenantID:    scenario.TenantID,
			WorkspaceID: scenario.WorkspaceID,
		})
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}
		store.memories[memory.ID] = memory
	}
	if len(errs) > 0 {
		return Result{Scenario: scenario.Name, Passed: false, Errors: errs}
	}

	applyEngine, err := graph.NewStoreBackedApplyEngine(store)
	if err != nil {
		return Result{Scenario: scenario.Name, Passed: false, Errors: []string{err.Error()}}
	}
	stage1 := &outageStage1Extractor{failAttempts: scenario.Stage1OutageAttempts}
	stage2 := &outageStage2Resolver{
		failAttempts: scenario.Stage2OutageAttempts,
		operations:   append([]reasoning.GraphOperation(nil), scenario.Operations...),
	}
	reasoner, err := reasoning.NewPipelineOrchestrator(stage1, stage2, nil)
	if err != nil {
		return Result{Scenario: scenario.Name, Passed: false, Errors: []string{err.Error()}}
	}
	jobs := newWorkerBacklogJobStore(scenario)
	rawEvents := newWorkerBacklogRawEventStore(scenario)
	processor, err := worker.NewProcessor(worker.Dependencies{
		WorkerID:    "worker:eval",
		BatchSize:   1,
		Jobs:        jobs,
		RawEvents:   rawEvents,
		Reasoner:    reasoner,
		ApplyEngine: applyEngine,
		Dreaming:    newWorkerBacklogDreamingService(),
		Clock: func() time.Time {
			return workerBacklogNow()
		},
	})
	if err != nil {
		return Result{Scenario: scenario.Name, Passed: false, Errors: []string{err.Error()}}
	}

	observedState := runWorkerBacklogPasses(ctx, scenario, processor, jobs, store)
	if observedState.firstSuccessPass > 0 && scenario.ReplayAfterSuccessCount > 0 {
		req := workerBacklogApplyRequest(scenario, stage1.output())
		for range scenario.ReplayAfterSuccessCount {
			if _, err := applyEngine.Apply(ctx, req); err != nil {
				observedState.errors = append(observedState.errors, "replay after success failed: "+err.Error())
			}
		}
		observedState.memoryCount = store.memoryCount()
		observedState.traceCount = store.traceCount()
		observedState.edgeCount = store.edgeCount()
	}

	errs = append(errs, compareWorkerBacklogExpectation(scenario.Expect, observedState)...)
	observed := Observed{
		Sources: workerBacklogSources(observedState),
		Tokens:  observedState.appliedOperationCount,
	}
	return Result{
		Scenario: scenario.Name,
		Passed:   len(errs) == 0,
		Errors:   errs,
		Observed: observed,
	}
}

func runWorkerBacklogPasses(ctx context.Context, scenario WorkerBacklogScenario, processor *worker.Processor, jobs *workerBacklogJobStore, store *graphReplayMemoryStore) workerBacklogObserved {
	maxPasses := scenario.MaxWorkerPasses
	if maxPasses <= 0 {
		maxPasses = 1 + scenario.Stage1OutageAttempts + scenario.Stage2OutageAttempts + scenario.ReplayAfterSuccessCount
	}
	if maxPasses < 1 {
		maxPasses = 1
	}
	initialMemoryCount := store.memoryCount()
	initialTraceCount := store.traceCount()
	initialEdgeCount := store.edgeCount()
	observed := workerBacklogObserved{}
	for pass := 1; pass <= maxPasses; pass++ {
		result, err := processor.RunOnce(ctx)
		if err != nil {
			observed.errors = append(observed.errors, err.Error())
		}
		observed.appliedOperationCount += result.AppliedOperationCount
		if result.Completed > 0 && observed.firstSuccessPass == 0 {
			observed.firstSuccessPass = pass
		}
		if observed.firstWritePass == 0 && (store.memoryCount() != initialMemoryCount || store.traceCount() != initialTraceCount || store.edgeCount() != initialEdgeCount) {
			observed.firstWritePass = pass
		}
		if !jobs.hasQueuedJob() {
			break
		}
	}
	observed.completedJobs = jobs.statusCount("complete")
	observed.failedAttempts = jobs.failedAttempts
	observed.blockedJobs = jobs.statusCount("blocked")
	observed.queuedJobs = jobs.statusCount("queued")
	observed.memoryCount = store.memoryCount()
	observed.traceCount = store.traceCount()
	observed.edgeCount = store.edgeCount()
	return observed
}

func validateWorkerBacklogScenario(scenario WorkerBacklogScenario) []string {
	var errs []string
	required := map[string]string{
		"scenario name": scenario.Name,
		"tenant_id":     scenario.TenantID,
		"workspace_id":  scenario.WorkspaceID,
		"job_id":        scenario.JobID,
	}
	for name, value := range required {
		if strings.TrimSpace(value) == "" {
			errs = append(errs, name+" is required")
		}
	}
	if len(scenario.RawEvents) == 0 {
		errs = append(errs, "raw_events are required")
	}
	if len(scenario.Operations) == 0 {
		errs = append(errs, "operations are required")
	}
	return errs
}

func compareWorkerBacklogExpectation(expect WorkerBacklogExpectation, observed workerBacklogObserved) []string {
	var errs []string
	checks := []struct {
		name string
		got  int
		want int
	}{
		{name: "completed jobs", got: observed.completedJobs, want: expect.CompletedJobs},
		{name: "failed attempts", got: observed.failedAttempts, want: expect.FailedAttempts},
		{name: "blocked jobs", got: observed.blockedJobs, want: expect.BlockedJobs},
		{name: "queued jobs", got: observed.queuedJobs, want: expect.QueuedJobs},
		{name: "applied operation count", got: observed.appliedOperationCount, want: expect.AppliedOperationCount},
		{name: "memory count", got: observed.memoryCount, want: expect.MemoryCount},
		{name: "trace count", got: observed.traceCount, want: expect.TraceCount},
		{name: "edge count", got: observed.edgeCount, want: expect.EdgeCount},
	}
	for _, check := range checks {
		if check.got != check.want {
			errs = append(errs, fmt.Sprintf("%s got %d want %d", check.name, check.got, check.want))
		}
	}
	if expect.NoGraphWritesBeforeSuccess && observed.firstWritePass > 0 && observed.firstSuccessPass > 0 && observed.firstWritePass < observed.firstSuccessPass {
		errs = append(errs, fmt.Sprintf("graph writes occurred on pass %d before successful pass %d", observed.firstWritePass, observed.firstSuccessPass))
	}
	if expect.ErrorContains != "" {
		joined := strings.Join(observed.errors, "\n")
		if !strings.Contains(joined, expect.ErrorContains) {
			errs = append(errs, fmt.Sprintf("worker errors %q do not contain %q", joined, expect.ErrorContains))
		}
	}
	return errs
}

func workerBacklogApplyRequest(scenario WorkerBacklogScenario, stage1 reasoning.Stage1Output) *graph.ApplyRequest {
	return &graph.ApplyRequest{
		JobID:       scenario.JobID,
		TenantID:    scenario.TenantID,
		WorkspaceID: scenario.WorkspaceID,
		RawEventIDs: workerBacklogRawEventIDs(scenario.RawEvents),
		Reasoning: &reasoning.ProcessTurnResult{
			Stage1: stage1,
			Stage2: reasoning.Stage2Output{
				Operations:     append([]reasoning.GraphOperation(nil), scenario.Operations...),
				ProfileDelta:   json.RawMessage(`{}`),
				SessionSummary: "",
				PlanDelta:      json.RawMessage(`{}`),
				Trace: reasoning.Trace{
					SchemaVersion: "vibegravity.eval.worker_backlog.v1",
					Stage:         reasoning.StageNameResolve,
					Codes:         []string{"mock_codex_recovered"},
					MetadataJSON:  json.RawMessage(`{"client":"mock_codex_outage_eval"}`),
				},
			},
		},
	}
}

func workerBacklogSources(observed workerBacklogObserved) []string {
	sources := []string{
		fmt.Sprintf("complete=%d", observed.completedJobs),
		fmt.Sprintf("failed=%d", observed.failedAttempts),
		fmt.Sprintf("blocked=%d", observed.blockedJobs),
		fmt.Sprintf("queued=%d", observed.queuedJobs),
		fmt.Sprintf("memories=%d", observed.memoryCount),
		fmt.Sprintf("traces=%d", observed.traceCount),
		fmt.Sprintf("edges=%d", observed.edgeCount),
	}
	sort.Strings(sources)
	return sources
}

type outageStage1Extractor struct {
	failAttempts int
	calls        int
}

func (e *outageStage1Extractor) Extract(_ context.Context, _ reasoning.Stage1Input) (reasoning.Stage1Output, error) {
	e.calls++
	if e.calls <= e.failAttempts {
		return reasoning.Stage1Output{}, fmt.Errorf("mock codex stage1 outage attempt %d", e.calls)
	}
	return e.output(), nil
}

func (e *outageStage1Extractor) output() reasoning.Stage1Output {
	return reasoning.Stage1Output{
		CandidateEntities: []reasoning.CandidateEntity{},
		CandidateMemories: []reasoning.CandidateMemory{},
		SummaryHint:       "mock_codex_stage1_recovered",
		TaskHint:          "mock_codex_stage2_apply_operations",
	}
}

type outageStage2Resolver struct {
	failAttempts int
	calls        int
	operations   []reasoning.GraphOperation
}

func (r *outageStage2Resolver) Resolve(_ context.Context, _ reasoning.Stage2Input) (reasoning.Stage2Output, error) {
	r.calls++
	if r.calls <= r.failAttempts {
		return reasoning.Stage2Output{}, fmt.Errorf("mock codex stage2 outage attempt %d", r.calls)
	}
	return reasoning.Stage2Output{
		Operations:     append([]reasoning.GraphOperation(nil), r.operations...),
		ProfileDelta:   json.RawMessage(`{}`),
		SessionSummary: "",
		PlanDelta:      json.RawMessage(`{}`),
		Trace: reasoning.Trace{
			SchemaVersion: "vibegravity.eval.worker_backlog.v1",
			Stage:         reasoning.StageNameResolve,
			Codes:         []string{"mock_codex_recovered"},
			MetadataJSON:  json.RawMessage(`{"client":"mock_codex_outage_eval"}`),
		},
	}, nil
}

type workerBacklogJobStore struct {
	jobs           map[string]*core.IngestJob
	order          []string
	failedAttempts int
}

func newWorkerBacklogJobStore(scenario WorkerBacklogScenario) *workerBacklogJobStore {
	job := &core.IngestJob{
		ID:          scenario.JobID,
		TenantID:    scenario.TenantID,
		WorkspaceID: scenario.WorkspaceID,
		JobKind:     core.JobKindProcessTurnEvent,
		Status:      "queued",
		RawEventIDs: workerBacklogRawEventIDs(scenario.RawEvents),
		PayloadJSON: json.RawMessage(`{"session_id":"session_1"}`),
		AvailableAt: workerBacklogNow(),
		CreatedAt:   workerBacklogNow(),
		UpdatedAt:   workerBacklogNow(),
	}
	return &workerBacklogJobStore{
		jobs:  map[string]*core.IngestJob{job.ID: job},
		order: []string{job.ID},
	}
}

func (s *workerBacklogJobStore) EnqueueJobs(context.Context, []*core.IngestJob) ([]string, error) {
	return nil, core.ErrNotImplemented
}

func (s *workerBacklogJobStore) ClaimJobs(_ context.Context, workerID string, limit int) ([]*core.IngestJob, error) {
	if limit <= 0 {
		limit = 1
	}
	claimed := make([]*core.IngestJob, 0, limit)
	for _, id := range s.order {
		job := s.jobs[id]
		if job == nil || job.Status != "queued" {
			continue
		}
		job.Status = "running"
		job.LockedBy = &workerID
		now := workerBacklogNow()
		job.LockedAt = &now
		claimed = append(claimed, cloneWorkerBacklogJob(job))
		if len(claimed) >= limit {
			break
		}
	}
	return claimed, nil
}

func (s *workerBacklogJobStore) CompleteJob(_ context.Context, jobID string) error {
	job, ok := s.jobs[jobID]
	if !ok {
		return core.ErrNotFound
	}
	job.Status = "complete"
	job.LockedBy = nil
	job.LockedAt = nil
	job.UpdatedAt = workerBacklogNow()
	return nil
}

func (s *workerBacklogJobStore) FailJob(_ context.Context, jobID string, err error) error {
	job, ok := s.jobs[jobID]
	if !ok {
		return core.ErrNotFound
	}
	job.Status = "queued"
	job.Attempts++
	job.LockedBy = nil
	job.LockedAt = nil
	message := err.Error()
	job.LastError = &message
	job.UpdatedAt = workerBacklogNow()
	s.failedAttempts++
	return nil
}

func (s *workerBacklogJobStore) BlockJob(_ context.Context, jobID string, err error) error {
	job, ok := s.jobs[jobID]
	if !ok {
		return core.ErrNotFound
	}
	job.Status = "blocked"
	job.Attempts++
	job.LockedBy = nil
	job.LockedAt = nil
	message := err.Error()
	job.LastError = &message
	job.UpdatedAt = workerBacklogNow()
	return nil
}

func (s *workerBacklogJobStore) hasQueuedJob() bool {
	return s.statusCount("queued") > 0
}

func (s *workerBacklogJobStore) statusCount(status string) int {
	count := 0
	for _, job := range s.jobs {
		if job.Status == status {
			count++
		}
	}
	return count
}

type workerBacklogRawEventStore struct {
	events map[string]*core.RawEvent
}

func newWorkerBacklogRawEventStore(scenario WorkerBacklogScenario) *workerBacklogRawEventStore {
	events := make(map[string]*core.RawEvent, len(scenario.RawEvents))
	for i, fixture := range scenario.RawEvents {
		eventID := fixture.EventID
		if eventID == "" {
			eventID = fmt.Sprintf("evt_%d", i+1)
		}
		sessionID := fixture.SessionID
		if sessionID == "" {
			sessionID = "session_1"
		}
		actorID := fixture.ActorID
		if actorID == "" {
			actorID = "agent:hermes-main"
		}
		eventKind := fixture.EventKind
		if eventKind == "" {
			eventKind = "message"
		}
		source := fixture.Source
		if source == "" {
			source = "hermes"
		}
		payload := fixture.PayloadJSON
		if len(payload) == 0 {
			payload = json.RawMessage(`{"text":"eval event"}`)
		}
		events[eventID] = &core.RawEvent{
			ID:          eventID,
			TenantID:    scenario.TenantID,
			WorkspaceID: scenario.WorkspaceID,
			SessionID:   sessionID,
			ActorID:     actorID,
			EventKind:   eventKind,
			Source:      source,
			Fingerprint: "fp_" + eventID,
			OccurredAt:  workerBacklogNow().Add(time.Duration(i) * time.Minute),
			PayloadJSON: append(json.RawMessage(nil), payload...),
			CreatedAt:   workerBacklogNow().Add(time.Duration(i) * time.Minute),
		}
	}
	return &workerBacklogRawEventStore{events: events}
}

func (s *workerBacklogRawEventStore) AppendRawEvents(context.Context, []*core.RawEvent) ([]string, error) {
	return nil, core.ErrNotImplemented
}

func (s *workerBacklogRawEventStore) GetRawEvents(_ context.Context, ids []string) ([]*core.RawEvent, error) {
	events := make([]*core.RawEvent, 0, len(ids))
	for _, id := range ids {
		if event, ok := s.events[id]; ok {
			events = append(events, cloneWorkerBacklogRawEvent(event))
		}
	}
	return events, nil
}

func newWorkerBacklogDreamingService() *graph.DreamingService {
	service, err := graph.NewDreamingService(graph.DreamingDependencies{
		Store: workerBacklogDreamingStore{},
		Clock: func() time.Time {
			return workerBacklogNow()
		},
	})
	if err != nil {
		panic(err)
	}
	return service
}

type workerBacklogDreamingStore struct{}

func (workerBacklogDreamingStore) LoadDreamingSessionInput(context.Context, *core.DreamSessionRequest) (*core.DreamingSessionInput, error) {
	return &core.DreamingSessionInput{}, nil
}

func (workerBacklogDreamingStore) PromoteMemories(context.Context, *core.DreamingPromotionRequest) (*core.DreamingPromotionResult, error) {
	return &core.DreamingPromotionResult{}, nil
}

func (workerBacklogDreamingStore) UpsertSessionSummary(context.Context, *core.SessionSummary) error {
	return nil
}

func (workerBacklogDreamingStore) GetSessionSummary(context.Context, string) (*core.SessionSummary, error) {
	return nil, core.ErrNotFound
}

func workerBacklogRawEventIDs(events []WorkerRawEventFixture) []string {
	ids := make([]string, 0, len(events))
	for i, event := range events {
		if event.EventID != "" {
			ids = append(ids, event.EventID)
			continue
		}
		ids = append(ids, fmt.Sprintf("evt_%d", i+1))
	}
	return ids
}

func cloneWorkerBacklogJob(job *core.IngestJob) *core.IngestJob {
	if job == nil {
		return nil
	}
	out := *job
	out.RawEventIDs = append([]string(nil), job.RawEventIDs...)
	out.PayloadJSON = append(json.RawMessage(nil), job.PayloadJSON...)
	if job.LockedBy != nil {
		value := *job.LockedBy
		out.LockedBy = &value
	}
	if job.LockedAt != nil {
		value := *job.LockedAt
		out.LockedAt = &value
	}
	if job.LastError != nil {
		value := *job.LastError
		out.LastError = &value
	}
	return &out
}

func cloneWorkerBacklogRawEvent(event *core.RawEvent) *core.RawEvent {
	if event == nil {
		return nil
	}
	out := *event
	out.PayloadJSON = append(json.RawMessage(nil), event.PayloadJSON...)
	return &out
}

func workerBacklogNow() time.Time {
	return time.Date(2026, time.April, 24, 0, 0, 0, 0, time.UTC)
}
