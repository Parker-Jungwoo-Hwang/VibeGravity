// ============================================================
// FILE     : internal/worker/processor.go
// PURPOSE  : Claims worker jobs and runs turn-processing and dreaming background pipelines.
// LAYER    : application
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : Dependencies, Processor, RunResult, JobFailure, NewProcessor
// DEPENDS  : context, errors, fmt, internal/core, internal/graph, internal/reasoning, internal/store
// USED_BY  : cmd/worker, internal/worker tests
// ------------------------------------------------------------
// AGENT_NOTE: Keep this as orchestration only; do not add local text extraction logic here.
// ============================================================

package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/graph"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/reasoning"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/store"
)

const defaultBatchSize = 1

var errUnsupportedApplyWork = errors.New("unsupported apply work")

// Dependencies collects the stores and application services used by the worker processor.
type Dependencies struct {
	WorkerID    string
	BatchSize   int
	Jobs        store.JobStore
	RawEvents   store.RawEventStore
	Reasoner    reasoning.Orchestrator
	ApplyEngine graph.ApplyEngine
	Dreaming    *graph.DreamingService
	Clock       func() time.Time
}

// Processor claims jobs and dispatches them to the correct worker pipeline.
type Processor struct {
	workerID    string
	batchSize   int
	jobs        store.JobStore
	rawEvents   store.RawEventStore
	reasoner    reasoning.Orchestrator
	applyEngine graph.ApplyEngine
	dreaming    *graph.DreamingService
	clock       func() time.Time
}

// RunResult summarizes one processor polling pass.
type RunResult struct {
	Claimed               int
	Completed             int
	Failed                int
	Blocked               int
	AppliedOperationCount int
	MemoryIDCount         int
	TraceWrittenCount     int
	SessionDreamCount     int
	WorkspaceDreamCount   int
	Failures              []JobFailure
}

// JobFailure describes one job failure in a processor polling pass.
type JobFailure struct {
	JobID   string
	JobKind core.JobKind
	Error   string
}

type jobProcessResult struct {
	AppliedOperationCount int
	MemoryIDCount         int
	TraceWrittenCount     int
	SessionDreamCount     int
	WorkspaceDreamCount   int
}

// NewProcessor builds a worker processor.
func NewProcessor(deps Dependencies) (*Processor, error) {
	if deps.Jobs == nil {
		return nil, fmt.Errorf("%w: worker job store is required", core.ErrInvalidArgument)
	}
	if deps.RawEvents == nil {
		return nil, fmt.Errorf("%w: worker raw event store is required", core.ErrInvalidArgument)
	}
	if deps.Reasoner == nil {
		return nil, fmt.Errorf("%w: worker reasoning orchestrator is required", core.ErrInvalidArgument)
	}
	if deps.ApplyEngine == nil {
		return nil, fmt.Errorf("%w: worker apply engine is required", core.ErrInvalidArgument)
	}
	if deps.Dreaming == nil {
		return nil, fmt.Errorf("%w: worker dreaming service is required", core.ErrInvalidArgument)
	}
	batchSize := deps.BatchSize
	if batchSize <= 0 {
		batchSize = defaultBatchSize
	}
	workerID := deps.WorkerID
	if workerID == "" {
		workerID = "worker:default"
	}
	clock := deps.Clock
	if clock == nil {
		clock = time.Now
	}
	return &Processor{
		workerID:    workerID,
		batchSize:   batchSize,
		jobs:        deps.Jobs,
		rawEvents:   deps.RawEvents,
		reasoner:    deps.Reasoner,
		applyEngine: deps.ApplyEngine,
		dreaming:    deps.Dreaming,
		clock:       clock,
	}, nil
}

