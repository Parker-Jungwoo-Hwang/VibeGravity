# Agent Skills

`.agents` is internal tooling for VibeGravity development. It is not a public
Hermes Memory product workflow yet.

## Skill Classes

- `policy-only`: mandatory policy and risk guardrails. It can include a
  handoff template, but it is not a runnable checklist.
- `checklist`: repeatable human/agent procedure with concrete gates or
  review output.
- `executable-backed`: procedure tied to an actual repo command, fixture, or
  script that can be run and verified.

## Current Skills

| Skill | Class | Backing | Notes |
|---|---|---|
| `code-headers.md` | executable-backed | `make check-headers` | Explains required Go file header policy and the canonical header gate. |
| `contract-check.md` | checklist | Contract docs under `plans/` | Guides contract review against runtime and storage invariants. |
| `eval-regression.md` | executable-backed | `make eval`, `tests/golden/replay_eval.json` | Maps deterministic eval scenarios to the actual runner and manifest. |
| `plan-implement-verify.md` | checklist | Repo gate bundle | Guides scoped implementation, verification, and handoff. |
| `source-provenance.md` | policy-only | AGENTS open-source policy | Captures open-source provenance and similarity-risk rules. |

## Gates

Use these gates for real work:

- `go test ./...`
- `make lint`
- `make check-headers`
- `git diff --check`
- `make eval`
- `make integration-postgres` when live DB behavior is touched

Do not add aspirational skill names. Names such as `context-router`,
`team-manager`, or `release-manager` should only be created after the repo has a
real command, owner, tests, and output contract for that workflow.
