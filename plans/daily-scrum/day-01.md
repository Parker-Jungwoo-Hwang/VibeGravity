# VibeGravity Day 01 일일 업무계획서

## 오늘의 한 줄 목표

오늘의 목표는 **제품 정의와 구조 경계를 잠가서, Day 2부터 저장소 bootstrap과 인터페이스 설계가 문서만 보고 바로 착수 가능하게 만드는 것**이다.

오늘은 기능을 많이 만드는 날이 아니다.  
오늘은 아래를 고정하는 날이다.

- Hermes-first 제품 방향
- local embedding-only / Codex-first reasoning 분리
- `sync_turn()` hot path와 background worker의 경계
- memory scope separation
- artifact / contract / squad ownership

Day 1의 성패는 코드 양이 아니라 **PRD freeze, ADR 0, squad 책임 경계 확정**으로 판단한다.

---

## 오늘 이 계획의 근거로 읽은 문서

이 일일 계획서는 아래 문서를 기준으로 작성한다.

- `plans/README.md`
- `plans/00_read-this-first_for-building-agents.md`
- `plans/01_rfp_vibegravity_hermes-first.md`
- `plans/02_product-contract_and_direction.md`
- `plans/03_target-architecture_codex-first.md`
- `plans/04_memory-scopes_dreaming_ontology-lite.md`
- `plans/05_runtime-contracts_ingest-recall-apply.md`
- `plans/06_data-model_and_storage-invariants.md`
- `plans/07_workpack_foundation-and-repo-setup.md`
- `plans/08_workpack_ingest-and-recall.md`
- `plans/09_workpack_memory-graph-and-dreaming.md`
- `plans/10_workpack_hermes-provider-and-external-surfaces.md`
- `plans/11_workpack_quality-ops-and-evals.md`
- `plans/12_agent-coding_playbook_codex-claude.md`
- `plans/13_handoff-prompts_and_response-templates.md`
- `plans/templates/PLANS.md`
- `plans/templates/AGENTS.md`
- `plans/templates/CLAUDE.md`

---

## 오늘 절대 흔들리면 안 되는 기준

아래는 Day 1 내내 재검증해야 하는 고정 축이다.

1. **Hermes가 첫 고객이다.**
   - 제품 의미, API, provider, demo path 모두 Hermes lifecycle 기준으로 닫는다.
2. **local runtime은 v1에서 embedding / cheap retrieval helper 중심이다.**
   - local extractor 의존 main path 금지.
3. **텍스트 해석과 graph operation 생성은 Codex-first다.**
   - reasoning 결과는 structured JSON이어야 한다.
4. **raw event와 derived memory는 절대 섞지 않는다.**
5. **scope는 명시적이어야 한다.**
   - `agent_private`
   - `workspace_shared`
   - `group_shared`
   - `session_scratch`
6. **recall은 budget-aware여야 한다.**
   - typed block assembly 먼저, 최종 text render는 나중.
7. **human correction은 first-class다.**
   - correction path와 traceability를 처음부터 설계에 넣는다.

---

## 오늘의 비범위

오늘 아래 항목은 논의만 하더라도 결정으로 끌고 가지 않는다.

- Claude Code direct surface 구현
- Codex client direct surface 구현
- full UI
- graph DB 교체 검토
- distributed queue 재설계
- connector 확장
- local extractor 재도입
- 세부 성능 최적화
- Day 2 이후 스키마 상세 구현

오늘은 **방향과 경계만 잠그는 날**이다.

---

## 오늘 끝나면 반드시 남아 있어야 하는 산출물

오늘 산출물은 문서 기준으로 남겨야 한다.  
구두 합의로 끝내지 않는다.

### 필수 산출물 경로

- `/Users/parker/Documents/VibeGravity/plans/day-01/prd-freeze.md`
- `/Users/parker/Documents/VibeGravity/plans/day-01/adr-000-hermes-first-foundation.md`
- `/Users/parker/Documents/VibeGravity/plans/day-01/squad-boundaries.md`
- `/Users/parker/Documents/VibeGravity/plans/day-01/domain-artifact-classes-v0.md`
- `/Users/parker/Documents/VibeGravity/plans/day-01/runtime-boundaries-v0.md`
- `/Users/parker/Documents/VibeGravity/plans/day-01/recall-pack-minimum-v0.md`
- `/Users/parker/Documents/VibeGravity/plans/day-01/codex-reasoning-contract-v0.md`
- `/Users/parker/Documents/VibeGravity/plans/day-01/hermes-integration-scope-v0.md`
- `/Users/parker/Documents/VibeGravity/plans/day-01/day-02-input-checklist.md`

