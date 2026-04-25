# Agent D Contract Gates

## Summary

Added lightweight contract-gate tests for the Work Pack 03 integration fixes.
The gates are intentionally narrow and do not implement feature behavior,
call Codex, add local extraction, or enable `update_memory` writes.

## Gates added

- Migration contract gate:
  - `tests/migration_contract_test.go`
  - Verifies `memory_edges_single_updates_target_idx` targets
    `to_memory_id` for `edge_kind = 'updates'`.
  - Verifies the index does not regress back to `from_memory_id`.

- Job failure contract gate:
  - `tests/migration_contract_test.go`
  - Verifies retryable `FailJob` SQL keeps `status = 'queued'` and the
    30-second retry interval.
  - Verifies permanent unsupported `BlockJob` SQL keeps `status = 'blocked'`
    and does not schedule automatic retry.

- Reasoning envelope contract gate:
  - `internal/reasoning/orchestrator_test.go`
  - Verifies the stub orchestrator rejects Stage 2 envelopes without
    `RequiredOutputSchema`.
  - Verifies the prepared-schema path still returns structured resolve-stage
    output from the stub orchestrator.

## Files changed

- `tests/migration_contract_test.go`
- `internal/reasoning/orchestrator_test.go`
- `docs/review-packets/agent-d-contract-gates.md`

## Tests run

- `gofmt -w tests/migration_contract_test.go internal/reasoning/orchestrator_test.go` - passed.
- `go test ./internal/reasoning` - passed.
- `go test ./tests` - passed.
- `go test ./...` - blocked by concurrent recovery-lane changes:
  - `cmd/cli/main_test.go` references `runCLI`, `blockedJobStoreFactory`, and `blockedJobStore` before those symbols exist.
  - `internal/store/postgres/jobs_test.go` references `listBlockedJobsStatement`, `scanIngestJobRows`, and `requeueBlockedJob` before those symbols exist.
- `make lint` - blocked by the same concurrent compile/typecheck failures.
- `make check-headers` - passed.
- `git diff --check` - passed.

## Remaining risks

- The job failure contract gate is a static source contract because this lane
  avoids editing Agent A's blocked-job recovery implementation files. If Agent A
  refactors job SQL into constants or helpers, this gate may need a small update
  while preserving the same behavior assertions.
- Full repo verification currently depends on Agent A completing or removing
  its in-progress blocked-job recovery and CLI tests.
- These gates protect the current contract only; they do not replace future
  `update_memory` transaction tests.

## Source Review

- Estimated source: first-principles VibeGravity plans, ADR-009, and in-repo integration review packet.
- Suspected license: project-internal original work plus Go standard library usage.
- Similarity risk: low.
- Review required: normal integration review recommended after all four parallel lanes land.
- Notes: no external project code, GPL-family material, or structured external snippets were used.
