// ============================================================
// FILE     : internal/store/postgres/groups.go
// PURPOSE  : Implements PostgreSQL persistence for memory groups and memberships.
// LAYER    : infra
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : CreateMemoryGroup, AddMembership, ListMemberships, ListMembershipsForEntity
// DEPENDS  : context, fmt, internal/core
// USED_BY  : recall assembler, Stage 2 source adapters, future group APIs
// ------------------------------------------------------------
// AGENT_NOTE: Group visibility must be membership-backed before group_shared memory is returned.
// ============================================================

package postgres

import (
	"context"
	"fmt"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

// CreateMemoryGroup writes a memory group.
func (s *Store) CreateMemoryGroup(ctx context.Context, group *core.MemoryGroup) error {
	if group == nil {
		return fmt.Errorf("%w: memory group is required", core.ErrInvalidArgument)
	}
	if group.ID == "" {
		id, err := newID("group")
		if err != nil {
			return err
		}
		group.ID = id
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO memory_groups (id, tenant_id, workspace_id, name, description, created_at)
		VALUES ($1,$2,$3,$4,$5,$6)
	`, group.ID, group.TenantID, group.WorkspaceID, group.Name, group.Description, timeOrNow(group.CreatedAt))
	if err != nil {
		return fmt.Errorf("insert memory group: %w", err)
	}
	return nil
}

// AddMembership adds an entity to a memory group.
func (s *Store) AddMembership(ctx context.Context, membership *core.MemoryGroupMembership) error {
	if membership == nil {
		return fmt.Errorf("%w: memory group membership is required", core.ErrInvalidArgument)
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO memory_group_memberships (group_id, entity_id, created_at)
		VALUES ($1,$2,$3)
		ON CONFLICT (group_id, entity_id) DO NOTHING
	`, membership.GroupID, membership.EntityID, timeOrNow(membership.CreatedAt))
	if err != nil {
		return fmt.Errorf("insert memory group membership: %w", err)
	}
	return nil
}

// ListMemberships loads memberships for a memory group.
func (s *Store) ListMemberships(ctx context.Context, groupID string) ([]*core.MemoryGroupMembership, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT group_id, entity_id, created_at
		FROM memory_group_memberships
		WHERE group_id = $1
		ORDER BY created_at ASC
	`, groupID)
	if err != nil {
		return nil, fmt.Errorf("query memory group memberships: %w", err)
	}
	defer rows.Close()
	return scanMemoryGroupMemberships(rows)
}

// ListMembershipsForEntity loads groups visible to an entity in one workspace.
func (s *Store) ListMembershipsForEntity(ctx context.Context, tenantID string, workspaceID string, entityID string) ([]*core.MemoryGroupMembership, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT gm.group_id, gm.entity_id, gm.created_at
		FROM memory_group_memberships gm
		JOIN memory_groups g ON g.id = gm.group_id
		WHERE g.tenant_id = $1
		  AND g.workspace_id = $2
		  AND gm.entity_id = $3
		ORDER BY gm.created_at ASC
	`, tenantID, workspaceID, entityID)
	if err != nil {
		return nil, fmt.Errorf("query entity memory group memberships: %w", err)
	}
	defer rows.Close()
	return scanMemoryGroupMemberships(rows)
}

type memoryGroupMembershipRows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}

func scanMemoryGroupMemberships(rows memoryGroupMembershipRows) ([]*core.MemoryGroupMembership, error) {
	memberships := make([]*core.MemoryGroupMembership, 0, 8)
	for rows.Next() {
		membership := &core.MemoryGroupMembership{}
		if err := rows.Scan(&membership.GroupID, &membership.EntityID, &membership.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan memory group membership: %w", err)
		}
		memberships = append(memberships, membership)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate memory group memberships: %w", err)
	}
	return memberships, nil
}
