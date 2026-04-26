// ============================================================
// FILE     : cmd/worker/main.go
// PURPOSE  : Starts the background worker process for ingest jobs and maintenance.
// LAYER    : interface
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : main
// DEPENDS  : internal/config, internal/runtime
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
	"github.com/parker-jungwoo-hwang/vibegravity/internal/runtime"
)

func main() {
	log.Println("Starting VibeGravity Background Worker...")

	cfg := config.LoadConfig()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	processor, closeRuntime, err := runtime.OpenWorkerProcessor(ctx, cfg, runtime.WorkerOptions{
		WorkerID: workerID(),
		Logger:   log.Default(),
	})
	if err != nil {
		log.Fatalf("Failed to initialize worker runtime: %v", err)
	}
	defer closeRuntime()

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
