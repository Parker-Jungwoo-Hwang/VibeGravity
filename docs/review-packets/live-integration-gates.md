# Live Integration Gates

## Summary

Added an explicit `make integration-postgres` target and local operator docs for
running live PostgreSQL trust-loop checks without changing the default local
development gate.

## Finding or slice fixed

Local deterministic gates are necessary but not sufficient for VibeGravity's
current risk profile. `go test ./...`, `make eval`, `make lint`,
`make check-headers`, and `git diff --check` can pass while the live
PostgreSQL path that enforces row locks, foreign keys, transaction rollback, and
extension-backed schema behavior remains untested.

The new target keeps that concern explicit:

- no `VIBEGRAVITY_DB_URL`: print a clear skip message and exit successfully;
- `VIBEGRAVITY_DB_URL` set: run live Postgres tests in
  `./internal/store/postgres` plus the DB-backed smoke tests in `./tests`;
- normal `go test ./...` remains deterministic and must not require a live DB.

## Files changed

- `Makefile`
- `tests/README.md`
- `docs/review-packets/live-integration-gates.md`

## Tests run

Completed in this slice:

- `env -u VIBEGRAVITY_DB_URL make integration-postgres` - passed; printed the
  expected skip message and scratch-DB instructions.
- `make check-headers` - passed.
- `git diff --check` - passed.

Blocked by concurrent active lanes:

- `go test ./...` - failed in `internal/kernel` because
  `CorrectMemory` currently returns `not implemented: correct memory`; that
  implementation/test surface is claimed by `codex-agent1-correction-provenance`.
- `go test ./internal/store/postgres ./tests` - failed in
  `TestPostgresReplaySourceContractsRequireFullEvidenceComparison`; the replay
  idempotency test/code surface is claimed by `codex-agent4-replay-idempotency`
  and `codex-agent5-evidence-safe-replay`.
- `make lint` - failed on gofmt for `internal/store/postgres/jobs.go`, which is
  claimed by `codex-agent1-correction-provenance`.

If a live scratch DB is available:

- `export VIBEGRAVITY_DB_URL='postgres://localhost:5432/vibegravity_integration?sslmode=disable'`
- `migrate -path migrations -database "$VIBEGRAVITY_DB_URL" up`
- `make integration-postgres`

## Live gate scope

The live PostgreSQL trust-loop gate should cover these behaviors as guarded tests
instead of prose-only claims:

- correction supersession foreign-key safety for replacement memory, mandatory
  `memory_trace`, `updates` edge, and prior-memory supersession;
- update concurrency where exactly one active/latest replacement wins and losing
  workers leave no dangling memory, trace, or edge rows;
- replay idempotency for deterministic graph operations;
- `explain_memory` provenance queries against real trace data;
- read-only `timeline` behavior over memories, traces, and correction artifacts;
- recall/search suppression for corrected or superseded memory.

Current automated live coverage is narrower: the existing guarded Postgres
concurrency test verifies real row-lock and direct-target `updates` behavior,
and `./tests` verifies a DB-backed health smoke. The target is the integration
home for the remaining trust-loop tests as they land.

## Remaining risks

- This slice does not automate scratch database creation or migrations. The
  exact manual migration command is documented in `tests/README.md`.
- The live target currently depends on tests being self-guarded and safe for a
  scratch database. Do not point `VIBEGRAVITY_DB_URL` at shared or production
  data.
- Correction supersession, explain, timeline, replay idempotency, and
  recall/search suppression still need live DB tests before the trust loop can
  be called fully live-verified.

## Source Review

- Estimated source: in-repo VibeGravity Makefile, tests, plans, and review
  packets.
- Suspected license: project-internal original material.
- Similarity risk: low; implementation is a small original Make target and
  project-specific documentation.
- Human review required: yes, before treating the live gate as release
  acceptance.
