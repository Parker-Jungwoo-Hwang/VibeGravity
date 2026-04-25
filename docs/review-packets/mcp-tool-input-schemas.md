# MCP Tool Input Schemas

Date: 2026-04-25
Scope: MCP protocol tool discovery for the Hermes Memory trust loop.

## Summary

`internal/mcp.Server` now advertises concrete JSON input schemas from
`tools/list` instead of returning the same generic object schema for every
tool.

This keeps MCP clients aligned with the operator trust loop: recall preview,
correction, timeline, and explain-memory discovery now show the tenant,
workspace, actor, memory, correction, scope, and evidence fields needed to call
the tools safely.

## Finding or slice fixed

The stdio MCP server delegated tool calls correctly, but tool discovery did not
tell clients which inputs mattered. That made `recall_preview`,
`correct_memory`, `view_timeline`, and `explain_memory` harder to use from MCP
clients without separately reading Go DTOs.

This slice changes discovery only. It does not change core service behavior,
storage semantics, retry behavior, graph writes, or Hermes configuration.

## Files changed

- `internal/mcp/protocol.go`
- `internal/mcp/protocol_test.go`
- `plans/10_workpack_hermes-provider-and-external-surfaces.md`
- `docs/review-packets/mcp-tool-input-schemas.md`

## Tests run

- `go test ./internal/mcp` - passed.
- `go test ./...` - passed.
- `make eval` - passed.
- `make lint` - passed.
- `make check-headers` - passed.
- `git diff --check` - passed.

## Remaining risks

- Schemas are still hand-maintained beside the DTOs. If DTO fields change,
  `tools/list` schema tests should be updated in the same slice.
- The schemas describe the accepted JSON shape but do not replace service-level
  validation.
- Real Hermes runtime roundtrip remains unverified.

## Source Review

- Estimated source: in-repo VibeGravity DTOs and MCP protocol code.
- Suspected license: project-internal original code and documentation.
- Similarity risk: low.
- Human review required: yes, because tool discovery is an external protocol
  contract.
