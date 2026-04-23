// Package store defines storage contracts for VibeGravity persistence.
package store

import (
	"context"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

// RawEventStore persists immutable raw events.
type RawEventStore interface {
	// AppendRawEvents inserts raw events idempotently and returns their IDs.
	AppendRawEvents(ctx context.Context, events []*core.RawEvent) ([]string, error)
	// GetRawEvents loads raw events by ID.
	GetRawEvents(ctx context.Context, ids []string) ([]*core.RawEvent, error)
}

// MemoryStore persists derived memories, edges, and traces.
type MemoryStore interface {
	// UpsertMemory writes a memory after apply validation.
	UpsertMemory(ctx context.Context, memory *core.Memory) error
	// GetMemory loads one memory by ID.
	GetMemory(ctx context.Context, memoryID string) (*core.Memory, error)
	// SearchMemories searches memories with the core search request contract.
	SearchMemories(ctx context.Context, req *core.SearchMemoriesRequest) (*core.SearchMemoriesResponse, error)
	// UpsertMemoryEdge writes a graph edge.
	UpsertMemoryEdge(ctx context.Context, edge *core.MemoryEdge) error
	// WriteMemoryTrace writes mandatory provenance for a memory.
	WriteMemoryTrace(ctx context.Context, trace *core.MemoryTrace) error
	// ExplainMemory loads provenance for one memory.
	ExplainMemory(ctx context.Context, req *core.ExplainMemoryRequest) (*core.ExplainMemoryResponse, error)
}

// JobStore persists and claims worker jobs.
type JobStore interface {
	// EnqueueJobs inserts jobs created by the hot ingest path.
	EnqueueJobs(ctx context.Context, jobs []*core.IngestJob) ([]string, error)
	// ClaimJobs claims available jobs for a worker.
	ClaimJobs(ctx context.Context, workerID string, limit int) ([]*core.IngestJob, error)
	// CompleteJob marks a job complete.
	CompleteJob(ctx context.Context, jobID string) error
	// FailJob records a failed attempt and retry state.
	FailJob(ctx context.Context, jobID string, err error) error
}

// ProfileStore persists rebuildable profile snapshots.
type ProfileStore interface {
	// GetProfile loads a profile snapshot.
	GetProfile(ctx context.Context, entityID string, scope core.MemoryScope) (*core.Profile, error)
	// UpsertProfile writes a profile snapshot.
	UpsertProfile(ctx context.Context, profile *core.Profile) error
}

// NoteStore persists human-authored notes.
type NoteStore interface {
	// AddNote writes a note.
	AddNote(ctx context.Context, note *core.Note) error
	// ListPinnedNotes loads pinned notes for recall.
	ListPinnedNotes(ctx context.Context, workspaceID string, scopes []core.MemoryScope) ([]*core.Note, error)
}

// PlanStore persists structured plans and plan items.
type PlanStore interface {
	// CreatePlan writes a plan and its initial items.
	CreatePlan(ctx context.Context, plan *core.Plan, items []*core.PlanItem) error
	// UpdatePlan updates a plan and its items.
	UpdatePlan(ctx context.Context, plan *core.Plan, items []*core.PlanItem) error
	// GetActivePlans loads active plans for recall.
	GetActivePlans(ctx context.Context, workspaceID string, scopes []core.MemoryScope) ([]*core.Plan, error)
}

// DocumentStore persists documents and searchable chunks.
type DocumentStore interface {
	// AddDocument writes a document.
	AddDocument(ctx context.Context, document *core.Document) error
	// AddDocumentChunks writes retrieval chunks for a document.
	AddDocumentChunks(ctx context.Context, chunks []*core.DocumentChunk) error
	// SearchDocuments searches document chunks with the core search contract.
	SearchDocuments(ctx context.Context, req *core.SearchDocumentsRequest) (*core.SearchDocumentsResponse, error)
}

// SessionSummaryStore persists per-session summaries.
type SessionSummaryStore interface {
	// UpsertSessionSummary writes a session summary.
	UpsertSessionSummary(ctx context.Context, summary *core.SessionSummary) error
	// GetSessionSummary loads the current summary for a session.
	GetSessionSummary(ctx context.Context, sessionID string) (*core.SessionSummary, error)
}

// GroupStore persists memory groups and memberships.
type GroupStore interface {
	// CreateMemoryGroup writes a memory group.
	CreateMemoryGroup(ctx context.Context, group *core.MemoryGroup) error
	// AddMembership adds an entity to a memory group.
	AddMembership(ctx context.Context, membership *core.MemoryGroupMembership) error
	// ListMemberships loads memberships for a memory group.
	ListMemberships(ctx context.Context, groupID string) ([]*core.MemoryGroupMembership, error)
}
