# Next Agent Integration Fixes

## Summary

This pass closes the Work Pack 03 integration review blockers that needed to be
resolved before a real Codex bridge or `update_memory` write path lands.

The worker now has a durable terminal path for deterministic unsupported apply
work, Stage 2 envelope construction goes through `Stage2InputPreparer`, and the
`updates` direct-target uniqueness guard is corrected/documented without
implementing `update_memory`.

## Findings fixed

1. Unsupported apply work no longer retries forever.
   - Added `store.JobStore.BlockJob`.
   - Implemented PostgreSQL `BlockJob` as `status = 'blocked'` with no 30-second retry schedule.
   - Worker routes apply `core.ErrNotImplemented` through the blocked path.
   - Transient reasoning/apply failures still use retrying `FailJob`.

2. Worker Stage 2 input now uses the preparer path.
   - Added a `Stage2InputPreparer` dependency to the worker processor, with an empty-source default.
   - Worker envelope construction now calls `Stage2InputPreparer.Prepare`.
   - Prepared Stage 2 input carries `required_output_schema`.
   - Existing profile/memory/document/plan/note source interfaces remain intact.
   - No local extraction and no real Codex call were added.

3. `updates` edge uniqueness is fixed and bounded.
   - Corrected the bootstrap migration partial unique index from `from_memory_id` to `to_memory_id`.
   - Added ADR-009 to document edge direction and the future `update_memory` transaction rule.
   - `update_memory` remains rejected until target latest locking, new memory/trace/edge write, and prior memory supersession are specified and tested together.

## Files changed

- `cmd/worker/main.go`
- `internal/store/store.go`
- `internal/store/postgres/jobs.go`
- `internal/store/postgres/jobs_test.go`
- `internal/worker/processor.go`
- `internal/worker/processor_test.go`
- `internal/ingest/service_test.go`
- `internal/reasoning/orchestrator.go`
- `migrations/000002_create_core_tables.up.sql`
- `docs/adr-009-updates-edge-lineage-guard.md`
- `plans/05_runtime-contracts_ingest-recall-apply.md`
- `plans/06_data-model_and_storage-invariants.md`
- `docs/review-packets/00-workpack-03-review-index.md`
- `docs/review-packets/team-coordination-log.md`
- `docs/review-packets/next-agent-integration-fixes.md`

## Tests run

- `gofmt -w internal/store/store.go internal/store/postgres/jobs.go internal/store/postgres/jobs_test.go internal/worker/processor.go internal/worker/processor_test.go internal/ingest/service_test.go internal/reasoning/orchestrator.go cmd/worker/main.go` — passed.
- `go test ./internal/worker` — passed.
- `go test ./internal/store/postgres` — passed.
- `go test ./internal/reasoning` — passed.
- `go test ./...` — passed.
- `make lint` — passed.
- `make check-headers` — passed.
- `git diff --check` — passed.

## Remaining risks

- Stage 2 preparation is wired into the worker envelope before the real Codex
  Stage 1 bridge exists. The current stub still returns empty Stage 1 output;
  when the real bridge lands, the worker/orchestrator boundary should prepare
  Stage 2 after actual Stage 1 candidates exist.
- `blocked` jobs are durable and will not be claimed by the current queued-only
  worker query. A later operator command or migration should define how to
  requeue blocked jobs after the unsupported operation becomes implemented.
- The corrected `updates` index prevents two direct updates to the same target
  memory, but it is not the full lineage/latest rule. `update_memory` still
  needs an atomic transaction and tests before write enablement.
- Existing Work Pack 03 files were already present in the working tree before
  this pass; this fix set builds on them and does not attempt to separate or
  revert prior team edits.

## Source Review

- Estimated source: first-principles VibeGravity plans, in-repo review packets, and existing local contracts.
- Suspected license: project-internal original work plus Go standard library and existing pgx dependency usage.
- Similarity risk: low.
- Review required: yes, normal integration review recommended before enabling real Codex or `update_memory`.
- Notes: no external project code, GPL-family material, or structured external snippets were used.
