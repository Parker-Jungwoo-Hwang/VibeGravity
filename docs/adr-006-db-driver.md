# ADR-006: Database Driver Policy

## Status

Accepted

## Context

VibeGravity v1부터 `pgvector`를 쓰고, `memories.embedding`과 `document_chunks.embedding`을 다룰 예정이다.
Go 표준 라이브러리인 `database/sql`과 `github.com/lib/pq` 조합보다 PostgreSQL 전용 기능을 더 잘 지원하는 드라이버가 필요하다.

## Decision

VibeGravity v1은 DB driver로 `github.com/jackc/pgx/v5`를 사용한다.
런타임 연결은 `pgxpool`을 기본으로 한다.
pgvector 연동은 `pgvector-go`의 pgx adapter(`github.com/pgvector/pgvector-go/pgx`)를 사용한다.

Migration은 기존 결정(ADR-001)대로 `golang-migrate`를 유지하되, 가능하면 `pgx5` database driver를 사용한다.

앱 런타임에서는 server, worker, cli doctor 모두 같은 DB connection factory(`internal/db/pool.go`)를 공유한다.

## Consequences

- DB access는 `github.com/jackc/pgx/v5/pgxpool`을 기본으로 사용한다. `database/sql + lib/pq`는 v1 기본 경로에서 사용하지 않는다.
- Postgres store 구현체는 `internal/store/postgres`에 두고, 생성자는 `*pgxpool.Pool`을 받는다.
- Core service는 pgx를 직접 import하지 않는다.

## Impact on Hermes-first Roadmap

pgvector와의 원활한 통합을 통해 Hermes의 semantic retrieval(`prefetch()`) 성능과 안정성을 극대화한다.
