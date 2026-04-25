# 06 Gpt Pro External Surfaces Hermes Mcp

Generated: 2026-04-25

This file is part of the GPT-Pro review material bundle for VibeGravity.

## Included Sources

- `consulting/00_consulting_request.md`
- `consulting/01_reading_order.md`
- `consulting/02_product_one_pager.md`
- `consulting/03_engine_positioning_and_narrative.md`
- `consulting/04_customer_and_use_cases.md`
- `consulting/05_mvp_scope_and_non_goals.md`
- `consulting/06_runtime_and_product_contract.md`
- `consulting/07_current_state_and_roadmap.md`
- `consulting/08_risks_and_open_decisions.md`
- `consulting/09_consulting_questionnaire.md`
- `internal/hermes/doc.go`
- `internal/hermes/provider.go`
- `internal/hermes/provider_test.go`
- `internal/mcp/doc.go`
- `internal/mcp/protocol.go`
- `internal/mcp/protocol_test.go`
- `internal/mcp/surface.go`
- `internal/mcp/surface_test.go`

## Source Contents


<!-- Source: consulting/00_consulting_request.md | bytes=3641 | lines=79 | sha16=f30291c72745268f -->

```md
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


```



<!-- Source: consulting/01_reading_order.md | bytes=2431 | lines=74 | sha16=e18425867e7503bd -->

```md
# Consulting Packet Reading Order

## Purpose

This packet prepares VibeGravity for a Product Owner-style consulting review.

The goal is not to audit every implementation detail. The goal is to answer:

- What product is this?
- Who is it for first?
- What must V1 prove?
- What should be cut, clarified, or surfaced?

## Recommended 30-Minute Path

Read these first:

1. `00_consulting_request.md`
2. `02_product_one_pager.md`
3. `03_engine_positioning_and_narrative.md`
4. `05_mvp_scope_and_non_goals.md`
5. `08_risks_and_open_decisions.md`
6. `09_consulting_questionnaire.md`

This gives enough context for a strategic product response.

## Recommended 90-Minute Path

Read all files in order:

1. `00_consulting_request.md` - the consulting ask.
2. `01_reading_order.md` - how to use the packet.
3. `02_product_one_pager.md` - product overview.
4. `03_engine_positioning_and_narrative.md` - category and narrative.
5. `04_customer_and_use_cases.md` - first customer and scenarios.
6. `05_mvp_scope_and_non_goals.md` - V1 boundaries.
7. `06_runtime_and_product_contract.md` - engine behavior and invariants.
8. `07_current_state_and_roadmap.md` - implementation status and next milestones.
9. `08_risks_and_open_decisions.md` - risk and decision backlog.
10. `09_consulting_questionnaire.md` - specific questions and requested output format.

## Source Material Used

This consulting packet was synthesized from:

- `PLANS.md`
- `plans/00_read-this-first_for-building-agents.md`
- `plans/01_rfp_vibegravity_hermes-first.md`
- `plans/02_product-contract_and_direction.md`
- `plans/03_target-architecture_codex-first.md`
- `plans/04_memory-scopes_dreaming_ontology-lite.md`
- `plans/05_runtime-contracts_ingest-recall-apply.md`
- `plans/06_data-model_and_storage-invariants.md`
- `plans/07_workpack_foundation-and-repo-setup.md`
- `plans/08_workpack_ingest-and-recall.md`
- `plans/09_workpack_memory-graph-and-dreaming.md`
- `plans/10_workpack_hermes-provider-and-external-surfaces.md`
- `plans/11_workpack_quality-ops-and-evals.md`
- `docs/review-packets/current-state-and-next-agent-handoff.md`

## Review Posture Requested

Please be direct.

We are less interested in generic encouragement and more interested in:

- product category clarity;
- wedge quality;
- V1 ruthlessness;
- trust surface requirements;
- whether the product promise is understandable to a real user;
- what would make the first demo feel unavoidable.


```



<!-- Source: consulting/02_product_one_pager.md | bytes=3121 | lines=89 | sha16=1fedab00ec6fabfe -->

```md
# VibeGravity Product One-Pager

## Product Definition

VibeGravity is a shared memory engine for AI agents.

It lets an agent host record what happened, derive structured memories, and retrieve only the useful context before future work.

## Problem

AI agents lose continuity across sessions.

The user repeats the same rules, preferences, project decisions, and current task state. Multiple agents working in the same workspace do not reliably share what should be shared, and private agent memory can easily blur with workspace memory if the system is not designed carefully.

Current memory approaches often fall into weak patterns:

- raw chat logs become the memory;
- vector search returns noisy context;
- private and shared memory are not strongly separated;
- corrections do not reliably change future behavior;
- users cannot inspect why a memory exists;
- memory gets longer instead of more useful.

## Target User

The first customer is Hermes Agent.

The first human user is the person operating Hermes and expecting continuity across sessions, workspaces, plans, notes, and corrections.

Later users may include coding agents, operators, or other local agent runtimes that connect through HTTP or MCP.

## Solution

VibeGravity sits behind the agent host as a memory engine.

It provides:

- `sync_turn()` after an agent turn to record raw events;
- worker processing to derive structured memory;
- graph apply to create, extend, update, and supersede memories;
- `prefetch()` before the next turn to return a compact recall pack;
- manual tools for search, notes, plans, correction, timeline, and provenance explanation.

## Product Promise

The user should not need to repeat important context.

The agent should remember durable facts, active plans, preferences, decisions, and corrections across sessions while keeping private, workspace, group, and session memory boundaries explicit.

## Why It Matters

If agents become daily collaborators, memory quality becomes product quality.

Bad memory is worse than no memory. It can leak private context, revive outdated facts, ignore user correction, or flood the prompt with noise.

VibeGravity's product bet is that memory needs an engine, not just a database.

## V1 Success Statement

V1 succeeds if Hermes can:

- send turns to VibeGravity without slowing the hot path;
- receive useful recall before the next turn;
- keep `agent_private`, `workspace_shared`, and `group_shared` memory separate;
- show why a memory exists;
- accept human correction;
- suppress superseded memory;
- keep working in degraded mode when Codex reasoning is unavailable;
- prove behavior through replay and golden scenarios.

## What VibeGravity Is Not

VibeGravity is not:

- a chat application;
- a standalone AI assistant;
- a general workflow automation system;
- a raw transcript archive;
- a generic vector database;
- a heavyweight knowledge graph platform.

## Product Tagline Candidates

- Memory engine for AI agents.
- Shared memory kernel for Hermes and agent teams.
- Scoped, correctable memory for long-running agents.
- The memory layer agents can trust.


```



<!-- Source: consulting/03_engine_positioning_and_narrative.md | bytes=3185 | lines=97 | sha16=f6465ceab4e4f2ec -->

```md
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


```



<!-- Source: consulting/04_customer_and_use_cases.md | bytes=3186 | lines=98 | sha16=c996982cd104f5b4 -->

```md
# Customer and Use Cases

## First Customer

The first customer is Hermes Agent.

That means V1 should optimize for Hermes lifecycle integration:

- before turn: call `prefetch()`;
- after turn: call `sync_turn()`;
- during operation: expose search, note, plan, correction, timeline, and explain tools;
- after session: trigger dreaming or consolidation hints.

## First Human User

The first human user is the operator or builder who uses Hermes for ongoing work.

This person cares less about the internal memory graph and more about:

- not repeating durable context;
- keeping current plans alive;
- correcting wrong memory;
- seeing why a memory exists;
- avoiding private/shared leakage;
- trusting that memory will not silently rot.

## Primary Jobs To Be Done

### 1. Continue Work Without Repeating Context

When I start a new Hermes session, I want Hermes to remember the relevant project rules, current plan, and recent decisions so I can continue work without re-explaining everything.

### 2. Keep Shared and Private Memory Separate

When multiple agents work in one workspace, I want workspace memory to be shared while private agent memory stays private, so collaboration does not become leakage.

### 3. Correct Memory When It Is Wrong

When the system remembers something outdated or wrong, I want to correct it once and have future recall reflect the correction.

### 4. Inspect Why Memory Exists

When memory affects agent behavior, I want to see where it came from and what reasoning or event produced it.

### 5. Keep Useful Context Short

When Hermes asks for recall, I want the engine to return compact, useful blocks instead of dumping a long transcript.

## Initial Use Cases

### Hermes Continuity

Hermes uses VibeGravity to remember project preferences, current task status, user constraints, and corrections across sessions.

Success looks like:

- Hermes resumes the correct work;
- active plan and pinned notes appear in recall;
- stale superseded memory is suppressed;
- user correction changes later behavior.

### Agent Collaboration in One Workspace

Multiple agents work on the same project while VibeGravity separates private and workspace memory.

Success looks like:

- workspace decisions are shared;
- `agent_private` memory only returns to the owning agent;
- group memory only appears for members;
- provenance remains inspectable.

### Operator Memory Control

The user uses tools to search, add notes, create/update plans, correct memory, and inspect timeline.

Success looks like:

- user can repair the memory system;
- operator actions are visible;
- correction is not a hidden prompt trick;
- the memory engine becomes more trustworthy over time.

## Later Use Cases

These should not drive V1 unless the Product Owner strongly disagrees:

- generic memory backend for many unrelated agent runtimes;
- standalone web UI;
- marketplace of memory integrations;
- advanced organization-wide knowledge graph;
- fully autonomous forgetting without operator visibility.

## Customer Question For Consulting

Is Hermes-first enough as the first customer, or should V1 explicitly define a human operator persona with a visible control surface?


```



