# Work Pack 04: Hermes Provider and External Surfaces

## 1. Goal

이 work pack의 목표는 VibeGravity를 실제 첫 고객에게 연결하는 것이다.
그 첫 고객은 Hermes Agent다.

V1 제품 언어에서는 이 slice를 **Hermes Memory, powered by VibeGravity**로
본다. Hermes는 integration host이고, 실제 사용자는 Hermes operator다.
따라서 단순 lifecycle 연결만으로는 부족하다. Operator가 Hermes가 받을 recall을
미리 보고, memory의 이유를 확인하고, 잘못된 memory를 고칠 수 있어야 한다.

## 2. Deliverables

- Hermes memory provider plugin
- provider tools
- recall preview / rendered context inspection
- explain and timeline trust surfaces
- bootstrap and doctor flow
- MCP surface
- external integration docs

## 3. Hermes Scope

v1에서 Hermes는 아래 lifecycle만 잘 연결하면 된다.

- pre-turn `prefetch()`
- post-turn `sync_turn()`
- provider tools
- optional session end dreaming hint

## 4. Plugin Responsibilities

| hook | role |
|---|---|
| `is_available()` | config and health check |
| `prefetch()` | `/v1/prefetch` call |
| `sync_turn()` | `/v1/sync-turn` call |
| `render_context()` | JSON to compact text |
| `get_tools()` | provider tools |
| `on_session_end()` | dreaming hint |

## 5. Provider Tools

최소 아래 도구는 제공한다.

- search memory
- add note
- show plan
- correct memory
- view timeline
- explain memory
- recall preview
- degraded status

## 6. MCP Surface

MCP는 operator와 coding agent를 위한 도구 표면이다.
Hermes plugin과 같은 core를 호출해야 한다.

## 7. Future Adapters

Claude Code와 Codex client는 나중에 붙인다.
이 work pack에서 API meaning을 generic하게 유지한다.

## 8. Tests

- mocked API로 provider tests
- real session replay with Hermes
- provider tool roundtrip
- built-in memory coexistence

## 9. Done When

- Hermes가 external memory provider로 연결된다
- prefetch와 sync가 세션에서 실제 돈다
- operator가 Hermes가 받을 recall preview를 볼 수 있다
- operator가 memory를 explain하고 correction할 수 있다
- corrected memory가 다음 relevant recall에서 old memory를 억제한다
- scope와 degraded/freshness 상태가 operator-visible surface에 드러난다
- tool calls가 memory state를 바꾼다
- plugin failure가 Hermes 전체를 죽이지 않는다

## 10. Current Narrow Adapter

`internal/hermes.Provider` is the first in-repo adapter layer. It does not
install or modify a local Hermes configuration. It maps provider lifecycle names
to the shared `core.VibeGravityService`:

- `IsAvailable` checks a minimal `Prefetch`.
- `Prefetch` delegates to core recall.
- `SyncTurn` delegates to core ingest.
- `RenderContext` turns typed recall blocks into compact text.
- `GetTools` advertises `recall_preview`, `search_memory`, `add_note`,
  `show_plan`, `explain_memory`, `correct_memory`, `view_timeline`, and
  `degraded_status`.
- `CallTool` dispatches the core-backed provider tools through the same DTOs
  used by HTTP and MCP. `degraded_status` returns recall metadata from
  `Prefetch`; `show_plan` is explicitly blocked with `ErrNotImplemented` until
  a read-only plan-list API exists.
- `OnSessionEnd` is a no-op hook reserved for dreaming integration.

The remaining work is external packaging/bootstrap and real Hermes runtime
roundtrip testing.

## 11. Current MCP Surface

`internal/mcp.Surface` is the in-repo MCP-style adapter layer. It lists tool
names and decodes JSON tool inputs before delegating to the same
`core.VibeGravityService` used by HTTP and Hermes.

Current tools:

- `prefetch`
- `recall_preview`
- `sync_turn`
- `search_memory`
- `search_documents`
- `add_note`
- `create_plan`
- `update_plan`
- `correct_memory`
- `view_timeline`
- `explain_memory`

`internal/mcp.Server` wraps that surface in the real MCP JSON-RPC protocol over
stdio. The first supported protocol version is `2025-11-25`, with these
minimum methods:

- `initialize`
- `notifications/initialized`
- `ping`
- `tools/list`
- `tools/call`

`tools/list` advertises concrete JSON input schemas for the current tool set
rather than a generic object placeholder. The trust-loop tools make their
operator-critical inputs visible to MCP clients:

- `recall_preview`: tenant/workspace plus optional session, actor, query,
  budget, and mode.
- `correct_memory`: target memory, operator, correction text, idempotency key,
  evidence, and optional visibility inputs (`entity_id`, `visible_group_ids`)
  for private/group-shared authorization.
- `view_timeline`: tenant/workspace, actor entity, scopes, time bounds, and
  limit.
- `explain_memory`: tenant/workspace, memory id, optional actor entity, and
  visible group ids for private/group-shared provenance checks.

Before new Hermes-facing features are added, the next external-surface slice
should prove protocol correctness for the existing trust loop. Required fields
must stay aligned across MCP, Hermes provider tools, HTTP DTOs, and core
validation. In particular, `correct_memory`, `view_timeline`, and
`explain_memory` must not lose tenant/workspace/actor/evidence or visibility
inputs at the protocol boundary.

This is a DB/protocol correctness slice, not a feature expansion slice. The
external path is V1-ready only when Hermes-facing clients can preview recall,
correct a memory, inspect explain/timeline evidence, and observe supersession
and degraded freshness with the same semantics the PostgreSQL-backed service
enforces.

`cmd/cli` now exposes the server through:

```bash
go run ./cmd/cli mcp serve --stdio
```

The Hermes-facing bootstrap path is:

```bash
go run ./cmd/cli hermes bootstrap --command "$(pwd)/bin/cli"
```

That prints a registration command shaped like:

```bash
hermes mcp add vibegravity --command /path/to/cli --args mcp serve --stdio
hermes mcp test vibegravity
```

Current limitation: Hermes' `memory` CLI only exposes a fixed external-provider
registry in this environment, so this slice uses Hermes' supported MCP server
registration path for real external roundtrip. The in-repo
`internal/hermes.Provider` remains the semantic provider adapter and should be
kept aligned with MCP/HTTP core behavior until Hermes exposes a custom memory
provider packaging hook.

Documents and dreaming should remain available engine capabilities behind these
surfaces, but they are not the external V1 promise. The V1 promise is that
Hermes Memory can preview, explain, correct, supersede, and honestly label
freshness/degradation through the protocol path Hermes can actually use.
