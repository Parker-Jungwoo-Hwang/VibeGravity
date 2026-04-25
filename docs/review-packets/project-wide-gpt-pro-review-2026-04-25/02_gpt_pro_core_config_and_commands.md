# 02 Gpt Pro Core Config And Commands

Generated: 2026-04-25

This file is part of the GPT-Pro review material bundle for VibeGravity.

## Included Sources

- `cmd/cli/main.go`
- `cmd/cli/main_test.go`
- `cmd/server/main.go`
- `cmd/worker/main.go`
- `internal/config/config.go`
- `internal/core/doc.go`
- `internal/core/document.go`
- `internal/core/dreaming.go`
- `internal/core/dto.go`
- `internal/core/entity.go`
- `internal/core/errors.go`
- `internal/core/group.go`
- `internal/core/job.go`
- `internal/core/kind.go`
- `internal/core/memory.go`
- `internal/core/note.go`
- `internal/core/plan.go`
- `internal/core/profile.go`
- `internal/core/raw_event.go`
- `internal/core/scope.go`
- `internal/core/service.go`
- `internal/core/service_test.go`
- `internal/core/session.go`
- `internal/db/pool.go`
- `internal/embed/doc.go`

## Source Contents


<!-- Source: cmd/cli/main.go | bytes=18769 | lines=644 | sha16=d33ce794ed5440db -->

```go
// ============================================================
// FILE     : cmd/cli/main.go
// PURPOSE  : Starts the CLI and runs local operator checks such as doctor and job metrics/recovery.
// LAYER    : interface
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : main
// DEPENDS  : context, errors, fmt, io, net/http, os, strconv, strings, time, internal/config, internal/core, internal/db, internal/eval, internal/mcp, internal/store/postgres
// USED_BY  : Makefile, local operators
// ------------------------------------------------------------
// AGENT_NOTE: Keep operator recovery explicit; blocked jobs must not requeue without a CLI action.
// ============================================================

// Package main starts the VibeGravity CLI process.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/config"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/db"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/eval"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/ingest"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/kernel"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/mcp"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/recall"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/store/postgres"
)

const defaultBlockedJobsCLIListLimit = 20
const defaultGoldenScenarioPath = "tests/golden/replay_eval.json"

type jobOperatorStore interface {
	ListBlockedJobs(ctx context.Context, limit int) ([]*core.IngestJob, error)
	RequeueBlockedJob(ctx context.Context, jobID string) error
	GetJobBacklogMetrics(ctx context.Context, req *core.JobBacklogMetricsRequest) (*core.JobBacklogMetrics, error)
}

type jobOperatorStoreFactory func(ctx context.Context) (jobOperatorStore, func(), error)
type serviceFactory func(ctx context.Context) (core.VibeGravityService, func(), error)

func main() {
	os.Exit(runCLI(context.Background(), os.Args[1:], os.Stdin, os.Stdout, openPostgresBlockedJobStore, openPostgresService))
}

func runCLI(ctx context.Context, args []string, in io.Reader, out io.Writer, openStore jobOperatorStoreFactory, openService serviceFactory) int {
	if len(args) < 1 {
		printUsage(out)
		return 1
	}

	cmd := args[0]
	switch cmd {
	case "doctor":
		runDoctor()
		return 0
	case "eval":
		return runEvalCommand(ctx, args[1:], out)
	case "jobs":
		return runJobsCommand(ctx, args[1:], out, openStore)
	case "mcp":
		return runMCPCommand(ctx, args[1:], in, out, openService)
	case "hermes":
		return runHermesCommand(args[1:], out)
	default:
		writef(out, "Unknown command: %s\n", cmd)
		printUsage(out)
		return 1
	}
}

func runMCPCommand(ctx context.Context, args []string, in io.Reader, out io.Writer, openService serviceFactory) int {
	if len(args) < 1 {
		printMCPUsage(out)
		return 1
	}
	switch args[0] {
	case "serve":
		return runMCPServe(ctx, args[1:], in, out, openService)
	default:
		writef(out, "Unknown mcp command: %s\n", args[0])
		printMCPUsage(out)
		return 1
	}
}

func runMCPServe(ctx context.Context, args []string, in io.Reader, out io.Writer, openService serviceFactory) int {
	if len(args) > 1 || (len(args) == 1 && args[0] != "--stdio") {
		printMCPUsage(out)
		return 1
	}
	service, closeService, err := openService(ctx)
	if err != nil {
		writef(out, "ERROR: open VibeGravity service: %v\n", err)
		return 1
	}
	defer closeService()

	surface, err := mcp.NewSurface(service)
	if err != nil {
		writef(out, "ERROR: create MCP surface: %v\n", err)
		return 1
	}
	server, err := mcp.NewServer(surface)
	if err != nil {
		writef(out, "ERROR: create MCP server: %v\n", err)
		return 1
	}
	if err := server.ServeStdio(ctx, in, out); err != nil {
		writef(out, "ERROR: serve MCP stdio: %v\n", err)
		return 1
	}
	return 0
}

func runEvalCommand(ctx context.Context, args []string, out io.Writer) int {
	if len(args) < 1 {
		printEvalUsage(out)
		return 1
	}
	switch args[0] {
	case "golden":
		path, err := parseGoldenEvalPath(args[1:])
		if err != nil {
			writef(out, "ERROR: %v\n", err)
			return 1
		}
		summary, err := eval.RunFile(ctx, path)
		if err != nil {
			writef(out, "ERROR: run golden eval: %v\n", err)
			return 1
		}
		printEvalSummary(out, summary, "Golden eval")
		if !summary.Passed {
			return 1
		}
		return 0
	case "demo":
		if len(args) > 1 {
			writef(out, "ERROR: unknown demo eval option: %s\n", args[1])
			return 1
		}
		summary := eval.RunHermesMemoryDemo(ctx)
		printEvalSummary(out, summary, "Hermes Memory demo eval")
		if !summary.Passed {
			return 1
		}
		return 0
	default:
		writef(out, "Unknown eval command: %s\n", args[0])
		printEvalUsage(out)
		return 1
	}
}

func parseGoldenEvalPath(args []string) (string, error) {
	path := defaultGoldenScenarioPath
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--path":
			if i+1 >= len(args) || args[i+1] == "" {
				return "", fmt.Errorf("--path requires a value")
			}
			path = args[i+1]
			i++
		default:
			return "", fmt.Errorf("unknown golden eval option: %s", args[i])
		}
	}
	return path, nil
}

func printEvalSummary(out io.Writer, summary *eval.Summary, label string) {
	if summary == nil {
		writeln(out, "No eval summary.")
		return
	}
	for _, result := range summary.Results {
		status := "PASS"
		if !result.Passed {
			status = "FAIL"
		}
		writef(out, "%s\t%s\tblocks=%v\ttokens=%d\tsources=%v\n", status, result.Scenario, result.Observed.BlockKinds, result.Observed.Tokens, result.Observed.Sources)
		for _, err := range result.Errors {
			writef(out, "  - %s\n", err)
		}
	}
	if summary.Passed {
		writef(out, "%s passed.\n", label)
		return
	}
	writef(out, "%s failed.\n", label)
}

func runHermesCommand(args []string, out io.Writer) int {
	if len(args) < 1 {
		printHermesUsage(out)
		return 1
	}
	switch args[0] {
	case "bootstrap":
		return runHermesBootstrap(args[1:], out)
	default:
		writef(out, "Unknown hermes command: %s\n", args[0])
		printHermesUsage(out)
		return 1
	}
}

func runHermesBootstrap(args []string, out io.Writer) int {
	name := "vibegravity"
	command := os.Args[0]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--name":
			if i+1 >= len(args) || args[i+1] == "" {
				writeln(out, "ERROR: --name requires a value")
				return 1
			}
			name = args[i+1]
			i++
		case "--command":
			if i+1 >= len(args) || args[i+1] == "" {
				writeln(out, "ERROR: --command requires a value")
				return 1
			}
			command = args[i+1]
			i++
		default:
			writef(out, "ERROR: unknown bootstrap option: %s\n", args[i])
			printHermesUsage(out)
			return 1
		}
	}
	if strings.TrimSpace(name) == "" || strings.ContainsAny(name, " \t\n") {
		writeln(out, "ERROR: Hermes MCP server name must be non-empty and contain no whitespace")
		return 1
	}
	if strings.TrimSpace(command) == "" {
		writeln(out, "ERROR: command must be non-empty")
		return 1
	}
	writeln(out, "Hermes MCP bootstrap command:")
	writef(out, "  hermes mcp add %s --command %s --args mcp serve --stdio\n", shellQuote(name), shellQuote(command))
	writeln(out)
	writeln(out, "After the VibeGravity API database is configured, verify with:")
	writef(out, "  hermes mcp test %s\n", shellQuote(name))
	return 0
}

func runJobsCommand(ctx context.Context, args []string, out io.Writer, openStore jobOperatorStoreFactory) int {
	if len(args) < 1 {
		printJobsUsage(out)
		return 1
	}

	switch args[0] {
	case "metrics":
		return runJobMetrics(ctx, args[1:], out, openStore)
	case "blocked":
		return runListBlockedJobs(ctx, args[1:], out, openStore)
	case "requeue-blocked":
		return runRequeueBlockedJob(ctx, args[1:], out, openStore)
	default:
		writef(out, "Unknown jobs command: %s\n", args[0])
		printJobsUsage(out)
		return 1
	}
}

func runJobMetrics(ctx context.Context, args []string, out io.Writer, openStore jobOperatorStoreFactory) int {
	req, err := parseJobMetricsArgs(args)
	if err != nil {
		writef(out, "ERROR: %v\n", err)
		return 1
	}
	store, closeStore, err := openStore(ctx)
	if err != nil {
		writef(out, "ERROR: open job store: %v\n", err)
		return 1
	}
	defer closeStore()

	metrics, err := store.GetJobBacklogMetrics(ctx, req)
	if err != nil {
		writef(out, "ERROR: get job metrics: %v\n", err)
		return 1
	}
	printJobMetrics(out, metrics)
	return 0
}

func runListBlockedJobs(ctx context.Context, args []string, out io.Writer, openStore jobOperatorStoreFactory) int {
	limit, err := parseBlockedJobsLimit(args)
	if err != nil {
		writef(out, "ERROR: %v\n", err)
		return 1
	}
	store, closeStore, err := openStore(ctx)
	if err != nil {
		writef(out, "ERROR: open job store: %v\n", err)
		return 1
	}
	defer closeStore()

	jobs, err := store.ListBlockedJobs(ctx, limit)
	if err != nil {
		writef(out, "ERROR: list blocked jobs: %v\n", err)
		return 1
	}
	printBlockedJobs(out, jobs)
	return 0
}

func runRequeueBlockedJob(ctx context.Context, args []string, out io.Writer, openStore jobOperatorStoreFactory) int {
	if len(args) != 1 || args[0] == "" {
		writeln(out, "ERROR: requeue-blocked requires one blocked job id")
		printJobsUsage(out)
		return 1
	}
	jobID := args[0]
	store, closeStore, err := openStore(ctx)
	if err != nil {
		writef(out, "ERROR: open job store: %v\n", err)
		return 1
	}
	defer closeStore()

	if err := store.RequeueBlockedJob(ctx, jobID); err != nil {
		if errors.Is(err, core.ErrNotFound) {
			writef(out, "ERROR: blocked job not found: %s\n", jobID)
			return 1
		}
		writef(out, "ERROR: requeue blocked job: %v\n", err)
		return 1
	}
	writef(out, "requeued blocked job %s\n", jobID)
	return 0
}

func parseJobMetricsArgs(args []string) (*core.JobBacklogMetricsRequest, error) {
	req := &core.JobBacklogMetricsRequest{DrainWindow: 15 * time.Minute}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--window":
			if i+1 >= len(args) || args[i+1] == "" {
				return nil, fmt.Errorf("--window requires a value")
			}
			window, err := time.ParseDuration(args[i+1])
			if err != nil {
				return nil, fmt.Errorf("window must be a duration: %w", err)
			}
			if window < time.Second {
				return nil, fmt.Errorf("window must be at least 1s")
			}
			if window > 24*time.Hour {
				return nil, fmt.Errorf("window must be at most 24h")
			}
			req.DrainWindow = window
			i++
		case "--tenant":
			if i+1 >= len(args) || args[i+1] == "" {
				return nil, fmt.Errorf("--tenant requires a value")
			}
			req.TenantID = args[i+1]
			i++
		case "--workspace":
			if i+1 >= len(args) || args[i+1] == "" {
				return nil, fmt.Errorf("--workspace requires a value")
			}
			req.WorkspaceID = args[i+1]
			i++
		default:
			return nil, fmt.Errorf("unknown jobs metrics option: %s", args[i])
		}
	}
	return req, nil
}

func parseBlockedJobsLimit(args []string) (int, error) {
	limit := defaultBlockedJobsCLIListLimit
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--limit":
			if i+1 >= len(args) {
				return 0, fmt.Errorf("--limit requires a value")
			}
			parsed, err := strconv.Atoi(args[i+1])
			if err != nil {
				return 0, fmt.Errorf("limit must be an integer: %w", err)
			}
			if parsed <= 0 {
				return 0, fmt.Errorf("limit must be greater than 0")
			}
			limit = parsed
			i++
		default:
			return 0, fmt.Errorf("unknown blocked jobs option: %s", args[i])
		}
	}
	return limit, nil
}

func printBlockedJobs(out io.Writer, jobs []*core.IngestJob) {
	if len(jobs) == 0 {
		writeln(out, "No blocked jobs.")
		return
	}
	writeln(out, "ID\tKIND\tTENANT\tWORKSPACE\tATTEMPTS\tUPDATED_AT\tLAST_ERROR")
	for _, job := range jobs {
		if job == nil {
			continue
		}
		writef(
			out,
			"%s\t%s\t%s\t%s\t%d\t%s\t%s\n",
			job.ID,
			job.JobKind,
			job.TenantID,
			job.WorkspaceID,
			job.Attempts,
			job.UpdatedAt.UTC().Format(time.RFC3339),
			optionalString(job.LastError),
		)
	}
}

func printJobMetrics(out io.Writer, metrics *core.JobBacklogMetrics) {
	if metrics == nil {
		writeln(out, "No job metrics.")
		return
	}
	writeln(out, "JOB BACKLOG")
	writef(out, "queued:       %d\n", metrics.Counts.Queued)
	writef(out, "ready queued: %d\n", metrics.Counts.ReadyQueued)
	writef(out, "running:      %d\n", metrics.Counts.Running)
	writef(out, "failed:       %d\n", metrics.Counts.Failed)
	writef(out, "blocked:      %d\n", metrics.Counts.Blocked)
	writef(out, "complete:     %d\n", metrics.Counts.Complete)
	writef(out, "retryable queued attempts: %d\n", metrics.RetryableQueuedAttempts)
	writeln(out)
	if metrics.OldestQueuedAgeSeconds == nil {
		writeln(out, "oldest queued age: unavailable")
	} else {
		writef(out, "oldest queued age: %s\n", (time.Duration(*metrics.OldestQueuedAgeSeconds) * time.Second).String())
	}
	if metrics.OldestRunningAgeSeconds == nil {
		writeln(out, "oldest running age: unavailable")
	} else {
		writef(out, "oldest running age: %s\n", (time.Duration(*metrics.OldestRunningAgeSeconds) * time.Second).String())
	}
	writef(out, "drain window:      %s\n", (time.Duration(metrics.DrainWindowSeconds) * time.Second).String())
	writef(out, "completed/window:  %d\n", metrics.CompletedInWindow)
	if metrics.DrainRateJobsPerMinute == nil {
		writeln(out, "drain rate:        unavailable")
	} else {
		writef(out, "drain rate:        %.2f jobs/min\n", *metrics.DrainRateJobsPerMinute)
	}
	if metrics.RecoveryETASeconds == nil {
		writeln(out, "recovery ETA:      unavailable")
	} else {
		writef(out, "recovery ETA:      %s\n", (time.Duration(*metrics.RecoveryETASeconds) * time.Second).String())
	}
	writef(out, "generated at:      %s\n", metrics.GeneratedAt.UTC().Format(time.RFC3339))
}

func optionalString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func writeln(out io.Writer, values ...any) {
	_, _ = fmt.Fprintln(out, values...)
}

func writef(out io.Writer, format string, values ...any) {
	_, _ = fmt.Fprintf(out, format, values...)
}

func openPostgresBlockedJobStore(ctx context.Context) (jobOperatorStore, func(), error) {
	cfg := config.LoadConfig()
	pool, err := db.NewPool(ctx, cfg)
	if err != nil {
		return nil, func() {}, err
	}
	return postgres.NewStore(pool), pool.Close, nil
}

func openPostgresService(ctx context.Context) (core.VibeGravityService, func(), error) {
	cfg := config.LoadConfig()
	pool, err := db.NewPool(ctx, cfg)
	if err != nil {
		return nil, func() {}, err
	}
	pgStore := postgres.NewStore(pool)
	ingestService, err := ingest.NewService(ingest.Dependencies{
		RawEvents: pgStore,
		Jobs:      pgStore,
	})
	if err != nil {
		pool.Close()
		return nil, func() {}, err
	}
	recallAssembler := recall.NewAssembler(recall.Dependencies{
		Notes:     pgStore,
		Plans:     pgStore,
		Memories:  pgStore,
		Documents: pgStore,
		Profiles:  pgStore,
		Summaries: pgStore,
		Groups:    pgStore,
		Freshness: recall.BacklogFreshnessProvider{Jobs: pgStore},
	})
	coreService, err := kernel.NewService(kernel.Dependencies{
		Ingest:      ingestService,
		Recall:      recallAssembler,
		Notes:       pgStore,
		Plans:       pgStore,
		Memories:    pgStore,
		Corrections: pgStore,
		Timeline:    pgStore,
		Documents:   pgStore,
	})
	if err != nil {
		pool.Close()
		return nil, func() {}, err
	}
	return coreService, pool.Close, nil
}

func printUsage(out io.Writer) {
	writeln(out, "Usage: cli <command>")
	writeln(out, "\nCommands:")
	writeln(out, "  doctor    Check system configuration and dependencies")
	writeln(out, "  eval      Run deterministic quality evals")
	writeln(out, "  hermes    Print Hermes bootstrap commands")
	writeln(out, "  jobs      Inspect and recover worker jobs")
	writeln(out, "  mcp       Serve the VibeGravity MCP protocol")
}

func printEvalUsage(out io.Writer) {
	writeln(out, "Usage: cli eval <command>")
	writeln(out)
	writeln(out, "Commands:")
	writeln(out, "  golden [--path FILE]  Run golden scenario regression evals")
	writeln(out, "  demo                  Run the local Hermes Memory trust-loop demo eval")
}

func printJobsUsage(out io.Writer) {
	writeln(out, "Usage: cli jobs <command>")
	writeln(out, "\nCommands:")
	writeln(out, "  metrics [--window D] [--tenant ID] [--workspace ID]  Show read-only worker backlog metrics")
	writeln(out, "  blocked [--limit N]       List blocked jobs without requeueing them")
	writeln(out, "  requeue-blocked <job_id>  Manually return one blocked job to the queued worker pool")
}

func printMCPUsage(out io.Writer) {
	writeln(out, "Usage: cli mcp serve [--stdio]")
	writeln(out)
	writeln(out, "Commands:")
	writeln(out, "  serve [--stdio]  Serve VibeGravity tools over MCP stdio")
}

func printHermesUsage(out io.Writer) {
	writeln(out, "Usage: cli hermes bootstrap [--name vibegravity] [--command /path/to/cli]")
	writeln(out)
	writeln(out, "Commands:")
	writeln(out, "  bootstrap  Print a Hermes MCP registration command for VibeGravity")
}

func shellQuote(value string) string {
	if value == "" {
		return "''"
	}
	if strings.IndexFunc(value, func(r rune) bool {
		return !(r >= 'A' && r <= 'Z') &&
			!(r >= 'a' && r <= 'z') &&
			!(r >= '0' && r <= '9') &&
			!strings.ContainsRune("@%_+=:,./-", r)
	}) == -1 {
		return value
	}
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
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

```



