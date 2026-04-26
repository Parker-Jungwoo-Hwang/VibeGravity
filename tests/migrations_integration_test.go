// ============================================================
// FILE     : tests/migrations_integration_test.go
// PURPOSE  : Runs opt-in migration apply and rollback smoke checks against an empty PostgreSQL database.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : migration integration tests
// DEPENDS  : context, os, os/exec, path/filepath, testing, time, pgxpool
// USED_BY  : make integration-postgres
// ------------------------------------------------------------
// AGENT_NOTE: This test uses VIBEGRAVITY_MIGRATION_TEST_DB_URL so the normal live DB gate is never dropped or rolled back by accident.
// ============================================================

package tests

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestMigrationsApplyOnEmptyDatabaseAndRollbackSmoke(t *testing.T) {
	dbURL := os.Getenv("VIBEGRAVITY_MIGRATION_TEST_DB_URL")
	if dbURL == "" {
		t.Skip("Skipping migration apply/rollback smoke because VIBEGRAVITY_MIGRATION_TEST_DB_URL is not set")
	}
	migrateBin, err := exec.LookPath("migrate")
	if err != nil {
		t.Fatalf("VIBEGRAVITY_MIGRATION_TEST_DB_URL is set, but golang-migrate CLI is unavailable: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("connect migration test database: %v", err)
	}
	defer pool.Close()

	assertMigrationDatabaseEmpty(ctx, t, pool)

	root := repoRoot(t)
	migrationPath := filepath.Join(root, "migrations")
	runMigrate(t, root, migrateBin, migrationPath, dbURL, "up")
	assertPublicTableExists(ctx, t, pool, "memories")
	assertPublicTableExists(ctx, t, pool, "memory_trace")
	assertPublicTableExists(ctx, t, pool, "ingest_jobs")

	runMigrate(t, root, migrateBin, migrationPath, dbURL, "down", "1")
	assertMigrationVersion(ctx, t, pool, 4, false)

	runMigrate(t, root, migrateBin, migrationPath, dbURL, "up", "1")
	assertMigrationVersion(ctx, t, pool, 5, false)
}

func assertMigrationDatabaseEmpty(ctx context.Context, t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	var tableCount int
	if err := pool.QueryRow(ctx, `
		SELECT count(*)
		FROM information_schema.tables
		WHERE table_schema = 'public'
		  AND table_type = 'BASE TABLE'
		  AND table_name <> 'schema_migrations'
	`).Scan(&tableCount); err != nil {
		t.Fatalf("inspect migration test database: %v", err)
	}
	if tableCount != 0 {
		t.Fatalf("migration test database must be empty; found %d existing public tables", tableCount)
	}
}

func assertPublicTableExists(ctx context.Context, t *testing.T, pool *pgxpool.Pool, table string) {
	t.Helper()

	var exists bool
	if err := pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.tables
			WHERE table_schema = 'public'
			  AND table_name = $1
		)
	`, table).Scan(&exists); err != nil {
		t.Fatalf("check table %s: %v", table, err)
	}
	if !exists {
		t.Fatalf("expected migrated table %s to exist", table)
	}
}

func assertMigrationVersion(ctx context.Context, t *testing.T, pool *pgxpool.Pool, version int, dirty bool) {
	t.Helper()

	var gotVersion int
	var gotDirty bool
	if err := pool.QueryRow(ctx, `SELECT version, dirty FROM schema_migrations`).Scan(&gotVersion, &gotDirty); err != nil {
		t.Fatalf("read schema_migrations: %v", err)
	}
	if gotVersion != version || gotDirty != dirty {
		t.Fatalf("unexpected migration version: got version=%d dirty=%v want version=%d dirty=%v", gotVersion, gotDirty, version, dirty)
	}
}

func runMigrate(t *testing.T, root string, migrateBin string, migrationPath string, dbURL string, args ...string) {
	t.Helper()

	cmdArgs := append([]string{"-path", migrationPath, "-database", dbURL}, args...)
	cmd := exec.Command(migrateBin, cmdArgs...)
	cmd.Dir = root
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("migrate %v failed: %v\n%s", args, err, output)
	}
}
