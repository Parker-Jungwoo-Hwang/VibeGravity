# QA and Release Readiness Review

Date: 2026-04-26  
Reviewer: Agent 09  
Role: CTO review  
Scope: Tests, QA, reliability, observability, release process, and regression prevention.  
Repo: `/Users/parker/Documents/VibeGravity`

## Review Request

Agent 09 was asked to review the GitHub repository `VibeGravity` from a CTO
perspective, focused on whether the repo can ship changes without breaking
users.

The requested review questions were:

- Are there automated tests?
- Are core CLI commands tested?
- Are scripts tested?
- Is there CI?
- Is there linting or formatting?
- Is there a release checklist?
- Is the changelog accurate and useful?
- Is semantic versioning handled well?
- Can users report bugs easily?
- Are failures observable through logs or clear errors?

The review was read-only. No source files were edited as part of the review.

## Sources Reviewed

Primary source files and folders reviewed:

- `AGENTS.md`
- `CLAUDE.md`
- `PLANS.md`
- `Makefile`
- `.golangci.yml`
- `go.mod`
- `COMMIT_MESSAGE_RULES.md`
- `cmd/cli/main.go`
- `cmd/cli/main_test.go`
- `cmd/cli/hermes_bootstrap_stopline_test.go`
- `cmd/server/main.go`
- `cmd/worker/main.go`
- `cmd/worker/main_test.go`
- `internal/httpapi/router.go`
- `internal/httpapi/router_test.go`
- `internal/hermes/provider_test.go`
- `internal/mcp/protocol_test.go`
- `internal/mcp/stdio_smoke_test.go`
- `internal/mcp/surface_test.go`
- `internal/kernel/correction_trust_loop_integration_test.go`
- `internal/store/postgres/concurrency_integration_test.go`
- `internal/store/postgres/memories_replay_test.go`
- `internal/store/postgres/jobs.go`
- `tests/README.md`
- `tests/baseline_test.go`
- `tests/migration_contract_test.go`
- `tests/golden/replay_eval.json`
- `plans/00_read-this-first_for-building-agents.md`
- `plans/01_rfp_vibegravity_hermes-first.md`
- `plans/02_product-contract_and_direction.md`
- `plans/03_target-architecture_codex-first.md`
- `plans/05_runtime-contracts_ingest-recall-apply.md`
- `plans/06_data-model_and_storage-invariants.md`
- `plans/11_workpack_quality-ops-and-evals.md`
- `.agents/coordination/README.md`
- `.agents/coordination/WORK_PROGRESS.md`
- `.agents/hermes-orchestration/README.md`
- `docs/adr-001-migration-versioning.md`
- `docs/review-packets/v1-trust-loop-readiness-report.md`
- `docs/review-packets/push-readiness-review-fixes.md`

Searches were also run for:

- `.github/workflows`
- `CHANGELOG.md`
- release files
- version files
- issue templates
- `setup.py`
- `scripts/`
- `release-manager`
- `qa-engineer`
- `code-reviewer`

## Verification Commands

The following commands were run from `/Users/parker/Documents/VibeGravity`:

```bash
go test ./...
make eval
make lint
make check-headers
git diff --check
go build ./cmd/server ./cmd/worker ./cmd/cli
make integration-postgres
go test -count=1 ./...
```

Observed results:

- `go test ./...` passed.
- `go test -count=1 ./...` passed.
- `make eval` passed.
- `make lint` passed.
- `make check-headers` passed.
- `git diff --check` passed.
- `go build ./cmd/server ./cmd/worker ./cmd/cli` passed.
- `make integration-postgres` skipped because `VIBEGRAVITY_DB_URL` was not set.

The repository had no tracked `.github` files, no root `README.md`, no
`CHANGELOG.md`, no release checklist, no version file, no issue templates, no
`setup.py`, and no root `scripts/` directory observed during this review.

## Executive Verdict

