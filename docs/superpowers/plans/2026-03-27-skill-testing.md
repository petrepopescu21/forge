# Forge Skill Testing & Validation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add dependency-aware testing infrastructure so structural lint runs on every PR and behavioral smoke tests run on-demand via `/test-skills` comment.

**Architecture:** Central dependency manifest (`skills/dependencies.yaml`) drives test selection. Go test harness with three assertion levels: file structure + quality gates (L2), infrastructure (L3), CI self-validation (L4). Two GitHub Actions workflows — `ci.yml` for structural lint, `test-skills.yml` for behavioral tests gated to an allowlist.

**Tech Stack:** Go 1.22+ (`go test`), `gopkg.in/yaml.v3` for YAML parsing, Claude Code CLI for behavioral tests, `gh` CLI for temp repo lifecycle, GitHub Actions.

**Spec:** `docs/superpowers/specs/2026-03-27-skill-testing-design.md`

---

### Task 1: Go Module + Dependencies Manifest

**Files:**
- Create: `go.mod`
- Create: `go.sum`
- Create: `skills/dependencies.yaml`

- [ ] **Step 1: Initialize Go module**

```bash
cd /Users/petre/personal/forge
go mod init github.com/petrepopescu21/forge
```

- [ ] **Step 2: Add YAML dependency**

```bash
go get gopkg.in/yaml.v3
```

- [ ] **Step 3: Create `skills/dependencies.yaml`**

```yaml
# Skill dependency graph — source of truth.
# Update this file when adding/removing forge:<name> references in skill.md files.
# CI validates consistency on every PR.

skills:
  # Orchestrators
  bootstrap-project:
    type: orchestrator
    depends:
      - setup-go-module
      - setup-react
      - setup-makefile
      - setup-linting
      - setup-bdd
      - setup-playwright
      - setup-ci
      - setup-sonar
      - setup-helm
      - generate-claude-md

  add-feature:
    type: orchestrator
    depends:
      - bdd-feature
      - tdd-cycle
      - quality-check

  # Workflow skills
  bdd-feature:
    type: workflow
    depends: []
  tdd-cycle:
    type: workflow
    depends: []
  quality-check:
    type: workflow
    depends: []

  # Setup skills
  setup-go-module:
    type: setup
    depends: []
  setup-react:
    type: setup
    depends: []
  setup-makefile:
    type: setup
    depends: []
  setup-linting:
    type: setup
    depends: []
  setup-bdd:
    type: setup
    depends: []
  setup-playwright:
    type: setup
    depends: []
  setup-ci:
    type: setup
    depends: []
  setup-sonar:
    type: setup
    depends: []
  setup-helm:
    type: setup
    depends: []
  generate-claude-md:
    type: setup
    depends: []
```

- [ ] **Step 4: Commit**

```bash
git add go.mod go.sum skills/dependencies.yaml
git commit -m "chore: init Go module and skill dependency manifest"
```

---

### Task 2: Shared Test Helpers (`tests/helpers_test.go`)

**Files:**
- Create: `tests/helpers_test.go`

- [ ] **Step 1: Create the helpers file**

```go
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
// It returns the combined stdout+stderr output.
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
// Registers a t.Cleanup to delete the repo even if the test fails.
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
// Returns true if the run succeeded.
func waitForWorkflowGreen(t *testing.T, repo string, timeoutMinutes int) bool {
	t.Helper()
	deadline := timeoutMinutes * 60 // seconds
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
```

- [ ] **Step 2: Verify it compiles**

```bash
cd /Users/petre/personal/forge && go vet ./tests/
```

- [ ] **Step 3: Commit**

```bash
git add tests/helpers_test.go
git commit -m "test: add shared test helper functions"
```

---

### Task 3: Structural Lint Tests (`tests/lint_test.go`)

**Files:**
- Create: `tests/lint_test.go`

- [ ] **Step 1: Write the failing tests**

