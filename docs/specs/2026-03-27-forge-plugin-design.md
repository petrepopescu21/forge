# Forge — Claude Code Superpowers Plugin for Go + React/TypeScript Projects

**Date:** 2026-03-27
**Status:** Approved
**Author:** Petre + Claude

## Summary

Forge is a Claude Code superpowers plugin that encodes opinionated development workflows for Go + React/TypeScript projects. It covers greenfield project bootstrapping, BDD-driven feature development, TDD implementation, and continuous quality enforcement.

The plugin is published as a standalone repo, installable as a superpowers plugin from any project.

## Target Stack

- **Backend:** Go (standard library preferred)
- **Frontend:** React + TypeScript (strict mode), Vite, Bun
- **Testing:** Vitest + React Testing Library (unit), Playwright (E2E), godog (BDD/Go), playwright-bdd (BDD/frontend)
- **CI/CD:** GitHub Actions (consolidated pipeline), SonarCloud
- **Infrastructure:** Kubernetes (AKS prod, Kind local), Helm
- **Linting:** golangci-lint, ESLint (flat config)

## Target Audience

Single developer (Petre) using Claude Code. Skills are tuned for personal workflow, not team onboarding.

## Plugin Structure

```
forge/
├── plugin.json
├── skills/
│   ├── bootstrap-project/
│   │   └── skill.md
│   ├── setup-go-module/
│   │   └── skill.md
│   ├── setup-react/
│   │   └── skill.md
│   ├── setup-makefile/
│   │   └── skill.md
│   ├── setup-ci/
│   │   └── skill.md
│   ├── setup-linting/
│   │   └── skill.md
│   ├── setup-bdd/
│   │   └── skill.md
│   ├── setup-playwright/
│   │   └── skill.md
│   ├── setup-sonar/
│   │   └── skill.md
│   ├── setup-helm/
│   │   └── skill.md
│   ├── generate-claude-md/
│   │   └── skill.md
│   ├── add-feature/
│   │   └── skill.md
│   ├── bdd-feature/
│   │   └── skill.md
│   ├── tdd-cycle/
│   │   └── skill.md
│   └── quality-check/
│       └── skill.md
```

**15 skills total:** 2 orchestrators, 10 setup skills, 3 workflow skills.

## Skill Categories

### Orchestrators

#### `bootstrap-project`

Entry point for new projects.

**Flow:**
1. Ask project name and one-liner description
2. Ask which layers to include (Go backend, React frontend, Helm/K8s — all default yes, each toggleable)
3. Invoke selected setup skills in dependency order:
   `setup-go-module` → `setup-react` → `setup-makefile` → `setup-linting` → `setup-bdd` → `setup-playwright` → `setup-ci` → `setup-sonar` → `setup-helm` → `generate-claude-md`
4. Run `make lint` and `make typecheck` to verify scaffold is clean
5. Create initial git commit

#### `add-feature`

Daily driver for feature work. Enforced by generated `CLAUDE.md`.

**Flow:**
1. Receive rough prompt (e.g., "add lead reassignment")
2. Invoke `bdd-feature` — proposes Gherkin scenarios, gets approval
3. Invoke `tdd-cycle` — drives implementation from approved scenarios
4. Invoke `quality-check` — lint, test, typecheck
5. Summary of what was built

### Setup Skills (Scaffold Layer)

All setup skills are idempotent — they check for existing files before writing, so re-running is safe. The primary use case is greenfield bootstrapping via `bootstrap-project`; individual setup skills can also be invoked standalone to add a missing layer to an existing forge project.

#### `setup-go-module`
- `go mod init` with provided module path
- Create `cmd/<name>/main.go` with minimal server stub
- Create `internal/` directory structure (api, domain, store)
- Create `internal/domain/` with a placeholder entity

#### `setup-react`
- `bun create vite` with React + TypeScript template
- Configure `vite.config.ts` with path aliases (`@/`)
- Set up `tsconfig.json` with strict mode
- Install TanStack Query + TanStack Table
- Create `src/test/setup.ts` with testing-library
- Configure `vitest.config.ts` with jsdom, coverage (v8 + lcov)

#### `setup-makefile`
- Standard targets: `help`, `build`, `test`, `lint`, `typecheck`, `dev-api`, `dev-web`, `clean`
- E2E targets: `e2e`, `e2e-web`, `e2e-web-integration`
- Infrastructure targets: `cluster-up`, `deploy`, `migrate`, `sonar`
- AKS safety guard on destructive targets
- `scripts/` directory for multi-step logic (complex targets extract to scripts)

#### `setup-linting`
- `.golangci.yml` — whitelist approach: govet, errcheck, staticcheck, unused, gocritic, revive, tparallel, paralleltest
- `eslint.config.js` — flat config with typescript-eslint, react-hooks, react-refresh
- `@typescript-eslint/no-unused-vars` error with `^_` ignore pattern

