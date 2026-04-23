# ADR-004: Memory Semantic Retrieval Scope

## Status

Accepted

## Context

VibeGravity는 "문서 검색 + 대화 저장"이 아니라 Hermes-first shared memory kernel이다.
`prefetch()`는 session summary, profile, memory neighborhood, documents, notes, plans를 조합한다.
worker 파이프라인도 local embedding → neighborhood retrieval → Codex reasoning으로 진행한다.
따라서 memory 자체가 semantic retrieval 대상이어야 한다.

document chunk만 embedding을 가지면 recall이 RAG 쪽으로 기울고, memory graph는 검색보다 후처리 자료처럼 밀린다.

## Decision

`memories.embedding`은 v1 schema에 포함한다.

단, HNSW 같은 approximate index는 v1 bootstrap에 바로 넣지 않는다.
v1에서는 `memories.embedding` 컬럼과 exact vector search 경로를 먼저 둔다.
데이터가 쌓이고 recall benchmark가 생긴 뒤 후속 migration으로 HNSW index를 추가한다.

`document_chunks.embedding`과 `memories.embedding`은 같은 embedding model과 dimension 정책을 공유한다.
각 row에는 `embedding_model`, `embedding_dims`, `embedding_updated_at`도 함께 둔다.
embedding model 교체나 backfill을 추적하기 위해서다.

## Consequences

- memories 테이블에 vector 컬럼이 추가되어 row 크기가 커진다.
- memory 생성 시 worker가 embedding을 계산하여 저장해야 한다.
- `prefetch()`에서 query embedding으로 memory를 직접 검색할 수 있다.
- embedding model 교체 시 row 단위로 backfill 상태를 추적할 수 있다.

## Impact on Hermes-first Roadmap

memory semantic retrieval이 있어야 `prefetch()`가 진정한 의미의 "relevant memory"를 돌려줄 수 있다.
없으면 keyword match나 최근 시간순 정렬에만 의존하게 된다.
