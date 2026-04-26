# Agent Workflow and Orchestration Review

Date: 2026-04-26
Reviewer: Agent 06
Role: Hybrid CTO/CPO review
Scope: Agent workflow system, role design, orchestration, parallel delegation, and handoff quality
Repo: `/Users/parker/Documents/VibeGravity`

## Executive Verdict

VibeGravity has a serious internal coordination layer, but it is not yet a reliable production-grade agent workflow system for real users.
The repo-backed system that actually exists is `.agents/coordination` plus `.agents/hermes-orchestration`, not the requested `.agent/workflows` structure.
The current system is good for expert-supervised parallel work because it has exact file claims, active claim visibility, heartbeats, release/done commands, result packets, and profile-isolated Hermes dispatch.
It is weaker as a product workflow because role definitions are mostly embedded in ad hoc prompts instead of durable role documents.
Handoff quality is useful, but not strict enough to prevent vague delegation, stale context, conceptual collisions, or overconfident completion claims.
Parallel tasks are safe only when a coordinator pre-splits non-overlapping lanes and keeps final integration authority.
My CTO/CPO decision is that this agent system is promising as an internal operating model, but not reliable enough for real users until it has canonical workflow roles, a shared phase context, stricter handoff contracts, and leader-owned conflict resolution.

## Source-of-Truth Review Notes

The requested files were:

- `.agent/workflows/leader.md`
- `.agent/workflows/quickstart.md`
- `.agent/workflows/planner.md`
- `.agent/workflows/architect.md`
- `.agent/workflows/frontend-dev.md`
- `.agent/workflows/backend-dev.md`
- `.agent/workflows/qa-engineer.md`
- `.agent/workflows/security-engineer.md`
- `.agent/workflows/tech-writer.md`
- `phase_context.md`

I did not find those files in the live checkout.
The live repo contains these relevant workflow surfaces instead:

- `.agents/coordination/README.md`
- `.agents/coordination/UNIVERSAL_AGENT_PROMPT.md`
- `.agents/coordination/PROMPT_SNIPPET.md`
- `.agents/coordination/WORK_PROGRESS.md`
- `.agents/coordination/agent-work.sh`
- `.agents/hermes-orchestration/README.md`
- `.agents/hermes-orchestration/dispatch.sh`
- `.agents/hermes-orchestration/run-agent.sh`
- `.agents/hermes-orchestration/collect.sh`
- `.agents/hermes-orchestration/tasks/*`
- `plans/12_agent-coding_playbook_codex-claude.md`
- `plans/13_handoff-prompts_and_response-templates.md`
- `docs/review-packets/*`

That absence matters.
The repo has operational coordination mechanics, but not the named product workflow system the request expected.

## Workflow Map

The current workflow has four practical phases.

### Phase 1: Agent Startup

Agents are expected to start by reading the durable repo instructions and live coordination board.
The universal prompt requires the agent to read:

- `AGENTS.md`
- `.agents/coordination/WORK_PROGRESS.md`
- `PLANS.md`
- key planning docs under `plans/`
- relevant docs under `docs/review-packets/`

The prompt also instructs the agent to run:

```bash
/Users/parker/Documents/VibeGravity/.agents/coordination/agent-work.sh status
```

This is a strong startup sequence for internal agents.
It prevents the common failure mode where a new agent starts from stale chat context instead of repo state.

### Phase 2: Lane Selection and File Claims

Before editing, agents must claim concrete files:

```bash
.agents/coordination/agent-work.sh claim "<agent-id>" "<short task>" <file> [<file> ...]
```

The documented rules are clear:

- read `WORK_PROGRESS.md`;
- claim exact files before editing;
- stop if another active agent owns a file;
- heartbeat during long work;
- release files when finished;
- leave a result note or review packet at lane end.

This is the best part of the current system.
It gives agents a simple, local locking protocol that works without external services.

### Phase 3: Execution or Review

The universal prompt allows an agent to choose implementation, tests, docs, review, verification, or handoff.
The Hermes orchestration layer allows a coordinator to dispatch separate prompts to local Hermes profiles:

- `default`
- `vuitton`
- `bottega`

