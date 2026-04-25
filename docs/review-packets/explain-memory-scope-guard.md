# ExplainMemory Scope Guard

Date: 2026-04-25
Scope: operator provenance lookup trust-boundary fix.

## Summary

`ExplainMemory` now scopes provenance lookup to the requested tenant and
workspace before returning trace evidence.

## Finding or slice fixed

The PostgreSQL `ExplainMemory` path accepted `tenant_id` and `workspace_id` at
the service boundary, but the storage query loaded `memory_trace` by
`memory_id` alone. A guessed memory id from another tenant or workspace could
return trace metadata and source evidence.

This slice makes the trace query join through `memories` and require matching
tenant/workspace. It also scopes raw-event and document provenance reads to the
same tenant/workspace.

## Files changed

- `internal/store/postgres/memories.go`
- `internal/store/postgres/memories_test.go`
- `docs/review-packets/explain-memory-scope-guard.md`

## Tests run

- `go test ./internal/store/postgres`

## Remaining risks

- This packet's original remaining risk around private-memory owner filtering
  was closed by `docs/review-packets/explain-memory-visibility-guard.md`.
- Edge rows are still loaded by connected memory id after the scoped trace
  check. That is acceptable for normal graph invariants, but a future hardening
  pass could scope edge expansion through tenant/workspace joins as well.

## Source Review

- Estimated source: in-repo VibeGravity storage contracts and current code.
- Suspected license: project-internal original code.
- Similarity risk: low.
- Human review required: recommended because this is an operator-visible
  provenance trust-boundary fix.
