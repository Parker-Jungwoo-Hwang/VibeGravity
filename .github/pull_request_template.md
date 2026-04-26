# Summary

<!-- What changed, and why does it matter for VibeGravity / Hermes Memory? -->

Private-validation rule: do not claim public launch, public beta, or V1 readiness unless the live PostgreSQL and Hermes/MCP proof gates passed.

## Files Changed

<!-- List the main files or folders changed. -->

## Risk Level

<!-- Choose one and explain briefly: P0/private-validation blocker, P1/trust quality, P2/follow-up, low docs-only. -->

## Commands Run

<!-- Include exact commands and results. -->

## Commands Skipped And Why

<!-- Especially explain skipped live PostgreSQL, Hermes/MCP, govulncheck, lint, or migration checks. -->

## Product / Contract Impact

<!-- Check any that apply. -->

- [ ] No product or runtime contract change
- [ ] Affects recall / prefetch behavior
- [ ] Affects sync / ingest behavior
- [ ] Affects correction, supersession, explain, or timeline
- [ ] Affects scope visibility or access control
- [ ] Affects MCP, HTTP, Hermes, or CLI surface
- [ ] Affects PostgreSQL schema, migrations, or storage invariants
- [ ] Updates docs, plans, review packets, or release metadata only

## Verification

<!-- Mark what ran. Explain skips, especially live PostgreSQL skips. -->

- [ ] `go test ./...`
- [ ] `make eval`
- [ ] `make lint`
- [ ] `make check-headers`
- [ ] `git diff --check`
- [ ] `make integration-postgres`
- [ ] `go mod verify`
- [ ] `govulncheck ./...`
- [ ] Not run; reason:

## Documentation

- [ ] Docs were updated.
- [ ] Docs were not needed; reason:

## Security Impact

- [ ] Security impact was checked.
- [ ] Security impact was not checked; reason:

## Trust-Loop Checklist

<!-- For behavior changes, confirm these remain true or explain why not applicable. -->

- [ ] Raw events and derived memories remain separate.
- [ ] Writes remain idempotent.
- [ ] Memory scope remains explicit.
- [ ] Provenance remains visible through `memory_trace` or an equivalent
      operator-visible artifact.
- [ ] Corrected or superseded memory is suppressed from normal current recall.
- [ ] Degraded or stale recall is labeled honestly.
- [ ] Agent-private data requires owner matching.
- [ ] Group-shared data requires valid membership where applicable.
- [ ] Memory scope behavior was affected; explain below.
- [ ] Memory scope behavior was not affected.

## Database / Migration Notes

<!-- Include migration order, rollback notes, live gate result, and scratch DB setup if applicable. -->

- [ ] Migration impact exists.
- [ ] No migration impact.
- [ ] Live PostgreSQL testing is required.
- [ ] Live PostgreSQL testing is not required.
- [ ] Hermes/MCP testing is required.
- [ ] Hermes/MCP testing is not required.

## Source Review

- Estimated source:
- Suspected license:
- Similarity risk:
- Review required:
- Notes:

## Risks / Follow-Up

<!-- Known gaps, deferred checks, rollout cautions, or next slice. -->