<!-- Source: cmd/cli/main_test.go | bytes=12120 | lines=350 | sha16=3f2f3715ff40be13 -->

```go
// ============================================================
// FILE     : cmd/cli/main_test.go
// PURPOSE  : Verifies operator CLI commands for blocked job inspection, MCP serving, and Hermes bootstrap.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : CLI tests
// DEPENDS  : bytes, context, strings, testing, time, internal/core
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: CLI tests must not open a real database or call external services.
// ============================================================

package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

func TestRunCLIListsBlockedJobs(t *testing.T) {
	t.Parallel()

	lastError := "not implemented: update_memory"
	store := &fakeBlockedJobStore{jobs: []*core.IngestJob{
		{
			ID:          "job_blocked_1",
			TenantID:    "tenant_1",
			WorkspaceID: "workspace_1",
			JobKind:     core.JobKindProcessTurnEvent,
			Status:      "blocked",
			Attempts:    4,
			LastError:   &lastError,
			UpdatedAt:   time.Date(2026, time.April, 24, 8, 0, 0, 0, time.UTC),
		},
	}}
	var out bytes.Buffer

	code := runCLI(context.Background(), []string{"jobs", "blocked", "--limit", "3"}, nil, &out, fakeStoreFactory(store), fakeServiceFactory(&fakeCLIService{}))

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; output: %s", code, out.String())
	}
	if store.listLimit != 3 {
		t.Fatalf("expected list limit 3, got %d", store.listLimit)
	}
	output := out.String()
	if !strings.Contains(output, "job_blocked_1") || !strings.Contains(output, "not implemented: update_memory") {
		t.Fatalf("expected blocked job details in output, got: %s", output)
	}
}

func TestRunCLIPrintsJobMetrics(t *testing.T) {
	t.Parallel()

	drainRate := 2.5
	recoveryETA := int64(240)
	oldestAge := int64(125)
	oldestRunningAge := int64(360)
	oldestAt := time.Date(2026, time.April, 24, 8, 10, 0, 0, time.UTC)
	oldestRunningAt := time.Date(2026, time.April, 24, 8, 6, 5, 0, time.UTC)
	store := &fakeBlockedJobStore{metrics: &core.JobBacklogMetrics{
		Counts: core.JobStatusCounts{
			Queued:      10,
			ReadyQueued: 7,
			Running:     2,
			Failed:      0,
			Blocked:     3,
			Complete:    50,
		},
		OldestQueuedAt:          &oldestAt,
		OldestQueuedAgeSeconds:  &oldestAge,
		OldestRunningAt:         &oldestRunningAt,
		OldestRunningAgeSeconds: &oldestRunningAge,
		DrainWindowSeconds:      600,
		CompletedInWindow:       25,
		DrainRateJobsPerMinute:  &drainRate,
		RecoveryETASeconds:      &recoveryETA,
		RetryableQueuedAttempts: 4,
		GeneratedAt:             time.Date(2026, time.April, 24, 8, 12, 5, 0, time.UTC),
	}}
	var out bytes.Buffer

	code := runCLI(context.Background(), []string{"jobs", "metrics", "--window", "10m", "--tenant", "tenant_1", "--workspace", "workspace_1"}, nil, &out, fakeStoreFactory(store), fakeServiceFactory(&fakeCLIService{}))

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; output: %s", code, out.String())
	}
	if store.metricsReq == nil {
		t.Fatalf("expected metrics request")
	}
	if store.metricsReq.DrainWindow != 10*time.Minute || store.metricsReq.TenantID != "tenant_1" || store.metricsReq.WorkspaceID != "workspace_1" {
		t.Fatalf("unexpected metrics request: %#v", store.metricsReq)
	}
	output := out.String()
	for _, want := range []string{
		"JOB BACKLOG",
		"queued:       10",
		"ready queued: 7",
		"running:      2",
		"blocked:      3",
		"retryable queued attempts: 4",
		"oldest running age: 6m0s",
		"drain rate:        2.50 jobs/min",
		"recovery ETA:      4m0s",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected output to contain %q, got: %s", want, output)
		}
	}
}

func TestRunCLIJobMetricsDefaultsWindow(t *testing.T) {
	t.Parallel()

	store := &fakeBlockedJobStore{metrics: &core.JobBacklogMetrics{}}
	var out bytes.Buffer

	code := runCLI(context.Background(), []string{"jobs", "metrics"}, nil, &out, fakeStoreFactory(store), fakeServiceFactory(&fakeCLIService{}))

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; output: %s", code, out.String())
	}
	if store.metricsReq == nil || store.metricsReq.DrainWindow != 15*time.Minute {
		t.Fatalf("expected default 15m metrics window, got %#v", store.metricsReq)
	}
}

func TestRunCLIRejectsInvalidMetricsWindow(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	code := runCLI(context.Background(), []string{"jobs", "metrics", "--window", "0s"}, nil, &out, fakeStoreFactory(&fakeBlockedJobStore{}), fakeServiceFactory(&fakeCLIService{}))

	if code == 0 {
		t.Fatalf("expected non-zero exit for invalid metrics window")
	}
	if !strings.Contains(out.String(), "window must be at least 1s") {
		t.Fatalf("expected invalid window message, got: %s", out.String())
	}
}

func TestRunCLIRequeuesBlockedJob(t *testing.T) {
	t.Parallel()

	store := &fakeBlockedJobStore{}
	var out bytes.Buffer

	code := runCLI(context.Background(), []string{"jobs", "requeue-blocked", "job_blocked_1"}, nil, &out, fakeStoreFactory(store), fakeServiceFactory(&fakeCLIService{}))

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; output: %s", code, out.String())
	}
	if store.requeuedJobID != "job_blocked_1" {
		t.Fatalf("expected job_blocked_1 to be requeued, got %q", store.requeuedJobID)
	}
	if !strings.Contains(out.String(), "requeued blocked job job_blocked_1") {
		t.Fatalf("unexpected output: %s", out.String())
	}
}

func TestRunCLIRejectsInvalidBlockedJobLimit(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	code := runCLI(context.Background(), []string{"jobs", "blocked", "--limit", "0"}, nil, &out, fakeStoreFactory(&fakeBlockedJobStore{}), fakeServiceFactory(&fakeCLIService{}))

	if code == 0 {
		t.Fatalf("expected non-zero exit for invalid limit")
	}
	if !strings.Contains(out.String(), "limit must be greater than 0") {
		t.Fatalf("expected invalid limit message, got: %s", out.String())
	}
}

func TestRunCLIReportsRequeueStoreError(t *testing.T) {
	t.Parallel()

	store := &fakeBlockedJobStore{requeueErr: core.ErrNotFound}
	var out bytes.Buffer

	code := runCLI(context.Background(), []string{"jobs", "requeue-blocked", "missing_job"}, nil, &out, fakeStoreFactory(store), fakeServiceFactory(&fakeCLIService{}))

	if code == 0 {
		t.Fatalf("expected non-zero exit for missing blocked job")
	}
	if !strings.Contains(out.String(), "blocked job not found") {
		t.Fatalf("expected not found message, got: %s", out.String())
	}
}

func TestRunCLIServesMCPStdio(t *testing.T) {
	t.Parallel()

	input := strings.NewReader(strings.Join([]string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"test","version":"0"}}}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`,
		"",
	}, "\n"))
	var out bytes.Buffer

	code := runCLI(context.Background(), []string{"mcp", "serve", "--stdio"}, input, &out, fakeStoreFactory(&fakeBlockedJobStore{}), fakeServiceFactory(&fakeCLIService{}))

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; output: %s", code, out.String())
	}
	if !strings.Contains(out.String(), `"protocolVersion":"2025-11-25"`) || !strings.Contains(out.String(), `"tools"`) {
		t.Fatalf("expected MCP initialize and tools/list responses, got: %s", out.String())
	}
}

