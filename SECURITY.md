# Security Policy

VibeGravity stores and recalls agent memory. Security issues include ordinary
software vulnerabilities and memory-specific trust failures such as private
memory leakage, incorrect scope checks, provenance loss, and unsafe correction
or supersession behavior.

## Supported Versions

VibeGravity is pre-release software. Security fixes currently target the main
development branch until the first tagged release exists.

| Version | Supported |
|---|---|
| `main` | Yes |
| Tagged releases | Not yet available |

## Reporting a Vulnerability

Please report security issues privately.

Preferred path:

1. Use GitHub Private Vulnerability Reporting if it is enabled for this
   repository.
2. If private reporting is not available, contact the maintainer privately
   before opening a public issue.
3. Do not include sensitive memory contents, private workspace data, secrets,
   tokens, credentials, or raw production database dumps in a report.

If a public issue is the only available route, keep the report minimal and avoid
exploit details until a maintainer has a private follow-up channel.

## What To Report

Please report:

- `agent_private` memory visible to the wrong actor.
- `workspace_shared` or `group_shared` memory returned without the expected
  tenant, workspace, owner, or membership checks.
- Cross-tenant, cross-workspace, or cross-session data exposure.
- Recall, search, explain, timeline, MCP, HTTP, Hermes, or CLI paths that bypass
  scope checks.
- Missing or misleading provenance for operator-visible memory.
- Correction or supersession behavior that hides stale memory, corrupts
  lineage, or loses `memory_trace`.
- Replay or idempotency bugs that duplicate memories, traces, edges, or
  correction artifacts.
- SQL injection, command injection, path traversal, deserialization, auth,
  secret handling, or dependency vulnerabilities.
- Logs, errors, eval fixtures, docs, or generated artifacts that expose secrets
  or private user data.

## Trust Boundaries

Security-sensitive boundaries in VibeGravity include:

- tenant and workspace isolation;
- `agent_private`, `workspace_shared`, `group_shared`, and `session_scratch`
  scope separation;
- group membership checks for group-shared memory;
- raw event storage versus derived memory storage;
- worker reasoning output versus apply-engine validation;
- correction artifacts, replacement memories, `updates` edges, and
  supersession transactions;
- local embedding runtime configuration;
- Codex bridge configuration and disabled-by-default real network execution;
- MCP stdio and HTTP request input schemas.

## Safe Testing Guidelines

- Use a scratch database for security testing.
- Do not test against production or shared databases without explicit
  authorization.
- Do not exfiltrate, publish, or retain private memory contents.
- Prefer minimal proof-of-concept inputs over large raw transcripts.
- Redact tenant IDs, workspace IDs, actor IDs, secrets, and private text unless
  they are required to reproduce the issue.

## Disclosure Expectations

Maintainers should acknowledge a private vulnerability report, triage severity,
and coordinate a fix before public disclosure. Until a formal release process is
created, security fixes should include:

- a focused regression test when practical;
- documentation of affected trust boundaries;
- notes on whether live PostgreSQL verification is required;
- a changelog entry once releases are tagged.

## Source Review

Estimated source: first-principles security policy based on the VibeGravity
plans, repository trust-loop invariants, and local review documents.

Suspected license: none.

Similarity risk: low.

Review required: yes, before using this as a public security contact policy,
because the project still needs a confirmed private reporting channel.
