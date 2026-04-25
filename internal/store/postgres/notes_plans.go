// ============================================================
// FILE     : internal/store/postgres/notes_plans.go
// PURPOSE  : Implements PostgreSQL persistence for notes and structured plans.
// LAYER    : infra
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : AddNote, ListPinnedNotes, CreatePlan, UpdatePlan, GetActivePlans
// DEPENDS  : internal/core
// USED_BY  : recall assembler, future notes and plans APIs
// ------------------------------------------------------------
// AGENT_NOTE: Notes and plans are operator-intent artifacts; keep them separate from memories.
// ============================================================

package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

// AddNote writes a note.
func (s *Store) AddNote(ctx context.Context, note *core.Note) error {
	if note == nil {
		return fmt.Errorf("%w: note is required", core.ErrInvalidArgument)
	}
	id := note.ID
	var err error
	if id == "" {
		id, err = newID("note")
		if err != nil {
			return err
		}
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO notes (
			id, tenant_id, workspace_id, note_kind, scope, owner_entity_id,
			text, pinned, expires_at, created_at, updated_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
	`, id, note.TenantID, note.WorkspaceID, note.NoteKind, note.Scope, note.OwnerEntityID,
		note.Text, note.Pinned, note.ExpiresAt, timeOrNow(note.CreatedAt), timeOrNow(note.UpdatedAt))
	if err != nil {
		return fmt.Errorf("insert note: %w", err)
	}
	note.ID = id
	return nil
}

// ListPinnedNotes loads pinned notes for recall.
func (s *Store) ListPinnedNotes(ctx context.Context, req *core.ListPinnedNotesRequest) ([]*core.Note, error) {
	if req == nil {
		return nil, fmt.Errorf("%w: list pinned notes request is required", core.ErrInvalidArgument)
	}
	sql, args := listPinnedNotesStatement(req)
	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query pinned notes: %w", err)
	}
	defer rows.Close()

	notes := make([]*core.Note, 0, 20)
	for rows.Next() {
		note := &core.Note{}
		if err := rows.Scan(&note.ID, &note.TenantID, &note.WorkspaceID, &note.NoteKind,
			&note.Scope, &note.OwnerEntityID, &note.Text, &note.Pinned, &note.ExpiresAt,
			&note.CreatedAt, &note.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan note: %w", err)
		}
		notes = append(notes, note)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate pinned notes: %w", err)
	}
	return notes, nil
}

func listPinnedNotesStatement(req *core.ListPinnedNotesRequest) (string, []any) {
	return `
		SELECT id, tenant_id, workspace_id, note_kind, scope, owner_entity_id,
		       text, pinned, expires_at, created_at, updated_at
		FROM notes
		WHERE tenant_id = $1
		  AND workspace_id = $2
		  AND scope = ANY($3)
		  AND pinned = true
		  AND (expires_at IS NULL OR expires_at > now())
		  AND (
		    scope <> 'agent_private'
		    OR ($4 <> '' AND owner_entity_id = $4)
		  )
		ORDER BY updated_at DESC, created_at DESC
		LIMIT 20
	`, []any{req.TenantID, req.WorkspaceID, req.Scopes, req.OwnerEntityID}
}

// CreatePlan writes a plan and its initial items.
func (s *Store) CreatePlan(ctx context.Context, plan *core.Plan, items []*core.PlanItem) error {
	if plan == nil {
		return fmt.Errorf("%w: plan is required", core.ErrInvalidArgument)
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin create plan: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if plan.ID == "" {
		plan.ID, err = newID("plan")
		if err != nil {
			return err
		}
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO plans (
			id, tenant_id, workspace_id, title, status, scope,
			owner_entity_id, evidence_json, created_at, updated_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
	`, plan.ID, plan.TenantID, plan.WorkspaceID, plan.Title, plan.Status, plan.Scope,
		plan.OwnerEntityID, rawJSONOrEmpty(plan.EvidenceJSON), timeOrNow(plan.CreatedAt), timeOrNow(plan.UpdatedAt))
	if err != nil {
		return fmt.Errorf("insert plan: %w", err)
	}
	if err := insertPlanItems(ctx, tx, plan.ID, items); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit create plan: %w", err)
	}
	return nil
}