// RunOnce claims up to the configured batch size and processes each claimed job.
func (p *Processor) RunOnce(ctx context.Context) (RunResult, error) {
	jobs, err := p.jobs.ClaimJobs(ctx, p.workerID, p.batchSize)
	if err != nil {
		return RunResult{}, fmt.Errorf("claim jobs: %w", err)
	}
	result := RunResult{
		Claimed: len(jobs),
	}
	var firstErr error
	for _, job := range jobs {
		if job == nil {
			continue
		}
		jobResult, err := p.processClaimedJob(ctx, job)
		if err != nil {
			if isPermanentJobFailure(err) {
				result.Blocked++
			} else {
				result.Failed++
			}
			result.Failures = append(result.Failures, newJobFailure(job, err))
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		result.Completed++
		result.AppliedOperationCount += jobResult.AppliedOperationCount
		result.MemoryIDCount += jobResult.MemoryIDCount
		result.TraceWrittenCount += jobResult.TraceWrittenCount
		result.SessionDreamCount += jobResult.SessionDreamCount
		result.WorkspaceDreamCount += jobResult.WorkspaceDreamCount
	}
	return result, firstErr
}

func (p *Processor) processClaimedJob(ctx context.Context, job *core.IngestJob) (jobProcessResult, error) {
	jobResult, err := p.handleJob(ctx, job)
	if err != nil {
		failureErr := jobFailureError(job, err)
		if isPermanentJobFailure(err) {
			if blockErr := p.jobs.BlockJob(ctx, job.ID, failureErr); blockErr != nil {
				return jobProcessResult{}, fmt.Errorf("process job %s: %w; block job: %v", job.ID, failureErr, blockErr)
			}
			return jobProcessResult{}, fmt.Errorf("process job %s: %w", job.ID, failureErr)
		}
		if failErr := p.jobs.FailJob(ctx, job.ID, failureErr); failErr != nil {
			return jobProcessResult{}, fmt.Errorf("process job %s: %w; record failure: %v", job.ID, failureErr, failErr)
		}
		return jobProcessResult{}, fmt.Errorf("process job %s: %w", job.ID, failureErr)
	}
	if err := p.jobs.CompleteJob(ctx, job.ID); err != nil {
		return jobProcessResult{}, fmt.Errorf("complete job %s: %w", job.ID, err)
	}
	return jobResult, nil
}

func (p *Processor) handleJob(ctx context.Context, job *core.IngestJob) (jobProcessResult, error) {
	switch job.JobKind {
	case core.JobKindProcessTurnEvent:
		return p.processTurnEvent(ctx, job)
	case core.JobKindDreamSession:
		return p.dreamSession(ctx, job)
	case core.JobKindDreamWorkspace:
		return p.dreamWorkspace(ctx, job)
	default:
		return jobProcessResult{}, fmt.Errorf("%w: unsupported job kind %q", core.ErrInvalidArgument, job.JobKind)
	}
}

func (p *Processor) processTurnEvent(ctx context.Context, job *core.IngestJob) (jobProcessResult, error) {
	if len(job.RawEventIDs) == 0 {
		return jobProcessResult{}, fmt.Errorf("%w: process_turn_event requires raw_event_ids", core.ErrInvalidArgument)
	}
	events, err := p.rawEvents.GetRawEvents(ctx, job.RawEventIDs)
	if err != nil {
		return jobProcessResult{}, fmt.Errorf("load raw event bundle: %w", err)
	}
	events, err = validateAndOrderRawEventBundle(job, events)
	if err != nil {
		return jobProcessResult{}, err
	}

	envelope := p.buildProcessTurnEnvelope(ctx, job, events)
	reasoningResult, err := p.reasoner.ProcessTurn(ctx, envelope)
	if err != nil {
		return jobProcessResult{}, fmt.Errorf("reason process_turn_event: %w", err)
	}
	applyResult, err := p.applyEngine.Apply(ctx, &graph.ApplyRequest{
		JobID:       job.ID,
		TenantID:    job.TenantID,
		WorkspaceID: job.WorkspaceID,
		RawEventIDs: append([]string(nil), job.RawEventIDs...),
		Reasoning:   reasoningResult,
	})
	if err != nil {
		if errors.Is(err, core.ErrNotImplemented) {
			return jobProcessResult{}, fmt.Errorf("%w: apply process_turn_event: %w", errUnsupportedApplyWork, err)
		}
		return jobProcessResult{}, fmt.Errorf("apply process_turn_event: %w", err)
	}
	return jobProcessResultFromApply(applyResult)
}

func (p *Processor) dreamSession(ctx context.Context, job *core.IngestJob) (jobProcessResult, error) {
	payload, err := decodeDreamingJobPayload(job)
	if err != nil {
		return jobProcessResult{}, err
	}
	if strings.TrimSpace(payload.SessionID) == "" {
		return jobProcessResult{}, fmt.Errorf("%w: dream_session payload.session_id is required", core.ErrInvalidArgument)
	}
	result, err := p.dreaming.DreamSession(ctx, &core.DreamSessionRequest{
		JobID:       job.ID,
		TenantID:    job.TenantID,
		WorkspaceID: job.WorkspaceID,
		SessionID:   payload.SessionID,
		Now:         p.clock().UTC(),
	})
	if err != nil {
		return jobProcessResult{}, fmt.Errorf("dream session: %w", err)
	}
	return jobProcessResult{
		SessionDreamCount: result.MidTermPromoted + boolCount(result.SessionSummaryWritten),
	}, nil
}

func (p *Processor) dreamWorkspace(ctx context.Context, job *core.IngestJob) (jobProcessResult, error) {
	result, err := p.dreaming.DreamWorkspace(ctx, &core.DreamWorkspaceRequest{
		JobID:       job.ID,
		TenantID:    job.TenantID,
		WorkspaceID: job.WorkspaceID,
		Now:         p.clock().UTC(),
	})
	if err != nil {
		return jobProcessResult{}, fmt.Errorf("dream workspace: %w", err)
	}
	return jobProcessResult{
		WorkspaceDreamCount: result.LongTermPromoted + result.UltraLongTermPromoted,
	}, nil
}

type dreamingJobPayload struct {
	SessionID string `json:"session_id"`
}

func decodeDreamingJobPayload(job *core.IngestJob) (*dreamingJobPayload, error) {
	payload := &dreamingJobPayload{}
	if len(job.PayloadJSON) == 0 {
		return payload, nil
	}
	if err := json.Unmarshal(job.PayloadJSON, payload); err != nil {
		return nil, fmt.Errorf("%w: dreaming job payload_json must be valid JSON", core.ErrInvalidArgument)
	}
	return payload, nil
}

func boolCount(value bool) int {
	if value {
		return 1
	}
	return 0
}

func jobProcessResultFromApply(result *graph.ApplyResult) (jobProcessResult, error) {
	if result == nil {
		return jobProcessResult{}, fmt.Errorf("%w: apply process_turn_event returned nil result", core.ErrInvalidArgument)
	}
	if result.AppliedOperationCount > 0 && !result.TraceWritten {
		return jobProcessResult{}, fmt.Errorf("%w: apply result reports applied operations without a trace", core.ErrConflict)
	}
	traceWrittenCount := 0
	if result.TraceWritten {
		traceWrittenCount = 1
	}
	return jobProcessResult{
		AppliedOperationCount: result.AppliedOperationCount,
		MemoryIDCount:         len(result.MemoryIDs),
		TraceWrittenCount:     traceWrittenCount,
	}, nil
}

func validateAndOrderRawEventBundle(job *core.IngestJob, events []*core.RawEvent) ([]*core.RawEvent, error) {
	if len(events) != len(job.RawEventIDs) {
		return nil, fmt.Errorf("%w: raw event bundle incomplete for job %s: expected %d events, got %d", core.ErrNotFound, job.ID, len(job.RawEventIDs), len(events))
	}

	expected := make(map[string]struct{}, len(job.RawEventIDs))
	for _, id := range job.RawEventIDs {
		if id == "" {
			return nil, fmt.Errorf("%w: raw event bundle contains empty raw_event_id for job %s", core.ErrInvalidArgument, job.ID)
		}
		if _, exists := expected[id]; exists {
			return nil, fmt.Errorf("%w: raw event bundle contains duplicate raw_event_id %q for job %s", core.ErrInvalidArgument, id, job.ID)
		}
		expected[id] = struct{}{}
	}

	byID := make(map[string]*core.RawEvent, len(events))
	for i, event := range events {
		if event == nil {
			return nil, fmt.Errorf("%w: raw event bundle incomplete for job %s: nil event at index %d", core.ErrNotFound, job.ID, i)
		}
		if _, ok := expected[event.ID]; !ok {
			return nil, fmt.Errorf("%w: raw event bundle contains unexpected raw_event_id %q for job %s", core.ErrConflict, event.ID, job.ID)
		}
		if _, exists := byID[event.ID]; exists {
			return nil, fmt.Errorf("%w: raw event bundle contains duplicate returned raw_event_id %q for job %s", core.ErrConflict, event.ID, job.ID)
		}
		if event.TenantID != job.TenantID {
			return nil, fmt.Errorf("%w: raw event bundle tenant mismatch for raw_event_id %q", core.ErrConflict, event.ID)
		}
		if event.WorkspaceID != job.WorkspaceID {
			return nil, fmt.Errorf("%w: raw event bundle workspace mismatch for raw_event_id %q", core.ErrConflict, event.ID)
		}
		byID[event.ID] = event
	}

	ordered := make([]*core.RawEvent, 0, len(job.RawEventIDs))
	for _, id := range job.RawEventIDs {
		event, ok := byID[id]
		if !ok {
			return nil, fmt.Errorf("%w: raw event bundle incomplete for job %s: missing raw_event_id %q", core.ErrNotFound, job.ID, id)
		}
		ordered = append(ordered, event)
	}
	if _, err := requireSingleRawEventActor(job, ordered); err != nil {
		return nil, err
	}
	return ordered, nil
}

func requireSingleRawEventActor(job *core.IngestJob, events []*core.RawEvent) (string, error) {
	var actorID string
	for _, event := range events {
		currentActorID := strings.TrimSpace(event.ActorID)
		if currentActorID == "" {
			return "", fmt.Errorf("%w: raw event bundle contains empty actor_id for raw_event_id %q in job %s", core.ErrInvalidArgument, event.ID, job.ID)
		}
		if actorID == "" {
			actorID = currentActorID
			continue
		}
		if currentActorID != actorID {
			return "", fmt.Errorf("%w: raw event bundle contains mixed actor_id values for job %s", core.ErrConflict, job.ID)
		}
	}
	if actorID == "" {
		return "", fmt.Errorf("%w: raw event bundle has no actor_id for job %s", core.ErrInvalidArgument, job.ID)
	}
	return actorID, nil
}

func newJobFailure(job *core.IngestJob, err error) JobFailure {
	failure := JobFailure{}
	if job != nil {
		failure.JobID = job.ID
		failure.JobKind = job.JobKind
	}
	if err != nil {
		failure.Error = err.Error()
	}
	return failure
}

func jobFailureError(job *core.IngestJob, err error) error {
	if job == nil {
		return fmt.Errorf("worker job failed: %w", err)
	}
	return fmt.Errorf("worker job failed job_id=%s job_kind=%s raw_event_count=%d: %w", job.ID, job.JobKind, len(job.RawEventIDs), err)
}

func isPermanentJobFailure(err error) bool {
	return errors.Is(err, errUnsupportedApplyWork)
}

func (p *Processor) buildProcessTurnEnvelope(_ context.Context, job *core.IngestJob, events []*core.RawEvent) *reasoning.ProcessTurnEnvelope {
	rawEventIDs := append([]string(nil), job.RawEventIDs...)
	rawEvents := append([]*core.RawEvent(nil), events...)
	stage1 := reasoning.Stage1Input{
		JobID:       job.ID,
		TenantID:    job.TenantID,
		WorkspaceID: job.WorkspaceID,
		RawEvents:   rawEvents,
	}
	return &reasoning.ProcessTurnEnvelope{
		JobID:       job.ID,
		TenantID:    job.TenantID,
		WorkspaceID: job.WorkspaceID,
		RawEventIDs: rawEventIDs,
		RawEvents:   rawEvents,
		Stage1:      stage1,
		Stage2: reasoning.Stage2Input{
			JobID:                job.ID,
			TenantID:             job.TenantID,
			WorkspaceID:          job.WorkspaceID,
			RawEvents:            rawEvents,
			RelevantMemories:     []core.MemoryResult{},
			RelevantDocuments:    []core.DocumentChunkResult{},
			ActivePlans:          []*core.Plan{},
			PinnedNotes:          []*core.Note{},
			RequiredOutputName:   reasoning.StageNameResolve,
			RequiredOutputSchema: reasoning.Stage2ResolveOutputSchemaV0,
		},
	}
}
