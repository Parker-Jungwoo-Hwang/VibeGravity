# CLI and Packaging Review

Date: 2026-04-26  
Reviewer: Agent 05  
Role: CTO review  
Scope: CLI quality, Python packaging, installation, dependency management, and distribution  
Repo: `/Users/parker/Documents/VibeGravity`

## Original Request

Agent 05 was asked to review the GitHub repository `VibeGravity` with this scope:

- CLI quality
- Python packaging
- installation
- dependency management
- distribution

The specific files requested for review were:

- `setup.py`
- `MANIFEST.in`
- `VibeGravity/cli.py`
- `VERSION`
- README install instructions
- `CHANGELOG.md`
- any packaging files

The review was explicitly read-only. The repo was treated as the source of truth. Missing requested files were recorded as absent instead of inferred from historical project state.

## Review Questions

This report answers:

1. Is `setup.py` complete and clean?
2. Are dependencies pinned or bounded?
3. Are package data files included correctly?
4. Does the CLI handle errors safely?
5. Does the CLI print useful messages?
6. Are command names consistent?
7. Are update commands safe?
8. Can this be published to PyPI without confusion?
9. Is there a cleaner packaging path using `pyproject.toml`?

## Executive Verdict

VibeGravity is not currently a Python package.
The requested Python packaging files are absent from the live checkout.
The current install path is Go source build or Go command execution, not `pip install`.
The CLI is useful for local operators and internal validation, but it is not yet packaged as a public developer command.
The project can support GitHub-source local use today, but not PyPI publication.
The cleaner distribution path is to choose a Go-first release model first, then add a Python wrapper only if there is a real Python integration reason.
Public release should not proceed until binary naming, versioning, root documentation, migration packaging, and vulnerability gates are cleaned up.

## Sources Reviewed

Primary files and commands reviewed:

- `go.mod`
- `go.sum`
- `Makefile`
- `.gitignore`
- `PLANS.md`
- `CLAUDE.md`
- `cmd/cli/main.go`
- `cmd/cli/main_test.go`
- `cmd/cli/hermes_bootstrap_stopline_test.go`
- `internal/config/config.go`
- `internal/db/pool.go`
- `internal/store/postgres/jobs.go`
- `internal/store/postgres/jobs_test.go`
- `internal/mcp/protocol.go`
- `internal/mcp/surface.go`
- `internal/hermes/provider.go`
- `plans/10_workpack_hermes-provider-and-external-surfaces.md`
- `tests/README.md`
- `docs/adr-001-migration-versioning.md`
- `docs/adr-002-embedding-dimension-policy.md`
- `docs/adr-006-db-driver.md`
- `docs/adr-008-package-layout.md`
- `docs/review-packets/v1-trust-loop-readiness-report.md`

Packaging file inventory commands showed these requested public packaging files are absent:

- `setup.py`
- `MANIFEST.in`
- `VibeGravity/cli.py`
- `VERSION`
- root `README.md`
- root `CHANGELOG.md`
- `pyproject.toml`
- `setup.cfg`
- `requirements.txt`
- `Pipfile`
- `poetry.lock`
- `uv.lock`

## Install Path

The current install path is a Go developer path.

The Makefile builds local binaries:

```bash
go build -o bin/server ./cmd/server
go build -o bin/worker ./cmd/worker
go build -o bin/cli ./cmd/cli
```

The repo also supports direct execution:

```bash
go run ./cmd/cli
go run ./cmd/cli eval demo
go run ./cmd/cli mcp serve --stdio
go run ./cmd/cli hermes bootstrap --command "$(pwd)/bin/cli"
```

The Hermes-facing bootstrap path currently prints MCP registration commands:

```bash
hermes mcp add vibegravity --command /path/to/cli --args mcp serve --stdio
hermes mcp test vibegravity
```

That is useful, but it is not a complete installation process. It assumes:

- a built CLI binary exists;
- PostgreSQL is installed and migrated;
- `VIBEGRAVITY_DB_URL` points at the correct database;
- a local embedding endpoint exists or the operator understands the degraded path;
- Hermes has MCP support available;
- the operator understands whether MCP registration is persistent and how to roll it back.

The current install path is therefore suitable for repo contributors and local operators, not public developers.

## Packaging Issues

### 1. The requested Python package does not exist

There is no Python package surface in the live checkout.

Observed absent:

```text
setup.py
MANIFEST.in
VibeGravity/cli.py
VERSION
README.md
CHANGELOG.md
pyproject.toml
```

