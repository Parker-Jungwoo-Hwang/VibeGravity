# Commit Message Rules

This repo uses an Airbnb-inspired commit style based on the public commit
history in `airbnb/javascript`: concise imperative subjects, optional
bracketed scopes, and a short body when context matters.

## Subject Line Format

Use this shape:

```text
[scope] [subscope] Imperative summary
```

Examples:

```text
[docs] Add commit message rules for VibeGravity
[plans] Clarify external surface workpacks
[api] [recall] Return normalized memory filters
```

## Rules

1. Start with one or two bracketed scopes when they add signal.
2. Write the subject in imperative mood: `Add`, `Fix`, `Refactor`, `Remove`,
   `Update`, `Document`, `Test`.
3. Keep the full subject line concise. Target 72 characters.
4. Do not end the subject with a period.
5. Make one commit represent one logical change.
6. Add a blank line before the body.
7. Use the body to explain why the change exists and what matters for review.
8. Wrap body lines at about 72 characters when practical.
9. Add references only when they help future readers.

## Recommended Body Sections

When the change is non-trivial, use short labeled sections:

```text
Why:
- reason this change was necessary

What:
- main files or behaviors changed

Notes:
- follow-up work, caveats, or migration details

Refs:
- issue, PR, or design doc
```

## Suggested Scopes

- `[docs]`
- `[plans]`
- `[repo]`
- `[api]`
- `[server]`
- `[worker]`
- `[mcp]`
- `[tests]`
- `[deps]`
- `[release]`
- `[infra]`
- `[security]`

## Quick Examples

```text
[repo] Add Airbnb-style git commit template

Why:
- establish a consistent commit format before the first project history lands

What:
- add a reusable git commit template
- document repo commit rules
- ignore local OMX and Finder noise
```

```text
[api] [ingest] Reject empty normalized statements

Why:
- prevent low-signal rows from entering shared memory storage

What:
- validate normalized statements before persistence
- add regression coverage for blank input payloads
```
