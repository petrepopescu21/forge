package tests

import (
	"fmt"
	"os"
	"os/exec"
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

	dir := t.TempDir()

	// Use bootstrap-project to get a complete project with CI in one session
	runClaude(t, dir, "use forge:bootstrap-project with name testapp, module github.com/test/testapp, description 'a test project', all layers yes")

	assertFileExists(t, dir, ".github/workflows/ci.yml")
	assertFileContains(t, dir, ".github/workflows/ci.yml", "lint")
	assertFileContains(t, dir, ".github/workflows/ci.yml", "test")

	repoName := fmt.Sprintf("forge-test-%d", os.Getpid())
	ghToken := os.Getenv("FORGE_TEST_GITHUB_TOKEN")

	runCmd(t, dir, "gh", "repo", "create", repoName, "--private", "--confirm")
	t.Cleanup(func() {
		exec.Command("gh", "repo", "delete", repoName, "--yes").Run()
	})

	runCmd(t, dir, "git", "init")
	runCmd(t, dir, "git", "config", "user.name", "forge-test")
	runCmd(t, dir, "git", "config", "user.email", "forge-test@test.local")
	runCmd(t, dir, "git", "add", ".")
	runCmd(t, dir, "git", "commit", "-m", "initial: test scaffold")

	owner := strings.TrimSpace(runCmd(t, dir, "gh", "api", "user", "--jq", ".login"))
	remote := fmt.Sprintf("https://x-access-token:%s@github.com/%s/%s.git",
		ghToken, owner, repoName)
	runCmd(t, dir, "git", "remote", "add", "origin", remote)
	runCmd(t, dir, "git", "push", "-u", "origin", "main")

	fullRepo := fmt.Sprintf("%s/%s", owner, repoName)
	success := waitForWorkflowGreen(t, fullRepo, 15)
	if !success {
		out := runCmd(t, ".", "gh", "run", "list", "--repo", fullRepo, "--limit", "1",
			"--json", "databaseId,conclusion")
		t.Fatalf("CI workflow did not succeed. Run info: %s", out)
	}
}
