// ============================================================
// FILE     : tests/migration_contract_test.go
// PURPOSE  : Guards migration and job-state contracts that must not regress during parallel Work Pack 03 fixes.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : TestMigrationContractUpdatesUniquenessTargetsPriorMemory, TestJobFailureContractSeparatesBlockedFromRetryable
// DEPENDS  : os, path/filepath, runtime, strings, testing
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: Keep these tests contract-focused; do not turn them into feature implementation tests.
// ============================================================

package tests

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestMigrationContractUpdatesUniquenessTargetsPriorMemory(t *testing.T) {
	t.Parallel()

	sql := readRepoFile(t, "migrations", "000002_create_core_tables.up.sql")
	indexSQL := extractBetween(t, sql, "CREATE UNIQUE INDEX memory_edges_single_updates_target_idx", ";")
	followUpSQL := readRepoFile(t, "migrations", "000004_fix_updates_edge_target_index.up.sql")

	if !strings.Contains(indexSQL, "ON memory_edges (to_memory_id)") {
		t.Fatalf("updates uniqueness must guard the prior memory target, got:\n%s", indexSQL)
	}
	if strings.Contains(indexSQL, "ON memory_edges (from_memory_id)") {
		t.Fatalf("updates uniqueness must not be keyed by from_memory_id, got:\n%s", indexSQL)
	}
	if !strings.Contains(indexSQL, "WHERE edge_kind = 'updates'") {
		t.Fatalf("updates uniqueness must remain a partial updates-only index, got:\n%s", indexSQL)
	}
	for _, want := range []string{
		"DROP INDEX IF EXISTS memory_edges_single_updates_target_idx",
		"ON memory_edges (to_memory_id)",
		"WHERE edge_kind = 'updates'",
	} {
		if !strings.Contains(followUpSQL, want) {
			t.Fatalf("follow-up migration must preserve %q for existing DBs, got:\n%s", want, followUpSQL)
		}
	}
}

func TestMigrationContractAllowsCorrectionApplyJobs(t *testing.T) {
	t.Parallel()

	sql := readRepoFile(t, "migrations", "000002_create_core_tables.up.sql")
	jobTableSQL := extractBetween(t, sql, "CREATE TABLE ingest_jobs", ");")

	if !strings.Contains(jobTableSQL, "'correction_apply'") {
		t.Fatalf("ingest_jobs job_kind CHECK must allow synchronous correction provenance jobs, got:\n%s", jobTableSQL)
	}
}

func TestJobFailureContractSeparatesBlockedFromRetryable(t *testing.T) {
	t.Parallel()

	source := readRepoFile(t, "internal", "store", "postgres", "jobs.go")
	failJob := extractBetween(t, source, "func failJob", "// BlockJob records deterministic unsupported work")
	blockJob := extractBetween(t, source, "func blockJob", "func jobErrorString")

	if !strings.Contains(failJob, "status = 'queued'") {
		t.Fatalf("FailJob must return transient failures to the queued retry state, got:\n%s", failJob)
	}
	if !strings.Contains(failJob, "interval '30 seconds'") {
		t.Fatalf("FailJob must preserve retry scheduling, got:\n%s", failJob)
	}
	if strings.Contains(failJob, "status = 'blocked'") {
		t.Fatalf("FailJob must not use the permanent blocked state, got:\n%s", failJob)
	}

	if !strings.Contains(blockJob, "status = 'blocked'") {
		t.Fatalf("BlockJob must use the permanent blocked state, got:\n%s", blockJob)
	}
	if strings.Contains(blockJob, "interval '30 seconds'") {
		t.Fatalf("BlockJob must not schedule automatic retry, got:\n%s", blockJob)
	}
	if strings.Contains(blockJob, "status = 'queued'") {
		t.Fatalf("BlockJob must not return deterministic unsupported work to queued, got:\n%s", blockJob)
	}
}

