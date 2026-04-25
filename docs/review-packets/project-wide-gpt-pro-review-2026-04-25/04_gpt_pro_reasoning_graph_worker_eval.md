# 04 Gpt Pro Reasoning Graph Worker Eval

Generated: 2026-04-25

This file is part of the GPT-Pro review material bundle for VibeGravity.

## Included Sources

- `internal/eval/demo.go`
- `internal/eval/demo_test.go`
- `internal/eval/golden.go`
- `internal/eval/golden_test.go`
- `internal/eval/graph_replay.go`
- `internal/eval/worker_backlog.go`
- `internal/graph/apply.go`
- `internal/graph/apply_test.go`
- `internal/graph/doc.go`
- `internal/graph/dreaming.go`
- `internal/graph/dreaming_test.go`
- `internal/graph/store_apply.go`
- `internal/graph/store_apply_test.go`
- `internal/reasoning/codex_bridge.go`
- `internal/reasoning/codex_bridge_test.go`
- `internal/reasoning/contracts.go`
- `internal/reasoning/doc.go`
- `internal/reasoning/mock_codex_client.go`
- `internal/reasoning/orchestrator.go`
- `internal/reasoning/orchestrator_test.go`
- `internal/reasoning/stage2_input_preparer.go`
- `internal/reasoning/stage2_input_preparer_test.go`
- `internal/worker/doc.go`
- `internal/worker/processor.go`
- `internal/worker/processor_test.go`
- `internal/worker/stage2_sources.go`
- `internal/worker/stage2_sources_test.go`

## Source Contents


<!-- Source: internal/eval/demo.go | bytes=12764 | lines=337 | sha16=44b821210a6c0815 -->

```go
// ============================================================
// FILE     : internal/eval/demo.go
// PURPOSE  : Runs the local Hermes Memory trust-loop demo eval.
// LAYER    : application
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : RunHermesMemoryDemo
// DEPENDS  : context, encoding/json, strings, time, internal/core, internal/graph, internal/reasoning, internal/recall
// USED_BY  : cmd/cli, internal/eval tests
// ------------------------------------------------------------
// AGENT_NOTE: Demo evals must stay local-only and must not call Hermes, Codex, or external stores.
// ============================================================

package eval

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/graph"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/reasoning"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/recall"
)

const (
	demoTenantID       = "tenant_demo"
	demoWorkspaceID    = "workspace_demo"
	demoSessionID      = "session_demo_next"
	demoHermesActorID  = "agent:hermes-main"
	demoOtherActorID   = "agent:claude"
	demoWrongMemoryID  = "mem_demo_wrong"
	demoWrongRawEvent  = "evt_demo_wrong_memory"
	demoCorrectionJob  = "job_demo_correction"
	demoCorrectionOpID = "op_demo_correction"
)

// RunHermesMemoryDemo executes a deterministic local version of the 5-minute
// Hermes Memory demo. It proves the user-facing trust loop without real Hermes,
// Codex, a database, or network calls.
func RunHermesMemoryDemo(ctx context.Context) *Summary {
	state := newDemoState()
	results := []Result{
		state.demoInitialRecall(ctx),
		state.demoExplainRecalledMemory(ctx),
		state.demoApplyCorrection(ctx),
		state.demoRecallAfterCorrection(ctx),
		state.demoPrivateScopeSeparation(ctx),
	}
	passed := true
	for _, result := range results {
		if !result.Passed {
			passed = false
			break
		}
	}
	return &Summary{Passed: passed, Results: results}
}

type demoState struct {
	now     time.Time
	store   *graphReplayMemoryStore
	notes   noteFixtureStore
	plans   planFixtureStore
	profile profileFixtureStore
}

func newDemoState() *demoState {
	now := time.Date(2026, 4, 25, 6, 55, 0, 0, time.UTC)
	store := newGraphReplayMemoryStore()
	wrong := &core.Memory{
		ID:            demoWrongMemoryID,
		TenantID:      demoTenantID,
		WorkspaceID:   demoWorkspaceID,
		Scope:         core.MemoryScopeWorkspaceShared,
		OwnerEntityID: "workspace:" + demoWorkspaceID,
		Kind:          core.MemoryKindFact,
		ArtifactClass: core.ArtifactClassKnowledge,
		Text:          "Wrong memory: Hermes Memory should lead with generic shared memory kernel language.",
		Fingerprint:   "fp_demo_wrong",
		Confidence:    0.74,
		Status:        core.MemoryStatusActive,
		ValidFrom:     now.Add(-2 * time.Hour),
		LatestFlag:    true,
		MetadataJSON:  json.RawMessage(`{"demo":"wrong_memory"}`),
		CreatedAt:     now.Add(-2 * time.Hour),
		UpdatedAt:     now.Add(-2 * time.Hour),
	}
	store.memories[wrong.ID] = wrong
	store.traces[wrong.ID] = &core.MemoryTrace{
		MemoryID:              wrong.ID,
		RawEventIDs:           []string{demoWrongRawEvent},
		ReasoningJobID:        "job_demo_wrong_memory",
		ReasoningStage:        string(reasoning.StageNameResolve),
		CandidateSnapshotJSON: json.RawMessage(`{"demo":"wrong_memory_candidate"}`),
		AppliedOperationsJSON: json.RawMessage(`[{"operation_id":"op_demo_wrong_memory","kind":"create_memory"}]`),
		CreatedAt:             now.Add(-2 * time.Hour),
	}
	store.memories["mem_demo_other_private"] = &core.Memory{
		ID:            "mem_demo_other_private",
		TenantID:      demoTenantID,
		WorkspaceID:   demoWorkspaceID,
		Scope:         core.MemoryScopeAgentPrivate,
		OwnerEntityID: demoOtherActorID,
		Kind:          core.MemoryKindPreference,
		ArtifactClass: core.ArtifactClassKnowledge,
		Text:          "Other agent private memory: do not show this to Hermes.",
		Fingerprint:   "fp_demo_other_private",
		Confidence:    0.93,
		Status:        core.MemoryStatusActive,
		ValidFrom:     now.Add(-time.Hour),
		LatestFlag:    true,
		MetadataJSON:  json.RawMessage(`{}`),
		CreatedAt:     now.Add(-time.Hour),
		UpdatedAt:     now.Add(-time.Hour),
	}
	return &demoState{
		now:   now,
		store: store,
		notes: noteFixtureStore{{
			ID:            "note_demo_project_rule",
			TenantID:      demoTenantID,
			WorkspaceID:   demoWorkspaceID,
			NoteKind:      "project_rule",
			Scope:         core.MemoryScopeWorkspaceShared,
			OwnerEntityID: "workspace:" + demoWorkspaceID,
			Text:          "Project rule: Hermes Memory must show scope, source, and correction path.",
			Pinned:        true,
			CreatedAt:     now.Add(-3 * time.Hour),
			UpdatedAt:     now.Add(-3 * time.Hour),
		}},
		plans: planFixtureStore{{
			ID:            "plan_demo_trust_loop",
			TenantID:      demoTenantID,
			WorkspaceID:   demoWorkspaceID,
			Title:         "Active plan: demo recall preview, explain, correction, supersession, and scope separation.",
			Status:        "active",
			Scope:         core.MemoryScopeWorkspaceShared,
			OwnerEntityID: "workspace:" + demoWorkspaceID,
			EvidenceJSON:  json.RawMessage(`{"demo":"trust_loop"}`),
			CreatedAt:     now.Add(-3 * time.Hour),
			UpdatedAt:     now.Add(-3 * time.Hour),
		}},
		profile: profileFixtureStore{{
			EntityID:   demoHermesActorID,
			Scope:      core.MemoryScopeAgentPrivate,
			StaticJSON: json.RawMessage(`{"demo":"Hermes operator expects concise memory proof"}`),
			UpdatedAt:  now.Add(-time.Hour),
		}},
	}
}

func (s *demoState) demoInitialRecall(ctx context.Context) Result {
	resp, err := s.prefetch(ctx, "Hermes Memory correction path")
	if err != nil {
		return failedDemoResult("demo initial recall shows rule plan and trust metadata", "prefetch failed: "+err.Error())
	}
	observed := observe(resp)
	errs := compareExpectation(Expectation{
		Contains:  []string{"Project rule: Hermes Memory", "Active plan: demo recall preview", "generic shared memory kernel language"},
		MaxTokens: 2200,
	}, observed)
	errs = append(errs, requireBlockMetadata(resp, "pinned_note", core.MemoryScopeWorkspaceShared, "notes", "stored")...)
	errs = append(errs, requireBlockMetadata(resp, "active_plan", core.MemoryScopeWorkspaceShared, "plans", "stored")...)
	errs = append(errs, requireBlockMetadata(resp, "memory", core.MemoryScopeWorkspaceShared, "memories", "stored")...)
	return demoResult("demo initial recall shows rule plan and trust metadata", observed, errs)
}

func (s *demoState) demoExplainRecalledMemory(ctx context.Context) Result {
	resp, err := s.store.ExplainMemory(ctx, &core.ExplainMemoryRequest{
		TenantID:    demoTenantID,
		WorkspaceID: demoWorkspaceID,
		MemoryID:    demoWrongMemoryID,
	})
	if err != nil {
		return failedDemoResult("demo explain shows recalled memory provenance", "explain failed: "+err.Error())
	}
	observed := Observed{
		BlockKinds: []string{"explain_memory"},
		Sources:    []string{"memory_trace"},
		Text:       strings.Join(append([]string{resp.MemoryID, resp.Trace.ReasoningJobID}, resp.Trace.RawEventIDs...), "\n"),
	}
	var errs []string
	for _, want := range []string{demoWrongMemoryID, "job_demo_wrong_memory", demoWrongRawEvent} {
		if !strings.Contains(observed.Text, want) {
			errs = append(errs, fmt.Sprintf("missing explain evidence %q", want))
		}
	}
	return demoResult("demo explain shows recalled memory provenance", observed, errs)
}

func (s *demoState) demoApplyCorrection(ctx context.Context) Result {
	engine, err := graph.NewStoreBackedApplyEngine(s.store)
	if err != nil {
		return failedDemoResult("demo correction writes supersession", "apply engine failed: "+err.Error())
	}
	result, err := engine.Apply(ctx, &graph.ApplyRequest{
		JobID:       demoCorrectionJob,
		TenantID:    demoTenantID,
		WorkspaceID: demoWorkspaceID,
		RawEventIDs: []string{"evt_demo_operator_correction"},
		Reasoning: &reasoning.ProcessTurnResult{Stage2: reasoning.Stage2Output{
			Operations: []reasoning.GraphOperation{{
				OperationID: demoCorrectionOpID,
				Kind:        reasoning.OperationKindUpdateMemory,
				Memory: &reasoning.MemoryMutation{
					TargetID:      demoWrongMemoryID,
					Kind:          core.MemoryKindFact,
					ArtifactClass: core.ArtifactClassKnowledge,
					Scope:         core.MemoryScopeWorkspaceShared,
					OwnerEntityID: "workspace:" + demoWorkspaceID,
					Text:          "Corrected memory: Hermes Memory should lead with continuity, proof, and fix-memory-once trust loop.",
					Confidence:    0.97,
					MetadataJSON:  json.RawMessage(`{"source":"operator_correction","demo":"trust_loop"}`),
				},
				Edge: &reasoning.EdgeMutation{
					ToMemoryID: demoWrongMemoryID,
					EdgeKind:   core.EdgeKindUpdates,
					Confidence: 0.99,
				},
				RawEventIDs: []string{"evt_demo_operator_correction"},
				Metadata:    json.RawMessage(`{"source":"operator_correction"}`),
			}},
			ProfileDelta: json.RawMessage(`{}`),
			PlanDelta:    json.RawMessage(`{}`),
			Trace: reasoning.Trace{
				SchemaVersion: "vibegravity.eval.demo.v1",
				Stage:         reasoning.StageNameResolve,
				MetadataJSON:  json.RawMessage(`{"source":"operator_correction"}`),
			},
		}},
	})
	if err != nil {
		return failedDemoResult("demo correction writes supersession", "apply failed: "+err.Error())
	}
	observed := Observed{
		BlockKinds: []string{"update_memory"},
		Sources:    []string{"memories", "memory_trace", "memory_edges"},
		Text:       strings.Join(result.MemoryIDs, "\n"),
	}
	var errs []string
	if result.AppliedOperationCount != 1 || !result.TraceWritten || len(result.MemoryIDs) != 1 {
		errs = append(errs, fmt.Sprintf("unexpected apply result: %#v", result))
	}
	if !s.store.memoryHasStatus(demoWrongMemoryID, core.MemoryStatusSuperseded, false) {
		errs = append(errs, "wrong memory was not superseded")
	}
	return demoResult("demo correction writes supersession", observed, errs)
}

func (s *demoState) demoRecallAfterCorrection(ctx context.Context) Result {
	resp, err := s.prefetch(ctx, "Hermes Memory trust loop")
	if err != nil {
		return failedDemoResult("demo next recall uses correction", "prefetch failed: "+err.Error())
	}
	observed := observe(resp)
	errs := compareExpectation(Expectation{
		Contains:    []string{"Corrected memory: Hermes Memory should lead with continuity"},
		NotContains: []string{"generic shared memory kernel language"},
		MaxTokens:   2200,
	}, observed)
	return demoResult("demo next recall uses correction", observed, errs)
}

func (s *demoState) demoPrivateScopeSeparation(ctx context.Context) Result {
	resp, err := s.prefetch(ctx, "private memory")
	if err != nil {
		return failedDemoResult("demo private scope separation", "prefetch failed: "+err.Error())
	}
	observed := observe(resp)
	errs := compareExpectation(Expectation{
		NotContains: []string{"Other agent private memory"},
		MaxTokens:   2200,
	}, observed)
	return demoResult("demo private scope separation", observed, errs)
}

func (s *demoState) prefetch(ctx context.Context, query string) (*core.PrefetchResponse, error) {
	assembler := recall.NewAssembler(recall.Dependencies{
		Notes:     s.notes,
		Plans:     s.plans,
		Memories:  s.store,
		Documents: documentFixtureStore{},
		Profiles:  s.profile,
		Summaries: summaryFixtureStore{},
		Clock:     func() time.Time { return s.now },
	})
	return assembler.Prefetch(ctx, &core.PrefetchRequest{
		TenantID:     demoTenantID,
		WorkspaceID:  demoWorkspaceID,
		SessionID:    demoSessionID,
		ActorID:      demoHermesActorID,
		Query:        query,
		BudgetTokens: 2200,
		Mode:         "default",
	})
}

func requireBlockMetadata(resp *core.PrefetchResponse, kind string, scope core.MemoryScope, source string, freshness string) []string {
	for _, block := range resp.Blocks {
		if block.Kind != kind {
			continue
		}
		var errs []string
		if block.Scope != scope {
			errs = append(errs, fmt.Sprintf("%s scope got %q want %q", kind, block.Scope, scope))
		}
		if block.Source != source {
			errs = append(errs, fmt.Sprintf("%s source got %q want %q", kind, block.Source, source))
		}
		if block.Freshness != freshness {
			errs = append(errs, fmt.Sprintf("%s freshness got %q want %q", kind, block.Freshness, freshness))
		}
		if strings.TrimSpace(block.SourceID) == "" {
			errs = append(errs, fmt.Sprintf("%s source_id is required", kind))
		}
		return errs
	}
	return []string{fmt.Sprintf("missing block kind %q", kind)}
}

func demoResult(name string, observed Observed, errs []string) Result {
	return Result{
		Scenario: name,
		Passed:   len(errs) == 0,
		Errors:   errs,
		Observed: observed,
	}
}

func failedDemoResult(name string, errs ...string) Result {
	return Result{Scenario: name, Passed: false, Errors: errs}
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



<!-- Source: internal/eval/golden.go | bytes=16020 | lines=451 | sha16=2a2536d309263c58 -->

```go
// ============================================================
// FILE     : internal/eval/golden.go
// PURPOSE  : Runs golden recall scenarios that catch memory quality regressions.
// LAYER    : application
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : ScenarioFile, Scenario, Result, Summary, RunFile, RunScenarios
// DEPENDS  : encoding/json, os, strings, internal/core, internal/recall, internal/store
// USED_BY  : cmd/cli, internal/eval tests, Makefile eval target
// ------------------------------------------------------------
// AGENT_NOTE: Eval fixtures must stay deterministic and must not call real Codex or external stores.
// ============================================================

// Package eval runs deterministic, local-only quality checks for recall behavior.
package eval

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"slices"
	"strings"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/recall"
)

// ScenarioFile is the top-level JSON fixture for golden eval scenarios.
type ScenarioFile struct {
	Scenarios              []Scenario              `json:"scenarios"`
	GraphReplayScenarios   []GraphReplayScenario   `json:"graph_replay_scenarios"`
	WorkerBacklogScenarios []WorkerBacklogScenario `json:"worker_backlog_scenarios"`
}