The proven pattern is:

1. Write focused task prompts under `.agents/hermes-orchestration/tasks/<run-id>/`.
2. Dispatch profiles through `dispatch.sh`.
3. Collect outputs through `collect.sh`.
4. Inspect outputs.
5. Decide the next narrow repo-backed slice.

This is good for research, review, and implementation design.
It is more dangerous for direct concurrent code edits unless each prompt includes exact claim rules and lane ownership.

### Phase 4: Handoff and Completion

The universal prompt requires a final handoff with:

- Summary
- What was changed or reviewed
- Files changed or reviewed
- Tests and checks run
- Remaining risks or blockers
- Source Review

This is directionally good.
The weakness is that it is still prose-first and not strict enough for production workflow auditing.
It does not require lane ID, phase ID, dependency status, gate skip reasons, merge order, or final integration owner.

## Requested Questions Answered

### Are agent roles clear?

Partially.

The Hermes task prompts define clear temporary roles:

- `default` as product and contract auditor;
- `vuitton` as Go/Postgres implementation reviewer;
- `bottega` as QA and regression reviewer.

Those roles are clear within specific task prompts.
They are not durable enough as a workflow system.
There are no canonical role files for `leader`, `quickstart`, `planner`, `architect`, `frontend-dev`, `backend-dev`, `qa-engineer`, `security-engineer`, or `tech-writer`.

The repo also has role-like patterns in review packets, such as Agent 7, Agent 8, Agent 9, and Agent 10, but these are historical execution lanes rather than stable reusable role definitions.

Clear roles:

- Product/contract auditor
- Go/Postgres implementation reviewer
- QA/regression reviewer
- Main Codex implementer
- Codex reviewer

Unclear or missing roles:

- Leader
- Quickstart guide
- Planner
- Architect
- Frontend developer
- Backend developer
- QA engineer as a durable workflow role
- Security engineer
- Technical writer

### Are handoff templates strict enough?

No.

The current handoff templates are useful for humans, but too loose for reliable multi-agent orchestration.
The existing required final handoff asks for summary, reviewed or changed files, checks, risks, and source review.
That is a good minimum.

It is missing production-grade fields:

- `agent_id`
- `role`
- `phase_id`
- `lane_id`
- `claimed_files`
- `reviewed_files`
- `changed_files`
- `result_doc`
- `dependencies`
- `upstream_blockers`
- `downstream_impacts`
- `gates_run`
- `gates_skipped`
- `skip_reason`
- `open_questions`
- `next_owner`
- `merge_order`
- `rollback_or_revert_notes`

Without these fields, a coordinator has to reconstruct too much from prose.
That is manageable for one expert operator.
It is brittle for a real user or a production multi-agent workflow.

### Are parallel tasks truly independent?

Sometimes.

Parallel work is safe when the lanes are read-only or when write surfaces are disjoint.
The existing Hermes prompts show a good read-only pattern: product/contract review, implementation design review, and QA/regression review can run independently because they do not edit files.
The repo history also shows safe test-only or docs-only lanes when each owns exact files.

Parallel work is risky when lanes touch related contracts through different files.
Exact file claims prevent two agents from editing the same file, but they do not prevent conceptual conflicts.
For example, one agent can update service validation while another updates MCP schemas, tests, or docs in a way that disagrees with it.
The claim board sees different files, but the product contract is shared.

Safe parallel steps:

- Read-only product review
- Read-only implementation design review
- Read-only QA review
- Isolated test additions in separate test files
- Docs-only packets with clear ownership
- Release-readiness report refresh after implementation claims are released

Risky parallel steps:

- Multiple agents editing `PLANS.md` or core plan docs
- Multiple agents touching migrations and store code separately
- One agent changing DTO/service validation while another changes MCP schema
- One agent changing worker behavior while another changes eval expectations
- Multiple agents updating review packets that claim current truth
- Any lane touching `internal/store/postgres/memories.go`, graph apply semantics, correction semantics, scope filtering, or protocol contracts without leader sequencing

### Where can agent outputs conflict?

Outputs can conflict in five main places.

