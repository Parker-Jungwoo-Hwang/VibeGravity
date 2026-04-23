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
