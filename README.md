# Forge

A Claude Code superpowers plugin for bootstrapping and developing Go + React/TypeScript projects with BDD and TDD.

## Installation

Add to your Claude Code settings:

```json
{
  "plugins": ["github:petrepopescu21/forge"]
}
```

## Skills

### Orchestrators

| Skill | Description |
|-------|-------------|
| forge:bootstrap-project | Scaffold a complete project from scratch |
| forge:add-feature | Full feature cycle: prompt → BDD → TDD → quality check |

### Workflow

| Skill | Description |
|-------|-------------|
| forge:bdd-feature | Transform a prompt into Gherkin scenarios with step stubs |
| forge:tdd-cycle | Red/green/refactor with discipline enforcement |
| forge:quality-check | Pre-commit gate: lint, typecheck, test |

### Setup

| Skill | Description |
|-------|-------------|
| forge:setup-go-module | Go module + cmd/ + internal/ scaffold |
| forge:setup-react | Vite + React + TypeScript + Vitest |
| forge:setup-makefile | Makefile with standard targets |
| forge:setup-linting | golangci-lint + ESLint configs |
| forge:setup-bdd | godog + playwright-bdd infrastructure |
| forge:setup-playwright | Dual Playwright configs with fixtures |
| forge:setup-ci | Consolidated GitHub Actions workflows |
| forge:setup-sonar | SonarCloud with dual coverage |
| forge:setup-helm | Helm chart + Kind + dev scripts |
| forge:generate-claude-md | CLAUDE.md with skill enforcement |

## How It Works

1. Bootstrap: forge:bootstrap-project scaffolds and generates CLAUDE.md
2. Develop: forge:add-feature drives BDD → TDD → quality
3. Verify: forge:quality-check before every commit

The generated CLAUDE.md ensures all future Claude Code sessions automatically use these skills.

## Stack

- Go
- React + TypeScript / Vite / Bun
- godog + playwright-bdd + Gherkin
- Vitest + RTL + Playwright
- GitHub Actions + SonarCloud
- Kubernetes / AKS / Helm / Kind
- golangci-lint + ESLint
