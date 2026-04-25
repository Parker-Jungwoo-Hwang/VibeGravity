# CorrectMemory Review and GetTimeline Prep

Date: 2026-04-24
Purpose: parallel-safe review and next-slice planning while another agent implements narrow `CorrectMemory` intake.
Status: completed as prep material; `CorrectMemory` and the first read-only
`GetTimeline` slice have both landed. Keep this packet as review evidence and
historical prompt material, not as the current next task.

Historical note: this packet describes the narrow record-only correction intake
slice from 2026-04-24. The active V1 contract now includes correction-driven
supersession: the append-safe correction artifact remains, and the correction
text is applied as a replacement memory with mandatory trace, an `updates` edge,
and prior-memory supersession. Any checklist item below that says not to
supersede or mutate `latest_flag` is scoped to the old intake-only slice.

## Parallel-Safe Boundary

This packet should not require edits to Go implementation files while the
`CorrectMemory` agent is working. Use it as a review checklist and as the
ready-to-send prompt for the next `GetTimeline` slice after the correction
intake diff lands.

Do not start `GetTimeline` implementation until the `CorrectMemory` diff has
been reviewed, because the timeline shape depends on the final correction
artifact/store contract.

## CorrectMemory Review Checklist

### Scope

- `kernel.Service.CorrectMemory` no longer returns `core.ErrNotImplemented`.
- The implementation only records correction intent.
- It does not implement `update_memory`.
- It does not archive, supersede, or mutate `latest_flag`.
- It does not introduce real Codex calls, Hermes provider behavior, or MCP tools.
- It preserves raw events, derived memories, and correction artifacts as separate records.

### Validation

- Rejects nil request.
- Requires `tenant_id`.
- Requires `workspace_id`.
- Requires `memory_id`.
- Requires `operator_id`.
- Requires `idempotency_key`.
- Requires non-empty `correction_text`.
- Returns `core.ErrNotFound` or equivalent service error when the target memory does not exist.
- Confirms the target memory belongs to the requested tenant/workspace before accepting correction.

### Persistence

- Writes a raw correction event through the raw event store.
- Raw correction event uses stable idempotency semantics.
- Raw correction event payload includes at least target memory ID, operator ID, correction text, and optional evidence JSON.
- Stores an operator-visible correction artifact, preferably append-safe such as `memory_corrections`.
- Does not overwrite the original target memory's `memory_trace`.
- If the implementation touches `memory_trace`, it must not destroy original reasoning provenance.
- If a new table is added, migration is tenant/workspace scoped and has a unique idempotency guard.

### Idempotency

- Retrying the same correction request with the same idempotency key does not create duplicate correction records.
- The response remains stable enough for clients to treat retries as accepted.
- Duplicate correction requests do not enqueue background graph work unless a later explicit reprocess contract is added.

### Tests

- Validation failure test covers at least one required field.
- Missing target memory test returns not found.
- Success test proves raw correction event is written.
- Success test proves correction artifact is written.
- Duplicate idempotency test proves correction state does not grow.
- Tests prove `update_memory`, archive, supersession, and `latest_flag` behavior are not opened.

### Docs

- If a correction table/store contract is added, update `plans/06_data-model_and_storage-invariants.md`.
- If public behavior changes, update `plans/05_runtime-contracts_ingest-recall-apply.md`.
- Update `PLANS.md` and this review packet only after the slice is complete.

### Red Flags

- Existing `memory_trace` is replaced with correction-only provenance.
- `memories.status`, `valid_to`, or `latest_flag` changes in this slice.
- `memory_edges` gains an `updates` edge in this slice.
- `CorrectMemory` calls the reasoning bridge.
- The correction write is not idempotent.
- The correction artifact cannot be surfaced by a later operator/timeline path.

### Review Commands

Run these before accepting the `CorrectMemory` diff:

```bash
go test ./...
make lint
make check-headers
git diff --check
git diff --stat
git status --short --branch
```

## After CorrectMemory Is Accepted

The next implementation slice should be `GetTimeline`. The reason to wait is
simple: timeline should expose the correction artifact shape that actually
landed, not an imagined one.

Minimum timeline behavior should be read-only. It should assemble existing
artifacts without creating graph mutations, Codex calls, or background jobs.

## GetTimeline Slice Shape

### Product Goal

Make `/v1/timeline` a truthful operator view over existing memory activity.
For v1, timeline is an inspection surface, not a mutation or dreaming surface.

