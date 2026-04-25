# Team 1 Review Packet: Graph Apply Safe Lineage Write

## Summary

Team 1 expanded `StoreBackedApplyEngine` one safe lineage step beyond `create_memory` by implementing write-capable `extend_memory`.

The slice keeps the apply boundary conservative:

- `NoopApplyEngine` validation was not weakened.
- `extend_memory` still must pass the existing validation floor, including a required target and an `extends` edge.
- The write path creates one derived memory, one mandatory `memory_trace`, and one `extends` edge.
- The target memory is left alive; no supersession/latest demotion behavior is attempted.

## Files changed

- `internal/graph/store_apply.go`
  - Added `extend_memory` handling to `StoreBackedApplyEngine`.
  - Expanded the storage dependency with an atomic memory+trace+edge write method.
  - Kept `update_memory`, `archive_memory`, profile, summary, plan, and group-shared writes rejected.
- `internal/graph/store_apply_test.go`
  - Added TDD coverage for successful `extend_memory` writes.
  - Added coverage that edge persistence failure returns no successful apply result.
  - Updated unsupported-write expectations for update/archive to document why they remain rejected.
- `internal/store/postgres/memories.go`
  - Added `CreateMemoryWithTraceAndEdge` to write memory, trace, and edge in one PostgreSQL transaction.
  - Refactored edge upsert into a transaction-capable helper.
- `docs/review-packets/team-1-graph-apply.md`
  - This review packet.

## Behavior added

`extend_memory` now writes through the store-backed apply engine when validation succeeds:

1. Builds a deterministic new memory ID from tenant, workspace, job, and operation ID.
2. Writes the new memory as `active` and `latest_flag=true`.
3. Writes a mandatory `memory_trace` for the new memory using the operation raw event IDs and resolve-stage provenance.
4. Writes a `memory_edges` row from the new memory to the target memory with `edge_kind='extends'`.
5. Performs the memory, trace, and edge write atomically in PostgreSQL via `CreateMemoryWithTraceAndEdge`.

The extension edge originates from the actual written memory ID, not from a model-supplied `from_memory_id`. The target memory remains untouched.

## Explicitly rejected behavior

This slice intentionally still rejects:

- `update_memory`: latest/supersession behavior is still uncertain, and updates must eventually demote or otherwise resolve prior latest state safely.
- `archive_memory`: archive status writes and recall suppression behavior are still outside this slice.
- `group_shared` memory writes: membership validation is not implemented yet.
- Non-empty `profile_delta`, `session_summary`, or `plan_delta` writes.
- Natural-language extraction.
- Real Codex calls.
- Profile merge or session summary writes.
- Any raw-event mutation or blending of raw events into derived memory rows.

## Tests run

Targeted RED check before implementation:

```bash
go test ./internal/graph -run 'TestStoreBackedApplyEngine_(WritesExtendMemoryWithTraceAndEdge|ExtendEdgeFailureDoesNotReportSuccessfulApply)' -count=1
```

Result: failed as expected because `extend_memory` was still validation-only.

Targeted GREEN checks after implementation:

```bash
go test ./internal/graph -run TestStoreBackedApplyEngine -count=1
go test ./internal/graph -count=1
```

Result: passed.

Final requested verification:

```bash
gofmt -w internal/graph/store_apply.go internal/graph/store_apply_test.go internal/store/postgres/memories.go && go test ./... && make lint && make check-headers && git diff --check
```

Result: passed.

## Risks

- `extend_memory` creates the extension memory as active/latest and leaves the target memory unchanged. This matches the current safe interpretation of `extends keeps prior memory alive`, but richer lineage/latest query semantics still need a later design pass.
- Target existence is enforced by PostgreSQL foreign keys when using the canonical store. The apply engine itself does not perform a preflight target lookup in this slice.
- The storage interface used by `StoreBackedApplyEngine` is now narrower than the general `store.MemoryStore` interface and includes the atomic graph write method locally; that was intentional to avoid editing broader shared store contracts.

## Next recommended slice

Implement `update_memory` only after latest/supersession behavior is made explicit. The next slice should define and test:

- how the previous latest memory is demoted or marked superseded,
- whether an `updates` edge target must be active/latest at write time,
- transaction order for new memory, trace, updates edge, and target latest/status changes,
- replay/idempotency behavior when the same update operation is applied more than once.

If that uncertainty is not resolved, the safer next improvement is target preflight validation for `extend_memory` using a read-capable store method, without changing latest state.

## Source Review

- Estimated source: implemented from the in-repo project plans, existing graph apply code, and storage contracts.
- Suspected license: project-internal original work.
- Similarity risk: low; no external code or long snippets were used.
- Human review required: normal project review recommended, especially around the choice to leave target latest/status untouched for `extend_memory` and to rely on PostgreSQL FK enforcement for target existence.
