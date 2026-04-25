# 01 Gpt Pro Context And Reading Order

Generated: 2026-04-25

This file is part of the GPT-Pro review material bundle for VibeGravity.

## Included Sources

- `.agents/coordination/.gitignore`
- `.agents/coordination/PROMPT_SNIPPET.md`
- `.agents/coordination/README.md`
- `.agents/coordination/UNIVERSAL_AGENT_PROMPT.md`
- `.agents/coordination/WORK_PROGRESS.md`
- `.agents/coordination/activity.log`
- `.agents/coordination/agent-work.sh`
- `.agents/coordination/claims.tsv`
- `.agents/hermes-orchestration/README.md`
- `.agents/hermes-orchestration/collect.sh`
- `.agents/hermes-orchestration/dispatch.sh`
- `.agents/hermes-orchestration/run-agent.sh`
- `.agents/hermes-orchestration/status.sh`
- `.agents/hermes-orchestration/tasks/20260424-eval-graph-gates/bottega.md`
- `.agents/hermes-orchestration/tasks/20260424-eval-graph-gates/default.md`
- `.agents/hermes-orchestration/tasks/20260424-eval-graph-gates/manifest.tsv`
- `.agents/hermes-orchestration/tasks/20260424-eval-graph-gates/vuitton.md`
- `.agents/hermes-orchestration/tasks/20260424_backlog_metrics/bottega.md`
- `.agents/hermes-orchestration/tasks/20260424_backlog_metrics/default.md`
- `.agents/hermes-orchestration/tasks/20260424_backlog_metrics/manifest.tsv`
- `.agents/hermes-orchestration/tasks/20260424_backlog_metrics/vuitton.md`
- `.agents/hermes-orchestration/tasks/smoke/bottega.md`
- `.agents/hermes-orchestration/tasks/smoke/default.md`
- `.agents/hermes-orchestration/tasks/smoke/manifest.tsv`
- `.agents/hermes-orchestration/tasks/smoke/vuitton.md`
- `.agents/hermes-orchestration/tasks/v1-update-memory/bottega.md`
- `.agents/hermes-orchestration/tasks/v1-update-memory/default.md`
- `.agents/hermes-orchestration/tasks/v1-update-memory/manifest.tsv`
- `.agents/hermes-orchestration/tasks/v1-update-memory/vuitton.md`
- `.agents/skills/code-headers.md`
- `.agents/skills/contract-check.md`
- `.agents/skills/eval-regression.md`
- `.agents/skills/plan-implement-verify.md`
- `.agents/skills/source-provenance.md`
- `.gitignore`
- `.gitmessage.txt`
- `.golangci.yml`
- `AGENTS.md`
- `CLAUDE.md`
- `COMMIT_MESSAGE_RULES.md`
- `Makefile`
- `PLANS.md`
- `go.mod`
- `go.sum`

## Source Contents


<!-- Source: .agents/coordination/.gitignore | bytes=59 | lines=6 | sha16=43a7857fbc3c7b56 -->

```text
/WORK_PROGRESS.md
/activity.log
/claims.tsv
/.lock/
/*.tmp

```



<!-- Source: .agents/coordination/PROMPT_SNIPPET.md | bytes=1039 | lines=29 | sha16=b2438dc8817f7a23 -->

````md
# Multi-Agent Coordination Snippet

Paste this into every parallel Codex, Hermes, Claude, or review-agent prompt
that may edit this repo.

For a fully autonomous agent that should decide its own next useful task, use
`.agents/coordination/UNIVERSAL_AGENT_PROMPT.md` instead.

If the operator gives only that universal prompt path, read it and execute it.
Do not ask what to do with the file.

```text
Before editing any file, read:

- /Users/parker/Documents/VibeGravity/.agents/coordination/WORK_PROGRESS.md

Then claim the exact files you intend to edit:

/Users/parker/Documents/VibeGravity/.agents/coordination/agent-work.sh claim "<agent-id>" "<short task>" <file> [<file> ...]

Rules:

- Do not edit a file claimed by another active agent.
- Claim exact file paths before opening a new write surface.
- Send a heartbeat before widening scope or after a long debugging pass.
- Release files immediately when done with them, before moving to other files.
- Finish with `done` only after verification and result notes are complete.
```

````



<!-- Source: .agents/coordination/README.md | bytes=3336 | lines=111 | sha16=efbbde2bb0a45f9b -->

````md
# Multi-Agent Coordination

This directory is the repo-local coordination surface for concurrent Codex,
Hermes, Claude, and reviewer agents working in VibeGravity.

The live shared progress file is:

```text
.agents/coordination/WORK_PROGRESS.md
```

It is generated from `.agents/coordination/claims.tsv` and
`.agents/coordination/activity.log`. Those live state files are intentionally
git-ignored so normal coding diffs do not fill up with heartbeat noise.

## Required Loop

Every agent must follow this loop before editing files:

1. Read `.agents/coordination/WORK_PROGRESS.md`.
2. Claim the exact files to edit before modifying them.
3. Stop if another active agent already owns any of those files.
4. Send heartbeats during long work or before widening scope.
5. Release a file immediately when work on that file is done.
6. Run verification and leave a result note or review packet when the lane ends.

Do not claim broad globs such as `internal/**`. Claim concrete file paths.

## Universal Prompt

Use `.agents/coordination/UNIVERSAL_AGENT_PROMPT.md` when launching a new
autonomous agent and you want it to decide whether to implement, test, review,
document, or hand off the next useful slice.

If an operator gives an agent only this path, the expected behavior is to read
the file and execute it immediately:

```text
/Users/parker/Documents/VibeGravity/.agents/coordination/UNIVERSAL_AGENT_PROMPT.md
```

The agent should not ask what to do with that path.

Use `.agents/coordination/PROMPT_SNIPPET.md` when you already have a focused
task prompt and only need to inject the shared-file claim rules.

## Commands

Initialize or refresh the shared progress file:

```bash
.agents/coordination/agent-work.sh init
```

View current claims and the recent activity log:

```bash
.agents/coordination/agent-work.sh status
```

Claim files before editing:

```bash
.agents/coordination/agent-work.sh claim codex-main "add recall tests" internal/recall/assembler.go internal/recall/assembler_test.go
```

Send a heartbeat while continuing the same lane:

```bash
.agents/coordination/agent-work.sh heartbeat codex-main "tests failing on token budget case; still in recall files"
```

Release files immediately after finishing with them:

```bash
.agents/coordination/agent-work.sh release codex-main internal/recall/assembler.go internal/recall/assembler_test.go
```

Mark an agent lane done and release all files owned by that agent:

```bash
.agents/coordination/agent-work.sh done codex-main "implemented recall budget fix; go test ./internal/recall passed"
```

## Collision Rules

- If `claim` reports a conflict, do not edit that file.
- If two agents both need one hot file, one agent should finish and release it
  first, or the coordinator should split the lane differently.
- Result docs under `docs/review-packets/` should still be used for completed
  work. `WORK_PROGRESS.md` is the live lock board, not the final handoff.
- Manual edits to `WORK_PROGRESS.md` are fallback-only. Prefer the script so
  the live file, claim table, and activity log stay consistent.

## Agent IDs

Use stable, human-readable IDs:

- `codex-main`
- `codex-reviewer`
- `hermes-default`
- `hermes-vuitton`
- `hermes-bottega`

For short ad hoc lanes, append the purpose:

```text
codex-main-recall-budget
hermes-default-stage2-review
```

````



<!-- Source: .agents/coordination/UNIVERSAL_AGENT_PROMPT.md | bytes=6279 | lines=136 | sha16=6958f10612a6c23a -->

````md
# Universal VibeGravity Agent Prompt

This is an execution prompt, not a question or a file to summarize.

If you are an agent reading this file, start working immediately. Do not ask
what to do with this file. Do not ask whether to review, update, coordinate, or
prepare a handoff. This file is the operator's instruction to autonomously begin
work in the VibeGravity repo.

If the operator message was only this file path, treat that path as an instruction
to read and execute this prompt now:

```text
/Users/parker/Documents/VibeGravity/.agents/coordination/UNIVERSAL_AGENT_PROMPT.md
```

Use this prompt for any new Codex, Hermes, Claude, reviewer, or implementation
agent that should autonomously find useful work in this repo.

```text
BEGIN NOW.

You are an autonomous VibeGravity engineering agent working in:

/Users/parker/Documents/VibeGravity

Your job is to decide what useful work should happen next, coordinate with the
other active agents, do the work, verify it, and leave a clear handoff. You must
judge whether the right next action is implementation, tests, docs, review,
verification, or a blocker-driven handoff. Do not wait for the operator to
classify the task unless the next step would be destructive, ambiguous in a way
that risks data loss, or blocked by another active agent.

Do not reply with any of these:

- "What would you like me to do with this file?"
- "I am ready to review it, update it, use it, or prepare a handoff."
- "What would you like me to do next?"

Those responses are failures. Instead, begin the startup sequence below.

Start every run exactly like this:

1. Read /Users/parker/Documents/VibeGravity/AGENTS.md.
2. Read /Users/parker/Documents/VibeGravity/.agents/coordination/WORK_PROGRESS.md.
3. Run:
   /Users/parker/Documents/VibeGravity/.agents/coordination/agent-work.sh status
4. Read the current planning and review surfaces that match the work:
   - /Users/parker/Documents/VibeGravity/PLANS.md
   - /Users/parker/Documents/VibeGravity/plans/00_read-this-first_for-building-agents.md
   - /Users/parker/Documents/VibeGravity/plans/01_rfp_vibegravity_hermes-first.md
   - /Users/parker/Documents/VibeGravity/plans/02_product-contract_and_direction.md
   - /Users/parker/Documents/VibeGravity/plans/03_target-architecture_codex-first.md
   - /Users/parker/Documents/VibeGravity/plans/05_runtime-contracts_ingest-recall-apply.md
   - /Users/parker/Documents/VibeGravity/plans/06_data-model_and_storage-invariants.md
   - Relevant files under /Users/parker/Documents/VibeGravity/docs/review-packets/

Coordination rules:

- Before editing any file, claim the exact file paths you intend to modify:
  /Users/parker/Documents/VibeGravity/.agents/coordination/agent-work.sh claim "<agent-id>" "<short task>" <file> [<file> ...]
- Use a stable agent id such as codex-main, codex-reviewer, hermes-default,
  hermes-vuitton, hermes-bottega, or a short task-specific variant.
- Do not edit a file claimed by another active agent.
- If a claim is rejected, choose a non-overlapping useful lane instead:
  review, tests, docs, a result packet, or a smaller implementation slice.
- Claim concrete files only. Do not claim broad globs such as internal/**.
- Send a heartbeat before widening scope or after a long debugging pass:
  /Users/parker/Documents/VibeGravity/.agents/coordination/agent-work.sh heartbeat "<agent-id>" "<current status>"
- Release files immediately when finished with them:
  /Users/parker/Documents/VibeGravity/.agents/coordination/agent-work.sh release "<agent-id>" <file> [<file> ...]
- Finish by marking the lane done:
  /Users/parker/Documents/VibeGravity/.agents/coordination/agent-work.sh done "<agent-id>" "<summary and verification>"

Decision rules:

- If the operator gave a specific task, do that task within the active repo
  contracts and coordination rules.
- If no specific task was given, pick the smallest high-value unclaimed slice
  from PLANS.md and the current review packets.
- Prefer work that advances the Hermes Memory trust loop: recall preview,
  explain/timeline, correction, supersession, visible scope/provenance, degraded
  freshness truthfulness, Hermes MCP/provider integration, evals, and operator
  visibility.
- If implementation is safe and unclaimed, implement code and tests.
- If code ownership is contested or risky, perform a focused review with file
  and line findings, or write a blocker-driven next-agent handoff.
- If behavior changed, update the relevant docs or review packet.
- If a new architecture direction is required, stop and propose an ADR instead
  of hiding it inside code.
- Keep changes narrow. Do not clean up unrelated files.

VibeGravity invariants you must preserve:

- Raw events and derived memories stay separate.
- All write paths are idempotent.
- Every memory has explicit scope and provenance.
- agent_private, workspace_shared, and group_shared visibility must not blur.
- group_shared memory requires valid membership.
- Recall must be budget-aware and expose trust metadata where possible.
- Reasoning output remains schema-first structured JSON.
- The default worker path is local embeddings -> retrieval -> Codex Stage 1 ->
  Codex Stage 2 -> apply engine.
- Do not reintroduce local extractor dependence into the main path.
- Keep real Codex disabled by default unless the repo explicitly enables a safe
  configuration and failure mode.

Verification rules:

- Run focused tests for the files you changed.
- Before handoff, run as much of the repo gate as is appropriate for the change:
  go test ./...
  make lint
  make check-headers
  git diff --check
  make eval
- If a gate cannot run, state exactly why and what remains unverified.
- Review your own diff before reporting completion.

Required final handoff:

- Summary
- What you changed or reviewed
- Files changed or files reviewed
- Tests and checks run
- Remaining risks or blockers
- Source Review: estimated source, suspected license, similarity risk, and
  whether human review is required

When writing code, use original implementation from first principles. Do not
copy GPL, AGPL, LGPL, SSPL, Elastic License, or unknown-license code.

Never claim the work is done while you still own active file claims. Release or
mark done in the coordination board first.
```

