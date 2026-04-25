# Current State Review and Next Agent Handoff

Date: 2026-04-24
Scope: product planning docs, current Go implementation, Work Pack 03 review packets, and next executable slice.

## Executive Summary

VibeGravity is past the pure foundation stage but not yet V1-complete.

The product direction in `plans/` is coherent: VibeGravity is a Hermes-first shared memory kernel, not a chat UI or generic agent runtime. The implementation now matches that direction in its broad shape: Go-first server/worker/CLI, PostgreSQL store, schema-first reasoning boundary, idempotent ingest, typed recall, manual note/plan/document APIs, and a conservative graph apply path.

The main risk now is not product confusion. The main risk is widening semantics before the transactional write boundaries and operator-visible recovery paths are finished.

## Planning Doc Review

The durable product documents remain the right source of truth:

- `plans/00_read-this-first_for-building-agents.md`: product identity and implementation order.
- `plans/01_rfp_vibegravity_hermes-first.md`: acceptance criteria and Hermes-first obligations.
- `plans/02_product-contract_and_direction.md`: non-negotiable invariants.
- `plans/03_target-architecture_codex-first.md`: API/worker/Codex/local embedding split.
- `plans/05_runtime-contracts_ingest-recall-apply.md`: API, worker, Stage 1/2, apply, correction, and failure contracts.
- `plans/06_data-model_and_storage-invariants.md`: canonical PostgreSQL invariants.
- `plans/07` through `plans/11`: work pack sequencing.

One stale planning issue was fixed in this pass: root `PLANS.md` still said the current work pack was Work Pack 01 even though the repo now contains Work Pack 02/03 implementation slices. It now points to this review packet and the next concrete slice.

## Current Development State

Implemented or scaffolded:

- Core 11-method `VibeGravityService` contract.
- HTTP routes for the v1 surface.
- `sync_turn()` hot path: validate, write raw events, enqueue process job.
- `prefetch()` typed recall assembler with notes, plans, profiles, session summaries, memories, and documents in degraded mode.
- PostgreSQL store for raw events, jobs, memories, traces, edges, notes, plans, documents, profiles, and session summaries.
- Manual surfaces for memory search, document search, add note, create plan, update plan, add document, and explain memory.
- Atomic document ingestion: document row upsert and chunk replacement now share a single store transaction.
- `/healthz` returns 503 for a missing DB pool instead of panicking in embedded or test surfaces.
- Worker path with raw event bundle validation, blocked-job handling for deterministic unsupported apply work, and aggregate pass reporting.
- Schema-first reasoning interfaces with safe stub Stage 1 and Stage 2 runners.
- Disabled-by-default Codex JSON bridge boundary and strict JSON validation tests.
- Store-backed apply for `create_memory`, safe `extend_memory`, and
  `update_memory` with mandatory trace and updates-edge supersession.
- Human correction now records intent and applies operator-driven replacement
  memory supersession.
- Narrow eval coverage now replays `update_memory` and correction-shaped graph
  supersession through the store-backed apply engine, including deterministic
  retry/idempotency, mandatory trace/edge counts, later recall, and the current
  `group_shared` write stop-line.
- Narrow worker backlog eval coverage now runs mocked Stage 1/Stage 2 outage
  scenarios through the real worker processor and store-backed apply engine,
  proving retry without graph side effects, recovery/replay idempotency for
  memory/trace/edge rows, and blocked state for deterministic unsupported apply
  work.
- `cli jobs metrics [--window D] [--tenant ID] [--workspace ID]` now provides
  read-only operator visibility into total queued, ready queued, running,
  failed, blocked, and complete job counts, retryable queued attempts, oldest
  ready queued age, drain rate, and recovery ETA when calculable.
- Review packets for Work Pack 03 coordination.

Still intentionally incomplete:

- Real Codex execution.
- Group-shared membership-aware retrieval and writes.
- Profile merge, plan delta writes, and production dreaming quality.
- External Hermes provider registry packaging.
- Full session replay, Codex outage/backlog metrics, real Codex outage handling,
  and production-grade ops.

## Code Review Findings

### Fixed. Document add could leave partial state if chunk write failed

File: `internal/kernel/service.go`, lines 111-116.

`AddDocument` now delegates to `AddDocumentWithChunks`, and the PostgreSQL store writes the document row plus chunk replacement in one transaction. Regression coverage asserts service-level atomic delegation and the store transaction shape.

### P2. `PLANS.md` was stale against the implementation phase

File: `PLANS.md`, prior lines 3-37.

The file still identified Work Pack 01 as current even though the repo has Work Pack 02/03 slices. This was likely to mislead the next agent into redoing foundation work or avoiding the real current risk area. This pass updated it.

### Fixed. Health check assumed `DBPool` was always non-nil

File: `internal/httpapi/router.go`, lines 63-68.

`Healthz` now returns 503 with a database-unavailable response when `DBPool` is missing. This keeps embedded and test surfaces from panicking.

### Fixed. Timeline and correction routes were only thin exposure

File: `internal/httpapi/router.go`, lines 237-291.

