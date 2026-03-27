package tests

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// runClaude invokes the Claude Code CLI in print mode against the given directory.
func runClaude(t *testing.T, dir string, prompt string) string {
	t.Helper()
	cmd := exec.Command("claude", "-p", "--allowedTools", "*", prompt)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("claude command failed: %v\nOutput:\n%s", err, string(out))
	}
	return string(out)
}

// assertFileExists fails the test if the file does not exist relative to dir.
func assertFileExists(t *testing.T, dir string, relPath string) {
	t.Helper()
	full := filepath.Join(dir, relPath)
	if _, err := os.Stat(full); os.IsNotExist(err) {
		t.Errorf("expected file %s to exist, but it does not", relPath)
	}
}

// assertDirExists fails the test if the directory does not exist relative to dir.
func assertDirExists(t *testing.T, dir string, relPath string) {
	t.Helper()
	full := filepath.Join(dir, relPath)
	info, err := os.Stat(full)
	if os.IsNotExist(err) {
		t.Errorf("expected directory %s to exist, but it does not", relPath)
	} else if err == nil && !info.IsDir() {
		t.Errorf("expected %s to be a directory, but it is a file", relPath)
	}
}

// assertFileContains fails if the file doesn't contain the substring.
func assertFileContains(t *testing.T, dir string, relPath string, substr string) {
	t.Helper()
	full := filepath.Join(dir, relPath)
	data, err := os.ReadFile(full)
	if err != nil {
		t.Fatalf("failed to read %s: %v", relPath, err)
	}
	if !strings.Contains(string(data), substr) {
		t.Errorf("file %s does not contain %q", relPath, substr)
	}
}

// runCmd runs a shell command in dir and fails the test on non-zero exit.
func runCmd(t *testing.T, dir string, name string, args ...string) string {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command %s %v failed: %v\nOutput:\n%s", name, args, err, string(out))
	}
	return string(out)
}

// runMake runs a make target in dir and fails the test on non-zero exit.
func runMake(t *testing.T, dir string, target string) string {
	t.Helper()
	return runCmd(t, dir, "make", target)
}

// httpGet performs an HTTP GET and returns the status code.
func httpGet(t *testing.T, url string) int {
	t.Helper()
	cmd := exec.Command("curl", "-s", "-o", "/dev/null", "-w", "%{http_code}", url)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("curl %s failed: %v\nOutput:\n%s", url, err, string(out))
	}
	var code int
	fmt.Sscanf(strings.TrimSpace(string(out)), "%d", &code)
	return code
}

// createTempGitHubRepo creates a private repo and returns its name.
func createTempGitHubRepo(t *testing.T) string {
	t.Helper()
	name := fmt.Sprintf("forge-test-%d", os.Getpid())
	runCmd(t, ".", "gh", "repo", "create", name, "--private", "--confirm")
	t.Cleanup(func() {
		exec.Command("gh", "repo", "delete", name, "--yes").Run()
	})
	return name
}

// waitForWorkflowGreen polls gh run list until the latest run completes or times out.
func waitForWorkflowGreen(t *testing.T, repo string, timeoutMinutes int) bool {
	t.Helper()
	deadline := timeoutMinutes * 60
	for elapsed := 0; elapsed < deadline; elapsed += 30 {
		out := runCmd(t, ".", "gh", "run", "list", "--repo", repo, "--limit", "1", "--json", "status,conclusion")
		if strings.Contains(out, `"status":"completed"`) {
			return strings.Contains(out, `"conclusion":"success"`)
		}
		cmd := exec.Command("sleep", "30")
		cmd.Run()
	}
	t.Fatalf("workflow did not complete within %d minutes", timeoutMinutes)
	return false
}

// copyDir recursively copies src to dst.
func copyDir(t *testing.T, src, dst string) {
	t.Helper()
	runCmd(t, ".", "cp", "-r", src, dst)
}
