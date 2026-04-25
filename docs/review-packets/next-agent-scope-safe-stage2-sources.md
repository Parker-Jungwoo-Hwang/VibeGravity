# Scope-Safe Stage 2 Sources

## Summary

This pass fixes the Stage 2 source privacy blocker before any real Codex bridge
is enabled. Store-backed Stage 2 memory, pinned note, and active plan retrieval
now derives the visible actor from the validated raw event bundle and carries
that actor as `owner_entity_id` when `agent_private` is included in visible
scopes.

The PostgreSQL store contract now excludes `agent_private` rows unless the
request owner matches the row owner. Stage 2 also keeps a caller-side visibility
filter so a buggy or future source cannot leak another actor's private source
rows into the reasoning envelope. `workspace_shared` and `session_scratch`
retrieval remain enabled. `group_shared` remains excluded until membership-aware
filtering exists.

## Findings fixed

- Stage 2 memory search no longer asks for `agent_private` without an actor
  owner. `SearchMemoriesRequest` carries `OwnerEntityID`, and PostgreSQL memory
  search requires matching `owner_entity_id` for `agent_private` rows.
- Stage 2 pinned notes and active plans now use request objects with
  `TenantID`, `WorkspaceID`, `Scopes`, and `OwnerEntityID`; PostgreSQL note and
  plan queries apply the same private-owner predicate.
- Stage 2 source adapters defensively filter returned memories, notes, and
  plans so another actor's `agent_private` rows and all `group_shared` rows are
  dropped before the Stage 2 input is prepared.
- Prefetch recall now passes `PrefetchRequest.ActorID` through the same shared
  store contracts, preserving the generic scope-safety rule outside Stage 2.
- Runtime/data-model docs now state that private retrieval requires an
  owner-scoped request.

## Files changed

- `internal/core/dto.go`
- `internal/store/store.go`
- `internal/store/postgres/search.go`
- `internal/store/postgres/search_test.go`
- `internal/store/postgres/notes_plans.go`
- `internal/store/postgres/notes_plans_test.go`
- `internal/worker/stage2_sources.go`
- `internal/worker/stage2_sources_test.go`
- `internal/recall/assembler.go`
- `internal/recall/assembler_test.go`
- `plans/05_runtime-contracts_ingest-recall-apply.md`
- `plans/06_data-model_and_storage-invariants.md`
- `docs/review-packets/next-agent-scope-safe-stage2-sources.md`

## Tests run

- `gofmt -w internal/core/dto.go internal/store/store.go internal/store/postgres/search.go internal/store/postgres/search_test.go internal/store/postgres/notes_plans.go internal/store/postgres/notes_plans_test.go internal/recall/assembler.go internal/recall/assembler_test.go internal/worker/stage2_sources.go internal/worker/stage2_sources_test.go` - passed.
- `go test ./internal/worker ./internal/recall ./internal/store/postgres` - passed.
- `go test ./...` - passed.
- `make lint` - passed.
- `make check-headers` - passed.
- `git diff --check` - passed.

## Remaining risks

- Stage 2 preparation still runs before a real Stage 1 Codex bridge exists in
  the current worker skeleton. This pass does not enable real Codex.
- `group_shared` is still unavailable to Stage 2 sources. It should stay that
  way until membership-aware filtering is implemented and tested.
- Memory search remains lexical fallback only; embedding and neighborhood
  retrieval are still future work.
- Existing parallel-agent changes were already present in the working tree; this
  pass builds on them and does not separate or revert that prior work.

## Source Review

- Estimated source: first-principles changes from VibeGravity repo contracts,
  in-repo review packets, and existing local store/reasoning patterns.
- Suspected license: project-internal original work plus Go standard library and
  existing pgx usage.
- Similarity risk: low.
- Review required: yes, normal integration review recommended before enabling a
  real Codex bridge.
- Notes: no external project code, GPL-family material, or structured external
  snippets were used.
