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

## Severity Template

Use this severity scale when reporting findings:

```text
Severity: critical | high | medium | low
Contract: product | runtime | storage | API | eval | docs
Evidence: file path and line, command output, or scenario name
Impact: what can break for Hermes Memory or operator trust
Required fix: the smallest change needed before merge or handoff
Live DB required: yes | no
```

- `critical`: breaks a core invariant, leaks scope, loses provenance, corrupts
  correction/supersession, or makes a live trust-loop claim without PostgreSQL
  proof.
- `high`: likely product regression in recall, correction, explain/timeline,
  worker retry, idempotency, or external protocol behavior.
- `medium`: contract drift that should be fixed before the next work pack but
  does not immediately corrupt memory state.
- `low`: documentation, naming, or follow-up clarity issue.

## Output

- critical breaks (blocks merge)
- medium concerns (should fix before next pack)
- minor notes (improve when convenient)
- files to inspect next
