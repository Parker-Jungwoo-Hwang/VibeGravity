# ADR-007: HTTP Router Policy

## Status

Accepted

## Context

VibeGravity API 서버는 HTTP 요청을 받아 core service로 전달하는 역할을 한다.
복잡한 웹 프레임워크보다는 Go 표준에 가깝고 가벼우며 유지보수가 용이한 라우터가 필요하다.

## Decision

VibeGravity v1은 HTTP router로 `github.com/go-chi/chi/v5`를 사용한다.

chi는 `net/http`에 가깝고 가볍다. VibeGravity의 API server는 transport adapter일 뿐이므로 무거운 웹 프레임워크가 필요하지 않다. 라우팅, middleware, health check, graceful shutdown 정도만 담당하게 한다.

API layer는 thin transport adapter로 유지한다.
제품 규칙은 handler가 아니라 core service에 둔다. 즉, handler는 request를 DTO로 바꾸고 core service를 호출한 뒤 response를 반환하는 얇은 계층이어야 한다.

## Consequences

- 라우터 및 HTTP 핸들러 코드는 `internal/httpapi` 패키지에 위치한다.
- `chi.NewRouter()`를 사용하여 라우팅 트리를 구성한다.

## Impact on Hermes-first Roadmap

API 서버의 계층을 얇게 유지함으로써 비즈니스 로직(Hermes 연동 핵심 로직)을 core 패키지에 집중할 수 있게 한다.