VibeGravity has a credible local quality system, but it is not release-ready for
frequent user-facing shipping.
The automated Go test suite is real and broad, with 41 test-related files and
198 `Test*` functions observed in the checkout.
The strongest local gates are deterministic evals, linting, source header
checks, migration contract tests, storage contract tests, and protocol smoke
tests.
The biggest release blocker is that the canonical PostgreSQL gate exists but
skipped in this environment because `VIBEGRAVITY_DB_URL` is unset.
There is no CI, no changelog, no release checklist, no semantic versioning
surface, and no bug-reporting path.
Users and operators can see several clear errors through HTTP, CLI, and worker
logs, but observability is still mostly local logs plus CLI backlog metrics, not
production-grade telemetry.
My CTO decision is that the repo can support rapid internal development slices,
but it cannot yet support frequent user-facing releases without a release
system.

## Current Quality System

The repo has meaningful automated tests.
The test suite covers core service behavior, ingest, recall, graph apply,
reasoning boundaries, worker orchestration, HTTP transport, Hermes provider
dispatch, MCP protocol behavior, PostgreSQL query shape, migration contracts,
and deterministic golden/demo evals.

The most important local gates are documented in `tests/README.md`:

```bash
go test ./...
make eval
make lint
make check-headers
git diff --check
```

The Makefile also provides:

- `make build`
- `make test`
- `make eval`
- `make integration-postgres`
- `make lint`
- `make check-headers`
- `make setup`

The deterministic eval system is a strong asset.
`make eval` runs:

```bash
go run ./cmd/cli eval golden --path tests/golden/replay_eval.json
go run ./cmd/cli eval demo
```

Those evals cover important product regressions:

- pinned note and active plan priority;
- private scope separation;
- superseded memory suppression;
- degraded recall with profile and summary fallback;
- token budget truncation;
- update replay behavior;
- correction replay behavior;
- current `group_shared` graph-write stop-line;
- mocked Stage 1 and Stage 2 outage behavior;
- unsupported apply work becoming blocked;
- Hermes Memory demo recall, explain, correction, supersession, and private
  scope separation.

The live PostgreSQL gate is correctly separated from the default local gate.
`make integration-postgres` is documented to run database-backed tests when
`VIBEGRAVITY_DB_URL` is set, and to skip explicitly when it is not set.
That is a reasonable local-development pattern, but it is not enough for release
readiness unless CI runs the live gate against a scratch database.

## Questions Answered

### Are there automated tests?

Yes.

The repo has a substantial Go test suite.
I counted 41 test-related files and 198 `Test*` functions.
`go test -count=1 ./...` passed during this review.

The tests are not just superficial compile checks.
They cover service validation, recall behavior, graph apply behavior, worker
job handling, MCP/Hermes dispatch, HTTP transport behavior, migration contract
shape, and live-PostgreSQL-gated trust-loop paths.

The caveat is that the most important storage truth depends on live PostgreSQL,
and those tests skip unless `VIBEGRAVITY_DB_URL` is set.

### Are core CLI commands tested?

Partially.

The following CLI behavior has direct tests:

- `jobs blocked`
- `jobs metrics`
- invalid jobs metrics window handling
- `jobs requeue-blocked`
- invalid blocked-job limit handling
- requeue missing-job errors
- `mcp serve --stdio`
- `eval demo`
- `hermes bootstrap`
- database URL password masking

That is a good base for operator commands.
It proves several important paths do not require a real database in unit tests.

The gaps are:

- `doctor` is not directly exercised as a CLI command.
- `eval golden` is covered through `make eval`, but not directly as a
  CLI-unit path with invalid options and failure behavior.
- The built binary path is not smoke-tested in automation.
- CLI exit-code contracts are incomplete for production release use.

### Are scripts tested?

No.

There is no root `scripts/` directory in the current checkout.
There are shell helpers under `.agents/coordination/` and
`.agents/hermes-orchestration/`, including:

- `.agents/coordination/agent-work.sh`
- `.agents/hermes-orchestration/run-agent.sh`
- `.agents/hermes-orchestration/dispatch.sh`
- `.agents/hermes-orchestration/collect.sh`
- `.agents/hermes-orchestration/status.sh`

I found no automated shell tests, `shellcheck` gate, or Bats-style smoke tests
for those helpers.
Because the `.agents` scripts coordinate multi-agent work and file claims, they
are operationally important even if they are not part of the shipped product.
They should be tested before relying on them for high-throughput release work.