This means `setup.py` is not incomplete in the ordinary sense; it is absent. The repo cannot honestly be published as a Python package without first creating a real Python package or wrapper.

### 2. The public package name and command name are unclear

The Go module is:

```text
github.com/parker-jungwoo-hwang/vibegravity
```

But the CLI package path is `./cmd/cli`, which installs as a generic binary named `cli` by default:

```text
github.com/parker-jungwoo-hwang/vibegravity/cmd/cli -> /Users/parker/go/bin/cli
```

That is fine for internal development, but bad for public distribution. A public developer should install and run a product-named command such as:

```bash
vibegravity doctor
vibegravity eval demo
vibegravity mcp serve --stdio
vibegravity hermes bootstrap
```

### 3. No root README install instructions

There is no tracked root `README.md`.

The repo has strong internal documents, but public developers need a root-level quickstart:

1. what VibeGravity is;
2. what is supported today;
3. prerequisites;
4. local demo command;
5. PostgreSQL setup;
6. migration command;
7. build/install command;
8. Hermes MCP registration;
9. troubleshooting and rollback.

Without a root README, a public developer will have to infer install flow from `Makefile`, plans, tests, and CLI help.

### 4. No changelog or release history

There is no root `CHANGELOG.md`.

That matters because the repo is moving quickly and carries a major product transition from an earlier Python-era concept to the current Go-first Hermes Memory direction. A changelog should explain:

- current public status;
- breaking changes;
- supported install path;
- migration requirements;
- known limitations;
- release gates.

### 5. Package data is not handled for public installation

The runtime config defaults `MigrationPath` to `migrations`.

That works when running from the repo root. It is fragile when distributing a standalone binary, Homebrew formula, Docker image, or Python wrapper because the binary may not be executed from a directory containing `migrations/`.

The release process needs one of these decisions:

- embed migrations into the binary;
- install migrations into a documented share directory;
- require `VIBEGRAVITY_MIGRATION_PATH`;
- ship a Docker image with migrations in a fixed internal path.

The same concern applies to docs and golden eval fixtures if they become public commands.

### 6. Go toolchain requirement is too high and ambiguous for public release

`go.mod` declares:

```text
go 1.25.7
```

The repo instruction says Go 1.22+, but the module currently requires Go 1.25.7. That can surprise public developers and CI. If Go 1.25.7 is intentional, the README and release tooling must say so. If it is accidental, lower the module target to the minimum supported Go version and verify tests.

### 7. No version source of truth

There is no root `VERSION`, and the CLI has no visible `version` command.

The MCP protocol server advertises a hard-coded server version:

```text
0.1.0
```

That is not enough for release management. Public builds need a single version source, ideally injected into binaries at build time and shown by:

```bash
vibegravity version
```

## CLI Issues

### 1. CLI categories are useful but internal

Running the CLI with no arguments prints:

```text
Usage: cli <command>

Commands:
  doctor    Check system configuration and dependencies
  eval      Run deterministic quality evals
  hermes    Print Hermes bootstrap commands
  jobs      Inspect and recover worker jobs
  mcp       Serve the VibeGravity MCP protocol
```

This is clear for local operators. It is not enough for first-time public users because:

- the binary name is `cli`;
- there is no recommended first command;
- there is no quickstart mode;
- there is no version command;
- there is no explicit explanation of prerequisites.

### 2. `doctor` prints errors but returns success

`runCLI` always returns `0` after `runDoctor()` finishes.

`runDoctor()` prints DB and embedding endpoint errors, but it does not return a status value that can fail the process.

This is dangerous for automation. A CI script or install script could treat a failed database or embedding check as passing.

Recommended behavior:

```text
0 = all required checks pass
1 = required checks fail
2 = optional/degraded checks fail only if strict mode is enabled
```

Also add:

```bash
vibegravity doctor --strict
vibegravity doctor --json
```

### 3. `doctor` says config is OK even with placeholder embedding config

`internal/config/config.go` defaults to:

```text
VIBEGRAVITY_EMBEDDING_ENDPOINT=http://localhost:8080
VIBEGRAVITY_EMBEDDING_MODEL=pending
VIBEGRAVITY_EMBEDDING_DIMS=0
```

ADR-002 says the doctor command should warn while model and dimensions are pending. The current output style can give false confidence because it prints config values and then prints `Config OK`.

Recommended behavior:

- fail or warn on `EmbeddingModel == "pending"`;
- fail or warn on `EmbeddingDims == 0`;
- distinguish required checks from degraded/optional checks;
- print remediation commands.

