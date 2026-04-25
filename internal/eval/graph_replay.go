// ============================================================
// FILE     : internal/eval/graph_replay.go
// PURPOSE  : Runs deterministic graph apply replay scenarios for memory update gates.
// LAYER    : application
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : GraphReplayScenario, GraphReplayExpectation, RunGraphReplayScenarios
// DEPENDS  : context, encoding/json, fmt, sort, strings, time, internal/core, internal/graph, internal/reasoning, internal/recall
// USED_BY  : internal/eval golden runner, tests/golden/replay_eval.json
// ------------------------------------------------------------
// AGENT_NOTE: Replay evals must stay local-only and must not call Codex, Hermes, or external stores.
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
	"github.com/parker-jungwoo-hwang/vibegravity/internal/recall"
)

// GraphReplayScenario replays structured graph operations through the real apply engine.
type GraphReplayScenario struct {
	Name            string                     `json:"name"`
	Description     string                     `json:"description"`
	TenantID        string                     `json:"tenant_id"`
	WorkspaceID     string                     `json:"workspace_id"`
	JobID           string                     `json:"job_id"`
	RawEventIDs     []string                   `json:"raw_event_ids"`
	InitialMemories []GraphMemoryFixture       `json:"initial_memories"`
	Operations      []reasoning.GraphOperation `json:"operations"`
	RetryCount      int                        `json:"retry_count"`
	Prefetch        core.PrefetchRequest       `json:"prefetch"`
	Expect          GraphReplayExpectation     `json:"expect"`
	Stage1          reasoning.Stage1Output     `json:"stage_1"`
	Trace           reasoning.Trace            `json:"trace"`
}

// GraphMemoryFixture is the compact memory row shape used by replay fixtures.
type GraphMemoryFixture struct {
	MemoryID      string             `json:"memory_id"`
	TenantID      string             `json:"tenant_id,omitempty"`
	WorkspaceID   string             `json:"workspace_id,omitempty"`
	Kind          core.MemoryKind    `json:"kind"`
	ArtifactClass core.ArtifactClass `json:"artifact_class"`
	Text          string             `json:"text"`
	Confidence    float64            `json:"confidence"`
	Scope         core.MemoryScope   `json:"scope"`
	GroupID       *string            `json:"group_id,omitempty"`
	OwnerEntityID string             `json:"owner_entity_id,omitempty"`
	Status        core.MemoryStatus  `json:"status,omitempty"`
	LatestFlag    bool               `json:"latest_flag"`
}

// GraphReplayExpectation extends recall expectations with graph persistence checks.
type GraphReplayExpectation struct {
	Expectation
	AppliedOperationCount *int     `json:"applied_operation_count,omitempty"`
	MemoryCount           *int     `json:"memory_count,omitempty"`
	TraceCount            *int     `json:"trace_count,omitempty"`
	EdgeCount             *int     `json:"edge_count,omitempty"`
	ActiveMemoryIDs       []string `json:"active_memory_ids"`
	SupersededMemoryIDs   []string `json:"superseded_memory_ids"`
	TraceContains         []string `json:"trace_contains"`
	Rejected              bool     `json:"rejected"`
	ErrorContains         string   `json:"error_contains"`
}

type graphReplayMemoryStore struct {
	memories map[string]*core.Memory
	traces   map[string]*core.MemoryTrace
	edges    map[string]*core.MemoryEdge
}

// RunGraphReplayScenarios executes graph replay scenarios against in-memory stores.
func RunGraphReplayScenarios(ctx context.Context, scenarios []GraphReplayScenario) *Summary {
	results := make([]Result, 0, len(scenarios))
	passed := true
	for _, scenario := range scenarios {
		result := runGraphReplayScenario(ctx, scenario)
		if !result.Passed {
			passed = false
		}
		results = append(results, result)
	}
	return &Summary{Passed: passed, Results: results}
}

