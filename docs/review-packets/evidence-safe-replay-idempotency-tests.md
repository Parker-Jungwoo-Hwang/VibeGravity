# Evidence-Safe Replay Idempotency Tests

## Summary

Agent 4 added contract tests for replay idempotency where deterministic retries
must be accepted only when the replayed evidence is identical. The tests are
intentionally strict: changed payload evidence for the same deterministic memory
ID must return `core.ErrConflict`, not silently succeed as an idempotent replay.

## Finding or slice fixed

Covered replay mismatch cases:

- `update_memory` identical replay: same `reasoning_job_id`, operation-derived
  memory ID, replacement text, fingerprint, raw event IDs, applied operation JSON,
  target memory ID, and edge kind should be idempotent success.
- `update_memory` changed replacement text should return `core.ErrConflict`.
- `update_memory` changed raw event IDs should return `core.ErrConflict`.
- `update_memory` changed target memory ID should return `core.ErrConflict`.
- `update_memory` changed edge kind should return `core.ErrConflict`.
- `create_memory` deterministic retry with changed trace evidence should return
  `core.ErrConflict`; `memory_trace` must not be overwritten silently on
  `memory_id` conflict.

The live Postgres tests are skipped unless `VIBEGRAVITY_DB_URL` is set. A
source-level guard keeps the missing-hook risk covered even without a live
database by requiring the replay path to call strict memory, trace, and edge
evidence comparison helpers and by rejecting `memory_trace` overwrite-on-conflict
behavior.

## Files changed

- `internal/store/postgres/memories_replay_test.go`
- `docs/review-packets/evidence-safe-replay-idempotency-tests.md`

## Tests run

- `go test ./internal/store/postgres -run 'TestPostgres(UpdateMemoryReplayRequiresIdenticalEvidence|CreateMemoryReplayRejectsTraceEvidenceOverwrite|ReplaySourceContractsRequireFullEvidenceComparison)'`

Expected result:

- Without `VIBEGRAVITY_DB_URL`, the live Postgres replay cases skip and the
  source-level contract guard must pass.
- With `VIBEGRAVITY_DB_URL`, identical `update_memory` replay must pass, and
  changed replacement text, raw event IDs, target memory ID, edge kind, or create
  trace evidence must return `core.ErrConflict`.

## Remaining risks

- The changed edge kind case may still reach the storage method as invalid input
  depending on where Agent 5 lands replay-aware classification; the behavioral
  contract in this packet expects `core.ErrConflict` for same job/operation
  changed evidence.
- These tests should be run against a migrated Postgres database before closing
  the implementation follow-up, because the strongest cases are integration
  contracts.

## Source Review

- Estimated source: repo-local code and review instructions only.
- Suspected license: project-owned VibeGravity source.
- Similarity risk: low; tests were written from the local storage contract and
  current implementation behavior.
- Human review required: yes, because these tests intentionally define the
  Agent 5 implementation contract and may fail until that fix lands.
