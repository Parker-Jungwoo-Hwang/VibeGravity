# Recall Preview Trust Metadata

Date: 2026-04-25
Scope: first code slice for the Hermes Memory trust loop.

## Summary

Implemented the first product-code slice for the Hermes Memory trust loop.

`PrefetchResponse` recall blocks now carry operator-visible trust metadata:
scope, source, source id, status, freshness, owner, and stable block id where
available. Recall metadata also reports whether the response is degraded and
why supporting stores are unavailable.

Hermes rendering now preserves scope/source/freshness labels in compact text,
and MCP exposes `recall_preview` as an alias for `prefetch`.

## Finding or slice fixed

Before this slice, `prefetch()` returned useful typed blocks, but blocks only
carried `kind`, `priority`, and `text`. That was insufficient for the V1 promise:
"Hermes remembers the right project context, shows why, and lets the operator
fix memory once."

This slice opens the first trust surface without changing graph write semantics.

## Files changed

- `internal/core/dto.go`
- `internal/recall/assembler.go`
- `internal/recall/assembler_test.go`
- `internal/hermes/provider.go`
- `internal/hermes/provider_test.go`
- `internal/mcp/surface.go`
- `internal/mcp/surface_test.go`
- `plans/05_runtime-contracts_ingest-recall-apply.md`
- `plans/10_workpack_hermes-provider-and-external-surfaces.md`
- `docs/review-packets/recall-preview-trust-metadata.md`

## Tests run

- `go test ./internal/core ./internal/recall ./internal/hermes ./internal/mcp` - passed
- `go test ./...` - passed
- `make eval` - passed
- `make lint` - passed
- `make check-headers` - passed
- `git diff --check` - passed

## Remaining risks

- Worker/Codex freshness now has an operator-visible recall path in the follow-up
  packet `operator-visible-degraded-recall-freshness.md`; real Codex remains
  disabled by default.
- Hermes has an in-repo provider adapter and MCP bootstrap path, but the real
  Hermes runtime roundtrip still needs verification.
- `recall_preview` is currently an MCP alias over `prefetch`; a richer CLI or
  Hermes command can format it more explicitly later.

## Source Review

- Estimated source: in-repo VibeGravity contracts and implementation.
- Suspected license: project-internal original code.
- Similarity risk: low.
- Human review required: yes, because recall DTOs are part of the v1 API
  contract.
