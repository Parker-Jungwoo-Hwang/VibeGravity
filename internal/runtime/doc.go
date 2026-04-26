// ============================================================
// FILE     : internal/runtime/doc.go
// PURPOSE  : Documents process-level runtime composition for commands.
// LAYER    : infra
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : package runtime
// DEPENDS  : plans/03_target-architecture_codex-first.md
// USED_BY  : cmd/server, cmd/worker, cmd/cli, cmd/vibegravity
// ------------------------------------------------------------
// AGENT_NOTE: Keep process wiring here so command packages stay thin.
// ============================================================

// Package runtime composes VibeGravity process dependencies for HTTP, CLI/MCP,
// and worker entrypoints.
package runtime