The routes are now exposed and backed by core behavior. `CorrectMemory` records a raw event and append-safe correction artifact, then applies the correction text as a replacement memory with operator-correction trace, an `updates` edge, and target supersession. `GetTimeline` parses `scopes`, `from`, `to`, and `limit`, then returns a read-only timeline over memories, traces, and correction artifacts with `agent_private` owner filtering and `group_shared` exclusion until membership filtering exists.

## Big-Picture Development Plan

1. Stabilize current API/store boundaries.
   - Make opened write paths transactional.
   - Keep idempotency, provenance, and scope checks visible in tests.

2. Finish manual/operator control paths.
   - correction recall integration and operator timeline polish
   - blocked job list/requeue UX

3. Finish safe graph semantics.
   - Target preflight for `extend_memory`
   - archive/supersede suppression

4. Enable real reasoning behind explicit configuration.
   - Keep mocked client tests first.
   - No local extractor fallback.
   - Real Codex only after failure/degraded behavior is explicit.

5. Connect first customer surfaces.
   - Hermes provider lifecycle.
   - MCP tools with the same core semantics.

6. Add quality gates.
   - Replay harness.
   - Golden scenarios.
   - Metrics and operator docs.

## Immediate Next Task

Turn the mocked outage/backlog eval gate and read-only job metrics into
operator-facing degraded recall metadata before enabling real Codex execution.

Why this next:

- `update_memory` retry, correction-shaped supersession recall, mandatory
  trace/edge shape, and the `group_shared` write stop-line are now covered by
  deterministic local graph replay scenarios.
- The mocked worker behavior now has a deterministic local gate and read-only
  CLI metrics. The next risk is making Codex freshness/degraded recall state
  visible in prefetch metadata, then eventually wiring a real Codex client
  behind the same boundary.

Acceptance criteria:

- Add explicit degraded recall metadata for Codex/backlog freshness loss.
- Keep mocked Stage 1/Stage 2 outage evals and read-only job metrics passing
  while wiring recall visibility.
- Keep transient reasoning failures retryable and deterministic unsupported
  apply work blocked without graph side effects.
- Preserve replay idempotency for memory, trace, and edge rows.
- Keep scope separation and group membership constraints visible in eval data.
- Preserve `go test ./...`, `make lint`, `make check-headers`, and
  `git diff --check` as required gates.
- `go test ./...`, `make lint`, `make check-headers`, and `git diff --check` pass.

## Next Agent Prompt

```md
You are continuing VibeGravity in `/Users/parker/Documents/VibeGravity`.

Read first:
- `AGENTS.md`
- `PLANS.md`
- `plans/00_read-this-first_for-building-agents.md`
- `plans/01_rfp_vibegravity_hermes-first.md`
- `plans/02_product-contract_and_direction.md`
- `plans/03_target-architecture_codex-first.md`
- `plans/05_runtime-contracts_ingest-recall-apply.md`
- `plans/06_data-model_and_storage-invariants.md`
- `docs/review-packets/current-state-and-next-agent-handoff.md`
- `docs/review-packets/next-agent-integration-fixes.md`

Task:
Turn the deterministic mocked outage/backlog eval gate into operator-visible
Codex outage and worker backlog recovery reporting.

Context:
- `create_memory`, safe `extend_memory`, and `update_memory` have store-backed write paths.
- `CorrectMemory` records append-safe correction intent and applies operator correction supersession.
- `GetTimeline` provides read-only operator visibility over memory/correction activity.
- `internal/eval` now runs graph replay scenarios for `update_memory` retry,
  correction-shaped supersession recall, mandatory trace/edge counts, and the
  `group_shared` write stop-line.
- `internal/eval` now runs worker backlog scenarios for mocked Stage 1 outage,
  mocked Stage 2 outage, recovery replay idempotency, and blocked unsupported
  apply work.

Implement only this scope:
- Add operator-visible backlog counts and recovery metrics without changing
  retry/block semantics.
- Preserve existing deterministic outage/backlog eval gates.
- Preserve existing validation floor and reject unsupported profile/session/plan
  deltas.
- Update docs if worker, eval, CLI, or store behavior changes.

Do not do:
- Do not implement real Codex calls.
- Do not implement archive behavior beyond existing supersession.
- Do not implement real external Hermes packaging or MCP protocol serving.
- Do not weaken source provenance, code header, or scope-separation rules.

Verification:
- `gofmt` on touched Go files
- `go test ./...`
- `make eval`
- `make lint`
- `make check-headers`
- `git diff --check`

Return:
- Files changed
- Tests and checks run
- Any remaining risks
- Source review: estimated source, suspected license, similarity risk, review required
```

## Verification For This Review Pass

Commands run:

- `go test ./...` — passed.
- `make eval` — passed.
- `make lint` — passed.
- `make check-headers` — passed.
- `git diff --check` — passed.

## Source Review

- Estimated source: in-repo VibeGravity product contracts, current code, and review packets.
- Suspected license: project-internal original documentation.
- Similarity risk: low.
- Review required: yes, because the next slice touches graph update and supersession semantics.
