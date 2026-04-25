// ============================================================
// FILE     : internal/graph/store_apply.go
// PURPOSE  : Applies validated create, extend, and update memory operations to durable graph storage.
// LAYER    : application
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : MemoryTraceCreator, StoreBackedApplyEngine, NewStoreBackedApplyEngine
// DEPENDS  : context, crypto/sha256, encoding/hex, encoding/json, fmt, strings, time, internal/core, internal/reasoning
// USED_BY  : cmd/worker, internal/graph tests
// ------------------------------------------------------------
// AGENT_NOTE: Latest-changing updates must be one storage transaction with trace, edge, and prior-memory supersession.
// ============================================================

package graph

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/reasoning"
)

// MemoryTraceCreator persists derived memory writes with mandatory trace provenance.
type MemoryTraceCreator interface {
	CreateMemoryWithTrace(ctx context.Context, memory *core.Memory, trace *core.MemoryTrace) error
	CreateMemoryWithTraceAndEdge(ctx context.Context, memory *core.Memory, trace *core.MemoryTrace, edge *core.MemoryEdge) error
	CreateMemoryWithTraceAndUpdateEdge(ctx context.Context, memory *core.Memory, trace *core.MemoryTrace, edge *core.MemoryEdge) error
}

// StoreBackedApplyEngine writes validated create_memory, extend_memory, and update_memory operations to storage.
type StoreBackedApplyEngine struct {
	memories MemoryTraceCreator
	now      func() time.Time
}

// NewStoreBackedApplyEngine creates the first write-capable graph apply engine.
func NewStoreBackedApplyEngine(memories MemoryTraceCreator) (*StoreBackedApplyEngine, error) {
	if memories == nil {
		return nil, fmt.Errorf("%w: graph memory store is required", core.ErrInvalidArgument)
	}
	return &StoreBackedApplyEngine{
		memories: memories,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}, nil
}

// Apply validates Stage 2 output and writes safe memory operations with provenance.
func (e *StoreBackedApplyEngine) Apply(ctx context.Context, req *ApplyRequest) (*ApplyResult, error) {
	if err := validateApplyRequest(req); err != nil {
		return nil, err
	}
	if err := rejectUnsupportedWriteScope(req); err != nil {
		return nil, err
	}

	createdIDs := make([]string, 0, len(req.Reasoning.Stage2.Operations))
	for i, operation := range req.Reasoning.Stage2.Operations {
		memory, trace, err := e.buildMemoryTrace(req, i, operation)
		if err != nil {
			return nil, err
		}
		switch operation.Kind {
		case reasoning.OperationKindCreateMemory:
			if err := e.memories.CreateMemoryWithTrace(ctx, memory, trace); err != nil {
				return nil, fmt.Errorf("create memory with trace: %w", err)
			}
		case reasoning.OperationKindExtendMemory:
			edge, err := buildMemoryEdge(req, memory, operation)
			if err != nil {
				return nil, err
			}
			if err := e.memories.CreateMemoryWithTraceAndEdge(ctx, memory, trace, edge); err != nil {
				return nil, fmt.Errorf("create extension memory with trace and edge: %w", err)
			}
		case reasoning.OperationKindUpdateMemory:
			edge, err := buildMemoryEdge(req, memory, operation)
			if err != nil {
				return nil, err
			}
			if err := e.memories.CreateMemoryWithTraceAndUpdateEdge(ctx, memory, trace, edge); err != nil {
				return nil, fmt.Errorf("create update memory with trace and supersession edge: %w", err)
			}
		default:
			return nil, fmt.Errorf("%w: operations[%d].kind %q is validation-only in store-backed apply", core.ErrNotImplemented, i, operation.Kind)
		}
		createdIDs = append(createdIDs, memory.ID)
	}
	return &ApplyResult{
		AppliedOperationCount: len(createdIDs),
		MemoryIDs:             createdIDs,
		TraceWritten:          len(createdIDs) > 0,
	}, nil
}

func rejectUnsupportedWriteScope(req *ApplyRequest) error {
	if hasNonEmptyObject(req.Reasoning.Stage2.ProfileDelta) {
		return fmt.Errorf("%w: profile_delta writes are not implemented", core.ErrNotImplemented)
	}
	if strings.TrimSpace(req.Reasoning.Stage2.SessionSummary) != "" {
		return fmt.Errorf("%w: session summary writes are not implemented", core.ErrNotImplemented)
	}
	if hasNonEmptyObject(req.Reasoning.Stage2.PlanDelta) {
		return fmt.Errorf("%w: plan_delta writes are not implemented", core.ErrNotImplemented)
	}
	for i, operation := range req.Reasoning.Stage2.Operations {
		if operation.Memory.Scope == core.MemoryScopeGroupShared {
			return fmt.Errorf("%w: operations[%d].memory.group_shared requires membership validation before writes", core.ErrNotImplemented, i)
		}
		switch operation.Kind {
		case reasoning.OperationKindCreateMemory:
			if operation.Edge != nil {
				return fmt.Errorf("%w: operations[%d].edge writes are not implemented for create_memory", core.ErrNotImplemented, i)
			}
		case reasoning.OperationKindExtendMemory:
			// extend_memory is the safe lineage write: it adds an extends edge while leaving the target memory alive.
		case reasoning.OperationKindUpdateMemory:
			// update_memory writes a new active memory, links it to the prior latest memory, and supersedes that target atomically.
		case reasoning.OperationKindArchiveMemory:
			return fmt.Errorf("%w: operations[%d].kind %q requires archive status handling", core.ErrNotImplemented, i, operation.Kind)
		default:
			return fmt.Errorf("%w: operations[%d].kind %q is validation-only in store-backed apply", core.ErrNotImplemented, i, operation.Kind)
		}
	}
	return nil
}

