package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// captureOutput redirects stdout/stderr during fn and returns what was written.
func captureStdout(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestCLI_Help(t *testing.T) {
	output := captureStdout(func() {
		err := run([]string{"--help"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "forge") {
		t.Errorf("help output should mention 'forge', got:\n%s", output)
	}
	if !strings.Contains(output, "init") {
		t.Errorf("help output should mention 'init' command, got:\n%s", output)
	}
}

func TestCLI_InitHelp(t *testing.T) {
	output := captureStdout(func() {
		// --help on the init subcommand
		err := run([]string{"init", "--help"})
		// flag.ContinueOnError returns flag.ErrHelp on --help; that's fine
		if err != nil && !strings.Contains(err.Error(), "help requested") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "--name") {
		t.Errorf("init --help should describe --name flag, got:\n%s", output)
	}
}

func TestCLI_Init(t *testing.T) {
	dest := t.TempDir()

	err := run([]string{
		"init",
		"--name", "testapp",
		"--module", "github.com/test/testapp",
		"--description", "a test application",
		"--dest", dest,
		"--layers", "go-module",
	})
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// The go-module layer creates cmd/<name>/main.go.
	mainGo := filepath.Join(dest, "cmd", "testapp", "main.go")
	if _, err := os.Stat(mainGo); os.IsNotExist(err) {
		t.Errorf("expected file %s to exist", mainGo)
	}

	// The common layer (always rendered) creates .gitignore.
	gitignore := filepath.Join(dest, ".gitignore")
	if _, err := os.Stat(gitignore); os.IsNotExist(err) {
		t.Errorf("expected file %s to exist", gitignore)
	}
}

func TestCLI_InitAllLayers(t *testing.T) {
	dest := t.TempDir()

	// Omitting --layers should default to all layers.
	err := run([]string{
		"init",
		"--name", "myapp",
		"--module", "github.com/user/myapp",
		"--dest", dest,
	})
	if err != nil {
		t.Fatalf("init with all layers failed: %v", err)
	}

	// Spot-check a few files from different layers.
	checks := []string{
		filepath.Join(dest, "cmd", "myapp", "main.go"),        // go-module
		filepath.Join(dest, "web", "package.json"),            // react
		filepath.Join(dest, "Makefile"),                       // makefile
		filepath.Join(dest, ".gitignore"),                     // common
	}
	for _, f := range checks {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist after full scaffold", f)
		}
	}
}

func TestCLI_InitMissingName(t *testing.T) {
	err := run([]string{"init", "--module", "github.com/test/app"})
	if err == nil {
		t.Fatal("expected error when --name is missing")
	}
	if !strings.Contains(err.Error(), "--name") {
		t.Errorf("error should mention --name, got: %v", err)
	}
}

func TestCLI_InitMissingModule(t *testing.T) {
	err := run([]string{"init", "--name", "myapp"})
	if err == nil {
		t.Fatal("expected error when --module is missing")
	}
	if !strings.Contains(err.Error(), "--module") {
		t.Errorf("error should mention --module, got: %v", err)
	}
}

func TestCLI_InitUnknownLayer(t *testing.T) {
	dest := t.TempDir()
	err := run([]string{
		"init",
		"--name", "myapp",
		"--module", "github.com/test/myapp",
		"--dest", dest,
		"--layers", "nonexistent-layer",
	})
	if err == nil {
		t.Fatal("expected error for unknown layer")
	}
}

func TestCLI_UnknownCommand(t *testing.T) {
	err := run([]string{"unknown-command"})
	if err == nil {
		t.Fatal("expected error for unknown command")
	}
}

func TestCLI_NoArgs(t *testing.T) {
	err := run([]string{})
	if err == nil {
		t.Fatal("expected error when no args provided")
	}
}
