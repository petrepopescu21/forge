package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testConfig() Config {
	return Config{
		Name:        "testapp",
		Module:      "github.com/test/testapp",
		Description: "A test application",
		Layers:      AllLayers,
	}
}

func TestRun_CommonLayer(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cfg := Config{
		Name:        "testapp",
		Module:      "github.com/test/testapp",
		Description: "A test application",
		Layers:      []string{},
	}

	if err := Run(cfg, dir); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	for _, f := range []string{".gitignore", "Dockerfile.backend", "Dockerfile.web", "README.md"} {
		path := filepath.Join(dir, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected %s to exist", f)
		}
	}

	// Verify README.md has project name
	data, err := os.ReadFile(filepath.Join(dir, "README.md"))
	if err != nil {
		t.Fatalf("reading README.md: %v", err)
	}
	if !strings.Contains(string(data), "testapp") {
		t.Error("README.md should contain project name")
	}
	if !strings.Contains(string(data), "A test application") {
		t.Error("README.md should contain project description")
	}

	// Verify Dockerfile.backend has project name
	data, err = os.ReadFile(filepath.Join(dir, "Dockerfile.backend"))
	if err != nil {
		t.Fatalf("reading Dockerfile.backend: %v", err)
	}
	if !strings.Contains(string(data), "./cmd/testapp") {
		t.Error("Dockerfile.backend should reference cmd/testapp")
	}

	// Verify Dockerfile.web exists with nginx stage
	data, err = os.ReadFile(filepath.Join(dir, "Dockerfile.web"))
	if err != nil {
		t.Fatalf("reading Dockerfile.web: %v", err)
	}
	if !strings.Contains(string(data), "nginx") {
		t.Error("Dockerfile.web should use nginx")
	}
}

func TestRun_GoModuleLayer(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cfg := Config{
		Name:        "testapp",
		Module:      "github.com/test/testapp",
		Description: "A test application",
		Layers:      []string{"go-module"},
	}

	if err := Run(cfg, dir); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// Check main.go exists with correct import
	mainPath := filepath.Join(dir, "cmd", "testapp", "main.go")
	data, err := os.ReadFile(mainPath)
	if err != nil {
		t.Fatalf("reading main.go: %v", err)
	}
	if !strings.Contains(string(data), "github.com/test/testapp/internal/api") {
		t.Error("main.go should import github.com/test/testapp/internal/api")
	}

	// Check router.go exists
	routerPath := filepath.Join(dir, "internal", "api", "router.go")
	if _, err := os.Stat(routerPath); os.IsNotExist(err) {
		t.Error("expected internal/api/router.go to exist")
	}

	// Check .gitkeep files
	for _, d := range []string{"internal/domain/.gitkeep", "internal/store/.gitkeep"} {
		if _, err := os.Stat(filepath.Join(dir, d)); os.IsNotExist(err) {
			t.Errorf("expected %s to exist", d)
		}
	}
}

func TestRun_ReactLayer(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cfg := Config{
		Name:        "testapp",
		Module:      "github.com/test/testapp",
		Description: "A test application",
		Layers:      []string{"react"},
	}

	if err := Run(cfg, dir); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	files := []string{
		"web/package.json",
		"web/vite.config.ts",
		"web/vitest.config.ts",
		"web/tsconfig.json",
		"web/tsconfig.node.json",
		"web/index.html",
		"web/src/main.tsx",
		"web/src/App.tsx",
		"web/src/test/setup.ts",
	}
	for _, f := range files {
		if _, err := os.Stat(filepath.Join(dir, f)); os.IsNotExist(err) {
			t.Errorf("expected %s to exist", f)
		}
	}

	// Verify package.json has project name
	data, err := os.ReadFile(filepath.Join(dir, "web", "package.json"))
	if err != nil {
		t.Fatalf("reading package.json: %v", err)
	}
	if !strings.Contains(string(data), `"name": "testapp"`) {
		t.Error("package.json should contain project name")
	}
}

func TestRun_AllLayers(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cfg := testConfig()

	if err := Run(cfg, dir); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// Representative file from each layer
	representative := map[string]string{
		"common":     ".gitignore",
		"go-module":  "cmd/testapp/main.go",
		"react":      "web/package.json",
		"makefile":   "Makefile",
		"linting":    ".golangci.yml",
		"bdd":        "features/health.feature",
		"playwright": "web/playwright.config.ts",
		"ci":         ".github/workflows/ci.yml",
		"sonar":      "sonar-project.properties",
		"helm":       "deploy/helm/testapp/Chart.yaml",
		"claude-md":  "CLAUDE.md",
	}

	for layer, file := range representative {
		path := filepath.Join(dir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("layer %s: expected %s to exist", layer, file)
		}
	}
}

