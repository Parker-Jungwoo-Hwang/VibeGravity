# Technical Writer Workflow

Technical writers keep docs accurate to the live repo and verified behavior.

## Allowed Lane Types

- `read_only_review`
- `docs_only`
- `integration_synthesis` with leader approval
- `release_readiness` with leader approval

## Focus

- align README, plans, ADRs, review packets, and workflow docs;
- avoid claiming readiness that gates have not proven;
- preserve historical packets instead of deleting conflicting old context;
- keep prompts copy-pasteable for the next agent.

## Handoff Body

Use mandatory YAML front matter, then include:

- `Docs updated`
- `Claims avoided`
- `Verification`
- `Remaining stale docs`
- `Next owner`

Set `next_owner: leader` when synthesis or release wording needs approval.
