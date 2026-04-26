# Security and Privacy Review

Date: 2026-04-26
Reviewer: Agent 08
Role: CTO and security reviewer
Scope: Security, privacy, supply-chain risk, secret handling, local file access, install risk, and user trust
Repo: `/Users/parker/Documents/VibeGravity`

## Original Request

Act as a CTO and security reviewer for the GitHub repo "VibeGravity".

Review this repo as software that users may install on their own projects.

Requested source areas:

- `setup.py`
- CLI code
- scripts that call subprocess
- scripts that read or write files
- deployment tools
- update command
- team-manager
- context-manager
- context-router
- env-manager
- security-scanner

Requested questions:

1. Does any command delete, overwrite, or copy files?
2. Does any script run shell commands?
3. Does any script download binaries or external resources?
4. Does any script scan user code or store user data?
5. Does team memory create privacy risk?
6. Does deployment via tunnel create risk?
7. Are update commands safe?
8. Are dependencies safe and bounded?
9. Are there missing warnings or consent steps?
10. Are there hardcoded secrets, risky paths, or unsafe defaults?

Clarification from live repo state:

- No `setup.py` exists in the live checkout.
- No repo-local `team-manager`, `context-manager`, `context-router`, `env-manager`, or `security-scanner` implementation exists in the live checkout.
- No dedicated self-update command or deployment/tunnel script was found.
- The closest team-management implementation is `.agents/coordination/agent-work.sh`.
- The closest deployment-like/external orchestration surface is `.agents/hermes-orchestration`.
- The repo is Go-first; legacy Python/package expectations should be treated as stale unless reintroduced intentionally.

## Executive Verdict

VibeGravity is not safe for public install yet.
It is acceptable only for local-only developer use with trusted operators.
The most serious issue is that the HTTP and MCP surfaces expose memory read/write operations without authentication or a durable authorization boundary.
Storage queries contain useful scope filters, but many external surfaces still trust caller-supplied tenant, workspace, actor, entity, and visible group identifiers.
No real hardcoded production secrets were found, but the repo contains tracked prebuilt binaries, unpinned setup tooling, hardcoded local Hermes paths, and external Hermes/Codex orchestration scripts.
The code is not malicious; the problem is that its trust boundary is still an internal developer boundary, not a public-install boundary.
My CTO decision is local-only use until auth, identity binding, install warnings, and supply-chain cleanup are complete.

## Threat Model

A user installs VibeGravity inside a private project and starts the server, CLI, MCP bridge, worker, or Hermes orchestration scripts.
The user's project context, raw agent events, document chunks, plans, notes, corrections, traces, and memory graph data may be sensitive.
If the HTTP server is reachable from LAN, a local tunnel, a browser-adjacent process, or another local user, an attacker can call memory read/write APIs.
If an MCP client or Hermes profile is misconfigured, a tool caller can invoke recall, sync, search, correction, timeline, and explain operations.
Because identity is mostly passed in request DTOs, a caller that can reach the surface may impersonate another actor or entity by supplying IDs.
If Hermes orchestration scripts are run on prompt files containing private repo details, those prompts and outputs can be sent to external model tooling and stored in local run logs.
If public users follow install commands blindly, Go modules and an unpinned tool install can fetch external code into their environment.

## Method

Reviewed live repo files only.
The review treated missing requested surfaces as absent, not as implied future behavior.

Primary evidence reviewed:

- `AGENTS.md`
- `plans/00_read-this-first_for-building-agents.md`
- `plans/01_rfp_vibegravity_hermes-first.md`
- `plans/02_product-contract_and_direction.md`
- `plans/03_target-architecture_codex-first.md`
- `plans/05_runtime-contracts_ingest-recall-apply.md`
- `plans/06_data-model_and_storage-invariants.md`
- `Makefile`
- `go.mod`
- `go.sum`
- `cmd/server/main.go`
- `cmd/cli/main.go`
- `internal/httpapi/router.go`
- `internal/mcp/surface.go`
- `internal/mcp/protocol.go`
- `internal/hermes/provider.go`
- `internal/config/config.go`
- `internal/kernel/service.go`
- `internal/ingest/service.go`
- `internal/recall/assembler.go`
- `internal/worker/stage2_sources.go`
- `internal/store/postgres/*.go`
- `migrations/*.sql`
- `.agents/coordination/*`
- `.agents/hermes-orchestration/*`