### 각 산출물이 담아야 하는 핵심

| 산출물 | 반드시 담아야 할 내용 |
|---|---|
| `prd-freeze.md` | 제품 한 줄 정의, P0/P1/P2 범위, non-goal, 성공 조건 |
| `adr-000-hermes-first-foundation.md` | 왜 Hermes-first인지, runtime split, local vs Codex 경계, 결과적 영향 |
| `squad-boundaries.md` | 스쿼드별 소유 영역, 입력/출력, 승인권자, 의존성 |
| `domain-artifact-classes-v0.md` | artifact class 목록, 설명, 생성 주체, canonical store 연결 |
| `runtime-boundaries-v0.md` | `sync_turn()` / worker / apply engine 책임 분리, failure contract |
| `recall-pack-minimum-v0.md` | typed block 종류, 우선순위, budget mode, degraded mode |
| `codex-reasoning-contract-v0.md` | stage 1/2 입력·출력, JSON schema 범위, retry/timeout 원칙 |
| `hermes-integration-scope-v0.md` | provider lifecycle, P0 tools, demo path, out-of-scope |
| `day-02-input-checklist.md` | Day 2 bootstrap에 필요한 고정값과 미결정 항목 목록 |

---

## 시간대별 운영 계획

| 시간 | 목적 | 진행 방식 | 산출물 / 체크포인트 |
|---|---|---|---|
| 09:00-09:20 | 일일 스탠드업 | 막힌 의존성만 확인 | 오늘 잠글 항목 5개 재확인 |
| 09:20-10:00 | CTO 정렬 | 기준 문서 재확인, 스쿼드 brief 배포 | 공통 제약사항 1페이지 배포 |
| 10:00-11:30 | 스쿼드 집중 1차 | 회의 금지, 각 스쿼드 초안 작성 | 각 스쿼드 v0 문서 초안 1개 이상 |
| 11:30-11:45 | 비동기 체크포인트 | 문서 링크와 blocker만 공유 | blocker 없는 스쿼드는 계속 진행 |
| 11:45-14:00 | 스쿼드 집중 2차 | 문서 구체화, 서로 필요한 결정 요청 | 각 문서에 open question 최대 2개 이하 |
| 14:00-15:00 | CTO contradiction scan | 용어 충돌, 책임 충돌, 범위 충돌 제거 | PRD freeze 초안 + ADR 0 초안 |
| 15:00-16:30 | 통합 리허설 | 문서 기준 워크스루 | 전체 lifecycle walkthrough 1회 통과 |
| 16:30-17:30 | 수정 루프 | 남은 충돌 제거, 문구 고정 | 산출물 freeze candidate 작성 |
| 17:30-18:00 | 승인 정리 | squad boundary sign-off | Day 2 입력 체크리스트 작성 |
| 18:00-18:30 | Day 1 게이트 | CTO gate | PRD freeze / ADR 0 / squad 경계 확정 |

---

## CTO 운영 포인트

CTO는 오늘 아래 세 가지 역할에 집중한다.

1. **스코프를 잠근다.**
   - “좋아 보이지만 Day 1이 아닌 것”을 다 잘라낸다.
2. **의존성 충돌을 없앤다.**
   - Squad A~E 문서가 서로 다른 세계관을 말하지 않게 한다.
3. **저녁 통합 게이트를 넘기지 못한 애매함을 다음 날로 넘기지 않는다.**
   - 미결정은 backlog가 아니라 ADR 후보로 남긴다.

CTO의 오늘 직접 소유 문서는 최소 아래 두 개다.

- `prd-freeze.md`
- `adr-000-hermes-first-foundation.md`

---

## 스쿼드별 상세 업무

## Squad A — Core / Data

**오늘의 목표:** artifact class 초안을 확정해서 Day 3 DTO와 Day 4 schema v0의 공통 모국어를 만든다.

**주요 입력 문서:**
- `02_product-contract_and_direction.md`
- `04_memory-scopes_dreaming_ontology-lite.md`
- `06_data-model_and_storage-invariants.md`

**오늘 문서 산출물:**
- `/Users/parker/Documents/VibeGravity/plans/day-01/domain-artifact-classes-v0.md`