// Scenario defines one deterministic recall quality check.
type Scenario struct {
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Prefetch    core.PrefetchRequest `json:"prefetch"`
	Fixtures    Fixtures             `json:"fixtures"`
	Expect      Expectation          `json:"expect"`
}

// Fixtures contains the in-memory store rows visible to one scenario.
type Fixtures struct {
	Notes            []core.Note                `json:"notes"`
	Plans            []core.Plan                `json:"plans"`
	Profiles         []core.Profile             `json:"profiles"`
	SessionSummaries []core.SessionSummary      `json:"session_summaries"`
	Memories         []core.MemoryResult        `json:"memories"`
	Documents        []core.DocumentChunkResult `json:"documents"`
}

// Expectation describes what a scenario must preserve.
type Expectation struct {
	BlockKinds       []string                   `json:"block_kinds"`
	BlockMetadata    []BlockMetadataExpectation `json:"block_metadata"`
	Contains         []string                   `json:"contains"`
	NotContains      []string                   `json:"not_contains"`
	Sources          []string                   `json:"sources"`
	MaxTokens        int                        `json:"max_tokens"`
	AllowExtraBlocks bool                       `json:"allow_extra_blocks"`
}

// BlockMetadataExpectation asserts operator-visible trust metadata on recall blocks.
type BlockMetadataExpectation struct {
	Kind          string           `json:"kind"`
	Scope         core.MemoryScope `json:"scope,omitempty"`
	Source        string           `json:"source,omitempty"`
	SourceID      string           `json:"source_id,omitempty"`
	Status        string           `json:"status,omitempty"`
	Freshness     string           `json:"freshness,omitempty"`
	OwnerEntityID string           `json:"owner_entity_id,omitempty"`
}

// Result is the pass/fail record for one scenario.
type Result struct {
	Scenario string   `json:"scenario"`
	Passed   bool     `json:"passed"`
	Errors   []string `json:"errors,omitempty"`
	Observed Observed `json:"observed"`
}

// Observed captures the deterministic recall output used for comparison.
type Observed struct {
	BlockKinds []string        `json:"block_kinds"`
	Blocks     []ObservedBlock `json:"blocks"`
	Sources    []string        `json:"sources"`
	Tokens     int             `json:"tokens"`
	Text       string          `json:"text"`
}

// ObservedBlock captures trust metadata for eval failure reporting.
type ObservedBlock struct {
	Kind          string           `json:"kind"`
	Scope         core.MemoryScope `json:"scope,omitempty"`
	Source        string           `json:"source,omitempty"`
	SourceID      string           `json:"source_id,omitempty"`
	Status        string           `json:"status,omitempty"`
	Freshness     string           `json:"freshness,omitempty"`
	OwnerEntityID string           `json:"owner_entity_id,omitempty"`
}

// Summary reports aggregate eval status.
type Summary struct {
	Passed  bool     `json:"passed"`
	Results []Result `json:"results"`
}

// RunFile loads and runs golden scenarios from a JSON fixture file.
func RunFile(ctx context.Context, path string) (*Summary, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read golden scenarios: %w", err)
	}
	var file ScenarioFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("parse golden scenarios: %w", err)
	}
	summary := RunScenarios(ctx, file.Scenarios)
	graphSummary := RunGraphReplayScenarios(ctx, file.GraphReplayScenarios)
	summary.Results = append(summary.Results, graphSummary.Results...)
	summary.Passed = summary.Passed && graphSummary.Passed
	workerSummary := RunWorkerBacklogScenarios(ctx, file.WorkerBacklogScenarios)
	summary.Results = append(summary.Results, workerSummary.Results...)
	summary.Passed = summary.Passed && workerSummary.Passed
	return summary, nil
}

// RunScenarios executes scenarios against the recall assembler with in-memory stores.
func RunScenarios(ctx context.Context, scenarios []Scenario) *Summary {
	results := make([]Result, 0, len(scenarios))
	passed := true
	for _, scenario := range scenarios {
		result := runScenario(ctx, scenario)
		if !result.Passed {
			passed = false
		}
		results = append(results, result)
	}
	return &Summary{Passed: passed, Results: results}
}

func runScenario(ctx context.Context, scenario Scenario) Result {
	errs := validateScenario(scenario)
	assembler := recall.NewAssembler(recall.Dependencies{
		Notes:     noteFixtureStore(scenario.Fixtures.Notes),
		Plans:     planFixtureStore(scenario.Fixtures.Plans),
		Profiles:  profileFixtureStore(scenario.Fixtures.Profiles),
		Summaries: summaryFixtureStore(scenario.Fixtures.SessionSummaries),
		Memories:  memoryFixtureStore(scenario.Fixtures.Memories),
		Documents: documentFixtureStore(scenario.Fixtures.Documents),
	})
	resp, err := assembler.Prefetch(ctx, &scenario.Prefetch)
	if err != nil {
		errs = append(errs, "prefetch failed: "+err.Error())
		return Result{Scenario: scenario.Name, Passed: false, Errors: errs}
	}

	observed := observe(resp)
	errs = append(errs, compareExpectation(scenario.Expect, observed)...)
	return Result{
		Scenario: scenario.Name,
		Passed:   len(errs) == 0,
		Errors:   errs,
		Observed: observed,
	}
}

func validateScenario(scenario Scenario) []string {
	var errs []string
	if strings.TrimSpace(scenario.Name) == "" {
		errs = append(errs, "scenario name is required")
	}
	return errs
}

func observe(resp *core.PrefetchResponse) Observed {
	if resp == nil {
		return Observed{}
	}
	kinds := make([]string, 0, len(resp.Blocks))
	blocks := make([]ObservedBlock, 0, len(resp.Blocks))
	texts := make([]string, 0, len(resp.Blocks))
	for _, block := range resp.Blocks {
		kinds = append(kinds, block.Kind)
		blocks = append(blocks, ObservedBlock{
			Kind:          block.Kind,
			Scope:         block.Scope,
			Source:        block.Source,
			SourceID:      block.SourceID,
			Status:        block.Status,
			Freshness:     block.Freshness,
			OwnerEntityID: block.OwnerEntityID,
		})
		texts = append(texts, block.Text)
	}
	return Observed{
		BlockKinds: kinds,
		Blocks:     blocks,
		Sources:    append([]string(nil), resp.Meta.Sources...),
		Tokens:     resp.Meta.EstimatedTokens,
		Text:       strings.Join(texts, "\n"),
	}
}

func compareExpectation(expect Expectation, observed Observed) []string {
	var errs []string
	if len(expect.BlockKinds) > 0 {
		if expect.AllowExtraBlocks {
			if !hasKindPrefix(observed.BlockKinds, expect.BlockKinds) {
				errs = append(errs, fmt.Sprintf("block kinds prefix got %v want %v", observed.BlockKinds, expect.BlockKinds))
			}
		} else if !reflect.DeepEqual(observed.BlockKinds, expect.BlockKinds) {
			errs = append(errs, fmt.Sprintf("block kinds got %v want %v", observed.BlockKinds, expect.BlockKinds))
		}
	}
	if len(expect.BlockMetadata) > 0 {
		errs = append(errs, compareBlockMetadata(expect.BlockMetadata, observed.Blocks)...)
	}
	for _, needle := range expect.Contains {
		if !strings.Contains(observed.Text, needle) {
			errs = append(errs, fmt.Sprintf("missing expected text %q", needle))
		}
	}
	for _, needle := range expect.NotContains {
		if strings.Contains(observed.Text, needle) {
			errs = append(errs, fmt.Sprintf("unexpected text %q", needle))
		}
	}
	if len(expect.Sources) > 0 && !reflect.DeepEqual(observed.Sources, expect.Sources) {
		errs = append(errs, fmt.Sprintf("sources got %v want %v", observed.Sources, expect.Sources))
	}
	if expect.MaxTokens > 0 && observed.Tokens > expect.MaxTokens {
		errs = append(errs, fmt.Sprintf("tokens got %d max %d", observed.Tokens, expect.MaxTokens))
	}
	return errs
}

func compareBlockMetadata(expect []BlockMetadataExpectation, observed []ObservedBlock) []string {
	if len(observed) < len(expect) {
		return []string{fmt.Sprintf("block metadata count got %d want at least %d", len(observed), len(expect))}
	}
	var errs []string
	for i, want := range expect {
		got := observed[i]
		if want.Kind != "" && got.Kind != want.Kind {
			errs = append(errs, fmt.Sprintf("block[%d] kind got %q want %q", i, got.Kind, want.Kind))
		}
		if want.Scope != "" && got.Scope != want.Scope {
			errs = append(errs, fmt.Sprintf("block[%d] scope got %q want %q", i, got.Scope, want.Scope))
		}
		if want.Source != "" && got.Source != want.Source {
			errs = append(errs, fmt.Sprintf("block[%d] source got %q want %q", i, got.Source, want.Source))
		}
		if want.SourceID != "" && got.SourceID != want.SourceID {
			errs = append(errs, fmt.Sprintf("block[%d] source_id got %q want %q", i, got.SourceID, want.SourceID))
		}
		if want.Status != "" && got.Status != want.Status {
			errs = append(errs, fmt.Sprintf("block[%d] status got %q want %q", i, got.Status, want.Status))
		}
		if want.Freshness != "" && got.Freshness != want.Freshness {
			errs = append(errs, fmt.Sprintf("block[%d] freshness got %q want %q", i, got.Freshness, want.Freshness))
		}
		if want.OwnerEntityID != "" && got.OwnerEntityID != want.OwnerEntityID {
			errs = append(errs, fmt.Sprintf("block[%d] owner_entity_id got %q want %q", i, got.OwnerEntityID, want.OwnerEntityID))
		}
	}
	return errs
}

func hasKindPrefix(got []string, want []string) bool {
	if len(got) < len(want) {
		return false
	}
	return reflect.DeepEqual(got[:len(want)], want)
}

type noteFixtureStore []core.Note

func (s noteFixtureStore) AddNote(context.Context, *core.Note) error {
	return core.ErrNotImplemented
}

func (s noteFixtureStore) ListPinnedNotes(_ context.Context, req *core.ListPinnedNotesRequest) ([]*core.Note, error) {
	notes := make([]*core.Note, 0, len(s))
	for i := range s {
		note := s[i]
		if note.TenantID != req.TenantID || note.WorkspaceID != req.WorkspaceID || !note.Pinned {
			continue
		}
		if !scopeVisible(note.Scope, note.OwnerEntityID, req.OwnerEntityID, req.Scopes) {
			continue
		}
		notes = append(notes, &note)
	}
	return notes, nil
}

type planFixtureStore []core.Plan

func (s planFixtureStore) CreatePlan(context.Context, *core.Plan, []*core.PlanItem) error {
	return core.ErrNotImplemented
}

func (s planFixtureStore) UpdatePlan(context.Context, *core.Plan, []*core.PlanItem) error {
	return core.ErrNotImplemented
}

func (s planFixtureStore) GetActivePlans(_ context.Context, req *core.GetActivePlansRequest) ([]*core.Plan, error) {
	plans := make([]*core.Plan, 0, len(s))
	for i := range s {
		plan := s[i]
		if plan.TenantID != req.TenantID || plan.WorkspaceID != req.WorkspaceID {
			continue
		}
		if !scopeVisible(plan.Scope, plan.OwnerEntityID, req.OwnerEntityID, req.Scopes) {
			continue
		}
		plans = append(plans, &plan)
	}
	return plans, nil
}

type profileFixtureStore []core.Profile

func (s profileFixtureStore) GetProfile(_ context.Context, entityID string, scope core.MemoryScope) (*core.Profile, error) {
	for i := range s {
		profile := s[i]
		if profile.EntityID == entityID && profile.Scope == scope {
			return &profile, nil
		}
	}
	return nil, core.ErrNotFound
}

func (s profileFixtureStore) UpsertProfile(context.Context, *core.Profile) error {
	return core.ErrNotImplemented
}

type summaryFixtureStore []core.SessionSummary

func (s summaryFixtureStore) UpsertSessionSummary(context.Context, *core.SessionSummary) error {
	return core.ErrNotImplemented
}

func (s summaryFixtureStore) GetSessionSummary(_ context.Context, sessionID string) (*core.SessionSummary, error) {
	for i := range s {
		summary := s[i]
		if summary.SessionID == sessionID {
			return &summary, nil
		}
	}
	return nil, core.ErrNotFound
}

type memoryFixtureStore []core.MemoryResult

func (s memoryFixtureStore) UpsertMemory(context.Context, *core.Memory) error {
	return core.ErrNotImplemented
}

func (s memoryFixtureStore) GetMemory(context.Context, string) (*core.Memory, error) {
	return nil, core.ErrNotImplemented
}

func (s memoryFixtureStore) SearchMemories(_ context.Context, req *core.SearchMemoriesRequest) (*core.SearchMemoriesResponse, error) {
	results := make([]core.MemoryResult, 0, len(s))
	for _, memory := range s {
		if !scopeVisible(memory.Scope, memory.OwnerEntityID, req.OwnerEntityID, req.Scopes) {
			continue
		}
		if !artifactClassVisible(memory.ArtifactClass, req.ArtifactClasses) {
			continue
		}
		if query := strings.TrimSpace(strings.ToLower(req.Query)); query != "" {
			if !strings.Contains(strings.ToLower(memory.Text), query) && !wordOverlap(strings.ToLower(memory.Text), query) {
				continue
			}
		}
		results = append(results, memory)
	}
	return &core.SearchMemoriesResponse{Memories: results}, nil
}

func (s memoryFixtureStore) UpsertMemoryEdge(context.Context, *core.MemoryEdge) error {
	return core.ErrNotImplemented
}

func (s memoryFixtureStore) WriteMemoryTrace(context.Context, *core.MemoryTrace) error {
	return core.ErrNotImplemented
}

func (s memoryFixtureStore) ExplainMemory(context.Context, *core.ExplainMemoryRequest) (*core.ExplainMemoryResponse, error) {
	return nil, core.ErrNotImplemented
}

type documentFixtureStore []core.DocumentChunkResult

func (s documentFixtureStore) AddDocumentWithChunks(context.Context, *core.Document, []*core.DocumentChunk) error {
	return core.ErrNotImplemented
}

func (s documentFixtureStore) AddDocument(context.Context, *core.Document) error {
	return core.ErrNotImplemented
}

func (s documentFixtureStore) AddDocumentChunks(context.Context, []*core.DocumentChunk) error {
	return core.ErrNotImplemented
}

func (s documentFixtureStore) SearchDocuments(_ context.Context, req *core.SearchDocumentsRequest) (*core.SearchDocumentsResponse, error) {
	chunks := make([]core.DocumentChunkResult, 0, len(s))
	query := strings.TrimSpace(strings.ToLower(req.Query))
	for _, chunk := range s {
		if query == "" || strings.Contains(strings.ToLower(chunk.Text), query) || wordOverlap(strings.ToLower(chunk.Text), query) {
			chunks = append(chunks, chunk)
		}
	}
	return &core.SearchDocumentsResponse{Chunks: chunks}, nil
}

func scopeVisible(scope core.MemoryScope, owner string, actor string, scopes []core.MemoryScope) bool {
	if !slices.Contains(scopes, scope) {
		return false
	}
	if scope == core.MemoryScopeAgentPrivate {
		return owner == actor
	}
	return scope != core.MemoryScopeGroupShared
}

func artifactClassVisible(class core.ArtifactClass, classes []core.ArtifactClass) bool {
	return len(classes) == 0 || slices.Contains(classes, class)
}

