# Onboarding and Activation Review

Date: 2026-04-26
Reviewer: Agent 02
Role: CPO review
Scope: User journey, onboarding, activation, and first-run experience
Repo: `/Users/parker/Documents/VibeGravity`

## Executive Verdict

VibeGravity has a strong product promise, but the public onboarding path is not ready.
The repo currently explains the engine well to internal builders and reviewers, not to a new public user.
The clearest activation proof exists through `go run ./cmd/cli eval demo`, but that command is not presented as the first thing a user should run.
There is no tracked root `README.md`, no `setup.py`, and no `VibeGravityKit/cli.py` in the live checkout, so a user looking for conventional install instructions will immediately lose confidence.
The current Hermes bootstrap path is useful, but it prints MCP registration commands only; it does not install or configure Hermes automatically.
The first-run flow can create confidence after discovery, but the discovery path itself creates doubt.
My CPO decision is that onboarding is not clear enough for public users yet.

## Questions Answered

### Can a new user understand what to install?

Not reliably.

The Makefile implies Go module setup, binary builds, and `golangci-lint` installation:

- `make setup` runs `go mod download` and installs `golangci-lint`.
- `make build` builds `bin/server`, `bin/worker`, and `bin/cli`.

Runtime configuration implies PostgreSQL, migrations, and a local embedding endpoint:

- `VIBEGRAVITY_DB_URL`
- `VIBEGRAVITY_MIGRATION_PATH`
- `VIBEGRAVITY_EMBEDDING_ENDPOINT`
- `VIBEGRAVITY_EMBEDDING_MODEL`
- `VIBEGRAVITY_EMBEDDING_DIMS`

However, these are not assembled into a public install checklist. A new user must infer prerequisites by reading code, Makefile targets, tests, and planning docs.

### Can a new user run the first command without confusion?

No.

The best first confidence command is:

```bash
go run ./cmd/cli eval demo
```

That command produces a clear pass/fail result for the Hermes Memory trust loop. But the repo does not direct a new user to run it first. Running the CLI with no arguments shows command categories, not an onboarding sequence:

```text
Usage: cli <command>

Commands:
  doctor    Check system configuration and dependencies
  eval      Run deterministic quality evals
  hermes    Print Hermes bootstrap commands
  jobs      Inspect and recover worker jobs
  mcp       Serve the VibeGravity MCP protocol
```

This is useful for an operator who already understands the system, but not for a new user asking, "What should I do first?"

### Does the repo explain what will be created in their project?

No.

The current repo shows what the build creates:

- `bin/server`
- `bin/worker`
- `bin/cli`

The Hermes bootstrap command prints what should be registered with Hermes:

```bash
hermes mcp add vibegravity --command /path/to/cli --args mcp serve --stdio
hermes mcp test vibegravity
```

But the repo does not plainly explain what will be created or modified:

- whether binaries are created under `bin/`;
- whether a `.env` file is expected;
- whether database tables are created by migrations;
- whether Hermes config is modified automatically;
- whether MCP registration is persistent;
- whether VibeGravity writes files into a user's project.

This is a major trust gap because memory tooling touches sensitive project context.

### Does it explain IDE support clearly?

No.

The repo has agent and coordination material for Codex, Hermes, Claude, and reviewer agents, but not a public-facing IDE support section. `.agents/coordination/README.md` explains a concurrent agent file-claim workflow, not a user-facing IDE integration model.

The repo should explicitly answer:

- Does VibeGravity support VS Code?
- Does it support Cursor?
- Does it support Antigravity?
- Does it support Codex directly?
- Does it support Claude Code directly?
- Is Hermes currently the only intended first host?
- Is MCP the generic integration path?

Right now, a user can infer that Hermes and MCP matter, but the IDE support story is not clear.

### Does it explain the difference between quickstart mode and leader mode clearly?

No.

I found no clear public explanation of "quickstart mode" versus "leader mode" in the setup-oriented surfaces.
There is multi-agent coordination documentation and local orchestration material, but it is not framed as a public onboarding choice.

If these modes are intended product concepts, they need a plain-language table:

| Mode | Intended user | What it does | What it creates | When to use |
|---|---|---|---|---|
| Quickstart | First-time evaluator | Runs a local deterministic demo | No persistent memory system required | Prove the product in five minutes |
| Leader | Operator or team lead | Registers VibeGravity as a Hermes/MCP memory engine | Binaries, DB schema, MCP registration | Use VibeGravity in a real Hermes workflow |

If "leader mode" is only internal multi-agent orchestration, it should not appear in public onboarding until it has a product meaning.

### Does it show a real before-and-after example?

Not in the public onboarding path.

The repo has the ingredients:

- `cli eval demo` proves initial recall, explanation, correction, supersession, and private-scope separation.
- The MVP demo candidate describes the expected story: user gives a rule, Hermes syncs a turn, later recall returns context, user corrects memory, later recall suppresses the old memory.

But a public README should show an actual before/after in user language:

