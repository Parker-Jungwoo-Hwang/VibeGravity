# MCP Trust Surface Delegation Tests

## Summary

This pass adds focused MCP adapter coverage for the operator trust-loop tools
that inspect memory state after recall and correction.

## Finding or slice fixed

`internal/mcp.Surface` already exposed and delegated `view_timeline` and
`explain_memory`, but the test coverage only locked the tool list and the
`recall_preview` / `correct_memory` calls. The new tests prove those inspection
tools decode JSON input, call the shared core service once, and return encoded
core responses.

## Files changed

- `internal/mcp/surface_test.go`
- `docs/review-packets/mcp-trust-surface-delegation-tests.md`

## Tests run

- `go test ./internal/mcp`
- `git diff --check`

## Remaining risks

- This is adapter coverage only. It does not verify a real Hermes runtime
  roundtrip or the full stdio MCP process.
- The concurrent freshness lane owns the broader `Prefetch` freshness metadata
  behavior, so this pass intentionally did not touch recall freshness files.

## Source Review

- Estimated source: original implementation and tests written from the current
  repository contracts.
- Suspected license: project-owned VibeGravity code.
- Similarity risk: low; no external code or long snippets were used.
- Human review required: no for licensing; yes for normal trust-surface review
  before broadening MCP/Hermes runtime behavior.
