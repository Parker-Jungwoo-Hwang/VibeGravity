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
	profiles := &fakeProfileStore{profiles: map[string]*core.Profile{
		"agent:hermes-main|agent_private": {
			TenantID:    "tenant_1",
			WorkspaceID: "workspace_1",
			EntityID:    "agent:hermes-main",
			Scope:       core.MemoryScopeAgentPrivate,
			StaticJSON:  json.RawMessage(`{"style":"brief"}`),
		},
	}}
	assembler := NewAssembler(Dependencies{
		Notes:    notes,
		Plans:    plans,
		Profiles: profiles,
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
	if got, want := profiles.calls[0], (profileCall{tenantID: "tenant_1", workspaceID: "workspace_1", entityID: "agent:hermes-main", scope: core.MemoryScopeAgentPrivate}); got != want {
		t.Fatalf("profile request was not tenant/workspace scoped: got %#v want %#v", got, want)
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

func TestAssemblerPrefetch_ReturnsStoredContextWhileStale(t *testing.T) {
	t.Parallel()

	lagSeconds := int64(180)
	assembler := NewAssembler(Dependencies{
		Notes: &fakeNoteStore{notes: []*core.Note{{
			ID:            "note_guardrail",
			Text:          "Do not enable real Codex by default.",
			Scope:         core.MemoryScopeWorkspaceShared,
			OwnerEntityID: "workspace:workspace_1",
			Pinned:        true,
		}}},
		Plans: &fakePlanStore{plans: []*core.Plan{{
			ID:            "plan_degraded",
			Title:         "Keep degraded recall useful while worker catches up.",
			Status:        "active",
			Scope:         core.MemoryScopeAgentPrivate,
			OwnerEntityID: "agent:hermes-main",
		}}},
		Profiles: &fakeProfileStore{profiles: map[string]*core.Profile{
			"agent:hermes-main|agent_private": {
				TenantID:    "tenant_1",
				WorkspaceID: "workspace_1",
				EntityID:    "agent:hermes-main",
				Scope:       core.MemoryScopeAgentPrivate,
				StaticJSON:  json.RawMessage(`{"operator":"prefers truthful degraded status"}`),
			},
		}},
		Summaries: &fakeSessionSummaryStore{summary: &core.SessionSummary{
			ID:          "summary_1",
			TenantID:    "tenant_1",
			WorkspaceID: "workspace_1",
			SessionID:   "session_1",
			SummaryText: "Previous session verified the mocked Codex bridge boundary.",
		}},
		Memories: &fakeMemoryStore{resp: &core.SearchMemoriesResponse{
			Memories: []core.MemoryResult{{
				MemoryID:      "mem_stored",
				Text:          "Stored memory remains available during Codex outage.",
				Scope:         core.MemoryScopeAgentPrivate,
				OwnerEntityID: "agent:hermes-main",
				LatestFlag:    true,
			}},
		}},
		Documents: &fakeDocumentStore{resp: &core.SearchDocumentsResponse{
			Chunks: []core.DocumentChunkResult{{
				ChunkID:    "chunk_1",
				DocumentID: "doc_1",
				Text:       "Runtime contract says degraded mode never returns empty if useful context exists.",
				Score:      0.9,
			}},
		}},
		Groups: &fakeGroupStore{},
		Freshness: fakeFreshnessProvider{state: Freshness{
			Freshness:       "stale",
			LagSeconds:      &lagSeconds,
			Reasons:         []string{"worker_backlog_stale"},
			AffectedSources: []string{"memories", "profile", "session_summaries"},
		}},
	})

	resp, err := assembler.Prefetch(context.Background(), testPrefetchRequest())
	if err != nil {
		t.Fatalf("Prefetch returned error: %v", err)
	}

	if len(resp.Blocks) == 0 {
		t.Fatalf("degraded recall should still return useful stored context")
	}
	if !resp.Meta.Degraded || resp.Meta.Freshness != "stale" || resp.Meta.FreshnessLagSeconds == nil || *resp.Meta.FreshnessLagSeconds != lagSeconds {
		t.Fatalf("expected stale degraded metadata, got %#v", resp.Meta)
	}
	for _, wantKind := range []string{"pinned_note", "active_plan", "profile_static", "session_summary", "memory", "document"} {
		block, ok := findRecallBlock(resp.Blocks, wantKind)
		if !ok {
			t.Fatalf("expected degraded recall to include %s in %#v", wantKind, resp.Blocks)
		}
		switch wantKind {
		case "profile_static", "session_summary", "memory":
			if block.Freshness != "stale" {
				t.Fatalf("expected derived block %s to be stale, got %#v", wantKind, block)
			}
		default:
			if block.Freshness != "stored" {
				t.Fatalf("expected non-derived block %s to stay stored, got %#v", wantKind, block)
			}
		}
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

func findRecallBlock(blocks []core.RecallBlock, kind string) (core.RecallBlock, bool) {
	for _, block := range blocks {
		if block.Kind == kind {
			return block, true
		}
	}
	return core.RecallBlock{}, false
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
	calls    []profileCall
}

func (s *fakeProfileStore) GetProfile(_ context.Context, tenantID string, workspaceID string, entityID string, scope core.MemoryScope) (*core.Profile, error) {
	s.calls = append(s.calls, profileCall{tenantID: tenantID, workspaceID: workspaceID, entityID: entityID, scope: scope})
	profile, ok := s.profiles[entityID+"|"+string(scope)]
	if !ok {
		return nil, core.ErrNotFound
	}
	return profile, nil
}

type profileCall struct {
	tenantID    string
	workspaceID string
	entityID    string
	scope       core.MemoryScope
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

type fakeSessionSummaryStore struct {
	summary *core.SessionSummary
}

func (s *fakeSessionSummaryStore) UpsertSessionSummary(context.Context, *core.SessionSummary) error {
	return core.ErrNotImplemented
}

func (s *fakeSessionSummaryStore) GetSessionSummary(_ context.Context, tenantID string, workspaceID string, sessionID string) (*core.SessionSummary, error) {
	if s.summary == nil {
		return nil, core.ErrNotFound
	}
	if s.summary.TenantID != tenantID || s.summary.WorkspaceID != workspaceID || s.summary.SessionID != sessionID {
		return nil, core.ErrNotFound
	}
	return s.summary, nil
}

type fakeDocumentStore struct {
	resp *core.SearchDocumentsResponse
}

func (s *fakeDocumentStore) AddDocumentWithChunks(context.Context, *core.Document, []*core.DocumentChunk) error {
	return core.ErrNotImplemented
}

func (s *fakeDocumentStore) AddDocument(context.Context, *core.Document) error {
	return core.ErrNotImplemented
}

func (s *fakeDocumentStore) AddDocumentChunks(context.Context, []*core.DocumentChunk) error {
	return core.ErrNotImplemented
}

func (s *fakeDocumentStore) SearchDocuments(context.Context, *core.SearchDocumentsRequest) (*core.SearchDocumentsResponse, error) {
	return s.resp, nil
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
