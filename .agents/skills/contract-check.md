---
name: contract-check
description: Use this skill to compare code changes against VibeGravity product and architecture contracts.
---

# Contract Check

## Purpose

This skill checks whether the current implementation violates the documented product contract.

## Required docs

- `plans/02_product-contract_and_direction.md`
- `plans/03_target-architecture_codex-first.md`
- `plans/05_runtime-contracts_ingest-recall-apply.md`
- `plans/06_data-model_and_storage-invariants.md`

## Review checklist

- Hermes-first direction kept
- Local extractor not reintroduced into main path
- Scope separation preserved (agent_private / workspace_shared / group_shared)
- Raw and derived separation preserved
- Provenance path preserved (memory_trace mandatory)
- Recall budget logic preserved
- Stage 1 / Stage 2 / Apply boundary is structured JSON only
- Apply engine 12-step pipeline order respected
- Idempotency preserved on all write paths
- group_shared requires valid membership
- Docs updated if behavior changed

## Output

- critical breaks (blocks merge)
- medium concerns (should fix before next pack)
- minor notes (improve when convenient)
- files to inspect next
