# Code Header Policy

This policy captures the repo-local rule for VibeGravity source files.

## Default

Use the minimal structured header for Go files. It is compact, easy for agents
to parse, and enough to rebuild file responsibility and dependency maps from a
file list.

Use the narrative header only for modules where architectural rationale is more
important than the type list. Use the development-log header only for files that
are intentionally audited across repeated agent edits.

## Required Go Header

Every non-generated Go file must start with a header containing these fields:

- `FILE`
- `PURPOSE`
- `LAYER`
- `STATUS`
- `EXPORTS`
- `DEPENDS`
- `USED_BY`
- `AGENT_NOTE`

Run this before handing off a change:

```bash
make check-headers
```

## Layer Values

Use one of these values:

- `domain`: core business contracts and domain records
- `application`: service orchestration and use-case level code
- `interface`: HTTP, CLI, MCP, Hermes, or other external surface adapters
- `infra`: database, config, embedding clients, migrations, runtime plumbing
- `util`: reusable tools and low-level helpers
- `test`: tests and test fixtures

## Status Values

Use one of these values:

- `draft`: incomplete or newly introduced
- `active`: current default path
- `experimental`: intentionally unstable surface
- `deprecated`: kept only for compatibility or migration

## Field Guidance

`PURPOSE` should be one sentence explaining why the file exists.

`EXPORTS` should name public symbols when practical. Use a grouped phrase only
when a file exports many closely related DTOs or constants.

`DEPENDS` should list the most important local files or external packages that
an agent should inspect before editing. Keep it short.

`USED_BY` should list the main consumers. If the file is a leaf executable or
test, name the command or package-level purpose.

`AGENT_NOTE` should name the one rule most likely to prevent a bad edit.

## Rename Rule

When a file moves, update the `FILE`, `DEPENDS`, and `USED_BY` fields in the
same change. A rename is not complete until `make check-headers` passes.
