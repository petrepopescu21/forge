# Forge CLI Rewrite — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace 11 deterministic AI skills with a Go CLI that uses embedded templates, keeping only 3 AI-powered skills (bdd-feature, tdd-cycle, add-feature) and a thin bootstrap-project wrapper skill.

**Architecture:** Go binary in `cmd/forge/` uses `embed.FS` + `text/template` to render project scaffolds. Templates live in `internal/templates/<layer>/`. A single `internal/scaffold/` package walks the embedded tree, processes templates, and writes output. The bootstrap-project skill gathers user inputs interactively then shells out to `go run ./cmd/forge`.

**Tech Stack:** Go embed, text/template, cobra (CLI flags), Renovate (dependency updates for Makefile tool versions)

---

## File Structure

```
forge/
├── cmd/forge/
│   └── main.go                         # CLI: parse flags, call scaffold.Run()
├── internal/
│   ├── scaffold/
│   │   ├── scaffold.go                 # Config struct, Run() function, template walker
│   │   └── scaffold_test.go            # Unit tests: each layer produces expected files
│   └── templates/                      # embed.FS root
│       ├── common/                     # Always rendered
│       │   ├── .gitignore.tmpl
│       │   ├── Dockerfile.tmpl
│       │   └── README.md.tmpl
│       ├── go-module/
│       │   ├── cmd/%Name%/main.go.tmpl
│       │   └── internal/api/router.go.tmpl
│       ├── react/
│       │   ├── web/package.json.tmpl
│       │   ├── web/vite.config.ts.tmpl
│       │   ├── web/vitest.config.ts.tmpl
│       │   ├── web/tsconfig.json.tmpl
│       │   ├── web/tsconfig.node.json.tmpl
│       │   └── web/src/test/setup.ts.tmpl
│       ├── makefile/
│       │   └── Makefile.tmpl
│       ├── linting/
│       │   ├── .golangci.yml.tmpl
│       │   └── web/eslint.config.js.tmpl
│       ├── bdd/
│       │   ├── features/health.feature.tmpl
│       │   ├── internal/api/health_bdd_test.go.tmpl
│       │   ├── web/features/navigation.feature.tmpl
│       │   └── web/e2e/steps/navigation.steps.ts.tmpl
│       ├── playwright/
│       │   ├── web/playwright.config.ts.tmpl
│       │   ├── web/playwright-integration.config.ts.tmpl
│       │   ├── web/e2e/fixtures.ts.tmpl
│       │   ├── web/e2e/navigation.spec.ts.tmpl
│       │   └── scripts/e2e-web.sh.tmpl
│       ├── ci/
│       │   ├── .github/workflows/ci.yml.tmpl
│       │   ├── .github/workflows/deploy.yml.tmpl
│       │   └── renovate.json.tmpl
│       ├── sonar/
│       │   └── sonar-project.properties.tmpl
│       ├── helm/
│       │   ├── deploy/helm/%Name%/Chart.yaml.tmpl
│       │   ├── deploy/helm/%Name%/values.yaml.tmpl
│       │   ├── deploy/helm/%Name%/values-aks.yaml.tmpl
│       │   ├── deploy/kind/kind-config.yaml.tmpl
│       │   ├── Tiltfile.tmpl
│       │   ├── scripts/cluster-db.sh.tmpl
│       │   └── scripts/cluster-deps.sh.tmpl
│       └── claude-md/
│           └── CLAUDE.md.tmpl
├── skills/
│   ├── bootstrap-project/skill.md      # Thin wrapper: gather inputs → run CLI
│   ├── add-feature/skill.md            # Unchanged
│   ├── bdd-feature/skill.md            # Unchanged
│   └── tdd-cycle/skill.md              # Unchanged
├── tests/
│   ├── lint_test.go                    # Updated for 4 skills
│   ├── cmd/affected/main.go            # Updated
│   └── (sdk_helpers_test.go removed)
│   └── (skills_test.go removed)
├── plugin.json                         # 4 skills
├── skills/dependencies.yaml            # Updated
└── go.mod                              # Remove anthropic SDK dep
```

**Path convention:** Template paths use `%Name%` as a directory placeholder (replaced at walk time with `Config.Name`). File contents use Go template syntax (`{{.Name}}`, `{{.Module}}`, `{{.Description}}`).

---

### Task 1: Create the scaffold engine

**Files:**
- Create: `internal/scaffold/scaffold.go`
- Create: `internal/scaffold/scaffold_test.go`

- [ ] **Step 1: Write the failing test for scaffold.Run()**

