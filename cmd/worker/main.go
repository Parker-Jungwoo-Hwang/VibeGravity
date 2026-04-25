// ============================================================
// FILE     : cmd/worker/main.go
// PURPOSE  : Starts the background worker process for ingest jobs and maintenance.
// LAYER    : interface
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : main
// DEPENDS  : internal/config, internal/db, internal/graph, internal/reasoning, internal/store/postgres, internal/worker
// USED_BY  : Makefile, worker deployments
// ------------------------------------------------------------
// AGENT_NOTE: Keep Codex and embedding work off the API hot path.
// ============================================================

// Package main starts the VibeGravity background worker process.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/config"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/db"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/graph"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/reasoning"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/store/postgres"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/worker"
)

func main() {
	log.Println("Starting VibeGravity Background Worker...")

	cfg := config.LoadConfig()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := db.NewPool(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	pgStore := postgres.NewStore(pool)
	applyEngine, err := graph.NewStoreBackedApplyEngine(pgStore)
	if err != nil {
		log.Fatalf("Failed to initialize graph apply engine: %v", err)
	}
	dreamingService, err := graph.NewDreamingService(graph.DreamingDependencies{
		Store: pgStore,
	})
	if err != nil {
		log.Fatalf("Failed to initialize dreaming service: %v", err)
	}
	stage2InputPreparer := worker.NewStoreBackedStage2InputPreparer(worker.Stage2SourceStores{
		Profiles:  pgStore,
		Memories:  pgStore,
		Documents: pgStore,
		Plans:     pgStore,
		Notes:     pgStore,
		Groups:    pgStore,
	})
	reasoner, err := newReasoner(stage2InputPreparer)
	if err != nil {
		log.Fatalf("Failed to initialize reasoning orchestrator: %v", err)
	}
	processor, err := worker.NewProcessor(worker.Dependencies{
		WorkerID:    workerID(),
		Jobs:        pgStore,
		RawEvents:   pgStore,
		Reasoner:    reasoner,
		ApplyEngine: applyEngine,
		Dreaming:    dreamingService,
	})
	if err != nil {
		log.Fatalf("Failed to initialize worker processor: %v", err)
	}

	log.Println("Worker is running. Waiting for jobs...")

	// Setup graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down Worker gracefully...")

		cancel()
	}()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Worker stopped.")
			return
		case <-ticker.C:
			result, err := processor.RunOnce(ctx)
			if err != nil {
				log.Printf("Worker pass completed with error: %v", err)
			}
			if result.Claimed == 0 {
				log.Println("Worker is idle.")
				continue
			}
			log.Printf(
				"Worker pass: claimed=%d completed=%d failed=%d blocked=%d applied_operations=%d memory_ids=%d traces_written=%d session_dreams=%d workspace_dreams=%d",
				result.Claimed,
				result.Completed,
				result.Failed,
				result.Blocked,
				result.AppliedOperationCount,
				result.MemoryIDCount,
				result.TraceWrittenCount,
				result.SessionDreamCount,
				result.WorkspaceDreamCount,
			)
		}
	}
}

func workerID() string {
	if value := os.Getenv("VIBEGRAVITY_WORKER_ID"); value != "" {
		return value
	}
	host, err := os.Hostname()
	if err != nil || host == "" {
		return "worker:local"
	}
	return "worker:" + host
}

func newReasoner(stage2InputPreparer *reasoning.Stage2InputPreparer) (reasoning.Orchestrator, error) {
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
