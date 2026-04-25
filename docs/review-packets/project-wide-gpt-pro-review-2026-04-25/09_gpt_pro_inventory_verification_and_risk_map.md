# GPT-Pro Inventory, Verification, and Risk Map

Generated: 2026-04-25

## Inventory Summary

- `binary_or_non_utf8`: 6
- `included_candidate`: 224
- `inventoried_only`: 62

## Verification Evidence

```text
go test ./...                                      PASS
make eval                                          PASS
make lint                                          PASS
make check-headers                                 PASS
git diff --check                                  PASS
```

## Current Git Status Before This Packet

```text
M AGENTS.md
 M Makefile
 M PLANS.md
 M bin/cli
 M cmd/cli/main.go
 M cmd/server/main.go
 M cmd/worker/main.go
 M internal/config/config.go
 M internal/core/dto.go
 M internal/core/errors.go
 M internal/core/job.go
 M internal/core/kind.go
 M internal/core/memory.go
 M internal/httpapi/router.go
 M internal/store/postgres/store.go
 M internal/store/store.go
 M migrations/000002_create_core_tables.down.sql
 M migrations/000002_create_core_tables.up.sql
 M plans/02_product-contract_and_direction.md
 M plans/05_runtime-contracts_ingest-recall-apply.md
 M plans/06_data-model_and_storage-invariants.md
 M plans/09_workpack_memory-graph-and-dreaming.md
 M plans/10_workpack_hermes-provider-and-external-surfaces.md
 M plans/11_workpack_quality-ops-and-evals.md
?? .agents/coordination/
?? .agents/hermes-orchestration/
?? cmd/cli/main_test.go
?? consulting/
?? docs/adr-008-package-layout.md
?? docs/adr-009-updates-edge-lineage-guard.md
?? docs/review-packets/
?? internal/core/dreaming.go
?? internal/embed/
?? internal/eval/
?? internal/graph/
?? internal/hermes/
?? internal/httpapi/router_test.go
?? internal/ingest/
?? internal/kernel/
?? internal/mcp/
?? internal/reasoning/
?? internal/recall/
?? internal/store/postgres/concurrency_integration_test.go
?? internal/store/postgres/corrections.go
?? internal/store/postgres/corrections_test.go
?? internal/store/postgres/documents.go
?? internal/store/postgres/documents_test.go
?? internal/store/postgres/dreaming.go
?? internal/store/postgres/dreaming_test.go
?? internal/store/postgres/groups.go
?? internal/store/postgres/helpers.go
?? internal/store/postgres/jobs.go
?? internal/store/postgres/jobs_test.go
?? internal/store/postgres/memories.go
?? internal/store/postgres/memories_test.go
?? internal/store/postgres/notes_plans.go
?? internal/store/postgres/notes_plans_test.go
?? internal/store/postgres/profiles_summaries.go
?? internal/store/postgres/raw_events.go
?? internal/store/postgres/search.go
?? internal/store/postgres/search_test.go
?? internal/store/postgres/timeline.go
?? internal/store/postgres/timeline_test.go
?? internal/worker/
?? tests/golden/
?? tests/migration_contract_test.go
```

## Risk Map For GPT-Pro

- P0: `CorrectMemory` correction supersession uses synthetic correction IDs in fields constrained to `ingest_jobs(id)` by migration FKs.
- P1: MCP `tools/list` schemas understate required fields compared with service validation.
- P1: update-memory replay idempotency does not compare full operation payload evidence before accepting retry success.
- P2: local deterministic gates pass, but live Postgres/Hermes/Codex paths remain mostly mocked or opt-in.

## Full File Inventory

