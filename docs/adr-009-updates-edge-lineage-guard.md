# ADR-009: Updates Edge Lineage Guard

## Status

Accepted

## Context

`update_memory` is still intentionally unsupported in the write-capable apply
engine. Before enabling it, the edge direction and latest invariant need to be
unambiguous.

VibeGravity stores lineage edges from the newly written memory to the prior
memory. For an `updates` edge:

- `from_memory_id` is the new memory created by the update operation.
- `to_memory_id` is the prior memory being superseded.

The initial migration had a partial unique index for `updates` on
`from_memory_id`. That only guaranteed that one new memory could not update
multiple targets. It did not prevent two new memories from both updating the
same prior memory.

## Decision

The direct edge-level guard for `updates` belongs on `to_memory_id`.

`migrations/000002_create_core_tables.up.sql` now defines
`memory_edges_single_updates_target_idx` as a partial unique index on
`to_memory_id WHERE edge_kind = 'updates'`.

This is only the direct-target guard. The full latest invariant must still be
handled by the future `update_memory` store transaction:

1. Lock and verify the target memory is still active/latest.
2. Write the new memory and mandatory `memory_trace`.
3. Write the `updates` edge from new memory to target memory.
4. Mark the target memory `superseded` and `latest_flag=false`.
5. Commit all changes atomically.

That transaction now exists in the store-backed apply path. It locks and verifies
the target memory as active/latest, writes the replacement memory and mandatory
trace, writes the `updates` edge from replacement to prior memory, supersedes the
prior memory, and commits those changes together. A deterministic retry of an
already completed update is accepted only when the replacement memory, trace, and
edge are all present and match the same target.

## Rollout Note

For a database that already applied the old index, run the timestamped follow-up
migration `000004_fix_updates_edge_target_index` before enabling
`update_memory`:

```sql
DROP INDEX IF EXISTS memory_edges_single_updates_target_idx;

CREATE UNIQUE INDEX memory_edges_single_updates_target_idx
    ON memory_edges (to_memory_id)
    WHERE edge_kind = 'updates';
```

Fresh bootstrap databases receive the corrected index from
`000002_create_core_tables.up.sql`.

## Consequences

- Direct forks where two new memories both update the same prior memory are
  rejected at the database edge layer.
- The database still does not infer a complete lineage root. Latest state must
  remain an explicit transaction rule, not a side effect of the edge table.
- `extend_memory` behavior is unchanged because the partial index only applies
  to `edge_kind = 'updates'`.
