# ADR-008: Product Package Layout

## Status

Accepted

## Context

VibeGravity is a Hermes-first shared memory kernel. The product documents define
three runtime promises:

- `sync_turn()` records quickly on the API hot path.
- background workers derive, resolve, and apply memory changes.
- `prefetch()` returns a short, budget-aware recall pack before the next turn.

The repository must make those responsibilities visible without splitting into
premature microservices. The package layout should also keep Hermes, MCP, and
HTTP surfaces on the same core semantics.

## Decision

Use a Go-first monorepo with small internal packages organized by product
responsibility:

| Path | Responsibility |
|---|---|
| `cmd/server` | HTTP API process entrypoint only |
| `cmd/worker` | background worker process entrypoint only |
| `cmd/cli` | local operator CLI and doctor entrypoint |
| `internal/core` | v1 service contract, DTOs, and domain records |
| `internal/kernel` | concrete `core.VibeGravityService` composition over ingest and recall |
| `internal/ingest` | `sync_turn()` normalization, validation, idempotent raw event writes, and job enqueueing |
| `internal/recall` | `prefetch()` typed block assembly, ranking, suppression, and token budgeting |
| `internal/worker` | background job claiming, dispatch, retry handoff, and process-level orchestration |
| `internal/graph` | apply engine, memory edges, profile merge, summaries, corrections, and dreaming rules |
| `internal/reasoning` | Codex stage 1 extract and stage 2 resolve bridge contracts |
| `internal/embed` | local embedding client and lexical/vector retrieval helpers |
| `internal/hermes` | Hermes provider adapter semantics and lifecycle mapping |
| `internal/mcp` | MCP tool surface that calls the same core semantics |
| `internal/httpapi` | thin HTTP transport adapter from routes to core requests |
| `internal/store` | storage interfaces that preserve idempotency, provenance, and scope separation |
| `internal/store/postgres` | PostgreSQL implementation of storage contracts |
| `internal/db` | PostgreSQL pool construction and pgvector registration |
| `internal/config` | environment and file-backed runtime configuration |
| `migrations` | PostgreSQL schema migrations |
| `tests` | cross-package smoke, integration, and replay tests |
| `tools` | repository maintenance tools such as header checks |
| `.agents` | Codex skills and shared agent assets |
| `pkg` | reserved for stable reusable packages only; keep empty until an API is proven |

Package dependencies should flow inward:

```text
cmd/* -> transport/adapters -> kernel -> ingest/recall -> store
cmd/worker -> internal/worker -> reasoning/graph/embed -> store
httpapi/hermes/mcp -> core service contract -> kernel
```

Handlers and adapters must stay thin. Product rules belong in core services and
their domain-specific packages, not in `cmd/*` or transport handlers.

## Consequences

- Empty product packages get package documentation now, so future work lands in
  the intended boundary instead of growing inside entrypoints.
- `pkg` remains intentionally empty until VibeGravity has a stable public
  library surface.
- New architecture packages require an ADR when they change the runtime model,
  scope model, queue model, or reasoning/apply boundary.
- Transport packages must not redefine memory semantics for Hermes, MCP, or
  HTTP independently.

## Impact on Hermes-first Roadmap

This layout lets Hermes use `prefetch()` and `sync_turn()` first while keeping
MCP and future clients aligned with the same core contract. It also leaves room
for degraded modes: recall can use existing profile, notes, plans, lexical
search, and previous memory even when Codex or local embeddings are unavailable.
