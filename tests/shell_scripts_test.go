// ============================================================
// FILE     : tests/shell_scripts_test.go
// PURPOSE  : Guards repo-local agent shell scripts with syntax, shellcheck, and coordination workflow smoke tests.
// LAYER    : test
// STATUS   : active
// ------------------------------------------------------------
// EXPORTS  : shell script tests
// DEPENDS  : os, os/exec, path/filepath, strings, testing
// USED_BY  : go test ./..., make release-gate
// ------------------------------------------------------------
// AGENT_NOTE: Keep fixture tests isolated from the real .agents coordination state.
// ============================================================

package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestAgentShellScriptsHaveValidSyntax(t *testing.T) {
	t.Parallel()

	root := repoRoot(t)
	for _, script := range agentShellScripts(t, root) {
		script := script
		t.Run(filepath.ToSlash(script), func(t *testing.T) {
			t.Parallel()

			shell := syntaxShellForScript(t, script)
			cmd := exec.Command(shell, "-n", script)
			cmd.Dir = root
			if output, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("%s -n %s failed: %v\n%s", shell, script, err, output)
			}
		})
	}
}

func TestAgentShellScriptsPassShellcheckWhenAvailable(t *testing.T) {
	t.Parallel()

	shellcheck, err := exec.LookPath("shellcheck")
	if err != nil {
		t.Skip("shellcheck is not installed")
	}

	root := repoRoot(t)
	args := append([]string{"-x"}, agentShellScripts(t, root)...)
	cmd := exec.Command(shellcheck, args...)
	cmd.Dir = root
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("shellcheck failed: %v\n%s", err, output)
	}
}

func TestAgentWorkClaimReleaseDoneFixture(t *testing.T) {
	t.Parallel()

	root := repoRoot(t)
	fixtureRoot := t.TempDir()
	script := filepath.Join(fixtureRoot, ".agents", "coordination", "agent-work.sh")
	if err := os.MkdirAll(filepath.Dir(script), 0o755); err != nil {
		t.Fatalf("create fixture coordination dir: %v", err)
	}
	copyFile(t, filepath.Join(root, ".agents", "coordination", "agent-work.sh"), script, 0o755)

	runFixtureAgentWork(t, fixtureRoot, "init")
	runFixtureAgentWork(t, fixtureRoot, "claim", "agent-test", "fixture task", "internal/core/service.go", "README.md")
	claims := readFixtureFile(t, fixtureRoot, ".agents", "coordination", "claims.tsv")
	if !strings.Contains(claims, "internal/core/service.go\tagent-test\tfixture task") ||
		!strings.Contains(claims, "README.md\tagent-test\tfixture task") {
		t.Fatalf("expected claim rows, got:\n%s", claims)
	}

	runFixtureAgentWork(t, fixtureRoot, "release", "agent-test", "README.md")
	claims = readFixtureFile(t, fixtureRoot, ".agents", "coordination", "claims.tsv")
	if strings.Contains(claims, "README.md") || !strings.Contains(claims, "internal/core/service.go") {
		t.Fatalf("release should remove only the released file, got:\n%s", claims)
	}

	runFixtureAgentWork(t, fixtureRoot, "done", "agent-test", "fixture complete")
	claims = readFixtureFile(t, fixtureRoot, ".agents", "coordination", "claims.tsv")
	if strings.TrimSpace(claims) != "" {
		t.Fatalf("done should release remaining claims, got:\n%s", claims)
	}
	progress := readFixtureFile(t, fixtureRoot, ".agents", "coordination", "WORK_PROGRESS.md")
	if !strings.Contains(progress, "No active claims.") {
		t.Fatalf("expected rendered progress with no active claims, got:\n%s", progress)
	}
	log := readFixtureFile(t, fixtureRoot, ".agents", "coordination", "activity.log")
	for _, want := range []string{"\tclaim\tagent-test\t", "\trelease\tagent-test\t", "\tdone\tagent-test\t"} {
		if !strings.Contains(log, want) {
			t.Fatalf("expected activity log to contain %q, got:\n%s", want, log)
		}
	}
}

func agentShellScripts(t *testing.T, root string) []string {
	t.Helper()

	var scripts []string
	err := filepath.WalkDir(filepath.Join(root, ".agents"), func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".sh" {
			scripts = append(scripts, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk .agents scripts: %v", err)
	}
	if len(scripts) == 0 {
		t.Fatalf("expected .agents shell scripts")
	}
	return scripts
}

func repoRoot(t *testing.T) string {
	t.Helper()

	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	return root
}

func syntaxShellForScript(t *testing.T, script string) string {
	t.Helper()

	data, err := os.ReadFile(script)
	if err != nil {
		t.Fatalf("read script %s: %v", script, err)
	}
	firstLine := strings.SplitN(string(data), "\n", 2)[0]
	if strings.Contains(firstLine, "bash") {
		return "bash"
	}
	return "sh"
}

func runFixtureAgentWork(t *testing.T, fixtureRoot string, args ...string) {
	t.Helper()

	cmdArgs := append([]string{filepath.Join(fixtureRoot, ".agents", "coordination", "agent-work.sh")}, args...)
	cmd := exec.Command("sh", cmdArgs...)
	cmd.Dir = fixtureRoot
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("agent-work.sh %s failed: %v\n%s", strings.Join(args, " "), err, output)
	}
}

func copyFile(t *testing.T, src string, dst string, mode os.FileMode) {
	t.Helper()

	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("read %s: %v", src, err)
	}
	if err := os.WriteFile(dst, data, mode); err != nil {
		t.Fatalf("write %s: %v", dst, err)
	}
}

func readFixtureFile(t *testing.T, root string, parts ...string) string {
	t.Helper()

	path := filepath.Join(append([]string{root}, parts...)...)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture file %s: %v", path, err)
	}
	return string(data)
}
