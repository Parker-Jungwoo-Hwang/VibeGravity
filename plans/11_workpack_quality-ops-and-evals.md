# Work Pack 05: Quality, Ops, and Evals

## 1. Goal

이 work pack의 목표는 시스템을 믿을 수 있게 만드는 것이다.  
기억 시스템은 조용히 틀릴 수 있다.  
그래서 observability와 replay와 golden eval이 필수다.

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

## 4. Metrics

핵심 메트릭은 아래다.

- `api.sync_turn.latency_ms`
- `api.prefetch.latency_ms`
- `jobs.backlog.count`
- `reasoning.codex.fail.count`
- `recall.pack.tokens`
- `memory.upsert.count`
- `memory.duplicate.rate`
- `profile.coherence.score`
- `updates_vs_extends.error_rate`

## 5. Replay Harness

과거 session을 다시 흘려 보낼 수 있어야 한다.  
prompt 변경, schema 변경, embedding 모델 변경 뒤에 비교할 수 있어야 한다.

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
