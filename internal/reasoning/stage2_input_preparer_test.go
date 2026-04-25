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
