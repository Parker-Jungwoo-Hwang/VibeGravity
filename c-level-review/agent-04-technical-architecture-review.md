# Technical Architecture Review

Date: 2026-04-26
Reviewer: Agent 04
Role: CTO review
Scope: Technical architecture, codebase structure, runtime path, workflows, skills, scripts, IDE adapters, generated adapters, and maintainability
Repo: `/Users/parker/Documents/VibeGravity`

## Original Request

Act as a CTO reviewing the GitHub repo `VibeGravity`.

Review scope:

- Technical architecture
- Codebase structure
- Root tree
- VibeGravityKit package
- `.agent` / `.agents` structure
- Workflows
- Skills
- Scripts
- `setup.py`
- `MANIFEST.in`
- Generated adapters

The review was requested as read-only.
This report records the review result after the initial read-only review was completed.

## Source-of-Truth Clarification

The live checkout is the source of truth.
Several requested surfaces appear to describe an older or expected packaging shape, not the current repo.

Observed in the live checkout:

- The repo is a Go-first monorepo.
- The live agent folder is `.agents`, not `.agent`.
- There is no `VibeGravityKit` package.
- There is no `setup.py`.
- There is no `MANIFEST.in`.
- There is no `.github/workflows` directory.
- There are no tracked Python files in the live checkout.
- There are no generated IDE adapter packages.
- `pkg/` exists but is intentionally empty.
- `bin/server` and `bin/worker` are tracked despite `bin/` being ignored.

This is architecturally important.
The current repo has mostly completed the transition away from the older Python/package shape, but some prompts and historical context still expect that older structure.

## Review Questions

This report answers:

1. What are the main modules?
2. What is the runtime path after a user runs the CLI?
3. How are workflows, skills, scripts, and IDE adapters connected?
4. Is the codebase easy to reason about?
5. Are responsibilities cleanly separated?
6. Are there single points of failure?
7. Are there naming mismatches, duplicate concepts, or confusing folder boundaries?
8. What architecture should this repo move toward?
9. Is the architecture ready for production-grade maintenance?

## Sources Reviewed

Primary files and folders reviewed:

- `AGENTS.md`
- `PLANS.md`
- `Makefile`
- `go.mod`
- `.gitignore`
- `.golangci.yml`
- `cmd/cli/main.go`
- `cmd/server/main.go`
- `cmd/worker/main.go`
- `internal/core/service.go`
- `internal/core/dto.go`
- `internal/kernel/service.go`
- `internal/ingest/service.go`
- `internal/recall/assembler.go`
- `internal/worker/processor.go`
- `internal/worker/stage2_sources.go`
- `internal/reasoning/orchestrator.go`
- `internal/reasoning/codex_bridge.go`
- `internal/reasoning/mock_codex_client.go`
- `internal/graph/apply.go`
- `internal/graph/store_apply.go`
- `internal/hermes/provider.go`
- `internal/mcp/surface.go`
- `internal/mcp/protocol.go`
- `internal/httpapi/router.go`
- `internal/store/store.go`
- `internal/store/postgres/store.go`
- `internal/store/postgres/memories.go`
- `internal/config/config.go`
- `internal/db/pool.go`
- `internal/embed/doc.go`
- `migrations/`
- `tests/`
- `tools/headercheck/main.go`
- `.agents/coordination/`
- `.agents/hermes-orchestration/`
- `.agents/skills/`
- `plans/00_read-this-first_for-building-agents.md`
- `plans/01_rfp_vibegravity_hermes-first.md`
- `plans/02_product-contract_and_direction.md`
- `plans/03_target-architecture_codex-first.md`
- `plans/05_runtime-contracts_ingest-recall-apply.md`
- `plans/06_data-model_and_storage-invariants.md`
- `docs/adr-008-package-layout.md`
- `plans/10_workpack_hermes-provider-and-external-surfaces.md`
- `docs/review-packets/project-wide-gpt-pro-review-2026-04-25/00_codex_project_review.md`
- Existing C-level review files under `c-level-review/`

Repo inventory checks were also run for:

- `VibeGravityKit`
- `setup.py`
- `MANIFEST.in`
- `.github/workflows`
- Python files
- YAML workflow files
- adapter-related files
- tracked generated binaries

## Verification Commands

The following commands were run from `/Users/parker/Documents/VibeGravity`:

```bash
go test ./...
make lint
make check-headers
make eval
git diff --check
make integration-postgres
```

Results:

- `go test ./...` passed.
- `make lint` passed.
- `make check-headers` passed.
- `make eval` passed.
- `git diff --check` passed.
- `make integration-postgres` skipped because `VIBEGRAVITY_DB_URL` was not set.

