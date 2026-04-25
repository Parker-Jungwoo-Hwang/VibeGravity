# GPT-Pro Follow-Up Contract Alignment

Date: 2026-04-25
Owner: Agent 9, architecture documentation and active contract alignment

## Summary

Aligned the active project direction around **Hermes Memory, powered by
VibeGravity**. The product is Hermes Memory; VibeGravity remains the agent
memory engine underneath it.

This docs slice reinforces that VibeGravity is not a generic chat app, not a
raw transcript archive, and not just a vector database. V1 must prove the trust
loop that makes Hermes Memory feel real: correction, provenance, supersession,
scope separation, explain, timeline, and degraded recall.

## Finding Or Slice Fixed

The important drift was not a missing feature description. It was contract
priority. Some older prep docs correctly described `CorrectMemory` as
record-only for the intake slice, but the active direction now includes
correction-driven supersession: raw correction evidence, append-safe
`memory_corrections`, replacement memory, mandatory trace, `updates` edge,
prior-memory supersession, and next-recall suppression of the outdated memory.

The next implementation slice should therefore stay focused on DB/protocol
correctness before new feature breadth:

- P0 correction provenance.
- Live PostgreSQL correction trust-loop tests.
- MCP schema correctness for recall preview, correction, timeline, and explain.
- Evidence-safe replay idempotency.
- Stop-line guardrails around real Codex default enablement, custom Hermes
  provider packaging, documents as the V1 promise, and dreaming as the V1 value
  story.

## Files Changed

- `PLANS.md`
- `plans/02_product-contract_and_direction.md`
- `plans/05_runtime-contracts_ingest-recall-apply.md`
- `plans/06_data-model_and_storage-invariants.md`
- `plans/10_workpack_hermes-provider-and-external-surfaces.md`
- `plans/11_workpack_quality-ops-and-evals.md`
- `docs/review-packets/correctmemory-review-and-gettimeline-prep.md`
- `docs/review-packets/gpt-pro-followup-contract-alignment.md`

## Tests Run

- `git diff --check`

No Go files were edited in this docs-only slice. `make check-headers` is not
required unless a later pass touches Go source files.

## Remaining Risks

- Older review packets and merged GPT-Pro review exports still contain
  historical intake-only `CorrectMemory` wording. They should be treated as
  historical unless promoted into an active plan.
- Live PostgreSQL gates are the next implementation priority, but this slice did
  not implement or run them.
- Real Hermes runtime roundtrips and custom provider packaging remain out of
  scope until the MCP/protocol contract is proven.

## Source Review

- Estimated source: in-repo VibeGravity plans, review packets, and product
  review follow-up instructions.
- Suspected license: project-local documentation.
- Similarity risk: low; wording is original project contract alignment.
- Human review required: recommended, because this packet states active contract
  direction for follow-on agents.
