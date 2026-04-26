// ============================================================
// FILE     : internal/config/config.go
// PURPOSE  : Loads runtime configuration from .env and environment variables.
// LAYER    : infra
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : Config, LoadConfig, CodexConfig
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

// CodexConfig contains disabled-by-default reasoning bridge settings.
type CodexConfig struct {
	Enabled    bool
	ClientMode string
	Endpoint   string
	Model      string
}

const (
	// CodexClientModeMock keeps worker reasoning on the deterministic mocked bridge.
	CodexClientModeMock = "mock"
	// CodexClientModeReal is reserved for explicit future real Codex execution.
	CodexClientModeReal = "real"
)

// Config contains settings shared by the server, worker, and CLI.
type Config struct {
	DatabaseURL       string
	MigrationPath     string
	EmbeddingEndpoint string
	EmbeddingModel    string
	EmbeddingDims     int
	Codex             CodexConfig
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
		Codex: CodexConfig{
			Enabled:    getEnvAsBool("VIBEGRAVITY_CODEX_ENABLED", false),
			ClientMode: getEnv("VIBEGRAVITY_CODEX_CLIENT", CodexClientModeMock),
			Endpoint:   getEnv("VIBEGRAVITY_CODEX_ENDPOINT", ""),
			Model:      getEnv("VIBEGRAVITY_CODEX_MODEL", ""),
		},
	}
	return cfg
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
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

func getEnvAsBool(key string, defaultVal bool) bool {
	valStr := getEnv(key, "")
	if valStr == "" {
		return defaultVal
	}
	val, err := strconv.ParseBool(valStr)
	if err != nil {
		log.Printf("Warning: invalid boolean for %s: %s", key, valStr)
		return defaultVal
	}
	return val
}
