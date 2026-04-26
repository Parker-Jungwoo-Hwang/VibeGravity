# Go Binary Packaging

VibeGravity packages Go binaries first. The first supported release shape is a
private-validation binary drop for Hermes Memory, powered by VibeGravity.

## Current Decision

- Do not publish to PyPI.
- Do not build a Python wrapper.
- Publish Go binaries before any package-manager channel.
- Treat `cmd/vibegravity` as the primary operator binary.
- Keep `cmd/server` and `cmd/worker` as service binaries when a release needs
  separated processes.
- Keep `cmd/cli` as a compatibility entrypoint only; do not make it the release
  artifact name.
- Defer Homebrew formula work until after the binary release process is proven.
- Defer Docker image distribution until after live trust-loop proof and runtime
  configuration are stable.

## Toolchain

`go.mod` is the source of truth for the Go version. The current release
baseline is Go `1.26.2`.

Before preparing a release, verify the local or CI toolchain:

```bash
go version
go env GOVERSION GOTOOLCHAIN
```

Release candidates must use a supported Go patch release with no known reachable
standard-library vulnerabilities in `govulncheck ./...`.

## Migrations

VibeGravity does not embed migrations in the Go binaries for the current `v0.x`
private-validation release path.

Release artifacts must ship the `migrations/` directory beside the binary, and
operators must set or document the migration path explicitly:

```bash
export VIBEGRAVITY_MIGRATION_PATH="$(pwd)/migrations"
migrate -path "$VIBEGRAVITY_MIGRATION_PATH" -database "$VIBEGRAVITY_DB_URL" up
```

Do not publish a release where the migration source is ambiguous. A future
release may switch to embedded migrations, but that change must update this doc,
`README.md`, and `docs/release-checklist.md` in the same slice.

## Release Checks

The full release policy lives in `docs/release-process.md`.

Run the release gate before staging artifacts:

```bash
make release-gate
```

The release gate includes:

- `go test -count=1 ./...`
- `make eval`
- `make lint`
- `make check-headers`
- `git diff --check`
- `go build ./cmd/server ./cmd/worker ./cmd/vibegravity`
- `go mod verify`
- `govulncheck ./...`

## Checksums

Every release artifact must have a SHA-256 checksum before it is handed to an
operator or attached to a release.

Stage artifacts in `dist/`, then run:

```bash
make release-checksums
```

This writes:

```text
dist/checksums.txt
```

## SBOM

SBOM generation is reviewed but not yet a hard gate for the first
private-validation binary drop. If a release claims SBOM coverage, generate and
review it with:

```bash
make sbom
```

The current optional target expects `syft` and writes a CycloneDX JSON file:

```text
dist/sbom.cdx.json
```

Before public pre-release, decide whether SBOM generation becomes mandatory and
record the tool, format, and review owner in `docs/release-checklist.md`.

## Source Review

Estimated source: first-principles packaging policy from the live Go-first repo
state, official Go release policy, and local release checklist.

Suspected license: none.

Similarity risk: low.

Review required: yes before the first public release.
