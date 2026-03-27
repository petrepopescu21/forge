# Forge — Claude Code Plugin

## What This Is

A Claude Code plugin with two parts:
1. **CLI scaffold generator** (`cmd/forge/`) — deterministic Go binary that generates project files from embedded templates
2. **AI workflow skills** (`skills/`) — 4 skills for BDD/TDD feature development

## Project Structure

- `cmd/forge/` — CLI binary (`forge init --name X --module Y --layers Z`)
- `internal/scaffold/` — template engine (embed.FS + text/template)
- `internal/scaffold/templates/` — embedded templates per layer (common, go-module, react, makefile, linting, bdd, playwright, ci, sonar, helm, claude-md)
- `skills/` — 4 AI-powered skill definitions (markdown with YAML frontmatter)
- `skills/dependencies.yaml` — skill dependency graph (source of truth)
- `plugin.json` — plugin manifest listing all skills
- `tests/` — structural lint tests for skills
- `.github/workflows/` — CI (structural lint + scaffold tests)

## Skills

| Skill | Type | Purpose |
|-------|------|---------|
| `bootstrap-project` | orchestrator | Gathers inputs, runs forge CLI |
| `add-feature` | orchestrator | BDD → TDD → quality gates |
| `bdd-feature` | workflow | Turns prompts into Gherkin scenarios |
| `tdd-cycle` | workflow | Red/green/refactor discipline |

## Mandatory Rules

### Dependency Manifest

When modifying any `skill.md` that adds or removes a `forge:<name>` reference:
- **Update `skills/dependencies.yaml`** to reflect the new dependency
- **Update `plugin.json`** if adding or removing a skill directory

### Templates

When modifying templates in `internal/scaffold/templates/`:
- `.tmpl` files are processed through `text/template` — use `{{.Name}}`, `{{.Module}}`, `{{.Description}}`
- Non-`.tmpl` files are copied verbatim — use for files with `{{` syntax (Helm, GitHub Actions)
- `%Name%` in directory paths is replaced with the project name
- Run scaffold tests after changes: `go test ./internal/scaffold/ -v`

### Validation

Run before committing:

```bash
# Structural lint (skills)
go test ./tests/ -run 'TestFrontmatter|TestPluginJson|TestDependency|TestNoOrphan|TestNoCyclic' -v

# Scaffold tests (templates)
go test ./internal/scaffold/ -v -count=1

# CLI tests
go test ./cmd/forge/ -v -count=1
```

Or all at once:

```bash
go test ./... -v -count=1
```

## Testing Workflow

### Level 1 — Unit Tests (always, before committing)

```bash
go test ./... -v -count=1
```

Covers: structural lint, scaffold template rendering, CLI flag parsing.

### Level 2 — E2E Scaffold Validation (before merging)

Test that the CLI produces a working project:

```bash
mkdir -p /tmp/forge-e2e && cd /tmp/forge-e2e
go run /path/to/forge/cmd/forge init \
  --name testapp \
  --module github.com/test/testapp \
  --description "test project"
cd web && bun install && cd ..
go mod init github.com/test/testapp
make lint
make typecheck
make test
```

### Level 3 — Full Workflow Validation (major changes)

Bootstrap a project, then test the add-feature workflow:

```bash
# In a Claude Code session inside the scaffolded project:
# "use forge:add-feature: add a health check endpoint at GET /healthz that returns 200 OK"
make test
```
