# Skills and Internal Tools Review

Date: 2026-04-26
Reviewer: Agent 07
Role: CTO review
Scope: Skill implementation, scripts, internal tools, maintainability, and code quality inside `.agents/skills` and adjacent repo-local agent tooling
Repo: `/Users/parker/Documents/VibeGravity`

## Original Request

Act as a CTO reviewing the GitHub repo "VibeGravity".

Review scope:

- Skill implementation
- Scripts
- Internal tools
- Maintainability
- Code quality inside `.agent/skills`

Clarification from repo state:

- The live checkout contains `.agents/skills`, not `.agent/skills`.
- There are no repo-local `SKILL.md` files under `.agents/skills`.
- The live repo uses Markdown skill files such as `.agents/skills/code-headers.md`.
- The live repo also contains adjacent executable agent tooling under `.agents/coordination` and `.agents/hermes-orchestration`.

The requested named surfaces were:

- `.agent/skills`
- `SKILL.md` files
- `scripts` folders
- data files
- `context-router`
- `context-manager`
- `code-reviewer`
- `team-manager`
- `deployment-wizard`
- `template-marketplace`
- `test-generator`
- `release-manager`

Only the live repo surfaces were treated as source of truth. Missing requested surfaces were recorded as absent rather than inferred.

## Review Questions

This report answers:

1. Which skills have executable scripts?
2. Which skills are mostly prompt/docs?
3. Do `SKILL.md` files match the actual scripts?
4. Are script interfaces consistent?
5. Are outputs machine-readable?
6. Are there hidden assumptions?
7. Are scripts robust across Windows, macOS, and Linux?
8. Are there duplicate tools?
9. Which tools should become core?
10. Which tools should be removed or rewritten?

## Executive Verdict

VibeGravity has a small, disciplined repo-local skill surface, not the broader skill platform implied by the requested names.
The five files under `.agents/skills` are useful agent playbooks, but they are mostly Markdown procedures rather than executable skills.
The real internal tools are adjacent scripts: `.agents/coordination/agent-work.sh`, `.agents/hermes-orchestration/*.sh`, and the Go header checker wired through `make check-headers`.
At current size, this system is maintainable because it is compact and easy for agents to read.
It is not yet maintainable as a larger platform because interfaces, output schemas, tests, portability, and ownership boundaries are not standardized.
The best v1 direction is to keep the skill set small, promote only proven tools into core, and avoid adding named skills unless they ship with executable behavior and tests.
The current skills are good as operating doctrine, but weak as reliable automation.

## Method

Reviewed repo-local files and directories:

- `.agents/skills`
- `.agents/coordination`
- `.agents/hermes-orchestration`
- `tools/headercheck/main.go`
- `Makefile`
- `AGENTS.md`
- `plans/00_read-this-first_for-building-agents.md`
- `plans/01_rfp_vibegravity_hermes-first.md`
- `plans/02_product-contract_and_direction.md`
- `plans/03_target-architecture_codex-first.md`
- `plans/05_runtime-contracts_ingest-recall-apply.md`
- `plans/06_data-model_and_storage-invariants.md`
- `c-level-review/onboarding-and-activation-review.md`

Validation performed:

- Enumerated `.agents/skills` files.
- Enumerated executable files under `.agents`.
- Compared skill Markdown claims against available scripts and Makefile targets.
- Checked shell script syntax with `sh -n` and `bash -n`.
- Exercised help/usage paths for script interface shape.
- Inspected ignored/generated state behavior for coordination and Hermes run outputs.

I did not dispatch live Hermes agents as part of this review because this report is about maintainability and script reality, not running a new multi-agent review lane.

## Skill Inventory

### Repo-Local Skills

The live `.agents/skills` directory contains:

| Skill file | Type | Executable backing | Maturity |
|---|---|---|---|
| `.agents/skills/code-headers.md` | Procedure | `make check-headers`, `tools/headercheck/main.go` | High |
| `.agents/skills/contract-check.md` | Review checklist | None | Medium-low |
| `.agents/skills/eval-regression.md` | Evaluation checklist | Indirectly related to `make eval`, but not wired | Medium-low |
| `.agents/skills/plan-implement-verify.md` | Workflow checklist | Indirectly related to Makefile gates | Medium |
| `.agents/skills/source-provenance.md` | Policy checklist | None | Medium |

### Adjacent Internal Tools

The executable agent tooling lives outside `.agents/skills`:

| Tool | Path | Purpose | Maturity |
|---|---|---|---|
| Coordination board | `.agents/coordination/agent-work.sh` | Claim, heartbeat, release, and done workflow for multi-agent file ownership | Medium-high |
| Hermes single-run wrapper | `.agents/hermes-orchestration/run-agent.sh` | Run one Hermes profile with one prompt | Medium |
| Hermes dispatcher | `.agents/hermes-orchestration/dispatch.sh` | Run multiple profile prompts in parallel | Medium |
| Hermes collector | `.agents/hermes-orchestration/collect.sh` | Print collected meta, stdout, and stderr for a run | Medium-low |
| Hermes profile status | `.agents/hermes-orchestration/status.sh` | Show local Hermes profiles | Low |
| Go header checker | `tools/headercheck/main.go` | Enforce machine-readable Go file headers | High |

### Missing Requested Skills

The following requested skill/tool names do not exist in the live checkout:

- `context-router`
- `context-manager`
- `code-reviewer`
- `team-manager`
- `deployment-wizard`
- `template-marketplace`
- `test-generator`
- `release-manager`

I found no matching directories, `SKILL.md` files, script folders, or data files for these names.
They should not be treated as available repo tools.

## Executable Reality Check

### Skills With Executable Scripts

No skill under `.agents/skills` contains its own executable script folder.
The closest executable mappings are:

- `code-headers` maps to `make check-headers`.
- `plan-implement-verify` maps partially to `go test ./...` and `golangci-lint run`, but the repo's actual durable path is `make lint` and `make check-headers`.
- `eval-regression` maps conceptually to `make eval`, but the skill does not name specific golden files or commands.
- `contract-check` has no script.
- `source-provenance` has no script.

### Skills That Are Mostly Prompt or Docs

All five `.agents/skills` files are mostly prompt/docs.
This is not automatically bad. VibeGravity intentionally moves repeated procedures into skill files so agents can load them late.
However, these skills should not be described as automation.
They are operational doctrine with one strong enforcement hook: `code-headers`.

### Usable Tools

The most usable internal tools are:

1. `tools/headercheck/main.go`
2. `make check-headers`
3. `.agents/coordination/agent-work.sh`
4. `.agents/hermes-orchestration/run-agent.sh`
5. `.agents/hermes-orchestration/dispatch.sh`
6. `.agents/hermes-orchestration/collect.sh`

The header checker is the strongest because it is implemented in Go, dependency-light, cross-platform by Go standards, and integrated into the Makefile.

The coordination script is operationally useful but is a shell tool, not a durable product-grade command.

The Hermes orchestration scripts are useful for this local operator environment, but they are not portable repo tooling yet.

## Script Interface Review

### `.agents/coordination/agent-work.sh`

Supported commands:

- `init`
- `status`
- `claim <agent-id> <task> <file> [<file> ...]`
- `heartbeat <agent-id> <note>`
- `release <agent-id> <file> [<file> ...]`
- `done <agent-id> <note>`

Strengths:

- Clear command verbs.
- Handles lock acquisition with a lock directory.
- Uses exact file paths rather than broad ownership scopes.
- Normalizes absolute repo paths to relative paths.
- Produces a human-readable Markdown board.
- Keeps live state files git-ignored.
- The interface matches the multi-agent workflow in `AGENTS.md`.

Weaknesses:

- No JSON output.
- `status` rewrites `WORK_PROGRESS.md`, so read-only status is not actually read-only.
- No stale-claim expiry or owner liveness model.
- No schema validation for agent IDs, task names, or file existence.
- Uses TSV as internal state, which is simple but fragile if fields grow.
- Machine consumers must parse Markdown or TSV manually.

CTO view:

This should become core, but not in its current shell-only shape.
It deserves a stable JSON status mode and a small test suite.

### `.agents/hermes-orchestration/run-agent.sh`

Supported interface:

```bash
run-agent.sh <profile> <prompt-file> [run-id]
```

Environment overrides:

- `HERMES_MODEL`
- `HERMES_PROVIDER`
- `HERMES_MAX_TURNS`

Strengths:

- Simple interface.
- Captures output, stderr, and metadata per run.
- Avoids mutating Hermes config by setting `HERMES_HOME` for the child process.
- Supports predictable profile names: `default`, `vuitton`, `bottega`.
- Writes useful meta fields such as profile, model, provider, timestamps, and exit code.

Weaknesses:

- Hard-codes `/Users/parker/.hermes` paths.
- Assumes the `hermes` CLI is installed and compatible.
- Assumes Bash.
- Assumes the local operator profile layout.
- Uses free-form Markdown output from agents.
- Does not validate prompt size, run ID safety, or profile config existence beyond fixed cases.
- Does not produce a single machine-readable result envelope.

CTO view:

Keep it local and experimental unless profile configuration is externalized.

### `.agents/hermes-orchestration/dispatch.sh`

Supported interface:

```bash
dispatch.sh <run-id> <manifest.tsv>
```

Manifest format:

```text
<profile><TAB><prompt-file>
```

Strengths:

- Clear and minimal.
- Runs multiple profiles in parallel.
- Propagates failures into non-zero final status.
- Rejects manifest lines with too many fields.

Weaknesses:

- TSV manifest has no schema version.
- No JSON summary.
- No per-task timeout.
- No concurrency limit.
- No cancellation or cleanup behavior.
- No validation that prompt files include coordination snippets when edits are allowed.
- No portability beyond Unix-like shells.

CTO view:

Useful for local operator review lanes, not ready as a general team-manager primitive.

### `.agents/hermes-orchestration/collect.sh`

Supported interface:

```bash
collect.sh <run-id>
```

Strengths:

- Easy to read.
- Prints meta, output, and stderr by profile.
- Good enough for manual synthesis after local Hermes runs.

Weaknesses:

- `--help` exits with code 2 rather than 0.
- Output is deliberately human-readable, not machine-readable.
- Truncates output with `sed -n '1,220p'` and stderr with `sed -n '1,120p'`.
- Does not report missing profile outputs as structured failures.
- Does not aggregate exit codes into a machine-consumable summary.

CTO view:

Rewrite or extend this first if Hermes orchestration becomes a real internal platform.
The first addition should be `collect.sh --json <run-id>`.

### `.agents/hermes-orchestration/status.sh`

Supported behavior:

- Runs `hermes profile list`.
- Shows `default`, `vuitton`, and `bottega`.

Strengths:

- Tiny and readable.
- Useful for the current operator's local Hermes setup.

Weaknesses:

- Hard-coded profiles.
- No argument handling.
- No fallback if a profile is missing.
- No JSON mode.
- Not portable outside this machine.

CTO view:

This is a convenience script, not a core tool.

### `tools/headercheck/main.go`

Strengths:

- Real executable implementation.
- Dependency-light.
- Integrated through `make check-headers`.
- Enforces required fields:
  - `FILE`
  - `PURPOSE`
  - `LAYER`
  - `STATUS`
  - `EXPORTS`
  - `DEPENDS`
  - `USED_BY`
  - `AGENT_NOTE`
- Validates layer and status enums.
- Skips generated Go files.
- Uses sorted failures for stable output.

Weaknesses:

- Does not produce JSON output.
- Does not expose an autofix mode.
- Only covers Go files.

CTO view:

This is the most production-like internal tool in the reviewed scope.
It should remain core.

## Do `SKILL.md` Files Match Actual Scripts?

There are no repo-local `SKILL.md` files in the reviewed scope.

The live repo uses:

- `.agents/skills/code-headers.md`
- `.agents/skills/contract-check.md`
- `.agents/skills/eval-regression.md`
- `.agents/skills/plan-implement-verify.md`
- `.agents/skills/source-provenance.md`

The prompt's `SKILL.md` assumption does not match this repo.

Match assessment for actual skill Markdown:

| Skill file | Match to actual tooling | Verdict |
|---|---|---|
| `code-headers.md` | Matches `make check-headers` and header policy | Good |
| `contract-check.md` | Matches contract docs, but no executable checker | Partial |
| `eval-regression.md` | Lists scenarios, but does not map them to concrete eval commands | Partial |
| `plan-implement-verify.md` | Mentions tests and lint, but misses full current gate set | Partial |
| `source-provenance.md` | Matches repo policy, but no enforcement tool | Partial |

The biggest mismatch is that the skill layer sounds like a capability system, while the implementation is mostly guidance.
That is acceptable if documented honestly.
It becomes risky if future agents assume a skill name implies a runnable tool.

## Are Script Interfaces Consistent?

Partially.

Consistent elements:

- Shell scripts use explicit positional arguments.
- Help text exists for the main scripts.
- Exit code 2 is used for usage and validation errors.
- Hermes orchestration writes files under a predictable `runs/<run-id>` directory.
- Coordination tooling uses a stable command vocabulary.

