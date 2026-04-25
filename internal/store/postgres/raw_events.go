// ============================================================
// FILE     : internal/store/postgres/raw_events.go
// PURPOSE  : Implements PostgreSQL persistence for immutable raw events.
// LAYER    : infra
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : AppendRawEvents, GetRawEvents
// DEPENDS  : internal/core, github.com/jackc/pgx/v5
// USED_BY  : internal/ingest, worker pipeline
// ------------------------------------------------------------
// AGENT_NOTE: Preserve raw_events as immutable source records and use idempotency for duplicates.
// ============================================================

package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

// AppendRawEvents inserts raw events idempotently and returns newly inserted IDs.
func (s *Store) AppendRawEvents(ctx context.Context, events []*core.RawEvent) ([]string, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin raw event insert: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	ids := make([]string, 0, len(events))
	for _, event := range events {
		if event == nil {
			continue
		}
		id := event.ID
		if id == "" {
			id, err = newID("evt")
			if err != nil {
				return nil, err
			}
		}
		row := tx.QueryRow(ctx, `
			INSERT INTO raw_events (
				id, tenant_id, workspace_id, session_id, actor_id, event_kind,
				source, idempotency_key, fingerprint, occurred_at, payload_json, created_at
			)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
			ON CONFLICT (tenant_id, source, idempotency_key) DO NOTHING
			RETURNING id
		`, id, event.TenantID, event.WorkspaceID, event.SessionID, event.ActorID,
			event.EventKind, event.Source, event.IdempotencyKey, event.Fingerprint,
			timeOrNow(event.OccurredAt), rawJSONOrEmpty(event.PayloadJSON), timeOrNow(event.CreatedAt))
		var insertedID string
		if err := row.Scan(&insertedID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				continue
			}
			return nil, fmt.Errorf("insert raw event: %w", err)
		}
		ids = append(ids, insertedID)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit raw event insert: %w", err)
	}
	return ids, nil
}

// GetRawEvents loads raw events by ID.
func (s *Store) GetRawEvents(ctx context.Context, ids []string) ([]*core.RawEvent, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, tenant_id, workspace_id, session_id, actor_id, event_kind,
		       source, idempotency_key, fingerprint, occurred_at, payload_json, created_at
		FROM raw_events
		WHERE id = ANY($1)
		ORDER BY occurred_at ASC, created_at ASC
	`, ids)
	if err != nil {
		return nil, fmt.Errorf("query raw events: %w", err)
	}
	defer rows.Close()

	events := make([]*core.RawEvent, 0, len(ids))
	for rows.Next() {
		event := &core.RawEvent{}
		if err := rows.Scan(&event.ID, &event.TenantID, &event.WorkspaceID, &event.SessionID,
			&event.ActorID, &event.EventKind, &event.Source, &event.IdempotencyKey,
			&event.Fingerprint, &event.OccurredAt, &event.PayloadJSON, &event.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan raw event: %w", err)
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate raw events: %w", err)
	}
	return events, nil
}