```go
package tests

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

const (
	skillsDir    = "../skills"
	pluginFile   = "../plugin.json"
	depsFile     = "../skills/dependencies.yaml"
)

// --- YAML / JSON structures ---

type DepsManifest struct {
	Skills map[string]SkillEntry `yaml:"skills"`
}

type SkillEntry struct {
	Type    string   `yaml:"type"`
	Depends []string `yaml:"depends"`
}

type PluginJSON struct {
	Skills []string `json:"skills"`
}

// --- Helpers ---

func loadDepsManifest(t *testing.T) DepsManifest {
	t.Helper()
	data, err := os.ReadFile(depsFile)
	if err != nil {
		t.Fatalf("failed to read %s: %v", depsFile, err)
	}
	var m DepsManifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		t.Fatalf("failed to parse %s: %v", depsFile, err)
	}
	return m
}

func loadPluginJSON(t *testing.T) PluginJSON {
	t.Helper()
	data, err := os.ReadFile(pluginFile)
	if err != nil {
		t.Fatalf("failed to read %s: %v", pluginFile, err)
	}
	var p PluginJSON
	if err := json.Unmarshal(data, &p); err != nil {
		t.Fatalf("failed to parse %s: %v", pluginFile, err)
	}
	return p
}

func listSkillDirs(t *testing.T) []string {
	t.Helper()
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		t.Fatalf("failed to read %s: %v", skillsDir, err)
	}
	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, e.Name())
		}
	}
	return dirs
}

// --- Tests ---

func TestFrontmatterValid(t *testing.T) {
	frontmatterRe := regexp.MustCompile(`(?s)^---\n(.+?)\n---`)

	for _, dir := range listSkillDirs(t) {
		skillMd := filepath.Join(skillsDir, dir, "skill.md")
		data, err := os.ReadFile(skillMd)
		if err != nil {
			t.Errorf("skill %s: cannot read skill.md: %v", dir, err)
			continue
		}

		match := frontmatterRe.FindSubmatch(data)
		if match == nil {
			t.Errorf("skill %s: missing YAML frontmatter (---...---)", dir)
			continue
		}

		var fm map[string]interface{}
		if err := yaml.Unmarshal(match[1], &fm); err != nil {
			t.Errorf("skill %s: invalid YAML in frontmatter: %v", dir, err)
			continue
		}

		if _, ok := fm["name"]; !ok {
			t.Errorf("skill %s: frontmatter missing 'name' field", dir)
		}
		if _, ok := fm["description"]; !ok {
			t.Errorf("skill %s: frontmatter missing 'description' field", dir)
		}
	}
}

func TestPluginJsonConsistency(t *testing.T) {
	plugin := loadPluginJSON(t)
	dirs := listSkillDirs(t)

	// Build set from plugin.json entries (strip "skills/" prefix)
	pluginSkills := make(map[string]bool)
	for _, s := range plugin.Skills {
		name := strings.TrimPrefix(s, "skills/")
		pluginSkills[name] = true
	}

	// Build set from directories
	dirSet := make(map[string]bool)
	for _, d := range dirs {
		dirSet[d] = true
	}

	// Every directory must be in plugin.json
	for _, d := range dirs {
		if !pluginSkills[d] {
			t.Errorf("skill directory %q exists but is not listed in plugin.json", d)
		}
	}

	// Every plugin.json entry must have a directory
	for name := range pluginSkills {
		if !dirSet[name] {
			t.Errorf("plugin.json lists %q but no skill directory exists", name)
		}
	}
}

func TestDependencyManifestConsistency(t *testing.T) {
	manifest := loadDepsManifest(t)
	dirs := listSkillDirs(t)
	dirSet := make(map[string]bool)
	for _, d := range dirs {
		dirSet[d] = true
	}

	// Every depends entry must point to a real skill directory
	for skill, entry := range manifest.Skills {
		for _, dep := range entry.Depends {
			if !dirSet[dep] {
				t.Errorf("skill %q depends on %q, but no skill directory exists for it", skill, dep)
			}
			if _, ok := manifest.Skills[dep]; !ok {
				t.Errorf("skill %q depends on %q, but %q is not in dependencies.yaml", skill, dep, dep)
			}
		}
	}

	// Every forge:<name> reference in skill.md files must be in the manifest
	forgeRef := regexp.MustCompile(`forge:([a-z][-a-z0-9]*)`)
	for _, dir := range dirs {
		skillMd := filepath.Join(skillsDir, dir, "skill.md")
		data, err := os.ReadFile(skillMd)
		if err != nil {
			continue
		}
		matches := forgeRef.FindAllSubmatch(data, -1)
		for _, m := range matches {
			ref := string(m[1])
			if _, ok := manifest.Skills[ref]; !ok {
				t.Errorf("skill %q references forge:%s, but %s is not in dependencies.yaml", dir, ref, ref)
			}
		}
	}
}

func TestNoOrphanSkills(t *testing.T) {
	manifest := loadDepsManifest(t)
	plugin := loadPluginJSON(t)
	dirs := listSkillDirs(t)

	pluginSkills := make(map[string]bool)
	for _, s := range plugin.Skills {
		pluginSkills[strings.TrimPrefix(s, "skills/")] = true
	}

	dirSet := make(map[string]bool)
	for _, d := range dirs {
		dirSet[d] = true
	}

	for name := range manifest.Skills {
		if !pluginSkills[name] {
			t.Errorf("dependencies.yaml has %q but it is not in plugin.json", name)
		}
		if !dirSet[name] {
			t.Errorf("dependencies.yaml has %q but no skill directory exists", name)
		}
	}
}

func TestNoCyclicDependencies(t *testing.T) {
	manifest := loadDepsManifest(t)

	// Kahn's algorithm for topological sort
	inDegree := make(map[string]int)
	for name := range manifest.Skills {
		inDegree[name] = 0
	}
	for _, entry := range manifest.Skills {
		for _, dep := range entry.Depends {
			inDegree[dep]++
		}
	}

	var queue []string
	for name, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, name)
		}
	}

	visited := 0
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		visited++
		for _, dep := range manifest.Skills[node].Depends {
			inDegree[dep]--
			if inDegree[dep] == 0 {
				queue = append(queue, dep)
			}
		}
	}

	if visited != len(manifest.Skills) {
		// Find the cycle participants
		var cycleNodes []string
		for name, deg := range inDegree {
			if deg > 0 {
				cycleNodes = append(cycleNodes, name)
			}
		}
		t.Errorf("dependency cycle detected involving: %s", strings.Join(cycleNodes, ", "))
	}
}
```

