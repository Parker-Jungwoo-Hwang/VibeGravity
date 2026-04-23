# Source Notes

이 문서 세트는 아래 자료를 바탕으로 다시 정리했다.

## Internal direction sources

- `What Is VibeGravity?`
- 기존 VibeGravity 개발 문서 세트
- 방향 재정렬 문서 세트
- Supermemory 설명 문서
- Honcho 설명 문서

## Official agent-coding sources

- OpenAI Codex best practices
- OpenAI Codex AGENTS.md guide
- OpenAI Codex skills guide
- Anthropic Claude Code memory guide
- Anthropic Claude Code skills guide
- Anthropic Claude Code subagent guide

## What was adopted

### From VibeGravity documents

- `prefetch()` / `sync_turn()` 중심 구조
- raw and derived separation
- Go HTTP server + worker + Postgres 구조
- shared memory kernel 정의

### From Honcho

- reasoning-first memory 철학
- compact context retrieval
- profile and representation 중심 접근

### From Supermemory

- documents vs memories 분리
- graph edges and versioning
- static vs dynamic profile 구분

### From Codex and Claude Code docs

- short durable instruction files
- procedure-heavy content moved to skills
- plan-first on hard tasks
- test and review in done criteria
- bounded work delegated to subagents
- minimal tool surface first

## What was intentionally changed

기존 일부 문서에는 local extractor가 중심에 있었다.  
이번 세트는 그것을 local embedding 중심으로 바꿨다.  
또한 Hermes-first 방향과 memory scope separation을 더 강하게 끌어올렸다.


## External reference pages

- OpenAI Codex best practices
- OpenAI Codex custom instructions with AGENTS.md
- OpenAI Codex agent skills
- Anthropic Claude Code memory
- Anthropic Claude Code skills
- Anthropic Claude Code subagents

이 문서 세트는 위 페이지의 운영 원칙만 흡수했다.  
제품 계약의 최종 기준은 이 문서 세트 자체다.
