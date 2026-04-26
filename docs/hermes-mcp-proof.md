# Hermes MCP Proof

Status: not proven.

`vibegravity hermes bootstrap` is not proof by itself. It only prints the Hermes
MCP registration command.

## Commands To Run

```bash
bin/vibegravity mcp serve --stdio
bin/vibegravity hermes bootstrap --name vibegravity --command "$(pwd)/bin/vibegravity"
hermes mcp add vibegravity --command "$(pwd)/bin/vibegravity" --args mcp serve --stdio
hermes mcp test vibegravity
```

## Tools To Prove Through Real MCP

- `recall_preview`
- `correct_memory`
- `explain_memory`
- `view_timeline`
- `degraded_status`

## Evidence Log

- Environment: local repo build, no reachable PostgreSQL at the default
  `postgres://localhost:5432/vibegravity?sslmode=disable`.
- Command run:

  ```bash
  printf '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"smoke","version":"0"}}}\n' | bin/vibegravity mcp serve --stdio
  ```

- Output:

  ```text
  ERROR: open VibeGravity service: failed to ping database: failed to connect to `user=parker database=vibegravity`:
  	127.0.0.1:5432 (localhost): dial error: dial tcp 127.0.0.1:5432: connect: connection refused
  	[::1]:5432 (localhost): dial error: dial tcp [::1]:5432: connect: connection refused
  ```

- Passed: none for real MCP proof.
- Failed: MCP service startup because PostgreSQL was unreachable.
- Remaining unproven: real Hermes invoking the trust-loop tools through MCP.

## Rollback

```bash
hermes mcp remove vibegravity
```