func TestRunCLIRunsHermesMemoryDemoEval(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	code := runCLI(context.Background(), []string{"eval", "demo"}, strings.NewReader(""), &out, fakeStoreFactory(&fakeBlockedJobStore{}), fakeServiceFactory(&fakeCLIService{}))

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; output: %s", code, out.String())
	}
	output := out.String()
	for _, want := range []string{
		"demo initial recall shows rule plan and trust metadata",
		"demo explain shows recalled memory provenance",
		"demo next recall uses correction",
		"demo private scope separation",
		"Hermes Memory demo eval passed.",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected demo eval output to contain %q, got: %s", want, output)
		}
	}
}

func TestRunCLIHermesBootstrapPrintsRegistrationCommand(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	code := runCLI(context.Background(), []string{"hermes", "bootstrap", "--name", "vg", "--command", "/tmp/vibe cli"}, strings.NewReader(""), &out, fakeStoreFactory(&fakeBlockedJobStore{}), fakeServiceFactory(&fakeCLIService{}))

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; output: %s", code, out.String())
	}
	output := out.String()
	if !strings.Contains(output, "hermes mcp add vg --command '/tmp/vibe cli' --args mcp serve --stdio") {
		t.Fatalf("expected Hermes registration command, got: %s", output)
	}
	if !strings.Contains(output, "hermes mcp test vg") {
		t.Fatalf("expected verification command, got: %s", output)
	}
}

type fakeBlockedJobStore struct {
	jobs          []*core.IngestJob
	listLimit     int
	listErr       error
	requeuedJobID string
	requeueErr    error
	metrics       *core.JobBacklogMetrics
	metricsReq    *core.JobBacklogMetricsRequest
	metricsErr    error
}

func (s *fakeBlockedJobStore) ListBlockedJobs(_ context.Context, limit int) ([]*core.IngestJob, error) {
	s.listLimit = limit
	if s.listErr != nil {
		return nil, s.listErr
	}
	return s.jobs, nil
}

func (s *fakeBlockedJobStore) RequeueBlockedJob(_ context.Context, jobID string) error {
	s.requeuedJobID = jobID
	return s.requeueErr
}

func (s *fakeBlockedJobStore) GetJobBacklogMetrics(_ context.Context, req *core.JobBacklogMetricsRequest) (*core.JobBacklogMetrics, error) {
	s.metricsReq = req
	if s.metricsErr != nil {
		return nil, s.metricsErr
	}
	if s.metrics != nil {
		return s.metrics, nil
	}
	return &core.JobBacklogMetrics{}, nil
}

func fakeStoreFactory(store *fakeBlockedJobStore) jobOperatorStoreFactory {
	return func(context.Context) (jobOperatorStore, func(), error) {
		return store, func() {}, nil
	}
}

type fakeCLIService struct{}

func (s *fakeCLIService) Prefetch(context.Context, *core.PrefetchRequest) (*core.PrefetchResponse, error) {
	return &core.PrefetchResponse{Blocks: []core.RecallBlock{{Kind: "note", Text: "mcp cli ok"}}}, nil
}

func (s *fakeCLIService) SyncTurn(context.Context, *core.SyncTurnRequest) (*core.SyncTurnResponse, error) {
	return &core.SyncTurnResponse{Status: "accepted"}, nil
}

func (s *fakeCLIService) AddDocument(context.Context, *core.AddDocumentRequest) (*core.AddDocumentResponse, error) {
	return &core.AddDocumentResponse{Status: "created"}, nil
}

func (s *fakeCLIService) SearchMemories(context.Context, *core.SearchMemoriesRequest) (*core.SearchMemoriesResponse, error) {
	return &core.SearchMemoriesResponse{}, nil
}

func (s *fakeCLIService) SearchDocuments(context.Context, *core.SearchDocumentsRequest) (*core.SearchDocumentsResponse, error) {
	return &core.SearchDocumentsResponse{}, nil
}

func (s *fakeCLIService) AddNote(context.Context, *core.AddNoteRequest) (*core.AddNoteResponse, error) {
	return &core.AddNoteResponse{Status: "created"}, nil
}

func (s *fakeCLIService) CreatePlan(context.Context, *core.CreatePlanRequest) (*core.CreatePlanResponse, error) {
	return &core.CreatePlanResponse{Status: "created"}, nil
}

func (s *fakeCLIService) UpdatePlan(context.Context, *core.UpdatePlanRequest) (*core.UpdatePlanResponse, error) {
	return &core.UpdatePlanResponse{Status: "updated"}, nil
}

func (s *fakeCLIService) CorrectMemory(context.Context, *core.CorrectMemoryRequest) (*core.CorrectMemoryResponse, error) {
	return &core.CorrectMemoryResponse{Status: "recorded"}, nil
}

func (s *fakeCLIService) GetTimeline(context.Context, *core.GetTimelineRequest) (*core.GetTimelineResponse, error) {
	return &core.GetTimelineResponse{}, nil
}

func (s *fakeCLIService) ExplainMemory(context.Context, *core.ExplainMemoryRequest) (*core.ExplainMemoryResponse, error) {
	return &core.ExplainMemoryResponse{}, nil
}

func fakeServiceFactory(service core.VibeGravityService) serviceFactory {
	return func(context.Context) (core.VibeGravityService, func(), error) {
		return service, func() {}, nil
	}
}