**오늘 결정해야 할 것:**
- 무엇이 `memory`이고 무엇이 별도 artifact인지
- `raw_event`, `memory`, `memory_edge`, `memory_trace`, `profile`, `session_summary`, `note`, `plan`, `plan_item`, `document`, `document_chunk`, `entity`, `ingest_job`, `memory_group`를 어떤 위상으로 둘지
- 어떤 artifact가 immutable인지, derived인지, human-authored인지
- scope 적용 단위가 무엇인지

**세부 작업 순서:**
1. product language와 storage table을 한 표로 맞춘다.
2. artifact class 목록을 먼저 적고, 각 artifact의 목적을 한 줄로 정의한다.
3. 각 artifact마다 아래 필드를 채운다.
   - class name
   - 설명
   - 생성 주체
   - 변경 가능성
   - canonical store/table 후보
   - scope 적용 여부
   - provenance 필요 여부
4. `memory`와 `note` / `plan` / `document`를 섞지 않는 규칙을 명시한다.
5. Day 13 `artifact_class` 도입을 미리 의식하되, 오늘은 최소 집합만 잠근다.
6. Day 4 migration 설계에 바로 연결되는 용어를 고정한다.

**완료 기준:**
- 다른 스쿼드가 `memory 하나로 다 퉁치는` 표현을 쓰지 않게 된다.
- storage / runtime / recall 문서가 동일한 artifact vocabulary를 공유한다.
- Day 3 DTO와 Day 4 DB schema의 뼈대가 이 문서만 보고 그려진다.

**특별 주의:**
- ontology-heavy 방향 금지
- graph DB 전제 금지
- artifact를 너무 세분화해서 Day 2를 막는 것 금지

---

## Squad B — API / Worker

**오늘의 목표:** API hot path와 background worker의 경계를 명확히 해서, `sync_turn()`이 무엇을 하고 무엇을 하지 않는지 논쟁을 끝낸다.

**주요 입력 문서:**
- `03_target-architecture_codex-first.md`
- `05_runtime-contracts_ingest-recall-apply.md`
- `07_workpack_foundation-and-repo-setup.md`

**오늘 문서 산출물:**
- `/Users/parker/Documents/VibeGravity/plans/day-01/runtime-boundaries-v0.md`

**오늘 결정해야 할 것:**
- `sync_turn()` hot path의 6단계 책임
- worker의 핵심 job 책임
- apply engine이 worker 내부에서 어디서 시작되는지
- queue를 canonical DB와 어떻게 연결할지
- 실패 시 availability 우선 원칙

**세부 작업 순서:**
1. `sync_turn()` 책임을 아래 6개로 고정한다.
   - normalize
   - validate
   - compute idempotency
   - insert raw events
   - enqueue jobs
   - ack
2. hot path non-goal을 분명히 적는다.
   - deep reasoning 금지
   - graph updates 금지
   - profile recompute 금지
   - dreaming 금지
3. worker에서 담당할 범위를 표로 정리한다.
   - `process_turn_event`
   - `embed_document_chunks`
   - `dream_session`
   - `dream_workspace`
   - `rebuild_profile`
   - `maintenance`
4. failure contract를 적는다.
   - Codex unavailable
   - embedding unavailable
   - worker backlog high
   - worker crash
5. `sync_turn()`과 `prefetch()`는 worker backlog와 분리되어 살아 있어야 한다는 원칙을 문서에 못 박는다.
6. Day 2 repo bootstrap에 필요한 runtime shell 요구사항을 목록화한다.

**완료 기준:**
- Squad C와 D가 hot path에 deep reasoning을 넣지 않는다.
- Day 5 idempotent ingest와 Day 6 worker skeleton의 선행 계약이 끝난다.
- API route 설계 전에 boundary 문서가 우선 기준이 된다.

**특별 주의:**
- “간단한 reasoning 정도는 hot path에서 하자”라는 유혹 차단
- queue를 별도 대형 인프라 전제로 설계하지 말 것

---

## Squad C — Recall / Search

**오늘의 목표:** Day 12 `prefetch()` v1의 출구를 미리 규정하는 최소 recall pack 형태를 고정한다.

**주요 입력 문서:**
- `00_read-this-first_for-building-agents.md`
- `03_target-architecture_codex-first.md`
- `05_runtime-contracts_ingest-recall-apply.md`
- `08_workpack_ingest-and-recall.md`

**오늘 문서 산출물:**
- `/Users/parker/Documents/VibeGravity/plans/day-01/recall-pack-minimum-v0.md`

**오늘 결정해야 할 것:**
- typed block 최소 종류
- 우선순위 순서
- budget mode 기본값
- degraded mode 기본값
- empty recall을 피하는 규칙