Before correction:

```text
Hermes recalls: "The project uses document memory as the V1 headline."
```

User correction:

```text
Actually, V1 is Hermes Memory: recall, explain, correction, and trust metadata.
```

After correction:

```text
Hermes recalls: "V1 is Hermes Memory, powered by VibeGravity. Documents are supporting context, not the headline."
```

This is the emotional activation moment. It should not be hidden inside eval output.

### Does the first-run flow create confidence or doubt?

Both, in the wrong order.

It creates doubt first because there is no root README, no install narrative, and no explicit first command. A user has to inspect Go files, Makefile targets, or internal docs to discover the usable path.

It creates confidence after the user finds `go run ./cmd/cli eval demo`. That command is genuinely strong because it proves the product's trust-loop story without needing real Hermes, Codex, PostgreSQL, or network dependencies.

The product should invert the order: confidence first, complexity second.

### Where could the user get stuck?

1. Repo discovery: no root `README.md`.
2. Installation: no single prerequisite list.
3. Package expectation: no `setup.py` or Python package path despite older packaging references and the prompt's expected files.
4. First command: CLI help lists commands but does not recommend a first run.
5. Doctor: reports DB and embedding failures but still exits with success and says the check completed.
6. Database setup: user must know how to create PostgreSQL DB, run migrations, and set `VIBEGRAVITY_DB_URL`.
7. Embedding setup: endpoint defaults to `http://localhost:8080`, but the repo does not explain what server should run there.
8. Hermes setup: bootstrap prints registration commands but does not clarify persistence, config location, or rollback.
9. MCP setup: command exists, but public explanation of MCP client expectations is thin.
10. Product modes: quickstart versus leader mode is not clearly defined.

## First-Run Journey Map

Current journey from repo discovery to first successful use:

1. User finds the GitHub repo.
2. User looks for a root README and does not find one.
3. User sees many internal planning, consulting, review-packet, and agent coordination docs.
4. User may open `plans/README.md`, which explicitly says the document set is for AI agents rather than humans.
5. User may inspect `Makefile` and infer Go build/test commands.
6. User may run `make setup` and `make build`.
7. User may run `go run ./cmd/cli` and see top-level command categories.
8. If user discovers `go run ./cmd/cli eval demo`, they see a strong local trust-loop demo pass.
9. If user tries real local setup, they need PostgreSQL, migrations, a configured database URL, and a local embedding endpoint.
10. If user tries Hermes integration, they run `go run ./cmd/cli hermes bootstrap --command "$(pwd)/bin/cli"` and receive MCP registration commands.
11. Full confidence still depends on real Hermes roundtrip and live PostgreSQL gates, both of which are not yet packaged as a public first-run flow.

## Activation Moment

The activation moment is:

```bash
go run ./cmd/cli eval demo
```

The moment where a user would say "this works" is when the CLI returns:

```text
Hermes Memory demo eval passed.
```

This command proves:

- initial recall includes project rule, active plan, and trust metadata;
- explain-memory provenance works;
- correction writes a supersession;
- next recall uses the corrected memory;
- private memory does not leak into another actor's recall.

That is the right product moment. It should become the public quickstart.

## Friction Points with File Evidence

### No root README

Evidence:

- `git ls-files` shows no tracked `README.md` at repo root.
- Existing tracked README files are nested: `.agents/coordination/README.md`, `.agents/hermes-orchestration/README.md`, `plans/README.md`, and `tests/README.md`.

Impact:

A new user has no canonical public entry point.

### The available README is agent-facing

Evidence:

- `plans/README.md` says the document set is not for explaining the project to humans; it is for AI agents building VibeGravity.

Impact:

The repo presents itself as an internal build pack rather than an installable product.

### Install and production ops are explicitly incomplete

Evidence:

- `consulting/07_current_state_and_roadmap.md` lists production ops, install, backup, and restore flows as incomplete.

Impact:

Public users will not know whether setup failure is their fault or the product's current limitation.

### Hermes bootstrap does not install or configure Hermes

Evidence:

- `plans/10_workpack_hermes-provider-and-external-surfaces.md` states that `internal/hermes.Provider` does not install or modify local Hermes configuration.
- `cmd/cli/hermes_bootstrap_stopline_test.go` forbids bootstrap output from claiming provider packaging, install readiness, or config writing.

Impact:

The word "bootstrap" may overpromise. It currently prints commands; it does not complete setup.

### Doctor can look successful even when dependencies fail

Evidence:

- `cmd/cli/main.go` prints database and embedding errors, then prints `Doctor check completed.`
- In local execution, `doctor` returned exit code 0 even when PostgreSQL and embedding endpoint checks failed.

Impact:

New users may misread a failed environment as a completed setup.

### The demo is strong but hidden

Evidence:

- `cmd/cli/main.go` exposes `eval demo`.
- `docs/review-packets/hermes-memory-demo-eval.md` explains that the demo walks the five-minute Hermes Memory trust loop without real Hermes, Codex, PostgreSQL, or network dependencies.

