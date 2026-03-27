// Package scaffold provides a deterministic project scaffolding engine
// that replaces AI-driven skill execution with embedded Go templates.
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

//go:embed all:templates
var templates embed.FS

// AllLayers lists every valid layer name.
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

// Config holds the project scaffolding configuration.
type Config struct {
	Name        string
	Module      string
	Description string
	Layers      []string
}

// HasLayer reports whether the named layer is selected.
func (c Config) HasLayer(name string) bool {
	for _, l := range c.Layers {
		if l == name {
			return true
		}
	}
	return false
}

// Run renders the common layer plus every selected layer into destDir.
func Run(cfg Config, destDir string) error {
	// Validate layer names.
	valid := make(map[string]bool, len(AllLayers))
	for _, l := range AllLayers {
		valid[l] = true
	}
	for _, l := range cfg.Layers {
		if !valid[l] {
			return fmt.Errorf("unknown layer: %q", l)
		}
	}

	// Always render common layer.
	if err := renderLayer(cfg, destDir, "common"); err != nil {
		return fmt.Errorf("rendering common layer: %w", err)
	}

	// Render each selected layer.
	for _, layer := range cfg.Layers {
		if err := renderLayer(cfg, destDir, layer); err != nil {
			return fmt.Errorf("rendering layer %s: %w", layer, err)
		}
	}

	return nil
}

// renderLayer walks the embedded templates/<layer>/ tree and writes files
// into destDir. Files ending in .tmpl are processed through text/template;
// others are copied verbatim. Directory paths containing %Name% have that
// token replaced with cfg.Name. Shell scripts get mode 0755.
func renderLayer(cfg Config, destDir, layer string) error {
	root := filepath.Join("templates", layer)

	return fs.WalkDir(templates, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Compute the relative path from the layer root.
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}

		// Replace %Name% token in directory and file paths.
		rel = strings.ReplaceAll(rel, "%Name%", cfg.Name)

		outPath := filepath.Join(destDir, rel)

		if d.IsDir() {
			return os.MkdirAll(outPath, 0755)
		}

		// Read embedded file.
		data, err := templates.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}

		// Ensure parent directory exists.
		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return err
		}

		if strings.HasSuffix(path, ".tmpl") {
			// Process through text/template.
			tmpl, err := template.New(filepath.Base(path)).Parse(string(data))
			if err != nil {
				return fmt.Errorf("parsing template %s: %w", path, err)
			}

			// Strip the .tmpl suffix from the output path.
			outPath = strings.TrimSuffix(outPath, ".tmpl")

			f, err := os.Create(outPath)
			if err != nil {
				return err
			}
			defer f.Close()

			if err := tmpl.Execute(f, cfg); err != nil {
				return fmt.Errorf("executing template %s: %w", path, err)
			}
		} else {
			// Copy verbatim.
			if err := os.WriteFile(outPath, data, 0644); err != nil {
				return err
			}
		}

		// Shell scripts get executable permission.
		if strings.HasSuffix(outPath, ".sh") {
			if err := os.Chmod(outPath, 0755); err != nil {
				return err
			}
		}

		return nil
	})
}