Validation commands used:

```bash
find . -name setup.py -o -name '*team-manager*' -o -name '*context-manager*' -o -name '*context-router*' -o -name '*env-manager*' -o -name '*security-scanner*'
rg -n "exec\\.|Command\\(|os\\.Remove|RemoveAll|Rename\\(|WriteFile|Mkdir|Create\\(|OpenFile|ReadFile|Copy|cp |mv |rm -rf|curl|wget|http\\.Get|client\\.Get"
rg -n "AKIA|BEGIN (RSA|OPENSSH|EC|PRIVATE) KEY|sk-[A-Za-z0-9]|OPENAI_API_KEY|ANTHROPIC_API_KEY|password\\s*=|token\\s*=|secret"
git ls-files -s bin .env .agents/hermes-orchestration/runs .agents/coordination
file bin/server bin/worker
shasum -a 256 bin/server bin/worker
go mod verify
go env GOPROXY GOSUMDB GOPRIVATE
go tool govulncheck -h
```

Results:

- `go mod verify` passed.
- `govulncheck` was not available: `go: no such tool "govulncheck"`.
- No tracked `.env` file was found.
- No production API keys, private keys, or real secrets were found by the secret-pattern search.
- Test-only strings such as `super-secret` are present in password masking tests and are not live secrets.

## Direct Answers to Requested Questions

### 1. Does any command delete, overwrite, or copy files?

Yes.

File-system behaviors:

- `Makefile:7-9`: `make build` overwrites `bin/server`, `bin/worker`, and `bin/cli`.
- `Makefile:41`: `make clean` runs `rm -rf bin/`.
- `.agents/coordination/agent-work.sh:55`: lock cleanup uses `rm -rf "$LOCK_DIR"` through a shell trap.
- `.agents/coordination/agent-work.sh:58-60`: creates or truncates missing coordination state files.
- `.agents/coordination/agent-work.sh:84-144`: rewrites `WORK_PROGRESS.md` through a temp file and move.
- `.agents/coordination/agent-work.sh:175-204` and `234-256`: copies and rewrites claim state with temp files.
- `.agents/hermes-orchestration/run-agent.sh:56-76`: creates run directories and writes metadata files.
- `.agents/hermes-orchestration/run-agent.sh:79-87`: writes Hermes stdout and stderr logs.

Database overwrite/delete behaviors:

- `internal/store/postgres/documents.go:100-115`: document upsert overwrites source, title, metadata, version hint, and updated timestamp on fingerprint conflict.
- `internal/store/postgres/documents.go:126-128`: document chunk replacement deletes all existing chunks for a document before inserting replacements.
- `internal/store/postgres/notes_plans.go:164-166`: updating a plan with items deletes all existing plan items before reinserting.
- `internal/store/postgres/jobs.go:456-480`: manual requeue mutates blocked jobs back to queued.
- `internal/store/postgres/memories.go:296-307`: correction/update flow supersedes the prior target memory.

CTO assessment:

The file deletions are mostly bounded to build artifacts and coordination state, but the repo needs public warnings around `make clean`, build overwrites, generated run logs, and destructive DB replacement semantics.

### 2. Does any script run shell commands?

Yes.

Shell command surfaces:

- `Makefile` runs Go build, test, eval, server, worker, setup, lint, header checks, and `rm -rf bin/`.
- `.agents/coordination/agent-work.sh` runs shell builtins plus tools such as `mkdir`, `awk`, `date`, `mv`, `cp`, `cat`, and `rm -rf`.
- `.agents/hermes-orchestration/run-agent.sh` runs `git`, `mkdir`, `date`, `cat`, `env`, and `hermes chat`.
- `.agents/hermes-orchestration/dispatch.sh` runs `run-agent.sh` in background processes.
- `.agents/hermes-orchestration/collect.sh` reads and prints run metadata, stdout, and stderr files.
- `.agents/hermes-orchestration/status.sh` runs `hermes profile list` and `hermes profile show`.

Go code:

- I did not find `exec.Command` usage in the Go code.
- Search hits for `exec.Exec` are database executor interfaces, not OS subprocess execution.

CTO assessment:

The shell scripts are intentional internal operator tooling, but they should not be advertised as public-safe install automation without consent prompts and clearer data handling docs.

### 3. Does any script download binaries or external resources?

Yes, through the Go toolchain and external model tooling.

External download/call paths:

- `Makefile:53`: `go mod download` downloads Go module dependencies.
- `Makefile:54`: `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest` downloads and installs an unpinned external tool.
- `.agents/hermes-orchestration/run-agent.sh:79-87`: `hermes chat` can call external provider-backed model tooling; defaults are `HERMES_PROVIDER=openai-codex` and `HERMES_MODEL=gpt-5.5`.
- `cmd/cli/main.go:628-634`: `doctor` performs an HTTP GET to `VIBEGRAVITY_EMBEDDING_ENDPOINT`.

What was not found:

- No `curl` or `wget` download script was found.
- No dedicated binary installer script was found.
- No current self-update command was found.

CTO assessment:

Dependency fetching is normal for Go development, but `@latest` is not acceptable for a public install path.
External model calls need explicit user consent and a clear "what may leave your machine" warning.

### 4. Does any script scan user code or store user data?

Yes.

Scanning/read paths:

- `tools/headercheck/main.go` reads Go files to enforce source headers.
- `internal/eval/golden.go` reads a user-provided eval JSON file path.
- `tests/migration_contract_test.go` and several tests read migration/source files.
- `cmd/cli eval golden --path <path>` can read any path the operator provides.

Data storage paths:

- `internal/ingest/service.go:65-132` stores raw turn events and enqueues jobs.
- `migrations/000002_create_core_tables.up.sql:1-13` stores raw event `payload_json`.
- `migrations/000002_create_core_tables.up.sql:69-116` stores derived memories.
- `migrations/000002_create_core_tables.up.sql:21-44` stores job payloads and `last_error`.
- The schema also stores memory traces, corrections, profiles, session summaries, notes, plans, documents, document chunks, groups, and memberships.
- `.agents/hermes-orchestration/run-agent.sh:79-87` stores model output and error logs.
- `.agents/coordination/agent-work.sh:63-68` stores coordination activity.

CTO assessment:

The repo's whole purpose is local memory storage, so storing user data is expected.
The public trust gap is that retention, redaction, deletion, backup, and "what is stored where" are not explained strongly enough for install-time consent.

### 5. Does team memory create privacy risk?

Yes.

Positive controls:

- `internal/recall/assembler.go:228-239` derives visible group IDs from stored `memory_group_memberships` during prefetch.
- The project documents correctly treat `agent_private`, `workspace_shared`, and `group_shared` as separate product boundaries.

Remaining risk:

- `internal/httpapi/router.go:272-278` accepts `entity_id` and `visible_group_ids` from query parameters for explain-memory.
- `internal/httpapi/router.go:312-341` accepts timeline identity fields from query parameters.
- `internal/mcp/surface.go:67-96` forwards tool inputs directly into core service methods.
- `internal/kernel/service.go:526-547` still validates correction target visibility against request-supplied `EntityID` and `VisibleGroupIDs`.
- `internal/store/postgres/search.go:50-71` filters private and group memory based on request-supplied owner and group values.
- `internal/store/postgres/memories.go:745-763` filters explain access based on request-supplied entity and group values.

CTO assessment:

Group memory is product-critical and inherently sensitive.
The storage layer has useful predicates, but without authentication and server-bound identity, those predicates can become caller-controlled access checks.

### 6. Does deployment via tunnel create risk?

Yes, if the current server is exposed through any tunnel.

Evidence:

- `cmd/server/main.go:90-93` binds the API server to `:8080`, which means all interfaces by default.
- `internal/httpapi/router.go:50-62` exposes the full `/v1` memory API surface without auth middleware.
- `cmd/server/main.go:90-93` does not configure read, write, or idle timeouts.
- `internal/httpapi/router.go:306-310` decodes JSON without an explicit body size limit.

What was not found:

- No deployment script or tunnel setup script was found in the live checkout.

CTO assessment:

There is no tunnel script, but a user could easily expose the server with ngrok, Cloudflare Tunnel, Tailscale Funnel, SSH forwarding, or IDE preview tooling.
In the current state, that would turn local memory APIs into unauthenticated network APIs.

### 7. Are update commands safe?

Partially.

Absent update surface:

- No repo self-update command was found.
- No package-manager update command was found.

Risky mutation surfaces:

- `cmd/cli/main.go:325-348`: `jobs requeue-blocked` mutates worker job state without confirmation, actor attribution, or audit event.
- `internal/store/postgres/jobs.go:456-480`: requeue changes blocked jobs to queued.
- `internal/store/postgres/notes_plans.go:164-166`: plan update with items deletes existing plan items.
- `internal/store/postgres/documents.go:100-128`: document upsert can overwrite metadata and replace chunks.
- `internal/store/postgres/memories.go:120-212`: update/correction transaction creates replacement memory, trace, `updates` edge, and supersedes the target.

Positive controls:

- Correction/update storage paths use transactions and active/latest target checks.
- The current documents describe correction supersession and provenance as first-class.

CTO assessment:

The correction path is moving in the right direction, but public install trust requires explicit confirmations or `--yes` gates for destructive operator commands, plus audit records for manual recovery actions.

### 8. Are dependencies safe and bounded?

Mixed.

Positive controls:

- `go.mod` pins direct module versions: `chi`, `pgx`, `godotenv`, and `pgvector-go`.
- `go.sum` provides module checksums.
- `go mod verify` passed.
- `GOSUMDB` is `sum.golang.org`.

Risks:

- `Makefile:54` installs `golangci-lint@latest`, which is not bounded.
- `GOPROXY` includes `direct`, so failed proxy resolution can fetch directly from origins.
- `govulncheck` is not installed in this environment, so vulnerability status was not verified.
- Tracked prebuilt binaries exist in `bin/server` and `bin/worker`.

Tracked binary evidence:

```text
bin/server: Mach-O 64-bit executable arm64
bin/worker: Mach-O 64-bit executable arm64
d7ad6b7657d9e434393c56f6bf684f1581679377689afb21ab3cb8e88fceba21  bin/server
7202091deb5b8051e1644cfed1526293149c8c8ffd2547ac8888dc2dee646b0f  bin/worker
```

CTO assessment:

Go dependencies are mostly bounded, but public release must remove tracked binaries, pin setup tools, and add a repeatable vulnerability/dependency review workflow.

### 9. Are there missing warnings or consent steps?

Yes.

Missing warnings:

- Starting `cmd/server` exposes local memory APIs on `:8080`.
- Tunneling or LAN exposure is unsafe without auth.
- `sync_turn`, documents, notes, plans, corrections, traces, and raw events persist sensitive user/project data.
- Hermes orchestration can send prompt file contents to external model tooling.
- Hermes run outputs and errors are stored under `.agents/hermes-orchestration/runs/`.
- Coordination logs can retain agent activity and file paths.
- `make clean` deletes `bin/`.
- `make setup` downloads modules and installs an unpinned external tool.
- `doctor` performs an HTTP GET to the configured embedding endpoint.

CTO assessment:

Consent should be explicit before network calls, external model calls, destructive local actions, and persistent memory writes.

### 10. Are there hardcoded secrets, risky paths, or unsafe defaults?

No hardcoded production secrets were found.

Risky paths and defaults:

- `.agents/hermes-orchestration/run-agent.sh:35-49` hardcodes `/Users/parker/.hermes` profile paths.
- `.agents/hermes-orchestration/run-agent.sh:64-66` defaults to `openai-codex`, `gpt-5.5`, and 90 max turns.
- `cmd/server/main.go:90-93` defaults to `:8080` on all interfaces.
- `internal/config/config.go:48` defaults database URL to `postgres://localhost:5432/vibegravity?sslmode=disable`.
- `internal/config/config.go:50` defaults embedding endpoint to `http://localhost:8080`, which can collide conceptually with the API server default port.
- `Makefile:54` uses an unpinned `@latest` install.
- Tracked prebuilt binaries are present under `bin/`.

CTO assessment:

The repo is not leaking obvious secrets, but it has unsafe public-install defaults and local-machine assumptions that need cleanup before release.

