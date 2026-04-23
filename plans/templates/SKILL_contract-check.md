---
name: contract-check
description: Use this skill to compare code changes against VibeGravity product and architecture contracts.
---

# Contract Check

## Purpose

This skill checks whether the current implementation violates the documented product contract.

## Required docs

- `02_product-contract_and_direction.md`
- `03_target-architecture_codex-first.md`
- `05_runtime-contracts_ingest-recall-apply.md`
- `06_data-model_and_storage-invariants.md`

## Review checklist

- Hermes-first direction kept
- local extractor not reintroduced
- scope separation preserved
- raw and derived separation preserved
- provenance path preserved
- recall budget logic preserved
- docs updated if behavior changed

## Output

- critical breaks
- medium concerns
- minor notes
- files to inspect next