func TestRun_NoUnresolvedTemplates(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cfg := testConfig()

	if err := Run(cfg, dir); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// Walk all generated files and check none contain unresolved Go template syntax.
	// Exceptions: Helm template files which legitimately use {{ }} syntax,
	// and NOTES.txt which also uses Helm syntax.
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		// Skip Helm template files — they legitimately contain {{ }} syntax.
		rel, _ := filepath.Rel(dir, path)
		if strings.Contains(rel, filepath.Join("deploy", "helm", "testapp", "templates")) {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		content := string(data)

		// Check for unresolved Go template directives like {{.Name}} or {{if ...}}
		// We look for {{ followed by a dot or keyword, which indicates unresolved templates.
		// GitHub Actions ${{ }} syntax uses "{{" but always followed by a space, which is fine.
		// We specifically check for Go template patterns.
		if strings.Contains(content, "{{.") {
			t.Errorf("file %s contains unresolved template directive '{{.'", rel)
		}
		if strings.Contains(content, "{{if") {
			t.Errorf("file %s contains unresolved template directive '{{if'", rel)
		}
		if strings.Contains(content, "{{end") {
			t.Errorf("file %s contains unresolved template directive '{{end'", rel)
		}
		if strings.Contains(content, "{{range") {
			t.Errorf("file %s contains unresolved template directive '{{range'", rel)
		}

		return nil
	})
	if err != nil {
		t.Fatalf("walking output: %v", err)
	}
}

func TestRun_SelectiveLayers(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cfg := Config{
		Name:        "testapp",
		Module:      "github.com/test/testapp",
		Description: "A test application",
		Layers:      []string{"go-module", "makefile"},
	}

	if err := Run(cfg, dir); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// Should exist (common + selected layers)
	shouldExist := []string{
		".gitignore",
		"Dockerfile.backend",
		"Dockerfile.web",
		"cmd/testapp/main.go",
		"Makefile",
	}
	for _, f := range shouldExist {
		if _, err := os.Stat(filepath.Join(dir, f)); os.IsNotExist(err) {
			t.Errorf("expected %s to exist", f)
		}
	}

	// Should NOT exist (unselected layers)
	shouldNotExist := []string{
		"web/package.json",
		".golangci.yml",
		"features/health.feature",
		"web/playwright.config.ts",
		".github/workflows/ci.yml",
		"sonar-project.properties",
		"deploy/helm/testapp/Chart.yaml",
		"CLAUDE.md",
	}
	for _, f := range shouldNotExist {
		if _, err := os.Stat(filepath.Join(dir, f)); err == nil {
			t.Errorf("expected %s to NOT exist (layer not selected)", f)
		}
	}
}

func TestRun_InvalidLayer(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cfg := Config{
		Name:        "testapp",
		Module:      "github.com/test/testapp",
		Description: "A test application",
		Layers:      []string{"nonexistent-layer"},
	}

	err := Run(cfg, dir)
	if err == nil {
		t.Fatal("expected error for unknown layer")
	}
	if !strings.Contains(err.Error(), "unknown layer") {
		t.Errorf("error should mention 'unknown layer', got: %v", err)
	}
}

func TestRun_MakefileHasRenovateComments(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cfg := Config{
		Name:        "testapp",
		Module:      "github.com/test/testapp",
		Description: "A test application",
		Layers:      []string{"makefile"},
	}

	if err := Run(cfg, dir); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "Makefile"))
	if err != nil {
		t.Fatalf("reading Makefile: %v", err)
	}
	content := string(data)

	renovateAnnotations := []string{
		"# renovate: datasource=github-releases depName=kubernetes-sigs/kind",
		"# renovate: datasource=github-releases depName=tilt-dev/tilt",
		"# renovate: datasource=github-releases depName=helm/helm",
		"# renovate: datasource=github-releases depName=kubernetes-sigs/cloud-provider-kind",
		"# renovate: datasource=github-releases depName=golangci/golangci-lint",
	}

	for _, ann := range renovateAnnotations {
		if !strings.Contains(content, ann) {
			t.Errorf("Makefile should contain Renovate annotation: %s", ann)
		}
	}
}

