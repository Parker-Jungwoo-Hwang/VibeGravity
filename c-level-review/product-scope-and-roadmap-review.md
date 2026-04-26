# Product Scope and Roadmap Review

Date: 2026-04-26  
Reviewer: Agent 03, hybrid CPO/CTO review  
Scope: Feature scope, product coherence, and roadmap focus for the VibeGravity repository.

## Review Request

Agent 03 was asked to review the GitHub repository `VibeGravity` from a C-level product and technology perspective.

The requested review questions were:

- What are the core features?
- Which features support the main product promise?
- Which features feel secondary or premature?
- Are there too many agents, skills, or modes for the current maturity?
- Does the changelog show a coherent product path?
- What should be cut, hidden, delayed, or promoted?
- What should the next 30 days focus on?

This review is read-only and uses the repository as the source of truth.

## Sources Reviewed

Primary source files and folders reviewed:

- `PLANS.md`
- `plans/00_read-this-first_for-building-agents.md`
- `plans/01_rfp_vibegravity_hermes-first.md`
- `plans/02_product-contract_and_direction.md`
- `plans/03_target-architecture_codex-first.md`
- `plans/04_memory-scopes_dreaming_ontology-lite.md`
- `plans/05_runtime-contracts_ingest-recall-apply.md`
- `plans/06_data-model_and_storage-invariants.md`
- `plans/10_workpack_hermes-provider-and-external-surfaces.md`
- `plans/11_workpack_quality-ops-and-evals.md`
- `consulting/00_consulting_request.md`
- `consulting/02_product_one_pager.md`
- `consulting/03_engine_positioning_and_narrative.md`
- `consulting/04_customer_and_use_cases.md`
- `consulting/05_mvp_scope_and_non_goals.md`
- `consulting/06_runtime_and_product_contract.md`
- `consulting/07_current_state_and_roadmap.md`
- `consulting/08_risks_and_open_decisions.md`
- `docs/review-packets/hermes-memory-trust-loop-product-pivot.md`
- `docs/review-packets/gpt-pro-followup-product-contract-alignment.md`
- `docs/review-packets/hermes-memory-demo-eval.md`
- `docs/review-packets/v1-trust-loop-readiness-report.md`
- `.agents/skills/`
- `.agents/coordination/`
- `.agents/hermes-orchestration/`
- HTTP, MCP, Hermes provider, CLI, and eval surfaces under `cmd/` and `internal/`.

Not observed in the current repo root:

- `README.md`
- `CHANGELOG.md`
- `.github/workflows/`
- `web-docs/`

That absence matters for product readiness because the internal roadmap is coherent, but the public-facing product path is not yet packaged as a clean external story.

## Executive Verdict

The repo has a focused product thesis: **Hermes Memory, powered by VibeGravity**.

The strongest product promise is clear: Hermes should remember scoped project context across sessions, show why it remembered something, and let the operator correct memory once. That is a compelling wedge because it turns memory from an invisible infrastructure feature into a trust loop.

The current repo is directionally strong but surface-heavy. It exposes or plans HTTP APIs, MCP tools, Hermes provider hooks, CLI ops, notes, plans, documents, Dreaming, replay evals, backlog metrics, correction, timeline, explain, and multi-agent orchestration. Many of those are legitimate engineering support systems, but they should not all be visible as V1 product value.

The roadmap is coherent after the Hermes Memory pivot, but the repo still carries too much scaffold for its current maturity. The next product milestone should not be "more memory features." It should be proof that the Hermes Memory trust loop works through live PostgreSQL and the actual external protocol path Hermes will use.

Until that is proven, VibeGravity should reduce visible scope, hide secondary features, and package the project around one operator-obvious demo.

## Core Product Thesis

VibeGravity should be the engine behind Hermes Memory: scoped, explainable, correctable memory that makes Hermes continuous across sessions.

## Core Features

The true core features are the ones that directly prove the product promise.

### 1. Hot-path ingest

`sync_turn()` records raw agent activity after a turn without blocking on heavy reasoning.

This is core because the product cannot be trusted if recording agent activity is slow, lossy, or non-idempotent. The repo correctly treats raw events as separate from derived memories.

### 2. Typed recall

`prefetch()` returns compact, typed, budget-aware recall blocks before the next turn.

This is the first user-visible value moment. If recall is noisy, too long, stale without warning, or privacy-blind, the product fails even if the architecture is elegant.

### 3. Scope separation

The memory model distinguishes:

- `agent_private`
- `workspace_shared`
- `group_shared`
- `session_scratch`

This is not merely a permission detail. It is a product feature. Multi-agent memory only becomes useful if users believe private and shared memory will not blur.

### 4. Correction and supersession

The active product contract is no longer record-only correction. The current V1 direction requires correction-driven supersession:

- raw correction event
- append-safe correction artifact
- replacement memory
- mandatory trace
- `updates` edge
- prior memory supersession
- next recall suppresses outdated memory

This is the heart of the trust loop. The user should be able to fix memory once and see that future recall changed.

### 5. Provenance and inspection

`ExplainMemory` and `GetTimeline` are core trust surfaces.

They let the operator inspect why Hermes remembered something, where it came from, whether it was corrected, and whether it is still current.

### 6. Degraded freshness visibility

Recall must honestly label stale or degraded state when Codex, worker, or backlog conditions mean memory may be behind raw events.

This prevents a dangerous product failure mode: stale memory presented as fresh intelligence.

### 7. Local quality gates

Golden evals, replay checks, and the local Hermes Memory demo are core support infrastructure.

They are not customer-facing product features, but they are essential because memory products can fail silently.

## Features That Support the Main Product Promise

These features support Hermes Memory V1 but should remain subordinate to the trust loop.

### Notes

Pinned notes give operators direct control over recall priority. This supports the promise because human-authored context can outrank noisy learned memory.

### Plans

Active plans are a strong recall source because they help Hermes continue work across sessions. Plan creation and update are useful, but rich plan management should not become a separate product.

### Search memory

Memory search supports inspection and operator control. It should be available, but the product should not position itself as search-first.

### MCP surface

MCP is the practical protocol path for Hermes-facing and operator-facing tools. It supports the product promise if it preserves the exact same trust-loop semantics as core service validation.

### Hermes provider adapter

The in-repo Hermes provider adapter is important because Hermes is the first customer. However, custom provider packaging should not block V1 if MCP is the actual viable integration path right now.

### CLI doctor and job metrics

These support local operation and debugging. They are important for beta readiness, but they are not the main product promise.

### Document storage and search

Documents can support recall and Stage 2 reasoning. They should remain supporting context, not the V1 headline.

### Dreaming

Dreaming can improve quality over time as a background maintenance layer. It should not be used to sell V1 before recall, correction, provenance, and protocol correctness are proven.

## Features That Feel Secondary or Premature

### Rich document memory

Document ingestion and search are useful, but they are not required to make the first Hermes Memory demo compelling. Promote documents later, after the trust loop is working.

### Dreaming as a visible product concept

Dreaming is architecturally useful, but product-visible Dreaming is premature. It risks shifting attention away from the more urgent correction and provenance loop.

### Custom Hermes provider registry packaging

The repo notes that Hermes' current CLI constraints make MCP registration the practical route. Custom provider packaging should be delayed until the protocol path is proven.

### Real Codex enabled by default

The repo correctly keeps real Codex disabled by default. Enabling it before failure behavior, retry policy, and degraded-mode UX are explicit would create trust risk.

### Broad agent integrations

Claude Code, Codex as a client, and other future adapters should remain post-V1. Generic multi-agent platform ambition is a distraction before Hermes Memory is proven.

### Operator UI

A small operator UI may become valuable, especially for timeline and explain. But V1 can proceed with CLI/MCP/HTTP if the demo is clear and the protocol path is real.

### Large ontology or knowledge graph positioning

The ontology-lite model is useful internally. Positioning the product as a knowledge graph platform would blur the wedge.

## Are There Too Many Agents, Skills, or Modes?

There are not too many repo-local skills in the narrow sense. `.agents/skills/` contains a small set of useful procedural skills:

- code headers
- contract check
- eval regression
- plan/implement/verify
- source provenance

Those are appropriate for an agent-built codebase with strict contracts.

The larger concern is not the number of skills. It is the number of operational and product surfaces:

- HTTP API
- MCP surface
- Hermes provider adapter
- CLI doctor
- CLI evals
- CLI jobs metrics
- blocked-job recovery
- Hermes bootstrap
- notes
- plans
- documents
- search
- correction
- timeline
- explain
- Dreaming jobs
- replay evals
- repo-local Hermes multi-profile orchestration

For the current maturity, this is too much to present as product scope. It is acceptable as internal scaffolding if the repo is disciplined about what is visible in V1.

The recommended posture:

- Keep the agent coordination system internal.
- Keep repo-local Hermes orchestration internal.
- Keep Dreaming internal.
- Keep documents as support.
- Present only Hermes Memory trust-loop tools as V1 product surface.

## Changelog and Product Path Diagnosis

There is no `CHANGELOG.md` in the current repo root.

That means the changelog does not yet show a coherent product path because it does not exist as a public artifact. Internally, the product path is visible through `PLANS.md`, consulting docs, and review packets:

1. Go-first foundation.
2. Ingest and recall.
3. Graph apply and worker safety.
4. Correction and supersession.
5. Hermes/MCP external surfaces.
6. Quality, evals, degraded-mode, and trust-loop readiness.
7. Product pivot to Hermes Memory.