func (e *StoreBackedApplyEngine) buildMemoryTrace(req *ApplyRequest, index int, operation reasoning.GraphOperation) (*core.Memory, *core.MemoryTrace, error) {
	createdAt := e.now().UTC()
	mutation := operation.Memory
	memory := &core.Memory{
		ID:            deterministicID("mem", req.TenantID, req.WorkspaceID, req.JobID, operation.OperationID),
		TenantID:      req.TenantID,
		WorkspaceID:   req.WorkspaceID,
		Scope:         mutation.Scope,
		GroupID:       mutation.GroupID,
		OwnerEntityID: mutation.OwnerEntityID,
		Kind:          mutation.Kind,
		ArtifactClass: mutation.ArtifactClass,
		Text:          mutation.Text,
		Fingerprint:   memoryFingerprint(req, operation),
		Confidence:    mutation.Confidence,
		Status:        core.MemoryStatusActive,
		ValidFrom:     createdAt,
		LatestFlag:    true,
		MetadataJSON:  mutation.MetadataJSON,
		CreatedAt:     createdAt,
		UpdatedAt:     createdAt,
	}

	candidateSnapshot, err := json.Marshal(req.Reasoning.Stage1)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal stage 1 candidate snapshot: %w", err)
	}
	appliedOperations, err := json.Marshal([]reasoning.GraphOperation{operation})
	if err != nil {
		return nil, nil, fmt.Errorf("marshal applied operation %d: %w", index, err)
	}
	trace := &core.MemoryTrace{
		MemoryID:              memory.ID,
		RawEventIDs:           append([]string(nil), operation.RawEventIDs...),
		ReasoningJobID:        req.JobID,
		ReasoningStage:        string(reasoning.StageNameResolve),
		CandidateSnapshotJSON: json.RawMessage(candidateSnapshot),
		AppliedOperationsJSON: json.RawMessage(appliedOperations),
		RelatedDocumentIDs:    []string{},
		CreatedAt:             createdAt,
	}
	return memory, trace, nil
}

func buildMemoryEdge(req *ApplyRequest, memory *core.Memory, operation reasoning.GraphOperation) (*core.MemoryEdge, error) {
	if operation.Edge == nil {
		return nil, fmt.Errorf("%w: memory edge is required", core.ErrInvalidArgument)
	}
	if operation.Edge.ToMemoryID == "" {
		return nil, fmt.Errorf("%w: memory edge target is required", core.ErrInvalidArgument)
	}
	if operation.Edge.ToMemoryID == memory.ID {
		return nil, fmt.Errorf("%w: memory edge cannot target itself", core.ErrInvalidArgument)
	}
	return &core.MemoryEdge{
		FromMemoryID:   memory.ID,
		ToMemoryID:     operation.Edge.ToMemoryID,
		EdgeKind:       operation.Edge.EdgeKind,
		Confidence:     operation.Edge.Confidence,
		CreatedByJobID: req.JobID,
		CreatedAt:      memory.CreatedAt,
	}, nil
}

func deterministicID(prefix string, parts ...string) string {
	sum := hashParts(parts...)
	return prefix + "_" + hex.EncodeToString(sum[:12])
}

func memoryFingerprint(req *ApplyRequest, operation reasoning.GraphOperation) string {
	payload := struct {
		TenantID      string             `json:"tenant_id"`
		WorkspaceID   string             `json:"workspace_id"`
		Scope         core.MemoryScope   `json:"scope"`
		GroupID       *string            `json:"group_id,omitempty"`
		OwnerEntityID string             `json:"owner_entity_id"`
		Kind          core.MemoryKind    `json:"kind"`
		ArtifactClass core.ArtifactClass `json:"artifact_class"`
		Text          string             `json:"text"`
		RawEventIDs   []string           `json:"raw_event_ids"`
	}{
		TenantID:      req.TenantID,
		WorkspaceID:   req.WorkspaceID,
		Scope:         operation.Memory.Scope,
		GroupID:       operation.Memory.GroupID,
		OwnerEntityID: operation.Memory.OwnerEntityID,
		Kind:          operation.Memory.Kind,
		ArtifactClass: operation.Memory.ArtifactClass,
		Text:          operation.Memory.Text,
		RawEventIDs:   operation.RawEventIDs,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return deterministicID("fp", req.JobID, operation.OperationID)
	}
	sum := sha256.Sum256(data)
	return "fp_" + hex.EncodeToString(sum[:16])
}

func hashParts(parts ...string) [32]byte {
	hash := sha256.New()
	for _, part := range parts {
		hash.Write([]byte(part))
		hash.Write([]byte{0})
	}
	var sum [32]byte
	copy(sum[:], hash.Sum(nil))
	return sum
}

func hasNonEmptyObject(raw json.RawMessage) bool {
	if len(raw) == 0 {
		return false
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil {
		return true
	}
	return len(object) > 0
}
