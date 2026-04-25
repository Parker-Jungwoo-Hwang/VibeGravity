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
