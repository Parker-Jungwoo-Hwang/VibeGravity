# ADR-005: artifact_class 도입 시점

## Status

Accepted

## Context

memory를 더 구조적으로 분류하기 위한 상위 분류 체계가 필요하다.
`MemoryKind` (fact, preference, trait, ...)는 자세한 의미 타입이다.
`artifact_class`는 retrieval lane과 token packing을 위한 더 큰 묶음이다.

지금 넣지 않으면 Work Pack 02에서 recall block, ranking, token packing을 만들 때 다시 도메인 타입과 migration을 건드리게 된다.
Foundation 단계에서 enum과 column만 넣어 두면 비용이 작다.
기능 구현은 미뤄도 된다. 하지만 구조는 지금 박아야 한다.

## Decision

`artifact_class`는 Foundation에 포함한다.
이 값은 무거운 온톨로지가 아니다.
retrieval lane과 token packing을 위한 상위 분류다.

Foundation에서는 Go enum, Memory domain type, DB column, DTO field, search filter까지만 넣는다.
실제 class-aware ranking과 token packing은 Work Pack 02에서 구현한다.

기본 값은 네 가지로 시작한다.

| artifact_class | 의미 |
|---|---|
| `context` | 현재 세션과 짧은 작업 문맥 |
| `knowledge` | 사실, 선호, 규칙, 절차, 문서 기반 지식 |
| `timeline` | 사건, 결정, 정정, 변화 기록 |
| `plan` | 목표, 작업 상태, 다음 행동 |

`MemoryKind`는 그대로 유지한다.
`artifact_class`는 더 큰 묶음이고, `MemoryKind`는 더 자세한 의미 타입이다.

### 적용 범위

`artifact_class`를 모든 테이블에 억지로 넣지 않는다.
Foundation에서는 `memories` 테이블과 recall DTO 중심이면 충분하다.
`documents`, `notes`, `plans`는 이미 테이블 자체가 class 역할을 한다.
나중에 unified artifact search가 필요해지면 그때 view를 만든다.

## Consequences

- `memories` 테이블에 `artifact_class` 컬럼이 추가된다 (default: `knowledge`).
- recall assembler가 class 기반으로 retrieval lane을 라우팅할 수 있는 기반이 생긴다.
- Work Pack 02에서 migration 추가 없이 class-aware ranking을 구현할 수 있다.

## Impact on Hermes-first Roadmap

class-aware retrieval이 도입되면 Hermes의 `prefetch()` 품질이 향상된다.
Foundation 단계에서 구조만 넣으므로 Hermes 연결 자체에는 영향 없다.