- [ ] **Step 2: Run the tests to verify they pass against current state**

```bash
cd /Users/petre/personal/forge && go test ./tests/ -run 'TestFrontmatter|TestPluginJson|TestDependency|TestNoOrphan|TestNoCyclic' -v
```

Expected: all 5 tests PASS (the current repo state is consistent).

- [ ] **Step 3: Verify tests catch real errors by introducing a deliberate break**

Temporarily add a bad entry to `dependencies.yaml` (e.g., a skill that references a non-existent directory), run the tests, confirm they fail, then revert.

- [ ] **Step 4: Commit**

```bash
git add tests/lint_test.go
git commit -m "test: add structural lint tests for skill validation"
```

---

### Task 4: Dependency-Aware Test Selector (`tests/cmd/affected/main.go`)

**Files:**
- Create: `tests/cmd/affected/main.go`

- [ ] **Step 1: Create the affected CLI**

```go
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type DepsManifest struct {
	Skills map[string]SkillEntry `yaml:"skills"`
}

type SkillEntry struct {
	Type    string   `yaml:"type"`
	Depends []string `yaml:"depends"`
}

func main() {
	base := "origin/main"
	if len(os.Args) > 2 && os.Args[1] == "--base" {
		base = os.Args[2]
	}

	// Get changed files
	out, err := exec.Command("git", "diff", "--name-only", base+"...HEAD").Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "git diff failed: %v\n", err)
		os.Exit(1)
	}

	changedFiles := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(changedFiles) == 1 && changedFiles[0] == "" {
		// No changes
		os.Exit(0)
	}

	// Check for infra changes that trigger all tests
	for _, f := range changedFiles {
		if f == "skills/dependencies.yaml" || f == "plugin.json" ||
			strings.HasPrefix(f, "tests/") {
			fmt.Println(".*") // run all tests
			os.Exit(0)
		}
	}

	// Load manifest
	data, err := os.ReadFile("skills/dependencies.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read dependencies.yaml: %v\n", err)
		os.Exit(1)
	}
	var manifest DepsManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse dependencies.yaml: %v\n", err)
		os.Exit(1)
	}

	// Map changed files to affected skills
	affected := make(map[string]bool)
	for _, f := range changedFiles {
		if !strings.HasPrefix(f, "skills/") {
			continue
		}
		// Extract skill name: skills/<name>/skill.md -> <name>
		parts := strings.SplitN(strings.TrimPrefix(f, "skills/"), "/", 2)
		if len(parts) == 0 {
			continue
		}
		skillName := parts[0]
		if _, ok := manifest.Skills[skillName]; ok {
			affected[skillName] = true
		}
	}

	if len(affected) == 0 {
		// No skill files changed
		os.Exit(0)
	}

	// Resolve parent orchestrators: if a leaf skill changed, its parent needs testing too
	for name, entry := range manifest.Skills {
		for _, dep := range entry.Depends {
			if affected[dep] {
				affected[name] = true
			}
		}
	}

	// Convert to Go test -run regex
	// Skill name "setup-go-module" -> test name "TestSetupGoModule"
	var testNames []string
	for name := range affected {
		testName := "Test" + toTestName(name)
		testNames = append(testNames, testName)
	}

	fmt.Println(strings.Join(testNames, "|"))
}

// toTestName converts "setup-go-module" to "SetupGoModule"
func toTestName(skillName string) string {
	parts := strings.Split(skillName, "-")
	var result strings.Builder
	for _, p := range parts {
		if len(p) > 0 {
			result.WriteString(strings.ToUpper(p[:1]))
			result.WriteString(p[1:])
		}
	}
	return result.String()
}
```

