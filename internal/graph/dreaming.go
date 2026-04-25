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
