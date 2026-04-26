# Architecture Runtime Cleanup

## Summary

Moved process composition into `internal/runtime`, reduced `internal/kernel` to
a facade, split correction/document/plan/timeline behavior into owned
application packages, and tightened MCP/schema, mocked Codex, and retrieval
truthfulness notes.

## Finding or slice fixed

- `cmd/server`, `cmd/cli`, `cmd/vibegravity`, and `cmd/worker` no longer each
  own their own service or worker dependency graph.
- `internal/kernel.Service` delegates product behavior instead of carrying all
  correction, document, plan, and timeline logic directly.
- MCP required input schemas now have a contract test that exercises service
  validation for every required field exposed through `tools/list`.
- Worker reasoning logs `MockCodexJSONClient` use and only treats real Codex as
  a future explicit `VIBEGRAVITY_CODEX_CLIENT=real` path.
- `internal/embed` remains out of scope for this slice; current retrieval is
  lexical/store-backed.

## Files changed

- `cmd/server/main.go`
- `cmd/cli/main.go`
- `cmd/vibegravity/main.go`
- `cmd/worker/main.go`
- `internal/runtime/*`
- `internal/kernel/service.go`
- `internal/corrections/*`
- `internal/documents/*`
- `internal/plans/*`
- `internal/timeline/*`
- `internal/mcp/protocol_test.go`
- `internal/config/config.go`
- `internal/embed/doc.go`
- `README.md`
- `PLANS.md`
- `docs/status.md`
- `plans/05_runtime-contracts_ingest-recall-apply.md`

## Tests run

- `go test ./internal/kernel ./internal/mcp ./internal/config ./cmd/worker ./cmd/cli ./cmd/vibegravity ./cmd/server ./internal/runtime`

## Remaining risks

- Full `go test ./...`, lint, and header gates still need to run after this
  slice because the worktree already contains unrelated edits.
- Real Codex client implementation is still intentionally absent.
- Embedding/vector retrieval is still intentionally absent; the current path is
  lexical/store-backed.

## Source Review

Estimated source: repo-local AGENTS.md, PLANS.md, controlling plan docs, and
current Go code.

Suspected license: internal project material only.

Similarity risk: low. The implementation follows the existing repo patterns and
does not reproduce external snippets.

Human review required: recommended because this is process-composition
architecture work.