- [ ] **Step 2: Verify it compiles**

```bash
cd /Users/petre/personal/forge && go build ./tests/cmd/affected/
```

- [ ] **Step 3: Test manually with a fake diff**

Create a test branch, touch a skill file, run the tool, verify output:

```bash
git checkout -b test-affected
touch skills/setup-helm/test-change
git add skills/setup-helm/test-change
git commit -m "test: dummy change"
go run ./tests/cmd/affected/ --base main
```

Expected output: `TestSetupHelm|TestBootstrapProject`

Then clean up:

```bash
git checkout main
git branch -D test-affected
```

- [ ] **Step 4: Commit**

```bash
git add tests/cmd/affected/main.go
git commit -m "feat: add dependency-aware test selector CLI"
```

---

### Task 5: Behavioral Tests — Setup Skills Level 2 (`tests/skills_test.go`)

**Files:**
- Create: `tests/skills_test.go`

These tests require Claude Code CLI and are skipped if `ANTHROPIC_API_KEY` is not set.

- [ ] **Step 1: Write the behavioral test file with setup skill tests**

```go
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

// --- Level 2: Setup Skill Tests ---

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
	// Linting needs Go module + React to have something to lint
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
```

- [ ] **Step 2: Verify it compiles**

```bash
cd /Users/petre/personal/forge && go vet ./tests/
```

- [ ] **Step 3: Commit**

```bash
git add tests/skills_test.go
git commit -m "test: add Level 2 behavioral tests for setup skills"
```

---

### Task 6: Behavioral Tests — Level 3 Infrastructure (`tests/infra_test.go`)

**Files:**
- Create: `tests/infra_test.go`

- [ ] **Step 1: Write the infrastructure test**

