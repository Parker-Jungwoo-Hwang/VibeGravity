# Product Positioning Review

Date: 2026-04-26  
Reviewer: Agent 01  
Role: CPO review  
Repo: `/Users/parker/Documents/VibeGravity`

## Review Scope

This review evaluates VibeGravity from the view of product positioning, customer value, and target users.

The review used the repository as the source of truth and focused on public or product-promise surfaces:

- root public docs if present: `README.md`, `README_VI.md`, `CHANGELOG.md`
- `PLANS.md`
- `consulting/`
- product and architecture plans under `plans/`
- review packets under `docs/review-packets/`
- implementation surfaces that prove or limit the product promise

No files were edited during the original CPO review. This report records that review result for C-level follow-up.

## Executive Verdict

VibeGravity is best understood as **Hermes Memory, powered by VibeGravity**, not as a generic shared memory platform.

The product is for technical Hermes operators who keep restarting long-running coding or project work and hate restating context.

The painful problem is not "lack of storage"; it is unreliable memory: stale context, private/shared leakage, no provenance, and corrections that do not change future behavior.

The repo proves the local deterministic trust-loop story, but it does not yet prove a real public product experience.

The biggest adoption blocker is that there is no root `README.md`, no `CHANGELOG.md`, and no landing/onboarding surface.

The positioning is strongest when it says: "Hermes remembers, shows why, and lets you fix memory once."

It becomes weaker whenever it says "shared memory kernel" or expands toward every agent runtime too early.

## Best Current Positioning

For solo builders operating Hermes across long-running projects, **Hermes Memory, powered by VibeGravity** gives agents scoped, explainable, correctable project memory across sessions, so users stop repeating context and can trust what the agent remembers.

## Clearest Answers To The Review Questions

### Who Is The Product For?

The first real user is the technical Hermes operator or builder.

The repo repeatedly names Hermes Agent as the first customer, but the human buyer/user is the person operating Hermes across ongoing projects. That user cares about continuity, correction, provenance, and scope separation more than they care about the internal memory graph.

The best first segment is a solo builder or highly technical indie hacker who uses Hermes as a daily project collaborator and can tolerate early local setup friction.

### What Painful Problem Does It Solve?

AI agents lose continuity across sessions and force users to repeat rules, preferences, current plans, project decisions, and corrections.

The more painful version of that problem is trust failure:

- raw chat logs are treated as memory;
- vector search returns noisy or stale context;
- private and shared memory blur together;
- users cannot inspect why memory exists;
- corrections do not reliably change future behavior;
- memory grows longer instead of more useful.

VibeGravity's strongest product problem is therefore not "agents need more context." It is "agents need memory that can be trusted, inspected, corrected, and scoped."

### Why Would A User Choose It Over Manually Prompting AI Coding Tools?

Manual prompting is labor. It depends on the user remembering what to paste, when to paste it, and how to avoid stale or private context.

VibeGravity offers a stronger promise:

- remember durable project context across sessions;
- retrieve compact recall packs instead of long copied prompt bundles;
- keep agent-private, workspace-shared, and group-shared memory separate;
- show source, scope, status, and freshness metadata;
- let a human correct wrong memory once;
- suppress superseded memory in later recall;
- keep degraded or stale recall visible instead of pretending it is fresh.

The product wins over manual prompting only if this trust loop is visible and reliable. A hidden backend that silently stores more context is not enough.

### What Is The Main Promise?

The main promise is:

> Hermes remembers the right project context across sessions, shows why it remembered it, and lets the operator fix memory once.

A shorter public version:

> Stop repeating context. Fix memory once. See why Hermes remembered it.

### Is The Promise Proven By The Repo?

Partially.

The repo has meaningful local proof:

- `go test ./...` passed during review.
- `make eval` passed during review.
- `cli eval demo` proves a local deterministic Hermes Memory trust-loop scenario.
- `Prefetch`/`recall_preview` expose trust metadata.
- correction, supersession, explain/timeline, stale/degraded metadata, and scope checks have test and review-packet coverage.

But the promise is not fully proven as a product:

