// ============================================================
// FILE     : internal/worker/doc.go
// PURPOSE  : Documents the background job processor package.
// LAYER    : application
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : package worker
// DEPENDS  : docs/adr-008-package-layout.md, plans/05_runtime-contracts_ingest-recall-apply.md
// USED_BY  : cmd/worker, tests
// ------------------------------------------------------------
// AGENT_NOTE: This package orchestrates jobs; semantic extraction remains Codex-first through internal/reasoning.
// ============================================================

// Package worker claims ingest_jobs and dispatches them to reasoning and graph apply services.
package worker
