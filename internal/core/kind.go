// ============================================================
// FILE     : internal/core/kind.go
// PURPOSE  : Defines canonical enum-like values for memory, edge, job, and artifact classes.
// LAYER    : domain
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : MemoryKind, EdgeKind, MemoryStatus, JobKind, ArtifactClass
// DEPENDS  : plans/02_product-contract_and_direction.md, docs/adr-005-artifact-class-timing.md
// USED_BY  : internal/core records, storage, reasoning/apply contracts
// ------------------------------------------------------------
// AGENT_NOTE: Treat value changes as contract changes and update docs/tests.
// ============================================================

package core

// MemoryKind describes the semantic type of a derived memory.
type MemoryKind string

const (
	// MemoryKindFact is a comparatively verifiable fact.
	MemoryKindFact MemoryKind = "fact"
	// MemoryKindPreference records a like, dislike, or preference.
	MemoryKindPreference MemoryKind = "preference"
	// MemoryKindTrait records a durable tendency or characteristic.
	MemoryKindTrait MemoryKind = "trait"
	// MemoryKindGoal records an intended outcome.
	MemoryKindGoal MemoryKind = "goal"
	// MemoryKindConstraint records a limiting rule or requirement.
	MemoryKindConstraint MemoryKind = "constraint"
	// MemoryKindRelationship records a person, team, project, or object relation.
	MemoryKindRelationship MemoryKind = "relationship"
	// MemoryKindDecision records a decision that has been made.
	MemoryKindDecision MemoryKind = "decision"
	// MemoryKindProcedure records a reusable way of working.
	MemoryKindProcedure MemoryKind = "procedure"
	// MemoryKindTaskState records current task state.
	MemoryKindTaskState MemoryKind = "task_state"
	// MemoryKindDocFact records a fact extracted from a document.
	MemoryKindDocFact MemoryKind = "doc_fact"
	// MemoryKindSummary records a compressed session or topic summary.
	MemoryKindSummary MemoryKind = "summary"
	// MemoryKindHypothesis records an uncertain inference.
	MemoryKindHypothesis MemoryKind = "hypothesis"
)

// EdgeKind describes the relationship between two memories or artifacts.
type EdgeKind string

const (
	// EdgeKindUpdates means the source memory supersedes the target memory.
	EdgeKindUpdates EdgeKind = "updates"
	// EdgeKindExtends means the source memory adds detail to the target memory.
	EdgeKindExtends EdgeKind = "extends"
	// EdgeKindSupports means the source memory strengthens the target memory.
	EdgeKindSupports EdgeKind = "supports"
	// EdgeKindContradicts means the source memory conflicts with the target memory.
	EdgeKindContradicts EdgeKind = "contradicts"
	// EdgeKindDerivedFrom means the source memory was derived from the target artifact.
	EdgeKindDerivedFrom EdgeKind = "derived_from"
	// EdgeKindReferencesDoc means the source memory is grounded in a document.
	EdgeKindReferencesDoc EdgeKind = "references_doc"
	// EdgeKindBelongsTo means the source memory belongs to an entity or scope.
	EdgeKindBelongsTo EdgeKind = "belongs_to"
	// EdgeKindCorrectedBy means the source memory was changed by an operator correction.
	EdgeKindCorrectedBy EdgeKind = "corrected_by"
)

// MemoryStatus describes whether a memory participates in recall.
type MemoryStatus string

const (
	// MemoryStatusActive is eligible for recall.
	MemoryStatusActive MemoryStatus = "active"
	// MemoryStatusSuperseded was replaced by a newer memory.
	MemoryStatusSuperseded MemoryStatus = "superseded"
	// MemoryStatusArchived is retained for provenance but suppressed by default.
	MemoryStatusArchived MemoryStatus = "archived"
	// MemoryStatusDeleted marks a memory removed by an explicit deletion policy.
	MemoryStatusDeleted MemoryStatus = "deleted"
)

// JobKind identifies a worker queue job.
type JobKind string

const (
	// JobKindProcessTurnEvent runs the event-to-memory pipeline.
	JobKindProcessTurnEvent JobKind = "process_turn_event"
	// JobKindEmbedDocumentChunks embeds document retrieval units.
	JobKindEmbedDocumentChunks JobKind = "embed_document_chunks"
	// JobKindDreamSession consolidates a session after activity.
	JobKindDreamSession JobKind = "dream_session"
	// JobKindDreamWorkspace consolidates workspace-level memory.
	JobKindDreamWorkspace JobKind = "dream_workspace"
	// JobKindRebuildProfile recomputes profile snapshots.
	JobKindRebuildProfile JobKind = "rebuild_profile"
	// JobKindMaintenance runs cleanup and backfill work.
	JobKindMaintenance JobKind = "maintenance"
)

// ArtifactClass is the broad retrieval lane for a memory.
type ArtifactClass string

const (
	// ArtifactClassContext represents current-session or short-lived work context.
	ArtifactClassContext ArtifactClass = "context"
	// ArtifactClassKnowledge represents durable facts, preferences, rules, and procedures.
	ArtifactClassKnowledge ArtifactClass = "knowledge"
	// ArtifactClassTimeline represents events, decisions, corrections, and changes.
	ArtifactClassTimeline ArtifactClass = "timeline"
	// ArtifactClassPlan represents goals, task state, and next actions.
	ArtifactClassPlan ArtifactClass = "plan"
)