First, docs can conflict with code.
VibeGravity relies heavily on planning docs and review packets.
If one agent updates code and another updates plans from older assumptions, the system can end with polished but stale guidance.

Second, schema and service validation can drift.
The repo has multiple external surfaces: HTTP, MCP, Hermes provider, CLI, and core service DTOs.
Separate agents can make each surface locally plausible while creating a cross-surface mismatch.

Third, tests can encode the wrong contract.
If QA writes tests from stale docs while backend changes behavior from current code, both lanes can appear complete until integration review.

Fourth, review packets can become competing truth.
The repo has many `docs/review-packets/` documents.
They are useful, but without a canonical index and current-state marker, agents can cite older packets as if they are still authoritative.

Fifth, file claims can be syntactically valid but semantically wrong.
The activity log already shows one claim that captured words like `--`, `verify`, and `MCP` as file names.
That suggests the claim tool needs path validation before it can be trusted as a production lock mechanism.

### Where can agents miss context?

Agents can miss context when they read only the live claim board and skip planning or review packets.
`WORK_PROGRESS.md` shows active claims and recent activity, but not the full project state.
It does not explain which review packets are stale, which gates are mandatory for the current phase, or which prior decisions are still binding.

Agents can also miss context when task prompts are too narrow.
For example, a backend prompt can inspect store files and miss product stop-lines in plan docs.
A QA prompt can inspect tests and miss an active architecture decision.
A docs prompt can inspect plans and miss current code behavior.

The universal prompt reduces this risk by requiring key plans and review packets.
However, it does not specify how to decide which review packets are current.
That leaves context selection to the agent's judgment.

### Is `phase_context.md` enough as shared memory?

No, because it does not exist in this checkout.

A future `phase_context.md` would be valuable, but only if it contains more than a status note.
It should be a canonical phase contract that includes:

- current phase name;
- current product promise;
- active source-of-truth docs;
- stale or superseded docs;
- active lanes;
- lane dependencies;
- owner per lane;
- current blockers;
- required verification gates;
- stop-lines;
- open decisions;
- next integration owner.

Even then, `phase_context.md` should not replace `WORK_PROGRESS.md`.
The better split is:

- `WORK_PROGRESS.md`: live lock board and activity trail.
- `phase_context.md`: shared product and execution memory for the current phase.
- `docs/review-packets/*`: durable lane results and deeper evidence.

### Does the system prevent runaway work?

Only partially.

The repo has several useful safeguards:

- claim exact files before editing;
- do not edit claimed files;
- heartbeat before widening scope;
- keep changes narrow;
- stop and propose an ADR for architecture changes;
- preserve explicit stop-lines such as no local extractor and no real Codex by default;
- Hermes `run-agent.sh` has a default `HERMES_MAX_TURNS` setting.

These are good.
They do not fully prevent runaway work.

Missing controls:

- maximum files per lane;
- maximum allowed scope expansion;
- required leader approval before widening from review to implementation;
- stale claim TTL;
- automatic warning when an agent touches files outside its claim;
- hard distinction between read-only, docs-only, tests-only, and code-edit lanes;
- required integration owner before multiple lanes merge;
- phase-level budget or timebox;
- automatic gate mapping based on changed file type.

### Does it prevent vague delegation?

It reduces vague delegation but does not eliminate it.

The universal prompt is strong because it tells agents not to ask what to do and gives them a startup sequence.
It also says to pick the smallest high-value unclaimed slice from `PLANS.md` and review packets when no specific task is given.

That is good for expert autonomous work.
It is still vague as a production delegation contract.
An agent can choose a slice that is useful but not the coordinator's intended next dependency.
It can also choose a low-collision side lane that generates output without advancing the critical path.

The system needs a stricter lane contract before work starts:

```yaml
lane_id:
role:
goal:
non_goals:
input_docs:
files_allowed:
files_forbidden:
expected_output:
required_gates:
handoff_path:
dependency_of:
```

### Does quickstart mode make sense as a product experience?

Yes as a concept, but not in the current repo.

A quickstart mode should be designed for a first-time user.
It should be read-only or low-risk, explain what it is about to do, run one confidence-building path, and produce a clear pass/fail result.