### 4. Error handling is generally safe in narrower command paths

The CLI does several things correctly:

- unknown top-level commands return nonzero;
- unknown subcommands return nonzero;
- invalid Hermes MCP names are rejected when they contain whitespace;
- command paths use explicit error messages;
- DB password masking is tested;
- `jobs metrics` is read-only;
- `jobs requeue-blocked` only updates jobs currently in `blocked` status;
- requeue does not increment attempts or schedule retry delay.

This is a good foundation for internal operator use.

### 5. Update and recovery commands need more public safety affordances

The CLI has a manual recovery command:

```bash
cli jobs requeue-blocked <job_id>
```

The underlying store requires `WHERE id = $1 AND status = 'blocked'`, which is good. However, for public operators this should also include:

- a `--dry-run` mode;
- a `--reason` or audit note;
- clear warning that requeueing blocked jobs can repeat failed work;
- tenant/workspace display before action;
- optional confirmation when run interactively.

The broader update tools are exposed through MCP/Hermes/core service paths:

- `update_plan`
- `correct_memory`
- correction-driven `update_memory` through the apply path

Those paths appear much better protected than a raw CLI update command, but public docs should explain idempotency keys, correction evidence, and replay behavior.

### 6. Command names are not fully consistent across surfaces

The MCP surface exposes:

```text
prefetch
recall_preview
sync_turn
search_memory
search_documents
add_note
create_plan
update_plan
correct_memory
view_timeline
explain_memory
degraded_status
```

The Hermes provider exposes:

```text
recall_preview
search_memory
add_note
show_plan
explain_memory
correct_memory
view_timeline
degraded_status
```

The CLI exposes:

```text
doctor
eval golden
eval demo
jobs metrics
jobs blocked
jobs requeue-blocked
mcp serve --stdio
hermes bootstrap
```

The naming is understandable, but not fully productized:

- `search_memory` is singular while `search_documents` is plural;
- Hermes has `show_plan`, while MCP has `create_plan` and `update_plan`;
- public CLI binary name is `cli`, not `vibegravity`;
- `prefetch` is an internal contract name while `recall_preview` is the operator-facing name.

Recommendation: keep internal contract names in code, but make public docs lead with operator names:

- `recall preview`
- `sync turn`
- `correct memory`
- `explain memory`
- `timeline`
- `doctor`

## Dependency Risks

### 1. Go dependencies are pinned by module versions

Direct Go dependencies in `go.mod` are exact versions:

```text
github.com/go-chi/chi/v5 v5.2.5
github.com/jackc/pgx/v5 v5.9.2
github.com/joho/godotenv v1.5.1
github.com/pgvector/pgvector-go v0.3.0
```

The module also records hashes in `go.sum`.

Validation performed:

```bash
go mod verify
```

Result:

```text
all modules verified
```

### 2. No Python dependency risk because there is no Python package

There are no Python dependency pins or bounds because the current repo has no Python packaging files.

If a Python package is added, it must use `pyproject.toml`, not a legacy full `setup.py`, and dependencies should use conservative lower and upper bounds:

```toml
dependencies = [
  "httpx>=0.27,<1",
  "pydantic>=2,<3"
]
```

Do not publish an empty shim package unless it has a clear purpose such as installing a wrapper around a released binary.

### 3. Standard-library vulnerability scan blocks release

Validation performed:

