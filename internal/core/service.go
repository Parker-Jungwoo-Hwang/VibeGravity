// ============================================================
// FILE     : internal/core/service.go
// PURPOSE  : Defines the primary v1 service contract for all runtime surfaces.
// LAYER    : domain
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : VibeGravityService
// DEPENDS  : context, internal/core/dto.go
// USED_BY  : HTTP API, MCP, Hermes provider, tests
// ------------------------------------------------------------
// AGENT_NOTE: Do not change this interface without updating public runtime docs and tests.
// ============================================================

package core

import (
	"context"
)

// VibeGravityService defines the primary v1 interface for every runtime surface.
type VibeGravityService interface {
	Prefetch(ctx context.Context, req *PrefetchRequest) (*PrefetchResponse, error)
	SyncTurn(ctx context.Context, req *SyncTurnRequest) (*SyncTurnResponse, error)
	AddDocument(ctx context.Context, req *AddDocumentRequest) (*AddDocumentResponse, error)
	SearchMemories(ctx context.Context, req *SearchMemoriesRequest) (*SearchMemoriesResponse, error)
	SearchDocuments(ctx context.Context, req *SearchDocumentsRequest) (*SearchDocumentsResponse, error)
	AddNote(ctx context.Context, req *AddNoteRequest) (*AddNoteResponse, error)
	CreatePlan(ctx context.Context, req *CreatePlanRequest) (*CreatePlanResponse, error)
	UpdatePlan(ctx context.Context, req *UpdatePlanRequest) (*UpdatePlanResponse, error)
	CorrectMemory(ctx context.Context, req *CorrectMemoryRequest) (*CorrectMemoryResponse, error)
	GetTimeline(ctx context.Context, req *GetTimelineRequest) (*GetTimelineResponse, error)
	ExplainMemory(ctx context.Context, req *ExplainMemoryRequest) (*ExplainMemoryResponse, error)
}
