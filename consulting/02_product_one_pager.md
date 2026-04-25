# VibeGravity Product One-Pager

## Product Definition

VibeGravity is a shared memory engine for AI agents.

It lets an agent host record what happened, derive structured memories, and retrieve only the useful context before future work.

## Problem

AI agents lose continuity across sessions.

The user repeats the same rules, preferences, project decisions, and current task state. Multiple agents working in the same workspace do not reliably share what should be shared, and private agent memory can easily blur with workspace memory if the system is not designed carefully.

Current memory approaches often fall into weak patterns:

- raw chat logs become the memory;
- vector search returns noisy context;
- private and shared memory are not strongly separated;
- corrections do not reliably change future behavior;
- users cannot inspect why a memory exists;
- memory gets longer instead of more useful.

## Target User

The first customer is Hermes Agent.

The first human user is the person operating Hermes and expecting continuity across sessions, workspaces, plans, notes, and corrections.

Later users may include coding agents, operators, or other local agent runtimes that connect through HTTP or MCP.

## Solution

VibeGravity sits behind the agent host as a memory engine.

It provides:

- `sync_turn()` after an agent turn to record raw events;
- worker processing to derive structured memory;
- graph apply to create, extend, update, and supersede memories;
- `prefetch()` before the next turn to return a compact recall pack;
- manual tools for search, notes, plans, correction, timeline, and provenance explanation.

## Product Promise

The user should not need to repeat important context.

The agent should remember durable facts, active plans, preferences, decisions, and corrections across sessions while keeping private, workspace, group, and session memory boundaries explicit.

## Why It Matters

If agents become daily collaborators, memory quality becomes product quality.

Bad memory is worse than no memory. It can leak private context, revive outdated facts, ignore user correction, or flood the prompt with noise.

VibeGravity's product bet is that memory needs an engine, not just a database.

## V1 Success Statement

V1 succeeds if Hermes can:

- send turns to VibeGravity without slowing the hot path;
- receive useful recall before the next turn;
- keep `agent_private`, `workspace_shared`, and `group_shared` memory separate;
- show why a memory exists;
- accept human correction;
- suppress superseded memory;
- keep working in degraded mode when Codex reasoning is unavailable;
- prove behavior through replay and golden scenarios.

## What VibeGravity Is Not

VibeGravity is not:

- a chat application;
- a standalone AI assistant;
- a general workflow automation system;
- a raw transcript archive;
- a generic vector database;
- a heavyweight knowledge graph platform.

## Product Tagline Candidates

- Memory engine for AI agents.
- Shared memory kernel for Hermes and agent teams.
- Scoped, correctable memory for long-running agents.
- The memory layer agents can trust.