That internal path is coherent. It shows a repo moving from infrastructure toward a sharper product wedge. But externally, a new reader would not see that path quickly because the root-level packaging is missing.

Recommendation: add a concise `CHANGELOG.md` before any public beta or external review. It should not list every agent lane. It should tell the product story:

- foundation
- recall
- correction
- provenance
- Hermes/MCP path
- trust-loop proof
- remaining blockers

## Roadmap Diagnosis

The roadmap is strategically focused but operationally bloated.

The strategic focus is good: Hermes Memory trust loop, PostgreSQL correctness, MCP/Hermes protocol path, and degraded truthfulness.

The operational breadth is high: the repo already carries documents, Dreaming, multi-surface APIs, CLI ops, evals, orchestration, and future adapter planning. This breadth is understandable for a memory engine, but it can dilute delivery if all of it competes for V1 attention.

The safest product diagnosis is:

> Focused thesis, bloated surface, incomplete proof.

The roadmap should now narrow around proof, not expansion.

## What Should Be Promoted

Promote these as V1 product language and demos:

- Hermes Memory, powered by VibeGravity.
- Recall preview before a Hermes turn.
- Scope labels on recalled memory.
- Explain why Hermes remembered something.
- Correct memory once.
- Supersede outdated memory.
- Show degraded or stale freshness honestly.
- Prove private memory does not leak into workspace recall.

These are the features that make the product emotionally and operationally legible.

## What Should Be Hidden

Hide these from the main V1 story:

- Dreaming terminology.
- Document memory as a headline.
- Internal worker pipeline details.
- Multi-agent orchestration.
- Repo-local Hermes profile dispatch.
- Broad future adapters.
- Golden eval implementation detail.
- Graph terminology except where needed for provenance.

These can stay in technical docs, but they should not be the first product narrative.

## What Should Be Delayed

Delay these until after the trust loop is proven:

- Real Codex default enablement.
- Custom Hermes memory provider packaging.
- Claude Code and Codex client integrations.
- Advanced document recall.
- Rich timeline visualization.
- Small operator UI beyond what is needed for proof.
- Production-grade Dreaming quality.
- Profile coherence scoring.
- Organization-wide admin or multi-tenant management surfaces.

## What Should Be Cut or Removed From V1

Cut these from V1 scope, even if implementation scaffolding remains:

- Generic chat UI.
- Large web app.
- Heavy ontology platform.
- General-purpose vector database positioning.
- Marketplace of memory integrations.
- Fully autonomous forgetting without operator visibility.
- Multi-node distributed queue redesign.
- Any claim that documents or Dreaming are the V1 value story.

Also cut or quarantine old correction language that says correction is record-only or must not supersede. That was historically true for an earlier slice, but it conflicts with the active V1 trust-loop contract.

## 30-Day Roadmap

The next 30 days should focus on turning the trust loop from a local deterministic story into a real first-customer proof.

### 1. Prove live PostgreSQL correction trust loop

Run the correction flow against a migrated scratch PostgreSQL database:

- `CorrectMemory`
- replacement memory
- mandatory `memory_trace`
- `updates` edge
- target supersession
- idempotent retry
- `ExplainMemory`
- `GetTimeline`
- next `Prefetch` suppresses stale memory

This is the highest-impact work because the repo itself says V1 readiness cannot be claimed from SQLite, mocks, or in-memory evals alone.

### 2. Prove the Hermes-facing protocol path

Use the real MCP stdio path that Hermes can register:

- `cli mcp serve --stdio`
- `cli hermes bootstrap`
- `hermes mcp test vibegravity`
- tool calls for recall preview, correction, explain, timeline, and degraded status

The product is Hermes-first. Therefore the external path must work, not just the core service.

### 3. Package the 5-minute Hermes Memory demo

Create one scripted demo that proves:

1. A project rule and active plan exist.
2. Hermes receives compact recall.
3. A wrong memory is visible.
4. The operator explains it.
5. The operator corrects it.
6. Later recall includes the correction and suppresses the old memory.
7. Private memory does not leak into shared recall.
8. Degraded state is labeled honestly.

This demo should become the product gate.

### 4. Add public-facing repo packaging

Add or prepare:

- `README.md`
- `CHANGELOG.md`
- clear V1 status
- install/bootstrap notes
- trust-loop demo instructions
- known limitations

The repo currently has strong internal plans but weak external orientation.

### 5. Freeze visible V1 surface

For 30 days, do not add new feature categories. Keep visible V1 to:

- recall preview
- sync turn
- search memory
- add note
- create/update plan only as recall control
- correct memory
- explain memory
- view timeline
- degraded status

Documents and Dreaming should remain support layers.

## 90-Day Roadmap

