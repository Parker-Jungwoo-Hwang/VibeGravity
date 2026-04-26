// ============================================================
// FILE     : internal/runtime/worker.go
// PURPOSE  : Composes worker processor dependencies and mocked Codex bridge selection.
// LAYER    : infra
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : WorkerOptions, OpenWorkerProcessor, NewReasoner
// DEPENDS  : internal/config, internal/db, internal/graph, internal/reasoning, internal/store/postgres, internal/worker
// USED_BY  : cmd/worker, runtime tests
// ------------------------------------------------------------
// AGENT_NOTE: Real Codex must stay explicit opt-in; the default worker uses MockCodexJSONClient.
// ============================================================

package runtime

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/config"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/db"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/graph"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/reasoning"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/store/postgres"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/worker"
)

// Logger is the subset of log.Logger used by runtime composition.
type Logger interface {
	Printf(format string, v ...any)
}

// WorkerOptions controls process-local worker composition.
type WorkerOptions struct {
	WorkerID string
	Logger   Logger
}

// OpenWorkerProcessor opens PostgreSQL and composes the background worker.
func OpenWorkerProcessor(ctx context.Context, cfg config.Config, opts WorkerOptions) (*worker.Processor, func(), error) {
	pool, err := db.NewPool(ctx, cfg)
	if err != nil {
		return nil, func() {}, err
	}
	pgStore := postgres.NewStore(pool)
	applyEngine, err := graph.NewStoreBackedApplyEngine(pgStore)
	if err != nil {
		pool.Close()
		return nil, func() {}, err
	}
	dreamingService, err := graph.NewDreamingService(graph.DreamingDependencies{
		Store: pgStore,
	})
	if err != nil {
		pool.Close()
		return nil, func() {}, err
	}
	stage2InputPreparer := worker.NewStoreBackedStage2InputPreparer(worker.Stage2SourceStores{
		Profiles:  pgStore,
		Memories:  pgStore,
		Documents: pgStore,
		Plans:     pgStore,
		Notes:     pgStore,
		Groups:    pgStore,
	})
	reasoner, err := NewReasoner(cfg.Codex, stage2InputPreparer, opts.Logger)
	if err != nil {
		pool.Close()
		return nil, func() {}, err
	}
	processor, err := worker.NewProcessor(worker.Dependencies{
		WorkerID:    opts.WorkerID,
		Jobs:        pgStore,
		RawEvents:   pgStore,
		Reasoner:    reasoner,
		ApplyEngine: applyEngine,
		Dreaming:    dreamingService,
	})
	if err != nil {
		pool.Close()
		return nil, func() {}, err
	}
	return processor, pool.Close, nil
}

// NewReasoner builds the worker reasoning orchestrator.
func NewReasoner(cfg config.CodexConfig, stage2InputPreparer *reasoning.Stage2InputPreparer, logger Logger) (reasoning.Orchestrator, error) {
	if logger == nil {
		logger = log.Default()
	}
	clientMode := strings.TrimSpace(cfg.ClientMode)
	if clientMode == "" {
		clientMode = config.CodexClientModeMock
	}
	if clientMode == config.CodexClientModeReal {
		if !cfg.Enabled || strings.TrimSpace(cfg.Endpoint) == "" || strings.TrimSpace(cfg.Model) == "" {
			return nil, fmt.Errorf("%w: real Codex requires VIBEGRAVITY_CODEX_ENABLED=true, VIBEGRAVITY_CODEX_CLIENT=real, VIBEGRAVITY_CODEX_ENDPOINT, and VIBEGRAVITY_CODEX_MODEL", core.ErrInvalidArgument)
		}
		return nil, fmt.Errorf("%w: real Codex client is not implemented in this runtime slice", core.ErrNotImplemented)
	}
	logger.Printf(
		"VibeGravity worker reasoning: using MockCodexJSONClient (codex_enabled=%t codex_client=%s endpoint_configured=%t model_configured=%t); no real Codex API calls will be made.",
		cfg.Enabled,
		clientMode,
		strings.TrimSpace(cfg.Endpoint) != "",
		strings.TrimSpace(cfg.Model) != "",
	)
	mockCodex := reasoning.NewMockCodexJSONClient()
	bridgeConfig := reasoning.CodexBridgeConfig{Enabled: true}
	stage1, err := reasoning.NewCodexStage1Extractor(bridgeConfig, mockCodex)
	if err != nil {
		return nil, err
	}
	stage2, err := reasoning.NewCodexStage2Resolver(bridgeConfig, mockCodex)
	if err != nil {
		return nil, err
	}
	return reasoning.NewPipelineOrchestrator(stage1, stage2, stage2InputPreparer)
}
