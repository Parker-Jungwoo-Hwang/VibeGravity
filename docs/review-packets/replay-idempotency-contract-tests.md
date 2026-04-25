# Replay Idempotency Contract Tests

## Summary

Agent 3 verified and tightened the replay idempotency contract tests for
evidence-safe memory graph writes. The core contract is strict: replaying the
same `reasoning_job_id` and operation-derived deterministic memory ID is
idempotent only when the new attempt matches the original memory, trace, and
edge evidence.

## Finding or slice fixed

Covered mismatch cases:

- Identical `update_memory` replay succeeds without duplicate memory, trace, or
  `updates` edge rows.
- `update_memory` replay with changed replacement text returns
  `core.ErrConflict`.
- `update_memory` replay with changed raw event IDs returns `core.ErrConflict`.
- `update_memory` replay with changed target memory ID returns
  `core.ErrConflict`.
- `update_memory` replay with changed edge kind returns `core.ErrConflict`.
- `create_memory` replay with the same deterministic memory ID but changed trace
  evidence returns `core.ErrConflict` and must not overwrite the existing trace.
- `extend_memory` replay through `CreateMemoryWithTraceAndEdge` succeeds when
  identical, but changed trace evidence returns `core.ErrConflict` and must not
  overwrite the existing trace.

The source-level guard also requires the Postgres replay path to compare memory
text, fingerprint, confidence, raw event IDs, applied operation JSON, operation
IDs, target memory ID, edge kind, and job evidence before treating an already
applied write as idempotent success.

## Files changed

- `internal/store/postgres/memories_replay_test.go`
- `docs/review-packets/replay-idempotency-contract-tests.md`

## Tests run

- `go test ./internal/store/postgres -run 'TestPostgres(UpdateMemoryReplayRequiresIdenticalEvidence|CreateMemoryReplayRejectsTraceEvidenceOverwrite|ExtendMemoryReplayRejectsTraceEvidenceOverwrite|ReplaySourceContractsRequireFullEvidenceComparison)'`
- `go test ./...`
- `make lint`
- `make check-headers`
- `git diff --check`

Expected behavior:

- Without `VIBEGRAVITY_DB_URL`, live Postgres replay cases skip and the
  source-level contract guard passes.
- With `VIBEGRAVITY_DB_URL`, identical replay attempts pass and changed evidence
  attempts return `core.ErrConflict`.

## Remaining risks

- Live Postgres replay coverage is still conditional on `VIBEGRAVITY_DB_URL`.
  This packet should be re-run against a migrated PostgreSQL database before a
  release-readiness claim.
- The eval in-memory replay store still models idempotent update replay at a
  higher level and does not compare full replay evidence like the Postgres store.

## Source Review

- Estimated source: repo-local contracts, `agent-c-update-memory-lineage-spec.md`,
  and current Postgres replay tests.
- Suspected license: project-owned VibeGravity source.
- Similarity risk: low; no external project code or structured external snippets
  were used.
- Human review required: yes, because these tests define replay-safety behavior
  that must stay aligned with future operation-application storage.
