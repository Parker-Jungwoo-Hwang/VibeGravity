# Privacy and Data Handling

VibeGravity is an agent memory engine. It stores raw agent activity, derives
structured memories, and returns scoped recall for later work. That means the
project may handle sensitive project context, operator instructions, private
agent memory, workspace-shared memory, correction history, provenance, and
document-derived context.

This document is a project data-handling guide, not a formal legal privacy
policy.

## Data VibeGravity Handles

VibeGravity can handle:

- raw events from agent turns;
- derived memory objects;
- memory edges such as `updates` and `extends`;
- provenance traces in `memory_trace`;
- human corrections and correction evidence;
- notes and plans;
- documents and document chunks;
- profiles and session summaries;
- worker job state and freshness metadata;
- local eval fixtures and test data.

Operators should assume that prompts, tool results, notes, plans, corrections,
and document text may become sensitive stored data unless they are explicitly
synthetic or redacted before ingestion.

## Storage Model

PostgreSQL is the canonical store for shared, replayable memory behavior.
SQLite and in-memory stores are for tests and lightweight local development
only.

The main storage categories are:

- `raw_events`: immutable input history;
- `memories`: derived structured memory;
- `memory_trace`: provenance for derived memory;
- `memory_corrections`: append-safe correction records;
- `notes` and `plans`: human-authored operator context;
- `documents` and `document_chunks`: supporting context for recall and Stage 2
  reasoning;
- `profiles`: rebuildable current-view snapshots;
- `session_summaries`: compressed session context;
- `memory_groups`: named group visibility boundaries;
- `memory_group_memberships`: actor membership records for group-shared memory;
- `ingest_jobs`: background worker state.
- job payloads: queued worker input, retry state, and error context;
- logs: local process output, CI output, debug messages, and attached report
  logs.

Raw events and derived memories must stay separate. Derived memory must keep
explicit scope and provenance.

## Scope Boundaries

Memory visibility is part of the product contract.

VibeGravity uses these scopes:

- `agent_private`: visible only to the owning actor and allowed operator paths;
- `workspace_shared`: visible within the workspace;
- `group_shared`: visible only when group membership permits it;
- `session_scratch`: short-lived session context.

Privacy-sensitive bugs include:

- `agent_private` memory returned to the wrong actor;
- `group_shared` memory returned without membership;
- workspace or tenant data crossing boundaries;
- explain, timeline, search, MCP, Hermes, or CLI paths exposing memory that
  recall would not be allowed to show;
- stale or superseded memory being presented as current truth.

Report security-sensitive cases through `SECURITY.md`, not a public issue.

## External Processing

The V1 architecture is local-runtime embedding-first and Codex-first for
structured reasoning.

Current rules:

- local runtime is used for embeddings and retrieval helpers;
- text interpretation and graph operations go through schema-first Codex
  reasoning contracts;
- real Codex execution is not the default worker path until explicitly enabled;
- mocked Codex paths and deterministic evals should not require network access;
- documents are supporting context, not the V1 product promise.

When real external reasoning is enabled, operators should treat submitted raw
events, candidate memories, notes, plans, document snippets, and related
provenance as potentially sent to the configured reasoning provider. Do not
enable external reasoning for private data unless that deployment's privacy,
retention, and access requirements are understood.

## Logs, Evals, and Reports

Do not place secrets, credentials, private memory text, raw production database
dumps, or unredacted customer/project data into:

- public GitHub issues;
- pull requests;
- review packets;
- eval fixtures;
- logs attached to reports;
- screenshots or terminal captures;
- generated docs.

Use synthetic data for tests and evals whenever possible. If real examples are
needed for diagnosis, reduce them to the smallest redacted reproduction.

## Corrections and Provenance

Corrections are data. A correction can preserve the original memory, write a
replacement memory, create an `updates` edge, and retain provenance through
`memory_trace`.

Data-handling expectations:

- do not overwrite original provenance to hide a prior memory;
- keep correction artifacts append-safe;
- suppress corrected or superseded memory from normal current recall;
- keep explain and timeline views scope-aware;
- make stale or degraded recall visible instead of presenting it as fresh.

## Development and Testing Guidance

For local development:

- use synthetic fixtures when possible;
- use a scratch PostgreSQL database for live integration tests;
- never run destructive experiments against shared or production databases;
- redact environment variables and database URLs before sharing logs;
- mention when `VIBEGRAVITY_DB_URL` is unset and live gates were skipped.

Default local verification:

```bash
go test ./...
make eval
make lint
make check-headers
git diff --check
```

Live PostgreSQL verification:

```bash
make integration-postgres
```

## Current Limitations

VibeGravity is pre-release software.

Not yet provided as stable product guarantees:

- hosted service privacy commitments;
- enterprise data-processing agreements;
- production retention controls;
- admin UI for export, deletion, or audit review;
- formal incident response SLA;
- default real Codex execution policy for production use.

Until those exist, treat deployments as operator-managed local or private
validation environments.

## Deletion, Export, and Retention

VibeGravity does not yet provide a stable user-facing delete, export, or
retention command. Operators should manage retention at the database and backup
layer during private validation.

Current guidance:

- use scratch databases for demos and tests;
- export with PostgreSQL-native tooling such as `pg_dump` only when the dump can
  be protected as sensitive memory data;
- delete scratch data by dropping the scratch database when validation ends;
- do not promise per-user deletion, selective memory purge, or retention windows
  until product commands and tests exist;
- document any manual deletion performed during a validation run, including
  affected tenant, workspace, actor, and time range.

## Source Review

Estimated source: first-principles documentation based on VibeGravity's plans,
storage invariants, `SECURITY.md`, `SUPPORT.md`, and current trust-loop
contracts.

Suspected license: none.

Similarity risk: low.

Review required: yes before presenting this as a formal legal privacy policy.
