package tests

import (
	"os"
	"path/filepath"
	"testing"
)

func skipIfNoAPI(t *testing.T) {
	t.Helper()
	if os.Getenv("ANTHROPIC_API_KEY") == "" && os.Getenv("ANTHROPIC_AUTH_TOKEN") == "" {
		t.Skip("neither ANTHROPIC_API_KEY nor ANTHROPIC_AUTH_TOKEN set, skipping behavioral test")
	}
}

// --- Standalone setup skill tests (no prerequisites) ---

func TestSetupGoModule(t *testing.T) {
	skipIfNoAPI(t)
	dir := t.TempDir()
	runClaude(t, dir, "use forge:setup-go-module with project name testapp and module github.com/test/testapp")
	assertFileExists(t, dir, "go.mod")
	assertFileContains(t, dir, "go.mod", "module github.com/test/testapp")
	assertFileExists(t, dir, "cmd/testapp/main.go")
	assertDirExists(t, dir, "internal")
	runCmd(t, dir, "go", "build", "./...")
}

func TestSetupReact(t *testing.T) {
	skipIfNoAPI(t)
	dir := t.TempDir()
	runClaude(t, dir, "use forge:setup-react with project name testapp")
	assertFileExists(t, dir, "web/package.json")
	assertDirExists(t, dir, "web/src")
	assertFileExists(t, dir, "web/vite.config.ts")
}

func TestSetupMakefile(t *testing.T) {
	skipIfNoAPI(t)
	dir := t.TempDir()
	runClaude(t, dir, "use forge:setup-makefile with project name testapp")
	assertFileExists(t, dir, "Makefile")
	assertFileContains(t, dir, "Makefile", "lint")
	assertFileContains(t, dir, "Makefile", "test")
	assertFileContains(t, dir, "Makefile", "typecheck")
	assertFileContains(t, dir, "Makefile", "build")
}

func TestSetupSonar(t *testing.T) {
	skipIfNoAPI(t)
	dir := t.TempDir()
	runClaude(t, dir, "use forge:setup-sonar with project name testapp and project key test_testapp")
	assertFileExists(t, dir, "sonar-project.properties")
	assertFileContains(t, dir, "sonar-project.properties", "test_testapp")
}

func TestGenerateClaudeMd(t *testing.T) {
	skipIfNoAPI(t)
	dir := t.TempDir()
	runClaude(t, dir, "use forge:generate-claude-md with project name testapp, module github.com/test/testapp, description 'a test project'")
	assertFileExists(t, dir, "CLAUDE.md")
	assertFileContains(t, dir, "CLAUDE.md", "forge")
	assertFileContains(t, dir, "CLAUDE.md", "add-feature")
}

func TestSetupCi(t *testing.T) {
	skipIfNoAPI(t)
	dir := t.TempDir()
	runClaude(t, dir, "use forge:setup-ci with project name testapp")
	assertFileExists(t, dir, ".github/workflows/ci.yml")
	assertFileContains(t, dir, ".github/workflows/ci.yml", "lint")
	assertFileContains(t, dir, ".github/workflows/ci.yml", "test")
}

// --- Tests that need a bootstrapped project (use shared fixture) ---

func TestSetupLinting(t *testing.T) {
	skipIfNoAPI(t)
	if fixtureDir == "" {
		t.Skip("bootstrap fixture not available")
	}
	dir := t.TempDir()
	copyDir(t, fixtureDir, filepath.Join(dir, "project"))
	projectDir := filepath.Join(dir, "project")
	assertFileExists(t, projectDir, ".golangci.yml")
	runMake(t, projectDir, "lint")
}

func TestSetupBdd(t *testing.T) {
	skipIfNoAPI(t)
	if fixtureDir == "" {
		t.Skip("bootstrap fixture not available")
	}
	dir := t.TempDir()
	copyDir(t, fixtureDir, filepath.Join(dir, "project"))
	projectDir := filepath.Join(dir, "project")
	assertDirExists(t, projectDir, "features")
	runMake(t, projectDir, "test")
}

func TestSetupPlaywright(t *testing.T) {
	skipIfNoAPI(t)
	if fixtureDir == "" {
		t.Skip("bootstrap fixture not available")
	}
	dir := t.TempDir()
	copyDir(t, fixtureDir, filepath.Join(dir, "project"))
	projectDir := filepath.Join(dir, "project")
	assertFileExists(t, projectDir, "playwright.config.ts")
	assertDirExists(t, projectDir, "web/e2e")
}
