// ============================================================
// FILE     : internal/kernel/doc.go
// PURPOSE  : Provides package documentation for the concrete VibeGravity service composition.
// LAYER    : application
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : package kernel
// DEPENDS  : internal/core, internal/ingest, internal/recall
// USED_BY  : cmd/server, tests, future Hermes and MCP adapters
// ------------------------------------------------------------
// AGENT_NOTE: Keep this package as orchestration glue; product rules belong in the domain packages it composes.
// ============================================================

// Package kernel composes VibeGravity application services behind the core contract.
package kernel