### Recommended Initial Sources

- `memories` for derived memory items.
- `memory_trace` for provenance timestamps and raw event linkage.
- correction artifact table/store from the `CorrectMemory` slice, if added.
- raw correction events when they are the only correction artifact available.

Do not add notes, plans, documents, profiles, or session summaries to the first
timeline slice unless the implementation stays small and fully tested.

### Request Handling

The current HTTP handler only forwards `tenant_id`, `workspace_id`, and
`entity_id`. The `GetTimeline` slice should parse and validate:

- `tenant_id`
- `workspace_id`
- `entity_id`
- `scopes`
- `from`
- `to`
- `limit`

Default limit should be bounded. Reject invalid time ranges and invalid limits.

### Scope Rules

- `workspace_shared` can be returned within the same tenant/workspace.
- `session_scratch` can be returned within the same tenant/workspace.
- `agent_private` requires `entity_id` and must match `owner_entity_id`.
- `group_shared` should stay excluded until membership-aware filtering exists.
- Do not leak private memories through timeline just because it is an operator endpoint.

### Store Contract

Prefer a dedicated store method such as:

```go
GetTimeline(ctx context.Context, req *core.GetTimelineRequest) (*core.GetTimelineResponse, error)
```

Keep the query tenant/workspace scoped. Keep ordering deterministic:

1. newest `occurred_at` or provenance timestamp first
2. stable ID tie-breaker

### Output Rules

- Use `core.TimelineItem`.
- Set `ArtifactClassTimeline` for correction events.
- Keep memory-derived items typed with the existing memory kind/artifact class.
- Include `memory_id` for memory/correction-linked items.
- Include `raw_event_id` when a source raw event is available.
- Do not render giant raw payloads into `Text`; use short operator-readable text.

## Ready-To-Send GetTimeline Prompt

```md
You are continuing VibeGravity in `/Users/parker/Documents/VibeGravity`.

Read first:
- `AGENTS.md`
- `PLANS.md`
- `plans/00_read-this-first_for-building-agents.md`
- `plans/01_rfp_vibegravity_hermes-first.md`
- `plans/02_product-contract_and_direction.md`
- `plans/03_target-architecture_codex-first.md`
- `plans/05_runtime-contracts_ingest-recall-apply.md`
- `plans/06_data-model_and_storage-invariants.md`
- `docs/review-packets/current-state-and-next-agent-handoff.md`
- `docs/review-packets/correctmemory-review-and-gettimeline-prep.md`

Task:
Implement the first read-only `GetTimeline` slice.

Context:
- `CorrectMemory` narrow intake should already be implemented and reviewed.
- `/v1/timeline` currently delegates to `kernel.Service.GetTimeline`, but the service behavior is not implemented.
- Timeline is an operator inspection surface, not a graph mutation path.

Implement only this scope:
- Parse and validate `tenant_id`, `workspace_id`, `entity_id`, `scopes`, `from`, `to`, and `limit` in the HTTP handler.
- Implement service-level `GetTimeline`.
- Add a store-level timeline read path over existing memories/traces and the correction artifact that landed in the `CorrectMemory` slice.
- Preserve scope separation: `agent_private` requires owner/entity match; exclude `group_shared` until membership filtering exists.
- Return deterministic newest-first `core.TimelineItem` rows.
- Include correction events/artifacts in timeline when available.
- Add focused tests for query parsing, validation, scope filtering, correction visibility, and deterministic ordering.
- Update docs only if public behavior or store contract changes.

Do not do:
- Do not implement `update_memory`.
- Do not archive, supersede, or mutate `latest_flag`.
- Do not implement real Codex calls.
- Do not implement Hermes provider or MCP tools.
- Do not create dreaming/profile/session-summary behavior.
- Do not weaken source provenance, code header, or scope-separation rules.
- Do not revert unrelated dirty worktree changes.

Verification:
- `gofmt` on touched Go files
- `go test ./...`
- `make lint`
- `make check-headers`
- `git diff --check`

Return:
- Files changed
- Tests/checks run
- Remaining risks
- Whether docs were updated
- Source Review:
  - Estimated source
  - Suspected license
  - Similarity risk
  - Review required
```

## Source Review

- Estimated source: first-principles VibeGravity plans and current repo contracts.
- Suspected license: none.
- Similarity risk: low.
- Review required: yes, because follow-on implementation will touch correction and timeline semantics.
