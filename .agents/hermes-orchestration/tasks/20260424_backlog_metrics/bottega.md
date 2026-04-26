You are the bottega Hermes QA/regression agent for VibeGravity.

Work in the VibeGravity repository root.

Do not mutate Hermes settings or profiles. Do not run hermes config/profile mutation commands.

Read:
- AGENTS.md
- PLANS.md
- plans/05_runtime-contracts_ingest-recall-apply.md
- plans/11_workpack_quality-ops-and-evals.md
- docs/review-packets/current-state-and-next-agent-handoff.md
- cmd/cli/main_test.go
- internal/store/postgres/jobs_test.go
- internal/eval/worker_backlog.go
- tests/golden/replay_eval.json

Task:
Define QA/regression coverage for the next slice: operator-visible worker backlog counts and recovery metrics.

Please focus on:
1. Edge cases for queue status counts and oldest queued age.
2. How to prove blocked jobs do not look retryable.
3. How to keep the existing mocked Codex outage eval gates passing.
4. Release-gate commands and any doc assertions.
5. Any test that would catch accidental real Codex/local extractor behavior.

Return a concise markdown report. Do not edit files.
