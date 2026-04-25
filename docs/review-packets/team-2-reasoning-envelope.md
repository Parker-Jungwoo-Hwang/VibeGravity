# Team 2 Review Packet: Reasoning Envelope Preparation

## Summary

Team 2 added an interface-driven Stage 2 input preparation layer under `internal/reasoning`.

The new layer assembles the Stage 2 resolve input from:

- current raw events
- Stage 1 candidate output
- existing profile snapshot
- relevant memories
- relevant document chunks
- active plans
- pinned notes
- required Stage 2 output schema marker

The implementation is intentionally preparation-only. It does not perform local text extraction, does not call Codex, and does not make `StubOrchestrator` infer anything.

## Interfaces added

Added `internal/reasoning/stage2_input_preparer.go`:

- `Stage2InputRequest`
  - Carries job identity, current raw events, and Stage 1 output into preparation.
- `Stage2ProfileSource`
  - `LoadStage2Profile(ctx, req)`
- `Stage2MemorySource`
  - `LoadStage2Memories(ctx, req)`
- `Stage2DocumentSource`
  - `LoadStage2Documents(ctx, req)`
- `Stage2PlanSource`
  - `LoadStage2ActivePlans(ctx, req)`
- `Stage2NoteSource`
  - `LoadStage2PinnedNotes(ctx, req)`
- `Stage2InputSources`
  - Groups the optional profile/memory/document/plan/note sources.
- `Stage2InputPreparer`
  - `Prepare(ctx, req)` assembles `Stage2Input`.
- `Stage2ResolveOutputSchemaV0`
  - Required output schema marker for the Stage 2 resolve contract.

## Files changed

- `internal/reasoning/contracts.go`
  - Added `required_output_schema` to `Stage2Input` while preserving existing Stage 2 output semantics.
- `internal/reasoning/stage2_input_preparer.go`
  - New preparation layer and source interfaces.
- `internal/reasoning/stage2_input_preparer_test.go`
  - New tests for full context assembly, missing-source degraded behavior, and request validation.
- `docs/review-packets/team-2-reasoning-envelope.md`
  - This review packet.

## What remains stubbed

- Codex Stage 1 and Stage 2 execution remains stubbed through the existing `StubOrchestrator`.
- The worker still builds its minimal envelope directly; this change provides the later callable preparation layer but does not wire it into `internal/worker/processor.go`.
- Retrieval implementations are not included here. The new source interfaces are boundaries for later store-backed or retrieval-backed adapters.
- No document text extraction was added.
- No graph/store interface expansion was added.

## Tests run

Targeted TDD cycle:

```bash
go test ./internal/reasoning
```

Result: passed.

Final verification commands requested for handoff:

```bash
gofmt -w internal/reasoning/stage2_input_preparer.go internal/reasoning/stage2_input_preparer_test.go internal/reasoning/contracts.go
go test ./...
make lint
make check-headers
git diff --check
```

Results:

- `gofmt`: passed.
- `go test ./internal/reasoning`: passed.
- `go test ./...`: failed in existing graph/worker compile path: `internal/graph/store_apply.go` references missing `buildMemoryTrace` and `buildMemoryEdge` helpers. Team 2 did not edit `internal/graph`.
- `make lint`: failed because `/Users/parker/.hermes/profiles/vuitton/home/go/bin/golangci-lint` is not installed in this environment.
- `make check-headers`: passed.
- `git diff --check`: passed.

## Risks

- `Stage2InputPreparer` currently performs shallow slice copies. This is sufficient for envelope assembly but does not deep-copy pointed-to `RawEvent`, `Plan`, or `Note` records.
- `required_output_schema` is a new Stage 2 input field. It is additive and does not alter Stage 2 output semantics, but downstream prompt/bridge code should treat it as the authoritative schema marker when the real Codex bridge lands.
- Because store interfaces were intentionally not expanded, real context loading still needs adapter work by the retrieval/store integration team.

## Integration notes for worker team

When worker integration is ready, replace the current direct `Stage2Input` construction with a `Stage2InputPreparer` call after Stage 1 output exists.

Expected flow:

1. Build `Stage1Input` from current raw events.
2. Run Stage 1 through the real reasoning bridge.
3. Call `Stage2InputPreparer.Prepare(ctx, Stage2InputRequest{...})` with the Stage 1 output and current raw events.
4. Pass the prepared `Stage2Input` into Stage 2 resolve.
5. Send only structured `Stage2Output` to the apply engine.

Important boundaries:

- The worker should not do local extraction.
- Retrieval adapters behind the source interfaces may use embeddings, lexical search, and stored records only.
- If a future integration requires expanding shared store interfaces, write the intended change first in `docs/review-packets/team-coordination-log.md` before editing shared store files.

## Source Review

- Estimated source: implemented from the project plans, existing contracts, and in-repo domain types.
- Suspected license: project-internal original work.
- Similarity risk: low; no external code or long snippets were used.
- Human review required: normal project review recommended, especially around the additive `required_output_schema` input marker and later worker integration timing.
