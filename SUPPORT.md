# Support

VibeGravity is pre-release software. Support is currently best-effort and
focused on validating the Hermes Memory trust loop: scoped recall, provenance,
correction, supersession, timeline/explain, idempotent replay, and degraded
state visibility.

## Where To Ask

Use GitHub Issues for:

- reproducible bugs;
- documentation gaps;
- local setup problems;
- deterministic eval failures;
- PostgreSQL integration gate failures;
- memory trust-loop regressions;
- feature requests that fit the active V1 product direction.

Use pull requests for:

- focused fixes;
- docs improvements;
- tests or eval scenarios;
- small protocol, storage, or recall improvements that preserve the current
  contracts.

Use `SECURITY.md` for:

- private memory leakage;
- cross-tenant, cross-workspace, or cross-actor data exposure;
- scope-check bypasses;
- secrets or credentials exposure;
- vulnerability reports that should not be public.

Do not open public issues containing private memory contents, credentials,
tokens, raw production database dumps, or exploit details.

## What To Include In A Support Request

For setup and runtime issues, include:

- operating system and Go version;
- command run;
- full error output with secrets redacted;
- whether `go test ./...` passes;
- whether `make eval` passes;
- whether `VIBEGRAVITY_DB_URL` was set;
- whether the issue uses SQLite/in-memory paths or live PostgreSQL.

For memory behavior issues, include:

- tenant, workspace, actor, and group identifiers, redacted if needed;
- memory scope involved: `agent_private`, `workspace_shared`, `group_shared`,
  or `session_scratch`;
- surface involved: HTTP, MCP, Hermes adapter, CLI, worker, or test/eval;
- whether the issue affects recall, search, explain, timeline, correction, or
  supersession;
- expected result;
- actual result;
- minimal reproduction steps.

## Current Support Boundaries

Supported:

- local development setup;
- deterministic test and eval gates;
- scratch PostgreSQL integration gates;
- V1 trust-loop behavior;
- docs, contribution, license, and release-readiness questions.

Not yet supported as stable product surface:

- production deployment guidance;
- hosted service support;
- enterprise support commitments;
- custom Hermes memory provider registry packaging;
- real Codex execution as the default worker path;
- broad integrations beyond the current HTTP, MCP, CLI, and Hermes-facing
  adapter surfaces.

## Response Expectations

This project does not yet have a formal service-level agreement. Maintainers
will prioritize:

1. security and private-memory leakage reports;
2. data integrity, provenance, correction, and supersession bugs;
3. regressions in `go test ./...`, `make eval`, or live PostgreSQL gates;
4. setup and documentation issues;
5. feature requests.

## Before Asking For Help

Run the local deterministic gate when practical:

```bash
go test ./...
make eval
make lint
make check-headers
git diff --check
```

For PostgreSQL behavior, use a scratch database and run:

```bash
make integration-postgres
```

If the live gate skips because `VIBEGRAVITY_DB_URL` is unset, mention that in
the support request.
