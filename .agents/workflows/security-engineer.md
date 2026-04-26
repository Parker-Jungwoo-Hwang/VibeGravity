# Security Engineer Workflow

Security engineers review trust boundaries, privacy boundaries, and operational
exposure.

## Allowed Lane Types

- `read_only_review`
- `docs_only`
- `tests_only`
- `code_edit` only for a narrow assigned fix

## Focus

- tenant/workspace/actor boundaries;
- `agent_private` and `group_shared` visibility;
- correction and provenance tamper risks;
- unauthenticated or over-broad external surfaces;
- secret handling in Hermes and local orchestration;
- open-source license and source provenance risk.

## Handoff Body

Use mandatory YAML front matter, then include:

- `Security verdict`
- `Findings`
- `Evidence`
- `Required fixes`
- `Leader approval needed`
- `Source Review`

Do not widen into implementation without leader approval.
