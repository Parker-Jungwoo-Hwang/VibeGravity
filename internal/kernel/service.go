// ============================================================
// FILE     : internal/kernel/service.go
// PURPOSE  : Implements core.VibeGravityService by delegating to product services.
// LAYER    : application
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : Dependencies, Service, NewService
// DEPENDS  : encoding/json, errors, internal/core, internal/ingest, internal/recall, internal/store
// USED_BY  : cmd/server, tests, future Hermes and MCP adapters
// ------------------------------------------------------------
// AGENT_NOTE: Do not hide product behavior here; route calls to the package that owns the contract.
// ============================================================

package kernel

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
	"github.com/parker-jungwoo-hwang/vibegravity/internal/ingest"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/recall"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/store"
)

// Dependencies contains the product services composed by the kernel.
type Dependencies struct {
	Ingest      *ingest.Service
	Recall      *recall.Assembler
	Notes       store.NoteStore
	Plans       store.PlanStore
	Memories    store.MemoryStore
	Corrections store.CorrectionStore
	Jobs        store.CorrectionApplyJobStore
	Timeline    store.TimelineStore
	Documents   store.DocumentStore
}

// Service is the concrete v1 VibeGravity service.
type Service struct {
	ingest      *ingest.Service
	recall      *recall.Assembler
	notes       store.NoteStore
	plans       store.PlanStore
	memories    store.MemoryStore
	corrections store.CorrectionStore
	jobs        store.CorrectionApplyJobStore
	timeline    store.TimelineStore
	documents   store.DocumentStore
}

var _ core.VibeGravityService = (*Service)(nil)

// NewService creates the concrete VibeGravity service.
func NewService(deps Dependencies) (*Service, error) {
	if deps.Ingest == nil {
		return nil, fmt.Errorf("%w: ingest service is required", core.ErrInvalidArgument)
	}
	if deps.Recall == nil {
		return nil, fmt.Errorf("%w: recall assembler is required", core.ErrInvalidArgument)
	}
	return &Service{
		ingest:      deps.Ingest,
		recall:      deps.Recall,
		notes:       deps.Notes,
		plans:       deps.Plans,
		memories:    deps.Memories,
		corrections: deps.Corrections,
		jobs:        deps.Jobs,
		timeline:    deps.Timeline,
		documents:   deps.Documents,
	}, nil
}

// Prefetch assembles a next-turn recall pack.
func (s *Service) Prefetch(ctx context.Context, req *core.PrefetchRequest) (*core.PrefetchResponse, error) {
	return s.recall.Prefetch(ctx, req)
}

// SyncTurn records turn events on the hot path.
func (s *Service) SyncTurn(ctx context.Context, req *core.SyncTurnRequest) (*core.SyncTurnResponse, error) {
	return s.ingest.SyncTurn(ctx, req)
}

const documentChunkMaxRunes = 1800

