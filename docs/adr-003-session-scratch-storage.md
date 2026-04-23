# ADR-003: session_scratch Storage Policy

## Status

Accepted

## Context

`session_scratch`는 짧은 수명의 작업 문맥이다.
Redis나 in-memory로 분리하면 저장 경계가 나뉘어 replay 가능성이 줄어든다.
v1에서는 저장 일관성과 replay 가능성이 최적화보다 중요하다.

## Decision

1. **v1에서는 Postgres 안에서 scope로 관리한다.** `memories` 테이블에서 `scope = 'session_scratch'`로 구분한다.
2. **수명 관리**: `expires_at` 필드를 사용한다. `memories` 테이블에 이미 `valid_to` (nullable)가 있으므로 이를 활용한다.
3. **정리 방식**: `maintenance` job kind로 expired session_scratch를 주기적으로 archived 상태로 전환한다. Hard delete는 하지 않는다.
4. **v1.5 이후 최적화 옵션**: 데이터 규모가 커지면 session_scratch를 별도 파티션이나 TTL 테이블로 분리하는 것을 검토한다.

## Consequences

- 모든 scope의 memory가 단일 테이블에 있어 쿼리와 replay가 단순하다.
- session_scratch가 많이 쌓이면 테이블이 커질 수 있지만, maintenance job이 정리한다.
- Postgres 트랜잭션 안에서 모든 scope를 함께 다룰 수 있다.

## Impact on Hermes-first Roadmap

Hermes의 session lifecycle과 자연스럽게 맞는다.
session 종료 시 scratch를 정리하는 dreaming hint와 연결된다.
