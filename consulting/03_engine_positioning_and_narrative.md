# Engine Positioning and Narrative

## Recommended Category

VibeGravity should be positioned as an agent memory engine.

This is clearer than calling it an app, and more specific than calling it infrastructure.

Suggested category language:

> VibeGravity is an agent memory engine that turns raw agent activity into scoped, correctable, reusable memory.

## Why Engine, Not App

An app owns the primary user experience.

VibeGravity does not aim to be the main chat surface. It sits behind Hermes and other agent hosts. Its job is to make memory reliable, inspectable, and reusable.

The visible product surfaces are supporting surfaces:

- CLI and doctor commands;
- HTTP API;
- MCP tools;
- Hermes provider adapter;
- timeline and explain views;
- possible future admin/operator UI.

These surfaces help operate the engine, but they are not the product center.

## Why Not Just a Database

A database stores data.

VibeGravity must decide what becomes memory, what gets suppressed, what remains private, what is superseded, and what should be recalled under a token budget.

The product value lives in these memory behaviors:

- raw and derived separation;
- scope-aware retrieval;
- provenance;
- human correction;
- recall packing;
- graph updates;
- degraded mode;
- replay and evaluation.

## Product Narrative

Agents are becoming long-running collaborators, but they still behave like they wake up with partial amnesia.

VibeGravity gives them a disciplined memory layer:

1. Capture raw activity without pretending raw logs are memory.
2. Derive structured memories behind the scenes.
3. Keep memory scoped to the right audience.
4. Let humans correct memory and see where it came from.
5. Return short, useful context before the next turn.

The emotional promise is continuity without chaos.

## Strategic Wedge

Hermes-first is the wedge.

This is strong because it gives VibeGravity a real host, lifecycle, and operational context. It avoids the trap of building a generic platform before one user path is excellent.

The risk is that Hermes-first may sound too narrow unless the narrative says:

> Hermes is the first proof environment. The product is a reusable memory engine.

## Differentiation

VibeGravity should differentiate on trust and behavior, not feature count.

Strong differentiators:

- explicit private/shared/group/session memory scopes;
- correction as a first-class product behavior;
- explainable provenance;
- budget-aware recall;
- separation of hot path and background reasoning;
- graph updates with supersession;
- replayable quality gates.

Weak differentiators:

- "we store more context";
- "we have vector search";
- "we support many integrations";
- "we have a large ontology";
- "we use an LLM to summarize chats".

## Positioning Statement

For people building or operating long-running AI agents, VibeGravity is an agent memory engine that gives agents scoped, correctable, explainable memory across sessions. Unlike raw transcript storage or generic vector memory, VibeGravity separates raw events from derived memory, keeps private and shared memory boundaries explicit, and returns compact recall packs that agents can actually use.
