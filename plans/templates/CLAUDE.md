# CLAUDE.md

## Project

This repo builds VibeGravity.  
VibeGravity is a shared memory kernel for Hermes-first agent workflows.

## Hold these facts in every session

- Hermes is the first customer
- local runtime is embedding-focused in v1
- Codex-first reasoning handles extraction and graph resolution
- memory scopes must stay separate
- raw events and derived memories must stay separate
- recall must be compact and token-aware

## Read these docs before major changes

- `00_read-this-first_for-building-agents.md`
- `01_rfp_vibegravity_hermes-first.md`
- `02_product-contract_and_direction.md`
- `03_target-architecture_codex-first.md`
- `05_runtime-contracts_ingest-recall-apply.md`

## Use skills for procedures

If you need a multi-step workflow, use a skill.  
Do not turn CLAUDE.md into a long procedure manual.

## Preferred working pattern

Plan first on hard tasks.  
Implement one coherent work unit at a time.  
Run checks.  
Review the diff.  
Report files changed, checks run, results, and risks.

## Watch for these failures

- scope leakage
- duplicate memory growth
- missing provenance
- empty or noisy recall
- silent contract drift
