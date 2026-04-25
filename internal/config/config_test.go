// ============================================================
// FILE     : internal/config/config_test.go
// PURPOSE  : Guards disabled-by-default runtime configuration contracts.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : config tests
// DEPENDS  : testing
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: Keep real Codex opt-in only; defaults must not enable networked reasoning.
// ============================================================

package config

import "testing"

func TestLoadConfigDisablesRealCodexByDefault(t *testing.T) {
	t.Setenv("VIBEGRAVITY_CODEX_ENABLED", "")
	t.Setenv("VIBEGRAVITY_CODEX_ENDPOINT", "")
	t.Setenv("VIBEGRAVITY_CODEX_MODEL", "")

	cfg := LoadConfig()

	if cfg.Codex.Enabled {
		t.Fatalf("real Codex must be disabled by default")
	}
	if cfg.Codex.Endpoint != "" {
		t.Fatalf("default Codex endpoint must be empty, got %q", cfg.Codex.Endpoint)
	}
	if cfg.Codex.Model != "" {
		t.Fatalf("default Codex model must be empty, got %q", cfg.Codex.Model)
	}
}

func TestLoadConfigRequiresExplicitCodexEnablement(t *testing.T) {
	t.Setenv("VIBEGRAVITY_CODEX_ENABLED", "true")
	t.Setenv("VIBEGRAVITY_CODEX_ENDPOINT", "https://codex.invalid")
	t.Setenv("VIBEGRAVITY_CODEX_MODEL", "codex-contract-test")

	cfg := LoadConfig()

	if !cfg.Codex.Enabled {
		t.Fatalf("explicit Codex enablement was not loaded")
	}
	if cfg.Codex.Endpoint != "https://codex.invalid" {
		t.Fatalf("unexpected Codex endpoint: %q", cfg.Codex.Endpoint)
	}
	if cfg.Codex.Model != "codex-contract-test" {
		t.Fatalf("unexpected Codex model: %q", cfg.Codex.Model)
	}
}
