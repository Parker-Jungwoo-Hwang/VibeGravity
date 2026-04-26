// ============================================================
// FILE     : internal/timeline/doc.go
// PURPOSE  : Documents the read-only timeline application service package.
// LAYER    : application
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : package timeline
// DEPENDS  : internal/core, internal/store
// USED_BY  : internal/kernel
// ------------------------------------------------------------
// AGENT_NOTE: Timeline is read-only operator visibility; never mutate memory state here.
// ============================================================

// Package timeline owns read-only memory and correction timeline use cases.
package timeline
