// ============================================================
// FILE     : internal/graph/apply.go
// PURPOSE  : Defines the graph apply engine contract and a validating no-op implementation.
// LAYER    : application
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : ApplyEngine, ApplyRequest, ApplyResult, NoopApplyEngine, NewNoopApplyEngine
// DEPENDS  : context, encoding/json, fmt, slices, internal/core, internal/reasoning
// USED_BY  : internal/worker, cmd/worker, tests
// ------------------------------------------------------------
// AGENT_NOTE: Apply must validate structured Stage 2 output before any memory, edge, profile, or trace write.
// ============================================================

package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/reasoning"
)

// ApplyEngine validates and applies structured Stage 2 reasoning output.
type ApplyEngine interface {
	Apply(ctx context.Context, req *ApplyRequest) (*ApplyResult, error)
}

// ApplyRequest is the apply boundary input for one worker job.
type ApplyRequest struct {
	JobID       string                       `json:"job_id"`
	TenantID    string                       `json:"tenant_id"`
	WorkspaceID string                       `json:"workspace_id"`
	RawEventIDs []string                     `json:"raw_event_ids"`
	Reasoning   *reasoning.ProcessTurnResult `json:"reasoning"`
}

// ApplyResult reports what the apply engine committed.
type ApplyResult struct {
	AppliedOperationCount int      `json:"applied_operation_count"`
	MemoryIDs             []string `json:"memory_ids"`
	TraceWritten          bool     `json:"trace_written"`
}

// NoopApplyEngine validates the apply request without writing derived state.
type NoopApplyEngine struct{}

// NewNoopApplyEngine creates a no-op graph apply engine for worker wiring tests.
func NewNoopApplyEngine() *NoopApplyEngine {
	return &NoopApplyEngine{}
}

// Apply validates schema-shaped Stage 2 output and intentionally commits nothing.
func (e *NoopApplyEngine) Apply(_ context.Context, req *ApplyRequest) (*ApplyResult, error) {
	if err := validateApplyRequest(req); err != nil {
		return nil, err
	}
	return &ApplyResult{
		AppliedOperationCount: 0,
		MemoryIDs:             []string{},
		TraceWritten:          false,
	}, nil
}

func validateApplyRequest(req *ApplyRequest) error {
	if req == nil {
		return fmt.Errorf("%w: apply request is required", core.ErrInvalidArgument)
	}
	if req.JobID == "" {
		return fmt.Errorf("%w: apply job_id is required", core.ErrInvalidArgument)
	}
	if req.TenantID == "" {
		return fmt.Errorf("%w: apply tenant_id is required", core.ErrInvalidArgument)
	}
	if req.WorkspaceID == "" {
		return fmt.Errorf("%w: apply workspace_id is required", core.ErrInvalidArgument)
	}
	if len(req.RawEventIDs) == 0 {
		return fmt.Errorf("%w: apply raw_event_ids are required", core.ErrInvalidArgument)
	}
	if req.Reasoning == nil {
		return fmt.Errorf("%w: apply reasoning result is required", core.ErrInvalidArgument)
	}
	if req.Reasoning.Stage2.Trace.SchemaVersion == "" {
		return fmt.Errorf("%w: apply resolve trace schema_version is required", core.ErrInvalidArgument)
	}
	if req.Reasoning.Stage2.Trace.Stage != reasoning.StageNameResolve {
		return fmt.Errorf("%w: apply requires resolve-stage trace", core.ErrInvalidArgument)
	}
	if err := validateJSONObject("profile_delta", req.Reasoning.Stage2.ProfileDelta); err != nil {
		return err
	}
	if err := validateJSONObject("plan_delta", req.Reasoning.Stage2.PlanDelta); err != nil {
		return err
	}
	if err := validateJSONObject("trace.metadata_json", req.Reasoning.Stage2.Trace.MetadataJSON); err != nil {
		return err
	}
	for i, operation := range req.Reasoning.Stage2.Operations {
		if err := validateOperation(i, operation, req.RawEventIDs); err != nil {
			return err
		}
	}
	return nil
}

