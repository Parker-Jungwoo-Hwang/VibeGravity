// ============================================================
// FILE     : internal/store/postgres/timeline.go
// PURPOSE  : Implements PostgreSQL read paths for operator timeline views.
// LAYER    : infra
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : GetTimeline
// DEPENDS  : internal/core
// USED_BY  : kernel GetTimeline path
// ------------------------------------------------------------
// AGENT_NOTE: Timeline is read-only here; do not mutate graph, trace, or recall state.
// ============================================================

package postgres

import (
	"context"
	"fmt"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

// GetTimeline loads a read-only timeline view over memories and correction artifacts.
func (s *Store) GetTimeline(ctx context.Context, req *core.GetTimelineRequest) (*core.GetTimelineResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("%w: get timeline request is required", core.ErrInvalidArgument)
	}
	sql, args := timelineStatement(req)
	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query timeline: %w", err)
	}
	defer rows.Close()

	items := make([]core.TimelineItem, 0, req.Limit)
	for rows.Next() {
		var item core.TimelineItem
		if err := rows.Scan(&item.ID, &item.Kind, &item.ArtifactClass, &item.Text,
			&item.OccurredAt, &item.MemoryID, &item.RawEventID); err != nil {
			return nil, fmt.Errorf("scan timeline item: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate timeline items: %w", err)
	}
	return &core.GetTimelineResponse{Items: items}, nil
}

func timelineStatement(req *core.GetTimelineRequest) (string, []any) {
	return `
		WITH memory_items AS (
			SELECT
				m.id AS id,
				m.kind AS kind,
				m.artifact_class AS artifact_class,
				m.text AS text,
				COALESCE(mt.created_at, m.valid_from, m.created_at) AS occurred_at,
				m.id AS memory_id,
				'' AS raw_event_id
			FROM memories m
			LEFT JOIN memory_trace mt ON mt.memory_id = m.id
			WHERE m.tenant_id = $1
			  AND m.workspace_id = $2
			  AND m.scope = ANY($3)
			  AND m.scope <> 'group_shared'
			  AND m.status <> 'deleted'
			  AND ($4::timestamptz IS NULL OR COALESCE(mt.created_at, m.valid_from, m.created_at) >= $4)
			  AND ($5::timestamptz IS NULL OR COALESCE(mt.created_at, m.valid_from, m.created_at) <= $5)
			  AND (
			    m.scope <> 'agent_private'
			    OR ($6 <> '' AND m.owner_entity_id = $6)
			  )
		),
		correction_items AS (
			SELECT
				c.id AS id,
				'correction' AS kind,
				'timeline' AS artifact_class,
				'Correction for memory ' || c.memory_id || ': ' || c.correction_text AS text,
				c.created_at AS occurred_at,
				c.memory_id AS memory_id,
				c.raw_event_id AS raw_event_id
			FROM memory_corrections c
			JOIN memories m ON m.id = c.memory_id
			WHERE c.tenant_id = $1
			  AND c.workspace_id = $2
			  AND m.tenant_id = $1
			  AND m.workspace_id = $2
			  AND m.scope = ANY($3)
			  AND m.scope <> 'group_shared'
			  AND ($4::timestamptz IS NULL OR c.created_at >= $4)
			  AND ($5::timestamptz IS NULL OR c.created_at <= $5)
			  AND (
			    m.scope <> 'agent_private'
			    OR ($6 <> '' AND m.owner_entity_id = $6)
			  )
		)
		SELECT id, kind, artifact_class, text, occurred_at, memory_id, raw_event_id
		FROM memory_items
		UNION ALL
		SELECT id, kind, artifact_class, text, occurred_at, memory_id, raw_event_id
		FROM correction_items
		ORDER BY occurred_at DESC, id DESC
		LIMIT $7
	`, []any{req.TenantID, req.WorkspaceID, req.Scopes, req.From, req.To, req.EntityID, req.Limit}
}
