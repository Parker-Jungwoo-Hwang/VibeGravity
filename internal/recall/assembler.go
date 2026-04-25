// ============================================================
// FILE     : internal/recall/assembler.go
// PURPOSE  : Assembles budget-aware typed recall blocks for prefetch requests.
// LAYER    : application
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : Dependencies, Assembler, NewAssembler
// DEPENDS  : internal/core, internal/store
// USED_BY  : internal/kernel, tests
// ------------------------------------------------------------
// AGENT_NOTE: Keep recall typed before rendering so Hermes, MCP, and HTTP share one meaning.
// ============================================================

package recall

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/store"
)

const (
	defaultBudgetTokens = 2200
	smallBudgetTokens   = 1000
	richBudgetTokens    = 4000
)

// Dependencies collects optional recall candidate stores.
type Dependencies struct {
	Notes     store.NoteStore
	Plans     store.PlanStore
	Memories  store.MemoryStore
	Documents store.DocumentStore
	Profiles  store.ProfileStore
	Summaries store.SessionSummaryStore
	Groups    store.GroupStore
	Freshness FreshnessProvider
	Clock     func() time.Time
}

// Assembler builds prefetch recall packs from typed candidate pools.
type Assembler struct {
	notes     store.NoteStore
	plans     store.PlanStore
	memories  store.MemoryStore
	documents store.DocumentStore
	profiles  store.ProfileStore
	summaries store.SessionSummaryStore
	groups    store.GroupStore
	freshness FreshnessProvider
	clock     func() time.Time
}

type candidateBlock struct {
	block  core.RecallBlock
	source string
	rank   float64
}

// NewAssembler builds a recall assembler. Missing stores are treated as degraded mode.
func NewAssembler(deps Dependencies) *Assembler {
	clock := deps.Clock
	if clock == nil {
		clock = time.Now
	}
	return &Assembler{
		notes:     deps.Notes,
		plans:     deps.Plans,
		memories:  deps.Memories,
		documents: deps.Documents,
		profiles:  deps.Profiles,
		summaries: deps.Summaries,
		groups:    deps.Groups,
		freshness: deps.Freshness,
		clock:     clock,
	}
}

// Prefetch assembles a budget-aware typed recall pack.
func (a *Assembler) Prefetch(ctx context.Context, req *core.PrefetchRequest) (*core.PrefetchResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("%w: prefetch request is required", core.ErrInvalidArgument)
	}
	if err := validatePrefetchRequest(req); err != nil {
		return nil, err
	}

	groupIDs, err := a.visibleGroupIDs(ctx, req)
	if err != nil {
		return nil, err
	}
	controlScopes := baseVisibleScopes()
	memoryScopes := visibleScopes(groupIDs)
	candidates := make([]candidateBlock, 0, 16)

	candidates, err = a.addPinnedNotes(ctx, req, controlScopes, candidates)
	if err != nil {
		return nil, err
	}
	candidates, err = a.addActivePlans(ctx, req, controlScopes, candidates)
	if err != nil {
		return nil, err
	}
	candidates, err = a.addProfiles(ctx, req, candidates)
	if err != nil {
		return nil, err
	}
	candidates, err = a.addSessionSummary(ctx, req, candidates)
	if err != nil {
		return nil, err
	}
	candidates, err = a.addMemories(ctx, req, memoryScopes, groupIDs, candidates)
	if err != nil {
		return nil, err
	}
	candidates, err = a.addDocuments(ctx, req, candidates)
	if err != nil {
		return nil, err
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].block.Priority != candidates[j].block.Priority {
			return candidates[i].block.Priority > candidates[j].block.Priority
		}
		return candidates[i].rank > candidates[j].rank
	})

	blocks, sources, estimated := packCandidates(candidates, budgetFor(req))
	freshness := a.recallFreshness(ctx, req)
	blocks = applyRecallFreshness(blocks, freshness)
	degradedReasons := append(a.degradedReasons(), freshness.Reasons...)
	degradedReasons = uniqueStrings(degradedReasons)
	return &core.PrefetchResponse{
		Blocks: blocks,
		Meta: core.RecallMeta{
			EstimatedTokens:     estimated,
			Sources:             sources,
			Freshness:           freshness.Freshness,
			FreshnessLagSeconds: freshness.LagSeconds,
			Degraded:            len(degradedReasons) > 0,
			DegradedReasons:     degradedReasons,
		},
	}, nil
}

