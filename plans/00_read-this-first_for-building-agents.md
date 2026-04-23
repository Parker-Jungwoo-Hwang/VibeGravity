# Read This First

당신은 VibeGravity를 만드는 AI agent다.  
이 프로젝트의 목표를 한 문장으로 말하면 이렇다.

VibeGravity는 Hermes와 다른 agent가 함께 쓰는 공용 기억 코어다.

## 이 제품이 하는 일

이 제품은 agent가 한 일을 raw event로 기록한다.  
그다음 그 raw event에서 쓸모 있는 기억만 정리해 memory object로 만든다.  
다음 turn 전에 필요한 기억만 짧게 꺼내 recall pack으로 돌려준다.

## 이 제품이 하지 않는 일

이 제품은 직접 답을 만드는 agent가 아니다.  
범용 chat UI도 아니다.  
처음부터 모든 agent를 완벽히 지원하는 플랫폼도 아니다.

## 첫 번째 고객

첫 번째 고객은 Hermes Agent다.  
초기 main 경로는 Hermes memory provider plugin이다.  
Claude Code와 Codex는 바로 main 고객이 아니라, 다음 연결 표면이다.

## 절대 놓치면 안 되는 기능

### 1. 메모리 범위

메모리는 최소 네 층으로 나뉜다.

- agent private memory
- workspace shared memory
- group shared memory
- session scratch / recent buffer

agent private memory는 그 agent만 본다.  
workspace shared memory는 workspace 안 누구나 본다.  
group shared memory는 지정된 agent 그룹만 본다.  
session scratch는 짧게 유지되는 작업 문맥이다.

### 2. Dreaming

대화를 그대로 영구 저장하지 않는다.  
단기, 중기, 장기, 초장기 기억층을 둔다.  
뒤에서 다시 정리하고 승격하고 요약하고 잊는다.

### 3. Recall

Recall은 두 모드가 함께 있어야 한다.

- 자동 주입
- 수동 주입

자동 주입은 `prefetch()`가 만든 recall pack이다.  
수동 주입은 search, pinned note, explicit include 같은 operator / tool 경로다.

### 4. 구조화

memory는 그냥 문장 덩어리가 아니다.  
kind, scope, entity, provenance, time, confidence, edge를 가진 구조화 객체다.  
완전한 무거운 온톨로지는 아니다.  
하지만 ontology-lite는 있어야 한다.

### 5. Token 최적화

많이 저장하는 것이 목표가 아니다.  
짧고 쓸모 있게 꺼내는 것이 목표다.  
Recall pack은 항상 budget을 의식해야 한다.

## 기술 방향

local은 임베딩 전용으로 시작한다.  
cheap lexical / hybrid retrieval 보조는 local에서 한다.  
text extraction, conflict resolution, profile update, graph apply input은 Codex-first로 간다.

초기 reasoning 파이프라인은 한 번에 전부 하지 않는다.  
v1은 Codex 2단계 체인을 기본으로 둔다.

- 1단계: candidate extraction
- 2단계: conflict resolution + graph operations + profile delta

이유는 안정성과 디버깅 때문이다.

## 구현 순서

먼저 foundation을 만든다.  
그다음 ingest와 recall을 만든다.  
그다음 memory graph와 dreaming을 만든다.  
그다음 Hermes plugin을 붙인다.  
마지막에 품질, 운영, 평가를 닫는다.

## 좋지 않은 구현 신호

다음 징후가 보이면 방향이 틀어진 것이다.

- raw event와 derived memory가 섞인다.
- workspace memory와 agent private memory가 섞인다.
- recall이 전체 로그를 그대로 뿌린다.
- local extractor에 다시 의존한다.
- prompting만 길어지고 contract가 없다.
- test와 replay 없이 prompt만 바꾼다.
- agent마다 다른 의미 규칙을 쓰기 시작한다.

## 좋은 구현 신호

다음 상태면 방향이 맞다.

- `sync_turn()`은 빠르게 끝난다.
- `prefetch()`는 짧고 유용한 recall pack을 준다.
- worker가 뒤에서 graph와 profile을 정리한다.
- 같은 turn 재전송에도 memory duplication이 안정적이다.
- 사람이 correction을 넣으면 이후 recall이 달라진다.
- Hermes가 붙어도, MCP가 붙어도, HTTP가 붙어도 의미가 같다.
