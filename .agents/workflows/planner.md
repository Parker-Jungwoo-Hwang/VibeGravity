# Planner Workflow

The planner turns product and review context into small, claimable lanes.

## Allowed Lane Types

- `read_only_review`
- `docs_only`
- `integration_synthesis` with leader approval

The planner does not edit production code.

## Required Checks

- Read current plans and review packets before proposing lanes.
- Compare proposed lanes against `.agents/coordination/WORK_PROGRESS.md`.
- Keep each lane's write surface disjoint.
- Mark every lane with a lane type from `.agents/workflows/README.md`.

## Handoff Body

Use mandatory YAML front matter, then include:

- `Recommended lanes`
- `Claim boundaries`
- `Dependencies`
- `Leader approvals needed`
- `Suggested next owner`

Set `next_owner: leader` when lane widening, integration synthesis, or release
readiness needs approval.
