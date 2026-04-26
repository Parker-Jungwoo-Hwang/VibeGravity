// ============================================================
// FILE     : internal/timeline/service.go
// PURPOSE  : Validates and delegates read-only timeline requests.
// LAYER    : application
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : Service, NewService
// DEPENDS  : internal/core, internal/store
// USED_BY  : internal/kernel, internal/timeline tests
// ------------------------------------------------------------
// AGENT_NOTE: Keep group_shared excluded until membership-aware timeline filtering exists.
// ============================================================

package timeline

import (
	"context"
	"fmt"
	"strings"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/store"
)

const (
	timelineDefaultLimit = 50
	timelineMaxLimit     = 100
)

// Service owns timeline use cases.
type Service struct {
	timeline store.TimelineStore
}

// NewService builds a timeline service.
func NewService(timeline store.TimelineStore) *Service {
	return &Service{timeline: timeline}
}

// GetTimeline assembles a read-only operator timeline over existing artifacts.
func (s *Service) GetTimeline(ctx context.Context, req *core.GetTimelineRequest) (*core.GetTimelineResponse, error) {
	if s == nil || s.timeline == nil {
		return nil, fmt.Errorf("%w: get timeline", core.ErrNotImplemented)
	}
	if req == nil {
		return nil, fmt.Errorf("%w: get timeline request is required", core.ErrInvalidArgument)
	}
	if err := requireFields(map[string]string{
		"tenant_id":    req.TenantID,
		"workspace_id": req.WorkspaceID,
		"entity_id":    req.EntityID,
	}); err != nil {
		return nil, err
	}
	if req.From != nil && req.To != nil && req.From.After(*req.To) {
		return nil, fmt.Errorf("%w: from must be before to", core.ErrInvalidArgument)
	}
	normalized := *req
	scopes, err := normalizeTimelineScopes(req.Scopes)
	if err != nil {
		return nil, err
	}
	normalized.Scopes = scopes
	if normalized.Limit == 0 {
		normalized.Limit = timelineDefaultLimit
	}
	if normalized.Limit < 0 || normalized.Limit > timelineMaxLimit {
		return nil, fmt.Errorf("%w: limit must be between 1 and %d", core.ErrInvalidArgument, timelineMaxLimit)
	}
	return s.timeline.GetTimeline(ctx, &normalized)
}

func requireFields(fields map[string]string) error {
	for name, value := range fields {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%w: %s is required", core.ErrInvalidArgument, name)
		}
	}
	return nil
}

func normalizeTimelineScopes(scopes []core.MemoryScope) ([]core.MemoryScope, error) {
	if len(scopes) == 0 {
		return []core.MemoryScope{
			core.MemoryScopeAgentPrivate,
			core.MemoryScopeWorkspaceShared,
			core.MemoryScopeSessionScratch,
		}, nil
	}
	seen := make(map[core.MemoryScope]struct{}, len(scopes))
	normalized := make([]core.MemoryScope, 0, len(scopes))
	for _, scope := range scopes {
		switch scope {
		case core.MemoryScopeAgentPrivate, core.MemoryScopeWorkspaceShared, core.MemoryScopeSessionScratch:
			if _, ok := seen[scope]; !ok {
				seen[scope] = struct{}{}
				normalized = append(normalized, scope)
			}
		case core.MemoryScopeGroupShared:
			continue
		default:
			return nil, fmt.Errorf("%w: unsupported timeline scope %q", core.ErrInvalidArgument, scope)
		}
	}
	return normalized, nil
}
