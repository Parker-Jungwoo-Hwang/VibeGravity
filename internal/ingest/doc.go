// ============================================================
// FILE     : internal/ingest/doc.go
// PURPOSE  : Provides package documentation for the sync_turn hot write path.
// LAYER    : application
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : package ingest
// DEPENDS  : plans/05_runtime-contracts_ingest-recall-apply.md
// USED_BY  : core service implementations, HTTP API, Hermes adapter
// ------------------------------------------------------------
// AGENT_NOTE: Keep sync_turn fast: normalize, validate, dedupe, insert raw events, enqueue jobs, ack.
// ============================================================

// Package ingest owns the sync_turn hot path for raw event writes and job enqueueing.
package ingest
