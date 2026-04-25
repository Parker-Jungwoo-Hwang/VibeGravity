// ============================================================
// FILE     : internal/store/postgres/documents.go
// PURPOSE  : Implements PostgreSQL persistence for documents and document chunks.
// LAYER    : infra
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : AddDocumentWithChunks, AddDocument, AddDocumentChunks
// DEPENDS  : internal/core, github.com/jackc/pgx/v5
// USED_BY  : document API, recall document search
// ------------------------------------------------------------
// AGENT_NOTE: Keep document storage separate from derived memory storage.
// ============================================================

package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

// AddDocumentWithChunks writes a document and replaces its chunks in one transaction.
func (s *Store) AddDocumentWithChunks(ctx context.Context, document *core.Document, chunks []*core.DocumentChunk) error {
	if document == nil {
		return fmt.Errorf("%w: document is required", core.ErrInvalidArgument)
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin add document with chunks: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := addDocumentInTx(ctx, tx, document); err != nil {
		return err
	}
	if err := addDocumentChunksInTx(ctx, tx, document.ID, chunks); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit add document with chunks: %w", err)
	}
	return nil
}

// AddDocument writes a document.
func (s *Store) AddDocument(ctx context.Context, document *core.Document) error {
	if document == nil {
		return fmt.Errorf("%w: document is required", core.ErrInvalidArgument)
	}
	return addDocumentInTx(ctx, s.pool, document)
}

// AddDocumentChunks writes retrieval chunks for a document.
func (s *Store) AddDocumentChunks(ctx context.Context, chunks []*core.DocumentChunk) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin add document chunks: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	replacedDocuments := make(map[string]struct{})
	for _, chunk := range chunks {
		if chunk == nil {
			continue
		}
		if _, ok := replacedDocuments[chunk.DocumentID]; !ok {
			if err := replaceDocumentChunksInTx(ctx, tx, chunk.DocumentID, []*core.DocumentChunk{chunk}); err != nil {
				return err
			}
			replacedDocuments[chunk.DocumentID] = struct{}{}
			continue
		}
		if err := insertDocumentChunkInTx(ctx, tx, chunk.DocumentID, chunk); err != nil {
			return err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit add document chunks: %w", err)
	}
	return nil
}

type documentExecutor interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func addDocumentInTx(ctx context.Context, exec documentExecutor, document *core.Document) error {
	var err error
	if document.ID == "" {
		document.ID, err = newID("doc")
		if err != nil {
			return err
		}
	}
	err = exec.QueryRow(ctx, `
		INSERT INTO documents (
			id, tenant_id, workspace_id, source, title, fingerprint,
			metadata_json, version_hint, created_at, updated_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		ON CONFLICT (tenant_id, workspace_id, fingerprint) DO UPDATE
		SET source = EXCLUDED.source,
		    title = EXCLUDED.title,
		    metadata_json = EXCLUDED.metadata_json,
		    version_hint = EXCLUDED.version_hint,
		    updated_at = EXCLUDED.updated_at
		RETURNING id
	`, document.ID, document.TenantID, document.WorkspaceID, document.Source,
		document.Title, document.Fingerprint, rawJSONOrEmpty(document.MetadataJSON),
		document.VersionHint, timeOrNow(document.CreatedAt), timeOrNow(document.UpdatedAt)).Scan(&document.ID)
	if err != nil {
		return fmt.Errorf("upsert document: %w", err)
	}
	return nil
}

func addDocumentChunksInTx(ctx context.Context, exec documentExecutor, documentID string, chunks []*core.DocumentChunk) error {
	return replaceDocumentChunksInTx(ctx, exec, documentID, chunks)
}

func replaceDocumentChunksInTx(ctx context.Context, exec documentExecutor, documentID string, chunks []*core.DocumentChunk) error {
	if _, err := exec.Exec(ctx, `DELETE FROM document_chunks WHERE document_id = $1`, documentID); err != nil {
		return fmt.Errorf("delete existing document chunks: %w", err)
	}
	for _, chunk := range chunks {
		if chunk == nil {
			continue
		}
		if err := insertDocumentChunkInTx(ctx, exec, documentID, chunk); err != nil {
			return err
		}
	}
	return nil
}

func insertDocumentChunkInTx(ctx context.Context, exec documentExecutor, documentID string, chunk *core.DocumentChunk) error {
	var err error
	if chunk.ID == "" {
		chunk.ID, err = newID("chunk")
		if err != nil {
			return err
		}
	}
	chunk.DocumentID = documentID
	_, err = exec.Exec(ctx, `
			INSERT INTO document_chunks (
				id, document_id, chunk_index, text, heading_path, metadata_json,
				neighbor_chunk_ids, embedding_model, embedding_dims, embedding_updated_at,
				created_at, updated_at
			)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		`, chunk.ID, chunk.DocumentID, chunk.ChunkIndex, chunk.Text, chunk.HeadingPath,
		rawJSONOrEmpty(chunk.MetadataJSON), chunk.NeighborChunkIDs,
		valueOr(chunk.EmbeddingModel, "pending"), chunk.EmbeddingDims, chunk.EmbeddingUpdatedAt,
		timeOrNow(chunk.CreatedAt), timeOrNow(chunk.UpdatedAt))
	if err != nil {
		return fmt.Errorf("insert document chunk: %w", err)
	}
	return nil
}
