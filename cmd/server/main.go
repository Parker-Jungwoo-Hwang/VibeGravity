// ============================================================
// FILE     : cmd/server/main.go
// PURPOSE  : Starts the HTTP API process and wires runtime dependencies.
// LAYER    : interface
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : main
// DEPENDS  : internal/config, internal/db, internal/httpapi, internal/store/postgres
// USED_BY  : Makefile, deployments
// ------------------------------------------------------------
// AGENT_NOTE: Keep API hot path behavior separate from worker reasoning work.
// ============================================================

// Package main starts the VibeGravity API server process.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/config"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/db"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/httpapi"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/store/postgres"
)

func main() {
	log.Println("Starting VibeGravity API Server...")

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

	app := &httpapi.App{
		// Service: coreService (to be wired later)
		DBPool: pool,
	}

	router := httpapi.NewRouter(app)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// Setup graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down API Server gracefully...")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("Error during server shutdown: %v", err)
		}
		cancel()
	}()

	log.Printf("API Server listening on %s", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Server failed: %v", err)
	}

	// Wait for context cancellation to complete
	<-ctx.Done()
	log.Println("API Server stopped.")
}
