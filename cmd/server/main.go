// ============================================================
// FILE     : cmd/server/main.go
// PURPOSE  : Starts the HTTP API process and wires runtime dependencies.
// LAYER    : interface
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : main
// DEPENDS  : internal/config, internal/httpapi, internal/runtime
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
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/config"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/httpapi"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/runtime"
)

func main() {
	log.Println("Starting VibeGravity API Server...")

	cfg := config.LoadConfig()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app, closeRuntime, err := runtime.OpenHTTPApp(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to initialize HTTP runtime: %v", err)
	}
	defer closeRuntime()

	router := httpapi.NewRouter(app)

	addr := serverAddr()
	if !isLoopbackAddr(addr) && os.Getenv("VIBEGRAVITY_UNSAFE_ALLOW_NON_LOOPBACK") != "true" {
		log.Fatalf("Refusing non-loopback bind %q. Set VIBEGRAVITY_UNSAFE_ALLOW_NON_LOOPBACK=true only for explicitly trusted local validation.", addr)
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
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

func serverAddr() string {
	return firstNonEmpty(os.Getenv("VIBEGRAVITY_HTTP_ADDR"), "127.0.0.1:8080")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func isLoopbackAddr(addr string) bool {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return false
	}
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
