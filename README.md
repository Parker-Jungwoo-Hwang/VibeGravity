# VibeGravity

VibeGravity is the agent memory engine behind **Hermes Memory**.

The V1 promise is simple: Hermes remembers the right project context across
sessions, shows why it remembered it, and lets the operator fix memory once.

VibeGravity is not a chat UI, standalone coding agent, raw transcript archive,
or generic vector database. It records agent activity, derives scoped structured
memory, and returns compact recall before future work.

## What It Does

- Accepts agent turns through `sync_turn()` and stores immutable raw events.
- Derives memories in a background worker instead of slowing the hot path.
- Builds typed, budget-aware recall packs through `prefetch()`.
- Keeps `agent_private`, `workspace_shared`, `group_shared`, and
  `session_scratch` memory separate.
- Preserves provenance through `memory_trace`.
- Lets operators inspect, correct, supersede, and explain memory.
- Exposes the same core semantics through HTTP, MCP, and Hermes-facing adapter
  surfaces.

## Current Status

The Go-first foundation is active. The repo has runnable core contracts,
PostgreSQL migrations, HTTP handlers, sync and prefetch paths, note, plan, and
document surfaces, a store-backed graph apply path, worker recovery behavior,
Hermes and MCP adapter skeletons, deterministic evals, and trust-loop tests.

V1 is not complete yet. The current priority is DB and protocol correctness for
the Hermes Memory trust loop: recall preview, explain and timeline, correction,
correction-driven supersession, scope visibility, idempotent replay, and honest
degraded recall through real PostgreSQL and external protocol paths.

## Architecture

```text
Hermes / MCP / HTTP client
        |
        v
VibeGravity Core
        |
        +-- PostgreSQL canonical store
        +-- ingest_jobs queue
        +-- Recall assembler
        +-- Graph apply engine
        |
        v
Background worker
        |
        +-- local embedding runtime
        +-- Codex Stage 1 extract
        +-- Codex Stage 2 resolve
```

The worker pipeline is:

```text
local embeddings -> neighborhood retrieval -> Codex Stage 1 extract
-> Codex Stage 2 resolve -> apply engine
```

Local model use is embedding-first in V1. Text interpretation and graph
operations are Codex-first, schema-first, and structured JSON only.

## Repository Map

```text
cmd/server      HTTP API entrypoint
cmd/worker      background worker entrypoint
cmd/cli         doctor, eval, jobs, Hermes, and MCP commands
internal/core   service contract and domain types
internal/ingest sync_turn write path
internal/recall prefetch assembler
internal/graph  reasoning/apply boundary and graph writes
internal/kernel core service composition
internal/store  PostgreSQL store and persistence interfaces
internal/hermes Hermes provider adapter
internal/mcp    MCP surface and stdio protocol server
migrations      PostgreSQL schema migrations
tests           integration and golden test assets
plans           controlling product and architecture plans
docs            ADRs, policies, and review packets
```

## Quick Start

Install dependencies:

```bash
make setup
```

Build:

```bash
make build
```

Run the deterministic local gate:

```bash
go test ./...
make eval
make lint
make check-headers
git diff --check
```

Run the server and worker:

```bash
make dev-server
make dev-worker
```

The CLI currently exposes:

```text
doctor    Check system configuration and dependencies
eval      Run deterministic quality evals
hermes    Print Hermes bootstrap commands
jobs      Inspect and recover worker jobs
mcp       Serve the VibeGravity MCP protocol
```

## PostgreSQL Gate

PostgreSQL is the canonical store. SQLite and in-memory behavior are useful for
tests and local development, but they are not enough to declare the trust loop
ready.

Use a scratch database for the live integration gate:

```bash
createdb vibegravity_integration
export VIBEGRAVITY_DB_URL='postgres://localhost:5432/vibegravity_integration?sslmode=disable'
migrate -path migrations -database "$VIBEGRAVITY_DB_URL" up
make integration-postgres
```

If `VIBEGRAVITY_DB_URL` is unset, `make integration-postgres` exits with an
explicit skip message.

## HTTP API

The V1 surface starts with:

| Method | Path | Purpose |
|---|---|---|
| `POST` | `/v1/prefetch` | Build recall pack |
| `POST` | `/v1/sync-turn` | Record turn |
| `POST` | `/v1/documents` | Add document |
| `POST` | `/v1/search/memories` | Search memories |
| `POST` | `/v1/search/documents` | Search documents |
| `POST` | `/v1/notes` | Add note |
| `POST` | `/v1/plans` | Create plan |
| `PATCH` | `/v1/plans/{id}` | Update plan |
| `POST` | `/v1/memory/correct` | Correct memory |
| `GET` | `/v1/memory/{id}/explain` | Explain provenance |
| `GET` | `/v1/timeline` | Read timeline |

## Core Invariants

- Raw events and derived memories stay separate.
- Every write path is idempotent.
- Every memory has explicit scope.
- Every memory has provenance.
- Recall is budget-aware.
- Human correction is first-class.
- `updates` can only target one latest memory at a time.
- Group-shared memory requires valid membership.
- Profiles are rebuildable from raw events, memories, and edges.

## Reading Order

Start here when changing product or runtime behavior:

1. `plans/00_read-this-first_for-building-agents.md`
2. `plans/01_rfp_vibegravity_hermes-first.md`
3. `plans/02_product-contract_and_direction.md`
4. `plans/03_target-architecture_codex-first.md`
5. `plans/05_runtime-contracts_ingest-recall-apply.md`
6. `plans/06_data-model_and_storage-invariants.md`
7. `PLANS.md`

For verification details, read `tests/README.md`. For current review context,
start with the active packets listed in `PLANS.md`.

## Source Review

Estimated source: written from the repository's own `AGENTS.md`, `PLANS.md`,
controlling plan documents, `Makefile`, and local command output.

Suspected license risk: low. This README does not reproduce external project
code or distinctive third-party implementation text.

Similarity risk: low. No external structured snippet of 10 or more consecutive
lines was used.

Human review required: recommended for product wording and release-readiness
claims before publishing publicly.
