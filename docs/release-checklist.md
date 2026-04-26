# Release Checklist

Hermes Memory / VibeGravity uses SemVer. Use `v0.x.y` until the live trust loop
and Hermes/MCP path are proven. Do not tag `v1.0.0` yet.

This checklist is for release candidates, private validation drops, and future
public tags. It is intentionally conservative because VibeGravity handles scoped
agent memory, provenance, corrections, and potentially sensitive project data.

## Release Type

- [ ] Private validation drop.
- [ ] Public pre-release tag.
- [ ] Public stable tag.

Version:

```text
v0.x.y
```

Release owner:

```text
TBD
```

## Required Local Checks

- [ ] Clean worktree.
- [ ] Public release worktree confirmed clean with `git status --short --branch`.
- [ ] Changelog entry.
- [ ] Version decision.
- [ ] SemVer decision is `v0.x.y` or `v0.x.y-rc.n`; `v1.0.0` is blocked until
      the V1 conditions in `docs/release-process.md` are met.
- [ ] Local Go toolchain matches `go.mod` (`go version` and `go env GOVERSION`).
- [ ] Migration review.
- [ ] Migration rollback status reviewed in `docs/migration-rollback-matrix.md`.
- [ ] Rollback plan reviewed in `docs/rollback-guide.md`.
- [ ] Release notes drafted from `docs/release-notes-template.md`.
- [ ] `make release-gate`
- [ ] `go test -count=1 ./...`
- [ ] `make eval`
- [ ] `make lint`
- [ ] `make check-headers`
- [ ] `git diff --check`
- [ ] `go build ./cmd/server ./cmd/worker ./cmd/vibegravity`
- [ ] `go mod verify`
- [ ] `govulncheck ./...`

## Required Trust-Loop Checks

- [ ] Live PostgreSQL gate: `make integration-postgres`
- [ ] Live PostgreSQL gate did not skip because `VIBEGRAVITY_DB_URL` was unset.
- [ ] Correction writes replacement memory, mandatory trace, `updates` edge, and
      prior-memory supersession.
- [ ] Retry/idempotency does not duplicate memories, traces, edges, or
      correction artifacts.
- [ ] Superseded memory is suppressed from normal current recall.
- [ ] `explain_memory` and `view_timeline` remain scope-aware.
- [ ] `agent_private` recall requires owner matching.
- [ ] `group_shared` memory requires valid membership where applicable.
- [ ] Stale/degraded recall is labeled honestly.
- [ ] Hermes/MCP smoke covers `recall_preview`, `correct_memory`,
      `explain_memory`, `view_timeline`, and degraded status.

## CLI Smoke

- [ ] `bin/vibegravity version`
- [ ] `bin/vibegravity eval demo`
- [ ] `bin/vibegravity hermes bootstrap --name vibegravity --command "$(pwd)/bin/vibegravity"`

## Documentation Checks

- [ ] `README.md` matches current commands and status.
- [ ] `CHANGELOG.md` has the release entry.
- [ ] `LICENSE` is present and correct.
- [ ] `CONTRIBUTING.md` describes current gates.
- [ ] `SECURITY.md` has a private reporting path or caveat.
- [ ] `SUPPORT.md` sets support boundaries.
- [ ] `CODE_OF_CONDUCT.md` is present if accepting public contributions.
- [ ] `docs/status.md` does not overclaim readiness.
- [ ] `docs/privacy-and-data-handling.md` reflects data handling and current
      limitations.
- [ ] `docs/live-postgres-proof.md` is updated for the release candidate.
- [ ] `docs/hermes-mcp-proof.md` is updated for the release candidate.

## Packaging Notes

- [ ] Go binary release only; no Python wrapper or Python package artifact.
- [ ] Primary operator artifact is named `vibegravity`.
- [ ] Homebrew distribution is explicitly deferred.
- [ ] Docker image distribution is explicitly deferred.
- [ ] Build artifacts are reproducible.
- [ ] Generated binaries are not accidentally tracked unless explicitly part of
      the release.
- [ ] Migrations are not embedded for the current `v0.x` release path.
- [ ] Release artifacts ship `migrations/` and document
      `VIBEGRAVITY_MIGRATION_PATH`.
- [ ] Migration order and rollback notes are documented.
- [ ] Annotated tag message includes gate results, migration rollback summary,
      and known risks.
- [ ] No lightweight release tag is used.
- [ ] SHA-256 checksums exist for every release artifact
      (`make release-checksums`).
- [ ] SBOM decision is recorded. If SBOM coverage is claimed, attach
      `dist/sbom.cdx.json` from `make sbom`.
- [ ] Known limitations are listed in release notes.
- [ ] Install or bootstrap instructions were tested from a clean checkout.

## No-Go Conditions

- Live PostgreSQL gate skipped or failed.
- Hermes/MCP smoke skipped or failed.
- License decision unclear for a public release.
- Auth, privacy, or trust-boundary docs missing.
- Migrations ambiguous.
- Missing artifact checksums.
- Go toolchain version is older than `go.mod`.
- `go mod verify` or `govulncheck ./...` skipped or failed.
- Scope leakage, provenance loss, or correction supersession regression is open.
- Release notes imply V1 readiness before live trust-loop proof exists.

## Final Steps

- [ ] Create release notes.
- [ ] Confirm the release tag name.
- [ ] Create annotated tag with
      `git tag -a v0.x.y -F dist/release-note-v0.x.y.md`.
- [ ] Push tag.
- [ ] Update `docs/status.md` after the release.

## Source Review

Estimated source: current `PLANS.md`, `docs/status.md`,
`docs/privacy-and-data-handling.md`, existing release-checklist draft, and local
verification commands.

Suspected license: none.

Similarity risk: low.

Review required: yes before first public tag.
