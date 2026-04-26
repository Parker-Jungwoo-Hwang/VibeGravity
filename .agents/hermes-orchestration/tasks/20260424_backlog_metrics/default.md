You are the default Hermes review agent for VibeGravity.

Work in the VibeGravity repository root.

Do not mutate Hermes settings or profiles. Do not run hermes config/profile mutation commands.

Read:
- AGENTS.md
- PLANS.md
- plans/00_read-this-first_for-building-agents.md
- plans/01_rfp_vibegravity_hermes-first.md
- plans/02_product-contract_and_direction.md
- plans/03_target-architecture_codex-first.md
- plans/05_runtime-contracts_ingest-recall-apply.md
- plans/06_data-model_and_storage-invariants.md
- docs/review-packets/current-state-and-next-agent-handoff.md

Task:
Review the next smallest V1.0 slice: operator-visible worker backlog counts and recovery metrics, without changing retry/block semantics and without real Codex calls.

Please report:
1. Product/contract acceptance criteria for this slice.
2. Stop-lines: what must not be implemented in this slice.
3. What docs must be updated.
4. Any risk that could accidentally blur private/shared/group scope or raw/derived separation.

Return a concise markdown report. Do not edit files.
