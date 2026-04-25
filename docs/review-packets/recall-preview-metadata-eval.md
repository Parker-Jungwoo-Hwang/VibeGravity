# Recall Preview Metadata Eval

Date: 2026-04-25
Scope: deterministic golden eval coverage for Hermes Memory recall trust metadata.

## Summary

Strengthened the golden recall eval so it can fail on missing or incorrect
operator-visible recall block metadata.

The first recall scenario now asserts that pinned notes, active plans, and
memory blocks expose the expected scope, source, source id, status, freshness,
and owner metadata. This keeps the Hermes Memory recall preview contract under
`make eval`, instead of checking only rendered text and source names.

## Finding or slice fixed

Before this slice, golden recall scenarios verified block kinds, text, sources,
and token budget. That protected recall usefulness, but not the trust-loop
metadata that tells Hermes operators why a block is visible and whether it is
stored or stale.

This slice adds a small `block_metadata` expectation shape to the eval runner
and covers regression reporting with a focused unit test.

## Files changed

- `internal/eval/golden.go`
- `internal/eval/golden_test.go`
- `tests/golden/replay_eval.json`
- `plans/11_workpack_quality-ops-and-evals.md`
- `docs/review-packets/recall-preview-metadata-eval.md`

## Tests run

- `go test ./internal/eval` - passed
- `go run ./cmd/cli eval golden --path tests/golden/replay_eval.json` - passed

## Remaining risks

- This checks deterministic recall metadata shape, not a live Hermes runtime
  roundtrip.
- The current eval compares block metadata by position; that is intentional for
  priority/order regressions, but broader unordered matching can be added if a
  later scenario needs it.

## Source Review

- Estimated source: in-repo VibeGravity eval and recall contracts.
- Suspected license: project-internal original code.
- Similarity risk: low.
- Human review required: no for source provenance; yes for product judgment on
  whether these are the right first metadata fields to lock.
