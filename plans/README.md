# VibeGravity AI Agent Build Pack

이 문서 세트는 사람에게 설명하기 위한 문서가 아니다.  
VibeGravity를 실제로 만들 AI agent가 방향을 잃지 않고 구현하게 만드는 작업 문서다.

이 세트의 목표는 세 가지다.

첫째, 제품의 북극성을 고정한다.  
둘째, 구현 순서를 고정한다.  
셋째, Codex와 Claude Code 같은 coding agent가 좋은 결과를 내는 운영 방식까지 같이 고정한다.

## 먼저 읽을 문서

가장 먼저 `00_read-this-first_for-building-agents.md`를 읽는다.  
그다음 `01_rfp_vibegravity_hermes-first.md`를 읽는다.  
그다음 `02_product-contract_and_direction.md`와 `03_target-architecture_codex-first.md`를 읽는다.  
구현에 들어갈 때는 `05_runtime-contracts_ingest-recall-apply.md`와 각 workpack 문서를 본다.  
agent 운영 방식은 `12_agent-coding_playbook_codex-claude.md`를 본다.

## 문서 목록

| 파일 | 역할 |
|---|---|
| `00_read-this-first_for-building-agents.md` | 한 장 브리프 |
| `01_rfp_vibegravity_hermes-first.md` | RFP. 요구사항과 납품 기준 |
| `02_product-contract_and_direction.md` | 제품 방향과 절대 규칙 |
| `03_target-architecture_codex-first.md` | 목표 구조 |
| `04_memory-scopes_dreaming_ontology-lite.md` | 메모리 범위, dreaming, 구조화 규칙 |
| `05_runtime-contracts_ingest-recall-apply.md` | API, worker, apply 계약 |
| `06_data-model_and_storage-invariants.md` | 저장 구조와 불변 조건 |
| `07_workpack_foundation-and-repo-setup.md` | 기반 작업 팩 |
| `08_workpack_ingest-and-recall.md` | ingest와 recall 작업 팩 |
| `09_workpack_memory-graph-and-dreaming.md` | graph와 dreaming 작업 팩 |
| `10_workpack_hermes-provider-and-external-surfaces.md` | Hermes 중심 외부 연결 작업 팩 |
| `11_workpack_quality-ops-and-evals.md` | 테스트, 운영, 평가 작업 팩 |
| `12_agent-coding_playbook_codex-claude.md` | Codex와 Claude Code 운영 방식 |
| `13_handoff-prompts_and_response-templates.md` | 시작 프롬프트와 응답 틀 |
| `14_source-notes.md` | 배경 자료와 참고 기준 |
| `templates/AGENTS.md` | Codex용 기본 지시 파일 |
| `templates/CLAUDE.md` | Claude Code용 기본 지시 파일 |
| `templates/PLANS.md` | 실행 계획 템플릿 |
| `templates/SKILL_plan-implement-verify.md` | 작업용 스킬 템플릿 |
| `templates/SKILL_contract-check.md` | 계약 점검 스킬 템플릿 |
| `templates/SKILL_eval-regression.md` | 회귀 평가 스킬 템플릿 |
| `templates/RFP_RESPONSE_TEMPLATE.md` | AI agent의 제안 응답 템플릿 |

## 이 세트의 핵심 방향

VibeGravity는 shared memory kernel이다.  
Hermes와 다른 agent가 공통으로 붙는 기억 계층이다.

VibeGravity는 모델이 아니다.  
답을 직접 만드는 agent runtime도 아니다.  
기억을 저장하고, 정리하고, 다시 짧게 꺼내 주는 백엔드다.

이 문서 세트는 다음 방향을 고정한다.

- Hermes를 첫 고객으로 잡는다.
- local LLM은 임베딩과 검색 보조에만 쓴다.
- 텍스트 해석과 graph 정리는 Codex-first로 간다.
- agent memory, workspace memory, group shared memory를 분리한다.
- dreaming은 maintenance layer로 넣는다.
- recall은 자동 주입과 수동 주입을 함께 지원한다.
- memory는 ontology-lite 방식으로 구조화한다.
- token 최적화는 제품 핵심 기능으로 다룬다.

## 이 문서 세트 사용 원칙

문서에 맞게 코드를 만든다.  
코드에 맞춰 문서를 끌려가게 만들지 않는다.  
충돌이 생기면 ADR을 쓰고 문서를 갱신한다.  
모호하면 먼저 계약 문서를 본다.  
그래도 모호하면 RFP의 목표와 non-goal을 다시 본다.
