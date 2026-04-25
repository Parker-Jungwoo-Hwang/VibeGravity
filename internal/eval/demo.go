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