func TestMigrationContractAddsAppendSafeMemoryCorrections(t *testing.T) {
	t.Parallel()

	sql := readRepoFile(t, "migrations", "000002_create_core_tables.up.sql")
	tableSQL := extractBetween(t, sql, "CREATE TABLE memory_corrections", ");")
	indexSQL := extractBetween(t, sql, "CREATE UNIQUE INDEX memory_corrections_tenant_workspace_idempotency_key_idx", ";")

	for _, want := range []string{
		"memory_id TEXT NOT NULL REFERENCES memories (id)",
		"raw_event_id TEXT NOT NULL REFERENCES raw_events (id)",
		"correction_text TEXT NOT NULL",
		"status TEXT NOT NULL DEFAULT 'recorded'",
	} {
		if !strings.Contains(tableSQL, want) {
			t.Fatalf("memory_corrections must preserve %q, got:\n%s", want, tableSQL)
		}
	}
	if !strings.Contains(indexSQL, "ON memory_corrections (tenant_id, workspace_id, idempotency_key)") {
		t.Fatalf("memory correction idempotency must be tenant/workspace scoped, got:\n%s", indexSQL)
	}
}

func TestMigrationContractScopesProfilesAndSessionSummaries(t *testing.T) {
	t.Parallel()

	sql := readRepoFile(t, "migrations", "000002_create_core_tables.up.sql")
	profilesSQL := extractBetween(t, sql, "CREATE TABLE profiles", ");")
	profileIndexSQL := extractBetween(t, sql, "CREATE INDEX profiles_tenant_workspace_entity_updated_at_idx", ";")
	summaryIndexSQL := extractBetween(t, sql, "CREATE INDEX session_summaries_tenant_workspace_session_updated_at_idx", ";")
	followUpSQL := readRepoFile(t, "migrations", "000005_scope_profiles_and_summaries.up.sql")

	for _, want := range []string{
		"tenant_id TEXT NOT NULL",
		"workspace_id TEXT NOT NULL",
		"PRIMARY KEY (tenant_id, workspace_id, entity_id, scope)",
	} {
		if !strings.Contains(profilesSQL, want) {
			t.Fatalf("profiles migration must preserve %q, got:\n%s", want, profilesSQL)
		}
	}
	if !strings.Contains(profileIndexSQL, "ON profiles (tenant_id, workspace_id, entity_id, updated_at DESC)") {
		t.Fatalf("profile lookup index must be tenant/workspace scoped, got:\n%s", profileIndexSQL)
	}
	if !strings.Contains(summaryIndexSQL, "ON session_summaries (tenant_id, workspace_id, session_id, updated_at DESC)") {
		t.Fatalf("session summary lookup index must be tenant/workspace scoped, got:\n%s", summaryIndexSQL)
	}
	for _, want := range []string{
		"ADD COLUMN IF NOT EXISTS tenant_id",
		"ADD COLUMN IF NOT EXISTS workspace_id",
		"ADD PRIMARY KEY (tenant_id, workspace_id, entity_id, scope)",
		"session_summaries_tenant_workspace_session_updated_at_idx",
	} {
		if !strings.Contains(followUpSQL, want) {
			t.Fatalf("follow-up migration must preserve %q for existing DBs, got:\n%s", want, followUpSQL)
		}
	}
}

func readRepoFile(t *testing.T, parts ...string) string {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not locate current test file")
	}
	repoRoot := filepath.Dir(filepath.Dir(currentFile))
	pathParts := append([]string{repoRoot}, parts...)
	data, err := os.ReadFile(filepath.Join(pathParts...))
	if err != nil {
		t.Fatalf("read repo file %v: %v", parts, err)
	}
	return string(data)
}

func extractBetween(t *testing.T, text, startMarker, endMarker string) string {
	t.Helper()

	start := strings.Index(text, startMarker)
	if start < 0 {
		t.Fatalf("missing start marker %q", startMarker)
	}
	remainder := text[start:]
	end := strings.Index(remainder, endMarker)
	if end < 0 {
		t.Fatalf("missing end marker %q after %q", endMarker, startMarker)
	}
	return remainder[:end+len(endMarker)]
}
