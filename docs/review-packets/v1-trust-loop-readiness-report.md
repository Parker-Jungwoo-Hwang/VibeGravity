# V1 Trust Loop Readiness Report

Agent: codex-agent10-release-readiness-refresh
Date: 2026-04-25
Scope: release-readiness verification for Hermes Memory, powered by VibeGravity

## Executive Verdict

No-go for claiming V1 correction trust-loop readiness.

The repo is release-shaping in the right direction: all required local
deterministic gates passed after the other agent lanes released their claims.
The P0 correction provenance fix, live PostgreSQL test harness, MCP schema
alignment, replay idempotency hardening, scope/group_shared stop-line tests,
Codex disabled-by-default guardrails, and DB/protocol-first docs all appear to
have landed in the current worktree.

The blocker is proof, not intent. `VIBEGRAVITY_DB_URL` is unset in this
verification shell, so the canonical live PostgreSQL correction trust-loop gate
was skipped. Local SQLite, mock, SQL-shape, and deterministic eval evidence are
useful, but they are not sufficient to claim the real correction loop is ready.

## Readiness Claim

VibeGravity cannot yet claim correction trust-loop readiness.

It can claim local deterministic trust-loop gates are green as of this run. It
cannot claim live DB trust-loop readiness until `make integration-postgres`
runs against a migrated scratch PostgreSQL database and proves correction
intake, replacement memory, mandatory trace, `updates` edge, target
supersession, retry idempotency, explain/timeline evidence, and next-recall
suppression through the canonical store.

## Completed Fixes Observed

| Area | Status | Evidence |
|---|---|---|
| P0 correction provenance fix | Landed locally; not live-verified in this run | `docs/review-packets/correctmemory-postgres-provenance-fix.md`; `internal/kernel/service_test.go`; `internal/store/postgres/jobs.go`; `internal/store/postgres/memories.go`. |
| Live PostgreSQL correction trust-loop test | Landed but skipped here | `internal/kernel/correction_trust_loop_integration_test.go`; `Makefile` target `integration-postgres`; `tests/README.md`. |
| MCP schema and service validation alignment | Landed locally | `internal/mcp/protocol_test.go` and `internal/mcp/surface_test.go` assert trust-loop schema fields and validation behavior. |
| Evidence-safe replay idempotency | Landed locally | `internal/store/postgres/memories.go`, `internal/store/postgres/memories_test.go`, `internal/store/postgres/memories_replay_test.go`, and `make eval` replay scenarios passed. |
| Scope and group_shared stop-line tests | Landed locally | `internal/store/postgres/search_test.go`, `internal/store/postgres/scope_safety_test.go`, `internal/recall/scope_safety_test.go`, `internal/graph/store_apply_test.go`; `make eval` keeps `group shared graph write remains rejected` green. |
| Codex disabled-by-default and no-local-extractor guardrails | Landed locally | `internal/config/config_test.go`, `cmd/worker/main_test.go`, `cmd/cli/main_test.go`, `internal/recall/assembler_test.go`; real Codex remains opt-in and mocked by default. |
| DB/protocol correctness prioritized in docs | Landed locally | `PLANS.md`, `plans/02_product-contract_and_direction.md`, `plans/05_runtime-contracts_ingest-recall-apply.md`, `plans/06_data-model_and_storage-invariants.md`, `plans/10_workpack_hermes-provider-and-external-surfaces.md`, `plans/11_workpack_quality-ops-and-evals.md`. |

## Gates Run

| Gate | Command | Outcome | Gate Type |
|---|---|---|---|
| Coordination status | `.agents/coordination/agent-work.sh status` | PASS for coordinator timing: all other active claims were released before final gate run; only this report claim remained. | Coordination |
| Worktree inspection | `git status --short --branch` | PASS for inspection; worktree remains intentionally dirty from landed agent lanes. | Inspection |
| Repo tests | `go test ./...` | PASS. All packages passed or had no test files. | Local deterministic |
| Golden/demo eval | `make eval` | PASS. Golden eval and Hermes Memory demo eval passed, including correction replay, replay idempotency, degraded recall, and group_shared write rejection scenarios. | Local deterministic |
| Lint | `make lint` | PASS. Ran `/Users/parker/go/bin/golangci-lint run`. | Local deterministic |
| Headers | `make check-headers` | PASS. Output: `code header check passed`. | Local deterministic |
| Whitespace | `git diff --check` | PASS. No output. | Local deterministic |
| Live Postgres availability | `if [ -n "${VIBEGRAVITY_DB_URL:-}" ]; then ...` | `VIBEGRAVITY_DB_URL=unset`. | Live integration precheck |
| Live Postgres gate | `make integration-postgres` | SKIPPED. Output: `Skipping live PostgreSQL integration gate: VIBEGRAVITY_DB_URL is not set.` | Live integration skipped |

## Skipped Gates