````



<!-- Source: .agents/coordination/WORK_PROGRESS.md | bytes=11766 | lines=55 | sha16=e5d961ec411e9820 -->

```md
# Agent Work Progress

Live shared progress board for concurrent VibeGravity agents.

Generated by `.agents/coordination/agent-work.sh`. Prefer the script over manual edits.

Last rendered: 2026-04-25T06:58:15Z

## Active Claims

No active claims.

## Recent Activity

| Time | Action | Agent | Files | Note |
|---|---|---|---|---|
| 2026-04-25T06:33:30Z | done | codex-main | .agents/coordination/UNIVERSAL_AGENT_PROMPT.md, .agents/coordination/README.md, .agents/coordination/PROMPT_SNIPPET.md, AGENTS.md | universal prompt made self-starting; AGENTS dispatch rule added; diff-check passed |
| 2026-04-25T06:35:38Z | claim | codex-main-running-freshness | internal/core/job.go, internal/store/postgres/jobs.go, internal/store/postgres/jobs_test.go, internal/recall/freshness.go, internal/recall/assembler_test.go, cmd/cli/main.go, cmd/cli/main_test.go, PLANS.md, docs/review-packets/operator-visible-degraded-recall-freshness.md | surface running job freshness lag |
| 2026-04-25T06:36:23Z | claim | codex-main-explain-scope | internal/store/postgres/memories.go, internal/store/postgres/memories_test.go, docs/review-packets/explain-memory-scope-guard.md | scope-bound explain memory provenance |
| 2026-04-25T06:36:48Z | claim | codex-main-mcp-trust-tests | internal/mcp/surface_test.go, docs/review-packets/mcp-trust-surface-delegation-tests.md | lock MCP trust surface delegation |
| 2026-04-25T06:36:51Z | claim | codex-main-mcp-schemas | internal/mcp/protocol.go, internal/mcp/protocol_test.go, plans/10_workpack_hermes-provider-and-external-surfaces.md, docs/review-packets/mcp-tool-input-schemas.md | add MCP tool input schemas |
| 2026-04-25T06:37:23Z | done | codex-universal-router |  | startup completed; found active agents already own freshness, explain-scope, and MCP schema/trust-test lanes; no files edited to avoid claim conflicts; reviewed AGENTS, PLANS, required plans, review packets, coordination status |
| 2026-04-25T06:37:23Z | heartbeat | codex-main-running-freshness | internal/core/job.go, internal/store/postgres/jobs.go, internal/store/postgres/jobs_test.go, internal/recall/freshness.go, internal/recall/assembler_test.go, cmd/cli/main.go, cmd/cli/main_test.go, PLANS.md, docs/review-packets/operator-visible-degraded-recall-freshness.md | patch added for running-job backlog freshness; formatting and focused tests next |
| 2026-04-25T06:37:25Z | heartbeat | codex-main-explain-scope | internal/store/postgres/memories.go, internal/store/postgres/memories_test.go, docs/review-packets/explain-memory-scope-guard.md | patched explain memory tenant/workspace guards; running focused tests |
| 2026-04-25T06:37:38Z | heartbeat | codex-main-running-freshness | internal/core/job.go, internal/store/postgres/jobs.go, internal/store/postgres/jobs_test.go, internal/recall/freshness.go, internal/recall/assembler_test.go, cmd/cli/main.go, cmd/cli/main_test.go, PLANS.md, docs/review-packets/operator-visible-degraded-recall-freshness.md | focused tests passed; running full repo gates |
| 2026-04-25T06:37:54Z | done | codex-main-mcp-trust-tests | internal/mcp/surface_test.go, docs/review-packets/mcp-trust-surface-delegation-tests.md | added MCP view_timeline/explain_memory delegation tests and review packet; go test ./internal/mcp, go test ./..., make check-headers, git diff --check, make eval passed; make lint currently blocked by active codex-main-mcp-schemas gofmt issue in internal/mcp/protocol.go |
| 2026-04-25T06:38:18Z | done | codex-main-explain-scope | internal/store/postgres/memories.go, internal/store/postgres/memories_test.go, docs/review-packets/explain-memory-scope-guard.md | ExplainMemory tenant/workspace provenance guard added; go test ./internal/store/postgres, go test ./..., make check-headers, git diff --check, make eval passed; make lint blocked by active codex-main-mcp-schemas gofmt issue in internal/mcp/protocol.go |
| 2026-04-25T06:38:21Z | heartbeat | codex-main-mcp-schemas | internal/mcp/protocol.go, internal/mcp/protocol_test.go, plans/10_workpack_hermes-provider-and-external-surfaces.md, docs/review-packets/mcp-tool-input-schemas.md | MCP schema implementation drafted; formatting and focused tests next |
| 2026-04-25T06:38:30Z | done | codex-main-running-freshness | internal/core/job.go, internal/store/postgres/jobs.go, internal/store/postgres/jobs_test.go, internal/recall/freshness.go, internal/recall/assembler_test.go, cmd/cli/main.go, cmd/cli/main_test.go, PLANS.md, docs/review-packets/operator-visible-degraded-recall-freshness.md | running-job freshness lag surfaced; focused tests, go test ./..., make eval, scoped golangci-lint, check-headers, and diff-check passed; full make lint blocked by codex-main-mcp-schemas gofmt issue in internal/mcp/protocol.go |
| 2026-04-25T06:39:09Z | done | codex-main-mcp-schemas | internal/mcp/protocol.go, internal/mcp/protocol_test.go, plans/10_workpack_hermes-provider-and-external-surfaces.md, docs/review-packets/mcp-tool-input-schemas.md | MCP tools/list now advertises concrete input schemas for trust-loop tools; go test ./internal/mcp, go test ./..., make eval, make lint, make check-headers, git diff --check passed |
| 2026-04-25T06:52:34Z | claim | codex-main-explain-visibility | internal/core/dto.go, internal/store/postgres/memories.go, internal/store/postgres/memories_test.go, internal/kernel/service.go, internal/kernel/service_test.go, internal/httpapi/router.go, internal/httpapi/router_test.go, internal/mcp/protocol.go, internal/mcp/protocol_test.go, docs/review-packets/explain-memory-scope-guard.md, docs/review-packets/explain-memory-visibility-guard.md | owner-aware explain memory provenance |
| 2026-04-25T06:53:06Z | claim | codex-recall-metadata-eval | internal/eval/golden.go, internal/eval/golden_test.go, tests/golden/replay_eval.json, docs/review-packets/recall-preview-metadata-eval.md, plans/11_workpack_quality-ops-and-evals.md | assert recall preview trust metadata in golden eval |
| 2026-04-25T06:53:14Z | claim | codex-main-hermes-tool-dispatch | internal/hermes/provider.go, internal/hermes/provider_test.go, plans/10_workpack_hermes-provider-and-external-surfaces.md, docs/review-packets/hermes-provider-tool-dispatch.md | add Hermes provider tool dispatch guard |
| 2026-04-25T06:53:28Z | claim | codex-main-demo-eval | internal/eval/demo.go, internal/eval/demo_test.go, cmd/cli/main.go, cmd/cli/main_test.go, Makefile, docs/review-packets/hermes-memory-demo-eval.md | add Hermes Memory demo eval |
| 2026-04-25T06:54:14Z | claim | codex-main-explain-visibility | internal/mcp/surface_test.go | owner-aware explain memory provenance |
| 2026-04-25T06:55:03Z | heartbeat | codex-recall-metadata-eval | internal/eval/golden.go, internal/eval/golden_test.go, tests/golden/replay_eval.json, docs/review-packets/recall-preview-metadata-eval.md, plans/11_workpack_quality-ops-and-evals.md | metadata expectation patch added; running focused tests |
| 2026-04-25T06:55:22Z | claim | codex-main-demo-eval | internal/eval/graph_replay.go | add Hermes Memory demo eval |
| 2026-04-25T06:55:24Z | done | codex-main-hermes-tool-dispatch | internal/hermes/provider.go, internal/hermes/provider_test.go, plans/10_workpack_hermes-provider-and-external-surfaces.md, docs/review-packets/hermes-provider-tool-dispatch.md | Hermes provider CallTool dispatch added for core-backed trust-loop tools; show_plan explicitly blocked until read-only plan API; go test ./internal/hermes, go test ./..., make lint, make check-headers, make eval, git diff --check passed |
| 2026-04-25T06:55:37Z | heartbeat | codex-main-explain-visibility | internal/core/dto.go, internal/store/postgres/memories.go, internal/store/postgres/memories_test.go, internal/kernel/service.go, internal/kernel/service_test.go, internal/httpapi/router.go, internal/httpapi/router_test.go, internal/mcp/protocol.go, internal/mcp/protocol_test.go, docs/review-packets/explain-memory-scope-guard.md, docs/review-packets/explain-memory-visibility-guard.md, internal/mcp/surface_test.go | patched ExplainMemory actor/group visibility; formatting and focused tests next |
| 2026-04-25T06:55:53Z | heartbeat | codex-recall-metadata-eval | internal/eval/golden.go, internal/eval/golden_test.go, tests/golden/replay_eval.json, docs/review-packets/recall-preview-metadata-eval.md, plans/11_workpack_quality-ops-and-evals.md | focused golden eval passed; full go test/make eval currently blocked by active codex-main-demo-eval duplicate ExplainMemory method |
| 2026-04-25T06:56:29Z | claim | codex-main-explain-visibility | plans/05_runtime-contracts_ingest-recall-apply.md, plans/10_workpack_hermes-provider-and-external-surfaces.md | owner-aware explain memory provenance |
| 2026-04-25T06:57:03Z | done | codex-recall-metadata-eval | internal/eval/golden.go, internal/eval/golden_test.go, tests/golden/replay_eval.json, docs/review-packets/recall-preview-metadata-eval.md, plans/11_workpack_quality-ops-and-evals.md | recall preview block metadata expectations added to golden eval; focused golden eval and check-headers/diff-check passed; full go test/make eval blocked by active codex-main-demo-eval duplicate ExplainMemory method in internal/eval/demo.go vs graph_replay.go |
| 2026-04-25T06:57:14Z | heartbeat | codex-main-demo-eval | internal/eval/demo.go, internal/eval/demo_test.go, cmd/cli/main.go, cmd/cli/main_test.go, Makefile, docs/review-packets/hermes-memory-demo-eval.md, internal/eval/graph_replay.go | demo eval implementation drafted; formatting and focused tests next |
| 2026-04-25T06:57:29Z | heartbeat | codex-main-explain-visibility | internal/core/dto.go, internal/store/postgres/memories.go, internal/store/postgres/memories_test.go, internal/kernel/service.go, internal/kernel/service_test.go, internal/httpapi/router.go, internal/httpapi/router_test.go, internal/mcp/protocol.go, internal/mcp/protocol_test.go, docs/review-packets/explain-memory-scope-guard.md, docs/review-packets/explain-memory-visibility-guard.md, internal/mcp/surface_test.go, plans/05_runtime-contracts_ingest-recall-apply.md, plans/10_workpack_hermes-provider-and-external-surfaces.md | ExplainMemory visibility patch verified; full go test/make eval currently blocked by active codex-main-demo-eval expectation mismatch |
| 2026-04-25T06:57:57Z | done | codex-main-explain-visibility | internal/core/dto.go, internal/store/postgres/memories.go, internal/store/postgres/memories_test.go, internal/kernel/service.go, internal/kernel/service_test.go, internal/httpapi/router.go, internal/httpapi/router_test.go, internal/mcp/protocol.go, internal/mcp/protocol_test.go, docs/review-packets/explain-memory-scope-guard.md, docs/review-packets/explain-memory-visibility-guard.md, internal/mcp/surface_test.go, plans/05_runtime-contracts_ingest-recall-apply.md, plans/10_workpack_hermes-provider-and-external-surfaces.md | ExplainMemory owner/group visibility guard added and documented; focused tests, make lint, make check-headers, git diff --check passed; full go test ./... and make eval blocked by active codex-main-demo-eval expectation mismatch in internal/eval/demo.go |
| 2026-04-25T06:58:15Z | done | codex-main-demo-eval | internal/eval/demo.go, internal/eval/demo_test.go, cmd/cli/main.go, cmd/cli/main_test.go, Makefile, docs/review-packets/hermes-memory-demo-eval.md, internal/eval/graph_replay.go | Hermes Memory demo eval added; cli eval demo and make eval now exercise local trust loop; go test ./internal/eval, go test ./cmd/cli, go test ./..., make eval, make lint, make check-headers, git diff --check passed |

## Protocol

- Read this file before starting and before widening scope.
- Claim exact files before editing.
- Do not edit files claimed by another active agent.
- Heartbeat during long work.
- Release files immediately when finished.

```