The skipped integration gate matters.
The local deterministic architecture is healthy, but live PostgreSQL production behavior is not proven by this review run.

## Executive Verdict

VibeGravity's current architecture is maintainable as a Go-first backend, but not yet production-grade.
The main code boundaries mostly match the accepted package layout in `docs/adr-008-package-layout.md`.
The core shape is clear: thin commands and transports, a shared service contract, explicit ingest and recall packages, a background worker, schema-first reasoning, graph apply, and a PostgreSQL store.
The runtime story is understandable for engineers, but still depends on duplicated composition code in `cmd/server`, `cmd/cli`, and `cmd/worker`.
The repo has strong local gates and meaningful evals, but production claims remain blocked by missing live Postgres and real external-runtime proof.
The Python-era package expectations are absent from the live repo and should either be retired or deliberately reintroduced.
My CTO decision is that the codebase has a solid architecture foundation, but it needs composition consolidation, runtime integration proof, CI, and cleanup before production-grade maintenance.

## Architecture Map

### Runtime Processes

The repo is organized around three Go command entrypoints:

| Path | Role | Current responsibility |
|---|---|---|
| `cmd/server` | HTTP API process | Loads config, opens Postgres, wires service dependencies, exposes `/healthz` and `/v1` routes |
| `cmd/worker` | Background worker process | Opens Postgres, claims jobs, runs reasoning, applies graph operations, runs dreaming jobs |
| `cmd/cli` | Operator CLI and MCP bridge | Runs doctor/eval/jobs commands, serves MCP stdio, prints Hermes bootstrap commands |

### Core Application Packages

| Path | Role | CTO assessment |
|---|---|---|
| `internal/core` | Domain DTOs, errors, service interface, domain types | Strong central contract |
| `internal/kernel` | Concrete `core.VibeGravityService` facade | Useful, but becoming too broad |
| `internal/ingest` | `sync_turn` hot path | Clean separation from reasoning and graph writes |
| `internal/recall` | Typed recall pack assembly | Strong product-critical package |
| `internal/worker` | Job claim, dispatch, Stage 2 source adapters | Reasonable orchestration boundary |
| `internal/reasoning` | Stage 1/Stage 2 schema contracts and Codex bridge boundary | Good contract shape, still mock-backed |
| `internal/graph` | Apply engine, updates/extends semantics, dreaming | Correctly conservative |
| `internal/store` | Storage interfaces | Useful contracts, but broad |
| `internal/store/postgres` | PostgreSQL store implementation | Product-critical and growing large |
| `internal/eval` | Deterministic regression and demo evals | Strong quality signal |

### Transport and Adapter Packages

| Path | Role | CTO assessment |
|---|---|---|
| `internal/httpapi` | HTTP router and handlers | Thin enough and aligned with core service |
| `internal/mcp` | MCP-style tool surface and JSON-RPC stdio server | Useful external surface, but schema duplication is a drift risk |
| `internal/hermes` | Hermes semantic provider adapter | Good thin adapter, not a real packaged Hermes provider yet |
| `internal/config` | Env and `.env` loading | Simple, but config contract needs stronger operator docs |
| `internal/db` | pgx pool and pgvector registration | Clean infra boundary |
| `internal/embed` | Intended embedding client package | Currently only documentation, not a working runtime module |

### Support and Governance

| Path | Role | CTO assessment |
|---|---|---|
| `migrations` | PostgreSQL schema evolution | Core production substrate |
| `tests` | Cross-package and golden tests | Good but live DB is opt-in |
| `tools/headercheck` | Go header policy enforcement | Simple and effective |
| `.agents/coordination` | File-claim coordination for concurrent agents | Strong internal multi-agent operating surface |
| `.agents/hermes-orchestration` | Profile-isolated Hermes dispatch scripts | Useful internal review/execution launcher |
| `.agents/skills` | Agent playbooks and policy checklists | Mostly Markdown procedures, not executable skills |
| `plans` | Product and architecture contracts | Valuable, but high volume can obscure current truth |
| `docs/review-packets` | Review, handoff, and verification artifacts | Useful historical evidence, but needs indexing discipline |

## Runtime Path After CLI

### `cli mcp serve --stdio`

This is the most important CLI runtime path for external integration.

Flow:

