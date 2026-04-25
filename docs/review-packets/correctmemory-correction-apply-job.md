# CorrectMemory Correction Apply Job

## Summary

Fixed the PostgreSQL FK-safety blocker in `CorrectMemory` correction supersession.
The replacement memory trace and `updates` edge now use a real deterministic
`ingest_jobs` row instead of synthetic `correction:<id>` identifiers.

## Finding or slice fixed

`CorrectMemory` previously built correction supersession provenance with
synthetic job IDs derived from the correction artifact. That could pass fake
store tests but violate the PostgreSQL foreign keys on
`memory_trace.reasoning_job_id` and `memory_edges.created_by_job_id`.

This slice preserves the FK contract by adding a deterministic completed
`correction_apply` job row. The job ID is stable for the tenant, workspace,
correction ID, target memory ID, and correction idempotency key. The kernel
creates or reuses that job before writing the replacement memory trace and
`updates` edge.

If the correction apply job cannot be created, supersession does not run. If the
graph supersession fails after the correction artifact is recorded, the service
returns the store error and does not report `applied`.

## Files changed

- `internal/core/kind.go`
- `internal/core/job.go`
- `internal/store/store.go`
- `internal/store/postgres/jobs.go`
- `internal/store/postgres/jobs_test.go`
- `internal/kernel/service.go`
- `internal/kernel/service_test.go`
- `cmd/server/main.go`
- `cmd/cli/main.go`
- `docs/review-packets/correctmemory-correction-apply-job.md`

## Tests run

- `go test ./internal/kernel` passed.
- `go test ./internal/store/postgres` passed after the active replay implementation lane settled.
- `go test ./internal/store/postgres -run 'TestCorrectionApplyJob|TestScanIngestJobRowsReturnsBlockedJobs|TestListBlockedJobsStatementOnlySelectsBlockedStatus'` passed.
- `go test ./cmd/server ./cmd/cli` passed.
- `go test ./...` passed.
- `make check-headers` passed.
- `git diff --check` passed.
- `/Users/parker/go/bin/golangci-lint run ./internal/kernel ./internal/core ./internal/store ./cmd/server ./cmd/cli` passed.

`make lint` is currently blocked by the active replay-idempotency lane in
`internal/store/postgres/memories_replay_test.go` with `unparam` findings for
`replayUpdateMemory`, `replayUpdateEdge`, and `cleanupPostgresReplayRows`. I did
not edit that claimed file.

## Remaining risks

- I did not run a live PostgreSQL correction supersession because this slice only
  added mocked/unit coverage and the live DB gate is being handled in another
  active lane.
- The correction apply job is inserted before the graph supersession transaction.
  If graph apply fails, the completed provenance job can exist without a trace or
  edge referencing it; the API response remains truthful and returns the failure.

## Source Review

- Estimated source: implemented from VibeGravity repository contracts and the
  GPT-Pro finding in the task prompt.
- Suspected license: project-owned code only.
- Similarity risk: low; no external code snippets or distinctive third-party
  implementation patterns were used.
- Human review required: yes, for live PostgreSQL FK validation and the final
  transaction-shape decision around pre-created correction apply jobs.
