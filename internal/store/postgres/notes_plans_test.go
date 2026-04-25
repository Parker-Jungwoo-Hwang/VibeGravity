// ============================================================
// FILE     : internal/store/postgres/notes_plans_test.go
// PURPOSE  : Verifies note and plan lookup SQL preserves actor-private visibility boundaries.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : postgres note and plan statement tests
// DEPENDS  : strings, testing, internal/core
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: Keep pinned note and active plan retrieval tenant- and actor-scoped.
// ============================================================

package postgres

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

func TestListPinnedNotesStatementScopesAgentPrivateToOwner(t *testing.T) {
	t.Parallel()

	sql, args := listPinnedNotesStatement(&core.ListPinnedNotesRequest{
		TenantID:      "tenant_1",
		WorkspaceID:   "workspace_1",
		OwnerEntityID: "agent:hermes-main",
		Scopes:        []core.MemoryScope{core.MemoryScopeAgentPrivate, core.MemoryScopeWorkspaceShared},
	})

	if !strings.Contains(sql, "tenant_id = $1") {
		t.Fatalf("expected tenant predicate for pinned notes, got:\n%s", sql)
	}
	if !strings.Contains(sql, "scope <> 'agent_private'") {
		t.Fatalf("expected non-private note scopes to bypass owner filter, got:\n%s", sql)
	}
	if !strings.Contains(sql, "owner_entity_id = $4") {
		t.Fatalf("expected agent_private note owner predicate, got:\n%s", sql)
	}
	if len(args) != 4 || args[3] != "agent:hermes-main" {
		t.Fatalf("unexpected pinned notes args: %#v", args)
	}
}

func readCurrentFile(t *testing.T, name string) string {
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

func extractSourceBetween(t *testing.T, source, startMarker, endMarker string) string {
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

func TestGetActivePlansStatementScopesAgentPrivateToOwner(t *testing.T) {
	t.Parallel()

	sql, args := getActivePlansStatement(&core.GetActivePlansRequest{
		TenantID:      "tenant_1",
		WorkspaceID:   "workspace_1",
		OwnerEntityID: "agent:hermes-main",
		Scopes:        []core.MemoryScope{core.MemoryScopeAgentPrivate, core.MemoryScopeWorkspaceShared},
	})

	if !strings.Contains(sql, "tenant_id = $1") {
		t.Fatalf("expected tenant predicate for active plans, got:\n%s", sql)
	}
	if !strings.Contains(sql, "scope <> 'agent_private'") {
		t.Fatalf("expected non-private plan scopes to bypass owner filter, got:\n%s", sql)
	}
	if !strings.Contains(sql, "owner_entity_id = $4") {
		t.Fatalf("expected agent_private plan owner predicate, got:\n%s", sql)
	}
	if len(args) != 4 || args[3] != "agent:hermes-main" {
		t.Fatalf("unexpected active plans args: %#v", args)
	}
}

func TestUpdatePlanSourceUsesTenantWorkspaceAndPatchSemantics(t *testing.T) {
	t.Parallel()

	source := readCurrentFile(t, "notes_plans.go")
	updateSource := extractSourceBetween(t, source, "func (s *Store) UpdatePlan", "// GetActivePlans loads active plans for recall.")

	for _, want := range []string{
		"COALESCE(NULLIF($2, ''), title)",
		"COALESCE(NULLIF($3, ''), status)",
		"COALESCE($4, evidence_json)",
		"AND tenant_id = $5",
		"AND workspace_id = $6",
		"if items != nil",
	} {
		if !strings.Contains(updateSource, want) {
			t.Fatalf("UpdatePlan must preserve %q, got:\n%s", want, updateSource)
		}
	}
}
