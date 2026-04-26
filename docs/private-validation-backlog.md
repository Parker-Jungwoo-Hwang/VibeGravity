# Private Validation Execution Backlog

This backlog turns the 2026-04-26 C-level review reports into execution work.
It is scoped to moving Hermes Memory, powered by VibeGravity, from internal-only
developer use to private-validation-ready.

Do not broaden the product. Do not prepare a public launch. Do not add product
categories, non-Go package distribution, enterprise positioning, or public beta
language.

## Verdict

Private validation remains blocked until live PostgreSQL proof and real
Hermes/MCP proof pass. Public launch remains blocked.

## P0: Private Validation Blockers

| Work | Status | Evidence |
|---|---|---|
| Root README with Hermes Memory framing | Done | `README.md` |
| License decision surface | Done, pending legal decision | `LICENSE` says license pending |
| Changelog | Done | `CHANGELOG.md` |
| Contribution guide | Done | `CONTRIBUTING.md` |
| Security policy | Done | `SECURITY.md` |
| Support policy | Done | `SUPPORT.md` |
| GitHub issue templates | Done | `.github/ISSUE_TEMPLATE/` |
| GitHub PR template | Done | `.github/pull_request_template.md` |
| Five-minute local demo package | Done | `docs/demo.md`, `examples/hermes-memory-trust-loop/` |
| Document expected demo output | Done | `examples/hermes-memory-trust-loop/expected-output.txt` |
| Run local demo | Done | `go run ./cmd/vibegravity eval demo` passed on 2026-04-26 |
| Live PostgreSQL proof | Blocked | `VIBEGRAVITY_DB_URL` unset, see `docs/live-postgres-proof.md` |
| Hermes/MCP proof | Blocked | PostgreSQL unreachable for MCP startup, see `docs/hermes-mcp-proof.md` |
| Server loopback-only default | Done | `cmd/server/main.go` |
| Unsafe non-loopback opt-in | Done | `VIBEGRAVITY_UNSAFE_ALLOW_NON_LOOPBACK=true` |
| HTTP timeouts | Done | `cmd/server/main.go` |
| JSON body size limit | Done | `internal/httpapi/router.go` |
| Stored data documentation | Done | `docs/privacy-and-data-handling.md` |
| External model call risk documentation | Done | `README.md`, `docs/privacy-and-data-handling.md` |
| Delete/export/retention guidance | Done with current limitations | `docs/privacy-and-data-handling.md` |
| Doctor exit code behavior | Done | `cmd/vibegravity`, compatibility `cmd/cli` |
| `doctor --strict` | Done | `cmd/vibegravity`, compatibility `cmd/cli` |
| `doctor --json` | Done | `cmd/vibegravity`, compatibility `cmd/cli` |
| `version` command | Done | `cmd/vibegravity`, compatibility `cmd/cli` |
| Public binary name `vibegravity` | Done | `cmd/vibegravity`, `Makefile` |
| Go-first packaging policy | Done | `docs/packaging.md` |
| Filesystem migration packaging decision | Done | `docs/packaging.md`, `README.md` |
| Release artifact checksums | Done as release target | `make release-checksums` |
| SBOM generation review | Done as optional target | `make sbom`, `docs/packaging.md` |
| Remove tracked binaries | Done in working tree | `bin/server`, `bin/worker` removed from Git tracking |
| Pin setup tooling | Done | `Makefile` |
| Add `govulncheck` release gate | Done | `Makefile`, `.github/workflows/ci.yml` |
| CI | Done | `.github/workflows/ci.yml` |
| Release checklist | Done | `docs/release-checklist.md` |
| SemVer policy | Done | `CHANGELOG.md`, `docs/release-checklist.md` |
| Release process | Done | `docs/release-process.md` |
| Rollback guide | Done | `docs/rollback-guide.md` |
| Migration rollback matrix | Done | `docs/migration-rollback-matrix.md` |
| Release notes template | Done | `docs/release-notes-template.md` |

## P1: Trust And Release Quality

| Work | Status | Next action |
|---|---|---|
| Shared runtime composition layer | Not done | Create `internal/runtime` and move common server/CLI/worker wiring there |
| Reduce repeated service wiring | Not done | Refactor after `internal/runtime` exists |
| Keep `kernel.Service` as facade | Ongoing | Avoid adding feature logic to `internal/kernel` |
| Move feature logic into smaller packages | Not started | Only after runtime proof gates are stable |
| MCP schema contract tests | Partially done | Expand tests for `recall_preview`, `correct_memory`, `explain_memory`, `view_timeline`, `degraded_status` |
| Mock Codex mode explicit | Partially done | Add runtime log/config signal in worker startup |
| Embedding implementation status clarity | Done in docs | Keep docs honest until endpoint proof exists |
| Shell script checks for `.agents` tooling | Not done | Add shell syntax/smoke tests before relying on it for release work |
| Agent skill registry | Done | `.agents/skills/README.md` |
| Workflow role docs / `phase_context.md` | Not done | Keep internal-only if added |
| Stronger handoff templates | Not done | Add strict lane/result fields before high-concurrency work |
| Safer `agent-work.sh` validation | Not done | Add `status --json`, `status --no-render`, stale warnings, path validation, option token rejection, broad glob rejection |

## P2: Useful Later

| Work | Status | Timing |
|---|---|---|
| Web docs | Deferred | After P0 proof passes |
| Business model doc | Deferred | After private validation signal |
| More examples | Deferred | After core trust-loop demo is proven live |
| Screenshots | Deferred | After stable operator flow exists |
| Demo video script | Deferred | After live proof docs pass |
| Expanded docs for scopes/correction/provenance/PostgreSQL/Hermes MCP | Partially done | Continue after P0 proof |
| Private beta CTA | Blocked | Only after live PostgreSQL and Hermes/MCP proof pass |

## Next Seven Days

1. Run live PostgreSQL proof against a migrated scratch database.
2. Run real Hermes/MCP proof through the trust-loop tools.
3. Refactor shared service wiring into `internal/runtime`.
4. Add MCP schema contract tests for the five trust-loop tools.
5. Add worker startup logging for mock versus real Codex mode.
6. Harden `.agents/coordination/agent-work.sh` validation or document exact TODOs.
7. Re-run the full release gate and update `docs/status.md`, `docs/live-postgres-proof.md`, and `docs/hermes-mcp-proof.md`.

## Source Review

Estimated source: first-principles backlog synthesized from the C-level review
reports in `c-level-review/` and the live repository state.

Suspected license: none.

Similarity risk: low.

Review required: yes before treating this as an external roadmap.
