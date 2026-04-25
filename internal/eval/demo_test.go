// ============================================================
// FILE     : internal/eval/demo_test.go
// PURPOSE  : Verifies the local Hermes Memory trust-loop demo eval.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : Hermes Memory demo eval tests
// DEPENDS  : context, testing
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: Demo eval tests must stay local-only and deterministic.
// ============================================================

package eval

import (
	"context"
	"testing"
)

func TestRunHermesMemoryDemoPassesTrustLoop(t *testing.T) {
	t.Parallel()

	summary := RunHermesMemoryDemo(context.Background())
	if summary == nil || !summary.Passed {
		t.Fatalf("expected demo eval to pass, got %#v", summary)
	}

	names := make(map[string]bool, len(summary.Results))
	for _, result := range summary.Results {
		names[result.Scenario] = true
	}
	for _, want := range []string{
		"demo initial recall shows rule plan and trust metadata",
		"demo explain shows recalled memory provenance",
		"demo correction writes supersession",
		"demo next recall uses correction",
		"demo private scope separation",
	} {
		if !names[want] {
			t.Fatalf("expected demo step %q in %#v", want, summary.Results)
		}
	}
}