- `make integration-postgres` skipped during review because `VIBEGRAVITY_DB_URL` was unset.
- the demo is local and deterministic, not a real Hermes runtime roundtrip;
- real Codex execution remains disabled by default;
- real Hermes runtime testing and packaging are still incomplete;
- there is no root public onboarding surface.

The repo proves product direction and local behavior. It does not yet prove public readiness.

### Where Does Messaging Feel Too Broad, Vague, Or Overclaimed?

Messaging is too broad when it leads with:

- "shared memory kernel";
- "agent memory engine" without a concrete user workflow;
- "universal memory platform";
- broad multi-agent runtime support;
- Dreaming;
- document memory;
- ontology-lite;
- a large v1 surface before the first trust loop is demonstrably polished.

The repo is strongest when it narrows to:

- Hermes-first;
- one operator;
- one painful workflow;
- recall preview;
- correction;
- explain/timeline;
- supersession;
- visible scope and degraded status.

The current docs already recognize this tension. The remaining issue is that the public-facing repo surface does not yet reflect the sharper story.

### Which User Type Should Be The First Target?

The first target should be **solo builder**.

That user is closest to the current repo reality: local-first, Hermes-first, technical, tolerant of CLI/MCP setup, and highly motivated by continuity across sessions.

Indie hackers are the next best target. Startup teams, agencies, and enterprise teams should come later because they require stronger packaging, onboarding, runtime proof, security posture, admin controls, and operational readiness.

## Target User Ranking

### 1. Solo Builder

Best first target.

The repo is local-first, Hermes-first, CLI/MCP-heavy, and aimed at one technical operator who can tolerate rough setup while caring deeply about continuity. A solo builder also feels the repetition pain directly: every new session loses the working context unless the user restates it.

This segment also matches the product's first proof: a single Hermes operator can see recall, correction, explain, and supersession in a controlled workflow.

### 2. Indie Hacker

Strong second target.

Indie hackers building agent workflows will understand the pain and accept early beta friction. They are likely to value memory continuity, private/shared scope, and correction because they run many experiments and switch contexts often.

The gap is packaging. Indie hackers still need a clean README, demo script, install path, and crisp explanation of what is simulated versus live.

### 3. Startup Team

Plausible later.

The workspace/shared/private memory model fits small teams working with multiple agents. Startup teams would care about shared project rules, current plans, and avoiding repeated onboarding of agents.

However, startup adoption needs stronger onboarding, live PostgreSQL proof, backup/restore, clearer operator UX, and a more stable external protocol story.

### 4. Agency

Real pain, weaker first target.

Agencies repeat context across clients and projects, so memory continuity is a real need. The private/shared/group scope model could become valuable for client boundaries and project teams.

But agencies create sharper trust requirements: tenant separation, auditability, explainability, rollback, and client-specific privacy guarantees. The repo is not yet packaged or proven enough for that adoption path.

### 5. Enterprise Team

Not first.

The repo lacks enterprise-facing packaging, admin controls, compliance language, observability, production deployment guidance, support model, and proven production ops.

Enterprise could eventually value scoped and auditable agent memory, but that is not the first wedge.

## Evidence From Repo

### Missing Public Product Surface

- `/Users/parker/Documents/VibeGravity/README.md`: not present.
- `/Users/parker/Documents/VibeGravity/README_VI.md`: not present.
- `/Users/parker/Documents/VibeGravity/CHANGELOG.md`: not present.

This is a product-readiness blocker. A repo without a root README cannot explain the promise, target user, quickstart, demo, limitations, or readiness level to a new visitor.

### Product Promise And Framing

