# Scope Safety and Group Shared Guardrails

## Summary

Agent 7 added focused guardrail coverage for memory scope boundaries across
recall, PostgreSQL search/provenance lookup, and graph apply.

This slice is test-only. It does not implement `group_shared` writes and does
not add any shortcut membership rule.

## Finding or slice fixed

Covered leakage paths:

- `agent_private` recall requests must carry the requesting actor as
  `owner_entity_id`.
- `workspace_shared` memory remains visible through tenant/workspace-bounded
  recall and search scopes.
- `group_shared` recall is requested only when explicit memberships produce
  `visible_group_ids`.
- PostgreSQL memory search keeps `group_shared` tied to `group_id = ANY($7)`;
  no visible groups means no group IDs are opened.
- `ExplainMemory` provenance lookup remains tenant/workspace bounded and keeps
  the same `agent_private` owner and `group_shared` visible-group predicates so
  explain cannot become a bypass around search visibility.
- Store-backed graph apply rejects `group_shared` create, extend, and update
  attempts with `ErrNotImplemented` before any storage write while membership
  validation is incomplete.

## Files changed

- `internal/recall/scope_safety_test.go`
- `internal/store/postgres/search_test.go`
- `internal/store/postgres/scope_safety_test.go`
- `internal/graph/store_apply_test.go`
- `docs/review-packets/scope-safety-and-group-shared-guardrails.md`

## Tests run

- `go test ./internal/store/postgres`
- `go test ./internal/recall`
- `go test ./internal/graph`
- `go test ./...`
- `make eval`
- `make lint`
- `make check-headers`
- `git diff --check`

## Remaining risks

- This slice adds deterministic unit/SQL-shape coverage. It does not add a live
  PostgreSQL membership integration test.
- The current `group_shared` write stop-line remains intentionally closed.
  Opening it later needs a real write-transaction membership check, not a
  caller-provided shortcut.
- Active concurrent agents are editing `internal/store/postgres/memories.go`,
  `internal/store/postgres/memories_test.go`,
  `internal/store/postgres/memories_replay_test.go`,
  `internal/recall/assembler_test.go`, and several docs/plans, so final
  repo-wide gate results must be interpreted against that shared worktree.

## Source Review

- Estimated source: internal repo contracts and existing VibeGravity test
  patterns.
- Suspected license: project-owned code only.
- Similarity risk: low; tests were written from the local product contract and
  existing public function behavior.
- Human review required: normal review recommended for the scope predicates and
  the decision to keep this slice test-only.
