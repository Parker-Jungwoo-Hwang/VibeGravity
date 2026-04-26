# Roadmap

This roadmap is intentionally conservative. VibeGravity is the agent memory
engine behind Hermes Memory, and V1 should not broaden beyond the trust loop
until that loop is proven through real storage and protocol paths.

## Current Stage

Internal use / private-validation candidate hardening.

The current work pack is Hermes Memory trust loop and first-customer
integration.

## V1 Product Goal

V1 proves one felt outcome:

> Hermes remembers the right project context across sessions, shows why it
> remembered it, and lets the operator fix memory once.

The product proof is:

- recall preview;
- explain and timeline;
- correction;
- correction-driven supersession;
- visible scope and provenance;
- idempotent replay;
- degraded-state truthfulness.

## Private Validation Gate

- README explains the product in under five minutes.
- Local demo works.
- Live PostgreSQL trust loop passes without skipping.
- Real Hermes/MCP roundtrip passes.
- Server is loopback-only by default.
- Unsafe exposure requires explicit opt-in.
- Doctor has useful exit codes and JSON output.
- Release gates and known limitations are visible.

Until the live PostgreSQL gate and real Hermes/MCP roundtrip pass, the repo can
be treated as a private-validation candidate, not a proven private-validation
drop.

## Near-Term Roadmap

### 1. Prove DB and protocol correctness

- Keep correction provenance append-safe and explainable.
- Prove replacement memory, mandatory trace, `updates` edge, and prior
  supersession in one PostgreSQL-backed trust path.
- Keep replay idempotency evidence-safe.
- Keep MCP schemas aligned with the core service contract.
- Keep live PostgreSQL gates visible and non-optional for readiness claims.

### 2. Package a 5-minute Hermes Memory demo

- Show project rule recall.
- Show active plan recall.
- Show wrong memory correction.
- Show supersession and next recall using the corrected memory.
- Show explain/timeline provenance.
- Show private/shared scope separation.
- Show stale/degraded status when worker state lags.

### 3. Harden operator-facing proof docs

- Keep `docs/status.md` current.
- Keep `docs/live-postgres-proof.md` current.
- Keep `docs/hermes-mcp-proof.md` current.
- Keep `docs/privacy-and-data-handling.md` honest about limitations.
- Keep `docs/release-checklist.md` as the release gate.

### 4. Prepare first private validation drop

- Use `v0.x.y` tags only.
- Do not claim V1 readiness.
- Include known limitations.
- Keep generated binaries out of source control unless release packaging
  explicitly requires them.
- Test bootstrap instructions from a clean checkout.

## Post-V1 Candidates

These are plausible after the trust loop is proven:

- production deployment guide;
- install/package command for Hermes MCP bootstrap;
- fuller session replay metrics;
- real Codex execution behind explicit configuration;
- broader eval suite for recall quality and correction behavior;
- richer Dreaming maintenance quality;
- export/delete/audit controls;
- broader adapters after Hermes-first semantics are stable.

## Deferred Until Trust Loop Is Proven

- Public beta.
- Product Hunt or Hacker News launch.
- Enterprise pitch.
- Non-Go package-manager distribution.
- Python wrapper.
- Homebrew formula.
- Docker image.
- Large web app or operator UI.
- Generic vector database positioning.
- Marketplace.
- Broad adapter support.
- Claude Code or Codex client integration.
- Dreaming as the headline.
- Document memory as the headline.

## No-Go Signals

- Live PostgreSQL trust-loop gate skips or fails.
- Hermes/MCP proof is not reproducible.
- `agent_private` or `group_shared` scope safety regresses.
- Correction does not suppress outdated memory in later recall.
- Explain/timeline cannot show provenance safely.
- Docs imply public or V1 readiness before proof exists.
- New feature breadth displaces DB/protocol correctness.

## Source Review

Estimated source: current `PLANS.md`, `docs/status.md`, existing roadmap draft,
and VibeGravity public-readiness docs.

Suspected license: none.

Similarity risk: low.

Review required: yes before using this as an external launch roadmap.
