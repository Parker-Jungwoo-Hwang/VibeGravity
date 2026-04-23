# Work Pack 01: Foundation and Repo Setup

## 1. Goal

이 work pack의 목표는 구현 기반을 고정하는 것이다.  
이 단계가 끝나면 repo, core interfaces, schema skeleton, dev loop가 살아 있어야 한다.

## 2. Inputs

반드시 먼저 읽을 문서:

- `00_read-this-first_for-building-agents.md`
- `01_rfp_vibegravity_hermes-first.md`
- `02_product-contract_and_direction.md`
- `03_target-architecture_codex-first.md`

## 3. Deliverables

- monorepo skeleton
- shared core package
- Go HTTP server baseline
- worker baseline
- Postgres migrations baseline
- config loader
- health check
- doctor command
- root `AGENTS.md`
- root `CLAUDE.md`
- root `PLANS.md`

## 4. Required Interfaces

최소 인터페이스를 먼저 고정한다.

```go
type VibeGravityService interface {
    Prefetch(ctx context.Context, req *PrefetchRequest) (*PrefetchResponse, error)
    SyncTurn(ctx context.Context, req *SyncTurnRequest) (*SyncTurnResponse, error)
    AddDocument(ctx context.Context, req *AddDocumentRequest) (*AddDocumentResponse, error)
    SearchMemories(ctx context.Context, req *SearchMemoriesRequest) (*SearchMemoriesResponse, error)
    SearchDocuments(ctx context.Context, req *SearchDocumentsRequest) (*SearchDocumentsResponse, error)
}
```

## 5. Required Decisions

이 단계에서 아래를 ADR 또는 문서로 고정한다.

- package layout
- config strategy
- db migration tool
- queue strategy using postgres table
- Go version and tooling
- code style and test runner

## 6. Tasks

### Task A. Repo bootstrap

Go-first monorepo를 만든다.  
packages, migrations, tests, docs 디렉터리를 만든다.

### Task B. Core contracts

request / response DTO와 service protocols를 만든다.

### Task C. Storage baseline

핵심 테이블 migration skeleton을 만든다.

### Task D. Runtime shell

api server와 worker를 따로 띄울 수 있게 만든다.

### Task E. Config shell

dev config와 local paths를 고정한다.  
`CODEX_HOME`과 local embedding endpoint도 설정값으로 둔다.

## 7. Tests

- app boots
- db migrates
- worker boots
- health check returns ok
- config loads in dev

## 8. Done When

- 로컬에서 server와 worker가 둘 다 뜬다
- health check가 된다
- migration이 적용된다
- core interfaces가 import 가능하다
- repo root instruction files가 생겼다
- 다음 work pack이 바로 이어질 수 있다

## 9. Common Failure Modes

- 구조를 너무 일찍 microservice로 찢는 것
- core contract 없이 route부터 만드는 것
- AGENTS / CLAUDE 없이 coding agent를 바로 돌리는 것
- dev loop를 문서화하지 않는 것
