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
