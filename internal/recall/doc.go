// ============================================================
// FILE     : internal/recall/doc.go
// PURPOSE  : Provides package documentation for budget-aware recall pack assembly.
// LAYER    : application
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : package recall
// DEPENDS  : internal/core, internal/store
// USED_BY  : core service implementations, HTTP API, Hermes adapter, MCP tools
// ------------------------------------------------------------
// AGENT_NOTE: Build typed blocks before rendering and keep recall scope-aware and budget-aware.
// ============================================================

// Package recall owns prefetch candidate assembly, ranking, suppression, and token packing.
package recall
