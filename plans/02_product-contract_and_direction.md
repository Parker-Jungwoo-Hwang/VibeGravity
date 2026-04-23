# Product Contract and Direction

## 1. Product North Star

VibeGravity는 agent가 한 일을 기억으로 바꾸는 shared memory kernel이다.

이 kernel은 세 가지를 잘해야 한다.

첫째, 빨리 기록한다.  
둘째, 천천히 정리한다.  
셋째, 다음 turn 전에 짧고 쓸모 있게 꺼낸다.

## 2. Product Promise

사용자는 같은 말을 매번 다시 하지 않아도 된다.  
Agent는 세션이 바뀌어도 중요한 규칙과 계획을 계속 이어갈 수 있다.  
여러 agent가 한 workspace 안에서 같은 기억을 공유할 수 있다.  
하지만 private memory는 private하게 남는다.

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
- giant knowledge graph platform
- pure note-taking app

## 5. User-Facing Mental Model

사람에게 설명할 때는 이렇게 말한다.

VibeGravity는 AI를 위한 공용 기억장치다.  
AI가 답하기 전에 필요한 기억을 짧게 불러오고, 대화 뒤에는 새 기억을 정리해 둔다.

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

## 8. Product Priorities

우선순위는 아래 순서를 따른다.

1. 정확한 write path
2. 안정적인 scope separation
3. 유용한 recall
4. graph and profile quality
5. Hermes integration
6. dreaming quality
7. operator experience
8. broader integrations

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

다섯 질문 중 셋 이상에 아니오가 나오면 방향이 틀렸을 가능성이 높다.