- `/Users/parker/Documents/VibeGravity/PLANS.md:7`: V1 is framed as **Hermes Memory, powered by VibeGravity**.
- `/Users/parker/Documents/VibeGravity/PLANS.md:11`: the first release must prove one felt outcome.
- `/Users/parker/Documents/VibeGravity/PLANS.md:13`: Hermes remembers the right project context across sessions.
- `/Users/parker/Documents/VibeGravity/PLANS.md:14`: Hermes shows why it remembered it and lets the operator fix memory once.
- `/Users/parker/Documents/VibeGravity/PLANS.md:16`: VibeGravity remains the engine and internal architecture name.
- `/Users/parker/Documents/VibeGravity/PLANS.md:37`: documents and rich dreaming are no longer the v1 product headline.
- `/Users/parker/Documents/VibeGravity/PLANS.md:47`: V1 is not ready until the correction trust loop is proven against real PostgreSQL and external protocol paths.

### Problem And User

- `/Users/parker/Documents/VibeGravity/consulting/02_product_one_pager.md:5`: VibeGravity is described as a shared memory engine for AI agents.
- `/Users/parker/Documents/VibeGravity/consulting/02_product_one_pager.md:11`: AI agents lose continuity across sessions.
- `/Users/parker/Documents/VibeGravity/consulting/02_product_one_pager.md:13`: users repeat rules, preferences, decisions, and task state.
- `/Users/parker/Documents/VibeGravity/consulting/02_product_one_pager.md:17`: raw chat logs becoming memory is called a weak pattern.
- `/Users/parker/Documents/VibeGravity/consulting/02_product_one_pager.md:18`: noisy vector search is called a weak pattern.
- `/Users/parker/Documents/VibeGravity/consulting/02_product_one_pager.md:19`: weak private/shared separation is called a weak pattern.
- `/Users/parker/Documents/VibeGravity/consulting/02_product_one_pager.md:20`: corrections not changing future behavior is called a weak pattern.
- `/Users/parker/Documents/VibeGravity/consulting/02_product_one_pager.md:21`: lack of inspectability is called a weak pattern.
- `/Users/parker/Documents/VibeGravity/consulting/02_product_one_pager.md:24`: the target user section names Hermes Agent as the first customer.
- `/Users/parker/Documents/VibeGravity/consulting/02_product_one_pager.md:28`: the first human user is the person operating Hermes and expecting continuity.

### Differentiation

- `/Users/parker/Documents/VibeGravity/consulting/03_engine_positioning_and_narrative.md:5`: recommended category is agent memory engine.
- `/Users/parker/Documents/VibeGravity/consulting/03_engine_positioning_and_narrative.md:17`: VibeGravity sits behind Hermes and other agent hosts.
- `/Users/parker/Documents/VibeGravity/consulting/03_engine_positioning_and_narrative.md:34`: it decides what becomes memory, what is suppressed, what remains private, what is superseded, and what is recalled under budget.
- `/Users/parker/Documents/VibeGravity/consulting/03_engine_positioning_and_narrative.md:59`: the emotional promise is continuity without chaos.
- `/Users/parker/Documents/VibeGravity/consulting/03_engine_positioning_and_narrative.md:63`: Hermes-first is the wedge.
- `/Users/parker/Documents/VibeGravity/consulting/03_engine_positioning_and_narrative.md:73`: differentiation should be trust and behavior, not feature count.
- `/Users/parker/Documents/VibeGravity/consulting/03_engine_positioning_and_narrative.md:77`: explicit private/shared/group/session scopes are a strong differentiator.
- `/Users/parker/Documents/VibeGravity/consulting/03_engine_positioning_and_narrative.md:78`: correction as first-class product behavior is a strong differentiator.
- `/Users/parker/Documents/VibeGravity/consulting/03_engine_positioning_and_narrative.md:79`: explainable provenance is a strong differentiator.
- `/Users/parker/Documents/VibeGravity/consulting/03_engine_positioning_and_narrative.md:80`: budget-aware recall is a strong differentiator.

### Jobs To Be Done

