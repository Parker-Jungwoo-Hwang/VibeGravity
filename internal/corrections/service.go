// ============================================================
// FILE     : internal/corrections/service.go
// PURPOSE  : Records operator corrections and applies provenance-backed supersession.
// LAYER    : application
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : Service, NewService, SupersessionStore
// DEPENDS  : crypto/sha256, encoding/json, internal/core, internal/store
// USED_BY  : internal/kernel, internal/corrections tests
// ------------------------------------------------------------
// AGENT_NOTE: Do not mutate prior memory trace; write replacement memory, trace, edge, and correction state together.
// ============================================================

package corrections

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/store"
)

// SupersessionStore applies the replacement memory, trace, edge, and correction status transaction.
type SupersessionStore interface {
	CreateCorrectionSupersession(ctx context.Context, memory *core.Memory, trace *core.MemoryTrace, edge *core.MemoryEdge, correctionID string) error
}

// Service owns correction use cases.
type Service struct {
	memories    store.MemoryStore
	corrections store.CorrectionStore
	jobs        store.CorrectionApplyJobStore
	clock       func() time.Time
}

// NewService builds a correction service.
func NewService(memories store.MemoryStore, corrections store.CorrectionStore, jobs store.CorrectionApplyJobStore) *Service {
	return &Service{
		memories:    memories,
		corrections: corrections,
		jobs:        jobs,
		clock:       time.Now,
	}
}

// CorrectMemory records human correction intent and applies an operator-driven supersession.
func (s *Service) CorrectMemory(ctx context.Context, req *core.CorrectMemoryRequest) (*core.CorrectMemoryResponse, error) {
	if s == nil || s.memories == nil || s.corrections == nil || s.jobs == nil {
		return nil, fmt.Errorf("%w: correct memory", core.ErrNotImplemented)
	}
	if req == nil {
		return nil, fmt.Errorf("%w: correct memory request is required", core.ErrInvalidArgument)
	}
	if err := requireFields(map[string]string{
		"tenant_id":       req.TenantID,
		"workspace_id":    req.WorkspaceID,
		"memory_id":       req.MemoryID,
		"operator_id":     req.OperatorID,
		"idempotency_key": req.IdempotencyKey,
		"correction_text": req.CorrectionText,
	}); err != nil {
		return nil, err
	}
	recorded, err := s.corrections.GetMemoryCorrectionByIdempotency(ctx, req.TenantID, req.WorkspaceID, req.IdempotencyKey)
	if err == nil {
		if err := validateCorrectionReplay(req, recorded); err != nil {
			return nil, err
		}
		memory, err := s.memories.GetMemory(ctx, recorded.MemoryID)
		if err != nil {
			return nil, err
		}
		if memory == nil || memory.TenantID != req.TenantID || memory.WorkspaceID != req.WorkspaceID {
			return nil, core.ErrNotFound
		}
		if err := validateCorrectionTargetVisible(req, memory); err != nil {
			return nil, err
		}
		return s.applyRecordedCorrection(ctx, req, memory, recorded, s.clock().UTC())
	}
	if err != nil && !errors.Is(err, core.ErrNotFound) {
		return nil, err
	}
	memory, err := s.memories.GetMemory(ctx, req.MemoryID)
	if err != nil {
		return nil, err
	}
	if memory == nil {
		return nil, core.ErrNotFound
	}
	if memory.TenantID != req.TenantID || memory.WorkspaceID != req.WorkspaceID {
		return nil, core.ErrNotFound
	}
	if err := validateCorrectionTargetVisible(req, memory); err != nil {
		return nil, err
	}
	if memory.Status != core.MemoryStatusActive || !memory.LatestFlag {
		return nil, fmt.Errorf("%w: correction target memory must be active latest", core.ErrConflict)
	}
	now := s.clock().UTC()
	payload, err := correctionPayload(req)
	if err != nil {
		return nil, err
	}
	event := &core.RawEvent{
		TenantID:       req.TenantID,
		WorkspaceID:    req.WorkspaceID,
		SessionID:      "correction:" + req.MemoryID,
		ActorID:        req.OperatorID,
		EventKind:      "memory_correction",
		Source:         "operator_correction",
		IdempotencyKey: req.IdempotencyKey,
		Fingerprint:    correctionFingerprint(req),
		OccurredAt:     now,
		PayloadJSON:    payload,
		CreatedAt:      now,
	}
	correction := &core.MemoryCorrection{
		TenantID:       req.TenantID,
		WorkspaceID:    req.WorkspaceID,
		MemoryID:       req.MemoryID,
		OperatorID:     req.OperatorID,
		IdempotencyKey: req.IdempotencyKey,
		CorrectionText: strings.TrimSpace(req.CorrectionText),
		EvidenceJSON:   jsonOrEmpty(req.EvidenceJSON),
		Status:         "recorded",
		CreatedAt:      now,
	}
	recorded, err = s.corrections.RecordMemoryCorrection(ctx, event, correction)
	if err != nil {
		return nil, err
	}
	if err := validateCorrectionReplay(req, recorded); err != nil {
		return nil, err
	}
	return s.applyRecordedCorrection(ctx, req, memory, recorded, now)
}