1. `cmd/cli/main.go` parses `mcp serve --stdio`.
2. `openPostgresService` loads config through `internal/config`.
3. `internal/db.NewPool` opens a PostgreSQL pool and registers pgvector types.
4. `postgres.NewStore` creates the root store.
5. `ingest.NewService` receives raw event and job stores.
6. `recall.NewAssembler` receives note, plan, memory, document, profile, summary, group, and freshness stores.
7. `kernel.NewService` composes the service facade.
8. `mcp.NewSurface` wraps the core service as named tools.
9. `mcp.NewServer` exposes those tools as newline-delimited JSON-RPC over stdio.

Architecture assessment:

- The path is coherent.
- It correctly shares core semantics with HTTP and Hermes.
- It duplicates service wiring that also exists in `cmd/server`.
- It depends on Postgres even for MCP startup.

### `cli hermes bootstrap`

Flow:

1. `cmd/cli/main.go` parses `hermes bootstrap`.
2. It validates the server name and command path.
3. It prints a command shaped like:

```bash
hermes mcp add vibegravity --command /path/to/cli --args mcp serve --stdio
hermes mcp test vibegravity
```

Architecture assessment:

- This is intentionally non-mutating.
- It is a bootstrap printer, not an installer.
- It does not prove a real Hermes roundtrip by itself.
- It is aligned with the current external path: MCP registration rather than custom Hermes provider packaging.

### `cli jobs ...`

Flow:

1. `cmd/cli/main.go` parses `jobs metrics`, `jobs blocked`, or `jobs requeue-blocked`.
2. It opens the PostgreSQL store.
3. It calls job store methods directly.

Architecture assessment:

- Good operational surface.
- Read-only metrics are well separated from explicit requeue behavior.
- It bypasses `kernel.Service`, which is acceptable for operator-only queue operations but should remain narrow.

### `cli eval ...`

Flow:

1. `cli eval golden` loads deterministic scenarios from `tests/golden/replay_eval.json`.
2. `cli eval demo` runs the local Hermes Memory trust-loop demo.
3. The eval layer exercises recall, correction-shaped supersession, scope separation, outage/recovery, and blocked work behavior.

Architecture assessment:

- This is one of the strongest parts of the repo.
- It makes product contracts executable.
- It does not replace live Postgres or real Hermes tests.

### `cli doctor`

Flow:

1. Loads config.
2. Prints masked database URL and embedding settings.
3. Attempts DB connection.
4. Attempts an HTTP GET to the configured embedding endpoint.

Architecture assessment:

- Useful as a first diagnostic.
- Still basic for production operation.
- It should eventually distinguish warnings from fatal readiness failures more explicitly.

## Server Runtime Path

`cmd/server/main.go`:

1. Loads config.
2. Opens Postgres.
3. Creates `postgres.Store`.
4. Wires `ingest.Service`.
5. Wires `recall.Assembler`.
6. Wires `kernel.Service`.
7. Creates `httpapi.App`.
8. Serves chi routes.

Important routes:

- `GET /healthz`
- `POST /v1/prefetch`
- `POST /v1/sync-turn`
- `POST /v1/documents`
- `POST /v1/search/memories`
- `POST /v1/search/documents`
- `POST /v1/notes`
- `POST /v1/plans`
- `PATCH /v1/plans/{id}`
- `POST /v1/memory/correct`
- `GET /v1/memory/{id}/explain`
- `GET /v1/timeline`

Assessment:

- The HTTP layer is appropriately thin.
- JSON decoding rejects unknown fields.
- Error mapping is centralized and simple.
- There is no authentication/authorization layer in the HTTP adapter yet.
- Service wiring should be extracted to a reusable runtime builder.

## Worker Runtime Path

`cmd/worker/main.go`:

1. Loads config.
2. Opens Postgres.
3. Creates `postgres.Store`.
4. Creates `graph.StoreBackedApplyEngine`.
5. Creates `graph.DreamingService`.
6. Creates store-backed Stage 2 source preparer.
7. Creates a mocked Codex reasoner.
8. Creates `worker.Processor`.
9. Polls every five seconds.
10. Claims jobs.
11. Dispatches by job kind.

Supported job kinds in `worker.Processor`:

- `process_turn_event`
- `dream_session`
- `dream_workspace`

Default `process_turn_event` path:

1. Load raw event bundle.
2. Validate bundle completeness, tenant, workspace, and actor consistency.
3. Build reasoning envelope.
4. Run Stage 1 extractor.
5. Prepare Stage 2 sources from stores.
6. Run Stage 2 resolver.
7. Validate structured output.
8. Apply graph operations through `graph.ApplyEngine`.
9. Mark job complete, failed, or blocked.

Assessment:

