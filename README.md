# Forge

A Claude Code plugin for bootstrapping and developing Go + React/TypeScript projects with BDD and TDD discipline.

## Installation

Install from the Claude Code marketplace:

```bash
claude install petrepopescu21/forge
```

## Commands

| Command | Description |
|---------|-------------|
| `/bootstrap` | Scaffold a complete project — collects inputs, runs the forge CLI, installs deps, verifies quality gates |
| `/feature` | Full feature cycle: prompt → BDD scenarios → TDD implementation → quality verification |

## Skills

| Skill | Type | Purpose |
|-------|------|---------|
| `forge:bootstrap-project` | orchestrator | Gathers inputs, runs forge CLI to scaffold projects in ~50ms |
| `forge:add-feature` | orchestrator | BDD → TDD → quality gates |
| `forge:bdd-feature` | workflow | Turns prompts into Gherkin scenarios with step stubs |
| `forge:tdd-cycle` | workflow | Red/green/refactor discipline |

## How It Works

Forge has two parts:

1. **CLI scaffold generator** (`cmd/forge/`) — a deterministic Go binary that generates project files from embedded templates across 11 layers (go-module, react, makefile, linting, bdd, playwright, ci, sonar, helm, claude-md, and common)
2. **AI workflow skills** — 4 skills that drive BDD/TDD feature development

### Workflow

1. **Bootstrap:** `/bootstrap` scaffolds a complete project with all layers
2. **Develop:** `/feature` drives BDD → TDD → quality for each feature
3. **Verify:** Quality gates (`make lint && make typecheck && make test`) run automatically

The generated `CLAUDE.md` ensures all future Claude Code sessions automatically use these skills.

## Stack

- Go
- React + TypeScript / Vite / Bun
- godog + playwright-bdd + Gherkin
- Vitest + RTL + Playwright
- GitHub Actions + SonarCloud
- Kubernetes / AKS / Helm / Kind
- golangci-lint + ESLint
