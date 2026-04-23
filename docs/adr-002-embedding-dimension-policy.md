# ADR-002: Embedding Dimension Policy

## Status

Accepted (dimension value pending model selection)

## Context

VibeGravity v1부터 pgvector를 도입하여 semantic retrieval을 기준 구조로 잡는다.
BYTEA로 시작하고 나중에 vector로 전환하면 migration 비용이 더 크다.
그러나 embedding 차원은 모델에 따라 다르다 (예: 768, 1024, 1536, 3072).
v1의 embedding 모델이 아직 최종 확정되지 않았을 수 있다.

## Decision

1. **pgvector v1 즉시 도입**: `CREATE EXTENSION IF NOT EXISTS vector`는 별도 migration으로 분리한다.
2. **차원 하드코딩 금지**: `vector(768)` 같은 숫자를 migration에 직접 쓰지 않는다.
3. **모델과 차원을 함께 저장**: 설정에 `embedding_model`과 `embedding_dims`를 두고, migration에서는 이 값을 기반으로 컬럼을 생성한다.
4. **Migration 분리 순서**:
   - Migration A: `CREATE EXTENSION IF NOT EXISTS vector`
   - Migration B: vector column 생성 (exact search로 시작)
   - Migration C (후속): HNSW approximate index 추가 (데이터 축적 후)
5. **v1 기본 모델 후보**: 아래 중 하나를 ADR-002b에서 최종 확정한다.
   - `nomic-embed-text` (768d, 로컬 실행 가능)
   - `text-embedding-3-small` (1536d, OpenAI API)
   - `mxbai-embed-large` (1024d, 로컬 실행 가능)
6. 모델 확정 전까지 config에 `embedding_model: "pending"`, `embedding_dims: 0`을 두고, 첫 부트 시 doctor command가 경고를 출력한다.

## Consequences

- pgvector 의존성이 v1부터 들어간다. PostgreSQL에 pgvector extension이 필수다.
- 차원 변경 시 migration으로 컬럼을 재생성해야 한다. 이는 `06_data-model` §13의 "vector 차원 변경은 ADR 대상"과 일치한다.
- exact search로 시작하므로 초기 성능은 데이터 규모에 비례한다. HNSW는 10만 건 이상 축적 후 도입한다.

## Impact on Hermes-first Roadmap

Hermes의 `prefetch()` 경로에서 semantic retrieval이 가능해진다.
embedding model이 미확정이어도 config 기반으로 진행할 수 있다.
