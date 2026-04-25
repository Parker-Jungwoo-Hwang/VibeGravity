# Team 3 — Worker Reliability Review Packet

## Summary

Team 3 hardened the `process_turn_event` worker orchestration path without implementing extraction, calling real Codex, changing graph write semantics, or weakening apply failure behavior.

The worker now records richer per-job failure context, treats store-backed apply `ErrNotImplemented` as explicit unsupported apply work, rejects incomplete or mismatched raw event bundles before reasoning/apply side effects, returns applied operation counts in `RunResult`, and exposes those counts in the worker process log line.

## Worker behavior changed

- **Job failure reporting**
  - `RunResult` now includes `Failures []JobFailure` with job ID, job kind, and the wrapped error string for each failed claimed job.
  - Errors recorded through `FailJob` now include job ID, job kind, and raw event count in addition to the root cause.
- **Unsupported apply operation handling**
  - When the apply engine returns `core.ErrNotImplemented`, the worker keeps the job failed and wraps the error as unsupported apply work.
  - The job is not completed, and the failure remains visible through both `FailJob` and `RunResult.Failures`.
- **Incomplete raw event bundles**
  - The worker now validates that returned raw events exactly match the job's requested raw event IDs.
  - It rejects missing events, nil events, duplicate returned events, unexpected event IDs, tenant mismatches, and workspace mismatches before building the reasoning envelope.
  - Valid bundles are ordered by `job.RawEventIDs` before entering the reasoning envelope.
- **Retry-safe behavior**
  - Reasoning and apply are skipped when the raw event bundle is incomplete or mismatched, avoiding derived-memory side effects for invalid source bundles.
  - Apply failures still do not complete jobs; deterministic store-backed apply remains responsible for idempotent replay when a job is retried.
  - Apply result inconsistencies that claim applied operations without a trace are treated as conflicts instead of successful completion.
- **Observability/loggability**
  - `RunResult` now aggregates `AppliedOperationCount`, `MemoryIDCount`, and `TraceWrittenCount`.
  - `cmd/worker` logs those aggregate counts for each non-idle worker pass.

## Files changed

- `internal/worker/processor.go`
- `internal/worker/processor_test.go`
- `cmd/worker/main.go`
- `docs/review-packets/00-workpack-03-review-index.md`
- `docs/review-packets/team-3-worker-reliability.md`
- `docs/review-packets/team-coordination-log.md`

## Tests run

- `go test ./internal/worker` — red phase first failed because the new `RunResult` reporting fields did not exist yet.
- `gofmt -w internal/worker/processor.go internal/worker/processor_test.go cmd/worker/main.go`
- `go test ./internal/worker ./cmd/worker` — passed.
- `go test ./...` — passed.
- `gofmt -w cmd/worker/main.go internal/worker/processor.go internal/worker/processor_test.go && go test ./... && make lint && make check-headers && git diff --check` — passed after installing the missing `golangci-lint` binary at the profile GOPATH path used by `make lint`.

## Remaining reliability risks

- The `JobStore` interface still exposes only `FailJob`, so deterministic non-implemented work is still scheduled through the existing retry path rather than a distinct dead-letter/permanent-failure state.
- Completion failure after a successful apply is reported to the caller, but the worker/store contract still lacks a durable "applied but completion failed" handoff state; stale-running reclaim/timeout remains future work because current Postgres claiming only selects queued jobs.
- Per-job failure reports are in-memory `RunResult` details; durable failure detail remains bounded by the current `ingest_jobs.last_error` string.
- Applied operation count observability is aggregate per worker pass; `TraceWrittenCount` counts jobs whose apply result reported a trace, not individual trace rows, and there is not yet structured metrics export or per-job metrics emission.

## Notes for final integration review

- Team 3 did not edit `internal/graph/*`, `internal/reasoning/*`, or `internal/store/postgres/memories.go`.
- Worker raw bundle validation should be compatible with Postgres returning raw events in any order because the worker reorders validated events by the job's requested raw event IDs.
- Store-backed apply remains the place that validates and rejects unsupported graph write semantics; the worker only preserves and reports the failure.
- Raw events remain read-only worker input. Derived memories remain apply engine output.
- The central index links packets that may be created by other teams later; missing linked files should be resolved during final integration review.
- Independent review of Team 3 changes returned APPROVED with no critical or important issues.

## Source Review

- Estimated source: implementation from project-local contracts and existing VibeGravity worker/apply code only.
- External sources used: none.
- Suspected license exposure: none beyond the repository's own code and Go standard library usage.
- Similarity risk: low; changes are small orchestration, validation, and reporting code written from first principles.
- Human review required: recommended for final integration because multiple teams are concurrently editing adjacent Work Pack 03 surfaces.
