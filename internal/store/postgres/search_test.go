// ============================================================
// FILE     : internal/store/postgres/search_test.go
// PURPOSE  : Verifies PostgreSQL search SQL preserves actor-private visibility boundaries.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : postgres search statement tests
// DEPENDS  : strings, testing, internal/core
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: These tests lock privacy predicates without requiring a live database.
// ============================================================

package postgres

import (
	"strings"
	"testing"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

func TestSearchMemoriesStatementScopesAgentPrivateToOwner(t *testing.T) {
	t.Parallel()

	sql, args := searchMemoriesStatement(&core.SearchMemoriesRequest{
		TenantID:        "tenant_1",
		WorkspaceID:     "workspace_1",
		OwnerEntityID:   "agent:hermes-main",
		Query:           "stage 2",
		Scopes:          []core.MemoryScope{core.MemoryScopeAgentPrivate, core.MemoryScopeWorkspaceShared},
		ArtifactClasses: []core.ArtifactClass{core.ArtifactClassKnowledge},
	})

	if !strings.Contains(sql, "scope <> 'agent_private'") {
		t.Fatalf("expected non-private scopes to bypass owner filter, got:\n%s", sql)
	}
	if !strings.Contains(sql, "owner_entity_id = $6") {
		t.Fatalf("expected agent_private owner predicate, got:\n%s", sql)
	}
	if !strings.Contains(sql, "group_id, owner_entity_id, valid_from") {
		t.Fatalf("memory search must return owner_entity_id for caller-side visibility checks, got:\n%s", sql)
	}
	if len(args) != 7 || args[5] != "agent:hermes-main" {
		t.Fatalf("unexpected memory search args: %#v", args)
	}
}

func TestSearchMemoriesStatementKeepsWorkspaceSharedVisibleByWorkspace(t *testing.T) {
	t.Parallel()

	sql, args := searchMemoriesStatement(&core.SearchMemoriesRequest{
		TenantID:        "tenant_1",
		WorkspaceID:     "workspace_1",
		Query:           "workspace rule",
		Scopes:          []core.MemoryScope{core.MemoryScopeWorkspaceShared},
		ArtifactClasses: []core.ArtifactClass{core.ArtifactClassKnowledge},
	})

	for _, want := range []string{
		"tenant_id = $1",
		"workspace_id = $2",
		"scope = ANY($3)",
		"scope <> 'agent_private'",
		"scope <> 'group_shared'",
	} {
		if !strings.Contains(sql, want) {
			t.Fatalf("expected workspace search SQL to preserve %q, got:\n%s", want, sql)
		}
	}
	if len(args) != 7 || args[0] != "tenant_1" || args[1] != "workspace_1" {
		t.Fatalf("unexpected memory search args: %#v", args)
	}
	gotScopes, ok := args[2].([]core.MemoryScope)
	if !ok || len(gotScopes) != 1 || gotScopes[0] != core.MemoryScopeWorkspaceShared {
		t.Fatalf("unexpected workspace scopes: %#v", args[2])
	}
}

func TestSearchMemoriesStatementScopesGroupSharedToMemberships(t *testing.T) {
	t.Parallel()

	sql, args := searchMemoriesStatement(&core.SearchMemoriesRequest{
		TenantID:        "tenant_1",
		WorkspaceID:     "workspace_1",
		OwnerEntityID:   "agent:hermes-main",
		VisibleGroupIDs: []string{"group_design"},
		Query:           "stage 2",
		Scopes:          []core.MemoryScope{core.MemoryScopeGroupShared},
		ArtifactClasses: []core.ArtifactClass{core.ArtifactClassKnowledge},
	})

	if !strings.Contains(sql, "scope <> 'group_shared'") {
		t.Fatalf("expected non-group scopes to bypass group filter, got:\n%s", sql)
	}
	if !strings.Contains(sql, "group_id = ANY($7)") {
		t.Fatalf("expected group_shared membership predicate, got:\n%s", sql)
	}
	if len(args) != 7 {
		t.Fatalf("unexpected arg count: %#v", args)
	}
	gotGroups, ok := args[6].([]string)
	if !ok || len(gotGroups) != 1 || gotGroups[0] != "group_design" {
		t.Fatalf("unexpected group args: %#v", args[6])
	}
}

func TestSearchMemoriesStatementDoesNotOpenGroupSharedWithoutVisibleGroups(t *testing.T) {
	t.Parallel()

	sql, args := searchMemoriesStatement(&core.SearchMemoriesRequest{
		TenantID:        "tenant_1",
		WorkspaceID:     "workspace_1",
		Query:           "group decision",
		Scopes:          []core.MemoryScope{core.MemoryScopeGroupShared},
		ArtifactClasses: []core.ArtifactClass{core.ArtifactClassKnowledge},
	})

	if !strings.Contains(sql, "group_id = ANY($7)") {
		t.Fatalf("group_shared search must depend on visible group ids, got:\n%s", sql)
	}
	gotGroups, ok := args[6].([]string)
	if !ok {
		t.Fatalf("visible group argument should be a string slice, got %#v", args[6])
	}
	if len(gotGroups) != 0 {
		t.Fatalf("group_shared search without memberships should pass no visible groups, got %#v", gotGroups)
	}
}
