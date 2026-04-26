# Rollback Guide

This guide covers release rollback for Hermes Memory / VibeGravity. Rollback
must protect scoped memory, provenance, corrections, and operator trust.

## Rollback Principles

- Prefer stopping rollout before applying database rollback to live data.
- Take a database backup or snapshot before every release migration.
- Never assume a migration is safely reversible just because a `.down.sql` file
  exists.
- Keep old binaries available until the new release has passed live gates.
- Record whether rollback used binary downgrade, migration down, restore from
  backup, or roll-forward fix.

## Pre-Release Backup

Before applying migrations for a release candidate:

```bash
pg_dump "$VIBEGRAVITY_DB_URL" > "backup-before-v0.x.y.sql"
```

For larger or production-like databases, use the platform snapshot mechanism and
record the snapshot id in the release notes.

## Binary Rollback

1. Stop the new `vibegravity`, server, and worker processes.
2. Restore the previous binary artifact and checksum-verify it.
3. Restart the previous processes with the previous env file.
4. Run the local CLI smoke:

```bash
bin/vibegravity version
bin/vibegravity eval demo
```

5. If the previous binary does not understand the current migrated schema, stop
   and choose either migration rollback or backup restore.

## Migration Rollback

Use `docs/migration-rollback-matrix.md` before running any down migration.

If the matrix says `Safe on empty or scratch DB only`, do not run it against
operator data unless a restore point exists and the release owner accepts data
loss.

General command shape:

```bash
migrate -path "$VIBEGRAVITY_MIGRATION_PATH" -database "$VIBEGRAVITY_DB_URL" down 1
```

After a migration rollback:

```bash
migrate -path "$VIBEGRAVITY_MIGRATION_PATH" -database "$VIBEGRAVITY_DB_URL" version
make integration-postgres
```

If integration fails after rollback, restore from backup instead of stacking
manual SQL fixes.

## Hermes/MCP Rollback

If a release registered a new MCP command with Hermes, remove or replace it:

```bash
hermes mcp remove vibegravity
```

If the local Hermes CLI uses a different command, inspect:

```bash
hermes mcp --help
```

Record the exact rollback command and result in `docs/hermes-mcp-proof.md`.

## Roll-Forward Decision

Roll forward instead of rolling back when:

- a migration has already transformed live operator data in a way that its down
  file would drop or mis-shape;
- the failure is in docs, packaging, checksums, or tag metadata only;
- the previous binary cannot safely run against the current schema;
- backup restore would lose accepted operator work and a small patch can repair
  the issue faster.

Roll-forward patches still need a new SemVer tag. Do not move the broken tag.

## Incident Record

For every rollback, record:

- release tag;
- commit SHA;
- database version before rollback;
- database version after rollback;
- backup or snapshot id;
- commands run;
- gate results after rollback;
- remaining risk;
- follow-up issue or review packet.

## Source Review

Estimated source: first-principles rollback guide from VibeGravity's PostgreSQL,
Hermes/MCP, and Go-binary release constraints.

Suspected license: none.

Similarity risk: low.

Review required: yes before first public tag.