- `/Users/parker/Documents/VibeGravity/consulting/04_customer_and_use_cases.md:16`: the first human user is the operator or builder who uses Hermes for ongoing work.
- `/Users/parker/Documents/VibeGravity/consulting/04_customer_and_use_cases.md:20`: this user wants to avoid repeating durable context.
- `/Users/parker/Documents/VibeGravity/consulting/04_customer_and_use_cases.md:22`: this user wants to correct wrong memory.
- `/Users/parker/Documents/VibeGravity/consulting/04_customer_and_use_cases.md:23`: this user wants to see why a memory exists.
- `/Users/parker/Documents/VibeGravity/consulting/04_customer_and_use_cases.md:24`: this user wants to avoid private/shared leakage.
- `/Users/parker/Documents/VibeGravity/consulting/04_customer_and_use_cases.md:29`: first job is continuing work without repeating context.
- `/Users/parker/Documents/VibeGravity/consulting/04_customer_and_use_cases.md:37`: another job is correcting memory when it is wrong.
- `/Users/parker/Documents/VibeGravity/consulting/04_customer_and_use_cases.md:41`: another job is inspecting why memory exists.

### MVP Scope

- `/Users/parker/Documents/VibeGravity/consulting/05_mvp_scope_and_non_goals.md:5`: V1 should prove Hermes is more continuous, safer, and more correctable across sessions.
- `/Users/parker/Documents/VibeGravity/consulting/05_mvp_scope_and_non_goals.md:7`: V1 does not need to prove VibeGravity is a universal memory platform.
- `/Users/parker/Documents/VibeGravity/consulting/05_mvp_scope_and_non_goals.md:19`: `prefetch()` returns typed recall blocks.
- `/Users/parker/Documents/VibeGravity/consulting/05_mvp_scope_and_non_goals.md:20`: recall is scope-aware.
- `/Users/parker/Documents/VibeGravity/consulting/05_mvp_scope_and_non_goals.md:21`: recall respects a token budget.
- `/Users/parker/Documents/VibeGravity/consulting/05_mvp_scope_and_non_goals.md:23`: superseded memory is suppressed.
- `/Users/parker/Documents/VibeGravity/consulting/05_mvp_scope_and_non_goals.md:28`: raw events and derived memories are separate.
- `/Users/parker/Documents/VibeGravity/consulting/05_mvp_scope_and_non_goals.md:30`: every memory has provenance.
- `/Users/parker/Documents/VibeGravity/consulting/05_mvp_scope_and_non_goals.md:32`: human correction can create a replacement memory and supersede the target.
- `/Users/parker/Documents/VibeGravity/consulting/05_mvp_scope_and_non_goals.md:71`: generic chat UI is not V1.
- `/Users/parker/Documents/VibeGravity/consulting/05_mvp_scope_and_non_goals.md:74`: every agent runtime integration is not V1.

### Runtime And Trust Contract

- `/Users/parker/Documents/VibeGravity/consulting/06_runtime_and_product_contract.md:5`: VibeGravity has two main lifecycle calls, `sync_turn()` and `prefetch()`.
- `/Users/parker/Documents/VibeGravity/consulting/06_runtime_and_product_contract.md:30`: `Prefetch` and `SyncTurn` are the engine heartbeat.
- `/Users/parker/Documents/VibeGravity/consulting/06_runtime_and_product_contract.md:31`: search, notes, plans, correction, timeline, and explain are trust and control surfaces.
- `/Users/parker/Documents/VibeGravity/consulting/06_runtime_and_product_contract.md:32`: documents are supporting context, not the main product.
- `/Users/parker/Documents/VibeGravity/consulting/06_runtime_and_product_contract.md:64`: raw events and derived memories stay separate.
- `/Users/parker/Documents/VibeGravity/consulting/06_runtime_and_product_contract.md:66`: every memory has provenance.
- `/Users/parker/Documents/VibeGravity/consulting/06_runtime_and_product_contract.md:67`: every memory has explicit scope.
- `/Users/parker/Documents/VibeGravity/consulting/06_runtime_and_product_contract.md:69`: human correction is first-class.
- `/Users/parker/Documents/VibeGravity/consulting/06_runtime_and_product_contract.md:83`: if the scope model is not trusted, the memory engine is not trusted.
- `/Users/parker/Documents/VibeGravity/consulting/06_runtime_and_product_contract.md:87`: Codex unavailability should pause new graph updates, not hide degraded state.

### Current State And Readiness

