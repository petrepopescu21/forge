// Command forge provides a CLI for scaffolding Go + React/TypeScript projects.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/petrepopescu21/forge/internal/scaffold"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("expected a command (e.g. init). Run with --help for usage")
	}

	switch args[0] {
	case "init":
		return runInit(args[1:])
	case "--help", "-h", "help":
		printUsage()
		return nil
	default:
		return fmt.Errorf("unknown command %q. Run with --help for usage", args[0])
	}
}

func printUsage() {
	fmt.Fprintf(os.Stdout, `forge — project scaffolding CLI

Usage:
  forge init --name <name> --module <module> [flags]

Commands:
  init    Scaffold a new project

Flags (init):
  --name         Project name (required)
  --module       Go module path (required)
  --description  One-liner description (optional)
  --dest         Output directory (default ".")
  --layers       Comma-separated layer names (default: all layers)
                 Available: %s
`, strings.Join(scaffold.AllLayers, ", "))
}

func runInit(args []string) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)

	var (
		name        string
		module      string
		description string
		dest        string
		layers      string
	)

	fs.StringVar(&name, "name", "", "project name (required)")
	fs.StringVar(&module, "module", "", "Go module path (required)")
	fs.StringVar(&description, "description", "", "one-liner description (optional)")
	fs.StringVar(&dest, "dest", ".", "output directory (default \".\")")
	fs.StringVar(&layers, "layers", "", "comma-separated layer names (default: all layers)")

	fs.Usage = func() {
		fmt.Fprintf(os.Stdout, `Usage: forge init --name <name> --module <module> [flags]

Flags:
`)
		fs.PrintDefaults()
		fmt.Fprintf(os.Stdout, "\nAvailable layers: %s\n", strings.Join(scaffold.AllLayers, ", "))
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	if name == "" {
		return fmt.Errorf("--name is required")
	}
	if module == "" {
		return fmt.Errorf("--module is required")
	}

	var selectedLayers []string
	if layers == "" {
		selectedLayers = scaffold.AllLayers
	} else {
		for _, l := range strings.Split(layers, ",") {
			l = strings.TrimSpace(l)
			if l != "" {
				selectedLayers = append(selectedLayers, l)
			}
		}
	}

	cfg := scaffold.Config{
		Name:        name,
		Module:      module,
		Description: description,
		Layers:      selectedLayers,
	}

	if err := scaffold.Run(cfg, dest); err != nil {
		return fmt.Errorf("scaffolding failed: %w", err)
	}

	fmt.Printf("Project %q scaffolded successfully in %q\n", name, dest)
	fmt.Printf("Layers generated: %s\n", strings.Join(selectedLayers, ", "))
	return nil
}