### Is there CI?

No.

No `.github` directory or `.github/workflows` files were present in the live
checkout.
That means the repo currently depends on local developer discipline for:

- tests;
- linting;
- formatting;
- header checks;
- builds;
- evals;
- live PostgreSQL gates.

This is the single largest release-process gap after the live database gate.
Frequent releases require an external, repeatable quality gate.

### Is there linting or formatting?

Yes, locally.

The repo has `.golangci.yml` and `make lint`.
The linter configuration enables:

- `errcheck`
- `gosimple`
- `govet`
- `ineffassign`
- `staticcheck`
- `typecheck`
- `unused`
- `gofmt`
- `goimports`
- `misspell`
- `unparam`
- `unconvert`
- `revive`

`make lint` passed during this review.
`make check-headers` also passed and enforces the VibeGravity Go source header
policy.

The missing part is CI enforcement.
Local linting is useful, but it does not protect branches unless the same gate
runs automatically.

### Is there a release checklist?

No.

There are review packets and a readiness report, but I did not find a canonical
release checklist.
The closest release-adjacent artifacts are:

- `docs/review-packets/v1-trust-loop-readiness-report.md`
- `docs/review-packets/push-readiness-review-fixes.md`
- `tests/README.md`
- `COMMIT_MESSAGE_RULES.md`

Those documents are useful, but they are not a release process.
A release checklist should explicitly require:

- clean worktree or declared intentional dirt;
- `go test -count=1 ./...`;
- `make eval`;
- `make lint`;
- `make check-headers`;
- `git diff --check`;
- `go build ./cmd/server ./cmd/worker ./cmd/cli`;
- live PostgreSQL gate against a migrated scratch DB;
- migration forward and rollback review;
- changelog entry;
- version/tag decision;
- release notes;
- rollback notes;
- known-risk list.

### Is the changelog accurate and useful?

No changelog exists in the current checkout.

Because `CHANGELOG.md` is absent, it cannot help users understand:

- what changed;
- whether a migration is required;
- whether behavior changed;
- whether an API or CLI command changed;
- whether a release is safe to adopt;
- what known risks remain.

The repo has strong internal review packets, but review packets are not a
substitute for a user-facing changelog.
They are too detailed, too internal, and too numerous for release consumers.

### Is semantic versioning handled well?

No.

I found no version file, no release tags, no SemVer policy, and no release
automation.
`go.mod` defines the module path and Go version, but not a product version.
`git tag --list` returned no release tags.

For the current maturity, the repo should use `v0.x.y` releases until the V1
trust loop is live-verified.
Breaking API, schema, MCP, or CLI behavior should increment minor versions while
the project is pre-1.0.
Patch versions should be reserved for bug fixes that do not alter contracts.

V1 should not be tagged until live PostgreSQL correction trust-loop behavior and
Hermes-facing protocol roundtrips are proven.

### Can users report bugs easily?

No.

I found no:

- issue template;
- bug report template;
- `CONTRIBUTING.md`;
- `SECURITY.md`;
- support policy;
- root README with a bug-report path.

This matters because memory systems fail in subtle ways.
Users need a structured bug template that asks for:

- VibeGravity version or commit;
- OS and Go version;
- PostgreSQL version;
- migration version;
- exact command or HTTP/MCP request;
- redacted config;
- whether `make eval` passes;
- whether `make integration-postgres` passes;
- logs for server, worker, and CLI;
- expected memory behavior versus observed memory behavior;
- privacy/scope impact.

Without that structure, bug reports will be hard to reproduce and easy to
misclassify.

### Are failures observable through logs or clear errors?

Partially.

Positive signals:

- HTTP handlers map core errors to status codes:
  - invalid argument to `400`;
  - not found to `404`;
  - conflict to `409`;
  - not implemented to `501`;
  - unknown server errors to `500`.
- `/healthz` returns `503` when the database pool is missing or unavailable.
- Worker logs show claimed, completed, failed, blocked, applied operations,
  memory IDs, trace counts, and dreaming counts.
