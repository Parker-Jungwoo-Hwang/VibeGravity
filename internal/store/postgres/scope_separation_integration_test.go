// ============================================================
// FILE     : internal/store/postgres/scope_separation_integration_test.go
// PURPOSE  : Strengthens live PostgreSQL scope-separation coverage for memory search visibility.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : TestPostgresSearchMemoriesLiveScopeSeparation
// DEPENDS  : context, os, slices, testing, time, internal/core, pgxpool
// USED_BY  : make integration-postgres
// ------------------------------------------------------------
// AGENT_NOTE: Keep this test opt-in through VIBEGRAVITY_DB_URL; it must prove private and group scopes stay separate in live SQL.
// ============================================================

package postgres

import (
	"context"
	"fmt"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

const scopeSeparationWorkspaceID = "workspace_scope_sep"

func TestPostgresSearchMemoriesLiveScopeSeparation(t *testing.T) {
	dbURL := os.Getenv("VIBEGRAVITY_DB_URL")
	if dbURL == "" {
		t.Skip("Skipping live Postgres scope-separation test because VIBEGRAVITY_DB_URL is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("connect test database: %v", err)
	}
	defer pool.Close()

	store := NewStore(pool)
	tenantID := fmt.Sprintf("tenant_scope_sep_%d", time.Now().UnixNano())
	workspaceID := scopeSeparationWorkspaceID
	hermesID := "agent:hermes-main"
	otherID := "agent:other"
	groupID := "group_scope_sep"
	startedAt := time.Now().UTC()

	cleanupPostgresScopeSeparationRows(ctx, t, pool, tenantID)
	defer cleanupPostgresScopeSeparationRows(context.Background(), t, pool, tenantID)

	mustSeedJob(ctx, t, pool, tenantID, workspaceID, "job_scope_sep")
	if err := store.CreateMemoryGroup(ctx, &core.MemoryGroup{
		ID:          groupID,
		TenantID:    tenantID,
		WorkspaceID: workspaceID,
		Name:        "scope separation group",
		CreatedAt:   startedAt,
	}); err != nil {
		t.Fatalf("seed memory group: %v", err)
	}
	for _, memory := range []*core.Memory{
		scopeSeparationMemory(tenantID, "mem_scope_workspace", core.MemoryScopeWorkspaceShared, nil, "workspace:scope", "Workspace scope separation visible rule.", "fp_scope_workspace", startedAt),
		scopeSeparationMemory(tenantID, "mem_scope_private_hermes", core.MemoryScopeAgentPrivate, nil, hermesID, "Hermes private scope separation visible rule.", "fp_scope_private_hermes", startedAt),
		scopeSeparationMemory(tenantID, "mem_scope_private_other", core.MemoryScopeAgentPrivate, nil, otherID, "Other private scope separation hidden rule.", "fp_scope_private_other", startedAt),
		scopeSeparationMemory(tenantID, "mem_scope_group", core.MemoryScopeGroupShared, &groupID, "group:scope", "Group scope separation visible only with membership.", "fp_scope_group", startedAt),
	} {
		trace := &core.MemoryTrace{
			MemoryID:              memory.ID,
			RawEventIDs:           []string{"evt_scope_sep"},
			ReasoningJobID:        "job_scope_sep",
			ReasoningStage:        "resolve",
			CandidateSnapshotJSON: []byte(`{"scope_separation":true}`),
			AppliedOperationsJSON: []byte(fmt.Sprintf(`[{"operation_id":"%s"}]`, memory.ID)),
			RelatedDocumentIDs:    []string{},
			CreatedAt:             startedAt,
		}
		if err := store.CreateMemoryWithTrace(ctx, memory, trace); err != nil {
			t.Fatalf("seed memory %s: %v", memory.ID, err)
		}
	}

	withoutGroup, err := store.SearchMemories(ctx, scopeSeparationSearchRequest(tenantID, workspaceID, hermesID, nil))
	if err != nil {
		t.Fatalf("search without group membership: %v", err)
	}
	assertScopeSeparationIDs(t, withoutGroup, []string{"mem_scope_workspace", "mem_scope_private_hermes"}, []string{"mem_scope_private_other", "mem_scope_group"})

	if err := store.AddMembership(ctx, &core.MemoryGroupMembership{
		GroupID:   groupID,
		EntityID:  hermesID,
		CreatedAt: startedAt,
	}); err != nil {
		t.Fatalf("seed group membership: %v", err)
	}
	memberships, err := store.ListMembershipsForEntity(ctx, tenantID, workspaceID, hermesID)
	if err != nil {
		t.Fatalf("load group memberships: %v", err)
	}
	visibleGroups := make([]string, 0, len(memberships))
	for _, membership := range memberships {
		visibleGroups = append(visibleGroups, membership.GroupID)
	}

	withGroup, err := store.SearchMemories(ctx, scopeSeparationSearchRequest(tenantID, workspaceID, hermesID, visibleGroups))
	if err != nil {
		t.Fatalf("search with group membership: %v", err)
	}
	assertScopeSeparationIDs(t, withGroup, []string{"mem_scope_workspace", "mem_scope_private_hermes", "mem_scope_group"}, []string{"mem_scope_private_other"})
}

func scopeSeparationMemory(tenantID string, id string, scope core.MemoryScope, groupID *string, ownerID string, text string, fingerprint string, now time.Time) *core.Memory {
	return &core.Memory{
		ID:            id,
		TenantID:      tenantID,
		WorkspaceID:   scopeSeparationWorkspaceID,
		Scope:         scope,
		GroupID:       groupID,
		OwnerEntityID: ownerID,
		Kind:          core.MemoryKindFact,
		ArtifactClass: core.ArtifactClassKnowledge,
		Text:          text,
		Fingerprint:   fingerprint,
		Confidence:    0.9,
		Status:        core.MemoryStatusActive,
		ValidFrom:     now,
		LatestFlag:    true,
		MetadataJSON:  []byte(`{}`),
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func scopeSeparationSearchRequest(tenantID string, workspaceID string, ownerID string, visibleGroupIDs []string) *core.SearchMemoriesRequest {
	return &core.SearchMemoriesRequest{
		TenantID:        tenantID,
		WorkspaceID:     workspaceID,
		OwnerEntityID:   ownerID,
		VisibleGroupIDs: visibleGroupIDs,
		Query:           "scope separation",
		Scopes: []core.MemoryScope{
			core.MemoryScopeAgentPrivate,
			core.MemoryScopeWorkspaceShared,
			core.MemoryScopeGroupShared,
		},
		ArtifactClasses: []core.ArtifactClass{core.ArtifactClassKnowledge},
	}
}

func assertScopeSeparationIDs(t *testing.T, resp *core.SearchMemoriesResponse, wantPresent []string, wantAbsent []string) {
	t.Helper()

	ids := make([]string, 0, len(resp.Memories))
	for _, memory := range resp.Memories {
		ids = append(ids, memory.MemoryID)
	}
	for _, want := range wantPresent {
		if !slices.Contains(ids, want) {
			t.Fatalf("expected memory %s in search results, got %v", want, ids)
		}
	}
	for _, blocked := range wantAbsent {
		if slices.Contains(ids, blocked) {
			t.Fatalf("memory %s should not be visible in search results %v", blocked, ids)
		}
	}
}

func cleanupPostgresScopeSeparationRows(ctx context.Context, t testing.TB, pool *pgxpool.Pool, tenantID string) {
	t.Helper()
	if _, err := pool.Exec(ctx, `
		DELETE FROM memory_group_memberships
		WHERE group_id IN (SELECT id FROM memory_groups WHERE tenant_id = $1)
	`, tenantID); err != nil {
		t.Fatalf("cleanup memory group memberships: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		DELETE FROM memories
		WHERE tenant_id = $1
	`, tenantID); err != nil {
		t.Fatalf("cleanup memories: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		DELETE FROM ingest_jobs
		WHERE tenant_id = $1
	`, tenantID); err != nil {
		t.Fatalf("cleanup ingest jobs: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		DELETE FROM memory_groups
		WHERE tenant_id = $1
	`, tenantID); err != nil {
		t.Fatalf("cleanup memory groups: %v", err)
	}
}