- The orchestration boundary is conceptually good.
- Permanent unsupported apply work becomes `blocked`, which is the right operational behavior.
- Real Codex is not wired; `cmd/worker` currently uses `MockCodexJSONClient`.
- The mock is intentionally deterministic, but the `Enabled: true` bridge config around a mock client can confuse operators reading the code.
- Embedding is not wired into the actual path yet, despite being part of the documented default pipeline.

## Workflows, Skills, Scripts, and IDE Adapters

### Workflows

The requested `.agent/workflows` structure does not exist.
The actual workflow system is:

- `.agents/coordination/README.md`
- `.agents/coordination/UNIVERSAL_AGENT_PROMPT.md`
- `.agents/coordination/PROMPT_SNIPPET.md`
- `.agents/coordination/agent-work.sh`
- `.agents/coordination/WORK_PROGRESS.md`

The coordination model is file-claim based:

1. Read `WORK_PROGRESS.md`.
2. Claim exact files.
3. Heartbeat during long work.
4. Release files when done.
5. Mark the lane done.

This is effective for expert-supervised multi-agent development.
It is not yet a polished product workflow system.

### Skills

The live `.agents/skills` files are Markdown procedures:

- `code-headers.md`
- `contract-check.md`
- `eval-regression.md`
- `plan-implement-verify.md`
- `source-provenance.md`

Only `code-headers.md` has a strong executable backing through `make check-headers` and `tools/headercheck`.
The rest are useful procedural checklists but not automated skill implementations.

### Scripts

Repo-local scripts:

- `.agents/coordination/agent-work.sh`
- `.agents/hermes-orchestration/run-agent.sh`
- `.agents/hermes-orchestration/dispatch.sh`
- `.agents/hermes-orchestration/collect.sh`
- `.agents/hermes-orchestration/status.sh`

Assessment:

- `agent-work.sh` is the most important script and is reasonably portable because it uses `/bin/sh`.
- Hermes orchestration scripts use Bash and are appropriate for macOS-local operations.
- Hermes scripts are intentionally profile-isolated through `HERMES_HOME`.
- Outputs are mostly human-readable, not stable machine-readable APIs.
- Coordination generated files are ignored, which is correct.

### IDE Adapters

No first-class IDE adapters were found.

There is no dedicated adapter for:

- VS Code
- Cursor
- Antigravity
- Claude Code
- Codex client

The real external adapter path is MCP stdio.
For V1, this is acceptable if the docs say it clearly.
The repo should avoid implying direct IDE support until it exists.

### Generated Adapters

No generated adapter package was found.

Generated or build artifacts observed:

- `bin/server`
- `bin/worker`

These are tracked in Git even though `bin/` is ignored.
That is a structural hygiene issue, not a core architecture flaw.

## Is The Codebase Easy To Reason About?

Mostly yes for backend engineers.

The good parts:

- The module map matches the product architecture.
- File headers make ownership and layer intent visible.
- The central `core.VibeGravityService` interface gives a stable mental model.
- `ingest`, `recall`, `worker`, `reasoning`, and `graph` correspond to real product phases.
- Tests and evals make important behavior concrete.

The harder parts:

- There are many planning and review documents, and not all are current truth.
- `kernel.Service` contains several feature implementations directly.
- `postgres.Store` implements many store interfaces from one root type.
- Runtime dependency construction is repeated.
- Historical Python/package references can mislead reviewers.
- The current external integration story is split across `internal/hermes`, `internal/mcp`, and CLI bootstrap output.

The codebase is reasoned about best through this path:

1. `AGENTS.md`
2. `PLANS.md`
3. `docs/adr-008-package-layout.md`
4. `internal/core/service.go`
5. `cmd/server/main.go`
6. `cmd/worker/main.go`
7. `cmd/cli/main.go`
8. `internal/kernel/service.go`
9. `internal/ingest`, `internal/recall`, `internal/worker`, `internal/reasoning`, `internal/graph`
10. `internal/store/postgres`

## Responsibility Separation

### Clean Separations

The following responsibilities are cleanly separated:

- Raw ingest is separate from derived memory apply.
- API hot path is separate from worker reasoning.
- Recall is typed before rendering.
- HTTP, MCP, and Hermes adapters share core semantics.
- Store interfaces sit between application code and Postgres.
- Eval scenarios are separate from unit tests.
- Agent coordination lives outside product runtime.

### Blurry Separations

The following responsibilities are starting to blur:

- `kernel.Service` is both facade and feature implementation surface.
- `postgres.Store` is a single concrete root for many persistence concerns.
- MCP tool schemas duplicate validation knowledge from service DTOs.
- `cmd/*` packages do too much composition work.
- Hermes provider and MCP surface overlap as external integration concepts.
- `internal/embed` is documented as important but not implemented.