```go
package tests

import (
	"testing"
	"time"
)

func TestSetupHelm(t *testing.T) {
	skipIfNoAPI(t)
	if testing.Short() {
		t.Skip("skipping infrastructure test in short mode")
	}

	dir := t.TempDir()

	// Bootstrap enough of a project for Helm to work
	runClaude(t, dir, "use forge:setup-go-module with project name testapp and module github.com/test/testapp")
	runClaude(t, dir, "use forge:setup-makefile with project name testapp")
	runClaude(t, dir, "use forge:setup-helm with project name testapp")

	// L2: file structure
	assertFileExists(t, dir, "deploy/helm/testapp/Chart.yaml")
	assertFileContains(t, dir, "deploy/helm/testapp/Chart.yaml", "name: testapp")
	runCmd(t, dir, "helm", "lint", "deploy/helm/testapp")

	// L3: Kind cluster + Helm deploy
	runMake(t, dir, "cluster-create")
	t.Cleanup(func() {
		// Tear down the Kind cluster even if the test fails
		runCmd(t, dir, "make", "cluster-delete")
	})

	// Build the Docker image and load into Kind
	runMake(t, dir, "build")
	runCmd(t, dir, "docker", "build", "-t", "testapp:latest", ".")
	runCmd(t, dir, "kind", "load", "docker-image", "testapp:latest", "--name", "testapp")

	// Deploy with Helm
	runCmd(t, dir, "helm", "install", "testapp", "deploy/helm/testapp",
		"--set", "image.repository=testapp",
		"--set", "image.tag=latest",
		"--set", "image.pullPolicy=Never",
		"--wait", "--timeout", "120s")

	// Wait for pod readiness, then check HTTP
	time.Sleep(10 * time.Second)
	code := httpGet(t, "http://localhost:8080/healthz")
	if code != 200 {
		t.Errorf("expected HTTP 200 from /healthz, got %d", code)
	}
}
```

- [ ] **Step 2: Verify it compiles**

```bash
cd /Users/petre/personal/forge && go vet ./tests/
```

- [ ] **Step 3: Commit**

```bash
git add tests/infra_test.go
git commit -m "test: add Level 3 infrastructure test for setup-helm"
```

---

### Task 7: Behavioral Tests — Level 4 CI Self-Validation (`tests/ci_test.go`)

**Files:**
- Create: `tests/ci_test.go`

- [ ] **Step 1: Write the CI self-validation test**

```go
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

	// Generate the project with CI
	runClaude(t, dir, "use forge:setup-go-module with project name testapp and module github.com/test/testapp")
	runClaude(t, dir, "use forge:setup-react with project name testapp")
	runClaude(t, dir, "use forge:setup-makefile with project name testapp")
	runClaude(t, dir, "use forge:setup-linting with project name testapp")
	runClaude(t, dir, "use forge:setup-ci with project name testapp")

	// L2: file structure assertions
	assertFileExists(t, dir, ".github/workflows/ci.yml")
	assertFileContains(t, dir, ".github/workflows/ci.yml", "lint")
	assertFileContains(t, dir, ".github/workflows/ci.yml", "test")

	// L4: push to temp GitHub repo and validate Actions run
	repoName := fmt.Sprintf("forge-test-%d", os.Getpid())
	ghToken := os.Getenv("FORGE_TEST_GITHUB_TOKEN")

	// Create the temp repo
	runCmd(t, dir, "gh", "repo", "create", repoName, "--private", "--confirm")
	t.Cleanup(func() {
		exec.Command("gh", "repo", "delete", repoName, "--yes").Run()
	})

	// Init git, commit, push
	runCmd(t, dir, "git", "init")
	runCmd(t, dir, "git", "config", "user.name", "forge-test")
	runCmd(t, dir, "git", "config", "user.email", "forge-test@test.local")
	runCmd(t, dir, "git", "add", ".")
	runCmd(t, dir, "git", "commit", "-m", "initial: test scaffold")

	owner := runCmd(t, dir, "gh", "api", "user", "--jq", ".login")
	remote := fmt.Sprintf("https://x-access-token:%s@github.com/%s/%s.git",
		ghToken, strings.TrimSpace(owner), repoName)
	runCmd(t, dir, "git", "remote", "add", "origin", remote)
	runCmd(t, dir, "git", "push", "-u", "origin", "main")

	// Wait for workflow to complete
	fullRepo := fmt.Sprintf("%s/%s", strings.TrimSpace(owner), repoName)
	success := waitForWorkflowGreen(t, fullRepo, 15)
	if !success {
		// Fetch logs for debugging
		out := runCmd(t, ".", "gh", "run", "list", "--repo", fullRepo, "--limit", "1",
			"--json", "databaseId,conclusion")
		t.Fatalf("CI workflow did not succeed. Run info: %s", out)
	}
}
```