func (s *Service) applyRecordedCorrection(ctx context.Context, req *core.CorrectMemoryRequest, memory *core.Memory, recorded *core.MemoryCorrection, now time.Time) (*core.CorrectMemoryResponse, error) {
	supersessions, ok := s.memories.(SupersessionStore)
	if !ok {
		return nil, fmt.Errorf("%w: correct memory supersession store", core.ErrNotImplemented)
	}
	correctionJob := core.NewCorrectionApplyJob(core.CorrectionApplyJobInput{
		TenantID:       req.TenantID,
		WorkspaceID:    req.WorkspaceID,
		CorrectionID:   recorded.ID,
		TargetMemoryID: recorded.MemoryID,
		IdempotencyKey: req.IdempotencyKey,
		RawEventID:     recorded.RawEventID,
		OperatorID:     req.OperatorID,
		AppliedAt:      now,
	})
	correctionJobID, err := s.jobs.EnsureCorrectionApplyJob(ctx, correctionJob)
	if err != nil {
		return nil, err
	}
	replacement, trace, edge, err := buildCorrectionSupersession(memory, recorded, correctionJobID, now)
	if err != nil {
		return nil, err
	}
	if err := supersessions.CreateCorrectionSupersession(ctx, replacement, trace, edge, recorded.ID); err != nil {
		return nil, err
	}
	return &core.CorrectMemoryResponse{
		MemoryID:           recorded.MemoryID,
		RawEventID:         recorded.RawEventID,
		CorrectionID:       recorded.ID,
		CorrectionRecorded: true,
		TraceWritten:       true,
		Status:             "applied",
	}, nil
}

func requireFields(fields map[string]string) error {
	for name, value := range fields {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%w: %s is required", core.ErrInvalidArgument, name)
		}
	}
	return nil
}

func validateCorrectionTargetVisible(req *core.CorrectMemoryRequest, memory *core.Memory) error {
	switch memory.Scope {
	case core.MemoryScopeAgentPrivate:
		entityID := strings.TrimSpace(req.EntityID)
		if entityID == "" || entityID != memory.OwnerEntityID {
			return fmt.Errorf("%w: correction target memory is not visible to entity", core.ErrNotFound)
		}
	case core.MemoryScopeGroupShared:
		if memory.GroupID == nil {
			return fmt.Errorf("%w: correction target group is missing", core.ErrNotFound)
		}
		targetGroupID := strings.TrimSpace(*memory.GroupID)
		if targetGroupID == "" {
			return fmt.Errorf("%w: correction target group is missing", core.ErrNotFound)
		}
		for _, visibleGroupID := range req.VisibleGroupIDs {
			if strings.TrimSpace(visibleGroupID) == targetGroupID {
				return nil
			}
		}
		return fmt.Errorf("%w: correction target memory is not visible to group", core.ErrNotFound)
	case core.MemoryScopeWorkspaceShared, core.MemoryScopeSessionScratch:
		return nil
	default:
		return fmt.Errorf("%w: correction target memory has unsupported scope", core.ErrInvalidArgument)
	}
	return nil
}