### CTO Judgment

The boundaries are clean enough for the current size.
They will become harder to maintain as soon as real Codex, real embeddings, auth, packaging, and production ops land.
The next refactor should happen before those are all added.

## Single Points Of Failure

### Runtime Composition

The same service construction pattern appears in multiple command entrypoints.
If one command receives a new dependency and another does not, behavior can drift.

Recommended fix:

- Add `internal/runtime` or `internal/app` to own process composition.

### PostgreSQL Store Root

`postgres.Store` implements many interfaces:

- Raw events
- Jobs
- Memories
- Corrections
- Timeline
- Profiles
- Notes
- Plans
- Documents
- Summaries
- Dreaming
- Groups

This is convenient but creates a large shared dependency knot.

Recommended fix:

- Keep one DB pool, but split implementation files and narrow feature-facing interfaces.
- Avoid passing the full store when a feature only needs one small interface.

### Kernel Facade Growth

`internal/kernel/service.go` is taking on product behavior for:

- Documents
- Notes
- Plans
- Corrections
- Timeline
- Explain memory
- Helper validation
- Fingerprints
- Correction supersession construction

Recommended fix:

- Keep `kernel.Service` as a facade.
- Move feature logic into smaller application services before more endpoints land.

### MCP Schema Duplication

`internal/mcp/protocol.go` manually defines tool input schemas.
Those schemas can drift from DTO and service validation rules.

Recommended fix:

- Add contract tests that compare MCP required fields to service validation.
- Consider a schema helper layer close to DTO definitions.

### Mock Reasoning Boundary

The worker currently creates a mock Codex JSON client with bridge config enabled.
This is safe as a deterministic local bridge, but it is not the same architecture as real Codex.

Recommended fix:

- Make `MockCodexJSONClient` selection explicit in config and logs.
- Keep real Codex disabled by default.
- Add a real client only behind clear config, retry policy, timeout policy, and degraded-mode behavior.

### Missing Live Integration Gate

The most important production path depends on live Postgres behavior.
This review could not run that path because `VIBEGRAVITY_DB_URL` was unset.

Recommended fix:

- Provide a scratch Postgres setup path and run `make integration-postgres` in CI or release gating.

## Naming Mismatches And Confusing Boundaries

### `.agent` vs `.agents`

The prompt requested `.agent` surfaces.
The repo uses `.agents`.

Decision needed:

- Standardize on `.agents`.
- Update all reviewer prompts and public docs accordingly.

### VibeGravityKit / Python Package vs Go Monorepo

The prompt requested `VibeGravityKit`, `setup.py`, and `MANIFEST.in`.
The live repo has none of these.

Decision needed:

- If Go-first is final, retire Python packaging language.
- If a Python package is still desired, define it as a new adapter package rather than leaving it as a ghost expectation.

### Hermes Provider vs MCP Server

There are two external concepts:

- `internal/hermes.Provider` as an in-repo semantic adapter.
- `cli mcp serve --stdio` as the actual current Hermes registration path.

This is technically reasonable but easy to explain poorly.

Decision needed:

- Document that MCP is the real V1 external protocol path.
- Keep `internal/hermes.Provider` as semantic mapping and test harness until custom provider packaging exists.

### `internal/embed`

The architecture says local embeddings are part of the default path.
The live package only has `doc.go`.

Decision needed:

- Either implement the embedding client soon or clearly label embedding as planned but not wired.

### `pkg/`

`pkg/` is empty.

This is acceptable if it is explicitly reserved for future stable APIs.
It should not receive unstable internal code.

### `bin/`

`bin/` is ignored but `bin/server` and `bin/worker` are tracked.

Decision needed:

- Remove tracked binaries from Git.
- Keep build outputs local.

### Go Version

`AGENTS.md` says Go 1.22+.
`go.mod` says `go 1.25.7`.

Decision needed:

- Align the declared minimum supported Go version with the actual toolchain requirement.

## Strong Parts

### Product-Aligned Package Layout

The package layout maps to the product:

- `sync_turn` is ingest.
- `prefetch` is recall.
- worker handles background reasoning.
- graph apply owns memory writes.
- store owns Postgres semantics.
- adapters remain thin.

This is the right shape for a memory kernel.

### Central Service Contract

`internal/core/service.go` gives the repo one API meaning.
HTTP, MCP, and Hermes all call into the same contract.
This reduces the risk of surface-specific behavior drift.

### Hot Path Discipline

