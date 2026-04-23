# RFP: Build VibeGravity as a Hermes-First Shared Memory Kernel

## 1. Request Summary

VibeGravity를 구현할 AI agent를 찾는다.  
이 시스템은 Hermes와 다른 agent가 함께 쓰는 공용 기억 백엔드다.  
초기 main 고객은 Hermes Agent다.

이 프로젝트의 최우선 목표는 긴 대화 로그를 저장하는 것이 아니다.  
중복이 적고, 구조화되어 있고, 빠르게 다시 꺼낼 수 있는 shared memory kernel을 만드는 것이다.

## 2. Product Definition

VibeGravity는 아래를 수행해야 한다.

- turn 전 `prefetch()`로 compact recall pack 생성
- turn 후 `sync_turn()`로 raw event 기록
- worker에서 background memory derivation 수행
- graph memory, profile, notes, plans, documents 관리
- shared scope와 access rule에 맞는 recall 수행
- Hermes provider plugin으로 lifecycle 연결
- 나중에 다른 agent surface가 붙어도 같은 의미 유지

## 3. Mandatory Product Capabilities

### 3.1 Integration

필수 지원 범위는 아래와 같다.

| 범위 | 우선순위 | 설명 |
|---|---|---|
| Hermes Agent | P0 | 첫 번째 main integration |
| HTTP API | P0 | 코어 서버 표면 |
| MCP tools | P1 | operator와 coding agent 도구 표면 |
| Claude Code | P2 | 이후 surface. 직접 main은 아님 |
| Codex as client | P2 | 이후 surface. reasoning backend와 별도 |

### 3.2 Memory Scope

다음 memory scope를 반드시 구현한다.

| scope | 설명 |
|---|---|
| `agent_private` | 특정 agent만 접근 |
| `workspace_shared` | workspace 전체 공유 |
| `group_shared` | 지정된 agent group 공유 |
| `session_scratch` | 짧은 작업 문맥 |

### 3.3 Memory Types

다음 typed memory를 기본 지원한다.

| type | 설명 |
|---|---|
| episodic | 사건과 대화 기록에서 추출된 기억 |
| semantic | 사실, 선호, 규칙, 관계 |
| procedural | 어떻게 일하는지에 대한 규칙과 절차 |
| task | 현재 작업 상태 |
| plan | 구조화된 계획 |
| note | 사람이 직접 넣는 메모 |
| doc_fact | 문서에서 나온 사실 |
| summary | 세션 또는 주제 요약 |

### 3.4 Dreaming

Dreaming은 선택 기능이 아니다.  
maintenance layer로 구현해야 한다.

| tier | 역할 |
|---|---|
| short-term | 최근 raw tail, scratch, session buffer |
| mid-term | session summary, active topics, active tasks |
| long-term | stable memory, decisions, preferences, reusable facts |
| ultra-long-term | canonical profile, durable procedures, high-value decisions |

### 3.5 Recall

Recall은 혼합 모드여야 한다.

- 자동 주입: `prefetch()`
- 수동 주입: `search`, `pin`, `include`, `correct`, `explain`

### 3.6 Structure

Memory는 ontology-lite 구조여야 한다.  
각 memory는 최소한 아래를 가져야 한다.

- kind
- scope
- owner or visibility rule
- entity ids
- provenance
- timestamps
- confidence
- fingerprint
- status
- graph edges

### 3.7 Token Optimization

Recall pack은 budget aware여야 한다.  
중복을 줄여야 한다.  
반박된 오래된 memory는 숨겨야 한다.  
manual note와 active plan은 높은 우선순위를 가져야 한다.

## 4. Architecture Constraints

### 4.1 Runtime split

다음 runtime split을 지킨다.

- API hot path
- background worker
- PostgreSQL as canonical store
- local embedding runtime
- Codex reasoning runtime

### 4.2 Local vs Codex

local runtime은 embedding 중심으로만 사용한다.  
local extractor 의존 구조로 돌아가면 안 된다.  
텍스트 해석과 graph operation 생성은 Codex-first로 간다.