```go
// internal/scaffold/scaffold_test.go
package scaffold_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/petrepopescu21/forge/internal/scaffold"
)

func TestRun_GoModuleLayer(t *testing.T) {
	t.Parallel()
	dest := t.TempDir()

	cfg := scaffold.Config{
		Name:        "testapp",
		Module:      "github.com/test/testapp",
		Description: "A test application",
		Layers:      []string{"go-module"},
	}

	if err := scaffold.Run(cfg, dest); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// Verify key files exist
	wantFiles := []string{
		"cmd/testapp/main.go",
		"internal/api/router.go",
	}
	for _, f := range wantFiles {
		path := filepath.Join(dest, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", f)
		}
	}

	// Verify template variables were substituted
	data, err := os.ReadFile(filepath.Join(dest, "cmd/testapp/main.go"))
	if err != nil {
		t.Fatalf("reading main.go: %v", err)
	}
	content := string(data)
	if !contains(content, "github.com/test/testapp/internal/api") {
		t.Errorf("main.go should contain module import, got:\n%s", content)
	}
	if contains(content, "{{") {
		t.Errorf("main.go contains unresolved template syntax")
	}
}

func TestRun_CommonLayer(t *testing.T) {
	t.Parallel()
	dest := t.TempDir()

	cfg := scaffold.Config{
		Name:        "testapp",
		Module:      "github.com/test/testapp",
		Description: "A test application",
		Layers:      []string{}, // common is always rendered
	}

	if err := scaffold.Run(cfg, dest); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	wantFiles := []string{
		".gitignore",
		"Dockerfile",
		"README.md",
	}
	for _, f := range wantFiles {
		path := filepath.Join(dest, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", f)
		}
	}
}

func TestRun_AllLayers(t *testing.T) {
	t.Parallel()
	dest := t.TempDir()

	cfg := scaffold.Config{
		Name:        "testapp",
		Module:      "github.com/test/testapp",
		Description: "A test application",
		Layers:      scaffold.AllLayers,
	}

	if err := scaffold.Run(cfg, dest); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// Spot-check representative files from each layer
	wantFiles := []string{
		".gitignore",
		"Dockerfile",
		"README.md",
		"cmd/testapp/main.go",
		"internal/api/router.go",
		"web/package.json",
		"web/vite.config.ts",
		"Makefile",
		".golangci.yml",
		"web/eslint.config.js",
		"features/health.feature",
		"web/playwright.config.ts",
		".github/workflows/ci.yml",
		"sonar-project.properties",
		"deploy/helm/testapp/values.yaml",
		"deploy/kind/kind-config.yaml",
		"Tiltfile",
		"CLAUDE.md",
		"renovate.json",
	}
	for _, f := range wantFiles {
		path := filepath.Join(dest, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", f)
		}
	}
}

func TestRun_NoUnresolvedTemplates(t *testing.T) {
	t.Parallel()
	dest := t.TempDir()

	cfg := scaffold.Config{
		Name:        "testapp",
		Module:      "github.com/test/testapp",
		Description: "A test application",
		Layers:      scaffold.AllLayers,
	}

	if err := scaffold.Run(cfg, dest); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// Walk all generated files and check for unresolved {{
	err := filepath.Walk(dest, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if contains(string(data), "{{") {
			rel, _ := filepath.Rel(dest, path)
			t.Errorf("file %s contains unresolved template syntax", rel)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking output dir: %v", err)
	}
}

func TestRun_SelectiveLayers(t *testing.T) {
	t.Parallel()
	dest := t.TempDir()

	cfg := scaffold.Config{
		Name:        "testapp",
		Module:      "github.com/test/testapp",
		Description: "A test application",
		Layers:      []string{"go-module", "makefile"},
	}

	if err := scaffold.Run(cfg, dest); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// go-module files should exist
	if _, err := os.Stat(filepath.Join(dest, "cmd/testapp/main.go")); os.IsNotExist(err) {
		t.Error("expected cmd/testapp/main.go to exist")
	}

	// react files should NOT exist
	if _, err := os.Stat(filepath.Join(dest, "web/package.json")); !os.IsNotExist(err) {
		t.Error("expected web/package.json to NOT exist with react layer disabled")
	}
}

func TestRun_InvalidLayer(t *testing.T) {
	t.Parallel()
	dest := t.TempDir()

	cfg := scaffold.Config{
		Name:   "testapp",
		Module: "github.com/test/testapp",
		Layers: []string{"nonexistent"},
	}

	err := scaffold.Run(cfg, dest)
	if err == nil {
		t.Fatal("expected error for invalid layer, got nil")
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && indexOf(s, substr) >= 0
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/petre/personal/forge && go test ./internal/scaffold/ -v -run TestRun -count=1`
Expected: FAIL — package does not exist yet

- [ ] **Step 3: Write the scaffold engine**

```go
// internal/scaffold/scaffold.go
package scaffold

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed templates
var templates embed.FS

// AllLayers lists every available layer name.
var AllLayers = []string{
	"go-module",
	"react",
	"makefile",
	"linting",
	"bdd",
	"playwright",
	"ci",
	"sonar",
	"helm",
	"claude-md",
}

// Config holds the template variables for scaffold generation.
type Config struct {
	Name        string
	Module      string
	Description string
	Layers      []string
}

// Run renders all selected layers (plus "common") into destDir.
func Run(cfg Config, destDir string) error {
	layerSet := map[string]bool{"common": true}
	for _, l := range cfg.Layers {
		if !isValidLayer(l) {
			return fmt.Errorf("unknown layer: %q", l)
		}
		layerSet[l] = true
	}

	for layer := range layerSet {
		root := filepath.Join("templates", layer)
		err := fs.WalkDir(templates, root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}

			// Compute output path: strip "templates/<layer>/" prefix and ".tmpl" suffix
			rel := strings.TrimPrefix(path, root+"/")
			rel = strings.TrimSuffix(rel, ".tmpl")

			// Replace %Name% directory placeholder
			rel = strings.ReplaceAll(rel, "%Name%", cfg.Name)

			outPath := filepath.Join(destDir, rel)

			// Read template content
			data, err := templates.ReadFile(path)
			if err != nil {
				return fmt.Errorf("reading template %s: %w", path, err)
			}

			// Parse and execute template
			tmpl, err := template.New(filepath.Base(path)).Parse(string(data))
			if err != nil {
				return fmt.Errorf("parsing template %s: %w", path, err)
			}

			// Create parent directories
			if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
				return fmt.Errorf("creating directory for %s: %w", outPath, err)
			}

			f, err := os.Create(outPath)
			if err != nil {
				return fmt.Errorf("creating %s: %w", outPath, err)
			}
			defer f.Close()

			if err := tmpl.Execute(f, cfg); err != nil {
				return fmt.Errorf("executing template %s: %w", path, err)
			}

			// Make shell scripts executable
			if strings.HasSuffix(rel, ".sh") {
				if err := os.Chmod(outPath, 0o755); err != nil {
					return fmt.Errorf("chmod %s: %w", outPath, err)
				}
			}

			return nil
		})
		if err != nil {
			return fmt.Errorf("processing layer %s: %w", layer, err)
		}
	}

	return nil
}

func isValidLayer(name string) bool {
	for _, l := range AllLayers {
		if l == name {
			return true
		}
	}
	return false
}
```