```bash
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

Result: affected vulnerabilities were found in the Go 1.25.7 standard library, with fixes in Go 1.25.8 or Go 1.25.9.

Examples included:

- `crypto/x509` issues fixed in Go 1.25.9;
- `crypto/tls` issue fixed in Go 1.25.9;
- `os` and `net/url` issues fixed in Go 1.25.8.

The project should upgrade the Go toolchain and rerun `govulncheck` before any public release.

### 4. Supply-chain posture is acceptable but not release-complete

Strengths:

- small direct dependency set;
- exact module versions;
- `go.sum` hashes;
- `go mod verify` passes;
- DB driver choice is documented in ADR-006;
- pgvector dependency is explicit and product-aligned.

Missing before release:

- automated `govulncheck`;
- SBOM generation;
- checksum files for release artifacts;
- signed tags or provenance attestations;
- documented dependency update policy;
- release CI against a pinned Go toolchain.

## Distribution Readiness

Current readiness classification:

```text
Local use only / GitHub-source contributor use.
```

Not ready for:

- PyPI;
- Homebrew;
- public binary release;
- Docker release;
- production local operator install;
- broad public developer onboarding.

The repo is closest to being ready for:

```text
GitHub install from source with documented Go build steps.
```

It can plausibly move to public binary distribution after:

- root README exists;
- binary is named `vibegravity`;
- version command exists;
- Go toolchain vulnerability scan passes;
- migrations are embedded or packaged;
- `doctor` has meaningful exit codes;
- live PostgreSQL gate is documented and green in release CI;
- Hermes MCP bootstrap is documented with rollback.

## Answers to Requested Questions

### Is `setup.py` complete and clean?

No. `setup.py` is absent.

This is acceptable only if VibeGravity is intentionally not a Python package. It is not acceptable if the intended public install path is `pip install vibe-gravity`.

### Are dependencies pinned or bounded?

Go dependencies are pinned by exact module versions in `go.mod` and hashed in `go.sum`.

Python dependencies are not pinned or bounded because there is no Python package.

The Go toolchain itself is currently a release risk because `go.mod` requires Go 1.25.7 and `govulncheck` found standard-library vulnerabilities fixed in later Go patch versions.

### Are package data files included correctly?

No public package data model exists.

For the current Go binary path, migrations are plain repo files and are not embedded. A binary installed outside the repo will need a configured migration path or embedded migrations.

For any future Python package, `MANIFEST.in` or equivalent `pyproject.toml` package-data configuration must include:

- migrations;
- README;
- changelog;
- license;
- any CLI templates or generated default config;
- test/demo fixtures only if public commands depend on them.

### Does the CLI handle errors safely?

Partly.

Safe behavior observed:

- unknown commands fail;
- invalid arguments fail;
- Hermes bootstrap validates non-empty no-whitespace server names;
- password masking is tested;
- blocked job requeue only touches blocked jobs;
- MCP stdio returns structured JSON-RPC errors.

Unsafe or weak behavior:

- `doctor` prints failures but exits success;
- `doctor` does not distinguish warning versus failure;
- no JSON/structured diagnostic output;
- no `--dry-run` for recovery actions;
- no release-grade version output.

### Does the CLI print useful messages?

Yes for internal operators.

Not enough for public developers.

The messages are concise and usually actionable once the user already understands the architecture. They do not yet teach installation, first-run flow, database setup, embedding setup, or Hermes integration.

### Are command names consistent?

Mostly consistent internally, but not polished publicly.

The biggest issue is the generic binary name `cli`. The second issue is naming drift between Hermes and MCP surfaces, especially `show_plan` versus `create_plan` / `update_plan`.

Recommendation: public command should be `vibegravity`; public docs should lead with `recall_preview`, `correct_memory`, `explain_memory`, `view_timeline`, and `doctor`.

### Are update commands safe?

The direct CLI recovery command is reasonably guarded because it only requeues currently blocked jobs.

The correction/update memory flow is designed around idempotency, provenance, and transaction safety, but full public release confidence still depends on live PostgreSQL and Hermes runtime evidence. Existing readiness docs already treat missing live DB proof as a release blocker.

Public update/recovery commands still need operator affordances:

- dry run;
- audit reason;
- confirmation for destructive or repeat-triggering actions;
- clear before/after output;
- tenant/workspace scoping in output.

### Can this be published to PyPI without confusion?

No.

Publishing to PyPI now would be actively confusing because:

- there is no Python package;
- there is no `pyproject.toml`;
- there is no Python CLI;
- the product is Go-first;
- the repo does not define what a Python package would install;
- users would expect `pip install` to give them a working command, but the current runtime depends on Go-built binaries, PostgreSQL, migrations, and an embedding endpoint.

### Is there a cleaner packaging path using `pyproject.toml`?

Yes, if the project truly needs a Python wrapper.

But the cleaner near-term path is Go-first distribution, not Python packaging.

Recommended ordering:

1. Make `vibegravity` a real Go binary command.
2. Add root README and changelog.
3. Add version injection.
4. Embed or package migrations.
5. Add release CI and vulnerability gates.
6. Add Homebrew/tarball/Docker distribution.
7. Only then decide whether a Python package should install a wrapper or client SDK.

If Python packaging is still desired, use `pyproject.toml` and keep `setup.py` as either absent or a tiny compatibility shim.

## Recommended Fixes

### `setup.py`

Do not add a full legacy `setup.py`.

Either:

- omit it because VibeGravity is Go-first; or
- keep only a tiny compatibility shim if a real `pyproject.toml` package is added.

Do not publish a placeholder package.

### `pyproject.toml`

If Python support is needed, add a modern `pyproject.toml`:

```toml
[build-system]
requires = ["hatchling>=1.24,<2"]
build-backend = "hatchling.build"