Inconsistent elements:

- `run-agent.sh --help` and `dispatch.sh --help` exit 0, while `collect.sh --help` exits 2.
- `agent-work.sh status` has side effects by rendering `WORK_PROGRESS.md`.
- Coordination output is Markdown and TSV.
- Hermes output is key-value meta, Markdown stdout, stderr logs, and human-readable collection output.
- `make lint` is the repo's durable lint wrapper, but `plan-implement-verify.md` names `golangci-lint run` directly.
- There is no shared convention for `--json`, `--quiet`, `--no-write`, or `--strict`.

CTO standard to adopt:

Every durable internal tool should support:

```text
<tool> --help
<tool> --json ...
<tool> --dry-run ...
```

Every durable internal tool should document:

- arguments
- environment variables
- outputs
- exit codes
- files written
- portability assumptions

## Are Outputs Machine-Readable?

Mostly no.

Machine-readable or semi-readable outputs:

- `agent-work.sh` internal state uses TSV in `.agents/coordination/claims.tsv` and `.agents/coordination/activity.log`.
- `run-agent.sh` writes simple key-value `.meta` files.
- `tools/headercheck/main.go` produces stable text failures.

Human-readable outputs:

- `WORK_PROGRESS.md`
- Hermes `.out.md` result files
- `collect.sh` output
- skill Markdown checklists

Missing:

- JSON status for coordination.
- JSON run summary for Hermes orchestration.
- JSON or JUnit-style output for eval regression scenarios.
- A structured contract-check result format.
- A machine-readable inventory of skills, scripts, and expected commands.

Recommendation:

Add machine-readable modes only to tools that should become core.
Do not over-engineer throwaway local scripts.

## Hidden Assumptions

### Repo Path and Naming

The request references `.agent/skills`, but the repo uses `.agents/skills`.
This is a small naming mismatch, but it matters for automation.
Any future installer, CI job, or agent prompt must use the live `.agents` path.

### Skill Format

The request references `SKILL.md`, but this repo uses one Markdown file per skill.
No local skill loader contract is visible in the repo.
Agents understand these files by convention, not because the repo enforces a skill package format.

### Operating System

Shell tools assume Unix-like behavior.
`agent-work.sh` is POSIX shell.
Hermes orchestration scripts are Bash.
Windows support would require WSL, Git Bash, or rewrites.

### User Environment

Hermes orchestration assumes:

- user home path `/Users/parker`;
- Hermes config exists under `/Users/parker/.hermes`;
- profiles `default`, `vuitton`, and `bottega` exist;
- `hermes` CLI is installed;
- `openai-codex` provider works;
- `gpt-5.5` is the default model.

These are valid local operator assumptions, but not general repo assumptions.

### Tool Availability

Scripts assume:

- `git`
- `bash`
- `sh`
- `awk`
- `sed`
- `tail`
- `date`
- `mkdir`
- `mv`
- `cp`
- `hermes`

These are fine on macOS/Linux, weak on native Windows.

### State Model

Coordination claims are advisory, not enforceable file locks.
They prevent agent collisions only if agents follow the protocol.

### Review Scope Drift

The skill docs repeat concepts already present in `AGENTS.md`, plan docs, and review packets.
This helps context loading, but creates drift risk.

## Cross-Platform Robustness

### macOS

Current tools are strongest on macOS because the local operator environment is macOS and Hermes profile paths are macOS-specific.
The scripts are likely usable in the current checkout.

### Linux

`agent-work.sh` should mostly work on Linux.
Hermes orchestration may work only if Hermes is installed and profile paths are adapted.
Hard-coded `/Users/parker` paths are the main blocker.

### Windows

Native Windows support is weak.
Problems:

- Bash dependency.
- POSIX tools dependency.
- path assumptions.
- Hermes profile path assumptions.
- Makefile dependency.

Recommendation:

Do not claim Windows support for `.agents` tooling.
If cross-platform support becomes important, move core internal tools into Go commands under `cmd/cli` or `tools/`.

## Duplicate Tools and Overlap

### Duplicates

No exact duplicate executable tools were found.
However, there is conceptual overlap:

- `contract-check.md`, `plan-implement-verify.md`, `AGENTS.md`, and `UNIVERSAL_AGENT_PROMPT.md` all describe verification and contract review expectations.
- `eval-regression.md`, `make eval`, and review-packet gate lists overlap but are not linked by a single manifest.
- `code-headers.md`, `docs/code-header-policy.md`, templates under `plans/templates`, and `tools/headercheck/main.go` overlap intentionally and are well aligned.
- Hermes orchestration task prompts repeat read-first documents and stop-lines per profile.

