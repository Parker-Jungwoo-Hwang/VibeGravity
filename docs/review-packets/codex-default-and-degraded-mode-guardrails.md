# Codex Default And Degraded Mode Guardrails

## Summary

Agent 8 strengthened the guardrail coverage around real-Codex defaults, worker
reasoning stop-lines, and degraded recall truthfulness. The implementation stays
test-only plus this review packet: real Codex remains disabled by default, the
worker default remains the mocked `CodexJSONClient` bridge, and no local
extractor fallback was added.

## Finding or slice fixed

- `VIBEGRAVITY_CODEX_ENABLED` is still disabled by default. Invalid boolean
  values also fall back to disabled even if endpoint/model strings are present.
- The default worker reasoner ignores real-Codex environment variables and still
  runs through the deterministic mocked Codex bridge. The mocked path returns no
  extracted memories and no graph operations.
- Degraded recall still returns useful stored context from notes, active plans,
  profile snapshots, session summaries, memories, and documents while marking
  derived sources (`profile`, `session_summaries`, `memories`) stale when
  worker/Codex lag is visible.
- Operator-facing `cli jobs metrics` already exists and remains read-only in the
  tested surface: it fetches backlog metrics without listing blocked jobs or
  requeueing anything.

## Files changed

- `internal/config/config_test.go`
- `cmd/worker/main_test.go`
- `internal/recall/assembler_test.go`
- `cmd/cli/main_test.go`
- `docs/review-packets/codex-default-and-degraded-mode-guardrails.md`

## Tests run

- `go test ./internal/config ./cmd/worker ./internal/recall ./cmd/cli`
- `go test ./...`
- `make lint`
- `make check-headers`
- `git diff --check`
- `make eval` was not run because eval fixtures or eval code were not touched.

## Remaining risks

- Real Codex is not implemented behind a production client yet; this slice only
  protects the disabled/mock default boundary.
- Freshness is inferred from worker backlog metrics, not from a dedicated Codex
  outage state table.
- Live PostgreSQL degraded-mode behavior should still be validated with a real
  database before claiming production readiness.

## Source Review

- Estimated source: repo-local tests and contracts only.
- Suspected license: project-owned VibeGravity code.
- Similarity risk: low; no external code or structured snippets were used.
- Human review required: recommended for product wording and final go/no-go on
  real Codex enablement, not for licensing.
