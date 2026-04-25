// ============================================================
// FILE     : internal/ingest/service.go
// PURPOSE  : Implements the sync_turn hot path for raw events and worker jobs.
// LAYER    : application
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : Dependencies, Service, NewService
// DEPENDS  : internal/core, internal/store
// USED_BY  : internal/kernel, tests
// ------------------------------------------------------------
// AGENT_NOTE: This path must not perform reasoning, graph updates, profile updates, or dreaming.
// ============================================================

package ingest

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/store"
)

const defaultSyncSource = "sync_turn"

// Dependencies collects stores and runtime hooks needed by the ingest service.
type Dependencies struct {
	RawEvents store.RawEventStore
	Jobs      store.JobStore
	Clock     func() time.Time
}

// Service owns the sync_turn hot path.
type Service struct {
	rawEvents store.RawEventStore
	jobs      store.JobStore
	clock     func() time.Time
}

// NewService builds an ingest service.
func NewService(deps Dependencies) (*Service, error) {
	if deps.RawEvents == nil {
		return nil, fmt.Errorf("%w: ingest raw event store is required", core.ErrInvalidArgument)
	}
	if deps.Jobs == nil {
		return nil, fmt.Errorf("%w: ingest job store is required", core.ErrInvalidArgument)
	}
	clock := deps.Clock
	if clock == nil {
		clock = time.Now
	}
	return &Service{
		rawEvents: deps.RawEvents,
		jobs:      deps.Jobs,
		clock:     clock,
	}, nil
}

// SyncTurn records raw turn events and enqueues one process_turn_event job.
func (s *Service) SyncTurn(ctx context.Context, req *core.SyncTurnRequest) (*core.SyncTurnResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("%w: sync turn request is required", core.ErrInvalidArgument)
	}
	if err := validateSyncTurnRequest(req); err != nil {
		return nil, err
	}

	now := s.clock().UTC()
	events, err := s.buildRawEvents(req, now)
	if err != nil {
		return nil, err
	}

	eventIDs, err := s.rawEvents.AppendRawEvents(ctx, events)
	if err != nil {
		if errors.Is(err, core.ErrDuplicate) {
			return &core.SyncTurnResponse{
				Status:         "accepted",
				SessionID:      req.SessionID,
				EventIDs:       nil,
				JobIDs:         nil,
				DuplicateCount: len(events),
			}, nil
		}
		return nil, fmt.Errorf("append raw events: %w", err)
	}
	if len(eventIDs) > len(events) {
		return nil, fmt.Errorf("%w: raw event store returned too many ids", core.ErrConflict)
	}

	jobIDs := make([]string, 0, 1)
	if len(eventIDs) > 0 {
		payload, err := json.Marshal(map[string]string{
			"session_id": req.SessionID,
			"actor_id":   req.ActorID,
			"source":     defaultSyncSource,
		})
		if err != nil {
			return nil, fmt.Errorf("encode ingest job payload: %w", err)
		}
		jobs := []*core.IngestJob{{
			ID:          stableID("job", req.TenantID, req.WorkspaceID, req.SessionID, req.IdempotencyKey, strings.Join(eventIDs, ",")),
			TenantID:    req.TenantID,
			WorkspaceID: req.WorkspaceID,
			JobKind:     core.JobKindProcessTurnEvent,
			Status:      "queued",
			RawEventIDs: eventIDs,
			PayloadJSON: payload,
			Attempts:    0,
			AvailableAt: now,
			CreatedAt:   now,
			UpdatedAt:   now,
		}}
		jobIDs, err = s.jobs.EnqueueJobs(ctx, jobs)
		if err != nil {
			return nil, fmt.Errorf("enqueue ingest jobs: %w", err)
		}
	}

	return &core.SyncTurnResponse{
		Status:         "accepted",
		SessionID:      req.SessionID,
		EventIDs:       eventIDs,
		JobIDs:         jobIDs,
		DuplicateCount: len(events) - len(eventIDs),
	}, nil
}

func validateSyncTurnRequest(req *core.SyncTurnRequest) error {
	required := map[string]string{
		"tenant_id":       req.TenantID,
		"workspace_id":    req.WorkspaceID,
		"session_id":      req.SessionID,
		"actor_id":        req.ActorID,
		"idempotency_key": req.IdempotencyKey,
	}
	for name, value := range required {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%w: %s is required", core.ErrInvalidArgument, name)
		}
	}
	if len(req.TurnEvents) == 0 {
		return fmt.Errorf("%w: at least one turn event is required", core.ErrInvalidArgument)
	}
	for i, event := range req.TurnEvents {
		if strings.TrimSpace(event.EventKind) == "" {
			return fmt.Errorf("%w: turn_events[%d].event_kind is required", core.ErrInvalidArgument, i)
		}
		if len(normalizePayload(event.PayloadJSON)) > 0 && !json.Valid(normalizePayload(event.PayloadJSON)) {
			return fmt.Errorf("%w: turn_events[%d].payload_json must be valid JSON", core.ErrInvalidArgument, i)
		}
	}
	return nil
}

func (s *Service) buildRawEvents(req *core.SyncTurnRequest, now time.Time) ([]*core.RawEvent, error) {
	events := make([]*core.RawEvent, 0, len(req.TurnEvents))
	for i, event := range req.TurnEvents {
		payload := normalizePayload(event.PayloadJSON)
		if !json.Valid(payload) {
			return nil, fmt.Errorf("%w: turn_events[%d].payload_json must be valid JSON", core.ErrInvalidArgument, i)
		}
		source := strings.TrimSpace(event.Source)
		if source == "" {
			source = defaultSyncSource
		}
		occurredAt := event.OccurredAt
		if occurredAt.IsZero() {
			occurredAt = now
		}
		fingerprint := strings.TrimSpace(event.Fingerprint)
		if fingerprint == "" {
			fingerprint = stableID("fp", req.TenantID, req.WorkspaceID, req.SessionID, event.EventKind, source, string(payload))
		}
		idempotencyKey := fmt.Sprintf("%s:%03d:%s", req.IdempotencyKey, i, fingerprint)
		events = append(events, &core.RawEvent{
			ID:             stableID("evt", req.TenantID, source, idempotencyKey),
			TenantID:       req.TenantID,
			WorkspaceID:    req.WorkspaceID,
			SessionID:      req.SessionID,
			ActorID:        req.ActorID,
			EventKind:      event.EventKind,
			Source:         source,
			IdempotencyKey: idempotencyKey,
			Fingerprint:    fingerprint,
			OccurredAt:     occurredAt.UTC(),
			PayloadJSON:    append(json.RawMessage(nil), payload...),
			CreatedAt:      now,
		})
	}
	return events, nil
}

func normalizePayload(payload json.RawMessage) json.RawMessage {
	if len(payload) == 0 {
		return json.RawMessage(`{}`)
	}
	return append(json.RawMessage(nil), payload...)
}

func stableID(prefix string, parts ...string) string {
	h := sha256.New()
	for _, part := range parts {
		_, _ = h.Write([]byte(part))
		_, _ = h.Write([]byte{0})
	}
	sum := h.Sum(nil)
	return prefix + "_" + hex.EncodeToString(sum)[:24]
}