```



<!-- Source: cmd/server/main.go | bytes=3348 | lines=119 | sha16=a0d456034c689678 -->

```go
// ============================================================
// FILE     : cmd/server/main.go
// PURPOSE  : Starts the HTTP API process and wires runtime dependencies.
// LAYER    : interface
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : main
// DEPENDS  : internal/config, internal/db, internal/httpapi, internal/ingest, internal/kernel, internal/recall, internal/store/postgres
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
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/config"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/db"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/httpapi"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/ingest"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/kernel"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/recall"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/store/postgres"
)

func main() {
	log.Println("Starting VibeGravity API Server...")

	cfg := config.LoadConfig()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := db.NewPool(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	pgStore := postgres.NewStore(pool)
	ingestService, err := ingest.NewService(ingest.Dependencies{
		RawEvents: pgStore,
		Jobs:      pgStore,
	})
	if err != nil {
		log.Fatalf("Failed to initialize ingest service: %v", err)
	}
	recallAssembler := recall.NewAssembler(recall.Dependencies{
		Notes:     pgStore,
		Plans:     pgStore,
		Memories:  pgStore,
		Documents: pgStore,
		Profiles:  pgStore,
		Summaries: pgStore,
		Groups:    pgStore,
		Freshness: recall.BacklogFreshnessProvider{Jobs: pgStore},
	})
	coreService, err := kernel.NewService(kernel.Dependencies{
		Ingest:      ingestService,
		Recall:      recallAssembler,
		Notes:       pgStore,
		Plans:       pgStore,
		Memories:    pgStore,
		Corrections: pgStore,
		Timeline:    pgStore,
		Documents:   pgStore,
	})
	if err != nil {
		log.Fatalf("Failed to initialize VibeGravity service: %v", err)
	}

	app := &httpapi.App{
		Service: coreService,
		DBPool:  pool,
	}

	router := httpapi.NewRouter(app)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
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

```



<!-- Source: cmd/worker/main.go | bytes=4354 | lines=151 | sha16=8accb3b24c8382c8 -->

```go
// ============================================================
// FILE     : cmd/worker/main.go
// PURPOSE  : Starts the background worker process for ingest jobs and maintenance.
// LAYER    : interface
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : main
// DEPENDS  : internal/config, internal/db, internal/graph, internal/reasoning, internal/store/postgres, internal/worker
// USED_BY  : Makefile, worker deployments
// ------------------------------------------------------------
// AGENT_NOTE: Keep Codex and embedding work off the API hot path.
// ============================================================

// Package main starts the VibeGravity background worker process.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/config"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/db"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/graph"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/reasoning"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/store/postgres"
	"github.com/parker-jungwoo-hwang/vibegravity/internal/worker"
)

func main() {
	log.Println("Starting VibeGravity Background Worker...")

	cfg := config.LoadConfig()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := db.NewPool(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	pgStore := postgres.NewStore(pool)
	applyEngine, err := graph.NewStoreBackedApplyEngine(pgStore)
	if err != nil {
		log.Fatalf("Failed to initialize graph apply engine: %v", err)
	}
	dreamingService, err := graph.NewDreamingService(graph.DreamingDependencies{
		Store: pgStore,
	})
	if err != nil {
		log.Fatalf("Failed to initialize dreaming service: %v", err)
	}
	stage2InputPreparer := worker.NewStoreBackedStage2InputPreparer(worker.Stage2SourceStores{
		Profiles:  pgStore,
		Memories:  pgStore,
		Documents: pgStore,
		Plans:     pgStore,
		Notes:     pgStore,
		Groups:    pgStore,
	})
	reasoner, err := newReasoner(stage2InputPreparer)
	if err != nil {
		log.Fatalf("Failed to initialize reasoning orchestrator: %v", err)
	}
	processor, err := worker.NewProcessor(worker.Dependencies{
		WorkerID:    workerID(),
		Jobs:        pgStore,
		RawEvents:   pgStore,
		Reasoner:    reasoner,
		ApplyEngine: applyEngine,
		Dreaming:    dreamingService,
	})
	if err != nil {
		log.Fatalf("Failed to initialize worker processor: %v", err)
	}

	log.Println("Worker is running. Waiting for jobs...")

	// Setup graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down Worker gracefully...")

		cancel()
	}()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Worker stopped.")
			return
		case <-ticker.C:
			result, err := processor.RunOnce(ctx)
			if err != nil {
				log.Printf("Worker pass completed with error: %v", err)
			}
			if result.Claimed == 0 {
				log.Println("Worker is idle.")
				continue
			}
			log.Printf(
				"Worker pass: claimed=%d completed=%d failed=%d blocked=%d applied_operations=%d memory_ids=%d traces_written=%d session_dreams=%d workspace_dreams=%d",
				result.Claimed,
				result.Completed,
				result.Failed,
				result.Blocked,
				result.AppliedOperationCount,
				result.MemoryIDCount,
				result.TraceWrittenCount,
				result.SessionDreamCount,
				result.WorkspaceDreamCount,
			)
		}
	}
}

func workerID() string {
	if value := os.Getenv("VIBEGRAVITY_WORKER_ID"); value != "" {
		return value
	}
	host, err := os.Hostname()
	if err != nil || host == "" {
		return "worker:local"
	}
	return "worker:" + host
}

func newReasoner(stage2InputPreparer *reasoning.Stage2InputPreparer) (reasoning.Orchestrator, error) {
	mockCodex := reasoning.NewMockCodexJSONClient()
	bridgeConfig := reasoning.CodexBridgeConfig{Enabled: true}
	stage1, err := reasoning.NewCodexStage1Extractor(bridgeConfig, mockCodex)
	if err != nil {
		return nil, err
	}
	stage2, err := reasoning.NewCodexStage2Resolver(bridgeConfig, mockCodex)
	if err != nil {
		return nil, err
	}
	return reasoning.NewPipelineOrchestrator(stage1, stage2, stage2InputPreparer)
}

```



<!-- Source: internal/config/config.go | bytes=2667 | lines=94 | sha16=7f8c2c953c20ec6c -->

```go
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
	Enabled  bool
	Endpoint string
	Model    string
}

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
			Enabled:  getEnvAsBool("VIBEGRAVITY_CODEX_ENABLED", false),
			Endpoint: getEnv("VIBEGRAVITY_CODEX_ENDPOINT", ""),
			Model:    getEnv("VIBEGRAVITY_CODEX_MODEL", ""),
		},
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

```



<!-- Source: internal/core/doc.go | bytes=700 | lines=16 | sha16=78a8a9146f43a578 -->

```go
// ============================================================
// FILE     : internal/core/doc.go
// PURPOSE  : Provides package documentation for the core domain contract.
// LAYER    : domain
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : package core
// DEPENDS  : plans/02_product-contract_and_direction.md
// USED_BY  : Go documentation, internal/core
// ------------------------------------------------------------
// AGENT_NOTE: Keep this package summary aligned with the product contract.
// ============================================================

// Package core defines VibeGravity's domain contract and service DTOs.
package core

```



<!-- Source: internal/core/document.go | bytes=2068 | lines=50 | sha16=75e4b417e32a7cbf -->

```go
// ============================================================
// FILE     : internal/core/document.go
// PURPOSE  : Defines document and chunk records used by document retrieval.
// LAYER    : domain
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : Document, DocumentChunk
// DEPENDS  : encoding/json, time
// USED_BY  : internal/store, internal/core/dto.go, recall and search paths
// ------------------------------------------------------------
// AGENT_NOTE: Keep documents separate from derived memories and memory trace.
// ============================================================

package core

import (
	"encoding/json"
	"time"
)

// Document is a deduplicated document-level source artifact.
type Document struct {
	ID           string          `json:"id"`
	TenantID     string          `json:"tenant_id"`
	WorkspaceID  string          `json:"workspace_id"`
	Source       string          `json:"source"`
	Title        string          `json:"title"`
	Fingerprint  string          `json:"fingerprint"`
	MetadataJSON json.RawMessage `json:"metadata_json"`
	VersionHint  string          `json:"version_hint"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// DocumentChunk is a searchable retrieval unit for a document.
type DocumentChunk struct {
	ID                 string          `json:"id"`
	DocumentID         string          `json:"document_id"`
	ChunkIndex         int             `json:"chunk_index"`
	Text               string          `json:"text"`
	HeadingPath        string          `json:"heading_path"`
	MetadataJSON       json.RawMessage `json:"metadata_json"`
	NeighborChunkIDs   []string        `json:"neighbor_chunk_ids"`
	EmbeddingModel     string          `json:"embedding_model"`
	EmbeddingDims      int             `json:"embedding_dims"`
	EmbeddingUpdatedAt *time.Time      `json:"embedding_updated_at,omitempty"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}

```



<!-- Source: internal/core/dreaming.go | bytes=3287 | lines=81 | sha16=ccc313acb4fda789 -->

```go
// ============================================================
// FILE     : internal/core/dreaming.go
// PURPOSE  : Defines background dreaming requests, inputs, and promotion results.
// LAYER    : domain
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : DreamingTier, DreamSessionRequest, DreamWorkspaceRequest, DreamingSessionInput, DreamingPromotionRequest, DreamingPromotionResult, DreamingResult
// DEPENDS  : time, internal/core/memory.go
// USED_BY  : graph dreaming service, worker, storage
// ------------------------------------------------------------
// AGENT_NOTE: Dreaming changes consolidation metadata and summaries without blurring scope boundaries.
// ============================================================

package core

import "time"

// DreamingTier describes consolidation depth for a memory.
type DreamingTier string

const (
	// DreamingTierShortTerm is recent scratch or raw-tail material.
	DreamingTierShortTerm DreamingTier = "short-term"
	// DreamingTierMidTerm is session or active-topic consolidation.
	DreamingTierMidTerm DreamingTier = "mid-term"
	// DreamingTierLongTerm is stable reusable memory.
	DreamingTierLongTerm DreamingTier = "long-term"
	// DreamingTierUltraLongTerm is canonical, repeatedly confirmed memory.
	DreamingTierUltraLongTerm DreamingTier = "ultra-long-term"
)

// DreamSessionRequest asks dreaming to consolidate one session.
type DreamSessionRequest struct {
	JobID       string    `json:"job_id"`
	TenantID    string    `json:"tenant_id"`
	WorkspaceID string    `json:"workspace_id"`
	SessionID   string    `json:"session_id"`
	Now         time.Time `json:"now"`
}

// DreamWorkspaceRequest asks dreaming to promote stable workspace memories.
type DreamWorkspaceRequest struct {
	JobID       string    `json:"job_id"`
	TenantID    string    `json:"tenant_id"`
	WorkspaceID string    `json:"workspace_id"`
	Now         time.Time `json:"now"`
}

// DreamingSessionInput is the source material for session consolidation.
type DreamingSessionInput struct {
	RawEventIDs []string  `json:"raw_event_ids"`
	Memories    []*Memory `json:"memories"`
}

// DreamingPromotionRequest selects existing memories for tier promotion.
type DreamingPromotionRequest struct {
	JobID             string       `json:"job_id"`
	TenantID          string       `json:"tenant_id"`
	WorkspaceID       string       `json:"workspace_id"`
	SessionID         string       `json:"session_id,omitempty"`
	MemoryIDs         []string     `json:"memory_ids,omitempty"`
	Tier              DreamingTier `json:"tier"`
	MinConfidence     float64      `json:"min_confidence"`
	RequireStableKind bool         `json:"require_stable_kind"`
	Now               time.Time    `json:"now"`
}