func runGraphReplayScenario(ctx context.Context, scenario GraphReplayScenario) Result {
	errs := validateGraphReplayScenario(scenario)
	store := newGraphReplayMemoryStore()
	for _, fixture := range scenario.InitialMemories {
		memory, err := fixture.toMemory(scenario)
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}
		store.memories[memory.ID] = memory
	}
	if len(errs) > 0 {
		return Result{Scenario: scenario.Name, Passed: false, Errors: errs}
	}

	engine, err := graph.NewStoreBackedApplyEngine(store)
	if err != nil {
		return Result{Scenario: scenario.Name, Passed: false, Errors: []string{err.Error()}}
	}

	req := graphReplayApplyRequest(scenario)
	var applyResult *graph.ApplyResult
	var applyErr error
	attempts := scenario.RetryCount
	if attempts < 1 {
		attempts = 1
	}
	for range attempts {
		applyResult, applyErr = engine.Apply(ctx, req)
	}

	errs = append(errs, compareGraphApplyResult(scenario.Expect, store, applyResult, applyErr)...)
	observed := Observed{}
	if shouldRunGraphPrefetch(scenario) {
		var prefetchErr error
		observed, prefetchErr = observeGraphReplayPrefetch(ctx, scenario, store)
		if prefetchErr != nil {
			errs = append(errs, prefetchErr.Error())
		} else {
			errs = append(errs, compareExpectation(scenario.Expect.Expectation, observed)...)
		}
	}
	return Result{
		Scenario: scenario.Name,
		Passed:   len(errs) == 0,
		Errors:   errs,
		Observed: observed,
	}
}

func graphReplayApplyRequest(scenario GraphReplayScenario) *graph.ApplyRequest {
	trace := scenario.Trace
	if trace.SchemaVersion == "" {
		trace = reasoning.Trace{
			SchemaVersion: "vibegravity.eval.graph_replay.v1",
			Stage:         reasoning.StageNameResolve,
			MetadataJSON:  json.RawMessage(`{}`),
		}
	}
	if trace.Stage == "" {
		trace.Stage = reasoning.StageNameResolve
	}
	return &graph.ApplyRequest{
		JobID:       scenario.JobID,
		TenantID:    scenario.TenantID,
		WorkspaceID: scenario.WorkspaceID,
		RawEventIDs: append([]string(nil), scenario.RawEventIDs...),
		Reasoning: &reasoning.ProcessTurnResult{
			Stage1: scenario.Stage1,
			Stage2: reasoning.Stage2Output{
				Operations:   append([]reasoning.GraphOperation(nil), scenario.Operations...),
				ProfileDelta: json.RawMessage(`{}`),
				PlanDelta:    json.RawMessage(`{}`),
				Trace:        trace,
			},
		},
	}
}

func validateGraphReplayScenario(scenario GraphReplayScenario) []string {
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
	if len(scenario.RawEventIDs) == 0 {
		errs = append(errs, "raw_event_ids are required")
	}
	if len(scenario.Operations) == 0 {
		errs = append(errs, "operations are required")
	}
	return errs
}

