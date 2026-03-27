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
- Behavioral tests require `ANTHROPIC_API_KEY` (API)
- Level 4 tests (CI self-validation) require `FORGE_TEST_GITHUB_TOKEN`

## Testing Workflow

After modifying skills, run the appropriate level of testing based on context.

### Level 1 — Structural Lint (always, before committing)

```bash
go test ./tests/ -run 'TestFrontmatter|TestPluginJson|TestDependency|TestNoOrphan|TestNoCyclic' -v
```

### Level 2 — Behavioral Unit Tests (SDK-based, fast)

Tests that the skill produces correct tool calls (Write/Bash) via the Anthropic API. No files are created — assertions are on intent, not side effects. Takes ~1 minute for all skills.

```bash
go test ./tests/ -run 'TestSetup|TestGenerate' -v -timeout 5m
```

To run only tests affected by your changes:

```bash
AFFECTED=$(go run ./tests/cmd/affected/ --base origin/main)
go test ./tests/ -run "$AFFECTED" -v -timeout 5m
```

### Level 3 — E2E Skill Validation (local, interactive)

Full end-to-end validation where skills are actually executed via Claude Code and the output is verified. Run this when making significant skill changes or before merging.

#### On a PR branch (comparing old vs new)

1. Identify which skills changed:
   ```bash
   go run ./tests/cmd/affected/ --base origin/main
   ```

2. For each changed skill, test the **new version** (current branch):
   ```bash
   mkdir -p /tmp/forge-e2e-new && cd /tmp/forge-e2e-new
   # Use the skill interactively — e.g. for setup-go-module:
   # "use forge:setup-go-module with project name testapp and module github.com/test/testapp"
   ```

3. Test the **old version** (main branch) for comparison:
   ```bash
   mkdir -p /tmp/forge-e2e-old && cd /tmp/forge-e2e-old
   git stash && git checkout main
   # Run the same skill prompt
   git checkout - && git stash pop
   ```

4. Compare outputs:
   ```bash
   diff -r /tmp/forge-e2e-old /tmp/forge-e2e-new
   ```

5. For orchestrator skills (`bootstrap-project`, `add-feature`), run the full workflow:
   ```bash
   mkdir -p /tmp/forge-e2e-bootstrap && cd /tmp/forge-e2e-bootstrap
   # "use forge:bootstrap-project with name testapp, module github.com/test/testapp, description 'test project', all layers yes"
   # Then verify: make lint && make typecheck && make test
   ```

#### On main branch (regression check)

Run a full bootstrap and verify quality gates:

```bash
mkdir -p /tmp/forge-e2e && cd /tmp/forge-e2e
# "use forge:bootstrap-project with name testapp, module github.com/test/testapp, description 'test project', all layers yes"
make lint
make typecheck
make test
```

Then test the add-feature workflow:

```bash
# "use forge:add-feature: add a health check endpoint at GET /healthz that returns 200 OK"
make test
```

### Level 4 — CI Self-Validation (local, requires GitHub token)

Validates that generated CI workflows actually run in GitHub Actions.

1. Bootstrap a project:
   ```bash
   mkdir -p /tmp/forge-e2e-ci && cd /tmp/forge-e2e-ci
   # "use forge:bootstrap-project with name testapp, module github.com/test/testapp, description 'test project', all layers yes"
   ```

2. Push to a temp GitHub repo:
   ```bash
   gh repo create forge-test-$(date +%s) --private --clone
   cp -r /tmp/forge-e2e-ci/* .
   git add . && git commit -m "test scaffold" && git push
   ```

3. Verify the CI workflow runs and passes:
   ```bash
   gh run list --limit 1
   gh run watch  # wait for completion
   ```

4. Clean up:
   ```bash
   gh repo delete forge-test-<timestamp> --yes
   ```
