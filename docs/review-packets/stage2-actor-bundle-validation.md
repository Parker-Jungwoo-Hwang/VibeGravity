# Stage 2 Actor Bundle Validation

## Summary

This pass closes the remaining actor-validation follow-up after the
scope-safe Stage 2 source retrieval fix. `process_turn_event` raw event bundles
now fail validation unless every loaded raw event has the same non-empty
`actor_id`, so Stage 2 source retrieval cannot derive private visibility from an
ambiguous bundle.

Stage 2 RequiredOutputSchema is unchanged. Real Codex remains disabled in the
current skeleton, no local extraction was added, and `update_memory` writes were
not implemented.

## Finding fixed

- Mixed or missing raw event actors in one `process_turn_event` bundle could
  reach Stage 2 input preparation, where source retrieval chooses the first
  non-empty actor. The worker now rejects empty actor IDs with
  `core.ErrInvalidArgument` and mixed actor IDs with `core.ErrConflict` before
  Stage 2 source loading runs.

## Files changed

- `internal/worker/processor.go`
- `internal/worker/processor_test.go`
- `plans/05_runtime-contracts_ingest-recall-apply.md`
- `docs/review-packets/stage2-actor-bundle-validation.md`

## Tests run

- `gofmt -w internal/worker/processor.go internal/worker/processor_test.go` - passed.
- `go test ./internal/worker` - passed.
- `go test ./...` - passed.
- `make lint` - passed.
- `make check-headers` - passed.
- `git diff --check` - passed.

## Remaining risks

- Stage 2 still uses a stub Stage 1 result in the current worker skeleton; this
  pass only hardens raw-event actor validation before source loading.
- `group_shared` remains excluded from Stage 2 source retrieval until
  membership-aware filtering is implemented.
- `update_memory` remains validation-only/unsupported for store-backed apply.

## Source Review

- Estimated source: first-principles change from VibeGravity contracts,
  AGENTS.md, and in-repo review packets.
- Suspected license: project-internal original work plus Go standard library.
- Similarity risk: low.
- Review required: yes, normal integration review recommended before enabling
  real Codex.
- Notes: no external project code, GPL-family material, or structured external
  snippets were used.