- `/Users/parker/Documents/VibeGravity/consulting/07_current_state_and_roadmap.md:5`: as of 2026-04-25, VibeGravity is beyond initial foundation but not V1-complete.
- `/Users/parker/Documents/VibeGravity/consulting/07_current_state_and_roadmap.md:32`: real Codex execution is not enabled by default.
- `/Users/parker/Documents/VibeGravity/consulting/07_current_state_and_roadmap.md:39`: custom Hermes provider registry packaging is blocked by Hermes CLI constraints.
- `/Users/parker/Documents/VibeGravity/consulting/07_current_state_and_roadmap.md:40`: full real Hermes runtime roundtrip tests are still needed.
- `/Users/parker/Documents/VibeGravity/consulting/07_current_state_and_roadmap.md:41`: production ops, install, backup, and restore flows are incomplete.
- `/Users/parker/Documents/VibeGravity/consulting/07_current_state_and_roadmap.md:57`: the suggested first milestone is a V1 trust slice.
- `/Users/parker/Documents/VibeGravity/consulting/07_current_state_and_roadmap.md:82`: a product demo should make value obvious in under 5 minutes.

### Product Risks Already Identified In Repo

- `/Users/parker/Documents/VibeGravity/consulting/08_risks_and_open_decisions.md:5`: product may sound like infrastructure, not value.
- `/Users/parker/Documents/VibeGravity/consulting/08_risks_and_open_decisions.md:13`: Hermes-first may be too narrow or hidden.
- `/Users/parker/Documents/VibeGravity/consulting/08_risks_and_open_decisions.md:19`: trust UX may be underbuilt.
- `/Users/parker/Documents/VibeGravity/consulting/08_risks_and_open_decisions.md:27`: scope separation is a product promise.
- `/Users/parker/Documents/VibeGravity/consulting/08_risks_and_open_decisions.md:35`: Codex dependency needs a clear story.
- `/Users/parker/Documents/VibeGravity/consulting/08_risks_and_open_decisions.md:43`: MVP may be too broad.

### Implementation And Test Evidence

- `/Users/parker/Documents/VibeGravity/Makefile:16`: `make eval` runs golden eval and demo eval.
- `/Users/parker/Documents/VibeGravity/Makefile:22`: `make integration-postgres` is opt-in and depends on `VIBEGRAVITY_DB_URL`.
- `/Users/parker/Documents/VibeGravity/internal/eval/demo.go:41`: `RunHermesMemoryDemo` executes the deterministic local 5-minute Hermes Memory demo.
- `/Users/parker/Documents/VibeGravity/internal/eval/demo.go:42`: the demo proves the trust loop without real Hermes, Codex, database, or network calls.
- `/Users/parker/Documents/VibeGravity/internal/eval/demo.go:156`: initial recall verifies rule, plan, memory, and trust metadata.
- `/Users/parker/Documents/VibeGravity/internal/eval/demo.go:172`: explain-memory verifies provenance.
- `/Users/parker/Documents/VibeGravity/internal/eval/demo.go:195`: correction writes supersession.
- `/Users/parker/Documents/VibeGravity/internal/eval/demo.go:254`: later recall uses correction and suppresses the old memory.
- `/Users/parker/Documents/VibeGravity/internal/eval/demo.go:268`: private scope separation is checked.
- `/Users/parker/Documents/VibeGravity/internal/hermes/provider.go:99`: Hermes provider advertises recall preview, search, add note, show plan, explain, correct, timeline, and degraded status.
- `/Users/parker/Documents/VibeGravity/internal/hermes/provider.go:149`: `show_plan` is still blocked until a read-only plan API exists.
- `/Users/parker/Documents/VibeGravity/internal/mcp/surface.go:43`: MCP tool surface lists the current externally callable tools.
- `/Users/parker/Documents/VibeGravity/internal/mcp/surface.go:67`: `recall_preview` maps to the same core `Prefetch` behavior as `prefetch`.
- `/Users/parker/Documents/VibeGravity/internal/kernel/correction_trust_loop_integration_test.go:34`: live PostgreSQL correction trust-loop test exists.
- `/Users/parker/Documents/VibeGravity/internal/kernel/correction_trust_loop_integration_test.go:35`: live test skips when `VIBEGRAVITY_DB_URL` is not set.
- `/Users/parker/Documents/VibeGravity/tests/README.md:17`: local deterministic gate does not prove real PostgreSQL locking, foreign keys, transaction rollback, or extension availability.
- `/Users/parker/Documents/VibeGravity/tests/README.md:21`: live PostgreSQL gate requires a scratch database.

