// ============================================================
// FILE     : internal/eval/golden_test.go
// PURPOSE  : Verifies the deterministic golden scenario runner.
// LAYER    : test
// STATUS   : draft
// ------------------------------------------------------------
// EXPORTS  : TestRunFilePassesGoldenFixture, TestRunScenariosReportsRegression
// DEPENDS  : context, path/filepath, testing, internal/eval
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: Eval tests should fail loudly when fixture expectations drift.
// ============================================================

package eval

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/parker-jungwoo-hwang/vibegravity/internal/core"
)

func TestRunFilePassesGoldenFixture(t *testing.T) {
	t.Parallel()

	summary, err := RunFile(context.Background(), filepath.Join("..", "..", "tests", "golden", "replay_eval.json"))
	if err != nil {
		t.Fatalf("RunFile returned error: %v", err)
	}
	if !summary.Passed {
		t.Fatalf("expected golden fixture to pass, got %#v", summary.Results)
	}
	if len(summary.Results) < 4 {
		t.Fatalf("expected multiple golden scenarios, got %d", len(summary.Results))
	}
}

func TestRunFileCoversGraphReplayGates(t *testing.T) {
	t.Parallel()

	summary, err := RunFile(context.Background(), filepath.Join("..", "..", "tests", "golden", "replay_eval.json"))
	if err != nil {
		t.Fatalf("RunFile returned error: %v", err)
	}
	names := make(map[string]bool, len(summary.Results))
	for _, result := range summary.Results {
		names[result.Scenario] = true
	}
	for _, want := range []string{
		"update memory replay suppresses prior fact",
		"correction replay changes later recall",
		"group shared graph write remains rejected",
	} {
		if !names[want] {
			t.Fatalf("expected graph replay scenario %q", want)
		}
	}
}

func TestRunFileCoversWorkerBacklogGates(t *testing.T) {
	t.Parallel()

	summary, err := RunFile(context.Background(), filepath.Join("..", "..", "tests", "golden", "replay_eval.json"))
	if err != nil {
		t.Fatalf("RunFile returned error: %v", err)
	}
	names := make(map[string]bool, len(summary.Results))
	for _, result := range summary.Results {
		names[result.Scenario] = true
	}
	for _, want := range []string{
		"stage1 outage retries without graph writes",
		"stage2 outage recovery replay is idempotent",
		"unsupported apply work becomes blocked",
	} {
		if !names[want] {
			t.Fatalf("expected worker backlog scenario %q", want)
		}
	}
}

func TestRunScenariosReportsRegression(t *testing.T) {
	t.Parallel()

	summary := RunScenarios(context.Background(), []Scenario{{
		Name: "missing pinned note is visible",
		Prefetch: core.PrefetchRequest{
			TenantID:     "tenant_1",
			WorkspaceID:  "workspace_1",
			SessionID:    "session_1",
			ActorID:      "agent:hermes-main",
			Query:        "plan",
			BudgetTokens: 100,
		},
		Expect: Expectation{Contains: []string{"must appear"}},
	}})

	if summary.Passed {
		t.Fatalf("expected scenario to fail")
	}
	if len(summary.Results) != 1 || len(summary.Results[0].Errors) == 0 {
		t.Fatalf("expected failure details, got %#v", summary.Results)
	}
}

func TestRunScenariosReportsBlockMetadataRegression(t *testing.T) {
	t.Parallel()

	summary := RunScenarios(context.Background(), []Scenario{{
		Name: "trust metadata mismatch is visible",
		Prefetch: core.PrefetchRequest{
			TenantID:     "tenant_1",
			WorkspaceID:  "workspace_1",
			SessionID:    "session_1",
			ActorID:      "agent:hermes-main",
			Query:        "Hermes",
			BudgetTokens: 100,
		},
		Fixtures: Fixtures{
			Memories: []core.MemoryResult{{
				MemoryID:      "mem_1",
				Kind:          core.MemoryKindFact,
				ArtifactClass: core.ArtifactClassKnowledge,
				Text:          "Hermes remembers scoped project context.",
				Confidence:    0.9,
				Scope:         core.MemoryScopeWorkspaceShared,
				LatestFlag:    true,
			}},
		},
		Expect: Expectation{
			BlockMetadata: []BlockMetadataExpectation{{
				Kind:      "memory",
				Scope:     core.MemoryScopeAgentPrivate,
				Source:    "memories",
				SourceID:  "mem_1",
				Status:    "active",
				Freshness: "stored",
			}},
		},
	}})

	if summary.Passed {
		t.Fatalf("expected scenario to fail")
	}
	if got := summary.Results[0].Errors; len(got) == 0 || !strings.Contains(strings.Join(got, "\n"), "scope got") {
		t.Fatalf("expected metadata failure details, got %#v", summary.Results)
	}
}
