# VibeGravity Agent Workflows

This directory defines the repo-local workflow contract for VibeGravity agents.
Use it with `.agents/coordination/agent-work.sh`, not as a replacement for the
live claim board.

## Required Startup

Every agent starts by reading:

- `AGENTS.md`
- `.agents/coordination/WORK_PROGRESS.md`
- `.agents/workflows/quickstart.md`
- The role file for its assigned role
- `.agents/workflows/phase_context.md`

If the task may edit files, the agent must claim exact file paths before
editing. If the task is read-only, the agent must not claim or edit files.

## Authority Model

Only the leader can approve lane widening.

Only the leader can approve final synthesis.

Non-leader agents may recommend widening or final synthesis, but they must leave
the recommendation in their handoff and wait for leader approval before acting.

## Lane Types

Use exactly one of these lane types in every handoff:

- `read_only_review`
- `docs_only`
- `tests_only`
- `code_edit`
- `integration_synthesis`
- `release_readiness`

Lane type controls the allowed write surface:

| Lane type | Allowed work |
|---|---|
| `read_only_review` | Inspect files and write findings in chat only. No repo edits. |
| `docs_only` | Edit documentation, prompts, review packets, or planning docs. |
| `tests_only` | Add or edit tests without changing production behavior. |
| `code_edit` | Change production code with tests and docs as needed. |
| `integration_synthesis` | Combine completed lane results into one synthesis packet. Leader approval required. |
| `release_readiness` | Verify release gates and publish readiness notes. No scope expansion without leader approval. |

## Handoff Front Matter

Every saved handoff, review packet, synthesis packet, or role result document
must begin with YAML front matter.

Required fields:

```yaml
---
agent_id: codex-main-example
role: backend-dev
phase_id: phase-05-quality
lane_id: recall-budget-tests
lane_type: tests_only
claimed_files:
  - internal/recall/assembler_test.go
reviewed_files:
  - internal/recall/assembler.go
changed_files:
  - internal/recall/assembler_test.go
gates_run:
  - go test ./internal/recall
gates_skipped:
  - make lint
skip_reasons:
  make lint: "Docs/test lane did not touch shared lint-sensitive code."
next_owner: leader
---
```

Rules:

- `agent_id`, `role`, `phase_id`, and `lane_id` are always required. Use the
  role file name when a role file exists, or `hermes-orchestration` for
  generated multi-Hermes synthesis packets.
- `claimed_files`, `reviewed_files`, and `changed_files` are always required.
  Use `[]` when empty.
- `gates_run`, `gates_skipped`, and `skip_reasons` are always required. Use
  `[]` or `{}` when empty.
- `next_owner` is always required. Use `leader` when the next decision needs
  lane widening, final synthesis, or release judgment.
- If `lane_type` is `read_only_review`, `claimed_files` and `changed_files`
  must be `[]`.

## Multi-Hermes Runs

Every multi-Hermes run must leave a synthesis packet under:

```text
.agents/hermes-orchestration/runs/<run-id>/synthesis.md
```

The packet must use the same YAML front matter contract and summarize:

- profiles dispatched
- task prompts used
- exit codes
- changed files reported by each profile
- gates reported by each profile
- conflicts, blockers, and leader decision needed

The synthesis packet is a record of results, not approval to merge, widen, or
declare completion. Final synthesis still requires leader approval.
