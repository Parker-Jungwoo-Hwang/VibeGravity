# V1 Trust Loop Readiness Report

Agent: codex-agent10-release-readiness
Date: 2026-04-25
Scope: release-readiness verification for Hermes Memory, powered by VibeGravity

## Executive verdict

No-go for claiming full correction trust-loop readiness, but local
push-readiness blockers from the follow-up code review are now addressed.

The repo is in a much better state than the prior review: local deterministic
gates pass, the P0 correction provenance fix appears present in the worktree,
MCP schema/service alignment appears present, replay/idempotency and stop-line
local gates are covered, and the product docs now frame V1 as Hermes Memory,
powered by VibeGravity.

The remaining blocker is live integration proof. `VIBEGRAVITY_DB_URL` is not set
in this verification shell, so the real PostgreSQL correction trust loop was not
exercised. Local green gates are useful and the repo can be prepared for GitHub
handoff, but they are still not enough to claim V1 correction trust-loop
readiness.

## Completed fixes observed

| Area | Status | Evidence |
|---|---|---|
| P0 correction provenance / FK safety | Present locally, not live-verified | `internal/core/job.go` adds deterministic `correction_apply` job IDs; `internal/store/postgres/jobs.go` adds `EnsureCorrectionApplyJob`; `internal/kernel/service.go` calls the job store before correction supersession; local tests pass. |
| `correction_apply` migration contract | Fixed locally, not live-verified | `migrations/000002_create_core_tables.up.sql` now allows `correction_apply`; `tests/migration_contract_test.go` guards it. |
| Existing DB `updates` index rollout | Fixed locally, not live-verified | `migrations/000004_fix_updates_edge_target_index.*.sql` drops/recreates `memory_edges_single_updates_target_idx` for already-migrated DBs. |
| Correction replay evidence safety | Fixed locally, not live-verified | `CorrectMemory` checks existing correction artifacts before active/latest rejection, validates same-key evidence, and PostgreSQL correction insert conflicts on mismatched evidence. |
| Correction artifact applied status | Fixed locally, not live-verified | PostgreSQL correction supersession now marks `memory_corrections.status = 'applied'` inside the same transaction as replacement memory, trace, edge, and target supersession. |
| Generated binary hygiene | Fixed locally | `bin/cli` is removed from Git tracking and `.gitignore` ignores `bin/`. |
| MCP schema/service alignment | Present locally | `internal/mcp/protocol.go` requires `idempotency_key` for `correct_memory`; `internal/mcp/protocol_test.go` asserts exact required fields for trust-loop tools. |
| Evidence-safe replay idempotency | Locally covered | `make eval` passes update replay, correction replay, and Stage 2 outage recovery replay scenarios; `internal/store/postgres/memories.go` has idempotent completed-update handling. |
| Stop-line contract tests | Present locally | `make eval` passes `group shared graph write remains rejected` and `unsupported apply work becomes blocked`; code still documents real Codex disabled by default and no local extractor path. |
| Product/doc alignment | Present locally | `PLANS.md`, `plans/02_product-contract_and_direction.md`, `plans/05_runtime-contracts_ingest-recall-apply.md`, and `plans/11_workpack_quality-ops-and-evals.md` use Hermes Memory trust-loop language and prioritize DB/protocol correctness. |
| Live integration gate home | Present but not exercised live | `Makefile` has `integration-postgres`; `tests/README.md` documents scratch DB setup and says current live automated coverage is still narrower than the full trust loop. |

## Remaining blockers

1. Live PostgreSQL correction trust loop is unverified in this run because
   `VIBEGRAVITY_DB_URL` is unset.
2. The current `make integration-postgres` target exists, but the docs say its
   current automated live coverage is still limited to Postgres concurrency and
   DB health smoke. Full correction supersession, explain, timeline,
   replay-idempotency, and recall/search suppression need live guarded tests.
3. Real Hermes runtime roundtrip remains unverified by these gates. Local Hermes
   provider/MCP tests are meaningful, but they do not prove a configured Hermes
   runtime session against a live DB.