func validateOperation(index int, operation reasoning.GraphOperation, applyRawEventIDs []string) error {
	if operation.OperationID == "" {
		return fmt.Errorf("%w: operations[%d].operation_id is required", core.ErrInvalidArgument, index)
	}
	if operation.Kind == "" {
		return fmt.Errorf("%w: operations[%d].kind is required", core.ErrInvalidArgument, index)
	}
	if !isSupportedOperationKind(operation.Kind) {
		return fmt.Errorf("%w: operations[%d].kind is unsupported", core.ErrInvalidArgument, index)
	}
	if len(operation.RawEventIDs) == 0 {
		return fmt.Errorf("%w: operations[%d].raw_event_ids are required", core.ErrInvalidArgument, index)
	}
	for _, rawEventID := range operation.RawEventIDs {
		if rawEventID == "" {
			return fmt.Errorf("%w: operations[%d].raw_event_ids cannot contain empty ids", core.ErrInvalidArgument, index)
		}
		if !slices.Contains(applyRawEventIDs, rawEventID) {
			return fmt.Errorf("%w: operations[%d].raw_event_ids must reference the apply bundle", core.ErrInvalidArgument, index)
		}
	}
	if err := validateJSONObject(fmt.Sprintf("operations[%d].metadata", index), operation.Metadata); err != nil {
		return err
	}

	switch operation.Kind {
	case reasoning.OperationKindCreateMemory:
		return validateMemoryMutation(index, operation.Memory, false)
	case reasoning.OperationKindUpdateMemory:
		if err := validateMemoryMutation(index, operation.Memory, true); err != nil {
			return err
		}
		return validateEdgeMutation(index, operation.Edge, core.EdgeKindUpdates, operation.Memory)
	case reasoning.OperationKindExtendMemory:
		if err := validateMemoryMutation(index, operation.Memory, true); err != nil {
			return err
		}
		return validateEdgeMutation(index, operation.Edge, core.EdgeKindExtends, operation.Memory)
	case reasoning.OperationKindArchiveMemory:
		return validateArchiveMutation(index, operation.Memory)
	default:
		return fmt.Errorf("%w: operations[%d].kind is unsupported", core.ErrInvalidArgument, index)
	}
}

func validateMemoryMutation(index int, memory *reasoning.MemoryMutation, targetRequired bool) error {
	if memory == nil {
		return fmt.Errorf("%w: operations[%d].memory is required", core.ErrInvalidArgument, index)
	}
	if targetRequired && memory.TargetID == "" {
		return fmt.Errorf("%w: operations[%d].memory.target_id is required", core.ErrInvalidArgument, index)
	}
	if !isSupportedMemoryKind(memory.Kind) {
		return fmt.Errorf("%w: operations[%d].memory.kind is unsupported", core.ErrInvalidArgument, index)
	}
	if !isSupportedArtifactClass(memory.ArtifactClass) {
		return fmt.Errorf("%w: operations[%d].memory.artifact_class is unsupported", core.ErrInvalidArgument, index)
	}
	if !isSupportedScope(memory.Scope) {
		return fmt.Errorf("%w: operations[%d].memory.scope is required", core.ErrInvalidArgument, index)
	}
	if memory.Scope == core.MemoryScopeGroupShared && (memory.GroupID == nil || *memory.GroupID == "") {
		return fmt.Errorf("%w: operations[%d].memory.group_id is required for group_shared scope", core.ErrInvalidArgument, index)
	}
	if memory.OwnerEntityID == "" {
		return fmt.Errorf("%w: operations[%d].memory.owner_entity_id is required", core.ErrInvalidArgument, index)
	}
	if memory.Text == "" {
		return fmt.Errorf("%w: operations[%d].memory.text is required", core.ErrInvalidArgument, index)
	}
	if !isValidConfidence(memory.Confidence) {
		return fmt.Errorf("%w: operations[%d].memory.confidence must be greater than 0 and less than or equal to 1", core.ErrInvalidArgument, index)
	}
	if err := validateJSONObject(fmt.Sprintf("operations[%d].memory.metadata_json", index), memory.MetadataJSON); err != nil {
		return err
	}
	return nil
}

