# Leader Workflow

The leader owns coordination decisions, lane widening approval, and final
synthesis approval.

## Responsibilities

- Choose or confirm the phase and lane split.
- Keep lanes collision-aware and file-scoped.
- Approve or reject lane widening requests.
- Approve or reject final synthesis.
- Ensure multi-Hermes runs leave a synthesis packet.
- Convert completed lane handoffs into the next safe action.

## Exclusive Authority

Only the leader may approve:

- widening a lane beyond its initial `lane_id`, `lane_type`, or claimed files;
- merging multiple lane outputs into final synthesis;
- declaring a release-readiness verdict.

If a non-leader needs broader scope, the leader must receive a handoff whose
`next_owner` is `leader` and whose body states:

- current lane boundary;
- requested wider boundary;
- reason widening is necessary;
- collision risk;
- exact files expected after widening.

## Leader Handoff Body

After reviewing lanes, write a body with:

- `Decision`
- `Approved lane changes`
- `Rejected lane changes`
- `Final synthesis status`
- `Next owner`
- `Risks`

The YAML front matter from `.agents/workflows/README.md` is mandatory.