### 4.3 Core invariants

다음은 절대 깨지면 안 된다.

- raw event와 derived memory를 분리한다.
- 모든 write path는 idempotent해야 한다.
- 모든 memory는 provenance를 가져야 한다.
- 모든 reasoning 결과는 structured JSON이어야 한다.
- 모든 surface는 같은 core semantics를 공유해야 한다.

## 5. Required Deliverables

구현 agent는 아래를 납품해야 한다.

### 5.1 Code

- monorepo baseline
- Go HTTP server
- worker
- PostgreSQL schema and migrations
- core service layer
- Hermes provider plugin
- MCP surface
- Codex reasoner bridge
- local embedding provider
- CLI or doctor command

### 5.2 Documents

- architecture
- data model
- API contracts
- work log or ADRs
- operator guide
- test and eval guide
- integration guide

### 5.3 Quality assets

- unit tests
- integration tests
- e2e replay path
- golden evaluation scenarios
- seed data or fixtures
- doctor checks

## 6. Out of Scope for MVP

아래는 처음부터 하지 않는다.

- 범용 chat product
- 모든 connector 생태계
- 완전한 GUI
- 모든 agent runtime 동시 지원
- fully automatic forgetting without operator visibility
- heavyweight ontology platform
- multi-node distributed queue redesign

## 7. Expected Delivery Order

### Phase 1

foundation, schema, `sync_turn()`, `prefetch()` skeleton

### Phase 2

ingest idempotency, recall pack, local embeddings, basic search

### Phase 3

Codex reasoning chain, graph apply engine, profile, notes, plans

### Phase 4

dreaming tiers, memory promotion and suppression, Hermes plugin

### Phase 5

quality, ops, evals, replay harness, docs hardening

## 8. Acceptance Criteria

### Product acceptance

- Hermes가 `prefetch()`와 `sync_turn()` lifecycle에 붙는다.
- 같은 turn 재전송에도 duplicate memory가 폭증하지 않는다.
- workspace shared와 agent private memory가 섞이지 않는다.
- active plan과 pinned note가 recall에 들어간다.
- correction 뒤에 recall 결과가 달라진다.
- Codex 실패 시 degraded mode가 동작한다.

### Technical acceptance

- API hot path는 worker backlog와 분리되어 살아 있다.
- reasoning output은 schema validation과 semantic validation을 모두 통과한다.
- replay harness로 과거 세션을 재처리할 수 있다.
- golden scenarios에서 `updates`와 `extends`가 허용 범위 안에 있다.
- token budget 1000 / 2200 / 4000에서 recall이 graceful하게 잘린다.

## 9. Proposal Format for the Implementing Agent

응답은 아래 순서를 따라야 한다.

1. 이해한 제품 정의
2. 구현 범위 재정리
3. 위험 요소
4. 단계별 계획
5. 첫 번째 작업 묶음
6. 테스트 계획
7. 남는 질문 또는 가정

템플릿은 `templates/RFP_RESPONSE_TEMPLATE.md`를 쓴다.

## 10. Evaluation Rubric

구현 agent 평가는 아래 기준으로 한다.

| 항목 | 비중 | 설명 |
|---|---|---|
| 방향 일치 | 25 | 제품 정의를 흐리지 않았는가 |
| 구조 정확성 | 20 | core invariants를 지켰는가 |
| Hermes 우선순위 | 15 | Hermes-first delivery를 지켰는가 |
| recall 품질 | 15 | 짧고 유용한 context를 만드는가 |
| memory scope 안전성 | 10 | shared/private/group 경계를 지키는가 |
| test and eval | 10 | 검증 자산을 같이 만들었는가 |
| 문서 품질 | 5 | 다음 agent가 이어받기 쉬운가 |

## 11. Final Note

이 RFP의 핵심은 기능 수가 아니다.  
방향성 유지다.

VibeGravity는 shared memory kernel이다.  
Hermes-first다.  
Codex-first reasoning이다.  
local은 embedding 중심이다.  
scope와 provenance와 token optimization은 제품 중심부다.
