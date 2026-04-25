// ============================================================
// FILE     : cmd/cli/hermes_bootstrap_stopline_test.go
// PURPOSE  : Guards the narrow Hermes bootstrap contract before broad packaging work.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : Hermes bootstrap stop-line tests
// DEPENDS  : bytes, context, strings, testing
// USED_BY  : go test ./...
// ------------------------------------------------------------
// AGENT_NOTE: Bootstrap output should stay MCP-registration-only until the trust loop is correct.
// ============================================================

package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestRunCLIHermesBootstrapStaysNarrowMCPRegistrationOnly(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	code := runCLI(context.Background(), []string{"hermes", "bootstrap", "--name", "vibegravity", "--command", "/opt/vibegravity/bin/cli"}, strings.NewReader(""), &out, fakeStoreFactory(&fakeBlockedJobStore{}), fakeServiceFactory(&fakeCLIService{}))
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; output: %s", code, out.String())
	}

	output := out.String()
	for _, want := range []string{
		"Hermes MCP bootstrap command:",
		"hermes mcp add vibegravity --command /opt/vibegravity/bin/cli --args mcp serve --stdio",
		"hermes mcp test vibegravity",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected narrow bootstrap output to contain %q, got: %s", want, output)
		}
	}
	for _, forbidden := range []string{
		"memory provider",
		"plugins/memory",
		"package",
		"install",
		"write config",
	} {
		if strings.Contains(strings.ToLower(output), forbidden) {
			t.Fatalf("bootstrap must not claim broad Hermes packaging readiness via %q, got: %s", forbidden, output)
		}
	}
}
