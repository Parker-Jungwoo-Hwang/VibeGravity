# Team Coordination Log

This file is append-only coordination space for Work Pack 03 teams. Preserve existing entries when adding updates.

## Entries

### 2026-04-24 07:17 KST — Team 3 — Worker reliability scope

- Created the coordination log; no pre-existing `docs/review-packets/` entries were present when Team 3 started.
- Team 3 changed worker orchestration only, plus the worker process log line:
  - `internal/worker/processor.go`
  - `internal/worker/processor_test.go`
  - `cmd/worker/main.go`
- Team 3 did not edit the coordination-gated graph/reasoning/storage files:
  - `internal/graph/*`
  - `internal/reasoning/*`
  - `internal/store/postgres/memories.go`
- Team 3 did not implement extraction, did not call real Codex, and did not change graph write semantics.
- Integration note: worker now fails unsupported apply work through `ErrNotImplemented`, records richer job failure context, rejects incomplete/mismatched raw event bundles before reasoning/apply, and reports applied operation/memory/trace counts through `RunResult` and `cmd/worker` logs.

### 2026-04-24 09:00 KST — Integration fixes — Blocked jobs and Stage 2 wiring

- Expanded `store.JobStore` with `BlockJob` so deterministic unsupported apply work can leave the retry queue.
- Worker now routes apply `core.ErrNotImplemented` through the blocked path and leaves transient reasoning/apply errors on `FailJob`.
- Worker Stage 2 envelope construction now goes through `Stage2InputPreparer`, preserving the profile/memory/document/plan/note source interfaces and carrying `required_output_schema`.
- Corrected the `updates` edge direct-target guard to `to_memory_id` in the bootstrap migration and recorded ADR-009 for the future `update_memory` transaction rule.

### 2026-04-24 10:45 KST — Graph concurrency validation — Live Postgres load smoke

- Added a skippable live-Postgres concurrency test for `CreateMemoryWithTraceAndUpdateEdge`.
- Test file:
  - `internal/store/postgres/concurrency_integration_test.go`
- The test launches 16 simultaneous `update_memory` storage attempts against the same active/latest target and asserts:
  - exactly one update commits,
  - the target becomes `superseded` and `latest_flag=false`,
  - the winning update has both `memory_trace` and an `updates` edge,
  - losing workers leave no dangling memory/trace rows.
- This validates the row-lock plus direct-target unique-index contract under real PostgreSQL when `VIBEGRAVITY_DB_URL` is set.
