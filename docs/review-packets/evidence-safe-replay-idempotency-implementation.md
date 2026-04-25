# Evidence-Safe Replay Idempotency Implementation

## Summary

Implemented evidence-safe replay handling for Postgres memory graph writes.

The write helpers now treat deterministic memory ID collisions as replay
attempts, not as permission to overwrite existing evidence. A replay is accepted
only when the existing memory row, memory trace, and lineage edge match the new
attempt's semantic evidence. Mismatches return `core.ErrConflict`.

## Finding or slice fixed

The previous path protected row counts but could accept changed replay evidence:

- `memory_trace` used `ON CONFLICT (memory_id) DO UPDATE`.
- deterministic memory ID conflicts could update replacement memory fields.
- update retry success checked for row presence, trace presence, and edge
  presence without comparing the trace and payload evidence.

This slice changes graph write helpers to use insert-or-validate behavior:

- `CreateMemoryWithTrace`
- `CreateMemoryWithTraceAndEdge`
- `CreateMemoryWithTraceAndUpdateEdge`
- `WriteMemoryTrace`

`UpsertMemory` and `UpsertMemoryEdge` retain their broader helper names and were
not expanded into a new operation table in this slice.

## Comparison strategy

For deterministic replay, the store compares:

- replacement memory tenant/workspace/scope/group/owner boundary;
- replacement memory kind, artifact class, text, fingerprint, confidence,
  status, latest flag, and metadata JSON;
- trace `raw_event_ids`;
- trace `reasoning_job_id`;
- trace `reasoning_stage`;
- trace candidate snapshot JSON;
- trace applied operation JSON;
- extracted applied `operation_id` values;
- trace operator-correction flag and related document IDs;
- lineage edge target, edge kind, confidence, and `created_by_job_id`.

JSON comparison is canonicalized before comparison, so formatting-only JSON
differences do not create conflicts. Array order remains significant for raw
event IDs, related document IDs, and applied operations.

## Files changed

- `internal/store/postgres/memories.go`
- `internal/store/postgres/memories_test.go`
- `docs/review-packets/evidence-safe-replay-idempotency-implementation.md`

## Tests run

- `go test ./internal/store/postgres`
- `go test ./internal/graph`
- `go test ./...`
- `make lint`
- `make check-headers`
- `git diff --check`

## Remaining risks

- This slice keeps the trace-based operation lookup strategy. A dedicated
  applied-operation table would make long-term operation identity queries
  simpler, but was intentionally not introduced here.
- `UpsertMemory` and `UpsertMemoryEdge` remain generic helpers. The graph write
  paths now use stricter insert-or-validate helpers.
- Live PostgreSQL replay coverage still depends on `VIBEGRAVITY_DB_URL`; unit
  coverage locks the comparison behavior without requiring a live database.

## Source Review

- Estimated source: project-internal requirements from `AGENTS.md`,
  `plans/05_runtime-contracts_ingest-recall-apply.md`,
  `plans/06_data-model_and_storage-invariants.md`,
  `docs/review-packets/agent-c-update-memory-lineage-spec.md`, ADR-009, and
  live store code.
- Suspected license: project-internal original implementation.
- Similarity risk: low; no external project code or structured external snippets
  were used.
- Human review required: yes, because this changes replay conflict semantics in
  graph persistence.
