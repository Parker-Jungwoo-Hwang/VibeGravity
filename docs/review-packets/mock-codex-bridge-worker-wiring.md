# Mock Codex Bridge Worker Wiring

## Summary

This slice replaces the production worker's direct Stub Stage 1 / Stage 2 runner
wiring with the explicit Codex bridge runners backed by a deterministic mocked
`CodexJSONClient`.

The worker still does not call a real Codex API. The mocked client returns strict
structured JSON through the same `CodexRequest` / `CodexResponse` boundary that a
future real client must implement.

## Changed

- Added `internal/reasoning/mock_codex_client.go`.
- Wired `cmd/worker/main.go` through:
  - `CodexStage1Extractor`
  - `CodexStage2Resolver`
  - `MockCodexJSONClient`
- Added a bridge-level test proving the mocked client runs through the real
  Stage 1 / Stage 2 runners.
- Updated `plans/05_runtime-contracts_ingest-recall-apply.md` to describe the
  current mocked bridge state.

## Boundaries Preserved

- No real Codex API call.
- No local extractor fallback.
- No free-form reasoning output crossing into apply.
- No graph write behavior changed.
- The future real client should implement `reasoning.CodexJSONClient` and keep
  strict structured JSON validation at the runner boundary.

## Verification

Run before handoff:

```bash
gofmt -w cmd/worker/main.go internal/reasoning/mock_codex_client.go internal/reasoning/codex_bridge_test.go
go test ./internal/reasoning ./internal/worker
go test ./...
make lint
make check-headers
git diff --check
```

## Source Review

- Estimated source: first-principles implementation from VibeGravity in-repo
  runtime contracts and existing reasoning interfaces.
- Suspected license: project-internal original work plus Go standard library.
- Similarity risk: low.
- Human review required: yes, normal integration review recommended before a
  real Codex client is added.
- External code or restricted-license material used: none.
