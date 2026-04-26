// ============================================================
// FILE     : internal/kernel/service.go
// PURPOSE  : Implements core.VibeGravityService as a facade over product services.
// LAYER    : application
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : Dependencies, Service, NewService
// DEPENDS  : internal/core, internal/corrections, internal/documents, internal/ingest, internal/plans, internal/recall, internal/timeline
// USED_BY  : internal/runtime, tests, HTTP, Hermes, and MCP adapters
// ------------------------------------------------------------
// AGENT_NOTE: Keep this facade thin; route product behavior to the package that owns the contract.
// ============================================================

package kernel

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/corrections"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/documents"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/ingest"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/plans"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/recall"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/store"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/timeline"
)

// Dependencies contains the product services composed by the kernel.
type Dependencies struct {
	Ingest      *ingest.Service
	Recall      *recall.Assembler
	Notes       store.NoteStore
	Plans       store.PlanStore
	Memories    store.MemoryStore
	Corrections store.CorrectionStore
	Jobs        store.CorrectionApplyJobStore
	Timeline    store.TimelineStore
	Documents   store.DocumentStore
}

// Service is the concrete v1 VibeGravity service facade.
type Service struct {
	ingest      *ingest.Service
	recall      *recall.Assembler
	notes       store.NoteStore
	memories    store.MemoryStore
	documents   *documents.Service
	plans       *plans.Service
	corrections *corrections.Service
	timeline    *timeline.Service
}

var _ core.VibeGravityService = (*Service)(nil)

// NewService creates the concrete VibeGravity service.
func NewService(deps Dependencies) (*Service, error) {
	if deps.Ingest == nil {
		return nil, fmt.Errorf("%w: ingest service is required", core.ErrInvalidArgument)
	}
	if deps.Recall == nil {
		return nil, fmt.Errorf("%w: recall assembler is required", core.ErrInvalidArgument)
	}
	return &Service{
		ingest:      deps.Ingest,
		recall:      deps.Recall,
		notes:       deps.Notes,
		memories:    deps.Memories,
		documents:   documents.NewService(deps.Documents),
		plans:       plans.NewService(deps.Plans),
		corrections: corrections.NewService(deps.Memories, deps.Corrections, deps.Jobs),
		timeline:    timeline.NewService(deps.Timeline),
	}, nil
}

// Prefetch assembles a next-turn recall pack.
func (s *Service) Prefetch(ctx context.Context, req *core.PrefetchRequest) (*core.PrefetchResponse, error) {
	return s.recall.Prefetch(ctx, req)
}

// SyncTurn records turn events on the hot path.
func (s *Service) SyncTurn(ctx context.Context, req *core.SyncTurnRequest) (*core.SyncTurnResponse, error) {
	return s.ingest.SyncTurn(ctx, req)
}

// AddDocument stores a document and its initial lexical retrieval chunks.
func (s *Service) AddDocument(ctx context.Context, req *core.AddDocumentRequest) (*core.AddDocumentResponse, error) {
	return s.documents.AddDocument(ctx, req)
}

// SearchMemories delegates memory search to storage.
func (s *Service) SearchMemories(ctx context.Context, req *core.SearchMemoriesRequest) (*core.SearchMemoriesResponse, error) {
	if s.memories == nil {
		return nil, fmt.Errorf("%w: search memories", core.ErrNotImplemented)
	}
	if req == nil {
		return nil, fmt.Errorf("%w: search memories request is required", core.ErrInvalidArgument)
	}
	if err := requireFields(map[string]string{
		"tenant_id":    req.TenantID,
		"workspace_id": req.WorkspaceID,
	}); err != nil {
		return nil, err
	}
	if len(req.Scopes) == 0 {
		req.Scopes = []core.MemoryScope{
			core.MemoryScopeAgentPrivate,
			core.MemoryScopeWorkspaceShared,
			core.MemoryScopeSessionScratch,
		}
	}
	if len(req.ArtifactClasses) == 0 {
		req.ArtifactClasses = []core.ArtifactClass{
			core.ArtifactClassContext,
			core.ArtifactClassKnowledge,
			core.ArtifactClassTimeline,
			core.ArtifactClassPlan,
		}
	}
	return s.memories.SearchMemories(ctx, req)
}

// SearchDocuments delegates document search to storage.
func (s *Service) SearchDocuments(ctx context.Context, req *core.SearchDocumentsRequest) (*core.SearchDocumentsResponse, error) {
	return s.documents.SearchDocuments(ctx, req)
}

// AddNote creates a human-authored recall control note.
func (s *Service) AddNote(ctx context.Context, req *core.AddNoteRequest) (*core.AddNoteResponse, error) {
	if s.notes == nil {
		return nil, fmt.Errorf("%w: add note", core.ErrNotImplemented)
	}
	if req == nil {
		return nil, fmt.Errorf("%w: add note request is required", core.ErrInvalidArgument)
	}
	if err := requireFields(map[string]string{
		"tenant_id":       req.TenantID,
		"workspace_id":    req.WorkspaceID,
		"scope":           string(req.Scope),
		"owner_entity_id": req.OwnerEntityID,
		"text":            req.Text,
	}); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	note := &core.Note{
		TenantID:      req.TenantID,
		WorkspaceID:   req.WorkspaceID,
		NoteKind:      valueOr(req.NoteKind, "operator"),
		Scope:         req.Scope,
		OwnerEntityID: req.OwnerEntityID,
		Text:          req.Text,
		Pinned:        req.Pinned,
		ExpiresAt:     req.ExpiresAt,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := s.notes.AddNote(ctx, note); err != nil {
		return nil, err
	}
	return &core.AddNoteResponse{NoteID: note.ID, Status: "created"}, nil
}

// CreatePlan creates a structured plan and its initial items.
func (s *Service) CreatePlan(ctx context.Context, req *core.CreatePlanRequest) (*core.CreatePlanResponse, error) {
	return s.plans.CreatePlan(ctx, req)
}

// UpdatePlan updates a structured plan and optionally replaces provided items.
func (s *Service) UpdatePlan(ctx context.Context, req *core.UpdatePlanRequest) (*core.UpdatePlanResponse, error) {
	return s.plans.UpdatePlan(ctx, req)
}

// CorrectMemory records human correction intent and applies an operator-driven supersession.
func (s *Service) CorrectMemory(ctx context.Context, req *core.CorrectMemoryRequest) (*core.CorrectMemoryResponse, error) {
	return s.corrections.CorrectMemory(ctx, req)
}

// GetTimeline assembles a read-only operator timeline over existing artifacts.
func (s *Service) GetTimeline(ctx context.Context, req *core.GetTimelineRequest) (*core.GetTimelineResponse, error) {
	return s.timeline.GetTimeline(ctx, req)
}

// ExplainMemory delegates provenance lookup to storage.
func (s *Service) ExplainMemory(ctx context.Context, req *core.ExplainMemoryRequest) (*core.ExplainMemoryResponse, error) {
	if s.memories == nil {
		return nil, fmt.Errorf("%w: explain memory", core.ErrNotImplemented)
	}
	if req == nil {
		return nil, fmt.Errorf("%w: explain memory request is required", core.ErrInvalidArgument)
	}
	if err := requireFields(map[string]string{
		"tenant_id":    req.TenantID,
		"workspace_id": req.WorkspaceID,
		"memory_id":    req.MemoryID,
	}); err != nil {
		return nil, err
	}
	return s.memories.ExplainMemory(ctx, req)
}

func requireFields(fields map[string]string) error {
	for name, value := range fields {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%w: %s is required", core.ErrInvalidArgument, name)
		}
	}
	return nil
}

func valueOr(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
