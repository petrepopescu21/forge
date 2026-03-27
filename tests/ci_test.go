package tests

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func skipIfNoGitHubToken(t *testing.T) {
	t.Helper()
	if os.Getenv("FORGE_TEST_GITHUB_TOKEN") == "" {
		t.Skip("FORGE_TEST_GITHUB_TOKEN not set, skipping CI validation test")
	}
}

func TestSetupCiValidation(t *testing.T) {
	skipIfNoAPI(t)
	skipIfNoGitHubToken(t)
	if testing.Short() {
		t.Skip("skipping CI validation test in short mode")
	}
	if fixtureDir == "" {
		t.Skip("bootstrap fixture not available")
	}

	dir := t.TempDir()
	copyDir(t, fixtureDir, filepath.Join(dir, "project"))
	projectDir := filepath.Join(dir, "project")

	assertFileExists(t, projectDir, ".github/workflows/ci.yml")
	assertFileContains(t, projectDir, ".github/workflows/ci.yml", "lint")
	assertFileContains(t, projectDir, ".github/workflows/ci.yml", "test")

	repoName := fmt.Sprintf("forge-test-%d", os.Getpid())
	ghToken := os.Getenv("FORGE_TEST_GITHUB_TOKEN")

	runCmd(t, projectDir, "gh", "repo", "create", repoName, "--private", "--confirm")
	t.Cleanup(func() {
		exec.Command("gh", "repo", "delete", repoName, "--yes").Run()
	})

	runCmd(t, projectDir, "git", "init")
	runCmd(t, projectDir, "git", "config", "user.name", "forge-test")
	runCmd(t, projectDir, "git", "config", "user.email", "forge-test@test.local")
	runCmd(t, projectDir, "git", "add", ".")
	runCmd(t, projectDir, "git", "commit", "-m", "initial: test scaffold")

	owner := strings.TrimSpace(runCmd(t, projectDir, "gh", "api", "user", "--jq", ".login"))
	remote := fmt.Sprintf("https://x-access-token:%s@github.com/%s/%s.git",
		ghToken, owner, repoName)
	runCmd(t, projectDir, "git", "remote", "add", "origin", remote)
	runCmd(t, projectDir, "git", "push", "-u", "origin", "main")

	fullRepo := fmt.Sprintf("%s/%s", owner, repoName)
	success := waitForWorkflowGreen(t, fullRepo, 15)
	if !success {
		out := runCmd(t, ".", "gh", "run", "list", "--repo", fullRepo, "--limit", "1",
			"--json", "databaseId,conclusion")
		t.Fatalf("CI workflow did not succeed. Run info: %s", out)
	}
}
