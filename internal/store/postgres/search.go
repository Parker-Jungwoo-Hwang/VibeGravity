// ============================================================
// FILE     : internal/store/postgres/search.go
// PURPOSE  : Implements minimal lexical memory and document search for degraded recall.
// LAYER    : infra
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : SearchMemories, SearchDocuments
// DEPENDS  : internal/core
// USED_BY  : recall assembler, future search APIs
// ------------------------------------------------------------
// AGENT_NOTE: Keep this as lexical fallback; vector ranking belongs behind explicit embedding paths.
// ============================================================

package postgres

import (
	"context"
	"fmt"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

// SearchMemories searches active latest memories with lexical fallback.
func (s *Store) SearchMemories(ctx context.Context, req *core.SearchMemoriesRequest) (*core.SearchMemoriesResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("%w: search memories request is required", core.ErrInvalidArgument)
	}
	sql, args := searchMemoriesStatement(req)
	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query memories: %w", err)
	}
	defer rows.Close()

	results := make([]core.MemoryResult, 0, 20)
	for rows.Next() {
		var result core.MemoryResult
		if err := rows.Scan(&result.MemoryID, &result.Kind, &result.ArtifactClass, &result.Text,
			&result.Confidence, &result.Scope, &result.GroupID, &result.OwnerEntityID, &result.ValidFrom, &result.LatestFlag); err != nil {
			return nil, fmt.Errorf("scan memory result: %w", err)
		}
		results = append(results, result)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate memory results: %w", err)
	}
	return &core.SearchMemoriesResponse{Memories: results}, nil
}

func searchMemoriesStatement(req *core.SearchMemoriesRequest) (string, []any) {
	return `
		SELECT id, kind, artifact_class, text, confidence, scope, group_id, owner_entity_id, valid_from, latest_flag
		FROM memories
		WHERE tenant_id = $1
		  AND workspace_id = $2
		  AND scope = ANY($3)
		  AND artifact_class = ANY($4)
		  AND status = 'active'
		  AND latest_flag = true
		  AND ($5 = '' OR text ILIKE '%' || $5 || '%')
		  AND (
		    scope <> 'agent_private'
		    OR ($6 <> '' AND owner_entity_id = $6)
		  )
		  AND (
		    scope <> 'group_shared'
		    OR ($7::text[] IS NOT NULL AND group_id = ANY($7))
		  )
		ORDER BY valid_from DESC, updated_at DESC
		LIMIT 20
	`, []any{req.TenantID, req.WorkspaceID, req.Scopes, req.ArtifactClasses, req.Query, req.OwnerEntityID, req.VisibleGroupIDs}
}

// SearchDocuments searches document chunks with lexical fallback.
func (s *Store) SearchDocuments(ctx context.Context, req *core.SearchDocumentsRequest) (*core.SearchDocumentsResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("%w: search documents request is required", core.ErrInvalidArgument)
	}
	rows, err := s.pool.Query(ctx, `
		SELECT c.id, c.document_id, c.text, 1.0::double precision AS score
		FROM document_chunks c
		JOIN documents d ON d.id = c.document_id
		WHERE d.tenant_id = $1
		  AND d.workspace_id = $2
		  AND ($3 = '' OR c.text ILIKE '%' || $3 || '%')
		ORDER BY d.updated_at DESC, c.chunk_index ASC
		LIMIT 20
	`, req.TenantID, req.WorkspaceID, req.Query)
	if err != nil {
		return nil, fmt.Errorf("query document chunks: %w", err)
	}
	defer rows.Close()

	chunks := make([]core.DocumentChunkResult, 0, 20)
	for rows.Next() {
		var chunk core.DocumentChunkResult
		if err := rows.Scan(&chunk.ChunkID, &chunk.DocumentID, &chunk.Text, &chunk.Score); err != nil {
			return nil, fmt.Errorf("scan document chunk result: %w", err)
		}
		chunks = append(chunks, chunk)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate document chunk results: %w", err)
	}
	return &core.SearchDocumentsResponse{Chunks: chunks}, nil
}