<!-- Source: .agents/coordination/activity.log | bytes=12165 | lines=38 | sha16=a0b8ea383977c0b6 -->

```text
2026-04-25T06:18:42Z	claim	codex-main	AGENTS.md, .agents/coordination/README.md, .agents/coordination/PROMPT_SNIPPET.md, .agents/coordination/agent-work.sh	bootstrap multi-agent coordination protocol
2026-04-25T06:19:02Z	claim	codex-main	.agents/hermes-orchestration/README.md	wire Hermes orchestration to coordination protocol
2026-04-25T06:20:50Z	done	codex-main	AGENTS.md, .agents/coordination/README.md, .agents/coordination/PROMPT_SNIPPET.md, .agents/coordination/agent-work.sh, .agents/hermes-orchestration/README.md	multi-agent coordination protocol installed; go test, lint, headers, eval, diff-check passed
2026-04-25T06:25:30Z	claim	codex-main	.agents/coordination/README.md, .agents/coordination/PROMPT_SNIPPET.md, .agents/coordination/UNIVERSAL_AGENT_PROMPT.md	add universal autonomous agent prompt
2026-04-25T06:26:39Z	done	codex-main	.agents/coordination/README.md, .agents/coordination/PROMPT_SNIPPET.md, .agents/coordination/UNIVERSAL_AGENT_PROMPT.md	universal autonomous agent prompt added; docs linked; whitespace/diff checks passed
2026-04-25T06:32:45Z	claim	codex-main	.agents/coordination/UNIVERSAL_AGENT_PROMPT.md, .agents/coordination/README.md, .agents/coordination/PROMPT_SNIPPET.md	make universal prompt self-starting
2026-04-25T06:33:00Z	claim	codex-main	AGENTS.md	add universal prompt dispatch rule
2026-04-25T06:33:30Z	done	codex-main	.agents/coordination/UNIVERSAL_AGENT_PROMPT.md, .agents/coordination/README.md, .agents/coordination/PROMPT_SNIPPET.md, AGENTS.md	universal prompt made self-starting; AGENTS dispatch rule added; diff-check passed
2026-04-25T06:35:38Z	claim	codex-main-running-freshness	internal/core/job.go, internal/store/postgres/jobs.go, internal/store/postgres/jobs_test.go, internal/recall/freshness.go, internal/recall/assembler_test.go, cmd/cli/main.go, cmd/cli/main_test.go, PLANS.md, docs/review-packets/operator-visible-degraded-recall-freshness.md	surface running job freshness lag
2026-04-25T06:36:23Z	claim	codex-main-explain-scope	internal/store/postgres/memories.go, internal/store/postgres/memories_test.go, docs/review-packets/explain-memory-scope-guard.md	scope-bound explain memory provenance
2026-04-25T06:36:48Z	claim	codex-main-mcp-trust-tests	internal/mcp/surface_test.go, docs/review-packets/mcp-trust-surface-delegation-tests.md	lock MCP trust surface delegation
2026-04-25T06:36:51Z	claim	codex-main-mcp-schemas	internal/mcp/protocol.go, internal/mcp/protocol_test.go, plans/10_workpack_hermes-provider-and-external-surfaces.md, docs/review-packets/mcp-tool-input-schemas.md	add MCP tool input schemas
2026-04-25T06:37:23Z	done	codex-universal-router		startup completed; found active agents already own freshness, explain-scope, and MCP schema/trust-test lanes; no files edited to avoid claim conflicts; reviewed AGENTS, PLANS, required plans, review packets, coordination status
2026-04-25T06:37:23Z	heartbeat	codex-main-running-freshness	internal/core/job.go, internal/store/postgres/jobs.go, internal/store/postgres/jobs_test.go, internal/recall/freshness.go, internal/recall/assembler_test.go, cmd/cli/main.go, cmd/cli/main_test.go, PLANS.md, docs/review-packets/operator-visible-degraded-recall-freshness.md	patch added for running-job backlog freshness; formatting and focused tests next
2026-04-25T06:37:25Z	heartbeat	codex-main-explain-scope	internal/store/postgres/memories.go, internal/store/postgres/memories_test.go, docs/review-packets/explain-memory-scope-guard.md	patched explain memory tenant/workspace guards; running focused tests
2026-04-25T06:37:38Z	heartbeat	codex-main-running-freshness	internal/core/job.go, internal/store/postgres/jobs.go, internal/store/postgres/jobs_test.go, internal/recall/freshness.go, internal/recall/assembler_test.go, cmd/cli/main.go, cmd/cli/main_test.go, PLANS.md, docs/review-packets/operator-visible-degraded-recall-freshness.md	focused tests passed; running full repo gates
2026-04-25T06:37:54Z	done	codex-main-mcp-trust-tests	internal/mcp/surface_test.go, docs/review-packets/mcp-trust-surface-delegation-tests.md	added MCP view_timeline/explain_memory delegation tests and review packet; go test ./internal/mcp, go test ./..., make check-headers, git diff --check, make eval passed; make lint currently blocked by active codex-main-mcp-schemas gofmt issue in internal/mcp/protocol.go
2026-04-25T06:38:18Z	done	codex-main-explain-scope	internal/store/postgres/memories.go, internal/store/postgres/memories_test.go, docs/review-packets/explain-memory-scope-guard.md	ExplainMemory tenant/workspace provenance guard added; go test ./internal/store/postgres, go test ./..., make check-headers, git diff --check, make eval passed; make lint blocked by active codex-main-mcp-schemas gofmt issue in internal/mcp/protocol.go
2026-04-25T06:38:21Z	heartbeat	codex-main-mcp-schemas	internal/mcp/protocol.go, internal/mcp/protocol_test.go, plans/10_workpack_hermes-provider-and-external-surfaces.md, docs/review-packets/mcp-tool-input-schemas.md	MCP schema implementation drafted; formatting and focused tests next
2026-04-25T06:38:30Z	done	codex-main-running-freshness	internal/core/job.go, internal/store/postgres/jobs.go, internal/store/postgres/jobs_test.go, internal/recall/freshness.go, internal/recall/assembler_test.go, cmd/cli/main.go, cmd/cli/main_test.go, PLANS.md, docs/review-packets/operator-visible-degraded-recall-freshness.md	running-job freshness lag surfaced; focused tests, go test ./..., make eval, scoped golangci-lint, check-headers, and diff-check passed; full make lint blocked by codex-main-mcp-schemas gofmt issue in internal/mcp/protocol.go
2026-04-25T06:39:09Z	done	codex-main-mcp-schemas	internal/mcp/protocol.go, internal/mcp/protocol_test.go, plans/10_workpack_hermes-provider-and-external-surfaces.md, docs/review-packets/mcp-tool-input-schemas.md	MCP tools/list now advertises concrete input schemas for trust-loop tools; go test ./internal/mcp, go test ./..., make eval, make lint, make check-headers, git diff --check passed
2026-04-25T06:52:34Z	claim	codex-main-explain-visibility	internal/core/dto.go, internal/store/postgres/memories.go, internal/store/postgres/memories_test.go, internal/kernel/service.go, internal/kernel/service_test.go, internal/httpapi/router.go, internal/httpapi/router_test.go, internal/mcp/protocol.go, internal/mcp/protocol_test.go, docs/review-packets/explain-memory-scope-guard.md, docs/review-packets/explain-memory-visibility-guard.md	owner-aware explain memory provenance
2026-04-25T06:53:06Z	claim	codex-recall-metadata-eval	internal/eval/golden.go, internal/eval/golden_test.go, tests/golden/replay_eval.json, docs/review-packets/recall-preview-metadata-eval.md, plans/11_workpack_quality-ops-and-evals.md	assert recall preview trust metadata in golden eval
2026-04-25T06:53:14Z	claim	codex-main-hermes-tool-dispatch	internal/hermes/provider.go, internal/hermes/provider_test.go, plans/10_workpack_hermes-provider-and-external-surfaces.md, docs/review-packets/hermes-provider-tool-dispatch.md	add Hermes provider tool dispatch guard
2026-04-25T06:53:28Z	claim	codex-main-demo-eval	internal/eval/demo.go, internal/eval/demo_test.go, cmd/cli/main.go, cmd/cli/main_test.go, Makefile, docs/review-packets/hermes-memory-demo-eval.md	add Hermes Memory demo eval
2026-04-25T06:54:14Z	claim	codex-main-explain-visibility	internal/mcp/surface_test.go	owner-aware explain memory provenance
2026-04-25T06:55:03Z	heartbeat	codex-recall-metadata-eval	internal/eval/golden.go, internal/eval/golden_test.go, tests/golden/replay_eval.json, docs/review-packets/recall-preview-metadata-eval.md, plans/11_workpack_quality-ops-and-evals.md	metadata expectation patch added; running focused tests
2026-04-25T06:55:22Z	claim	codex-main-demo-eval	internal/eval/graph_replay.go	add Hermes Memory demo eval
2026-04-25T06:55:24Z	done	codex-main-hermes-tool-dispatch	internal/hermes/provider.go, internal/hermes/provider_test.go, plans/10_workpack_hermes-provider-and-external-surfaces.md, docs/review-packets/hermes-provider-tool-dispatch.md	Hermes provider CallTool dispatch added for core-backed trust-loop tools; show_plan explicitly blocked until read-only plan API; go test ./internal/hermes, go test ./..., make lint, make check-headers, make eval, git diff --check passed
2026-04-25T06:55:37Z	heartbeat	codex-main-explain-visibility	internal/core/dto.go, internal/store/postgres/memories.go, internal/store/postgres/memories_test.go, internal/kernel/service.go, internal/kernel/service_test.go, internal/httpapi/router.go, internal/httpapi/router_test.go, internal/mcp/protocol.go, internal/mcp/protocol_test.go, docs/review-packets/explain-memory-scope-guard.md, docs/review-packets/explain-memory-visibility-guard.md, internal/mcp/surface_test.go	patched ExplainMemory actor/group visibility; formatting and focused tests next
2026-04-25T06:55:53Z	heartbeat	codex-recall-metadata-eval	internal/eval/golden.go, internal/eval/golden_test.go, tests/golden/replay_eval.json, docs/review-packets/recall-preview-metadata-eval.md, plans/11_workpack_quality-ops-and-evals.md	focused golden eval passed; full go test/make eval currently blocked by active codex-main-demo-eval duplicate ExplainMemory method
2026-04-25T06:56:29Z	claim	codex-main-explain-visibility	plans/05_runtime-contracts_ingest-recall-apply.md, plans/10_workpack_hermes-provider-and-external-surfaces.md	owner-aware explain memory provenance
2026-04-25T06:57:03Z	done	codex-recall-metadata-eval	internal/eval/golden.go, internal/eval/golden_test.go, tests/golden/replay_eval.json, docs/review-packets/recall-preview-metadata-eval.md, plans/11_workpack_quality-ops-and-evals.md	recall preview block metadata expectations added to golden eval; focused golden eval and check-headers/diff-check passed; full go test/make eval blocked by active codex-main-demo-eval duplicate ExplainMemory method in internal/eval/demo.go vs graph_replay.go
2026-04-25T06:57:14Z	heartbeat	codex-main-demo-eval	internal/eval/demo.go, internal/eval/demo_test.go, cmd/cli/main.go, cmd/cli/main_test.go, Makefile, docs/review-packets/hermes-memory-demo-eval.md, internal/eval/graph_replay.go	demo eval implementation drafted; formatting and focused tests next
2026-04-25T06:57:29Z	heartbeat	codex-main-explain-visibility	internal/core/dto.go, internal/store/postgres/memories.go, internal/store/postgres/memories_test.go, internal/kernel/service.go, internal/kernel/service_test.go, internal/httpapi/router.go, internal/httpapi/router_test.go, internal/mcp/protocol.go, internal/mcp/protocol_test.go, docs/review-packets/explain-memory-scope-guard.md, docs/review-packets/explain-memory-visibility-guard.md, internal/mcp/surface_test.go, plans/05_runtime-contracts_ingest-recall-apply.md, plans/10_workpack_hermes-provider-and-external-surfaces.md	ExplainMemory visibility patch verified; full go test/make eval currently blocked by active codex-main-demo-eval expectation mismatch
2026-04-25T06:57:57Z	done	codex-main-explain-visibility	internal/core/dto.go, internal/store/postgres/memories.go, internal/store/postgres/memories_test.go, internal/kernel/service.go, internal/kernel/service_test.go, internal/httpapi/router.go, internal/httpapi/router_test.go, internal/mcp/protocol.go, internal/mcp/protocol_test.go, docs/review-packets/explain-memory-scope-guard.md, docs/review-packets/explain-memory-visibility-guard.md, internal/mcp/surface_test.go, plans/05_runtime-contracts_ingest-recall-apply.md, plans/10_workpack_hermes-provider-and-external-surfaces.md	ExplainMemory owner/group visibility guard added and documented; focused tests, make lint, make check-headers, git diff --check passed; full go test ./... and make eval blocked by active codex-main-demo-eval expectation mismatch in internal/eval/demo.go
2026-04-25T06:58:15Z	done	codex-main-demo-eval	internal/eval/demo.go, internal/eval/demo_test.go, cmd/cli/main.go, cmd/cli/main_test.go, Makefile, docs/review-packets/hermes-memory-demo-eval.md, internal/eval/graph_replay.go	Hermes Memory demo eval added; cli eval demo and make eval now exercise local trust loop; go test ./internal/eval, go test ./cmd/cli, go test ./..., make eval, make lint, make check-headers, git diff --check passed

```



