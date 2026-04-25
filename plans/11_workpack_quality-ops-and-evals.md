# Work Pack 05: Quality, Ops, and Evals

## 1. Goal

이 work pack의 목표는 시스템을 믿을 수 있게 만드는 것이다.
기억 시스템은 조용히 틀릴 수 있다.
그래서 observability와 replay와 golden eval이 필수다.

V1 quality bar는 **Hermes Memory trust loop**를 기준으로 잡는다. 좋은 eval은
엔진 내부 shape만 확인하지 않고, Hermes operator가 다음을 믿을 수 있는지
검증해야 한다: recall preview, visible scope, explainable provenance,
correction, supersession, degraded status.

The immediate quality priority is DB/protocol correctness, not new feature
breadth. Documents are supporting recall/reasoning context, and dreaming is a
maintenance layer; neither can substitute for a proven correction trust loop
against real PostgreSQL and the Hermes-facing MCP/protocol path.

## 2. Deliverables

- structured logs
- core metrics
- trace and provenance export
- doctor command
- replay harness
- golden scenarios
- latency checks
- backup and restore notes
- security notes for Codex auth and private data

## 3. Golden Scenario Set

최소 아래 시나리오는 있어야 한다.

- user preference accumulation
- user correction updates old fact
- workspace shared vs agent private separation
- group shared memory visible only to members
- active plan influences next recall
- pinned note overrides noisy memory
- session dreaming promotes useful fact
- superseded memory is suppressed
- manual correction changes profile
- recall preview shows the scope, source, status, and freshness of memory blocks
- operator correction appears in the next relevant recall and suppresses the old
  memory
- explain/timeline can show why Hermes remembered a recalled memory
- Codex outage keeps `prefetch()` useful from previous profile, notes, plans,
  session summaries, and existing memories
- stale worker/Codex freshness state marks recall metadata degraded and downgrades
  derived block freshness without changing notes/plans/document source labels
- Codex recovery drains worker backlog without duplicate derived memories,
  missing `memory_trace`, or cross-scope leakage

## 4. Metrics

핵심 메트릭은 아래다.

- `api.sync_turn.latency_ms`
- `api.prefetch.latency_ms`
- `jobs.backlog.count`
- `jobs.backlog.oldest_age_seconds`
- `jobs.backlog.drain_rate_per_minute`
- `jobs.backlog.recovery_eta_seconds`
- `reasoning.codex.fail.count`
- `reasoning.codex.outage.duration_seconds`
- `recall.degraded_mode.count`
- `recall.degraded_mode.useful_block_count`
- `recall.degraded_mode.freshness_lag_seconds`
- `recall.pack.tokens`
- `memory.upsert.count`
- `memory.duplicate.rate`
- `profile.coherence.score`
- `updates_vs_extends.error_rate`
- `memory.correction.propagation_success_rate`
- `memory.superseded.recall_leak.count`
- `recall.preview.scope_label_coverage`
- `memory.explain.coverage`
- `operator.restatement.reduction_rate`
- `operator.correction.usage_count`

## 4.1 Codex Outage UX Gate

Codex failure is acceptable only if the user experience degrades visibly but
predictably. The system must prove the following before real Codex execution is
enabled by default.

- `prefetch()` still returns non-empty recall when any useful stored context
  exists.
- Recall metadata marks degraded mode and reports which sources were available.
- Manual notes, pinned notes, active plans, and previous profile blocks remain
  higher priority than stale memory blocks.
- The response stays within the requested token budget.
- Freshness loss is observable as `recall.degraded_mode.freshness_lag_seconds`,
  not hidden as a successful fresh graph update.
- Agent-private, workspace-shared, and group-shared boundaries remain identical
  to normal mode.
- Operator-facing surfaces must say that stored memory is being used while new
  graph/profile updates are delayed.

Minimum acceptance: during a simulated Codex outage, Hermes can continue at
least a planning or continuation workflow using prior profile, active plan,
pinned notes, session summaries, and existing memories. New semantic graph
updates may be delayed, but the user must not receive an empty recall pack when
stored context exists.

## 4.2 Worker Backlog Recovery Gate

Recovery is acceptable only if the backlog drains measurably and replay remains
idempotent.

- Measure queued, running, failed, and blocked job counts separately.
- Track the oldest queued job age and drain rate after Codex recovers.
- Estimate recovery ETA from backlog count and observed completed jobs per
  minute.
- Replay the same raw event bundle after recovery and verify no duplicate
  memory, edge, or trace rows are created.
- Verify blocked deterministic unsupported work does not enter an automatic
  retry loop.
- Verify transient Codex failures retry, then complete after recovery without
  bypassing apply validation.

Minimum acceptance: after a fixed simulated outage window, the worker must drain
the eligible backlog to zero under the configured worker concurrency, while
leaving unsupported deterministic jobs blocked and preserving mandatory
`memory_trace` for every applied memory.

## 5. Replay Harness