The current universal prompt is not a good public quickstart.
It is an autonomous engineering prompt.
It reads many internal docs, decides what work to do, may edit files, and assumes the user wants agentic repo work.

A good VibeGravity quickstart experience would:

1. inspect repo state;
2. detect whether dependencies are available;
3. run the deterministic Hermes Memory demo or local eval;
4. explain what passed;
5. recommend the next setup step;
6. avoid file edits unless explicitly approved.

Quickstart should create confidence before complexity.
The current workflow creates power before confidence.

### Does leader mode make sense as a production experience?

Yes, but it needs to be made real.

Leader mode is the right production concept for multi-agent work.
It should be the mode that turns a broad user goal into ordered lanes, assigns roles, enforces file ownership, reconciles outputs, and publishes the final truth.

The current repo has pieces of leader behavior:

- coordination board;
- universal prompt;
- file claims;
- Hermes dispatch;
- review packets;
- handoff templates.

But the leader role itself is not defined as a durable workflow.
There is no `leader.md` that owns decomposition, dependency graph, conflict resolution, scope escalation, merge order, and final verification.
Today, leader mode is mostly performed by the human operator or main Codex session.
That is acceptable internally.
It is not enough for a production product experience.

## Role Clarity Review

### Clear Roles

#### Product and Contract Auditor

This role is clear in Hermes `default` prompts.
It reviews product acceptance criteria, stop-lines, docs, and contract risks.

Strength:
It aligns well with VibeGravity's contract-first culture.

Weakness:
It is task-local, not codified as a reusable workflow role.

#### Go/Postgres Implementation Reviewer

This role is clear in Hermes `vuitton` prompts.
It inspects Go, PostgreSQL, transaction shape, interface shape, and implementation risks.

Strength:
It maps well to VibeGravity's highest-risk backend surfaces.

Weakness:
It can overlap with architect and backend-dev unless those roles are separated.

#### QA and Regression Reviewer

This role is clear in Hermes `bottega` prompts.
It focuses on tests, eval coverage, edge cases, and release gates.

Strength:
It is a natural parallel lane because it can often work read-only.

Weakness:
It needs authority to mark gates insufficient even when implementation tests pass.

### Overlapping Roles

#### Planner vs Architect

Planner should own sequencing and acceptance criteria.
Architect should own boundaries, ADRs, runtime model, and system invariants.
Without explicit docs, both can rewrite plans and create conflicting product truth.

#### Architect vs Backend Developer

Backend developer should implement within accepted contracts.
Architect should not casually rewrite code-level implementation details.
Backend developer should not make architecture changes without ADR.
The current docs say this in principle, but no role workflow enforces it.

#### QA Engineer vs Security Engineer

QA owns behavior and regression coverage.
Security owns trust boundaries, scope leakage, private/shared/group separation, and abuse paths.
In VibeGravity, these overlap heavily because scope safety is both product quality and security.
The roles need a shared checklist but separate decision authority.

#### Tech Writer vs Planner

Tech writer should update docs from accepted truth.
Planner should decide the next plan.
If both update `PLANS.md` or core plan docs in parallel, they can conflict.

#### Frontend Developer

This role is not currently meaningful in the live backend-first workflow unless it is scoped to a future operator UI.
It should not be included in default VibeGravity agent workflows until there is a real frontend surface to own.

## Parallelization Review

### Safe Parallelization Patterns

Safe patterns are read-only, sidecar, or file-disjoint:

- Product/contract review in one lane.
- Implementation design review in one lane.
- QA/regression review in one lane.
- Security/scope review in one lane.
- Docs-only update after implementation has landed.
- Test-only guardrails in isolated test files.
- Release-readiness report after other claims are released.

These work because they minimize direct file collisions and preserve main-thread integration.

### Risky Parallelization Patterns

Risky patterns are concept-shared even when files differ:

- Backend changes service validation while protocol lane updates MCP schemas.
- Store lane changes migrations while graph lane changes apply semantics.
- QA lane encodes tests from stale docs while docs lane updates product contract.
- Tech writer updates current-state docs while planner updates `PLANS.md`.
- Security lane changes scope assumptions while recall lane changes retrieval behavior.

