# Hermes Memory Demo Eval

Date: 2026-04-25
Scope: local-only V1 trust-loop demo gate.

## Summary

Added `cli eval demo`, a deterministic local eval that walks the 5-minute
Hermes Memory trust loop without real Hermes, Codex, PostgreSQL, or network
dependencies.

The demo gate proves:

- next-session recall returns a project rule, active plan, and memory with
  scope/source/freshness metadata;
- explain-memory provenance can show why a recalled memory exists;
- operator correction writes a replacement memory, trace, and `updates` edge;
- later recall includes the corrected memory and suppresses the old one;
- another actor's `agent_private` memory does not appear in Hermes recall.

## Finding or slice fixed

The quality plan described a 5-minute Hermes Memory demo, but the repo only had
separate golden, graph replay, and worker backlog gates. Those gates were useful
for implementation safety, but there was no single operator-shaped demo command
that exercised the trust-loop story end to end.

This slice adds that demo as a local eval and wires it into `make eval`.

## Files changed

- `internal/eval/demo.go`
- `internal/eval/demo_test.go`
- `internal/eval/graph_replay.go`
- `cmd/cli/main.go`
- `cmd/cli/main_test.go`
- `Makefile`
- `docs/review-packets/hermes-memory-demo-eval.md`

## Tests run

- `go test ./internal/eval` - passed.
- `go test ./cmd/cli` - passed.
- `go test ./...` - passed.
- `make eval` - passed.
- `make lint` - passed.
- `make check-headers` - passed.
- `git diff --check` - passed.

## Remaining risks

- This is still a local deterministic demo, not a real Hermes runtime roundtrip.
- The demo uses in-memory stores and mocked structured graph operations; it does
  not prove real Codex extraction or a production database session replay.
- `make eval` now runs both golden scenarios and the demo, so future demo drift
  will block the local eval gate.

## Source Review

- Estimated source: in-repo VibeGravity eval, recall, and graph apply code.
- Suspected license: project-internal original code and documentation.
- Similarity risk: low.
- Human review required: yes, because this adds a release-gate style demo.
