# GPT-Pro Review Request Prompt

You are GPT-Pro reviewing VibeGravity, a Go-first, PostgreSQL-backed shared memory kernel for Hermes Memory.

Please review the attached material files in this order:

1. `00_codex_project_review.md`
2. `01_gpt_pro_context_and_reading_order.md`
3. `02_gpt_pro_core_config_and_commands.md`
4. `03_gpt_pro_ingest_recall_kernel_http.md`
5. `04_gpt_pro_reasoning_graph_worker_eval.md`
6. `05_gpt_pro_postgres_store_and_migrations.md`
7. `06_gpt_pro_external_surfaces_hermes_mcp.md`
8. `07_gpt_pro_tests_and_golden_scenarios.md`
9. `08_gpt_pro_docs_plans_adrs_review_packets.md`
10. `09_gpt_pro_inventory_verification_and_risk_map.md`

Review goals:

1. Validate or refute the P0 finding: correction supersession appears to write synthetic `correction:<id>` values into `memory_trace.reasoning_job_id` and `memory_edges.created_by_job_id`, while the migration constrains both to `ingest_jobs(id)`.
2. Validate the MCP schema mismatch finding: `tools/list` advertises fewer required fields than the service actually requires.
3. Validate the replay/idempotency finding: update/create/extend retries appear count-idempotent but may not reject changed payload evidence.
4. Judge whether the project is coherent as **Hermes Memory, powered by VibeGravity** and whether the current next slice should be DB/protocol correctness rather than new features.
5. Produce a prioritized fix plan that preserves these stop-lines: no local extractor, no real Codex by default, no group_shared writes before membership validation, no broad Hermes packaging before trust-loop correctness.

Please return:

- Executive verdict: V1 readiness and biggest blocker.
- Findings ordered by severity with file/line evidence.
- Fix plan for the next 1-3 implementation slices.
- Tests/integration gates to add, especially live Postgres coverage.
- Any architectural drift between docs and code.
- Whether you agree that the current local green gates are insufficient for the correction trust loop.

Assume today is 2026-04-25 and the repo is under active development with a dirty working tree. Do not recommend reverting unrelated work.
