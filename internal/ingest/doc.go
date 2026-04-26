// ============================================================
// FILE     : internal/ingest/doc.go
// PURPOSE  : Provides package documentation for the sync_turn hot write path.
// LAYER    : application
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : package ingest
// DEPENDS  : internal/core, internal/store
// USED_BY  : core service implementations, HTTP API, Hermes adapter
// ------------------------------------------------------------
// AGENT_NOTE: Keep sync_turn fast: normalize, validate, dedupe, insert raw events, enqueue jobs, ack.
// ============================================================

// Package ingest owns the sync_turn hot path for raw event writes and job enqueueing.
package ingest