4. Active coordination claims existed during verification, so this report treats
   unrelated dirty files and in-progress lanes as unverified rather than
   recommending any revert.

## Gates run

| Gate | Command | Outcome | Gate type |
|---|---|---|---|
| Repo tests | `go test ./...` | PASS | Local deterministic |
| Golden/demo eval | `make eval` | PASS; golden eval and Hermes Memory demo eval passed | Local deterministic |
| Lint | `make lint` | PASS; ran `/Users/parker/go/bin/golangci-lint run` | Local deterministic |
| Headers | `make check-headers` | PASS; `code header check passed` | Local deterministic |
| Whitespace | `git diff --check` | PASS; no output | Local deterministic |
| Live Postgres target availability | `make integration-postgres` | SKIP; printed `VIBEGRAVITY_DB_URL is not set` | Live integration skipped |

## Skipped gates

| Gate | Reason |
|---|---|
| Live PostgreSQL correction trust loop | `VIBEGRAVITY_DB_URL` is unset. |
| Real Hermes runtime roundtrip | No configured live DB/runtime gate was available in this verification run. |
| Real Codex execution | Intentionally disabled by current contract; tests remain mocked-only. |

## Risk table

| Risk | Current status | Release impact | Next action |
|---|---|---|---|
| P0 correction provenance | Locally fixed-looking; not live-verified | No-go until live DB confirms replacement memory, trace, edge, and job FK safety | Run a migrated scratch DB and add/execute a guarded `correct_memory -> explain -> timeline -> prefetch` live test. |
| MCP schema/service | Locally aligned | Lower risk after tests | Keep schema tests exact when adding/changing tools. |
| Replay idempotency | Local eval green; live DB incomplete | Medium risk | Add live replay idempotency assertions for deterministic graph operations. |
| Live Postgres coverage | Gate exists but skipped here and still incomplete | Release blocker | Populate `VIBEGRAVITY_DB_URL`, migrate scratch DB, run `make integration-postgres`, then broaden live tests. |
| Hermes runtime coverage | Provider/MCP tests green locally, no runtime roundtrip | Release blocker for Hermes-first V1 | Add configured Hermes runtime smoke against the same scratch DB. |
| Codex default behavior | Real Codex remains disabled by default | Acceptable for current safety posture | Do not enable by default until failure UX, retries, and operator visibility are proven. |
| `group_shared` safety | Local stop-line and membership-aware read gates are present | Acceptable locally; keep guarded | Keep `group_shared` writes rejected until membership/write semantics are explicit and live-tested. |

## Dirty worktree and coordination notes

The worktree is intentionally dirty from concurrent lanes. I did not revert or
normalize unrelated changes. At verification time, active coordination claims
included correction provenance, MCP schema, replay idempotency, external protocol
smoke, doc alignment, and live integration gate work. This report only claimed:

- `docs/review-packets/v1-trust-loop-readiness-report.md`

Unrelated dirty files should be reviewed by their owning lanes. Notable dirty
surfaces include `Makefile`, `PLANS.md`, `cmd/`, `internal/`, `migrations/`,
`plans/`, `.agents/coordination/`, `.agents/hermes-orchestration/`,
`docs/review-packets/`, and `tests/`.

## Next recommended slice

Build the live DB trust-loop gate before any new product feature work:

1. Create and migrate a scratch PostgreSQL database.
2. Export `VIBEGRAVITY_DB_URL`.
3. Add a guarded live test that seeds or writes an active memory, calls
   `CorrectMemory`, verifies the deterministic `correction_apply` job row,
   replacement memory, mandatory `memory_trace`, `updates` edge, prior
   supersession, `ExplainMemory`, `GetTimeline`, and next `Prefetch` suppression.
4. Run `make integration-postgres` plus the local deterministic gates.

Only after that passes should the project claim correction trust-loop readiness.

## Source Review

- Estimated source: in-repo VibeGravity files, local command outputs, and
  project memory notes for prior live-DB caution.
- Suspected license: project-internal original material.
- Similarity risk: low; this report is original verification prose.
- Human review required: yes, because release readiness depends on live
  PostgreSQL and Hermes runtime evidence not available in this run.
