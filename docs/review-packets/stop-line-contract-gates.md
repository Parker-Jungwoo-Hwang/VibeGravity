# Stop-Line Contract Gates

## Summary

Agent 8 added focused contract gates for the four review stop-lines. This slice
does not add product features, broaden Hermes packaging, enable real Codex, or
open group-shared writes. It locks the current safety boundaries while the trust
loop continues to mature.

## Finding or slice fixed

- Real Codex remains disabled by default at the config boundary.
- The default worker composition runs through the mocked Codex bridge and does
  not locally extract memories from raw event text.
- `group_shared` graph writes remain rejected before storage until membership
  validation exists in the write path.
- Hermes bootstrap output remains a narrow MCP registration/test command and
  does not claim provider packaging or install readiness.

## Files changed

- `internal/config/config_test.go`
- `cmd/worker/main_test.go`
- `internal/graph/stop_line_contract_test.go`
- `cmd/cli/hermes_bootstrap_stopline_test.go`
- `docs/review-packets/stop-line-contract-gates.md`

## Tests run

- `go test ./internal/config`
- `go test ./cmd/worker`
- `go test ./internal/graph`
- `go test ./cmd/cli`
- `go test ./...` blocked by active replay evidence lane:
  `internal/store/postgres/memories_test.go:288` expected `ErrConflict`, got
  `not found` in `TestValidateReplayEvidenceRejectsChangedSemanticEvidence`.
- `make lint` blocked by active replay helper lint:
  `internal/store/postgres/memories_replay_test.go` has `unparam` findings in
  `replayUpdateMemory`, `replayUpdateEdge`, and `cleanupPostgresReplayRows`.
- `make check-headers`
- `git diff --check`

## Remaining risks

- The full `plans/05_runtime-contracts_ingest-recall-apply.md` doc was not
  edited in this lane because it was actively claimed by another agent.
- Full-repo test and lint are currently blocked by active replay evidence work
  outside this lane; focused stop-line tests, header check, and diff check pass.

## Source Review

- Estimated source: repo-local contracts and current implementation only.
- Suspected license: project-owned VibeGravity code and docs.
- Similarity risk: low; tests were written from observed repo behavior and
  project stop-line requirements.
- Human review required: normal review recommended because these tests encode
  product safety boundaries.