func wordOverlap(text string, query string) bool {
	for _, word := range strings.Fields(query) {
		word = strings.Trim(word, ".,:;!?()[]{}\"'")
		if len(word) < 4 {
			continue
		}
		if strings.Contains(text, word) {
			return true
		}
	}
	return false
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



<!-- Source: internal/eval/graph_replay.go | bytes=21379 | lines=626 | sha16=a6ac5712d6a963e0 -->

```go
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

```



<!-- Source: internal/eval/worker_backlog.go | bytes=19985 | lines=599 | sha16=32a9019d93d3ed65 -->

```go
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

```



<!-- Source: internal/graph/apply.go | bytes=11962 | lines=314 | sha16=88a0b2fc50cb1369 -->

```go
// ============================================================
// FILE     : internal/graph/apply.go
// PURPOSE  : Defines the graph apply engine contract and a validating no-op implementation.
// LAYER    : application
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : ApplyEngine, ApplyRequest, ApplyResult, NoopApplyEngine, NewNoopApplyEngine
// DEPENDS  : context, encoding/json, fmt, slices, internal/core, internal/reasoning
// USED_BY  : internal/worker, cmd/worker, tests
// ------------------------------------------------------------
// AGENT_NOTE: Apply must validate structured Stage 2 output before any memory, edge, profile, or trace write.
// ============================================================

package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/reasoning"
)

// ApplyEngine validates and applies structured Stage 2 reasoning output.
type ApplyEngine interface {
	Apply(ctx context.Context, req *ApplyRequest) (*ApplyResult, error)
}

// ApplyRequest is the apply boundary input for one worker job.
type ApplyRequest struct {
	JobID       string                       `json:"job_id"`
	TenantID    string                       `json:"tenant_id"`
	WorkspaceID string                       `json:"workspace_id"`
	RawEventIDs []string                     `json:"raw_event_ids"`
	Reasoning   *reasoning.ProcessTurnResult `json:"reasoning"`
}

// ApplyResult reports what the apply engine committed.
type ApplyResult struct {
	AppliedOperationCount int      `json:"applied_operation_count"`
	MemoryIDs             []string `json:"memory_ids"`
	TraceWritten          bool     `json:"trace_written"`
}

// NoopApplyEngine validates the apply request without writing derived state.
type NoopApplyEngine struct{}

// NewNoopApplyEngine creates a no-op graph apply engine for worker wiring tests.
func NewNoopApplyEngine() *NoopApplyEngine {
	return &NoopApplyEngine{}
}

// Apply validates schema-shaped Stage 2 output and intentionally commits nothing.
func (e *NoopApplyEngine) Apply(_ context.Context, req *ApplyRequest) (*ApplyResult, error) {
	if err := validateApplyRequest(req); err != nil {
		return nil, err
	}
	return &ApplyResult{
		AppliedOperationCount: 0,
		MemoryIDs:             []string{},
		TraceWritten:          false,
	}, nil
}

func validateApplyRequest(req *ApplyRequest) error {
	if req == nil {
		return fmt.Errorf("%w: apply request is required", core.ErrInvalidArgument)
	}
	if req.JobID == "" {
		return fmt.Errorf("%w: apply job_id is required", core.ErrInvalidArgument)
	}
	if req.TenantID == "" {
		return fmt.Errorf("%w: apply tenant_id is required", core.ErrInvalidArgument)
	}
	if req.WorkspaceID == "" {
		return fmt.Errorf("%w: apply workspace_id is required", core.ErrInvalidArgument)
	}
	if len(req.RawEventIDs) == 0 {
		return fmt.Errorf("%w: apply raw_event_ids are required", core.ErrInvalidArgument)
	}
	if req.Reasoning == nil {
		return fmt.Errorf("%w: apply reasoning result is required", core.ErrInvalidArgument)
	}
	if req.Reasoning.Stage2.Trace.SchemaVersion == "" {
		return fmt.Errorf("%w: apply resolve trace schema_version is required", core.ErrInvalidArgument)
	}
	if req.Reasoning.Stage2.Trace.Stage != reasoning.StageNameResolve {
		return fmt.Errorf("%w: apply requires resolve-stage trace", core.ErrInvalidArgument)
	}
	if err := validateJSONObject("profile_delta", req.Reasoning.Stage2.ProfileDelta); err != nil {
		return err
	}
	if err := validateJSONObject("plan_delta", req.Reasoning.Stage2.PlanDelta); err != nil {
		return err
	}
	if err := validateJSONObject("trace.metadata_json", req.Reasoning.Stage2.Trace.MetadataJSON); err != nil {
		return err
	}
	for i, operation := range req.Reasoning.Stage2.Operations {
		if err := validateOperation(i, operation, req.RawEventIDs); err != nil {
			return err
		}
	}
	return nil
}

func validateOperation(index int, operation reasoning.GraphOperation, applyRawEventIDs []string) error {
	if operation.OperationID == "" {
		return fmt.Errorf("%w: operations[%d].operation_id is required", core.ErrInvalidArgument, index)
	}
	if operation.Kind == "" {
		return fmt.Errorf("%w: operations[%d].kind is required", core.ErrInvalidArgument, index)
	}
	if !isSupportedOperationKind(operation.Kind) {
		return fmt.Errorf("%w: operations[%d].kind is unsupported", core.ErrInvalidArgument, index)
	}
	if len(operation.RawEventIDs) == 0 {
		return fmt.Errorf("%w: operations[%d].raw_event_ids are required", core.ErrInvalidArgument, index)
	}
	for _, rawEventID := range operation.RawEventIDs {
		if rawEventID == "" {
			return fmt.Errorf("%w: operations[%d].raw_event_ids cannot contain empty ids", core.ErrInvalidArgument, index)
		}
		if !slices.Contains(applyRawEventIDs, rawEventID) {
			return fmt.Errorf("%w: operations[%d].raw_event_ids must reference the apply bundle", core.ErrInvalidArgument, index)
		}
	}
	if err := validateJSONObject(fmt.Sprintf("operations[%d].metadata", index), operation.Metadata); err != nil {
		return err
	}

	switch operation.Kind {
	case reasoning.OperationKindCreateMemory:
		return validateMemoryMutation(index, operation.Memory, false)
	case reasoning.OperationKindUpdateMemory:
		if err := validateMemoryMutation(index, operation.Memory, true); err != nil {
			return err
		}
		return validateEdgeMutation(index, operation.Edge, core.EdgeKindUpdates, operation.Memory)
	case reasoning.OperationKindExtendMemory:
		if err := validateMemoryMutation(index, operation.Memory, true); err != nil {
			return err
		}
		return validateEdgeMutation(index, operation.Edge, core.EdgeKindExtends, operation.Memory)
	case reasoning.OperationKindArchiveMemory:
		return validateArchiveMutation(index, operation.Memory)
	default:
		return fmt.Errorf("%w: operations[%d].kind is unsupported", core.ErrInvalidArgument, index)
	}
}

func validateMemoryMutation(index int, memory *reasoning.MemoryMutation, targetRequired bool) error {
	if memory == nil {
		return fmt.Errorf("%w: operations[%d].memory is required", core.ErrInvalidArgument, index)
	}
	if targetRequired && memory.TargetID == "" {
		return fmt.Errorf("%w: operations[%d].memory.target_id is required", core.ErrInvalidArgument, index)
	}
	if !isSupportedMemoryKind(memory.Kind) {
		return fmt.Errorf("%w: operations[%d].memory.kind is unsupported", core.ErrInvalidArgument, index)
	}
	if !isSupportedArtifactClass(memory.ArtifactClass) {
		return fmt.Errorf("%w: operations[%d].memory.artifact_class is unsupported", core.ErrInvalidArgument, index)
	}
	if !isSupportedScope(memory.Scope) {
		return fmt.Errorf("%w: operations[%d].memory.scope is required", core.ErrInvalidArgument, index)
	}
	if memory.Scope == core.MemoryScopeGroupShared && (memory.GroupID == nil || *memory.GroupID == "") {
		return fmt.Errorf("%w: operations[%d].memory.group_id is required for group_shared scope", core.ErrInvalidArgument, index)
	}
	if memory.OwnerEntityID == "" {
		return fmt.Errorf("%w: operations[%d].memory.owner_entity_id is required", core.ErrInvalidArgument, index)
	}
	if memory.Text == "" {
		return fmt.Errorf("%w: operations[%d].memory.text is required", core.ErrInvalidArgument, index)
	}
	if !isValidConfidence(memory.Confidence) {
		return fmt.Errorf("%w: operations[%d].memory.confidence must be greater than 0 and less than or equal to 1", core.ErrInvalidArgument, index)
	}
	if err := validateJSONObject(fmt.Sprintf("operations[%d].memory.metadata_json", index), memory.MetadataJSON); err != nil {
		return err
	}
	return nil
}

func validateArchiveMutation(index int, memory *reasoning.MemoryMutation) error {
	if memory == nil {
		return fmt.Errorf("%w: operations[%d].memory is required", core.ErrInvalidArgument, index)
	}
	if memory.TargetID == "" && memory.MemoryID == "" {
		return fmt.Errorf("%w: operations[%d].memory target is required", core.ErrInvalidArgument, index)
	}
	if !isSupportedScope(memory.Scope) {
		return fmt.Errorf("%w: operations[%d].memory.scope is required", core.ErrInvalidArgument, index)
	}
	if memory.Scope == core.MemoryScopeGroupShared && (memory.GroupID == nil || *memory.GroupID == "") {
		return fmt.Errorf("%w: operations[%d].memory.group_id is required for group_shared scope", core.ErrInvalidArgument, index)
	}
	if memory.OwnerEntityID == "" {
		return fmt.Errorf("%w: operations[%d].memory.owner_entity_id is required", core.ErrInvalidArgument, index)
	}
	if memory.Confidence != 0 && !isValidConfidence(memory.Confidence) {
		return fmt.Errorf("%w: operations[%d].memory.confidence must be greater than 0 and less than or equal to 1", core.ErrInvalidArgument, index)
	}
	if err := validateJSONObject(fmt.Sprintf("operations[%d].memory.metadata_json", index), memory.MetadataJSON); err != nil {
		return err
	}
	return nil
}

func validateEdgeMutation(index int, edge *reasoning.EdgeMutation, expectedKind core.EdgeKind, memory *reasoning.MemoryMutation) error {
	if edge == nil {
		return fmt.Errorf("%w: operations[%d].edge is required", core.ErrInvalidArgument, index)
	}
	if edge.EdgeKind != expectedKind {
		return fmt.Errorf("%w: operations[%d].edge.edge_kind must be %q", core.ErrInvalidArgument, index, expectedKind)
	}
	if edge.ToMemoryID == "" {
		return fmt.Errorf("%w: operations[%d].edge.to_memory_id is required", core.ErrInvalidArgument, index)
	}
	if edge.ToMemoryID != memory.TargetID {
		return fmt.Errorf("%w: operations[%d].edge.to_memory_id must match memory.target_id", core.ErrInvalidArgument, index)
	}
	if memory.MemoryID != "" && edge.FromMemoryID != "" && edge.FromMemoryID != memory.MemoryID {
		return fmt.Errorf("%w: operations[%d].edge.from_memory_id must match memory.memory_id", core.ErrInvalidArgument, index)
	}
	if edge.FromMemoryID == edge.ToMemoryID {
		return fmt.Errorf("%w: operations[%d].edge cannot target itself", core.ErrInvalidArgument, index)
	}
	if !isValidConfidence(edge.Confidence) {
		return fmt.Errorf("%w: operations[%d].edge.confidence must be greater than 0 and less than or equal to 1", core.ErrInvalidArgument, index)
	}
	return nil
}

func isSupportedOperationKind(kind reasoning.OperationKind) bool {
	switch kind {
	case reasoning.OperationKindCreateMemory,
		reasoning.OperationKindUpdateMemory,
		reasoning.OperationKindExtendMemory,
		reasoning.OperationKindArchiveMemory:
		return true
	default:
		return false
	}
}

func isSupportedMemoryKind(kind core.MemoryKind) bool {
	switch kind {
	case core.MemoryKindFact,
		core.MemoryKindPreference,
		core.MemoryKindTrait,
		core.MemoryKindGoal,
		core.MemoryKindConstraint,
		core.MemoryKindRelationship,
		core.MemoryKindDecision,
		core.MemoryKindProcedure,
		core.MemoryKindTaskState,
		core.MemoryKindDocFact,
		core.MemoryKindSummary,
		core.MemoryKindHypothesis:
		return true
	default:
		return false
	}
}

func isSupportedArtifactClass(class core.ArtifactClass) bool {
	switch class {
	case core.ArtifactClassContext,
		core.ArtifactClassKnowledge,
		core.ArtifactClassTimeline,
		core.ArtifactClassPlan:
		return true
	default:
		return false
	}
}

func isSupportedScope(scope core.MemoryScope) bool {
	switch scope {
	case core.MemoryScopeAgentPrivate,
		core.MemoryScopeWorkspaceShared,
		core.MemoryScopeGroupShared,
		core.MemoryScopeSessionScratch:
		return true
	default:
		return false
	}
}

func isValidConfidence(confidence float64) bool {
	return confidence > 0 && confidence <= 1
}

func validateJSONObject(field string, raw json.RawMessage) error {
	if len(raw) == 0 {
		return nil
	}
	if !json.Valid(raw) {
		return fmt.Errorf("%w: %s must be valid JSON", core.ErrInvalidArgument, field)
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil {
		return fmt.Errorf("%w: %s must be a JSON object", core.ErrInvalidArgument, field)
	}
	if object == nil {
		return fmt.Errorf("%w: %s must be a JSON object", core.ErrInvalidArgument, field)
	}
	return nil
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



<!-- Source: internal/graph/doc.go | bytes=805 | lines=16 | sha16=8832ab33000f7d49 -->

```go
// ============================================================
// FILE     : internal/graph/doc.go
// PURPOSE  : Provides package documentation for graph apply, profile, correction, and dreaming logic.
// LAYER    : application
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : package graph
// DEPENDS  : plans/05_runtime-contracts_ingest-recall-apply.md
// USED_BY  : worker pipeline, core service implementations
// ------------------------------------------------------------
// AGENT_NOTE: Validate structured reasoning output before writing memories, edges, traces, or profiles.
// ============================================================

// Package graph owns memory graph apply, profile updates, corrections, and dreaming rules.
package graph

```



<!-- Source: internal/graph/dreaming.go | bytes=7752 | lines=244 | sha16=bed8d68de0f20863 -->

```go
// ============================================================
// FILE     : internal/graph/dreaming.go
// PURPOSE  : Runs background memory consolidation and promotion without creating duplicate memories.
// LAYER    : application
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : DreamingStore, DreamingService, DreamingDependencies, NewDreamingService
// DEPENDS  : context, fmt, strings, time, internal/core, internal/store
// USED_BY  : internal/worker, graph dreaming tests
// ------------------------------------------------------------
// AGENT_NOTE: Dreaming may promote tiers and summaries, but it must not reinterpret raw text locally.
// ============================================================

package graph

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/store"
)

const (
	sessionMemorySummaryLimit = 5
	longTermMinConfidence     = 0.85
	ultraLongTermConfidence   = 0.95
)

// DreamingStore is the storage surface used by background consolidation.
type DreamingStore interface {
	store.DreamingStore
	store.SessionSummaryStore
}

// DreamingDependencies collects storage and time dependencies for dreaming.
type DreamingDependencies struct {
	Store DreamingStore
	Clock func() time.Time
}

// DreamingService runs session and workspace consolidation jobs.
type DreamingService struct {
	store DreamingStore
	clock func() time.Time
}

// NewDreamingService builds a background dreaming service.
func NewDreamingService(deps DreamingDependencies) (*DreamingService, error) {
	if deps.Store == nil {
		return nil, fmt.Errorf("%w: dreaming store is required", core.ErrInvalidArgument)
	}
	clock := deps.Clock
	if clock == nil {
		clock = time.Now
	}
	return &DreamingService{
		store: deps.Store,
		clock: clock,
	}, nil
}

// DreamSession consolidates one session into a summary and marks its derived memories mid-term.
func (s *DreamingService) DreamSession(ctx context.Context, req *core.DreamSessionRequest) (*core.DreamingResult, error) {
	if req == nil {
		return nil, fmt.Errorf("%w: dream_session request is required", core.ErrInvalidArgument)
	}
	if err := validateDreamSessionRequest(req); err != nil {
		return nil, err
	}
	now := s.now(req.Now)
	req.Now = now

	input, err := s.store.LoadDreamingSessionInput(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("load dream_session input: %w", err)
	}
	if input == nil {
		input = &core.DreamingSessionInput{}
	}

	result := &core.DreamingResult{}
	summary := buildSessionSummary(req, input, now)
	if strings.TrimSpace(summary.SummaryText) != "" {
		if err := s.store.UpsertSessionSummary(ctx, summary); err != nil {
			return nil, fmt.Errorf("write dream_session summary: %w", err)
		}
		result.SessionSummaryWritten = true
	}

	memoryIDs := memoryIDs(input.Memories)
	if len(memoryIDs) > 0 {
		promotion, err := s.store.PromoteMemories(ctx, &core.DreamingPromotionRequest{
			JobID:         req.JobID,
			TenantID:      req.TenantID,
			WorkspaceID:   req.WorkspaceID,
			SessionID:     req.SessionID,
			MemoryIDs:     memoryIDs,
			Tier:          core.DreamingTierMidTerm,
			Now:           now,
			MinConfidence: 0,
		})
		if err != nil {
			return nil, fmt.Errorf("promote session memories: %w", err)
		}
		result.MidTermPromoted = promotion.PromotedCount
	}

	workspaceResult, err := s.DreamWorkspace(ctx, &core.DreamWorkspaceRequest{
		JobID:       req.JobID,
		TenantID:    req.TenantID,
		WorkspaceID: req.WorkspaceID,
		Now:         now,
	})
	if err != nil {
		return nil, err
	}
	result.LongTermPromoted = workspaceResult.LongTermPromoted
	result.UltraLongTermPromoted = workspaceResult.UltraLongTermPromoted
	return result, nil
}

