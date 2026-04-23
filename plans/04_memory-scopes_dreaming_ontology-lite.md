# Memory Scopes, Dreaming, and Ontology-Lite

## 1. Why This Document Exists

VibeGravity가 그냥 로그 저장소가 되지 않으려면 세 가지가 먼저 정해져야 한다.

- 누가 어떤 memory를 볼 수 있는가
- 어떤 기억이 오래 살아남는가
- memory를 어떤 구조로 저장하는가

## 2. Memory Scope Model

기본 scope는 네 가지다.

| scope | 설명 | 기본 접근자 |
|---|---|---|
| `agent_private` | 한 agent만 쓰는 기억 | owner agent, operator |
| `workspace_shared` | workspace 전체 공유 기억 | workspace members |
| `group_shared` | 지정된 agent group 공유 기억 | group members |
| `session_scratch` | 짧은 세션용 문맥 | current session only |

### 2.1 agent_private

특정 agent의 로컬 작업 습관, 내부 working notes, tool preference 같은 것이 들어간다.

### 2.2 workspace_shared

프로젝트 결정, 공용 규칙, 팀 선호, 공통 constraints가 들어간다.

### 2.3 group_shared

일부 agent끼리만 공유하는 전략, handoff pattern, 협업 힌트가 들어간다.

### 2.4 session_scratch

최근 context, temporary task state, 아직 장기 기억으로 승격되지 않은 재료다.

## 3. Access Rules

scope만 적고 끝내지 않는다.  
각 memory는 visibility rule을 가져야 한다.

필수 필드는 아래다.

- scope
- owner_entity_id
- workspace_id
- optional group_id
- read policy
- write policy
- correction policy

## 4. Memory Type Model

VibeGravity는 아래 typed memory를 기본으로 둔다.

| kind | 설명 |
|---|---|
| `fact` | 비교적 검증 가능한 사실 |
| `preference` | 좋아함, 싫어함, 선호 |
| `trait` | 오래 지속되는 성향 |
| `goal` | 이루려는 목표 |
| `constraint` | 제한 조건 |
| `relationship` | 사람, 팀, 프로젝트 관계 |
| `decision` | 이미 내린 결정 |
| `procedure` | 어떻게 일하는지에 대한 규칙 |
| `task_state` | 현재 작업 상태 |
| `doc_fact` | 문서에서 추출된 사실 |
| `summary` | 세션 또는 주제 요약 |
| `hypothesis` | 아직 확실하지 않은 추정 |

## 5. Ontology-Lite Shape

완전한 온톨로지는 피한다.  
하지만 아래 객체는 일급으로 둔다.

- entity
- memory
- edge
- profile
- plan
- note
- document
- document_chunk

### 5.1 Entity

예시:

- `user:parker`
- `agent:hermes-main`
- `workspace:vibegravity`
- `project:vibegravity-core`
- `group:reviewers`

### 5.2 Memory

모든 memory는 아래 최소 필드를 가진다.

```json
{
  "id": "mem_123",
  "kind": "preference",
  "scope": "workspace_shared",
  "text": "User prefers short Korean answers.",
  "entity_ids": ["user:parker"],
  "owner_entity_id": "workspace:vibegravity",
  "confidence": 0.94,
  "status": "active",
  "fingerprint": "fp_abc",
  "source_event_ids": ["evt_1"],
  "valid_from": "2026-04-23T00:00:00Z",
  "valid_to": null
}
```

### 5.3 Edge

기본 edge는 아래를 둔다.

| edge_kind | 의미 |
|---|---|
| `updates` | 이전 기억을 교체 |
| `extends` | 이전 기억을 보완 |
| `supports` | 다른 기억을 강화 |
| `contradicts` | 다른 기억을 반박 |
| `derived_from` | reasoning 결과의 근원 |
| `references_doc` | 문서 근거 연결 |
| `belongs_to` | entity or scope 귀속 |
| `corrected_by` | operator correction 연결 |

## 6. Dreaming Tiers

Dreaming은 대화를 재평가하는 maintenance layer다.  
기본 tier는 아래와 같다.

### 6.1 Short-term

최근 events와 scratch 상태다.  
빠르게 쌓이고 빠르게 사라진다.

예:

- recent tail
- temporary task hints
- unresolved references

### 6.2 Mid-term

세션과 최근 며칠 수준의 요약 계층이다.

예:

- session summary
- active plan hints
- recent topic summaries
- recent dynamic profile facts

### 6.3 Long-term

오래 유지할 가치가 있는 memory다.

예:

- stable preference
- important decision
- durable project rule
- reusable procedure

### 6.4 Ultra-long-term

정말 자주 바뀌지 않는 canonical layer다.

예:

- canonical profile facts
- deep long-term preferences
- core workspace rules
- durable shared procedures

## 7. Promotion Rules

모든 기억을 곧바로 장기로 올리지 않는다.

| from | to | 조건 |
|---|---|---|
| short-term | mid-term | session end or topic closure |
| mid-term | long-term | repeated, useful, stable |
| long-term | ultra-long-term | very stable, high-value, repeatedly confirmed |

## 8. Forget and Supersede

잊는다는 것은 무조건 delete가 아니다.  
기본 전략은 supersede와 archive다.

### 8.1 Supersede

새 사실이 예전 사실을 바꾸면 `updates`를 건다.  
옛 memory는 `superseded`가 된다.

### 8.2 Archive

현재 recall에 필요 없지만 provenance상 남겨야 하면 archived로 보낸다.

### 8.3 Hard delete

법적 또는 사용자 명시 요청이 있을 때만 별도 정책으로 다룬다.

## 9. Mixed Recall and Manual Control

자동 recall만 있으면 오작동을 잡기 어렵다.  
수동 control도 넣는다.

필수 manual 기능은 아래다.

- memory search
- note pin
- plan create and update
- memory correction
- explicit include by id
- provenance explain

## 10. Profiles

프로필은 snapshot이다.  
memory 전체와 같지 않다.

```json
{
  "entity_id": "user:parker",
  "static": [
    "짧고 직접적인 한국어 답변을 선호한다"
  ],
  "dynamic": [
    "VibeGravity를 Hermes-first shared memory kernel로 만들고 있다"
  ]
}
```

정적은 느리게 바뀐다.  
동적은 최근 활동을 더 강하게 반영한다.

## 11. Scope-Aware Recall Rules

Recall할 때는 다음 순서를 따른다.

1. system and session critical
2. matching scope
3. active plan and pinned notes
4. stable profile
5. dynamic profile
6. relevant memories
7. relevant documents
8. recent tail

scope mismatch memory는 candidate pool에서 먼저 제거한다.

## 12. Quality Rules

이 문서의 구조가 잘 지켜졌는지 보려면 아래를 본다.

- 같은 사실이 scope별로 뒤섞이지 않는가
- `updates`와 `extends`가 구분되는가
- short-term이 long-term을 오염시키지 않는가
- correction 뒤에 supersession이 생기는가
- profile.static과 dynamic이 섞이지 않는가
