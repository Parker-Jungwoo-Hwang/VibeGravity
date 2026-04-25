# VibeGravity Product Consulting Request

Date: 2026-04-25

## One-Line Ask

Please review VibeGravity as a product, not just as a technical system, and tell us whether the current direction can become a compelling first product.

## Project Summary

VibeGravity is a shared memory engine for Hermes and other AI agents.

It records agent activity as raw events, derives useful structured memories, and returns compact recall packs before the next turn. It is not a chat UI, not a standalone coding agent, and not a general agent runtime.

The current product thesis is:

> Agents need persistent, scoped, correctable memory. VibeGravity should be the memory engine that lets Hermes and later agent surfaces remember important facts, decisions, plans, preferences, and corrections across sessions.

## Why We Want Product Owner Consulting

The engineering direction is becoming clear, but the product framing needs senior judgment.

We want an excellent Product Owner to pressure-test:

- whether this should be positioned as an engine, developer tool, agent infrastructure layer, or packaged Hermes feature;
- what the first lovable V1 should include;
- whether Hermes-first is enough as a wedge;
- what user-facing surface is required for trust, correction, and adoption;
- what should be cut until after V1;
- what metrics prove that the memory engine is actually valuable.

## Current Working Assumption

VibeGravity should be treated primarily as an engine.

The product surface may include CLI, MCP tools, doctor commands, timeline/explain views, and later a small operator UI. But the core value is not a front-end app. The core value is reliable memory behavior for agent hosts.

## Requested Consulting Output

Please return a concise but decisive product review with:

1. Product category recommendation.
2. Ideal first customer and first use case.
3. V1 scope recommendation.
4. What to remove or delay.
5. Must-have operator/user surfaces.
6. Trust and safety requirements for memory correction, provenance, and private/shared scope separation.
7. Adoption strategy for Hermes-first rollout.
8. Success metrics for V1.
9. Top product risks.
10. A 30-day product execution plan.

## Decisions We Need Help With

- Should VibeGravity be sold or explained as "agent memory infrastructure", "shared memory kernel", "Hermes memory engine", or something else?
- Is the first user Hermes itself, the person operating Hermes, or coding agents that connect through MCP?
- Does V1 need a visible timeline/explain UI, or are CLI/MCP/HTTP enough?
- How much "dreaming" and memory promotion is required before the product feels real?
- Should document memory, plans, notes, and corrections all be in V1, or should V1 narrow further?
- What is the smallest demo that makes the value obvious in under 5 minutes?

## Important Constraints

- Hermes-first delivery is currently the primary wedge.
- Local runtime is embedding-only in V1.
- Codex is the reasoning backend for text interpretation and graph operations.
- PostgreSQL is the canonical shared store.
- Memory scopes must stay separate: `agent_private`, `workspace_shared`, `group_shared`, and `session_scratch`.
- Raw events and derived memories must stay separate.
- Every memory must have provenance.
- Human correction is first-class.

## Materials Included

This folder contains 10 files total, including this request. The recommended reading path starts with `01_reading_order.md`.

The packet is intentionally product-oriented. It summarizes the existing planning docs and current implementation status so the reviewer can focus on product judgment instead of code archaeology.