// DreamWorkspace promotes stable existing memories deeper into long-term tiers.
func (s *DreamingService) DreamWorkspace(ctx context.Context, req *core.DreamWorkspaceRequest) (*core.DreamingResult, error) {
	if req == nil {
		return nil, fmt.Errorf("%w: dream_workspace request is required", core.ErrInvalidArgument)
	}
	if err := validateDreamWorkspaceRequest(req); err != nil {
		return nil, err
	}
	now := s.now(req.Now)

	longTerm, err := s.store.PromoteMemories(ctx, &core.DreamingPromotionRequest{
		JobID:             req.JobID,
		TenantID:          req.TenantID,
		WorkspaceID:       req.WorkspaceID,
		Tier:              core.DreamingTierLongTerm,
		MinConfidence:     longTermMinConfidence,
		RequireStableKind: true,
		Now:               now,
	})
	if err != nil {
		return nil, fmt.Errorf("promote workspace long-term memories: %w", err)
	}
	ultraLongTerm, err := s.store.PromoteMemories(ctx, &core.DreamingPromotionRequest{
		JobID:             req.JobID,
		TenantID:          req.TenantID,
		WorkspaceID:       req.WorkspaceID,
		Tier:              core.DreamingTierUltraLongTerm,
		MinConfidence:     ultraLongTermConfidence,
		RequireStableKind: true,
		Now:               now,
	})
	if err != nil {
		return nil, fmt.Errorf("promote workspace ultra-long-term memories: %w", err)
	}
	return &core.DreamingResult{
		LongTermPromoted:      longTerm.PromotedCount,
		UltraLongTermPromoted: ultraLongTerm.PromotedCount,
	}, nil
}

func (s *DreamingService) now(value time.Time) time.Time {
	if !value.IsZero() {
		return value.UTC()
	}
	return s.clock().UTC()
}

func validateDreamSessionRequest(req *core.DreamSessionRequest) error {
	if strings.TrimSpace(req.JobID) == "" {
		return fmt.Errorf("%w: dream_session job_id is required", core.ErrInvalidArgument)
	}
	if err := validateDreamWorkspaceFields(req.TenantID, req.WorkspaceID); err != nil {
		return err
	}
	if strings.TrimSpace(req.SessionID) == "" {
		return fmt.Errorf("%w: dream_session session_id is required", core.ErrInvalidArgument)
	}
	return nil
}

func validateDreamWorkspaceRequest(req *core.DreamWorkspaceRequest) error {
	if strings.TrimSpace(req.JobID) == "" {
		return fmt.Errorf("%w: dream_workspace job_id is required", core.ErrInvalidArgument)
	}
	return validateDreamWorkspaceFields(req.TenantID, req.WorkspaceID)
}

func validateDreamWorkspaceFields(tenantID string, workspaceID string) error {
	if strings.TrimSpace(tenantID) == "" {
		return fmt.Errorf("%w: dreaming tenant_id is required", core.ErrInvalidArgument)
	}
	if strings.TrimSpace(workspaceID) == "" {
		return fmt.Errorf("%w: dreaming workspace_id is required", core.ErrInvalidArgument)
	}
	return nil
}

