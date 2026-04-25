// ============================================================
// FILE     : internal/store/store.go
// PURPOSE  : Defines persistence interfaces for raw, memory, job, profile, note, plan, document, group, and dreaming stores.
// LAYER    : infra
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : RawEventStore, MemoryStore, CorrectionStore, TimelineStore, JobStore, JobMetricsStore, ProfileStore, NoteStore, PlanStore, DocumentStore, SessionSummaryStore, GroupStore, DreamingStore
// DEPENDS  : context, internal/core
// USED_BY  : internal/store/postgres, service implementations
// ------------------------------------------------------------
// AGENT_NOTE: Store contracts must preserve idempotency, provenance, and scope separation.
// ============================================================

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

// CorrectionStore persists human correction intent beside immutable raw events.
type CorrectionStore interface {
	// RecordMemoryCorrection writes the raw correction event and operator-visible artifact idempotently.
	RecordMemoryCorrection(ctx context.Context, event *core.RawEvent, correction *core.MemoryCorrection) (*core.MemoryCorrection, error)
	// GetMemoryCorrectionByIdempotency loads an existing correction artifact for replay validation.
	GetMemoryCorrectionByIdempotency(ctx context.Context, tenantID string, workspaceID string, idempotencyKey string) (*core.MemoryCorrection, error)
}

// CorrectionApplyJobStore creates completed correction apply provenance jobs.
type CorrectionApplyJobStore interface {
	// EnsureCorrectionApplyJob creates or reuses the deterministic completed job row used by memory_trace and memory_edges FKs.
	EnsureCorrectionApplyJob(ctx context.Context, job *core.IngestJob) (string, error)
}

// TimelineStore reads operator-visible memory activity.
type TimelineStore interface {
	// GetTimeline loads a read-only timeline view.
	GetTimeline(ctx context.Context, req *core.GetTimelineRequest) (*core.GetTimelineResponse, error)
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
	// BlockJob records deterministic unsupported work without scheduling retry.
	BlockJob(ctx context.Context, jobID string, err error) error
}

// JobMetricsStore reads worker queue health without mutating job state.
type JobMetricsStore interface {
	// GetJobBacklogMetrics returns operator-visible backlog counts and recovery estimates.
	GetJobBacklogMetrics(ctx context.Context, req *core.JobBacklogMetricsRequest) (*core.JobBacklogMetrics, error)
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
	ListPinnedNotes(ctx context.Context, req *core.ListPinnedNotesRequest) ([]*core.Note, error)
}

// PlanStore persists structured plans and plan items.
type PlanStore interface {
	// CreatePlan writes a plan and its initial items.
	CreatePlan(ctx context.Context, plan *core.Plan, items []*core.PlanItem) error
	// UpdatePlan updates a plan and its items.
	UpdatePlan(ctx context.Context, plan *core.Plan, items []*core.PlanItem) error
	// GetActivePlans loads active plans for recall.
	GetActivePlans(ctx context.Context, req *core.GetActivePlansRequest) ([]*core.Plan, error)
}

// DocumentStore persists documents and searchable chunks.
type DocumentStore interface {
	// AddDocumentWithChunks writes a document and replaces its chunks atomically.
	AddDocumentWithChunks(ctx context.Context, document *core.Document, chunks []*core.DocumentChunk) error
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

// DreamingStore persists and loads background consolidation state.
type DreamingStore interface {
	// LoadDreamingSessionInput loads raw event and derived memory inputs for one session.
	LoadDreamingSessionInput(ctx context.Context, req *core.DreamSessionRequest) (*core.DreamingSessionInput, error)
	// PromoteMemories marks existing memories with a deeper dreaming tier without changing scope.
	PromoteMemories(ctx context.Context, req *core.DreamingPromotionRequest) (*core.DreamingPromotionResult, error)
}

// GroupStore persists memory groups and memberships.
type GroupStore interface {
	// CreateMemoryGroup writes a memory group.
	CreateMemoryGroup(ctx context.Context, group *core.MemoryGroup) error
	// AddMembership adds an entity to a memory group.
	AddMembership(ctx context.Context, membership *core.MemoryGroupMembership) error
	// ListMemberships loads memberships for a memory group.
	ListMemberships(ctx context.Context, groupID string) ([]*core.MemoryGroupMembership, error)
	// ListMembershipsForEntity loads groups visible to an entity in one workspace.
	ListMembershipsForEntity(ctx context.Context, tenantID string, workspaceID string, entityID string) ([]*core.MemoryGroupMembership, error)
}