<!-- Source: consulting/05_mvp_scope_and_non_goals.md | bytes=2711 | lines=96 | sha16=b88bcf6759ccfc8b -->

```md
# MVP Scope and Non-Goals

## V1 Product Goal

V1 should prove that VibeGravity makes Hermes more continuous, safer, and more correctable across sessions.

It does not need to prove that VibeGravity is a universal memory platform.

## Must Have

### Hot Path

- `sync_turn()` records raw events quickly.
- The API stays available even if background reasoning is delayed.
- Ingest is idempotent.

### Recall

- `prefetch()` returns typed recall blocks.
- Recall is scope-aware.
- Recall respects a token budget.
- Active plans and pinned notes can outrank noisy memory.
- Superseded memory is suppressed.
- Degraded mode still returns useful existing context when possible.

### Memory Engine

- Raw events and derived memories are separate.
- Every memory has explicit scope.
- Every memory has provenance.
- `create_memory`, `extend_memory`, and safe `update_memory` have guarded write semantics.
- Human correction can create a replacement memory and supersede the target.

### Operator Control

- Search memory.
- Add note.
- Create/update plan.
- Correct memory.
- View timeline.
- Explain memory provenance.

### Hermes Integration

- Hermes can call pre-turn recall and post-turn sync.
- A practical MCP or provider path exists.
- Failure in VibeGravity does not kill Hermes.

### Quality Gate

- Golden scenarios cover recall usefulness, scope separation, correction, supersession, degraded mode, and replay idempotency.

## Should Have

- CLI doctor for local setup.
- Read-only backlog metrics.
- Basic worker recovery visibility.
- Session summary recall.
- Dreaming jobs that promote memory by metadata without changing scope.

## Could Have

- Small operator UI.
- Rich timeline visualization.
- Real Hermes provider registry packaging if Hermes exposes the needed hook.
- More advanced document recall.
- Profile coherence scoring.

## Not V1

- Generic chat UI.
- Large web app.
- Fully autonomous forgetting without operator visibility.
- Every agent runtime integration.
- Heavy ontology platform.
- Organization-wide admin console.
- Multi-node distributed queue redesign.
- Real Codex enabled by default before failure/degraded behavior is explicit.

## MVP Demo Candidate

A strong V1 demo should show:

1. User gives Hermes a project rule and plan.
2. Hermes completes a turn and calls `sync_turn()`.
3. Worker derives memory.
4. Next session calls `prefetch()` and gets a compact recall pack.
5. User corrects a wrong memory.
6. Later recall suppresses the old memory and includes the correction.
7. Timeline/explain shows where the memory and correction came from.

## Product Owner Question

Is this V1 still too broad? If yes, what is the single narrow V1 that best proves the product?


```



<!-- Source: consulting/06_runtime_and_product_contract.md | bytes=2768 | lines=104 | sha16=b31eec8c1a01835e -->

```md
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


```



<!-- Source: consulting/07_current_state_and_roadmap.md | bytes=3386 | lines=110 | sha16=a5f8707aaa8d7a0b -->

```md
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


```



<!-- Source: consulting/08_risks_and_open_decisions.md | bytes=3216 | lines=88 | sha16=7676dcf460ec8fe2 -->

```md
# Risks and Open Decisions

## Product Risks

### 1. The Product May Sound Like Infrastructure, Not Value

"Shared memory kernel" is accurate but abstract.

Risk: users understand the architecture but not why they should care.

Decision needed: choose a product narrative that starts from pain, not architecture.

### 2. Hermes-First May Be Too Narrow or Too Hidden

Hermes is a strong first host, but if the memory engine is invisible, the user may not notice the value.

Decision needed: define the first visible proof moment.

### 3. Trust UX May Be Underbuilt

Memory is sensitive. Users need ways to inspect and correct it.

Risk: a powerful engine without timeline/explain/correction UX feels unsafe.

Decision needed: what minimum operator surface must ship with V1?

### 4. Scope Separation Is a Product Promise

Private/shared/group memory separation is not optional.

Risk: if this is hard to explain or inspect, users will not trust multi-agent memory.

Decision needed: how should scope be displayed and controlled?

### 5. Codex Dependency Needs a Clear Story

Codex-first reasoning is deliberate, but outages and auth state can affect memory freshness.

Risk: users misread stale recall as fresh intelligence.

Decision needed: how visible should degraded mode be?

### 6. The MVP May Be Too Broad

V1 currently includes ingest, recall, notes, plans, documents, correction, timeline, explain, graph updates, dreaming, Hermes, MCP, evals, and ops.

Risk: too many surfaces delay the first compelling demo.

Decision needed: what can be postponed without weakening the product promise?

## Open Product Decisions

| Decision | Current Lean | Needs PO Judgment |
|---|---|---|
| Category | Agent memory engine | Is this understandable enough? |
| First customer | Hermes Agent | Is the first human persona clear? |
| First proof | Recall + correction + explain | Is this the right demo? |
| Operator UI | CLI/MCP/timeline first | Is a minimal UI required? |
| V1 scope | Memory continuity and trust | What should be cut? |
| Dreaming | Maintenance layer, metadata-first | How much is needed for V1? |
| Documents | Supporting context | Keep in V1 or defer? |
| Codex | Reasoning backend, disabled by default until gates | How to message dependency and outages? |
| Packaging | MCP bootstrap first | Is that enough for Hermes-first adoption? |

## Critical Product Questions

1. What would make a user say "I need this" after a demo?
2. What is the one workflow V1 must make dramatically better?
3. What should the product refuse to do?
4. What is the simplest visible surface for trust?
5. Which memory mistakes are unacceptable?
6. What must be measurable in the first beta?
7. Is VibeGravity a product brand, a component inside Hermes, or both?

## Suggested Success Metrics

- Recall usefulness score from golden scenarios.
- Correction-to-recall propagation success.
- Scope leakage regression count.
- Duplicate memory rate.
- Superseded memory suppression rate.
- `sync_turn()` latency.
- `prefetch()` latency.
- Degraded recall non-empty rate when stored context exists.
- Backlog recovery time after mocked Codex outage.
- Number of sessions where the user did not need to restate known context.


```



<!-- Source: consulting/09_consulting_questionnaire.md | bytes=2404 | lines=81 | sha16=eff3d084284b5cac -->

```md
# Product Owner Consulting Questionnaire

## Main Question

Is VibeGravity currently shaped like a product people will understand and trust, or is it still mostly an excellent technical substrate?

## Questions For The Consultant

### Product Category

1. What category should VibeGravity claim?
2. Is "agent memory engine" clear enough?
3. Should the public story lead with Hermes, agents, memory continuity, or trust?

### First Customer

4. Who is the true first user: Hermes, the Hermes operator, or developers building agent systems?
5. What is the first use case that should define all V1 tradeoffs?
6. Does Hermes-first create a strong enough wedge?

### V1 Scope

7. What is the minimum V1 that proves value?
8. Which current capabilities should be removed from V1?
9. Which capabilities are non-negotiable for trust?
10. Is document memory necessary for V1?
11. Is dreaming necessary for V1, or can it be a post-V1 quality layer?

### User Experience

12. Does V1 need a UI, or are CLI/MCP/HTTP surfaces enough?
13. What should the user see when memory is wrong?
14. What should the user see when Codex reasoning is unavailable?
15. How should private vs shared vs group memory be explained?

### Trust and Safety

16. What are the unacceptable failure modes?
17. What should be auditable before beta?
18. How much provenance must be shown to a normal user?
19. What should the correction workflow feel like?

### Demo and Go-To-Market

20. What is the best 5-minute demo?
21. What is the best first landing-page promise?
22. What would make early adopters switch this on in Hermes?
23. What should the first beta cohort look like?

## Requested Scorecard

Please score each area from 1 to 5 and explain the reason.

| Area | Score | Notes |
|---|---:|---|
| Product clarity |  |  |
| First customer clarity |  |  |
| V1 scope sharpness |  |  |
| Trust and correction UX |  |  |
| Differentiation |  |  |
| Hermes-first wedge |  |  |
| Demo strength |  |  |
| Adoption readiness |  |  |
| Risk control |  |  |

## Requested Final Deliverable Format

Please return:

1. Executive verdict.
2. Product category recommendation.
3. First customer and first use case.
4. V1 scope: keep, cut, defer.
5. Required user/operator surfaces.
6. Top 5 product risks.
7. Recommended 5-minute demo.
8. 30-day execution plan.
9. Suggested product language.
10. Any hard disagreement with the current direction.


```



<!-- Source: internal/hermes/doc.go | bytes=769 | lines=16 | sha16=d5fd2d580b30889f -->