The current system catches direct file collisions.
It does not catch cross-file contract collisions.

### Required Independence Test

Before launching a parallel lane, the leader should answer:

1. Does this lane have a single owner?
2. Does it have a unique lane ID?
3. Are files allowed and forbidden explicit?
4. Can it finish without waiting for another active lane?
5. If it depends on another lane, is it marked as blocked or review-only?
6. Does it write a result packet?
7. Can its output be accepted or rejected without rewriting another lane?

If the answer to any of these is no, the work is not truly parallel.

## Coordination Risks

### P0 Risks

#### P0. Requested workflow system is absent

The requested `.agent/workflows` role files and `phase_context.md` do not exist in the live checkout.
That means the repo cannot reliably execute the requested role workflow as specified.

Impact:
Agents will fall back to universal prompts, local judgment, and historical patterns.
This is too implicit for real users.

Recommendation:
Add the workflow docs or explicitly rename the intended path to `.agents/workflows`.

#### P0. No production leader authority

The repo has coordination mechanics, but no durable leader role.
There is no single documented owner for:

- lane decomposition;
- dependency ordering;
- conflict resolution;
- scope escalation;
- stale context cleanup;
- final verification;
- merge-readiness verdict.

Impact:
Parallel work can produce many plausible outputs without a reliable final decision.

Recommendation:
Create `leader.md` and make it the only role allowed to approve lane widening, resolve cross-lane conflicts, and publish final phase state.

### P1 Risks

#### P1. `WORK_PROGRESS.md` is not enough shared memory

`WORK_PROGRESS.md` is a live claim board.
It does not contain phase goals, blockers, canonical docs, stale docs, or acceptance criteria.

Impact:
Agents can avoid file collisions while still missing the real product context.

Recommendation:
Add `phase_context.md` and require agents to read it after `WORK_PROGRESS.md`.

#### P1. Handoffs are not machine-checkable

Current handoffs are prose-first.
They are understandable, but hard to audit automatically.

Impact:
A leader cannot reliably tell which gates were skipped, which dependencies remain, or which output is canonical.

Recommendation:
Require a YAML front matter block in every handoff and review packet.

#### P1. Exact file claims do not prevent conceptual collisions

Two agents can work on different files but change the same contract.

Impact:
The system can pass claim checks while producing incompatible code, docs, and tests.

Recommendation:
Leader must classify each lane by contract surface, not only file path.

#### P1. Autonomous task selection can wander off critical path

The universal prompt tells an agent to pick the smallest high-value unclaimed slice if no task is given.

Impact:
Agents may produce useful side outputs while the actual blocker remains unresolved.

Recommendation:
Require quickstart or leader mode to generate an explicit lane contract before autonomous work begins.

#### P1. Runaway control is incomplete

The system has heartbeats and narrow-scope language, but no hard scope limits.

Impact:
An agent can expand from review into docs, tests, and code if it judges the path useful.

Recommendation:
Add lane type constraints: `read_only`, `docs_only`, `tests_only`, `code_edit`, `integration_owner`.

### P2 Risks

#### P2. Hermes outputs are not automatically promoted into canonical packets

Hermes outputs land under `.agents/hermes-orchestration/runs/<run-id>/`.
That is useful but not the same as a curated review packet.

Impact:
Good review findings may remain hidden in run logs.

Recommendation:
Require a synthesis packet under `docs/review-packets/` for every multi-Hermes run.

#### P2. Claim tool accepts non-path tokens

The activity log shows a claim that treated words like `--`, `verify`, and `MCP` as files.

Impact:
The claim board can become noisy or misleading.

Recommendation:
Validate claim paths against repo-root path syntax and reject option-like or non-path tokens.

#### P2. Frontend role is premature

The repo is backend-first and Hermes/MCP-oriented.

Impact:
Adding a frontend-dev workflow now can confuse role routing.

Recommendation:
Define frontend-dev only when there is an operator UI or web console.

## Product Experience Review

### Quickstart Mode

