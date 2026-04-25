# Agent C Review Packet: `update_memory` Lineage and Latest Specification

## Summary

This packet defines the required lineage, latest-state, provenance, idempotency, and rollback rules that must be satisfied before any `update_memory` write implementation is enabled.

`update_memory` remains a future write path. This document is specification-only and does not authorize weakening `NoopApplyEngine` validation or allowing `StoreBackedApplyEngine` to write update operations before the transaction and tests below exist.

The core model is:

- An `update_memory` operation writes a **new derived memory row**.
- The new memory **supersedes exactly one prior latest memory**.
- The lineage edge direction is always **new memory -> prior memory**.
- The prior memory is made non-latest in the same transaction.
- The new memory and its `memory_trace` must commit atomically with the `updates` edge and prior-memory status change.

Explicit edge interpretation:

- `from_memory_id` = the newly-created memory produced by this `update_memory` operation.
- `to_memory_id` = the existing target memory being superseded.

This follows ADR-009 and the storage invariant that the direct `updates` target guard belongs on `memory_edges(to_memory_id) WHERE edge_kind = 'updates'`.

## Proposed invariant

### Invariant: one latest memory per direct update lineage head

For any `update_memory` operation:

1. The operation must target exactly one existing memory.
2. The target memory must be in the same `tenant_id` and `workspace_id` as the job.
3. The target memory must be `status = active` and `latest_flag = true` at the time the transaction locks it.
4. The newly created memory must be inserted as `status = active` and `latest_flag = true`.
5. The target memory must be updated to `status = superseded` and `latest_flag = false` in the same transaction.
6. The `updates` edge must be inserted as:
   - `from_memory_id = new_memory.id`
   - `to_memory_id = target_memory.id`
   - `edge_kind = updates`
7. The direct target uniqueness guard must prevent two different new memories from both directly updating the same target memory.
8. No memory row may be considered successfully written unless its `memory_trace` row is also written.

### Invariant: update is replacement, not mutation-in-place

`update_memory` must never rewrite the target memory text in place. It creates a new memory row and records lineage through `memory_edges`. This preserves provenance, rollback clarity, and timeline/explainability.

### Invariant: update is not extend

`update_memory` means the new memory supersedes the prior one and should become the latest recall candidate. `extend_memory` means additive detail and does not unset the prior memory. Therefore:

- `update_memory` requires an `updates` edge.
- `extend_memory` requires an `extends` edge.
- An `update_memory` operation with an `extends`, `supports`, `contradicts`, or missing edge is invalid.
- An `extend_memory` operation must not perform latest/supersession behavior.

## Transaction rules

The future write implementation must execute each `update_memory` operation inside one database transaction. If multiple operations in one Stage 2 result are applied together, either the whole apply request should be one transaction, or each operation must have explicit idempotency/resume behavior. The safer first implementation is one transaction for the entire apply request.

### Pre-transaction validation

Before opening the write transaction or before mutating rows, the apply engine must retain the existing validation floor:

- operation kind is supported and non-empty
- operation id is present
- operation raw event IDs are inside the apply request raw event bundle
- `profile_delta`, `plan_delta`, trace metadata, operation metadata, and memory metadata are JSON objects where required
- memory payload has kind, artifact class, scope, owner, text, and confidence
- group-shared payload has `group_id`
- `update_memory` payload has a target memory ID
- `update_memory` edge is present and has `edge_kind = updates`
- operation tenant/workspace context is not allowed to cross boundaries

This spec does not weaken `NoopApplyEngine`; `NoopApplyEngine` should continue to validate shape only and not perform writes.

### Required transaction order

For a single `update_memory` operation, execute in this order:

1. **Begin transaction.**
2. **Idempotency check by operation identity.**
   - Look for an already-applied operation for the same `reasoning_job_id` + `operation_id`.
   - If already applied with the same target and new memory identity, return success without writing duplicates.
   - If the same `reasoning_job_id` + `operation_id` exists but points to different payload/target, fail as an idempotency conflict.
   - If there is no durable operation table yet, derive the check from `memory_trace.reasoning_job_id` plus `applied_operations_json.operation_id` and document that this is a temporary query shape.
3. **Lock the target latest memory.**
   - Select target memory by ID, tenant, and workspace using a row lock (`FOR UPDATE` on PostgreSQL).
   - Verify the locked row is still `status = active` and `latest_flag = true`.
   - If the target is missing, cross-tenant, cross-workspace, superseded, archived, deleted, or not latest, reject the operation and roll back.
4. **Validate direct update edge target against the locked target.**
   - `edge.to_memory_id` must equal `memory.target_id` / operation target.
   - `edge.from_memory_id`, if supplied by Stage 2, must either be empty or equal the deterministic new memory ID that the apply engine will insert.
   - The apply engine should prefer generating/confirming the new memory ID, not trusting a hallucinated ID blindly.