과거 session을 다시 흘려 보낼 수 있어야 한다.
prompt 변경, schema 변경, embedding 모델 변경 뒤에 비교할 수 있어야 한다.

The replay harness must support a Codex outage profile:

- fail Stage 1, Stage 2, or both for a bounded time window
- continue accepting `sync_turn()` writes during the outage
- run `prefetch()` during the outage and score degraded recall usefulness
- restore Codex and measure backlog drain time
- compare memory, edge, trace, profile, and recall outputs before and after
  replay

The V1 demo replay must cover one complete trust loop:

1. Hermes stores a project rule and active plan through `sync_turn()`.
2. Next-session `prefetch()` returns those blocks with visible scope/source.
3. Operator explains one recalled memory.
4. Operator corrects a wrong memory.
5. The next relevant recall includes the replacement and suppresses the old
   memory.
6. An `agent_private` memory does not appear in `workspace_shared` recall.

Current narrow eval gate:

- `tests/golden/replay_eval.json` contains deterministic golden recall scenarios.
- Recall golden scenarios can assert block-level trust metadata, including
  scope, source, source id, status, freshness, and owner, so recall preview
  regressions are caught before Hermes-facing surfaces render misleading memory.
- `tests/golden/replay_eval.json` also contains narrow graph replay scenarios
  for `update_memory`, correction-shaped supersession, deterministic retry, and
  the current `group_shared` write stop-line.
- `tests/golden/replay_eval.json` contains mocked worker backlog scenarios for
  deterministic Stage 1 outage, deterministic Stage 2 outage, recovery replay
  idempotency, and unsupported apply work blocking.
- `internal/eval` runs recall scenarios against in-memory stores and the real
  recall assembler.
- `internal/eval` runs graph replay scenarios through the real
  `graph.StoreBackedApplyEngine`, then checks state shape and later recall
  against an in-memory store that enforces the current update boundary.
- `internal/eval` runs worker backlog scenarios through the real
  `worker.Processor`, mocked Stage 1/Stage 2 runners, and the real
  `graph.StoreBackedApplyEngine`, then checks that transient reasoning failure
  goes through retry, failed reasoning writes no graph state, recovery/replay
  does not duplicate memory/trace/edge rows, and deterministic unsupported
  apply work lands in blocked state.
- `cli jobs metrics [--window D] [--tenant ID] [--workspace ID]` exposes the
  first operator-visible backlog metrics: total queued and ready queued counts,
  other status counts, retryable queued attempts, oldest ready queued age,
  completed jobs in the drain window, drain rate, and recovery ETA when
  calculable.
- `cli eval golden --path tests/golden/replay_eval.json` prints pass/fail, observed block kinds, sources, and token estimates.
- `make eval` is the local quality gate for pinned notes, active plans, private
  scope separation, superseded suppression, degraded profile/summary recall,
  budget behavior, replay idempotency, mandatory trace/edge shape,
  membership-blocked `group_shared` graph writes, mocked Codex outage retry,
  worker backlog recovery, and blocked unsupported work.

This is not the full session replay harness yet. It is the first regression gate
that keeps quality-sensitive recall behavior visible while reasoning, apply,
Hermes packaging, and real replay evolve. It now simulates narrow mocked Stage 1
and Stage 2 outage/recovery behavior, but it does not yet measure full Codex
outage windows, backlog drain rates, recovery ETA, production replay sessions,
real Codex auth/client behavior, or real Hermes runtime roundtrips.

## 5.1 Next Quality Slice

Prioritize these gates before broadening V1 features:

- Product contract alignment: test and docs should treat Hermes Memory,
  powered by VibeGravity, as the active frame. VibeGravity remains the agent
  memory engine, not a generic chat app, raw transcript archive, or vector DB.
- P0 correction provenance: correction artifact, replacement memory, trace,
  `updates` edge, prior supersession, explain/timeline visibility, and next
  recall suppression.
- MCP schema correctness: `tools/list` and `tools/call` preserve the service
  contract fields for recall preview, correction, timeline, and explain memory.
- Evidence-safe replay idempotency: retry/replay must not duplicate memories,
  traces, edges, or correction artifacts, and stale/degraded state must be
  visible.
- Live PostgreSQL integration: correction supersession and update lineage need
  real database gates, especially transaction and uniqueness behavior.
- Stop-line protection: keep real Codex disabled by default, keep custom Hermes
  provider packaging out of scope until protocol roundtrips are proven, and do
  not promote documents or dreaming into the V1 product promise.

## 6. Security Notes

- `CODEX_HOME` is sensitive
- auth cache must stay on trusted host
- local model logs should be minimal
- export endpoints must respect scope and permissions

## 7. Done When

- one memory can be traced end-to-end
- broken job can be replayed
- golden regressions are visible
- release gate exists
- operator can understand what changed and why
- the 5-minute Hermes Memory demo can show continuity, explain, correction,
  supersession, and private/shared scope separation
