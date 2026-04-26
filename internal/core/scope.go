// ============================================================
// FILE     : internal/core/scope.go
// PURPOSE  : Defines explicit visibility scopes for memory artifacts.
// LAYER    : domain
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : MemoryScope
// DEPENDS  : internal/core/memory.go, internal/core/note.go, internal/core/plan.go
// USED_BY  : every memory, note, plan, profile, and recall path
// ------------------------------------------------------------
// AGENT_NOTE: Scope must never be implicit or nullable in memory writes.
// ============================================================

package core

// MemoryScope identifies the visibility boundary for a memory artifact.
type MemoryScope string

const (
	// MemoryScopeAgentPrivate is visible to one owning agent and the operator.
	MemoryScopeAgentPrivate MemoryScope = "agent_private"
	// MemoryScopeWorkspaceShared is visible to members of the workspace.
	MemoryScopeWorkspaceShared MemoryScope = "workspace_shared"
	// MemoryScopeGroupShared is visible to members of a named memory group.
	MemoryScopeGroupShared MemoryScope = "group_shared"
	// MemoryScopeSessionScratch is short-lived session-local context.
	MemoryScopeSessionScratch MemoryScope = "session_scratch"
)
