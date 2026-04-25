# VibeGravity Project Review

Date: 2026-04-25
Reviewer: Codex
Scope: live checkout at `/Users/parker/Documents/VibeGravity`, excluding `.git` internals. Binary build outputs, `.DS_Store`, `.omx` local telemetry/state, and prior Hermes run logs are inventoried but not embedded as raw GPT-Pro material.

## Executive Verdict

VibeGravity is directionally coherent and much further than a scaffold. The current repo has a Go-first shared memory kernel with a clear `VibeGravityService` contract, HTTP/MCP/Hermes-facing adapters, a PostgreSQL store, idempotent raw ingest, typed recall with trust metadata, correction/timeline/explain surfaces, mocked schema-first Codex runners, worker backlog visibility, and local eval gates.

I would not call it V1-ready yet. The strongest product line is now correct: **Hermes Memory, powered by VibeGravity**. The riskiest gap is no longer product framing; it is proving the trust loop against the real PostgreSQL schema and real external runtimes. In particular, I found one blocking DB-contract issue that can break `CorrectMemory` supersession in real Postgres.

## Findings

### P0. Operator correction supersession likely fails against real PostgreSQL foreign keys

Evidence:

- `migrations/000002_create_core_tables.up.sql:136` defines `memory_edges.created_by_job_id TEXT REFERENCES ingest_jobs (id)`.
- `migrations/000002_create_core_tables.up.sql:152` defines `memory_trace.reasoning_job_id TEXT REFERENCES ingest_jobs (id)`.
- `internal/kernel/service.go:567` sets correction trace `ReasoningJobID` to `"correction:" + correction.ID`.
- `internal/kernel/service.go:580` sets correction edge `CreatedByJobID` to `"correction:" + correction.ID`.

Why it matters: `CorrectMemory` first writes a raw correction event and `memory_corrections` row, then calls `CreateMemoryWithTraceAndUpdateEdge`. On a real DB, the replacement memory trace and updates edge reference synthetic IDs that do not exist in `ingest_jobs`, so the transaction should violate FK constraints. Existing tests use fake stores for kernel correction behavior and source-shape tests for correction persistence, so this can pass `go test ./...` while failing in production.

Recommended fix: either allow correction provenance fields to be nullable/non-FK for operator-correction traces and edges, or create a real deterministic correction/apply job row and use that job ID. Then add a live Postgres or migration-contract test that specifically exercises `CorrectMemory` through the real store.

### P1. MCP tool schemas advertise missing required fields for core-required calls

Evidence:

- `internal/mcp/protocol.go:179-190` makes `prefetch` / `recall_preview` require only `tenant_id` and `workspace_id`, but `internal/recall/assembler.go` requires `session_id` and `actor_id` for real prefetch.
- `internal/mcp/protocol.go:189-196` marks `sync_turn.idempotency_key` optional, but `internal/ingest/service.go:136-142` requires it.
- `internal/mcp/protocol.go:261-268` marks `correct_memory.idempotency_key` optional, but `internal/kernel/service.go:335-341` requires it.
- `internal/mcp/protocol.go` also leaves `owner_entity_id` optional for `add_note` / `create_plan`, while `internal/kernel/service.go:185-190` and `internal/kernel/service.go:221-227` require it.
- `internal/mcp/protocol_test.go:58` currently bakes in the too-short `recall_preview` required list.

Why it matters: MCP discovery is the first Hermes-facing protocol path. A client following `tools/list` can send apparently valid tool calls that then fail service validation. That weakens the trust-loop integration story even though the core service behavior is stricter.

Recommended fix: align schemas with DTO/service validation or relax service defaults intentionally. Add protocol tests that use the real service validation path, not only a fake service.

### P1. Update replay idempotency does not fully enforce payload mismatch conflicts

Evidence:

- `docs/review-packets/agent-c-update-memory-lineage-spec.md:81-84` says the same `reasoning_job_id` + `operation_id` with different payload/target must fail as an idempotency conflict.
- `docs/review-packets/agent-c-update-memory-lineage-spec.md:158-160` repeats that replaying the same operation with different payload, target, edge kind, or raw event IDs must fail.
- `internal/store/postgres/memories.go:149-154` returns idempotent success when `completedUpdateAlreadyApplied` is true.
- `internal/store/postgres/memories.go:208-244` checks tenant/workspace/scope/group/owner/status/trace/edge presence, but not replacement text, fingerprint, confidence, raw event IDs, applied operation JSON, or operation ID.
- `internal/store/postgres/memories.go:425-440` upserts trace on `memory_id` conflict, so create/extend retry with the same deterministic memory ID can overwrite trace evidence instead of surfacing a mismatch.

