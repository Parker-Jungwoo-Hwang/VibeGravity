# Work Pack 02: Ingest and Recall

## 1. Goal

이 work pack의 목표는 제품의 입구와 출구를 먼저 안정화하는 것이다.

입구는 `sync_turn()`이다.  
출구는 `prefetch()`다.

## 2. Deliverables

- `sync_turn()` 구현
- `prefetch()` 구현
- idempotent event ingest
- recall typed blocks
- notes and plans minimal recall inclusion
- lexical + embedding candidate retrieval
- degraded mode

## 3. Tasks

### Task A. `sync_turn()` write path

- normalize input
- compute idempotency
- insert raw events
- enqueue `process_turn_event`
- ack response

### Task B. `prefetch()` assembler

candidate pools를 만든다.

- session summary
- active plan
- pinned notes
- stable profile
- dynamic profile
- memory neighborhood
- document chunks
- recent tail

### Task C. ranking and suppression

- scope filter
- recency filter
- superseded suppression
- duplicate suppression
- budget-aware packing

### Task D. cache

session head version 기준 recall cache를 둔다.

## 4. APIs To Finish

- `POST /v1/sync-turn`
- `POST /v1/prefetch`
- `POST /v1/notes`
- `POST /v1/plans`

## 5. Tests

### Ingest

- same request sent twice
- partial failure retry
- out-of-order events
- worker crash and reclaim

### Recall

- no Codex available
- no embeddings available
- pinned note included
- active plan included
- superseded memory hidden
- different budget sizes

## 6. Done When

- `sync_turn()` is fast
- `prefetch()` returns useful blocks
- duplicate_count works
- notes and plans affect recall
- empty recall is rare and explainable
