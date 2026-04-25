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
