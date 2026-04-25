# Product Contract and Direction

## 1. Product North Star

VibeGravity는 agent가 한 일을 기억으로 바꾸는 shared memory kernel이다.

이 kernel은 세 가지를 잘해야 한다.

첫째, 빨리 기록한다.
둘째, 천천히 정리한다.
셋째, 다음 turn 전에 짧고 쓸모 있게 꺼낸다.

## 1.1 V1 Product Framing

V1의 첫 사용자-facing 제품 언어는 **Hermes Memory, powered by
VibeGravity**다.

VibeGravity는 엔진 이름이자 내부 architecture 이름으로 유지한다. 하지만 첫
고객에게는 "shared memory kernel"보다 아래 약속을 먼저 말한다.

> Hermes remembers the right project context across sessions, shows why it
> remembered it, and lets the operator fix memory once.

따라서 V1은 범용 memory platform이 아니라 Hermes operator가 체감하는 memory
trust loop를 닫는 제품이다.

Active contract: V1 is not done because documents can be stored, searched, or
fed into reasoning. Documents are supporting context. V1 is done only when the
Hermes Memory trust loop works through real storage and external protocol
surfaces: recall preview, explain/timeline, correction, correction-driven
supersession, and the next recall that suppresses the outdated memory.

## 2. Product Promise

사용자는 같은 말을 매번 다시 하지 않아도 된다.
Agent는 세션이 바뀌어도 중요한 규칙과 계획을 계속 이어갈 수 있다.
여러 agent가 한 workspace 안에서 같은 기억을 공유할 수 있다.
하지만 private memory는 private하게 남는다.

V1에서 이 약속은 네 가지 사용자 행동으로 검증한다.

- Hermes가 다음 세션에서 프로젝트 규칙, 현재 계획, 최근 결정을 다시 불러온다.
- operator가 memory의 source, scope, status를 볼 수 있다.
- operator가 잘못된 memory를 한 번 correction하면 이후 recall이 달라진다.
- Codex나 worker가 지연될 때 stale/degraded 상태가 숨겨지지 않는다.

## 3. Non-Negotiable Features

이 프로젝트는 아래 기능이 빠지면 제품 정체성이 무너진다.

### Hermes-first integration

Hermes가 첫 고객이다.
문서와 코드 구조 모두 Hermes lifecycle을 기준으로 잡는다.

### Memory scope separation

agent private memory와 workspace shared memory를 분리한다.
group shared memory를 따로 둔다.

### Dreaming

단기, 중기, 장기, 초장기를 나눈다.
대화를 다 저장하는 제품이 아니라, 가치 있는 것을 다시 정리하는 제품이다.

V1 headline은 dreaming이 아니다. Dreaming은 maintenance layer로 남기되,
첫 제품 감동은 recall, explain, correction, supersession에서 와야 한다.
Dreaming can improve memory quality over time, but it must not be used to mask
an unproven correction loop or weak protocol contract.

### Mixed recall

자동 recall과 수동 control이 같이 있어야 한다.

### Ontology-lite

memory는 typed object다.
kind, edge, provenance, status가 있어야 한다.

### Token optimization

Recall pack은 짧아야 한다.
반복, 장황함, 오래된 노이즈를 억제해야 한다.

## 4. What VibeGravity Is Not

VibeGravity는 다음 제품이 아니다.

- main chat product
- standalone coding agent
- only-vector memory DB
- only-chat-log archive
- raw transcript archive with nicer search
- giant knowledge graph platform
- pure note-taking app

The active product frame is narrower and stronger: Hermes Memory is the
operator-facing product, powered by VibeGravity as the agent memory engine.
VibeGravity earns its place by turning agent activity into scoped, explainable,
correctable memory, not by becoming a generic chat surface or generic vector
store.

## 5. User-Facing Mental Model

사람에게 설명할 때는 이렇게 말한다.

Hermes Memory는 Hermes가 프로젝트 맥락을 세션 사이에서도 기억하게 해준다.
VibeGravity는 그 뒤에서 raw agent activity를 scoped, correctable memory로
바꾸는 agent memory engine이다.

짧은 문구는 아래를 쓴다.