```go
// ============================================================
// FILE     : internal/hermes/doc.go
// PURPOSE  : Provides package documentation for Hermes provider adapter semantics.
// LAYER    : interface
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : package hermes
// DEPENDS  : plans/10_workpack_hermes-provider-and-external-surfaces.md
// USED_BY  : Hermes provider integration, tests
// ------------------------------------------------------------
// AGENT_NOTE: Hermes is the first customer; adapter behavior must not fork core semantics.
// ============================================================

// Package hermes maps Hermes memory-provider lifecycle hooks to VibeGravity core calls.
package hermes

```



<!-- Source: internal/hermes/provider.go | bytes=6682 | lines=181 | sha16=57d05fe04a287d1e -->

```go
// ============================================================
// FILE     : internal/hermes/provider.go
// PURPOSE  : Maps Hermes memory-provider lifecycle hooks to VibeGravity core service calls.
// LAYER    : interface
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : Provider, ProviderTool, NewProvider
// DEPENDS  : context, fmt, strings, internal/core
// USED_BY  : Hermes integration tests, future plugin runtime
// ------------------------------------------------------------
// AGENT_NOTE: Keep this adapter thin so Hermes, HTTP, and MCP share the same core semantics.
// ============================================================

package hermes

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

// Provider maps Hermes memory-provider lifecycle hooks to the core service.
type Provider struct {
	service core.VibeGravityService
}

// ProviderTool describes one operator tool exposed through the Hermes provider.
type ProviderTool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// NewProvider creates a Hermes-facing adapter over the shared VibeGravity service.
func NewProvider(service core.VibeGravityService) (*Provider, error) {
	if service == nil {
		return nil, fmt.Errorf("%w: hermes provider service is required", core.ErrInvalidArgument)
	}
	return &Provider{service: service}, nil
}

// IsAvailable checks whether the backing VibeGravity service can answer a minimal prefetch.
func (p *Provider) IsAvailable(ctx context.Context, req *core.PrefetchRequest) bool {
	if p == nil || p.service == nil || req == nil {
		return false
	}
	_, err := p.service.Prefetch(ctx, req)
	return err == nil
}

// Prefetch calls the core recall assembler before a Hermes turn.
func (p *Provider) Prefetch(ctx context.Context, req *core.PrefetchRequest) (*core.PrefetchResponse, error) {
	if p == nil || p.service == nil {
		return nil, fmt.Errorf("%w: hermes provider is not initialized", core.ErrInvalidArgument)
	}
	return p.service.Prefetch(ctx, req)
}

// SyncTurn records a completed Hermes turn through the hot ingest path.
func (p *Provider) SyncTurn(ctx context.Context, req *core.SyncTurnRequest) (*core.SyncTurnResponse, error) {
	if p == nil || p.service == nil {
		return nil, fmt.Errorf("%w: hermes provider is not initialized", core.ErrInvalidArgument)
	}
	return p.service.SyncTurn(ctx, req)
}

// RenderContext turns typed recall blocks into compact Hermes text context.
func (p *Provider) RenderContext(resp *core.PrefetchResponse) string {
	if resp == nil || len(resp.Blocks) == 0 {
		return ""
	}
	lines := make([]string, 0, len(resp.Blocks))
	for _, block := range resp.Blocks {
		text := strings.TrimSpace(block.Text)
		if text == "" {
			continue
		}
		kind := strings.TrimSpace(block.Kind)
		if kind == "" {
			kind = "memory"
		}
		labels := []string{kind, fmt.Sprintf("%d", block.Priority)}
		if block.Scope != "" {
			labels = append(labels, string(block.Scope))
		}
		if source := strings.TrimSpace(block.Source); source != "" {
			labels = append(labels, source)
		}
		if freshness := strings.TrimSpace(block.Freshness); freshness != "" {
			labels = append(labels, freshness)
		}
		lines = append(lines, fmt.Sprintf("[%s] %s", strings.Join(labels, ":"), text))
	}
	return strings.Join(lines, "\n")
}

// GetTools lists the minimum Hermes provider tools backed by the v1 core API.
func (p *Provider) GetTools() []ProviderTool {
	return []ProviderTool{
		{Name: "recall_preview", Description: "Preview the scoped recall Hermes will receive."},
		{Name: "search_memory", Description: "Search visible VibeGravity memories."},
		{Name: "add_note", Description: "Create a human-authored recall control note."},
		{Name: "show_plan", Description: "Show active structured plans."},
		{Name: "explain_memory", Description: "Show provenance for a remembered item."},
		{Name: "correct_memory", Description: "Record a human correction for a memory."},
		{Name: "view_timeline", Description: "View memory and correction activity."},
		{Name: "degraded_status", Description: "Show whether recall is fresh or degraded."},
	}
}

// CallTool dispatches a Hermes provider tool to the shared core service.
func (p *Provider) CallTool(ctx context.Context, name string, input json.RawMessage) (json.RawMessage, error) {
	if p == nil || p.service == nil {
		return nil, fmt.Errorf("%w: hermes provider is not initialized", core.ErrInvalidArgument)
	}
	switch name {
	case "recall_preview":
		var req core.PrefetchRequest
		return callJSON(ctx, input, &req, p.service.Prefetch)
	case "search_memory":
		var req core.SearchMemoriesRequest
		return callJSON(ctx, input, &req, p.service.SearchMemories)
	case "add_note":
		var req core.AddNoteRequest
		return callJSON(ctx, input, &req, p.service.AddNote)
	case "explain_memory":
		var req core.ExplainMemoryRequest
		return callJSON(ctx, input, &req, p.service.ExplainMemory)
	case "correct_memory":
		var req core.CorrectMemoryRequest
		return callJSON(ctx, input, &req, p.service.CorrectMemory)
	case "view_timeline":
		var req core.GetTimelineRequest
		return callJSON(ctx, input, &req, p.service.GetTimeline)
	case "degraded_status":
		var req core.PrefetchRequest
		return callJSON(ctx, input, &req, func(ctx context.Context, req *core.PrefetchRequest) (*core.RecallMeta, error) {
			resp, err := p.service.Prefetch(ctx, req)
			if err != nil {
				return nil, err
			}
			if resp == nil {
				return &core.RecallMeta{}, nil
			}
			return &resp.Meta, nil
		})
	case "show_plan":
		return nil, fmt.Errorf("%w: hermes provider tool %q needs a read-only plan API", core.ErrNotImplemented, name)
	default:
		return nil, fmt.Errorf("%w: unknown hermes provider tool %q", core.ErrInvalidArgument, name)
	}
}

// OnSessionEnd records a session-end hint for future dreaming integration.
func (p *Provider) OnSessionEnd(context.Context, string) error {
	if p == nil || p.service == nil {
		return fmt.Errorf("%w: hermes provider is not initialized", core.ErrInvalidArgument)
	}
	return nil
}

func callJSON[Req any, Resp any](ctx context.Context, input json.RawMessage, req *Req, call func(context.Context, *Req) (*Resp, error)) (json.RawMessage, error) {
	if len(input) == 0 {
		input = json.RawMessage(`{}`)
	}
	if err := json.Unmarshal(input, req); err != nil {
		return nil, fmt.Errorf("%w: decode hermes tool input: %v", core.ErrInvalidArgument, err)
	}
	resp, err := call(ctx, req)
	if err != nil {
		return nil, err
	}
	output, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("encode hermes tool output: %w", err)
	}
	return output, nil
}

```



<!-- Source: internal/hermes/provider_test.go | bytes=10507 | lines=274 | sha16=315a530430c500d8 -->

