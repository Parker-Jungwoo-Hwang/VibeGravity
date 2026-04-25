// ============================================================
// FILE     : internal/store/postgres/documents_test.go
// PURPOSE  : Verifies document storage source preserves idempotent chunk replacement.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : postgres document source tests
// DEPENDS  : os, path/filepath, runtime, strings, testing
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: These tests lock storage contracts without requiring a live database.
// ============================================================

package postgres

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestAddDocumentChunksReplacesExistingChunksForDocument(t *testing.T) {
	t.Parallel()

	source := readPostgresSourceFile(t, "documents.go")
	addChunksSource := extractPostgresSourceBetween(t, source, "func (s *Store) AddDocumentChunks", "if err := tx.Commit")
	replaceChunksSource := extractPostgresSourceBetween(t, source, "func replaceDocumentChunksInTx", "func insertDocumentChunkInTx")

	if !strings.Contains(replaceChunksSource, "DELETE FROM document_chunks WHERE document_id = $1") {
		t.Fatalf("document chunk replacement must delete old chunks for idempotent document upserts, got:\n%s", replaceChunksSource)
	}
	if !strings.Contains(addChunksSource, "replacedDocuments") {
		t.Fatalf("AddDocumentChunks should delete once per document, got:\n%s", addChunksSource)
	}
	if !strings.Contains(addChunksSource, "replaceDocumentChunksInTx(ctx, tx, chunk.DocumentID") {
		t.Fatalf("AddDocumentChunks should route first chunk per document through replacement helper, got:\n%s", addChunksSource)
	}
}

func TestAddDocumentWithChunksUsesSingleTransaction(t *testing.T) {
	t.Parallel()

	source := readPostgresSourceFile(t, "documents.go")
	atomicSource := extractPostgresSourceBetween(t, source, "func (s *Store) AddDocumentWithChunks", "func (s *Store) AddDocument(")

	for _, required := range []string{
		"s.pool.Begin(ctx)",
		"defer func() { _ = tx.Rollback(ctx) }()",
		"addDocumentInTx(ctx, tx, document)",
		"addDocumentChunksInTx(ctx, tx, document.ID, chunks)",
		"tx.Commit(ctx)",
	} {
		if !strings.Contains(atomicSource, required) {
			t.Fatalf("AddDocumentWithChunks must keep document and chunks in one transaction; missing %q in:\n%s", required, atomicSource)
		}
	}
}

func readPostgresSourceFile(t *testing.T, name string) string {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not locate current test file")
	}
	data, err := os.ReadFile(filepath.Join(filepath.Dir(currentFile), name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return string(data)
}

func extractPostgresSourceBetween(t *testing.T, source, startMarker, endMarker string) string {
	t.Helper()

	start := strings.Index(source, startMarker)
	if start < 0 {
		t.Fatalf("missing start marker %q", startMarker)
	}
	remainder := source[start:]
	end := strings.Index(remainder, endMarker)
	if end < 0 {
		t.Fatalf("missing end marker %q", endMarker)
	}
	return remainder[:end]
}
