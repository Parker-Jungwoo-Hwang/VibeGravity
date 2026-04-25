# ExplainMemory Visibility Guard

Date: 2026-04-25
Scope: operator provenance lookup owner and group visibility.

## Summary

`ExplainMemory` now carries actor visibility into the storage trace lookup.
Private memory explanations require the requesting actor to match
`owner_entity_id`, and group-shared explanations require the requested memory's
group id to be included in `visible_group_ids`.

## Finding or slice fixed

The prior scope guard fixed tenant/workspace isolation, but a caller inside the
same workspace could still explain a private memory by guessing its memory id.
That left the explain surface weaker than search, recall, and timeline.

This slice adds optional `entity_id` and `visible_group_ids` to
`ExplainMemoryRequest`, wires them through HTTP and MCP tool input schemas, and
enforces the visibility predicate in the PostgreSQL trace query.

## Files changed

- `internal/core/dto.go`
- `internal/store/postgres/memories.go`
- `internal/store/postgres/memories_test.go`
- `internal/kernel/service_test.go`
- `internal/httpapi/router.go`
- `internal/httpapi/router_test.go`
- `internal/mcp/protocol.go`
- `internal/mcp/protocol_test.go`
- `internal/mcp/surface_test.go`
- `plans/05_runtime-contracts_ingest-recall-apply.md`
- `plans/10_workpack_hermes-provider-and-external-surfaces.md`
- `docs/review-packets/explain-memory-scope-guard.md`
- `docs/review-packets/explain-memory-visibility-guard.md`

## Tests run

- `go test ./internal/store/postgres ./internal/kernel ./internal/httpapi ./internal/mcp`
- `make lint`
- `make check-headers`
- `git diff --check`

Attempted but blocked by the active `codex-main-demo-eval` lane:

- `go test ./...` failed in `cmd/cli` and `internal/eval` because the new
  Hermes Memory demo eval expected memory/profile ordering that no longer
  matched observed recall.
- `make eval` failed for the same `cli eval demo` expectation mismatch. The
  golden eval portion passed.

## Remaining risks

- Existing HTTP callers that omit `entity_id` can still explain
  `workspace_shared` memory, but not `agent_private` memory. That is intended
  for compatibility and safety.
- Edge rows are still returned after the requested memory passes visibility
  checks. They include IDs and edge kinds, not memory text, but a future
  hardening slice could scope edge expansion through memory joins too.

## Source Review

- Estimated source: in-repo VibeGravity contracts and implementation.
- Suspected license: project-internal original code.
- Similarity risk: low.
- Human review required: recommended because this changes operator-visible
  provenance authorization behavior.