5. **Compute/confirm the new memory ID and fingerprint.**
   - The new memory fingerprint should be based on the new memory payload, not the target memory payload.
   - If deterministic IDs are used, they must include `reasoning_job_id` + `operation_id` or an equivalent idempotency key.
6. **Insert the new memory row.**
   - `status = active`
   - `latest_flag = true`
   - explicit `scope`, `owner_entity_id`, `kind`, `artifact_class`, `text`, `confidence`
   - same tenant/workspace as the job
   - same `group_id` rules as `create_memory`
7. **Insert the mandatory memory trace for the new memory.**
   - `memory_id = new_memory.id`
   - `raw_event_ids = operation.raw_event_ids`
   - `reasoning_job_id = apply_request.job_id`
   - `reasoning_stage = resolve`
   - `candidate_snapshot_json = Stage 1 output used to produce the operation`
   - `applied_operations_json = the exact structured operation or operation subset applied for this memory`
   - `operator_correction_flag = true` only when the source operation is an explicit human correction path; otherwise false
   - `related_document_ids` populated from Stage 2 document references when available, otherwise empty
8. **Insert the `updates` edge.**
   - `from_memory_id = new_memory.id`
   - `to_memory_id = target_memory.id`
   - `edge_kind = updates`
   - `confidence = operation.edge.confidence` or a validated default policy
   - `created_by_job_id = apply_request.job_id`
   - The database partial unique index on `to_memory_id WHERE edge_kind = 'updates'` must reject a second direct update to the same target.
9. **Supersede the target memory.**
   - Set `status = superseded`.
   - Set `latest_flag = false`.
   - Set `valid_to` to the new memory `valid_from` or transaction timestamp, using one consistent policy.
   - Update `updated_at`.
10. **Optionally update profile/session/plan only after memory/trace/edge succeed.**
    - For the first `update_memory` write slice, profile/session/plan deltas may remain rejected as in the current store-backed apply slice.
11. **Commit.**
12. **Return written IDs.**
    - Include the new memory ID, target memory ID, and edge identity in the apply result/logging path.

### Locking and concurrency rules

- The target latest row lock is mandatory. Do not rely only on the unique edge index.
- The row lock protects the semantic latest check before supersession.
- The unique edge index protects the direct lineage target from race-created forks.
- If two transactions attempt to update the same target:
  - one may acquire the lock first and commit;
  - the second must re-check after waiting and then reject because the target is no longer `active/latest`, or fail on the unique edge guard;
  - the second must not create a dangling memory or trace.
- A target memory that is already `superseded`, `archived`, or `deleted` is not valid for `update_memory`.
- Cross-scope updates should be rejected unless a future ADR explicitly permits them. The default rule is that the new memory keeps the same scope and group boundary as the target unless the operation is a validated correction flow with an explicit scope policy.

### Failure rollback rules

Any failure before commit must roll back all side effects for that operation/apply transaction:

- If new memory insert succeeds but trace insert fails, roll back the new memory.
- If memory and trace insert succeed but edge insert fails, roll back memory and trace.
- If edge insert succeeds but target supersession fails, roll back memory, trace, and edge.
- If target supersession succeeds but commit fails, transaction rollback semantics must leave no partial latest change.
- If idempotency conflict is detected, roll back and return a deterministic validation/apply error.
- If the operation is unsupported by the current write slice, return `core.ErrNotImplemented` so worker can block rather than retry forever.

### Idempotency rules

`update_memory` must be safe under worker retry and job replay.

Required behavior:

1. Replaying the same job and same `operation_id` after a successful commit must not create another new memory, trace, or edge.
2. Replaying after a transaction rollback should perform the write exactly once.
3. Replaying the same `operation_id` with different payload, target, edge kind, or raw event IDs must fail as an idempotency conflict.
4. The idempotency check must happen before locking/writing when possible, and must be repeated or protected inside the transaction to avoid races.
5. The trace must contain enough structured operation evidence to explain replay behavior.

Suggested first-slice implementation strategy:

- Use deterministic new memory IDs from `job_id + operation_id` for update-created memories, or add a durable operation-applied table before enabling updates.
- If deterministic IDs are not acceptable, define a unique operation application key before implementation begins. Do not rely on random memory IDs plus best-effort trace search alone for long-term correctness.

### Recall/latest behavior after commit

After a successful update:

- Default recall should suppress the superseded target memory because it is no longer `active/latest`.
- Timeline/explain endpoints should still be able to show the superseded target through trace and `updates` edge lineage.
- The new memory becomes the default latest candidate.
- The old memory remains stored for provenance and correction/explainability.

