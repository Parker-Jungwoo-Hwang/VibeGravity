# Runtime and Product Contract

## Engine Contract

VibeGravity has two main lifecycle calls:

- `sync_turn()` records what happened after a turn.
- `prefetch()` returns useful context before a turn.

Everything else supports those two calls.

## Core Surface

The current v1 service surface includes:

- `Prefetch`
- `SyncTurn`
- `AddDocument`
- `SearchMemories`
- `SearchDocuments`
- `AddNote`
- `CreatePlan`
- `UpdatePlan`
- `CorrectMemory`
- `GetTimeline`
- `ExplainMemory`

Product interpretation:

- `Prefetch` and `SyncTurn` are the engine heartbeat.
- Search, notes, plans, correction, timeline, and explain are trust and control surfaces.
- Documents are supporting context, not the main product.

## Runtime Shape

VibeGravity is designed as:

- Go HTTP API;
- background worker;
- PostgreSQL canonical store;
- local embedding runtime;
- Codex reasoning bridge;
- MCP and Hermes adapter surfaces.

Hot path and background reasoning are separated.

## Reasoning Path

The default worker pipeline is:

1. load raw event bundle;
2. use local embeddings and retrieval helpers;
3. run Codex Stage 1 extraction;
4. run Codex Stage 2 resolve;
5. validate structured operations;
6. apply graph changes through guarded storage writes.

Local LLM extraction is intentionally not the default path.

## Product Invariants

These invariants are product features, not just engineering details:

- raw events and derived memories stay separate;
- every write path is idempotent;
- every memory has provenance;
- every memory has explicit scope;
- recall is budget-aware;
- human correction is first-class;
- `updates` can only target one latest memory at a time;
- group-shared memory requires membership;
- profile is rebuildable from raw events, memories, and edges.

## Memory Scopes

VibeGravity distinguishes:

- `agent_private`: visible only to the owning agent and authorized operator paths;
- `workspace_shared`: visible across the workspace;
- `group_shared`: visible only to explicit group members;
- `session_scratch`: short-lived session context.

This is central to the product. If the scope model is not trusted, the memory engine is not trusted.

## Degraded Mode

If Codex is unavailable:

- new graph updates may pause;
- previous profile, notes, plans, session summaries, and existing memory should still feed recall;
- degraded state should be visible in metadata;
- backlog and freshness loss should be observable.

If embeddings are unavailable:

- lexical fallback should keep service usable;
- recall quality may drop;
- service should not fail closed unless privacy or correctness requires it.

## Product Owner Question

Which of these technical contracts must be visible to the user as product UX, and which can remain invisible infrastructure?
