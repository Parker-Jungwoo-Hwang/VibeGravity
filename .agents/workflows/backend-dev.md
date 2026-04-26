# Backend Developer Workflow

Backend developers implement narrow Go/runtime/database slices.

## Allowed Lane Types

- `tests_only`
- `code_edit`
- `docs_only` when documenting behavior changed in the lane

## Required Loop

1. Read the active plan and runtime/data model contracts.
2. Claim exact files before editing.
3. Implement the smallest behavior change that satisfies the lane.
4. Add or update focused tests.
5. Run focused tests first, then appropriate repo gates.
6. Release claims before final handoff.

## Stop Lines

- Do not reintroduce local extractor dependence.
- Do not blur memory scopes.
- Do not change schema or architecture direction without docs or ADR coverage.
- Do not widen into another hot file without leader approval.

## Handoff Body

Use mandatory YAML front matter, then include:

- `Summary`
- `Behavior changed`
- `Tests`
- `Gates skipped and why`
- `Remaining risks`
- `Source Review`