func TestHasLayer(t *testing.T) {
	t.Parallel()
	cfg := Config{
		Layers: []string{"go-module", "react"},
	}

	if !cfg.HasLayer("go-module") {
		t.Error("HasLayer should return true for go-module")
	}
	if !cfg.HasLayer("react") {
		t.Error("HasLayer should return true for react")
	}
	if cfg.HasLayer("helm") {
		t.Error("HasLayer should return false for helm")
	}
}

func TestRun_ShellScriptsExecutable(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cfg := Config{
		Name:        "testapp",
		Module:      "github.com/test/testapp",
		Description: "A test application",
		Layers:      []string{"playwright", "helm"},
	}

	if err := Run(cfg, dir); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	scripts := []string{
		"scripts/e2e-web.sh",
		"scripts/cluster-db.sh",
		"scripts/cluster-deps.sh",
	}

	for _, s := range scripts {
		path := filepath.Join(dir, s)
		info, err := os.Stat(path)
		if os.IsNotExist(err) {
			t.Errorf("expected %s to exist", s)
			continue
		}
		if err != nil {
			t.Errorf("stat %s: %v", s, err)
			continue
		}
		if info.Mode()&0111 == 0 {
			t.Errorf("expected %s to be executable, got mode %v", s, info.Mode())
		}
	}
}

func TestRun_HelmTemplatesRawCopy(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cfg := Config{
		Name:        "testapp",
		Module:      "github.com/test/testapp",
		Description: "A test application",
		Layers:      []string{"helm"},
	}

	if err := Run(cfg, dir); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// Verify Helm template files exist and contain raw {{ }} syntax
	helmTemplates := []string{
		"deploy/helm/testapp/templates/deployment.yaml",
		"deploy/helm/testapp/templates/deployment-web.yaml",
		"deploy/helm/testapp/templates/service.yaml",
		"deploy/helm/testapp/templates/service-web.yaml",
		"deploy/helm/testapp/templates/gateway.yaml",
		"deploy/helm/testapp/templates/virtualservice.yaml",
		"deploy/helm/testapp/templates/serviceaccount.yaml",
		"deploy/helm/testapp/templates/_helpers.tpl",
		"deploy/helm/testapp/templates/NOTES.txt",
	}

	for _, f := range helmTemplates {
		path := filepath.Join(dir, f)
		data, err := os.ReadFile(path)
		if os.IsNotExist(err) {
			t.Errorf("expected %s to exist", f)
			continue
		}
		if err != nil {
			t.Errorf("reading %s: %v", f, err)
			continue
		}
		// These files should contain Helm template syntax
		if !strings.Contains(string(data), "{{") {
			t.Errorf("%s should contain Helm template syntax ({{ }})", f)
		}
	}
}

func TestRun_ClaudeMdConditionalSections(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		layers     []string
		shouldHave []string
		shouldNot  []string
	}{
		{
			name:       "all layers",
			layers:     []string{"claude-md", "go-module", "react", "bdd", "helm"},
			shouldHave: []string{"### Backend", "### Frontend", "### Testing", "### Infrastructure", "### Go", "### TypeScript"},
			shouldNot:  []string{},
		},
		{
			name:       "go only",
			layers:     []string{"claude-md", "go-module"},
			shouldHave: []string{"### Backend", "### Go"},
			shouldNot:  []string{"### Frontend", "### Testing", "### Infrastructure", "### TypeScript"},
		},
		{
			name:       "react only",
			layers:     []string{"claude-md", "react"},
			shouldHave: []string{"### Frontend", "### TypeScript"},
			shouldNot:  []string{"### Backend", "### Testing", "### Infrastructure", "### Go"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			cfg := Config{
				Name:        "testapp",
				Module:      "github.com/test/testapp",
				Description: "A test application",
				Layers:      tt.layers,
			}

			if err := Run(cfg, dir); err != nil {
				t.Fatalf("Run() error: %v", err)
			}

			data, err := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
			if err != nil {
				t.Fatalf("reading CLAUDE.md: %v", err)
			}
			content := string(data)

			for _, s := range tt.shouldHave {
				if !strings.Contains(content, s) {
					t.Errorf("CLAUDE.md should contain %q", s)
				}
			}
			for _, s := range tt.shouldNot {
				if strings.Contains(content, s) {
					t.Errorf("CLAUDE.md should NOT contain %q", s)
				}
			}
		})
	}
}
