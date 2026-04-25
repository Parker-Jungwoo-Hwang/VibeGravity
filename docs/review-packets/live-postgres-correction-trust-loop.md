# Live Postgres Correction Trust Loop

## Summary

Added an opt-in live PostgreSQL integration test for the full `CorrectMemory`
trust loop through `kernel.Service` and the real PostgreSQL store.

The gate is guarded by `VIBEGRAVITY_DB_URL` and skips cleanly when that env var
is unset. When pointed at a scratch migrated database, it exercises real
migrations, foreign keys, row state, transaction boundaries, explain, timeline,
search, and prefetch behavior.

## Finding or slice fixed

Local fake-store tests can pass while the real database rejects correction
supersession. The risky fields are:

- `memory_trace.reasoning_job_id`
- `memory_edges.created_by_job_id`

Both are constrained to `ingest_jobs(id)` in PostgreSQL. The new integration
test proves the correction path creates or reuses a real correction-apply job row
before writing the replacement trace and `updates` edge.

The test also locks the product trust-loop behavior:

- seeds a target memory with a real seed job and trace;
- calls the real `CorrectMemory` service path;
- verifies the replacement memory, mandatory trace, `updates` edge, and target
  supersession;
- verifies trace and edge provenance both point at existing `ingest_jobs` rows;
- verifies `ExplainMemory` shows the replacement trace, correction source event,
  and `updates` edge;
- verifies `GetTimeline` shows both the correction artifact and replacement
  memory;
- verifies `SearchMemories` and `Prefetch` return the corrected latest memory
  and suppress the superseded old memory;
- verifies same-key retry does not duplicate memory, trace, edge, correction,
  raw event, or job rows;
- verifies same-key changed-text retry conflicts without adding rows.

## Files changed

- `internal/kernel/correction_trust_loop_integration_test.go`
- `Makefile`
- `tests/README.md`
- `docs/review-packets/live-postgres-correction-trust-loop.md`

## Tests run

Completed without live PostgreSQL:

- `go test ./internal/kernel` - passed; live trust-loop test skipped because
  `VIBEGRAVITY_DB_URL` is unset.
- `env -u VIBEGRAVITY_DB_URL go test ./...` - passed; live DB tests skipped.
- `env -u VIBEGRAVITY_DB_URL make integration-postgres` - passed; printed the
  explicit skip message.
- `make lint` - passed.
- `make check-headers` - passed.
- `git diff --check` - passed.

Live PostgreSQL was not available in this environment at the time this packet
was written, so the live branch of the new test has not been executed here.

To run the live gate:

```bash
export VIBEGRAVITY_DB_URL='postgres://localhost:5432/vibegravity_integration?sslmode=disable'
migrate -path migrations -database "$VIBEGRAVITY_DB_URL" up
make integration-postgres
```

## Remaining risks

- The test assumes the operator provides a scratch migrated DB. It does not
  create or drop databases.
- The new live test is intentionally strict about idempotent retry and changed
  correction text. If it fails against live PostgreSQL, treat that as a
  readiness blocker rather than weakening the test.
- The migration now accepts the current `correction_apply` job kind because the
  FK-safe correction path uses that job row for trace and edge provenance.
- Existing databases that already applied the old `updates` index need
  `000004_fix_updates_edge_target_index` before relying on target-side
  `updates` uniqueness.
- The test does not cover `group_shared` correction visibility; this slice stays
  on `workspace_shared` to prove the trust loop without membership setup.

## Source Review

- Estimated source: in-repo VibeGravity contracts, storage code, and review
  packets.
- Suspected license: project-internal original code and documentation.
- Similarity risk: low; the test is project-specific and written from the local
  schema and service contract.
- Human review required: yes, before using the live gate as a release-readiness
  claim.
