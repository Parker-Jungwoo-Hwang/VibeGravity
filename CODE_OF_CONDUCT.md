# Code of Conduct

VibeGravity is a pre-release project intended for open-source release. It
handles sensitive ideas: private workspace context, shared memory boundaries,
correction history, provenance, and operator trust. Contributors are expected to
work with care toward both people and data.

## Expected Behavior

- Be respectful, direct, and constructive.
- Assume good intent, while still reviewing code and claims rigorously.
- Keep discussions focused on the work, evidence, and user impact.
- Protect private memory contents, secrets, credentials, and user data.
- Redact sensitive identifiers and examples when reporting issues.
- Make room for questions from new contributors.
- Accept maintainer decisions about scope, safety, and release readiness.

## Unacceptable Behavior

- Harassment, threats, insults, or personal attacks.
- Discriminatory language or conduct.
- Publishing another person's private information without permission.
- Posting secrets, private workspace data, raw memory contents, or production
  database dumps in public channels.
- Deliberately bypassing scope, privacy, or security boundaries.
- Misrepresenting test results, readiness, provenance, or security impact.
- Repeatedly derailing technical discussions after maintainers ask to refocus.

## Memory-Specific Care

Because VibeGravity is a memory system, contributors should treat privacy and
scope bugs as serious product issues. In particular:

- `agent_private` data must not leak into shared recall.
- `group_shared` data must respect membership boundaries.
- correction, timeline, and explain views must not expose memory outside the
  requester's allowed scope.
- reports should use minimal reproductions and redact sensitive content.

## Enforcement

Maintainers may edit, hide, or remove comments, issues, pull requests, or other
contributions that violate this code of conduct. Maintainers may also limit or
remove participation for repeated or severe violations.

Security-sensitive issues should be reported through `SECURITY.md`, not public
discussion. General support requests should follow `SUPPORT.md`.

## Scope

This code of conduct applies to project spaces, including the repository,
issues, pull requests, discussions, review channels, and project-related
community spaces when they exist.

## Source Review

Estimated source: original first-principles policy written for VibeGravity's
memory, privacy, and contributor-readiness needs.

Suspected license: none.

Similarity risk: low.

Review required: no for repository use; yes before presenting it as a formal
foundation or company policy.
