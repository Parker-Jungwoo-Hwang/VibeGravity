# Hermes Memory, powered by VibeGravity

Stop repeating context.
Fix memory once.
See why Hermes remembered it.

VibeGravity is the engine behind Hermes Memory. It turns Hermes activity into scoped, explainable, correctable memory so a solo builder or technical Hermes operator can carry project context across sessions without pasting the same rules again and again.

VibeGravity is currently for solo builders and technical Hermes operators. The project is in internal use and private-validation hardening. It is not ready for public launch. The local deterministic demo works. Live PostgreSQL proof and a real Hermes/MCP roundtrip are required before a private-validation drop is treated as proven.

## First Command

```bash
go run ./cmd/vibegravity eval demo
```

This demo needs no database, no Hermes runtime, no Codex, and no network. It is the first confidence path because it shows the product value locally in under five minutes.

Expected output:

```text
PASS	demo initial recall shows rule plan and trust metadata	blocks=[pinned_note active_plan profile_static memory]	tokens=56	sources=[notes plans profile memories]
PASS	demo explain shows recalled memory provenance	blocks=[explain_memory]	tokens=0	sources=[memory_trace]
PASS	demo correction writes supersession	blocks=[update_memory]	tokens=0	sources=[memories memory_trace memory_edges]
PASS	demo next recall uses correction	blocks=[pinned_note active_plan memory profile_static]	tokens=57	sources=[notes plans memories profile]
PASS	demo private scope separation	blocks=[pinned_note active_plan profile_static memory]	tokens=57	sources=[notes plans profile memories]
Hermes Memory demo eval passed.
```

Each line proves one trust step:

- Initial recall includes a project rule, an active plan, a profile block, a memory block, scope/source metadata, and a compact token estimate.
- Explain shows why a recalled memory exists by exposing `memory_trace`.
- Correction writes a replacement memory, provenance trace, and `updates` edge.
- Next recall uses the corrected memory and suppresses the stale one.
- Private scope separation keeps another actor's private memory out of Hermes recall.

## What This Is

Hermes Memory is a private-validation candidate for Hermes operators. VibeGravity is the Go engine underneath it. It stores raw events separately from derived memories, assembles compact recall packs, exposes provenance, and lets an operator correct wrong memory once so later recall changes.

## What This Is Not

This is not a public beta, chat UI, generic agent runtime, generic vector database, hosted SaaS, enterprise launch, Product Hunt launch, or broad GitHub growth project. Documents, Dreaming, and future adapters are supporting layers, not the V1 headline.

## Who It Is For

The first target user is a solo builder or technical Hermes operator who already runs local developer tools and wants Hermes to remember scoped project context across sessions.

## Why It Beats Manual Prompting

Manual prompting makes the operator repeat rules, plans, corrections, and private/public boundaries. Hermes Memory makes those facts structured, scoped, inspectable, and correctable. The operator can see what Hermes is about to receive, ask why a memory appeared, fix it once, and expect later recall to suppress the old memory.

## What Works Today

- Local deterministic demo: `go run ./cmd/vibegravity eval demo`.
- Golden evals through `make eval`.
- Go core service contracts for ingest, recall, notes, plans, documents, correction, explain, and timeline.
- PostgreSQL migrations and a live integration gate entrypoint.
- MCP stdio server implementation and Hermes bootstrap command printer.
- Scope-aware recall guardrails for `agent_private`, `workspace_shared`, and membership-gated `group_shared` memory in the implemented paths.
- Correction-driven supersession in the core/store-backed trust loop tests.

## What Is Not Ready Yet

- V1 readiness is not proven until the live PostgreSQL trust-loop gate passes in a scratch database.
- Real Hermes/MCP roundtrip is not proven until Hermes invokes `recall_preview`, `correct_memory`, `explain_memory`, `view_timeline`, and `degraded_status` through the MCP path.
- Real Codex calls are disabled by default. Current worker wiring logs and uses `MockCodexJSONClient`; future real Codex requires explicit `VIBEGRAVITY_CODEX_ENABLED=true`, `VIBEGRAVITY_CODEX_CLIENT=real`, endpoint, and model config.
- Embedding behavior is not a validated product claim. Current recall and Stage 2 retrieval are store-backed lexical paths; `internal/embed` is reserved for a later explicit embedding client slice.
- HTTP identity is not a complete authentication system. The server is local-only by default while auth is incomplete.

## Known Limitations

- Internal use / private-validation hardening only.
- No public launch support, hosted support, enterprise support, or broad adapter support.
- `VIBEGRAVITY_DB_URL` is required for live PostgreSQL gates.
- `VIBEGRAVITY_EMBEDDING_MODEL=pending` and `VIBEGRAVITY_EMBEDDING_DIMS=0` are degraded setup signals.
- Current retrieval is lexical/store-backed, not semantic vector retrieval.
- Request DTOs still carry identity fields in several HTTP paths; do not expose the server beyond loopback until authenticated identity is implemented.
- Local evals prove deterministic behavior, not real Hermes or real PostgreSQL behavior.