<!-- Source: .agents/coordination/agent-work.sh | bytes=7255 | lines=314 | sha16=2bd901f9bdb73213 -->

```bash
#!/bin/sh
set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
STATE_DIR="$ROOT_DIR/.agents/coordination"
PROGRESS_FILE="$STATE_DIR/WORK_PROGRESS.md"
CLAIMS_FILE="$STATE_DIR/claims.tsv"
LOG_FILE="$STATE_DIR/activity.log"
LOCK_DIR="$STATE_DIR/.lock"

usage() {
	cat <<'EOF'
Usage:
  .agents/coordination/agent-work.sh init
  .agents/coordination/agent-work.sh status
  .agents/coordination/agent-work.sh claim <agent-id> <task> <file> [<file> ...]
  .agents/coordination/agent-work.sh heartbeat <agent-id> <note>
  .agents/coordination/agent-work.sh release <agent-id> <file> [<file> ...]
  .agents/coordination/agent-work.sh done <agent-id> <note>

Claim exact files before editing them. Release each file as soon as that file is
finished, before moving to another write surface.
EOF
}

timestamp() {
	date -u "+%Y-%m-%dT%H:%M:%SZ"
}

sanitize() {
	printf "%s" "$1" | tr '\t\r\n' '   '
}

repo_path() {
	path=$1
	case "$path" in
		"$ROOT_DIR"/*)
			path=${path#"$ROOT_DIR"/}
			;;
	esac
	printf "%s" "$path"
}

acquire_lock() {
	mkdir -p "$STATE_DIR"
	attempts=0
	while ! mkdir "$LOCK_DIR" 2>/dev/null; do
		attempts=$((attempts + 1))
		if [ "$attempts" -ge 30 ]; then
			echo "Timed out waiting for coordination lock: $LOCK_DIR" >&2
			exit 3
		fi
		sleep 1
	done
	trap 'rm -rf "$LOCK_DIR"' EXIT INT TERM
}

ensure_state() {
	[ -f "$CLAIMS_FILE" ] || : >"$CLAIMS_FILE"
	[ -f "$LOG_FILE" ] || : >"$LOG_FILE"
}

append_log() {
	action=$(sanitize "$1")
	agent=$(sanitize "$2")
	files=$(sanitize "$3")
	note=$(sanitize "$4")
	printf "%s\t%s\t%s\t%s\t%s\n" "$(timestamp)" "$action" "$agent" "$files" "$note" >>"$LOG_FILE"
}

join_files() {
	out=
	for raw_path in "$@"; do
		path=$(sanitize "$(repo_path "$raw_path")")
		if [ -z "$out" ]; then
			out=$path
		else
			out="$out, $path"
		fi
	done
	printf "%s" "$out"
}

render_progress() {
	now=$(timestamp)
	tmp="$STATE_DIR/WORK_PROGRESS.md.tmp.$$"
	{
		echo "# Agent Work Progress"
		echo
		echo "Live shared progress board for concurrent VibeGravity agents."
		echo
		echo "Generated by \`.agents/coordination/agent-work.sh\`. Prefer the script over manual edits."
		echo
		echo "Last rendered: $now"
		echo
		echo "## Active Claims"
		echo
		if [ -s "$CLAIMS_FILE" ]; then
			echo "| File | Agent | Task | Claimed at | Last update | Note |"
			echo "|---|---|---|---|---|---|"
			awk -F '\t' '
				function esc(s) {
					gsub(/\|/, "\\|", s)
					return s
				}
				NF >= 5 {
					note = ""
					if (NF >= 6) {
						note = $6
					}
					printf "| `%s` | %s | %s | %s | %s | %s |\n", esc($1), esc($2), esc($3), esc($4), esc($5), esc(note)
				}
			' "$CLAIMS_FILE"
		else
			echo "No active claims."
		fi
		echo
		echo "## Recent Activity"
		echo
		if [ -s "$LOG_FILE" ]; then
			echo "| Time | Action | Agent | Files | Note |"
			echo "|---|---|---|---|---|"
			tail -n 30 "$LOG_FILE" | awk -F '\t' '
				function esc(s) {
					gsub(/\|/, "\\|", s)
					return s
				}
				NF >= 5 {
					printf "| %s | %s | %s | %s | %s |\n", esc($1), esc($2), esc($3), esc($4), esc($5)
				}
			'
		else
			echo "No activity yet."
		fi
		echo
		echo "## Protocol"
		echo
		echo "- Read this file before starting and before widening scope."
		echo "- Claim exact files before editing."
		echo "- Do not edit files claimed by another active agent."
		echo "- Heartbeat during long work."
		echo "- Release files immediately when finished."
	} >"$tmp"
	mv "$tmp" "$PROGRESS_FILE"
}

claim_cmd() {
	if [ "$#" -lt 3 ]; then
		usage >&2
		exit 2
	fi
	agent=$(sanitize "$1")
	task=$(sanitize "$2")
	shift 2
	now=$(timestamp)
	conflicts=
	for raw_path in "$@"; do
		path=$(sanitize "$(repo_path "$raw_path")")
		owner=$(awk -F '\t' -v file="$path" -v agent="$agent" '$1 == file && $2 != agent { print $2; exit }' "$CLAIMS_FILE")
		if [ -n "$owner" ]; then
			if [ -z "$conflicts" ]; then
				conflicts="$path is already claimed by $owner"
			else
				conflicts="$conflicts
$path is already claimed by $owner"
			fi
		fi
	done
	if [ -n "$conflicts" ]; then
		echo "Claim rejected:" >&2
		printf "%s\n" "$conflicts" >&2
		exit 2
	fi

	tmp="$STATE_DIR/claims.tsv.tmp.$$"
	cp "$CLAIMS_FILE" "$tmp"
	for raw_path in "$@"; do
		path=$(sanitize "$(repo_path "$raw_path")")
		next="$tmp.next"
		awk -F '\t' -v file="$path" '$1 != file { print }' "$tmp" >"$next"
		mv "$next" "$tmp"
		printf "%s\t%s\t%s\t%s\t%s\t%s\n" "$path" "$agent" "$task" "$now" "$now" "$task" >>"$tmp"
	done
	mv "$tmp" "$CLAIMS_FILE"
	append_log "claim" "$agent" "$(join_files "$@")" "$task"
}

heartbeat_cmd() {
	if [ "$#" -ne 2 ]; then
		usage >&2
		exit 2
	fi
	agent=$(sanitize "$1")
	note=$(sanitize "$2")
	now=$(timestamp)
	tmp="$STATE_DIR/claims.tsv.tmp.$$"
	awk -F '\t' -v OFS='\t' -v agent="$agent" -v now="$now" -v note="$note" '
		$2 == agent {
			$5 = now
			$6 = note
		}
		{ print }
	' "$CLAIMS_FILE" >"$tmp"
	mv "$tmp" "$CLAIMS_FILE"
	append_log "heartbeat" "$agent" "$(awk -F '\t' -v agent="$agent" '$2 == agent { if (out == "") out = $1; else out = out ", " $1 } END { print out }' "$CLAIMS_FILE")" "$note"
}

release_cmd() {
	if [ "$#" -lt 2 ]; then
		usage >&2
		exit 2
	fi
	agent=$(sanitize "$1")
	shift
	conflicts=
	for raw_path in "$@"; do
		path=$(sanitize "$(repo_path "$raw_path")")
		owner=$(awk -F '\t' -v file="$path" '$1 == file { print $2; exit }' "$CLAIMS_FILE")
		if [ -n "$owner" ] && [ "$owner" != "$agent" ]; then
			if [ -z "$conflicts" ]; then
				conflicts="$path is claimed by $owner, not $agent"
			else
				conflicts="$conflicts
$path is claimed by $owner, not $agent"
			fi
		fi
	done
	if [ -n "$conflicts" ]; then
		echo "Release rejected:" >&2
		printf "%s\n" "$conflicts" >&2
		exit 2
	fi

	tmp="$STATE_DIR/claims.tsv.tmp.$$"
	cp "$CLAIMS_FILE" "$tmp"
	for raw_path in "$@"; do
		path=$(sanitize "$(repo_path "$raw_path")")
		next="$tmp.next"
		awk -F '\t' -v file="$path" -v agent="$agent" '!(($1 == file) && ($2 == agent)) { print }' "$tmp" >"$next"
		mv "$next" "$tmp"
	done
	mv "$tmp" "$CLAIMS_FILE"
	append_log "release" "$agent" "$(join_files "$@")" "released claimed files"
}

done_cmd() {
	if [ "$#" -ne 2 ]; then
		usage >&2
		exit 2
	fi
	agent=$(sanitize "$1")
	note=$(sanitize "$2")
	files=$(awk -F '\t' -v agent="$agent" '$2 == agent { if (out == "") out = $1; else out = out ", " $1 } END { print out }' "$CLAIMS_FILE")
	tmp="$STATE_DIR/claims.tsv.tmp.$$"
	awk -F '\t' -v agent="$agent" '$2 != agent { print }' "$CLAIMS_FILE" >"$tmp"
	mv "$tmp" "$CLAIMS_FILE"
	append_log "done" "$agent" "$files" "$note"
}

if [ "$#" -lt 1 ]; then
	usage >&2
	exit 2
fi

cmd=$1
shift

case "$cmd" in
	init)
		acquire_lock
		ensure_state
		render_progress
		echo "$PROGRESS_FILE"
		;;
	status)
		acquire_lock
		ensure_state
		render_progress
		cat "$PROGRESS_FILE"
		;;
	claim)
		acquire_lock
		ensure_state
		claim_cmd "$@"
		render_progress
		echo "Claim recorded. See $PROGRESS_FILE"
		;;
	heartbeat)
		acquire_lock
		ensure_state
		heartbeat_cmd "$@"
		render_progress
		echo "Heartbeat recorded. See $PROGRESS_FILE"
		;;
	release)
		acquire_lock
		ensure_state
		release_cmd "$@"
		render_progress
		echo "Release recorded. See $PROGRESS_FILE"
		;;
	done)
		acquire_lock
		ensure_state
		done_cmd "$@"
		render_progress
		echo "Done recorded. See $PROGRESS_FILE"
		;;
	*)
		usage >&2
		exit 2
		;;
esac

```



<!-- Source: .agents/coordination/claims.tsv | bytes=0 | lines=0 | sha16=e3b0c44298fc1c14 -->

```tsv

```



<!-- Source: .agents/hermes-orchestration/README.md | bytes=2118 | lines=68 | sha16=2d356e2a742b9869 -->

````md
# Hermes Orchestration

This folder holds repo-local helper scripts for dispatching work to the local
Hermes profiles used for VibeGravity coordination.

The scripts do not modify Hermes configuration. They do not run
`hermes config set`, `hermes profile use`, or alias management commands. Profile
selection is done only by setting `HERMES_HOME` for the child process.

## Profiles

| profile | HERMES_HOME |
|---|---|
| default | `/Users/parker/.hermes` |
| vuitton | `/Users/parker/.hermes/profiles/vuitton` |
| bottega | `/Users/parker/.hermes/profiles/bottega` |

All commands default to `openai-codex` with `gpt-5.5`.

## Single Agent

```bash
.agents/hermes-orchestration/run-agent.sh default .agents/hermes-orchestration/tasks/smoke/default.md
```

Outputs are written to:

```text
.agents/hermes-orchestration/runs/<run-id>/<profile>.out.md
.agents/hermes-orchestration/runs/<run-id>/<profile>.meta
```

## Dispatch Multiple Agents

Create a TSV manifest with one task per line:

```text
default	.agents/hermes-orchestration/tasks/smoke/default.md
vuitton	.agents/hermes-orchestration/tasks/smoke/vuitton.md
bottega	.agents/hermes-orchestration/tasks/smoke/bottega.md
```