## High-Risk Code Paths

| Risk | File | Command or behavior | Assessment |
|---|---|---|---|
| Unauthenticated network API | `cmd/server/main.go:90-93` | Server binds to `:8080` | P0 if exposed beyond trusted localhost |
| Unauthenticated API routes | `internal/httpapi/router.go:50-62` | Full memory API exposed under `/v1` | P0 |
| No body limit | `internal/httpapi/router.go:306-310` | JSON decode without `MaxBytesReader` | P1 denial-of-service risk |
| No server timeouts | `cmd/server/main.go:90-93` | `http.Server` lacks read/write/idle timeouts | P1 denial-of-service risk |
| Caller-supplied explain identity | `internal/httpapi/router.go:272-278` | `entity_id` and `visible_group_ids` come from query | P0 auth boundary risk |
| Caller-supplied timeline identity | `internal/httpapi/router.go:312-341` | timeline identity comes from query | P0 auth boundary risk |
| MCP read/write surface | `internal/mcp/surface.go:45-96` | Tools delegate directly to core service | P0 in untrusted MCP contexts |
| Request-controlled private search | `internal/store/postgres/search.go:50-71` | Private/group filters depend on request values | P0 without auth binding |
| Request-controlled explain access | `internal/store/postgres/memories.go:745-763` | Explain visibility depends on request values | P0 without auth binding |
| Correction visibility | `internal/kernel/service.go:526-547` | Uses request entity/group values | P0/P1 depending on deployment boundary |
| External model orchestration | `.agents/hermes-orchestration/run-agent.sh:79-87` | Reads prompt and runs `hermes chat` | P1 privacy/consent risk |
| Unpinned setup tool | `Makefile:54` | Installs `golangci-lint@latest` | P1 supply-chain risk |
| Tracked binaries | `bin/server`, `bin/worker` | Mach-O arm64 executables in repo | P1 supply-chain risk |
| Manual job mutation | `cmd/cli/main.go:325-348` | Requeues blocked jobs | P1 operator safety/audit risk |
| Destructive clean | `Makefile:41` | `rm -rf bin/` | P2 if documented; P1 if hidden in install path |
| Document chunk replacement | `internal/store/postgres/documents.go:126-128` | Deletes chunks before replacement | P2/P1 depending on user expectations |
| Plan item replacement | `internal/store/postgres/notes_plans.go:164-166` | Deletes plan items before replacement | P2/P1 depending on user expectations |

## Privacy Risks

### Local memory

VibeGravity stores highly sensitive project and agent context by design.
This includes raw events, derived memories, memory traces, correction text, evidence JSON, notes, plans, document chunks, profiles, session summaries, groups, and memberships.
That is acceptable for a local memory engine only if install docs clearly explain what is stored, where it is stored, how long it persists, how to delete it, and what can leave the machine.

### Team profile and group memory

`group_shared` memory is a real privacy boundary, not just a recall filter.
The code has promising membership-backed flows for prefetch and worker source assembly, but direct HTTP and MCP surfaces still need authenticated identity binding.
Until identity is server-derived, group memory should be considered unsafe for untrusted clients.

### Code scanning

The repo does not contain broad hidden source-code scanning automation.
The header checker reads Go files, eval reads specified scenario JSON, and tests read migration/source files.
This is acceptable, but public docs should say what is scanned and which commands read user-provided paths.

### Logs

Logs and generated state can retain sensitive information:

- `ingest_jobs.last_error`
- Hermes `*.out.md` and `*.err.log`
- coordination `activity.log`
- CLI printed errors
- potential server logs through chi middleware

These are mostly ignored/generated locally, but they still need retention and redaction guidance.

### External calls

Current worker reasoning uses a deterministic mocked Codex client, which is safer than calling a real model by surprise.
However, Hermes orchestration can call external provider-backed tooling, and the future Codex bridge is explicitly part of the architecture.
The public product must make external model calls opt-in, visible, and auditable.

## Supply-Chain Risks

### Dependencies

The direct Go dependencies are small and pinned in `go.mod`.
`go.sum` checksums exist and `go mod verify` passed.
This is a good baseline.

Risks remain:

- No `govulncheck` result was available.
- No visible dependency automation or CI policy was verified in this checkout.
- `go install ...@latest` is unpinned.
- `GOPROXY` includes `direct`, which is normal but should be documented for controlled enterprise environments.

### Install flow

There is no public-safe install flow yet.
`make setup` downloads modules and installs an unpinned tool.
The server default bind address is not safe for public installation.
The repo has no root README in the current checkout, so users are likely to infer setup from internal docs and Makefile commands.

### Update flow

No self-update command was found.
That is safer than an unsafe updater, but the repo should explicitly state that updates are Git/source-driven until a signed release process exists.

### Generated files

Generated and local-state files include build outputs, Hermes run logs, and coordination state.
Tracked prebuilt `bin/server` and `bin/worker` should be removed from version control.
Public releases should provide reproducible build instructions or signed artifacts, not unreviewed binaries in the repository.

## Risk Ranking

### P0

1. Unauthenticated HTTP API exposes memory read/write operations.
2. MCP tools expose memory read/write operations without a trust boundary.
3. Identity and group visibility are request-controlled on direct surfaces.
4. Tunnel or LAN exposure would turn local memory operations into unauthenticated network operations.

### P1

1. Tracked prebuilt binaries under `bin/`.
2. Unpinned `golangci-lint@latest` install in `make setup`.
3. Hermes orchestration can send prompt files to external provider-backed tooling without a strong consent gate.
4. Manual DB mutation commands lack confirmation, actor attribution, and audit events.
5. Server lacks request body limits and read/write/idle timeouts.
6. Public docs do not sufficiently warn about persisted memory, external calls, generated logs, or tunnel exposure.

### P2

1. Hardcoded `/Users/parker/.hermes` profile paths reduce portability and expose local assumptions.
2. Coordination and Hermes run logs can retain sensitive task context.
3. `doctor` reaches a configured HTTP endpoint without an explicit "this will make a network request" warning.
4. Document and plan replacement semantics need clearer user-facing warnings.
5. Stale requested surfaces such as `setup.py` and Python package expectations can confuse public users.

## Required Fixes Before Public Release

1. Add authentication to HTTP and MCP surfaces.
2. Bind actor, entity, tenant, workspace, and group visibility to authenticated identity, not request DTOs.
3. Make the server bind to `127.0.0.1` by default and require an explicit flag/env var for non-loopback binding.
4. Add read, write, idle, and header timeouts to `http.Server`.
5. Add request body size limits for all JSON endpoints.
6. Add public warnings for tunnel exposure and require an explicit unsafe-network opt-in.
7. Add consent prompts or explicit flags before Hermes/Codex external model calls.
8. Document exactly what data is stored in PostgreSQL and generated logs.
9. Add deletion/export/retention guidance for raw events, memories, traces, corrections, documents, logs, and generated run files.
10. Remove tracked binaries from the repository and publish source-first build instructions.
11. Pin setup tools; replace `golangci-lint@latest` with a fixed version.
12. Add `govulncheck` and dependency review to the release gate.
13. Add confirmation or `--yes` gates for destructive commands such as clean, requeue, replace-document chunks, and replace-plan items.
14. Add audit records for manual operator recovery actions.
15. Replace hardcoded `/Users/parker/...` paths with repo-relative or user-configurable paths.
16. Add a public install README that distinguishes local-only dev mode from any future beta/public mode.

## Final CTO Security Decision

VibeGravity is safe for local-only use by trusted developers.
It is not safe for public install.
It is not ready for broad beta install.

The reason is not code quality alone.
The reason is trust boundary maturity: unauthenticated local surfaces, request-controlled identity, external orchestration without strong consent, tracked binaries, and incomplete install warnings.
Once auth, identity binding, loopback-safe defaults, dependency pinning, binary cleanup, and consent documentation land, the project can be reassessed for a limited beta.

## Source Review

- Estimated source: repo-local Go code, shell scripts, migrations, Makefile, and project docs.
- Suspected license exposure: no external code copied during this report.
- Similarity risk: low; this report is original review prose based on live repo inspection.
- Human review required: yes, because this is a release/security decision document and should be validated against any concurrent agent findings before public-facing publication.