func buildSessionSummary(req *core.DreamSessionRequest, input *core.DreamingSessionInput, now time.Time) *core.SessionSummary {
	sourceMemoryIDs := memoryIDs(input.Memories)
	lines := []string{
		fmt.Sprintf("Session %s consolidated %d raw events and %d derived memories.", req.SessionID, len(input.RawEventIDs), len(sourceMemoryIDs)),
	}
	for i, memory := range input.Memories {
		if i >= sessionMemorySummaryLimit {
			break
		}
		if memory == nil || strings.TrimSpace(memory.Text) == "" {
			continue
		}
		lines = append(lines, "- "+strings.TrimSpace(memory.Text))
	}
	return &core.SessionSummary{
		ID:              "sum_" + req.JobID,
		TenantID:        req.TenantID,
		WorkspaceID:     req.WorkspaceID,
		SessionID:       req.SessionID,
		SummaryText:     strings.Join(lines, "\n"),
		SourceEventIDs:  append([]string(nil), input.RawEventIDs...),
		SourceMemoryIDs: sourceMemoryIDs,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

func memoryIDs(memories []*core.Memory) []string {
	ids := make([]string, 0, len(memories))
	seen := make(map[string]struct{}, len(memories))
	for _, memory := range memories {
		if memory == nil || memory.ID == "" {
			continue
		}
		if _, ok := seen[memory.ID]; ok {
			continue
		}
		ids = append(ids, memory.ID)
		seen[memory.ID] = struct{}{}
	}
	return ids
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



<!-- Source: internal/graph/store_apply.go | bytes=9899 | lines=255 | sha16=9cf8c2ab390ac3ca -->

```go
// ============================================================
// FILE     : internal/graph/store_apply.go
// PURPOSE  : Applies validated create, extend, and update memory operations to durable graph storage.
// LAYER    : application
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : MemoryTraceCreator, StoreBackedApplyEngine, NewStoreBackedApplyEngine
// DEPENDS  : context, crypto/sha256, encoding/hex, encoding/json, fmt, strings, time, internal/core, internal/reasoning
// USED_BY  : cmd/worker, internal/graph tests
// ------------------------------------------------------------
// AGENT_NOTE: Latest-changing updates must be one storage transaction with trace, edge, and prior-memory supersession.
// ============================================================

package graph

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/reasoning"
)

// MemoryTraceCreator persists derived memory writes with mandatory trace provenance.
type MemoryTraceCreator interface {
	CreateMemoryWithTrace(ctx context.Context, memory *core.Memory, trace *core.MemoryTrace) error
	CreateMemoryWithTraceAndEdge(ctx context.Context, memory *core.Memory, trace *core.MemoryTrace, edge *core.MemoryEdge) error
	CreateMemoryWithTraceAndUpdateEdge(ctx context.Context, memory *core.Memory, trace *core.MemoryTrace, edge *core.MemoryEdge) error
}

// StoreBackedApplyEngine writes validated create_memory, extend_memory, and update_memory operations to storage.
type StoreBackedApplyEngine struct {
	memories MemoryTraceCreator
	now      func() time.Time
}

// NewStoreBackedApplyEngine creates the first write-capable graph apply engine.
func NewStoreBackedApplyEngine(memories MemoryTraceCreator) (*StoreBackedApplyEngine, error) {
	if memories == nil {
		return nil, fmt.Errorf("%w: graph memory store is required", core.ErrInvalidArgument)
	}
	return &StoreBackedApplyEngine{
		memories: memories,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}, nil
}

// Apply validates Stage 2 output and writes safe memory operations with provenance.
func (e *StoreBackedApplyEngine) Apply(ctx context.Context, req *ApplyRequest) (*ApplyResult, error) {
	if err := validateApplyRequest(req); err != nil {
		return nil, err
	}
	if err := rejectUnsupportedWriteScope(req); err != nil {
		return nil, err
	}

	createdIDs := make([]string, 0, len(req.Reasoning.Stage2.Operations))
	for i, operation := range req.Reasoning.Stage2.Operations {
		memory, trace, err := e.buildMemoryTrace(req, i, operation)
		if err != nil {
			return nil, err
		}
		switch operation.Kind {
		case reasoning.OperationKindCreateMemory:
			if err := e.memories.CreateMemoryWithTrace(ctx, memory, trace); err != nil {
				return nil, fmt.Errorf("create memory with trace: %w", err)
			}
		case reasoning.OperationKindExtendMemory:
			edge, err := buildMemoryEdge(req, memory, operation)
			if err != nil {
				return nil, err
			}
			if err := e.memories.CreateMemoryWithTraceAndEdge(ctx, memory, trace, edge); err != nil {
				return nil, fmt.Errorf("create extension memory with trace and edge: %w", err)
			}
		case reasoning.OperationKindUpdateMemory:
			edge, err := buildMemoryEdge(req, memory, operation)
			if err != nil {
				return nil, err
			}
			if err := e.memories.CreateMemoryWithTraceAndUpdateEdge(ctx, memory, trace, edge); err != nil {
				return nil, fmt.Errorf("create update memory with trace and supersession edge: %w", err)
			}
		default:
			return nil, fmt.Errorf("%w: operations[%d].kind %q is validation-only in store-backed apply", core.ErrNotImplemented, i, operation.Kind)
		}
		createdIDs = append(createdIDs, memory.ID)
	}
	return &ApplyResult{
		AppliedOperationCount: len(createdIDs),
		MemoryIDs:             createdIDs,
		TraceWritten:          len(createdIDs) > 0,
	}, nil
}

func rejectUnsupportedWriteScope(req *ApplyRequest) error {
	if hasNonEmptyObject(req.Reasoning.Stage2.ProfileDelta) {
		return fmt.Errorf("%w: profile_delta writes are not implemented", core.ErrNotImplemented)
	}
	if strings.TrimSpace(req.Reasoning.Stage2.SessionSummary) != "" {
		return fmt.Errorf("%w: session summary writes are not implemented", core.ErrNotImplemented)
	}
	if hasNonEmptyObject(req.Reasoning.Stage2.PlanDelta) {
		return fmt.Errorf("%w: plan_delta writes are not implemented", core.ErrNotImplemented)
	}
	for i, operation := range req.Reasoning.Stage2.Operations {
		if operation.Memory.Scope == core.MemoryScopeGroupShared {
			return fmt.Errorf("%w: operations[%d].memory.group_shared requires membership validation before writes", core.ErrNotImplemented, i)
		}
		switch operation.Kind {
		case reasoning.OperationKindCreateMemory:
			if operation.Edge != nil {
				return fmt.Errorf("%w: operations[%d].edge writes are not implemented for create_memory", core.ErrNotImplemented, i)
			}
		case reasoning.OperationKindExtendMemory:
			// extend_memory is the safe lineage write: it adds an extends edge while leaving the target memory alive.
		case reasoning.OperationKindUpdateMemory:
			// update_memory writes a new active memory, links it to the prior latest memory, and supersedes that target atomically.
		case reasoning.OperationKindArchiveMemory:
			return fmt.Errorf("%w: operations[%d].kind %q requires archive status handling", core.ErrNotImplemented, i, operation.Kind)
		default:
			return fmt.Errorf("%w: operations[%d].kind %q is validation-only in store-backed apply", core.ErrNotImplemented, i, operation.Kind)
		}
	}
	return nil
}

func (e *StoreBackedApplyEngine) buildMemoryTrace(req *ApplyRequest, index int, operation reasoning.GraphOperation) (*core.Memory, *core.MemoryTrace, error) {
	createdAt := e.now().UTC()
	mutation := operation.Memory
	memory := &core.Memory{
		ID:            deterministicID("mem", req.TenantID, req.WorkspaceID, req.JobID, operation.OperationID),
		TenantID:      req.TenantID,
		WorkspaceID:   req.WorkspaceID,
		Scope:         mutation.Scope,
		GroupID:       mutation.GroupID,
		OwnerEntityID: mutation.OwnerEntityID,
		Kind:          mutation.Kind,
		ArtifactClass: mutation.ArtifactClass,
		Text:          mutation.Text,
		Fingerprint:   memoryFingerprint(req, operation),
		Confidence:    mutation.Confidence,
		Status:        core.MemoryStatusActive,
		ValidFrom:     createdAt,
		LatestFlag:    true,
		MetadataJSON:  mutation.MetadataJSON,
		CreatedAt:     createdAt,
		UpdatedAt:     createdAt,
	}

	candidateSnapshot, err := json.Marshal(req.Reasoning.Stage1)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal stage 1 candidate snapshot: %w", err)
	}
	appliedOperations, err := json.Marshal([]reasoning.GraphOperation{operation})
	if err != nil {
		return nil, nil, fmt.Errorf("marshal applied operation %d: %w", index, err)
	}
	trace := &core.MemoryTrace{
		MemoryID:              memory.ID,
		RawEventIDs:           append([]string(nil), operation.RawEventIDs...),
		ReasoningJobID:        req.JobID,
		ReasoningStage:        string(reasoning.StageNameResolve),
		CandidateSnapshotJSON: json.RawMessage(candidateSnapshot),
		AppliedOperationsJSON: json.RawMessage(appliedOperations),
		RelatedDocumentIDs:    []string{},
		CreatedAt:             createdAt,
	}
	return memory, trace, nil
}

func buildMemoryEdge(req *ApplyRequest, memory *core.Memory, operation reasoning.GraphOperation) (*core.MemoryEdge, error) {
	if operation.Edge == nil {
		return nil, fmt.Errorf("%w: memory edge is required", core.ErrInvalidArgument)
	}
	if operation.Edge.ToMemoryID == "" {
		return nil, fmt.Errorf("%w: memory edge target is required", core.ErrInvalidArgument)
	}
	if operation.Edge.ToMemoryID == memory.ID {
		return nil, fmt.Errorf("%w: memory edge cannot target itself", core.ErrInvalidArgument)
	}
	return &core.MemoryEdge{
		FromMemoryID:   memory.ID,
		ToMemoryID:     operation.Edge.ToMemoryID,
		EdgeKind:       operation.Edge.EdgeKind,
		Confidence:     operation.Edge.Confidence,
		CreatedByJobID: req.JobID,
		CreatedAt:      memory.CreatedAt,
	}, nil
}

func deterministicID(prefix string, parts ...string) string {
	sum := hashParts(parts...)
	return prefix + "_" + hex.EncodeToString(sum[:12])
}

func memoryFingerprint(req *ApplyRequest, operation reasoning.GraphOperation) string {
	payload := struct {
		TenantID      string             `json:"tenant_id"`
		WorkspaceID   string             `json:"workspace_id"`
		Scope         core.MemoryScope   `json:"scope"`
		GroupID       *string            `json:"group_id,omitempty"`
		OwnerEntityID string             `json:"owner_entity_id"`
		Kind          core.MemoryKind    `json:"kind"`
		ArtifactClass core.ArtifactClass `json:"artifact_class"`
		Text          string             `json:"text"`
		RawEventIDs   []string           `json:"raw_event_ids"`
	}{
		TenantID:      req.TenantID,
		WorkspaceID:   req.WorkspaceID,
		Scope:         operation.Memory.Scope,
		GroupID:       operation.Memory.GroupID,
		OwnerEntityID: operation.Memory.OwnerEntityID,
		Kind:          operation.Memory.Kind,
		ArtifactClass: operation.Memory.ArtifactClass,
		Text:          operation.Memory.Text,
		RawEventIDs:   operation.RawEventIDs,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return deterministicID("fp", req.JobID, operation.OperationID)
	}
	sum := sha256.Sum256(data)
	return "fp_" + hex.EncodeToString(sum[:16])
}

func hashParts(parts ...string) [32]byte {
	hash := sha256.New()
	for _, part := range parts {
		hash.Write([]byte(part))
		hash.Write([]byte{0})
	}
	var sum [32]byte
	copy(sum[:], hash.Sum(nil))
	return sum
}

func hasNonEmptyObject(raw json.RawMessage) bool {
	if len(raw) == 0 {
		return false
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil {
		return true
	}
	return len(object) > 0
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



<!-- Source: internal/reasoning/codex_bridge.go | bytes=10052 | lines=264 | sha16=6b527b1e7886d32e -->

```go
// ============================================================
// FILE     : internal/reasoning/codex_bridge.go
// PURPOSE  : Defines the disabled-by-default Codex JSON bridge boundary.
// LAYER    : application
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : CodexBridgeConfig, CodexJSONClient, CodexRequest, CodexResponse, CodexStage1Extractor, CodexStage2Resolver
// DEPENDS  : bytes, context, encoding/json, fmt, internal/core
// USED_BY  : internal/reasoning tests, future worker wiring
// ------------------------------------------------------------
// AGENT_NOTE: Keep this bridge schema-first; do not add local extraction fallback here.
// ============================================================

package reasoning

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

// Stage1ExtractOutputSchemaV0 marks the required structured Stage 1 output contract.
const Stage1ExtractOutputSchemaV0 = "stage1.extract.output.v0"

// CodexBridgeConfig controls whether the Codex bridge runners are allowed to call a client.
type CodexBridgeConfig struct {
	Enabled bool
}

// CodexRequest is the schema-marked request sent to a Codex JSON client.
type CodexRequest struct {
	Stage                StageName       `json:"stage"`
	RequiredOutputName   StageName       `json:"required_output_name"`
	RequiredOutputSchema string          `json:"required_output_schema"`
	InputJSON            json.RawMessage `json:"input_json"`
}

// CodexResponse is the raw structured JSON returned by a Codex JSON client.
type CodexResponse struct {
	OutputJSON json.RawMessage `json:"output_json"`
}

// CodexJSONClient is the narrow client boundary for schema-first Codex calls.
type CodexJSONClient interface {
	CompleteJSON(ctx context.Context, req CodexRequest) (CodexResponse, error)
}

// CodexStage1Extractor runs Stage 1 through the Codex JSON client when explicitly enabled.
type CodexStage1Extractor struct {
	cfg    CodexBridgeConfig
	client CodexJSONClient
}

// NewCodexStage1Extractor creates a disabled-by-default Stage 1 Codex runner.
func NewCodexStage1Extractor(cfg CodexBridgeConfig, client CodexJSONClient) (*CodexStage1Extractor, error) {
	if cfg.Enabled && client == nil {
		return nil, fmt.Errorf("%w: enabled codex stage1 extractor requires a client", core.ErrInvalidArgument)
	}
	return &CodexStage1Extractor{cfg: cfg, client: client}, nil
}

// Extract sends Stage 1 input to Codex only when the bridge is explicitly enabled.
func (e *CodexStage1Extractor) Extract(ctx context.Context, input Stage1Input) (Stage1Output, error) {
	if err := validateStage1Input(input); err != nil {
		return Stage1Output{}, err
	}
	if e == nil || !e.cfg.Enabled {
		return Stage1Output{}, fmt.Errorf("%w: codex stage1 extractor is disabled", core.ErrNotImplemented)
	}
	inputJSON, err := json.Marshal(input)
	if err != nil {
		return Stage1Output{}, fmt.Errorf("marshal stage1 input: %w", err)
	}
	resp, err := e.client.CompleteJSON(ctx, CodexRequest{
		Stage:                StageNameExtract,
		RequiredOutputName:   StageNameExtract,
		RequiredOutputSchema: Stage1ExtractOutputSchemaV0,
		InputJSON:            inputJSON,
	})
	if err != nil {
		return Stage1Output{}, fmt.Errorf("codex stage1 complete: %w", err)
	}
	output, err := decodeStrictJSON[Stage1Output]("stage1 output", resp.OutputJSON)
	if err != nil {
		return Stage1Output{}, err
	}
	if err := validateStage1Output(output); err != nil {
		return Stage1Output{}, err
	}
	return output, nil
}

// CodexStage2Resolver runs Stage 2 through the Codex JSON client when explicitly enabled.
type CodexStage2Resolver struct {
	cfg    CodexBridgeConfig
	client CodexJSONClient
}

// NewCodexStage2Resolver creates a disabled-by-default Stage 2 Codex runner.
func NewCodexStage2Resolver(cfg CodexBridgeConfig, client CodexJSONClient) (*CodexStage2Resolver, error) {
	if cfg.Enabled && client == nil {
		return nil, fmt.Errorf("%w: enabled codex stage2 resolver requires a client", core.ErrInvalidArgument)
	}
	return &CodexStage2Resolver{cfg: cfg, client: client}, nil
}

// Resolve sends prepared Stage 2 input to Codex only when the bridge is explicitly enabled.
func (r *CodexStage2Resolver) Resolve(ctx context.Context, input Stage2Input) (Stage2Output, error) {
	if err := validatePreparedStage2Input(input); err != nil {
		return Stage2Output{}, err
	}
	if input.RequiredOutputSchema != Stage2ResolveOutputSchemaV0 {
		return Stage2Output{}, fmt.Errorf("%w: stage2 required output schema must be %q", core.ErrInvalidArgument, Stage2ResolveOutputSchemaV0)
	}
	if r == nil || !r.cfg.Enabled {
		return Stage2Output{}, fmt.Errorf("%w: codex stage2 resolver is disabled", core.ErrNotImplemented)
	}
	inputJSON, err := json.Marshal(input)
	if err != nil {
		return Stage2Output{}, fmt.Errorf("marshal stage2 input: %w", err)
	}
	resp, err := r.client.CompleteJSON(ctx, CodexRequest{
		Stage:                StageNameResolve,
		RequiredOutputName:   input.RequiredOutputName,
		RequiredOutputSchema: input.RequiredOutputSchema,
		InputJSON:            inputJSON,
	})
	if err != nil {
		return Stage2Output{}, fmt.Errorf("codex stage2 complete: %w", err)
	}
	output, err := decodeStrictJSON[Stage2Output]("stage2 output", resp.OutputJSON)
	if err != nil {
		return Stage2Output{}, err
	}
	if err := validateStage2Output(output); err != nil {
		return Stage2Output{}, err
	}
	if err := validateStage2OutputJSONFields(output); err != nil {
		return Stage2Output{}, err
	}
	return output, nil
}

func validateStage1Input(input Stage1Input) error {
	if input.JobID == "" {
		return fmt.Errorf("%w: stage1 job_id is required", core.ErrInvalidArgument)
	}
	if input.TenantID == "" {
		return fmt.Errorf("%w: stage1 tenant_id is required", core.ErrInvalidArgument)
	}
	if input.WorkspaceID == "" {
		return fmt.Errorf("%w: stage1 workspace_id is required", core.ErrInvalidArgument)
	}
	if len(input.RawEvents) == 0 {
		return fmt.Errorf("%w: stage1 raw event bundle is required", core.ErrInvalidArgument)
	}
	return nil
}

func validateStage1Output(output Stage1Output) error {
	for i, entity := range output.CandidateEntities {
		if entity.EntityKind == "" {
			return fmt.Errorf("%w: stage1 candidate_entities[%d].entity_kind is required", core.ErrInvalidArgument, i)
		}
		if entity.DisplayName == "" {
			return fmt.Errorf("%w: stage1 candidate_entities[%d].display_name is required", core.ErrInvalidArgument, i)
		}
		if !validConfidence(entity.Confidence) {
			return fmt.Errorf("%w: stage1 candidate_entities[%d].confidence must be greater than 0 and less than or equal to 1", core.ErrInvalidArgument, i)
		}
		if err := validateJSONObject(fmt.Sprintf("stage1 candidate_entities[%d].metadata_json", i), entity.MetadataJSON); err != nil {
			return err
		}
		if entity.SourceEventID == "" {
			return fmt.Errorf("%w: stage1 candidate_entities[%d].source_event_id is required", core.ErrInvalidArgument, i)
		}
	}
	for i, memory := range output.CandidateMemories {
		if memory.Kind == "" {
			return fmt.Errorf("%w: stage1 candidate_memories[%d].kind is required", core.ErrInvalidArgument, i)
		}
		if memory.ArtifactClass == "" {
			return fmt.Errorf("%w: stage1 candidate_memories[%d].artifact_class is required", core.ErrInvalidArgument, i)
		}
		if memory.Scope == "" {
			return fmt.Errorf("%w: stage1 candidate_memories[%d].scope is required", core.ErrInvalidArgument, i)
		}
		if memory.Text == "" {
			return fmt.Errorf("%w: stage1 candidate_memories[%d].text is required", core.ErrInvalidArgument, i)
		}
		if !validConfidence(memory.Confidence) {
			return fmt.Errorf("%w: stage1 candidate_memories[%d].confidence must be greater than 0 and less than or equal to 1", core.ErrInvalidArgument, i)
		}
		if len(memory.RawEventIDs) == 0 {
			return fmt.Errorf("%w: stage1 candidate_memories[%d].raw_event_ids are required", core.ErrInvalidArgument, i)
		}
		for _, rawEventID := range memory.RawEventIDs {
			if rawEventID == "" {
				return fmt.Errorf("%w: stage1 candidate_memories[%d].raw_event_ids cannot contain empty ids", core.ErrInvalidArgument, i)
			}
		}
	}
	return nil
}

func validateStage2OutputJSONFields(output Stage2Output) error {
	if err := validateJSONObject("stage2 profile_delta", output.ProfileDelta); err != nil {
		return err
	}
	if err := validateJSONObject("stage2 plan_delta", output.PlanDelta); err != nil {
		return err
	}
	if err := validateJSONObject("stage2 trace.metadata_json", output.Trace.MetadataJSON); err != nil {
		return err
	}
	for i, operation := range output.Operations {
		if err := validateJSONObject(fmt.Sprintf("stage2 operations[%d].metadata", i), operation.Metadata); err != nil {
			return err
		}
		if operation.Memory != nil {
			if err := validateJSONObject(fmt.Sprintf("stage2 operations[%d].memory.metadata_json", i), operation.Memory.MetadataJSON); err != nil {
				return err
			}
		}
	}
	return nil
}

func decodeStrictJSON[T any](label string, raw json.RawMessage) (T, error) {
	var value T
	if len(bytes.TrimSpace(raw)) == 0 {
		return value, fmt.Errorf("%w: %s JSON is required", core.ErrInvalidArgument, label)
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return value, fmt.Errorf("%w: decode %s: %v", core.ErrInvalidArgument, label, err)
	}
	var trailing struct{}
	if err := decoder.Decode(&trailing); err != io.EOF {
		return value, fmt.Errorf("%w: decode %s: trailing JSON is not allowed", core.ErrInvalidArgument, label)
	}
	return value, nil
}

func validateJSONObject(label string, raw json.RawMessage) error {
	if len(bytes.TrimSpace(raw)) == 0 {
		return fmt.Errorf("%w: %s must be a JSON object", core.ErrInvalidArgument, label)
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil {
		return fmt.Errorf("%w: %s must be a JSON object", core.ErrInvalidArgument, label)
	}
	return nil
}

func validConfidence(confidence float64) bool {
	return confidence > 0 && confidence <= 1
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



<!-- Source: internal/reasoning/contracts.go | bytes=6989 | lines=163 | sha16=d682199eaaaae999 -->

```go
// ============================================================
// FILE     : internal/reasoning/contracts.go
// PURPOSE  : Defines schema-first contracts for Codex extract and resolve stages.
// LAYER    : application
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : StageName, process turn contracts, stage contracts, graph operation DTOs, Trace
// DEPENDS  : encoding/json, internal/core
// USED_BY  : internal/worker, internal/graph, tests
// ------------------------------------------------------------
// AGENT_NOTE: Stage outputs must be structured data only; free-form reasoning must not cross apply.
// ============================================================

package reasoning

import (
	"encoding/json"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

// StageName identifies a reasoning stage in memory trace records.
type StageName string

const (
	// StageNameExtract is the first Codex pass that proposes candidates only.
	StageNameExtract StageName = "extract"
	// StageNameResolve is the second Codex pass that produces apply operations.
	StageNameResolve StageName = "resolve"
)

// ProcessTurnEnvelope is the worker-built input bundle for a process_turn_event job.
type ProcessTurnEnvelope struct {
	JobID       string           `json:"job_id"`
	TenantID    string           `json:"tenant_id"`
	WorkspaceID string           `json:"workspace_id"`
	RawEventIDs []string         `json:"raw_event_ids"`
	RawEvents   []*core.RawEvent `json:"raw_events"`
	Stage1      Stage1Input      `json:"stage_1"`
	Stage2      Stage2Input      `json:"stage_2"`
}

// ProcessTurnResult contains the structured outputs from both reasoning stages.
type ProcessTurnResult struct {
	Stage1 Stage1Output `json:"stage_1"`
	Stage2 Stage2Output `json:"stage_2"`
}

// Stage1Input is the schema for candidate extraction.
type Stage1Input struct {
	JobID       string           `json:"job_id"`
	TenantID    string           `json:"tenant_id"`
	WorkspaceID string           `json:"workspace_id"`
	RawEvents   []*core.RawEvent `json:"raw_events"`
}

// Stage1Output is candidate-only output; it is not an apply contract.
type Stage1Output struct {
	CandidateEntities []CandidateEntity `json:"candidate_entities"`
	CandidateMemories []CandidateMemory `json:"candidate_memories"`
	SummaryHint       string            `json:"summary_hint"`
	TaskHint          string            `json:"task_hint"`
}

// CandidateEntity describes an entity mention proposed by Stage 1.
type CandidateEntity struct {
	EntityID      string          `json:"entity_id,omitempty"`
	EntityKind    string          `json:"entity_kind"`
	DisplayName   string          `json:"display_name"`
	Confidence    float64         `json:"confidence"`
	MetadataJSON  json.RawMessage `json:"metadata_json"`
	SourceEventID string          `json:"source_event_id"`
}

// CandidateMemory describes a possible memory proposed by Stage 1.
type CandidateMemory struct {
	Kind          core.MemoryKind    `json:"kind"`
	ArtifactClass core.ArtifactClass `json:"artifact_class"`
	Scope         core.MemoryScope   `json:"scope"`
	Text          string             `json:"text"`
	Confidence    float64            `json:"confidence"`
	RawEventIDs   []string           `json:"raw_event_ids"`
}

// Stage2Input is the schema for conflict resolution and operation planning.
type Stage2Input struct {
	JobID                string                     `json:"job_id"`
	TenantID             string                     `json:"tenant_id"`
	WorkspaceID          string                     `json:"workspace_id"`
	RawEvents            []*core.RawEvent           `json:"raw_events"`
	Stage1               Stage1Output               `json:"stage_1"`
	ExistingProfile      *core.Profile              `json:"existing_profile,omitempty"`
	RelevantMemories     []core.MemoryResult        `json:"relevant_memories"`
	RelevantDocuments    []core.DocumentChunkResult `json:"relevant_documents"`
	ActivePlans          []*core.Plan               `json:"active_plans"`
	PinnedNotes          []*core.Note               `json:"pinned_notes"`
	RequiredOutputName   StageName                  `json:"required_output_name"`
	RequiredOutputSchema string                     `json:"required_output_schema"`
}

// Stage2Output is the only reasoning output that may cross into the apply engine.
type Stage2Output struct {
	Operations     []GraphOperation `json:"operations"`
	ProfileDelta   json.RawMessage  `json:"profile_delta"`
	SessionSummary string           `json:"session_summary"`
	PlanDelta      json.RawMessage  `json:"plan_delta"`
	Trace          Trace            `json:"trace"`
}

// GraphOperation is a structured memory graph operation produced by Stage 2.
type GraphOperation struct {
	OperationID string          `json:"operation_id"`
	Kind        OperationKind   `json:"kind"`
	Memory      *MemoryMutation `json:"memory,omitempty"`
	Edge        *EdgeMutation   `json:"edge,omitempty"`
	RawEventIDs []string        `json:"raw_event_ids"`
	Metadata    json.RawMessage `json:"metadata"`
}

// OperationKind identifies the apply action requested by Stage 2.
type OperationKind string

const (
	// OperationKindCreateMemory requests a new derived memory.
	OperationKindCreateMemory OperationKind = "create_memory"
	// OperationKindUpdateMemory requests an updates edge and latest resolution.
	OperationKindUpdateMemory OperationKind = "update_memory"
	// OperationKindExtendMemory requests an extends edge while keeping prior memory alive.
	OperationKindExtendMemory OperationKind = "extend_memory"
	// OperationKindArchiveMemory requests recall suppression for an existing memory.
	OperationKindArchiveMemory OperationKind = "archive_memory"
)

// MemoryMutation is the structured memory payload for create/update operations.
type MemoryMutation struct {
	MemoryID      string             `json:"memory_id,omitempty"`
	TargetID      string             `json:"target_id,omitempty"`
	Kind          core.MemoryKind    `json:"kind"`
	ArtifactClass core.ArtifactClass `json:"artifact_class"`
	Scope         core.MemoryScope   `json:"scope"`
	GroupID       *string            `json:"group_id,omitempty"`
	OwnerEntityID string             `json:"owner_entity_id"`
	Text          string             `json:"text"`
	Confidence    float64            `json:"confidence"`
	MetadataJSON  json.RawMessage    `json:"metadata_json"`
}

// EdgeMutation is the structured edge payload for graph operations.
type EdgeMutation struct {
	FromMemoryID string        `json:"from_memory_id"`
	ToMemoryID   string        `json:"to_memory_id"`
	EdgeKind     core.EdgeKind `json:"edge_kind"`
	Confidence   float64       `json:"confidence"`
}

// Trace stores structured debugging evidence for the reasoning run.
type Trace struct {
	SchemaVersion string          `json:"schema_version"`
	Stage         StageName       `json:"stage"`
	Codes         []string        `json:"codes"`
	MetadataJSON  json.RawMessage `json:"metadata_json"`
}

```



<!-- Source: internal/reasoning/doc.go | bytes=814 | lines=16 | sha16=c2f6ca806d819a48 -->

```go
// ============================================================
// FILE     : internal/reasoning/doc.go
// PURPOSE  : Provides package documentation for Codex-first structured reasoning.
// LAYER    : application
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : package reasoning
// DEPENDS  : plans/03_target-architecture_codex-first.md, plans/05_runtime-contracts_ingest-recall-apply.md
// USED_BY  : worker pipeline, graph apply engine
// ------------------------------------------------------------
// AGENT_NOTE: Stage outputs must remain schema-first JSON and never bypass apply validation.
// ============================================================

// Package reasoning owns Codex stage 1 extraction and stage 2 resolution contracts.
package reasoning

```



<!-- Source: internal/reasoning/mock_codex_client.go | bytes=3898 | lines=97 | sha16=c2b14050b8b6b166 -->

```go
// ============================================================
// FILE     : internal/reasoning/mock_codex_client.go
// PURPOSE  : Provides a deterministic mocked Codex JSON client for the worker reasoning bridge.
// LAYER    : application
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : MockCodexJSONClient, NewMockCodexJSONClient
// DEPENDS  : context, encoding/json, fmt, internal/core
// USED_BY  : cmd/worker, internal/reasoning tests
// ------------------------------------------------------------
// AGENT_NOTE: This mock must exercise the Codex bridge interface without becoming a local extractor.
// ============================================================

package reasoning

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

// MockCodexJSONClient is a deterministic local client for exercising the Codex bridge boundary.
type MockCodexJSONClient struct{}

// NewMockCodexJSONClient creates a mocked Codex JSON client for non-networked worker wiring.
func NewMockCodexJSONClient() *MockCodexJSONClient {
	return &MockCodexJSONClient{}
}

// CompleteJSON returns schema-marked structured JSON without calling a real Codex API.
func (c *MockCodexJSONClient) CompleteJSON(_ context.Context, req CodexRequest) (CodexResponse, error) {
	if c == nil {
		return CodexResponse{}, fmt.Errorf("%w: mock codex client is required", core.ErrInvalidArgument)
	}
	switch req.Stage {
	case StageNameExtract:
		return c.completeStage1(req)
	case StageNameResolve:
		return c.completeStage2(req)
	default:
		return CodexResponse{}, fmt.Errorf("%w: unsupported mock codex stage %q", core.ErrInvalidArgument, req.Stage)
	}
}

func (c *MockCodexJSONClient) completeStage1(req CodexRequest) (CodexResponse, error) {
	if req.RequiredOutputName != StageNameExtract {
		return CodexResponse{}, fmt.Errorf("%w: mock stage1 required output name must be extract", core.ErrInvalidArgument)
	}
	if req.RequiredOutputSchema != Stage1ExtractOutputSchemaV0 {
		return CodexResponse{}, fmt.Errorf("%w: mock stage1 required output schema must be %q", core.ErrInvalidArgument, Stage1ExtractOutputSchemaV0)
	}
	input, err := decodeStrictJSON[Stage1Input]("mock stage1 input", req.InputJSON)
	if err != nil {
		return CodexResponse{}, err
	}
	if err := validateStage1Input(input); err != nil {
		return CodexResponse{}, err
	}
	output := Stage1Output{
		CandidateEntities: []CandidateEntity{},
		CandidateMemories: []CandidateMemory{},
		SummaryHint:       "mock_codex_stage1_no_candidates",
		TaskHint:          "mock_codex_stage2_no_operations",
	}
	return marshalMockCodexOutput("mock stage1 output", output)
}

func (c *MockCodexJSONClient) completeStage2(req CodexRequest) (CodexResponse, error) {
	if req.RequiredOutputName != StageNameResolve {
		return CodexResponse{}, fmt.Errorf("%w: mock stage2 required output name must be resolve", core.ErrInvalidArgument)
	}
	if req.RequiredOutputSchema != Stage2ResolveOutputSchemaV0 {
		return CodexResponse{}, fmt.Errorf("%w: mock stage2 required output schema must be %q", core.ErrInvalidArgument, Stage2ResolveOutputSchemaV0)
	}
	input, err := decodeStrictJSON[Stage2Input]("mock stage2 input", req.InputJSON)
	if err != nil {
		return CodexResponse{}, err
	}
	if err := validatePreparedStage2Input(input); err != nil {
		return CodexResponse{}, err
	}
	output := emptyStage2Output()
	output.Trace.Codes = []string{"mock_codex_bridge_no_operations"}
	output.Trace.MetadataJSON = json.RawMessage(`{"client":"mock_codex_json_client"}`)
	return marshalMockCodexOutput("mock stage2 output", output)
}

func marshalMockCodexOutput(label string, output any) (CodexResponse, error) {
	outputJSON, err := json.Marshal(output)
	if err != nil {
		return CodexResponse{}, fmt.Errorf("marshal %s: %w", label, err)
	}
	return CodexResponse{OutputJSON: outputJSON}, nil
}

```



<!-- Source: internal/reasoning/orchestrator.go | bytes=8558 | lines=236 | sha16=7d5afb1c8e60ef31 -->

```go
// ============================================================
// FILE     : internal/reasoning/orchestrator.go
// PURPOSE  : Provides schema-first reasoning orchestration interfaces and safe stubs.
// LAYER    : application
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : Orchestrator, Stage1Extractor, Stage2Resolver, StubOrchestrator, stub runners, PipelineOrchestrator
// DEPENDS  : context, encoding/json, fmt, internal/core
// USED_BY  : internal/worker, cmd/worker, tests
// ------------------------------------------------------------
// AGENT_NOTE: Runners may call Codex later, but this package must not add local extraction.
// ============================================================

package reasoning

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

var emptyJSONObject = json.RawMessage(`{}`)

// Orchestrator runs schema-first Stage 1 extract and Stage 2 resolve.
type Orchestrator interface {
	ProcessTurn(ctx context.Context, envelope *ProcessTurnEnvelope) (*ProcessTurnResult, error)
}

// Stage1Extractor runs the schema-first extract pass.
type Stage1Extractor interface {
	Extract(ctx context.Context, input Stage1Input) (Stage1Output, error)
}

// Stage2Resolver runs the schema-first resolve pass.
type Stage2Resolver interface {
	Resolve(ctx context.Context, input Stage2Input) (Stage2Output, error)
}

// StubStage1Extractor is a contract-preserving placeholder until the Codex extract bridge lands.
type StubStage1Extractor struct{}

// NewStubStage1Extractor creates a Stage 1 extractor that returns empty candidates.
func NewStubStage1Extractor() *StubStage1Extractor {
	return &StubStage1Extractor{}
}

// Extract validates Stage 1 input and returns empty structured candidate output.
func (e *StubStage1Extractor) Extract(_ context.Context, input Stage1Input) (Stage1Output, error) {
	if input.JobID == "" {
		return Stage1Output{}, fmt.Errorf("%w: stage1 job_id is required", core.ErrInvalidArgument)
	}
	if input.TenantID == "" {
		return Stage1Output{}, fmt.Errorf("%w: stage1 tenant_id is required", core.ErrInvalidArgument)
	}
	if input.WorkspaceID == "" {
		return Stage1Output{}, fmt.Errorf("%w: stage1 workspace_id is required", core.ErrInvalidArgument)
	}
	if len(input.RawEvents) == 0 {
		return Stage1Output{}, fmt.Errorf("%w: stage1 raw event bundle is required", core.ErrInvalidArgument)
	}
	return Stage1Output{
		CandidateEntities: []CandidateEntity{},
		CandidateMemories: []CandidateMemory{},
	}, nil
}

// StubStage2Resolver is a contract-preserving placeholder until the Codex resolve bridge lands.
type StubStage2Resolver struct{}

// NewStubStage2Resolver creates a Stage 2 resolver that returns no graph operations.
func NewStubStage2Resolver() *StubStage2Resolver {
	return &StubStage2Resolver{}
}

// Resolve validates the prepared Stage 2 input and returns empty structured output.
func (r *StubStage2Resolver) Resolve(_ context.Context, input Stage2Input) (Stage2Output, error) {
	if err := validatePreparedStage2Input(input); err != nil {
		return Stage2Output{}, err
	}
	return emptyStage2Output(), nil
}

// PipelineOrchestrator runs Stage 1, prepares Stage 2 from the Stage 1 output, then runs Stage 2.
type PipelineOrchestrator struct {
	stage1              Stage1Extractor
	stage2              Stage2Resolver
	stage2InputPreparer *Stage2InputPreparer
}

// NewPipelineOrchestrator creates a mockable two-stage reasoning orchestrator.
func NewPipelineOrchestrator(stage1 Stage1Extractor, stage2 Stage2Resolver, preparer *Stage2InputPreparer) (*PipelineOrchestrator, error) {
	if stage1 == nil {
		return nil, fmt.Errorf("%w: stage1 extractor is required", core.ErrInvalidArgument)
	}
	if stage2 == nil {
		return nil, fmt.Errorf("%w: stage2 resolver is required", core.ErrInvalidArgument)
	}
	if preparer == nil {
		preparer = NewStage2InputPreparer(Stage2InputSources{})
	}
	return &PipelineOrchestrator{
		stage1:              stage1,
		stage2:              stage2,
		stage2InputPreparer: preparer,
	}, nil
}

// ProcessTurn executes the schema-first two-stage reasoning chain.
func (o *PipelineOrchestrator) ProcessTurn(ctx context.Context, envelope *ProcessTurnEnvelope) (*ProcessTurnResult, error) {
	if err := validateProcessTurnEnvelope(envelope); err != nil {
		return nil, err
	}
	stage1Output, err := o.stage1.Extract(ctx, envelope.Stage1)
	if err != nil {
		return nil, fmt.Errorf("stage1 extract: %w", err)
	}
	stage2Input, err := o.stage2InputPreparer.Prepare(ctx, Stage2InputRequest{
		JobID:       envelope.JobID,
		TenantID:    envelope.TenantID,
		WorkspaceID: envelope.WorkspaceID,
		RawEvents:   envelope.RawEvents,
		Stage1:      stage1Output,
	})
	if err != nil {
		return nil, fmt.Errorf("prepare stage2 input: %w", err)
	}
	stage2Output, err := o.stage2.Resolve(ctx, stage2Input)
	if err != nil {
		return nil, fmt.Errorf("stage2 resolve: %w", err)
	}
	if err := validateStage2Output(stage2Output); err != nil {
		return nil, err
	}
	return &ProcessTurnResult{
		Stage1: stage1Output,
		Stage2: stage2Output,
	}, nil
}

// StubOrchestrator is a compatibility wrapper around the safe two-stage stub pipeline.
type StubOrchestrator struct {
	pipeline *PipelineOrchestrator
}

// NewStubOrchestrator creates a reasoning orchestrator that returns empty structured output.
func NewStubOrchestrator() *StubOrchestrator {
	pipeline, err := NewPipelineOrchestrator(NewStubStage1Extractor(), NewStubStage2Resolver(), nil)
	if err != nil {
		panic(err)
	}
	return &StubOrchestrator{pipeline: pipeline}
}

// ProcessTurn validates the envelope and returns schema-shaped empty Stage 1 and Stage 2 output through the pipeline.
func (o *StubOrchestrator) ProcessTurn(ctx context.Context, envelope *ProcessTurnEnvelope) (*ProcessTurnResult, error) {
	return o.pipeline.ProcessTurn(ctx, envelope)
}

func validateProcessTurnEnvelope(envelope *ProcessTurnEnvelope) error {
	if envelope == nil {
		return fmt.Errorf("%w: reasoning envelope is required", core.ErrInvalidArgument)
	}
	if envelope.JobID == "" {
		return fmt.Errorf("%w: reasoning job_id is required", core.ErrInvalidArgument)
	}
	if envelope.TenantID == "" {
		return fmt.Errorf("%w: reasoning tenant_id is required", core.ErrInvalidArgument)
	}
	if envelope.WorkspaceID == "" {
		return fmt.Errorf("%w: reasoning workspace_id is required", core.ErrInvalidArgument)
	}
	if len(envelope.RawEvents) == 0 {
		return fmt.Errorf("%w: reasoning raw event bundle is required", core.ErrInvalidArgument)
	}
	if envelope.Stage2.RequiredOutputName != StageNameResolve {
		return fmt.Errorf("%w: reasoning stage 2 required output name must be resolve", core.ErrInvalidArgument)
	}
	if envelope.Stage2.RequiredOutputSchema == "" {
		return fmt.Errorf("%w: reasoning stage 2 required output schema is required", core.ErrInvalidArgument)
	}
	return nil
}

func validatePreparedStage2Input(input Stage2Input) error {
	if input.JobID == "" {
		return fmt.Errorf("%w: stage2 job_id is required", core.ErrInvalidArgument)
	}
	if input.TenantID == "" {
		return fmt.Errorf("%w: stage2 tenant_id is required", core.ErrInvalidArgument)
	}
	if input.WorkspaceID == "" {
		return fmt.Errorf("%w: stage2 workspace_id is required", core.ErrInvalidArgument)
	}
	if len(input.RawEvents) == 0 {
		return fmt.Errorf("%w: stage2 raw event bundle is required", core.ErrInvalidArgument)
	}
	if input.RequiredOutputName != StageNameResolve {
		return fmt.Errorf("%w: stage2 required output name must be resolve", core.ErrInvalidArgument)
	}
	if input.RequiredOutputSchema == "" {
		return fmt.Errorf("%w: stage2 required output schema is required", core.ErrInvalidArgument)
	}
	return nil
}

func validateStage2Output(output Stage2Output) error {
	if output.Trace.SchemaVersion == "" {
		return fmt.Errorf("%w: stage2 trace schema_version is required", core.ErrInvalidArgument)
	}
	if output.Trace.Stage != StageNameResolve {
		return fmt.Errorf("%w: stage2 trace must be resolve", core.ErrInvalidArgument)
	}
	return nil
}

func emptyStage2Output() Stage2Output {
	return Stage2Output{
		Operations:     []GraphOperation{},
		ProfileDelta:   cloneRawJSON(emptyJSONObject),
		SessionSummary: "",
		PlanDelta:      cloneRawJSON(emptyJSONObject),
		Trace: Trace{
			SchemaVersion: "v0",
			Stage:         StageNameResolve,
			Codes:         []string{"stub_reasoning_no_operations"},
			MetadataJSON:  cloneRawJSON(emptyJSONObject),
		},
	}
}

func cloneRawJSON(raw json.RawMessage) json.RawMessage {
	return append(json.RawMessage(nil), raw...)
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



<!-- Source: internal/reasoning/stage2_input_preparer.go | bytes=7350 | lines=220 | sha16=caf0f05162d3134b -->

```go
// ============================================================
// FILE     : internal/reasoning/stage2_input_preparer.go
// PURPOSE  : Assembles the schema-first Stage 2 input envelope from retrieval context.
// LAYER    : application
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : Stage2ResolveOutputSchemaV0, Stage2InputRequest, Stage2InputSources, Stage2InputPreparer
// DEPENDS  : context, fmt, internal/core
// USED_BY  : internal/worker, internal/reasoning tests
// ------------------------------------------------------------
// AGENT_NOTE: This layer only prepares retrieval context; it must not extract text or call Codex.
// ============================================================

package reasoning

import (
	"context"
	"fmt"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

// Stage2ResolveOutputSchemaV0 marks the required structured Stage 2 output contract.
const Stage2ResolveOutputSchemaV0 = "stage2.resolve.output.v0"

// Stage2InputRequest carries current job data and Stage 1 candidates into context preparation.
type Stage2InputRequest struct {
	JobID       string
	TenantID    string
	WorkspaceID string
	RawEvents   []*core.RawEvent
	Stage1      Stage1Output
}

// Stage2ProfileSource loads the current rebuildable profile snapshot for Stage 2.
type Stage2ProfileSource interface {
	LoadStage2Profile(ctx context.Context, req Stage2InputRequest) (*core.Profile, error)
}

// Stage2MemorySource loads relevant derived memories for Stage 2 conflict resolution.
type Stage2MemorySource interface {
	LoadStage2Memories(ctx context.Context, req Stage2InputRequest) ([]core.MemoryResult, error)
}

// Stage2DocumentSource loads relevant document chunks for Stage 2 grounding.
type Stage2DocumentSource interface {
	LoadStage2Documents(ctx context.Context, req Stage2InputRequest) ([]core.DocumentChunkResult, error)
}

// Stage2PlanSource loads active plans that may affect Stage 2 plan deltas.
type Stage2PlanSource interface {
	LoadStage2ActivePlans(ctx context.Context, req Stage2InputRequest) ([]*core.Plan, error)
}

// Stage2NoteSource loads pinned notes that must be visible to Stage 2 reasoning.
type Stage2NoteSource interface {
	LoadStage2PinnedNotes(ctx context.Context, req Stage2InputRequest) ([]*core.Note, error)
}

// Stage2InputSources groups retrieval/context dependencies for Stage 2 preparation.
type Stage2InputSources struct {
	Profiles  Stage2ProfileSource
	Memories  Stage2MemorySource
	Documents Stage2DocumentSource
	Plans     Stage2PlanSource
	Notes     Stage2NoteSource
}

// Stage2InputPreparer assembles Stage2Input without performing extraction or reasoning.
type Stage2InputPreparer struct {
	sources Stage2InputSources
}

// NewStage2InputPreparer creates an interface-driven Stage 2 input preparer.
func NewStage2InputPreparer(sources Stage2InputSources) *Stage2InputPreparer {
	return &Stage2InputPreparer{sources: sources}
}

// Prepare assembles a structured Stage2Input from current events, Stage 1 candidates, and retrieval sources.
func (p *Stage2InputPreparer) Prepare(ctx context.Context, req Stage2InputRequest) (Stage2Input, error) {
	if err := validateStage2InputRequest(req); err != nil {
		return Stage2Input{}, err
	}

	profile, err := p.loadProfile(ctx, req)
	if err != nil {
		return Stage2Input{}, err
	}
	memories, err := p.loadMemories(ctx, req)
	if err != nil {
		return Stage2Input{}, err
	}
	documents, err := p.loadDocuments(ctx, req)
	if err != nil {
		return Stage2Input{}, err
	}
	plans, err := p.loadPlans(ctx, req)
	if err != nil {
		return Stage2Input{}, err
	}
	notes, err := p.loadNotes(ctx, req)
	if err != nil {
		return Stage2Input{}, err
	}

	return Stage2Input{
		JobID:                req.JobID,
		TenantID:             req.TenantID,
		WorkspaceID:          req.WorkspaceID,
		RawEvents:            append([]*core.RawEvent(nil), req.RawEvents...),
		Stage1:               req.Stage1,
		ExistingProfile:      profile,
		RelevantMemories:     cloneMemoryResults(memories),
		RelevantDocuments:    cloneDocumentChunkResults(documents),
		ActivePlans:          clonePlans(plans),
		PinnedNotes:          cloneNotes(notes),
		RequiredOutputName:   StageNameResolve,
		RequiredOutputSchema: Stage2ResolveOutputSchemaV0,
	}, nil
}

func validateStage2InputRequest(req Stage2InputRequest) error {
	if req.JobID == "" {
		return fmt.Errorf("%w: stage2 input job_id is required", core.ErrInvalidArgument)
	}
	if req.TenantID == "" {
		return fmt.Errorf("%w: stage2 input tenant_id is required", core.ErrInvalidArgument)
	}
	if req.WorkspaceID == "" {
		return fmt.Errorf("%w: stage2 input workspace_id is required", core.ErrInvalidArgument)
	}
	if len(req.RawEvents) == 0 {
		return fmt.Errorf("%w: stage2 input raw events are required", core.ErrInvalidArgument)
	}
	return nil
}

func cloneMemoryResults(memories []core.MemoryResult) []core.MemoryResult {
	if len(memories) == 0 {
		return []core.MemoryResult{}
	}
	return append([]core.MemoryResult(nil), memories...)
}

func cloneDocumentChunkResults(documents []core.DocumentChunkResult) []core.DocumentChunkResult {
	if len(documents) == 0 {
		return []core.DocumentChunkResult{}
	}
	return append([]core.DocumentChunkResult(nil), documents...)
}

func clonePlans(plans []*core.Plan) []*core.Plan {
	if len(plans) == 0 {
		return []*core.Plan{}
	}
	return append([]*core.Plan(nil), plans...)
}

func cloneNotes(notes []*core.Note) []*core.Note {
	if len(notes) == 0 {
		return []*core.Note{}
	}
	return append([]*core.Note(nil), notes...)
}

func (p *Stage2InputPreparer) loadProfile(ctx context.Context, req Stage2InputRequest) (*core.Profile, error) {
	if p == nil || p.sources.Profiles == nil {
		return nil, nil
	}
	profile, err := p.sources.Profiles.LoadStage2Profile(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("load stage2 profile: %w", err)
	}
	return profile, nil
}

func (p *Stage2InputPreparer) loadMemories(ctx context.Context, req Stage2InputRequest) ([]core.MemoryResult, error) {
	if p == nil || p.sources.Memories == nil {
		return []core.MemoryResult{}, nil
	}
	memories, err := p.sources.Memories.LoadStage2Memories(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("load stage2 memories: %w", err)
	}
	return memories, nil
}

func (p *Stage2InputPreparer) loadDocuments(ctx context.Context, req Stage2InputRequest) ([]core.DocumentChunkResult, error) {
	if p == nil || p.sources.Documents == nil {
		return []core.DocumentChunkResult{}, nil
	}
	documents, err := p.sources.Documents.LoadStage2Documents(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("load stage2 documents: %w", err)
	}
	return documents, nil
}

func (p *Stage2InputPreparer) loadPlans(ctx context.Context, req Stage2InputRequest) ([]*core.Plan, error) {
	if p == nil || p.sources.Plans == nil {
		return []*core.Plan{}, nil
	}
	plans, err := p.sources.Plans.LoadStage2ActivePlans(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("load stage2 active plans: %w", err)
	}
	return plans, nil
}

func (p *Stage2InputPreparer) loadNotes(ctx context.Context, req Stage2InputRequest) ([]*core.Note, error) {
	if p == nil || p.sources.Notes == nil {
		return []*core.Note{}, nil
	}
	notes, err := p.sources.Notes.LoadStage2PinnedNotes(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("load stage2 pinned notes: %w", err)
	}
	return notes, nil
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



<!-- Source: internal/worker/doc.go | bytes=785 | lines=16 | sha16=7a74e678aaa8e353 -->

```go
// ============================================================
// FILE     : internal/worker/doc.go
// PURPOSE  : Documents the background job processor package.
// LAYER    : application
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : package worker
// DEPENDS  : docs/adr-008-package-layout.md, plans/05_runtime-contracts_ingest-recall-apply.md
// USED_BY  : cmd/worker, tests
// ------------------------------------------------------------
// AGENT_NOTE: This package orchestrates jobs; semantic extraction remains Codex-first through internal/reasoning.
// ============================================================

// Package worker claims ingest_jobs and dispatches them to reasoning and graph apply services.
package worker

```



<!-- Source: internal/worker/processor.go | bytes=14608 | lines=435 | sha16=ea025d0b2546ebae -->

```go
// ============================================================
// FILE     : internal/worker/processor.go
// PURPOSE  : Claims worker jobs and runs turn-processing and dreaming background pipelines.
// LAYER    : application
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : Dependencies, Processor, RunResult, JobFailure, NewProcessor
// DEPENDS  : context, errors, fmt, internal/core, internal/graph, internal/reasoning, internal/store
// USED_BY  : cmd/worker, internal/worker tests
// ------------------------------------------------------------
// AGENT_NOTE: Keep this as orchestration only; do not add local text extraction logic here.
// ============================================================

package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/graph"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/reasoning"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/store"
)

const defaultBatchSize = 1

var errUnsupportedApplyWork = errors.New("unsupported apply work")

// Dependencies collects the stores and application services used by the worker processor.
type Dependencies struct {
	WorkerID    string
	BatchSize   int
	Jobs        store.JobStore
	RawEvents   store.RawEventStore
	Reasoner    reasoning.Orchestrator
	ApplyEngine graph.ApplyEngine
	Dreaming    *graph.DreamingService
	Clock       func() time.Time
}

// Processor claims jobs and dispatches them to the correct worker pipeline.
type Processor struct {
	workerID    string
	batchSize   int
	jobs        store.JobStore
	rawEvents   store.RawEventStore
	reasoner    reasoning.Orchestrator
	applyEngine graph.ApplyEngine
	dreaming    *graph.DreamingService
	clock       func() time.Time
}

// RunResult summarizes one processor polling pass.
type RunResult struct {
	Claimed               int
	Completed             int
	Failed                int
	Blocked               int
	AppliedOperationCount int
	MemoryIDCount         int
	TraceWrittenCount     int
	SessionDreamCount     int
	WorkspaceDreamCount   int
	Failures              []JobFailure
}

// JobFailure describes one job failure in a processor polling pass.
type JobFailure struct {
	JobID   string
	JobKind core.JobKind
	Error   string
}

type jobProcessResult struct {
	AppliedOperationCount int
	MemoryIDCount         int
	TraceWrittenCount     int
	SessionDreamCount     int
	WorkspaceDreamCount   int
}

// NewProcessor builds a worker processor.
func NewProcessor(deps Dependencies) (*Processor, error) {
	if deps.Jobs == nil {
		return nil, fmt.Errorf("%w: worker job store is required", core.ErrInvalidArgument)
	}
	if deps.RawEvents == nil {
		return nil, fmt.Errorf("%w: worker raw event store is required", core.ErrInvalidArgument)
	}
	if deps.Reasoner == nil {
		return nil, fmt.Errorf("%w: worker reasoning orchestrator is required", core.ErrInvalidArgument)
	}
	if deps.ApplyEngine == nil {
		return nil, fmt.Errorf("%w: worker apply engine is required", core.ErrInvalidArgument)
	}
	if deps.Dreaming == nil {
		return nil, fmt.Errorf("%w: worker dreaming service is required", core.ErrInvalidArgument)
	}
	batchSize := deps.BatchSize
	if batchSize <= 0 {
		batchSize = defaultBatchSize
	}
	workerID := deps.WorkerID
	if workerID == "" {
		workerID = "worker:default"
	}
	clock := deps.Clock
	if clock == nil {
		clock = time.Now
	}
	return &Processor{
		workerID:    workerID,
		batchSize:   batchSize,
		jobs:        deps.Jobs,
		rawEvents:   deps.RawEvents,
		reasoner:    deps.Reasoner,
		applyEngine: deps.ApplyEngine,
		dreaming:    deps.Dreaming,
		clock:       clock,
	}, nil
}

// RunOnce claims up to the configured batch size and processes each claimed job.
func (p *Processor) RunOnce(ctx context.Context) (RunResult, error) {
	jobs, err := p.jobs.ClaimJobs(ctx, p.workerID, p.batchSize)
	if err != nil {
		return RunResult{}, fmt.Errorf("claim jobs: %w", err)
	}
	result := RunResult{
		Claimed: len(jobs),
	}
	var firstErr error
	for _, job := range jobs {
		if job == nil {
			continue
		}
		jobResult, err := p.processClaimedJob(ctx, job)
		if err != nil {
			if isPermanentJobFailure(err) {
				result.Blocked++
			} else {
				result.Failed++
			}
			result.Failures = append(result.Failures, newJobFailure(job, err))
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		result.Completed++
		result.AppliedOperationCount += jobResult.AppliedOperationCount
		result.MemoryIDCount += jobResult.MemoryIDCount
		result.TraceWrittenCount += jobResult.TraceWrittenCount
		result.SessionDreamCount += jobResult.SessionDreamCount
		result.WorkspaceDreamCount += jobResult.WorkspaceDreamCount
	}
	return result, firstErr
}

func (p *Processor) processClaimedJob(ctx context.Context, job *core.IngestJob) (jobProcessResult, error) {
	jobResult, err := p.handleJob(ctx, job)
	if err != nil {
		failureErr := jobFailureError(job, err)
		if isPermanentJobFailure(err) {
			if blockErr := p.jobs.BlockJob(ctx, job.ID, failureErr); blockErr != nil {
				return jobProcessResult{}, fmt.Errorf("process job %s: %w; block job: %v", job.ID, failureErr, blockErr)
			}
			return jobProcessResult{}, fmt.Errorf("process job %s: %w", job.ID, failureErr)
		}
		if failErr := p.jobs.FailJob(ctx, job.ID, failureErr); failErr != nil {
			return jobProcessResult{}, fmt.Errorf("process job %s: %w; record failure: %v", job.ID, failureErr, failErr)
		}
		return jobProcessResult{}, fmt.Errorf("process job %s: %w", job.ID, failureErr)
	}
	if err := p.jobs.CompleteJob(ctx, job.ID); err != nil {
		return jobProcessResult{}, fmt.Errorf("complete job %s: %w", job.ID, err)
	}
	return jobResult, nil
}

func (p *Processor) handleJob(ctx context.Context, job *core.IngestJob) (jobProcessResult, error) {
	switch job.JobKind {
	case core.JobKindProcessTurnEvent:
		return p.processTurnEvent(ctx, job)
	case core.JobKindDreamSession:
		return p.dreamSession(ctx, job)
	case core.JobKindDreamWorkspace:
		return p.dreamWorkspace(ctx, job)
	default:
		return jobProcessResult{}, fmt.Errorf("%w: unsupported job kind %q", core.ErrInvalidArgument, job.JobKind)
	}
}

func (p *Processor) processTurnEvent(ctx context.Context, job *core.IngestJob) (jobProcessResult, error) {
	if len(job.RawEventIDs) == 0 {
		return jobProcessResult{}, fmt.Errorf("%w: process_turn_event requires raw_event_ids", core.ErrInvalidArgument)
	}
	events, err := p.rawEvents.GetRawEvents(ctx, job.RawEventIDs)
	if err != nil {
		return jobProcessResult{}, fmt.Errorf("load raw event bundle: %w", err)
	}
	events, err = validateAndOrderRawEventBundle(job, events)
	if err != nil {
		return jobProcessResult{}, err
	}

	envelope := p.buildProcessTurnEnvelope(ctx, job, events)
	reasoningResult, err := p.reasoner.ProcessTurn(ctx, envelope)
	if err != nil {
		return jobProcessResult{}, fmt.Errorf("reason process_turn_event: %w", err)
	}
	applyResult, err := p.applyEngine.Apply(ctx, &graph.ApplyRequest{
		JobID:       job.ID,
		TenantID:    job.TenantID,
		WorkspaceID: job.WorkspaceID,
		RawEventIDs: append([]string(nil), job.RawEventIDs...),
		Reasoning:   reasoningResult,
	})
	if err != nil {
		if errors.Is(err, core.ErrNotImplemented) {
			return jobProcessResult{}, fmt.Errorf("%w: apply process_turn_event: %w", errUnsupportedApplyWork, err)
		}
		return jobProcessResult{}, fmt.Errorf("apply process_turn_event: %w", err)
	}
	return jobProcessResultFromApply(applyResult)
}

func (p *Processor) dreamSession(ctx context.Context, job *core.IngestJob) (jobProcessResult, error) {
	payload, err := decodeDreamingJobPayload(job)
	if err != nil {
		return jobProcessResult{}, err
	}
	if strings.TrimSpace(payload.SessionID) == "" {
		return jobProcessResult{}, fmt.Errorf("%w: dream_session payload.session_id is required", core.ErrInvalidArgument)
	}
	result, err := p.dreaming.DreamSession(ctx, &core.DreamSessionRequest{
		JobID:       job.ID,
		TenantID:    job.TenantID,
		WorkspaceID: job.WorkspaceID,
		SessionID:   payload.SessionID,
		Now:         p.clock().UTC(),
	})
	if err != nil {
		return jobProcessResult{}, fmt.Errorf("dream session: %w", err)
	}
	return jobProcessResult{
		SessionDreamCount: result.MidTermPromoted + boolCount(result.SessionSummaryWritten),
	}, nil
}

func (p *Processor) dreamWorkspace(ctx context.Context, job *core.IngestJob) (jobProcessResult, error) {
	result, err := p.dreaming.DreamWorkspace(ctx, &core.DreamWorkspaceRequest{
		JobID:       job.ID,
		TenantID:    job.TenantID,
		WorkspaceID: job.WorkspaceID,
		Now:         p.clock().UTC(),
	})
	if err != nil {
		return jobProcessResult{}, fmt.Errorf("dream workspace: %w", err)
	}
	return jobProcessResult{
		WorkspaceDreamCount: result.LongTermPromoted + result.UltraLongTermPromoted,
	}, nil
}

type dreamingJobPayload struct {
	SessionID string `json:"session_id"`
}

func decodeDreamingJobPayload(job *core.IngestJob) (*dreamingJobPayload, error) {
	payload := &dreamingJobPayload{}
	if len(job.PayloadJSON) == 0 {
		return payload, nil
	}
	if err := json.Unmarshal(job.PayloadJSON, payload); err != nil {
		return nil, fmt.Errorf("%w: dreaming job payload_json must be valid JSON", core.ErrInvalidArgument)
	}
	return payload, nil
}

func boolCount(value bool) int {
	if value {
		return 1
	}
	return 0
}

func jobProcessResultFromApply(result *graph.ApplyResult) (jobProcessResult, error) {
	if result == nil {
		return jobProcessResult{}, fmt.Errorf("%w: apply process_turn_event returned nil result", core.ErrInvalidArgument)
	}
	if result.AppliedOperationCount > 0 && !result.TraceWritten {
		return jobProcessResult{}, fmt.Errorf("%w: apply result reports applied operations without a trace", core.ErrConflict)
	}
	traceWrittenCount := 0
	if result.TraceWritten {
		traceWrittenCount = 1
	}
	return jobProcessResult{
		AppliedOperationCount: result.AppliedOperationCount,
		MemoryIDCount:         len(result.MemoryIDs),
		TraceWrittenCount:     traceWrittenCount,
	}, nil
}

func validateAndOrderRawEventBundle(job *core.IngestJob, events []*core.RawEvent) ([]*core.RawEvent, error) {
	if len(events) != len(job.RawEventIDs) {
		return nil, fmt.Errorf("%w: raw event bundle incomplete for job %s: expected %d events, got %d", core.ErrNotFound, job.ID, len(job.RawEventIDs), len(events))
	}

	expected := make(map[string]struct{}, len(job.RawEventIDs))
	for _, id := range job.RawEventIDs {
		if id == "" {
			return nil, fmt.Errorf("%w: raw event bundle contains empty raw_event_id for job %s", core.ErrInvalidArgument, job.ID)
		}
		if _, exists := expected[id]; exists {
			return nil, fmt.Errorf("%w: raw event bundle contains duplicate raw_event_id %q for job %s", core.ErrInvalidArgument, id, job.ID)
		}
		expected[id] = struct{}{}
	}

	byID := make(map[string]*core.RawEvent, len(events))
	for i, event := range events {
		if event == nil {
			return nil, fmt.Errorf("%w: raw event bundle incomplete for job %s: nil event at index %d", core.ErrNotFound, job.ID, i)
		}
		if _, ok := expected[event.ID]; !ok {
			return nil, fmt.Errorf("%w: raw event bundle contains unexpected raw_event_id %q for job %s", core.ErrConflict, event.ID, job.ID)
		}
		if _, exists := byID[event.ID]; exists {
			return nil, fmt.Errorf("%w: raw event bundle contains duplicate returned raw_event_id %q for job %s", core.ErrConflict, event.ID, job.ID)
		}
		if event.TenantID != job.TenantID {
			return nil, fmt.Errorf("%w: raw event bundle tenant mismatch for raw_event_id %q", core.ErrConflict, event.ID)
		}
		if event.WorkspaceID != job.WorkspaceID {
			return nil, fmt.Errorf("%w: raw event bundle workspace mismatch for raw_event_id %q", core.ErrConflict, event.ID)
		}
		byID[event.ID] = event
	}

	ordered := make([]*core.RawEvent, 0, len(job.RawEventIDs))
	for _, id := range job.RawEventIDs {
		event, ok := byID[id]
		if !ok {
			return nil, fmt.Errorf("%w: raw event bundle incomplete for job %s: missing raw_event_id %q", core.ErrNotFound, job.ID, id)
		}
		ordered = append(ordered, event)
	}
	if _, err := requireSingleRawEventActor(job, ordered); err != nil {
		return nil, err
	}
	return ordered, nil
}

func requireSingleRawEventActor(job *core.IngestJob, events []*core.RawEvent) (string, error) {
	var actorID string
	for _, event := range events {
		currentActorID := strings.TrimSpace(event.ActorID)
		if currentActorID == "" {
			return "", fmt.Errorf("%w: raw event bundle contains empty actor_id for raw_event_id %q in job %s", core.ErrInvalidArgument, event.ID, job.ID)
		}
		if actorID == "" {
			actorID = currentActorID
			continue
		}
		if currentActorID != actorID {
			return "", fmt.Errorf("%w: raw event bundle contains mixed actor_id values for job %s", core.ErrConflict, job.ID)
		}
	}
	if actorID == "" {
		return "", fmt.Errorf("%w: raw event bundle has no actor_id for job %s", core.ErrInvalidArgument, job.ID)
	}
	return actorID, nil
}

func newJobFailure(job *core.IngestJob, err error) JobFailure {
	failure := JobFailure{}
	if job != nil {
		failure.JobID = job.ID
		failure.JobKind = job.JobKind
	}
	if err != nil {
		failure.Error = err.Error()
	}
	return failure
}

func jobFailureError(job *core.IngestJob, err error) error {
	if job == nil {
		return fmt.Errorf("worker job failed: %w", err)
	}
	return fmt.Errorf("worker job failed job_id=%s job_kind=%s raw_event_count=%d: %w", job.ID, job.JobKind, len(job.RawEventIDs), err)
}

func isPermanentJobFailure(err error) bool {
	return errors.Is(err, errUnsupportedApplyWork)
}

func (p *Processor) buildProcessTurnEnvelope(_ context.Context, job *core.IngestJob, events []*core.RawEvent) *reasoning.ProcessTurnEnvelope {
	rawEventIDs := append([]string(nil), job.RawEventIDs...)
	rawEvents := append([]*core.RawEvent(nil), events...)
	stage1 := reasoning.Stage1Input{
		JobID:       job.ID,
		TenantID:    job.TenantID,
		WorkspaceID: job.WorkspaceID,
		RawEvents:   rawEvents,
	}
	return &reasoning.ProcessTurnEnvelope{
		JobID:       job.ID,
		TenantID:    job.TenantID,
		WorkspaceID: job.WorkspaceID,
		RawEventIDs: rawEventIDs,
		RawEvents:   rawEvents,
		Stage1:      stage1,
		Stage2: reasoning.Stage2Input{
			JobID:                job.ID,
			TenantID:             job.TenantID,
			WorkspaceID:          job.WorkspaceID,
			RawEvents:            rawEvents,
			RelevantMemories:     []core.MemoryResult{},
			RelevantDocuments:    []core.DocumentChunkResult{},
			ActivePlans:          []*core.Plan{},
			PinnedNotes:          []*core.Note{},
			RequiredOutputName:   reasoning.StageNameResolve,
			RequiredOutputSchema: reasoning.Stage2ResolveOutputSchemaV0,
		},
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



<!-- Source: internal/worker/stage2_sources.go | bytes=11172 | lines=362 | sha16=8c3893a7c55a1e95 -->

```go
// ============================================================
// FILE     : internal/worker/stage2_sources.go
// PURPOSE  : Adapts existing stores into Stage2InputPreparer source interfaces.
// LAYER    : application
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : Stage2SourceStores, NewStoreBackedStage2InputPreparer, NewStoreBackedStage2InputSources
// DEPENDS  : context, errors, fmt, strings, internal/core, internal/reasoning, internal/store
// USED_BY  : cmd/worker, internal/worker tests
// ------------------------------------------------------------
// AGENT_NOTE: Source adapters may retrieve stored context only; never extract raw text or call Codex here.
// ============================================================

package worker

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/reasoning"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/store"
)

// Stage2SourceStores collects existing store contracts used to prepare Stage 2 context.
type Stage2SourceStores struct {
	Profiles  store.ProfileStore
	Memories  store.MemoryStore
	Documents store.DocumentStore
	Plans     store.PlanStore
	Notes     store.NoteStore
	Groups    store.GroupStore
}

// NewStoreBackedStage2InputPreparer builds a Stage2InputPreparer from existing stores.
func NewStoreBackedStage2InputPreparer(stores Stage2SourceStores) *reasoning.Stage2InputPreparer {
	return reasoning.NewStage2InputPreparer(NewStoreBackedStage2InputSources(stores))
}

// NewStoreBackedStage2InputSources adapts configured stores to Stage2InputPreparer source interfaces.
func NewStoreBackedStage2InputSources(stores Stage2SourceStores) reasoning.Stage2InputSources {
	sources := reasoning.Stage2InputSources{}
	if stores.Profiles != nil {
		sources.Profiles = storeBackedStage2ProfileSource{profiles: stores.Profiles}
	}
	if stores.Memories != nil {
		sources.Memories = storeBackedStage2MemorySource{memories: stores.Memories, groups: stores.Groups}
	}
	if stores.Documents != nil {
		sources.Documents = storeBackedStage2DocumentSource{documents: stores.Documents}
	}
	if stores.Plans != nil {
		sources.Plans = storeBackedStage2PlanSource{plans: stores.Plans}
	}
	if stores.Notes != nil {
		sources.Notes = storeBackedStage2NoteSource{notes: stores.Notes}
	}
	return sources
}

type storeBackedStage2ProfileSource struct {
	profiles store.ProfileStore
}

func (s storeBackedStage2ProfileSource) LoadStage2Profile(ctx context.Context, req reasoning.Stage2InputRequest) (*core.Profile, error) {
	targets := make([]stage2ProfileTarget, 0, 2)
	if actorID := stage2ActorID(req.RawEvents); actorID != "" {
		targets = append(targets, stage2ProfileTarget{entityID: actorID, scope: core.MemoryScopeAgentPrivate})
	}
	if req.WorkspaceID != "" {
		targets = append(targets, stage2ProfileTarget{entityID: "workspace:" + req.WorkspaceID, scope: core.MemoryScopeWorkspaceShared})
	}
	for _, target := range targets {
		profile, err := s.profiles.GetProfile(ctx, target.entityID, target.scope)
		if errors.Is(err, core.ErrNotFound) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("get stage2 profile: %w", err)
		}
		return profile, nil
	}
	return nil, nil
}

type stage2ProfileTarget struct {
	entityID string
	scope    core.MemoryScope
}

type storeBackedStage2MemorySource struct {
	memories store.MemoryStore
	groups   store.GroupStore
}

func (s storeBackedStage2MemorySource) LoadStage2Memories(ctx context.Context, req reasoning.Stage2InputRequest) ([]core.MemoryResult, error) {
	actorID := stage2ActorID(req.RawEvents)
	groupIDs, err := s.visibleGroupIDs(ctx, req, actorID)
	if err != nil {
		return nil, err
	}
	resp, err := s.memories.SearchMemories(ctx, &core.SearchMemoriesRequest{
		TenantID:        req.TenantID,
		WorkspaceID:     req.WorkspaceID,
		OwnerEntityID:   actorID,
		VisibleGroupIDs: groupIDs,
		Query:           stage2StructuredSearchQuery(req),
		Scopes:          stage2VisibleScopes(groupIDs),
		ArtifactClasses: stage2ArtifactClasses(),
	})
	if errors.Is(err, core.ErrNotFound) {
		return []core.MemoryResult{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("search stage2 memories: %w", err)
	}
	if resp == nil || len(resp.Memories) == 0 {
		return []core.MemoryResult{}, nil
	}
	return filterStage2MemoryResults(resp.Memories, actorID, groupIDs), nil
}

func (s storeBackedStage2MemorySource) visibleGroupIDs(ctx context.Context, req reasoning.Stage2InputRequest, actorID string) ([]string, error) {
	if s.groups == nil || strings.TrimSpace(actorID) == "" {
		return []string{}, nil
	}
	memberships, err := s.groups.ListMembershipsForEntity(ctx, req.TenantID, req.WorkspaceID, actorID)
	if errors.Is(err, core.ErrNotFound) {
		return []string{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("list stage2 visible memory groups: %w", err)
	}
	return stage2MembershipGroupIDs(memberships), nil
}

type storeBackedStage2DocumentSource struct {
	documents store.DocumentStore
}

func (s storeBackedStage2DocumentSource) LoadStage2Documents(ctx context.Context, req reasoning.Stage2InputRequest) ([]core.DocumentChunkResult, error) {
	resp, err := s.documents.SearchDocuments(ctx, &core.SearchDocumentsRequest{
		TenantID:    req.TenantID,
		WorkspaceID: req.WorkspaceID,
		Query:       stage2StructuredSearchQuery(req),
	})
	if errors.Is(err, core.ErrNotFound) {
		return []core.DocumentChunkResult{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("search stage2 documents: %w", err)
	}
	if resp == nil || len(resp.Chunks) == 0 {
		return []core.DocumentChunkResult{}, nil
	}
	return append([]core.DocumentChunkResult(nil), resp.Chunks...), nil
}

type storeBackedStage2PlanSource struct {
	plans store.PlanStore
}

func (s storeBackedStage2PlanSource) LoadStage2ActivePlans(ctx context.Context, req reasoning.Stage2InputRequest) ([]*core.Plan, error) {
	actorID := stage2ActorID(req.RawEvents)
	plans, err := s.plans.GetActivePlans(ctx, &core.GetActivePlansRequest{
		TenantID:      req.TenantID,
		WorkspaceID:   req.WorkspaceID,
		OwnerEntityID: actorID,
		Scopes:        stage2BaseVisibleScopes(),
	})
	if errors.Is(err, core.ErrNotFound) {
		return []*core.Plan{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get stage2 active plans: %w", err)
	}
	if len(plans) == 0 {
		return []*core.Plan{}, nil
	}
	return filterStage2Plans(plans, actorID), nil
}

type storeBackedStage2NoteSource struct {
	notes store.NoteStore
}

func (s storeBackedStage2NoteSource) LoadStage2PinnedNotes(ctx context.Context, req reasoning.Stage2InputRequest) ([]*core.Note, error) {
	actorID := stage2ActorID(req.RawEvents)
	notes, err := s.notes.ListPinnedNotes(ctx, &core.ListPinnedNotesRequest{
		TenantID:      req.TenantID,
		WorkspaceID:   req.WorkspaceID,
		OwnerEntityID: actorID,
		Scopes:        stage2BaseVisibleScopes(),
	})
	if errors.Is(err, core.ErrNotFound) {
		return []*core.Note{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("list stage2 pinned notes: %w", err)
	}
	if len(notes) == 0 {
		return []*core.Note{}, nil
	}
	return filterStage2Notes(notes, actorID), nil
}

func stage2ActorID(events []*core.RawEvent) string {
	for _, event := range events {
		if event == nil {
			continue
		}
		if actorID := strings.TrimSpace(event.ActorID); actorID != "" {
			return actorID
		}
	}
	return ""
}

func filterStage2MemoryResults(memories []core.MemoryResult, actorID string, groupIDs []string) []core.MemoryResult {
	if len(memories) == 0 {
		return []core.MemoryResult{}
	}
	visibleGroups := stringSet(groupIDs)
	filtered := make([]core.MemoryResult, 0, len(memories))
	for _, memory := range memories {
		if !stage2ScopeVisibleToActor(memory.Scope, memory.OwnerEntityID, memory.GroupID, actorID, visibleGroups) {
			continue
		}
		filtered = append(filtered, memory)
	}
	return filtered
}

func filterStage2Plans(plans []*core.Plan, actorID string) []*core.Plan {
	if len(plans) == 0 {
		return []*core.Plan{}
	}
	filtered := make([]*core.Plan, 0, len(plans))
	for _, plan := range plans {
		if plan == nil || !stage2ScopeVisibleToActor(plan.Scope, plan.OwnerEntityID, nil, actorID, nil) {
			continue
		}
		filtered = append(filtered, plan)
	}
	return filtered
}

func filterStage2Notes(notes []*core.Note, actorID string) []*core.Note {
	if len(notes) == 0 {
		return []*core.Note{}
	}
	filtered := make([]*core.Note, 0, len(notes))
	for _, note := range notes {
		if note == nil || !stage2ScopeVisibleToActor(note.Scope, note.OwnerEntityID, nil, actorID, nil) {
			continue
		}
		filtered = append(filtered, note)
	}
	return filtered
}

func stage2ScopeVisibleToActor(scope core.MemoryScope, ownerEntityID string, groupID *string, actorID string, visibleGroups map[string]struct{}) bool {
	switch scope {
	case core.MemoryScopeAgentPrivate:
		return actorID != "" && ownerEntityID == actorID
	case core.MemoryScopeWorkspaceShared, core.MemoryScopeSessionScratch:
		return true
	case core.MemoryScopeGroupShared:
		if groupID == nil || visibleGroups == nil {
			return false
		}
		_, ok := visibleGroups[*groupID]
		return ok
	default:
		return false
	}
}

func stage2VisibleScopes(groupIDs []string) []core.MemoryScope {
	scopes := stage2BaseVisibleScopes()
	if len(groupIDs) > 0 {
		scopes = append(scopes, core.MemoryScopeGroupShared)
	}
	return scopes
}

func stage2BaseVisibleScopes() []core.MemoryScope {
	return []core.MemoryScope{
		core.MemoryScopeAgentPrivate,
		core.MemoryScopeWorkspaceShared,
		core.MemoryScopeSessionScratch,
	}
}

func stage2MembershipGroupIDs(memberships []*core.MemoryGroupMembership) []string {
	if len(memberships) == 0 {
		return []string{}
	}
	groupIDs := make([]string, 0, len(memberships))
	seen := make(map[string]struct{}, len(memberships))
	for _, membership := range memberships {
		if membership == nil {
			continue
		}
		groupID := strings.TrimSpace(membership.GroupID)
		if groupID == "" {
			continue
		}
		if _, ok := seen[groupID]; ok {
			continue
		}
		seen[groupID] = struct{}{}
		groupIDs = append(groupIDs, groupID)
	}
	return groupIDs
}

func stringSet(values []string) map[string]struct{} {
	if len(values) == 0 {
		return nil
	}
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		set[trimmed] = struct{}{}
	}
	return set
}

func stage2ArtifactClasses() []core.ArtifactClass {
	return []core.ArtifactClass{
		core.ArtifactClassContext,
		core.ArtifactClassKnowledge,
		core.ArtifactClassTimeline,
		core.ArtifactClassPlan,
	}
}

func stage2StructuredSearchQuery(req reasoning.Stage2InputRequest) string {
	parts := make([]string, 0, 4+len(req.Stage1.CandidateEntities)+len(req.Stage1.CandidateMemories))
	appendPart := func(value string) {
		value = strings.TrimSpace(value)
		if value != "" {
			parts = append(parts, value)
		}
	}
	appendPart(req.Stage1.SummaryHint)
	appendPart(req.Stage1.TaskHint)
	for _, entity := range req.Stage1.CandidateEntities {
		appendPart(entity.DisplayName)
	}
	for _, memory := range req.Stage1.CandidateMemories {
		appendPart(memory.Text)
	}
	return strings.Join(parts, "\n")
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