// DreamingPromotionResult reports metadata-only memory promotions.
type DreamingPromotionResult struct {
	PromotedCount int      `json:"promoted_count"`
	MemoryIDs     []string `json:"memory_ids"`
}

// DreamingResult reports one background dreaming job outcome.
type DreamingResult struct {
	SessionSummaryWritten bool `json:"session_summary_written"`
	MidTermPromoted       int  `json:"mid_term_promoted"`
	LongTermPromoted      int  `json:"long_term_promoted"`
	UltraLongTermPromoted int  `json:"ultra_long_term_promoted"`
}

```



<!-- Source: internal/core/dto.go | bytes=13099 | lines=339 | sha16=030f5bbfc21cab80 -->

```go
// ============================================================
// FILE     : internal/core/dto.go
// PURPOSE  : Defines v1 request and response DTOs shared by runtime surfaces.
// LAYER    : domain
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : PrefetchRequest, SyncTurnRequest, search/note/plan/memory DTOs
// DEPENDS  : encoding/json, time, plans/05_runtime-contracts_ingest-recall-apply.md
// USED_BY  : internal/core/service.go, internal/httpapi, tests
// ------------------------------------------------------------
// AGENT_NOTE: Keep DTO changes synchronized with runtime contract docs.
// ============================================================

package core

import (
	"encoding/json"
	"time"
)

// PrefetchRequest asks the recall assembler for a typed next-turn recall pack.
type PrefetchRequest struct {
	TenantID     string `json:"tenant_id"`
	WorkspaceID  string `json:"workspace_id"`
	SessionID    string `json:"session_id"`
	ActorID      string `json:"actor_id"`
	Query        string `json:"query"`
	BudgetTokens int    `json:"budget_tokens"`
	Mode         string `json:"mode"`
}

// PrefetchResponse returns budget-aware typed recall blocks.
type PrefetchResponse struct {
	Blocks []RecallBlock `json:"blocks"`
	Meta   RecallMeta    `json:"meta"`
}

// RecallBlock is one typed item in a recall pack.
type RecallBlock struct {
	ID            string      `json:"id,omitempty"`
	Kind          string      `json:"kind"`
	Priority      int         `json:"priority"`
	Text          string      `json:"text"`
	Scope         MemoryScope `json:"scope,omitempty"`
	Source        string      `json:"source,omitempty"`
	SourceID      string      `json:"source_id,omitempty"`
	Status        string      `json:"status,omitempty"`
	Freshness     string      `json:"freshness,omitempty"`
	OwnerEntityID string      `json:"owner_entity_id,omitempty"`
}

// RecallMeta describes recall assembly and token budget metadata.
type RecallMeta struct {
	EstimatedTokens     int      `json:"estimated_tokens"`
	Sources             []string `json:"sources"`
	Freshness           string   `json:"freshness,omitempty"`
	FreshnessLagSeconds *int64   `json:"freshness_lag_seconds,omitempty"`
	Degraded            bool     `json:"degraded"`
	DegradedReasons     []string `json:"degraded_reasons,omitempty"`
}

// SyncTurnRequest records a complete turn through the hot ingest path.
type SyncTurnRequest struct {
	TenantID       string            `json:"tenant_id"`
	WorkspaceID    string            `json:"workspace_id"`
	SessionID      string            `json:"session_id"`
	ActorID        string            `json:"actor_id"`
	IdempotencyKey string            `json:"idempotency_key"`
	TurnEvents     []RawEventPayload `json:"turn_events"`
}

// RawEventPayload is the event payload accepted by SyncTurn.
type RawEventPayload struct {
	EventKind   string          `json:"event_kind"`
	Source      string          `json:"source"`
	Fingerprint string          `json:"fingerprint"`
	OccurredAt  time.Time       `json:"occurred_at"`
	PayloadJSON json.RawMessage `json:"payload_json"`
}

// SyncTurnResponse acknowledges accepted raw events and queued jobs.
type SyncTurnResponse struct {
	Status         string   `json:"status"`
	SessionID      string   `json:"session_id"`
	EventIDs       []string `json:"event_ids"`
	JobIDs         []string `json:"job_ids"`
	DuplicateCount int      `json:"duplicate_count"`
}

// AddDocumentRequest adds a document for chunking and document search.
type AddDocumentRequest struct {
	TenantID     string          `json:"tenant_id"`
	WorkspaceID  string          `json:"workspace_id"`
	Source       string          `json:"source"`
	Title        string          `json:"title"`
	Content      string          `json:"content"`
	Fingerprint  string          `json:"fingerprint"`
	MetadataJSON json.RawMessage `json:"metadata_json"`
	VersionHint  string          `json:"version_hint"`
}

// AddDocumentResponse reports the created document and chunk identifiers.
type AddDocumentResponse struct {
	DocumentID string   `json:"document_id"`
	ChunkIDs   []string `json:"chunk_ids"`
	Status     string   `json:"status"`
}

// SearchMemoriesRequest searches memories with scope and artifact-class filters.
type SearchMemoriesRequest struct {
	TenantID        string          `json:"tenant_id"`
	WorkspaceID     string          `json:"workspace_id"`
	OwnerEntityID   string          `json:"owner_entity_id,omitempty"`
	VisibleGroupIDs []string        `json:"visible_group_ids,omitempty"`
	Query           string          `json:"query"`
	Scopes          []MemoryScope   `json:"scopes"`
	ArtifactClasses []ArtifactClass `json:"artifact_classes"`
}

// SearchMemoriesResponse returns matching memory records.
type SearchMemoriesResponse struct {
	Memories []MemoryResult `json:"memories"`
}

// MemoryResult is a recall-safe search result for a memory.
type MemoryResult struct {
	MemoryID      string        `json:"memory_id"`
	Kind          MemoryKind    `json:"kind"`
	ArtifactClass ArtifactClass `json:"artifact_class"`
	Text          string        `json:"text"`
	Confidence    float64       `json:"confidence"`
	Scope         MemoryScope   `json:"scope"`
	GroupID       *string       `json:"group_id,omitempty"`
	OwnerEntityID string        `json:"owner_entity_id,omitempty"`
	ValidFrom     time.Time     `json:"valid_from"`
	LatestFlag    bool          `json:"latest_flag"`
}

// SearchDocumentsRequest searches document chunks.
type SearchDocumentsRequest struct {
	TenantID    string `json:"tenant_id"`
	WorkspaceID string `json:"workspace_id"`
	Query       string `json:"query"`
}

// SearchDocumentsResponse returns matching document chunks.
type SearchDocumentsResponse struct {
	Chunks []DocumentChunkResult `json:"chunks"`
}

// DocumentChunkResult is a search result for a document retrieval unit.
type DocumentChunkResult struct {
	ChunkID    string  `json:"chunk_id"`
	DocumentID string  `json:"document_id"`
	Text       string  `json:"text"`
	Score      float64 `json:"score"`
}

// AddNoteRequest creates a human-authored note.
type AddNoteRequest struct {
	TenantID      string      `json:"tenant_id"`
	WorkspaceID   string      `json:"workspace_id"`
	NoteKind      string      `json:"note_kind"`
	Scope         MemoryScope `json:"scope"`
	OwnerEntityID string      `json:"owner_entity_id"`
	Text          string      `json:"text"`
	Pinned        bool        `json:"pinned"`
	ExpiresAt     *time.Time  `json:"expires_at,omitempty"`
}

// AddNoteResponse reports the created note.
type AddNoteResponse struct {
	NoteID string `json:"note_id"`
	Status string `json:"status"`
}

// ListPinnedNotesRequest loads visible pinned notes for recall or Stage 2 context.
type ListPinnedNotesRequest struct {
	TenantID      string        `json:"tenant_id"`
	WorkspaceID   string        `json:"workspace_id"`
	OwnerEntityID string        `json:"owner_entity_id,omitempty"`
	Scopes        []MemoryScope `json:"scopes"`
}

// PlanItemInput is an item payload used when creating or updating a plan.
type PlanItemInput struct {
	ID           string          `json:"id,omitempty"`
	Title        string          `json:"title"`
	Status       string          `json:"status"`
	EvidenceJSON json.RawMessage `json:"evidence_json"`
}

// CreatePlanRequest creates a structured plan.
type CreatePlanRequest struct {
	TenantID      string          `json:"tenant_id"`
	WorkspaceID   string          `json:"workspace_id"`
	Title         string          `json:"title"`
	Status        string          `json:"status"`
	Scope         MemoryScope     `json:"scope"`
	OwnerEntityID string          `json:"owner_entity_id"`
	EvidenceJSON  json.RawMessage `json:"evidence_json"`
	Items         []PlanItemInput `json:"items"`
}

// CreatePlanResponse reports the created plan and item identifiers.
type CreatePlanResponse struct {
	PlanID  string   `json:"plan_id"`
	ItemIDs []string `json:"item_ids"`
	Status  string   `json:"status"`
}

// GetActivePlansRequest loads visible active plans for recall or Stage 2 context.
type GetActivePlansRequest struct {
	TenantID      string        `json:"tenant_id"`
	WorkspaceID   string        `json:"workspace_id"`
	OwnerEntityID string        `json:"owner_entity_id,omitempty"`
	Scopes        []MemoryScope `json:"scopes"`
}

// UpdatePlanRequest updates a structured plan.
type UpdatePlanRequest struct {
	TenantID     string          `json:"tenant_id"`
	WorkspaceID  string          `json:"workspace_id"`
	PlanID       string          `json:"plan_id"`
	Title        *string         `json:"title,omitempty"`
	Status       *string         `json:"status,omitempty"`
	EvidenceJSON json.RawMessage `json:"evidence_json"`
	Items        []PlanItemInput `json:"items"`
}

// UpdatePlanResponse reports the updated plan.
type UpdatePlanResponse struct {
	PlanID string `json:"plan_id"`
	Status string `json:"status"`
}

// CorrectMemoryRequest records a human correction for a memory.
type CorrectMemoryRequest struct {
	TenantID       string          `json:"tenant_id"`
	WorkspaceID    string          `json:"workspace_id"`
	MemoryID       string          `json:"memory_id"`
	OperatorID     string          `json:"operator_id"`
	IdempotencyKey string          `json:"idempotency_key"`
	CorrectionText string          `json:"correction_text"`
	EvidenceJSON   json.RawMessage `json:"evidence_json"`
}

// CorrectMemoryResponse reports the correction side effects.
type CorrectMemoryResponse struct {
	MemoryID           string `json:"memory_id"`
	RawEventID         string `json:"raw_event_id"`
	CorrectionID       string `json:"correction_id,omitempty"`
	CorrectionRecorded bool   `json:"correction_recorded"`
	TraceWritten       bool   `json:"trace_written"`
	Status             string `json:"status"`
}

