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
	"github.com/parker-jungwoo-hwang/vibegravity/internal/corrections"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/documents"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/plans"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/timeline"
)

func TestAddDocumentStoresDocumentAndChunks(t *testing.T) {
	t.Parallel()

	documents := &fakeDocumentStore{}
	service := newTestService(t, Dependencies{Documents: documents})

	resp, err := service.AddDocument(context.Background(), &core.AddDocumentRequest{
		TenantID:    "tenant_1",
		WorkspaceID: "workspace_1",
		Source:      "operator_upload",
		Title:       "Runtime Notes",
		Content:     strings.Repeat("A", 1805),
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
	service := newTestService(t, Dependencies{Documents: documents})

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
	service := newTestService(t, Dependencies{Plans: plans})
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

	service := newTestService(t, Dependencies{
		Memories:    &fakeKernelMemoryStore{},
		Corrections: &fakeCorrectionStore{},
		Jobs:        &fakeCorrectionApplyJobStore{},
	})

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

	service := newTestService(t, Dependencies{
		Memories:    &fakeKernelMemoryStore{err: core.ErrNotFound},
		Corrections: &fakeCorrectionStore{},
		Jobs:        &fakeCorrectionApplyJobStore{},
	})

	_, err := service.CorrectMemory(context.Background(), validCorrectionRequest())
	if !errors.Is(err, core.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestCorrectMemoryRecordsRawEventAndCorrection(t *testing.T) {
	t.Parallel()

	corrections := &fakeCorrectionStore{}
	memories := &fakeKernelMemoryStore{memory: validCorrectionTargetMemory()}
	service := newTestService(t, Dependencies{
		Memories:    memories,
		Corrections: corrections,
		Jobs:        &fakeCorrectionApplyJobStore{},
	})

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
	if memories.updateMemory == nil || memories.updateTrace == nil || memories.updateEdge == nil {
		t.Fatalf("correction did not apply graph supersession: memory=%#v trace=%#v edge=%#v", memories.updateMemory, memories.updateTrace, memories.updateEdge)
	}
	if memories.updateMemory.Text != "Use the newer fact." || memories.updateMemory.Scope != core.MemoryScopeWorkspaceShared || memories.updateMemory.OwnerEntityID != "agent:hermes-main" {
		t.Fatalf("replacement memory did not preserve target boundary and corrected text: %#v", memories.updateMemory)
	}
	if memories.updateTrace.RawEventIDs[0] != "evt_correction" || !memories.updateTrace.OperatorCorrectionFlag || memories.updateTrace.ReasoningStage != "operator_correction" {
		t.Fatalf("correction trace did not preserve operator provenance: %#v", memories.updateTrace)
	}
	if memories.updateTrace.ReasoningJobID == "" || strings.HasPrefix(memories.updateTrace.ReasoningJobID, "correction:") {
		t.Fatalf("correction trace must use a real correction apply job id, got %#v", memories.updateTrace)
	}
	if memories.updateEdge.FromMemoryID != memories.updateMemory.ID || memories.updateEdge.ToMemoryID != "mem_1" || memories.updateEdge.EdgeKind != core.EdgeKindUpdates {
		t.Fatalf("correction updates edge mismatch: %#v", memories.updateEdge)
	}
	if memories.updateEdge.CreatedByJobID != memories.updateTrace.ReasoningJobID {
		t.Fatalf("trace and edge must share correction apply job id: trace=%q edge=%q", memories.updateTrace.ReasoningJobID, memories.updateEdge.CreatedByJobID)
	}
}

func TestCorrectMemoryIdempotentRetryReturnsRecordedArtifact(t *testing.T) {
	t.Parallel()

	corrections := &fakeCorrectionStore{
		recorded: &core.MemoryCorrection{
			ID:             "corr_existing",
			TenantID:       "tenant_1",
			WorkspaceID:    "workspace_1",
			MemoryID:       "mem_1",
			RawEventID:     "evt_existing",
			IdempotencyKey: "correction_1",
			CorrectionText: "Use the newer fact.",
			OperatorID:     "operator_1",
			EvidenceJSON:   json.RawMessage(`{}`),
			Status:         "recorded",
		},
	}
	service := newTestService(t, Dependencies{
		Memories:    &fakeKernelMemoryStore{memory: validCorrectionTargetMemory()},
		Corrections: corrections,
		Jobs:        &fakeCorrectionApplyJobStore{},
	})

	resp, err := service.CorrectMemory(context.Background(), validCorrectionRequest())
	if err != nil {
		t.Fatalf("CorrectMemory retry returned error: %v", err)
	}
	if resp.RawEventID != "evt_existing" || resp.CorrectionID != "corr_existing" || resp.Status != "applied" || !resp.TraceWritten {
		t.Fatalf("idempotent retry did not return existing correction artifact: %#v", resp)
	}
}

func TestCorrectMemoryIdempotentRetryBypassesNonLatestTargetPrecheck(t *testing.T) {
	t.Parallel()

	target := validCorrectionTargetMemory()
	target.Status = core.MemoryStatusSuperseded
	target.LatestFlag = false
	corrections := &fakeCorrectionStore{
		recorded: &core.MemoryCorrection{
			ID:             "corr_existing",
			TenantID:       "tenant_1",
			WorkspaceID:    "workspace_1",
			MemoryID:       "mem_1",
			RawEventID:     "evt_existing",
			IdempotencyKey: "correction_1",
			CorrectionText: "Use the newer fact.",
			OperatorID:     "operator_1",
			EvidenceJSON:   json.RawMessage(`{}`),
			Status:         "applied",
		},
	}
	service := newTestService(t, Dependencies{
		Memories:    &fakeKernelMemoryStore{memory: target},
		Corrections: corrections,
		Jobs:        &fakeCorrectionApplyJobStore{},
	})

	resp, err := service.CorrectMemory(context.Background(), validCorrectionRequest())
	if err != nil {
		t.Fatalf("CorrectMemory replay returned error: %v", err)
	}
	if resp.CorrectionID != "corr_existing" || resp.RawEventID != "evt_existing" || resp.Status != "applied" {
		t.Fatalf("idempotent replay did not return existing applied correction: %#v", resp)
	}
}

func TestCorrectMemoryRejectsReusedIdempotencyKeyWithDifferentEvidence(t *testing.T) {
	t.Parallel()

	corrections := &fakeCorrectionStore{
		recorded: &core.MemoryCorrection{
			ID:             "corr_existing",
			TenantID:       "tenant_1",
			WorkspaceID:    "workspace_1",
			MemoryID:       "mem_other",
			RawEventID:     "evt_existing",
			IdempotencyKey: "correction_1",
			CorrectionText: "Different text.",
			OperatorID:     "operator_1",
			EvidenceJSON:   json.RawMessage(`{}`),
			Status:         "recorded",
		},
	}
	service := newTestService(t, Dependencies{
		Memories:    &fakeKernelMemoryStore{memory: validCorrectionTargetMemory()},
		Corrections: corrections,
		Jobs:        &fakeCorrectionApplyJobStore{},
	})

	_, err := service.CorrectMemory(context.Background(), validCorrectionRequest())
	if !errors.Is(err, core.ErrConflict) {
		t.Fatalf("expected ErrConflict for reused correction idempotency key, got %v", err)
	}
}

func TestCorrectMemoryRejectsNonLatestTargetBeforeRecordingCorrection(t *testing.T) {
	t.Parallel()

	target := validCorrectionTargetMemory()
	target.Status = core.MemoryStatusSuperseded
	target.LatestFlag = false
	corrections := &fakeCorrectionStore{}
	service := newTestService(t, Dependencies{
		Memories:    &fakeKernelMemoryStore{memory: target},
		Corrections: corrections,
		Jobs:        &fakeCorrectionApplyJobStore{},
	})

	_, err := service.CorrectMemory(context.Background(), validCorrectionRequest())
	if !errors.Is(err, core.ErrConflict) {
		t.Fatalf("expected ErrConflict for non-latest correction target, got %v", err)
	}
	if corrections.event != nil || corrections.correction != nil {
		t.Fatalf("non-latest target must not record correction side effects: event=%#v correction=%#v", corrections.event, corrections.correction)
	}
}

func TestCorrectMemoryRejectsPrivateTargetWithoutEntityVisibility(t *testing.T) {
	t.Parallel()

	target := validCorrectionTargetMemory()
	target.Scope = core.MemoryScopeAgentPrivate
	target.OwnerEntityID = "agent:hermes-main"
	corrections := &fakeCorrectionStore{}
	service := newTestService(t, Dependencies{
		Memories:    &fakeKernelMemoryStore{memory: target},
		Corrections: corrections,
		Jobs:        &fakeCorrectionApplyJobStore{},
	})

	req := validCorrectionRequest()
	_, err := service.CorrectMemory(context.Background(), req)
	if !errors.Is(err, core.ErrNotFound) {
		t.Fatalf("expected private target without entity visibility to be hidden, got %v", err)
	}
	if corrections.event != nil || corrections.correction != nil {
		t.Fatalf("invisible private target must not record correction side effects: event=%#v correction=%#v", corrections.event, corrections.correction)
	}
}

func TestCorrectMemoryAllowsPrivateTargetForVisibleOwner(t *testing.T) {
	t.Parallel()

	target := validCorrectionTargetMemory()
	target.Scope = core.MemoryScopeAgentPrivate
	target.OwnerEntityID = "agent:hermes-main"
	service := newTestService(t, Dependencies{
		Memories:    &fakeKernelMemoryStore{memory: target},
		Corrections: &fakeCorrectionStore{},
		Jobs:        &fakeCorrectionApplyJobStore{},
	})

	req := validCorrectionRequest()
	req.EntityID = "agent:hermes-main"
	resp, err := service.CorrectMemory(context.Background(), req)
	if err != nil {
		t.Fatalf("CorrectMemory returned error for visible private target: %v", err)
	}
	if resp.Status != "applied" {
		t.Fatalf("unexpected private correction response: %#v", resp)
	}
}

func TestCorrectMemoryRejectsGroupSharedTargetWithoutVisibleGroup(t *testing.T) {
	t.Parallel()

	groupID := "group_design"
	target := validCorrectionTargetMemory()
	target.Scope = core.MemoryScopeGroupShared
	target.GroupID = &groupID
	corrections := &fakeCorrectionStore{}
	service := newTestService(t, Dependencies{
		Memories:    &fakeKernelMemoryStore{memory: target},
		Corrections: corrections,
		Jobs:        &fakeCorrectionApplyJobStore{},
	})

	req := validCorrectionRequest()
	req.VisibleGroupIDs = []string{"group_ops"}
	_, err := service.CorrectMemory(context.Background(), req)
	if !errors.Is(err, core.ErrNotFound) {
		t.Fatalf("expected group target without visible group to be hidden, got %v", err)
	}
	if corrections.event != nil || corrections.correction != nil {
		t.Fatalf("invisible group target must not record correction side effects: event=%#v correction=%#v", corrections.event, corrections.correction)
	}
}

func TestCorrectMemoryAllowsGroupSharedTargetForVisibleGroup(t *testing.T) {
	t.Parallel()

	groupID := "group_design"
	target := validCorrectionTargetMemory()
	target.Scope = core.MemoryScopeGroupShared
	target.GroupID = &groupID
	service := newTestService(t, Dependencies{
		Memories:    &fakeKernelMemoryStore{memory: target},
		Corrections: &fakeCorrectionStore{},
		Jobs:        &fakeCorrectionApplyJobStore{},
	})

	req := validCorrectionRequest()
	req.VisibleGroupIDs = []string{"group_design"}
	resp, err := service.CorrectMemory(context.Background(), req)
	if err != nil {
		t.Fatalf("CorrectMemory returned error for visible group target: %v", err)
	}
	if resp.Status != "applied" {
		t.Fatalf("unexpected group correction response: %#v", resp)
	}
}

func TestCorrectMemoryDoesNotReportSuccessWhenSupersessionFails(t *testing.T) {
	t.Parallel()

	storeErr := errors.New("supersession failed")
	memories := &fakeKernelMemoryStore{memory: validCorrectionTargetMemory(), updateErr: storeErr}
	service := newTestService(t, Dependencies{
		Memories:    memories,
		Corrections: &fakeCorrectionStore{},
		Jobs:        &fakeCorrectionApplyJobStore{},
	})

	resp, err := service.CorrectMemory(context.Background(), validCorrectionRequest())
	if !errors.Is(err, storeErr) {
		t.Fatalf("CorrectMemory error = %v, want %v", err, storeErr)
	}
	if resp != nil {
		t.Fatalf("failed supersession must not return success: %#v", resp)
	}
}

func TestCorrectMemoryUsesStableCorrectionApplyJobForSupersession(t *testing.T) {
	t.Parallel()

	jobs := &fakeCorrectionApplyJobStore{}
	memories := &fakeKernelMemoryStore{memory: validCorrectionTargetMemory()}
	service := newTestService(t, Dependencies{
		Memories:    memories,
		Corrections: &fakeCorrectionStore{},
		Jobs:        jobs,
	})

	_, err := service.CorrectMemory(context.Background(), validCorrectionRequest())
	if err != nil {
		t.Fatalf("CorrectMemory returned error: %v", err)
	}
	wantJobID := core.CorrectionApplyJobID("tenant_1", "workspace_1", "corr_1", "mem_1", "correction_1")
	if jobs.job == nil || jobs.job.ID != wantJobID || jobs.job.Status != "complete" || jobs.job.JobKind != core.JobKindCorrectionApply {
		t.Fatalf("correction apply job was not deterministic and complete: %#v", jobs.job)
	}
	if len(jobs.job.RawEventIDs) != 1 || jobs.job.RawEventIDs[0] != "evt_correction" {
		t.Fatalf("correction apply job lost raw correction event provenance: %#v", jobs.job)
	}
	if memories.updateTrace.ReasoningJobID != wantJobID || memories.updateEdge.CreatedByJobID != wantJobID {
		t.Fatalf("supersession did not use correction apply job id: trace=%q edge=%q want=%q", memories.updateTrace.ReasoningJobID, memories.updateEdge.CreatedByJobID, wantJobID)
	}
}

func TestCorrectMemoryDoesNotApplySupersessionWhenCorrectionApplyJobFails(t *testing.T) {
	t.Parallel()

	jobErr := errors.New("job insert failed")
	memories := &fakeKernelMemoryStore{memory: validCorrectionTargetMemory()}
	service := newTestService(t, Dependencies{
		Memories:    memories,
		Corrections: &fakeCorrectionStore{},
		Jobs:        &fakeCorrectionApplyJobStore{err: jobErr},
	})

	resp, err := service.CorrectMemory(context.Background(), validCorrectionRequest())
	if !errors.Is(err, jobErr) {
		t.Fatalf("CorrectMemory error = %v, want %v", err, jobErr)
	}
	if resp != nil {
		t.Fatalf("failed correction apply job must not return success: %#v", resp)
	}
	if memories.updateMemory != nil || memories.updateTrace != nil || memories.updateEdge != nil {
		t.Fatalf("supersession should not run without a real correction apply job: memory=%#v trace=%#v edge=%#v", memories.updateMemory, memories.updateTrace, memories.updateEdge)
	}
}

func TestGetTimelineDefaultsScopesLimitAndDelegates(t *testing.T) {
	t.Parallel()

	timeline := &fakeTimelineStore{}
	service := newTestService(t, Dependencies{Timeline: timeline})

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
	if timeline.req == nil || timeline.req.Limit != 50 {
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

	service := newTestService(t, Dependencies{Timeline: &fakeTimelineStore{}})

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
		Limit:       101,
	})
	if !errors.Is(err, core.ErrInvalidArgument) {
		t.Fatalf("expected invalid limit error, got %v", err)
	}
}

func TestGetTimelineExcludesGroupSharedUntilMembershipFiltering(t *testing.T) {
	t.Parallel()

	timeline := &fakeTimelineStore{}
	service := newTestService(t, Dependencies{Timeline: timeline})

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
	service := newTestService(t, Dependencies{Memories: memories})

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

func newTestService(t *testing.T, deps Dependencies) *Service {
	t.Helper()
	return &Service{
		recall:      deps.Recall,
		notes:       deps.Notes,
		memories:    deps.Memories,
		documents:   documents.NewService(deps.Documents),
		plans:       plans.NewService(deps.Plans),
		corrections: corrections.NewService(deps.Memories, deps.Corrections, deps.Jobs),
		timeline:    timeline.NewService(deps.Timeline),
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

func (s *fakeKernelMemoryStore) CreateCorrectionSupersession(_ context.Context, memory *core.Memory, trace *core.MemoryTrace, edge *core.MemoryEdge, _ string) error {
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

type fakeCorrectionApplyJobStore struct {
	job *core.IngestJob
	err error
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

func (s *fakeCorrectionStore) GetMemoryCorrectionByIdempotency(context.Context, string, string, string) (*core.MemoryCorrection, error) {
	if s.recorded != nil {
		return s.recorded, nil
	}
	return nil, core.ErrNotFound
}

func (s *fakeCorrectionApplyJobStore) EnsureCorrectionApplyJob(_ context.Context, job *core.IngestJob) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	s.job = job
	return job.ID, nil
}