## Verification Run During Review

The review ran the local proof gates that map directly to the product claim.

### `go test ./...`

Result: passed.

All Go packages passed or had no test files.

### `make eval`

Result: passed.

The golden eval passed scenarios for:

- pinned note and active plan priority;
- agent-private owner scoping;
- superseded memory suppression;
- degraded recall returning useful stored context;
- budget truncation;
- update memory replay;
- correction replay;
- group-shared graph write rejection;
- Stage 1 outage retry without graph writes;
- Stage 2 outage recovery idempotency;
- unsupported apply work blocking.

The Hermes Memory demo eval passed scenarios for:

- initial recall with project rule, active plan, and trust metadata;
- explain-memory provenance;
- correction supersession;
- next recall using correction;
- private scope separation.

### `make integration-postgres`

Result: skipped.

Reason:

```text
VIBEGRAVITY_DB_URL is not set.
```

Product implication: live PostgreSQL readiness cannot be claimed from this review run.

## Main Product Risks

### P0 - Blocks Adoption

#### No Public Onboarding Surface

The repository lacks root-level `README.md`, `README_VI.md`, and `CHANGELOG.md`.

Without a root README, a visitor cannot quickly understand:

- what VibeGravity is;
- who it is for;
- why it is better than manual prompting;
- how to run it;
- what is proven;
- what is still simulated or incomplete.

This blocks adoption even if the underlying architecture is strong.

#### End-To-End Promise Is Not Live-Proven

The local deterministic trust loop is strong, but the repo has not proven the complete live path:

```text
Hermes runtime -> MCP/provider -> VibeGravity service -> live PostgreSQL -> recall/correction/explain/timeline -> next Hermes recall
```

The live PostgreSQL gate skipped during review, and real Hermes roundtrip remains incomplete.

#### First Human Persona Is Not Sharp Enough In Public Language

"Hermes Agent" is a valid first customer internally, but a product needs a human target.

The first public persona should be:

> Solo builders and technical operators using Hermes for long-running projects.

### P1 - Hurts Trust

#### Messaging Still Drifts Toward Infrastructure

"Shared memory kernel" is accurate, but not emotionally or commercially sharp.

It should be reserved for architecture docs. Public language should lead with:

- stop repeating context;
- fix memory once;
- see why Hermes remembered it;
- scoped memory that does not leak private context.

#### Trust UX Is Not Yet Product-Shaped

The code has CLI/MCP/provider surfaces, but the product experience still needs a coherent demo or user flow:

1. preview what Hermes is about to remember;
2. inspect why it remembered it;
3. correct wrong memory;
4. see old memory suppressed;
5. see stale/degraded state honestly.

#### Codex Dependency Needs Plain-Language Handling

The repo is wise to keep real Codex disabled by default until degraded behavior is explicit. But a user needs to know:

- what works without Codex;
- what pauses when Codex is unavailable;
- how stale recall is labeled;
- how backlog recovery works;
- when real Codex is required.

### P2 - Can Wait

#### Naming Split Needs Discipline

Use:

- **Hermes Memory** for the first user-facing product;
- **VibeGravity** for the engine and architecture.

Do not lead public docs with the engine name unless the page is developer-facing.

#### Documents And Dreaming Should Stay Supporting Capabilities

Documents and Dreaming may matter architecturally, but they should not be part of the first product headline.

The first product value is the trust loop.

#### Future Integrations Should Not Be Over-Sold

Claude Code, Codex as a client, broader MCP clients, and other agent runtimes should be described as later expansion, not as part of the initial adoption promise.

## Recommended Changes

### README

