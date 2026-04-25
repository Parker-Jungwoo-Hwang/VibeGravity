# Hermes Provider Tool Dispatch

## Summary

This pass turns the in-repo Hermes provider tool list into a thin executable
adapter for the core-backed trust-loop tools.

## Finding or slice fixed

`internal/hermes.Provider` already advertised `recall_preview`, memory search,
correction, explain, timeline, and degraded-status tools, but there was no
provider-level dispatch helper proving those tool names reached the shared core
service. `CallTool` now decodes JSON into the existing core DTOs and delegates
to the same service methods used by HTTP and MCP.

`degraded_status` intentionally calls `Prefetch` and returns only
`RecallMeta`, so Hermes-facing status stays tied to the actual recall freshness
signal. `show_plan` remains explicit `ErrNotImplemented` because the core
service does not yet expose a read-only plan-list API.

## Files changed

- `internal/hermes/provider.go`
- `internal/hermes/provider_test.go`
- `plans/10_workpack_hermes-provider-and-external-surfaces.md`
- `docs/review-packets/hermes-provider-tool-dispatch.md`

## Tests run

- `gofmt -w internal/hermes/provider.go internal/hermes/provider_test.go`
- `go test ./internal/hermes`
- `go test ./...`
- `make lint`
- `make check-headers`
- `make eval`
- `git diff --check`

## Remaining risks

- This is still an in-repo adapter test, not a real Hermes runtime roundtrip.
- `show_plan` needs a future read-only core API before it can be executable as
  a provider tool.

## Source Review

- Estimated source: original implementation based on existing VibeGravity DTOs
  and adapter patterns.
- Suspected license: project-owned VibeGravity code.
- Similarity risk: low; no external code or long snippets were used.
- Human review required: normal adapter review recommended before relying on
  provider tool dispatch outside the MCP bootstrap path.
