# Customer and Use Cases

## First Customer

The first customer is Hermes Agent.

That means V1 should optimize for Hermes lifecycle integration:

- before turn: call `prefetch()`;
- after turn: call `sync_turn()`;
- during operation: expose search, note, plan, correction, timeline, and explain tools;
- after session: trigger dreaming or consolidation hints.

## First Human User

The first human user is the operator or builder who uses Hermes for ongoing work.

This person cares less about the internal memory graph and more about:

- not repeating durable context;
- keeping current plans alive;
- correcting wrong memory;
- seeing why a memory exists;
- avoiding private/shared leakage;
- trusting that memory will not silently rot.

## Primary Jobs To Be Done

### 1. Continue Work Without Repeating Context

When I start a new Hermes session, I want Hermes to remember the relevant project rules, current plan, and recent decisions so I can continue work without re-explaining everything.

### 2. Keep Shared and Private Memory Separate

When multiple agents work in one workspace, I want workspace memory to be shared while private agent memory stays private, so collaboration does not become leakage.

### 3. Correct Memory When It Is Wrong

When the system remembers something outdated or wrong, I want to correct it once and have future recall reflect the correction.

### 4. Inspect Why Memory Exists

When memory affects agent behavior, I want to see where it came from and what reasoning or event produced it.

### 5. Keep Useful Context Short

When Hermes asks for recall, I want the engine to return compact, useful blocks instead of dumping a long transcript.

## Initial Use Cases

### Hermes Continuity

Hermes uses VibeGravity to remember project preferences, current task status, user constraints, and corrections across sessions.

Success looks like:

- Hermes resumes the correct work;
- active plan and pinned notes appear in recall;
- stale superseded memory is suppressed;
- user correction changes later behavior.

### Agent Collaboration in One Workspace

Multiple agents work on the same project while VibeGravity separates private and workspace memory.

Success looks like:

- workspace decisions are shared;
- `agent_private` memory only returns to the owning agent;
- group memory only appears for members;
- provenance remains inspectable.

### Operator Memory Control

The user uses tools to search, add notes, create/update plans, correct memory, and inspect timeline.

Success looks like:

- user can repair the memory system;
- operator actions are visible;
- correction is not a hidden prompt trick;
- the memory engine becomes more trustworthy over time.

## Later Use Cases

These should not drive V1 unless the Product Owner strongly disagrees:

- generic memory backend for many unrelated agent runtimes;
- standalone web UI;
- marketplace of memory integrations;
- advanced organization-wide knowledge graph;
- fully autonomous forgetting without operator visibility.

## Customer Question For Consulting

Is Hermes-first enough as the first customer, or should V1 explicitly define a human operator persona with a visible control surface?
