// ============================================================
// FILE     : internal/store/postgres/corrections.go
// PURPOSE  : Implements PostgreSQL persistence for human memory corrections.
// LAYER    : infra
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : RecordMemoryCorrection
// DEPENDS  : internal/core, github.com/jackc/pgx/v5
// USED_BY  : kernel CorrectMemory path
// ------------------------------------------------------------
// AGENT_NOTE: Corrections are append-safe operator artifacts; do not mutate memory_trace or latest_flag here.
// ============================================================

package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

// RecordMemoryCorrection writes a raw correction event and operator-visible artifact idempotently.
func (s *Store) RecordMemoryCorrection(ctx context.Context, event *core.RawEvent, correction *core.MemoryCorrection) (*core.MemoryCorrection, error) {
	if event == nil {
		return nil, fmt.Errorf("%w: raw correction event is required", core.ErrInvalidArgument)
	}
	if correction == nil {
		return nil, fmt.Errorf("%w: memory correction is required", core.ErrInvalidArgument)
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin memory correction transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	rawEventID, err := insertCorrectionRawEvent(ctx, tx, event)
	if err != nil {
		return nil, err
	}
	correction.RawEventID = rawEventID
	recorded, err := insertMemoryCorrection(ctx, tx, correction)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit memory correction transaction: %w", err)
	}
	return recorded, nil
}

func insertCorrectionRawEvent(ctx context.Context, tx pgx.Tx, event *core.RawEvent) (string, error) {
	id := event.ID
	var err error
	if id == "" {
		id, err = newID("evt")
		if err != nil {
			return "", err
		}
	}
	var rawEventID string
	err = tx.QueryRow(ctx, `
		INSERT INTO raw_events (
			id, tenant_id, workspace_id, session_id, actor_id, event_kind,
			source, idempotency_key, fingerprint, occurred_at, payload_json, created_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		ON CONFLICT (tenant_id, source, idempotency_key) DO UPDATE
		SET idempotency_key = EXCLUDED.idempotency_key
		WHERE raw_events.workspace_id = EXCLUDED.workspace_id
		RETURNING id
	`, id, event.TenantID, event.WorkspaceID, event.SessionID, event.ActorID,
		event.EventKind, event.Source, event.IdempotencyKey, event.Fingerprint,
		timeOrNow(event.OccurredAt), rawJSONOrEmpty(event.PayloadJSON), timeOrNow(event.CreatedAt)).Scan(&rawEventID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", fmt.Errorf("%w: correction idempotency key belongs to a different workspace", core.ErrConflict)
	}
	if err != nil {
		return "", fmt.Errorf("insert correction raw event: %w", err)
	}
	return rawEventID, nil
}

func insertMemoryCorrection(ctx context.Context, tx pgx.Tx, correction *core.MemoryCorrection) (*core.MemoryCorrection, error) {
	id := correction.ID
	var err error
	if id == "" {
		id, err = newID("corr")
		if err != nil {
			return nil, err
		}
	}
	recorded := &core.MemoryCorrection{}
	err = tx.QueryRow(ctx, `
		INSERT INTO memory_corrections (
			id, tenant_id, workspace_id, memory_id, operator_id, raw_event_id,
			idempotency_key, correction_text, evidence_json, status, created_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (tenant_id, workspace_id, idempotency_key) DO UPDATE
		SET idempotency_key = EXCLUDED.idempotency_key
		WHERE memory_corrections.memory_id = EXCLUDED.memory_id
		  AND memory_corrections.operator_id = EXCLUDED.operator_id
		  AND memory_corrections.raw_event_id = EXCLUDED.raw_event_id
		  AND memory_corrections.correction_text = EXCLUDED.correction_text
		  AND memory_corrections.evidence_json = EXCLUDED.evidence_json
		RETURNING id, tenant_id, workspace_id, memory_id, operator_id, raw_event_id,
		          idempotency_key, correction_text, evidence_json, status, created_at
	`, id, correction.TenantID, correction.WorkspaceID, correction.MemoryID, correction.OperatorID,
		correction.RawEventID, correction.IdempotencyKey, correction.CorrectionText,
		rawJSONOrEmpty(correction.EvidenceJSON), valueOr(correction.Status, "recorded"),
		timeOrNow(correction.CreatedAt)).Scan(&recorded.ID, &recorded.TenantID,
		&recorded.WorkspaceID, &recorded.MemoryID, &recorded.OperatorID,
		&recorded.RawEventID, &recorded.IdempotencyKey, &recorded.CorrectionText,
		&recorded.EvidenceJSON, &recorded.Status, &recorded.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("%w: correction idempotency key belongs to different evidence", core.ErrConflict)
	}
	if err != nil {
		return nil, fmt.Errorf("insert memory correction: %w", err)
	}
	return recorded, nil
}

// GetMemoryCorrectionByIdempotency loads an existing correction artifact for replay validation.
func (s *Store) GetMemoryCorrectionByIdempotency(ctx context.Context, tenantID string, workspaceID string, idempotencyKey string) (*core.MemoryCorrection, error) {
	recorded := &core.MemoryCorrection{}
	err := s.pool.QueryRow(ctx, `
		SELECT id, tenant_id, workspace_id, memory_id, operator_id, raw_event_id,
		       idempotency_key, correction_text, evidence_json, status, created_at
		FROM memory_corrections
		WHERE tenant_id = $1
		  AND workspace_id = $2
		  AND idempotency_key = $3
	`, tenantID, workspaceID, idempotencyKey).Scan(&recorded.ID, &recorded.TenantID,
		&recorded.WorkspaceID, &recorded.MemoryID, &recorded.OperatorID,
		&recorded.RawEventID, &recorded.IdempotencyKey, &recorded.CorrectionText,
		&recorded.EvidenceJSON, &recorded.Status, &recorded.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, core.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get memory correction by idempotency: %w", err)
	}
	return recorded, nil
}
