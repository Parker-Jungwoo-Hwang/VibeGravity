// ============================================================
// FILE     : internal/store/postgres/dreaming_test.go
// PURPOSE  : Verifies PostgreSQL dreaming promotion query contracts without a live database.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : postgres dreaming helper tests
// DEPENDS  : strings, testing, time, internal/core
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: Promotion SQL must update metadata only, never scope or provenance fields.
// ============================================================

package postgres

import (
	"strings"
	"testing"
	"time"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

func TestDreamingPromotionStatementMarksMetadataOnly(t *testing.T) {
	t.Parallel()

	sql, args := dreamingPromotionStatement(&core.DreamingPromotionRequest{
		JobID:         "job_dream_1",
		TenantID:      "tenant_1",
		WorkspaceID:   "workspace_1",
		MemoryIDs:     []string{"mem_1", "mem_2"},
		Tier:          core.DreamingTierMidTerm,
		Now:           time.Date(2026, time.April, 24, 3, 0, 0, 0, time.UTC),
		MinConfidence: 0.5,
	})

	if !strings.Contains(sql, "jsonb_set") || !strings.Contains(sql, "'{dreaming}'") {
		t.Fatalf("expected dreaming metadata update, got: %s", sql)
	}
	if strings.Contains(sql, "scope =") || strings.Contains(sql, "owner_entity_id =") || strings.Contains(sql, "memory_trace") {
		t.Fatalf("promotion SQL must not mutate scope, owner, or trace: %s", sql)
	}
	if !strings.Contains(sql, "id = ANY($7::text[])") {
		t.Fatalf("expected explicit memory id filter, got: %s", sql)
	}
	if len(args) != 7 {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestDreamingPromotionStatementStableWorkspaceFilters(t *testing.T) {
	t.Parallel()

	sql, _ := dreamingPromotionStatement(&core.DreamingPromotionRequest{
		JobID:             "job_dream_workspace",
		TenantID:          "tenant_1",
		WorkspaceID:       "workspace_1",
		Tier:              core.DreamingTierLongTerm,
		MinConfidence:     0.85,
		RequireStableKind: true,
	})

	if !strings.Contains(sql, "kind IN ('fact','preference','trait','goal','constraint','decision','procedure')") {
		t.Fatalf("expected stable-kind filter, got: %s", sql)
	}
	if !strings.Contains(sql, "scope <> 'session_scratch'") {
		t.Fatalf("expected session scratch exclusion, got: %s", sql)
	}
	if !strings.Contains(sql, "FOR UPDATE SKIP LOCKED") {
		t.Fatalf("expected safe concurrent promotion locking, got: %s", sql)
	}
}

func TestValidateDreamingPromotionRequestRejectsUnsupportedTier(t *testing.T) {
	t.Parallel()

	err := validateDreamingPromotionRequest(&core.DreamingPromotionRequest{
		JobID:       "job_dream_1",
		TenantID:    "tenant_1",
		WorkspaceID: "workspace_1",
		Tier:        core.DreamingTierShortTerm,
	})
	if err == nil {
		t.Fatalf("expected unsupported short-term promotion to fail")
	}
}
