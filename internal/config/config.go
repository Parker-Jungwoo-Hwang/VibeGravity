// ============================================================
// FILE     : internal/config/config.go
// PURPOSE  : Loads runtime configuration from .env and environment variables.
// LAYER    : infra
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : Config, LoadConfig
// DEPENDS  : github.com/joho/godotenv, os, strconv
// USED_BY  : cmd/server, cmd/worker, cmd/cli, tests
// ------------------------------------------------------------
// AGENT_NOTE: Treat env names as runtime contract and update docs when they change.
// ============================================================

// Package config defines VibeGravity runtime configuration.
package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config contains settings shared by the server, worker, and CLI.
type Config struct {
	DatabaseURL       string
	MigrationPath     string
	EmbeddingEndpoint string
	EmbeddingModel    string
	EmbeddingDims     int
}

// LoadConfig loads configuration from .env and environment variables.
func LoadConfig() Config {
	// Ignore error if .env doesn't exist
	_ = godotenv.Load()

	cfg := Config{
		DatabaseURL:       getEnv("VIBEGRAVITY_DB_URL", "postgres://localhost:5432/vibegravity?sslmode=disable"),
		MigrationPath:     getEnv("VIBEGRAVITY_MIGRATION_PATH", "migrations"),
		EmbeddingEndpoint: getEnv("VIBEGRAVITY_EMBEDDING_ENDPOINT", "http://localhost:8080"),
		EmbeddingModel:    getEnv("VIBEGRAVITY_EMBEDDING_MODEL", "pending"),
		EmbeddingDims:     getEnvAsInt("VIBEGRAVITY_EMBEDDING_DIMS", 0),
	}
	return cfg
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	valStr := getEnv(key, "")
	if valStr == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		log.Printf("Warning: invalid integer for %s: %s", key, valStr)
		return defaultVal
	}
	return val
}