## Real Setup Prerequisites

- Go 1.26.2, matching `go.mod`.
- PostgreSQL with the VibeGravity migrations applied.
- `golang-migrate` for migration application.
- Optional for live recall validation: local embedding runtime, explicit embedding model, and embedding dims.
- Optional for external proof: Hermes CLI with MCP support.
- Optional release/security gate: `govulncheck`.

## Setup And Local Artifacts

`make setup` performs external downloads: it runs `go mod download` and installs
pinned Go tools including `golangci-lint` and `govulncheck`.

`make build` writes local binaries under `bin/`. Build outputs are local
artifacts only; `bin/` stays gitignored and should not be committed.

`make clean` removes the entire `bin/` directory.

## PostgreSQL Setup

Use a scratch database for private validation:

```bash
createdb vibegravity_integration
export VIBEGRAVITY_DB_URL='postgres://localhost:5432/vibegravity_integration?sslmode=disable'
```

## Migration Setup

VibeGravity currently uses filesystem migrations. Keep the migration path explicit:

```bash
export VIBEGRAVITY_MIGRATION_PATH="$(pwd)/migrations"
migrate -path "$VIBEGRAVITY_MIGRATION_PATH" -database "$VIBEGRAVITY_DB_URL" up
make integration-postgres
```

Current `v0.x` Go binary packaging does not embed migrations. Release artifacts
must ship `migrations/` beside the binary and document `VIBEGRAVITY_MIGRATION_PATH`.
See `docs/packaging.md`.

## Packaging Policy

VibeGravity is Go binary-first:

- no PyPI publishing;
- no Python wrapper;
- primary release binary is `vibegravity`;
- Homebrew and Docker distribution are deferred;
- release artifacts need SHA-256 checksums;
- SBOM generation is reviewed but not yet a hard gate.

Release process, rollback, migration rollback status, and release note template
live in `docs/release-process.md`, `docs/rollback-guide.md`,
`docs/migration-rollback-matrix.md`, and `docs/release-notes-template.md`.

## Hermes MCP Registration

Build the validation binary:

```bash
make build
```

Print the registration command:

```bash
bin/vibegravity hermes bootstrap --name vibegravity --command "$(pwd)/bin/vibegravity"
```

The bootstrap command only prints a Hermes MCP registration command. It does not modify Hermes config automatically.

The shape is:

```bash
hermes mcp add vibegravity --command "$(pwd)/bin/vibegravity" --args mcp serve --stdio
hermes mcp test vibegravity
```

## Undo Hermes MCP Registration

```bash
hermes mcp remove vibegravity
```

If your Hermes CLI uses a different removal command, inspect `hermes mcp --help` and record the exact rollback command in `docs/hermes-mcp-proof.md`.

## What Data Gets Stored

VibeGravity may store raw events, derived memories, memory traces, corrections, notes, plans, documents, document chunks, profiles, session summaries, groups, memberships, job payloads, and logs. See `docs/privacy-and-data-handling.md` for deletion, export, and retention guidance.

## What May Leave The Machine

The deterministic demo sends nothing over the network. In live operation, external calls may occur only when explicitly configured: future local embedding endpoint calls, Hermes MCP stdio traffic, and future Codex reasoning calls when `VIBEGRAVITY_CODEX_ENABLED=true` and `VIBEGRAVITY_CODEX_CLIENT=real`. Do not enable external model calls with private memory unless you have reviewed the data path.

## Security Warnings

The HTTP server binds to `127.0.0.1:8080` by default. Non-loopback binding requires `VIBEGRAVITY_HTTP_ADDR` plus `VIBEGRAVITY_UNSAFE_ALLOW_NON_LOOPBACK=true`.

Do not expose VibeGravity through LAN binds, tunnels, ngrok, Cloudflare Tunnel, Tailscale Funnel, IDE preview ports, or public reverse proxies without authentication. Until full auth is implemented, do not trust request-supplied tenant, workspace, actor, entity, or visible group IDs across a network boundary.

## User-Facing Binary

Private-validation docs use:

```bash
vibegravity demo
vibegravity quickstart
vibegravity doctor
vibegravity doctor --strict
vibegravity doctor --json
vibegravity version
vibegravity eval demo
vibegravity mcp serve --stdio
vibegravity hermes bootstrap
```

`cmd/cli` remains as a compatibility entrypoint during the rename, but new private-validation docs should prefer `cmd/vibegravity` or the built `vibegravity` binary.

## Verification

Local gates:

```bash
go test -count=1 ./...
make eval
make lint
make check-headers
git diff --check
go mod verify
govulncheck ./...
```

Release-candidate gate:

```bash
make release-gate
```

Live gates:

```bash
make integration-postgres
hermes mcp test vibegravity
```

Do not claim a proven private-validation drop if the live PostgreSQL proof or Hermes/MCP proof is skipped.