`internal/ingest/service.go` keeps the hot path narrow:

- validate
- normalize
- append raw events
- enqueue jobs
- return accepted

It does not perform reasoning, graph updates, profile recompute, or dreaming.
That is architecturally correct.

### Typed Recall

`internal/recall/assembler.go` builds typed blocks before rendering.
This supports the product promise because recall blocks carry:

- kind
- priority
- source
- source id
- scope
- status
- freshness
- owner

This is better than returning a single undifferentiated text blob.

### Conservative Graph Apply

`internal/graph/store_apply.go` supports a narrow set of writes and rejects unsupported work.
This is healthy for an early production backend.
It is better to block unsupported deterministic work than silently half-apply it.

### Operator-Visible Evals

`make eval` is valuable.
It exercises product-level behavior, not only implementation details.
The demo eval is an especially strong proof point for the Hermes Memory trust loop.

### Coordination Discipline

`.agents/coordination` is a strong internal operating system for concurrent agents.
It reduces accidental file collisions and makes live ownership visible.

## Weak Parts

### Composition Duplication

Runtime composition appears in `cmd/server` and `cmd/cli`.
Worker composition is separate in `cmd/worker`.

Risk:

- Dependencies may drift across entrypoints.
- Future config changes will be applied inconsistently.

### Broad Kernel Service

`kernel.Service` is useful, but it is collecting feature logic.

Risk:

- It becomes the place every new endpoint goes.
- Feature-specific invariants become harder to test and own.

### Monolithic Store Root

`postgres.Store` implements many interfaces.

Risk:

- Storage changes can become cross-feature changes.
- Testing small features may require broad fake stores.

### Mocked Critical Runtime

Real Codex and real Hermes are not part of the default verified path.

Risk:

- The architecture may look done locally while external runtime behavior remains unproven.

### Missing Embedding Implementation

The documented runtime path includes local embeddings.
The live package is not implemented.

Risk:

- Architecture docs overstate the implemented retrieval path.

### No CI

There is no `.github/workflows` directory.

Risk:

- Good local gates exist, but they depend on human or agent discipline.

### Tracked Build Outputs

Tracked `bin/server` and `bin/worker` conflict with the ignored build-output policy.

Risk:

- Diffs become noisy.
- Build artifacts can become stale or architecture-confusing.

## Structural Risks

### P0 Risks

#### P0. Live PostgreSQL Trust Loop Is Not Proven In This Review

Evidence:

- `make integration-postgres` skipped because `VIBEGRAVITY_DB_URL` was unset.
- The product contract says V1 is not ready until correction, provenance, supersession, explain/timeline, and next recall are proven against real PostgreSQL and external protocol paths.

Impact:

- Local deterministic tests are not enough for production-grade maintenance.
- The most important stateful behavior remains unverified in this run.

Recommended action:

- Add a repeatable scratch Postgres setup.
- Run `make integration-postgres` before any production readiness claim.
- Add this gate to CI or a release checklist.

#### P0. Real External Runtime Path Is Not Fully Proven

Evidence:

- `internal/hermes.Provider` is an in-repo adapter.
- `cli hermes bootstrap` prints commands but does not perform or verify real configuration.
- MCP stdio exists, but real Hermes runtime roundtrip was not run in this review.

Impact:

- The first customer integration remains partially theoretical.

Recommended action:

- Add a real MCP stdio smoke test against Hermes where the environment allows it.
- Keep bootstrap explicitly non-mutating until installation semantics are decided.

### P1 Risks

#### P1. Runtime Composition Drift

Evidence:

- `cmd/server/main.go` and `cmd/cli/main.go` separately construct the service.
- `cmd/worker/main.go` separately constructs worker dependencies.

Impact:

- New dependencies or config can land in one process but not another.

Recommended action:

- Create `internal/runtime` for shared composition.

#### P1. Kernel Facade Becoming A Feature Bucket

Evidence:

- `internal/kernel/service.go` contains direct logic for documents, notes, plans, correction, timeline, explain, helper validation, and fingerprinting.

Impact:

- Feature ownership becomes less clear.
- Tests become harder to focus.

Recommended action:

- Move feature logic into smaller application packages while preserving `core.VibeGravityService` as facade.

#### P1. MCP Schema Drift

Evidence:

- Tool input schemas are hand-written in `internal/mcp/protocol.go`.
- Core validation lives elsewhere.

Impact:

- MCP clients can discover one contract and hit a different runtime contract.

Recommended action:

- Add service-backed MCP schema contract tests.
- Consider generating or centralizing required field definitions.

#### P1. Mock Codex Wiring Can Mislead Operators

