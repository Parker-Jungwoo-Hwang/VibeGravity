# Summary

Closed the three push-readiness review blockers from the CTO/CPO inspection:
correction visibility authorization, tenant/workspace scoped profile and
session-summary recall, and safe doctor output.

# Finding or slice fixed

- `CorrectMemory` now authorizes the target before recording side effects.
  `agent_private` corrections require `entity_id` to match the memory owner, and
  `group_shared` corrections require the memory group to appear in
  `visible_group_ids`.
- Profile lookup and session-summary recall now include tenant and workspace
  identity, with a fresh migration and migration contract guard.
- `doctor` now masks URL userinfo passwords, `password=` keyword DSNs, and URL
  query passwords before printing `DatabaseURL`.

# Files changed

- `internal/core/dto.go`
- `internal/core/profile.go`
- `internal/kernel/service.go`
- `internal/store/store.go`
- `internal/store/postgres/profiles_summaries.go`
- `internal/recall/assembler.go`
- `internal/worker/stage2_sources.go`
- `internal/mcp/protocol.go`
- `cmd/cli/main.go`
- `migrations/000002_create_core_tables.up.sql`
- `migrations/000005_scope_profiles_and_summaries.up.sql`
- `migrations/000005_scope_profiles_and_summaries.down.sql`
- focused tests in `internal/kernel`, `internal/recall`, `internal/worker`,
  `internal/mcp`, `cmd/cli`, and `tests`
- contract docs in `plans/05`, `plans/06`, and `plans/10`

# Tests run

- `go test ./internal/kernel`
- `go test ./internal/recall`
- `go test ./internal/worker`
- `go test ./internal/mcp`
- `go test ./cmd/cli`
- `go test ./tests`
- `go test -count=1 ./...`
- `go build ./cmd/server ./cmd/worker ./cmd/cli`
- `make eval`
- `make lint`
- `make check-headers`
- `git diff --check`
- `go run ./cmd/cli eval demo`
- `make integration-postgres` skipped because `VIBEGRAVITY_DB_URL` is unset.

# Remaining risks

`make integration-postgres` still needs a real `VIBEGRAVITY_DB_URL` to prove the
new migration and scoped recall behavior against live PostgreSQL.

# Source Review

Estimated source: repo-local implementation from the review findings and
existing VibeGravity contracts. Suspected license: project-owned. Similarity
risk: low; no external code or structured snippets were used. Human review
required: yes, for migration/backfill policy before applying to any non-scratch
database with existing profile rows.
