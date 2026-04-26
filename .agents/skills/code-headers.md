---
name: code-headers
description: Use this skill when creating, renaming, or materially editing Go source files so VibeGravity code headers stay parseable and useful.
---

# Code Headers

## Purpose

Keep source files self-orienting for humans and agents without turning headers
into long in-file histories.

## Required Inputs

- `docs/code-header-policy.md`
- `plans/templates/code-header-minimal-go.md`

## Steps

1. Choose the minimal structured header by default.
2. Use the narrative template only when architectural rationale is essential.
3. Use the development-log template only when a file is explicitly audit-heavy.
4. Fill `FILE`, `PURPOSE`, `LAYER`, `STATUS`, `EXPORTS`, `DEPENDS`, `USED_BY`,
   and `AGENT_NOTE`.
5. Keep `DEPENDS` and `USED_BY` short and action-oriented.
6. After edits or renames, run `make check-headers`.
7. If a header rule changes, update `docs/code-header-policy.md` first.

## Standard Gate

The canonical header validation command is:

```bash
make check-headers
```

Do not substitute `go test`, `make lint`, or manual inspection for this gate.

## Done When

- New or changed Go files have current headers.
- `make check-headers` passes.
- Any intentional exception is documented in the policy.
