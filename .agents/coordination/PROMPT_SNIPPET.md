# Multi-Agent Coordination Snippet

Paste this into every parallel Codex, Hermes, Claude, or review-agent prompt
that may edit this repo.

For a fully autonomous agent that should decide its own next useful task, use
`.agents/coordination/UNIVERSAL_AGENT_PROMPT.md` instead.

If the operator gives only that universal prompt path, read it and execute it.
Do not ask what to do with the file.

```text
Before editing any file, read:

- .agents/coordination/WORK_PROGRESS.md
- .agents/workflows/quickstart.md
- the relevant .agents/workflows/<role>.md file

Then claim the exact files you intend to edit:

.agents/coordination/agent-work.sh claim "<agent-id>" "<short task>" <file> [<file> ...]

Rules:

- Do not edit a file claimed by another active agent.
- Claim exact file paths before opening a new write surface.
- Do not use `--`, globs, directories, parent traversal, or whitespace paths as
  claim paths.
- Send a heartbeat before widening scope or after a long debugging pass.
- Only the leader can approve lane widening.
- Only the leader can approve final synthesis.
- Release files immediately when done with them, before moving to other files.
- Finish with `done` only after verification and result notes are complete.
- Every saved handoff must start with YAML front matter containing agent_id,
  role, phase_id, lane_id, lane_type, claimed_files, reviewed_files,
  changed_files, gates_run, gates_skipped, skip_reasons, and next_owner.
```