func compareGraphApplyResult(expect GraphReplayExpectation, store *graphReplayMemoryStore, result *graph.ApplyResult, applyErr error) []string {
	var errs []string
	if expect.Rejected {
		if applyErr == nil {
			errs = append(errs, "expected graph apply to be rejected")
		} else if expect.ErrorContains != "" && !strings.Contains(applyErr.Error(), expect.ErrorContains) {
			errs = append(errs, fmt.Sprintf("apply error %q does not contain %q", applyErr.Error(), expect.ErrorContains))
		}
	} else if applyErr != nil {
		errs = append(errs, "graph apply failed: "+applyErr.Error())
	}
	if expect.AppliedOperationCount != nil {
		got := 0
		if result != nil {
			got = result.AppliedOperationCount
		}
		if got != *expect.AppliedOperationCount {
			errs = append(errs, fmt.Sprintf("applied operation count got %d want %d", got, *expect.AppliedOperationCount))
		}
	}
	if expect.MemoryCount != nil && store.memoryCount() != *expect.MemoryCount {
		errs = append(errs, fmt.Sprintf("memory count got %d want %d", store.memoryCount(), *expect.MemoryCount))
	}
	if expect.TraceCount != nil && store.traceCount() != *expect.TraceCount {
		errs = append(errs, fmt.Sprintf("trace count got %d want %d", store.traceCount(), *expect.TraceCount))
	}
	if expect.EdgeCount != nil && store.edgeCount() != *expect.EdgeCount {
		errs = append(errs, fmt.Sprintf("edge count got %d want %d", store.edgeCount(), *expect.EdgeCount))
	}
	for _, memoryID := range expect.ActiveMemoryIDs {
		if !store.memoryHasStatus(memoryID, core.MemoryStatusActive, true) {
			errs = append(errs, fmt.Sprintf("memory %s is not active latest", memoryID))
		}
	}
	for _, memoryID := range expect.SupersededMemoryIDs {
		if !store.memoryHasStatus(memoryID, core.MemoryStatusSuperseded, false) {
			errs = append(errs, fmt.Sprintf("memory %s is not superseded", memoryID))
		}
	}
	traceText := store.traceText()
	for _, needle := range expect.TraceContains {
		if !strings.Contains(traceText, needle) {
			errs = append(errs, fmt.Sprintf("trace is missing expected text %q", needle))
		}
	}
	return errs
}

func shouldRunGraphPrefetch(scenario GraphReplayScenario) bool {
	return strings.TrimSpace(scenario.Prefetch.TenantID) != "" ||
		strings.TrimSpace(scenario.Prefetch.WorkspaceID) != "" ||
		strings.TrimSpace(scenario.Prefetch.SessionID) != "" ||
		strings.TrimSpace(scenario.Prefetch.ActorID) != ""
}

func observeGraphReplayPrefetch(ctx context.Context, scenario GraphReplayScenario, store *graphReplayMemoryStore) (Observed, error) {
	assembler := recall.NewAssembler(recall.Dependencies{Memories: store})
	resp, err := assembler.Prefetch(ctx, &scenario.Prefetch)
	if err != nil {
		return Observed{}, fmt.Errorf("prefetch after graph replay failed: %w", err)
	}
	return observe(resp), nil
}

func newGraphReplayMemoryStore() *graphReplayMemoryStore {
	return &graphReplayMemoryStore{
		memories: make(map[string]*core.Memory),
		traces:   make(map[string]*core.MemoryTrace),
		edges:    make(map[string]*core.MemoryEdge),
	}
}

