// ============================================================
// FILE     : cmd/cli/main.go
// PURPOSE  : Starts the CLI and runs local operator checks such as doctor and job metrics/recovery.
// LAYER    : interface
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : main
// DEPENDS  : context, errors, fmt, io, net/http, net/url, os, strconv, strings, time, internal/config, internal/core, internal/db, internal/eval, internal/mcp, internal/store/postgres
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
	"net/url"
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
		Jobs:        pgStore,
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

func maskPassword(raw string) string {
	if masked, ok := maskKeywordPassword(raw); ok {
		return masked
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	changed := false
	if parsed.User != nil {
		if _, ok := parsed.User.Password(); ok {
			parsed.User = url.UserPassword(parsed.User.Username(), "xxxxx")
			changed = true
		}
	}
	query := parsed.Query()
	if _, ok := query["password"]; ok {
		query.Set("password", "xxxxx")
		parsed.RawQuery = query.Encode()
		changed = true
	}
	if !changed {
		return raw
	}
	return parsed.String()
}

func maskKeywordPassword(raw string) (string, bool) {
	fields := strings.Fields(raw)
	if len(fields) == 0 {
		return raw, false
	}
	changed := false
	for i, field := range fields {
		if strings.HasPrefix(strings.ToLower(field), "password=") {
			fields[i] = "password=xxxxx"
			changed = true
		}
	}
	if !changed {
		return raw, false
	}
	return strings.Join(fields, " "), true
}