// UpdatePlan updates a plan and replaces its provided items when items is non-nil.
func (s *Store) UpdatePlan(ctx context.Context, plan *core.Plan, items []*core.PlanItem) error {
	if plan == nil || plan.ID == "" {
		return fmt.Errorf("%w: plan id is required", core.ErrInvalidArgument)
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin update plan: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	tag, err := tx.Exec(ctx, `
		UPDATE plans
		SET title = COALESCE(NULLIF($2, ''), title),
		    status = COALESCE(NULLIF($3, ''), status),
		    evidence_json = COALESCE($4, evidence_json),
		    updated_at = now()
		WHERE id = $1
		  AND tenant_id = $5
		  AND workspace_id = $6
	`, plan.ID, plan.Title, plan.Status, rawJSONOrNil(plan.EvidenceJSON), plan.TenantID, plan.WorkspaceID)
	if err != nil {
		return fmt.Errorf("update plan: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return core.ErrNotFound
	}
	if items != nil {
		if _, err := tx.Exec(ctx, `DELETE FROM plan_items WHERE plan_id = $1`, plan.ID); err != nil {
			return fmt.Errorf("delete plan items: %w", err)
		}
		if err := insertPlanItems(ctx, tx, plan.ID, items); err != nil {
			return err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit update plan: %w", err)
	}
	return nil
}

// GetActivePlans loads active plans for recall.
func (s *Store) GetActivePlans(ctx context.Context, req *core.GetActivePlansRequest) ([]*core.Plan, error) {
	if req == nil {
		return nil, fmt.Errorf("%w: get active plans request is required", core.ErrInvalidArgument)
	}
	sql, args := getActivePlansStatement(req)
	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query active plans: %w", err)
	}
	defer rows.Close()

	plans := make([]*core.Plan, 0, 20)
	for rows.Next() {
		plan := &core.Plan{}
		if err := rows.Scan(&plan.ID, &plan.TenantID, &plan.WorkspaceID, &plan.Title,
			&plan.Status, &plan.Scope, &plan.OwnerEntityID, &plan.EvidenceJSON,
			&plan.CreatedAt, &plan.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan plan: %w", err)
		}
		plans = append(plans, plan)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate active plans: %w", err)
	}
	return plans, nil
}

func getActivePlansStatement(req *core.GetActivePlansRequest) (string, []any) {
	return `
		SELECT id, tenant_id, workspace_id, title, status, scope,
		       owner_entity_id, evidence_json, created_at, updated_at
		FROM plans
		WHERE tenant_id = $1
		  AND workspace_id = $2
		  AND scope = ANY($3)
		  AND lower(status) NOT IN ('done', 'completed', 'archived', 'deleted', 'cancelled')
		  AND (
		    scope <> 'agent_private'
		    OR ($4 <> '' AND owner_entity_id = $4)
		  )
		ORDER BY updated_at DESC, created_at DESC
		LIMIT 20
	`, []any{req.TenantID, req.WorkspaceID, req.Scopes, req.OwnerEntityID}
}

type planItemExecutor interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

func insertPlanItems(ctx context.Context, tx planItemExecutor, planID string, items []*core.PlanItem) error {
	for _, item := range items {
		if item == nil {
			continue
		}
		var err error
		if item.ID == "" {
			item.ID, err = newID("plan_item")
			if err != nil {
				return err
			}
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO plan_items (
				id, plan_id, title, status, evidence_json, created_at, updated_at
			)
			VALUES ($1,$2,$3,$4,$5,$6,$7)
		`, item.ID, planID, item.Title, item.Status, rawJSONOrEmpty(item.EvidenceJSON),
			timeOrNow(item.CreatedAt), timeOrNow(item.UpdatedAt))
		if err != nil {
			return fmt.Errorf("insert plan item: %w", err)
		}
	}
	return nil
}
