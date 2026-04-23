# AGENTS.md

## Repo purpose

This repo builds VibeGravity.  
VibeGravity is a shared memory kernel for Hermes and other agents.  
It is not a chat UI and not a generic agent runtime.

## Direction

Keep Hermes-first delivery.  
Keep local runtime embedding-only in v1.  
Keep Codex-first reasoning for text interpretation and graph operations.  
Keep agent_private, workspace_shared, and group_shared memory separate.

## Read before work

Always read these files before making non-trivial changes:

- `00_read-this-first_for-building-agents.md`
- `01_rfp_vibegravity_hermes-first.md`
- `02_product-contract_and_direction.md`
- `03_target-architecture_codex-first.md`
- `05_runtime-contracts_ingest-recall-apply.md`
- `06_data-model_and_storage-invariants.md`

## Core invariants

- raw events and derived memories must stay separate
- all write paths must be idempotent
- every memory must keep provenance
- every memory must have explicit scope
- recall must be budget-aware
- human correction is first-class

## Workflow

For complex tasks, plan first.  
For repeated procedures, use skills.  
For bounded exploration, use subagents.  
After coding, run checks, then review your own diff, then update docs.

## Done means

A task is not done until:

- code is implemented
- tests or checks are run
- docs are updated if behavior changed
- risks are reported

## Do not

- reintroduce local extractor dependence into the main path
- blur agent_private and workspace_shared memory
- hide contract changes inside code without docs
- skip tests because the change seems small