// GetTimelineRequest asks for a timeline view assembled from existing artifacts.
type GetTimelineRequest struct {
	TenantID    string        `json:"tenant_id"`
	WorkspaceID string        `json:"workspace_id"`
	Scopes      []MemoryScope `json:"scopes"`
	EntityID    string        `json:"entity_id"`
	From        *time.Time    `json:"from,omitempty"`
	To          *time.Time    `json:"to,omitempty"`
	Limit       int           `json:"limit"`
}

// GetTimelineResponse returns timeline view items without requiring a cache table.
type GetTimelineResponse struct {
	Items []TimelineItem `json:"items"`
}

// TimelineItem is one row in the timeline view.
type TimelineItem struct {
	ID            string        `json:"id"`
	Kind          MemoryKind    `json:"kind"`
	ArtifactClass ArtifactClass `json:"artifact_class"`
	Text          string        `json:"text"`
	OccurredAt    time.Time     `json:"occurred_at"`
	MemoryID      string        `json:"memory_id"`
	RawEventID    string        `json:"raw_event_id"`
}

// ExplainMemoryRequest asks for provenance for one memory.
type ExplainMemoryRequest struct {
	TenantID        string   `json:"tenant_id"`
	WorkspaceID     string   `json:"workspace_id"`
	MemoryID        string   `json:"memory_id"`
	EntityID        string   `json:"entity_id,omitempty"`
	VisibleGroupIDs []string `json:"visible_group_ids,omitempty"`
}

// ExplainMemoryResponse returns trace, edges, and source event evidence.
type ExplainMemoryResponse struct {
	MemoryID     string                   `json:"memory_id"`
	Trace        MemoryTraceResult        `json:"trace"`
	Edges        []MemoryEdgeResult       `json:"edges"`
	SourceEvents []ProvenanceEventResult  `json:"source_events"`
	Documents    []ProvenanceDocumentLink `json:"documents"`
}

// MemoryTraceResult is the DTO shape for memory provenance.
type MemoryTraceResult struct {
	RawEventIDs            []string        `json:"raw_event_ids"`
	ReasoningJobID         string          `json:"reasoning_job_id"`
	ReasoningStage         string          `json:"reasoning_stage"`
	CandidateSnapshotJSON  json.RawMessage `json:"candidate_snapshot_json"`
	AppliedOperationsJSON  json.RawMessage `json:"applied_operations_json"`
	OperatorCorrectionFlag bool            `json:"operator_correction_flag"`
	RelatedDocumentIDs     []string        `json:"related_document_ids"`
	CreatedAt              time.Time       `json:"created_at"`
}

// MemoryEdgeResult is the DTO shape for one memory edge.
type MemoryEdgeResult struct {
	FromMemoryID string    `json:"from_memory_id"`
	ToMemoryID   string    `json:"to_memory_id"`
	EdgeKind     EdgeKind  `json:"edge_kind"`
	Confidence   float64   `json:"confidence"`
	CreatedAt    time.Time `json:"created_at"`
}

// ProvenanceEventResult summarizes a source raw event for provenance display.
type ProvenanceEventResult struct {
	EventID     string          `json:"event_id"`
	EventKind   string          `json:"event_kind"`
	Source      string          `json:"source"`
	Fingerprint string          `json:"fingerprint"`
	OccurredAt  time.Time       `json:"occurred_at"`
	PayloadJSON json.RawMessage `json:"payload_json"`
}

// ProvenanceDocumentLink summarizes a document used as memory evidence.
type ProvenanceDocumentLink struct {
	DocumentID string `json:"document_id"`
	Title      string `json:"title"`
}

```



<!-- Source: internal/core/entity.go | bytes=1171 | lines=32 | sha16=ef308696a68a6bfd -->

```go
// ============================================================
// FILE     : internal/core/entity.go
// PURPOSE  : Defines entity records for users, agents, workspaces, projects, and groups.
// LAYER    : domain
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : Entity
// DEPENDS  : encoding/json, time
// USED_BY  : group membership, profile, and scope-aware storage paths
// ------------------------------------------------------------
// AGENT_NOTE: Preserve tenant and workspace fields on every persisted entity.
// ============================================================

package core

import (
	"encoding/json"
	"time"
)

