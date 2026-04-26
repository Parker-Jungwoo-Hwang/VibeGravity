# Hermes Orchestration

This folder holds repo-local helper scripts for dispatching work to the local
Hermes profiles used for VibeGravity coordination.

The scripts do not modify Hermes configuration. They do not run
`hermes config set`, `hermes profile use`, or alias management commands. Profile
selection is done only by setting `HERMES_HOME` for the child process.

## Profiles

Profile paths are configured outside the script body. Copy the example manifest
when local profile paths differ from the default `$HOME/.hermes` layout:

```bash
cp .agents/hermes-orchestration/profiles.example.tsv .agents/hermes-orchestration/profiles.tsv
```

Manifest format:

```text
default	$HOME/.hermes
vuitton	$HOME/.hermes/profiles/vuitton
bottega	$HOME/.hermes/profiles/bottega
```

`profiles.tsv` is ignored because it is local machine configuration. To use a
different config path, set `HERMES_PROFILE_MANIFEST`. If no manifest exists,
`run-agent.sh` falls back to `HERMES_PROFILE_ROOT` or `$HOME/.hermes`.

All commands default to `openai-codex` with `gpt-5.5`.

## Single Agent

```bash
.agents/hermes-orchestration/run-agent.sh default .agents/hermes-orchestration/tasks/smoke/default.md
```

Outputs are written to:

```text
.agents/hermes-orchestration/runs/<run-id>/<profile>.out.md
.agents/hermes-orchestration/runs/<run-id>/<profile>.meta
.agents/hermes-orchestration/runs/<run-id>/<profile>.result.json
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
.agents/hermes-orchestration/collect.sh --json smoke-001
```

Every multi-Hermes dispatch writes a synthesis packet:

```text
.agents/hermes-orchestration/runs/<run-id>/synthesis.md
```

`collect.sh` refreshes the same packet when results are collected. The packet
uses the YAML front matter required by `.agents/workflows/README.md` and sets
`next_owner: leader`. It records evidence only; it does not approve final
synthesis or lane widening.

## Orchestration Loop

1. Write focused task prompts under `tasks/<run-id>/`.
2. Dispatch to profiles with `dispatch.sh`.
3. Collect results with `collect.sh`.
4. Inspect `runs/<run-id>/synthesis.md`.
5. The leader approves or rejects final synthesis and any lane widening.
6. Create the next manifest if follow-up work is needed.
7. Keep all repo/code changes in this Codex session unless a Hermes result is
   explicitly selected for implementation.

## Coordination Requirement

Every Hermes task prompt that may edit repo files must include the coordination
snippet from `.agents/coordination/PROMPT_SNIPPET.md`.

Before editing, each Hermes profile must read
`.agents/coordination/WORK_PROGRESS.md`, claim exact files with
`.agents/coordination/agent-work.sh claim ...`, heartbeat during long work, and
release files immediately after finishing them.
