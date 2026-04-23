// ============================================================
// FILE     : cmd/worker/main.go
// PURPOSE  : Starts the background worker process for ingest jobs and maintenance.
// LAYER    : interface
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : main
// DEPENDS  : internal/config, internal/db, internal/store/postgres
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
	"github.com/parker-jungwoo-hwang/vibegravity/internal/store/postgres"
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

	// Store initialization (mock implementation for now)
	_ = postgres.NewStore(pool)

	log.Println("Worker is running. Waiting for jobs...")

	// Setup graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down Worker gracefully...")

		// In a real worker, we'd wait for the current job to finish
		time.Sleep(1 * time.Second)
		cancel()
	}()

	// Mock Main Loop
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Worker stopped.")
			return
		case <-ticker.C:
			// In the future: Claim job using the pgxpool within a transaction
			log.Println("Ping: worker is idle...")
		}
	}
}
