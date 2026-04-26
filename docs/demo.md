# Demo

The local demo is the first confidence path for Hermes Memory.

Run:

```bash
go run ./cmd/vibegravity eval demo
```

Or after building:

```bash
bin/vibegravity demo
```

The demo needs no database, no Hermes runtime, no Codex, and no network.

## Story

A project rule exists. An active plan exists. Hermes receives compact recall. A
wrong memory is visible. The operator asks why it was remembered. The operator
corrects the memory. Later recall includes the correction. The old memory is
suppressed. Private memory does not leak. Degraded or stale state is labeled
honestly when the runtime can detect it.

## Before / Correction / After

Before correction:

Hermes recalls the wrong V1 headline.

Correction:

Actually, V1 is Hermes Memory: recall, explain, correction, and trust metadata.

After correction:

Hermes recalls that V1 is Hermes Memory, powered by VibeGravity. Documents are
supporting context, not the headline.

This is the activation moment. The user should understand the value within five
minutes: Hermes can remember project context, show why it remembered, and stop
using the old memory after correction.

## Expected Output

See `examples/hermes-memory-trust-loop/expected-output.txt`.