func (a *Assembler) recallFreshness(ctx context.Context, req *core.PrefetchRequest) Freshness {
	if a.freshness == nil {
		return Freshness{Freshness: "stored"}
	}
	state, err := a.freshness.RecallFreshness(ctx, req)
	if err != nil {
		return Freshness{
			Freshness:       "degraded",
			Reasons:         []string{"recall_freshness_probe_unavailable"},
			AffectedSources: derivedRecallSources(),
		}
	}
	if strings.TrimSpace(state.Freshness) == "" {
		state.Freshness = "stored"
	}
	return state
}

func (a *Assembler) degradedReasons() []string {
	var reasons []string
	if a.notes == nil {
		reasons = append(reasons, "notes_unavailable")
	}
	if a.plans == nil {
		reasons = append(reasons, "plans_unavailable")
	}
	if a.memories == nil {
		reasons = append(reasons, "memories_unavailable")
	}
	if a.documents == nil {
		reasons = append(reasons, "documents_unavailable")
	}
	if a.profiles == nil {
		reasons = append(reasons, "profiles_unavailable")
	}
	if a.summaries == nil {
		reasons = append(reasons, "session_summaries_unavailable")
	}
	if a.groups == nil {
		reasons = append(reasons, "group_membership_unavailable")
	}
	return reasons
}

func validatePrefetchRequest(req *core.PrefetchRequest) error {
	required := map[string]string{
		"tenant_id":    req.TenantID,
		"workspace_id": req.WorkspaceID,
		"session_id":   req.SessionID,
		"actor_id":     req.ActorID,
	}
	for name, value := range required {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%w: %s is required", core.ErrInvalidArgument, name)
		}
	}
	return nil
}

func visibleScopes(groupIDs []string) []core.MemoryScope {
	scopes := baseVisibleScopes()
	if len(groupIDs) > 0 {
		scopes = append(scopes, core.MemoryScopeGroupShared)
	}
	return scopes
}

func baseVisibleScopes() []core.MemoryScope {
	return []core.MemoryScope{
		core.MemoryScopeAgentPrivate,
		core.MemoryScopeWorkspaceShared,
		core.MemoryScopeSessionScratch,
	}
}

