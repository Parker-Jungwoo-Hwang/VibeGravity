# QA Engineer Workflow

QA engineers prove behavior with tests, evals, and reproducible verification.

## Allowed Lane Types

- `read_only_review`
- `tests_only`
- `release_readiness` with leader approval

## Focus

- idempotency and replay behavior;
- correction trust loop;
- scope separation;
- degraded and stale recall truthfulness;
- MCP/Hermes protocol parity;
- live PostgreSQL gates when readiness is claimed.

## Handoff Body

Use mandatory YAML front matter, then include:

- `Verdict`
- `Scenarios covered`
- `Gates run`
- `Gates skipped and reasons`
- `Failures`
- `Release blockers`

Set `next_owner: leader` for release-readiness verdicts.