// Entity represents a user, agent, workspace, project, or group.
type Entity struct {
	ID           string          `json:"id"`
	TenantID     string          `json:"tenant_id"`
	WorkspaceID  string          `json:"workspace_id"`
	EntityKind   string          `json:"entity_kind"`
	DisplayName  string          `json:"display_name"`
	MetadataJSON json.RawMessage `json:"metadata_json"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

```



<!-- Source: internal/core/errors.go | bytes=1297 | lines=32 | sha16=e9368e3ff71d9a4a -->

```go
// ============================================================
// FILE     : internal/core/errors.go
// PURPOSE  : Defines shared domain errors for service and storage boundaries.
// LAYER    : domain
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : ErrNotFound, ErrDuplicate, ErrInvalidArgument, ErrConflict, ErrNotImplemented
// DEPENDS  : errors
// USED_BY  : internal/core services, internal/store implementations
// ------------------------------------------------------------
// AGENT_NOTE: Prefer these sentinel errors over transport-specific failures.
// ============================================================

package core

import "errors"

// ErrNotFound reports that a requested record does not exist.
var ErrNotFound = errors.New("not found")

// ErrDuplicate reports that an idempotent write already exists.
var ErrDuplicate = errors.New("duplicate")

// ErrInvalidArgument reports a contract validation failure.
var ErrInvalidArgument = errors.New("invalid argument")

// ErrConflict reports that a request would violate storage invariants.
var ErrConflict = errors.New("conflict")

// ErrNotImplemented reports that a contract exists but the behavior has not landed yet.
var ErrNotImplemented = errors.New("not implemented")

```



<!-- Source: internal/core/group.go | bytes=1200 | lines=34 | sha16=317648488f4a0e3b -->

```go
// ============================================================
// FILE     : internal/core/group.go
// PURPOSE  : Defines group shared memory records and memberships.
// LAYER    : domain
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : MemoryGroup, MemoryGroupMembership
// DEPENDS  : time
// USED_BY  : internal/store, scope-aware recall and apply paths
// ------------------------------------------------------------
// AGENT_NOTE: group_shared memory requires valid membership before visibility.
// ============================================================

package core

import "time"

// MemoryGroup defines a named group for group-shared memory.
type MemoryGroup struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	WorkspaceID string    `json:"workspace_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// MemoryGroupMembership links an entity to a memory group.
type MemoryGroupMembership struct {
	GroupID   string    `json:"group_id"`
	EntityID  string    `json:"entity_id"`
	CreatedAt time.Time `json:"created_at"`
}

```



<!-- Source: internal/core/job.go | bytes=3133 | lines=71 | sha16=aeee01a0b530e277 -->

```go
// ============================================================
// FILE     : internal/core/job.go
// PURPOSE  : Defines PostgreSQL-backed worker queue job records.
// LAYER    : domain
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : IngestJob, JobBacklogMetricsRequest, JobStatusCounts, JobBacklogMetrics
// DEPENDS  : encoding/json, time, internal/core/kind.go
// USED_BY  : internal/store, cmd/worker, ingest pipeline
// ------------------------------------------------------------
// AGENT_NOTE: Jobs must support retry without duplicate apply side effects.
// ============================================================

package core

import (
	"encoding/json"
	"time"
)

// IngestJob is a PostgreSQL-backed worker queue item.
type IngestJob struct {
	ID          string          `json:"id"`
	TenantID    string          `json:"tenant_id"`
	WorkspaceID string          `json:"workspace_id"`
	JobKind     JobKind         `json:"job_kind"`
	Status      string          `json:"status"`
	RawEventIDs []string        `json:"raw_event_ids"`
	PayloadJSON json.RawMessage `json:"payload_json"`
	Attempts    int             `json:"attempts"`
	AvailableAt time.Time       `json:"available_at"`
	LockedBy    *string         `json:"locked_by,omitempty"`
	LockedAt    *time.Time      `json:"locked_at,omitempty"`
	LastError   *string         `json:"last_error,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// JobBacklogMetricsRequest scopes read-only worker backlog metrics for operators.
type JobBacklogMetricsRequest struct {
	TenantID     string        `json:"tenant_id,omitempty"`
	WorkspaceID  string        `json:"workspace_id,omitempty"`
	DrainWindow  time.Duration `json:"drain_window,omitempty"`
	GeneratedNow time.Time     `json:"generated_now,omitempty"`
}

// JobStatusCounts summarizes ingest job rows by durable queue status.
type JobStatusCounts struct {
	Queued      int `json:"queued"`
	ReadyQueued int `json:"ready_queued"`
	Running     int `json:"running"`
	Failed      int `json:"failed"`
	Blocked     int `json:"blocked"`
	Complete    int `json:"complete"`
}

// JobBacklogMetrics is a read-only operator view of worker queue health.
type JobBacklogMetrics struct {
	Counts                  JobStatusCounts `json:"counts"`
	OldestQueuedAt          *time.Time      `json:"oldest_queued_at,omitempty"`
	OldestQueuedAgeSeconds  *int64          `json:"oldest_queued_age_seconds,omitempty"`
	OldestRunningAt         *time.Time      `json:"oldest_running_at,omitempty"`
	OldestRunningAgeSeconds *int64          `json:"oldest_running_age_seconds,omitempty"`
	DrainWindowSeconds      int64           `json:"drain_window_seconds"`
	CompletedInWindow       int             `json:"completed_in_window"`
	DrainRateJobsPerMinute  *float64        `json:"drain_rate_jobs_per_minute,omitempty"`
	RecoveryETASeconds      *int64          `json:"recovery_eta_seconds,omitempty"`
	RetryableQueuedAttempts int             `json:"retryable_queued_attempts"`
	GeneratedAt             time.Time       `json:"generated_at"`
}

```



<!-- Source: internal/core/kind.go | bytes=5271 | lines=115 | sha16=89b01eb4bd7f1f10 -->

```go
// ============================================================
// FILE     : internal/core/kind.go
// PURPOSE  : Defines canonical enum-like values for memory, edge, job, and artifact classes.
// LAYER    : domain
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : MemoryKind, EdgeKind, MemoryStatus, JobKind, ArtifactClass
// DEPENDS  : plans/02_product-contract_and_direction.md, docs/adr-005-artifact-class-timing.md
// USED_BY  : internal/core records, storage, reasoning/apply contracts
// ------------------------------------------------------------
// AGENT_NOTE: Treat value changes as contract changes and update docs/tests.
// ============================================================

package core

// MemoryKind describes the semantic type of a derived memory.
type MemoryKind string

const (
	// MemoryKindFact is a comparatively verifiable fact.
	MemoryKindFact MemoryKind = "fact"
	// MemoryKindPreference records a like, dislike, or preference.
	MemoryKindPreference MemoryKind = "preference"
	// MemoryKindTrait records a durable tendency or characteristic.
	MemoryKindTrait MemoryKind = "trait"
	// MemoryKindGoal records an intended outcome.
	MemoryKindGoal MemoryKind = "goal"
	// MemoryKindConstraint records a limiting rule or requirement.
	MemoryKindConstraint MemoryKind = "constraint"
	// MemoryKindRelationship records a person, team, project, or object relation.
	MemoryKindRelationship MemoryKind = "relationship"
	// MemoryKindDecision records a decision that has been made.
	MemoryKindDecision MemoryKind = "decision"
	// MemoryKindProcedure records a reusable way of working.
	MemoryKindProcedure MemoryKind = "procedure"
	// MemoryKindTaskState records current task state.
	MemoryKindTaskState MemoryKind = "task_state"
	// MemoryKindDocFact records a fact extracted from a document.
	MemoryKindDocFact MemoryKind = "doc_fact"
	// MemoryKindSummary records a compressed session or topic summary.
	MemoryKindSummary MemoryKind = "summary"
	// MemoryKindHypothesis records an uncertain inference.
	MemoryKindHypothesis MemoryKind = "hypothesis"
	// MemoryKindCorrection records an operator correction intent in timeline views.
	MemoryKindCorrection MemoryKind = "correction"
)

// EdgeKind describes the relationship between two memories or artifacts.
type EdgeKind string

const (
	// EdgeKindUpdates means the source memory supersedes the target memory.
	EdgeKindUpdates EdgeKind = "updates"
	// EdgeKindExtends means the source memory adds detail to the target memory.
	EdgeKindExtends EdgeKind = "extends"
	// EdgeKindSupports means the source memory strengthens the target memory.
	EdgeKindSupports EdgeKind = "supports"
	// EdgeKindContradicts means the source memory conflicts with the target memory.
	EdgeKindContradicts EdgeKind = "contradicts"
	// EdgeKindDerivedFrom means the source memory was derived from the target artifact.
	EdgeKindDerivedFrom EdgeKind = "derived_from"
	// EdgeKindReferencesDoc means the source memory is grounded in a document.
	EdgeKindReferencesDoc EdgeKind = "references_doc"
	// EdgeKindBelongsTo means the source memory belongs to an entity or scope.
	EdgeKindBelongsTo EdgeKind = "belongs_to"
	// EdgeKindCorrectedBy means the source memory was changed by an operator correction.
	EdgeKindCorrectedBy EdgeKind = "corrected_by"
)

// MemoryStatus describes whether a memory participates in recall.
type MemoryStatus string

const (
	// MemoryStatusActive is eligible for recall.
	MemoryStatusActive MemoryStatus = "active"
	// MemoryStatusSuperseded was replaced by a newer memory.
	MemoryStatusSuperseded MemoryStatus = "superseded"
	// MemoryStatusArchived is retained for provenance but suppressed by default.
	MemoryStatusArchived MemoryStatus = "archived"
	// MemoryStatusDeleted marks a memory removed by an explicit deletion policy.
	MemoryStatusDeleted MemoryStatus = "deleted"
)

// JobKind identifies a worker queue job.
type JobKind string

const (
	// JobKindProcessTurnEvent runs the event-to-memory pipeline.
	JobKindProcessTurnEvent JobKind = "process_turn_event"
	// JobKindEmbedDocumentChunks embeds document retrieval units.
	JobKindEmbedDocumentChunks JobKind = "embed_document_chunks"
	// JobKindDreamSession consolidates a session after activity.
	JobKindDreamSession JobKind = "dream_session"
	// JobKindDreamWorkspace consolidates workspace-level memory.
	JobKindDreamWorkspace JobKind = "dream_workspace"
	// JobKindRebuildProfile recomputes profile snapshots.
	JobKindRebuildProfile JobKind = "rebuild_profile"
	// JobKindMaintenance runs cleanup and backfill work.
	JobKindMaintenance JobKind = "maintenance"
)

// ArtifactClass is the broad retrieval lane for a memory.
type ArtifactClass string

const (
	// ArtifactClassContext represents current-session or short-lived work context.
	ArtifactClassContext ArtifactClass = "context"
	// ArtifactClassKnowledge represents durable facts, preferences, rules, and procedures.
	ArtifactClassKnowledge ArtifactClass = "knowledge"
	// ArtifactClassTimeline represents events, decisions, corrections, and changes.
	ArtifactClassTimeline ArtifactClass = "timeline"
	// ArtifactClassPlan represents goals, task state, and next actions.
	ArtifactClassPlan ArtifactClass = "plan"
)

```



<!-- Source: internal/core/memory.go | bytes=3839 | lines=83 | sha16=eb74a58ad1cda330 -->

```go
// ============================================================
// FILE     : internal/core/memory.go
// PURPOSE  : Defines derived memory, graph edge, and provenance trace records.
// LAYER    : domain
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : Memory, MemoryEdge, MemoryTrace, MemoryCorrection
// DEPENDS  : encoding/json, time, internal/core/kind.go, internal/core/scope.go
// USED_BY  : apply engine, recall, storage, explain-memory path
// ------------------------------------------------------------
// AGENT_NOTE: Never blur raw events, derived memories, and memory_trace.
// ============================================================

package core

import (
	"encoding/json"
	"time"
)

// Memory is a derived structured object used by recall and graph operations.
type Memory struct {
	ID                 string          `json:"id"`
	TenantID           string          `json:"tenant_id"`
	WorkspaceID        string          `json:"workspace_id"`
	Scope              MemoryScope     `json:"scope"`
	GroupID            *string         `json:"group_id,omitempty"`
	OwnerEntityID      string          `json:"owner_entity_id"`
	Kind               MemoryKind      `json:"kind"`
	ArtifactClass      ArtifactClass   `json:"artifact_class"`
	Text               string          `json:"text"`
	Fingerprint        string          `json:"fingerprint"`
	Confidence         float64         `json:"confidence"`
	Status             MemoryStatus    `json:"status"`
	ValidFrom          time.Time       `json:"valid_from"`
	ValidTo            *time.Time      `json:"valid_to,omitempty"`
	LatestFlag         bool            `json:"latest_flag"`
	MetadataJSON       json.RawMessage `json:"metadata_json"`
	EmbeddingModel     string          `json:"embedding_model"`
	EmbeddingDims      int             `json:"embedding_dims"`
	EmbeddingUpdatedAt *time.Time      `json:"embedding_updated_at,omitempty"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}

// MemoryEdge records a relationship between two memory objects.
type MemoryEdge struct {
	FromMemoryID   string    `json:"from_memory_id"`
	ToMemoryID     string    `json:"to_memory_id"`
	EdgeKind       EdgeKind  `json:"edge_kind"`
	Confidence     float64   `json:"confidence"`
	CreatedByJobID string    `json:"created_by_job_id"`
	CreatedAt      time.Time `json:"created_at"`
}

// MemoryTrace stores provenance and reasoning evidence for one memory.
type MemoryTrace struct {
	MemoryID               string          `json:"memory_id"`
	RawEventIDs            []string        `json:"raw_event_ids"`
	ReasoningJobID         string          `json:"reasoning_job_id"`
	ReasoningStage         string          `json:"reasoning_stage"`
	CandidateSnapshotJSON  json.RawMessage `json:"candidate_snapshot_json"`
	AppliedOperationsJSON  json.RawMessage `json:"applied_operations_json"`
	OperatorCorrectionFlag bool            `json:"operator_correction_flag"`
	RelatedDocumentIDs     []string        `json:"related_document_ids"`
	CreatedAt              time.Time       `json:"created_at"`
}

// MemoryCorrection records a human correction intent without superseding a memory.
type MemoryCorrection struct {
	ID             string          `json:"id"`
	TenantID       string          `json:"tenant_id"`
	WorkspaceID    string          `json:"workspace_id"`
	MemoryID       string          `json:"memory_id"`
	OperatorID     string          `json:"operator_id"`
	RawEventID     string          `json:"raw_event_id"`
	IdempotencyKey string          `json:"idempotency_key"`
	CorrectionText string          `json:"correction_text"`
	EvidenceJSON   json.RawMessage `json:"evidence_json"`
	Status         string          `json:"status"`
	CreatedAt      time.Time       `json:"created_at"`
}

```



<!-- Source: internal/core/note.go | bytes=1222 | lines=32 | sha16=9624b0400f63cd00 -->

```go
// ============================================================
// FILE     : internal/core/note.go
// PURPOSE  : Defines human-authored note records that can influence recall.
// LAYER    : domain
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : Note
// DEPENDS  : time, internal/core/scope.go
// USED_BY  : internal/store, recall assembler, note API
// ------------------------------------------------------------
// AGENT_NOTE: Notes are operator intent and must stay distinct from memories.
// ============================================================

package core

import "time"

// Note is a human-authored memory control artifact.
type Note struct {
	ID            string      `json:"id"`
	TenantID      string      `json:"tenant_id"`
	WorkspaceID   string      `json:"workspace_id"`
	NoteKind      string      `json:"note_kind"`
	Scope         MemoryScope `json:"scope"`
	OwnerEntityID string      `json:"owner_entity_id"`
	Text          string      `json:"text"`
	Pinned        bool        `json:"pinned"`
	ExpiresAt     *time.Time  `json:"expires_at,omitempty"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

```



<!-- Source: internal/core/plan.go | bytes=1652 | lines=45 | sha16=1e504a1d49b3ced5 -->

```go
// ============================================================
// FILE     : internal/core/plan.go
// PURPOSE  : Defines structured plan records and their task items.
// LAYER    : domain
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : Plan, PlanItem
// DEPENDS  : encoding/json, time, internal/core/scope.go
// USED_BY  : internal/store, recall assembler, plan API
// ------------------------------------------------------------
// AGENT_NOTE: Active plans get recall priority, so preserve scope and evidence.
// ============================================================

package core

import (
	"encoding/json"
	"time"
)

// Plan is a structured operator or agent plan.
type Plan struct {
	ID            string          `json:"id"`
	TenantID      string          `json:"tenant_id"`
	WorkspaceID   string          `json:"workspace_id"`
	Title         string          `json:"title"`
	Status        string          `json:"status"`
	Scope         MemoryScope     `json:"scope"`
	OwnerEntityID string          `json:"owner_entity_id"`
	EvidenceJSON  json.RawMessage `json:"evidence_json"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

// PlanItem is a task within a structured plan.
type PlanItem struct {
	ID           string          `json:"id"`
	PlanID       string          `json:"plan_id"`
	Title        string          `json:"title"`
	Status       string          `json:"status"`
	EvidenceJSON json.RawMessage `json:"evidence_json"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

```



<!-- Source: internal/core/profile.go | bytes=1144 | lines=31 | sha16=1119cd5e4fd3c748 -->

```go
// ============================================================
// FILE     : internal/core/profile.go
// PURPOSE  : Defines rebuildable static and dynamic profile snapshots.
// LAYER    : domain
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : Profile
// DEPENDS  : encoding/json, time, internal/core/scope.go
// USED_BY  : recall assembler, dreaming, storage
// ------------------------------------------------------------
// AGENT_NOTE: Profiles must remain rebuildable from raw events, memories, and edges.
// ============================================================

package core

import (
	"encoding/json"
	"time"
)

// Profile is a rebuildable static and dynamic snapshot for an entity.
type Profile struct {
	EntityID        string          `json:"entity_id"`
	Scope           MemoryScope     `json:"scope"`
	StaticJSON      json.RawMessage `json:"static_json"`
	DynamicJSON     json.RawMessage `json:"dynamic_json"`
	SourceMemoryIDs []string        `json:"source_memory_ids"`
	UpdatedAt       time.Time       `json:"updated_at"`
	Version         int64           `json:"version"`
}

```



<!-- Source: internal/core/raw_event.go | bytes=1381 | lines=36 | sha16=29e43a0eea44e2af -->

```go
// ============================================================
// FILE     : internal/core/raw_event.go
// PURPOSE  : Defines immutable ingest records before memory derivation.
// LAYER    : domain
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : RawEvent
// DEPENDS  : encoding/json, time
// USED_BY  : ingest hot path, worker jobs, memory_trace
// ------------------------------------------------------------
// AGENT_NOTE: Raw events are source records; do not mix derived memory fields into them.
// ============================================================

package core

import (
	"encoding/json"
	"time"
)

// RawEvent is an immutable ingest record before memory derivation.
type RawEvent struct {
	ID             string          `json:"id"`
	TenantID       string          `json:"tenant_id"`
	WorkspaceID    string          `json:"workspace_id"`
	SessionID      string          `json:"session_id"`
	ActorID        string          `json:"actor_id"`
	EventKind      string          `json:"event_kind"`
	Source         string          `json:"source"`
	IdempotencyKey string          `json:"idempotency_key"`
	Fingerprint    string          `json:"fingerprint"`
	OccurredAt     time.Time       `json:"occurred_at"`
	PayloadJSON    json.RawMessage `json:"payload_json"`
	CreatedAt      time.Time       `json:"created_at"`
}

```



<!-- Source: internal/core/scope.go | bytes=1275 | lines=29 | sha16=c565b3a83e0b77eb -->

```go
// ============================================================
// FILE     : internal/core/scope.go
// PURPOSE  : Defines explicit visibility scopes for memory artifacts.
// LAYER    : domain
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : MemoryScope
// DEPENDS  : plans/04_memory-scopes_dreaming_ontology-lite.md
// USED_BY  : every memory, note, plan, profile, and recall path
// ------------------------------------------------------------
// AGENT_NOTE: Scope must never be implicit or nullable in memory writes.
// ============================================================

package core

// MemoryScope identifies the visibility boundary for a memory artifact.
type MemoryScope string

const (
	// MemoryScopeAgentPrivate is visible to one owning agent and the operator.
	MemoryScopeAgentPrivate MemoryScope = "agent_private"
	// MemoryScopeWorkspaceShared is visible to members of the workspace.
	MemoryScopeWorkspaceShared MemoryScope = "workspace_shared"
	// MemoryScopeGroupShared is visible to members of a named memory group.
	MemoryScopeGroupShared MemoryScope = "group_shared"
	// MemoryScopeSessionScratch is short-lived session-local context.
	MemoryScopeSessionScratch MemoryScope = "session_scratch"
)

```



<!-- Source: internal/core/service.go | bytes=1769 | lines=34 | sha16=f998808e3a6d6238 -->

```go
// ============================================================
// FILE     : internal/core/service.go
// PURPOSE  : Defines the primary v1 service contract for all runtime surfaces.
// LAYER    : domain
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : VibeGravityService
// DEPENDS  : context, internal/core/dto.go
// USED_BY  : HTTP API, MCP, Hermes provider, tests
// ------------------------------------------------------------
// AGENT_NOTE: Do not change this interface without updating AGENTS.md and runtime docs.
// ============================================================

package core

import (
	"context"
)

// VibeGravityService defines the primary v1 interface for every runtime surface.
type VibeGravityService interface {
	Prefetch(ctx context.Context, req *PrefetchRequest) (*PrefetchResponse, error)
	SyncTurn(ctx context.Context, req *SyncTurnRequest) (*SyncTurnResponse, error)
	AddDocument(ctx context.Context, req *AddDocumentRequest) (*AddDocumentResponse, error)
	SearchMemories(ctx context.Context, req *SearchMemoriesRequest) (*SearchMemoriesResponse, error)
	SearchDocuments(ctx context.Context, req *SearchDocumentsRequest) (*SearchDocumentsResponse, error)
	AddNote(ctx context.Context, req *AddNoteRequest) (*AddNoteResponse, error)
	CreatePlan(ctx context.Context, req *CreatePlanRequest) (*CreatePlanResponse, error)
	UpdatePlan(ctx context.Context, req *UpdatePlanRequest) (*UpdatePlanResponse, error)
	CorrectMemory(ctx context.Context, req *CorrectMemoryRequest) (*CorrectMemoryResponse, error)
	GetTimeline(ctx context.Context, req *GetTimelineRequest) (*GetTimelineResponse, error)
	ExplainMemory(ctx context.Context, req *ExplainMemoryRequest) (*ExplainMemoryResponse, error)
}

```



<!-- Source: internal/core/service_test.go | bytes=3659 | lines=120 | sha16=acc8d58cea98b1e5 -->

```go
// ============================================================
// FILE     : internal/core/service_test.go
// PURPOSE  : Verifies the core service interface and domain records compile together.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : TestVibeGravityService_Baseline, TestDomainTypes_Compile
// DEPENDS  : internal/core
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: Keep this as a fast contract smoke test for domain changes.
// ============================================================

package core

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestVibeGravityService_Baseline(t *testing.T) {
	t.Helper()
	var _ VibeGravityService = (*contractService)(nil)
}

func TestDomainTypes_Compile(t *testing.T) {
	now := time.Date(2026, time.April, 24, 0, 0, 0, 0, time.UTC)
	payload := json.RawMessage(`{"text":"hello"}`)

	memory := Memory{
		ID:             "mem_1",
		TenantID:       "tenant_1",
		WorkspaceID:    "workspace_1",
		Scope:          MemoryScopeWorkspaceShared,
		OwnerEntityID:  "agent:hermes-main",
		Kind:           MemoryKindDecision,
		ArtifactClass:  ArtifactClassKnowledge,
		Text:           "VibeGravity is Hermes-first.",
		Fingerprint:    "fp_1",
		Confidence:     0.99,
		Status:         MemoryStatusActive,
		ValidFrom:      now,
		LatestFlag:     true,
		MetadataJSON:   payload,
		EmbeddingModel: "pending",
		EmbeddingDims:  0,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if memory.Scope != MemoryScopeWorkspaceShared {
		t.Fatalf("unexpected scope: %s", memory.Scope)
	}
	if memory.ArtifactClass != ArtifactClassKnowledge {
		t.Fatalf("unexpected artifact class: %s", memory.ArtifactClass)
	}

	trace := MemoryTrace{
		MemoryID:               memory.ID,
		RawEventIDs:            []string{"evt_1"},
		ReasoningJobID:         "job_1",
		ReasoningStage:         "resolve",
		CandidateSnapshotJSON:  payload,
		AppliedOperationsJSON:  payload,
		OperatorCorrectionFlag: false,
		RelatedDocumentIDs:     []string{"doc_1"},
		CreatedAt:              now,
	}
	if trace.RawEventIDs[0] != "evt_1" {
		t.Fatalf("unexpected trace event id: %s", trace.RawEventIDs[0])
	}
}

type contractService struct{}

func (contractService) Prefetch(context.Context, *PrefetchRequest) (*PrefetchResponse, error) {
	return nil, nil
}

func (contractService) SyncTurn(context.Context, *SyncTurnRequest) (*SyncTurnResponse, error) {
	return nil, nil
}

func (contractService) AddDocument(context.Context, *AddDocumentRequest) (*AddDocumentResponse, error) {
	return nil, nil
}

func (contractService) SearchMemories(context.Context, *SearchMemoriesRequest) (*SearchMemoriesResponse, error) {
	return nil, nil
}

func (contractService) SearchDocuments(context.Context, *SearchDocumentsRequest) (*SearchDocumentsResponse, error) {
	return nil, nil
}

func (contractService) AddNote(context.Context, *AddNoteRequest) (*AddNoteResponse, error) {
	return nil, nil
}

func (contractService) CreatePlan(context.Context, *CreatePlanRequest) (*CreatePlanResponse, error) {
	return nil, nil
}

func (contractService) UpdatePlan(context.Context, *UpdatePlanRequest) (*UpdatePlanResponse, error) {
	return nil, nil
}

func (contractService) CorrectMemory(context.Context, *CorrectMemoryRequest) (*CorrectMemoryResponse, error) {
	return nil, nil
}

func (contractService) GetTimeline(context.Context, *GetTimelineRequest) (*GetTimelineResponse, error) {
	return nil, nil
}

func (contractService) ExplainMemory(context.Context, *ExplainMemoryRequest) (*ExplainMemoryResponse, error) {
	return nil, nil
}

```



<!-- Source: internal/core/session.go | bytes=1157 | lines=30 | sha16=b66cd6dc0865d6cb -->

```go
// ============================================================
// FILE     : internal/core/session.go
// PURPOSE  : Defines rebuildable summaries for session-level memory consolidation.
// LAYER    : domain
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : SessionSummary
// DEPENDS  : time
// USED_BY  : dreaming jobs, recall assembler, storage
// ------------------------------------------------------------
// AGENT_NOTE: Session summaries are derived artifacts and must keep source IDs.
// ============================================================

package core

import "time"

// SessionSummary is a rebuildable summary for one session.
type SessionSummary struct {
	ID              string    `json:"id"`
	TenantID        string    `json:"tenant_id"`
	WorkspaceID     string    `json:"workspace_id"`
	SessionID       string    `json:"session_id"`
	SummaryText     string    `json:"summary_text"`
	SourceEventIDs  []string  `json:"source_event_ids"`
	SourceMemoryIDs []string  `json:"source_memory_ids"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

```



<!-- Source: internal/db/pool.go | bytes=1778 | lines=54 | sha16=6eb5f49d25779769 -->

```go
// ============================================================
// FILE     : internal/db/pool.go
// PURPOSE  : Builds PostgreSQL connection pools with pgvector registration.
// LAYER    : infra
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : NewPool
// DEPENDS  : internal/config, github.com/jackc/pgx/v5, github.com/pgvector/pgvector-go
// USED_BY  : cmd/server, cmd/worker, cmd/cli, tests
// ------------------------------------------------------------
// AGENT_NOTE: PostgreSQL is canonical; keep pgvector setup explicit here.
// ============================================================

// Package db manages the database connection pool using pgxpool.
package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pgxvector "github.com/pgvector/pgvector-go/pgx"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/config"
)

// NewPool initializes a new PostgreSQL connection pool.
// It configures the pgxpool to register pgvector types automatically on new connections.
func NewPool(ctx context.Context, cfg config.Config) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Register pgvector types on each connection
	poolConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		return pgxvector.RegisterTypes(ctx, conn)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}

```



<!-- Source: internal/embed/doc.go | bytes=765 | lines=16 | sha16=8f44d8fa17f9fda3 -->

```go
// ============================================================
// FILE     : internal/embed/doc.go
// PURPOSE  : Provides package documentation for local embedding and retrieval helpers.
// LAYER    : infra
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : package embed
// DEPENDS  : plans/03_target-architecture_codex-first.md
// USED_BY  : worker pipeline, recall, reasoning neighborhood retrieval
// ------------------------------------------------------------
// AGENT_NOTE: Keep local runtime embedding-focused; do not add a local extractor here.
// ============================================================

// Package embed owns local embedding clients and lexical/vector retrieval helpers.
package embed

```
