# Hermes Memory Trust Loop Product Pivot

Date: 2026-04-25
Scope: product direction documents and V1 planning posture.

## Summary

V1 is now framed as **Hermes Memory, powered by VibeGravity**.

The VibeGravity name remains the engine and internal architecture name. The
first user-facing product story should no longer lead with "shared memory
kernel". The V1 promise is:

> Hermes remembers the right project context across sessions, shows why it
> remembered it, and lets the operator fix memory once.

This keeps the engine direction intact while making the first product wedge
clearer to a Hermes operator.

## Finding or slice fixed

The planning docs were too engine-first for the first product story. That was
technically accurate but weak as a user-facing V1 promise.

This pass fixes the product framing by moving the next work from broad
integration toward the **Hermes Memory trust loop**:

- recall preview
- visible scope
- explain/timeline provenance
- correction
- supersession
- degraded freshness metadata
- next relevant recall reflects the correction

Documents and rich dreaming remain engine capabilities, but they are no longer
the V1 headline.

## Files changed

- `PLANS.md`
- `plans/02_product-contract_and_direction.md`
- `plans/05_runtime-contracts_ingest-recall-apply.md`
- `plans/06_data-model_and_storage-invariants.md`
- `plans/10_workpack_hermes-provider-and-external-surfaces.md`
- `plans/11_workpack_quality-ops-and-evals.md`
- `docs/review-packets/hermes-memory-trust-loop-product-pivot.md`

## Tests run

- `git diff --check` - passed for the documentation files touched by this slice.

Not run because this is a documentation-only product-direction change:

- `go test ./...`
- `make eval`
- `make lint`
- `make check-headers`

## Remaining risks

- The current code may not yet expose a polished `recall preview` command or
  Hermes-facing degraded status even though the product docs now require it.
- `internal/hermes.Provider` and `internal/mcp.Surface` should be reviewed next
  against the trust-loop surface list.
- Real Codex remains disabled by default until failure behavior and freshness
  state are operator-visible.
- A 5-minute Hermes Memory demo still needs to be scripted and verified against
  a real local workflow.

## Source Review

- Estimated source: in-repo VibeGravity planning docs, current consulting
  packet, and user-provided Product Owner consulting response.
- Suspected license: project-internal original material.
- Similarity risk: low.
- Human review required: yes, because the change updates product direction and
  V1 scope rather than only wording.