// AddDocument stores a document and its initial lexical retrieval chunks.
func (s *Service) AddDocument(ctx context.Context, req *core.AddDocumentRequest) (*core.AddDocumentResponse, error) {
	if s.documents == nil {
		return nil, fmt.Errorf("%w: add document", core.ErrNotImplemented)
	}
	if req == nil {
		return nil, fmt.Errorf("%w: add document request is required", core.ErrInvalidArgument)
	}
	if err := requireFields(map[string]string{
		"tenant_id":    req.TenantID,
		"workspace_id": req.WorkspaceID,
		"source":       req.Source,
		"title":        req.Title,
		"content":      req.Content,
	}); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	document := &core.Document{
		TenantID:     req.TenantID,
		WorkspaceID:  req.WorkspaceID,
		Source:       req.Source,
		Title:        req.Title,
		Fingerprint:  valueOr(req.Fingerprint, documentFingerprint(req)),
		MetadataJSON: jsonOrEmpty(req.MetadataJSON),
		VersionHint:  req.VersionHint,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	chunks := buildDocumentChunks("", req.Content, now)
	if err := s.documents.AddDocumentWithChunks(ctx, document, chunks); err != nil {
		return nil, err
	}
	chunkIDs := make([]string, 0, len(chunks))
	for _, chunk := range chunks {
		chunkIDs = append(chunkIDs, chunk.ID)
	}
	return &core.AddDocumentResponse{DocumentID: document.ID, ChunkIDs: chunkIDs, Status: "created"}, nil
}

// SearchMemories delegates memory search to storage.
func (s *Service) SearchMemories(ctx context.Context, req *core.SearchMemoriesRequest) (*core.SearchMemoriesResponse, error) {
	if s.memories == nil {
		return nil, fmt.Errorf("%w: search memories", core.ErrNotImplemented)
	}
	if req == nil {
		return nil, fmt.Errorf("%w: search memories request is required", core.ErrInvalidArgument)
	}
	if err := requireFields(map[string]string{
		"tenant_id":    req.TenantID,
		"workspace_id": req.WorkspaceID,
	}); err != nil {
		return nil, err
	}
	if len(req.Scopes) == 0 {
		req.Scopes = []core.MemoryScope{
			core.MemoryScopeAgentPrivate,
			core.MemoryScopeWorkspaceShared,
			core.MemoryScopeSessionScratch,
		}
	}
	if len(req.ArtifactClasses) == 0 {
		req.ArtifactClasses = []core.ArtifactClass{
			core.ArtifactClassContext,
			core.ArtifactClassKnowledge,
			core.ArtifactClassTimeline,
			core.ArtifactClassPlan,
		}
	}
	return s.memories.SearchMemories(ctx, req)
}

// SearchDocuments delegates document search to storage.
func (s *Service) SearchDocuments(ctx context.Context, req *core.SearchDocumentsRequest) (*core.SearchDocumentsResponse, error) {
	if s.documents == nil {
		return nil, fmt.Errorf("%w: search documents", core.ErrNotImplemented)
	}
	if req == nil {
		return nil, fmt.Errorf("%w: search documents request is required", core.ErrInvalidArgument)
	}
	if err := requireFields(map[string]string{
		"tenant_id":    req.TenantID,
		"workspace_id": req.WorkspaceID,
	}); err != nil {
		return nil, err
	}
	return s.documents.SearchDocuments(ctx, req)
}

// AddNote creates a human-authored recall control note.
func (s *Service) AddNote(ctx context.Context, req *core.AddNoteRequest) (*core.AddNoteResponse, error) {
	if s.notes == nil {
		return nil, fmt.Errorf("%w: add note", core.ErrNotImplemented)
	}
	if req == nil {
		return nil, fmt.Errorf("%w: add note request is required", core.ErrInvalidArgument)
	}
	if err := requireFields(map[string]string{
		"tenant_id":       req.TenantID,
		"workspace_id":    req.WorkspaceID,
		"scope":           string(req.Scope),
		"owner_entity_id": req.OwnerEntityID,
		"text":            req.Text,
	}); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	note := &core.Note{
		TenantID:      req.TenantID,
		WorkspaceID:   req.WorkspaceID,
		NoteKind:      valueOr(req.NoteKind, "operator"),
		Scope:         req.Scope,
		OwnerEntityID: req.OwnerEntityID,
		Text:          req.Text,
		Pinned:        req.Pinned,
		ExpiresAt:     req.ExpiresAt,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := s.notes.AddNote(ctx, note); err != nil {
		return nil, err
	}
	return &core.AddNoteResponse{NoteID: note.ID, Status: "created"}, nil
}

// CreatePlan creates a structured plan and its initial items.
func (s *Service) CreatePlan(ctx context.Context, req *core.CreatePlanRequest) (*core.CreatePlanResponse, error) {
	if s.plans == nil {
		return nil, fmt.Errorf("%w: create plan", core.ErrNotImplemented)
	}
	if req == nil {
		return nil, fmt.Errorf("%w: create plan request is required", core.ErrInvalidArgument)
	}
	if err := requireFields(map[string]string{
		"tenant_id":       req.TenantID,
		"workspace_id":    req.WorkspaceID,
		"title":           req.Title,
		"scope":           string(req.Scope),
		"owner_entity_id": req.OwnerEntityID,
	}); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	plan := &core.Plan{
		TenantID:      req.TenantID,
		WorkspaceID:   req.WorkspaceID,
		Title:         req.Title,
		Status:        valueOr(req.Status, "active"),
		Scope:         req.Scope,
		OwnerEntityID: req.OwnerEntityID,
		EvidenceJSON:  jsonOrEmpty(req.EvidenceJSON),
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	items := make([]*core.PlanItem, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, &core.PlanItem{
			ID:           item.ID,
			Title:        item.Title,
			Status:       valueOr(item.Status, "open"),
			EvidenceJSON: jsonOrEmpty(item.EvidenceJSON),
			CreatedAt:    now,
			UpdatedAt:    now,
		})
	}
	if err := s.plans.CreatePlan(ctx, plan, items); err != nil {
		return nil, err
	}
	itemIDs := make([]string, 0, len(items))
	for _, item := range items {
		itemIDs = append(itemIDs, item.ID)
	}
	return &core.CreatePlanResponse{PlanID: plan.ID, ItemIDs: itemIDs, Status: "created"}, nil
}

// UpdatePlan updates a structured plan and optionally replaces provided items.
func (s *Service) UpdatePlan(ctx context.Context, req *core.UpdatePlanRequest) (*core.UpdatePlanResponse, error) {
	if s.plans == nil {
		return nil, fmt.Errorf("%w: update plan", core.ErrNotImplemented)
	}
	if req == nil {
		return nil, fmt.Errorf("%w: update plan request is required", core.ErrInvalidArgument)
	}
	if err := requireFields(map[string]string{
		"tenant_id":    req.TenantID,
		"workspace_id": req.WorkspaceID,
		"plan_id":      req.PlanID,
	}); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	plan := &core.Plan{
		ID:           req.PlanID,
		TenantID:     req.TenantID,
		WorkspaceID:  req.WorkspaceID,
		EvidenceJSON: req.EvidenceJSON,
		UpdatedAt:    now,
	}
	if req.Title != nil {
		plan.Title = strings.TrimSpace(*req.Title)
		if plan.Title == "" {
			return nil, fmt.Errorf("%w: title cannot be empty", core.ErrInvalidArgument)
		}
	}
	if req.Status != nil {
		plan.Status = strings.TrimSpace(*req.Status)
		if plan.Status == "" {
			return nil, fmt.Errorf("%w: status cannot be empty", core.ErrInvalidArgument)
		}
	}
	items := make([]*core.PlanItem, 0, len(req.Items))
	if req.Items != nil {
		for _, item := range req.Items {
			title := strings.TrimSpace(item.Title)
			if title == "" {
				return nil, fmt.Errorf("%w: plan item title is required", core.ErrInvalidArgument)
			}
			items = append(items, &core.PlanItem{
				ID:           item.ID,
				Title:        title,
				Status:       valueOr(item.Status, "open"),
				EvidenceJSON: jsonOrEmpty(item.EvidenceJSON),
				CreatedAt:    now,
				UpdatedAt:    now,
			})
		}
	} else {
		items = nil
	}
	if err := s.plans.UpdatePlan(ctx, plan, items); err != nil {
		return nil, err
	}
	return &core.UpdatePlanResponse{PlanID: req.PlanID, Status: "updated"}, nil
}

// CorrectMemory records human correction intent and applies an operator-driven supersession.
func (s *Service) CorrectMemory(ctx context.Context, req *core.CorrectMemoryRequest) (*core.CorrectMemoryResponse, error) {
	if s.memories == nil || s.corrections == nil || s.jobs == nil {
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
		return s.applyRecordedCorrection(ctx, req, memory, recorded, time.Now().UTC())
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
	if memory.Status != core.MemoryStatusActive || !memory.LatestFlag {
		return nil, fmt.Errorf("%w: correction target memory must be active latest", core.ErrConflict)
	}
	now := time.Now().UTC()
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
	supersessions, ok := s.memories.(correctionSupersessionStore)
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

type correctionSupersessionStore interface {
	CreateCorrectionSupersession(ctx context.Context, memory *core.Memory, trace *core.MemoryTrace, edge *core.MemoryEdge, correctionID string) error
}

const (
	timelineDefaultLimit = 50
	timelineMaxLimit     = 100
)

// GetTimeline assembles a read-only operator timeline over existing artifacts.
func (s *Service) GetTimeline(ctx context.Context, req *core.GetTimelineRequest) (*core.GetTimelineResponse, error) {
	if s.timeline == nil {
		return nil, fmt.Errorf("%w: get timeline", core.ErrNotImplemented)
	}
	if req == nil {
		return nil, fmt.Errorf("%w: get timeline request is required", core.ErrInvalidArgument)
	}
	if err := requireFields(map[string]string{
		"tenant_id":    req.TenantID,
		"workspace_id": req.WorkspaceID,
		"entity_id":    req.EntityID,
	}); err != nil {
		return nil, err
	}
	if req.From != nil && req.To != nil && req.From.After(*req.To) {
		return nil, fmt.Errorf("%w: from must be before to", core.ErrInvalidArgument)
	}
	normalized := *req
	scopes, err := normalizeTimelineScopes(req.Scopes)
	if err != nil {
		return nil, err
	}
	normalized.Scopes = scopes
	if normalized.Limit == 0 {
		normalized.Limit = timelineDefaultLimit
	}
	if normalized.Limit < 0 || normalized.Limit > timelineMaxLimit {
		return nil, fmt.Errorf("%w: limit must be between 1 and %d", core.ErrInvalidArgument, timelineMaxLimit)
	}
	return s.timeline.GetTimeline(ctx, &normalized)
}

// ExplainMemory delegates provenance lookup to storage.
func (s *Service) ExplainMemory(ctx context.Context, req *core.ExplainMemoryRequest) (*core.ExplainMemoryResponse, error) {
	if s.memories == nil {
		return nil, fmt.Errorf("%w: explain memory", core.ErrNotImplemented)
	}
	if req == nil {
		return nil, fmt.Errorf("%w: explain memory request is required", core.ErrInvalidArgument)
	}
	if err := requireFields(map[string]string{
		"tenant_id":    req.TenantID,
		"workspace_id": req.WorkspaceID,
		"memory_id":    req.MemoryID,
	}); err != nil {
		return nil, err
	}
	return s.memories.ExplainMemory(ctx, req)
}

func requireFields(fields map[string]string) error {
	for name, value := range fields {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%w: %s is required", core.ErrInvalidArgument, name)
		}
	}
	return nil
}

func jsonOrEmpty(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 {
		return json.RawMessage(`{}`)
	}
	return raw
}

func valueOr(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func documentFingerprint(req *core.AddDocumentRequest) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{
		req.TenantID,
		req.WorkspaceID,
		req.Source,
		req.Title,
		req.Content,
	}, "\x00")))
	return fmt.Sprintf("sha256:%x", sum[:])
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

func normalizeTimelineScopes(scopes []core.MemoryScope) ([]core.MemoryScope, error) {
	if len(scopes) == 0 {
		return []core.MemoryScope{
			core.MemoryScopeAgentPrivate,
			core.MemoryScopeWorkspaceShared,
			core.MemoryScopeSessionScratch,
		}, nil
	}
	seen := make(map[core.MemoryScope]struct{}, len(scopes))
	normalized := make([]core.MemoryScope, 0, len(scopes))
	for _, scope := range scopes {
		switch scope {
		case core.MemoryScopeAgentPrivate, core.MemoryScopeWorkspaceShared, core.MemoryScopeSessionScratch:
			if _, ok := seen[scope]; !ok {
				seen[scope] = struct{}{}
				normalized = append(normalized, scope)
			}
		case core.MemoryScopeGroupShared:
			continue
		default:
			return nil, fmt.Errorf("%w: unsupported timeline scope %q", core.ErrInvalidArgument, scope)
		}
	}
	return normalized, nil
}

func buildDocumentChunks(documentID, content string, now time.Time) []*core.DocumentChunk {
	paragraphs := strings.Split(strings.TrimSpace(content), "\n\n")
	chunks := make([]*core.DocumentChunk, 0, len(paragraphs))
	var builder strings.Builder
	flush := func() {
		text := strings.TrimSpace(builder.String())
		if text == "" {
			builder.Reset()
			return
		}
		chunks = append(chunks, &core.DocumentChunk{
			DocumentID:     documentID,
			ChunkIndex:     len(chunks),
			Text:           text,
			MetadataJSON:   json.RawMessage(`{}`),
			EmbeddingModel: "pending",
			EmbeddingDims:  0,
			CreatedAt:      now,
			UpdatedAt:      now,
		})
		builder.Reset()
	}
	for _, paragraph := range paragraphs {
		paragraph = strings.TrimSpace(paragraph)
		if paragraph == "" {
			continue
		}
		if builder.Len() > 0 && builder.Len()+len(paragraph)+2 > documentChunkMaxRunes {
			flush()
		}
		if len([]rune(paragraph)) > documentChunkMaxRunes {
			flush()
			for _, part := range splitRunes(paragraph, documentChunkMaxRunes) {
				builder.WriteString(part)
				flush()
			}
			continue
		}
		if builder.Len() > 0 {
			builder.WriteString("\n\n")
		}
		builder.WriteString(paragraph)
	}
	flush()
	return chunks
}

func splitRunes(text string, maxRunes int) []string {
	runes := []rune(text)
	parts := make([]string, 0, (len(runes)/maxRunes)+1)
	for start := 0; start < len(runes); start += maxRunes {
		end := start + maxRunes
		if end > len(runes) {
			end = len(runes)
		}
		parts = append(parts, string(runes[start:end]))
	}
	return parts
}
