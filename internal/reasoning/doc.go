// ============================================================
// FILE     : internal/reasoning/doc.go
// PURPOSE  : Provides package documentation for Codex-first structured reasoning.
// LAYER    : application
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : package reasoning
// DEPENDS  : plans/03_target-architecture_codex-first.md, plans/05_runtime-contracts_ingest-recall-apply.md
// USED_BY  : worker pipeline, graph apply engine
// ------------------------------------------------------------
// AGENT_NOTE: Stage outputs must remain schema-first JSON and never bypass apply validation.
// ============================================================

// Package reasoning owns Codex stage 1 extraction and stage 2 resolution contracts.
package reasoning
