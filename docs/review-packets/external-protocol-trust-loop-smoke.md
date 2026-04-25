# External Protocol Trust Loop Smoke

Date: 2026-04-25
Owner: Agent 7
Scope: MCP stdio and Hermes provider adapter smoke coverage for trust-loop tools.

## Summary

This slice adds external-surface smoke coverage for the Hermes Memory trust
loop without broadening into Hermes packaging, registry setup, or runtime
installation behavior.

The new MCP stdio smoke test drives newline-delimited JSON-RPC through
`initialize`, `tools/list`, and `tools/call`, then verifies the trust-loop tool
requests reach the shared core service DTOs unchanged. Hermes provider tests now
also assert that `CallTool` delegates the same trust-loop request DTOs to the
same core service methods.

## Finding or slice fixed

Covered tools:

- `recall_preview`: MCP and Hermes both decode to `core.PrefetchRequest` and
  delegate to `Prefetch`.
- `correct_memory`: MCP and Hermes both decode to `core.CorrectMemoryRequest`
  and delegate to `CorrectMemory`.
- `explain_memory`: MCP and Hermes both decode to `core.ExplainMemoryRequest`
  and delegate to `ExplainMemory`.
- `view_timeline`: MCP and Hermes both decode to `core.GetTimelineRequest` and
  delegate to `GetTimeline`.
- `degraded_status`: MCP and Hermes both decode to `core.PrefetchRequest`,
  delegate to `Prefetch`, and return `RecallMeta` only.

The MCP smoke test includes a negative `correct_memory` call missing
`memory_id`; the fake core service rejects it as a required-field error and the
protocol reports it as a tool error rather than inventing alternate behavior.

## Files changed

- `internal/mcp/stdio_smoke_test.go`
- `internal/mcp/surface.go`
- `internal/mcp/protocol.go`
- `internal/hermes/provider_test.go`
- `docs/review-packets/external-protocol-trust-loop-smoke.md`

## Tests run

- `go test ./internal/mcp`
- `go test ./internal/hermes`
- `go test ./cmd/cli`
- `go test ./...`
- `make lint`
- `make check-headers`
- `git diff --check`

## Remaining risks

- This is still local protocol smoke coverage with fake core services, not a
  real Hermes runtime roundtrip.
- MCP discovery schemas remain hand-maintained next to DTO/service validation.
- `degraded_status` intentionally reports recall freshness metadata only; it
  does not create a separate health subsystem.
- Broad Hermes packaging, registry behavior, and profile installation remain
  out of scope for this lane.

## Source Review

- Estimated source: original implementation based on existing VibeGravity core
  DTOs, MCP protocol adapter, and Hermes provider adapter patterns.
- Suspected license: project-owned VibeGravity code.
- Similarity risk: low; no external code or long snippets were used.
- Human review required: normal protocol-contract review recommended before
  relying on this as an external client compatibility guarantee.