```go
// ============================================================
// FILE     : internal/hermes/provider_test.go
// PURPOSE  : Verifies Hermes provider lifecycle hooks delegate to VibeGravity core semantics.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : Hermes provider adapter tests
// DEPENDS  : context, errors, strings, testing, internal/core
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: These tests use a fake core service; they do not call a real Hermes runtime.
// ============================================================

package hermes

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

func TestProviderDelegatesPrefetchAndRendersContext(t *testing.T) {
	t.Parallel()

	service := &fakeService{
		prefetchResp: &core.PrefetchResponse{
			Blocks: []core.RecallBlock{
				{Kind: "pinned_note", Priority: 100, Text: "Keep Hermes first.", Scope: core.MemoryScopeWorkspaceShared, Source: "notes", Freshness: "stored"},
				{Kind: "active_plan", Priority: 95, Text: "Finish V1 core semantics."},
			},
			Meta: core.RecallMeta{EstimatedTokens: 12, Sources: []string{"notes", "plans"}},
		},
	}
	provider := newTestProvider(t, service)

	resp, err := provider.Prefetch(context.Background(), &core.PrefetchRequest{
		TenantID: "tenant_1", WorkspaceID: "workspace_1", SessionID: "session_1", ActorID: "agent:hermes-main",
	})
	if err != nil {
		t.Fatalf("Prefetch returned error: %v", err)
	}
	if service.prefetchCalls != 1 {
		t.Fatalf("expected one prefetch call, got %d", service.prefetchCalls)
	}
	rendered := provider.RenderContext(resp)
	if !strings.Contains(rendered, "[pinned_note:100:workspace_shared:notes:stored] Keep Hermes first.") {
		t.Fatalf("rendered context lost pinned note: %q", rendered)
	}
	if !strings.Contains(rendered, "[active_plan:95] Finish V1 core semantics.") {
		t.Fatalf("rendered context lost active plan: %q", rendered)
	}
}

func TestProviderDelegatesSyncTurn(t *testing.T) {
	t.Parallel()

	service := &fakeService{syncResp: &core.SyncTurnResponse{Status: "accepted", EventIDs: []string{"evt_1"}, JobIDs: []string{"job_1"}}}
	provider := newTestProvider(t, service)

	resp, err := provider.SyncTurn(context.Background(), &core.SyncTurnRequest{
		TenantID: "tenant_1", WorkspaceID: "workspace_1", SessionID: "session_1", ActorID: "agent:hermes-main",
	})
	if err != nil {
		t.Fatalf("SyncTurn returned error: %v", err)
	}
	if service.syncCalls != 1 {
		t.Fatalf("expected one sync call, got %d", service.syncCalls)
	}
	if resp.Status != "accepted" || len(resp.JobIDs) != 1 {
		t.Fatalf("unexpected sync response: %#v", resp)
	}
}

func TestProviderAvailabilityReflectsPrefetchHealth(t *testing.T) {
	t.Parallel()

	okProvider := newTestProvider(t, &fakeService{prefetchResp: &core.PrefetchResponse{}})
	if !okProvider.IsAvailable(context.Background(), &core.PrefetchRequest{TenantID: "tenant_1"}) {
		t.Fatalf("expected provider to be available when prefetch succeeds")
	}

	failingProvider := newTestProvider(t, &fakeService{prefetchErr: errors.New("database unavailable")})
	if failingProvider.IsAvailable(context.Background(), &core.PrefetchRequest{TenantID: "tenant_1"}) {
		t.Fatalf("expected provider to be unavailable when prefetch fails")
	}
}

func TestProviderToolsExposeV1HermesSurface(t *testing.T) {
	t.Parallel()

	provider := newTestProvider(t, &fakeService{})
	tools := provider.GetTools()
	names := make(map[string]bool, len(tools))
	for _, tool := range tools {
		names[tool.Name] = true
	}
	for _, want := range []string{"recall_preview", "search_memory", "add_note", "show_plan", "explain_memory", "correct_memory", "view_timeline", "degraded_status"} {
		if !names[want] {
			t.Fatalf("expected provider tool %q in %#v", want, tools)
		}
	}
}

func TestProviderCallToolDelegatesRecallPreview(t *testing.T) {
	t.Parallel()

	service := &fakeService{prefetchResp: &core.PrefetchResponse{Blocks: []core.RecallBlock{{Text: "Use scoped recall."}}}}
	provider := newTestProvider(t, service)

	raw, err := provider.CallTool(context.Background(), "recall_preview", json.RawMessage(`{"tenant_id":"tenant_1","workspace_id":"workspace_1"}`))
	if err != nil {
		t.Fatalf("CallTool returned error: %v", err)
	}
	if service.prefetchCalls != 1 {
		t.Fatalf("expected one prefetch call, got %d", service.prefetchCalls)
	}
	if !strings.Contains(string(raw), "Use scoped recall.") {
		t.Fatalf("expected encoded recall preview output, got %s", string(raw))
	}
}

func TestProviderCallToolDelegatesTrustLoopTools(t *testing.T) {
	t.Parallel()

	service := &fakeService{
		searchResp:   &core.SearchMemoriesResponse{Memories: []core.MemoryResult{{MemoryID: "mem_1", Text: "Project rule"}}},
		addNoteResp:  &core.AddNoteResponse{NoteID: "note_1", Status: "created"},
		explainResp:  &core.ExplainMemoryResponse{MemoryID: "mem_1", Trace: core.MemoryTraceResult{ReasoningJobID: "job_1"}},
		correctResp:  &core.CorrectMemoryResponse{MemoryID: "mem_1", CorrectionRecorded: true},
		timelineResp: &core.GetTimelineResponse{Items: []core.TimelineItem{{ID: "item_1", Text: "Correction recorded"}}},
	}
	provider := newTestProvider(t, service)

	cases := []struct {
		name string
		raw  json.RawMessage
		want string
	}{
		{name: "search_memory", raw: json.RawMessage(`{"tenant_id":"tenant_1","workspace_id":"workspace_1","query":"rule"}`), want: "Project rule"},
		{name: "add_note", raw: json.RawMessage(`{"tenant_id":"tenant_1","workspace_id":"workspace_1","text":"Pin this"}`), want: "note_1"},
		{name: "explain_memory", raw: json.RawMessage(`{"tenant_id":"tenant_1","workspace_id":"workspace_1","memory_id":"mem_1"}`), want: "job_1"},
		{name: "correct_memory", raw: json.RawMessage(`{"tenant_id":"tenant_1","workspace_id":"workspace_1","memory_id":"mem_1"}`), want: "correction_recorded"},
		{name: "view_timeline", raw: json.RawMessage(`{"tenant_id":"tenant_1","workspace_id":"workspace_1","entity_id":"agent:hermes-main"}`), want: "Correction recorded"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := provider.CallTool(context.Background(), tc.name, tc.raw)
			if err != nil {
				t.Fatalf("CallTool returned error: %v", err)
			}
			if !strings.Contains(string(raw), tc.want) {
				t.Fatalf("expected output to contain %q, got %s", tc.want, string(raw))
			}
		})
	}
}

func TestProviderCallToolReturnsDegradedStatusFromPrefetchMeta(t *testing.T) {
	t.Parallel()

	service := &fakeService{prefetchResp: &core.PrefetchResponse{Meta: core.RecallMeta{Freshness: "stale", Degraded: true, DegradedReasons: []string{"worker backlog"}}}}
	provider := newTestProvider(t, service)

	raw, err := provider.CallTool(context.Background(), "degraded_status", json.RawMessage(`{"tenant_id":"tenant_1","workspace_id":"workspace_1"}`))
	if err != nil {
		t.Fatalf("CallTool returned error: %v", err)
	}
	if service.prefetchCalls != 1 {
		t.Fatalf("expected one prefetch call, got %d", service.prefetchCalls)
	}
	if !strings.Contains(string(raw), `"freshness":"stale"`) || !strings.Contains(string(raw), "worker backlog") {
		t.Fatalf("expected encoded recall meta, got %s", string(raw))
	}
}

func TestProviderCallToolRejectsUnknownInvalidAndUnbackedTools(t *testing.T) {
	t.Parallel()

	provider := newTestProvider(t, &fakeService{})
	if _, err := provider.CallTool(context.Background(), "unknown", json.RawMessage(`{}`)); !errors.Is(err, core.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument for unknown tool, got %v", err)
	}
	if _, err := provider.CallTool(context.Background(), "recall_preview", json.RawMessage(`{`)); !errors.Is(err, core.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument for invalid JSON, got %v", err)
	}
	if _, err := provider.CallTool(context.Background(), "show_plan", json.RawMessage(`{}`)); !errors.Is(err, core.ErrNotImplemented) {
		t.Fatalf("expected ErrNotImplemented for show_plan, got %v", err)
	}
}

func newTestProvider(t *testing.T, service core.VibeGravityService) *Provider {
	t.Helper()

	provider, err := NewProvider(service)
	if err != nil {
		t.Fatalf("NewProvider returned error: %v", err)
	}
	return provider
}

type fakeService struct {
	prefetchCalls int
	syncCalls     int
	searchCalls   int
	addNoteCalls  int
	explainCalls  int
	correctCalls  int
	timelineCalls int
	prefetchResp  *core.PrefetchResponse
	prefetchErr   error
	syncResp      *core.SyncTurnResponse
	syncErr       error
	searchResp    *core.SearchMemoriesResponse
	addNoteResp   *core.AddNoteResponse
	explainResp   *core.ExplainMemoryResponse
	correctResp   *core.CorrectMemoryResponse
	timelineResp  *core.GetTimelineResponse
}

func (s *fakeService) Prefetch(context.Context, *core.PrefetchRequest) (*core.PrefetchResponse, error) {
	s.prefetchCalls++
	return s.prefetchResp, s.prefetchErr
}

func (s *fakeService) SyncTurn(context.Context, *core.SyncTurnRequest) (*core.SyncTurnResponse, error) {
	s.syncCalls++
	return s.syncResp, s.syncErr
}

func (s *fakeService) AddDocument(context.Context, *core.AddDocumentRequest) (*core.AddDocumentResponse, error) {
	return nil, core.ErrNotImplemented
}

func (s *fakeService) SearchMemories(context.Context, *core.SearchMemoriesRequest) (*core.SearchMemoriesResponse, error) {
	s.searchCalls++
	return s.searchResp, nil
}

func (s *fakeService) SearchDocuments(context.Context, *core.SearchDocumentsRequest) (*core.SearchDocumentsResponse, error) {
	return nil, core.ErrNotImplemented
}

func (s *fakeService) AddNote(context.Context, *core.AddNoteRequest) (*core.AddNoteResponse, error) {
	s.addNoteCalls++
	return s.addNoteResp, nil
}

func (s *fakeService) CreatePlan(context.Context, *core.CreatePlanRequest) (*core.CreatePlanResponse, error) {
	return nil, core.ErrNotImplemented
}

func (s *fakeService) UpdatePlan(context.Context, *core.UpdatePlanRequest) (*core.UpdatePlanResponse, error) {
	return nil, core.ErrNotImplemented
}

func (s *fakeService) CorrectMemory(context.Context, *core.CorrectMemoryRequest) (*core.CorrectMemoryResponse, error) {
	s.correctCalls++
	return s.correctResp, nil
}

func (s *fakeService) GetTimeline(context.Context, *core.GetTimelineRequest) (*core.GetTimelineResponse, error) {
	s.timelineCalls++
	return s.timelineResp, nil
}

func (s *fakeService) ExplainMemory(context.Context, *core.ExplainMemoryRequest) (*core.ExplainMemoryResponse, error) {
	s.explainCalls++
	return s.explainResp, nil
}

```



<!-- Source: internal/mcp/doc.go | bytes=756 | lines=16 | sha16=999ee2bb757e8a9d -->

```go
// ============================================================
// FILE     : internal/mcp/doc.go
// PURPOSE  : Provides package documentation for the MCP operator and agent tool surface.
// LAYER    : interface
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : package mcp
// DEPENDS  : plans/10_workpack_hermes-provider-and-external-surfaces.md
// USED_BY  : MCP clients, operator tools, coding agent integrations
// ------------------------------------------------------------
// AGENT_NOTE: MCP tools must call the same core semantics used by Hermes and HTTP.
// ============================================================

// Package mcp exposes VibeGravity memory operations as MCP tools.
package mcp

```



<!-- Source: internal/mcp/protocol.go | bytes=15442 | lines=426 | sha16=921203ceb6bd6f45 -->

```go
// ============================================================
// FILE     : internal/mcp/protocol.go
// PURPOSE  : Serves the VibeGravity MCP tool surface over JSON-RPC transports.
// LAYER    : interface
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : ProtocolVersion, Server, NewServer
// DEPENDS  : internal/mcp/surface.go, encoding/json, bufio, io
// USED_BY  : cmd/cli, MCP protocol roundtrip tests
// ------------------------------------------------------------
// AGENT_NOTE: Stdout must contain only newline-delimited JSON-RPC messages.
// ============================================================

package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// ProtocolVersion is the MCP version this server advertises.
const ProtocolVersion = "2025-11-25"

const (
	jsonRPCVersion = "2.0"

	errParseError     = -32700
	errInvalidRequest = -32600
	errMethodNotFound = -32601
	errInvalidParams  = -32602
	errInternalError  = -32603
)

// Server handles MCP JSON-RPC messages for a Surface.
type Server struct {
	surface *Surface
}

// NewServer creates an MCP protocol server over the shared tool surface.
func NewServer(surface *Surface) (*Server, error) {
	if surface == nil {
		return nil, fmt.Errorf("mcp surface is required")
	}
	return &Server{surface: surface}, nil
}

// ServeStdio serves newline-delimited MCP JSON-RPC over stdin/stdout style streams.
func (s *Server) ServeStdio(ctx context.Context, in io.Reader, out io.Writer) error {
	if s == nil || s.surface == nil {
		return fmt.Errorf("mcp server is not initialized")
	}
	scanner := bufio.NewScanner(in)
	const maxMessageBytes = 1024 * 1024
	scanner.Buffer(make([]byte, 0, 64*1024), maxMessageBytes)
	encoder := json.NewEncoder(out)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		resp, respond := s.HandleMessage(ctx, line)
		if !respond {
			continue
		}
		if err := encoder.Encode(resp); err != nil {
			return fmt.Errorf("write mcp response: %w", err)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read mcp input: %w", err)
	}
	return nil
}

// HandleMessage handles one MCP JSON-RPC message. Notifications return respond=false.
func (s *Server) HandleMessage(ctx context.Context, raw json.RawMessage) (json.RawMessage, bool) {
	var req rpcRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		return marshalRPCError(nil, errParseError, "Parse error", nil), true
	}
	if req.JSONRPC != jsonRPCVersion || req.Method == "" {
		return marshalRPCError(req.ID, errInvalidRequest, "Invalid Request", nil), true
	}
	if len(req.ID) == 0 {
		if req.Method == "notifications/initialized" {
			return nil, false
		}
		return nil, false
	}

	switch req.Method {
	case "initialize":
		return marshalRPCResult(req.ID, initializeResult{
			ProtocolVersion: ProtocolVersion,
			Capabilities: serverCapabilities{
				Tools: toolsCapability{ListChanged: false},
			},
			ServerInfo: implementationInfo{
				Name:        "vibegravity",
				Title:       "VibeGravity MCP Server",
				Version:     "0.1.0",
				Description: "Hermes-first shared memory kernel tools.",
			},
			Instructions: "Use VibeGravity tools for memory recall, sync, notes, plans, corrections, and timeline visibility.",
		}), true
	case "ping":
		return marshalRPCResult(req.ID, map[string]any{}), true
	case "tools/list":
		return marshalRPCResult(req.ID, listToolsResult{Tools: s.protocolTools()}), true
	case "tools/call":
		result, err := s.callTool(ctx, req.Params)
		if err != nil {
			return marshalRPCError(req.ID, errInvalidParams, err.Error(), nil), true
		}
		return marshalRPCResult(req.ID, result), true
	default:
		return marshalRPCError(req.ID, errMethodNotFound, "Method not found", map[string]string{"method": req.Method}), true
	}
}

func (s *Server) protocolTools() []protocolTool {
	tools := s.surface.Tools()
	out := make([]protocolTool, 0, len(tools))
	for _, tool := range tools {
		out = append(out, protocolTool{
			Name:        tool.Name,
			Title:       tool.Name,
			Description: tool.Description,
			InputSchema: toolInputSchema(tool.Name),
		})
	}
	return out
}

func toolInputSchema(name string) map[string]any {
	base := func(required []string, properties map[string]any) map[string]any {
		schema := map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties":           properties,
		}
		if len(required) > 0 {
			schema["required"] = required
		}
		return schema
	}
	stringProp := func(description string) map[string]any {
		return map[string]any{"type": "string", "description": description}
	}
	intProp := func(description string) map[string]any {
		return map[string]any{"type": "integer", "description": description}
	}
	boolProp := func(description string) map[string]any {
		return map[string]any{"type": "boolean", "description": description}
	}
	jsonProp := func(description string) map[string]any {
		return map[string]any{"description": description}
	}
	stringArrayProp := func(description string) map[string]any {
		return map[string]any{
			"type":        "array",
			"description": description,
			"items":       map[string]any{"type": "string"},
		}
	}

	scopeArray := stringArrayProp("Visible memory scopes such as agent_private, workspace_shared, group_shared, or session_scratch.")
	switch name {
	case "prefetch", "recall_preview":
		return base([]string{"tenant_id", "workspace_id"}, map[string]any{
			"tenant_id":     stringProp("Tenant identifier."),
			"workspace_id":  stringProp("Workspace identifier."),
			"session_id":    stringProp("Session identifier."),
			"actor_id":      stringProp("Actor requesting recall."),
			"query":         stringProp("Question or task used to assemble recall."),
			"budget_tokens": intProp("Maximum approximate recall token budget."),
			"mode":          stringProp("Recall mode such as default, small, or rich."),
		})
	case "sync_turn":
		return base([]string{"tenant_id", "workspace_id", "session_id", "actor_id", "turn_events"}, map[string]any{
			"tenant_id":       stringProp("Tenant identifier."),
			"workspace_id":    stringProp("Workspace identifier."),
			"session_id":      stringProp("Session identifier."),
			"actor_id":        stringProp("Actor that produced the turn."),
			"idempotency_key": stringProp("Optional request-level idempotency key."),
			"turn_events": map[string]any{
				"type":        "array",
				"description": "Raw turn events to record.",
				"items": map[string]any{
					"type":                 "object",
					"additionalProperties": false,
					"required":             []string{"event_kind", "source", "payload_json"},
					"properties": map[string]any{
						"event_kind":   stringProp("Raw event kind."),
						"source":       stringProp("Producer source."),
						"fingerprint":  stringProp("Optional event fingerprint."),
						"occurred_at":  stringProp("RFC3339 event timestamp."),
						"payload_json": jsonProp("Raw event payload object."),
					},
				},
			},
		})
	case "search_memory":
		return base([]string{"tenant_id", "workspace_id", "query"}, map[string]any{
			"tenant_id":         stringProp("Tenant identifier."),
			"workspace_id":      stringProp("Workspace identifier."),
			"owner_entity_id":   stringProp("Actor used for agent_private filtering."),
			"visible_group_ids": stringArrayProp("Group identifiers visible to the actor."),
			"query":             stringProp("Memory search query."),
			"scopes":            scopeArray,
			"artifact_classes":  stringArrayProp("Memory artifact classes to include."),
		})
	case "search_documents":
		return base([]string{"tenant_id", "workspace_id", "query"}, map[string]any{
			"tenant_id":    stringProp("Tenant identifier."),
			"workspace_id": stringProp("Workspace identifier."),
			"query":        stringProp("Document search query."),
		})
	case "add_note":
		return base([]string{"tenant_id", "workspace_id", "scope", "text"}, map[string]any{
			"tenant_id":       stringProp("Tenant identifier."),
			"workspace_id":    stringProp("Workspace identifier."),
			"note_kind":       stringProp("Note kind."),
			"scope":           stringProp("Memory scope for the note."),
			"owner_entity_id": stringProp("Owner when scope is agent_private."),
			"text":            stringProp("Note text."),
			"pinned":          boolProp("Whether the note should receive recall priority."),
			"expires_at":      stringProp("Optional RFC3339 expiration timestamp."),
		})
	case "create_plan":
		return base([]string{"tenant_id", "workspace_id", "title", "status", "scope"}, map[string]any{
			"tenant_id":       stringProp("Tenant identifier."),
			"workspace_id":    stringProp("Workspace identifier."),
			"title":           stringProp("Plan title."),
			"status":          stringProp("Plan status."),
			"scope":           stringProp("Plan visibility scope."),
			"owner_entity_id": stringProp("Owner when scope is agent_private."),
			"evidence_json":   jsonProp("Optional structured evidence."),
			"items":           planItemsSchema(),
		})
	case "update_plan":
		return base([]string{"tenant_id", "workspace_id", "plan_id"}, map[string]any{
			"tenant_id":     stringProp("Tenant identifier."),
			"workspace_id":  stringProp("Workspace identifier."),
			"plan_id":       stringProp("Plan identifier."),
			"title":         stringProp("Optional replacement title."),
			"status":        stringProp("Optional replacement status."),
			"evidence_json": jsonProp("Optional structured evidence."),
			"items":         planItemsSchema(),
		})
	case "correct_memory":
		return base([]string{"tenant_id", "workspace_id", "memory_id", "operator_id", "correction_text"}, map[string]any{
			"tenant_id":       stringProp("Tenant identifier."),
			"workspace_id":    stringProp("Workspace identifier."),
			"memory_id":       stringProp("Memory being corrected."),
			"operator_id":     stringProp("Human or operator actor applying the correction."),
			"idempotency_key": stringProp("Optional correction idempotency key."),
			"correction_text": stringProp("Replacement truth or correction instruction."),
			"evidence_json":   jsonProp("Optional correction evidence."),
		})
	case "view_timeline":
		return base([]string{"tenant_id", "workspace_id", "entity_id"}, map[string]any{
			"tenant_id":    stringProp("Tenant identifier."),
			"workspace_id": stringProp("Workspace identifier."),
			"scopes":       scopeArray,
			"entity_id":    stringProp("Actor used for private timeline filtering."),
			"from":         stringProp("Optional RFC3339 lower time bound."),
			"to":           stringProp("Optional RFC3339 upper time bound."),
			"limit":        intProp("Maximum number of timeline items."),
		})
	case "explain_memory":
		return base([]string{"tenant_id", "workspace_id", "memory_id"}, map[string]any{
			"tenant_id":         stringProp("Tenant identifier."),
			"workspace_id":      stringProp("Workspace identifier."),
			"memory_id":         stringProp("Memory identifier to explain."),
			"entity_id":         stringProp("Actor used for private memory visibility."),
			"visible_group_ids": stringArrayProp("Group identifiers visible to the actor."),
		})
	default:
		return map[string]any{"type": "object"}
	}
}

func planItemsSchema() map[string]any {
	return map[string]any{
		"type":        "array",
		"description": "Structured plan items.",
		"items": map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"id":            map[string]any{"type": "string", "description": "Existing item identifier."},
				"title":         map[string]any{"type": "string", "description": "Item title."},
				"status":        map[string]any{"type": "string", "description": "Item status."},
				"evidence_json": map[string]any{"description": "Optional structured evidence."},
			},
		},
	}
}

func (s *Server) callTool(ctx context.Context, params json.RawMessage) (callToolResult, error) {
	var req callToolRequest
	if len(params) == 0 {
		return callToolResult{}, errors.New("tools/call params are required")
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return callToolResult{}, fmt.Errorf("decode tools/call params: %w", err)
	}
	if req.Name == "" {
		return callToolResult{}, errors.New("tools/call name is required")
	}
	if len(req.Arguments) == 0 {
		req.Arguments = json.RawMessage(`{}`)
	}
	raw, err := s.surface.Call(ctx, req.Name, req.Arguments)
	if err != nil {
		return callToolResult{
			Content: []textContent{{Type: "text", Text: err.Error()}},
			IsError: true,
		}, nil
	}
	var structured map[string]any
	if err := json.Unmarshal(raw, &structured); err != nil {
		return callToolResult{}, fmt.Errorf("decode tool output: %w", err)
	}
	return callToolResult{
		Content:           []textContent{{Type: "text", Text: string(raw)}},
		StructuredContent: structured,
		IsError:           false,
	}, nil
}

func marshalRPCResult(id json.RawMessage, result any) json.RawMessage {
	resp := rpcResponse{JSONRPC: jsonRPCVersion, ID: id, Result: result}
	raw, err := json.Marshal(resp)
	if err != nil {
		return marshalRPCError(id, errInternalError, "Internal error", nil)
	}
	return raw
}

func marshalRPCError(id json.RawMessage, code int, message string, data any) json.RawMessage {
	resp := rpcResponse{JSONRPC: jsonRPCVersion, ID: id, Error: &rpcError{Code: code, Message: message, Data: data}}
	raw, _ := json.Marshal(resp)
	return raw
}

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type implementationInfo struct {
	Name        string `json:"name"`
	Title       string `json:"title,omitempty"`
	Version     string `json:"version"`
	Description string `json:"description,omitempty"`
}

type initializeResult struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    serverCapabilities `json:"capabilities"`
	ServerInfo      implementationInfo `json:"serverInfo"`
	Instructions    string             `json:"instructions,omitempty"`
}

type serverCapabilities struct {
	Tools toolsCapability `json:"tools"`
}

type toolsCapability struct {
	ListChanged bool `json:"listChanged"`
}

type listToolsResult struct {
	Tools []protocolTool `json:"tools"`
}

type protocolTool struct {
	Name        string         `json:"name"`
	Title       string         `json:"title,omitempty"`
	Description string         `json:"description,omitempty"`
	InputSchema map[string]any `json:"inputSchema"`
}

type callToolRequest struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
}

type callToolResult struct {
	Content           []textContent  `json:"content"`
	StructuredContent map[string]any `json:"structuredContent,omitempty"`
	IsError           bool           `json:"isError"`
}

type textContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

```



<!-- Source: internal/mcp/protocol_test.go | bytes=7232 | lines=180 | sha16=c284eed301d1b8bd -->

```go
// ============================================================
// FILE     : internal/mcp/protocol_test.go
// PURPOSE  : Verifies MCP JSON-RPC lifecycle, tool listing, and tool call roundtrips.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : MCP protocol server tests
// DEPENDS  : internal/mcp/protocol.go, internal/mcp/surface_test.go
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: Protocol tests should prove real JSON-RPC shape, not only adapter delegation.
// ============================================================

package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

func TestServerHandlesInitializeAndToolRoundtrip(t *testing.T) {
	t.Parallel()

	surface := newTestSurface(t, &fakeService{
		prefetchResp: &core.PrefetchResponse{
			Blocks: []core.RecallBlock{{Kind: "pinned_note", Priority: 100, Text: "Keep Hermes first."}},
			Meta:   core.RecallMeta{EstimatedTokens: 4, Sources: []string{"notes"}},
		},
	})
	server := newProtocolServer(t, surface)

	initResp, respond := server.HandleMessage(context.Background(), json.RawMessage(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"test","version":"0"}}}`))
	if !respond {
		t.Fatalf("initialize did not produce a response")
	}
	var initEnvelope map[string]any
	decodeJSONMessage(t, initResp, &initEnvelope)
	result := initEnvelope["result"].(map[string]any)
	if result["protocolVersion"] != ProtocolVersion {
		t.Fatalf("unexpected protocol version: %#v", result["protocolVersion"])
	}
	if _, ok := result["capabilities"].(map[string]any)["tools"]; !ok {
		t.Fatalf("initialize response did not advertise tools: %s", string(initResp))
	}

	listResp, respond := server.HandleMessage(context.Background(), json.RawMessage(`{"jsonrpc":"2.0","id":"tools","method":"tools/list"}`))
	if !respond {
		t.Fatalf("tools/list did not produce a response")
	}
	if !strings.Contains(string(listResp), `"prefetch"`) || !strings.Contains(string(listResp), `"inputSchema"`) {
		t.Fatalf("tools/list missing expected tool schema: %s", string(listResp))
	}
	if !strings.Contains(string(listResp), `"required":["tenant_id","workspace_id"]`) || !strings.Contains(string(listResp), `"recall_preview"`) {
		t.Fatalf("tools/list did not expose recall preview required inputs: %s", string(listResp))
	}

	callResp, respond := server.HandleMessage(context.Background(), json.RawMessage(`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"prefetch","arguments":{"tenant_id":"tenant_1","workspace_id":"workspace_1","session_id":"session_1","actor_id":"agent:hermes-main"}}}`))
	if !respond {
		t.Fatalf("tools/call did not produce a response")
	}
	if !strings.Contains(string(callResp), `"structuredContent"`) || !strings.Contains(string(callResp), "Keep Hermes first.") {
		t.Fatalf("tools/call did not return structured prefetch output: %s", string(callResp))
	}
}

