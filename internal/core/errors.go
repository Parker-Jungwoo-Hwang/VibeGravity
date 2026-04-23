// ============================================================
// FILE     : internal/core/errors.go
// PURPOSE  : Defines shared domain errors for service and storage boundaries.
// LAYER    : domain
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : ErrNotFound, ErrDuplicate, ErrInvalidArgument, ErrConflict
// DEPENDS  : errors
// USED_BY  : internal/core services, internal/store implementations
// ------------------------------------------------------------
// AGENT_NOTE: Prefer these sentinel errors over transport-specific failures.
// ============================================================

package core

import "errors"

// ErrNotFound reports that a requested record does not exist.
var ErrNotFound = errors.New("not found")

// ErrDuplicate reports that an idempotent write already exists.
var ErrDuplicate = errors.New("duplicate")

// ErrInvalidArgument reports a contract validation failure.
var ErrInvalidArgument = errors.New("invalid argument")

// ErrConflict reports that a request would violate storage invariants.
var ErrConflict = errors.New("conflict")