| Gate | Reason | Release Impact |
|---|---|---|
| Live PostgreSQL correction trust loop | `VIBEGRAVITY_DB_URL` is unset. | Release blocker for any correction trust-loop readiness claim. |
| Real Hermes runtime roundtrip | No configured Hermes runtime plus live DB gate was available in this verification run. | Release blocker for Hermes-first V1 runtime readiness. |
| Real Codex execution | Intentionally disabled by current contract; tests stay mocked-only. | Acceptable for current safety posture, but not proof of production Codex behavior. |

## Risk Table

| Risk | Current Status | Release Impact | Next Action |
|---|---|---|---|
| Correction provenance | Local fix and tests present; live DB unverified. | No-go for readiness until FK/transaction behavior is proven in PostgreSQL. | Run migrated scratch DB gate and inspect replacement memory, trace, edge, job, correction artifact, and target supersession rows. |
| Live PostgreSQL coverage | `integration-postgres` target exists but skipped. | Primary blocker. | Export `VIBEGRAVITY_DB_URL` for a migrated scratch DB and run `make integration-postgres`. |
| MCP schema | Local schema/service tests pass. | Lower risk locally; still needs protocol runtime confidence with Hermes path. | Keep exact required-field tests and add live protocol smoke tied to the DB trust-loop fixture. |
| Replay idempotency | Local deterministic and SQL-shape coverage pass. | Medium until live replay cases run against PostgreSQL. | Include replay assertions in the live integration gate. |
| Scope leakage | Local tests cover private ownership and membership-gated group_shared memory reads. | Medium-low locally; live DB proof still useful. | Keep scope tests in local gate and add representative live DB rows. |
| group_shared writes | Local stop-line remains rejected. | Acceptable for current contract. | Do not enable group_shared writes before explicit membership/write semantics and live tests. |
| Codex default behavior | Real Codex disabled by default; mocked bridge remains default. | Acceptable and safer for this slice. | Keep real Codex opt-in until failure UX, retries, and stale/degraded visibility are live-tested. |
| Degraded visibility | Local eval and tests show stale/degraded metadata paths. | Medium until real backlog/worker behavior is observed. | Add live backlog/degraded smoke with real store state. |
| Hermes runtime coverage | Provider/MCP surfaces have tests, but no live Hermes runtime roundtrip ran here. | Release blocker for Hermes-first V1. | Run Hermes against the same scratch DB through MCP/provider path after DB gate passes. |

## Dirty Worktree Notes

I did not revert unrelated dirty work. These files were dirty after the landed
agent lanes and are treated as owned by their respective lanes unless this
report is the file:

- Modified: `PLANS.md`, `cmd/cli/main_test.go`, `cmd/worker/main_test.go`,
  `docs/review-packets/correctmemory-review-and-gettimeline-prep.md`,
  `docs/review-packets/evidence-safe-replay-idempotency-implementation.md`,
  `docs/review-packets/mcp-schema-service-contract-alignment.md`,
  `internal/config/config_test.go`, `internal/graph/store_apply_test.go`,
  `internal/mcp/surface_test.go`, `internal/recall/assembler_test.go`,
  `internal/store/postgres/memories.go`,
  `internal/store/postgres/memories_replay_test.go`,
  `internal/store/postgres/memories_test.go`,
  `internal/store/postgres/search_test.go`,
  `plans/02_product-contract_and_direction.md`,
  `plans/05_runtime-contracts_ingest-recall-apply.md`,
  `plans/06_data-model_and_storage-invariants.md`,
  `plans/10_workpack_hermes-provider-and-external-surfaces.md`,
  `plans/11_workpack_quality-ops-and-evals.md`.
- Untracked: `docs/review-packets/codex-default-and-degraded-mode-guardrails.md`,
  `docs/review-packets/correctmemory-postgres-provenance-fix.md`,
  `docs/review-packets/gpt-pro-followup-contract-alignment.md`,
  `docs/review-packets/replay-idempotency-contract-tests.md`,
  `docs/review-packets/scope-safety-and-group-shared-guardrails.md`,
  `internal/recall/scope_safety_test.go`,
  `internal/store/postgres/scope_safety_test.go`.
- This report: `docs/review-packets/v1-trust-loop-readiness-report.md`.

## Next Recommended Slice

Make the next slice live-DB-only and release-blocking:

1. Create and migrate a scratch PostgreSQL database.
2. Export `VIBEGRAVITY_DB_URL`.
3. Run `make integration-postgres`.
4. If it fails, fix only the live DB trust-loop path.
5. If it passes, extend the live gate to include `CorrectMemory -> ExplainMemory -> GetTimeline -> Prefetch` and record row-count assertions for correction artifact, replacement memory, trace, `updates` edge, target supersession, and idempotent retry.
6. After DB proof is green, add a Hermes runtime smoke against the same scratch DB.

## Source Review

- Estimated source: in-repo VibeGravity files, local command outputs, and prior
  VibeGravity project memory used only to preserve the known live-DB caution.
- Suspected license: project-internal original material.
- Similarity risk: low; this is original verification prose.
- Human review required: yes, because release readiness depends on live
  PostgreSQL and Hermes runtime evidence not available in this run.