- CLI job metrics expose queued, ready queued, running, failed, blocked,
  complete, retryable attempts, oldest queued/running age, drain rate, and
  recovery ETA.
- `doctor` masks database passwords before printing configuration.
- `mcp serve --stdio` returns protocol-level errors for unknown methods and
  invalid tool calls.

Remaining gaps:

- Logs are standard `log` output, not structured logs.
- There is no metrics endpoint.
- There is no tracing.
- There is no release-grade dashboard or alerting guide.
- `doctor` prints errors but still returns success from `runCLI` because the
  `doctor` path calls `runDoctor()` and returns `0`.
- Live failure behavior around real PostgreSQL and real Hermes is not yet
  automated in CI.

The repo has useful operator-visible signals for local development and
debugging, but not enough observability for frequent releases.

## Missing Quality Gates

The highest-priority missing gates are:

1. GitHub Actions or equivalent CI.
2. CI job for `go test -count=1 ./...`.
3. CI job for `make eval`.
4. CI job for `make lint`.
5. CI job for `make check-headers`.
6. CI job for `go build ./cmd/server ./cmd/worker ./cmd/cli`.
7. CI job with PostgreSQL service and migrations that runs
   `make integration-postgres`.
8. CLI binary smoke test after build.
9. Shell helper tests for `.agents` scripts.
10. Release checklist.
11. Changelog enforcement.
12. Version/tag policy.
13. Bug-report and security-report templates.
14. Coverage reporting or at least package-level coverage trend tracking.
15. Migration rollback verification for every new migration.

## Highest-Risk Regression Areas

### 1. Live PostgreSQL correction trust loop

This is the top risk.
The product promise depends on the operator correcting memory once and the
system preserving provenance while changing future recall.
The live trust loop must prove:

- raw correction event creation;
- append-safe correction artifact creation;
- replacement memory creation;
- mandatory trace creation;
- `updates` edge creation;
- prior target supersession;
- retry idempotency;
- conflict behavior for changed evidence;
- explain/timeline visibility;
- next-recall suppression of the old memory.

The test exists, but it skipped in this environment.
That makes it a release blocker until CI runs it.

### 2. MCP and Hermes schema parity

The repo now tests MCP and Hermes dispatch, but protocol drift remains high
risk because external clients depend on exact required fields.
The highest-risk tools are:

- `recall_preview`
- `correct_memory`
- `explain_memory`
- `view_timeline`
- `degraded_status`

Any schema mismatch can make the core service correct but unusable through the
surface Hermes actually calls.

### 3. Scope separation

Scope leakage is existential for a memory product.
The risky boundaries are:

- `agent_private` owner matching;
- `workspace_shared` visibility;
- `group_shared` membership filtering;
- correction visibility before side effects;
- explain/timeline visibility;
- Stage 2 source assembly.

The repo has good local guardrails here, but representative live DB fixtures
should be part of release gating.

### 4. Replay idempotency and concurrent updates

The repo has strong evidence-safe replay tests and a skippable concurrent
PostgreSQL test.
This remains high risk because memory duplication or split latest state can
silently corrupt user trust.

### 5. Worker backlog and degraded freshness

The repo has local backlog metrics and degraded recall metadata.
The next risk is truthfulness under real worker delays:

- stale recall must be labeled as stale;
- recovery ETA must not overclaim;
- blocked jobs must not look like retryable work;
- unrelated maintenance jobs must not falsely make user recall look stale.

### 6. CLI doctor and setup reliability

`doctor` is an important trust command, but it is not yet strict enough as a
release gate.
It should return non-zero when required dependencies are unavailable unless
called in an explicitly advisory mode.

## Recommended Test Plan

### Unit tests

Add or strengthen:

- `doctor` exit-code tests;
- `doctor` DB failure tests;
- `doctor` embedding endpoint failure tests;
- `eval golden` CLI argument and failure tests;
- CLI usage and unknown-command tests;
- release checklist parser or static release-doc tests once release docs exist;
- changelog format tests once `CHANGELOG.md` exists.

### Integration tests

Make these required in CI with PostgreSQL:

