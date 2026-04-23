---
name: source-provenance
description: Use this skill before adding code influenced by external material so VibeGravity stays safe for open-source release.
---

# Source Provenance

## Purpose

Keep VibeGravity's implementation original, reviewable, and safe to publish as
open source.

## Allowed Reference Classes

- First-principles implementation from VibeGravity's own plans and contracts.
- Official language, standard library, or dependency documentation.
- Commercially usable permissive patterns from MIT, BSD, or Apache-2.0 sources.

## Blocked Reference Classes

- GPL, AGPL, LGPL, SSPL, Elastic License, or related copyleft/source-available
  license families.
- Verbatim code, comments, function names, file layouts, or distinctive
  implementation structure from a specific external project.
- Structured external snippets of 10 or more consecutive lines.

## Steps

1. Identify whether the implementation is first-principles or externally
   influenced.
2. If external influence exists, record the estimated source class and suspected
   license before coding.
3. Rewrite any structured snippet of 10 or more consecutive lines from first
   principles.
4. Change names, boundaries, and file placement to match VibeGravity's own
   architecture, not the referenced project.
5. Stop and warn if substantial similarity risk remains.
6. Include the source review block in the handoff.

## Source Review Block

```text
Source Review:
- Estimated source: first-principles VibeGravity plans / official docs / permissive pattern / unknown
- Suspected license: none / MIT / BSD / Apache-2.0 / unknown
- Similarity risk: low / medium / high
- Review required: no / yes
- Notes: short rationale
```

## Done When

- The code is newly implemented for VibeGravity.
- No blocked license family was used as a coding reference.
- The handoff includes the source review block for any code-bearing change.