Why it matters: the retry path protects counts, but not all semantic evidence. If a deterministic job is retried with divergent Stage 2 output, the store can accept an existing update as success or overwrite trace state. That is exactly the kind of silent drift the contract is trying to avoid.

Recommended fix: store or derive a durable operation identity and compare payload hash, target, edge kind, raw event IDs, and applied operation JSON before treating retry as success. Add negative replay tests for changed text/raw events/target.

### P2. Current green gates are strong locally but do not cover the highest-risk real-runtime path

Evidence:

- `go test ./...`, `make eval`, `make lint`, and `make check-headers` all pass in this checkout.
- The Postgres concurrency test is guarded by `VIBEGRAVITY_DB_URL` and skips without a live database.
- Hermes provider and MCP tests use in-repo adapters/fakes; real Hermes runtime roundtrip remains unverified.
- Worker still uses `MockCodexJSONClient`; real Codex execution is intentionally not enabled.

Why it matters: local deterministic gates are valuable and should stay. But the next readiness claim needs one live Postgres correction/timeline/explain flow and one real stdio MCP smoke flow before calling the trust loop reliable.

Recommended fix: add a small opt-in integration target that migrates a scratch Postgres DB and runs `sync_turn -> process/replay or seed memory -> prefetch -> correct_memory -> explain -> timeline -> prefetch` through real stores.

## What Is Working Well

- Product direction is now sharper: V1 is a Hermes Memory trust loop, not a vague memory platform.
- Raw events, derived memories, traces, corrections, notes, plans, documents, and jobs are conceptually separated.
- The `VibeGravityService` interface is stable enough to keep HTTP, MCP, and Hermes adapters thin.
- Recall blocks now carry scope/source/status/freshness metadata, which directly supports the product promise.
- Worker failure handling distinguishes retryable failures from deterministic unsupported work that should become `blocked`.
- Graph apply is conservative: create, extend, update are open; archive/profile/session/plan deltas are still rejected rather than half-applied.
- The eval gate is meaningful for local regression: replay, correction-shaped supersession, group-shared stop-line, outage/recovery, and demo trust-loop scenarios are all represented.

## Project Status

Implemented enough to review as a real backend:

- Go commands: server, worker, CLI.
- Core DTOs and service contract.
- HTTP routes for v1 endpoints.
- PostgreSQL migrations and store implementations.
- Hot-path raw ingest and queued jobs.
- Typed recall with notes, plans, profiles, summaries, memories, documents, freshness metadata.
- Store-backed graph apply for `create_memory`, `extend_memory`, and `update_memory`.
- Append-safe correction intake and attempted correction supersession.
- Timeline and explain-memory trust surfaces.
- MCP stdio protocol server and Hermes provider adapter.
- Golden/demo/worker backlog evals.

Not yet V1 complete:

- Real Codex API client, prompt builder, retry policy, and operator-facing failure mode.
- Live Hermes runtime packaging/roundtrip.
- Live Postgres trust-loop integration coverage for correction supersession.
- Full profile/session/plan-delta apply semantics.
- Group-shared writes and group-aware notes/plans.
- Production deployment/ops story.

## Verification Run

Commands run on 2026-04-25:

```text
go test ./...                                      PASS
make eval                                          PASS
make lint                                          PASS
make check-headers                                 PASS
git diff --check                                  PASS
```

`make eval` passed the golden replay scenarios and Hermes Memory demo eval in this checkout.

## GPT-Pro Packet Files

The nine GPT-Pro material files are:

1. `01_gpt_pro_context_and_reading_order.md`
2. `02_gpt_pro_core_config_and_commands.md`
3. `03_gpt_pro_ingest_recall_kernel_http.md`
4. `04_gpt_pro_reasoning_graph_worker_eval.md`
5. `05_gpt_pro_postgres_store_and_migrations.md`
6. `06_gpt_pro_external_surfaces_hermes_mcp.md`
7. `07_gpt_pro_tests_and_golden_scenarios.md`
8. `08_gpt_pro_docs_plans_adrs_review_packets.md`
9. `09_gpt_pro_inventory_verification_and_risk_map.md`

Review prompt:

- `10_gpt_pro_review_prompt.md`

## Source Review

- Estimated source: live in-repo VibeGravity source, plans, ADRs, review packets, and local test/eval outputs.
- Suspected license: project-internal original material unless otherwise noted in repo policy docs.
- Similarity risk: low; this review and generated packet are original summaries and verbatim aggregation of local project files for review.
- Human review required: yes. The P0 finding changes DB/provenance semantics and should be reviewed before implementation.
