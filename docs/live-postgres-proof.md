# Live PostgreSQL Proof

Status: not proven.

This file records the private-validation live PostgreSQL trust-loop gate. Do not
claim V1 readiness until this proof passes against a migrated scratch database.

## Setup

```bash
createdb vibegravity_integration
export VIBEGRAVITY_DB_URL='postgres://localhost:5432/vibegravity_integration?sslmode=disable'
export VIBEGRAVITY_MIGRATION_PATH="$(pwd)/migrations"
migrate -path "$VIBEGRAVITY_MIGRATION_PATH" -database "$VIBEGRAVITY_DB_URL" up
make integration-postgres
```

## Required Evidence

- CorrectMemory works on live DB.
- Replacement memory is created.
- `memory_trace` is created.
- `updates` edge is created.
- Old memory is superseded.
- Retry does not create duplicate memory.
- ExplainMemory shows provenance.
- Timeline shows correction flow.
- Next Prefetch suppresses stale memory.
- Scope separation still holds.

## Latest Result

Command run:

```bash
make integration-postgres
```

Output:

```text
Skipping live PostgreSQL integration gate: VIBEGRAVITY_DB_URL is not set.
Prepare a migrated scratch DB, export VIBEGRAVITY_DB_URL, then rerun: make integration-postgres
```

Exact blocker: `VIBEGRAVITY_DB_URL` was unset, so the live PostgreSQL trust-loop
proof did not run in this environment.