Note: This file uses `strings` — add the import. The `strings` package is already imported in `helpers_test.go` but since this is a separate file in the same package, it needs its own import.

- [ ] **Step 2: Add missing import and verify compilation**

Ensure the file has `"strings"` in its import block. Then:

```bash
cd /Users/petre/personal/forge && go vet ./tests/
```

- [ ] **Step 3: Commit**

```bash
git add tests/ci_test.go
git commit -m "test: add Level 4 CI self-validation test for setup-ci"
```

---

### Task 8: Behavioral Tests — Orchestrators + Workflow Skills (`tests/orchestrator_test.go`)

**Files:**
- Create: `tests/orchestrator_test.go`

- [ ] **Step 1: Write the orchestrator and workflow tests**

```go
package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fixtureDir holds a pre-bootstrapped project shared across workflow tests.
// Set by TestMain in main_test.go.
var fixtureDir string

func TestBootstrapProject(t *testing.T) {
	skipIfNoAPI(t)
	dir := t.TempDir()

	runClaude(t, dir, "use forge:bootstrap-project with name testapp, module github.com/test/testapp, description 'a test project', all layers yes")

	// L2: file structure + quality gates
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

	// BDD scenario created
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

	// Tests pass after implementation
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

	// Should create a .feature file with valid Gherkin
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

	// First create a BDD scenario to drive TDD
	runClaude(t, projectDir, "use forge:bdd-feature with prompt: 'user can get server version at GET /version'")
	runClaude(t, projectDir, "use forge:tdd-cycle to implement the version endpoint step definitions")

	// Tests should pass
	runMake(t, projectDir, "test")

	// Git log should show red/green/refactor pattern
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

	// Quality check itself doesn't produce files — it runs gates.
	// The fact that runClaude didn't fail means the skill ran.
	// Verify gates pass directly too.
	runMake(t, projectDir, "lint")
	runMake(t, projectDir, "typecheck")
	runMake(t, projectDir, "test")
}
```

Note: This file uses `strings` — needs its own import.

- [ ] **Step 2: Create `tests/main_test.go` for shared bootstrap fixture**

```go
package tests

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
)

func TestMain(m *testing.M) {
	// If running behavioral tests (API key present), bootstrap a shared fixture
	if os.Getenv("ANTHROPIC_API_KEY") != "" {
		dir, err := os.MkdirTemp("", "forge-fixture-*")
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to create fixture dir: %v\n", err)
			os.Exit(1)
		}
		defer os.RemoveAll(dir)

		// Bootstrap a project for workflow tests to share
		cmd := exec.Command("claude", "-p", "--allowedTools", "*",
			"use forge:bootstrap-project with name testapp, module github.com/test/testapp, description 'a test project', all layers yes")
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Fprintf(os.Stderr, "bootstrap fixture failed: %v\nOutput:\n%s\n", err, string(out))
			// Don't exit — let tests skip via fixtureDir == ""
		} else {
			fixtureDir = dir
		}
	}

	os.Exit(m.Run())
}
```

Note: needs `"os/exec"` import.

- [ ] **Step 3: Verify compilation**

```bash
cd /Users/petre/personal/forge && go vet ./tests/
```

- [ ] **Step 4: Commit**

```bash
git add tests/orchestrator_test.go tests/main_test.go
git commit -m "test: add orchestrator and workflow behavioral tests"
```

---

### Task 9: GitHub Actions — `ci.yml`

