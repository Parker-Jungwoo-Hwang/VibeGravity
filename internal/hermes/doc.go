// ============================================================
// FILE     : internal/hermes/doc.go
// PURPOSE  : Provides package documentation for Hermes provider adapter semantics.
// LAYER    : interface
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : package hermes
// DEPENDS  : plans/10_workpack_hermes-provider-and-external-surfaces.md
// USED_BY  : Hermes provider integration, tests
// ------------------------------------------------------------
// AGENT_NOTE: Hermes is the first customer; adapter behavior must not fork core semantics.
// ============================================================

// Package hermes maps Hermes memory-provider lifecycle hooks to VibeGravity core calls.
package hermes