func (f GraphMemoryFixture) toMemory(scenario GraphReplayScenario) (*core.Memory, error) {
	if strings.TrimSpace(f.MemoryID) == "" {
		return nil, fmt.Errorf("initial memory id is required")
	}
	if f.Scope == "" {
		return nil, fmt.Errorf("initial memory %s scope is required", f.MemoryID)
	}
	if f.Kind == "" {
		return nil, fmt.Errorf("initial memory %s kind is required", f.MemoryID)
	}
	if f.ArtifactClass == "" {
		return nil, fmt.Errorf("initial memory %s artifact_class is required", f.MemoryID)
	}
	now := time.Date(2026, 4, 24, 0, 0, 0, 0, time.UTC)
	tenantID := f.TenantID
	if tenantID == "" {
		tenantID = scenario.TenantID
	}
	workspaceID := f.WorkspaceID
	if workspaceID == "" {
		workspaceID = scenario.WorkspaceID
	}
	status := f.Status
	if status == "" {
		status = core.MemoryStatusActive
		if !f.LatestFlag {
			status = core.MemoryStatusSuperseded
		}
	}
	confidence := f.Confidence
	if confidence == 0 {
		confidence = 0.8
	}
	return &core.Memory{
		ID:            f.MemoryID,
		TenantID:      tenantID,
		WorkspaceID:   workspaceID,
		Scope:         f.Scope,
		GroupID:       cloneStringPtr(f.GroupID),
		OwnerEntityID: defaultOwnerEntityID(f.Scope, f.OwnerEntityID, workspaceID),
		Kind:          f.Kind,
		ArtifactClass: f.ArtifactClass,
		Text:          f.Text,
		Confidence:    confidence,
		Status:        status,
		ValidFrom:     now,
		LatestFlag:    f.LatestFlag,
		MetadataJSON:  json.RawMessage(`{}`),
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

func defaultOwnerEntityID(scope core.MemoryScope, ownerEntityID string, workspaceID string) string {
	if ownerEntityID != "" {
		return ownerEntityID
	}
	if scope == core.MemoryScopeWorkspaceShared {
		return "workspace:" + workspaceID
	}
	return ownerEntityID
}

func (s *graphReplayMemoryStore) CreateMemoryWithTrace(ctx context.Context, memory *core.Memory, trace *core.MemoryTrace) error {
	_ = ctx
	return s.writeMemoryTraceAndOptionalEdge(memory, trace, nil)
}

func (s *graphReplayMemoryStore) CreateMemoryWithTraceAndEdge(ctx context.Context, memory *core.Memory, trace *core.MemoryTrace, edge *core.MemoryEdge) error {
	_ = ctx
	if _, ok := s.memories[edge.ToMemoryID]; !ok {
		return fmt.Errorf("%w: target memory %s is required", core.ErrNotFound, edge.ToMemoryID)
	}
	return s.writeMemoryTraceAndOptionalEdge(memory, trace, edge)
}

func (s *graphReplayMemoryStore) CreateMemoryWithTraceAndUpdateEdge(ctx context.Context, memory *core.Memory, trace *core.MemoryTrace, edge *core.MemoryEdge) error {
	_ = ctx
	if s.completedUpdateAlreadyApplied(memory, edge) {
		return nil
	}
	target, ok := s.memories[edge.ToMemoryID]
	if !ok {
		return fmt.Errorf("%w: update target memory %s is required", core.ErrNotFound, edge.ToMemoryID)
	}
	if target.TenantID != memory.TenantID || target.WorkspaceID != memory.WorkspaceID {
		return fmt.Errorf("%w: update target crosses tenant or workspace boundary", core.ErrInvalidArgument)
	}
	if target.Status != core.MemoryStatusActive || !target.LatestFlag {
		return fmt.Errorf("%w: update target must be active latest", core.ErrInvalidArgument)
	}
	if target.Scope != memory.Scope || target.OwnerEntityID != memory.OwnerEntityID || !sameStringPtr(target.GroupID, memory.GroupID) {
		return fmt.Errorf("%w: update target scope, group, and owner must match replacement", core.ErrInvalidArgument)
	}
	if err := s.writeMemoryTraceAndOptionalEdge(memory, trace, edge); err != nil {
		return err
	}
	target.ValidTo = cloneTimePtr(&memory.ValidFrom)
	target.Status = core.MemoryStatusSuperseded
	target.LatestFlag = false
	target.UpdatedAt = memory.CreatedAt
	return nil
}

func (s *graphReplayMemoryStore) UpsertMemory(ctx context.Context, memory *core.Memory) error {
	_ = ctx
	if memory == nil || memory.ID == "" {
		return fmt.Errorf("%w: memory is required", core.ErrInvalidArgument)
	}
	s.memories[memory.ID] = cloneMemory(memory)
	return nil
}

func (s *graphReplayMemoryStore) GetMemory(ctx context.Context, memoryID string) (*core.Memory, error) {
	_ = ctx
	memory, ok := s.memories[memoryID]
	if !ok {
		return nil, core.ErrNotFound
	}
	return cloneMemory(memory), nil
}

func (s *graphReplayMemoryStore) SearchMemories(ctx context.Context, req *core.SearchMemoriesRequest) (*core.SearchMemoriesResponse, error) {
	_ = ctx
	if req == nil {
		return nil, fmt.Errorf("%w: search memories request is required", core.ErrInvalidArgument)
	}
	results := make([]core.MemoryResult, 0, len(s.memories))
	query := strings.TrimSpace(strings.ToLower(req.Query))
	for _, memory := range s.memories {
		if memory.TenantID != req.TenantID || memory.WorkspaceID != req.WorkspaceID {
			continue
		}
		if memory.Status != core.MemoryStatusActive || !memory.LatestFlag {
			continue
		}
		if !scopeVisible(memory.Scope, memory.OwnerEntityID, req.OwnerEntityID, req.Scopes) {
			continue
		}
		if !artifactClassVisible(memory.ArtifactClass, req.ArtifactClasses) {
			continue
		}
		text := strings.ToLower(memory.Text)
		if query != "" && !strings.Contains(text, query) && !wordOverlap(text, query) {
			continue
		}
		results = append(results, core.MemoryResult{
			MemoryID:      memory.ID,
			Kind:          memory.Kind,
			ArtifactClass: memory.ArtifactClass,
			Text:          memory.Text,
			Confidence:    memory.Confidence,
			Scope:         memory.Scope,
			OwnerEntityID: memory.OwnerEntityID,
			ValidFrom:     memory.ValidFrom,
			LatestFlag:    memory.LatestFlag,
		})
	}
	sort.SliceStable(results, func(i, j int) bool {
		if !results[i].ValidFrom.Equal(results[j].ValidFrom) {
			return results[i].ValidFrom.After(results[j].ValidFrom)
		}
		return results[i].MemoryID < results[j].MemoryID
	})
	return &core.SearchMemoriesResponse{Memories: results}, nil
}

func (s *graphReplayMemoryStore) UpsertMemoryEdge(ctx context.Context, edge *core.MemoryEdge) error {
	_ = ctx
	if edge == nil {
		return fmt.Errorf("%w: memory edge is required", core.ErrInvalidArgument)
	}
	s.edges[edgeKey(edge)] = cloneEdge(edge)
	return nil
}

func (s *graphReplayMemoryStore) WriteMemoryTrace(ctx context.Context, trace *core.MemoryTrace) error {
	_ = ctx
	if trace == nil || trace.MemoryID == "" {
		return fmt.Errorf("%w: memory trace is required", core.ErrInvalidArgument)
	}
	s.traces[trace.MemoryID] = cloneTrace(trace)
	return nil
}

func (s *graphReplayMemoryStore) ExplainMemory(ctx context.Context, req *core.ExplainMemoryRequest) (*core.ExplainMemoryResponse, error) {
	_ = ctx
	if req == nil {
		return nil, fmt.Errorf("%w: explain memory request is required", core.ErrInvalidArgument)
	}
	memory, ok := s.memories[req.MemoryID]
	if !ok || memory.TenantID != req.TenantID || memory.WorkspaceID != req.WorkspaceID {
		return nil, core.ErrNotFound
	}
	trace, ok := s.traces[req.MemoryID]
	if !ok {
		return nil, core.ErrNotFound
	}
	edges := make([]core.MemoryEdgeResult, 0, len(s.edges))
	for _, edge := range s.edges {
		if edge.FromMemoryID != req.MemoryID && edge.ToMemoryID != req.MemoryID {
			continue
		}
		edges = append(edges, core.MemoryEdgeResult{
			FromMemoryID: edge.FromMemoryID,
			ToMemoryID:   edge.ToMemoryID,
			EdgeKind:     edge.EdgeKind,
			Confidence:   edge.Confidence,
			CreatedAt:    edge.CreatedAt,
		})
	}
	return &core.ExplainMemoryResponse{
		MemoryID: req.MemoryID,
		Trace: core.MemoryTraceResult{
			RawEventIDs:            append([]string(nil), trace.RawEventIDs...),
			ReasoningJobID:         trace.ReasoningJobID,
			ReasoningStage:         trace.ReasoningStage,
			CandidateSnapshotJSON:  append(json.RawMessage(nil), trace.CandidateSnapshotJSON...),
			AppliedOperationsJSON:  append(json.RawMessage(nil), trace.AppliedOperationsJSON...),
			OperatorCorrectionFlag: trace.OperatorCorrectionFlag,
			RelatedDocumentIDs:     append([]string(nil), trace.RelatedDocumentIDs...),
			CreatedAt:              trace.CreatedAt,
		},
		Edges: edges,
	}, nil
}

func (s *graphReplayMemoryStore) writeMemoryTraceAndOptionalEdge(memory *core.Memory, trace *core.MemoryTrace, edge *core.MemoryEdge) error {
	if memory == nil || memory.ID == "" {
		return fmt.Errorf("%w: memory is required", core.ErrInvalidArgument)
	}
	if trace == nil || trace.MemoryID != memory.ID {
		return fmt.Errorf("%w: memory trace must target the written memory", core.ErrInvalidArgument)
	}
	s.memories[memory.ID] = cloneMemory(memory)
	s.traces[trace.MemoryID] = cloneTrace(trace)
	if edge != nil {
		s.edges[edgeKey(edge)] = cloneEdge(edge)
	}
	return nil
}

func (s *graphReplayMemoryStore) completedUpdateAlreadyApplied(memory *core.Memory, edge *core.MemoryEdge) bool {
	if memory == nil || edge == nil {
		return false
	}
	if _, ok := s.memories[memory.ID]; !ok {
		return false
	}
	if _, ok := s.traces[memory.ID]; !ok {
		return false
	}
	if _, ok := s.edges[edgeKey(edge)]; !ok {
		return false
	}
	target, ok := s.memories[edge.ToMemoryID]
	return ok && target.Status == core.MemoryStatusSuperseded && !target.LatestFlag
}

func (s *graphReplayMemoryStore) memoryCount() int {
	return len(s.memories)
}

func (s *graphReplayMemoryStore) traceCount() int {
	return len(s.traces)
}

func (s *graphReplayMemoryStore) edgeCount() int {
	return len(s.edges)
}

func (s *graphReplayMemoryStore) memoryHasStatus(memoryID string, status core.MemoryStatus, latest bool) bool {
	memory, ok := s.memories[memoryID]
	return ok && memory.Status == status && memory.LatestFlag == latest
}

func (s *graphReplayMemoryStore) traceText() string {
	ids := make([]string, 0, len(s.traces))
	for id := range s.traces {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	parts := make([]string, 0, len(ids)*4)
	for _, id := range ids {
		trace := s.traces[id]
		parts = append(parts,
			strings.Join(trace.RawEventIDs, " "),
			trace.ReasoningJobID,
			string(trace.CandidateSnapshotJSON),
			string(trace.AppliedOperationsJSON),
		)
	}
	return strings.Join(parts, "\n")
}

func edgeKey(edge *core.MemoryEdge) string {
	if edge == nil {
		return ""
	}
	return edge.FromMemoryID + "\x00" + edge.ToMemoryID + "\x00" + string(edge.EdgeKind)
}

func cloneMemory(memory *core.Memory) *core.Memory {
	if memory == nil {
		return nil
	}
	out := *memory
	out.GroupID = cloneStringPtr(memory.GroupID)
	out.ValidTo = cloneTimePtr(memory.ValidTo)
	out.EmbeddingUpdatedAt = cloneTimePtr(memory.EmbeddingUpdatedAt)
	out.MetadataJSON = append(json.RawMessage(nil), memory.MetadataJSON...)
	return &out
}

func cloneTrace(trace *core.MemoryTrace) *core.MemoryTrace {
	if trace == nil {
		return nil
	}
	out := *trace
	out.RawEventIDs = append([]string(nil), trace.RawEventIDs...)
	out.CandidateSnapshotJSON = append(json.RawMessage(nil), trace.CandidateSnapshotJSON...)
	out.AppliedOperationsJSON = append(json.RawMessage(nil), trace.AppliedOperationsJSON...)
	out.RelatedDocumentIDs = append([]string(nil), trace.RelatedDocumentIDs...)
	return &out
}

func cloneEdge(edge *core.MemoryEdge) *core.MemoryEdge {
	if edge == nil {
		return nil
	}
	out := *edge
	return &out
}

func cloneStringPtr(value *string) *string {
	if value == nil {
		return nil
	}
	out := *value
	return &out
}

func cloneTimePtr(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	out := *value
	return &out
}

func sameStringPtr(left *string, right *string) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}