**Files:**
- Create: `.github/workflows/ci.yml`

- [ ] **Step 1: Create the CI workflow**

```yaml
name: CI

on:
  pull_request:
    branches: [main]

jobs:
  structural-lint:
    name: Structural Lint
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: "1.22"

      - name: Run structural lint tests
        run: go test ./tests/ -run 'TestFrontmatter|TestPluginJson|TestDependency|TestNoOrphan|TestNoCyclic' -v
```

- [ ] **Step 2: Commit**

```bash
mkdir -p .github/workflows
git add .github/workflows/ci.yml
git commit -m "ci: add structural lint workflow for every PR"
```

---

### Task 10: GitHub Actions — `test-skills.yml`

**Files:**
- Create: `.github/workflows/test-skills.yml`

- [ ] **Step 1: Create the behavioral test workflow**

```yaml
name: Behavioral Skill Tests

on:
  issue_comment:
    types: [created]

jobs:
  test-skills:
    name: Behavioral Tests
    if: >
      github.event.issue.pull_request &&
      contains(github.event.comment.body, '/test-skills') &&
      contains(fromJson('["petrepopescu21"]'), github.event.comment.user.login)
    runs-on: ubuntu-latest
    timeout-minutes: 45
    steps:
      - name: Acknowledge trigger
        uses: actions/github-script@v7
        with:
          script: |
            github.rest.reactions.createForIssueComment({
              owner: context.repo.owner,
              repo: context.repo.repo,
              comment_id: context.payload.comment.id,
              content: 'eyes'
            });

      - name: Checkout PR branch
        uses: actions/checkout@v4
        with:
          ref: refs/pull/${{ github.event.issue.number }}/head
          fetch-depth: 0

      - name: Fetch base branch
        run: git fetch origin main

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: "1.22"

      - name: Setup Node
        uses: actions/setup-node@v4
        with:
          node-version: "20"

      - name: Install Claude Code
        run: npm install -g @anthropic-ai/claude-code

      - name: Install Bun
        uses: oven-sh/setup-bun@v2

      - name: Install Helm
        uses: azure/setup-helm@v4

      - name: Determine affected tests
        id: affected
        run: |
          AFFECTED=$(go run ./tests/cmd/affected/ --base origin/main)
          echo "pattern=$AFFECTED" >> "$GITHUB_OUTPUT"
          if [ -z "$AFFECTED" ]; then
            echo "No skills affected by this PR"
            echo "skip=true" >> "$GITHUB_OUTPUT"
          else
            echo "Running tests matching: $AFFECTED"
            echo "skip=false" >> "$GITHUB_OUTPUT"
          fi

      - name: Run behavioral tests
        if: steps.affected.outputs.skip != 'true'
        env:
          ANTHROPIC_API_KEY: ${{ secrets.ANTHROPIC_API_KEY }}
          FORGE_TEST_GITHUB_TOKEN: ${{ secrets.FORGE_TEST_GITHUB_TOKEN }}
        run: |
          go test ./tests/ -run '${{ steps.affected.outputs.pattern }}' -v -timeout 30m 2>&1 | tee test-output.txt

      - name: Post results
        if: always() && steps.affected.outputs.skip != 'true'
        uses: actions/github-script@v7
        with:
          script: |
            const fs = require('fs');
            let output = '';
            try {
              output = fs.readFileSync('test-output.txt', 'utf8');
            } catch (e) {
              output = 'No test output captured';
            }

            // Extract pass/fail summary
            const lines = output.split('\n');
            const resultLines = lines.filter(l =>
              l.startsWith('--- PASS') || l.startsWith('--- FAIL') || l.startsWith('--- SKIP') || l.startsWith('ok') || l.startsWith('FAIL')
            );

            const status = '${{ job.status }}' === 'success' ? '✅' : '❌';
            const body = [
              `${status} **Behavioral skill tests: ${{ job.status }}**`,
              '',
              'Pattern: `${{ steps.affected.outputs.pattern }}`',
              '',
              '<details><summary>Test Results</summary>',
              '',
              '```',
              resultLines.join('\n') || output.slice(-2000),
              '```',
              '</details>'
            ].join('\n');

            github.rest.issues.createComment({
              owner: context.repo.owner,
              repo: context.repo.repo,
              issue_number: context.issue.number,
              body: body
            });
