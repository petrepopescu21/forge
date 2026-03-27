package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBootstrapProject(t *testing.T) {
	skipIfNoAPI(t)
	dir := t.TempDir()

	runClaudeWithModel(t, dir, "use forge:bootstrap-project with name testapp, module github.com/test/testapp, description 'a test project', all layers yes", "claude-sonnet-4-6")

	assertFileExists(t, dir, "go.mod")
	assertFileContains(t, dir, "go.mod", "module github.com/test/testapp")
	assertFileExists(t, dir, "cmd/testapp/main.go")
	assertDirExists(t, dir, "internal")
	assertFileExists(t, dir, "web/package.json")
	assertFileExists(t, dir, "Makefile")
	assertFileExists(t, dir, "CLAUDE.md")
	assertFileExists(t, dir, ".golangci.yml")
	assertFileExists(t, dir, ".github/workflows/ci.yml")
	assertFileExists(t, dir, "deploy/helm/testapp/Chart.yaml")
	assertFileExists(t, dir, "sonar-project.properties")
	assertDirExists(t, dir, "features")

	runMake(t, dir, "lint")
	runMake(t, dir, "typecheck")
	runMake(t, dir, "test")
}

func TestAddFeature(t *testing.T) {
	skipIfNoAPI(t)
	if fixtureDir == "" {
		t.Skip("bootstrap fixture not available")
	}

	dir := t.TempDir()
	copyDir(t, fixtureDir, filepath.Join(dir, "project"))
	projectDir := filepath.Join(dir, "project")

	runClaude(t, projectDir, "use forge:add-feature: add a health check endpoint at GET /healthz that returns 200 OK with JSON body {\"status\": \"ok\"}")

	found := false
	filepath.Walk(filepath.Join(projectDir, "features"), func(path string, info os.FileInfo, err error) error {
		if err == nil && filepath.Ext(path) == ".feature" {
			data, _ := os.ReadFile(path)
			if strings.Contains(string(data), "health") || strings.Contains(string(data), "healthz") {
				found = true
			}
		}
		return nil
	})
	if !found {
		t.Error("expected a .feature file mentioning health check, found none")
	}

	runMake(t, projectDir, "test")
}

func TestBddFeature(t *testing.T) {
	skipIfNoAPI(t)
	if fixtureDir == "" {
		t.Skip("bootstrap fixture not available")
	}

	dir := t.TempDir()
	copyDir(t, fixtureDir, filepath.Join(dir, "project"))
	projectDir := filepath.Join(dir, "project")

	runClaude(t, projectDir, "use forge:bdd-feature with prompt: 'user can list all items with pagination'")

	found := false
	filepath.Walk(filepath.Join(projectDir, "features"), func(path string, info os.FileInfo, err error) error {
		if err == nil && filepath.Ext(path) == ".feature" {
			data, _ := os.ReadFile(path)
			content := string(data)
			if strings.Contains(content, "Scenario") && strings.Contains(content, "Given") {
				found = true
			}
		}
		return nil
	})
	if !found {
		t.Error("expected a .feature file with Gherkin scenarios, found none")
	}
}

func TestTddCycle(t *testing.T) {
	skipIfNoAPI(t)
	if fixtureDir == "" {
		t.Skip("bootstrap fixture not available")
	}

	dir := t.TempDir()
	copyDir(t, fixtureDir, filepath.Join(dir, "project"))
	projectDir := filepath.Join(dir, "project")

	runClaude(t, projectDir, "use forge:bdd-feature with prompt: 'user can get server version at GET /version'")
	runClaude(t, projectDir, "use forge:tdd-cycle to implement the version endpoint step definitions")

	runMake(t, projectDir, "test")

	out := runCmd(t, projectDir, "git", "log", "--oneline", "-10")
	if !strings.Contains(out, "red") && !strings.Contains(out, "green") {
		t.Log("Warning: git log does not show clear red/green/refactor commits")
		t.Log("Git log:\n" + out)
	}
}

func TestQualityCheck(t *testing.T) {
	skipIfNoAPI(t)
	if fixtureDir == "" {
		t.Skip("bootstrap fixture not available")
	}

	dir := t.TempDir()
	copyDir(t, fixtureDir, filepath.Join(dir, "project"))
	projectDir := filepath.Join(dir, "project")

	runClaude(t, projectDir, "use forge:quality-check")

	runMake(t, projectDir, "lint")
	runMake(t, projectDir, "typecheck")
	runMake(t, projectDir, "test")
}
