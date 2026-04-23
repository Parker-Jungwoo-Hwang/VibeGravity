# Work Pack 03: Memory Graph and Dreaming

## 1. Goal

이 work pack의 목표는 VibeGravity를 진짜 memory system으로 만드는 것이다.  
여기서 graph, profile, dreaming이 살아난다.

## 2. Deliverables

- Codex stage 1 extract
- Codex stage 2 resolve
- apply engine
- memory graph
- profile snapshots
- session summaries
- dreaming jobs
- correction path
- supersede / archive logic

## 3. Tasks

### Task A. Codex extract stage

현재 event에서 candidate memory와 entity를 뽑는다.

### Task B. Codex resolve stage

기존 graph, profile, note, plan, doc와 비교해 operations를 만든다.

### Task C. Apply engine

schema validation과 semantic validation 뒤에 commit한다.

### Task D. Profile builder

static과 dynamic을 분리한 snapshot을 만든다.

### Task E. Dreaming

- dream_session
- dream_workspace
- memory promotion
- stale dynamic suppression

### Task F. Correction

human correction을 strong hint로 반영한다.

## 4. Required Contracts

### Output schema

반드시 structured JSON이다.

```json
{
  "operations": [],
  "profile_delta": {},
  "session_summary": "",
  "plan_delta": {},
  "trace": {}
}
```

### Apply rules

- `updates` lowers previous latest
- `extends` keeps prior memory alive
- `contradicts` weakens recall priority
- `hypothesis` is conservative in profile merge

## 5. Dreaming Rules

- short-term to mid-term on session boundary
- repeated useful dynamic facts may become long-term
- canonical profile only from repeatedly supported memories
- noisy or contradicted memories decay or archive

## 6. Tests

- correction creates `updates`
- additive detail creates `extends`
- profile.static stays stable
- profile.dynamic changes with recency
- dreaming does not duplicate memory
- repeated sessions promote useful facts

## 7. Done When

- graph edges exist and are queryable
- profile snapshot rebuild works
- dreaming jobs run without hot path impact
- correction changes later recall
