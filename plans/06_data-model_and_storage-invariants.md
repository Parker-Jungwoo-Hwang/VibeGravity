# Data Model and Storage Invariants

## 1. Storage Philosophy

PostgreSQL이 canonical store다.
shared, concurrent, replayable memory system이기 때문이다.

SQLite는 테스트와 lightweight local dev용이다.
의미 규칙의 기준 저장소는 PostgreSQL이다.
Hermes Memory V1 cannot be declared ready from SQLite or in-memory behavior
alone. The correction trust loop, `updates` lineage, replay idempotency, and
external protocol paths must have live PostgreSQL gates before new feature
breadth becomes the priority.

## 2. Core Tables

v1 핵심 테이블은 아래다.

| table | role |
|---|---|
| `raw_events` | immutable ingest log |
| `ingest_jobs` | worker queue |
| `entities` | users, agents, workspaces, projects, groups |
| `memories` | deduped structured memory |
| `memory_edges` | memory relationships |
| `memory_trace` | provenance and reasoning traces |
| `memory_corrections` | append-safe human correction intents |
| `profiles` | static and dynamic snapshots |
| `session_summaries` | per-session summaries |
| `notes` | human-authored notes |
| `plans` | structured plans |
| `plan_items` | plan tasks |
| `documents` | static document units |
| `document_chunks` | searchable document units |
| `memory_groups` | shared group definitions |
| `memory_group_memberships` | which agents belong to which groups |

## 3. `raw_events`

이 테이블은 절대 원본 의미를 잃지 않는다.

필수 필드:

- tenant_id
- workspace_id
- session_id
- actor_id
- event_kind
- source
- idempotency_key
- fingerprint
- occurred_at
- payload_json
- created_at

## 4. `memories`

`memories`는 raw event와 분리된 derived object다.
필수 필드는 아래다.

- id
- tenant_id
- workspace_id
- scope
- group_id nullable
- owner_entity_id
- kind
- text
- fingerprint
- confidence
- status
- valid_from
- valid_to nullable
- latest_flag
- metadata_json
- created_at
- updated_at

## 5. `memory_edges`

기본 필드:

- from_memory_id
- to_memory_id
- edge_kind
- confidence
- created_by_job_id
- created_at

`updates`와 `extends`는 v1에서 가장 중요한 edge다.

## 6. `memory_trace`

trace는 선택이 아니다.
필수다.

필수 필드:

- memory_id
- raw_event_ids
- reasoning_job_id
- reasoning_stage
- candidate_snapshot_json
- applied_operations_json
- operator_correction_flag
- related_document_ids
- created_at

## 7. `memory_corrections`

`memory_corrections`는 사람이 특정 memory에 대해 남긴 정정 의도를 보존한다.
이 테이블 자체는 `memory_trace`를 덮어쓰지 않고, `latest_flag`를 직접 바꾸지
않으며 operator-visible artifact를 append-safe하게 남긴다. Correction workflow
전체는 별도 storage transaction에서 replacement memory, mandatory trace,
`updates` edge, prior target supersession, and correction artifact status
transition to `applied`을 수행할 수 있다. 같은 idempotency key가 다른 memory,
operator, correction text, or evidence에 재사용되면 replay가 아니라 conflict다.

필수 필드:

- id
- tenant_id
- workspace_id
- memory_id
- operator_id
- raw_event_id
- idempotency_key
- correction_text
- evidence_json
- status
- created_at

## 8. `profiles`

profile은 snapshot이다.
materialized facts가 아니라 current view다.

필드 예시:

- tenant_id
- workspace_id
- entity_id
- scope
- static_json
- dynamic_json
- source_memory_ids
- updated_at
- version

## 9. `memory_groups`

group shared memory를 위해 group model을 둔다.

필수 필드:

- id
- tenant_id
- workspace_id
- name
- description
- created_at

멤버십은 별도 테이블로 둔다.

## 10. `notes` and `plans`

note와 plan은 자동 memory와 구분한다.
인간의 의도와 구조가 더 강하기 때문이다.

note 기본 필드:

- note_kind
- scope
- owner_entity_id
- text
- pinned
- expires_at

plan 기본 필드:

- title
- status
- scope
- owner_entity_id
- evidence_json
- created_at
- updated_at

## 11. `documents` and `document_chunks`

문서와 기억은 저장 층부터 분리한다.

Document memory is supporting context for recall, search, and Stage 2 reasoning.
It is not the V1 product promise. V1 readiness still depends on correction,
provenance, supersession, scope separation, explain/timeline, degraded recall,
and protocol correctness against PostgreSQL.

documents:

- document-level dedup
- metadata
- version hints

document_chunks:

- retrieval unit
- embedding
- heading/path metadata
- neighbor info

## 12. Storage Invariants

### Invariant A. Every memory has scope

scope null 금지

### Invariant B. Every memory has provenance

trace 없는 memory 금지
단, explicit human note는 note trace로 대체 가능

V1 operator-facing surfaces must expose enough provenance for trust: source
event or artifact, scope, status, correction state, and whether a memory has
been superseded.

