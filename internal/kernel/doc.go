// ============================================================
// FILE     : internal/kernel/doc.go
// PURPOSE  : Provides package documentation for the VibeGravity service facade.
// LAYER    : application
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : package kernel
// DEPENDS  : internal/core, internal/ingest, internal/recall, product service packages
// USED_BY  : internal/runtime, tests, HTTP, Hermes, and MCP adapters
// ------------------------------------------------------------
// AGENT_NOTE: Keep this package as orchestration glue; product rules belong in the domain packages it composes.
// ============================================================

// Package kernel exposes the VibeGravity service facade behind the core contract.
package kernel