The next 90 days should convert the proven trust loop into a local beta-quality product.

### 1. Local beta packaging

Make the local run path boring and repeatable:

- setup
- migration
- doctor
- bootstrap
- start server
- start worker
- run MCP
- run demo
- backup/restore notes

### 2. Production-quality replay and evals

Move beyond narrow deterministic evals toward full session replay:

- compare replay outputs across code changes
- measure duplicate memory rate
- measure correction propagation
- measure superseded-memory leakage
- measure scope-label coverage
- measure degraded recall usefulness

### 3. Real Codex enablement behind explicit gates

Only enable real Codex after:

- prompt builder exists
- retry policy exists
- failure mode is operator-visible
- degraded freshness remains truthful
- no local extractor fallback is reintroduced
- tests prove structured JSON boundaries

### 4. Dreaming as quality, not headline

Improve Dreaming after the trust loop works:

- session summaries
- tier metadata
- promotion quality
- suppression of low-value memory
- no scope mutation

Do not sell it as the V1 product center.

### 5. Documents as supporting memory

Promote document memory only after:

- source provenance is visible
- document recall is budget-aware
- document facts do not crowd out active plans, notes, or corrected memory

### 6. Second integration only after Hermes success

Consider Claude Code or Codex-as-client after Hermes Memory is demonstrably useful. The second adapter should validate that the core semantics generalize without changing the product.

## Kill or Defer List

| Feature or Surface | Decision | Reason |
|---|---|---|
| Generic chat UI | Cut from V1 | Not the product. |
| Large web app | Defer | Trust loop can be proven without it. |
| Heavy ontology platform | Cut from V1 | Ontology-lite is enough. |
| Documents as headline | Hide/defer | Supporting context, not V1 promise. |
| Dreaming as headline | Hide/defer | Maintenance layer, not first proof. |
| Real Codex default enablement | Defer | Requires explicit failure/degraded behavior. |
| Custom Hermes provider registry packaging | Defer | MCP is the current practical path. |
| Claude Code integration | Defer | Post-Hermes expansion. |
| Codex client integration | Defer | Do not confuse reasoning backend with client surface. |
| Group-shared writes | Defer | Needs explicit membership/write semantics. |
| Operator UI | Defer lightly | Useful later, not required before protocol proof. |
| Multi-agent orchestration as product | Hide | Internal delivery mechanism. |
| Historical correction prep docs | Quarantine | Some describe older record-only semantics. |

## Answer to Each Requested Question

### What are the core features?

The core features are `sync_turn()`, `prefetch()`, scoped memory, provenance, correction, supersession, explain/timeline, degraded freshness labeling, and local quality gates.

### Which features support the main product promise?

Notes, plans, search, MCP, Hermes adapter, CLI doctor, backlog metrics, evals, documents, and Dreaming all support the promise when they remain subordinate to recall, correction, scope, and provenance.

### Which features feel secondary or premature?

Rich document memory, visible Dreaming, custom provider packaging, real Codex default enablement, broad adapter support, operator UI, and profile scoring are secondary or premature.

### Are there too many agents, skills, or modes for the current maturity?

There are not too many local skills. There are too many visible surfaces if the repo tries to present all of them as V1. The coordination and Hermes orchestration layers should remain internal.

### Does the changelog show a coherent product path?

No, because there is no root `CHANGELOG.md`. Internally the product path is coherent through plans and review packets, but externally it is not yet packaged.

### What should be cut, hidden, delayed, or promoted?

Promote Hermes Memory trust-loop features. Hide Dreaming, documents, orchestration, and graph internals. Delay real Codex default, custom provider packaging, broad adapters, and UI. Cut generic chat, vector database, and heavy ontology positioning.

### What should the next 30 days focus on?

The next 30 days should focus on live PostgreSQL trust-loop proof, real Hermes/MCP protocol roundtrip, the 5-minute demo, root README/CHANGELOG packaging, and freezing visible V1 scope.

## Final CTO/CPO Decision

VibeGravity needs focus before more feature work.

The repo is not conceptually scattered anymore; the Hermes Memory pivot gives it a strong product center. But the implementation and documentation still expose too many secondary surfaces for the current maturity.

The next milestone should be a narrow, evidence-backed V1 trust loop:

> Hermes recalls scoped context, the operator can inspect why, the operator corrects a bad memory once, later recall changes, stale memory is suppressed, private memory stays private, and degraded state is visible.

Until that path passes through live PostgreSQL and the Hermes-facing protocol, new feature breadth should be treated as a distraction.

## Source Review

- Estimated source: VibeGravity repository documents, implementation surfaces, review packets, and prior read-only Agent 03 product review.
- Suspected license: project-internal original material.
- Similarity risk: low; this report is original synthesis from the repository.
- Human review required: yes, because this is a product and roadmap decision document.
