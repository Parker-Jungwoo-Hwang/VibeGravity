# Contributing to VibeGravity

VibeGravity is the agent memory engine behind Hermes Memory. Contributions
should protect the V1 trust loop: recall preview, explain and timeline,
correction, supersession, provenance, idempotent replay, scope visibility, and
honest degraded recall.

## Start Here

Before changing product, runtime, storage, or protocol behavior, read the
public operator-facing docs:

1. `README.md`
2. `docs/privacy-and-data-handling.md`
3. `docs/packaging.md`
4. `docs/demo.md`
5. `tests/README.md`

Internal planning, agent coordination notes, and review packets are intentionally
not part of the public source distribution.

## Product Boundaries

- Keep Hermes-first delivery.
- Keep VibeGravity as an agent memory engine, not a chat UI or generic vector
  database.
- Keep raw events and derived memories separate.
- Keep `agent_private`, `workspace_shared`, `group_shared`, and
  `session_scratch` boundaries explicit.
- Keep local runtime embedding-first in V1.
- Keep Codex reasoning schema-first and structured JSON only.
- Do not broaden into real Codex default enablement, custom Hermes registry
  packaging, or new product promises unless the current plan explicitly calls
  for it.

## Development Workflow

Use focused changes. Prefer the smallest slice that proves the contract or fixes
the issue.

For code changes:

1. Update or add tests near the behavior.
2. Keep Go file headers consistent with neighboring files.
3. Update public docs when behavior changes.
4. Run the local deterministic gate.
5. Report remaining risks honestly.

Keep changes focused on the user-visible program, runtime behavior, tests, or
public documentation.

## Verification

Run the deterministic local gate before handoff:

```bash
go test ./...
make eval
make lint
make check-headers
git diff --check
go mod verify
govulncheck ./...
```

For release candidates and private-validation drops, run:

```bash
make release-gate
```

For storage, correction, lineage, or protocol trust-loop work, also run the live
PostgreSQL gate against a scratch database when applicable:

```bash
createdb vibegravity_integration
export VIBEGRAVITY_DB_URL='postgres://localhost:5432/vibegravity_integration?sslmode=disable'
migrate -path migrations -database "$VIBEGRAVITY_DB_URL" up
make integration-postgres
```

If `VIBEGRAVITY_DB_URL` is unset, the live gate will skip explicitly. Do not
claim full PostgreSQL readiness from skipped live tests.

## Open-Source Code Policy

Code must be original to this repo or derived only from commercially usable
permissive patterns.

- Do not reference or closely reproduce GPL, AGPL, LGPL, SSPL, Elastic License,
  or related license-family code.
- Use MIT, BSD, Apache-2.0, official documentation, or first-principles design
  as the acceptable reference boundary.
- Do not copy an external project's file structure, comments, function names,
  or distinctive implementation shape.
- If similarity risk appears, stop and ask for review before coding.

For code-bearing handoffs, include a source review:

```text
Source Review:
- Estimated source:
- Suspected license:
- Similarity risk:
- Review required:
- Notes:
```

## Commit Messages

Use short imperative commit messages and keep one commit focused on one logical
change.
