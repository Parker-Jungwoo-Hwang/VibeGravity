// ============================================================
// FILE     : internal/reasoning/doc.go
// PURPOSE  : Provides package documentation for Codex-first structured reasoning.
// LAYER    : application
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : package reasoning
// DEPENDS  : internal/core, internal/graph, internal/worker
// USED_BY  : worker pipeline, graph apply engine
// ------------------------------------------------------------
// AGENT_NOTE: Stage outputs must remain schema-first JSON and never bypass apply validation.
// ============================================================

// Package reasoning owns Codex stage 1 extraction and stage 2 resolution contracts.
package reasoning