func jsonOrEmpty(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 {
		return json.RawMessage(`{}`)
	}
	return raw
}

func correctionPayload(req *core.CorrectMemoryRequest) (json.RawMessage, error) {
	payload := struct {
		MemoryID       string          `json:"memory_id"`
		OperatorID     string          `json:"operator_id"`
		CorrectionText string          `json:"correction_text"`
		EvidenceJSON   json.RawMessage `json:"evidence_json"`
	}{
		MemoryID:       req.MemoryID,
		OperatorID:     req.OperatorID,
		CorrectionText: strings.TrimSpace(req.CorrectionText),
		EvidenceJSON:   jsonOrEmpty(req.EvidenceJSON),
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal correction payload: %w", err)
	}
	return data, nil
}

func validateCorrectionReplay(req *core.CorrectMemoryRequest, recorded *core.MemoryCorrection) error {
	if recorded == nil {
		return fmt.Errorf("%w: recorded correction is required", core.ErrInvalidArgument)
	}
	if recorded.TenantID != req.TenantID ||
		recorded.WorkspaceID != req.WorkspaceID ||
		recorded.MemoryID != req.MemoryID ||
		recorded.OperatorID != req.OperatorID ||
		recorded.IdempotencyKey != req.IdempotencyKey ||
		strings.TrimSpace(recorded.CorrectionText) != strings.TrimSpace(req.CorrectionText) ||
		!jsonEqual(recorded.EvidenceJSON, jsonOrEmpty(req.EvidenceJSON)) {
		return fmt.Errorf("%w: correction idempotency key belongs to different evidence", core.ErrConflict)
	}
	return nil
}

func jsonEqual(left json.RawMessage, right json.RawMessage) bool {
	var leftValue any
	var rightValue any
	if err := json.Unmarshal(jsonOrEmpty(left), &leftValue); err != nil {
		return false
	}
	if err := json.Unmarshal(jsonOrEmpty(right), &rightValue); err != nil {
		return false
	}
	return reflect.DeepEqual(leftValue, rightValue)
}

func correctionFingerprint(req *core.CorrectMemoryRequest) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{
		req.TenantID,
		req.WorkspaceID,
		req.MemoryID,
		req.OperatorID,
		req.IdempotencyKey,
		strings.TrimSpace(req.CorrectionText),
		string(jsonOrEmpty(req.EvidenceJSON)),
	}, "\x00")))
	return fmt.Sprintf("sha256:%x", sum[:])
}

func buildCorrectionSupersession(target *core.Memory, correction *core.MemoryCorrection, reasoningJobID string, correctedAt time.Time) (*core.Memory, *core.MemoryTrace, *core.MemoryEdge, error) {
	if target == nil {
		return nil, nil, nil, fmt.Errorf("%w: correction target memory is required", core.ErrInvalidArgument)
	}
	if correction == nil {
		return nil, nil, nil, fmt.Errorf("%w: recorded correction is required", core.ErrInvalidArgument)
	}
	if strings.TrimSpace(reasoningJobID) == "" {
		return nil, nil, nil, fmt.Errorf("%w: correction apply job id is required", core.ErrInvalidArgument)
	}
	replacement := &core.Memory{
		ID:            correctionSupersessionID(target.ID, correction.IdempotencyKey),
		TenantID:      target.TenantID,
		WorkspaceID:   target.WorkspaceID,
		Scope:         target.Scope,
		GroupID:       target.GroupID,
		OwnerEntityID: target.OwnerEntityID,
		Kind:          target.Kind,
		ArtifactClass: target.ArtifactClass,
		Text:          strings.TrimSpace(correction.CorrectionText),
		Fingerprint:   correctionSupersessionFingerprint(target, correction),
		Confidence:    1.0,
		Status:        core.MemoryStatusActive,
		ValidFrom:     correctedAt.UTC(),
		LatestFlag:    true,
		MetadataJSON:  correctionSupersessionMetadata(target, correction),
		CreatedAt:     correctedAt.UTC(),
		UpdatedAt:     correctedAt.UTC(),
	}
	appliedOperations, err := correctionSupersessionOperation(target, replacement, correction)
	if err != nil {
		return nil, nil, nil, err
	}
	trace := &core.MemoryTrace{
		MemoryID:               replacement.ID,
		RawEventIDs:            []string{correction.RawEventID},
		ReasoningJobID:         reasoningJobID,
		ReasoningStage:         "operator_correction",
		CandidateSnapshotJSON:  json.RawMessage(`{"source":"operator_correction"}`),
		AppliedOperationsJSON:  appliedOperations,
		OperatorCorrectionFlag: true,
		RelatedDocumentIDs:     []string{},
		CreatedAt:              correctedAt.UTC(),
	}
	edge := &core.MemoryEdge{
		FromMemoryID:   replacement.ID,
		ToMemoryID:     target.ID,
		EdgeKind:       core.EdgeKindUpdates,
		Confidence:     1.0,
		CreatedByJobID: reasoningJobID,
		CreatedAt:      correctedAt.UTC(),
	}
	return replacement, trace, edge, nil
}

