# Status

Hermes Memory, powered by VibeGravity, is in internal use and
private-validation candidate hardening.

It is not ready for public launch, public beta, Product Hunt, Hacker News,
enterprise positioning, or broad GitHub growth.

## Product Position

VibeGravity is the agent memory engine behind Hermes Memory. The V1 promise is:

> Hermes remembers the right project context across sessions, shows why it
> remembered it, and lets the operator fix memory once.

The active product proof is the trust loop: recall preview, explain/timeline,
correction, correction-driven supersession, visible scope, provenance,
idempotent replay, and honest degraded-state labeling.

## Current Proof

- Local deterministic demo works: `go run ./cmd/vibegravity eval demo`.
- Local eval gate exists: `make eval`.
- Live PostgreSQL gate exists: `make integration-postgres`.
- MCP stdio server and Hermes bootstrap command printer exist.
- HTTP, MCP, CLI, and Hermes-facing adapter surfaces exist as current
  integration seams.
- `CorrectMemory`, `GetTimeline`, and `Prefetch` have trust-loop behavior in
  the current code path.
- Read-only worker backlog metrics and degraded recall freshness labeling exist.

## Not Yet Proven

- Live PostgreSQL trust loop must pass on a migrated scratch database before
  V1 readiness is claimed.
- Real Hermes/MCP roundtrip must prove `recall_preview`, `correct_memory`,
  `explain_memory`, `view_timeline`, and `degraded_status`.
- Full authenticated identity is not implemented.
- Real Codex calls are disabled by default. The worker currently logs and uses
  `MockCodexJSONClient`; future real Codex requires explicit
  `VIBEGRAVITY_CODEX_ENABLED=true`, `VIBEGRAVITY_CODEX_CLIENT=real`, endpoint,
  and model configuration.
- Embedding runtime behavior is out of the current slice. Retrieval is
  store-backed lexical today; semantic/vector retrieval is not a ready product
  claim until endpoint, model, dims, embedding writes, and retrieval proof are
  verified.
- Custom Hermes memory provider registry packaging is not done.
- Full session replay and production operations are not done.

## Local Verification

Default deterministic gate:

```bash
go test ./...
make eval
make lint
make check-headers
git diff --check
```

Live PostgreSQL gate:

```bash
make integration-postgres
```

If `VIBEGRAVITY_DB_URL` is unset, the live PostgreSQL gate skips explicitly. A
skip is useful signal, but it is not proof of PostgreSQL readiness.

## Current Private-Validation Work

The private-validation trust layer is being filled in:

- `README.md`
- `LICENSE`
- `CHANGELOG.md`
- `CONTRIBUTING.md`
- `SECURITY.md`
- `SUPPORT.md`
- `CODE_OF_CONDUCT.md`
- GitHub issue and pull request templates
- privacy and data-handling docs

These files are release-discipline and reviewer-confidence material. They are
not a public-launch package.

## Readiness Rule

Do not claim V1 readiness until live PostgreSQL proof and Hermes/MCP proof pass.

## Source Review

Estimated source: current `PLANS.md`, existing local status draft, and
VibeGravity public-readiness docs.

Suspected license: none.

Similarity risk: low.

Review required: yes before using this as public launch status.
