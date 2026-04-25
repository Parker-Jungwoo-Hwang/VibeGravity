// ============================================================
// FILE     : internal/store/postgres/corrections_test.go
// PURPOSE  : Verifies correction persistence SQL without requiring a live database.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : postgres correction source tests
// DEPENDS  : strings, testing
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: Lock the narrow correction intake boundary without testing supersession semantics here.
// ============================================================

package postgres

import (
	"strings"
	"testing"
)

func TestRecordMemoryCorrectionUsesSingleTransaction(t *testing.T) {
	t.Parallel()

	source := readPostgresSourceFile(t, "corrections.go")
	recordSource := extractPostgresSourceBetween(t, source, "func (s *Store) RecordMemoryCorrection", "func insertCorrectionRawEvent")

	for _, want := range []string{
		"s.pool.BeginTx(ctx, pgx.TxOptions{})",
		"defer func() { _ = tx.Rollback(ctx) }()",
		"insertCorrectionRawEvent(ctx, tx, event)",
		"insertMemoryCorrection(ctx, tx, correction)",
		"tx.Commit(ctx)",
	} {
		if !strings.Contains(recordSource, want) {
			t.Fatalf("RecordMemoryCorrection must preserve %q, got:\n%s", want, recordSource)
		}
	}
}

func TestCorrectionRawEventReturnsStableIDOnRetry(t *testing.T) {
	t.Parallel()

	source := readPostgresSourceFile(t, "corrections.go")
	insertSource := extractPostgresSourceBetween(t, source, "func insertCorrectionRawEvent", "func insertMemoryCorrection")

	for _, want := range []string{
		"ON CONFLICT (tenant_id, source, idempotency_key) DO UPDATE",
		"WHERE raw_events.workspace_id = EXCLUDED.workspace_id",
		"RETURNING id",
	} {
		if !strings.Contains(insertSource, want) {
			t.Fatalf("correction raw event insert must preserve stable retry IDs; missing %q in:\n%s", want, insertSource)
		}
	}
}

func TestMemoryCorrectionArtifactIsAppendSafe(t *testing.T) {
	t.Parallel()

	source := readPostgresSourceFile(t, "corrections.go")
	start := strings.Index(source, "func insertMemoryCorrection")
	if start < 0 {
		t.Fatal("missing insertMemoryCorrection")
	}
	insertSource := source[start:]
	if strings.Contains(insertSource, "latest_flag") || strings.Contains(insertSource, "memory_trace") || strings.Contains(insertSource, "memory_edges") {
		t.Fatalf("correction intake must not mutate graph apply/provenance tables, got:\n%s", insertSource)
	}
	for _, want := range []string{
		"INSERT INTO memory_corrections",
		"ON CONFLICT (tenant_id, workspace_id, idempotency_key) DO UPDATE",
		"memory_corrections.memory_id = EXCLUDED.memory_id",
		"memory_corrections.operator_id = EXCLUDED.operator_id",
		"memory_corrections.raw_event_id = EXCLUDED.raw_event_id",
		"memory_corrections.correction_text = EXCLUDED.correction_text",
		"memory_corrections.evidence_json = EXCLUDED.evidence_json",
		"RETURNING id, tenant_id, workspace_id, memory_id, operator_id, raw_event_id",
	} {
		if !strings.Contains(insertSource, want) {
			t.Fatalf("memory correction insert must preserve %q, got:\n%s", want, insertSource)
		}
	}
}

func TestGetMemoryCorrectionByIdempotencyLoadsReplayEvidence(t *testing.T) {
	t.Parallel()

	source := readPostgresSourceFile(t, "corrections.go")
	start := strings.Index(source, "func (s *Store) GetMemoryCorrectionByIdempotency")
	if start < 0 {
		t.Fatal("missing GetMemoryCorrectionByIdempotency")
	}
	loadSource := source[start:]

	for _, want := range []string{
		"FROM memory_corrections",
		"tenant_id = $1",
		"workspace_id = $2",
		"idempotency_key = $3",
		"id, tenant_id, workspace_id, memory_id, operator_id, raw_event_id",
	} {
		if !strings.Contains(loadSource, want) {
			t.Fatalf("correction replay lookup must preserve %q, got:\n%s", want, loadSource)
		}
	}
}
