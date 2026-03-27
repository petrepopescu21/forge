# Forge — Claude Code Superpowers Plugin

## What This Is

A Claude Code superpowers plugin that scaffolds and enforces BDD/TDD workflows for Go + React/TypeScript projects. This repo contains skill definitions (markdown), a Go test harness, and CI workflows.

## Project Structure

- `skills/` — 15 skill definitions (markdown files with YAML frontmatter)
- `skills/dependencies.yaml` — skill dependency graph (source of truth)
- `plugin.json` — plugin manifest listing all skills
- `tests/` — Go test harness (structural lint + behavioral smoke tests)
- `tests/cmd/affected/` — CLI for dependency-aware test selection
- `.github/workflows/` — CI (structural lint) + behavioral tests (on-demand)

## Mandatory Rules

### Dependency Manifest

When modifying any `skill.md` that adds or removes a `forge:<name>` reference:
- **Update `skills/dependencies.yaml`** to reflect the new dependency
- **Update `plugin.json`** if adding or removing a skill directory

When adding or removing a skill directory:
- **Update both `plugin.json` and `skills/dependencies.yaml`**

### Validation

Run structural lint before committing skill changes:

```bash
go test ./tests/ -run 'TestFrontmatter|TestPluginJson|TestDependency|TestNoOrphan|TestNoCyclic' -v
```

### Skill Structure

Every skill must have a `skill.md` with YAML frontmatter containing at minimum:
- `name` — skill identifier (matches directory name)
- `description` — trigger description for Claude Code

### Testing

- Structural tests run on every PR automatically
- Behavioral tests run via `/test-skills` PR comment (authorized users only)
- Behavioral tests require `ANTHROPIC_API_KEY` environment variable
- Level 4 tests (CI self-validation) require `FORGE_TEST_GITHUB_TOKEN`