Impact:

The best activation path exists, but public onboarding does not lead with it.

## Missing Explanations

The repo needs plain-language explanations for:

- VibeGravity versus Hermes Memory.
- Engine versus app.
- Quickstart versus real setup.
- Quickstart mode versus leader mode, if leader mode is intended as a user-facing concept.
- What MCP is in this product.
- What Hermes bootstrap does and does not do.
- What `doctor` checks.
- What database objects migrations create.
- What files and config are created locally.
- Why local embedding exists and what endpoint is expected.
- Why real Codex is disabled by default.
- What "trust loop" means.
- What "scope" means: `agent_private`, `workspace_shared`, `group_shared`, and `session_scratch`.
- What correction and supersession mean in user terms.
- What degraded recall means.
- How to undo or remove a Hermes MCP registration.

## Recommended README Flow

The public README should follow this order:

1. **One-sentence promise**

   VibeGravity powers Hermes Memory: scoped, correctable, explainable memory for long-running agents.

2. **What this is and is not**

   It is an agent memory engine. It is not a chat UI, not a standalone agent, and not a generic vector database.

3. **Five-minute quickstart**

   ```bash
   go run ./cmd/cli eval demo
   ```

   Explain that this runs a local deterministic trust-loop demo with no DB, Hermes, Codex, or network dependency.

4. **What success looks like**

   Show the expected pass output and explain each line in human terms.

5. **Real setup prerequisites**

   List Go, PostgreSQL, `golang-migrate`, Hermes CLI, and optional local embedding endpoint.

6. **Build**

   ```bash
   make setup
   make build
   ```

7. **Database setup**

   ```bash
   createdb vibegravity
   export VIBEGRAVITY_DB_URL='postgres://localhost:5432/vibegravity?sslmode=disable'
   migrate -path migrations -database "$VIBEGRAVITY_DB_URL" up
   ```

8. **Doctor**

   ```bash
   bin/cli doctor
   ```

   Explain pass, warning, and failure semantics.

9. **Hermes MCP registration**

   ```bash
   bin/cli hermes bootstrap --command "$(pwd)/bin/cli"
   hermes mcp add vibegravity --command "$(pwd)/bin/cli" --args mcp serve --stdio
   hermes mcp test vibegravity
   ```

10. **Before-and-after example**

    Show a wrong memory, correction, and later corrected recall.

11. **Modes**

    Define quickstart mode and leader/operator mode, or remove those terms from public onboarding.

12. **IDE and client support**

    Clarify Hermes-first, MCP-compatible clients, and what is not supported yet.

13. **Known limitations**

    Real Codex disabled by default, custom Hermes memory provider packaging incomplete, real Hermes roundtrip tests still pending, production backup/restore not ready.

## CLI Onboarding Fixes

### Command names

Recommended additions:

```bash
bin/cli quickstart
bin/cli demo
bin/cli setup check
bin/cli setup print-env
bin/cli hermes print-registration
```

Keep existing commands for compatibility, but make first-use commands obvious.

### Help text

Current top-level usage should add:

```text
First time here?
  cli quickstart     Run a local Hermes Memory demo. No DB or Hermes required.
  cli setup check    Check local DB, migrations, embedding endpoint, and Hermes readiness.
```

`cli eval demo` help should explain:

```text
Runs the local trust-loop demo:
recall -> explain -> correction -> supersession -> private-scope check.
No database, Hermes, Codex, or network required.
```

### Error messages

Database failure should say:

```text
PostgreSQL is not reachable.
Next steps:
  1. Start PostgreSQL.
  2. createdb vibegravity
  3. export VIBEGRAVITY_DB_URL='postgres://localhost:5432/vibegravity?sslmode=disable'
  4. migrate -path migrations -database "$VIBEGRAVITY_DB_URL" up
```

Embedding failure should say:

```text
Embedding endpoint is not reachable.
For quickstart, run: cli quickstart
For real setup, start your local embedding server and set VIBEGRAVITY_EMBEDDING_ENDPOINT.
```

Hermes bootstrap should say:

```text
This command prints a Hermes MCP registration command.
It does not modify Hermes config automatically.
```

### Examples

Add public examples for:

```bash
go run ./cmd/cli eval demo
make build
bin/cli doctor
bin/cli hermes bootstrap --command "$(pwd)/bin/cli"
bin/cli mcp serve --stdio
bin/cli jobs metrics --window 15m
```

## Final CPO Decision

Onboarding is not clear enough for public users.

The product has a credible activation moment and a compelling trust-loop demo, but the public repo does not yet guide users to it. The current state is suitable for internal development, expert review, and guided demos. It is not yet suitable for unguided public adoption.

The next onboarding priority should be a root `README.md` that leads with `cli eval demo` as the no-dependency quickstart, then separately explains real PostgreSQL, embedding, and Hermes MCP setup. The product should sell confidence before complexity.
