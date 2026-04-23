# Agent Coding Playbook for Codex and Claude Code

## 1. Purpose

이 문서는 VibeGravity를 구현할 coding agent가 좋은 결과를 내게 만드는 운영 문서다.  
핵심은 복잡한 문제를 잘게 나누고, durable facts와 reusable procedures를 분리하고, 항상 test와 review까지 닫는 것이다.

## 2. Core Rule

좋은 agent coding은 긴 프롬프트 하나로 끝나지 않는다.  
작은 계약, 짧은 계획, 검증 루프, 재사용 가능한 instruction 파일로 만든다.

## 3. Universal Workflow

모든 작업은 아래 순서를 따른다.

1. understand
2. plan
3. implement
4. verify
5. review
6. handoff

## 4. Prompt Shape

작업 프롬프트는 아래 네 칸을 반드시 가져야 한다.

- Goal
- Context
- Constraints
- Done when

이 네 칸이 비어 있으면 agent는 흔들리기 쉽다.

## 5. What Goes Into AGENTS.md and CLAUDE.md

두 파일에는 session마다 항상 필요한 사실만 넣는다.

예:

- repo layout
- build and test commands
- architecture rules
- do-not rules
- done definition
- review expectations

다단계 절차나 긴 playbook은 넣지 않는다.  
그건 skill로 뺀다.

## 6. What Goes Into Skills

반복되는 절차를 skill로 만든다.

예:

- plan -> implement -> verify
- contract check
- replay and eval
- Hermes integration review
- schema migration review

skill은 길어도 된다.  
필요할 때만 로드되기 때문이다.

## 7. When To Use Subagents

subagent는 bounded work에만 쓴다.

좋은 용도:

- code exploration
- test writing
- diff review
- migration risk scan
- eval result summary

좋지 않은 용도:

- architecture ownership을 완전히 넘기는 것
- main thread 없이 모두 병렬로 흩는 것
- 같은 파일을 여러 subagent가 동시에 수정하는 것

## 8. MCP Strategy

도구는 적게 붙인다.  
실제 workflow를 줄여 주는 것만 붙인다.  
처음부터 모든 MCP를 붙이지 않는다.

VibeGravity 구현에서 기본 MCP 후보는 아래 정도다.

- filesystem or local shell
- git
- optional issue tracker
- VibeGravity MCP for self-dogfooding after baseline

## 9. Codex-Specific Guidance

Codex는 plan-first가 잘 맞는다.  
복잡한 작업은 먼저 계획을 세우게 한다.  
repo-level `AGENTS.md`를 두고, repeated workflow는 skill로 만든다.  
main thread는 한 coherent unit of work를 유지하고, bounded work만 subagent로 보낸다.  
tests, lint, type check, review를 done criteria에 넣는다.

## 10. Claude Code-Specific Guidance

Claude Code는 `CLAUDE.md`와 auto memory를 같이 쓴다.  
하지만 durable project truth는 `CLAUDE.md`에 적는다.  
반복 절차는 skill로 뺀다.  
subagent는 read-only reviewer나 research assistant 같은 bounded role에 잘 맞는다.

## 11. Shared Rule for Both Tools

같은 mistake가 두 번 나오면 instruction file을 업데이트한다.  
chat에서만 고치고 끝내지 않는다.

## 12. VibeGravity-Specific Working Pattern

VibeGravity 구현은 아래 단위로 끊는 것이 좋다.

### A. Contract-first

API schema, DTO, storage invariants부터 고정한다.

### B. Hot path first

`sync_turn()`과 `prefetch()`를 먼저 만든다.

### C. Background second

worker, Codex chain, apply engine을 붙인다.

### D. Integration third

Hermes plugin을 붙인다.

### E. Eval last but mandatory

golden set과 replay harness로 닫는다.

## 13. Required Agent Output After Each Task

각 작업이 끝나면 agent는 아래를 보고해야 한다.

- changed files
- what was implemented
- commands run
- test results
- risks or follow-ups
- docs updated

## 14. Anti-Patterns

- giant single prompt without contracts
- architecture change without ADR
- skipping tests because code “looks right”
- durable rules staying only in chat
- loading too much context into main instructions
- letting multiple agents edit same hot files without coordination

## 15. Suggested Working Loop

짧은 작업은 한 thread에서 끝낸다.  
큰 작업은 work pack 단위로 나눈다.  
각 work pack마다 plan file을 만든다.  
작업이 끝나면 review skill을 돌린다.  
회귀가 있으면 eval skill을 돌린다.

## 16. Minimum Instruction Stack

VibeGravity repo에는 최소 아래 instruction stack이 있어야 한다.

- root `AGENTS.md`
- root `CLAUDE.md`
- `PLANS.md`
- skill for plan / implement / verify
- skill for contract check
- skill for eval regression
- skill for code headers

## 17. Source File Headers

Go source files use a parseable code header so agents can reconstruct file
purpose, layer, dependencies, and consumers without opening every file.

Default to `plans/templates/code-header-minimal-go.md`.
Use `.agents/skills/code-headers.md` whenever a Go file is created, renamed, or
materially changed.
Run `make check-headers` before handoff.

## 18. Why This Works

이 방식이 좋은 이유는 context를 나누기 때문이다.  
항상 필요한 사실은 instruction file로 남기고, 절차는 skill로 늦게 로드하고, bounded work는 subagent로 분리한다.  
그래서 token 낭비가 줄고, 결과가 더 일관되다.


## 19. Research Notes Behind This Playbook

이 문서의 운영 규칙은 감으로 만든 것이 아니다.  
공식 Codex 문서의 핵심은 plan-first, `AGENTS.md`, skill, bounded subagent, testing and review다.  
공식 Claude Code 문서의 핵심은 `CLAUDE.md`에는 durable facts를 두고, 절차는 skill로 빼고, subagent는 bounded role에 쓰는 것이다.  
그래서 VibeGravity 문서도 같은 구조로 맞췄다.
