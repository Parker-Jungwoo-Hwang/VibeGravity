# PLANS.md

## Current Work Pack

Hermes Memory trust loop and first-customer integration.

## V1 Product Promise

V1 is now framed as **Hermes Memory, powered by VibeGravity**.

The first release must prove one felt outcome:

> Hermes remembers the right project context across sessions, shows why it
> remembered it, and lets the operator fix memory once.

`VibeGravity` remains the engine and internal architecture name. Public and
first-customer language should lead with Hermes continuity, correction, and
proof rather than "shared memory kernel".

## Current State

The Go-first foundation is now beyond the original Work Pack 01 skeleton.
The repo has runnable core contracts, PostgreSQL migrations, HTTP handlers,
sync/prefetch paths, note/plan/document surfaces, store-backed graph apply for
`create_memory` and safe `extend_memory`, worker blocked-job handling, and
schema-first reasoning stubs.

The project is not V1-complete. Safe `update_memory` transaction semantics,
correction supersession, the first in-repo Hermes provider adapter, Hermes MCP
bootstrap output, real stdio MCP protocol serving, narrow graph replay eval
gates, deterministic mocked Codex outage / worker backlog recovery eval gates,
read-only worker backlog metrics, group-shared membership filtering, and
operator-visible recall freshness degradation now exist, but real Codex
execution, custom Hermes memory provider registry packaging, full session
replay, and production ops remain.

Documents and rich dreaming remain engine capabilities, but they are no longer
the V1 product headline. V1 should sell the trust loop: recall preview,
explain/timeline, correction, supersession, visible scope, and degraded-status
truthfulness.

Documents are supporting context for recall and Stage 2 reasoning, not the V1
product promise. Dreaming is a quality and maintenance layer, not the thing that
makes V1 feel real to a Hermes operator. The product earns trust when Hermes can
remember, explain, correct, supersede, and show stale/degraded state honestly.

V1 is not ready until the correction trust loop is proven against real
PostgreSQL and the external protocol paths that Hermes will actually use. In
practice, that means DB/protocol correctness comes before new feature breadth.
The local push-readiness follow-up has closed the known correction DB contract
gaps: `correction_apply` is allowed by the fresh migration, existing DBs get an
`updates` index repair migration, correction idempotency now validates replay
evidence before applying, correction artifacts are marked `applied` with the
supersession transaction, and generated `bin/` binaries are not tracked.

## Active Review Packet

- `docs/review-packets/hermes-memory-trust-loop-product-pivot.md`
- `docs/review-packets/operator-visible-degraded-recall-freshness.md`
- `docs/review-packets/gpt-pro-followup-contract-alignment.md`
- `docs/review-packets/gpt-pro-followup-product-contract-alignment.md`

## Next Concrete Slice

Prove DB/protocol correctness for the Hermes Memory trust slice before adding
new product features.

Goal:

- Treat VibeGravity as the engine for Hermes Memory: an agent memory engine, not
  a chat app, raw transcript archive, or vector database product.
- Lock P0 correction provenance: raw correction event, append-safe correction
  artifact, replacement memory, mandatory trace, `updates` edge, and prior
  supersession must remain explainable and idempotent.
- Lock MCP schema correctness for the trust-loop tools so Hermes-facing clients
  can discover and send the same tenant/workspace/actor/memory/evidence fields
  the core service validates.
- Keep replay idempotency evidence-safe: retries must not duplicate memories,
  traces, edges, or correction artifacts, and stale evidence must not be hidden.
- Add or preserve live PostgreSQL integration gates for the correction and
  lineage paths before treating SQLite/in-memory checks as enough.
- Protect stop-lines: do not broaden into document-product promises, dreaming as
  the V1 value story, real Codex default enablement, or custom Hermes memory
  provider packaging until the DB/protocol gates pass.
- Re-run `go test ./...`, `make lint`, `make check-headers`, and
  `git diff --check` after implementation slices.

Recently completed:

