You are the default Hermes profile reviewing the next VibeGravity implementation slice.

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
- plans/00_read-this-first_for-building-agents.md
- plans/01_rfp_vibegravity_hermes-first.md
- plans/02_product-contract_and_direction.md
- plans/05_runtime-contracts_ingest-recall-apply.md
- plans/11_workpack_quality-ops-and-evals.md
- docs/review-packets/current-state-and-next-agent-handoff.md

Task:
Review the product and contract acceptance criteria for the next smallest V1.0 slice:
"Add replay/eval gates for graph updates and human correction supersession."

Please return:
1. The exact acceptance criteria this slice should satisfy.
2. Stop-lines: what must stay out of scope.
3. Product risks if the eval is too shallow.
4. A recommended small implementation shape that stays within current contracts.
5. Any docs that must be updated.