Then run:

```bash
.agents/hermes-orchestration/dispatch.sh smoke-001 .agents/hermes-orchestration/tasks/smoke/manifest.tsv
.agents/hermes-orchestration/collect.sh smoke-001
```

## Orchestration Loop

1. Write focused task prompts under `tasks/<run-id>/`.
2. Dispatch to profiles with `dispatch.sh`.
3. Collect results with `collect.sh`.
4. Inspect output, decide follow-up work, and create the next manifest.
5. Keep all repo/code changes in this Codex session unless a Hermes result is
   explicitly selected for implementation.

## Coordination Requirement

Every Hermes task prompt that may edit repo files must include the coordination
snippet from `.agents/coordination/PROMPT_SNIPPET.md`.

Before editing, each Hermes profile must read
`.agents/coordination/WORK_PROGRESS.md`, claim exact files with
`.agents/coordination/agent-work.sh claim ...`, heartbeat during long work, and
release files immediately after finishing them.

````



<!-- Source: .agents/hermes-orchestration/collect.sh | bytes=906 | lines=41 | sha16=2450cef6a5166ecb -->

```bash
#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 || "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  echo "Usage: collect.sh <run-id>" >&2
  exit 2
fi

run_id="$1"
repo_root="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
run_dir="$repo_root/.agents/hermes-orchestration/runs/$run_id"

if [[ ! -d "$run_dir" ]]; then
  echo "Run directory not found: $run_dir" >&2
  exit 2
fi

for meta in "$run_dir"/*.meta; do
  [[ -e "$meta" ]] || continue
  profile="$(basename "$meta" .meta)"
  out_file="$run_dir/$profile.out.md"
  err_file="$run_dir/$profile.err.log"

  echo "===== $profile meta ====="
  sed -n '1,120p' "$meta"
  echo
  echo "===== $profile output ====="
  if [[ -s "$out_file" ]]; then
    sed -n '1,220p' "$out_file"
  else
    echo "(no output)"
  fi
  if [[ -s "$err_file" ]]; then
    echo
    echo "===== $profile stderr ====="
    sed -n '1,120p' "$err_file"
  fi
  echo
done


```



<!-- Source: .agents/hermes-orchestration/dispatch.sh | bytes=1396 | lines=75 | sha16=eb807aa84cb503d2 -->

```bash
#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage:
  dispatch.sh <run-id> <manifest.tsv>

Manifest format:
  <profile><TAB><prompt-file>

Blank lines and lines starting with # are ignored.
USAGE
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

if [[ $# -ne 2 ]]; then
  usage >&2
  exit 2
fi

run_id="$1"
manifest="$2"

if [[ ! -f "$manifest" ]]; then
  echo "Manifest not found: $manifest" >&2
  exit 2
fi

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
run_dir="$repo_root/.agents/hermes-orchestration/runs/$run_id"
mkdir -p "$run_dir"

pids=()
profiles=()

while IFS=$'\t' read -r profile prompt_file rest; do
  if [[ -z "${profile:-}" || "${profile:0:1}" == "#" ]]; then
    continue
  fi
  if [[ -n "${rest:-}" ]]; then
    echo "Invalid manifest line for profile $profile: too many fields" >&2
    exit 2
  fi

  "$script_dir/run-agent.sh" "$profile" "$prompt_file" "$run_id" &
  pids+=("$!")
  profiles+=("$profile")
done < "$manifest"

if [[ ${#pids[@]} -eq 0 ]]; then
  echo "No tasks found in manifest: $manifest" >&2
  exit 2
fi

status=0
for i in "${!pids[@]}"; do
  if wait "${pids[$i]}"; then
    echo "${profiles[$i]}: ok"
  else
    code=$?
    echo "${profiles[$i]}: failed ($code)" >&2
    status=1
  fi
done

echo "run_dir=$run_dir"
exit "$status"


```



<!-- Source: .agents/hermes-orchestration/run-agent.sh | bytes=1924 | lines=100 | sha16=ab5de56eeac3a035 -->

```bash
#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage:
  run-agent.sh <profile> <prompt-file> [run-id]

Profiles:
  default
  vuitton
  bottega

Environment overrides:
  HERMES_MODEL       default: gpt-5.5
  HERMES_PROVIDER    default: openai-codex
  HERMES_MAX_TURNS   default: 90
USAGE
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

if [[ $# -lt 2 || $# -gt 3 ]]; then
  usage >&2
  exit 2
fi

profile="$1"
prompt_file="$2"
run_id="${3:-$(date +%Y%m%d_%H%M%S)}"

case "$profile" in
  default)
    hermes_home="/Users/parker/.hermes"
    ;;
  vuitton)
    hermes_home="/Users/parker/.hermes/profiles/vuitton"
    ;;
  bottega)
    hermes_home="/Users/parker/.hermes/profiles/bottega"
    ;;
  *)
    echo "Unknown profile: $profile" >&2
    exit 2
    ;;
esac

if [[ ! -f "$prompt_file" ]]; then
  echo "Prompt file not found: $prompt_file" >&2
  exit 2
fi

repo_root="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
run_dir="$repo_root/.agents/hermes-orchestration/runs/$run_id"
mkdir -p "$run_dir"

out_file="$run_dir/$profile.out.md"
err_file="$run_dir/$profile.err.log"
meta_file="$run_dir/$profile.meta"

model="${HERMES_MODEL:-gpt-5.5}"
provider="${HERMES_PROVIDER:-openai-codex}"
max_turns="${HERMES_MAX_TURNS:-90}"

{
  echo "profile=$profile"
  echo "hermes_home=$hermes_home"
  echo "prompt_file=$prompt_file"
  echo "run_id=$run_id"
  echo "model=$model"
  echo "provider=$provider"
  echo "started_at=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
} > "$meta_file"

set +e
env HERMES_HOME="$hermes_home" hermes chat \
  -Q \
  --source tool \
  --provider "$provider" \
  -m "$model" \
  --max-turns "$max_turns" \
  -q "$(cat "$prompt_file")" \
  > "$out_file" \
  2> "$err_file"
status=$?
set -e

{
  echo "finished_at=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  echo "exit_code=$status"
  echo "out_file=$out_file"
  echo "err_file=$err_file"
} >> "$meta_file"

exit "$status"


```



<!-- Source: .agents/hermes-orchestration/status.sh | bytes=159 | lines=12 | sha16=b50df71268139ca6 -->

```bash
#!/usr/bin/env bash
set -euo pipefail

hermes profile list
echo
hermes profile show default
echo
hermes profile show vuitton
echo
hermes profile show bottega


```



<!-- Source: .agents/hermes-orchestration/tasks/20260424-eval-graph-gates/bottega.md | bytes=1224 | lines=35 | sha16=9178bb9124bd4cbe -->

```md
You are the bottega Hermes profile reviewing QA, edge cases, and release gates for VibeGravity.

Repo: /Users/parker/Documents/VibeGravity

Rules:
- Do not mutate Hermes settings.
- Do not run hermes config/profile mutation commands.
- Do not edit files. Return review guidance only.
- Do not propose real Codex calls for this slice.
- Do not reintroduce local extractor behavior.
- Do not enable group_shared writes without membership validation.

Read first:
- AGENTS.md
- PLANS.md
- plans/01_rfp_vibegravity_hermes-first.md
- plans/05_runtime-contracts_ingest-recall-apply.md
- plans/11_workpack_quality-ops-and-evals.md
- internal/eval/golden.go
- internal/eval/golden_test.go
- tests/golden/replay_eval.json
- internal/worker/processor.go
- internal/store/postgres/jobs.go

Task:
Review QA and regression coverage for the next smallest V1.0 slice:
"Add replay/eval gates for graph updates and human correction supersession."

Please return:
1. Edge cases the eval must cover.
2. Release gates to run before handoff.
3. How to keep Codex outage/backlog behavior visible without real Codex calls.
4. Risks around false confidence from fixture-only evals.
5. A concise recommended checklist for Codex before final report.

```



<!-- Source: .agents/hermes-orchestration/tasks/20260424-eval-graph-gates/default.md | bytes=1223 | lines=33 | sha16=e85e7ad14a5cd652 -->

```md
You are the default Hermes profile reviewing the next VibeGravity implementation slice.

Repo: /Users/parker/Documents/VibeGravity

Rules:
- Do not mutate Hermes settings.
- Do not run hermes config/profile mutation commands.
- Do not edit files. Return review guidance only.
- Do not propose real Codex calls for this slice.
- Do not reintroduce local extractor behavior.
- Do not enable group_shared writes without membership validation.

Read first:
- AGENTS.md
- PLANS.md
- plans/00_read-this-first_for-building-agents.md
- plans/01_rfp_vibegravity_hermes-first.md
- plans/02_product-contract_and_direction.md
- plans/05_runtime-contracts_ingest-recall-apply.md
- plans/11_workpack_quality-ops-and-evals.md
- docs/review-packets/current-state-and-next-agent-handoff.md

Task:
Review the product and contract acceptance criteria for the next smallest V1.0 slice:
"Add replay/eval gates for graph updates and human correction supersession."

Please return:
1. The exact acceptance criteria this slice should satisfy.
2. Stop-lines: what must stay out of scope.
3. Product risks if the eval is too shallow.
4. A recommended small implementation shape that stays within current contracts.
5. Any docs that must be updated.

```



<!-- Source: .agents/hermes-orchestration/tasks/20260424-eval-graph-gates/manifest.tsv | bytes=240 | lines=4 | sha16=e62362197aa481d9 -->

```tsv
default	.agents/hermes-orchestration/tasks/20260424-eval-graph-gates/default.md
vuitton	.agents/hermes-orchestration/tasks/20260424-eval-graph-gates/vuitton.md
bottega	.agents/hermes-orchestration/tasks/20260424-eval-graph-gates/bottega.md

```



<!-- Source: .agents/hermes-orchestration/tasks/20260424-eval-graph-gates/vuitton.md | bytes=1209 | lines=34 | sha16=f2493317267a4c00 -->

```md
You are the vuitton Hermes profile reviewing Go/Postgres/API implementation design for VibeGravity.

Repo: /Users/parker/Documents/VibeGravity

Rules:
- Do not mutate Hermes settings.
- Do not run hermes config/profile mutation commands.
- Do not edit files. Return review guidance only.
- Do not propose real Codex calls for this slice.
- Do not reintroduce local extractor behavior.
- Do not enable group_shared writes without membership validation.

Read first:
- AGENTS.md
- PLANS.md
- plans/05_runtime-contracts_ingest-recall-apply.md
- plans/06_data-model_and_storage-invariants.md
- internal/eval/golden.go
- internal/graph/store_apply.go
- internal/store/postgres/memories.go
- internal/kernel/service.go
- tests/golden/replay_eval.json

Task:
Review implementation options for the next smallest V1.0 slice:
"Add replay/eval gates for graph updates and human correction supersession."

Please return:
1. The lowest-risk Go implementation design.
2. Whether this should extend internal/eval, add a new fixture, or add a separate package.
3. Transaction/idempotency risks that the eval must catch.
4. Specific tests that should be added.
5. Files likely to change, with any risks of touching hot files.

```



<!-- Source: .agents/hermes-orchestration/tasks/20260424_backlog_metrics/bottega.md | bytes=1020 | lines=29 | sha16=975dc0ee436d9d88 -->

```md
You are the bottega Hermes QA/regression agent for VibeGravity.

Work in /Users/parker/Documents/VibeGravity.

Do not mutate Hermes settings or profiles. Do not run hermes config/profile mutation commands.

Read:
- AGENTS.md
- PLANS.md
- plans/05_runtime-contracts_ingest-recall-apply.md
- plans/11_workpack_quality-ops-and-evals.md
- docs/review-packets/current-state-and-next-agent-handoff.md
- cmd/cli/main_test.go
- internal/store/postgres/jobs_test.go
- internal/eval/worker_backlog.go
- tests/golden/replay_eval.json

Task:
Define QA/regression coverage for the next slice: operator-visible worker backlog counts and recovery metrics.

Please focus on:
1. Edge cases for queue status counts and oldest queued age.
2. How to prove blocked jobs do not look retryable.
3. How to keep the existing mocked Codex outage eval gates passing.
4. Release-gate commands and any doc assertions.
5. Any test that would catch accidental real Codex/local extractor behavior.

Return a concise markdown report. Do not edit files.

```



<!-- Source: .agents/hermes-orchestration/tasks/20260424_backlog_metrics/default.md | bytes=1059 | lines=28 | sha16=46d28d31a9de60a6 -->

