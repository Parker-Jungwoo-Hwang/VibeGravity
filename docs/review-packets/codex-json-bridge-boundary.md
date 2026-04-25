# Codex JSON Bridge Boundary

## Summary

This slice adds a real but disabled-by-default Codex bridge boundary for the
two-stage reasoning pipeline. It does not enable real Codex in production worker
wiring, does not add local extraction, and does not change graph apply writes.

The bridge now has mockable Stage 1 and Stage 2 Codex runner implementations
behind a narrow JSON client interface. Responses are strictly decoded and
validated before they can become `Stage1Output` or `Stage2Output`.

## Finding or slice fixed

- Added `CodexJSONClient`, schema-marked `CodexRequest` / `CodexResponse`, and
  disabled `CodexStage1Extractor` / `CodexStage2Resolver` runners.
- Added `Stage1ExtractOutputSchemaV0` and preserved the existing
  `Stage2ResolveOutputSchemaV0` / `RequiredOutputSchema` contract.
- Added strict JSON decoding with unknown-field and trailing-JSON rejection.
- Added validation that Stage 2 JSON object fields remain objects before apply.
- Added disabled-by-default config fields:
  `VIBEGRAVITY_CODEX_ENABLED`, `VIBEGRAVITY_CODEX_ENDPOINT`, and
  `VIBEGRAVITY_CODEX_MODEL`.
- Documented the explicit Codex enablement boundary in
  `plans/05_runtime-contracts_ingest-recall-apply.md`.

## Files changed

- `internal/reasoning/codex_bridge.go`
- `internal/reasoning/codex_bridge_test.go`
- `internal/config/config.go`
- `plans/05_runtime-contracts_ingest-recall-apply.md`
- `docs/review-packets/codex-json-bridge-boundary.md`

## Tests run

- `gofmt -w internal/reasoning/codex_bridge.go internal/reasoning/codex_bridge_test.go internal/config/config.go` - passed.
- `go test ./internal/reasoning` - passed.
- `go test ./internal/worker` - passed.
- `go test ./...` - passed.
- `make lint` - passed.
- `make check-headers` - passed.
- `git diff --check` - passed.

## Remaining risks

- Real Codex is still not enabled. The current worker remains wired to stub
  Stage 1 and Stage 2 runners.
- No HTTP/OpenAI client implementation, prompt builder, retry policy, or
  operator-facing runtime enablement path landed in this slice.
- Stage 2 semantic validation still relies on the apply layer for operation
  kind, scope, raw event, and lineage rules.
- `group_shared` remains excluded from Stage 2 retrieval until
  membership-aware filtering exists.
- `update_memory`, profile delta, session summary, plan delta, and group-shared
  writes remain unsupported in store-backed apply.

## Source Review

- Estimated source: first-principles implementation from VibeGravity in-repo
  runtime contracts, AGENTS.md, and existing reasoning interfaces.
- Suspected license: project-internal original work plus Go standard library.
- Similarity risk: low.
- Human review required: yes, normal integration review recommended before a
  real Codex client or production enablement is added.
- External code or restricted-license material used: none.
