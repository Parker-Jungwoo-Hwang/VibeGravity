# Test Gates

VibeGravity keeps the default local gate deterministic. `go test ./...` must pass
without a live database; database-backed tests skip unless `VIBEGRAVITY_DB_URL`
is set.

## Local deterministic gate

```bash
go test ./...
make eval
make lint
make check-headers
git diff --check
```

This gate catches local contract drift, golden replay regressions, lint issues,
header policy drift, and whitespace errors. It does not prove real PostgreSQL
locking, foreign keys, transaction rollback, or extension availability.

## Live PostgreSQL gate

Use a scratch database. Do not point this gate at a shared or production DB.

1. Create a scratch database and enable required extensions.

   ```bash
   createdb vibegravity_integration
   ```

2. Run migrations with `golang-migrate`.

   ```bash
   export VIBEGRAVITY_DB_URL='postgres://localhost:5432/vibegravity_integration?sslmode=disable'
   migrate -path migrations -database "$VIBEGRAVITY_DB_URL" up
   ```

3. Run the opt-in gate.

   ```bash
   make integration-postgres
   ```

With `VIBEGRAVITY_DB_URL` unset, `make integration-postgres` prints an explicit
skip message and exits successfully. With it set, the target runs live
PostgreSQL tests in `./internal/store/postgres`, the kernel trust-loop
integration tests in `./internal/kernel`, and the DB-backed smoke tests in
`./tests`.

## Trust-loop gates to cover

The live PostgreSQL gate exists because the highest-risk trust-loop behavior
depends on real database semantics:

- correction supersession foreign-key safety for replacement memory, trace, and
  `updates` edge writes;
- concurrent `update_memory` races where exactly one worker wins and losing
  workers leave no dangling memory, trace, or edge rows;
- replay idempotency for deterministic graph operations;
- `explain_memory` provenance loading against real trace rows;
- read-only `timeline` visibility over memories, traces, and correction
  artifacts;
- recall/search suppression of corrected or superseded memory.

Current automated live coverage includes the Postgres concurrency smoke, the
kernel correction trust-loop test, replay idempotency checks, and the DB health
smoke. Remaining trust-loop bullets should continue to land as guarded live
tests as those flows become safe to exercise end to end.
