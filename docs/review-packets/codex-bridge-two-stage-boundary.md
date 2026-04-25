# Codex Bridge Two-Stage Boundary

## Summary

This pass adds the next safe Codex bridge slice without enabling real Codex
calls. Reasoning now has explicit mockable Stage 1 and Stage 2 runner
interfaces, plus a pipeline orchestrator that runs Stage 1 before preparing the
Stage 2 input.

The production worker still uses stub runners, so no local extraction was added
and no test calls real Codex. The important integration fix is that store-backed
Stage 2 source retrieval now sits behind the reasoning orchestrator and receives
actual Stage 1 structured output when a real extractor is later plugged in.

## Slice fixed

- Added `Stage1Extractor` and `Stage2Resolver` interfaces in
  `internal/reasoning`.
- Added `PipelineOrchestrator`, which runs:
  Stage 1 extract -> Stage 2 input preparation -> Stage 2 resolve.
- Kept `StubOrchestrator` as a compatibility wrapper around stub Stage 1 and
  Stage 2 runners.
- Updated the worker envelope path so the worker validates/loads raw events and
  passes a schema-marked Stage 2 shell, but does not retrieve Stage 2 context
  before Stage 1.
- Updated `cmd/worker` to use the pipeline orchestrator with stub runners and
  the existing store-backed Stage 2 preparer.
- Added unit coverage proving Stage 2 sources receive Stage 1 output before the
  resolver runs and that Stage 2 is skipped when Stage 1 fails.

## Files changed

- `cmd/worker/main.go`
- `internal/reasoning/orchestrator.go`
- `internal/reasoning/orchestrator_test.go`
- `internal/worker/processor.go`
- `internal/worker/processor_test.go`
- `plans/05_runtime-contracts_ingest-recall-apply.md`
- `docs/review-packets/codex-bridge-two-stage-boundary.md`

## Tests run

- `gofmt -w cmd/worker/main.go internal/reasoning/orchestrator.go internal/reasoning/orchestrator_test.go internal/worker/processor.go internal/worker/processor_test.go` - passed.
- `go test ./internal/reasoning ./internal/worker` - passed.
- `go test ./internal/worker` - passed.
- `go test ./...` - passed.
- `make lint` - passed.
- `make check-headers` - passed.
- `git diff --check` - passed.

## Remaining risks

- Stage 1 and Stage 2 are still stubbed. This slice creates the bridge seam but
  does not add a Codex client, prompt builder, JSON schema validator, retry
  policy, or production configuration.
- The worker process now wires the pipeline with stub runners, so Stage 2 source
  retrieval still usually searches from empty Stage 1 output until a real Stage
  1 extractor lands.
- `group_shared` remains excluded from Stage 2 sources until membership-aware
  filtering exists.
- `update_memory`, profile delta, session summary, plan delta, and group-shared
  writes remain unsupported in store-backed apply.

## Source Review

- Estimated source: first-principles implementation from VibeGravity in-repo
  runtime contracts, AGENTS.md, and existing local reasoning/worker interfaces.
- Suspected license: project-internal original work plus Go standard library.
- Similarity risk: low.
- Review required: yes, normal integration review recommended before replacing
  stub runners with a real Codex client.
- Notes: no external project code, GPL-family material, or structured external
  snippets were used.