- `internal/eval` now runs deterministic golden recall scenarios from
  `tests/golden/replay_eval.json`, with `cli eval golden` and `make eval` as the
  first quality regression gate.
- `internal/eval` now also runs narrow graph replay scenarios through the real
  store-backed apply engine, checking `update_memory` retry idempotency,
  correction-shaped supersession recall, mandatory trace/edge counts, prior
  memory supersession, and the current `group_shared` write stop-line.
- `internal/eval` now runs worker backlog scenarios through the real
  `worker.Processor` and `graph.StoreBackedApplyEngine` with mocked Stage 1 and
  Stage 2 outage controls. The gate proves transient reasoning failure retries
  without derived graph side effects, recovery writes only after structured
  reasoning succeeds, replay remains idempotent for memory/trace/edge rows, and
  unsupported deterministic apply work becomes blocked instead of retrying
  forever.
- `cli jobs metrics [--window D] [--tenant ID] [--workspace ID]` now reports
  read-only operator backlog visibility: total queued, ready queued, running,
  failed, blocked, and complete counts, retryable queued attempts, oldest ready
  queued age, oldest running job age, drain rate, and recovery ETA when enough
  completed-job history exists. It does not claim, requeue, fail, complete, or
  unblock jobs.
- `update_memory` now writes a replacement memory, mandatory trace, `updates`
  edge, and prior-memory supersession inside one PostgreSQL transaction. The
  path locks and verifies the target as active/latest, rejects scope/owner
  boundary changes, and treats deterministic successful retries as idempotent.
- Operator blocked-job recovery exists through `cli jobs blocked [--limit N]`
  and `cli jobs requeue-blocked <job_id>`.
- `internal/hermes.Provider` maps Hermes lifecycle hooks to core `Prefetch` and
  `SyncTurn`, renders typed recall context, exposes the minimum tool list, and
  has mocked lifecycle tests.
- `internal/mcp.Surface` maps MCP-style tool names to the same core service
  calls and has mocked tool delegation tests.
- `internal/mcp.Server` serves `initialize`, `notifications/initialized`,
  `ping`, `tools/list`, and `tools/call` over newline-delimited MCP stdio JSON-RPC.
- `cli mcp serve --stdio` starts the real MCP protocol server, and
  `cli hermes bootstrap` prints the `hermes mcp add ... --args mcp serve --stdio`
  registration plus `hermes mcp test` verification command.
- `POST /v1/documents` now uses an atomic document+chunk store path.
- `/healthz` returns `503` for a missing DB pool instead of panicking in embedded/test surfaces.
- `CorrectMemory` now validates, records an idempotent raw correction event,
  writes an append-safe `memory_corrections` artifact, then applies the
  correction text as a replacement memory with mandatory trace, an `updates`
  edge, and prior-memory supersession while preserving the original memory
  trace.
- `GetTimeline` now parses timeline query parameters and returns a read-only, scope-aware timeline over memories, traces, and correction artifacts.
- `Prefetch` now consumes read-only worker backlog freshness signals and marks
  recall meta plus derived recall blocks stale when queued, retry, or long-running
  worker state means stored memory/profile/session-summary context may lag behind
  raw events.

## After That

1. Turn the mocked outage/backlog eval into full session replay metrics before
   real Codex is enabled by default.
2. Build the 5-minute Hermes Memory demo: project rule, active plan, wrong
   memory, correction, supersession, explain/timeline, and private/shared scope
   check.
3. Turn the printed Hermes MCP bootstrap into an install/package command once
   the distribution format is decided.
4. Add real Hermes runtime roundtrip tests against a configured local database.

## Done Gates

- Code paths opened by the service contract have tests.
- Unsupported deterministic graph work blocks jobs instead of retrying forever.
- Raw events and derived memories remain separate.
- Agent-private retrieval always requires owner matching.
- Every operator-visible memory/recall path exposes scope and provenance.
- Corrected memory changes the next relevant recall and suppresses the old row.
- Degraded recall is labeled instead of presented as fresh memory.
- Source provenance and code header checks pass.
- Docs and review packets match the current code state.
