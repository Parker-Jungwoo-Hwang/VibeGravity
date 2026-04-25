// ============================================================
// FILE     : internal/store/postgres/profiles_summaries.go
// PURPOSE  : Implements PostgreSQL persistence for profiles and session summaries.
// LAYER    : infra
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : GetProfile, UpsertProfile, UpsertSessionSummary, GetSessionSummary
// DEPENDS  : internal/core, github.com/jackc/pgx/v5
// USED_BY  : recall assembler, graph and dreaming slices
// ------------------------------------------------------------
// AGENT_NOTE: Profiles and summaries must remain rebuildable from source artifacts.
// ============================================================

package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

// GetProfile loads a profile snapshot.
func (s *Store) GetProfile(ctx context.Context, entityID string, scope core.MemoryScope) (*core.Profile, error) {
	profile := &core.Profile{}
	err := s.pool.QueryRow(ctx, `
		SELECT entity_id, scope, static_json, dynamic_json, source_memory_ids, updated_at, version
		FROM profiles
		WHERE entity_id = $1 AND scope = $2
	`, entityID, scope).Scan(&profile.EntityID, &profile.Scope, &profile.StaticJSON,
		&profile.DynamicJSON, &profile.SourceMemoryIDs, &profile.UpdatedAt, &profile.Version)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, core.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get profile: %w", err)
	}
	return profile, nil
}

// UpsertProfile writes a profile snapshot.
func (s *Store) UpsertProfile(ctx context.Context, profile *core.Profile) error {
	if profile == nil {
		return fmt.Errorf("%w: profile is required", core.ErrInvalidArgument)
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO profiles (
			entity_id, scope, static_json, dynamic_json, source_memory_ids, updated_at, version
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (entity_id, scope) DO UPDATE
		SET static_json = EXCLUDED.static_json,
		    dynamic_json = EXCLUDED.dynamic_json,
		    source_memory_ids = EXCLUDED.source_memory_ids,
		    updated_at = EXCLUDED.updated_at,
		    version = profiles.version + 1
	`, profile.EntityID, profile.Scope, rawJSONOrEmpty(profile.StaticJSON), rawJSONOrEmpty(profile.DynamicJSON),
		profile.SourceMemoryIDs, timeOrNow(profile.UpdatedAt), profile.Version)
	if err != nil {
		return fmt.Errorf("upsert profile: %w", err)
	}
	return nil
}

// UpsertSessionSummary writes a session summary.
func (s *Store) UpsertSessionSummary(ctx context.Context, summary *core.SessionSummary) error {
	if summary == nil {
		return fmt.Errorf("%w: session summary is required", core.ErrInvalidArgument)
	}
	var err error
	if summary.ID == "" {
		summary.ID, err = newID("sum")
		if err != nil {
			return err
		}
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO session_summaries (
			id, tenant_id, workspace_id, session_id, summary_text,
			source_event_ids, source_memory_ids, created_at, updated_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (id) DO UPDATE
		SET summary_text = EXCLUDED.summary_text,
		    source_event_ids = EXCLUDED.source_event_ids,
		    source_memory_ids = EXCLUDED.source_memory_ids,
		    updated_at = EXCLUDED.updated_at
	`, summary.ID, summary.TenantID, summary.WorkspaceID, summary.SessionID, summary.SummaryText,
		summary.SourceEventIDs, summary.SourceMemoryIDs, timeOrNow(summary.CreatedAt), timeOrNow(summary.UpdatedAt))
	if err != nil {
		return fmt.Errorf("upsert session summary: %w", err)
	}
	return nil
}

// GetSessionSummary loads the current summary for a session.
func (s *Store) GetSessionSummary(ctx context.Context, sessionID string) (*core.SessionSummary, error) {
	summary := &core.SessionSummary{}
	err := s.pool.QueryRow(ctx, `
		SELECT id, tenant_id, workspace_id, session_id, summary_text,
		       source_event_ids, source_memory_ids, created_at, updated_at
		FROM session_summaries
		WHERE session_id = $1
		ORDER BY updated_at DESC
		LIMIT 1
	`, sessionID).Scan(&summary.ID, &summary.TenantID, &summary.WorkspaceID,
		&summary.SessionID, &summary.SummaryText, &summary.SourceEventIDs,
		&summary.SourceMemoryIDs, &summary.CreatedAt, &summary.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, core.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get session summary: %w", err)
	}
	return summary, nil
}
