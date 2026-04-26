# Changelog

All notable changes to Hermes Memory / VibeGravity will be tracked here.

The project will use SemVer. Use `v0.x.y` until the live PostgreSQL trust loop
and real Hermes/MCP path are proven. Do not tag `v1.0.0` yet.

## [Unreleased]

### Added

- Added private-validation trust docs for Hermes Memory, powered by VibeGravity.
- Added release process, rollback guide, migration rollback matrix, and release
  notes template.
- Added public `cmd/vibegravity` binary entrypoint while keeping `cmd/cli` for compatibility.
- Added `vibegravity demo`, `quickstart`, `version`, `doctor --strict`, and `doctor --json`.
- Added local-only server bind guardrails, HTTP timeouts, and request body size limits.
- Added release, privacy, status, demo, live PostgreSQL proof, and Hermes/MCP proof docs.
- Added GitHub issue templates for bugs, docs issues, feature requests, and memory scope leaks.
- Added CI and release gate documentation including `govulncheck`.
- Added Go binary packaging policy, checksum target, and optional SBOM target.

### Changed

- Root README now starts with the Hermes Memory product framing and states internal use / private-validation hardening status.
- Public docs prefer `vibegravity` commands instead of the generic `cli` name.
- Hermes bootstrap output now states that it prints a command and does not modify Hermes config automatically.
- `make setup` pins setup tools instead of installing floating tool versions.
- Release packaging is Go binary-first and keeps migrations as shipped filesystem artifacts for `v0.x`.

### Fixed

- `doctor` no longer returns success when required checks fail.
- Server no longer binds to all interfaces by default.
- Tracked build artifacts are being removed from the source tree.

### Security

- Added explicit warnings for LAN exposure, tunnels, IDE previews, request-supplied identity, and external model calls.
- Added privacy and data-handling documentation for stored records, network egress, deletion, export, and retention.
- Raised the Go version requirement to 1.26.2.

### Migration Notes

- Runtime migration behavior remains filesystem-based through `VIBEGRAVITY_MIGRATION_PATH`.
- Private-validation operators should run `migrate -path "$VIBEGRAVITY_MIGRATION_PATH" -database "$VIBEGRAVITY_DB_URL" up`.
- Release packaging must not leave migrations ambiguous.
- Migrations are not embedded in the current `v0.x` Go binary release path.

### Known Risks

- Live PostgreSQL proof is required before readiness can be claimed.
- Real Hermes/MCP tool roundtrip is required before broader beta can be considered.
- Full authentication is not implemented; keep HTTP local-only.
- Real Codex execution is disabled by default.
- Embedding model and dims are configuration gates, not proven defaults.

### Verification

- Required local gates: `make release-gate`, including `go mod verify` and `govulncheck ./...`.
- Required live gates: `make integration-postgres` and real Hermes/MCP smoke when configured.
