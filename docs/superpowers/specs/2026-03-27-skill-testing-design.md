# Forge Skill Testing & Validation Design

**Date:** 2026-03-27
**Status:** Approved

## Summary

Add a dependency-aware testing infrastructure to the forge plugin that validates skill changes at PR time. Structural lint tests run on every PR; behavioral smoke tests (exercising skills via Claude Code CLI) run on-demand via `/test-skills` PR comment, restricted to an allowlist of authorized users.

## Goals

1. Catch structural regressions (broken frontmatter, stale dependencies, orphan skills) on every PR
2. Catch behavioral regressions (skills that produce wrong output) on-demand before merge
3. Only test what's affected — use a dependency manifest to resolve the blast radius of a change
4. Keep costs controlled — behavioral tests require Claude API calls and run only when an authorized user triggers them

## Non-Goals

- Full integration testing of generated projects (deploying to K8s, running SonarCloud, etc.)
- Performance benchmarking of skills
- Testing skills against multiple Claude model versions

## Architecture

### Dependency Manifest (`skills/dependencies.yaml`)

Central source of truth for the skill dependency graph. Every `forge:<name>` reference in a skill.md must have a corresponding entry.

```yaml
skills:
  bootstrap-project:
    type: orchestrator
    depends:
      - setup-go-module
      - setup-react
      - setup-makefile
      - setup-linting
      - setup-bdd
      - setup-playwright
      - setup-ci
      - setup-sonar
      - setup-helm
      - generate-claude-md

  add-feature:
    type: orchestrator
    depends:
      - bdd-feature
      - tdd-cycle
      - quality-check

  bdd-feature:
    type: workflow
    depends: []
  tdd-cycle:
    type: workflow
    depends: []
  quality-check:
    type: workflow
    depends: []

  setup-go-module:
    type: setup
    depends: []
  setup-react:
    type: setup
    depends: []
  setup-makefile:
    type: setup
    depends: []
  setup-linting:
    type: setup
    depends: []
  setup-bdd:
    type: setup
    depends: []
  setup-playwright:
    type: setup
    depends: []
  setup-ci:
    type: setup
    depends: []
  setup-sonar:
    type: setup
    depends: []
  setup-helm:
    type: setup
    depends: []
  generate-claude-md:
    type: setup
    depends: []
```

### Structural Lint Tests (`tests/lint_test.go`)

Pure file-parsing tests that run on every PR. No Claude CLI needed.

**Test cases:**

1. **TestFrontmatterValid** — every `skills/*/skill.md` has YAML frontmatter with `name` and `description` fields
2. **TestPluginJsonConsistency** — every directory in `skills/` is listed in `plugin.json` and vice versa
3. **TestDependencyManifestConsistency** — every `forge:<name>` reference in skill.md files has a matching entry in `dependencies.yaml`, and every `depends` entry points to a real skill directory
4. **TestNoOrphanSkills** — every skill in `dependencies.yaml` exists in both `plugin.json` and `skills/` directory
5. **TestNoCyclicDependencies** — the dependency graph is a DAG

### Behavioral Test Harness (`tests/skills_test.go`)

Each skill gets a Go test function that:

1. Creates a temp directory
2. Runs Claude Code CLI with a canned prompt targeting the skill
3. Asserts at the appropriate depth level (see below)

#### Assertion Levels

**Level 2 — Quality gates pass:** File structure correct, `make lint` / `make typecheck` / `make test` exit 0.

**Level 3 — Infrastructure works:** Docker builds, Kind cluster starts, Helm deploys, app responds on HTTP.

**Level 4 — CI validates itself:** Push to a temp GitHub repo, trigger Actions, wait for green, tear down.

#### Per-Skill Assertions