#### `setup-bdd`
- **Go:** install `godog`, create `features/` directory, create step definition scaffold in test files, wire into `go test` via `TestFeatures` function
- **Frontend:** install `playwright-bdd`, create `web/features/` directory, configure feature-file-to-Playwright-test generation
- Create example `.feature` file demonstrating the pattern for each layer

#### `setup-playwright`
- `playwright.config.ts` — unit E2E with Vite dev server auto-start, parallel, chromium
- `playwright-integration.config.ts` — K8s integration (no server, serial, single worker, external base URL)
- `web/e2e/` directory structure with `fixtures.ts` for API mocking
- Example navigation spec

#### `setup-ci`

**Consolidated into three workflows** (not six):

**`ci.yml`** — the single quality pipeline:
- Triggers: push to main/develop + PRs to main
- Reusable via `workflow_call`
- Jobs:
  1. **Lint** — golangci-lint + ESLint + Helm lint (parallel)
  2. **Test & Coverage** — Go tests with `coverage.out` + frontend tests with `lcov.info` + typecheck
  3. **E2E** — Kind cluster + Skaffold deploy + Go E2E + Playwright E2E
  4. **SonarCloud** — collects coverage artifacts from test job, runs scan
- Debug output on E2E failure (pod status, logs, events)

**`deploy.yml`** — deploy only:
- Triggers: push to main
- Depends on CI passing
- Build/push to ACR + Helm deploy to AKS

**`dependabot-auto-merge.yml`** — auto-squash patch/minor updates

#### `setup-sonar`
- `sonar-project.properties` with sources, tests, coverage paths, exclusions for generated code
- Coverage wiring: Go `coverage.out` + TypeScript `web/coverage/lcov.info`
- `sonar` Makefile target with Docker-based scanner for local runs

#### `setup-helm`
- Helm chart scaffold under `deploy/helm/<name>/`
- `values.yaml` + `values-aks.yaml`
- Kind config (`deploy/kind/kind-config.yaml`)
- `scripts/cluster-db.sh` for on-cluster PostgreSQL
- `scripts/deploy-local.sh` for Skaffold local deploy

#### `generate-claude-md`

Generates a `CLAUDE.md` that enforces the forge workflow in all future Claude sessions.

**Generated content:**
- Project description (from bootstrap input)
- **Required plugin** declaration
- **Mandatory workflows:**
  - `forge:add-feature` — must be invoked before any feature code
  - `forge:tdd-cycle` — all implementation follows red/green/refactor
  - `forge:quality-check` — must pass before every commit
- Architecture section (populated based on which setup skills were invoked)
- Quality gates (`make test`, `make lint`, `make typecheck`)
- Project layout (auto-generated tree reflecting actual scaffold)

### Workflow Skills (Daily Development)

#### `bdd-feature`

BDD analyst: transforms rough prompts into expert Gherkin.

**Flow:**
1. Receive rough prompt
2. Analyze — identify actors, actions, outcomes, edge cases
3. Propose complete `.feature` file:
   - Feature description with business context
   - Happy path scenarios
   - Error/edge case scenarios (unauthorized, not found, invalid state)
   - Scenario outlines for parameterized cases
4. Present for approval
5. Write `.feature` file:
   - Backend features → `features/`
   - Frontend features → `web/features/`
6. Generate step definition stubs (godog or playwright-bdd)

**Analyst behavior:** If the prompt is vague, the skill proposes what the feature *should* look like based on domain context from `CLAUDE.md` and existing features, then asks to confirm or adjust. It doesn't just ask "what do you mean?"

#### `tdd-cycle`

Drives red/green/refactor from BDD step definitions or standalone.

**Flow:**
1. Pick the first unimplemented step/test
2. **Red** — write a failing test (or verify BDD step stub fails)
3. **Green** — write minimum implementation to pass
4. **Refactor** — clean up without breaking tests
5. Run `make test` to confirm
6. Repeat for next step/test
7. After all steps pass, run full quality check

**Discipline enforcement:** Will not write implementation before a failing test exists. Runs tests between each step to prove red→green transitions.

#### `quality-check`

Pre-commit gate.

**Flow:**
1. `make lint` — Go + ESLint
2. `make typecheck` — TypeScript
3. `make test` — Go + frontend unit tests
4. Report pass/fail for each gate
5. If any fail, show what needs fixing (report only, no auto-fix)

## Key Design Decisions

| Decision | Choice | Reason |
|---|---|---|
| Plugin format | Superpowers plugin | Installable, versioned, referenceable from any project |
| Skill granularity | Microskills + orchestrators | Composable individually, full-stack via orchestrator |
| BDD tooling | godog + playwright-bdd + Gherkin | `.feature` files as single source of truth for behavior |
| CI structure | One consolidated pipeline | Single source of truth for "is this green" |
| CLAUDE.md generation | Skill enforcement via project instructions | Future sessions automatically use the forge workflow |
| Stack scope | Go + React/TypeScript only | Focused and opinionated over flexible |
| Audience | Single developer (Petre) | Tuned for personal workflow |