### Healthy Duplication

`code-headers` duplication is healthy because policy, template, and checker each have a separate job.

### Risky Duplication

`plan-implement-verify` and `contract-check` are at risk of drifting from the actual repo gates.
The active repo repeatedly relies on:

- `go test ./...`
- `make lint`
- `make check-headers`
- `git diff --check`
- `make eval`
- `make integration-postgres` when live PostgreSQL evidence is required

The skill docs should reflect that gate set.

## Which Tools Should Become Core?

Minimum core set for a strong v1:

1. `make check-headers` and `tools/headercheck/main.go`
2. `.agents/coordination/agent-work.sh`, after adding read-only JSON status
3. `.agents/skills/source-provenance.md`, kept as policy and handoff requirement
4. `.agents/skills/contract-check.md`, upgraded into a structured checklist tied to current plan docs
5. `.agents/skills/eval-regression.md`, upgraded into a scenario manifest tied to `make eval`
6. `make eval`, as the real regression entry point for trust-loop behavior

Hermes orchestration should remain optional local operator tooling until it has:

- externalized profile config;
- JSON output;
- tests;
- documented portability boundaries;
- clear separation between read-only review lanes and edit-capable lanes.

## Which Tools Should Be Removed or Rewritten?

### Remove

Do not remove current `.agents/skills` files.
They are small and useful.

Do not remove Hermes orchestration yet.
It provides real local value for the current operator workflow.

### Rewrite or Harden

Rewrite or harden these first:

1. `.agents/hermes-orchestration/collect.sh`
   - Add `--json`.
   - Make `--help` exit 0.
   - Report missing files and exit codes structurally.
   - Avoid silent truncation in machine mode.

2. `.agents/hermes-orchestration/run-agent.sh`
   - Move profile path mapping to config or env.
   - Validate `hermes` availability.
   - Validate `HERMES_MAX_TURNS` is numeric.
   - Emit a JSON result envelope.

3. `.agents/coordination/agent-work.sh`
   - Add `status --json`.
   - Add `status --no-render` or split read-only status from render.
   - Add optional stale-claim detection.
   - Add a small fixture-based test script.

4. `.agents/skills/eval-regression.md`
   - Map every required scenario to an actual test, golden eval, or missing-test TODO.
   - Name the exact command for each scenario.

5. `.agents/skills/plan-implement-verify.md`
   - Replace raw `golangci-lint run` with `make lint`.
   - Add `make check-headers`, `git diff --check`, and conditional `make eval`.
   - Add `make integration-postgres` for changes touching correction trust-loop readiness.

6. `.agents/skills/contract-check.md`
   - Add a structured output template:
     - blocking contract breaks;
     - risky deviations;
     - doc sync required;
     - tests required;
     - exact source docs consulted.

### Do Not Add Yet

Do not add new top-level skills named:

- `context-router`
- `context-manager`
- `code-reviewer`
- `team-manager`
- `deployment-wizard`
- `template-marketplace`
- `test-generator`
- `release-manager`

until each has:

- an owner;
- a real use case;
- a command or clear checklist;
- a test or verification story;
- a machine-readable output contract if it is automation.

Adding these names now would create "capability theater": many impressive labels, little enforceable behavior.

## Refactor Plan

### Phase 1: Align Existing Skill Docs

Files:

- `.agents/skills/plan-implement-verify.md`
- `.agents/skills/eval-regression.md`
- `.agents/skills/contract-check.md`

Recommendations:

- Update `plan-implement-verify` to use the repo's real gate vocabulary:
  - `go test ./...`
  - `make lint`
  - `make check-headers`
  - `git diff --check`
  - `make eval`
  - `make integration-postgres` when behavior depends on live PostgreSQL
- Update `eval-regression` to name concrete scenarios and their command/test location.
- Update `contract-check` to require a source-doc list and structured findings.

### Phase 2: Add Machine-Readable Core Tool Modes

Files:

- `.agents/coordination/agent-work.sh`
- `.agents/hermes-orchestration/collect.sh`
- `.agents/hermes-orchestration/run-agent.sh`

Recommendations:

- Add `agent-work.sh status --json --no-render`.
- Add `collect.sh --json <run-id>`.
- Add a `result.json` output next to each Hermes run.
- Keep human Markdown output for operator ergonomics.

### Phase 3: Externalize Hermes Profile Configuration

Files:

- `.agents/hermes-orchestration/README.md`
- `.agents/hermes-orchestration/run-agent.sh`
- new optional config file, for example `.agents/hermes-orchestration/profiles.example.tsv`

Recommendations:

- Remove hard-coded profile home paths from the script body.
- Keep `/Users/parker/.hermes` as documented local defaults, not core logic.
- Allow `HERMES_PROFILE_MANIFEST` or similar env var.

### Phase 4: Add Tests for Internal Tools

Files:

- new `tools/agenttools` or shell fixture tests
- existing `tools/headercheck/main.go`
- `.agents/coordination/agent-work.sh`

Recommendations:

- Keep `headercheck` in Go.
- Add fixture tests for claim conflict, release ownership, heartbeat, and done.
- Add a syntax test target for shell scripts.
- Prefer moving durable logic into Go if the scripts grow.

### Phase 5: Create a Skill Registry

File:

- new `.agents/skills/README.md`

Recommendations:

- List each skill.
- Say whether it is policy-only, checklist, or executable-backed.
- Name the executable command if one exists.
- Name expected output format.
- Name owner and maturity.

This one file would prevent future agents from assuming missing skills exist.

## File-Level Recommendations

### `.agents/skills/code-headers.md`

Keep.
It is short, clear, and tied to real enforcement.

Recommended change:

- Add direct pointer to `tools/headercheck/main.go`.
- Add exact command `make check-headers`.

### `.agents/skills/contract-check.md`

Keep, but strengthen.

Recommended change:

- Add explicit output template with severity levels.
- Require file and line references when reviewing code.
- Require docs updated or docs-not-needed rationale.

### `.agents/skills/eval-regression.md`

Keep, but rewrite from checklist to manifest.

Recommended change:

- For each scenario, add:
  - command;
  - expected artifact;
  - whether currently covered;
  - gap if uncovered.

### `.agents/skills/plan-implement-verify.md`

Keep, but align with actual gates.

Recommended change:

- Replace `golangci-lint run` with `make lint`.
- Add `make check-headers`.
- Add `git diff --check`.
- Add behavior-dependent `make eval`.
- Mention `make integration-postgres` for trust-loop storage changes.

### `.agents/skills/source-provenance.md`

Keep.
It is important for open-source safety.

Recommended change:

- Add one sentence clarifying that docs-only changes do not need a full source review block unless they include implementation material.

### `.agents/coordination/agent-work.sh`

Promote toward core.

Recommended change:

- Add read-only JSON status.
- Add stale claim warnings.
- Add tests.

### `.agents/hermes-orchestration/run-agent.sh`

Keep as local operator tooling.

Recommended change:

- Externalize profile paths.
- Add dependency checks.
- Add structured output.

### `.agents/hermes-orchestration/dispatch.sh`

Keep as local operator tooling.

Recommended change:

- Add manifest validation.
- Add optional concurrency control.
- Add structured summary.

### `.agents/hermes-orchestration/collect.sh`

Rewrite first among Hermes scripts.

Recommended change:

- Add JSON output.
- Make help behavior consistent.
- Avoid truncation in JSON mode.

### `.agents/hermes-orchestration/status.sh`

Keep only as convenience.

Recommended change:

- Do not make it core unless profile config becomes portable.

## Final CTO Decision

The skills system is maintainable at current size.
It is not maintainable as a larger internal platform without a stronger contract.

My decision:

- Keep the current five repo-local skills.
- Do not expand the skill list with aspirational names.
- Promote only `code-headers`, `source-provenance`, `contract-check`, `eval-regression`, and `agent-work` into the v1 core operating stack.
- Treat Hermes orchestration as useful local infrastructure, not portable product tooling.
- Require every future skill to declare whether it is policy-only, checklist-only, executable-backed, or product-core.

The current system works because it is small and human-readable.
If it grows without schemas and tests, it will become a pile of prompts.
The CTO line is simple: fewer skills, more executable proof.

## Source Review

- Estimated source: first-principles review from the live VibeGravity repo and repo-local instructions.
- Suspected license: none.
- Similarity risk: low.
- Review required: no for licensing; yes for product ownership decisions before promoting any tool into v1 core.
- Notes: This report summarizes live repo inspection and does not introduce implementation code.