**세부 작업 순서:**
1. recall pack을 문자열이 아니라 typed block 묶음으로 정의한다.
2. 최소 block 후보를 정리한다.
   - `pinned_note`
   - `active_plan`
   - `profile_static`
   - `profile_dynamic`
   - `memory_block`
   - `document_block`
   - `recent_block`
   - `meta`
3. block별 priority rough order를 문서화한다.
4. budget mode를 `small / default / rich`로 잡고, 1000 / 2200 / 4000 토큰 기준을 붙인다.
5. 아래 suppression 규칙을 최소 버전으로 적는다.
   - scope filter first
   - dedup before render
   - superseded suppression
   - plan / note uplift
   - degraded mode never empty if useful context exists
6. Hermes text render와 MCP JSON이 같은 의미를 공유하도록 typed block -> render 2단계 구조를 적는다.

**완료 기준:**
- Squad E가 Hermes provider context render를 이 문서 기준으로 잡을 수 있다.
- Squad D가 reasoning input의 retrieval neighborhood를 어떤 단위로 받을지 맞출 수 있다.
- Day 15 token packer가 오늘 문서의 연장선이 된다.

**특별 주의:**
- recall을 “최근 대화 붙이기”로 축소하지 말 것
- empty response를 정상 상태로 받아들이지 말 것
- raw log 장문 재주입 금지

---

## Squad D — Reasoning / Apply

**오늘의 목표:** Codex-first reasoning contract를 고정해, local extractor 회귀와 free-form LLM output을 구조적으로 막는다.

**주요 입력 문서:**
- `00_read-this-first_for-building-agents.md`
- `03_target-architecture_codex-first.md`
- `05_runtime-contracts_ingest-recall-apply.md`
- `09_workpack_memory-graph-and-dreaming.md`

**오늘 문서 산출물:**
- `/Users/parker/Documents/VibeGravity/plans/day-01/codex-reasoning-contract-v0.md`

**오늘 결정해야 할 것:**
- stage 1 extract 입력 / 출력
- stage 2 resolve 입력 / 출력
- structured JSON 범위
- apply 전 validation 계층
- profile / session summary / plan delta를 어느 단계에서 낼지

**세부 작업 순서:**
1. v1 reasoning chain을 2단계로 못 박는다.
   - Stage 1: extract
   - Stage 2: resolve
2. stage 1 출력 항목을 최소화해 적는다.
   - candidate entities
   - candidate memories
   - summary hint
   - task hint
3. stage 2 출력 항목을 표준화한다.
   - `operations`
   - `profile_delta`
   - `session_summary`
   - `plan_delta`
   - `trace`
4. JSON schema first 원칙을 적는다.
5. apply engine은 reasoning 결과를 그대로 믿지 않고 아래 순서를 탄다는 점을 적는다.
   - schema validation
   - semantic validation
   - entity ensure
   - fingerprint dedup
   - edge validity check
   - status / latest resolution
   - upsert / trace / commit
6. `updates`와 `extends`를 같은 뜻처럼 쓰지 못하게 의미 차이를 예시로 명시한다.

**완료 기준:**
- Day 16 Codex bridge, Day 17 apply engine, Day 18 profile/session summary 설계가 이 문서 위에서 바로 이어진다.
- structured JSON이 아니면 통과되지 않는다는 기준이 합의된다.
- Squad B와 hot path 분리, Squad C와 recall input shape가 충돌하지 않는다.

**특별 주의:**
- free-form reasoning output 허용 금지
- local LLM extractor fallback을 main path로 열어두지 말 것
- profile delta와 session summary의 소유 위치를 애매하게 두지 말 것

---

## Squad E — Hermes / MCP / Quality

**오늘의 목표:** Hermes-first 통합 범위를 고정해, v1에서 무엇을 붙이고 무엇을 미루는지 분명하게 만든다.

**주요 입력 문서:**
- `01_rfp_vibegravity_hermes-first.md`
- `03_target-architecture_codex-first.md`
- `10_workpack_hermes-provider-and-external-surfaces.md`
- `11_workpack_quality-ops-and-evals.md`
- `12_agent-coding_playbook_codex-claude.md`

**오늘 문서 산출물:**
- `/Users/parker/Documents/VibeGravity/plans/day-01/hermes-integration-scope-v0.md`
- `/Users/parker/Documents/VibeGravity/plans/day-01/day-02-input-checklist.md` 초안 협업

