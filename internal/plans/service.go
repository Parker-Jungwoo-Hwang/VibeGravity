// ============================================================
// FILE     : internal/plans/service.go
// PURPOSE  : Implements structured plan create and update use cases.
// LAYER    : application
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : Service, NewService
// DEPENDS  : encoding/json, internal/core, internal/store
// USED_BY  : internal/kernel, internal/plans tests
// ------------------------------------------------------------
// AGENT_NOTE: Preserve patch semantics; nil Items means do not replace plan items.
// ============================================================

package plans

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/store"
)

// Service owns plan use cases.
type Service struct {
	plans store.PlanStore
	clock func() time.Time
}

// NewService builds a plan service.
func NewService(plans store.PlanStore) *Service {
	return &Service{plans: plans, clock: time.Now}
}

// CreatePlan creates a structured plan and its initial items.
func (s *Service) CreatePlan(ctx context.Context, req *core.CreatePlanRequest) (*core.CreatePlanResponse, error) {
	if s == nil || s.plans == nil {
		return nil, fmt.Errorf("%w: create plan", core.ErrNotImplemented)
	}
	if req == nil {
		return nil, fmt.Errorf("%w: create plan request is required", core.ErrInvalidArgument)
	}
	if err := requireFields(map[string]string{
		"tenant_id":       req.TenantID,
		"workspace_id":    req.WorkspaceID,
		"title":           req.Title,
		"scope":           string(req.Scope),
		"owner_entity_id": req.OwnerEntityID,
	}); err != nil {
		return nil, err
	}
	now := s.clock().UTC()
	plan := &core.Plan{
		TenantID:      req.TenantID,
		WorkspaceID:   req.WorkspaceID,
		Title:         req.Title,
		Status:        valueOr(req.Status, "active"),
		Scope:         req.Scope,
		OwnerEntityID: req.OwnerEntityID,
		EvidenceJSON:  jsonOrEmpty(req.EvidenceJSON),
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	items := make([]*core.PlanItem, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, &core.PlanItem{
			ID:           item.ID,
			Title:        item.Title,
			Status:       valueOr(item.Status, "open"),
			EvidenceJSON: jsonOrEmpty(item.EvidenceJSON),
			CreatedAt:    now,
			UpdatedAt:    now,
		})
	}
	if err := s.plans.CreatePlan(ctx, plan, items); err != nil {
		return nil, err
	}
	itemIDs := make([]string, 0, len(items))
	for _, item := range items {
		itemIDs = append(itemIDs, item.ID)
	}
	return &core.CreatePlanResponse{PlanID: plan.ID, ItemIDs: itemIDs, Status: "created"}, nil
}

// UpdatePlan updates a structured plan and optionally replaces provided items.
func (s *Service) UpdatePlan(ctx context.Context, req *core.UpdatePlanRequest) (*core.UpdatePlanResponse, error) {
	if s == nil || s.plans == nil {
		return nil, fmt.Errorf("%w: update plan", core.ErrNotImplemented)
	}
	if req == nil {
		return nil, fmt.Errorf("%w: update plan request is required", core.ErrInvalidArgument)
	}
	if err := requireFields(map[string]string{
		"tenant_id":    req.TenantID,
		"workspace_id": req.WorkspaceID,
		"plan_id":      req.PlanID,
	}); err != nil {
		return nil, err
	}
	now := s.clock().UTC()
	plan := &core.Plan{
		ID:           req.PlanID,
		TenantID:     req.TenantID,
		WorkspaceID:  req.WorkspaceID,
		EvidenceJSON: req.EvidenceJSON,
		UpdatedAt:    now,
	}
	if req.Title != nil {
		plan.Title = strings.TrimSpace(*req.Title)
		if plan.Title == "" {
			return nil, fmt.Errorf("%w: title cannot be empty", core.ErrInvalidArgument)
		}
	}
	if req.Status != nil {
		plan.Status = strings.TrimSpace(*req.Status)
		if plan.Status == "" {
			return nil, fmt.Errorf("%w: status cannot be empty", core.ErrInvalidArgument)
		}
	}
	items := make([]*core.PlanItem, 0, len(req.Items))
	if req.Items != nil {
		for _, item := range req.Items {
			title := strings.TrimSpace(item.Title)
			if title == "" {
				return nil, fmt.Errorf("%w: plan item title is required", core.ErrInvalidArgument)
			}
			items = append(items, &core.PlanItem{
				ID:           item.ID,
				Title:        title,
				Status:       valueOr(item.Status, "open"),
				EvidenceJSON: jsonOrEmpty(item.EvidenceJSON),
				CreatedAt:    now,
				UpdatedAt:    now,
			})
		}
	} else {
		items = nil
	}
	if err := s.plans.UpdatePlan(ctx, plan, items); err != nil {
		return nil, err
	}
	return &core.UpdatePlanResponse{PlanID: req.PlanID, Status: "updated"}, nil
}

func requireFields(fields map[string]string) error {
	for name, value := range fields {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%w: %s is required", core.ErrInvalidArgument, name)
		}
	}
	return nil
}

func jsonOrEmpty(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 {
		return json.RawMessage(`{}`)
	}
	return raw
}

func valueOr(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
