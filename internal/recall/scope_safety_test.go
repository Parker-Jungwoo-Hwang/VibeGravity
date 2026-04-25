// ============================================================
// FILE     : internal/recall/scope_safety_test.go
// PURPOSE  : Verifies prefetch recall requests preserve private, workspace, and group visibility boundaries.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : recall scope safety tests
// DEPENDS  : context, reflect, testing, internal/core
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: Recall should request group_shared memory only from explicit membership-derived group IDs.
// ============================================================

package recall

import (
	"context"
	"reflect"
	"testing"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

func TestAssemblerPrefetch_ScopeSafetyWithoutGroupMembership(t *testing.T) {
	t.Parallel()

	memories := &fakeMemoryStore{resp: &core.SearchMemoriesResponse{
		Memories: []core.MemoryResult{{
			MemoryID:      "mem_private",
			Text:          "Private memory visible only to the requesting actor.",
			Scope:         core.MemoryScopeAgentPrivate,
			OwnerEntityID: "agent:hermes-main",
			LatestFlag:    true,
		}},
	}}
	assembler := NewAssembler(Dependencies{
		Memories: memories,
		Groups:   &fakeGroupStore{},
	})

	resp, err := assembler.Prefetch(context.Background(), testPrefetchRequest())
	if err != nil {
		t.Fatalf("Prefetch returned error: %v", err)
	}

	wantScopes := []core.MemoryScope{
		core.MemoryScopeAgentPrivate,
		core.MemoryScopeWorkspaceShared,
		core.MemoryScopeSessionScratch,
	}
	if !reflect.DeepEqual(memories.lastReq.Scopes, wantScopes) {
		t.Fatalf("recall without membership should not request group_shared: got %v want %v", memories.lastReq.Scopes, wantScopes)
	}
	if memories.lastReq.OwnerEntityID != "agent:hermes-main" {
		t.Fatalf("agent_private recall must carry actor owner, got %#v", memories.lastReq)
	}
	if len(memories.lastReq.VisibleGroupIDs) != 0 {
		t.Fatalf("recall without memberships should pass no visible groups, got %#v", memories.lastReq.VisibleGroupIDs)
	}
	if len(resp.Blocks) != 1 || resp.Blocks[0].Scope != core.MemoryScopeAgentPrivate || resp.Blocks[0].OwnerEntityID != "agent:hermes-main" {
		t.Fatalf("private recall block lost owner/scope metadata: %#v", resp.Blocks)
	}
}

func TestAssemblerPrefetch_ScopeSafetyWithWorkspaceAndGroupMembership(t *testing.T) {
	t.Parallel()

	groupID := "group_design"
	memories := &fakeMemoryStore{resp: &core.SearchMemoriesResponse{
		Memories: []core.MemoryResult{
			{
				MemoryID:      "mem_workspace",
				Text:          "Workspace recall keeps shared rules available.",
				Scope:         core.MemoryScopeWorkspaceShared,
				OwnerEntityID: "workspace:workspace_1",
				LatestFlag:    true,
			},
			{
				MemoryID:   "mem_group",
				Text:       "Design group recall requires explicit membership.",
				Scope:      core.MemoryScopeGroupShared,
				GroupID:    &groupID,
				LatestFlag: true,
			},
		},
	}}
	assembler := NewAssembler(Dependencies{
		Memories: memories,
		Groups: &fakeGroupStore{memberships: []*core.MemoryGroupMembership{{
			GroupID:  groupID,
			EntityID: "agent:hermes-main",
		}}},
	})

	resp, err := assembler.Prefetch(context.Background(), testPrefetchRequest())
	if err != nil {
		t.Fatalf("Prefetch returned error: %v", err)
	}

	wantScopes := []core.MemoryScope{
		core.MemoryScopeAgentPrivate,
		core.MemoryScopeWorkspaceShared,
		core.MemoryScopeSessionScratch,
		core.MemoryScopeGroupShared,
	}
	if !reflect.DeepEqual(memories.lastReq.Scopes, wantScopes) {
		t.Fatalf("recall with membership should request group_shared through visible groups: got %v want %v", memories.lastReq.Scopes, wantScopes)
	}
	if !reflect.DeepEqual(memories.lastReq.VisibleGroupIDs, []string{groupID}) {
		t.Fatalf("recall should pass membership-derived visible groups, got %#v", memories.lastReq.VisibleGroupIDs)
	}
	if len(resp.Blocks) != 2 {
		t.Fatalf("expected workspace and group memory blocks, got %#v", resp.Blocks)
	}
	gotScopes := []core.MemoryScope{resp.Blocks[0].Scope, resp.Blocks[1].Scope}
	if !containsScope(gotScopes, core.MemoryScopeWorkspaceShared) || !containsScope(gotScopes, core.MemoryScopeGroupShared) {
		t.Fatalf("workspace and group memory scopes should stay distinct in recall blocks: %#v", resp.Blocks)
	}
}

func containsScope(values []core.MemoryScope, needle core.MemoryScope) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}