func TestServerToolSchemasExposeTrustLoopInputs(t *testing.T) {
	t.Parallel()

	server := newProtocolServer(t, newTestSurface(t, &fakeService{}))
	tools := server.protocolTools()

	byName := make(map[string]protocolTool, len(tools))
	for _, tool := range tools {
		byName[tool.Name] = tool
	}
	assertRequiredInputs(t, byName["recall_preview"], "tenant_id", "workspace_id")
	assertRequiredInputs(t, byName["correct_memory"], "tenant_id", "workspace_id", "memory_id", "operator_id", "correction_text")
	assertRequiredInputs(t, byName["view_timeline"], "tenant_id", "workspace_id", "entity_id")
	assertRequiredInputs(t, byName["explain_memory"], "tenant_id", "workspace_id", "memory_id")

	correctionProps := byName["correct_memory"].InputSchema["properties"].(map[string]any)
	if _, ok := correctionProps["evidence_json"]; !ok {
		t.Fatalf("correct_memory schema should expose evidence_json for provenance")
	}
	timelineProps := byName["view_timeline"].InputSchema["properties"].(map[string]any)
	if _, ok := timelineProps["scopes"]; !ok {
		t.Fatalf("view_timeline schema should expose scopes for visibility review")
	}
	explainProps := byName["explain_memory"].InputSchema["properties"].(map[string]any)
	if _, ok := explainProps["entity_id"]; !ok {
		t.Fatalf("explain_memory schema should expose entity_id for private visibility")
	}
	if _, ok := explainProps["visible_group_ids"]; !ok {
		t.Fatalf("explain_memory schema should expose visible_group_ids for group visibility")
	}
}

