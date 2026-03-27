package tests

import (
	"os"
	"testing"
)

func skipIfNoAPI(t *testing.T) {
	t.Helper()
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping behavioral test")
	}
}

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
	runCmd(t, dir, "bun", "install")
	runCmd(t, dir, "bun", "run", "build")
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
	assertFileContains(t, dir, "Makefile", "dev-api")
	assertFileContains(t, dir, "Makefile", "dev-web")
	runMake(t, dir, "help")
}

func TestSetupLinting(t *testing.T) {
	skipIfNoAPI(t)
	dir := t.TempDir()
	runClaude(t, dir, "use forge:setup-go-module with project name testapp and module github.com/test/testapp")
	runClaude(t, dir, "use forge:setup-react with project name testapp")
	runClaude(t, dir, "use forge:setup-makefile with project name testapp")
	runClaude(t, dir, "use forge:setup-linting with project name testapp")
	assertFileExists(t, dir, ".golangci.yml")
	runMake(t, dir, "lint")
}

func TestSetupBdd(t *testing.T) {
	skipIfNoAPI(t)
	dir := t.TempDir()
	runClaude(t, dir, "use forge:setup-go-module with project name testapp and module github.com/test/testapp")
	runClaude(t, dir, "use forge:setup-makefile with project name testapp")
	runClaude(t, dir, "use forge:setup-bdd with project name testapp")
	assertDirExists(t, dir, "features")
	runMake(t, dir, "test")
}

func TestSetupPlaywright(t *testing.T) {
	skipIfNoAPI(t)
	dir := t.TempDir()
	runClaude(t, dir, "use forge:setup-react with project name testapp")
	runClaude(t, dir, "use forge:setup-playwright with project name testapp")
	assertFileExists(t, dir, "playwright.config.ts")
	assertDirExists(t, dir, "web/e2e")
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
