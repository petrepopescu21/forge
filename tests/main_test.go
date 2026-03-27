package tests

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
)

var fixtureDir string

func TestMain(m *testing.M) {
	if os.Getenv("ANTHROPIC_API_KEY") != "" || os.Getenv("ANTHROPIC_AUTH_TOKEN") != "" {
		dir, err := os.MkdirTemp("", "forge-fixture-*")
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to create fixture dir: %v\n", err)
			os.Exit(1)
		}
		defer os.RemoveAll(dir)

		cmd := exec.Command("claude", "--print", "--allowedTools", "*", "--",
			"use forge:bootstrap-project with name testapp, module github.com/test/testapp, description 'a test project', all layers yes")
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Fprintf(os.Stderr, "bootstrap fixture failed: %v\nOutput:\n%s\n", err, string(out))
		} else {
			fixtureDir = dir
		}
	}

	os.Exit(m.Run())
}
