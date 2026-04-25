# Agent A — Blocked Job Recovery

## Summary

Implemented an operator-facing recovery path for jobs that were moved to `blocked` after deterministic unsupported apply work. The path lets an operator inspect blocked jobs and manually requeue a specific blocked job after the unsupported operation has landed or after an operator has otherwise decided replay is safe.

This does not change worker retry behavior. Transient failures still use `FailJob` and retry scheduling, while blocked jobs remain out of the queued worker pool until an explicit CLI requeue command is run.

## Files changed

- `internal/store/postgres/jobs.go`
- `internal/store/postgres/jobs_test.go`
- `cmd/cli/main.go`
- `cmd/cli/main_test.go`
- `docs/review-packets/agent-a-blocked-job-recovery.md`

## Behavior added

- Added concrete PostgreSQL job recovery methods:
  - `ListBlockedJobs(ctx, limit)` lists newest blocked jobs with job metadata and `last_error` preserved for inspection.
  - `RequeueBlockedJob(ctx, jobID)` manually changes exactly one currently blocked job back to `queued`.
- Requeue is guarded by `WHERE id = $1 AND status = 'blocked'`, so complete/running/queued jobs are not accidentally touched.
- Manual requeue does not increment `attempts`, does not use the 30-second retry interval, and does not alter transient `FailJob` semantics.
- Added operator CLI surface:
  - `cli jobs blocked [--limit N]`
  - `cli jobs requeue-blocked <job_id>`
- CLI tests use an injected fake store and never open a real database.

## Tests run

- `go test ./internal/store/postgres ./cmd/cli` — red phase first failed on missing `ListBlockedJobs`, `RequeueBlockedJob`, and CLI command functions.
- `gofmt -w internal/store/postgres/jobs.go internal/store/postgres/jobs_test.go cmd/cli/main.go cmd/cli/main_test.go`
- `go test ./internal/store/postgres ./cmd/cli` — passed.
- `go test ./...` — passed after moving the requeue helper outside the source span used by the existing blocked-vs-retryable contract test.
- `gofmt -w internal/store/postgres/jobs.go internal/store/postgres/jobs_test.go cmd/cli/main.go cmd/cli/main_test.go && go test ./... && make lint && make check-headers && git diff --check` — passed.
- Independent review — approved with no critical or important issues; minor note about tenant/workspace filters and audit trail is reflected in remaining risks.

## Remaining risks

- `ListBlockedJobs` currently lists all blocked jobs globally with a limit; tenant/workspace filters may be needed before multi-tenant operations.
- Requeue is intentionally manual and immediate; operators need to know that the previously unsupported apply work is now implemented before replay.
- The CLI output is plain tab-separated text, not JSON; automation may later need `--json`.
- The current job status model still has no separate audit table for who requeued a blocked job or why.

## Source Review

- Estimated source: project-local VibeGravity contracts, existing queue code, and in-repo review packet guidance.
- External sources used: none.
- Suspected license exposure: none beyond Go standard library and existing pgx usage already present in the repo.
- Similarity risk: low; SQL helpers, CLI parsing, and tests were written from first principles for this repository.
- Human review required: recommended before relying on this for production recovery because operator requeue policy and audit requirements may evolve.