```md
You are the default Hermes review agent for VibeGravity.

Work in /Users/parker/Documents/VibeGravity.

Do not mutate Hermes settings or profiles. Do not run hermes config/profile mutation commands.

Read:
- AGENTS.md
- PLANS.md
- plans/00_read-this-first_for-building-agents.md
- plans/01_rfp_vibegravity_hermes-first.md
- plans/02_product-contract_and_direction.md
- plans/03_target-architecture_codex-first.md
- plans/05_runtime-contracts_ingest-recall-apply.md
- plans/06_data-model_and_storage-invariants.md
- docs/review-packets/current-state-and-next-agent-handoff.md

Task:
Review the next smallest V1.0 slice: operator-visible worker backlog counts and recovery metrics, without changing retry/block semantics and without real Codex calls.

Please report:
1. Product/contract acceptance criteria for this slice.
2. Stop-lines: what must not be implemented in this slice.
3. What docs must be updated.
4. Any risk that could accidentally blur private/shared/group scope or raw/derived separation.

Return a concise markdown report. Do not edit files.

```



<!-- Source: .agents/hermes-orchestration/tasks/20260424_backlog_metrics/manifest.tsv | bytes=237 | lines=4 | sha16=1ce1a01767fc6bd7 -->

```tsv
default	.agents/hermes-orchestration/tasks/20260424_backlog_metrics/default.md
vuitton	.agents/hermes-orchestration/tasks/20260424_backlog_metrics/vuitton.md
bottega	.agents/hermes-orchestration/tasks/20260424_backlog_metrics/bottega.md

```



<!-- Source: .agents/hermes-orchestration/tasks/20260424_backlog_metrics/vuitton.md | bytes=1189 | lines=42 | sha16=85523f19e6a50164 -->

```md
You are the vuitton Hermes implementation-design agent for VibeGravity.

Work in /Users/parker/Documents/VibeGravity.

Do not mutate Hermes settings or profiles. Do not run hermes config/profile mutation commands.

Read:
- AGENTS.md
- PLANS.md
- plans/05_runtime-contracts_ingest-recall-apply.md
- plans/06_data-model_and_storage-invariants.md
- cmd/cli/main.go
- cmd/cli/main_test.go
- internal/store/store.go
- internal/store/postgres/jobs.go
- internal/store/postgres/jobs_test.go
- internal/core/job.go

Task:
Design the smallest Go/Postgres/API implementation for operator-visible job backlog metrics.

Current CLI has:
- cli jobs blocked [--limit N]
- cli jobs requeue-blocked <job_id>

Needed slice:
- queued/running/failed/blocked/complete counts
- oldest queued age
- simple drain-rate/recovery-ETA shape if possible from current table fields
- no retry/block semantic changes
- no real Codex calls
- preserve existing blocked job recovery commands

Please report:
1. Recommended interface/type shape.
2. Recommended SQL shape and transaction/concurrency risks.
3. CLI command shape.
4. Tests to add.
5. Files likely touched.

Return a concise markdown report. Do not edit files.

```



<!-- Source: .agents/hermes-orchestration/tasks/smoke/bottega.md | bytes=71 | lines=4 | sha16=f834ccbd36ea5628 -->

```md
You are the bottega Hermes profile.
Reply exactly: BOTTEGA_WRAPPER_OK


```



<!-- Source: .agents/hermes-orchestration/tasks/smoke/default.md | bytes=71 | lines=4 | sha16=c952c4e9dc864440 -->

```md
You are the default Hermes profile.
Reply exactly: DEFAULT_WRAPPER_OK


```



<!-- Source: .agents/hermes-orchestration/tasks/smoke/manifest.tsv | bytes=181 | lines=5 | sha16=c3ab1eb70530fd59 -->

```tsv
default	.agents/hermes-orchestration/tasks/smoke/default.md
vuitton	.agents/hermes-orchestration/tasks/smoke/vuitton.md
bottega	.agents/hermes-orchestration/tasks/smoke/bottega.md


```



<!-- Source: .agents/hermes-orchestration/tasks/smoke/vuitton.md | bytes=71 | lines=4 | sha16=1a63e7eb1a3c71c3 -->

```md
You are the vuitton Hermes profile.
Reply exactly: VUITTON_WRAPPER_OK


```



<!-- Source: .agents/hermes-orchestration/tasks/v1-update-memory/bottega.md | bytes=454 | lines=20 | sha16=526e7b245a317d52 -->

```md
You are the bottega Hermes profile working as QA/test engineer.

Repo: /Users/parker/Documents/VibeGravity

Inspect tests around:
- internal/graph
- internal/store/postgres
- tests/migration_contract_test.go
- Makefile

Task:
Define the regression suite needed before considering update_memory safe for V1.0 progress.

Return only:
1. Store-level tests.
2. Apply-level tests.
3. Migration/constraint checks.
4. Commands to run and likely failure modes.


```



<!-- Source: .agents/hermes-orchestration/tasks/v1-update-memory/default.md | bytes=719 | lines=24 | sha16=bd8ddd0eee8d429b -->

```md
You are the default Hermes profile working as VibeGravity product/contract auditor.

Repo: /Users/parker/Documents/VibeGravity

Read:
- AGENTS.md
- PLANS.md
- plans/00_read-this-first_for-building-agents.md
- plans/01_rfp_vibegravity_hermes-first.md
- plans/02_product-contract_and_direction.md
- plans/05_runtime-contracts_ingest-recall-apply.md
- plans/06_data-model_and_storage-invariants.md
- docs/review-packets/current-state-and-next-agent-handoff.md

Task:
Identify the exact acceptance criteria for the next V1.0 slice: safe update_memory transaction semantics.

Return only:
1. Must-have behavior.
2. Files likely involved.
3. Tests that must exist.
4. Stop lines: what must not be implemented in this slice.


```



<!-- Source: .agents/hermes-orchestration/tasks/v1-update-memory/manifest.tsv | bytes=214 | lines=5 | sha16=2f1626d2b6705488 -->

```tsv
default	.agents/hermes-orchestration/tasks/v1-update-memory/default.md
vuitton	.agents/hermes-orchestration/tasks/v1-update-memory/vuitton.md
bottega	.agents/hermes-orchestration/tasks/v1-update-memory/bottega.md


```



<!-- Source: .agents/hermes-orchestration/tasks/v1-update-memory/vuitton.md | bytes=780 | lines=24 | sha16=5f8ea70584e01056 -->

```md
You are the vuitton Hermes profile working as Go/Postgres implementation reviewer.

Repo: /Users/parker/Documents/VibeGravity

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


```



<!-- Source: .agents/skills/code-headers.md | bytes=1064 | lines=34 | sha16=a045104a41d83025 -->

```md
---
name: code-headers
description: Use this skill when creating, renaming, or materially editing Go source files so VibeGravity code headers stay parseable and useful.
---

# Code Headers

## Purpose

Keep source files self-orienting for humans and agents without turning headers
into long in-file histories.

## Required Inputs

- `docs/code-header-policy.md`
- `plans/templates/code-header-minimal-go.md`

## Steps

1. Choose the minimal structured header by default.
2. Use the narrative template only when architectural rationale is essential.
3. Use the development-log template only when a file is explicitly audit-heavy.
4. Fill `FILE`, `PURPOSE`, `LAYER`, `STATUS`, `EXPORTS`, `DEPENDS`, `USED_BY`,
   and `AGENT_NOTE`.
5. Keep `DEPENDS` and `USED_BY` short and action-oriented.
6. After edits or renames, run `make check-headers`.
7. If a header rule changes, update `docs/code-header-policy.md` first.

## Done When

- New or changed Go files have current headers.
- `make check-headers` passes.
- Any intentional exception is documented in the policy.

```



<!-- Source: .agents/skills/contract-check.md | bytes=1171 | lines=39 | sha16=6ee5246ce5854512 -->

```md
---
name: contract-check
description: Use this skill to compare code changes against VibeGravity product and architecture contracts.
---

# Contract Check

## Purpose

This skill checks whether the current implementation violates the documented product contract.

## Required docs

- `plans/02_product-contract_and_direction.md`
- `plans/03_target-architecture_codex-first.md`
- `plans/05_runtime-contracts_ingest-recall-apply.md`
- `plans/06_data-model_and_storage-invariants.md`

## Review checklist

- Hermes-first direction kept
- Local extractor not reintroduced into main path
- Scope separation preserved (agent_private / workspace_shared / group_shared)
- Raw and derived separation preserved
- Provenance path preserved (memory_trace mandatory)
- Recall budget logic preserved
- Stage 1 / Stage 2 / Apply boundary is structured JSON only
- Apply engine 12-step pipeline order respected
- Idempotency preserved on all write paths
- group_shared requires valid membership
- Docs updated if behavior changed

## Output

- critical breaks (blocks merge)
- medium concerns (should fix before next pack)
- minor notes (improve when convenient)
- files to inspect next

```



<!-- Source: .agents/skills/eval-regression.md | bytes=888 | lines=33 | sha16=4aa75924b79f3a65 -->

```md
---
name: eval-regression
description: Use this skill to run or inspect golden scenarios and detect memory regressions.
---

# Eval Regression

## When to use

Use this after changing reasoning, recall, scopes, or profile logic.

## Required scenarios

- correction updates old fact (supersession works)
- workspace_shared vs agent_private separation (scope filter correct)
- group_shared visibility (membership check works)
- pinned note inclusion in recall
- active plan inclusion in recall
- superseded memory suppression (latest_flag logic)
- dreaming promotion (session → mid → long → ultra-long)
- degraded recall without Codex (existing profile + recent tail still returned)
- budget_tokens respected (output within budget)
- fingerprint dedup prevents duplicate memories

## Output

- scenario
- expected result
- observed result
- pass or fail
- suspected cause
- next fix

```



<!-- Source: .agents/skills/plan-implement-verify.md | bytes=764 | lines=33 | sha16=88f46f0e4ef22587 -->

```md
---
name: plan-implement-verify
description: Use this skill for feature work that needs a short plan, code changes, checks, and self-review.
---

# Plan Implement Verify

## When to use

Use this skill when the task changes code or contracts.
Do not use it for trivial one-line edits.

## Steps

1. Read the relevant contract docs from `plans/`.
2. Write a short plan in PLANS.md or as a comment.
3. Implement the smallest coherent slice.
4. Run the relevant checks:
   - `go test ./...`
   - `golangci-lint run`
5. Review the diff against the contract (load `contract-check` skill if needed).
6. Report changed files, commands, results, and risks.

## Output format

- plan
- implementation summary
- files changed
- checks run
- results
- risks
- docs updated

```



<!-- Source: .agents/skills/source-provenance.md | bytes=1905 | lines=56 | sha16=509c2e94b5266470 -->

````md
---
name: source-provenance
description: Use this skill before adding code influenced by external material so VibeGravity stays safe for open-source release.
---

# Source Provenance

## Purpose

Keep VibeGravity's implementation original, reviewable, and safe to publish as
open source.

## Allowed Reference Classes

- First-principles implementation from VibeGravity's own plans and contracts.
- Official language, standard library, or dependency documentation.
- Commercially usable permissive patterns from MIT, BSD, or Apache-2.0 sources.

## Blocked Reference Classes

- GPL, AGPL, LGPL, SSPL, Elastic License, or related copyleft/source-available
  license families.
- Verbatim code, comments, function names, file layouts, or distinctive
  implementation structure from a specific external project.
- Structured external snippets of 10 or more consecutive lines.

## Steps

1. Identify whether the implementation is first-principles or externally
   influenced.
2. If external influence exists, record the estimated source class and suspected
   license before coding.
3. Rewrite any structured snippet of 10 or more consecutive lines from first
   principles.
4. Change names, boundaries, and file placement to match VibeGravity's own
   architecture, not the referenced project.
5. Stop and warn if substantial similarity risk remains.
6. Include the source review block in the handoff.

## Source Review Block

```text
Source Review:
- Estimated source: first-principles VibeGravity plans / official docs / permissive pattern / unknown
- Suspected license: none / MIT / BSD / Apache-2.0 / unknown
- Similarity risk: low / medium / high
- Review required: no / yes
- Notes: short rationale
```

## Done When

- The code is newly implemented for VibeGravity.
- No blocked license family was used as a coding reference.
- The handoff includes the source review block for any code-bearing change.

````



<!-- Source: .gitignore | bytes=16 | lines=3 | sha16=fe048748c7e01757 -->

```text
.DS_Store
.omx/

```



<!-- Source: .gitmessage.txt | bytes=684 | lines=26 | sha16=2b28ba5552cc6ae5 -->

```text
# Format:
# [scope] [subscope] Imperative summary
#
# Examples:
# [docs] Add commit message rules for VibeGravity
# [api] [ingest] Tighten validation for memory writes
# [tests] Cover duplicate recall edge case
#
# Rules:
# - Keep the subject concise and imperative.
# - Prefer bracketed scopes, similar to Airbnb's public repo history.
# - Target 72 characters for the full subject line.
# - Do not end the subject with a period.
#
# Why:
# - Explain the reason for the change.
#
# What:
# - Summarize the main code or document changes.
#
# Notes:
# - Mention follow-up work, rollout caveats, or migration details if needed.
#
# Refs:
# - Link issue, PR, doc, or ticket if relevant.

```



