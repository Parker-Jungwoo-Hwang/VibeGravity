# Work Pack 04: Hermes Provider and External Surfaces

## 1. Goal

이 work pack의 목표는 VibeGravity를 실제 첫 고객에게 연결하는 것이다.  
그 첫 고객은 Hermes Agent다.

## 2. Deliverables

- Hermes memory provider plugin
- provider tools
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
- tool calls가 memory state를 바꾼다
- plugin failure가 Hermes 전체를 죽이지 않는다
