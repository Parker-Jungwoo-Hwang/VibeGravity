# Current State and Roadmap

## Snapshot

As of 2026-04-25, VibeGravity is beyond the initial foundation stage but not V1-complete.

The repo has moved into Work Pack 03 and early first-customer integration work.

## Implemented or Scaffolded

- Go-first core contracts.
- PostgreSQL migrations.
- HTTP handlers for the v1 surface.
- `sync_turn()` hot path.
- `prefetch()` typed recall assembler.
- Notes, plans, documents, memory search, document search.
- Store-backed graph apply for `create_memory`, safe `extend_memory`, and guarded `update_memory`.
- Mandatory memory trace for graph writes.
- Worker blocked-job handling for deterministic unsupported apply work.
- Schema-first Stage 1 and Stage 2 reasoning interfaces.
- Disabled-by-default Codex JSON bridge boundary with strict JSON validation.
- Human correction path that records a raw correction event, stores a correction artifact, writes a replacement memory, creates an `updates` edge, and supersedes the prior memory.
- Read-only `GetTimeline`.
- `ExplainMemory`.
- Hermes in-repo provider adapter.
- MCP-style surface and stdio MCP server.
- CLI Hermes bootstrap output for MCP registration.
- Golden eval gate for recall and graph replay scenarios.
- Mocked outage/backlog eval gate.
- Read-only job backlog metrics.

## Still Incomplete

- Real Codex execution is not enabled by default.
- Real Codex prompt builder, retry policy, and operator-facing failure behavior are not finished.
- Full production replay harness is not finished.
- Production-grade dreaming quality is not complete.
- Profile merge and plan delta writes remain narrow.
- Custom Hermes memory provider registry packaging is blocked by current Hermes CLI constraints.
- Full real Hermes runtime roundtrip tests are still needed.
- Production ops, install, backup, and restore flows are still incomplete.

## Current Strategic Tension

The engineering is moving toward a sophisticated memory kernel.

The product still needs a sharper answer to:

- what first user sees;
- what V1 demo proves;
- what surfaces are mandatory for trust;
- how much operator UX is required before Hermes-first rollout;
- whether the product should stay embedded-only or expose a small standalone control plane.

## Suggested Roadmap

### Milestone 1. V1 Trust Slice

Goal: prove the engine remembers, corrects, scopes, and explains.

Deliver:

- end-to-end Hermes recall/sync flow;
- memory correction visible in later recall;
- timeline and explain path usable by operator;
- degraded recall metadata;
- golden evals passing.

### Milestone 2. V1 Operational Slice

Goal: make the engine safe to run locally.

Deliver:

- doctor command;
- install/bootstrap command;
- backup and restore notes;
- backlog visibility;
- real Hermes roundtrip test;
- clear degraded-mode behavior.

### Milestone 3. V1 Product Demo

Goal: make the value obvious in under 5 minutes.

Deliver:

- scripted Hermes session;
- recall before/after correction;
- private vs shared memory example;
- timeline/explain proof;
- failure/degraded mode proof.

### Milestone 4. Post-V1 Expansion

Goal: extend beyond Hermes without changing core semantics.

Possible:

- broader MCP clients;
- Codex as client;
- Claude Code as consumer;
- optional operator UI;
- richer document memory.

## Product Owner Question

Is the roadmap sequenced for product proof, or is it still too engineering-led?
