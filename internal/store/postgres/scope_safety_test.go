// ============================================================
// FILE     : internal/store/postgres/scope_safety_test.go
// PURPOSE  : Verifies PostgreSQL provenance lookup keeps memory visibility boundaries explicit.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : postgres scope safety statement tests
// DEPENDS  : strings, testing
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: ExplainMemory must never become a bypass around search scope rules.
// ============================================================

package postgres

import (
	"strings"
	"testing"
)

func TestExplainMemoryTraceStatementPreservesVisibilityBoundaries(t *testing.T) {
	t.Parallel()

	sql := explainMemoryTraceStatement()

	for _, want := range []string{
		"JOIN memories m ON m.id = mt.memory_id",
		"m.tenant_id = $2",
		"m.workspace_id = $3",
		"m.scope <> 'agent_private'",
		"m.owner_entity_id = $4",
		"m.scope <> 'group_shared'",
		"m.group_id IS NOT NULL",
		"m.group_id = ANY($5)",
	} {
		if !strings.Contains(sql, want) {
			t.Fatalf("explain memory trace query must preserve %q, got:\n%s", want, sql)
		}
	}
}
