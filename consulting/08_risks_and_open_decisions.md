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
