# GPT-Pro Follow-Up Product Contract Alignment

Date: 2026-04-25
Owner: Agent 9, architecture documentation and product-contract alignment

## Summary

Aligned the active docs around the product verdict: V1 is **Hermes Memory,
powered by VibeGravity**. VibeGravity remains the engine and architecture name,
while the first product promise is the Hermes operator trust loop: recall
preview, explain/timeline, correction, supersession, visible scope, and honest
freshness/degraded status.

## Finding Or Slice Fixed

The main drift was historical correction language. Older prep material framed
`CorrectMemory` as record-only and explicitly avoided supersession. That was
true for the intake slice, but the active contract now includes
correction-driven supersession backed by raw correction evidence,
`memory_corrections`, replacement memory, mandatory trace, `updates` edge, and
prior-memory supersession.

The next slice is therefore not broad new features. It is DB/protocol
correctness:

- P0 correction provenance.
- MCP schema correctness for trust-loop tools.
- Evidence-safe replay idempotency.
- Live PostgreSQL integration gates.
- Stop-line protection around real Codex default enablement, custom Hermes
  provider packaging, documents as product promise, and dreaming as the V1 value
  story.

## Files Changed

- `PLANS.md`
- `plans/02_product-contract_and_direction.md`
- `plans/05_runtime-contracts_ingest-recall-apply.md`
- `plans/06_data-model_and_storage-invariants.md`
- `plans/10_workpack_hermes-provider-and-external-surfaces.md`
- `plans/11_workpack_quality-ops-and-evals.md`
- `docs/review-packets/correctmemory-review-and-gettimeline-prep.md`
- `docs/review-packets/gpt-pro-followup-product-contract-alignment.md`

## Tests Run

- `git diff --check`

No Go files were edited for this docs-only slice, so `make check-headers` was
not required by the definition of done.

## Remaining Risks

- Some older review packets may still contain intake-only correction wording.
  They should remain historical unless a future packet is promoted to active
  contract status.
- Live PostgreSQL integration gates are called out as the next priority, but
  this slice did not implement those gates.
- External Hermes custom memory-provider packaging remains intentionally out of
  scope until the MCP/protocol roundtrip is proven.

## Source Review

- Estimated source: in-repo VibeGravity plans, review packets, and current
  product-review follow-up direction.
- Suspected license: project-local documentation.
- Similarity risk: low; wording was written from current repo contracts and
  review instructions.
- Human review required: recommended, because this packet defines active product
  contract language for follow-on agents.