- [ ] **Step 4: Run test to verify it fails (no templates yet)**

Run: `cd /Users/petre/personal/forge && go test ./internal/scaffold/ -v -run TestRun_CommonLayer -count=1`
Expected: FAIL — templates directory is empty, no files rendered

- [ ] **Step 5: Commit scaffold engine skeleton**

```bash
git add internal/scaffold/scaffold.go internal/scaffold/scaffold_test.go
git commit -m "feat: add scaffold engine with embed + text/template"
```

---

### Task 2: Create common layer templates

**Files:**
- Create: `internal/scaffold/templates/common/.gitignore.tmpl`
- Create: `internal/scaffold/templates/common/Dockerfile.tmpl`
- Create: `internal/scaffold/templates/common/README.md.tmpl`

- [ ] **Step 1: Create .gitignore template**

Extract from `skills/bootstrap-project/skill.md` lines 96-156. This file has no template variables — copy as-is into `.gitignore.tmpl`.

- [ ] **Step 2: Create Dockerfile template**

Extract from `skills/bootstrap-project/skill.md` lines 163-219. Replace hardcoded project references:
- `$(basename $(pwd))` → `{{.Name}}`

- [ ] **Step 3: Create README.md template**

Extract from `skills/bootstrap-project/skill.md` lines 348-428. Replace:
- `myproject` → `{{.Name}}`
- `github.com/petrepopescu21/myproject` → `{{.Module}}`
- `A self-hosted CRM for field sales` → `{{.Description}}`

- [ ] **Step 4: Run common layer test**

Run: `cd /Users/petre/personal/forge && go test ./internal/scaffold/ -v -run TestRun_CommonLayer -count=1`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/scaffold/templates/common/
git commit -m "feat: add common layer templates (gitignore, Dockerfile, README)"
```

---

### Task 3: Create go-module layer templates

**Files:**
- Create: `internal/scaffold/templates/go-module/cmd/%Name%/main.go.tmpl`
- Create: `internal/scaffold/templates/go-module/internal/api/router.go.tmpl`

- [ ] **Step 1: Create main.go template**

Extract from `skills/setup-go-module/skill.md` lines 51-99. Replace:
- `github.com/petrepopescu21/myproject/internal/api` → `{{.Module}}/internal/api`

- [ ] **Step 2: Create router.go template**

Extract from `skills/setup-go-module/skill.md` lines 107-127. No template variables needed — this file is module-independent.

- [ ] **Step 3: Run go-module layer test**

Run: `cd /Users/petre/personal/forge && go test ./internal/scaffold/ -v -run TestRun_GoModuleLayer -count=1`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/scaffold/templates/go-module/
git commit -m "feat: add go-module layer templates"
```

---

### Task 4: Create react layer templates

**Files:**
- Create: `internal/scaffold/templates/react/web/vite.config.ts.tmpl`
- Create: `internal/scaffold/templates/react/web/vitest.config.ts.tmpl`
- Create: `internal/scaffold/templates/react/web/tsconfig.json.tmpl`
- Create: `internal/scaffold/templates/react/web/tsconfig.node.json.tmpl`
- Create: `internal/scaffold/templates/react/web/src/test/setup.ts.tmpl`
- Create: `internal/scaffold/templates/react/web/package.json.tmpl`
- Create: `internal/scaffold/templates/react/web/index.html.tmpl`
- Create: `internal/scaffold/templates/react/web/src/App.tsx.tmpl`
- Create: `internal/scaffold/templates/react/web/src/main.tsx.tmpl`

- [ ] **Step 1: Write a test for the react layer**

Add to `scaffold_test.go`:

```go
func TestRun_ReactLayer(t *testing.T) {
	t.Parallel()
	dest := t.TempDir()

	cfg := scaffold.Config{
		Name:        "testapp",
		Module:      "github.com/test/testapp",
		Description: "A test application",
		Layers:      []string{"react"},
	}

	if err := scaffold.Run(cfg, dest); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	wantFiles := []string{
		"web/package.json",
		"web/vite.config.ts",
		"web/vitest.config.ts",
		"web/tsconfig.json",
		"web/src/test/setup.ts",
	}
	for _, f := range wantFiles {
		path := filepath.Join(dest, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", f)
		}
	}
}
```

- [ ] **Step 2: Create react templates**

Extract from `skills/setup-react/skill.md`:
- `vite.config.ts` from lines 43-63
- `vitest.config.ts` from lines 70-94
- `tsconfig.json` from lines 109-130
- `setup.ts` from lines 101-102
- `package.json` — construct a minimal one with all deps from the skill (lines 29-36 for runtime deps, devDeps from lines 92-101 for lint skill). Replace `"name"` value with `{{.Name}}`

Also create minimal `index.html`, `src/App.tsx`, `src/main.tsx` so the scaffold actually builds:

`web/index.html.tmpl`:
```html
<!doctype html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>{{.Name}}</title>
  </head>
  <body>
    <div id="root"></div>
    <script type="module" src="/src/main.tsx"></script>
  </body>
</html>
```

`web/src/main.tsx.tmpl`:
```tsx
import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App'

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
)
```

