// ============================================================
// FILE     : internal/embed/doc.go
// PURPOSE  : Records the deferred local embedding runtime boundary.
// LAYER    : infra
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : package embed
// DEPENDS  : plans/03_target-architecture_codex-first.md
// USED_BY  : future embedding client implementation
// ------------------------------------------------------------
// AGENT_NOTE: Keep local runtime embedding-focused; do not add a local extractor here.
// ============================================================

// Package embed is reserved for the local embedding client.
//
// The current runtime slice intentionally does not implement an embedding HTTP
// client here. Recall and Stage 2 source preparation use store-backed lexical
// retrieval today. Vector search and embedding writes should land behind an
// explicit embedding endpoint/model/dims configuration and a retrieval proof,
// without adding any local text extraction path.
package embed
