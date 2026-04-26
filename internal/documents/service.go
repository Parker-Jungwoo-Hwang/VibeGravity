// ============================================================
// FILE     : internal/documents/service.go
// PURPOSE  : Implements document ingestion, lexical chunking, and document search use cases.
// LAYER    : application
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : Service, NewService
// DEPENDS  : crypto/sha256, encoding/json, internal/core, internal/store
// USED_BY  : internal/kernel, internal/documents tests
// ------------------------------------------------------------
// AGENT_NOTE: Keep chunks lexical until an explicit embedding client is wired.
// ============================================================

package documents

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/store"
)

const documentChunkMaxRunes = 1800

// Service owns document use cases.
type Service struct {
	documents store.DocumentStore
	clock     func() time.Time
}

// NewService builds a document service.
func NewService(documents store.DocumentStore) *Service {
	return &Service{documents: documents, clock: time.Now}
}

// AddDocument stores a document and its initial lexical retrieval chunks.
func (s *Service) AddDocument(ctx context.Context, req *core.AddDocumentRequest) (*core.AddDocumentResponse, error) {
	if s == nil || s.documents == nil {
		return nil, fmt.Errorf("%w: add document", core.ErrNotImplemented)
	}
	if req == nil {
		return nil, fmt.Errorf("%w: add document request is required", core.ErrInvalidArgument)
	}
	if err := requireFields(map[string]string{
		"tenant_id":    req.TenantID,
		"workspace_id": req.WorkspaceID,
		"source":       req.Source,
		"title":        req.Title,
		"content":      req.Content,
	}); err != nil {
		return nil, err
	}
	now := s.clock().UTC()
	document := &core.Document{
		TenantID:     req.TenantID,
		WorkspaceID:  req.WorkspaceID,
		Source:       req.Source,
		Title:        req.Title,
		Fingerprint:  valueOr(req.Fingerprint, documentFingerprint(req)),
		MetadataJSON: jsonOrEmpty(req.MetadataJSON),
		VersionHint:  req.VersionHint,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	chunks := buildDocumentChunks("", req.Content, now)
	if err := s.documents.AddDocumentWithChunks(ctx, document, chunks); err != nil {
		return nil, err
	}
	chunkIDs := make([]string, 0, len(chunks))
	for _, chunk := range chunks {
		chunkIDs = append(chunkIDs, chunk.ID)
	}
	return &core.AddDocumentResponse{DocumentID: document.ID, ChunkIDs: chunkIDs, Status: "created"}, nil
}

// SearchDocuments delegates document search to storage.
func (s *Service) SearchDocuments(ctx context.Context, req *core.SearchDocumentsRequest) (*core.SearchDocumentsResponse, error) {
	if s == nil || s.documents == nil {
		return nil, fmt.Errorf("%w: search documents", core.ErrNotImplemented)
	}
	if req == nil {
		return nil, fmt.Errorf("%w: search documents request is required", core.ErrInvalidArgument)
	}
	if err := requireFields(map[string]string{
		"tenant_id":    req.TenantID,
		"workspace_id": req.WorkspaceID,
	}); err != nil {
		return nil, err
	}
	return s.documents.SearchDocuments(ctx, req)
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

func documentFingerprint(req *core.AddDocumentRequest) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{
		req.TenantID,
		req.WorkspaceID,
		req.Source,
		req.Title,
		req.Content,
	}, "\x00")))
	return fmt.Sprintf("sha256:%x", sum[:])
}

func buildDocumentChunks(documentID, content string, now time.Time) []*core.DocumentChunk {
	paragraphs := strings.Split(strings.TrimSpace(content), "\n\n")
	chunks := make([]*core.DocumentChunk, 0, len(paragraphs))
	var builder strings.Builder
	flush := func() {
		text := strings.TrimSpace(builder.String())
		if text == "" {
			builder.Reset()
			return
		}
		chunks = append(chunks, &core.DocumentChunk{
			DocumentID:     documentID,
			ChunkIndex:     len(chunks),
			Text:           text,
			MetadataJSON:   json.RawMessage(`{}`),
			EmbeddingModel: "pending",
			EmbeddingDims:  0,
			CreatedAt:      now,
			UpdatedAt:      now,
		})
		builder.Reset()
	}
	for _, paragraph := range paragraphs {
		paragraph = strings.TrimSpace(paragraph)
		if paragraph == "" {
			continue
		}
		if builder.Len() > 0 && builder.Len()+len(paragraph)+2 > documentChunkMaxRunes {
			flush()
		}
		if len([]rune(paragraph)) > documentChunkMaxRunes {
			flush()
			for _, part := range splitRunes(paragraph, documentChunkMaxRunes) {
				builder.WriteString(part)
				flush()
			}
			continue
		}
		if builder.Len() > 0 {
			builder.WriteString("\n\n")
		}
		builder.WriteString(paragraph)
	}
	flush()
	return chunks
}

func splitRunes(text string, maxRunes int) []string {
	runes := []rune(text)
	parts := make([]string, 0, (len(runes)/maxRunes)+1)
	for start := 0; start < len(runes); start += maxRunes {
		end := start + maxRunes
		if end > len(runes) {
			end = len(runes)
		}
		parts = append(parts, string(runes[start:end]))
	}
	return parts
}