Quickstart mode is a strong product idea, but the current universal prompt should not be the product quickstart.
The universal prompt is written for autonomous engineering agents, not first-time users.
It tells agents to begin work immediately, read many internal docs, decide what kind of work is useful, and potentially edit files.

That is too much agency for a quickstart user experience.

A good quickstart should answer:

- What is VibeGravity?
- What will this command do?
- Will it edit my repo?
- Will it write to a database?
- Will it call Codex or Hermes?
- What output proves it worked?
- What is the next safest step?

Recommended quickstart shape:

```bash
go run ./cmd/cli quickstart
```

Expected behavior:

1. Run read-only environment checks.
2. Run deterministic demo eval if available.
3. Show the Hermes Memory trust-loop outcome.
4. Print next setup steps.
5. Avoid modifying user config automatically.

Product verdict:
Quickstart should be a confidence path, not an autonomous work launcher.

### Leader Mode

Leader mode makes sense as the production experience for serious multi-agent work.
It should turn a user goal into a managed workflow.

Recommended leader responsibilities:

- read product and phase context;
- produce a lane plan;
- identify which lanes are parallel and which are sequential;
- assign roles;
- require exact file claims;
- track active claims;
- collect outputs;
- resolve conflicts;
- update phase context;
- run or verify final gates;
- issue a final CTO-style verdict.

Leader mode should not be just another agent persona.
It should be the stateful workflow controller.

Recommended leader user experience:

```bash
go run ./cmd/cli agent leader --goal "prepare V1 trust-loop release readiness"
```

The command should produce:

- `phase_context.md`
- lane assignments
- expected output files
- gate checklist
- final synthesis packet

Product verdict:
Leader mode is the right production model, but it currently exists only as an implicit operating pattern.

## Recommended Workflow Changes

### 1. Add Canonical Workflow Files

Create:

```text
.agents/workflows/
  leader.md
  quickstart.md
  planner.md
  architect.md
  backend-dev.md
  qa-engineer.md
  security-engineer.md
  tech-writer.md
  phase_context.md
```

If compatibility with the requested `.agent/workflows` path matters, add a short forwarding README or duplicate only the public entrypoint there.

### 2. Define `leader.md`

Minimum content:

- purpose;
- authority;
- startup checklist;
- lane decomposition rules;
- parallelization rules;
- conflict resolution rules;
- scope widening policy;
- final verification policy;
- handoff synthesis requirements.

Leader must be the only role allowed to:

- approve broadening scope;
- merge parallel lane conclusions;
- mark phase complete;
- declare real-user reliability.

### 3. Define `quickstart.md`

Minimum content:

- quickstart is read-only by default;
- it does not edit source files without user approval;
- it explains what it will run;
- it runs the fastest confidence gate;
- it reports next step and risk.

Quickstart should not use the universal autonomous prompt.

### 4. Add `phase_context.md`

Suggested shape:

```yaml
phase_id:
phase_name:
date:
product_promise:
current_goal:
canonical_docs:
stale_docs:
active_lanes:
blocked_lanes:
parallel_safe_lanes:
sequential_lanes:
stop_lines:
required_gates:
open_decisions:
integration_owner:
last_updated_by:
```

This should be updated by leader mode, not every agent.

### 5. Strengthen Handoff Templates

Every result packet should start with:

```yaml
---
agent_id:
role:
phase_id:
lane_id:
lane_type:
status:
claimed_files:
reviewed_files:
changed_files:
result_doc:
dependencies:
gates_run:
gates_skipped:
skip_reasons:
remaining_blockers:
next_owner:
source_review:
---
```

This keeps handoffs usable by both humans and future workflow tools.

### 6. Tighten File Claim Validation

Update `agent-work.sh` to reject:

- empty paths;
- option-like tokens such as `--`;
- plain words that do not contain `/` unless they are known root files;
- duplicate file entries;
- claims outside the repo root;
- broad globs.

Also add:

- `agent-work.sh stale`;
- `agent-work.sh validate`;
- stale-claim TTL warnings;
- optional `lane_id` support.

### 7. Require Lane Type Declarations

Allowed lane types:

