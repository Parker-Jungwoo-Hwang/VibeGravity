You are the vuitton Hermes profile working as Go/Postgres implementation reviewer.

Repo: VibeGravity repository root

Inspect:
- internal/graph/apply.go
- internal/graph/store_apply.go
- internal/store/store.go
- internal/store/postgres/memories.go
- internal/store/postgres/helpers.go
- migrations/000002_create_core_tables.up.sql
- docs/adr-009-updates-edge-lineage-guard.md
- plans/05_runtime-contracts_ingest-recall-apply.md

Task:
Design the narrow code change for enabling update_memory writes. Focus on transaction shape, target latest guard, memory_trace rollback, updates edge direction, and scope constraints.

Return only:
1. Proposed store interface changes.
2. Proposed Postgres transaction steps.
3. Proposed graph apply changes.
4. Risk points to verify.