func TestServerServeStdioRoundtrip(t *testing.T) {
	t.Parallel()

	surface := newTestSurface(t, &fakeService{prefetchResp: &core.PrefetchResponse{Blocks: []core.RecallBlock{{Kind: "note", Text: "stdio ok"}}}})
	server := newProtocolServer(t, surface)
	input := strings.Join([]string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"test","version":"0"}}}`,
		`{"jsonrpc":"2.0","method":"notifications/initialized"}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"prefetch","arguments":{"tenant_id":"tenant_1","workspace_id":"workspace_1"}}}`,
		"",
	}, "\n")
	var out bytes.Buffer

	if err := server.ServeStdio(context.Background(), strings.NewReader(input), &out); err != nil {
		t.Fatalf("ServeStdio returned error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected two JSON-RPC responses, got %d: %q", len(lines), out.String())
	}
	if !strings.Contains(lines[0], `"protocolVersion":"2025-11-25"`) {
		t.Fatalf("first response was not initialize: %s", lines[0])
	}
	if !strings.Contains(lines[1], "stdio ok") {
		t.Fatalf("second response was not tool output: %s", lines[1])
	}
}

func TestServerReturnsProtocolErrorForUnknownMethod(t *testing.T) {
	t.Parallel()

	server := newProtocolServer(t, newTestSurface(t, &fakeService{}))
	raw, respond := server.HandleMessage(context.Background(), json.RawMessage(`{"jsonrpc":"2.0","id":1,"method":"missing"}`))
	if !respond {
		t.Fatalf("unknown method did not produce a response")
	}
	if !strings.Contains(string(raw), `"code":-32601`) {
		t.Fatalf("expected method-not-found error, got %s", string(raw))
	}
}

func newProtocolServer(t *testing.T, surface *Surface) *Server {
	t.Helper()

	server, err := NewServer(surface)
	if err != nil {
		t.Fatalf("NewServer returned error: %v", err)
	}
	return server
}

func decodeJSONMessage(t *testing.T, raw json.RawMessage, out any) {
	t.Helper()

	if err := json.Unmarshal(raw, out); err != nil {
		t.Fatalf("decode JSON message: %v; raw=%s", err, string(raw))
	}
}

func assertRequiredInputs(t *testing.T, tool protocolTool, want ...string) {
	t.Helper()

	required, ok := tool.InputSchema["required"].([]string)
	if !ok {
		t.Fatalf("%s schema missing required input list: %#v", tool.Name, tool.InputSchema)
	}
	got := make(map[string]bool, len(required))
	for _, field := range required {
		got[field] = true
	}
	for _, field := range want {
		if !got[field] {
			t.Fatalf("%s schema missing required field %q in %#v", tool.Name, field, required)
		}
	}
}

```