<!-- Source: .golangci.yml | bytes=352 | lines=26 | sha16=a9626d6f4cbc26dd -->

```yaml
run:
  timeout: 5m
  tests: true

linters:
  disable-all: true
  enable:
    - errcheck
    - gosimple
    - govet
    - ineffassign
    - staticcheck
    - typecheck
    - unused
    - gofmt
    - goimports
    - misspell
    - unparam
    - unconvert
    - revive

issues:
  exclude-use-default: false
  max-issues-per-linter: 0
  max-same-issues: 0

```



<!-- Source: AGENTS.md | bytes=8481 | lines=214 | sha16=60181277d06ab45a -->

````md
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

## Open-source code policy

VibeGravity is open source.
Do not reference or closely reproduce code under GPL, AGPL, LGPL, SSPL, or Elastic License families.
Use only commercially usable permissive patterns such as MIT, BSD, and Apache-2.0 as reference material, and implement anew from principles.
Do not reproduce function names, file structure, comments, or distinctive implementations from a specific open-source project.
If generated code may be substantially similar to external open-source code, stop and warn before coding, then offer an alternative design.
When handing off code, include a source review: estimated source, suspected license, similarity risk, and whether human review is required.
Treat structured external snippets of 10 or more consecutive lines as similarity risk and rewrite from first principles.
Use `.agents/skills/source-provenance.md` when adding code inspired by external material.

## Code file headers

All non-generated Go source files must start with the VibeGravity code header.
Default to the minimal structured header in `plans/templates/code-header-minimal-go.md`.
Use `.agents/skills/code-headers.md` when creating, renaming, or materially editing Go files.
Run `make check-headers` before handoff.

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
For parallel or multi-agent work, read `.agents/coordination/WORK_PROGRESS.md`
first and claim exact files with `.agents/coordination/agent-work.sh` before
editing. Release each file immediately when finished. See
`.agents/coordination/README.md` for the full protocol.
If the operator gives only
`.agents/coordination/UNIVERSAL_AGENT_PROMPT.md` or its absolute path, treat
that as an instruction to read and execute the prompt immediately; do not ask
what to do with it.
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

````



<!-- Source: CLAUDE.md | bytes=3637 | lines=90 | sha16=48ab49e752353f8c -->

````md
# CLAUDE.md

## Project

This repo builds VibeGravity.
VibeGravity is a shared memory kernel for Hermes-first agent workflows.

## Hold these facts in every session

- Hermes is the first customer
- Language is Go (1.22+), database is PostgreSQL + pgvector
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

````



<!-- Source: COMMIT_MESSAGE_RULES.md | bytes=1994 | lines=93 | sha16=0b712b5a6f88b161 -->

````md
# Commit Message Rules

This repo uses an Airbnb-inspired commit style based on the public commit
history in `airbnb/javascript`: concise imperative subjects, optional
bracketed scopes, and a short body when context matters.

## Subject Line Format

Use this shape:

```text
[scope] [subscope] Imperative summary
```

Examples:

```text
[docs] Add commit message rules for VibeGravity
[plans] Clarify external surface workpacks
[api] [recall] Return normalized memory filters
```

## Rules

1. Start with one or two bracketed scopes when they add signal.
2. Write the subject in imperative mood: `Add`, `Fix`, `Refactor`, `Remove`,
   `Update`, `Document`, `Test`.
3. Keep the full subject line concise. Target 72 characters.
4. Do not end the subject with a period.
5. Make one commit represent one logical change.
6. Add a blank line before the body.
7. Use the body to explain why the change exists and what matters for review.
8. Wrap body lines at about 72 characters when practical.
9. Add references only when they help future readers.

## Recommended Body Sections

When the change is non-trivial, use short labeled sections:

```text
Why:
- reason this change was necessary

What:
- main files or behaviors changed

Notes:
- follow-up work, caveats, or migration details

Refs:
- issue, PR, or design doc
```

## Suggested Scopes

- `[docs]`
- `[plans]`
- `[repo]`
- `[api]`
- `[server]`
- `[worker]`
- `[mcp]`
- `[tests]`
- `[deps]`
- `[release]`
- `[infra]`
- `[security]`

## Quick Examples

```text
[repo] Add Airbnb-style git commit template

Why:
- establish a consistent commit format before the first project history lands

What:
- add a reusable git commit template
- document repo commit rules
- ignore local OMX and Finder noise
```

```text
[api] [ingest] Reject empty normalized statements

Why:
- prevent low-signal rows from entering shared memory storage

What:
- validate normalized statements before persistence
- add regression coverage for blank input payloads
```

````



<!-- Source: Makefile | bytes=911 | lines=44 | sha16=a79d2a4f0ac5007c -->

```text
.PHONY: build test eval lint check-headers clean dev-server dev-worker setup

GOLANGCI_LINT ?= $(shell command -v golangci-lint 2>/dev/null || printf "%s/bin/golangci-lint" "$$(go env GOPATH)")

# Build all binaries
build:
	go build -o bin/server ./cmd/server
	go build -o bin/worker ./cmd/worker
	go build -o bin/cli ./cmd/cli

# Run all tests
test:
	go test -v ./...

# Run deterministic golden quality evals
eval:
	go run ./cmd/cli eval golden --path tests/golden/replay_eval.json
	go run ./cmd/cli eval demo

# Run linter
lint:
	$(GOLANGCI_LINT) run

# Check source file headers
check-headers:
	go run ./tools/headercheck

# Clean build artifacts
clean:
	rm -rf bin/

# Run dev server
dev-server:
	go run ./cmd/server

# Run dev worker
dev-worker:
	go run ./cmd/worker

# Setup environment (useful for new devs)
setup:
	go mod download
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

```



<!-- Source: PLANS.md | bytes=7057 | lines=136 | sha16=e0ac8c3ab5dbd36a -->

```md
# PLANS.md

## Current Work Pack

Hermes Memory trust loop and first-customer integration.

## V1 Product Promise

V1 is now framed as **Hermes Memory, powered by VibeGravity**.

The first release must prove one felt outcome:

> Hermes remembers the right project context across sessions, shows why it
> remembered it, and lets the operator fix memory once.

`VibeGravity` remains the engine and internal architecture name. Public and
first-customer language should lead with Hermes continuity, correction, and
proof rather than "shared memory kernel".

## Current State

The Go-first foundation is now beyond the original Work Pack 01 skeleton.
The repo has runnable core contracts, PostgreSQL migrations, HTTP handlers,
sync/prefetch paths, note/plan/document surfaces, store-backed graph apply for
`create_memory` and safe `extend_memory`, worker blocked-job handling, and
schema-first reasoning stubs.

The project is not V1-complete. Safe `update_memory` transaction semantics,
correction supersession, the first in-repo Hermes provider adapter, Hermes MCP
bootstrap output, real stdio MCP protocol serving, narrow graph replay eval
gates, deterministic mocked Codex outage / worker backlog recovery eval gates,
read-only worker backlog metrics, group-shared membership filtering, and
operator-visible recall freshness degradation now exist, but real Codex
execution, custom Hermes memory provider registry packaging, full session
replay, and production ops remain.

Documents and rich dreaming remain engine capabilities, but they are no longer
the V1 product headline. V1 should sell the trust loop: recall preview,
explain/timeline, correction, supersession, visible scope, and degraded-status
truthfulness.

## Active Review Packet

- `docs/review-packets/hermes-memory-trust-loop-product-pivot.md`
- `docs/review-packets/operator-visible-degraded-recall-freshness.md`

## Next Concrete Slice

Move from safe graph writes toward the Hermes Memory trust slice.

Goal:

- Make the operator-visible trust loop excellent before broadening product
  scope: recall preview, explain, correct, timeline, visible scope, and degraded
  freshness metadata.
- Use the stdio MCP server as the first Hermes-facing protocol integration while
  Hermes custom memory-provider registry support remains unavailable.
- Keep real Codex disabled by default until failure behavior, freshness loss,
  and backlog recovery are visible to the operator.
- Preserve the `update_memory` transaction boundary while broader surfaces land.
- Re-run `go test ./...`, `make lint`, `make check-headers`, and `git diff --check`.

Recently completed:

- `internal/eval` now runs deterministic golden recall scenarios from
  `tests/golden/replay_eval.json`, with `cli eval golden` and `make eval` as the
  first quality regression gate.
- `internal/eval` now also runs narrow graph replay scenarios through the real
  store-backed apply engine, checking `update_memory` retry idempotency,
  correction-shaped supersession recall, mandatory trace/edge counts, prior
  memory supersession, and the current `group_shared` write stop-line.
- `internal/eval` now runs worker backlog scenarios through the real
  `worker.Processor` and `graph.StoreBackedApplyEngine` with mocked Stage 1 and
  Stage 2 outage controls. The gate proves transient reasoning failure retries
  without derived graph side effects, recovery writes only after structured
  reasoning succeeds, replay remains idempotent for memory/trace/edge rows, and
  unsupported deterministic apply work becomes blocked instead of retrying
  forever.
- `cli jobs metrics [--window D] [--tenant ID] [--workspace ID]` now reports
  read-only operator backlog visibility: total queued, ready queued, running,
  failed, blocked, and complete counts, retryable queued attempts, oldest ready
  queued age, oldest running job age, drain rate, and recovery ETA when enough
  completed-job history exists. It does not claim, requeue, fail, complete, or
  unblock jobs.
- `update_memory` now writes a replacement memory, mandatory trace, `updates`
  edge, and prior-memory supersession inside one PostgreSQL transaction. The
  path locks and verifies the target as active/latest, rejects scope/owner
  boundary changes, and treats deterministic successful retries as idempotent.
- Operator blocked-job recovery exists through `cli jobs blocked [--limit N]`
  and `cli jobs requeue-blocked <job_id>`.
- `internal/hermes.Provider` maps Hermes lifecycle hooks to core `Prefetch` and
  `SyncTurn`, renders typed recall context, exposes the minimum tool list, and
  has mocked lifecycle tests.
- `internal/mcp.Surface` maps MCP-style tool names to the same core service
  calls and has mocked tool delegation tests.
- `internal/mcp.Server` serves `initialize`, `notifications/initialized`,
  `ping`, `tools/list`, and `tools/call` over newline-delimited MCP stdio JSON-RPC.
- `cli mcp serve --stdio` starts the real MCP protocol server, and
  `cli hermes bootstrap` prints the `hermes mcp add ... --args mcp serve --stdio`
  registration plus `hermes mcp test` verification command.
- `POST /v1/documents` now uses an atomic document+chunk store path.
- `/healthz` returns `503` for a missing DB pool instead of panicking in embedded/test surfaces.
- `CorrectMemory` now validates, records an idempotent raw correction event,
  writes an append-safe `memory_corrections` artifact, then applies the
  correction text as a replacement memory with mandatory trace, an `updates`
  edge, and prior-memory supersession while preserving the original memory
  trace.
- `GetTimeline` now parses timeline query parameters and returns a read-only, scope-aware timeline over memories, traces, and correction artifacts.
- `Prefetch` now consumes read-only worker backlog freshness signals and marks
  recall meta plus derived recall blocks stale when queued, retry, or long-running
  worker state means stored memory/profile/session-summary context may lag behind
  raw events.

## After That

1. Turn the mocked outage/backlog eval into full session replay metrics before
   real Codex is enabled by default.
2. Build the 5-minute Hermes Memory demo: project rule, active plan, wrong
   memory, correction, supersession, explain/timeline, and private/shared scope
   check.
3. Turn the printed Hermes MCP bootstrap into an install/package command once
   the distribution format is decided.
4. Add real Hermes runtime roundtrip tests against a configured local database.

## Done Gates

- Code paths opened by the service contract have tests.
- Unsupported deterministic graph work blocks jobs instead of retrying forever.
- Raw events and derived memories remain separate.
- Agent-private retrieval always requires owner matching.
- Every operator-visible memory/recall path exposes scope and provenance.
- Corrected memory changes the next relevant recall and suppresses the old row.
- Degraded recall is labeled instead of presented as fresh memory.
- Source provenance and code header checks pass.
- Docs and review packets match the current code state.

```



<!-- Source: go.mod | bytes=522 | lines=20 | sha16=1efd3dba3b79f946 -->

```go
module github.com/parker-jungwoo-hwang/vibegravity

go 1.25.7

require (
	github.com/go-chi/chi/v5 v5.2.5
	github.com/jackc/pgx/v5 v5.9.2
	github.com/joho/godotenv v1.5.1
	github.com/pgvector/pgvector-go v0.3.0
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/text v0.29.0 // indirect
)

```