### Invariant C. `updates` can only target one latest memory at a time

한 memory lineage에서 동시에 둘 이상의 latest truth가 기본값이 되면 안 된다.
`updates` edge는 새 memory에서 이전 memory로 향하므로 direct target guard는
`memory_edges(to_memory_id) WHERE edge_kind = 'updates'`에 둔다.
전체 lineage latest 보장은 `update_memory` transaction에서 target latest lock,
new memory/trace/edge write, prior memory supersede를 함께 commit해야 한다.
현재 write path는 target을 active/latest로 lock하고, 새 memory와 mandatory
trace, 새 memory -> prior memory 방향의 `updates` edge, prior memory의
`superseded/latest_flag=false/valid_to` 변경을 하나의 transaction으로 commit한다.
같은 deterministic job/operation retry는 이미 완성된 memory, trace, edge가
확인될 때 idempotent success로 처리한다.

### Invariant C1. correction supersession is append-safe and provenance-backed

`correct_memory`는 raw correction event와 `memory_corrections` row를 쓴다.
그 다음 correction text를 새 replacement memory로 쓰고, mandatory
`memory_trace(operator_correction_flag = true)`, `updates` edge, prior target
supersession, and `memory_corrections.status = 'applied'`을 하나의 storage
transaction으로 commit한다. 기존 target `memory_trace`는 덮어쓰지 않으며,
correction artifact는 operator-visible append-safe 기록으로 남는다. New
correction 대상은 active/latest memory여야 한다. Exact same-key replay는 이미
recorded/applied artifact와 completed replacement memory, trace, and edge를
검증한 뒤 idempotent success로 처리한다.
This is the active contract. Record-only correction documents describe an older
intake-only slice unless they are explicitly updated to this supersession
contract.

Live PostgreSQL gates should prove at least:

- the replacement memory, trace, edge, and target supersession commit together;
- retrying the same correction idempotency key does not create duplicate graph
  rows or correction artifacts;
- explain/timeline can still show both the correction artifact and the preserved
  original memory provenance;
- stale or superseded memory does not leak into normal recall as current truth.

### Invariant C2. timeline is read-only

`get_timeline`은 `memories`, `memory_trace`, `memory_corrections`를 읽어서
operator-visible view를 만든다. 이 경로는 graph, trace, correction, profile,
summary, job state를 쓰지 않는다. `agent_private` rows는 `entity_id`가
`owner_entity_id`와 일치할 때만 반환하며, `group_shared`는 membership-aware
filtering이 들어오기 전까지 반환하지 않는다.

### Invariant D. group shared memory requires valid membership

group_id만 있고 membership이 없으면 invalid state다.

### Invariant D1. agent private retrieval requires owner scope

`agent_private` memories, notes, and plans can only be returned to recall or
Stage 2 source assembly when the request carries the visible actor as
`owner_entity_id`, and the row owner matches that actor. `workspace_shared` and
`session_scratch` rows do not use this owner gate. `group_shared` memories can
only be returned when the visible actor has a `memory_group_memberships` row for
the memory `group_id`; group-shared notes and plans remain excluded until those
tables carry a group identifier.

### Invariant E. profile is rebuildable

profile은 raw + memories + edges에서 재생성 가능해야 한다.
Profile lookup is tenant/workspace scoped: `(tenant_id, workspace_id, entity_id,
scope)` is the identity boundary for stored snapshots. Session summary recall is
also tenant/workspace/session scoped; `session_id` alone is never enough in a
shared PostgreSQL store.

## 13. Suggested Indexes

최소 권장 index는 아래다.

- `raw_events(tenant_id, workspace_id, session_id, created_at desc)`
- `raw_events(tenant_id, source, idempotency_key unique)`
- `memories(tenant_id, workspace_id, scope, status)`
- `memories(fingerprint)`
- `memory_edges(from_memory_id, edge_kind)`
- `memory_edges(to_memory_id, edge_kind)`
- `memory_corrections(tenant_id, workspace_id, idempotency_key unique)`
- `memory_corrections(memory_id, created_at desc)`
- `profiles(tenant_id, workspace_id, entity_id, updated_at desc)`
- `session_summaries(tenant_id, workspace_id, session_id, updated_at desc)`
- `notes(workspace_id, pinned, expires_at)`
- `plans(workspace_id, status)`
- `document_chunks(document_id, chunk_index)`

## 14. Migration Rules

vector 차원 변경, edge 종류 변경, scope 구조 변경은 전부 ADR 대상이다.
큰 테이블 backfill은 온라인 작업으로 분리한다.
profile과 summary는 rebuild 가능해야 하므로 destructive migration 전후 재생성 path를 남긴다.

## 15. Storage Review Questions

- private and shared rows가 섞이지 않는가
- group membership 없이 group shared가 생기지 않는가
- raw event 없이 memory가 생기지 않는가
- profile이 특정 custom cache에만 갇히지 않는가
- correction과 supersession trace를 끝까지 따라갈 수 있는가