Evidence:

- `cmd/worker/main.go` creates `MockCodexJSONClient`.
- It passes `CodexBridgeConfig{Enabled: true}` to the mock runners.

Impact:

- It exercises the bridge, but does not mean real Codex is enabled.

Recommended action:

- Rename/log this as mock mode.
- Wire real Codex only through explicit config and failure-mode policy.

#### P1. Missing Embedding Runtime Implementation

Evidence:

- `internal/embed` only contains package documentation.
- Architecture docs describe local embedding as part of the default worker pipeline.

Impact:

- Retrieval architecture is not fully implemented.

Recommended action:

- Implement the embedding client or narrow the docs until it exists.

### P2 Risks

#### P2. Tracked Generated Binaries

Evidence:

- `git ls-files bin` lists `bin/server` and `bin/worker`.
- `.gitignore` ignores `bin/`.

Impact:

- Build artifacts can confuse code review and release state.

Recommended action:

- Run `git rm --cached bin/server bin/worker`.

#### P2. Missing Public Package Boundary

Evidence:

- `pkg/` is empty.
- No Python package exists.
- No public library API is defined.

Impact:

- The repo is clear as an internal backend, but not as an installable SDK or library.

Recommended action:

- Keep `pkg/` empty until a stable API exists.
- Do not reintroduce `VibeGravityKit` without a clear product reason.

#### P2. Historical Review Pack Volume

Evidence:

- `docs/review-packets` contains many detailed packets.
- Several older packet names describe predecessor states.

Impact:

- Current truth can be hard for a new maintainer to identify.

Recommended action:

- Maintain a current review-packet index.
- Mark old packets as historical when contracts have moved on.

#### P2. Go Version Messaging Mismatch

Evidence:

- `AGENTS.md` says Go 1.22+.
- `go.mod` says Go 1.25.7.

Impact:

- Contributors may use the wrong toolchain.

Recommended action:

- Align docs and `go.mod`.

## Recommended Architecture

The target architecture should remain a Go-first monorepo.
The repo should not split into microservices yet.
The next move should be a cleaner internal composition and ownership structure.

### Target Module Shape

Recommended high-level structure:

```text
cmd/
  server/
  worker/
  cli/

internal/
  runtime/          # shared process composition and dependency wiring
  core/             # stable DTOs, service interface, domain types
  kernel/           # facade implementing core.VibeGravityService
  ingest/           # sync_turn hot path
  recall/           # typed recall assembly
  corrections/      # correction intake and supersession application
  documents/        # document add/search orchestration
  plans/            # plan create/update/read orchestration
  timeline/         # operator timeline orchestration
  worker/           # job claim and dispatch
  reasoning/        # Codex structured-output boundary
  graph/            # apply engine and graph semantics
  embed/            # local embedding client and retrieval helpers
  httpapi/          # HTTP transport
  mcp/              # MCP transport and schema surface
  hermes/           # Hermes semantic adapter
  store/            # storage interfaces
  store/postgres/   # Postgres implementations
  config/
  db/
  eval/
```

This structure keeps the current good boundaries, but prevents `kernel.Service` from becoming the only application layer.

### Dependency Direction

Recommended dependency direction:

```text
cmd/* -> internal/runtime
internal/runtime -> config/db/store + application services
httpapi/mcp/hermes -> core.VibeGravityService
kernel -> feature services
feature services -> narrow store interfaces
worker -> reasoning/graph/embed/store
store/postgres -> store interfaces + core records
```

### Runtime Composition Principle

Every process should share composition where possible:

- Server composition
- CLI MCP composition
- Worker composition
- Test composition

This avoids hidden differences between HTTP and MCP behavior.

### Protocol Contract Principle

MCP, HTTP, and Hermes should not redefine required fields separately.
The repo should move toward a single contract source for:

- required fields
- default values
- visibility inputs
- scope handling
- error mapping
- degraded/freshness metadata

### Production Proof Principle

The architecture should not call itself production-grade until these are true:

- Live Postgres integration gate passes.
- MCP stdio smoke works against the intended client path.
- Hermes bootstrap has a tested roundtrip or is clearly documented as command printing only.
- Real Codex client is either disabled by default with honest degraded behavior or enabled with explicit retry/timeout/failure policy.
- Embedding runtime is either implemented or explicitly out of the current runtime path.

## Refactor Plan

### Step 1: Add Shared Runtime Composition

Create:

- `internal/runtime/service.go`
- `internal/runtime/worker.go`

Move repeated construction from:

- `cmd/server/main.go`
- `cmd/cli/main.go`
- `cmd/worker/main.go`

Acceptance criteria:

- `cmd/server` becomes mostly process lifecycle and HTTP startup.
- `cmd/cli` calls shared service construction for MCP.
- `cmd/worker` calls shared worker construction.
- Tests can construct runtime dependencies without duplicating command code.

### Step 2: Keep `kernel.Service` As Facade Only

Extract feature logic from:

- `internal/kernel/service.go`

Candidate packages:

- `internal/corrections`
- `internal/documents`
- `internal/timeline`
- `internal/plans`

Acceptance criteria:

- `kernel.Service` routes calls.
- Feature packages own validation, fingerprints, and operation construction.
- Existing `core.VibeGravityService` remains stable.

### Step 3: Stabilize MCP Contract Surface

Refactor:

- `internal/mcp/protocol.go`
- `internal/mcp/protocol_test.go`
- `internal/mcp/surface.go`

Acceptance criteria:

- Tool schemas match service-required fields.
- Tests exercise at least one real service-validation path per trust-loop tool.
- `recall_preview`, `correct_memory`, `view_timeline`, and `explain_memory` remain schema-complete.

### Step 4: Make Worker Reasoning Mode Explicit

Refactor:

- `cmd/worker/main.go`
- `internal/reasoning/codex_bridge.go`
- `internal/reasoning/mock_codex_client.go`
- `internal/config/config.go`

Acceptance criteria:

- Mock mode is explicitly named and logged.
- Real Codex remains disabled by default.
- Enabling real Codex requires endpoint/model/client construction and failure behavior.

### Step 5: Implement Or De-scope Embeddings

Work in:

- `internal/embed/`
- `internal/worker/`
- `internal/recall/`
- `plans/03_target-architecture_codex-first.md`
- `plans/05_runtime-contracts_ingest-recall-apply.md`

Acceptance criteria:

- Either a real local embedding client exists, or docs state that current retrieval is lexical/store-backed only.
- Worker path documentation matches live code.

### Step 6: Clean Generated Artifacts

Work in:

- `bin/server`
- `bin/worker`
- `.gitignore`

Acceptance criteria:

- `bin/server` and `bin/worker` are removed from tracked files.
- `bin/` remains ignored.
- Build output is generated locally by `make build`.

### Step 7: Add CI

Create:

- `.github/workflows/ci.yml`

Minimum CI commands:

```bash
go test ./...
make eval
make lint
make check-headers
git diff --check
```

Optional release/integration lane:

```bash
make integration-postgres
```

Acceptance criteria:

- CI protects the same local gates that agents already run.
- Live Postgres gate is opt-in or uses a service container if feasible.

### Step 8: Align Version And Packaging Truth

Work in:

- `go.mod`
- `AGENTS.md`
- `PLANS.md`
- future root `README.md`

Acceptance criteria:

- Go version requirement is clear.
- Python packaging references are removed or explicitly marked historical.
- `VibeGravityKit` is not referenced as a current package unless it is intentionally restored.

### Step 9: Current Truth Index

Work in:

- `PLANS.md`
- `docs/review-packets/00-workpack-03-review-index.md`
- possibly a new `docs/review-packets/current-index.md`

Acceptance criteria:

- New agents can quickly identify current source-of-truth packets.
- Historical predecessor packets are not mistaken for active contracts.

## Final CTO Decision

The architecture is not ready for production-grade maintenance.

It is ready for serious continued backend development.
The core architecture is much stronger than a scaffold:

- The package layout is coherent.
- The service contract is stable.
- The ingest and recall paths are clean.
- Worker and graph apply are conservative.
- Local tests and evals are meaningful.
- Adapter boundaries are mostly thin.

But production-grade maintenance needs more than a good internal shape.
The repo still needs:

- shared runtime composition;
- live PostgreSQL trust-loop proof;
- real external protocol smoke tests;
- explicit real-vs-mock reasoning mode;
- embedding implementation or honest de-scope;
- CI;
- tracked binary cleanup;
- package/version truth alignment.

CTO decision:

Do not broaden into new feature surfaces until the architecture cleanup and live integration proof are complete.
The next best investment is a narrow platform-hardening slice: shared runtime composition, CI, and live Postgres/MCP trust-loop gates.

## Source Review

- Estimated source: live in-repo VibeGravity source, plans, ADRs, review packets, local command output, and Agent 04 CTO analysis.
- Suspected license: project-internal original material.
- Similarity risk: low.
- Review required: yes.
- Notes: This report is an original architecture review based on local repo inspection. No external source code or license-restricted material was used.