## Required tests for future implementation

The implementation team should add tests before writing production code. Minimum required test set:

### Validation-floor tests

1. Reject `update_memory` with missing target.
2. Reject `update_memory` with missing edge.
3. Reject `update_memory` with edge kind other than `updates`.
4. Reject `update_memory` whose raw event IDs are outside the apply bundle.
5. Reject `update_memory` with missing memory kind/artifact class/scope/owner/text/confidence.
6. Reject `group_shared` update without `group_id`.
7. Confirm `NoopApplyEngine` validation behavior is not weakened.

### Transaction success tests

1. Updating an active/latest target creates one new active/latest memory.
2. The prior target becomes `status = superseded` and `latest_flag = false`.
3. An `updates` edge is written from the new memory to the prior memory.
4. A `memory_trace` row is written for the new memory.
5. The operation uses the apply request job ID as `created_by_job_id` / `reasoning_job_id`.
6. The new memory and prior memory remain in the same tenant/workspace.

### Edge direction tests

1. Assert `memory_edges.from_memory_id` equals the new memory ID.
2. Assert `memory_edges.to_memory_id` equals the superseded target memory ID.
3. Assert the direct unique guard rejects a second `updates` edge to the same `to_memory_id`.
4. Assert two different targets can each be updated once.

### Latest/concurrency tests

1. Reject update when target is already `superseded`.
2. Reject update when target is `active` but `latest_flag = false`.
3. Reject update when target is `archived` or `deleted`.
4. Simulate two concurrent updates of the same target; exactly one commits and no dangling memory/trace remains from the loser.
5. Verify target row is locked/rechecked before supersession.

Current live-Postgres coverage:

- `internal/store/postgres/concurrency_integration_test.go` adds `TestPostgresConcurrentUpdateMemoryAllowsOneWinnerNoDanglingWrites`.
- The test is skipped unless `VIBEGRAVITY_DB_URL` is set because it verifies real PostgreSQL row-lock and unique-index behavior.
- It launches 16 concurrent update attempts against one active/latest target and asserts exactly one active/latest successor, one `updates` edge with trace, the target marked superseded/non-latest, and zero dangling losing memory/trace rows.
- This is a load smoke test, not a benchmark. Keep adding heavier replay/benchmark coverage before claiming high-load production readiness.

### Idempotency/retry tests

1. Apply the same job and operation twice; the second run returns success/no-op without duplicate memory, trace, or edge.
2. Replay same `operation_id` with different payload; fail as idempotency conflict.
3. Replay after injected failure before commit; retry writes exactly once.
4. Replay after edge unique violation caused by already-applied operation; resolve as idempotent only if operation evidence matches, otherwise fail.

### Rollback tests

1. Inject trace insert failure; assert no new memory remains.
2. Inject edge insert failure; assert no new memory/trace remains.
3. Inject target supersession failure; assert new memory/trace/edge are rolled back and target remains active/latest.
4. Inject commit failure if the test harness supports it; assert no observable partial state.

### Recall/explain tests

1. Recall/search excludes the superseded target by default.
2. Recall/search includes the new latest memory.
3. Explain-memory lineage can traverse from new memory to prior memory through `updates` edge.
4. Memory trace for the new memory contains Stage 1 candidate snapshot and applied operation evidence.

## Open questions

1. Should the first write implementation use deterministic memory IDs derived from `job_id + operation_id`, or should it introduce a dedicated operation-application/idempotency table?
2. Should `valid_to` on the target use the new memory `valid_from`, transaction timestamp, or source raw event `occurred_at`?
3. Are scope changes during `update_memory` ever valid, or should scope/group boundary always match the target in v1?
4. Should profile/session/plan deltas remain rejected for the first `update_memory` slice, matching the current narrow write-capable apply approach?
5. How should operator corrections be distinguished in Stage 2 operations so `operator_correction_flag` is set only for correction-derived updates?
6. Should update lineage eventually enforce one latest per transitive lineage root, or is direct-target latest plus target-row lock sufficient for v1?
7. Should an update to a target with existing `extends` children require special handling, or are `extends` children preserved as historical/additive context under the superseded target?

## Source Review

- Estimated source: project-internal requirements from `AGENTS.md`, ADR-009, Work Pack 03 notes, runtime contracts, storage invariants, and current review packet guidance.
- Suspected license: project-internal original specification.
- Similarity risk: low; no external project code or structured external snippets were used.
- Human review required: yes. This spec should be reviewed before any `update_memory` write implementation, migration, or store interface change begins.
- Notes: this packet intentionally avoids editing implementation, store, graph, worker, migration, coordination-log, or review-index files.