Create `/Users/parker/Documents/VibeGravity/README.md`.

Recommended structure:

1. Hero headline:

   ```text
   Hermes Memory, powered by VibeGravity
   ```

2. Subheadline:

   ```text
   Stop repeating context. Fix memory once. See why Hermes remembered it.
   ```

3. One-paragraph product explanation:

   ```text
   Hermes Memory gives Hermes scoped, explainable, correctable project memory across sessions. VibeGravity is the local memory engine behind it: it records raw agent activity, derives structured memory, returns compact recall before the next turn, and keeps private, workspace, and group memory boundaries explicit.
   ```

4. Who it is for:

   ```text
   VibeGravity is currently for solo builders and technical operators using Hermes for long-running local projects.
   ```

5. Why not manual prompting:

   ```text
   Manual prompting makes you restate context. VibeGravity keeps durable context in a scoped memory system, shows where memory came from, lets you correct wrong memory once, and suppresses superseded facts in later recall.
   ```

6. Current readiness:

   ```text
   Current status: internal use. Local deterministic tests and demo evals pass. Live PostgreSQL and real Hermes runtime roundtrip are required before private beta.
   ```

7. Demo:

   ```bash
   go test ./...
   make eval
   ```

8. Live readiness gate:

   ```bash
   export VIBEGRAVITY_DB_URL='postgres://localhost:5432/vibegravity_integration?sslmode=disable'
   migrate -path migrations -database "$VIBEGRAVITY_DB_URL" up
   make integration-postgres
   ```

9. Non-goals:

   - not a chat UI;
   - not a raw transcript archive;
   - not a generic vector DB;
   - not every agent runtime integration in V1.

### Docs

Keep `plans/` and `consulting/` as internal source-of-truth planning docs.

Add a short product-facing docs entry, either:

- `/Users/parker/Documents/VibeGravity/docs/product.md`, or
- `/Users/parker/Documents/VibeGravity/docs/positioning.md`.

That doc should define:

- first user;
- first use case;
- product promise;
- proof path;
- readiness level;
- what remains unproven.

### Onboarding

Add a "5-minute trust-loop demo" guide.

Recommended file:

```text
/Users/parker/Documents/VibeGravity/docs/demo-hermes-memory-trust-loop.md
```

It should show:

1. project rule enters memory;
2. active plan appears in recall;
3. wrong memory is corrected;
4. old memory is suppressed;
5. explain/timeline shows provenance;
6. private/shared scope separation is visible;
7. degraded or stale recall is labeled.

Be explicit about which steps are local deterministic evals and which require live Postgres or real Hermes.

### Naming

Use this naming rule:

- Public first-customer product: **Hermes Memory**
- Engine: **VibeGravity**
- Category: scoped, correctable memory for Hermes and long-running agents
- Avoid as headline: shared memory kernel, universal memory platform, vector memory backend

Recommended tagline:

```text
Stop repeating context. Fix memory once. See why Hermes remembered it.
```

Recommended one-liner:

```text
Hermes Memory gives Hermes scoped, explainable, correctable project memory across sessions, powered by the local VibeGravity memory engine.
```

### Changelog

Create `/Users/parker/Documents/VibeGravity/CHANGELOG.md` before any external beta.

Minimum sections:

- Unreleased
- Added
- Changed
- Fixed
- Known Limitations

Known limitations should include:

- real Codex disabled by default;
- live PostgreSQL required for full trust-loop proof;
- real Hermes runtime roundtrip still required before beta;
- packaging/install flow incomplete.

## Final CPO Decision

**Internal use only.**

The product direction is strong enough for a focused private beta later, but the repo is not ready for public growth or early beta.

The product should not go public until three things are true:

1. A root README explains the promise, target user, quickstart, demo, and readiness level.
2. The live Postgres trust-loop gate passes against a scratch database.
3. A real Hermes runtime roundtrip proves recall preview, correction, explain/timeline, supersession, and degraded freshness through the same path a user will run.

Once those gates are true, the product can move from internal use to private beta for a small cohort of solo Hermes operators and indie builders.
