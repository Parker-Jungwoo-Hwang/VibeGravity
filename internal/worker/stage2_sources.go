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
		profile, err := s.profiles.GetProfile(ctx, req.TenantID, req.WorkspaceID, target.entityID, target.scope)
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
