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
