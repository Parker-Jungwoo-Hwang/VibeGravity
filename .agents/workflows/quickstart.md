# Quickstart Workflow

Quickstart is read-only.

Use this file when an agent needs orientation before choosing or accepting a
lane. Quickstart must not edit repo files, claim files, run formatters, or
change generated state.

## Read-Only Startup

1. Read `AGENTS.md`.
2. Read `.agents/coordination/WORK_PROGRESS.md`.
3. Run `.agents/coordination/agent-work.sh status`.
4. Read this file.
5. Read `.agents/workflows/phase_context.md`.
6. Read the role file if the operator already assigned a role.

## Output

Quickstart output should be a short recommendation:

- suggested `phase_id`;
- suggested `lane_id`;
- suggested `lane_type`;
- exact files likely to be claimed later;
- blockers or active claims observed;
- whether leader approval is required.

Do not claim files from quickstart. The implementation or docs lane claims files
only after the leader or operator selects that lane.