- Hermes Memory, powered by VibeGravity.
- Stop repeating context. Fix memory once. See why Hermes remembered it.
- Every memory has a scope, a source, and a correction path.

## 6. Internal Mental Model for Implementing Agents

구현 agent는 VibeGravity를 아래처럼 이해해야 한다.

- ingest kernel
- recall assembler
- memory graph engine
- profile and summary layer
- dreaming maintenance layer
- integration adapters

이 여섯 층이 같은 의미를 공유해야 한다.

## 7. Core Invariants

다음 불변 조건은 코드, 테스트, 문서에 모두 반영해야 한다.

### Invariant A. Raw and derived are separate

`raw_events`는 원본이다.
`memories`는 해석 결과다.
절대 섞지 않는다.

### Invariant B. Writes are idempotent

같은 turn을 여러 번 보내도 의미가 부풀지 않아야 한다.

### Invariant C. Memory always has provenance

어떤 event와 어떤 reasoning job에서 왔는지 끝까지 따라갈 수 있어야 한다.

### Invariant D. Scope is explicit

scope를 생략한 memory는 금지한다.

### Invariant E. Recall is budgeted

Recall은 무한히 길어질 수 없다.
항상 예산을 의식한다.

### Invariant F. Human correction is first-class

사람이 memory를 고치면 그 흔적이 남고 이후 reasoning에 반영되어야 한다.
The active V1 correction contract includes correction-driven supersession: the
operator correction remains append-safe evidence, and the corrected replacement
memory becomes the active/latest recall candidate through an `updates` edge and
mandatory provenance. Older record-only correction prep remains historical
material unless it is explicitly framed as the intake-only predecessor.

## 8. Product Priorities

우선순위는 아래 순서를 따른다.

1. 정확한 write path
2. 안정적인 scope separation
3. 유용한 recall
4. operator trust surface: recall preview, explain, timeline, correct
5. correction provenance and supersession proven against PostgreSQL
6. MCP/Hermes protocol schema correctness
7. evidence-safe replay idempotency
8. Hermes integration
9. graph and profile quality
10. degraded-mode truthfulness
11. dreaming quality
12. broader integrations

The next slice should favor DB/protocol correctness over new feature breadth:
P0 correction provenance, MCP schema correctness, evidence-safe replay
idempotency, live PostgreSQL integration gates, and stop-line protection.
Document memory should stay supporting context for recall and Stage 2 reasoning.
Dreaming should stay a maintenance and quality layer. Neither should displace
the P0 trust-loop proof: correction, provenance, supersession, scope separation,
explain, timeline, and degraded recall through real PostgreSQL and external
protocol paths.

## 9. Product Language

문서와 코드에서 용어를 통일한다.

| 용어 | 의미 |
|---|---|
| event | raw input |
| memory | structured derived object |
| scope | visibility boundary |
| profile | durable snapshot |
| summary | compressed narrative block |
| note | human-authored instruction or memo |
| plan | structured task object |
| recall pack | next-turn context bundle |
| dreaming | background consolidation and promotion |
| correction | human override or fix |
| Hermes Memory | first user-facing product wedge powered by VibeGravity |
| trust loop | recall preview, explain, correct, supersede, and next recall |

## 10. Future-Proofing Without Scope Creep

Claude Code와 Codex client integration은 염두에 둔다.
하지만 v1의 제품 의미는 Hermes-first로 닫는다.

즉 다음 원칙을 지킨다.

- interfaces are generic
- semantics are Hermes-proven first
- future adapters do not redefine core behavior

## 11. Decision Test

구현 중 선택지가 생기면 아래 질문을 한다.

이 선택이 `prefetch()`를 더 유용하게 만드는가.
이 선택이 `sync_turn()`를 더 안전하게 만드는가.
이 선택이 memory scope를 더 분명하게 만드는가.
이 선택이 사람이 correction할 길을 남기는가.
이 선택이 token 낭비를 줄이는가.
이 선택이 Hermes operator에게 memory를 보고, 믿고, 고칠 방법을 주는가.

여섯 질문 중 셋 이상에 아니오가 나오면 방향이 틀렸을 가능성이 높다.
