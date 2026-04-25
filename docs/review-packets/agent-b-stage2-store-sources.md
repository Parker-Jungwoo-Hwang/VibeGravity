# Agent B Review Packet: Stage 2 Store-Backed Sources

## Summary

Agent B wired real store-backed source adapters behind `Stage2InputPreparer` for the worker path.

The worker now constructs its Stage 2 envelope with a preparer backed by the PostgreSQL store for:

- existing profile snapshots
- memory search results
- document chunk search results
- active plans
- pinned notes

The implementation does not perform local extraction and does not call real Codex. It preserves the existing `reasoning.Stage2InputPreparer` source interfaces and keeps `RequiredOutputSchema` populated through the preparer.

## Adapter design

Added `internal/worker/stage2_sources.go` with `Stage2SourceStores`, `NewStoreBackedStage2InputPreparer`, and `NewStoreBackedStage2InputSources`.

The adapters intentionally sit in `internal/worker` so the reasoning package remains interface-driven and store-agnostic.

Design details:

- `Profiles`
  - Uses `store.ProfileStore.GetProfile`.
  - Looks up the first raw event actor as `agent_private` profile.
  - Falls back to `workspace:<workspace_id>` as `workspace_shared` profile.
  - Treats `core.ErrNotFound` as no profile rather than a job failure.
- `Memories`
  - Uses `store.MemoryStore.SearchMemories`.
  - Uses existing visible scopes: `agent_private`, `workspace_shared`, `session_scratch`.
  - Uses artifact classes: `context`, `knowledge`, `timeline`, `plan`.
  - Search query is built only from structured Stage 1 output hints/candidates if present. It does not parse raw event payload text.
  - With the current stub reasoner path, Stage 1 is empty, so query is empty and the existing store search returns recent active/latest rows.
- `Documents`
  - Uses `store.DocumentStore.SearchDocuments` with the same structured Stage 1 query string.
- `Plans`
  - Uses `store.PlanStore.GetActivePlans` with the same visible scopes.
- `Notes`
  - Uses `store.NoteStore.ListPinnedNotes` with the same visible scopes.

All adapters convert `core.ErrNotFound` into empty context and propagate other store errors so transient store failures still fail the worker job normally.

`cmd/worker/main.go` now injects a store-backed preparer using the single PostgreSQL store instance. The reasoner remains `reasoning.NewStubOrchestrator()`, so there is still no real Codex call.

## Files changed

- `internal/worker/stage2_sources.go`
  - New store-backed Stage 2 source adapters.
- `internal/worker/stage2_sources_test.go`
  - New tests for all source adapters, profile fallback, no raw payload extraction, required schema preservation, and error propagation.
- `cmd/worker/main.go`
  - Injects a store-backed `Stage2InputPreparer` into the worker processor.
- `cmd/cli/main.go`
  - Verification-only repair: stripped accidental embedded line-number prefixes from the working-tree copy so `go test ./...` could parse the existing CLI recovery implementation. This was not part of the Stage 2 adapter design.
- `docs/review-packets/agent-b-stage2-store-sources.md`
  - This review packet.

## Tests run

Targeted RED check before implementation:

```bash
go test ./internal/worker -run TestStoreBackedStage2InputPreparer -count=1
```

Result: failed as expected because `NewStoreBackedStage2InputPreparer` and `Stage2SourceStores` did not exist.

Targeted GREEN checks after implementation:

```bash
gofmt -w internal/worker/stage2_sources.go internal/worker/stage2_sources_test.go cmd/worker/main.go
go test ./internal/worker -run TestStoreBackedStage2InputPreparer -count=1
go test ./internal/worker ./cmd/worker -count=1
go test ./internal/worker -count=1
```

Result: passed.

Full-suite check run before this packet:

```bash
go test ./...
```

Result: passed after the working tree stabilized from concurrent edits in nearby job/CLI integration files.

Final required verification:

```bash
gofmt -w internal/worker/stage2_sources.go internal/worker/stage2_sources_test.go cmd/worker/main.go cmd/cli/main.go
go test ./...
make lint
make check-headers
git diff --check
```

Result: passed.

## Remaining risks

- Current worker architecture still uses the combined `Reasoner.ProcessTurn` stub. Stage 2 preparation therefore happens before a real Stage 1 Codex pass exists. The adapters are ready to consume Stage 1 output once the bridge is split or the orchestrator fills it, but today the query is usually empty.
- Memory/document retrieval is limited to existing store search interfaces and the current lexical fallback behavior. No embeddings or neighborhood expansion are wired here.
- `agent_private` memory search still depends on current store search scope filtering; the store search interface does not accept an owner/entity filter yet. This packet did not change shared store contracts.
- `group_shared` is intentionally not included because membership-aware source filtering is not implemented in the available store/search contract.
- Profile input remains singular because `Stage2Input` currently accepts one `ExistingProfile`; the adapter prefers actor private profile and falls back to workspace profile.

## Source Review

- Estimated source: first-principles implementation from VibeGravity in-repo plans, existing review packets, and existing store/reasoning contracts.
- Suspected license: project-internal original work.
- Similarity risk: low; no external project code or long structured snippets were used.
- Human review required: normal integration review recommended, especially around empty-query retrieval volume and the lack of owner filtering in `SearchMemories` for `agent_private` scope.
