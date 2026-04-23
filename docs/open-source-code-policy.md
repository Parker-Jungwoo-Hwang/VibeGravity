# Open-Source Code Policy

VibeGravity is intended for open-source development. Code must be original to
this repo or derived only from commercially usable permissive patterns.

## Rules

1. Do not reference or closely reproduce code under GPL, AGPL, LGPL, SSPL,
   Elastic License, or related license families.
2. Use MIT, BSD, Apache-2.0, official documentation, or first-principles design
   as the acceptable reference boundary.
3. Do not copy an external project's function names, file structure, comments,
   or distinctive implementation shape.
4. If code may be substantially similar to external open-source code, stop and
   warn before implementing it.
5. Treat structured external snippets of 10 or more consecutive lines as risky
   and rewrite them from first principles.
6. For code-bearing handoffs, include a source review block with estimated
   source, suspected license, similarity risk, and review requirement.

## Source Review Template

```text
Source Review:
- Estimated source: first-principles VibeGravity plans / official docs / permissive pattern / unknown
- Suspected license: none / MIT / BSD / Apache-2.0 / unknown
- Similarity risk: low / medium / high
- Review required: no / yes
- Notes: short rationale
```

## Default For This Repo

Default to first-principles implementation from the VibeGravity plans. Use
external docs for API behavior and dependency usage only. When in doubt, choose
a simpler original implementation and request review.