func correctionSupersessionID(targetMemoryID, idempotencyKey string) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{targetMemoryID, idempotencyKey}, "\x00")))
	return fmt.Sprintf("mem_corr_%x", sum[:12])
}

func correctionSupersessionFingerprint(target *core.Memory, correction *core.MemoryCorrection) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{
		target.TenantID,
		target.WorkspaceID,
		target.ID,
		string(target.Scope),
		target.OwnerEntityID,
		string(target.Kind),
		string(target.ArtifactClass),
		correction.IdempotencyKey,
		strings.TrimSpace(correction.CorrectionText),
	}, "\x00")))
	return fmt.Sprintf("fp_%x", sum[:16])
}

func correctionSupersessionMetadata(target *core.Memory, correction *core.MemoryCorrection) json.RawMessage {
	data, err := json.Marshal(struct {
		Source         string          `json:"source"`
		CorrectionID   string          `json:"correction_id"`
		TargetMemoryID string          `json:"target_memory_id"`
		OperatorID     string          `json:"operator_id"`
		EvidenceJSON   json.RawMessage `json:"evidence_json"`
	}{
		Source:         "operator_correction",
		CorrectionID:   correction.ID,
		TargetMemoryID: target.ID,
		OperatorID:     correction.OperatorID,
		EvidenceJSON:   jsonOrEmpty(correction.EvidenceJSON),
	})
	if err != nil {
		return json.RawMessage(`{"source":"operator_correction"}`)
	}
	return data
}

func correctionSupersessionOperation(target *core.Memory, replacement *core.Memory, correction *core.MemoryCorrection) (json.RawMessage, error) {
	data, err := json.Marshal([]struct {
		OperationID     string        `json:"operation_id"`
		Kind            string        `json:"kind"`
		MemoryID        string        `json:"memory_id"`
		TargetMemoryID  string        `json:"target_memory_id"`
		RawEventIDs     []string      `json:"raw_event_ids"`
		OperatorID      string        `json:"operator_id"`
		EdgeKind        core.EdgeKind `json:"edge_kind"`
		CorrectionID    string        `json:"correction_id"`
		CorrectionState string        `json:"correction_state"`
	}{{
		OperationID:     "operator_correction:" + correction.ID,
		Kind:            "update_memory",
		MemoryID:        replacement.ID,
		TargetMemoryID:  target.ID,
		RawEventIDs:     []string{correction.RawEventID},
		OperatorID:      correction.OperatorID,
		EdgeKind:        core.EdgeKindUpdates,
		CorrectionID:    correction.ID,
		CorrectionState: "applied",
	}})
	if err != nil {
		return nil, fmt.Errorf("marshal correction supersession operation: %w", err)
	}
	return json.RawMessage(data), nil
}
