// ============================================================
// FILE     : tests/baseline_test.go
// PURPOSE  : Provides baseline integration smoke tests for config and health checks.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : TestConfigLoad, TestHealthzEndpoint
// DEPENDS  : internal/config, internal/db, internal/httpapi
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: Keep database-dependent tests skippable when VIBEGRAVITY_DB_URL is unset.
// ============================================================

package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/config"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/db"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/httpapi"
)

func TestConfigLoad(t *testing.T) {
	_ = os.Setenv("VIBEGRAVITY_EMBEDDING_MODEL", "test-model-xyz")
	defer func() { _ = os.Unsetenv("VIBEGRAVITY_EMBEDDING_MODEL") }()

	cfg := config.LoadConfig()
	if cfg.EmbeddingModel != "test-model-xyz" {
		t.Errorf("Expected embedding model to be test-model-xyz, got %s", cfg.EmbeddingModel)
	}
}

func TestHealthzEndpoint(t *testing.T) {
	// Skip this test if no database is available
	dbURL := os.Getenv("VIBEGRAVITY_DB_URL")
	if dbURL == "" {
		t.Skip("Skipping TestHealthzEndpoint because VIBEGRAVITY_DB_URL is not set")
	}

	cfg := config.LoadConfig()
	ctx := context.Background()

	pool, err := db.NewPool(ctx, cfg)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	app := &httpapi.App{
		DBPool: pool,
	}

	router := httpapi.NewRouter(app)

	req, err := http.NewRequest("GET", "/healthz", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// We expect a JSON response with status "ok"
	expected := `{"status":"ok"}` + "\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}
