// ============================================================
// FILE     : internal/store/postgres/dreaming.go
// PURPOSE  : Implements PostgreSQL queries for background dreaming consolidation and tier promotion.
// LAYER    : infra
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : LoadDreamingSessionInput, PromoteMemories
// DEPENDS  : context, fmt, strings, time, internal/core, github.com/jackc/pgx/v5
// USED_BY  : graph dreaming service, worker
// ------------------------------------------------------------
// AGENT_NOTE: Dreaming metadata updates must never change memory scope, owner, or provenance.
// ============================================================

package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

const dreamingPromotionLimit = 100

// LoadDreamingSessionInput loads raw event IDs and derived memories for one session.
func (s *Store) LoadDreamingSessionInput(ctx context.Context, req *core.DreamSessionRequest) (*core.DreamingSessionInput, error) {
	if req == nil {
		return nil, fmt.Errorf("%w: dream_session request is required", core.ErrInvalidArgument)
	}
	if strings.TrimSpace(req.TenantID) == "" || strings.TrimSpace(req.WorkspaceID) == "" || strings.TrimSpace(req.SessionID) == "" {
		return nil, fmt.Errorf("%w: dream_session tenant_id, workspace_id, and session_id are required", core.ErrInvalidArgument)
	}

	rawEventIDs, err := s.dreamingRawEventIDs(ctx, req)
	if err != nil {
		return nil, err
	}
	memories, err := s.dreamingMemoriesForRawEvents(ctx, req, rawEventIDs)
	if err != nil {
		return nil, err
	}
	return &core.DreamingSessionInput{RawEventIDs: rawEventIDs, Memories: memories}, nil
}

// PromoteMemories marks existing memories with a deeper dreaming tier in metadata_json.
func (s *Store) PromoteMemories(ctx context.Context, req *core.DreamingPromotionRequest) (*core.DreamingPromotionResult, error) {
	if err := validateDreamingPromotionRequest(req); err != nil {
		return nil, err
	}
	sql, args := dreamingPromotionStatement(req)
	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("promote dreaming memories: %w", err)
	}
	defer rows.Close()

	result := &core.DreamingPromotionResult{MemoryIDs: []string{}}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan promoted memory id: %w", err)
		}
		result.MemoryIDs = append(result.MemoryIDs, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("scan promoted memories: %w", err)
	}
	result.PromotedCount = len(result.MemoryIDs)
	return result, nil
}

func validateDreamingPromotionRequest(req *core.DreamingPromotionRequest) error {
	if req == nil {
		return fmt.Errorf("%w: dreaming promotion request is required", core.ErrInvalidArgument)
	}
	if strings.TrimSpace(req.TenantID) == "" {
		return fmt.Errorf("%w: dreaming promotion tenant_id is required", core.ErrInvalidArgument)
	}
	if strings.TrimSpace(req.WorkspaceID) == "" {
		return fmt.Errorf("%w: dreaming promotion workspace_id is required", core.ErrInvalidArgument)
	}
	if strings.TrimSpace(req.JobID) == "" {
		return fmt.Errorf("%w: dreaming promotion job_id is required", core.ErrInvalidArgument)
	}
	switch req.Tier {
	case core.DreamingTierMidTerm, core.DreamingTierLongTerm, core.DreamingTierUltraLongTerm:
	case "":
		return fmt.Errorf("%w: dreaming promotion tier is required", core.ErrInvalidArgument)
	default:
		return fmt.Errorf("%w: unsupported dreaming promotion tier %q", core.ErrInvalidArgument, req.Tier)
	}
	if req.MinConfidence < 0 || req.MinConfidence > 1 {
		return fmt.Errorf("%w: dreaming promotion min_confidence must be between 0 and 1", core.ErrInvalidArgument)
	}
	return nil
}