`web/src/App.tsx.tmpl`:
```tsx
function App() {
  return (
    <div>
      <h1>{{.Name}}</h1>
      <p>{{.Description}}</p>
    </div>
  )
}

export default App
```

`web/tsconfig.node.json.tmpl`:
```json
{
  "compilerOptions": {
    "composite": true,
    "module": "ESNext",
    "moduleResolution": "bundler",
    "allowSyntheticDefaultImports": true
  },
  "include": ["vite.config.ts"]
}
```

- [ ] **Step 3: Run test**

Run: `cd /Users/petre/personal/forge && go test ./internal/scaffold/ -v -run TestRun_ReactLayer -count=1`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/scaffold/templates/react/
git commit -m "feat: add react layer templates"
```

---

### Task 5: Create makefile layer template

**Files:**
- Create: `internal/scaffold/templates/makefile/Makefile.tmpl`

- [ ] **Step 1: Create Makefile template**

Extract from `skills/setup-makefile/skill.md` lines 33-192. Replace all `<project-name>` with `{{.Name}}`.

- [ ] **Step 2: Run test**

Run: `cd /Users/petre/personal/forge && go test ./internal/scaffold/ -v -run TestRun_AllLayers -count=1`
Expected: Still fails (missing other layers) — but verify Makefile is generated

- [ ] **Step 3: Commit**

```bash
git add internal/scaffold/templates/makefile/
git commit -m "feat: add makefile layer template"
```

---

### Task 6: Create linting layer templates

**Files:**
- Create: `internal/scaffold/templates/linting/.golangci.yml.tmpl`
- Create: `internal/scaffold/templates/linting/web/eslint.config.js.tmpl`

- [ ] **Step 1: Create templates**

Extract `.golangci.yml` from `skills/setup-linting/skill.md` lines 17-87. No template variables.

Extract `web/eslint.config.js` from `skills/setup-linting/skill.md` lines 107-146. No template variables.

- [ ] **Step 2: Commit**

```bash
git add internal/scaffold/templates/linting/
git commit -m "feat: add linting layer templates"
```

---

### Task 7: Create bdd layer templates

**Files:**
- Create: `internal/scaffold/templates/bdd/features/health.feature.tmpl`
- Create: `internal/scaffold/templates/bdd/internal/api/health_bdd_test.go.tmpl`
- Create: `internal/scaffold/templates/bdd/web/features/navigation.feature.tmpl`
- Create: `internal/scaffold/templates/bdd/web/e2e/steps/navigation.steps.ts.tmpl`

- [ ] **Step 1: Create templates**

Extract from `skills/setup-bdd/skill.md`:
- `health.feature` from lines 25-35. No template variables.
- `health_bdd_test.go` from lines 40-141. Replace `github.com/petrepopescu21/myproject` → nothing (this file doesn't import the project module; it uses `httptest` directly). However, the `context` import is needed — ensure it's included.
- `navigation.feature` from lines 163-178. No template variables.
- `navigation.steps.ts` from lines 182-207. No template variables.

**Important:** The Go template file contains `{{` characters in Go string interpolation like `fmt.Errorf`. To prevent the template engine from parsing these, use `{{"{{"}}`  or raw string blocks, OR better — since these files have no actual template variables, we can avoid the `.tmpl` suffix and copy them verbatim. **Decision:** For files with no template variables that contain `{{`, don't use `.tmpl` suffix. Update the scaffold engine to handle non-`.tmpl` files as raw copies.

Actually, the simpler approach: update `scaffold.go` to only process `*.tmpl` files through the template engine. Files without `.tmpl` are copied verbatim. This handles the Go test file cleanly.

Update `scaffold.go` — in the `WalkDir` callback:

```go
// If not a .tmpl file, copy verbatim
if !strings.HasSuffix(path, ".tmpl") {
    return copyFile(data, outPath)
}

// Strip .tmpl suffix for template files
rel = strings.TrimSuffix(rel, ".tmpl")
```

- [ ] **Step 2: Update scaffold.go to support raw (non-tmpl) files**

In `scaffold.go`, split the logic: `.tmpl` files go through template engine, other files are raw-copied.

```go
// After reading data and computing outPath:

if strings.HasSuffix(path, ".tmpl") {
    rel = strings.TrimSuffix(rel, ".tmpl")
    outPath = filepath.Join(destDir, rel)
    // ... template parse + execute (existing code)
} else {
    // Raw copy
    if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
        return err
    }
    return os.WriteFile(outPath, data, 0o644)
}
```

- [ ] **Step 3: Commit**

```bash
git add internal/scaffold/templates/bdd/ internal/scaffold/scaffold.go
git commit -m "feat: add bdd layer templates, support raw file copies"
```

---

### Task 8: Create playwright layer templates

**Files:**
- Create: `internal/scaffold/templates/playwright/web/playwright.config.ts.tmpl`
- Create: `internal/scaffold/templates/playwright/web/playwright-integration.config.ts.tmpl`
- Create: `internal/scaffold/templates/playwright/web/e2e/fixtures.ts.tmpl`
- Create: `internal/scaffold/templates/playwright/web/e2e/navigation.spec.ts.tmpl`
- Create: `internal/scaffold/templates/playwright/scripts/e2e-web.sh.tmpl`

- [ ] **Step 1: Create templates**

Extract from `skills/setup-playwright/skill.md`:
- `playwright.config.ts` from lines 26-51. No template variables.
- `playwright-integration.config.ts` from lines 58-80. No template variables.
- `fixtures.ts` from lines 88-116. No template variables.
- `navigation.spec.ts` from lines 131-151. No template variables.
- `e2e-web.sh` from lines 157-217. Replace `pebblr` references with `{{.Name}}`.

Since most of these have no template variables but might contain `${}` or other syntax that conflicts with Go templates — use raw files (no `.tmpl` suffix) where there are no template variables.

- [ ] **Step 2: Commit**

```bash
git add internal/scaffold/templates/playwright/
git commit -m "feat: add playwright layer templates"
```

---

### Task 9: Create ci layer templates (with Renovate)

**Files:**
- Create: `internal/scaffold/templates/ci/.github/workflows/ci.yml.tmpl`
- Create: `internal/scaffold/templates/ci/.github/workflows/deploy.yml.tmpl`
- Create: `internal/scaffold/templates/ci/renovate.json.tmpl`

- [ ] **Step 1: Create ci.yml template**

Extract from `skills/setup-ci/skill.md` lines 54-186. Replace `pebblr` with `{{.Name}}`.

- [ ] **Step 2: Create deploy.yml template**

Extract from `skills/setup-ci/skill.md` lines 190-280. Replace `pebblr` with `{{.Name}}`.

- [ ] **Step 3: Create renovate.json**

This replaces the old `dependabot-auto-merge.yml`. Renovate handles Go modules, npm, GitHub Actions, AND custom regex for Makefile tool versions.

```json
{
  "$schema": "https://docs.renovatebot.com/renovate-schema.json",
  "extends": [
    "config:recommended",
    ":automergeMinor",
    ":automergeDigest",
    "group:allNonMajor"
  ],
  "packageRules": [
    {
      "matchManagers": ["gomod", "npm", "github-actions"],
      "automerge": true,
      "automergeType": "pr",
      "matchUpdateTypes": ["minor", "patch", "digest"]
    }
  ],
  "customManagers": [
    {
      "customType": "regex",
      "fileMatch": ["^Makefile$"],
      "matchStrings": [
        "KIND_VERSION\\s*:=\\s*(?<currentValue>v[\\d.]+)\\s*#\\s*renovate:\\s*datasource=(?<datasource>[^\\s]+)\\s+depName=(?<depName>[^\\s]+)",
        "TILT_VERSION\\s*:=\\s*(?<currentValue>[\\d.]+)\\s*#\\s*renovate:\\s*datasource=(?<datasource>[^\\s]+)\\s+depName=(?<depName>[^\\s]+)",
        "HELM_VERSION\\s*:=\\s*(?<currentValue>v[\\d.]+)\\s*#\\s*renovate:\\s*datasource=(?<datasource>[^\\s]+)\\s+depName=(?<depName>[^\\s]+)",
        "CLOUD_PROVIDER_KIND_VERSION\\s*:=\\s*(?<currentValue>v[\\d.]+)\\s*#\\s*renovate:\\s*datasource=(?<datasource>[^\\s]+)\\s+depName=(?<depName>[^\\s]+)",
        "GOLANGCI_LINT_VERSION\\s*:=\\s*(?<currentValue>v[\\d.]+)\\s*#\\s*renovate:\\s*datasource=(?<datasource>[^\\s]+)\\s+depName=(?<depName>[^\\s]+)"
      ]
    }
  ]
}
```

**Important:** The Makefile template needs Renovate-compatible comments on each version pin. Update the Makefile template (from Task 5) to include datasource hints:

```makefile
KIND_VERSION           := v0.27.0   # renovate: datasource=github-releases depName=kubernetes-sigs/kind
TILT_VERSION           := 0.33.22   # renovate: datasource=github-releases depName=tilt-dev/tilt
HELM_VERSION           := v3.17.3   # renovate: datasource=github-releases depName=helm/helm
CLOUD_PROVIDER_KIND_VERSION := v0.6.0 # renovate: datasource=github-releases depName=kubernetes-sigs/cloud-provider-kind
GOLANGCI_LINT_VERSION  := v2.1.6    # renovate: datasource=github-releases depName=golangci/golangci-lint
```

- [ ] **Step 4: Commit**

```bash
git add internal/scaffold/templates/ci/ internal/scaffold/templates/makefile/Makefile.tmpl
git commit -m "feat: add ci layer templates with Renovate for Makefile deps"
```

---

### Task 10: Create sonar layer template

**Files:**
- Create: `internal/scaffold/templates/sonar/sonar-project.properties.tmpl`

- [ ] **Step 1: Create template**

Extract from `skills/setup-sonar/skill.md` lines 38-58. The sonar org and project key are not known at scaffold time. Use `{{.Name}}` as the project key placeholder. The user will need to customize after generation.

```properties
# SonarCloud Project Configuration
# TODO: Update organization and project key
sonar.projectKey={{.Name}}
sonar.organization=FIXME
sonar.projectName={{.Name}}
sonar.projectVersion=1.0

sonar.sources=cmd,internal,web/src
sonar.tests=internal,web/src
sonar.test.inclusions=**/*_test.go,**/*.test.ts,**/*.test.tsx,**/*.spec.ts,**/*.spec.tsx

sonar.exclusions=**/node_modules/**,**/dist/**,**/vendor/**,**/*.gen.go

sonar.go.coverage.reportPath=coverage.out
sonar.typescript.lcov.reportPaths=web/coverage/lcov.info

sonar.sourceEncoding=UTF-8
```

- [ ] **Step 2: Commit**

```bash
git add internal/scaffold/templates/sonar/
git commit -m "feat: add sonar layer template"
```

---

### Task 11: Create helm layer templates

**Files:**
- Create: `internal/scaffold/templates/helm/deploy/helm/%Name%/Chart.yaml.tmpl`
- Create: `internal/scaffold/templates/helm/deploy/helm/%Name%/values.yaml.tmpl`
- Create: `internal/scaffold/templates/helm/deploy/helm/%Name%/values-aks.yaml.tmpl`
- Create: `internal/scaffold/templates/helm/deploy/kind/kind-config.yaml.tmpl`
- Create: `internal/scaffold/templates/helm/Tiltfile.tmpl`
- Create: `internal/scaffold/templates/helm/scripts/cluster-db.sh.tmpl`
- Create: `internal/scaffold/templates/helm/scripts/cluster-deps.sh` (raw — no template vars, has `${}` syntax)

- [ ] **Step 1: Create templates**

Extract from `skills/setup-helm/skill.md`:
- `values.yaml` from lines 18-93. Replace `pebblr` → `{{.Name}}`.
- `values-aks.yaml` from lines 98-151. Replace `pebblr` → `{{.Name}}`.
- `kind-config.yaml` from lines 157-176. Replace `pebblr` → `{{.Name}}`.
- `Tiltfile` from lines 360-388. Replace `<project-name>` → `{{.Name}}`.
- `cluster-db.sh` from lines 185-352. Replace `pebblr` → `{{.Name}}`. **Careful:** This file uses bash `${}` syntax heavily. Since Go templates use `{{}}` not `${}`, the bash syntax is safe. But any lines with `{{` would conflict — check for this. There are none in this file.
- `cluster-deps.sh` from lines 397-454. No template variables, uses `${}` — copy as raw file.
- `Chart.yaml` — create new:

```yaml
apiVersion: v2
name: {{.Name}}
description: {{.Description}}
type: application
version: 0.1.0
appVersion: "0.1.0"
```

Also need to include the standard Helm template files (`_helpers.tpl`, `deployment.yaml`, `service.yaml`, etc.). Since `helm create` generates these and the skill says "run `helm create`", we have two options:
1. Embed the standard Helm templates
2. Have the CLI run `helm create` then overlay values

**Decision:** The CLI should just emit values.yaml, values-aks.yaml, and Chart.yaml. The standard templates (`_helpers.tpl`, `deployment.yaml`, `service.yaml`, `serviceaccount.yaml`, `hpa.yaml`, `ingress.yaml`) are generated by `helm create`. The scaffold engine should call `helm create` as a subprocess for this layer, then overlay our custom files. But that introduces a runtime dependency on helm...

**Simpler decision:** Embed the standard Helm template files too. They're boilerplate and rarely change. Extract them from what `helm create` produces. This keeps the CLI dependency-free.

Create these additional files under `internal/scaffold/templates/helm/deploy/helm/%Name%/templates/`:
- `_helpers.tpl.tmpl`
- `deployment.yaml` (raw — uses Helm `{{` syntax, NOT Go templates)
- `service.yaml` (raw)
- `serviceaccount.yaml` (raw)
- `hpa.yaml` (raw)
- `ingress.yaml` (raw)
- `NOTES.txt` (raw)

**Critical:** Helm template files use `{{ .Values.x }}` syntax which collides with Go's `text/template`. These MUST be raw files (no `.tmpl` suffix) so the scaffold engine copies them verbatim.

- [ ] **Step 2: Commit**

```bash
git add internal/scaffold/templates/helm/
git commit -m "feat: add helm layer templates"
```

---

### Task 12: Create claude-md layer template

**Files:**
- Create: `internal/scaffold/templates/claude-md/CLAUDE.md.tmpl`

- [ ] **Step 1: Create template**

Extract the template from `skills/generate-claude-md/skill.md` lines 50-291. Replace all `[PLACEHOLDERS]`:
- `[PROJECT_NAME]` → `{{.Name}}`
- `[PROJECT_DESCRIPTION]` → `{{.Description}}`

Remove all `[Include only if X]` conditional comments — the template should include all sections. The user can prune after generation, or we can use Go template `{{if}}` conditionals based on which layers are selected.

**Decision:** Use Go template conditionals. Add a helper method or fields to Config:

```go
// Add to Config:
func (c Config) HasLayer(name string) bool {
    for _, l := range c.Layers {
        if l == name {
            return true
        }
    }
    return false
}
```

Then in the CLAUDE.md template:
```
{{if .HasLayer "go-module"}}
### Backend
...
{{end}}
```

- [ ] **Step 2: Commit**

```bash
git add internal/scaffold/templates/claude-md/ internal/scaffold/scaffold.go
git commit -m "feat: add claude-md layer template with conditional sections"
```

---

### Task 13: Run all scaffold tests green

- [ ] **Step 1: Run the full test suite**

Run: `cd /Users/petre/personal/forge && go test ./internal/scaffold/ -v -count=1`
Expected: ALL PASS

- [ ] **Step 2: Fix any failures**

Address any template syntax issues, missing files, or path problems.

- [ ] **Step 3: Commit fixes if needed**

```bash
git add internal/scaffold/
git commit -m "fix: resolve template test failures"
```

---

### Task 14: Create the CLI

**Files:**
- Create: `cmd/forge/main.go`

- [ ] **Step 1: Write a test for the CLI**

```go
// cmd/forge/main_test.go
package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCLI_Help(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "--help")
	out, err := cmd.CombinedOutput()
	if err != nil {
		// --help exits non-zero with flag package, that's OK
		_ = err
	}
	if len(out) == 0 {
		t.Fatal("expected help output, got empty")
	}
}

func TestCLI_Init(t *testing.T) {
	dest := t.TempDir()
	cmd := exec.Command("go", "run", ".",
		"init",
		"--name", "testapp",
		"--module", "github.com/test/testapp",
		"--description", "A test app",
		"--dest", dest,
		"--layers", "go-module,makefile",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI failed: %v\nOutput: %s", err, out)
	}

	if _, err := os.Stat(filepath.Join(dest, "cmd/testapp/main.go")); os.IsNotExist(err) {
		t.Error("expected cmd/testapp/main.go to be generated")
	}
	if _, err := os.Stat(filepath.Join(dest, "Makefile")); os.IsNotExist(err) {
		t.Error("expected Makefile to be generated")
	}
}
```

- [ ] **Step 2: Write the CLI**

```go
// cmd/forge/main.go
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/petrepopescu21/forge/internal/scaffold"
)

func main() {
	initCmd := flag.NewFlagSet("init", flag.ExitOnError)
	name := initCmd.String("name", "", "Project name (required)")
	module := initCmd.String("module", "", "Go module path (required)")
	desc := initCmd.String("description", "", "One-line project description")
	dest := initCmd.String("dest", ".", "Destination directory")
	layers := initCmd.String("layers", strings.Join(scaffold.AllLayers, ","), "Comma-separated layers to scaffold")

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: forge init [flags]\n")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "init":
		initCmd.Parse(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\nUsage: forge init [flags]\n", os.Args[1])
		os.Exit(1)
	}

	if *name == "" || *module == "" {
		fmt.Fprintln(os.Stderr, "Error: --name and --module are required")
		initCmd.Usage()
		os.Exit(1)
	}

	cfg := scaffold.Config{
		Name:        *name,
		Module:      *module,
		Description: *desc,
		Layers:      strings.Split(*layers, ","),
	}

	if err := scaffold.Run(cfg, *dest); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Scaffolded %s in %s\n", *name, *dest)
	fmt.Printf("Layers: %s\n", *layers)
}
```

- [ ] **Step 3: Run CLI tests**

Run: `cd /Users/petre/personal/forge && go test ./cmd/forge/ -v -count=1`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add cmd/forge/
git commit -m "feat: add forge CLI with init command"
```

---

### Task 15: Rewrite bootstrap-project skill as thin wrapper

**Files:**
- Modify: `skills/bootstrap-project/skill.md`

- [ ] **Step 1: Replace skill content**

The new skill gathers inputs interactively (same questions as before), then runs:

```bash
go run github.com/petrepopescu21/forge/cmd/forge init \
  --name "$NAME" \
  --module "$MODULE" \
  --description "$DESC" \
  --layers "$LAYERS"
```

Write the new `skills/bootstrap-project/skill.md`:

```markdown
---
name: bootstrap-project
description: Scaffold a complete Go + React/TypeScript project. Gathers project info interactively, then runs the forge CLI to generate all files deterministically. Trigger on "new project", "bootstrap project", "scaffold project", "create project", or "init project".
---

# Bootstrap Project

Bootstrapping a new project using the forge CLI.

## Prerequisites

- Go 1.22 or later (to build/run the forge CLI)
- Bun 1.1 or higher (for frontend dependencies after scaffolding)

## Process

### Step 1: Gather Project Information

Collect the following from the user:

1. **Project name** — lowercase alphanumeric and hyphens (e.g., `myproject`)
2. **Go module path** — e.g., `github.com/petrepopescu21/myproject`
3. **One-liner description** — e.g., "A self-hosted CRM for field sales"

### Step 2: Select Layers

Present the layer checklist. Default all to yes; user can deselect:

- Go backend (`go-module`)
- React + TypeScript (`react`)
- Helm + Kubernetes (`helm`)
- BDD + Gherkin (`bdd`)
- SonarCloud (`sonar`)
- Playwright E2E (`playwright`)

These are always included: `makefile`, `linting`, `ci`, `claude-md`.

Build the comma-separated layer string from selections.

### Step 3: Initialize Git Repository

```bash
git init
git commit --allow-empty -m "initial: project bootstrap"
```

### Step 4: Run the Forge CLI

Determine the path to the forge plugin repository. The plugin is installed at the path where this skill file lives — navigate up from the skill to find the repo root containing `cmd/forge/`.

```bash
FORGE_REPO="$(cd "$(dirname "$0")/../.." && pwd)"
go run "$FORGE_REPO/cmd/forge" init \
  --name "$NAME" \
  --module "$MODULE" \
  --description "$DESCRIPTION" \
  --layers "$LAYERS"
```

If the plugin path is not determinable, fall back to:

```bash
go run github.com/petrepopescu21/forge/cmd/forge@latest init \
  --name "$NAME" \
  --module "$MODULE" \
  --description "$DESCRIPTION" \
  --layers "$LAYERS"
```

### Step 5: Install Dependencies

If React layer was selected:

```bash
cd web && bun install
```

Run `go mod tidy` if Go layers were selected:

```bash
go mod tidy
```

### Step 6: Verify Quality Gates

```bash
make lint
make typecheck
make test
```

All must pass. Fix any issues before proceeding.

### Step 7: Commit

Commit the scaffolding in one commit:

```bash
git add -A
git commit -m "scaffold: initialize project with forge CLI"
```

### Step 8: Summary

Report what was generated:

- List all layers that were scaffolded
- Show `make help` output
- Remind user to:
  - Update `sonar-project.properties` with their org (if sonar layer)
  - Update `renovate.json` if they need custom rules (if ci layer)
  - Use `forge:add-feature` to begin feature development
```

- [ ] **Step 2: Commit**

```bash
git add skills/bootstrap-project/skill.md
git commit -m "feat: rewrite bootstrap-project as thin CLI wrapper"
```

---

### Task 16: Remove setup skills and quality-check

**Files:**
- Delete: `skills/setup-go-module/skill.md` (and directory)
- Delete: `skills/setup-react/skill.md` (and directory)
- Delete: `skills/setup-makefile/skill.md` (and directory)
- Delete: `skills/setup-linting/skill.md` (and directory)
- Delete: `skills/setup-bdd/skill.md` (and directory)
- Delete: `skills/setup-playwright/skill.md` (and directory)
- Delete: `skills/setup-ci/skill.md` (and directory)
- Delete: `skills/setup-sonar/skill.md` (and directory)
- Delete: `skills/setup-helm/skill.md` (and directory)
- Delete: `skills/generate-claude-md/skill.md` (and directory)
- Delete: `skills/quality-check/skill.md` (and directory)

- [ ] **Step 1: Remove skill directories**

```bash
rm -rf skills/setup-go-module skills/setup-react skills/setup-makefile \
       skills/setup-linting skills/setup-bdd skills/setup-playwright \
       skills/setup-ci skills/setup-sonar skills/setup-helm \
       skills/generate-claude-md skills/quality-check
```

- [ ] **Step 2: Commit**

```bash
git add -A
git commit -m "remove: delete 11 setup skills replaced by forge CLI"
```

---

### Task 17: Update plugin.json and dependencies.yaml

**Files:**
- Modify: `plugin.json`
- Modify: `skills/dependencies.yaml`

- [ ] **Step 1: Update plugin.json**

```json
{
  "name": "forge",
  "version": "0.2.0",
  "description": "Opinionated development workflows for Go + React/TypeScript projects — BDD, TDD, CI/CD, quality gates. Includes a CLI scaffold generator and AI-powered feature development skills.",
  "skills": [
    "skills/bootstrap-project",
    "skills/add-feature",
    "skills/bdd-feature",
    "skills/tdd-cycle"
  ]
}
```

- [ ] **Step 2: Update dependencies.yaml**

```yaml
# Skill dependency graph — source of truth.
# The setup-* skills have been replaced by the forge CLI (cmd/forge/).
# Only AI-powered workflow skills remain.

skills:
  # Orchestrators
  bootstrap-project:
    type: orchestrator
    depends: []

  add-feature:
    type: orchestrator
    depends:
      - bdd-feature
      - tdd-cycle

  # Workflow skills
  bdd-feature:
    type: workflow
    depends: []
  tdd-cycle:
    type: workflow
    depends: []
```

- [ ] **Step 3: Commit**

```bash
git add plugin.json skills/dependencies.yaml
git commit -m "feat: update plugin.json and dependencies for 4-skill structure"
```

---

### Task 18: Update tests

**Files:**
- Modify: `tests/lint_test.go` (minimal changes — it validates structure, should work with 4 skills)
- Delete: `tests/skills_test.go` (behavioral tests for removed skills)
- Delete: `tests/sdk_helpers_test.go` (SDK test harness no longer needed)
- Modify: `tests/cmd/affected/main.go` (update test name mapping)

- [ ] **Step 1: Remove behavioral test files**

```bash
rm tests/skills_test.go tests/sdk_helpers_test.go
```

- [ ] **Step 2: Verify lint tests still pass**

Run: `cd /Users/petre/personal/forge && go test ./tests/ -run 'TestFrontmatter|TestPluginJson|TestDependency|TestNoOrphan|TestNoCyclic' -v`
Expected: PASS (structural lint works with 4 skills)

- [ ] **Step 3: Remove anthropic SDK dependency**

```bash
cd /Users/petre/personal/forge && go mod tidy
```

This should remove `github.com/anthropics/anthropic-sdk-go` since nothing imports it anymore.

- [ ] **Step 4: Commit**

```bash
git add tests/ go.mod go.sum
git commit -m "feat: remove behavioral tests, keep structural lint for 4 skills"
```

---

### Task 19: Update CI workflows

**Files:**
- Modify: `.github/workflows/ci.yml`
- Modify or delete: `.github/workflows/test-skills.yml`

- [ ] **Step 1: Update ci.yml to also run scaffold tests**

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
          go-version-file: go.mod

      - name: Run structural lint tests
        run: go test ./tests/ -run 'TestFrontmatter|TestPluginJson|TestDependency|TestNoOrphan|TestNoCyclic' -v

  scaffold-tests:
    name: Scaffold Tests
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - name: Run scaffold tests
        run: go test ./internal/scaffold/ -v -count=1

      - name: Run CLI tests
        run: go test ./cmd/forge/ -v -count=1
```

- [ ] **Step 2: Remove test-skills.yml**

The behavioral test workflow is no longer needed — all setup skills are gone and testing is done with `go test`.

```bash
rm .github/workflows/test-skills.yml
```

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/
git commit -m "ci: add scaffold tests, remove behavioral test workflow"
```

---

### Task 20: Update CLAUDE.md

**Files:**
- Modify: `CLAUDE.md`

- [ ] **Step 1: Update CLAUDE.md to reflect new structure**

Key changes:
- Document the CLI (`cmd/forge/`) and template structure
- Update skill list (4 skills, not 15)
- Replace quality-check references with `make lint && make typecheck && make test`
- Update testing workflow (scaffold tests replace behavioral tests)
- Remove Level 2 behavioral test instructions (no more SDK tests)
- Update Level 3 to use CLI directly instead of skill invocation

- [ ] **Step 2: Commit**

```bash
git add CLAUDE.md
git commit -m "docs: update CLAUDE.md for CLI-based scaffolding"
```

---

### Task 21: Clean up stale files

- [ ] **Step 1: Remove the `affected` binary if present**

```bash
rm -f affected
```

- [ ] **Step 2: Remove stale plan/spec docs that reference old structure**

Check `docs/` for anything that should be updated or removed.

- [ ] **Step 3: Final verification**

Run: `cd /Users/petre/personal/forge && go test ./... -v -count=1`
Expected: ALL PASS (structural lint + scaffold tests + CLI tests)

- [ ] **Step 4: Commit any cleanup**

```bash
git add -A
git commit -m "chore: clean up stale artifacts"
```
