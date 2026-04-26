# Release Process

This document is the source of truth for VibeGravity release discipline.
Release work is conservative because Hermes Memory stores scoped agent memory,
operator corrections, provenance, and potentially sensitive project context.

## Version Policy

VibeGravity uses SemVer tags in the form `vMAJOR.MINOR.PATCH`.

Use only `v0.x.y` releases until the V1 tag conditions in this document are
met. A `v0.x.y` release may be useful for private validation, but it is not a
claim that the V1 Hermes Memory trust loop is broadly ready.

For `v0.x.y`:

- bump `PATCH` for docs-only changes, bug fixes, gate fixes, and compatible
  hardening;
- bump `MINOR` for new compatible commands, API fields, MCP tools, storage
  surfaces, or operator workflows;
- do not use `MAJOR` before V1;
- prefer pre-release suffixes such as `v0.2.0-rc.1` for release candidates;
- do not reuse a tag after it has been pushed.

## V1 Tag Conditions

Do not tag `v1.0.0` until every condition below is true and documented in the
release notes.

- Live PostgreSQL gate passes on a freshly migrated scratch database.
- The live PostgreSQL gate does not skip because `VIBEGRAVITY_DB_URL` is unset.
- Correction writes replacement memory, mandatory `memory_trace`, `updates`
  edge, target supersession, and applied correction status in one transaction.
- Retrying the same correction idempotency key does not create duplicate graph
  rows or correction artifacts.
- Normal current recall suppresses superseded memory.
- `explain_memory` and `view_timeline` show correction and preserved
  provenance without leaking private or group-scoped memory.
- Real Hermes/MCP smoke passes for `recall_preview`, `correct_memory`,
  `explain_memory`, `view_timeline`, and `degraded_status`.
- `agent_private`, `workspace_shared`, and `group_shared` scope behavior is
  proven through PostgreSQL-backed paths.
- Degraded or stale recall is labeled honestly.
- Release notes, rollback guidance, migration rollback status, checksums, and
  known risks are complete.
- Public documentation does not overclaim public launch, public beta, hosted
  service, or enterprise readiness.

## Required Release Files

Each release candidate must update or review:

- `CHANGELOG.md`
- `docs/release-checklist.md`
- `docs/release-notes-template.md`
- `docs/rollback-guide.md`
- `docs/migration-rollback-matrix.md`
- `docs/live-postgres-proof.md`
- `docs/hermes-mcp-proof.md`

`CHANGELOG.md` entries must use these sections in this order:

- Added
- Changed
- Fixed
- Security
- Migration Notes
- Known Risks

Use an empty section only when the release owner explicitly records `None`.

## Preflight

Start every release candidate with:

```bash
git status --short --branch
git diff --check
```

Public release candidates require a clean worktree before artifacts are built or
tags are created. Private validation drops may be prepared from a dirty
worktree only when the release notes list the exact dirty files and the drop is
not published as a public tag.

Confirm the Go toolchain:

```bash
go version
go env GOVERSION GOTOOLCHAIN
```

## Local Gates

Run:

```bash
make release-gate
```

The gate includes tests, evals, lint, header checks, whitespace checks, builds,
module verification, and `govulncheck ./...`.

## Live Gates

Live PostgreSQL is mandatory before any release is called proven:

```bash
export VIBEGRAVITY_MIGRATION_PATH="$(pwd)/migrations"
migrate -path "$VIBEGRAVITY_MIGRATION_PATH" -database "$VIBEGRAVITY_DB_URL" up
make integration-postgres
```

If `make integration-postgres` skips because `VIBEGRAVITY_DB_URL` is unset, the
release is no-go for V1 or public validation claims.

Hermes/MCP smoke is mandatory before any release is called proven:

```bash
make build
bin/vibegravity hermes bootstrap --name vibegravity --command "$(pwd)/bin/vibegravity"
hermes mcp test vibegravity
```

The release owner must also record trust-loop tool evidence in
`docs/hermes-mcp-proof.md`.

## Artifacts

Current `v0.x` releases are Go binary-first. Migrations are filesystem
artifacts and must ship beside the binary. The artifact set must make the
migration path unambiguous.

Stage artifacts under `dist/`, then run:

```bash
make release-checksums
```

If the release claims SBOM coverage, run:

```bash
make sbom
```

## Annotated Tag Policy

Use annotated Git tags for all release candidates and releases.

```bash
git tag -a v0.x.y -F dist/release-note-v0.x.y.md
git push origin v0.x.y
```

Tag messages must include:

- version;
- release type;
- commit SHA;
- local gate result;
- live PostgreSQL gate result;
- Hermes/MCP smoke result;
- migration rollback summary;
- known risks.

Do not create lightweight release tags. Do not force-push or move a pushed
release tag. If a tag is wrong after push, publish a new patch or release
candidate tag and document the superseded tag.

## Rollback

Every release must include a rollback decision:

- binary rollback target;
- migration rollback status from `docs/migration-rollback-matrix.md`;
- data backup or restore point;
- Hermes/MCP registration rollback command;
- no-go conditions that require stopping rollout instead of rolling forward.

Follow `docs/rollback-guide.md`.

## Source Review

Estimated source: first-principles release discipline from VibeGravity's live
Go-first repository state, release checklist, packaging notes, and current
Hermes Memory trust-loop gates.

Suspected license: none.

Similarity risk: low.

Review required: yes before first public tag.
