You are the vuitton Hermes implementation-design agent for VibeGravity.

Work in /Users/parker/Documents/VibeGravity.

Do not mutate Hermes settings or profiles. Do not run hermes config/profile mutation commands.

Read:
- AGENTS.md
- PLANS.md
- plans/05_runtime-contracts_ingest-recall-apply.md
- plans/06_data-model_and_storage-invariants.md
- cmd/cli/main.go
- cmd/cli/main_test.go
- internal/store/store.go
- internal/store/postgres/jobs.go
- internal/store/postgres/jobs_test.go
- internal/core/job.go

Task:
Design the smallest Go/Postgres/API implementation for operator-visible job backlog metrics.

Current CLI has:
- cli jobs blocked [--limit N]
- cli jobs requeue-blocked <job_id>

Needed slice:
- queued/running/failed/blocked/complete counts
- oldest queued age
- simple drain-rate/recovery-ETA shape if possible from current table fields
- no retry/block semantic changes
- no real Codex calls
- preserve existing blocked job recovery commands

Please report:
1. Recommended interface/type shape.
2. Recommended SQL shape and transaction/concurrency risks.
3. CLI command shape.
4. Tests to add.
5. Files likely touched.

Return a concise markdown report. Do not edit files.
