# Operator-Visible Degraded Recall Freshness

## Summary

`prefetch()` now has a narrow, read-only freshness signal path for worker/Codex
lag. When backlog metrics show stale ready work, long-running claimed work, or
retryable queued attempts, recall response metadata reports degraded freshness
and derived stored recall blocks are labeled `stale`.

## Finding or slice fixed

- Added `RecallMeta.freshness` and `RecallMeta.freshness_lag_seconds` so
  operators can distinguish normal stored recall from stale/degraded recall.
- Added `recall.FreshnessProvider` and `BacklogFreshnessProvider`, backed by
  existing read-only `JobMetricsStore`, without changing worker queue or graph
  write semantics.
- Added oldest running job age to read-only backlog metrics and recall freshness
  so stuck in-flight work can degrade derived recall before it returns to the
  retry queue.
- Downgraded only derived recall sources (`memories`, `profile`,
  `session_summaries`) when backlog/Codex retry state says memory may lag.
  Manual notes, active plans, and document chunks keep their own source labels.
- Preserved MCP `recall_preview` behavior because it remains an alias over the
  shared `prefetch()` response.

## Files changed

- `internal/core/dto.go`
- `internal/core/job.go`
- `internal/recall/assembler.go`
- `internal/recall/freshness.go`
- `internal/recall/assembler_test.go`
- `internal/store/postgres/jobs.go`
- `internal/store/postgres/jobs_test.go`
- `cmd/server/main.go`
- `cmd/cli/main.go`
- `cmd/cli/main_test.go`
- `PLANS.md`
- `plans/05_runtime-contracts_ingest-recall-apply.md`
- `plans/11_workpack_quality-ops-and-evals.md`
- `docs/review-packets/recall-preview-trust-metadata.md`
- `docs/review-packets/operator-visible-degraded-recall-freshness.md`

## Tests run

- `gofmt -w internal/core/dto.go internal/recall/assembler.go internal/recall/freshness.go internal/recall/assembler_test.go cmd/server/main.go cmd/cli/main.go`
- `go test ./...`
- `make eval`
- `make lint`
- `make check-headers`
- `git diff --check`

## Remaining risks

- Freshness is currently inferred from backlog metrics, not a dedicated Codex
  health heartbeat.
- Running-job age is measured from `locked_at` with `updated_at` as a defensive
  fallback; it still does not prove whether the worker is truly dead or just
  executing a slow valid job.
- Real Hermes runtime roundtrip remains unverified; MCP `recall_preview` is
  covered by the shared prefetch alias semantics.

## Source Review

- Estimated source: in-repo VibeGravity contracts and implementation.
- Suspected license: project-internal original code.
- Similarity risk: low; implementation is a small original adapter over existing
  local interfaces.
- Human review required: recommended before enabling real Codex by default.
