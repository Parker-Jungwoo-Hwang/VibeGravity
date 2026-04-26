# Release Notes Template

Copy this template into `dist/release-note-v0.x.y.md` before creating an
annotated tag.

```markdown
# VibeGravity v0.x.y

Release type: private validation drop | public pre-release | public stable
Commit: `<sha>`
Date: YYYY-MM-DD
Owner: `<name>`

## Summary

One paragraph describing what changed and why it matters for Hermes Memory,
powered by VibeGravity.

## Added

- None.

## Changed

- None.

## Fixed

- None.

## Security

- None.

## Migration Notes

- Migration range:
- `VIBEGRAVITY_MIGRATION_PATH`:
- Live PostgreSQL gate:
- Rollback status:
- Backup or snapshot:

## Known Risks

- None.

## Required Gates

- `git status --short --branch`:
- `make release-gate`:
- `make integration-postgres`:
- Hermes/MCP smoke:
- `make release-checksums`:
- `make sbom` if claimed:

## Artifacts

- Binary:
- Migrations:
- Checksums:
- SBOM:

## Rollback

- Previous version:
- Binary rollback:
- Migration rollback:
- Hermes/MCP rollback:
- Stop-line:

## Source Review

Estimated source:
Suspected license:
Similarity risk:
Review required:
```

## Tag Command

```bash
git tag -a v0.x.y -F dist/release-note-v0.x.y.md
git push origin v0.x.y
```

## Source Review

Estimated source: first-principles release note template from VibeGravity's
current changelog, gate, migration, and rollback requirements.

Suspected license: none.

Similarity risk: low.

Review required: yes before first public tag.
