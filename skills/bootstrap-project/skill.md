---
name: bootstrap-project
description: Scaffold a complete Go + React/TypeScript project. Gathers project info interactively, then runs the forge CLI to generate all files deterministically. Trigger on "new project", "bootstrap project", "scaffold project", "create project", or "init project".
user_invocable: true
command: bootstrap
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

The forge CLI lives in the plugin repository. Determine the plugin path from this skill file's location, then run:

```bash
FORGE_REPO="<path-to-forge-plugin-repo>"
go run "$FORGE_REPO/cmd/forge" init \
  --name "$NAME" \
  --module "$MODULE" \
  --description "$DESCRIPTION" \
  --layers "$LAYERS"
```

### Step 5: Install Dependencies

If Go layers were selected:

```bash
go mod tidy
```

If React layer was selected:

```bash
cd web && bun install
```

### Step 6: Verify Quality Gates

```bash
make lint
make typecheck
make test
```

All must pass. Fix any issues before proceeding.

### Step 7: Commit

```bash
git add -A
git commit -m "scaffold: initialize project with forge CLI"
```

### Step 8: Summary

Report what was generated:

- List all layers that were scaffolded
- Show `make help` output
- Remind user to:
  - Update `sonar-project.properties` with their SonarCloud org (if sonar layer)
  - Install Renovate GitHub app for automated dependency updates (if ci layer)
  - Use `forge:add-feature` to begin feature development
