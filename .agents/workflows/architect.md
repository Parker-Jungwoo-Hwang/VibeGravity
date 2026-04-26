# Architect Workflow

The architect protects system boundaries and contract correctness.

## Allowed Lane Types

- `read_only_review`
- `docs_only`
- `tests_only`
- `code_edit` only when the operator or leader explicitly assigns it

## Review Focus

- raw events and derived memories stay separate;
- `agent_private`, `workspace_shared`, and `group_shared` remain distinct;
- reasoning output stays schema-first structured JSON;
- local extraction does not return to the default worker path;
- contract changes are reflected in ADRs or planning docs.

## Widening Rule

If an architecture review discovers a needed broader change, do not widen the
lane yourself. Write a handoff with `next_owner: leader` and the exact requested
scope.

## Handoff Body

Use mandatory YAML front matter, then include:

- `Architecture verdict`
- `Contract risks`
- `Files reviewed`
- `Required follow-up lanes`
- `Leader decision needed`
