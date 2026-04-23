# AGENTS.md

## Repo purpose

This repo builds VibeGravity.  
VibeGravity is a shared memory kernel for Hermes and other agents.  
It is not a chat UI and not a generic agent runtime.

## Direction

Keep Hermes-first delivery.  
Keep local runtime embedding-only in v1.  
Keep Codex-first reasoning for text interpretation and graph operations.  
Keep agent_private, workspace_shared, and group_shared memory separate.

## Tech stack

- Language: Go (1.22+)
- Database: PostgreSQL (canonical store), SQLite (tests and lightweight local dev)
- HTTP framework: net/http + chi router
- Embedding runtime: local model server (HTTP endpoint, configurable)
- Reasoning: Codex API (bridge from worker)
- Queue: Postgres-backed job table (`ingest_jobs`)
- Migration tool: golang-migrate (confirmed, see docs/adr-001)
- Vector search: pgvector extension (v1, see docs/adr-002)
- Config: env vars + YAML file, loaded via internal/config
- Embedding config: embedding_model, embedding_dims stored per row

## Repo layout

```text
vibegravity/
├─ cmd/
│  ├─ server/          # HTTP API entrypoint
│  ├─ worker/          # background worker entrypoint
│  └─ cli/             # CLI and doctor command
├─ internal/
│  ├─ core/            # VibeGravityService interface and domain types
│  ├─ ingest/          # sync_turn write path
│  ├─ recall/          # prefetch assembler
│  ├─ graph/           # memory graph and apply engine
│  ├─ reasoning/       # Codex reasoning bridge
│  ├─ mcp/             # MCP surface
│  ├─ hermes/          # Hermes provider adapter
│  ├─ embed/           # local embedding client
│  ├─ config/          # config loader
│  └─ store/           # database layer (postgres + sqlite)
├─ pkg/                # reusable library packages (shared types, helpers)
├─ migrations/         # SQL migration files
├─ tests/              # integration and golden tests
├─ docs/               # ADRs and operational docs
├─ .agents/            # Codex skills and shared agent assets
├─ .claude/            # Claude Code project assets (if needed)
├─ plans/              # architecture and work pack documents
├─ AGENTS.md           # this file (Codex instruction)
├─ CLAUDE.md           # Claude Code instruction
└─ PLANS.md            # current work plan
```

## Reasoning contract

Keep Stage 1 (Extract) and Stage 2 (Resolve) schema-first and structured JSON only.  
Do not let free-form reasoning output cross the apply boundary.  
Apply engine validates before committing — see `05_runtime-contracts`.

## Read before work

Always read these files before making non-trivial changes:

- `plans/00_read-this-first_for-building-agents.md`
- `plans/01_rfp_vibegravity_hermes-first.md`
- `plans/02_product-contract_and_direction.md`
- `plans/03_target-architecture_codex-first.md`
- `plans/05_runtime-contracts_ingest-recall-apply.md`
- `plans/06_data-model_and_storage-invariants.md`

## Worker pipeline (default path)

local embeddings → neighborhood retrieval → Codex stage 1 extract → Codex stage 2 resolve → apply engine

Local LLM is embedding-only. Retrieval helpers and lexical fallback are allowed.  
Local extractor must not be reintroduced as the default path.

## Core interfaces

The full v1 service contract:

```go
type VibeGravityService interface {
    Prefetch(ctx context.Context, req *PrefetchRequest) (*PrefetchResponse, error)
    SyncTurn(ctx context.Context, req *SyncTurnRequest) (*SyncTurnResponse, error)
    AddDocument(ctx context.Context, req *AddDocumentRequest) (*AddDocumentResponse, error)
    SearchMemories(ctx context.Context, req *SearchMemoriesRequest) (*SearchMemoriesResponse, error)
    SearchDocuments(ctx context.Context, req *SearchDocumentsRequest) (*SearchDocumentsResponse, error)
    AddNote(ctx context.Context, req *AddNoteRequest) (*AddNoteResponse, error)
    CreatePlan(ctx context.Context, req *CreatePlanRequest) (*CreatePlanResponse, error)
    UpdatePlan(ctx context.Context, req *UpdatePlanRequest) (*UpdatePlanResponse, error)
    CorrectMemory(ctx context.Context, req *CorrectMemoryRequest) (*CorrectMemoryResponse, error)
    GetTimeline(ctx context.Context, req *GetTimelineRequest) (*GetTimelineResponse, error)
    ExplainMemory(ctx context.Context, req *ExplainMemoryRequest) (*ExplainMemoryResponse, error)
}
```

## API surface (v1)

| method | path | purpose |
|---|---|---|
| POST | /v1/prefetch | recall pack 생성 |
| POST | /v1/sync-turn | turn 기록 |
| POST | /v1/documents | 문서 추가 |
| POST | /v1/search/memories | memory 검색 |
| POST | /v1/search/documents | 문서 검색 |
| POST | /v1/notes | note 생성 |
| POST | /v1/plans | plan 생성 |
| PATCH | /v1/plans/{id} | plan 수정 |
| POST | /v1/memory/correct | memory 교정 |
| GET | /v1/memory/{id}/explain | provenance 추적 |
| GET | /v1/timeline | timeline 조회 |

## Core invariants

- raw events and derived memories must stay separate
- all write paths must be idempotent
- every memory must keep provenance (memory_trace is mandatory)
- every memory must have explicit scope (scope null 금지)
- recall must be budget-aware
- human correction is first-class
- `updates` edge can only target one latest memory at a time
- group shared memory requires valid membership
- profile is rebuildable from raw + memories + edges

## Build and test commands

```bash
# build
go build ./cmd/server
go build ./cmd/worker
go build ./cmd/cli

# test
go test ./...

# lint
golangci-lint run

# migrations
migrate -path migrations -database $DATABASE_URL up

# run dev
go run ./cmd/server
go run ./cmd/worker
```

## Core tables

raw_events, ingest_jobs, entities, memories, memory_edges, memory_trace,
profiles, session_summaries, notes, plans, plan_items, documents,
document_chunks, memory_groups, memory_group_memberships.

See `plans/06_data-model_and_storage-invariants.md` for full schema.

## Workflow

For complex tasks, plan first.  
For repeated procedures, use skills from `.agents/skills/`.  
For bounded exploration, use subagents.  
After coding, run checks, then review your own diff, then update docs.

## Done means

A task is not done until:

- code is implemented
- `go test ./...` passes
- `golangci-lint run` passes
- docs are updated if behavior changed
- risks are reported

## Do not

- reintroduce local extractor dependence into the main path
- blur agent_private and workspace_shared memory
- hide contract changes inside code without docs
- skip tests because the change seems small
- let multiple agents edit the same hot files without coordination
- make architecture changes without an ADR in docs/
- put long procedures in this file (use skills instead)
