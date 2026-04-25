// ============================================================
// FILE     : internal/store/postgres/timeline_test.go
// PURPOSE  : Verifies timeline query source preserves read-only scope filtering.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : postgres timeline source tests
// DEPENDS  : strings, testing
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: Timeline tests lock read-only query shape without requiring a live database.
// ============================================================

package postgres

import (
	"strings"
	"testing"
)

func TestTimelineStatementReadsMemoriesAndCorrections(t *testing.T) {
	t.Parallel()

	source := readPostgresSourceFile(t, "timeline.go")
	statementSource := extractPostgresSourceBetween(t, source, "func timelineStatement", "`, []any")

	for _, want := range []string{
		"FROM memories m",
		"LEFT JOIN memory_trace mt ON mt.memory_id = m.id",
		"FROM memory_corrections c",
		"JOIN memories m ON m.id = c.memory_id",
		"UNION ALL",
		"ORDER BY occurred_at DESC, id DESC",
	} {
		if !strings.Contains(statementSource, want) {
			t.Fatalf("timeline statement must preserve %q, got:\n%s", want, statementSource)
		}
	}
}

func TestTimelineStatementPreservesScopeAndOwnerFiltering(t *testing.T) {
	t.Parallel()

	source := readPostgresSourceFile(t, "timeline.go")
	statementSource := extractPostgresSourceBetween(t, source, "func timelineStatement", "`, []any")

	for _, want := range []string{
		"m.tenant_id = $1",
		"m.workspace_id = $2",
		"m.scope = ANY($3)",
		"m.scope <> 'group_shared'",
		"m.scope <> 'agent_private'",
		"m.owner_entity_id = $6",
	} {
		if !strings.Contains(statementSource, want) {
			t.Fatalf("timeline statement must preserve %q, got:\n%s", want, statementSource)
		}
	}
}

func TestTimelineStatementDoesNotMutateGraphState(t *testing.T) {
	t.Parallel()

	source := readPostgresSourceFile(t, "timeline.go")
	statementSource := extractPostgresSourceBetween(t, source, "func timelineStatement", "`, []any")

	for _, blocked := range []string{
		"UPDATE ",
		"INSERT ",
		"DELETE ",
		"latest_flag =",
		"memory_edges",
	} {
		if strings.Contains(statementSource, blocked) {
			t.Fatalf("timeline statement must stay read-only; found %q in:\n%s", blocked, statementSource)
		}
	}
}