| path | status | bytes | lines | sha16 | note |
|---|---:|---:|---:|---|---|
| `.DS_Store` | binary_or_non_utf8 | 8196 | 0 | `e4aba8354474b909` | not bundled as text |
| `.agents/coordination/.gitignore` | included_candidate | 59 | 6 | `43a7857fbc3c7b56` |  |
| `.agents/coordination/PROMPT_SNIPPET.md` | included_candidate | 1039 | 29 | `b2438dc8817f7a23` |  |
| `.agents/coordination/README.md` | included_candidate | 3336 | 111 | `efbbde2bb0a45f9b` |  |
| `.agents/coordination/UNIVERSAL_AGENT_PROMPT.md` | included_candidate | 6279 | 136 | `6958f10612a6c23a` |  |
| `.agents/coordination/WORK_PROGRESS.md` | included_candidate | 11766 | 55 | `e5d961ec411e9820` |  |
| `.agents/coordination/activity.log` | included_candidate | 12165 | 38 | `a0b8ea383977c0b6` |  |
| `.agents/coordination/agent-work.sh` | included_candidate | 7255 | 314 | `2bd901f9bdb73213` |  |
| `.agents/coordination/claims.tsv` | included_candidate | 0 | 0 | `e3b0c44298fc1c14` |  |
| `.agents/hermes-orchestration/README.md` | included_candidate | 2118 | 68 | `2d356e2a742b9869` |  |
| `.agents/hermes-orchestration/collect.sh` | included_candidate | 906 | 41 | `2450cef6a5166ecb` |  |
| `.agents/hermes-orchestration/dispatch.sh` | included_candidate | 1396 | 75 | `eb807aa84cb503d2` |  |
| `.agents/hermes-orchestration/run-agent.sh` | included_candidate | 1924 | 100 | `ab5de56eeac3a035` |  |
| `.agents/hermes-orchestration/runs/.gitignore` | inventoried_only | 24 | 4 | `245498699e2ac286` | binary/local telemetry/generated review output or runtime logs |
| `.agents/hermes-orchestration/runs/.gitkeep` | inventoried_only | 1 | 2 | `01ba4719c80b6fe9` | binary/local telemetry/generated review output or runtime logs |
| `.agents/hermes-orchestration/runs/20260424-eval-graph-gates/bottega.err.log` | inventoried_only | 36 | 3 | `55e7cace70c15f46` | binary/local telemetry/generated review output or runtime logs |
| `.agents/hermes-orchestration/runs/20260424-eval-graph-gates/bottega.meta` | inventoried_only | 538 | 12 | `a862a7c5f5168710` | binary/local telemetry/generated review output or runtime logs |
| `.agents/hermes-orchestration/runs/20260424-eval-graph-gates/bottega.out.md` | inventoried_only | 17073 | 375 | `d030268b66ac0b4d` | binary/local telemetry/generated review output or runtime logs |
| `.agents/hermes-orchestration/runs/20260424-eval-graph-gates/default.err.log` | inventoried_only | 36 | 3 | `9078a3374a92b16c` | binary/local telemetry/generated review output or runtime logs |
| `.agents/hermes-orchestration/runs/20260424-eval-graph-gates/default.meta` | inventoried_only | 521 | 12 | `d56f69346f2a8291` | binary/local telemetry/generated review output or runtime logs |
| `.agents/hermes-orchestration/runs/20260424-eval-graph-gates/default.out.md` | inventoried_only | 12264 | 251 | `30b6b69faaa0b48a` | binary/local telemetry/generated review output or runtime logs |
| `.agents/hermes-orchestration/runs/20260424-eval-graph-gates/vuitton.err.log` | inventoried_only | 36 | 3 | `e151e758d36ef00b` | binary/local telemetry/generated review output or runtime logs |
| `.agents/hermes-orchestration/runs/20260424-eval-graph-gates/vuitton.meta` | inventoried_only | 538 | 12 | `158c28fa7acbd5d5` | binary/local telemetry/generated review output or runtime logs |
| `.agents/hermes-orchestration/runs/20260424-eval-graph-gates/vuitton.out.md` | inventoried_only | 18286 | 530 | `48130b9773d72b68` | binary/local telemetry/generated review output or runtime logs |
| `.agents/hermes-orchestration/runs/20260424_backlog_metrics/bottega.err.log` | inventoried_only | 36 | 3 | `30324d88c718c38a` | binary/local telemetry/generated review output or runtime logs |
| `.agents/hermes-orchestration/runs/20260424_backlog_metrics/bottega.meta` | inventoried_only | 534 | 12 | `61136d852e511e3c` | binary/local telemetry/generated review output or runtime logs |
| `.agents/hermes-orchestration/runs/20260424_backlog_metrics/bottega.out.md` | inventoried_only | 10982 | 306 | `1b1d450c0723c052` | binary/local telemetry/generated review output or runtime logs |
| `.agents/hermes-orchestration/runs/20260424_backlog_metrics/default.err.log` | inventoried_only | 36 | 3 | `9d74b555fa4a5c02` | binary/local telemetry/generated review output or runtime logs |
| `.agents/hermes-orchestration/runs/20260424_backlog_metrics/default.meta` | inventoried_only | 517 | 12 | `44044c9c5ca38848` | binary/local telemetry/generated review output or runtime logs |
| `.agents/hermes-orchestration/runs/20260424_backlog_metrics/default.out.md` | inventoried_only | 7229 | 148 | `633a83acbf4803f9` | binary/local telemetry/generated review output or runtime logs |
| `.agents/hermes-orchestration/runs/20260424_backlog_metrics/vuitton.err.log` | inventoried_only | 36 | 3 | `575dd826ff76f844` | binary/local telemetry/generated review output or runtime logs |
| `.agents/hermes-orchestration/runs/20260424_backlog_metrics/vuitton.meta` | inventoried_only | 534 | 12 | `a5800e39ea6271f1` | binary/local telemetry/generated review output or runtime logs |
| `.agents/hermes-orchestration/runs/20260424_backlog_metrics/vuitton.out.md` | inventoried_only | 10295 | 308 | `5780903c4979dfc3` | binary/local telemetry/generated review output or runtime logs |
| `.agents/hermes-orchestration/runs/smoke-001/bottega.err.log` | inventoried_only | 36 | 3 | `ccc63f9194860543` | binary/local telemetry/generated review output or runtime logs |
| `.agents/hermes-orchestration/runs/smoke-001/bottega.meta` | inventoried_only | 470 | 12 | `69b52e60275cb0bb` | binary/local telemetry/generated review output or runtime logs |
| `.agents/hermes-orchestration/runs/smoke-001/bottega.out.md` | inventoried_only | 19 | 2 | `21e3a9e817df8f2a` | binary/local telemetry/generated review output or runtime logs |
| `.agents/hermes-orchestration/runs/smoke-001/default.err.log` | inventoried_only | 36 | 3 | `4f34611f4e6443da` | binary/local telemetry/generated review output or runtime logs |
| `.agents/hermes-orchestration/runs/smoke-001/default.meta` | inventoried_only | 453 | 12 | `6ff850ed70ae757b` | binary/local telemetry/generated review output or runtime logs |
| `.agents/hermes-orchestration/runs/smoke-001/default.out.md` | inventoried_only | 19 | 2 | `591875f095585895` | binary/local telemetry/generated review output or runtime logs |
| `.agents/hermes-orchestration/runs/smoke-001/vuitton.err.log` | inventoried_only | 36 | 3 | `a5bf4a605ab2e8ef` | binary/local telemetry/generated review output or runtime logs |
| `.agents/hermes-orchestration/runs/smoke-001/vuitton.meta` | inventoried_only | 470 | 12 | `2ad3ce0c9a9fa54b` | binary/local telemetry/generated review output or runtime logs |
| `.agents/hermes-orchestration/runs/smoke-001/vuitton.out.md` | inventoried_only | 19 | 2 | `f8b86440fcba9817` | binary/local telemetry/generated review output or runtime logs |
| `.agents/hermes-orchestration/runs/v1-update-memory-001/bottega.err.log` | inventoried_only | 0 | 0 | `e3b0c44298fc1c14` | binary/local telemetry/generated review output or runtime logs |
| `.agents/hermes-orchestration/runs/v1-update-memory-001/bottega.meta` | inventoried_only | 514 | 12 | `43c6b4c80ac610da` | binary/local telemetry/generated review output or runtime logs |
| `.agents/hermes-orchestration/runs/v1-update-memory-001/bottega.out.md` | inventoried_only | 11302 | 232 | `378636b633682f2a` | binary/local telemetry/generated review output or runtime logs |
| `.agents/hermes-orchestration/runs/v1-update-memory-001/default.err.log` | inventoried_only | 36 | 3 | `e84f6d081b2090cf` | binary/local telemetry/generated review output or runtime logs |
| `.agents/hermes-orchestration/runs/v1-update-memory-001/default.meta` | inventoried_only | 497 | 12 | `4ccec6389b2bbd28` | binary/local telemetry/generated review output or runtime logs |
| `.agents/hermes-orchestration/runs/v1-update-memory-001/default.out.md` | inventoried_only | 5655 | 121 | `f366179ee2ec973a` | binary/local telemetry/generated review output or runtime logs |
| `.agents/hermes-orchestration/runs/v1-update-memory-001/vuitton.err.log` | inventoried_only | 36 | 3 | `a8e51beaa311cdd8` | binary/local telemetry/generated review output or runtime logs |
| `.agents/hermes-orchestration/runs/v1-update-memory-001/vuitton.meta` | inventoried_only | 514 | 12 | `c89fb1205dea209f` | binary/local telemetry/generated review output or runtime logs |
| `.agents/hermes-orchestration/runs/v1-update-memory-001/vuitton.out.md` | inventoried_only | 10646 | 221 | `362fffc8db4ec7fb` | binary/local telemetry/generated review output or runtime logs |
| `.agents/hermes-orchestration/status.sh` | included_candidate | 159 | 12 | `b50df71268139ca6` |  |
| `.agents/hermes-orchestration/tasks/20260424-eval-graph-gates/bottega.md` | included_candidate | 1224 | 35 | `9178bb9124bd4cbe` |  |
| `.agents/hermes-orchestration/tasks/20260424-eval-graph-gates/default.md` | included_candidate | 1223 | 33 | `e85e7ad14a5cd652` |  |
| `.agents/hermes-orchestration/tasks/20260424-eval-graph-gates/manifest.tsv` | included_candidate | 240 | 4 | `e62362197aa481d9` |  |
| `.agents/hermes-orchestration/tasks/20260424-eval-graph-gates/vuitton.md` | included_candidate | 1209 | 34 | `f2493317267a4c00` |  |
| `.agents/hermes-orchestration/tasks/20260424_backlog_metrics/bottega.md` | included_candidate | 1020 | 29 | `975dc0ee436d9d88` |  |
| `.agents/hermes-orchestration/tasks/20260424_backlog_metrics/default.md` | included_candidate | 1059 | 28 | `46d28d31a9de60a6` |  |
| `.agents/hermes-orchestration/tasks/20260424_backlog_metrics/manifest.tsv` | included_candidate | 237 | 4 | `1ce1a01767fc6bd7` |  |
| `.agents/hermes-orchestration/tasks/20260424_backlog_metrics/vuitton.md` | included_candidate | 1189 | 42 | `85523f19e6a50164` |  |
| `.agents/hermes-orchestration/tasks/smoke/bottega.md` | included_candidate | 71 | 4 | `f834ccbd36ea5628` |  |
| `.agents/hermes-orchestration/tasks/smoke/default.md` | included_candidate | 71 | 4 | `c952c4e9dc864440` |  |
| `.agents/hermes-orchestration/tasks/smoke/manifest.tsv` | included_candidate | 181 | 5 | `c3ab1eb70530fd59` |  |
| `.agents/hermes-orchestration/tasks/smoke/vuitton.md` | included_candidate | 71 | 4 | `1a63e7eb1a3c71c3` |  |
| `.agents/hermes-orchestration/tasks/v1-update-memory/bottega.md` | included_candidate | 454 | 20 | `526e7b245a317d52` |  |
| `.agents/hermes-orchestration/tasks/v1-update-memory/default.md` | included_candidate | 719 | 24 | `bd8ddd0eee8d429b` |  |
| `.agents/hermes-orchestration/tasks/v1-update-memory/manifest.tsv` | included_candidate | 214 | 5 | `2f1626d2b6705488` |  |
| `.agents/hermes-orchestration/tasks/v1-update-memory/vuitton.md` | included_candidate | 780 | 24 | `5f8ea70584e01056` |  |
| `.agents/skills/code-headers.md` | included_candidate | 1064 | 34 | `a045104a41d83025` |  |
| `.agents/skills/contract-check.md` | included_candidate | 1171 | 39 | `6ee5246ce5854512` |  |
| `.agents/skills/eval-regression.md` | included_candidate | 888 | 33 | `4aa75924b79f3a65` |  |
| `.agents/skills/plan-implement-verify.md` | included_candidate | 764 | 33 | `88f46f0e4ef22587` |  |
| `.agents/skills/source-provenance.md` | included_candidate | 1905 | 56 | `509c2e94b5266470` |  |
| `.gitignore` | included_candidate | 16 | 3 | `fe048748c7e01757` |  |
| `.gitmessage.txt` | included_candidate | 684 | 26 | `2b28ba5552cc6ae5` |  |
| `.golangci.yml` | included_candidate | 352 | 26 | `a9626d6f4cbc26dd` |  |
| `.omx/logs/notify-fallback-2026-04-23.jsonl` | inventoried_only | 9689 | 47 | `42df9995eeaf81be` | binary/local telemetry/generated review output or runtime logs |
| `.omx/logs/notify-fallback-2026-04-25.jsonl` | inventoried_only | 90795 | 464 | `44002fb6341d0c4a` | binary/local telemetry/generated review output or runtime logs |
| `.omx/logs/omx-2026-04-23.jsonl` | inventoried_only | 858 | 7 | `0cc79e274ce1d262` | binary/local telemetry/generated review output or runtime logs |
| `.omx/logs/omx-2026-04-25.jsonl` | inventoried_only | 436 | 4 | `a7aadcf455af828f` | binary/local telemetry/generated review output or runtime logs |
| `.omx/logs/session-history.jsonl` | inventoried_only | 704 | 5 | `f5d63e4e84f69628` | binary/local telemetry/generated review output or runtime logs |
| `.omx/logs/tmux-hook-2026-04-23.jsonl` | inventoried_only | 3234 | 34 | `87313af0b8da2c02` | binary/local telemetry/generated review output or runtime logs |
| `.omx/logs/tmux-hook-2026-04-24.jsonl` | inventoried_only | 5292 | 55 | `63ff968fb5e9f586` | binary/local telemetry/generated review output or runtime logs |
| `.omx/logs/tmux-hook-2026-04-25.jsonl` | inventoried_only | 2842 | 30 | `4880c2c87f1ed0f5` | binary/local telemetry/generated review output or runtime logs |
| `.omx/logs/turns-2026-04-23.jsonl` | inventoried_only | 19176 | 34 | `0a7bc2aa3b6c4b67` | binary/local telemetry/generated review output or runtime logs |
| `.omx/logs/turns-2026-04-24.jsonl` | inventoried_only | 32437 | 55 | `b5f2830d697fe1c1` | binary/local telemetry/generated review output or runtime logs |
| `.omx/logs/turns-2026-04-25.jsonl` | inventoried_only | 19473 | 31 | `7db769c459636ba2` | binary/local telemetry/generated review output or runtime logs |
| `.omx/metrics.json` | included_candidate | 234 | 10 | `ebf5db3fc7a1f554` |  |
| `.omx/state/current-task-baseline.json` | inventoried_only | 261 | 13 | `68c0562e5e689ab8` | binary/local telemetry/generated review output or runtime logs |
| `.omx/state/hud-state.json` | inventoried_only | 194 | 6 | `4b8e18a43a78973b` | binary/local telemetry/generated review output or runtime logs |
| `.omx/state/notify-fallback-state.json` | inventoried_only | 2615 | 90 | `bab9cdf25aaecb59` | binary/local telemetry/generated review output or runtime logs |
| `.omx/state/notify-hook-state.json` | inventoried_only | 4713 | 45 | `998150b8ea0c3547` | binary/local telemetry/generated review output or runtime logs |
| `.omx/state/session.json` | inventoried_only | 178 | 7 | `6d94f30dd3a73920` | binary/local telemetry/generated review output or runtime logs |
| `.omx/state/sessions/omx-1776978217416-332lnt/notify-hook-state.json` | inventoried_only | 189 | 6 | `6b70c7885b6aa99f` | binary/local telemetry/generated review output or runtime logs |
| `.omx/state/sessions/omx-1777096731666-a1yw2b/notify-hook-state.json` | inventoried_only | 189 | 6 | `93f8b7334a6db233` | binary/local telemetry/generated review output or runtime logs |
| `.omx/state/sessions/omx-1777096742370-zjgx7m/AGENTS.md` | inventoried_only | 9047 | 226 | `9eb0df7cd6858f64` | binary/local telemetry/generated review output or runtime logs |
| `.omx/state/sessions/omx-1777096742370-zjgx7m/hud-state.json` | inventoried_only | 144 | 6 | `81d80c70cd21caaa` | binary/local telemetry/generated review output or runtime logs |
| `.omx/state/skill-active-state.json` | inventoried_only | 781 | 27 | `ab3576fb9fd38089` | binary/local telemetry/generated review output or runtime logs |
| `.omx/state/subagent-tracking.json` | inventoried_only | 1209 | 36 | `23a953ab2089abce` | binary/local telemetry/generated review output or runtime logs |
| `.omx/state/team-leader-nudge.json` | inventoried_only | 91 | 5 | `05a821e65d035c79` | binary/local telemetry/generated review output or runtime logs |
| `.omx/state/tmux-hook-state.json` | inventoried_only | 195 | 9 | `cdcdcd86bcf8dcce` | binary/local telemetry/generated review output or runtime logs |
| `AGENTS.md` | included_candidate | 8481 | 214 | `60181277d06ab45a` |  |
| `CLAUDE.md` | included_candidate | 3637 | 90 | `48ab49e752353f8c` |  |
| `COMMIT_MESSAGE_RULES.md` | included_candidate | 1994 | 93 | `0b712b5a6f88b161` |  |
| `Makefile` | included_candidate | 911 | 44 | `a79d2a4f0ac5007c` |  |
| `PLANS.md` | included_candidate | 7057 | 136 | `e0ac8c3ab5dbd36a` |  |
| `bin/cli` | binary_or_non_utf8 | 15012018 | 0 | `5c879a535124e51f` | not bundled as text |
| `bin/server` | binary_or_non_utf8 | 2371362 | 0 | `d7ad6b7657d9e434` | not bundled as text |
| `bin/worker` | binary_or_non_utf8 | 2371362 | 0 | `7202091deb5b8051` | not bundled as text |
| `cmd/cli/main.go` | included_candidate | 18769 | 644 | `d33ce794ed5440db` |  |
| `cmd/cli/main_test.go` | included_candidate | 12120 | 350 | `3f2f3715ff40be13` |  |
| `cmd/server/main.go` | included_candidate | 3348 | 119 | `a0d456034c689678` |  |
| `cmd/worker/main.go` | included_candidate | 4354 | 151 | `8accb3b24c8382c8` |  |
| `consulting/00_consulting_request.md` | included_candidate | 3641 | 79 | `f30291c72745268f` |  |
| `consulting/01_reading_order.md` | included_candidate | 2431 | 74 | `e18425867e7503bd` |  |
| `consulting/02_product_one_pager.md` | included_candidate | 3121 | 89 | `1fedab00ec6fabfe` |  |
| `consulting/03_engine_positioning_and_narrative.md` | included_candidate | 3185 | 97 | `f6465ceab4e4f2ec` |  |
| `consulting/04_customer_and_use_cases.md` | included_candidate | 3186 | 98 | `c996982cd104f5b4` |  |
| `consulting/05_mvp_scope_and_non_goals.md` | included_candidate | 2711 | 96 | `b88bcf6759ccfc8b` |  |
| `consulting/06_runtime_and_product_contract.md` | included_candidate | 2768 | 104 | `b31eec8c1a01835e` |  |
| `consulting/07_current_state_and_roadmap.md` | included_candidate | 3386 | 110 | `a5f8707aaa8d7a0b` |  |
| `consulting/08_risks_and_open_decisions.md` | included_candidate | 3216 | 88 | `7676dcf460ec8fe2` |  |
| `consulting/09_consulting_questionnaire.md` | included_candidate | 2404 | 81 | `eff3d084284b5cac` |  |
| `docs/.DS_Store` | binary_or_non_utf8 | 8196 | 0 | `336fbd0a01bf0976` | not bundled as text |
| `docs/adr-001-migration-versioning.md` | included_candidate | 1605 | 39 | `6f61350b399be039` |  |
| `docs/adr-002-embedding-dimension-policy.md` | included_candidate | 2088 | 39 | `0c358225fbf5a301` |  |
| `docs/adr-003-session-scratch-storage.md` | included_candidate | 1383 | 30 | `4e71afaebe507022` |  |
| `docs/adr-004-memory-semantic-retrieval.md` | included_candidate | 1784 | 39 | `697e3c4148948c7e` |  |
| `docs/adr-005-artifact-class-timing.md` | included_candidate | 2281 | 55 | `3dafe428274498a0` |  |
| `docs/adr-006-db-driver.md` | included_candidate | 1347 | 31 | `ef907812581b4077` |  |
| `docs/adr-007-http-router.md` | included_candidate | 1230 | 29 | `23cdf0ad466505ef` |  |
| `docs/adr-008-package-layout.md` | included_candidate | 3754 | 79 | `f334d48142abbce7` |  |
| `docs/adr-009-updates-edge-lineage-guard.md` | included_candidate | 2670 | 72 | `3a7e868ad1ef86ca` |  |
| `docs/code-header-policy.md` | included_candidate | 2161 | 74 | `49b48ba23e15880f` |  |
| `docs/open-source-code-policy.md` | included_candidate | 1477 | 37 | `df4be8bfa6c27e81` |  |
| `docs/review-packets/00-workpack-03-review-index.md` | included_candidate | 1058 | 21 | `134e77657bfee80d` |  |
| `docs/review-packets/agent-a-blocked-job-recovery.md` | included_candidate | 3425 | 52 | `2900ffe9c062222b` |  |
| `docs/review-packets/agent-b-stage2-store-sources.md` | included_candidate | 5549 | 115 | `c3a967ea4e0d099f` |  |
| `docs/review-packets/agent-c-update-memory-lineage-spec.md` | included_candidate | 16550 | 261 | `29607cc8e203e92b` |  |
| `docs/review-packets/agent-d-contract-gates.md` | included_candidate | 2897 | 67 | `7f2d14ca3db8d200` |  |
| `docs/review-packets/codex-bridge-two-stage-boundary.md` | included_candidate | 3207 | 74 | `4044338da217923d` |  |
| `docs/review-packets/codex-json-bridge-boundary.md` | included_candidate | 2859 | 67 | `94df1f7d47441330` |  |
| `docs/review-packets/correctmemory-review-and-gettimeline-prep.md` | included_candidate | 8990 | 236 | `a6a1398b63b5ed3f` |  |
| `docs/review-packets/current-state-and-next-agent-handoff.md` | included_candidate | 10917 | 230 | `cee97e891517f9dd` |  |
| `docs/review-packets/explain-memory-scope-guard.md` | included_candidate | 1599 | 47 | `a89aaae40f689935` |  |
| `docs/review-packets/explain-memory-visibility-guard.md` | included_candidate | 2627 | 70 | `086817183748abf6` |  |
| `docs/review-packets/hermes-memory-demo-eval.md` | included_candidate | 2163 | 64 | `31bc6d9644dcdde9` |  |
| `docs/review-packets/hermes-memory-trust-loop-product-pivot.md` | included_candidate | 2626 | 79 | `723fde81f272022c` |  |
| `docs/review-packets/hermes-provider-tool-dispatch.md` | included_candidate | 1827 | 52 | `f9a215c401f359ab` |  |
| `docs/review-packets/mcp-tool-input-schemas.md` | included_candidate | 1936 | 58 | `82323ab45b145740` |  |
| `docs/review-packets/mcp-trust-surface-delegation-tests.md` | included_candidate | 1376 | 41 | `30f96f08849f3ead` |  |
| `docs/review-packets/mock-codex-bridge-worker-wiring.md` | included_candidate | 1816 | 56 | `8a9d8891f05e9458` |  |
| `docs/review-packets/next-agent-integration-fixes.md` | included_candidate | 4260 | 85 | `3c48a84a8ab5dedf` |  |
| `docs/review-packets/next-agent-scope-safe-stage2-sources.md` | included_candidate | 3887 | 81 | `9c7da18dd0e58edb` |  |
| `docs/review-packets/operator-visible-degraded-recall-freshness.md` | included_candidate | 2784 | 70 | `ab771dde3969d163` |  |
| `docs/review-packets/recall-preview-metadata-eval.md` | included_candidate | 1927 | 54 | `49110a42f986de47` |  |
| `docs/review-packets/recall-preview-trust-metadata.md` | included_candidate | 2347 | 66 | `523dd2e0401f80fc` |  |
| `docs/review-packets/stage2-actor-bundle-validation.md` | included_candidate | 2139 | 57 | `83ab96a75e933fa3` |  |
| `docs/review-packets/team-1-graph-apply.md` | included_candidate | 5379 | 105 | `d5f79f8d2a57c5ba` |  |
| `docs/review-packets/team-2-reasoning-envelope.md` | included_candidate | 5050 | 121 | `3340266f32330042` |  |
| `docs/review-packets/team-3-worker-reliability.md` | included_candidate | 5370 | 69 | `c27d120974223fd6` |  |
| `docs/review-packets/team-coordination-log.md` | included_candidate | 2478 | 39 | `3a302f3aaf882431` |  |
| `go.mod` | included_candidate | 522 | 20 | `1efd3dba3b79f946` |  |
| `go.sum` | included_candidate | 6584 | 77 | `b3c0e3c91a88d001` |  |
| `internal/config/config.go` | included_candidate | 2667 | 94 | `7f8c2c953c20ec6c` |  |
| `internal/core/doc.go` | included_candidate | 700 | 16 | `78a8a9146f43a578` |  |
| `internal/core/document.go` | included_candidate | 2068 | 50 | `75e4b417e32a7cbf` |  |
| `internal/core/dreaming.go` | included_candidate | 3287 | 81 | `ccc313acb4fda789` |  |
| `internal/core/dto.go` | included_candidate | 13099 | 339 | `030f5bbfc21cab80` |  |
| `internal/core/entity.go` | included_candidate | 1171 | 32 | `ef308696a68a6bfd` |  |
| `internal/core/errors.go` | included_candidate | 1297 | 32 | `e9368e3ff71d9a4a` |  |
| `internal/core/group.go` | included_candidate | 1200 | 34 | `317648488f4a0e3b` |  |
| `internal/core/job.go` | included_candidate | 3133 | 71 | `aeee01a0b530e277` |  |
| `internal/core/kind.go` | included_candidate | 5271 | 115 | `89b01eb4bd7f1f10` |  |
| `internal/core/memory.go` | included_candidate | 3839 | 83 | `eb74a58ad1cda330` |  |
| `internal/core/note.go` | included_candidate | 1222 | 32 | `9624b0400f63cd00` |  |
| `internal/core/plan.go` | included_candidate | 1652 | 45 | `1e504a1d49b3ced5` |  |
| `internal/core/profile.go` | included_candidate | 1144 | 31 | `1119cd5e4fd3c748` |  |
| `internal/core/raw_event.go` | included_candidate | 1381 | 36 | `29e43a0eea44e2af` |  |
| `internal/core/scope.go` | included_candidate | 1275 | 29 | `c565b3a83e0b77eb` |  |
| `internal/core/service.go` | included_candidate | 1769 | 34 | `f998808e3a6d6238` |  |
| `internal/core/service_test.go` | included_candidate | 3659 | 120 | `acc8d58cea98b1e5` |  |
| `internal/core/session.go` | included_candidate | 1157 | 30 | `b66cd6dc0865d6cb` |  |
| `internal/db/pool.go` | included_candidate | 1778 | 54 | `6eb5f49d25779769` |  |
| `internal/embed/doc.go` | included_candidate | 765 | 16 | `8f44d8fa17f9fda3` |  |
| `internal/eval/demo.go` | included_candidate | 12764 | 337 | `44b821210a6c0815` |  |
| `internal/eval/demo_test.go` | included_candidate | 1350 | 45 | `f4c09398d1a1731b` |  |
| `internal/eval/golden.go` | included_candidate | 16020 | 451 | `2a2536d309263c58` |  |
| `internal/eval/golden_test.go` | included_candidate | 4463 | 151 | `edd99763b39b0c64` |  |
| `internal/eval/graph_replay.go` | included_candidate | 21379 | 626 | `a6ac5712d6a963e0` |  |
| `internal/eval/worker_backlog.go` | included_candidate | 19985 | 599 | `32a9019d93d3ed65` |  |
| `internal/graph/apply.go` | included_candidate | 11962 | 314 | `88a0b2fc50cb1369` |  |
| `internal/graph/apply_test.go` | included_candidate | 7743 | 260 | `dc45349205816ceb` |  |
| `internal/graph/doc.go` | included_candidate | 805 | 16 | `8832ab33000f7d49` |  |
| `internal/graph/dreaming.go` | included_candidate | 7752 | 244 | `bed8d68de0f20863` |  |
| `internal/graph/dreaming_test.go` | included_candidate | 4929 | 150 | `561304430d3453fd` |  |
| `internal/graph/store_apply.go` | included_candidate | 9899 | 255 | `9cf8c2ab390ac3ca` |  |
| `internal/graph/store_apply_test.go` | included_candidate | 14801 | 382 | `55a838bbb63fe671` |  |
| `internal/hermes/doc.go` | included_candidate | 769 | 16 | `d5fd2d580b30889f` |  |
| `internal/hermes/provider.go` | included_candidate | 6682 | 181 | `57d05fe04a287d1e` |  |
| `internal/hermes/provider_test.go` | included_candidate | 10507 | 274 | `315a530430c500d8` |  |
| `internal/httpapi/router.go` | included_candidate | 11550 | 417 | `e4817702894f30ef` |  |
| `internal/httpapi/router_test.go` | included_candidate | 11633 | 304 | `89f5d8be3435d63b` |  |
| `internal/ingest/doc.go` | included_candidate | 786 | 16 | `f59d228ec9f4c8f5` |  |
| `internal/ingest/service.go` | included_candidate | 6599 | 216 | `902c77ef6248ec9e` |  |
| `internal/ingest/service_test.go` | included_candidate | 5981 | 205 | `77803e93b205356b` |  |
| `internal/kernel/doc.go` | included_candidate | 804 | 16 | `79b40337a919e494` |  |
| `internal/kernel/service.go` | included_candidate | 23991 | 739 | `07000eff9f8d3ade` |  |
| `internal/kernel/service_test.go` | included_candidate | 18655 | 560 | `5bb741f0b98b04c5` |  |
| `internal/mcp/doc.go` | included_candidate | 756 | 16 | `999ee2bb757e8a9d` |  |
| `internal/mcp/protocol.go` | included_candidate | 15442 | 426 | `921203ceb6bd6f45` |  |
| `internal/mcp/protocol_test.go` | included_candidate | 7232 | 180 | `c284eed301d1b8bd` |  |
| `internal/mcp/surface.go` | included_candidate | 4457 | 118 | `884e24371a261ba5` |  |
| `internal/mcp/surface_test.go` | included_candidate | 8166 | 220 | `9bcfbc5d30cdf32b` |  |
| `internal/reasoning/codex_bridge.go` | included_candidate | 10052 | 264 | `6b527b1e7886d32e` |  |
| `internal/reasoning/codex_bridge_test.go` | included_candidate | 8015 | 262 | `36dd4168198b0b7c` |  |
| `internal/reasoning/contracts.go` | included_candidate | 6989 | 163 | `d682199eaaaae999` |  |
| `internal/reasoning/doc.go` | included_candidate | 814 | 16 | `c2f6ca806d819a48` |  |
| `internal/reasoning/mock_codex_client.go` | included_candidate | 3898 | 97 | `c2b14050b8b6b166` |  |
| `internal/reasoning/orchestrator.go` | included_candidate | 8558 | 236 | `7d5afb1c8e60ef31` |  |
| `internal/reasoning/orchestrator_test.go` | included_candidate | 7724 | 242 | `820b735f1b6b22d5` |  |
| `internal/reasoning/stage2_input_preparer.go` | included_candidate | 7350 | 220 | `caf0f05162d3134b` |  |
| `internal/reasoning/stage2_input_preparer_test.go` | included_candidate | 7663 | 212 | `437bf9fa6704fefc` |  |
| `internal/recall/assembler.go` | included_candidate | 19005 | 720 | `01d9ebe708bd528b` |  |
| `internal/recall/assembler_test.go` | included_candidate | 18185 | 552 | `3f556040623b812a` |  |
| `internal/recall/doc.go` | included_candidate | 802 | 16 | `93e9a6d62beac4d7` |  |
| `internal/recall/freshness.go` | included_candidate | 4483 | 146 | `80255c896c486b53` |  |
| `internal/store/postgres/concurrency_integration_test.go` | included_candidate | 8877 | 263 | `91a375d4ebc66513` |  |
| `internal/store/postgres/corrections.go` | included_candidate | 4579 | 122 | `00e0504c3c1904b2` |  |
| `internal/store/postgres/corrections_test.go` | included_candidate | 2869 | 79 | `7ca2ef4d479242e1` |  |
| `internal/store/postgres/documents.go` | included_candidate | 5611 | 166 | `54a5d2e39304c3f6` |  |
| `internal/store/postgres/documents_test.go` | included_candidate | 3155 | 89 | `a009591bdff366eb` |  |
| `internal/store/postgres/dreaming.go` | included_candidate | 8159 | 241 | `e1e683beb21ad07e` |  |
| `internal/store/postgres/dreaming_test.go` | included_candidate | 2957 | 87 | `4f07fc03d07611cb` |  |
| `internal/store/postgres/groups.go` | included_candidate | 4009 | 114 | `e1afbfa43581d7ee` |  |
| `internal/store/postgres/helpers.go` | included_candidate | 1304 | 52 | `a78a500ecf85ea17` |  |
| `internal/store/postgres/jobs.go` | included_candidate | 14440 | 467 | `b518243b8ea5aa24` |  |
| `internal/store/postgres/jobs_test.go` | included_candidate | 13185 | 435 | `35302f2b62beda48` |  |
| `internal/store/postgres/memories.go` | included_candidate | 20855 | 604 | `2160722111596dc8` |  |
| `internal/store/postgres/memories_test.go` | included_candidate | 6018 | 207 | `0dbecbace6b9a184` |  |
| `internal/store/postgres/notes_plans.go` | included_candidate | 7979 | 253 | `94abac90df49db52` |  |
| `internal/store/postgres/notes_plans_test.go` | included_candidate | 3881 | 122 | `947b68124da0336d` |  |
| `internal/store/postgres/profiles_summaries.go` | included_candidate | 4418 | 120 | `6f4646bf46791812` |  |
| `internal/store/postgres/raw_events.go` | included_candidate | 3361 | 101 | `9b50ae7943a247da` |  |
| `internal/store/postgres/search.go` | included_candidate | 3914 | 107 | `5d3a404273b22cfe` |  |
| `internal/store/postgres/search_test.go` | included_candidate | 2764 | 76 | `fa35492f854bd3eb` |  |
| `internal/store/postgres/store.go` | included_candidate | 1624 | 48 | `498264d80b8d1302` |  |
| `internal/store/postgres/timeline.go` | included_candidate | 3649 | 108 | `3afa25ff6b455ce3` |  |
| `internal/store/postgres/timeline_test.go` | included_candidate | 2386 | 79 | `6a9c29e82c10aded` |  |
| `internal/store/store.go` | included_candidate | 7299 | 144 | `1ff7387f4b7b2c16` |  |
| `internal/worker/doc.go` | included_candidate | 785 | 16 | `7a74e678aaa8e353` |  |
| `internal/worker/processor.go` | included_candidate | 14608 | 435 | `ea025d0b2546ebae` |  |
| `internal/worker/processor_test.go` | included_candidate | 24166 | 749 | `5f81fa4269df391a` |  |
| `internal/worker/stage2_sources.go` | included_candidate | 11172 | 362 | `8c3893a7c55a1e95` |  |
| `internal/worker/stage2_sources_test.go` | included_candidate | 19581 | 516 | `793cbc97a9123b29` |  |
| `migrations/000001_create_pgvector_extension.down.sql` | included_candidate | 33 | 2 | `92a6c13145ed5919` |  |
| `migrations/000001_create_pgvector_extension.up.sql` | included_candidate | 39 | 2 | `9e9b2cfec47519f4` |  |
| `migrations/000002_create_core_tables.down.sql` | included_candidate | 553 | 17 | `63920f9c5c460dd5` |  |
| `migrations/000002_create_core_tables.up.sql` | included_candidate | 9907 | 307 | `89c179f9dc95742b` |  |
| `migrations/000003_add_vector_columns.down.sql` | included_candidate | 390 | 12 | `49c02fb4bb2e4275` |  |
| `migrations/000003_add_vector_columns.up.sql` | included_candidate | 813 | 16 | `e3dd3c20d81b2074` |  |
| `plans/.DS_Store` | binary_or_non_utf8 | 6148 | 0 | `3b931e8cab30fa1c` | not bundled as text |
| `plans/00_read-this-first_for-building-agents.md` | included_candidate | 3828 | 115 | `abdfa73595abcf13` |  |
| `plans/01_rfp_vibegravity_hermes-first.md` | included_candidate | 7227 | 261 | `2259948093e592d3` |  |
| `plans/02_product-contract_and_direction.md` | included_candidate | 6381 | 202 | `d67888ee95e0e646` |  |
| `plans/03_target-architecture_codex-first.md` | included_candidate | 6125 | 240 | `af43da36dbec6f26` |  |
| `plans/04_memory-scopes_dreaming_ontology-lite.md` | included_candidate | 6231 | 267 | `6a6f26cbce677cd2` |  |
| `plans/05_runtime-contracts_ingest-recall-apply.md` | included_candidate | 18569 | 473 | `d75f3b200af62903` |  |
| `plans/06_data-model_and_storage-invariants.md` | included_candidate | 8033 | 291 | `45d797a6cb6bfb74` |  |
| `plans/07_workpack_foundation-and-repo-setup.md` | included_candidate | 2633 | 103 | `b2035c766efc17b2` |  |
| `plans/08_workpack_ingest-and-recall.md` | included_candidate | 1553 | 87 | `8804890f61f1b5fd` |  |
| `plans/09_workpack_memory-graph-and-dreaming.md` | included_candidate | 2479 | 102 | `b1657225d9b13733` |  |
| `plans/10_workpack_hermes-provider-and-external-surfaces.md` | included_candidate | 5453 | 174 | `7cca4659041a94e7` |  |
| `plans/11_workpack_quality-ops-and-evals.md` | included_candidate | 9056 | 205 | `5b8baf181e13bc04` |  |
| `plans/12_agent-coding_playbook_codex-claude.md` | included_candidate | 5598 | 204 | `c5035044a265ad70` |  |
| `plans/13_handoff-prompts_and_response-templates.md` | included_candidate | 2798 | 166 | `7f0d08b5ee5de526` |  |
| `plans/14_source-notes.md` | included_candidate | 2194 | 79 | `facd25249c7aa065` |  |
| `plans/README.md` | included_candidate | 3750 | 74 | `0adb219f34b4e250` |  |
| `plans/templates/AGENTS.md` | included_candidate | 1576 | 58 | `bd0bba2b77f82670` |  |
| `plans/templates/CLAUDE.md` | included_candidate | 1159 | 45 | `0a89c51269c4e23b` |  |
| `plans/templates/PLANS.md` | included_candidate | 290 | 35 | `9bb11cdd5dcf481b` |  |
| `plans/templates/RFP_RESPONSE_TEMPLATE.md` | included_candidate | 414 | 36 | `4d6f5b479ff6e131` |  |
| `plans/templates/SKILL_contract-check.md` | included_candidate | 793 | 35 | `92c5afcb76df0f97` |  |
| `plans/templates/SKILL_eval-regression.md` | included_candidate | 585 | 31 | `90785c6e2a400a83` |  |
| `plans/templates/SKILL_plan-implement-verify.md` | included_candidate | 636 | 31 | `3494086970c08efd` |  |
| `plans/templates/code-header-devlog-go.md` | included_candidate | 907 | 27 | `96de9f5c4856a91b` |  |
| `plans/templates/code-header-minimal-go.md` | included_candidate | 791 | 19 | `1dcf5d4de94ffd43` |  |
| `plans/templates/code-header-narrative-go.md` | included_candidate | 891 | 30 | `a7a84ea97f32f8a4` |  |
| `tests/baseline_test.go` | included_candidate | 2263 | 80 | `e273f196c7b9bb48` |  |
| `tests/golden/replay_eval.json` | included_candidate | 20565 | 609 | `397e4c1ae10655e0` |  |
| `tests/migration_contract_test.go` | included_candidate | 4572 | 121 | `20da94220156d03a` |  |
| `tools/headercheck/main.go` | included_candidate | 4298 | 180 | `1c1c9f51abff5a0f` |  |
