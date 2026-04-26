# Phase Context

Use `phase_id` to name the project slice a lane belongs to. Keep phase names
stable across handoffs so the leader can synthesize results.

## Current Phase IDs

| Phase ID | Meaning |
|---|---|
| `phase-01-foundation` | repo foundation, contracts, schema floor |
| `phase-02-ingest-recall` | `sync_turn()`, `prefetch()`, retrieval, recall pack |
| `phase-03-reasoning-graph` | Codex Stage 1/2, apply engine, graph writes |
| `phase-04-hermes-memory` | Hermes provider, MCP surface, trust-loop UX |
| `phase-05-quality-ops` | evals, replay, degraded modes, docs hardening |
| `phase-06-release-readiness` | release checks, packaging, public docs, readiness verdict |

## Lane Naming

Use `lane_id` values that are short, concrete, and file-scope friendly:

```text
correction-provenance-tests
mcp-schema-parity
recall-freshness-docs
agent-workflow-docs
release-readiness-report
```

## Phase Rules

- A lane belongs to one phase unless the leader approves synthesis.
- A lane may mention dependencies in other phases, but it must not edit across
  phases without leader approval.
- A release-readiness claim requires gates, not just green focused tests.
- A multi-Hermes run belongs to the phase named by its synthesis packet.