<!-- Source: go.sum | bytes=6584 | lines=77 | sha16=b3c0e3c91a88d001 -->

```text
entgo.io/ent v0.14.3 h1:wokAV/kIlH9TeklJWGGS7AYJdVckr0DloWjIcO9iIIQ=
entgo.io/ent v0.14.3/go.mod h1:aDPE/OziPEu8+OWbzy4UlvWmD2/kbRuWfK2A40hcxJM=
github.com/davecgh/go-spew v1.1.0/go.mod h1:J7Y8YcW2NihsgmVo/mv3lAwl/skON4iLHjSsI+c5H38=
github.com/davecgh/go-spew v1.1.1 h1:vj9j/u1bqnvCEfJOwUhtlOARqs3+rkHYY13jYWTU97c=
github.com/davecgh/go-spew v1.1.1/go.mod h1:J7Y8YcW2NihsgmVo/mv3lAwl/skON4iLHjSsI+c5H38=
github.com/go-chi/chi/v5 v5.2.5 h1:Eg4myHZBjyvJmAFjFvWgrqDTXFyOzjj7YIm3L3mu6Ug=
github.com/go-chi/chi/v5 v5.2.5/go.mod h1:X7Gx4mteadT3eDOMTsXzmI4/rwUpOwBHLpAfupzFJP0=
github.com/go-pg/pg/v10 v10.11.0 h1:CMKJqLgTrfpE/aOVeLdybezR2om071Vh38OLZjsyMI0=
github.com/go-pg/pg/v10 v10.11.0/go.mod h1:4BpHRoxE61y4Onpof3x1a2SQvi9c+q1dJnrNdMjsroA=
github.com/go-pg/zerochecker v0.2.0 h1:pp7f72c3DobMWOb2ErtZsnrPaSvHd2W4o9//8HtF4mU=
github.com/go-pg/zerochecker v0.2.0/go.mod h1:NJZ4wKL0NmTtz0GKCoJ8kym6Xn/EQzXRl2OnAe7MmDo=
github.com/google/uuid v1.6.0 h1:NIvaJDMOsjHA8n1jAhLSgzrAzy1Hgr+hNrb57e+94F0=
github.com/google/uuid v1.6.0/go.mod h1:TIyPZe4MgqvfeYDBFedMoGGpEw/LqOeaOT+nhxU+yHo=
github.com/jackc/pgpassfile v1.0.0 h1:/6Hmqy13Ss2zCq62VdNG8tM1wchn8zjSGOBJ6icpsIM=
github.com/jackc/pgpassfile v1.0.0/go.mod h1:CEx0iS5ambNFdcRtxPj5JhEz+xB6uRky5eyVu/W2HEg=
github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 h1:iCEnooe7UlwOQYpKFhBabPMi4aNAfoODPEFNiAnClxo=
github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761/go.mod h1:5TJZWKEWniPve33vlWYSoGYefn3gLQRzjfDlhSJ9ZKM=
github.com/jackc/pgx/v5 v5.9.2 h1:3ZhOzMWnR4yJ+RW1XImIPsD1aNSz4T4fyP7zlQb56hw=
github.com/jackc/pgx/v5 v5.9.2/go.mod h1:mal1tBGAFfLHvZzaYh77YS/eC6IX9OWbRV1QIIM0Jn4=
github.com/jackc/puddle/v2 v2.2.2 h1:PR8nw+E/1w0GLuRFSmiioY6UooMp6KJv0/61nB7icHo=
github.com/jackc/puddle/v2 v2.2.2/go.mod h1:vriiEXHvEE654aYKXXjOvZM39qJ0q+azkZFrfEOc3H4=
github.com/jinzhu/inflection v1.0.0 h1:K317FqzuhWc8YvSVlFMCCUb36O/S9MCKRDI7QkRKD/E=
github.com/jinzhu/inflection v1.0.0/go.mod h1:h+uFLlag+Qp1Va5pdKtLDYj+kHp5pxUVkryuEj+Srlc=
github.com/jinzhu/now v1.1.5 h1:/o9tlHleP7gOFmsnYNz3RGnqzefHA47wQpKrrdTIwXQ=
github.com/jinzhu/now v1.1.5/go.mod h1:d3SSVoowX0Lcu0IBviAWJpolVfI5UJVZZ7cO71lE/z8=
github.com/jmoiron/sqlx v1.3.5 h1:vFFPA71p1o5gAeqtEAwLU4dnX2napprKtHr7PYIcN3g=
github.com/jmoiron/sqlx v1.3.5/go.mod h1:nRVWtLre0KfCLJvgxzCsLVMogSvQ1zNJtpYr2Ccp0mQ=
github.com/joho/godotenv v1.5.1 h1:7eLL/+HRGLY0ldzfGMeQkb7vMd0as4CfYvUVzLqw0N0=
github.com/joho/godotenv v1.5.1/go.mod h1:f4LDr5Voq0i2e/R5DDNOoa2zzDfwtkZa6DnEwAbqwq4=
github.com/lib/pq v1.10.9 h1:YXG7RB+JIjhP29X+OtkiDnYaXQwpS4JEWq7dtCCRUEw=
github.com/lib/pq v1.10.9/go.mod h1:AlVN5x4E4T544tWzH6hKfbfQvm3HdbOxrmggDNAPY9o=
github.com/pgvector/pgvector-go v0.3.0 h1:Ij+Yt78R//uYqs3Zk35evZFvr+G0blW0OUN+Q2D1RWc=
github.com/pgvector/pgvector-go v0.3.0/go.mod h1:duFy+PXWfW7QQd5ibqutBO4GxLsUZ9RVXhFZGIBsWSA=
github.com/pmezard/go-difflib v1.0.0 h1:4DBwDE0NGyQoBHbLQYPwSUPoCMWR5BEzIk/f1lZbAQM=
github.com/pmezard/go-difflib v1.0.0/go.mod h1:iKH77koFhYxTK1pcRnkKkqfTogsbg7gZNVY4sRDYZ/4=
github.com/stretchr/objx v0.1.0/go.mod h1:HFkY916IF+rwdDfMAkV7OtwuqBVzrE8GR6GFx+wExME=
github.com/stretchr/testify v1.3.0/go.mod h1:M5WIy9Dh21IEIfnGCwXGc5bZfKNJtfHm1UVUgZn+9EI=
github.com/stretchr/testify v1.7.0/go.mod h1:6Fq8oRcR53rry900zMqJjRRixrwX3KX962/h/Wwjteg=
github.com/stretchr/testify v1.11.1 h1:7s2iGBzp5EwR7/aIZr8ao5+dra3wiQyKjjFuvgVKu7U=
github.com/stretchr/testify v1.11.1/go.mod h1:wZwfW3scLgRK+23gO65QZefKpKQRnfz6sD981Nm4B6U=
github.com/tmthrgd/go-hex v0.0.0-20190904060850-447a3041c3bc h1:9lRDQMhESg+zvGYmW5DyG0UqvY96Bu5QYsTLvCHdrgo=
github.com/tmthrgd/go-hex v0.0.0-20190904060850-447a3041c3bc/go.mod h1:bciPuU6GHm1iF1pBvUfxfsH0Wmnc2VbpgvbI9ZWuIRs=
github.com/uptrace/bun v1.1.12 h1:sOjDVHxNTuM6dNGaba0wUuz7KvDE1BmNu9Gqs2gJSXQ=
github.com/uptrace/bun v1.1.12/go.mod h1:NPG6JGULBeQ9IU6yHp7YGELRa5Agmd7ATZdz4tGZ6z0=
github.com/uptrace/bun/dialect/pgdialect v1.1.12 h1:m/CM1UfOkoBTglGO5CUTKnIKKOApOYxkcP2qn0F9tJk=
github.com/uptrace/bun/dialect/pgdialect v1.1.12/go.mod h1:Ij6WIxQILxLlL2frUBxUBOZJtLElD2QQNDcu/PWDHTc=
github.com/uptrace/bun/driver/pgdriver v1.1.12 h1:3rRWB1GK0psTJrHwxzNfEij2MLibggiLdTqjTtfHc1w=
github.com/uptrace/bun/driver/pgdriver v1.1.12/go.mod h1:ssYUP+qwSEgeDDS1xm2XBip9el1y9Mi5mTAvLoiADLM=
github.com/vmihailenco/bufpool v0.1.11 h1:gOq2WmBrq0i2yW5QJ16ykccQ4wH9UyEsgLm6czKAd94=
github.com/vmihailenco/bufpool v0.1.11/go.mod h1:AFf/MOy3l2CFTKbxwt0mp2MwnqjNEs5H/UxrkA5jxTQ=
github.com/vmihailenco/msgpack/v5 v5.3.5 h1:5gO0H1iULLWGhs2H5tbAHIZTV8/cYafcFOr9znI5mJU=
github.com/vmihailenco/msgpack/v5 v5.3.5/go.mod h1:7xyJ9e+0+9SaZT0Wt1RGleJXzli6Q/V5KbhBonMG9jc=
github.com/vmihailenco/tagparser v0.1.2 h1:gnjoVuB/kljJ5wICEEOpx98oXMWPLj22G67Vbd1qPqc=
github.com/vmihailenco/tagparser v0.1.2/go.mod h1:OeAg3pn3UbLjkWt+rN9oFYB6u/cQgqMEUPoW2WPyhdI=
github.com/vmihailenco/tagparser/v2 v2.0.0 h1:y09buUbR+b5aycVFQs/g70pqKVZNBmxwAhO7/IwNM9g=
github.com/vmihailenco/tagparser/v2 v2.0.0/go.mod h1:Wri+At7QHww0WTrCBeu4J6bNtoV6mEfg5OIWRZA9qds=
github.com/x448/float16 v0.8.4 h1:qLwI1I70+NjRFUR3zs1JPUCgaCXSh3SW62uAKT1mSBM=
github.com/x448/float16 v0.8.4/go.mod h1:14CWIYCyZA/cWjXOioeEpHeN/83MdbZDRQHoFcYsOfg=
golang.org/x/crypto v0.36.0 h1:AnAEvhDddvBdpY+uR+MyHmuZzzNqXSe/GvuDeob5L34=
golang.org/x/crypto v0.36.0/go.mod h1:Y4J0ReaxCR1IMaabaSMugxJES1EpwhBHhv2bDHklZvc=
golang.org/x/sync v0.17.0 h1:l60nONMj9l5drqw6jlhIELNv9I0A4OFgRsG9k2oT9Ug=
golang.org/x/sync v0.17.0/go.mod h1:9KTHXmSnoGruLpwFjVSX0lNNA75CykiMECbovNTZqGI=
golang.org/x/sys v0.31.0 h1:ioabZlmFYtWhL+TRYpcnNlLwhyxaM9kWTDEmfnprqik=
golang.org/x/sys v0.31.0/go.mod h1:BJP2sWEmIv4KK5OTEluFJCKSidICx8ciO85XgH3Ak8k=
golang.org/x/text v0.29.0 h1:1neNs90w9YzJ9BocxfsQNHKuAT4pkghyXc4nhZ6sJvk=
golang.org/x/text v0.29.0/go.mod h1:7MhJOA9CD2qZyOKYazxdYMF85OwPdEr9jTtBpO7ydH4=
gopkg.in/check.v1 v0.0.0-20161208181325-20d25e280405/go.mod h1:Co6ibVJAznAaIkqp8huTwlJQCZ016jof/cbN4VW5Yz0=
gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c/go.mod h1:K4uyk7z7BCEPqu6E+C64Yfv1cQ7kz7rIZviUmN+EgEM=
gopkg.in/yaml.v3 v3.0.1 h1:fxVm/GzAzEWqLHuvctI91KS9hhNmmWOoWu0XTYJS7CA=
gopkg.in/yaml.v3 v3.0.1/go.mod h1:K4uyk7z7BCEPqu6E+C64Yfv1cQ7kz7rIZviUmN+EgEM=
gorm.io/driver/postgres v1.5.4 h1:Iyrp9Meh3GmbSuyIAGyjkN+n9K+GHX9b9MqsTL4EJCo=
gorm.io/driver/postgres v1.5.4/go.mod h1:Bgo89+h0CRcdA33Y6frlaHHVuTdOf87pmyzwW9C/BH0=
gorm.io/gorm v1.25.5 h1:zR9lOiiYf09VNh5Q1gphfyia1JpiClIWG9hQaxB/mls=
gorm.io/gorm v1.25.5/go.mod h1:hbnx/Oo0ChWMn1BIhpy1oYozzpM15i4YPuHDmfYtwg8=
mellium.im/sasl v0.3.1 h1:wE0LW6g7U83vhvxjC1IY8DnXM+EU095yeo8XClvCdfo=
mellium.im/sasl v0.3.1/go.mod h1:xm59PUYpZHhgQ9ZqoJ5QaCqzWMi8IeS49dhp6plPCzw=

```
