You are the vuitton Hermes profile reviewing Go/Postgres/API implementation design for VibeGravity.

Repo: VibeGravity repository root

Rules:
- Do not mutate Hermes settings.
- Do not run hermes config/profile mutation commands.
- Do not edit files. Return review guidance only.
- Do not propose real Codex calls for this slice.
- Do not reintroduce local extractor behavior.
- Do not enable group_shared writes without membership validation.

Read first:
- AGENTS.md
- PLANS.md
- plans/05_runtime-contracts_ingest-recall-apply.md
- plans/06_data-model_and_storage-invariants.md
- internal/eval/golden.go
- internal/graph/store_apply.go
- internal/store/postgres/memories.go
- internal/kernel/service.go
- tests/golden/replay_eval.json

Task:
Review implementation options for the next smallest V1.0 slice:
"Add replay/eval gates for graph updates and human correction supersession."

Please return:
1. The lowest-risk Go implementation design.
2. Whether this should extend internal/eval, add a new fixture, or add a separate package.
3. Transaction/idempotency risks that the eval must catch.
4. Specific tests that should be added.
5. Files likely to change, with any risks of touching hot files.