func (a *Assembler) visibleGroupIDs(ctx context.Context, req *core.PrefetchRequest) ([]string, error) {
	if a.groups == nil {
		return []string{}, nil
	}
	memberships, err := a.groups.ListMembershipsForEntity(ctx, req.TenantID, req.WorkspaceID, req.ActorID)
	if errors.Is(err, core.ErrNotFound) {
		return []string{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("list visible memory groups: %w", err)
	}
	return membershipGroupIDs(memberships), nil
}

func membershipGroupIDs(memberships []*core.MemoryGroupMembership) []string {
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

func (a *Assembler) addPinnedNotes(ctx context.Context, req *core.PrefetchRequest, scopes []core.MemoryScope, candidates []candidateBlock) ([]candidateBlock, error) {
	if a.notes == nil {
		return candidates, nil
	}
	notes, err := a.notes.ListPinnedNotes(ctx, &core.ListPinnedNotesRequest{
		TenantID:      req.TenantID,
		WorkspaceID:   req.WorkspaceID,
		OwnerEntityID: req.ActorID,
		Scopes:        scopes,
	})
	if errors.Is(err, core.ErrNotFound) {
		return candidates, nil
	}
	if err != nil {
		return nil, fmt.Errorf("list pinned notes: %w", err)
	}
	now := a.clock().UTC()
	for _, note := range notes {
		if note == nil || strings.TrimSpace(note.Text) == "" {
			continue
		}
		if note.ExpiresAt != nil && !note.ExpiresAt.After(now) {
			continue
		}
		candidates = append(candidates, candidateBlock{
			block: core.RecallBlock{
				ID:            note.ID,
				Kind:          "pinned_note",
				Priority:      100,
				Text:          note.Text,
				Scope:         note.Scope,
				Source:        "notes",
				SourceID:      note.ID,
				Status:        "pinned",
				Freshness:     "stored",
				OwnerEntityID: note.OwnerEntityID,
			},
			source: "notes",
		})
	}
	return candidates, nil
}

func (a *Assembler) addActivePlans(ctx context.Context, req *core.PrefetchRequest, scopes []core.MemoryScope, candidates []candidateBlock) ([]candidateBlock, error) {
	if a.plans == nil {
		return candidates, nil
	}
	plans, err := a.plans.GetActivePlans(ctx, &core.GetActivePlansRequest{
		TenantID:      req.TenantID,
		WorkspaceID:   req.WorkspaceID,
		OwnerEntityID: req.ActorID,
		Scopes:        scopes,
	})
	if errors.Is(err, core.ErrNotFound) {
		return candidates, nil
	}
	if err != nil {
		return nil, fmt.Errorf("list active plans: %w", err)
	}
	for _, plan := range plans {
		if plan == nil || strings.TrimSpace(plan.Title) == "" || suppressedPlanStatus(plan.Status) {
			continue
		}
		candidates = append(candidates, candidateBlock{
			block: core.RecallBlock{
				ID:            plan.ID,
				Kind:          "active_plan",
				Priority:      95,
				Text:          plan.Title,
				Scope:         plan.Scope,
				Source:        "plans",
				SourceID:      plan.ID,
				Status:        plan.Status,
				Freshness:     "stored",
				OwnerEntityID: plan.OwnerEntityID,
			},
			source: "plans",
		})
	}
	return candidates, nil
}

func (a *Assembler) addProfiles(ctx context.Context, req *core.PrefetchRequest, candidates []candidateBlock) ([]candidateBlock, error) {
	if a.profiles == nil {
		return candidates, nil
	}
	profileTargets := []struct {
		entityID string
		scope    core.MemoryScope
	}{
		{entityID: req.ActorID, scope: core.MemoryScopeAgentPrivate},
		{entityID: "workspace:" + req.WorkspaceID, scope: core.MemoryScopeWorkspaceShared},
	}
	for _, target := range profileTargets {
		profile, err := a.profiles.GetProfile(ctx, target.entityID, target.scope)
		if errors.Is(err, core.ErrNotFound) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("get profile: %w", err)
		}
		if profile == nil {
			continue
		}
		if text := jsonText(profile.StaticJSON); text != "" {
			candidates = append(candidates, candidateBlock{
				block: core.RecallBlock{
					ID:        target.entityID,
					Kind:      "profile_static",
					Priority:  90,
					Text:      text,
					Scope:     target.scope,
					Source:    "profile",
					SourceID:  target.entityID,
					Status:    "snapshot",
					Freshness: "stored",
				},
				source: "profile",
			})
		}
		if text := jsonText(profile.DynamicJSON); text != "" {
			candidates = append(candidates, candidateBlock{
				block: core.RecallBlock{
					ID:        target.entityID,
					Kind:      "profile_dynamic",
					Priority:  85,
					Text:      text,
					Scope:     target.scope,
					Source:    "profile",
					SourceID:  target.entityID,
					Status:    "snapshot",
					Freshness: "stored",
				},
				source: "profile",
			})
		}
	}
	return candidates, nil
}

func (a *Assembler) addSessionSummary(ctx context.Context, req *core.PrefetchRequest, candidates []candidateBlock) ([]candidateBlock, error) {
	if a.summaries == nil {
		return candidates, nil
	}
	summary, err := a.summaries.GetSessionSummary(ctx, req.SessionID)
	if errors.Is(err, core.ErrNotFound) {
		return candidates, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get session summary: %w", err)
	}
	if summary != nil && strings.TrimSpace(summary.SummaryText) != "" {
		candidates = append(candidates, candidateBlock{
			block: core.RecallBlock{
				ID:        summary.ID,
				Kind:      "session_summary",
				Priority:  80,
				Text:      summary.SummaryText,
				Scope:     core.MemoryScopeSessionScratch,
				Source:    "session_summaries",
				SourceID:  summary.ID,
				Status:    "summary",
				Freshness: "stored",
			},
			source: "session_summaries",
		})
	}
	return candidates, nil
}

func (a *Assembler) addMemories(ctx context.Context, req *core.PrefetchRequest, scopes []core.MemoryScope, groupIDs []string, candidates []candidateBlock) ([]candidateBlock, error) {
	if a.memories == nil || strings.TrimSpace(req.Query) == "" {
		return candidates, nil
	}
	resp, err := a.memories.SearchMemories(ctx, &core.SearchMemoriesRequest{
		TenantID:        req.TenantID,
		WorkspaceID:     req.WorkspaceID,
		OwnerEntityID:   req.ActorID,
		VisibleGroupIDs: groupIDs,
		Query:           req.Query,
		Scopes:          scopes,
		ArtifactClasses: []core.ArtifactClass{
			core.ArtifactClassContext,
			core.ArtifactClassKnowledge,
			core.ArtifactClassTimeline,
			core.ArtifactClassPlan,
		},
	})
	if errors.Is(err, core.ErrNotFound) {
		return candidates, nil
	}
	if err != nil {
		return nil, fmt.Errorf("search memories: %w", err)
	}
	if resp == nil {
		return candidates, nil
	}
	for i, memory := range resp.Memories {
		if i >= 8 {
			break
		}
		if strings.TrimSpace(memory.Text) == "" || !memory.LatestFlag {
			continue
		}
		rank := scoreMemoryCandidate(req.Query, memory, a.clock().UTC())
		candidates = append(candidates, candidateBlock{
			block: core.RecallBlock{
				ID:            memory.MemoryID,
				Kind:          "memory",
				Priority:      70 + int(rank),
				Text:          memory.Text,
				Scope:         memory.Scope,
				Source:        "memories",
				SourceID:      memory.MemoryID,
				Status:        memoryRecallStatus(memory),
				Freshness:     "stored",
				OwnerEntityID: memory.OwnerEntityID,
			},
			source: "memories",
			rank:   rank,
		})
	}
	return candidates, nil
}

func (a *Assembler) addDocuments(ctx context.Context, req *core.PrefetchRequest, candidates []candidateBlock) ([]candidateBlock, error) {
	if a.documents == nil || strings.TrimSpace(req.Query) == "" {
		return candidates, nil
	}
	resp, err := a.documents.SearchDocuments(ctx, &core.SearchDocumentsRequest{
		TenantID:    req.TenantID,
		WorkspaceID: req.WorkspaceID,
		Query:       req.Query,
	})
	if errors.Is(err, core.ErrNotFound) {
		return candidates, nil
	}
	if err != nil {
		return nil, fmt.Errorf("search documents: %w", err)
	}
	if resp == nil {
		return candidates, nil
	}
	for i, chunk := range resp.Chunks {
		if i >= 5 {
			break
		}
		if strings.TrimSpace(chunk.Text) == "" {
			continue
		}
		rank := scoreDocumentCandidate(req.Query, chunk)
		candidates = append(candidates, candidateBlock{
			block: core.RecallBlock{
				ID:        chunk.ChunkID,
				Kind:      "document",
				Priority:  60 + int(rank),
				Text:      chunk.Text,
				Scope:     core.MemoryScopeWorkspaceShared,
				Source:    "documents",
				SourceID:  chunk.DocumentID,
				Status:    "supporting_context",
				Freshness: "stored",
			},
			source: "documents",
			rank:   rank,
		})
	}
	return candidates, nil
}

func packCandidates(candidates []candidateBlock, budget int) ([]core.RecallBlock, []string, int) {
	blocks := make([]core.RecallBlock, 0, len(candidates))
	sources := make([]string, 0, 8)
	seenSource := make(map[string]struct{})
	seenText := make(map[string]struct{})
	estimated := 0

	for _, candidate := range candidates {
		text := strings.TrimSpace(candidate.block.Text)
		if text == "" {
			continue
		}
		dedupKey := strings.ToLower(text)
		if _, ok := seenText[dedupKey]; ok {
			continue
		}
		remaining := budget - estimated
		if remaining <= 0 {
			break
		}
		blockBudget := minInt(remaining, maxBlockTokens(budget))
		candidate.block.Text = truncateToBudget(text, blockBudget)
		if candidate.block.Text == "" {
			continue
		}
		tokenCost := estimateTokens(candidate.block.Text)
		if estimated+tokenCost > budget {
			continue
		}
		blocks = append(blocks, candidate.block)
		seenText[dedupKey] = struct{}{}
		estimated += tokenCost
		if candidate.source != "" {
			if _, ok := seenSource[candidate.source]; !ok {
				sources = append(sources, candidate.source)
				seenSource[candidate.source] = struct{}{}
			}
		}
	}
	return blocks, sources, estimated
}

func maxBlockTokens(budget int) int {
	if budget <= 8 {
		return budget
	}
	limit := budget * 45 / 100
	if limit < 8 {
		return 8
	}
	return limit
}

func budgetFor(req *core.PrefetchRequest) int {
	if req.BudgetTokens > 0 {
		return req.BudgetTokens
	}
	switch strings.ToLower(strings.TrimSpace(req.Mode)) {
	case "small":
		return smallBudgetTokens
	case "rich":
		return richBudgetTokens
	default:
		return defaultBudgetTokens
	}
}

func estimateTokens(text string) int {
	words := strings.Fields(text)
	if len(words) == 0 {
		return 0
	}
	return len(words) + len(words)/3 + 1
}

func scoreMemoryCandidate(query string, memory core.MemoryResult, now time.Time) float64 {
	score := lexicalOverlapScore(query, memory.Text) * 20
	if memory.Confidence > 0 {
		score += memory.Confidence * 3
	}
	if !memory.ValidFrom.IsZero() {
		age := now.Sub(memory.ValidFrom)
		switch {
		case age < 0:
			score++
		case age <= 24*time.Hour:
			score += 3
		case age <= 7*24*time.Hour:
			score += 1.5
		}
	}
	switch memory.Kind {
	case core.MemoryKindConstraint, core.MemoryKindDecision, core.MemoryKindProcedure, core.MemoryKindTaskState:
		score++
	}
	return score
}

func scoreDocumentCandidate(query string, chunk core.DocumentChunkResult) float64 {
	score := lexicalOverlapScore(query, chunk.Text) * 12
	if chunk.Score > 0 {
		score += chunk.Score * 4
	}
	return score
}

func memoryRecallStatus(memory core.MemoryResult) string {
	if !memory.LatestFlag {
		return "suppressed"
	}
	return "active"
}

func lexicalOverlapScore(query, text string) float64 {
	queryTerms := recallTerms(query)
	if len(queryTerms) == 0 {
		return 0
	}
	textTerms := recallTerms(text)
	matches := 0
	for term := range queryTerms {
		if _, ok := textTerms[term]; ok {
			matches++
		}
	}
	return float64(matches) / float64(len(queryTerms))
}

func recallTerms(text string) map[string]struct{} {
	terms := make(map[string]struct{})
	for _, raw := range strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	}) {
		if len(raw) < 3 {
			continue
		}
		terms[raw] = struct{}{}
	}
	return terms
}

func minInt(left, right int) int {
	if left < right {
		return left
	}
	return right
}

func truncateToBudget(text string, budget int) string {
	if budget <= 0 {
		return ""
	}
	if estimateTokens(text) <= budget {
		return text
	}
	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}
	maxWords := (budget - 1) * 3 / 4
	if maxWords <= 0 {
		return ""
	}
	if maxWords > len(words) {
		maxWords = len(words)
	}
	return strings.Join(words[:maxWords], " ") + "..."
}

func jsonText(raw []byte) string {
	text := strings.TrimSpace(string(raw))
	if text == "" || text == "{}" || text == "null" {
		return ""
	}
	return text
}

func suppressedPlanStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "done", "completed", "archived", "deleted", "cancelled":
		return true
	default:
		return false
	}
}