<!-- Source: internal/mcp/surface.go | bytes=4457 | lines=118 | sha16=884e24371a261ba5 -->

```go
// ============================================================
// FILE     : internal/mcp/surface.go
// PURPOSE  : Exposes VibeGravity core operations as a small MCP-style tool surface.
// LAYER    : interface
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : Surface, Tool, NewSurface
// DEPENDS  : context, encoding/json, fmt, internal/core
// USED_BY  : MCP integration tests, future protocol server
// ------------------------------------------------------------
// AGENT_NOTE: This surface must stay a thin adapter over core service semantics.
// ============================================================

package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

// Tool describes one MCP-visible VibeGravity operation.
type Tool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Surface exposes core VibeGravity operations through tool names.
type Surface struct {
	service core.VibeGravityService
}

// NewSurface creates an MCP-style adapter over the shared core service.
func NewSurface(service core.VibeGravityService) (*Surface, error) {
	if service == nil {
		return nil, fmt.Errorf("%w: mcp service is required", core.ErrInvalidArgument)
	}
	return &Surface{service: service}, nil
}

// Tools lists the current MCP tool surface.
func (s *Surface) Tools() []Tool {
	return []Tool{
		{Name: "prefetch", Description: "Assemble a typed recall pack."},
		{Name: "recall_preview", Description: "Preview scoped recall with source and freshness metadata."},
		{Name: "sync_turn", Description: "Record raw turn events and enqueue memory processing."},
		{Name: "search_memory", Description: "Search visible memories."},
		{Name: "search_documents", Description: "Search document chunks."},
		{Name: "add_note", Description: "Create a human-authored recall note."},
		{Name: "create_plan", Description: "Create a structured plan."},
		{Name: "update_plan", Description: "Update a structured plan."},
		{Name: "correct_memory", Description: "Record a correction intent."},
		{Name: "view_timeline", Description: "Read memory and correction activity."},
		{Name: "explain_memory", Description: "Read provenance for one memory."},
	}
}

// Call decodes a JSON tool request, delegates to the core service, and returns JSON.
func (s *Surface) Call(ctx context.Context, name string, input json.RawMessage) (json.RawMessage, error) {
	if s == nil || s.service == nil {
		return nil, fmt.Errorf("%w: mcp surface is not initialized", core.ErrInvalidArgument)
	}
	switch name {
	case "prefetch", "recall_preview":
		var req core.PrefetchRequest
		return callJSON(ctx, input, &req, s.service.Prefetch)
	case "sync_turn":
		var req core.SyncTurnRequest
		return callJSON(ctx, input, &req, s.service.SyncTurn)
	case "search_memory":
		var req core.SearchMemoriesRequest
		return callJSON(ctx, input, &req, s.service.SearchMemories)
	case "search_documents":
		var req core.SearchDocumentsRequest
		return callJSON(ctx, input, &req, s.service.SearchDocuments)
	case "add_note":
		var req core.AddNoteRequest
		return callJSON(ctx, input, &req, s.service.AddNote)
	case "create_plan":
		var req core.CreatePlanRequest
		return callJSON(ctx, input, &req, s.service.CreatePlan)
	case "update_plan":
		var req core.UpdatePlanRequest
		return callJSON(ctx, input, &req, s.service.UpdatePlan)
	case "correct_memory":
		var req core.CorrectMemoryRequest
		return callJSON(ctx, input, &req, s.service.CorrectMemory)
	case "view_timeline":
		var req core.GetTimelineRequest
		return callJSON(ctx, input, &req, s.service.GetTimeline)
	case "explain_memory":
		var req core.ExplainMemoryRequest
		return callJSON(ctx, input, &req, s.service.ExplainMemory)
	default:
		return nil, fmt.Errorf("%w: unknown mcp tool %q", core.ErrInvalidArgument, name)
	}
}

func callJSON[Req any, Resp any](ctx context.Context, input json.RawMessage, req *Req, call func(context.Context, *Req) (*Resp, error)) (json.RawMessage, error) {
	if len(input) == 0 {
		input = json.RawMessage(`{}`)
	}
	if err := json.Unmarshal(input, req); err != nil {
		return nil, fmt.Errorf("%w: decode mcp tool input: %v", core.ErrInvalidArgument, err)
	}
	resp, err := call(ctx, req)
	if err != nil {
		return nil, err
	}
	output, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("encode mcp tool output: %w", err)
	}
	return output, nil
}

```