- `read_only_review`
- `docs_only`
- `tests_only`
- `code_edit`
- `integration_synthesis`
- `release_readiness`

Each lane type should have allowed outputs and required gates.

### 8. Promote Hermes Run Outputs

For every multi-profile Hermes run, require a synthesis file:

```text
docs/review-packets/<run-id>-synthesis.md
```

That packet should state:

- prompts dispatched;
- profiles used;
- findings accepted;
- findings rejected;
- next implementation lane;
- conflicts between agents;
- source review.

### 9. Add a Workflow Index

Add:

```text
.agents/workflows/README.md
```

It should explain:

- when to use quickstart;
- when to use leader;
- when to use planner;
- when to use architect;
- when to use backend-dev;
- when to use QA/security;
- where outputs must be written.

### 10. Keep Universal Prompt, But Reposition It

The universal prompt is useful.
It should remain an expert internal tool.
It should not be the public quickstart.

Recommended label:

```text
Universal autonomous engineering prompt for trusted internal agents.
```

## Final CTO/CPO Decision

The current agent workflow system is not reliable enough for real users.
It is reliable enough for an expert operator coordinating internal parallel work in a known repo.
The difference is important.

The existing coordination layer solves one hard problem: preventing direct file edit collisions.
It does not yet solve the broader product problem: reliable role boundaries, phase memory, dependency ordering, output reconciliation, and completion authority.

Before VibeGravity presents this as a production agent workflow system, it needs:

1. canonical workflow role files;
2. real leader mode;
3. read-only quickstart mode;
4. `phase_context.md`;
5. stricter machine-checkable handoffs;
6. stronger claim validation;
7. explicit lane types;
8. required synthesis after parallel work.

My final decision is:

```text
NOT RELIABLE ENOUGH FOR REAL USERS YET.
RELIABLE ENOUGH FOR INTERNAL, EXPERT-SUPERVISED PARALLEL DEVELOPMENT.
```

## Evidence Appendix

### Coordination Layer Evidence

`.agents/coordination/README.md` defines the repo-local coordination surface and required loop:

- read `WORK_PROGRESS.md`;
- claim exact files;
- stop on active ownership conflict;
- heartbeat during long work;
- release immediately after finishing;
- run verification and leave a result note or review packet.

This is strong operational hygiene.

### Universal Prompt Evidence

`.agents/coordination/UNIVERSAL_AGENT_PROMPT.md` is self-starting.
It instructs agents to begin immediately, read repo instructions and plans, check coordination status, choose useful work, preserve VibeGravity invariants, verify changes, and report a final handoff.

This is strong for trusted autonomous agents.
It is too powerful and too open-ended for public quickstart mode.

### Hermes Orchestration Evidence

`.agents/hermes-orchestration/README.md` defines profile-isolated dispatch:

- `default`
- `vuitton`
- `bottega`

It also states that scripts do not mutate Hermes configuration and select profiles by setting `HERMES_HOME` for the child process.
That is a good production instinct.
The orchestration still depends on hand-written prompts and manual synthesis.

### Existing Playbook Evidence

`plans/12_agent-coding_playbook_codex-claude.md` already captures good general principles:

- understand;
- plan;
- implement;
- verify;
- review;
- handoff;
- use bounded subagents;
- do not hand off architecture ownership completely;
- do not let multiple agents edit the same hot files.

This document is directionally correct.
It should be upgraded into concrete workflow role files.

### Existing Handoff Template Evidence

`plans/13_handoff-prompts_and_response-templates.md` provides practical prompt and response templates.
They are useful but not strict enough for production orchestration because they are prose-first and lack machine-checkable metadata.

## Source Review

- Estimated source: live VibeGravity repo files, especially `.agents/coordination`, `.agents/hermes-orchestration`, `plans/12_agent-coding_playbook_codex-claude.md`, `plans/13_handoff-prompts_and_response-templates.md`, and current review packet conventions.
- Suspected license: project-internal original documentation.
- Similarity risk: low. This report is original review synthesis based on local repo evidence.
- Human review required: yes. The recommendation affects the repo's agent workflow architecture and should be reviewed before implementation.