- `TestPostgresCorrectMemoryTrustLoop`;
- concurrent update winner/no-dangling-writes test;
- replay idempotency tests;
- `/healthz` with real DB pool;
- migration apply test from empty DB;
- migration down/up smoke where safe;
- correction to explain to timeline to prefetch full path.

### CLI smoke tests

After building binaries, run:

```bash
bin/cli eval golden --path tests/golden/replay_eval.json
bin/cli eval demo
bin/cli hermes bootstrap --name vibegravity --command "$(pwd)/bin/cli"
bin/cli mcp serve --stdio
```

For `doctor`, add two modes:

- advisory local mode, allowed to report missing optional services;
- release mode, required to fail when DB or embedding dependencies are missing.

### Packaging tests

Add a clean-clone style packaging job:

1. Checkout repo.
2. Run `make setup`.
3. Run `make build`.
4. Verify `bin/server`, `bin/worker`, and `bin/cli` exist and execute.
5. Create a scratch PostgreSQL database.
6. Run migrations.
7. Run `make integration-postgres`.
8. Run CLI smoke commands.
9. Archive binaries or release artifacts.

### Script tests

For `.agents` scripts:

- run `shellcheck`;
- test `agent-work.sh init/status/claim/release/done` on a temporary state dir
  if the script can support one;
- smoke-test `dispatch.sh` and `collect.sh` with a fake Hermes command or dry
  run mode.

## Release Process Fixes

### Versioning

Adopt explicit SemVer before the first public release.
Recommended starting policy:

- use `v0.x.y` while the trust loop is still hardening;
- bump minor for API, CLI, MCP, migration, or behavior changes;
- bump patch for compatible fixes;
- reserve `v1.0.0` for a live-verified Hermes Memory trust loop.

### Changelog

Add `CHANGELOG.md` with sections:

- Added;
- Changed;
- Fixed;
- Security;
- Migration Notes;
- Known Risks;
- Verification.

Every release entry should include:

- commit or tag;
- migration impact;
- CLI changes;
- API/MCP changes;
- operator-visible behavior changes;
- tests and gates run.

### Tags and release notes

Use annotated tags:

```bash
git tag -a v0.x.y -m "VibeGravity v0.x.y"
```

Release notes should include:

- product summary;
- upgrade steps;
- rollback notes;
- database migration notes;
- known limitations;
- verification commands;
- compatibility notes for Hermes/MCP.

### Release checklist

Create `docs/release-checklist.md` or `RELEASE.md`.
Minimum checklist:

- worktree state inspected;
- changelog entry added;
- version decided;
- migrations reviewed;
- `go test -count=1 ./...` passed;
- `make eval` passed;
- `make lint` passed;
- `make check-headers` passed;
- `git diff --check` passed;
- binaries built;
- live PostgreSQL gate passed;
- Hermes/MCP smoke passed;
- rollback notes written;
- tag created;
- release notes published.

### Rollback process

Document rollback separately from normal release notes.
At minimum:

- how to redeploy previous binaries;
- how to identify current migration version;
- whether each migration has a safe down path;
- when DB rollback is not safe and forward-fix is required;
- how to inspect blocked jobs;
- how to requeue blocked jobs;
- how to verify recall after rollback.

## Final CTO Decision

The repo can support rapid internal development and review-driven hardening.
It cannot yet support frequent user-facing releases.

The difference is automation and release discipline.
The local engineering culture is strong: tests, evals, lint, source headers,
review packets, and contract docs are all present.
But users are not protected by a CI system, release checklist, changelog,
version policy, bug-reporting path, or mandatory live PostgreSQL gate.

I would approve continued internal iteration.
I would not approve public release cadence until the release system exists and
the live PostgreSQL trust-loop gate runs automatically.

## Source Review

Estimated source: repo-local files and command outputs from
`/Users/parker/Documents/VibeGravity`.
Suspected license: project-owned internal review material.
Similarity risk: low; this report is original analysis based on the live
checkout.
Human review required: yes, because release readiness depends on live
PostgreSQL and Hermes runtime evidence that should be verified in an automated
release environment.
