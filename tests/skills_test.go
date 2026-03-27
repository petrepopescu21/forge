package tests

import (
	"os"
	"strings"
	"testing"
)

func skipIfNoAPI(t *testing.T) {
	t.Helper()
	if os.Getenv("ANTHROPIC_API_KEY") == "" && os.Getenv("ANTHROPIC_AUTH_TOKEN") == "" {
		t.Skip("neither ANTHROPIC_API_KEY nor ANTHROPIC_AUTH_TOKEN set, skipping behavioral test")
	}
}

func TestSetupGoModule(t *testing.T) {
	calls := runSkillTest(t, "setup-go-module",
		"Set up a Go module with project name testapp and module path github.com/test/testapp")
	assertToolCallExists(t, calls, "go mod init")
}

func TestSetupReact(t *testing.T) {
	calls := runSkillTest(t, "setup-react",
		"Set up React frontend with project name testapp")
	// May use bun create vite (Bash) or Write package.json
	assertToolCallExists(t, calls, "vite")
}

func TestSetupMakefile(t *testing.T) {
	calls := runSkillTest(t, "setup-makefile",
		"Create a Makefile for project name testapp")
	assertToolCallExists(t, calls, "Makefile")
	assertToolCallExists(t, calls, "lint")
}

func TestSetupLinting(t *testing.T) {
	calls := runSkillTest(t, "setup-linting",
		"Set up linting for project name testapp")
	assertToolCallExists(t, calls, "golangci")
}

func TestSetupBdd(t *testing.T) {
	calls := runSkillTest(t, "setup-bdd",
		"Set up BDD with godog for project name testapp")
	found := false
	for _, c := range calls {
		for _, v := range c.Input {
			if s, ok := v.(string); ok {
				if strings.Contains(s, "feature") || strings.Contains(s, "godog") || strings.Contains(s, "cucumber") {
					found = true
					break
				}
			}
		}
	}
	if !found {
		t.Errorf("expected tool calls related to BDD/features/godog, got: %s", summarizeCalls(calls))
	}
}

func TestSetupPlaywright(t *testing.T) {
	calls := runSkillTest(t, "setup-playwright",
		"Set up Playwright for project name testapp")
	// May reference playwright or e2e in Bash/Write
	assertToolCallExists(t, calls, "e2e")
}

func TestSetupSonar(t *testing.T) {
	calls := runSkillTest(t, "setup-sonar",
		"Set up SonarCloud for project testapp with project key test_testapp. Write the sonar-project.properties file.")
	assertToolCallExists(t, calls, "sonar")
}

func TestGenerateClaudeMd(t *testing.T) {
	calls := runSkillTest(t, "generate-claude-md",
		"Generate and write a CLAUDE.md file for project testapp, module github.com/test/testapp, description 'a test project'. The project has Go backend and React frontend with forge plugin. Write the CLAUDE.md file now without exploring first.")
	// May write CLAUDE.md directly or explore first depending on model
	found := false
	for _, c := range calls {
		for _, v := range c.Input {
			if s, ok := v.(string); ok {
				if strings.Contains(s, "CLAUDE") || strings.Contains(s, "forge") {
					found = true
					break
				}
			}
		}
	}
	if !found {
		t.Errorf("expected tool calls referencing CLAUDE.md or forge, got: %s", summarizeCalls(calls))
	}
}

func TestSetupCi(t *testing.T) {
	calls := runSkillTest(t, "setup-ci",
		"Set up GitHub Actions CI for project name testapp")
	assertToolCallExists(t, calls, "ci.yml")
	assertToolCallExists(t, calls, "lint")
}

func TestSetupHelm(t *testing.T) {
	calls := runSkillTest(t, "setup-helm",
		"Set up Helm chart for project name testapp")
	assertToolCallExists(t, calls, "helm")
}