[project]
name = "vibe-gravity"
version = "0.1.0"
description = "Hermes Memory, powered by VibeGravity."
readme = "README.md"
requires-python = ">=3.11"
license = { text = "TBD" }

[project.scripts]
vibegravity = "vibegravity.cli:main"
```

Only do this if there is a real Python module under `vibegravity/`.

### `MANIFEST.in`

Only add `MANIFEST.in` if Python packaging is chosen.

Include:

```text
include README.md
include CHANGELOG.md
include LICENSE
recursive-include migrations *.sql
recursive-include docs *.md
```

Do not include internal agent scratch state, `.omx`, `bin/`, `.DS_Store`, or generated coordination state.

### CLI help text

Rename the public binary and update help:

```text
Usage: vibegravity <command>

Start here:
  vibegravity eval demo     Run the local Hermes Memory trust-loop demo
  vibegravity doctor        Check local runtime configuration

Commands:
  doctor
  eval
  hermes
  jobs
  mcp
  version
```

Add command examples to the root README.

### CLI behavior

Change `doctor` to return an exit code.

Add:

```bash
vibegravity doctor --strict
vibegravity doctor --json
vibegravity version
vibegravity jobs requeue-blocked <job_id> --dry-run
```

### Release process

Add a release checklist:

1. Update `CHANGELOG.md`.
2. Run `go test ./...`.
3. Run `make eval`.
4. Run `make lint`.
5. Run `make check-headers`.
6. Run `git diff --check`.
7. Run `go mod verify`.
8. Run `govulncheck ./...`.
9. Run live PostgreSQL gate against migrated scratch DB.
10. Run Hermes MCP roundtrip smoke when available.
11. Build binaries with version metadata.
12. Generate checksums and SBOM.
13. Tag release.
14. Publish binaries or Docker image.

### Go-first packaging

Preferred public path:

- `cmd/vibegravity` as the public CLI entrypoint;
- optional aliases or subcommands for `server` and `worker`;
- GoReleaser or equivalent release automation;
- Homebrew formula;
- Docker image for server/worker plus migrations;
- explicit migration packaging story.

### PyPI path

Only publish to PyPI if one of these is true:

- there is a Python client SDK;
- there is a Python wrapper that downloads or invokes a released binary;
- Hermes or another first-class host needs Python package installation.

If PyPI is used, the package description must say clearly:

```text
This Python package is a wrapper/client for the Go VibeGravity engine.
The core VibeGravity server and worker are Go binaries.
```

## Final CTO Decision

Do not proceed with a public PyPI release.

Do not market this as publicly installable yet.

Proceed with local/GitHub-source usage while the team closes the packaging gap.
The repo has a real product and a real operator CLI foundation, but the distribution story is not ready for public developers.
The next best move is not to create a fake Python package; it is to make the Go-first install path obvious, named, versioned, testable, and secure.

## Verification Performed

Read-only verification commands run:

```bash
git status --short
find . -maxdepth 3 \( -name 'setup.py' -o -name 'pyproject.toml' -o -name 'setup.cfg' -o -name 'MANIFEST.in' -o -name 'VERSION' -o -name 'README*' -o -name 'CHANGELOG*' -o -name 'requirements*.txt' -o -name 'uv.lock' -o -name 'poetry.lock' \) -print | sort
git ls-files setup.py MANIFEST.in VibeGravity/cli.py VERSION README.md CHANGELOG.md pyproject.toml setup.cfg requirements.txt go.mod go.sum cmd/cli/main.go
go version
go list -m all
go mod verify
go test ./cmd/cli
go mod tidy -diff
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
go build -o /tmp/vibegravity-review-build/cli ./cmd/cli
go build -o /tmp/vibegravity-review-build-server/server ./cmd/server
go build -o /tmp/vibegravity-review-build-server/worker ./cmd/worker
```

Notable results:

- `go mod verify`: passed.
- `go test ./cmd/cli`: passed.
- `go mod tidy -diff`: no diff output.
- `go build` for CLI/server/worker: passed.
- `govulncheck`: failed because Go 1.25.7 standard-library vulnerabilities affect reachable symbols.

## Source Review

- Estimated source: current VibeGravity repo files and local command outputs.
- Suspected license: project-internal original material.
- Similarity risk: low.
- Human review required: yes, before any public release or package publication.
- Notes: This report is original review prose based on the live checkout. It does not copy external packaging templates beyond small illustrative snippets.