**오늘 결정해야 할 것:**
- Hermes provider가 Day 23에 반드시 제공할 lifecycle hook
- provider tools 최소 집합
- MCP와 provider의 의미 동등성
- doctor / health / demo path의 최소 범위
- built-in memory coexistence / failure isolation 기대치

**세부 작업 순서:**
1. Hermes v1 lifecycle을 아래로 고정한다.
   - pre-turn `prefetch()`
   - post-turn `sync_turn()`
   - `render_context()`
   - provider tools
   - optional session-end dreaming hint
2. hook별 역할을 표로 적는다.
   - `is_available()`
   - `prefetch()`
   - `sync_turn()`
   - `render_context()`
   - `get_tools()`
   - `on_session_end()`
3. provider tools 최소 세트를 적는다.
   - search memory
   - add note
   - show plan
   - correct memory
   - view timeline
4. MCP surface는 같은 core semantics를 호출해야 한다는 원칙을 명시한다.
5. Day 30 demo path를 문장 하나로 정의한다.
   - “Hermes가 prefetch로 context를 받고, 한 턴을 마친 뒤 sync_turn을 보내며, 다음 세션에서 같은 프로젝트를 계속 아는 것처럼 움직인다.”
6. plugin failure가 Hermes 전체를 죽이지 않아야 한다는 장애 격리 원칙을 적는다.

**완료 기준:**
- provider scope와 MCP scope가 충돌하지 않는다.
- Claude Code / Codex direct client를 Day 1 범위 밖으로 확실히 밀어낸다.
- Day 23~25 구현의 경계가 문서로 닫힌다.

**특별 주의:**
- Hermes adapter가 core semantics를 재정의하지 않게 할 것
- tool surface를 과하게 넓히지 말 것
- operator convenience 때문에 scope rule을 흐리지 말 것

---

## 스쿼드 간 의존성 맵

| 선행 | 후행 | 오늘 안에 맞춰야 하는 것 |
|---|---|---|
| Squad A | Squad B / C / D / E | artifact vocabulary |
| Squad B | Squad C / D / E | hot path vs worker 경계 |
| Squad C | Squad D / E | typed block / budget / render 전 단계 구조 |
| Squad D | Squad B / C / E | reasoning output schema / apply handoff |
| Squad E | 전체 | Hermes-first demo slice / provider scope |

**의존성 처리 원칙:**
- 30분 안에 안 풀리는 충돌은 CTO에게 올린다.
- 충돌이 제품 의미를 바꾸면 ADR 후보로 올린다.
- 충돌이 표현 차이면 `prd-freeze.md` 용어 표에서 통일한다.

---

## 11:30 비동기 체크포인트 양식

각 스쿼드는 아래 형식으로만 공유한다.

```md
## Squad [A-E]
- 초안 문서: [path]
- 오늘 잠근 문장 3개:
  - ...
  - ...
  - ...
- 남은 blocker:
  - ...
- 다른 스쿼드 결정 필요:
  - ...
```

회의로 길게 풀지 않는다.  
blocker만 올리고 바로 집중 시간으로 복귀한다.

---

## 15:00 통합 리허설 의제

문서만 놓고 아래 워크스루를 한다.

### 워크스루 1 — 기본 한 턴
1. Hermes가 turn 전에 `prefetch()` 호출
2. VibeGravity가 typed recall pack 반환
3. Hermes가 응답 생성
4. Hermes가 turn 후 `sync_turn()` 호출
5. hot path는 raw event 기록 + job enqueue + ack까지 수행
6. worker가 Codex reasoning과 apply를 뒤에서 수행
7. 다음 `prefetch()`에서 note / plan / profile / memory가 반영됨

### 워크스루 2 — 장애 / degrade
1. Codex unavailable일 때 무엇이 멈추고 무엇이 살아 있는가
2. embedding unavailable일 때 lexical fallback만으로 무엇이 가능한가
3. worker backlog가 높아도 `sync_turn()` / `prefetch()`가 살아 있는가
4. provider failure가 Hermes 전체를 죽이지 않는가

### 워크스루 3 — 의미 보존
1. raw와 derived가 섞이지 않는가
2. private / workspace / group scope가 문서상 명확한가
3. note / plan / memory / doc가 같은 객체처럼 뭉개지지 않는가
4. `updates`와 `extends`가 다른 동작을 가리키는가

리허설에서 막히는 항목은 바로 문서 수정으로 닫는다.  
“나중에 코드 짜면서 보자”는 금지다.

---

## 18:00 Day 1 게이트

