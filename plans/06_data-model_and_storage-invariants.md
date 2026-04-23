# Data Model and Storage Invariants

## 1. Storage Philosophy

PostgreSQL이 canonical store다.  
shared, concurrent, replayable memory system이기 때문이다.

SQLite는 테스트와 lightweight local dev용이다.  
의미 규칙의 기준 저장소는 PostgreSQL이다.

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

## 7. `profiles`

profile은 snapshot이다.  
materialized facts가 아니라 current view다.

필드 예시:

- entity_id
- scope
- static_json
- dynamic_json
- source_memory_ids
- updated_at
- version

## 8. `memory_groups`

group shared memory를 위해 group model을 둔다.

필수 필드:

- id
- tenant_id
- workspace_id
- name
- description
- created_at

멤버십은 별도 테이블로 둔다.

## 9. `notes` and `plans`

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

## 10. `documents` and `document_chunks`

문서와 기억은 저장 층부터 분리한다.

documents:

- document-level dedup
- metadata
- version hints

document_chunks:

- retrieval unit
- embedding
- heading/path metadata
- neighbor info

## 11. Storage Invariants

### Invariant A. Every memory has scope

scope null 금지

### Invariant B. Every memory has provenance

trace 없는 memory 금지  
단, explicit human note는 note trace로 대체 가능

### Invariant C. `updates` can only target one latest memory at a time

한 memory lineage에서 동시에 둘 이상의 latest truth가 기본값이 되면 안 된다.

### Invariant D. group shared memory requires valid membership

group_id만 있고 membership이 없으면 invalid state다.

### Invariant E. profile is rebuildable

profile은 raw + memories + edges에서 재생성 가능해야 한다.

## 12. Suggested Indexes

최소 권장 index는 아래다.

- `raw_events(tenant_id, workspace_id, session_id, created_at desc)`
- `raw_events(tenant_id, source, idempotency_key unique)`
- `memories(tenant_id, workspace_id, scope, status)`
- `memories(fingerprint)`
- `memory_edges(from_memory_id, edge_kind)`
- `memory_edges(to_memory_id, edge_kind)`
- `profiles(entity_id, updated_at desc)`
- `notes(workspace_id, pinned, expires_at)`
- `plans(workspace_id, status)`
- `document_chunks(document_id, chunk_index)`

## 13. Migration Rules

vector 차원 변경, edge 종류 변경, scope 구조 변경은 전부 ADR 대상이다.  
큰 테이블 backfill은 온라인 작업으로 분리한다.  
profile과 summary는 rebuild 가능해야 하므로 destructive migration 전후 재생성 path를 남긴다.

## 14. Storage Review Questions

- private and shared rows가 섞이지 않는가
- group membership 없이 group shared가 생기지 않는가
- raw event 없이 memory가 생기지 않는가
- profile이 특정 custom cache에만 갇히지 않는가
- correction과 supersession trace를 끝까지 따라갈 수 있는가
