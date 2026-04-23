---
name: eval-regression
description: Use this skill to run or inspect golden scenarios and detect memory regressions.
---

# Eval Regression

## When to use

Use this after changing reasoning, recall, scopes, or profile logic.

## Required scenarios

- correction updates old fact (supersession works)
- workspace_shared vs agent_private separation (scope filter correct)
- group_shared visibility (membership check works)
- pinned note inclusion in recall
- active plan inclusion in recall
- superseded memory suppression (latest_flag logic)
- dreaming promotion (session → mid → long → ultra-long)
- degraded recall without Codex (existing profile + recent tail still returned)
- budget_tokens respected (output within budget)
- fingerprint dedup prevents duplicate memories

## Output

- scenario
- expected result
- observed result
- pass or fail
- suspected cause
- next fix