아래를 모두 통과해야 Day 2로 넘어간다.

### Gate A — 방향 고정
- [ ] Hermes-first가 모든 문서에서 동일하게 표현된다.
- [ ] local은 embedding-only v1로 잠겼다.
- [ ] Codex-first reasoning이 hot path 바깥으로 분리되었다.

### Gate B — 핵심 계약 고정
- [ ] `sync_turn()` 책임과 non-goal이 고정되었다.
- [ ] `prefetch()`가 typed block / budget-aware 구조로 정의되었다.
- [ ] reasoning output이 structured JSON이라는 기준이 고정되었다.
- [ ] apply 전 validation 계층이 명시되었다.

### Gate C — 데이터 / 의미 경계 고정
- [ ] raw event와 derived memory가 분리되었다.
- [ ] note / plan / document / profile / memory가 구분되었다.
- [ ] scope 종류와 visibility rule이 명시되었다.
- [ ] provenance와 correction이 first-class로 남아 있다.

### Gate D — 조직 경계 고정
- [ ] squad 책임 경계가 문서화되었다.
- [ ] 승인권자와 의존성 경로가 명시되었다.
- [ ] Day 2 bootstrap 입력 체크리스트가 완성되었다.

### Gate E — 다음 날 인수인계 가능성
- [ ] Day 2가 repo / config / migration / interface skeleton에 바로 착수 가능하다.
- [ ] 오늘 결정이 chat이 아니라 파일에 남아 있다.
- [ ] open question은 0~3개 이내이며 owner가 붙어 있다.

한 항목이라도 실패하면 Day 1은 끝난 것이 아니다.

---

## 오늘 닫아야 하는 핵심 질문

아래 질문은 오늘 안에 문서로 닫아야 한다.

1. `memory`와 `note` / `plan` / `document`는 어떤 관계인가?
2. Day 3 interface freeze 전에 artifact vocabulary는 무엇으로 통일할 것인가?
3. `sync_turn()`은 어디까지 하고 어디서 worker에 넘기는가?
4. `prefetch()`의 최소 block 종류와 우선순위는 무엇인가?
5. reasoning 2-stage 체인은 어떤 입력과 어떤 출력 계약을 갖는가?
6. `updates`와 `extends`는 어떤 상황에서 다르게 쓰이는가?
7. Hermes provider가 v1에서 반드시 제공할 hook과 tool은 무엇인가?
8. Day 2 bootstrap에 필요한 package layout / config / migration / queue / tooling 결정은 무엇이 남아 있는가?

---

## 리스크와 완화책

| 리스크 | 오늘 나타나는 형태 | 완화책 |
|---|---|---|
| 설계 과잉 | ontology나 surface를 너무 크게 잡음 | artifact minimum set만 고정 |
| hot path 오염 | reasoning 일부를 API에 넣자는 주장 | Squad B 문서에서 non-goal 못 박기 |
| 용어 붕괴 | memory가 모든 것을 뜻하는 단어가 됨 | Squad A vocabulary 표를 기준 문서로 지정 |
| 통합 표면 과확장 | Hermes 외 surface를 같은 우선순위로 논의 | Hermes-first / others later를 PRD에 명시 |
| 문서 충돌 | 스쿼드별 문서가 다른 세계관을 말함 | 15:00 contradiction scan 강제 |
| 구두 합의 증발 | 회의에서만 결정되고 파일에 안 남음 | 18:00 gate에서 파일 존재 자체를 확인 |

---

## Day 2에 넘길 인수인계 기준

Day 2 팀은 아래 질문에 즉답할 수 있어야 한다.

- repo shape는 무엇인가?
- core service entrypoint는 무엇인가?
- 어떤 API부터 skeleton을 만들면 되는가?
- 어떤 artifact와 테이블이 핵심인가?
- queue와 worker baseline은 어떤 철학으로 가는가?
- 어떤 instruction files를 repo root에 둬야 하는가?
- 어떤 테스트 / health / doctor baseline이 필요한가?

즉, Day 1의 목표는 “좋은 생각을 많이 하는 것”이 아니라 **Day 2가 헤매지 않게 만드는 것**이다.

---

## 오늘의 최종 성공 정의

오늘 성공은 아래 한 문장으로 판단한다.

**VibeGravity의 제품 정의, runtime 경계, artifact vocabulary, reasoning contract, Hermes integration 범위가 문서로 잠겨서, Day 2부터 구현 스쿼드가 방향 논쟁 없이 곧바로 bootstrap 작업에 들어갈 수 있다.**