func (s *Store) dreamingRawEventIDs(ctx context.Context, req *core.DreamSessionRequest) ([]string, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id
		FROM raw_events
		WHERE tenant_id = $1
		  AND workspace_id = $2
		  AND session_id = $3
		ORDER BY occurred_at ASC, created_at ASC, id ASC
	`, req.TenantID, req.WorkspaceID, req.SessionID)
	if err != nil {
		return nil, fmt.Errorf("load dream_session raw events: %w", err)
	}
	defer rows.Close()

	ids := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan dream_session raw event id: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("scan dream_session raw event ids: %w", err)
	}
	return ids, nil
}

func (s *Store) dreamingMemoriesForRawEvents(ctx context.Context, req *core.DreamSessionRequest, rawEventIDs []string) ([]*core.Memory, error) {
	if len(rawEventIDs) == 0 {
		return []*core.Memory{}, nil
	}
	rows, err := s.pool.Query(ctx, `
		SELECT DISTINCT m.id, m.tenant_id, m.workspace_id, m.scope, m.group_id,
		       m.owner_entity_id, m.kind, m.artifact_class, m.text, m.fingerprint,
		       m.confidence, m.status, m.valid_from, m.valid_to, m.latest_flag,
		       m.metadata_json, m.embedding_model, m.embedding_dims,
		       m.embedding_updated_at, m.created_at, m.updated_at
		FROM memories m
		JOIN memory_trace mt ON mt.memory_id = m.id
		WHERE m.tenant_id = $1
		  AND m.workspace_id = $2
		  AND m.status = $3
		  AND m.latest_flag = true
		  AND mt.raw_event_ids && $4::text[]
		ORDER BY m.confidence DESC, m.updated_at DESC, m.id ASC
	`, req.TenantID, req.WorkspaceID, core.MemoryStatusActive, rawEventIDs)
	if err != nil {
		return nil, fmt.Errorf("load dream_session memories: %w", err)
	}
	defer rows.Close()
	return scanDreamingMemories(rows)
}

func scanDreamingMemories(rows pgx.Rows) ([]*core.Memory, error) {
	memories := []*core.Memory{}
	for rows.Next() {
		memory := &core.Memory{}
		if err := rows.Scan(
			&memory.ID, &memory.TenantID, &memory.WorkspaceID, &memory.Scope,
			&memory.GroupID, &memory.OwnerEntityID, &memory.Kind, &memory.ArtifactClass,
			&memory.Text, &memory.Fingerprint, &memory.Confidence, &memory.Status,
			&memory.ValidFrom, &memory.ValidTo, &memory.LatestFlag, &memory.MetadataJSON,
			&memory.EmbeddingModel, &memory.EmbeddingDims, &memory.EmbeddingUpdatedAt,
			&memory.CreatedAt, &memory.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan dream_session memory: %w", err)
		}
		memories = append(memories, memory)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("scan dream_session memories: %w", err)
	}
	return memories, nil
}

func dreamingPromotionStatement(req *core.DreamingPromotionRequest) (string, []any) {
	now := req.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
	args := []any{
		req.TenantID,
		req.WorkspaceID,
		req.Tier,
		req.JobID,
		now.UTC(),
		req.MinConfidence,
		nullableStringSlice(req.MemoryIDs),
	}
	filters := []string{
		"tenant_id = $1",
		"workspace_id = $2",
		"status = 'active'",
		"latest_flag = true",
		"confidence >= $6",
		"($7::text[] IS NULL OR id = ANY($7::text[]))",
		"(metadata_json #>> '{dreaming,tier}' IS NULL OR metadata_json #>> '{dreaming,tier}' <> $3)",
	}
	if req.RequireStableKind {
		filters = append(filters, "kind IN ('fact','preference','trait','goal','constraint','decision','procedure')")
		filters = append(filters, "scope <> 'session_scratch'")
	}
	where := strings.Join(filters, "\n		  AND ")
	sql := fmt.Sprintf(`
		WITH selected AS (
			SELECT id
			FROM memories
			WHERE %s
			ORDER BY confidence DESC, updated_at DESC, id ASC
			LIMIT %d
			FOR UPDATE SKIP LOCKED
		)
		UPDATE memories m
		SET metadata_json = jsonb_set(
		        m.metadata_json,
		        '{dreaming}',
		        COALESCE(m.metadata_json->'dreaming', '{}'::jsonb)
		            || jsonb_build_object(
		                'tier', $3,
		                'last_dream_job_id', $4,
		                'promoted_at', to_jsonb($5::timestamptz)
		            ),
		        true
		    ),
		    updated_at = $5
		FROM selected
		WHERE m.id = selected.id
		RETURNING m.id
	`, where, dreamingPromotionLimit)
	return sql, args
}

func nullableStringSlice(values []string) any {
	if len(values) == 0 {
		return nil
	}
	return values
}
