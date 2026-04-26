---
name: eval-regression
description: Use this skill to run or inspect golden scenarios and detect memory regressions.
---

# Eval Regression

## When to use

Use this after changing reasoning, recall, scopes, or profile logic.

## Commands

Run the executable-backed eval gate:

```bash
make eval
```

For a narrower golden-only loop, run the same command the Makefile uses:

```bash
go run ./cmd/vibegravity eval golden --path tests/golden/replay_eval.json
```

`make eval` also runs the Hermes Memory demo eval:

```bash
go run ./cmd/vibegravity eval demo
```

## Scenario Manifest

The manifest is `tests/golden/replay_eval.json`. It currently contains these
scenario groups:

| Scenario | Group | Contract covered |
|---|---|---|
| `pinned note and active plan outrank memory` | `scenarios` | Manual note and active plan priority in recall. |
| `agent private memory stays owner scoped` | `scenarios` | `agent_private` owner filtering. |
| `superseded memory is suppressed` | `scenarios` | `latest_flag` suppression. |
| `degraded recall still returns profile and summary` | `scenarios` | Stored context survives degraded recall. |
| `budget truncates low priority context` | `scenarios` | `budget_tokens` ceiling. |
| `update memory replay suppresses prior fact` | `graph_replay_scenarios` | `update_memory`, `updates`, trace, and replay idempotency. |
| `correction replay changes later recall` | `graph_replay_scenarios` | Correction-shaped supersession and next recall behavior. |
| `group shared graph write remains rejected` | `graph_replay_scenarios` | Stop-line for group-shared writes without membership validation. |
| `stage1 outage retries without graph writes` | `worker_backlog_scenarios` | Stage 1 outage retry without partial graph writes. |
| `stage2 outage recovery replay is idempotent` | `worker_backlog_scenarios` | Stage 2 recovery and replay idempotency. |
| `unsupported apply work becomes blocked` | `worker_backlog_scenarios` | Unsupported deterministic work becomes blocked, not retry spam. |

Live database behavior is not proven by these in-memory deterministic evals. If
the change touches PostgreSQL transactions, locks, migrations, FK behavior, or
live correction trust-loop behavior, also run:

```bash
make integration-postgres
```

## Output

- scenario
- expected result
- observed result
- pass or fail
- suspected cause
- next fix
