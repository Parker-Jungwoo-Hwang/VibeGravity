# CorrectMemory PostgreSQL Provenance Fix

## Summary

Verified the P0 correction trust-loop blocker is fixed in the current checkout.
`CorrectMemory` no longer writes synthetic `correction:<id>` values into
`memory_trace.reasoning_job_id` or `memory_edges.created_by_job_id`.

## Finding or slice fixed

The active implementation creates or reuses a deterministic completed
`ingest_jobs` row with `job_kind = correction_apply` before writing correction
supersession graph state. The deterministic job ID is derived from tenant ID,
workspace ID, correction ID, target memory ID, and correction idempotency key.

That real job ID is then used for both:

- `memory_trace.reasoning_job_id`
- `memory_edges.created_by_job_id`

The foreign key constraints remain intact. The synchronous correction apply job
is completed immediately because it represents operator correction provenance,
not queued worker work.

The response path is truthful: if the correction apply job insert/upsert fails,
supersession is not attempted; if graph supersession fails, the service returns
the store error and does not return an `applied` response.

## Files changed

- `docs/review-packets/correctmemory-postgres-provenance-fix.md`

Already-landed implementation evidence inspected:

- `internal/kernel/service.go`
- `internal/core/job.go`
- `internal/core/kind.go`
- `internal/store/store.go`
- `internal/store/postgres/jobs.go`
- `internal/kernel/service_test.go`
- `internal/store/postgres/jobs_test.go`
- `internal/kernel/correction_trust_loop_integration_test.go`

## Tests run

- `go test ./internal/kernel`
- `go test ./internal/store/postgres -run 'TestCorrectionApplyJob|TestJobBacklogMetricsStatement|TestNormalizeJobBacklogMetricsRequest|TestListBlockedJobsStatementOnlySelectsBlockedStatus'`
- `go test ./...`
- `make lint`
- `make check-headers`
- `git diff --check`

## Remaining risks

- The live PostgreSQL trust-loop test remains opt-in through
  `VIBEGRAVITY_DB_URL`; local deterministic gates can prove SQL shape and kernel
  behavior, but not a live database FK run unless that environment variable is
  configured.
- The correction apply job is upserted before the graph supersession
  transaction. If graph supersession fails afterward, the API remains truthful,
  but a completed provenance job may exist without a trace or edge reference.

## Source Review

- Estimated source: current VibeGravity repository contracts and the task
  prompt's P0 review finding.
- Suspected license: project-owned code and docs only.
- Similarity risk: low; no external code snippets or third-party implementation
  patterns were used.
- Human review required: recommended for live PostgreSQL FK validation using
  `VIBEGRAVITY_DB_URL` and `make integration-postgres`.
