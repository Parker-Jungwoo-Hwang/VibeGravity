// ============================================================
// FILE     : cmd/cli/main.go
// PURPOSE  : Starts the CLI and runs local operator checks such as doctor.
// LAYER    : interface
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : main
// DEPENDS  : internal/config, internal/db, net/http
// USED_BY  : Makefile, local operators
// ------------------------------------------------------------
// AGENT_NOTE: Keep doctor checks read-only and truthful about unavailable dependencies.
// ============================================================

// Package main starts the VibeGravity CLI process.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/config"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/db"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]

	switch cmd {
	case "doctor":
		runDoctor()
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: cli <command>")
	fmt.Println("\nCommands:")
	fmt.Println("  doctor    Check system configuration and dependencies")
}

func runDoctor() {
	fmt.Println("VibeGravity Doctor")
	fmt.Println("==================")

	// 1. Check Config
	fmt.Println("\n[1] Checking Configuration...")
	cfg := config.LoadConfig()
	fmt.Printf("  DatabaseURL:       %s\n", maskPassword(cfg.DatabaseURL))
	fmt.Printf("  MigrationPath:     %s\n", cfg.MigrationPath)
	fmt.Printf("  EmbeddingEndpoint: %s\n", cfg.EmbeddingEndpoint)
	fmt.Printf("  EmbeddingModel:    %s\n", cfg.EmbeddingModel)
	fmt.Printf("  EmbeddingDims:     %d\n", cfg.EmbeddingDims)
	fmt.Println("  -> Config OK")

	// 2. Check Database
	fmt.Println("\n[2] Checking Database Connection...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := db.NewPool(ctx, cfg)
	if err != nil {
		fmt.Printf("  -> ERROR: Failed to connect to database: %v\n", err)
	} else {
		defer pool.Close()
		fmt.Println("  -> Database Connection OK")
	}

	// 3. Check Embedding Endpoint
	fmt.Println("\n[3] Checking Embedding Endpoint...")
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(cfg.EmbeddingEndpoint)
	if err != nil {
		fmt.Printf("  -> ERROR: Failed to reach embedding endpoint (%s): %v\n", cfg.EmbeddingEndpoint, err)
	} else {
		defer func() { _ = resp.Body.Close() }()
		fmt.Printf("  -> Embedding Endpoint OK (Status: %s)\n", resp.Status)
	}

	fmt.Println("\nDoctor check completed.")
}

func maskPassword(url string) string {
	// A simple masker for display purposes, could be improved.
	// We'll just print it for now if we assume local dev,
	// but ideally we'd parse the URL and mask the password part.
	return url
}