<!-- Source: internal/mcp/surface_test.go | bytes=8166 | lines=220 | sha16=9bcfbc5d30cdf32b -->

```go
// ============================================================
// FILE     : internal/mcp/surface_test.go
// PURPOSE  : Verifies the MCP-style tool surface delegates to shared core semantics.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : MCP surface adapter tests
// DEPENDS  : context, encoding/json, errors, strings, testing, internal/core
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: These tests do not start a protocol server; they lock tool-to-core mapping.
// ============================================================

package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

func TestSurfaceListsV1Tools(t *testing.T) {
	t.Parallel()

	surface := newTestSurface(t, &fakeService{})
	tools := surface.Tools()
	names := make(map[string]bool, len(tools))
	for _, tool := range tools {
		names[tool.Name] = true
	}
	for _, want := range []string{"prefetch", "recall_preview", "sync_turn", "search_memory", "add_note", "correct_memory", "view_timeline", "explain_memory"} {
		if !names[want] {
			t.Fatalf("expected tool %q in %#v", want, tools)
		}
	}
}

func TestSurfaceCallsRecallPreviewAlias(t *testing.T) {
	t.Parallel()

	service := &fakeService{prefetchResp: &core.PrefetchResponse{Blocks: []core.RecallBlock{{Kind: "memory", Priority: 90, Text: "Preview"}}}}
	surface := newTestSurface(t, service)

	raw, err := surface.Call(context.Background(), "recall_preview", json.RawMessage(`{"tenant_id":"tenant_1","workspace_id":"workspace_1"}`))
	if err != nil {
		t.Fatalf("Call returned error: %v", err)
	}
	if service.prefetchCalls != 1 {
		t.Fatalf("expected recall_preview to delegate to prefetch, got %d calls", service.prefetchCalls)
	}
	if !strings.Contains(string(raw), "Preview") {
		t.Fatalf("expected encoded recall preview output, got %s", string(raw))
	}
}

func TestSurfaceCallsPrefetch(t *testing.T) {
	t.Parallel()

	service := &fakeService{prefetchResp: &core.PrefetchResponse{Blocks: []core.RecallBlock{{Kind: "note", Priority: 100, Text: "Pinned"}}}}
	surface := newTestSurface(t, service)

	raw, err := surface.Call(context.Background(), "prefetch", json.RawMessage(`{"tenant_id":"tenant_1","workspace_id":"workspace_1"}`))
	if err != nil {
		t.Fatalf("Call returned error: %v", err)
	}
	if service.prefetchCalls != 1 {
		t.Fatalf("expected one prefetch call, got %d", service.prefetchCalls)
	}
	if !strings.Contains(string(raw), "Pinned") {
		t.Fatalf("expected encoded prefetch output, got %s", string(raw))
	}
}

func TestSurfaceCallsCorrectMemory(t *testing.T) {
	t.Parallel()

	service := &fakeService{correctResp: &core.CorrectMemoryResponse{MemoryID: "mem_1", CorrectionRecorded: true, Status: "recorded"}}
	surface := newTestSurface(t, service)

	raw, err := surface.Call(context.Background(), "correct_memory", json.RawMessage(`{"tenant_id":"tenant_1","workspace_id":"workspace_1","memory_id":"mem_1"}`))
	if err != nil {
		t.Fatalf("Call returned error: %v", err)
	}
	if service.correctCalls != 1 {
		t.Fatalf("expected one correction call, got %d", service.correctCalls)
	}
	if !strings.Contains(string(raw), `"correction_recorded":true`) {
		t.Fatalf("expected encoded correction output, got %s", string(raw))
	}
}

func TestSurfaceCallsViewTimeline(t *testing.T) {
	t.Parallel()

	service := &fakeService{timelineResp: &core.GetTimelineResponse{Items: []core.TimelineItem{{ID: "item_1", MemoryID: "mem_1", Text: "Corrected project rule"}}}}
	surface := newTestSurface(t, service)

	raw, err := surface.Call(context.Background(), "view_timeline", json.RawMessage(`{"tenant_id":"tenant_1","workspace_id":"workspace_1","entity_id":"agent:hermes-main"}`))
	if err != nil {
		t.Fatalf("Call returned error: %v", err)
	}
	if service.timelineCalls != 1 {
		t.Fatalf("expected one timeline call, got %d", service.timelineCalls)
	}
	if !strings.Contains(string(raw), "Corrected project rule") {
		t.Fatalf("expected encoded timeline output, got %s", string(raw))
	}
}

func TestSurfaceCallsExplainMemory(t *testing.T) {
	t.Parallel()

	service := &fakeService{explainResp: &core.ExplainMemoryResponse{MemoryID: "mem_1", Trace: core.MemoryTraceResult{ReasoningJobID: "job_1"}}}
	surface := newTestSurface(t, service)

	raw, err := surface.Call(context.Background(), "explain_memory", json.RawMessage(`{"tenant_id":"tenant_1","workspace_id":"workspace_1","memory_id":"mem_1","entity_id":"agent:hermes-main","visible_group_ids":["group_design"]}`))
	if err != nil {
		t.Fatalf("Call returned error: %v", err)
	}
	if service.explainCalls != 1 {
		t.Fatalf("expected one explain call, got %d", service.explainCalls)
	}
	if service.explainReq == nil || service.explainReq.EntityID != "agent:hermes-main" {
		t.Fatalf("expected explain visibility request, got %#v", service.explainReq)
	}
	if got := service.explainReq.VisibleGroupIDs; len(got) != 1 || got[0] != "group_design" {
		t.Fatalf("expected visible group ids, got %#v", service.explainReq)
	}
	if !strings.Contains(string(raw), `"reasoning_job_id":"job_1"`) {
		t.Fatalf("expected encoded explain output, got %s", string(raw))
	}
}

func TestSurfaceRejectsUnknownToolAndInvalidJSON(t *testing.T) {
	t.Parallel()

	surface := newTestSurface(t, &fakeService{})
	if _, err := surface.Call(context.Background(), "unknown", json.RawMessage(`{}`)); !errors.Is(err, core.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument for unknown tool, got %v", err)
	}
	if _, err := surface.Call(context.Background(), "prefetch", json.RawMessage(`{`)); !errors.Is(err, core.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument for invalid JSON, got %v", err)
	}
}

func newTestSurface(t *testing.T, service core.VibeGravityService) *Surface {
	t.Helper()

	surface, err := NewSurface(service)
	if err != nil {
		t.Fatalf("NewSurface returned error: %v", err)
	}
	return surface
}

type fakeService struct {
	prefetchCalls int
	correctCalls  int
	timelineCalls int
	explainCalls  int
	prefetchResp  *core.PrefetchResponse
	correctResp   *core.CorrectMemoryResponse
	timelineResp  *core.GetTimelineResponse
	explainResp   *core.ExplainMemoryResponse
	explainReq    *core.ExplainMemoryRequest
}

func (s *fakeService) Prefetch(context.Context, *core.PrefetchRequest) (*core.PrefetchResponse, error) {
	s.prefetchCalls++
	return s.prefetchResp, nil
}

func (s *fakeService) SyncTurn(context.Context, *core.SyncTurnRequest) (*core.SyncTurnResponse, error) {
	return &core.SyncTurnResponse{Status: "accepted"}, nil
}

func (s *fakeService) AddDocument(context.Context, *core.AddDocumentRequest) (*core.AddDocumentResponse, error) {
	return &core.AddDocumentResponse{Status: "created"}, nil
}

func (s *fakeService) SearchMemories(context.Context, *core.SearchMemoriesRequest) (*core.SearchMemoriesResponse, error) {
	return &core.SearchMemoriesResponse{}, nil
}

func (s *fakeService) SearchDocuments(context.Context, *core.SearchDocumentsRequest) (*core.SearchDocumentsResponse, error) {
	return &core.SearchDocumentsResponse{}, nil
}

func (s *fakeService) AddNote(context.Context, *core.AddNoteRequest) (*core.AddNoteResponse, error) {
	return &core.AddNoteResponse{Status: "created"}, nil
}

func (s *fakeService) CreatePlan(context.Context, *core.CreatePlanRequest) (*core.CreatePlanResponse, error) {
	return &core.CreatePlanResponse{Status: "created"}, nil
}

func (s *fakeService) UpdatePlan(context.Context, *core.UpdatePlanRequest) (*core.UpdatePlanResponse, error) {
	return &core.UpdatePlanResponse{Status: "updated"}, nil
}

func (s *fakeService) CorrectMemory(context.Context, *core.CorrectMemoryRequest) (*core.CorrectMemoryResponse, error) {
	s.correctCalls++
	return s.correctResp, nil
}

func (s *fakeService) GetTimeline(context.Context, *core.GetTimelineRequest) (*core.GetTimelineResponse, error) {
	s.timelineCalls++
	return s.timelineResp, nil
}

func (s *fakeService) ExplainMemory(_ context.Context, req *core.ExplainMemoryRequest) (*core.ExplainMemoryResponse, error) {
	s.explainCalls++
	s.explainReq = req
	return s.explainResp, nil
}

```
