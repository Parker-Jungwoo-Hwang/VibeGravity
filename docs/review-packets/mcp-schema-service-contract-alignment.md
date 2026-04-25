# MCP Schema Service Contract Alignment

Date: 2026-04-25
Scope: MCP `tools/list` input schemas and service validation parity.

## Summary

MCP discovery now advertises the fields that the core service path actually
requires for trust-loop tools. This prevents MCP clients from constructing calls
that look valid from `tools/list` but are rejected by `Prefetch`, `SyncTurn`,
`AddNote`, `CreatePlan`, or `CorrectMemory` validation.

## Finding or slice fixed

The previous schemas understated several required fields:

- `prefetch`, `recall_preview`, and `degraded_status` now require
  `tenant_id`, `workspace_id`, `session_id`, and `actor_id`.
- `sync_turn` now requires `idempotency_key` and at least one `turn_events`
  item. Per-event `event_kind` remains required; `source` and `payload_json`
  are intentionally optional because ingest defaults them.
- `add_note` now requires `owner_entity_id`.
- `create_plan` now requires `owner_entity_id`; `status` remains optional
  because the service defaults it to `active`.
- `correct_memory` now requires `idempotency_key`.
- `search_memory`, `view_timeline`, and `explain_memory` stay aligned with the
  current service-required fields. Query and visibility-filter fields remain
  optional where core validation does not require them.

`degraded_status` was added to the MCP surface as the same thin Prefetch-meta
adapter already used by the Hermes provider so that the operator trust-loop
tool set has a discoverable MCP schema.

## Files changed

- `internal/mcp/protocol.go`
- `internal/mcp/protocol_test.go`
- `internal/mcp/surface.go`
- `internal/mcp/surface_test.go`
- `docs/review-packets/mcp-schema-service-contract-alignment.md`

## Tests run

- 2026-04-25 recheck: `go test ./internal/mcp` - passed.
- 2026-04-25 recheck: `go test ./internal/mcp -count=1` - passed.
- 2026-04-25 recheck: `go test ./internal/hermes` - passed.
- 2026-04-25 recheck: `go test ./...` - passed.
- 2026-04-25 recheck: `make lint` - passed.
- 2026-04-25 recheck: `make check-headers` - passed.
- 2026-04-25 recheck: `git diff --check` - passed.
- 2026-04-25 recheck: surface smoke payloads now use schema-complete
  `prefetch`, `recall_preview`, and `correct_memory` arguments, and the
  surface has a predictable incomplete `recall_preview` validation test for
  missing `actor_id`.
- 2026-04-25 recheck: direct `go run ./cmd/cli mcp serve --stdio`
  `tools/list` probe was blocked because the CLI opens the real service and
  local PostgreSQL was not listening on `127.0.0.1:5432`; in-process protocol
  tests verified the same `tools/list` handler without requiring a database.

## Remaining risks

- MCP schemas are still hand-maintained next to DTO/service validation.
- JSON schema validation is discovery-only here; service validation remains the
  enforcement point.
- The DB-opening CLI probe for `tools/list` still needs a live PostgreSQL
  service or a CLI test harness that can inject an in-memory service.

## Source Review

- Estimated source: in-repo VibeGravity DTOs, service validation, MCP adapter,
  and Hermes degraded-status adapter pattern.
- Suspected license: project-internal original code and documentation.
- Similarity risk: low.
- Human review required: yes, because MCP discovery is an external protocol
  contract.
