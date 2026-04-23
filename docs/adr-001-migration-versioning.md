# ADR-001: Migration Versioning Policy

## Status

Accepted

## Context

VibeGravity는 golang-migrate를 migration 도구로 사용한다.
팀 규모가 10명이고 병렬 작업이 많아질 수 있다.
순차 번호(000001)는 초기 부트스트랩에서 직관적이지만, 다수가 동시에 migration을 만들면 번호 충돌이 발생한다.

## Decision

1. **초기 부트스트랩(Work Pack 01)**: `000001_name.up.sql` / `000001_name.down.sql` 순차 번호를 사용한다.
2. **부트스트랩 이후(Work Pack 02~)**: Unix timestamp 기반 버전으로 전환한다. 형식은 `{unix_timestamp}_name.up.sql`이다. 예: `1714000000_add_memory_embedding.up.sql`
3. Timestamp 전환 시점은 `000001_create_core_tables` migration이 production에 적용된 이후로 한다.
4. Migration 생성은 `migrate create -ext sql -dir migrations -seq` 대신 `migrate create -ext sql -dir migrations` (timestamp default)를 사용한다.

## Options Considered

### Option A: Timestamp 전환 (선택됨)
- 장점: 병렬 작업 시 충돌 불가, 도구 기본값과 일치
- 단점: 순서가 직관적이지 않음

### Option B: Migration owner 1명 지정
- 장점: 중앙 관리로 품질 통제
- 단점: 병목, owner 부재 시 작업 차단

## Consequences

- 부트스트랩 migration은 번호가 깔끔하게 유지된다.
- 이후 migration은 timestamp로 자동 정렬되어 충돌이 없다.
- PR review 시 migration 파일의 timestamp를 확인해 순서를 검증한다.

## Impact on Hermes-first Roadmap

없음. 내부 도구 결정이며 제품 의미에 영향 없음.