```

- [ ] **Step 2: Commit**

```bash
git add .github/workflows/test-skills.yml
git commit -m "ci: add on-demand behavioral test workflow with author gating"
```

---

### Task 11: CLAUDE.md for the Forge Repo

**Files:**
- Create: `CLAUDE.md`

- [ ] **Step 1: Create CLAUDE.md**

```markdown
# Forge — Claude Code Superpowers Plugin

## What This Is

A Claude Code superpowers plugin that scaffolds and enforces BDD/TDD workflows for Go + React/TypeScript projects. This repo contains skill definitions (markdown), a Go test harness, and CI workflows.

## Project Structure

- `skills/` — 15 skill definitions (markdown files with YAML frontmatter)
- `skills/dependencies.yaml` — skill dependency graph (source of truth)
- `plugin.json` — plugin manifest listing all skills
- `tests/` — Go test harness (structural lint + behavioral smoke tests)
- `tests/cmd/affected/` — CLI for dependency-aware test selection
- `.github/workflows/` — CI (structural lint) + behavioral tests (on-demand)

## Mandatory Rules

### Dependency Manifest

When modifying any `skill.md` that adds or removes a `forge:<name>` reference:
- **Update `skills/dependencies.yaml`** to reflect the new dependency
- **Update `plugin.json`** if adding or removing a skill directory

When adding or removing a skill directory:
- **Update both `plugin.json` and `skills/dependencies.yaml`**

### Validation

Run structural lint before committing skill changes:
```bash
go test ./tests/ -run 'TestFrontmatter|TestPluginJson|TestDependency|TestNoOrphan|TestNoCyclic' -v
```

### Skill Structure

Every skill must have a `skill.md` with YAML frontmatter containing at minimum:
- `name` — skill identifier (matches directory name)
- `description` — trigger description for Claude Code

### Testing

- Structural tests run on every PR automatically
- Behavioral tests run via `/test-skills` PR comment (authorized users only)
- Behavioral tests require `ANTHROPIC_API_KEY` environment variable
- Level 4 tests (CI self-validation) require `FORGE_TEST_GITHUB_TOKEN`
```

- [ ] **Step 2: Commit**

```bash
git add CLAUDE.md
git commit -m "docs: add CLAUDE.md with repo conventions and enforcement rules"
```

---

### Task 12: Install skill-creator Plugin

**Files:**
- Modify: User's Claude Code settings (not a repo file)

- [ ] **Step 1: Check if skill-creator is available as a plugin**

```bash
claude plugins search skill-creator
```

Or add to the user's Claude Code settings manually. This is a configuration step, not a code change.

- [ ] **Step 2: Verify the plugin is accessible**

Test by asking Claude to use the skill-creator skill in this repo.

- [ ] **Step 3: Document in CLAUDE.md**

No code change needed — the CLAUDE.md from Task 11 already covers the repo's purpose as a skill development repo.

---

### Task 13: Final Verification

- [ ] **Step 1: Run structural lint tests locally**

```bash
cd /Users/petre/personal/forge && go test ./tests/ -run 'TestFrontmatter|TestPluginJson|TestDependency|TestNoOrphan|TestNoCyclic' -v
```

Expected: all 5 tests PASS.

- [ ] **Step 2: Run the affected CLI to verify it works**

```bash
cd /Users/petre/personal/forge && go run ./tests/cmd/affected/ --base HEAD~1
```

Expected: outputs a test pattern based on the last commit's changes.

- [ ] **Step 3: Verify Go module is tidy**

```bash
cd /Users/petre/personal/forge && go mod tidy
```

- [ ] **Step 4: Final commit if go mod tidy changed anything**

```bash
git add go.mod go.sum
git commit -m "chore: tidy Go module" || echo "nothing to commit"
```