| Skill | Level | Assertions |
|-------|-------|------------|
| `setup-go-module` | 2 | `go.mod` exists with correct module path, `cmd/<name>/main.go` exists, `internal/` exists, `go build ./...` succeeds |
| `setup-react` | 2 | `web/package.json` exists, `web/src/` exists, `web/vite.config.ts` exists, `bun install` succeeds, `bun run build` succeeds |
| `setup-makefile` | 2 | `Makefile` exists, contains expected targets (`lint`, `test`, `typecheck`, `build`, `dev-api`, `dev-web`), `make help` exits 0 |
| `setup-linting` | 2 | `.golangci.yml` exists, ESLint config exists, `make lint` exits 0 |
| `setup-bdd` | 2 | `features/` dir exists with at least one `.feature` file, step definition stubs exist, `make test` exits 0 |
| `setup-playwright` | 2 | `playwright.config.ts` exists, `web/e2e/` dir exists, `bun run playwright install` succeeds |
| `setup-sonar` | 2 | `sonar-project.properties` exists, contains correct project key, coverage paths configured |
| `generate-claude-md` | 2 | `CLAUDE.md` exists, contains `forge` plugin reference, contains `add-feature` workflow reference |
| `setup-helm` | 3 | `deploy/helm/<name>/Chart.yaml` exists, `helm lint` passes, `make cluster-create` starts Kind (vendored in `bin/`), `helm install` succeeds, HTTP GET to app returns 200 |
| `setup-ci` | 4 | `.github/workflows/ci.yml` exists with expected jobs, push to temp GitHub repo, Actions trigger, CI workflow reaches green, temp repo deleted |
| `bootstrap-project` | 2+3+4 | Runs all setup skills, `make lint` + `make typecheck` + `make test` pass, Kind cluster + Helm deploy works, CI workflow validates on temp repo |
| `add-feature` | 2 | Pre-bootstrap, then add canned feature ("health check endpoint"), `.feature` file created, step definitions created, handler code exists, `make test` passes with new tests |
| `bdd-feature` | 2 | Pre-bootstrap, run with canned prompt, `.feature` file created with valid Gherkin, step stubs generated |
| `tdd-cycle` | 2 | Pre-bootstrap with existing `.feature` + stubs, run cycle, implementation code created, tests pass, git log shows red/green/refactor commits |
| `quality-check` | 2 | Pre-bootstrap, run quality check, all three gates (lint, typecheck, test) reported as pass |

#### Temp GitHub Repo Lifecycle (Level 4)

For CI validation tests:

1. **Create:** `gh repo create forge-test-<timestamp> --private --clone` in a temp directory
2. **Populate:** Copy generated project files into the repo, commit and push
3. **Trigger:** Push triggers the generated `ci.yml` workflow automatically
4. **Wait:** Poll `gh run list` until the workflow completes (timeout: 15 min)
5. **Assert:** `gh run view <id>` shows success, all jobs passed
6. **Teardown:** `gh repo delete forge-test-<timestamp> --yes` (always, even on failure)

Requires a `GITHUB_TOKEN` secret with `repo` and `workflow` scopes. The test uses `t.Cleanup()` to guarantee repo deletion.

#### Shared Pre-Bootstrap Fixture

Workflow skill tests (add-feature, bdd-feature, tdd-cycle, quality-check) all need a bootstrapped project. To avoid re-bootstrapping for each test:

- `TestMain` runs bootstrap-project once into a shared fixture directory
- Individual workflow tests copy the fixture into their own temp dir via `cp -r`
- If bootstrap fails, all dependent tests skip with `t.Skip("bootstrap fixture failed")`

**Helper functions:**

- `runClaude(t, dir, prompt)` — invokes `claude -p --allowedTools '*' --dir <dir> "<prompt>"`, captures output
- `assertFileExists(t, dir, path)` — checks file exists
- `assertFileContains(t, dir, path, substring)` — checks file content
- `runMake(t, dir, target)` — runs `make <target>` in dir, fails test on non-zero exit

### Dependency-Aware Test Selection (`tests/cmd/affected/main.go`)

A small CLI that:

1. Reads `skills/dependencies.yaml`
2. Runs `git diff --name-only origin/main...HEAD` to find changed files
3. Maps changed files to affected skills (e.g., `skills/setup-helm/skill.md` -> `setup-helm`)
4. Resolves parent orchestrators that depend on affected skills
5. Outputs a Go test `-run` regex (e.g., `TestSetupHelm|TestBootstrapProject`)

**Special cases:**

- Changes to `dependencies.yaml`, test infrastructure, or `plugin.json` trigger all behavioral tests
- Changes to non-skill files (README, docs) trigger no behavioral tests

### GitHub Actions

#### `ci.yml` — Every PR

Runs structural lint tests only. Fast, free, deterministic.

```yaml
name: CI
on:
  pull_request:
    branches: [main]
jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - checkout
      - setup-go 1.22
      - go test ./tests/ -run 'TestFrontmatter|TestPluginJson|TestDependency|TestNoOrphan|TestNoCyclic' -v
```

#### `test-skills.yml` — On-demand via `/test-skills` comment

Runs behavioral tests for affected skills only. Requires Claude API key.

```yaml
name: Behavioral Skill Tests
on:
  issue_comment:
    types: [created]
```

**Gating:**

- Triggers only on PR comments containing `/test-skills`
- Author must be in an allowlist (JSON array in the workflow, initially `["petrepopescu21"]`)
- Checks out the PR head ref

**Runner requirements:**

- Docker pre-installed (ubuntu-latest has it)
- `gh` CLI for Level 4 tests (pre-installed on ubuntu-latest)
- Kind is NOT installed in CI — the generated Makefile vendors it into `bin/` with a pinned version. Level 3 tests use `make cluster-create` which handles Kind installation.

**Secrets required:**

- `ANTHROPIC_API_KEY` — for Claude Code CLI
- `FORGE_TEST_GITHUB_TOKEN` — PAT with `repo` + `workflow` scopes, for creating/deleting temp repos and triggering Actions. Separate from the default `GITHUB_TOKEN` because it needs cross-repo permissions.

**Steps:**

1. Checkout PR branch
2. Install Go, Node, Bun, Claude Code CLI, Kind
3. Run `tests/cmd/affected` to determine which tests to run
4. Run `go test ./tests/ -run "$AFFECTED" -v -timeout 30m`
5. Post results as a PR comment (pass/fail with per-test breakdown)

### CLAUDE.md

A `CLAUDE.md` at the repo root enforcing:

- When modifying any skill.md that adds/removes a `forge:<name>` reference, update `skills/dependencies.yaml`
- When adding/removing a skill directory, update both `plugin.json` and `dependencies.yaml`
- Run `go test ./tests/ -run TestLint` before committing skill changes

### Repo-Level Skills

Install **skill-creator** as a plugin — this repo's primary activity is building and iterating on skills.

No other repo-level skills needed. The standard superpowers (brainstorming, TDD, debugging) are available globally.

### Dogfooding

- The Go test harness is written using `superpowers:test-driven-development`
- The behavioral tests ARE the dogfooding — they exercise every forge skill end-to-end
- The forge plugin itself cannot fully dogfood on its own repo (it generates Go+React projects, this repo is markdown + Go tests)

## File Map

```
forge/
  skills/
    dependencies.yaml          # NEW — dependency manifest
  tests/
    lint_test.go               # NEW — structural validation
    skills_test.go             # NEW — behavioral smoke tests
    helpers_test.go            # NEW — shared test utilities
    cmd/
      affected/
        main.go                # NEW — dependency-aware test selector
  .github/
    workflows/
      ci.yml                   # NEW — structural lint on every PR
      test-skills.yml          # NEW — behavioral tests on /test-skills
  CLAUDE.md                    # NEW — repo conventions and enforcement
  go.mod                       # NEW — Go module for tests
```

## Open Questions

None — all decisions resolved during brainstorming.
