# Handoff Prompts and Response Templates

## 1. Purpose

이 문서는 구현 agent를 시작할 때 바로 붙여 넣을 수 있는 템플릿을 제공한다.

## 2. Kickoff Prompt for a Main Coding Agent

```md
You are building VibeGravity.

Read these files first:
- 00_read-this-first_for-building-agents.md
- 01_rfp_vibegravity_hermes-first.md
- 02_product-contract_and_direction.md
- 03_target-architecture_codex-first.md
- 05_runtime-contracts_ingest-recall-apply.md

Your task:
[describe one concrete task]

Goal:
[what must change]

Context:
[files, docs, errors, constraints]

Constraints:
- keep Hermes-first direction
- keep local runtime embedding-only
- do not blur agent_private and workspace_shared memory
- preserve idempotent write path
- update docs with code

Done when:
- [observable success 1]
- [observable success 2]
- [tests and checks]
- [docs updated]

Before coding, write a short plan.
After coding, run checks.
Then review your own diff.
Then report changed files, commands run, results, and risks.
```

## 3. Prompt for a Focused Work Pack

```md
Work pack: [name]

Read:
- [doc 1]
- [doc 2]

Implement only this scope:
- [scope item]
- [scope item]

Do not do:
- [out of scope item]
- [out of scope item]

Return:
1. plan
2. implementation
3. tests
4. doc updates
5. risks
```

## 4. Review Prompt

```md
Review the current diff against the VibeGravity product contract.

Check:
- Hermes-first direction
- scope separation
- raw vs derived separation
- recall budget safety
- human correction support
- tests and docs

Return:
- critical issues
- medium issues
- minor issues
- exact files to inspect
```

## 5. Contract Check Prompt

```md
Compare the changed code with:
- 02_product-contract_and_direction.md
- 03_target-architecture_codex-first.md
- 05_runtime-contracts_ingest-recall-apply.md
- 06_data-model_and_storage-invariants.md

List any contract breaks.
If none, say why the implementation is aligned.
```

## 6. Eval Prompt

```md
Run or inspect the golden scenarios for:
- correction updates old fact
- workspace shared vs agent private separation
- pinned notes in recall
- active plan in recall
- superseded memory suppression

Return:
- pass/fail per scenario
- suspected causes
- next fixes
```

## 7. Response Template for the Implementing Agent

```md
## What I understood

## Plan

## Changes made

## Files changed

## Checks run

## Results

## Risks and follow-ups

## Docs updated
```

## 8. ADR Prompt

```md
We hit an architectural decision.
Write an ADR with:
- context
- options considered
- decision
- consequences
- impact on Hermes-first roadmap
```

## 9. Migration Prompt

```md
Prepare a safe migration plan.

Include:
- current schema
- target schema
- backfill steps
- rollback steps
- replay impact
- profile rebuild impact
```
