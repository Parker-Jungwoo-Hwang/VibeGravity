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
2. Read `.agents/workflows/quickstart.md`.
3. Read the role file under `.agents/workflows/` when a role is assigned.
4. Claim the exact files to edit before modifying them.
5. Stop if another active agent already owns any of those files.
6. Send heartbeats during long work or before requesting wider scope.
7. Release a file immediately when work on that file is done.
8. Run verification and leave a result note or review packet when the lane ends.

Do not claim broad globs such as `internal/**`. Claim concrete file paths.
The claim tool rejects flag-like paths such as `--`, broad globs, directory
claims, parent traversal, and paths with whitespace.

## Workflow Roles

Workflow definitions live in:

```text
.agents/workflows/
```

Use `.agents/workflows/README.md` for the shared lane and handoff contract.

Only the leader can approve lane widening.

Only the leader can approve final synthesis.

Quickstart is read-only. It may recommend a lane, but it must not claim files or
edit the repo.

Every saved handoff, review packet, role result, or synthesis packet must start
with YAML front matter containing:

- `agent_id`, `role`, `phase_id`, `lane_id`, `lane_type`
- `claimed_files`, `reviewed_files`, `changed_files`
- `gates_run`, `gates_skipped`, `skip_reasons`
- `next_owner`

Allowed lane types are:

- `read_only_review`
- `docs_only`
- `tests_only`
- `code_edit`
- `integration_synthesis`
- `release_readiness`

## Universal Prompt

Use `.agents/coordination/UNIVERSAL_AGENT_PROMPT.md` when launching a new
autonomous agent and you want it to decide whether to implement, test, review,
document, or hand off the next useful slice.

If an operator gives an agent only this path, the expected behavior is to read
the file and execute it immediately:

```text
.agents/coordination/UNIVERSAL_AGENT_PROMPT.md
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

Get machine-readable current claims without rewriting `WORK_PROGRESS.md`:

```bash
.agents/coordination/agent-work.sh status --json --no-render
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
- If `status` reports a stale claim warning, treat it as a warning, not an
  automatic release. A leader or operator decides whether stale ownership can be
  cleared.
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
