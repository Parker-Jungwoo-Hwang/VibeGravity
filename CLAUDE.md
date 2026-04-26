# CLAUDE.md

## Project

This repo builds VibeGravity.  
VibeGravity is a shared memory kernel for Hermes-first agent workflows.

## Hold these facts in every session

- Hermes is the first customer
- Language is Go 1.26.2 (match `go.mod`), database is PostgreSQL + pgvector
- Migration tool is golang-migrate (see docs/adr-001)
- Local runtime is embedding-only in v1 (no local LLM text extraction)
- Worker pipeline: local embeddings → neighborhood retrieval → Codex stage 1 extract → Codex stage 2 resolve → apply engine
- Local extractor must not be reintroduced as the default path
- Codex-first reasoning handles extraction (Stage 1) and graph resolution (Stage 2)
- Memory scopes: agent_private, workspace_shared, group_shared, session_scratch — must stay separate
- Artifact classes: context, knowledge, timeline, plan — retrieval lane grouping above MemoryKind
- Raw events and derived memories must stay separate
- Recall must be compact and token-aware (budget_tokens)
- Stage 1, Stage 2, and apply handoff stay schema-first structured JSON only
- Apply engine validates before committing (12-step pipeline)
- Queue is Postgres-backed (ingest_jobs table), not external message broker
- Embedding config: embedding_model, embedding_dims, embedding_updated_at stored per row
- Both memories and document_chunks have vector embeddings (see docs/adr-004)
- ExplainMemory provides provenance tracing — correction write path alone is not enough

## Repo layout summary

- `cmd/server` — HTTP API entry
- `cmd/worker` — background worker entry
- `cmd/cli` — CLI + doctor
- `internal/core` — service interfaces and domain types
- `internal/ingest` — sync_turn write path
- `internal/recall` — prefetch assembler
- `internal/graph` — memory graph and apply engine
- `internal/reasoning` — Codex bridge
- `internal/store` — database layer
- `internal/config` — config loader
- `pkg/` — shared library packages
- `migrations/` — SQL migrations
- `plans/` — architecture docs (read before major changes)
- `.agents/skills/` — reusable agent skill files

## Read these docs before major changes

- `plans/00_read-this-first_for-building-agents.md`
- `plans/01_rfp_vibegravity_hermes-first.md`
- `plans/02_product-contract_and_direction.md`
- `plans/03_target-architecture_codex-first.md`
- `plans/05_runtime-contracts_ingest-recall-apply.md`
- `plans/06_data-model_and_storage-invariants.md`

## Build and test

```bash
go build ./cmd/server && go build ./cmd/worker && go build ./cmd/cli
go test ./...
golangci-lint run
```

## Use skills for procedures

If you need a multi-step workflow, load a skill from `.agents/skills/`.  
Available skills:

- `plan-implement-verify` — feature work with plan, code, checks, review
- `contract-check` — compare changes against product/arch contracts
- `eval-regression` — run golden scenarios to detect memory regressions

Do not turn CLAUDE.md into a long procedure manual.

## Preferred working pattern

Plan first on hard tasks.  
Implement one coherent work unit at a time.  
Run `go test ./...` and `golangci-lint run`.  
Review the diff against contracts.  
Report files changed, checks run, results, and risks.

## Watch for these failures

- scope leakage (agent_private appearing in workspace_shared recall)
- duplicate memory growth (fingerprint dedup not working)
- missing provenance (memory without memory_trace)
- empty or noisy recall (budget not respected, superseded not suppressed)
- silent contract drift (behavior changed without docs update)
- free-form reasoning output crossing the apply boundary
- group shared memory without valid membership
