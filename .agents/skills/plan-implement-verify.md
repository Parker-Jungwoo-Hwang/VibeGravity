---
name: plan-implement-verify
description: Use this skill for feature work that needs a short plan, code changes, checks, and self-review.
---

# Plan Implement Verify

## When to use

Use this skill when the task changes code or contracts.  
Do not use it for trivial one-line edits.

## Steps

1. Read the relevant contract docs from `plans/`.
2. Write a short plan in PLANS.md or as a comment.
3. Implement the smallest coherent slice.
4. Run the relevant checks:
   - `go test ./...`
   - `golangci-lint run`
5. Review the diff against the contract (load `contract-check` skill if needed).
6. Report changed files, commands, results, and risks.

## Output format

- plan
- implementation summary
- files changed
- checks run
- results
- risks
- docs updated