func validateArchiveMutation(index int, memory *reasoning.MemoryMutation) error {
	if memory == nil {
		return fmt.Errorf("%w: operations[%d].memory is required", core.ErrInvalidArgument, index)
	}
	if memory.TargetID == "" && memory.MemoryID == "" {
		return fmt.Errorf("%w: operations[%d].memory target is required", core.ErrInvalidArgument, index)
	}
	if !isSupportedScope(memory.Scope) {
		return fmt.Errorf("%w: operations[%d].memory.scope is required", core.ErrInvalidArgument, index)
	}
	if memory.Scope == core.MemoryScopeGroupShared && (memory.GroupID == nil || *memory.GroupID == "") {
		return fmt.Errorf("%w: operations[%d].memory.group_id is required for group_shared scope", core.ErrInvalidArgument, index)
	}
	if memory.OwnerEntityID == "" {
		return fmt.Errorf("%w: operations[%d].memory.owner_entity_id is required", core.ErrInvalidArgument, index)
	}
	if memory.Confidence != 0 && !isValidConfidence(memory.Confidence) {
		return fmt.Errorf("%w: operations[%d].memory.confidence must be greater than 0 and less than or equal to 1", core.ErrInvalidArgument, index)
	}
	if err := validateJSONObject(fmt.Sprintf("operations[%d].memory.metadata_json", index), memory.MetadataJSON); err != nil {
		return err
	}
	return nil
}

func validateEdgeMutation(index int, edge *reasoning.EdgeMutation, expectedKind core.EdgeKind, memory *reasoning.MemoryMutation) error {
	if edge == nil {
		return fmt.Errorf("%w: operations[%d].edge is required", core.ErrInvalidArgument, index)
	}
	if edge.EdgeKind != expectedKind {
		return fmt.Errorf("%w: operations[%d].edge.edge_kind must be %q", core.ErrInvalidArgument, index, expectedKind)
	}
	if edge.ToMemoryID == "" {
		return fmt.Errorf("%w: operations[%d].edge.to_memory_id is required", core.ErrInvalidArgument, index)
	}
	if edge.ToMemoryID != memory.TargetID {
		return fmt.Errorf("%w: operations[%d].edge.to_memory_id must match memory.target_id", core.ErrInvalidArgument, index)
	}
	if memory.MemoryID != "" && edge.FromMemoryID != "" && edge.FromMemoryID != memory.MemoryID {
		return fmt.Errorf("%w: operations[%d].edge.from_memory_id must match memory.memory_id", core.ErrInvalidArgument, index)
	}
	if edge.FromMemoryID == edge.ToMemoryID {
		return fmt.Errorf("%w: operations[%d].edge cannot target itself", core.ErrInvalidArgument, index)
	}
	if !isValidConfidence(edge.Confidence) {
		return fmt.Errorf("%w: operations[%d].edge.confidence must be greater than 0 and less than or equal to 1", core.ErrInvalidArgument, index)
	}
	return nil
}

func isSupportedOperationKind(kind reasoning.OperationKind) bool {
	switch kind {
	case reasoning.OperationKindCreateMemory,
		reasoning.OperationKindUpdateMemory,
		reasoning.OperationKindExtendMemory,
		reasoning.OperationKindArchiveMemory:
		return true
	default:
		return false
	}
}

func isSupportedMemoryKind(kind core.MemoryKind) bool {
	switch kind {
	case core.MemoryKindFact,
		core.MemoryKindPreference,
		core.MemoryKindTrait,
		core.MemoryKindGoal,
		core.MemoryKindConstraint,
		core.MemoryKindRelationship,
		core.MemoryKindDecision,
		core.MemoryKindProcedure,
		core.MemoryKindTaskState,
		core.MemoryKindDocFact,
		core.MemoryKindSummary,
		core.MemoryKindHypothesis:
		return true
	default:
		return false
	}
}

func isSupportedArtifactClass(class core.ArtifactClass) bool {
	switch class {
	case core.ArtifactClassContext,
		core.ArtifactClassKnowledge,
		core.ArtifactClassTimeline,
		core.ArtifactClassPlan:
		return true
	default:
		return false
	}
}

func isSupportedScope(scope core.MemoryScope) bool {
	switch scope {
	case core.MemoryScopeAgentPrivate,
		core.MemoryScopeWorkspaceShared,
		core.MemoryScopeGroupShared,
		core.MemoryScopeSessionScratch:
		return true
	default:
		return false
	}
}

func isValidConfidence(confidence float64) bool {
	return confidence > 0 && confidence <= 1
}

func validateJSONObject(field string, raw json.RawMessage) error {
	if len(raw) == 0 {
		return nil
	}
	if !json.Valid(raw) {
		return fmt.Errorf("%w: %s must be valid JSON", core.ErrInvalidArgument, field)
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil {
		return fmt.Errorf("%w: %s must be a JSON object", core.ErrInvalidArgument, field)
	}
	if object == nil {
		return fmt.Errorf("%w: %s must be a JSON object", core.ErrInvalidArgument, field)
	}
	return nil
}
