# Forge Plugin Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Claude Code superpowers plugin with 15 skills for Go + React/TypeScript project bootstrapping and BDD/TDD-driven development workflows.

**Architecture:** Microskills with orchestrators. Each skill is a standalone `skill.md` file. Workflow skills (bdd-feature, tdd-cycle, quality-check) define the core development loop. Setup skills scaffold project layers. Orchestrators compose them.

**Tech Stack:** Claude Code superpowers plugin format, Gherkin/godog (Go BDD), playwright-bdd (frontend BDD), Vitest, Playwright, GitHub Actions, SonarCloud, Helm, golangci-lint, ESLint.

**Spec:** `docs/specs/2026-03-27-forge-plugin-design.md`

---

## File Map

All files live under the forge repo root.

| File | Responsibility |
|------|---------------|
| `plugin.json` | Plugin manifest (already created) |
| `skills/quality-check/skill.md` | Pre-commit gate: lint, typecheck, test |
| `skills/tdd-cycle/skill.md` | Red/green/refactor loop with discipline enforcement |
| `skills/bdd-feature/skill.md` | Prompt → Gherkin scenarios → step stubs |
| `skills/setup-go-module/skill.md` | Go module + cmd/ + internal/ scaffold |
| `skills/setup-react/skill.md` | Vite + React + TS + Vitest scaffold |
| `skills/setup-makefile/skill.md` | Makefile with standard targets |
| `skills/setup-linting/skill.md` | golangci-lint + ESLint configs |
| `skills/setup-bdd/skill.md` | godog + playwright-bdd setup |
| `skills/setup-playwright/skill.md` | Playwright configs + fixtures |
| `skills/setup-ci/skill.md` | Consolidated GitHub Actions workflows |
| `skills/setup-sonar/skill.md` | SonarCloud config + coverage wiring |
| `skills/setup-helm/skill.md` | Helm chart + Kind + scripts |
| `skills/generate-claude-md/skill.md` | CLAUDE.md with skill enforcement |
| `skills/add-feature/skill.md` | Orchestrator: prompt → BDD → TDD → quality |
| `skills/bootstrap-project/skill.md` | Orchestrator: full project scaffold |
| `README.md` | Installation and usage docs |

---

### Task 1: quality-check skill

**Files:**
- Create: `skills/quality-check/skill.md`

- [ ] **Step 1: Create the skill file**

```markdown
---
name: quality-check
description: Run all quality gates (lint, typecheck, test) and report results. Use before every commit, after completing a feature, or when you need to verify the project is clean. Trigger when you see phrases like "check quality", "run checks", "is this clean", "before I commit", or after completing any implementation work.
---

# Quality Check

Pre-commit quality gate. Runs all checks and reports pass/fail for each.

**Announce:** "Running quality checks — lint, typecheck, and tests."

## Why This Exists

Catching lint errors, type errors, and test failures before they hit CI saves time and keeps the commit history clean. This skill runs all three gates in a predictable order and gives you a clear pass/fail report.

## Process

Run each gate in order. Do not stop on first failure — run all three and report everything.

### Step 1: Lint

```bash
make lint
```

This runs `golangci-lint run ./...` for Go and `cd web && bun run lint` for the frontend. Both must pass.

### Step 2: Typecheck

```bash
make typecheck
```

This runs `cd web && bun run typecheck` (tsc --noEmit). Must produce zero errors.

### Step 3: Test

```bash
make test
```

This runs `go test ./...` for Go and `cd web && bun run test` for frontend unit tests.

## Reporting

After all three gates have run, present a summary:

| Gate | Status |
|------|--------|
| Lint | PASS / FAIL |
| Typecheck | PASS / FAIL |
| Test | PASS / FAIL |

If any gate failed, show the relevant error output. Do not attempt to auto-fix — just report what needs attention.

## When NOT to Use

- Do not run this during active TDD red/green cycles (tests are expected to fail during red phase)
- Do not run E2E tests — this is for fast pre-commit checks only
```

- [ ] **Step 2: Verify file exists and is valid markdown**

Run: `cat skills/quality-check/skill.md | head -5`
Expected: frontmatter with `name: quality-check`

- [ ] **Step 3: Commit**

```bash
git add skills/quality-check/skill.md
git commit -m "feat: add quality-check skill — pre-commit gate for lint, typecheck, test"
```

---

### Task 2: tdd-cycle skill

**Files:**
- Create: `skills/tdd-cycle/skill.md`

- [ ] **Step 1: Create the skill file**

```markdown
---
name: tdd-cycle
description: Drive test-driven development with strict red/green/refactor discipline. Use when implementing any feature, fixing bugs, or writing new code. Trigger on phrases like "implement this", "write the code", "build this feature", "fix this bug", or when transitioning from BDD scenarios to implementation. This skill enforces that failing tests exist before implementation code is written.
---

# TDD Cycle

Drives implementation through strict red/green/refactor discipline.

**Announce:** "Using TDD cycle — I'll write failing tests first, then implement."

## The Iron Law

**No implementation code without a failing test first.**

This is not optional. Every piece of functionality gets a test that fails before the implementation exists, then passes after. If you find yourself writing implementation code without a red test, stop and write the test first.

## Process

### Starting Point

The TDD cycle can start from two places:

1. **From BDD step stubs** — the `bdd-feature` skill has generated `.feature` files and step definition stubs. Each unimplemented step is a failing test. Pick the first one.
2. **Standalone** — no BDD context. Identify the first unit of behavior to implement and write a test for it.

### The Loop

For each unit of behavior:

#### 1. Red — Write the Failing Test

Write a test that describes what the code should do. The test must fail because the implementation doesn't exist yet.

**Go pattern:**
```go
func TestLeadAssignment_AssignsToRep(t *testing.T) {
    t.Parallel()
    svc := NewLeadService(stubRepo{}, stubRBAC{})
    err := svc.Assign(ctx, leadID, repID)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
}
```

**React pattern:**
```typescript
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

it('assigns lead when rep is selected', async () => {
  render(<AssignmentDialog leadId="123" />)
  await userEvent.click(screen.getByRole('button', { name: 'Assign' }))
  expect(screen.getByText('Assigned')).toBeInTheDocument()
})
```

#### 2. Run — Verify It Fails

```bash
# Go
go test ./internal/leads/... -run TestLeadAssignment_AssignsToRep -v

# Frontend
cd web && bun run test -- --reporter=verbose AssignmentDialog
```

Confirm the test fails for the expected reason (missing function, wrong return value, etc.) — not for an unrelated error like a syntax mistake in the test itself.

#### 3. Green — Minimum Implementation

Write the smallest amount of code that makes the test pass. Do not add error handling, validation, or features that the test doesn't require. If you're tempted to add something "while you're here," resist — write a test for it first.

#### 4. Run — Verify It Passes

```bash
make test
```

All tests must pass — not just the new one. If an existing test breaks, fix the implementation, not the test (unless the test was wrong).

#### 5. Refactor — Clean Up

With green tests as your safety net, clean up:
- Remove duplication
- Improve naming
- Extract helpers if a pattern repeats 3+ times
- Simplify logic

Run `make test` again after refactoring to confirm nothing broke.

#### 6. Commit

```bash
git add -A
git commit -m "feat: <what this test+implementation achieves>"
```

Commit after each red→green→refactor cycle, not after batching multiple features.

#### 7. Repeat

Pick the next unimplemented behavior and go back to step 1.

### Completion

After all behaviors are implemented and all tests pass:

1. Run the full quality check: `make lint && make typecheck && make test`
2. If any gate fails, fix the issue and re-run
3. Summarize what was implemented and how many tests were added

## Test Design Guidelines

### Go Tests
- Use `_test` package suffix (e.g., `package leads_test`) to test the public API
- Use `t.Parallel()` on every test and subtest
- Use manual stubs over mocking frameworks — implement the interface with the behavior you need
- Use `t.Helper()` on test helper functions
- Wrap test setup in helper functions like `newTestService()` that return configured dependencies

### Frontend Tests
- Use React Testing Library — query by role, label, or text, never by CSS class or test ID
- Use `userEvent` (not `fireEvent`) for user interactions
- Test behavior, not implementation — don't assert on internal state
- Use `vi.fn()` for callback mocking
- Use `renderHook` from `@testing-library/react` for hook tests

## Red Flags

If you notice any of these, stop and correct course:

| Symptom | Problem |
|---------|---------|
| Writing implementation before a test | Violates the iron law — write the test first |
| Test passes immediately when written | The test isn't testing anything new, or the implementation already exists |
| Large implementation between test runs | Break it into smaller steps — each step should be one test |
| Fixing tests to match implementation | Tests define behavior — if the test is right, fix the implementation |
| Skipping refactor phase | Technical debt accumulates — take 30 seconds to clean up |
```

- [ ] **Step 2: Verify file exists**

Run: `cat skills/tdd-cycle/skill.md | head -5`
Expected: frontmatter with `name: tdd-cycle`

- [ ] **Step 3: Commit**

```bash
git add skills/tdd-cycle/skill.md
git commit -m "feat: add tdd-cycle skill — red/green/refactor with discipline enforcement"
```

---

### Task 3: bdd-feature skill

**Files:**
- Create: `skills/bdd-feature/skill.md`

- [ ] **Step 1: Create the skill file**

```markdown
---
name: bdd-feature
description: Transform a rough feature prompt into expert Gherkin scenarios and generate step definition stubs. Use when starting any new feature, adding behavior, or when the user provides a feature description (even vague ones). Trigger on phrases like "add feature", "I want users to be able to", "new capability", "implement X", or any description of desired behavior. This skill acts as a BDD analyst — it proposes concrete scenarios rather than asking clarifying questions.
---

# BDD Feature

Transforms rough prompts into expert Gherkin scenarios with step definition stubs.

**Announce:** "Using BDD feature skill — I'll propose Gherkin scenarios for your approval before any implementation."

## Why This Matters

Starting with Gherkin scenarios forces you to define behavior precisely before writing code. It catches ambiguity early, documents acceptance criteria, and produces tests that verify the feature works end-to-end. The scenarios become the single source of truth for what the feature does.

## Process

### Step 1: Analyze the Prompt

Read the user's prompt and extract:
- **Actor(s):** Who is performing the action? (e.g., rep, manager, admin, system)
- **Action(s):** What are they trying to do?
- **Outcome(s):** What should happen when they succeed?
- **Edge cases:** What can go wrong? (unauthorized, not found, invalid state, conflicts)

Also read:
- `CLAUDE.md` for domain context, roles, and existing conventions
- Existing `.feature` files for style consistency and to avoid duplicating scenarios

### Step 2: Propose Scenarios

Do NOT ask clarifying questions. Instead, propose what the feature should look like based on domain context. Present a complete `.feature` file for approval.

**The analyst behavior:** If the prompt is vague (e.g., "add assignments"), propose a fully-formed feature with your best interpretation of what it should do. The user will correct what's wrong — this is faster than a Q&A loop.

Structure the feature file like this:

```gherkin
Feature: Lead Assignment
  As a manager
  I want to assign leads to reps
  So that each lead has a responsible rep for follow-up

  Background:
    Given the system has the following reps:
      | name       | role    |
      | Alice      | rep     |
      | Bob        | rep     |
    And the following unassigned leads exist:
      | company    | status     |
      | Acme Corp  | new        |

  Scenario: Manager assigns a lead to a rep
    When the manager assigns "Acme Corp" to "Alice"
    Then the lead "Acme Corp" should be assigned to "Alice"
    And a "lead_assigned" event should be recorded

  Scenario: Cannot assign to a non-existent rep
    When the manager assigns "Acme Corp" to "Unknown"
    Then the assignment should fail with "rep not found"

  Scenario: Rep cannot assign leads
    Given the current user is a rep
    When the rep tries to assign "Acme Corp" to "Bob"
    Then the assignment should fail with "forbidden"

  Scenario Outline: Assignment updates lead status
    Given a lead with status "<initial_status>"
    When the manager assigns the lead to a rep
    Then the lead status should be "<final_status>"

    Examples:
      | initial_status | final_status |
      | new            | assigned     |
      | contacted      | assigned     |
```

**Include:**
- Happy path scenario(s)
- Error/edge cases (unauthorized, not found, invalid state, conflicts)
- Scenario Outlines for parameterized cases where 3+ scenarios differ only by data
- Background section for shared setup (when 2+ scenarios share Given steps)

**Feature file conventions:**
- Use domain language from `CLAUDE.md` (e.g., "rep", "manager", "lead", not "user", "item")
- Steps should be readable by a non-developer
- Avoid implementation details in steps (no "sends a POST request", no "queries the database")
- One Feature per file, focused on one capability

### Step 3: Get Approval

Present the proposed `.feature` file and ask:

> "Here are the scenarios I've drafted for this feature. Do these cover the right behaviors? Want me to add, remove, or change any scenarios?"

Wait for approval. Adjust if the user provides feedback.

### Step 4: Write the Feature File

After approval, write the `.feature` file to the correct location:
- **Backend behavior** → `features/<feature-name>.feature`
- **Frontend behavior** → `web/features/<feature-name>.feature`

Use kebab-case for filenames (e.g., `lead-assignment.feature`).

### Step 5: Generate Step Definition Stubs

#### Go (godog)

Create a test file with step definitions wired to the feature:

```go
// internal/leads/assignment_test.go
package leads_test

import (
    "testing"
    "github.com/cucumber/godog"
)

func TestAssignmentFeatures(t *testing.T) {
    suite := godog.TestSuite{
        ScenarioInitializer: InitializeAssignmentScenario,
        Options: &godog.Options{
            Format:   "pretty",
            Paths:    []string{"../../features/lead-assignment.feature"},
            TestingT: t,
        },
    }
    if suite.Run() != 0 {
        t.Fatal("non-zero exit from godog")
    }
}

func InitializeAssignmentScenario(ctx *godog.ScenarioContext) {
    ctx.Step(`^the system has the following reps:$`, theSystemHasTheFollowingReps)
    ctx.Step(`^the manager assigns "([^"]*)" to "([^"]*)"$`, theManagerAssigns)
    ctx.Step(`^the lead "([^"]*)" should be assigned to "([^"]*)"$`, theLeadShouldBeAssigned)
    // ... remaining steps
}

func theSystemHasTheFollowingReps(table *godog.Table) error {
    return godog.ErrPending
}

func theManagerAssigns(lead, rep string) error {
    return godog.ErrPending
}

func theLeadShouldBeAssigned(lead, rep string) error {
    return godog.ErrPending
}
```

#### Frontend (playwright-bdd)

Create step definitions for playwright-bdd:

```typescript
// web/e2e/steps/lead-assignment.steps.ts
import { createBdd } from 'playwright-bdd'
import { test } from '../fixtures'

const { Given, When, Then } = createBdd(test)

Given('the system has the following reps:', async ({ page }, table) => {
  // TODO: implement — mock API or seed data
})

When('the manager assigns {string} to {string}', async ({ page }, lead: string, rep: string) => {
  // TODO: implement
})

Then('the lead {string} should be assigned to {string}', async ({ page }, lead: string, rep: string) => {
  // TODO: implement
})
```

### Step 6: Verify Stubs Fail

Run the tests to confirm the stubs are wired correctly and fail with "pending" status:

```bash
# Go
go test ./internal/leads/... -run TestAssignmentFeatures -v

# Frontend
cd web && bun run test:e2e
```

The tests should fail because steps return `ErrPending` or have `TODO` implementations. This confirms the BDD scaffolding is wired correctly and ready for TDD implementation.

## After This Skill

The next step is `forge:tdd-cycle` to implement each step definition using red/green/refactor. The step stubs are the failing tests — implement them one at a time.
```

- [ ] **Step 2: Verify file exists**

Run: `cat skills/bdd-feature/skill.md | head -5`
Expected: frontmatter with `name: bdd-feature`

- [ ] **Step 3: Commit**

```bash
git add skills/bdd-feature/skill.md
git commit -m "feat: add bdd-feature skill — prompt to Gherkin scenarios with analyst behavior"
```

---

### Task 4: setup-go-module skill

**Files:**
- Create: `skills/setup-go-module/skill.md`

- [ ] **Step 1: Create the skill file**

```markdown
---
name: setup-go-module
description: Scaffold a Go backend with module init, cmd/ entrypoint, and internal/ package structure. Use when bootstrapping a new Go project or when the bootstrap-project orchestrator invokes this skill. Trigger on "set up Go", "init Go module", "create Go backend", or as part of project bootstrapping.
---

# Setup Go Module

Scaffolds a Go backend with the standard project layout.

**Announce:** "Setting up Go module with cmd/ and internal/ structure."

## Prerequisites

- Go installed (1.22+)
- Empty or new project directory

## Process

### Step 1: Ask for Module Path

If not provided by the orchestrator, ask:

> "What's the Go module path? (e.g., `github.com/petrepopescu21/myproject`)"

### Step 2: Initialize Module

```bash
go mod init <module-path>
```

### Step 3: Create Directory Structure

```
cmd/
  <project-name>/
    main.go
internal/
  api/
    router.go
  domain/
  store/
```

### Step 4: Write Entrypoint

Create `cmd/<project-name>/main.go`:

```go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown error: %v", err)
	}
}
```

### Step 5: Write Router Stub

Create `internal/api/router.go`:

```go
package api

import "net/http"

// NewRouter creates the application HTTP router with all routes registered.
func NewRouter() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	return mux
}
```

### Step 6: Verify

```bash
go build ./...
go vet ./...
```

Both must succeed with no errors.

### Step 7: Commit

```bash
git add -A
git commit -m "feat: scaffold Go module with cmd/ entrypoint and internal/ structure"
```
```

- [ ] **Step 2: Commit**

```bash
git add skills/setup-go-module/skill.md
git commit -m "feat: add setup-go-module skill — Go project scaffold"
```

---

### Task 5: setup-react skill

**Files:**
- Create: `skills/setup-react/skill.md`

- [ ] **Step 1: Create the skill file**

```markdown
---
name: setup-react
description: Scaffold a React + TypeScript frontend with Vite, Bun, TanStack Query/Table, Vitest, and React Testing Library. Use when bootstrapping a frontend or when the bootstrap-project orchestrator invokes this skill. Trigger on "set up frontend", "create React app", "add frontend", or as part of project bootstrapping.
---

# Setup React

Scaffolds a React + TypeScript frontend with Vite, Bun, and testing infrastructure.

**Announce:** "Setting up React + TypeScript frontend with Vite and Bun."

## Prerequisites

- Bun installed (1.1+)
- Project root directory exists

## Process

### Step 1: Create Vite Project

```bash
cd <project-root>
bun create vite web --template react-ts
cd web
bun install
```

### Step 2: Install Dependencies

```bash
bun add @tanstack/react-query @tanstack/react-table
bun add -d @testing-library/react @testing-library/jest-dom @testing-library/user-event \
  @vitejs/plugin-react vitest jsdom @vitest/coverage-v8
```

### Step 3: Configure Vite

Write `web/vite.config.ts`:

```typescript
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': new URL('./src', import.meta.url).pathname,
    },
  },
  server: {
    port: 5174,
    proxy: {
      '/api': 'http://localhost:8080',
    },
  },
})
```

### Step 4: Configure Vitest

Write `web/vitest.config.ts`:

```typescript
import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./src/test/setup.ts'],
    exclude: ['e2e/**', 'node_modules/**'],
    passWithNoTests: true,
    css: true,
    coverage: {
      provider: 'v8',
      reporter: ['text', 'lcov'],
      reportsDirectory: './coverage',
    },
  },
  resolve: {
    alias: {
      '@': new URL('./src', import.meta.url).pathname,
    },
  },
})
```

### Step 5: Create Test Setup

Write `web/src/test/setup.ts`:

```typescript
import '@testing-library/jest-dom'
```

### Step 6: Configure TypeScript

Ensure `web/tsconfig.json` has strict mode and path aliases:

```json
{
  "compilerOptions": {
    "target": "ES2020",
    "useDefineForClassFields": true,
    "lib": ["ES2020", "DOM", "DOM.Iterable"],
    "module": "ESNext",
    "skipLibCheck": true,
    "moduleResolution": "bundler",
    "allowImportingTsExtensions": true,
    "isolatedModules": true,
    "moduleDetection": "force",
    "noEmit": true,
    "jsx": "react-jsx",
    "strict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noFallthroughCasesInSwitch": true,
    "paths": {
      "@/*": ["./src/*"]
    }
  },
  "include": ["src"]
}
```

### Step 7: Add Package Scripts

Ensure `web/package.json` has these scripts:

```json
{
  "scripts": {
    "dev": "vite",
    "build": "tsc -b && vite build",
    "preview": "vite preview",
    "test": "vitest run",
    "test:watch": "vitest",
    "test:e2e": "playwright test",
    "test:e2e:headed": "playwright test --headed",
    "test:e2e:integration": "playwright test --config=playwright-integration.config.ts",
    "lint": "eslint .",
    "typecheck": "tsc --noEmit"
  }
}
```

### Step 8: Verify

```bash
cd web
bun run build
bun run test
bun run typecheck
```

All three must succeed.

### Step 9: Commit

```bash
git add -A
git commit -m "feat: scaffold React + TypeScript frontend with Vite, Vitest, TanStack"
```
```

- [ ] **Step 2: Commit**

```bash
git add skills/setup-react/skill.md
git commit -m "feat: add setup-react skill — React + TS + Vite + Vitest scaffold"
```

---

### Task 6: setup-makefile skill

**Files:**
- Create: `skills/setup-makefile/skill.md`

- [ ] **Step 1: Create the skill file**

```markdown
---
name: setup-makefile
description: Create a Makefile with standard targets for build, test, lint, typecheck, dev servers, E2E, infrastructure, and quality gates. Use when bootstrapping a project or when the bootstrap-project orchestrator invokes this skill. Trigger on "set up Makefile", "create build targets", or as part of project bootstrapping.
---

# Setup Makefile

Creates a Makefile with all standard targets for Go + React/TypeScript projects.

**Announce:** "Setting up Makefile with standard build, test, lint, and infrastructure targets."

## Conventions

- CI/CD pipelines call Makefile targets only — never raw commands
- Single-command targets go directly in the Makefile
- Multi-step or complex logic (2+ commands with branching/loops) extracts to `scripts/`
- All targets are `.PHONY`

## Process

### Step 1: Determine Project Name

Use the directory name or ask the user if invoked standalone.

### Step 2: Create the Makefile

Write the Makefile with these target groups. Adapt paths based on which setup skills have run (e.g., skip web targets if no frontend).

```makefile
# <Project Name> — Makefile
# CI/CD pipelines call these targets only.

.DEFAULT_GOAL := help
.PHONY: help build test lint typecheck dev-api dev-web clean e2e e2e-web e2e-web-integration e2e-cluster e2e-db e2e-deploy cluster-up cluster-deps deploy migrate sonar

# ── Paths ─────────────────────────────────────────────────────────────────────
GO_CMD     := cmd/<project-name>
WEB_DIR    := web
CLUSTER    := <project-name>-local
KIND_CFG   := deploy/kind/kind-config.yaml

# ── AKS safety guard ─────────────────────────────────────────────────────────
AKS_GUARD := @if kubectl get nodes -o jsonpath='{.items[*].metadata.name}' 2>/dev/null | grep -q 'aks-'; then echo 'ERROR: Refusing to run against AKS cluster. This target is for local Kind only.'; exit 1; fi

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*##' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' | sort

build: ## Build Go binary and React frontend
	@go build -o bin/<project-name> ./$(GO_CMD)
	@cd $(WEB_DIR) && bun install --frozen-lockfile && bun run build

test: ## Run Go tests and frontend tests
	@go test ./...
	@cd $(WEB_DIR) && bun run test

lint: ## Run golangci-lint and ESLint
	@golangci-lint run ./...
	@cd $(WEB_DIR) && bun run lint

typecheck: ## Run tsc --noEmit
	@cd $(WEB_DIR) && bun run typecheck

dev-api: ## Run Go API server locally with hot reload
	@air -c .air.toml || go run ./$(GO_CMD)

dev-web: ## Run Vite dev server
	@cd $(WEB_DIR) && bun run dev

e2e: ## Run Go E2E tests
	@go test -v -tags=e2e -count=1 -timeout=10m ./e2e/...

e2e-web: ## Run Playwright E2E tests
	@cd $(WEB_DIR) && bun run test:e2e

e2e-web-integration: ## Run Playwright integration tests against Kind cluster
	@scripts/e2e-web.sh

e2e-cluster: ## Create Kind cluster for E2E
	$(AKS_GUARD)
	@kind create cluster --name $(CLUSTER) --config $(KIND_CFG) --wait 120s

e2e-db: ## Deploy PostgreSQL and seed data for E2E
	$(AKS_GUARD)
	@scripts/cluster-db.sh <project-name>-e2e

e2e-deploy: ## Build and deploy to E2E namespace via Skaffold
	$(AKS_GUARD)
	@skaffold run -p e2e --default-repo="" --status-check=true

cluster-up: ## Recreate local Kind cluster with all dependencies
	$(AKS_GUARD)
	@kind delete cluster --name $(CLUSTER) 2>/dev/null || true
	@kind create cluster --name $(CLUSTER) --config $(KIND_CFG)
	@$(MAKE) cluster-deps

cluster-deps: ## Install cert-manager, ESO, and Envoy Gateway (idempotent)
	$(AKS_GUARD)
	@scripts/cluster-deps.sh

deploy: ## Build and deploy to local Kind cluster via Skaffold
	$(AKS_GUARD)
	@scripts/deploy-local.sh

migrate: ## Run database migrations
	$(AKS_GUARD)
	@go run ./cmd/migrate

sonar: ## Run SonarCloud analysis locally
	@docker run --rm --network=host -v $(CURDIR):/usr/src -w /usr/src \
		sonarsource/sonar-scanner-cli \
		-Dsonar.host.url=https://sonarcloud.io \
		-Dsonar.token=$${SONAR_TOKEN:?Set SONAR_TOKEN}

clean: ## Clean build artifacts
	@rm -rf bin/ web/dist/ web/node_modules/.vite
```

Replace all `<project-name>` placeholders with the actual project name.

### Step 3: Create scripts/ Directory

```bash
mkdir -p scripts
```

Create placeholder scripts that targets reference. These will be populated by other setup skills (setup-helm creates `cluster-db.sh`, `deploy-local.sh`, `cluster-deps.sh`).

### Step 4: Verify

```bash
make help
```

Should list all targets with descriptions.

### Step 5: Commit

```bash
git add Makefile scripts/
git commit -m "feat: add Makefile with standard build, test, lint, and infrastructure targets"
```
```

- [ ] **Step 2: Commit**

```bash
git add skills/setup-makefile/skill.md
git commit -m "feat: add setup-makefile skill — standard Makefile with all targets"
```

---

### Task 7: setup-linting skill

**Files:**
- Create: `skills/setup-linting/skill.md`

- [ ] **Step 1: Create the skill file**

```markdown
---
name: setup-linting
description: Configure golangci-lint for Go and ESLint (flat config) for TypeScript/React with opinionated rulesets. Use when bootstrapping a project or adding linting to an existing one. Trigger on "set up linting", "add linters", "configure ESLint", "configure golangci-lint", or as part of project bootstrapping.
---

# Setup Linting

Configures opinionated linting for Go and TypeScript/React.

**Announce:** "Setting up golangci-lint and ESLint with opinionated configs."

## Process

### Step 1: Create golangci-lint Config

Write `.golangci.yml` at the project root:

```yaml
version: "2"
run:
  go: "1.24"

linters:
  default: none
  enable:
    - govet
    - errcheck
    - staticcheck
    - unused
    - ineffassign
    - gocritic
    - revive
    - tparallel
    - paralleltest

exclusions:
  paths:
    - web/node_modules
  rules:
    - path: test/e2e/
      linters:
        - revive
        - paralleltest

settings:
  gocritic:
    enabled-tags:
      - diagnostic
      - style
      - performance
    disabled-checks:
      - hugeParam

  revive:
    rules:
      - blank-imports
      - context-as-argument
      - dot-imports
      - error-return
      - error-strings
      - error-naming
      - exported
      - increment-decrement
      - var-naming
      - range
      - receiver-naming
      - time-naming
      - unexported-return
      - indent-error-flow
      - errorf
      - empty-block
      - superfluous-else
      - unreachable-code

output:
  formats:
    text:
      path: stderr
  sort-order:
    - linter
    - severity
    - file
```

### Step 2: Create ESLint Config

Install ESLint dependencies:

```bash
cd web
bun add -d eslint @eslint/js typescript-eslint eslint-plugin-react-hooks eslint-plugin-react-refresh globals
```

Write `web/eslint.config.js`:

```javascript
import js from '@eslint/js'
import tseslint from 'typescript-eslint'
import reactHooks from 'eslint-plugin-react-hooks'
import reactRefresh from 'eslint-plugin-react-refresh'
import globals from 'globals'

export default tseslint.config(
  { ignores: ['dist'] },
  js.configs.recommended,
  ...tseslint.configs.recommended,
  {
    files: ['**/*.{ts,tsx}'],
    languageOptions: {
      ecmaVersion: 2020,
      globals: globals.browser,
    },
    plugins: {
      'react-hooks': reactHooks,
      'react-refresh': reactRefresh,
    },
    rules: {
      ...reactHooks.configs.recommended.rules,
      'react-refresh/only-export-components': [
        'warn',
        { allowConstantExport: true },
      ],
      '@typescript-eslint/no-unused-vars': ['error', { argsIgnorePattern: '^_' }],
      '@typescript-eslint/explicit-function-return-type': 'off',
      '@typescript-eslint/explicit-module-boundary-types': 'off',
    },
  },
)
```

### Step 3: Verify

```bash
golangci-lint run ./...
cd web && bun run lint
```

Both should pass (or only report issues in existing code, not config errors).

### Step 4: Commit

```bash
git add .golangci.yml web/eslint.config.js web/package.json web/bun.lockb
git commit -m "feat: configure golangci-lint and ESLint with opinionated rulesets"
```
```

- [ ] **Step 2: Commit**

```bash
git add skills/setup-linting/skill.md
git commit -m "feat: add setup-linting skill — golangci-lint + ESLint configs"
```

---

### Task 8: setup-bdd skill

**Files:**
- Create: `skills/setup-bdd/skill.md`

- [ ] **Step 1: Create the skill file**

```markdown
---
name: setup-bdd
description: Set up BDD infrastructure with godog for Go and playwright-bdd for the frontend, including feature directories, example feature files, and step definition scaffolds. Use when bootstrapping a project or adding BDD to an existing one. Trigger on "set up BDD", "add Cucumber", "add Gherkin", "set up godog", "set up playwright-bdd", or as part of project bootstrapping.
---

# Setup BDD

Sets up BDD infrastructure for both Go (godog) and frontend (playwright-bdd).

**Announce:** "Setting up BDD with godog and playwright-bdd."

## Process

### Step 1: Set Up Go BDD (godog)

Install godog:

```bash
go get github.com/cucumber/godog@latest
```

Create the features directory:

```bash
mkdir -p features
```

Write an example feature file at `features/health.feature`:

```gherkin
Feature: Health Check
  The API exposes a health endpoint for monitoring.

  Scenario: Health endpoint returns OK
    When I send a GET request to "/healthz"
    Then the response status should be 200
    And the response body should contain "ok"
```

Write the step definitions at `internal/api/health_bdd_test.go`:

```go
package api_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cucumber/godog"
)

type apiContext struct {
	resp *http.Response
	body string
}

func TestHealthFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: func(ctx *godog.ScenarioContext) {
			api := &apiContext{}
			ctx.Step(`^I send a GET request to "([^"]*)"$`, api.iSendAGETRequestTo)
			ctx.Step(`^the response status should be (\d+)$`, api.theResponseStatusShouldBe)
			ctx.Step(`^the response body should contain "([^"]*)"$`, api.theResponseBodyShouldContain)
		},
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/health.feature"},
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("non-zero exit from godog")
	}
}

func (a *apiContext) iSendAGETRequestTo(endpoint string) error {
	// Wire this to your actual router
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Get(srv.URL + endpoint)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	a.resp = resp
	body, _ := io.ReadAll(resp.Body)
	a.body = string(body)
	resp.Body.Close()
	return nil
}

func (a *apiContext) theResponseStatusShouldBe(expected int) error {
	if a.resp.StatusCode != expected {
		return fmt.Errorf("expected status %d, got %d", expected, a.resp.StatusCode)
	}
	return nil
}

func (a *apiContext) theResponseBodyShouldContain(expected string) error {
	if !strings.Contains(a.body, expected) {
		return fmt.Errorf("expected body to contain %q, got %q", expected, a.body)
	}
	return nil
}
```

### Step 2: Set Up Frontend BDD (playwright-bdd)

Install playwright-bdd:

```bash
cd web
bun add -d playwright-bdd
```

Create the features directory:

```bash
mkdir -p web/features
```

Write an example feature file at `web/features/navigation.feature`:

```gherkin
Feature: App Navigation
  Users can navigate between pages using the sidebar.

  Scenario: Navigate to the main page
    Given I am on the app
    When I navigate to the home page
    Then I should see the main heading
```

Write step definitions at `web/e2e/steps/navigation.steps.ts`:

```typescript
import { createBdd } from 'playwright-bdd'
import { test } from '../fixtures'

const { Given, When, Then } = createBdd(test)

Given('I am on the app', async ({ page }) => {
  await page.goto('/')
})

When('I navigate to the home page', async ({ page }) => {
  await page.goto('/')
})

Then('I should see the main heading', async ({ page }) => {
  // Adjust to match your actual app heading
  await page.locator('h1').first().waitFor({ state: 'visible' })
})
```

Update Playwright config to support BDD by adding to `web/playwright.config.ts`:

```typescript
import { defineConfig, devices } from '@playwright/test'
import { defineBddConfig } from 'playwright-bdd'

const testDir = defineBddConfig({
  features: './features/**/*.feature',
  steps: './e2e/steps/**/*.steps.ts',
})

export default defineConfig({
  testDir,
  // ... rest of config
})
```

### Step 3: Verify

```bash
# Go BDD
go test ./internal/api/... -run TestHealthFeatures -v

# Frontend BDD
cd web && bun run test:e2e
```

Both should pass (the health example is fully implemented).

### Step 4: Commit

```bash
git add -A
git commit -m "feat: set up BDD with godog and playwright-bdd, example feature files"
```
```

- [ ] **Step 2: Commit**

```bash
git add skills/setup-bdd/skill.md
git commit -m "feat: add setup-bdd skill — godog + playwright-bdd infrastructure"
```

---

### Task 9: setup-playwright skill

**Files:**
- Create: `skills/setup-playwright/skill.md`

- [ ] **Step 1: Create the skill file**

```markdown
---
name: setup-playwright
description: Configure Playwright with dual configs (unit E2E with auto-start dev server, integration E2E against K8s cluster), fixtures for API mocking, and example specs. Use when bootstrapping a project or adding E2E testing. Trigger on "set up Playwright", "add E2E tests", "configure E2E", or as part of project bootstrapping.
---

# Setup Playwright

Configures Playwright with two configs: unit E2E (local dev server) and integration E2E (K8s cluster).

**Announce:** "Setting up Playwright with dual E2E configs and fixtures."

## Process

### Step 1: Install Playwright

```bash
cd web
bun add -d @playwright/test
bunx playwright install chromium
```

### Step 2: Create Unit E2E Config

Write `web/playwright.config.ts`:

```typescript
import { defineConfig, devices } from '@playwright/test'

export default defineConfig({
  testDir: './e2e',
  testIgnore: ['**/integration/**'],
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: process.env.CI ? 'github' : 'list',
  use: {
    baseURL: 'http://localhost:5174',
    trace: 'on-first-retry',
  },
  projects: [
    { name: 'chromium', use: { ...devices['Desktop Chrome'] } },
  ],
  webServer: {
    command: 'bun run dev',
    url: 'http://localhost:5174',
    reuseExistingServer: !process.env.CI,
    timeout: 30_000,
  },
})
```

### Step 3: Create Integration E2E Config

Write `web/playwright-integration.config.ts`:

```typescript
import { defineConfig, devices } from '@playwright/test'

const baseURL = process.env.E2E_BASE_URL
if (!baseURL) {
  throw new Error('E2E_BASE_URL must be set for integration tests')
}

export default defineConfig({
  testDir: './e2e/integration',
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: 1,
  reporter: process.env.CI ? 'github' : 'list',
  timeout: 30_000,
  use: {
    baseURL,
    trace: 'on-first-retry',
  },
  projects: [
    { name: 'chromium', use: { ...devices['Desktop Chrome'] } },
  ],
})
```

### Step 4: Create Fixtures

Write `web/e2e/fixtures.ts`:

```typescript
import { test as base, type Page } from '@playwright/test'

export async function mockApi(page: Page) {
  await page.route('**/api/v1/**', async (route) => {
    const url = new URL(route.request().url())

    if (url.pathname === '/api/v1/health') {
      await route.fulfill({ status: 200, json: { status: 'ok' } })
      return
    }

    await route.fulfill({ status: 404, json: { error: { code: 'NOT_FOUND', message: 'not found' } } })
  })
}

export const test = base.extend<{}>({
  // Add custom fixtures here as the app grows
})

export { expect } from '@playwright/test'
```

### Step 5: Create Directory Structure

```bash
mkdir -p web/e2e/integration web/e2e/steps
```

### Step 6: Create Example Spec

Write `web/e2e/navigation.spec.ts`:

```typescript
import { test, expect } from '@playwright/test'
import { mockApi } from './fixtures'

test.describe('App navigation', () => {
  test.beforeEach(async ({ page }) => {
    await mockApi(page)
  })

  test('loads the home page', async ({ page }) => {
    await page.goto('/')
    await expect(page).toHaveTitle(/.+/)
  })
})
```

### Step 7: Create E2E Web Script

Write `scripts/e2e-web.sh`:

```bash
#!/usr/bin/env bash
set -euo pipefail

NAMESPACE="${E2E_NAMESPACE:-<project-name>-e2e}"
SERVICE="${E2E_SERVICE:-svc/<project-name>-e2e}"
SERVICE_PORT="${E2E_SERVICE_PORT:-8080}"
WAIT_TIMEOUT="${E2E_WAIT_TIMEOUT:-60}"

# Find a free port
PORT=$(python3 -c 'import socket; s=socket.socket(); s.bind(("",0)); print(s.getsockname()[1]); s.close()')

# Port-forward in background
kubectl port-forward -n "$NAMESPACE" "$SERVICE" "$PORT:$SERVICE_PORT" &
PF_PID=$!
trap "kill $PF_PID 2>/dev/null" EXIT

# Wait for the service to be ready
echo "Waiting for service at localhost:$PORT..."
for i in $(seq 1 "$WAIT_TIMEOUT"); do
  if curl -sf "http://localhost:$PORT/healthz" >/dev/null 2>&1; then
    echo "Service ready after ${i}s"
    break
  fi
  if [ "$i" -eq "$WAIT_TIMEOUT" ]; then
    echo "ERROR: Service not ready after ${WAIT_TIMEOUT}s"
    exit 1
  fi
  sleep 1
done

# Run Playwright integration tests
cd web
E2E_BASE_URL="http://localhost:$PORT" bun run test:e2e:integration "$@"
```

```bash
chmod +x scripts/e2e-web.sh
```

### Step 8: Verify

```bash
cd web && bun run test:e2e
```

Should pass with the example spec (auto-starts Vite dev server).

### Step 9: Commit

```bash
git add -A
git commit -m "feat: set up Playwright with dual E2E configs, fixtures, and example spec"
```
```

- [ ] **Step 2: Commit**

```bash
git add skills/setup-playwright/skill.md
git commit -m "feat: add setup-playwright skill — dual Playwright configs with fixtures"
```

---

### Task 10: setup-ci skill

**Files:**
- Create: `skills/setup-ci/skill.md`

- [ ] **Step 1: Create the skill file**

```markdown
---
name: setup-ci
description: Create consolidated GitHub Actions workflows — one CI pipeline (lint, test, coverage, E2E, SonarCloud), one deploy pipeline, and Dependabot auto-merge. Use when bootstrapping a project or adding CI/CD. Trigger on "set up CI", "add GitHub Actions", "configure CI/CD", "add pipelines", or as part of project bootstrapping.
---

# Setup CI

Creates three consolidated GitHub Actions workflows.

**Announce:** "Setting up GitHub Actions — consolidated CI, deploy, and Dependabot auto-merge."

## Why Consolidated

One CI pipeline that runs everything — lint, test, coverage, E2E, SonarCloud — is the single source of truth for "is this green." No more checking three different workflows to understand build status.

## Process

### Step 1: Create Workflow Directory

```bash
mkdir -p .github/workflows
```

### Step 2: Create CI Workflow

Write `.github/workflows/ci.yml`:

```yaml
name: CI

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]
  workflow_call:

jobs:
  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
          cache: true

      - uses: golangci/golangci-lint-action@v7
        with:
          version: latest

      - uses: oven-sh/setup-bun@v2
        with:
          bun-version: latest

      - run: cd web && bun install --frozen-lockfile
      - run: cd web && bun run lint

      - uses: azure/setup-helm@v4
      - run: helm lint deploy/helm/*/

  test:
    name: Test & Coverage
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
          cache: true

      - uses: oven-sh/setup-bun@v2
        with:
          bun-version: latest

      - run: go test -coverprofile=coverage.out ./...

      - run: cd web && bun install --frozen-lockfile
      - run: cd web && bun run test -- --coverage
      - run: cd web && bun run typecheck

      - uses: actions/upload-artifact@v4
        with:
          name: coverage
          path: |
            coverage.out
            web/coverage/lcov.info

  e2e:
    name: E2E
    runs-on: ubuntu-latest
    timeout-minutes: 20
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
          cache: true

      - uses: oven-sh/setup-bun@v2
        with:
          bun-version: latest

      - name: Install Kind
        run: |
          curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.27.0/kind-linux-amd64
          chmod +x ./kind
          sudo mv ./kind /usr/local/bin/kind

      - name: Install Skaffold
        run: |
          curl -Lo skaffold https://storage.googleapis.com/skaffold/releases/latest/skaffold-linux-amd64
          chmod +x skaffold
          sudo mv skaffold /usr/local/bin/

      - name: Set up E2E environment
        run: |
          make e2e-cluster
          make e2e-db
          make e2e-deploy

      - name: Run Go E2E tests
        run: make e2e

      - name: Install Playwright
        run: cd web && bun install --frozen-lockfile && bunx playwright install --with-deps chromium

      - name: Run Playwright E2E tests
        run: make e2e-web-integration

      - name: Debug output on failure
        if: failure()
        run: |
          echo "=== Pod status ==="
          kubectl get pods -A
          echo "=== App logs ==="
          kubectl logs -n ${E2E_NAMESPACE:-*-e2e} -l app=${PROJECT_NAME:-app} --tail=100 || true
          echo "=== Events ==="
          kubectl get events -A --sort-by='.lastTimestamp' | tail -30

  sonar:
    name: SonarCloud
    runs-on: ubuntu-latest
    needs: [test]
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/download-artifact@v4
        with:
          name: coverage

      - uses: SonarSource/sonarqube-scan-action@v6
        env:
          SONAR_TOKEN: ${{ secrets.SONAR_TOKEN }}
```

### Step 3: Create Deploy Workflow

Write `.github/workflows/deploy.yml`:

```yaml
name: Deploy

on:
  push:
    branches: [main]

concurrency:
  group: deploy
  cancel-in-progress: false

jobs:
  ci:
    uses: ./.github/workflows/ci.yml

  build-and-push:
    name: Build & Push
    runs-on: ubuntu-latest
    needs: [ci]
    steps:
      - uses: actions/checkout@v4

      - name: Generate image tag
        id: tag
        run: echo "tag=sha-${GITHUB_SHA::8}" >> "$GITHUB_OUTPUT"

      - name: Login to ACR
        run: echo "${{ secrets.ACR_PASSWORD }}" | docker login ${{ vars.ACR_REGISTRY }} -u ${{ secrets.ACR_USERNAME }} --password-stdin

      - name: Build and push
        run: |
          docker build -t ${{ vars.ACR_REGISTRY }}/${{ vars.IMAGE_NAME }}:${{ steps.tag.outputs.tag }} .
          docker tag ${{ vars.ACR_REGISTRY }}/${{ vars.IMAGE_NAME }}:${{ steps.tag.outputs.tag }} ${{ vars.ACR_REGISTRY }}/${{ vars.IMAGE_NAME }}:latest
          docker push ${{ vars.ACR_REGISTRY }}/${{ vars.IMAGE_NAME }}:${{ steps.tag.outputs.tag }}
          docker push ${{ vars.ACR_REGISTRY }}/${{ vars.IMAGE_NAME }}:latest

  deploy:
    name: Deploy to AKS
    runs-on: ubuntu-latest
    needs: [build-and-push]
    steps:
      - uses: actions/checkout@v4

      - name: Set up kubeconfig
        run: echo "${{ secrets.KUBE_CONFIG }}" | base64 -d > "$HOME/.kube/config"

      - uses: azure/setup-helm@v4

      - name: Helm upgrade
        run: |
          helm upgrade --install ${{ vars.RELEASE_NAME }} deploy/helm/*/ \
            -f deploy/helm/*/values-aks.yaml \
            --set image.tag=sha-${GITHUB_SHA::8} \
            --wait --timeout 120s

      - name: Verify rollout
        run: kubectl rollout status deployment/${{ vars.RELEASE_NAME }} --timeout=60s
```

### Step 4: Create Dependabot Auto-Merge Workflow

Write `.github/workflows/dependabot-auto-merge.yml`:

```yaml
name: Dependabot Auto-Merge

on:
  pull_request:
    types: [opened, synchronize]

permissions:
  contents: write
  pull-requests: write

jobs:
  auto-merge:
    if: github.actor == 'dependabot[bot]'
    runs-on: ubuntu-latest
    steps:
      - uses: dependabot/fetch-metadata@v2
        id: metadata

      - if: steps.metadata.outputs.update-type != 'version-update:semver-major'
        run: gh pr merge --auto --squash "$PR_URL"
        env:
          PR_URL: ${{ github.event.pull_request.html_url }}
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### Step 5: Commit

```bash
git add .github/
git commit -m "feat: add consolidated GitHub Actions — CI, deploy, Dependabot auto-merge"
```
```

- [ ] **Step 2: Commit**

```bash
git add skills/setup-ci/skill.md
git commit -m "feat: add setup-ci skill — consolidated GitHub Actions workflows"
```

---

### Task 11: setup-sonar skill

**Files:**
- Create: `skills/setup-sonar/skill.md`

- [ ] **Step 1: Create the skill file**

```markdown
---
name: setup-sonar
description: Configure SonarCloud with sonar-project.properties, coverage wiring for Go and TypeScript, and a local scan Makefile target. Use when bootstrapping a project or adding code quality analysis. Trigger on "set up SonarCloud", "add Sonar", "configure code quality", or as part of project bootstrapping.
---

# Setup SonarCloud

Configures SonarCloud integration with dual coverage reporting.

**Announce:** "Setting up SonarCloud with Go and TypeScript coverage wiring."

## Prerequisites

- SonarCloud organization and project created at sonarcloud.io
- `SONAR_TOKEN` secret added to GitHub repo settings

## Process

### Step 1: Ask for SonarCloud Details

If not provided by the orchestrator, ask:

> "What's your SonarCloud organization and project key? (e.g., org: `my-org`, key: `my-org_my-project`)"

### Step 2: Create SonarCloud Config

Write `sonar-project.properties`:

```properties
sonar.projectKey=<org>_<project>
sonar.organization=<org>
sonar.projectName=<Project Name>

# Sources
sonar.sources=cmd,internal,web/src
sonar.tests=internal,web/src
sonar.test.inclusions=**/*_test.go,**/*.test.ts,**/*.test.tsx,**/*.spec.ts,**/*.spec.tsx
sonar.exclusions=**/node_modules/**,**/dist/**,**/vendor/**,**/*.gen.go

# Coverage
sonar.go.coverage.reportPaths=coverage.out
sonar.typescript.lcov.reportPaths=web/coverage/lcov.info

sonar.sourceEncoding=UTF-8
```

Replace `<org>`, `<project>`, and `<Project Name>` with actual values.

### Step 3: Verify Coverage Generation

```bash
# Go coverage
go test -coverprofile=coverage.out ./...

# Frontend coverage
cd web && bun run test -- --coverage
ls web/coverage/lcov.info
```

Both coverage files should be generated.

### Step 4: Test Local Scan (Optional)

If `SONAR_TOKEN` is available:

```bash
make sonar
```

### Step 5: Commit

```bash
git add sonar-project.properties
git commit -m "feat: configure SonarCloud with dual Go/TypeScript coverage"
```
```

- [ ] **Step 2: Commit**

```bash
git add skills/setup-sonar/skill.md
git commit -m "feat: add setup-sonar skill — SonarCloud config with dual coverage"
```

---

### Task 12: setup-helm skill

**Files:**
- Create: `skills/setup-helm/skill.md`

- [ ] **Step 1: Create the skill file**

```markdown
---
name: setup-helm
description: Scaffold a Helm chart, Kind cluster config, and local dev scripts (cluster-db, deploy-local, cluster-deps). Use when bootstrapping a project or adding Kubernetes deployment. Trigger on "set up Helm", "add Kubernetes", "set up K8s", "configure deployment", or as part of project bootstrapping.
---

# Setup Helm

Scaffolds Helm chart, Kind cluster configuration, and local development scripts.

**Announce:** "Setting up Helm chart, Kind config, and local dev scripts."

## Process

### Step 1: Create Helm Chart

```bash
mkdir -p deploy/helm
helm create deploy/helm/<project-name>
```

This generates the standard Helm chart structure. Then customize:

**Update `deploy/helm/<project-name>/values.yaml`** with project defaults:

```yaml
replicaCount: 1

image:
  repository: <project-name>
  pullPolicy: IfNotPresent
  tag: "latest"

service:
  type: ClusterIP
  port: 8080

resources:
  limits:
    cpu: 200m
    memory: 256Mi
  requests:
    cpu: 100m
    memory: 128Mi

livenessProbe:
  httpGet:
    path: /healthz
    port: http
  initialDelaySeconds: 5
  periodSeconds: 10

readinessProbe:
  httpGet:
    path: /healthz
    port: http
  initialDelaySeconds: 5
  periodSeconds: 10
```

**Create `deploy/helm/<project-name>/values-aks.yaml`** for AKS overrides:

```yaml
image:
  repository: <acr-registry>/<project-name>

ingress:
  enabled: true
```

### Step 2: Create Kind Config

Write `deploy/kind/kind-config.yaml`:

```yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
  - role: control-plane
    extraPortMappings:
      - containerPort: 30080
        hostPort: 30080
        protocol: TCP
```

### Step 3: Create Local Dev Scripts

**Write `scripts/cluster-db.sh`:**

```bash
#!/usr/bin/env bash
set -euo pipefail

NAMESPACE="${1:?Usage: cluster-db.sh <namespace> [up|stop|reset]}"
ACTION="${2:-up}"
DB_NAMESPACE="<project-name>-db"
DB_PASSWORD="devpassword"

case "$ACTION" in
  stop)
    kubectl delete namespace "$NAMESPACE" --ignore-not-found
    echo "Namespace $NAMESPACE removed."
    ;;
  reset)
    kubectl delete namespace "$NAMESPACE" --ignore-not-found
    kubectl delete namespace "$DB_NAMESPACE" --ignore-not-found
    sleep 2
    exec "$0" "$NAMESPACE" up
    ;;
  up)
    # Create DB namespace
    kubectl create namespace "$DB_NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -

    # Deploy PostgreSQL
    kubectl apply -n "$DB_NAMESPACE" -f - <<DBEOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: postgres
spec:
  replicas: 1
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
    spec:
      containers:
        - name: postgres
          image: postgres:16
          env:
            - name: POSTGRES_DB
              value: app
            - name: POSTGRES_USER
              value: app
            - name: POSTGRES_PASSWORD
              value: "$DB_PASSWORD"
          ports:
            - containerPort: 5432
---
apiVersion: v1
kind: Service
metadata:
  name: postgres
spec:
  selector:
    app: postgres
  ports:
    - port: 5432
      targetPort: 5432
DBEOF

    kubectl wait -n "$DB_NAMESPACE" --for=condition=available deployment/postgres --timeout=60s

    # Create app namespace and secrets
    kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -

    DSN="postgres://app:${DB_PASSWORD}@postgres.${DB_NAMESPACE}.svc.cluster.local:5432/app?sslmode=disable"

    kubectl create secret generic app-secrets \
      -n "$NAMESPACE" \
      --from-literal=db-dsn="$DSN" \
      --from-literal=db-url="$DSN" \
      --from-literal=db-password="$DB_PASSWORD" \
      --from-literal=jwt-secret="dev-jwt-secret" \
      --dry-run=client -o yaml | kubectl apply -f -

    # Run migrations if available
    if [ -d "migrations" ]; then
      echo "Running migrations..."
      go run ./cmd/migrate 2>/dev/null || echo "No migrate command found, skipping."
    fi

    echo "Database ready in $DB_NAMESPACE, secrets in $NAMESPACE."
    ;;
esac
```

**Write `scripts/deploy-local.sh`:**

```bash
#!/usr/bin/env bash
set -euo pipefail

# Load env vars from web/.env.development if it exists
if [ -f web/.env.development ]; then
  set -a
  source web/.env.development
  set +a
fi

skaffold run --default-repo=""
```

**Write `scripts/cluster-deps.sh`:**

```bash
#!/usr/bin/env bash
set -euo pipefail

helm repo add jetstack https://charts.jetstack.io 2>/dev/null || true
helm repo add external-secrets https://charts.external-secrets.io 2>/dev/null || true
helm repo update

helm upgrade --install cert-manager jetstack/cert-manager \
  --namespace cert-manager --create-namespace \
  --set crds.enabled=true --wait

helm upgrade --install external-secrets external-secrets/external-secrets \
  --namespace external-secrets-operator --create-namespace --wait

echo "Cluster dependencies installed."
```

Make all scripts executable:

```bash
chmod +x scripts/cluster-db.sh scripts/deploy-local.sh scripts/cluster-deps.sh
```

### Step 4: Verify

```bash
helm lint deploy/helm/*/
```

Should pass with no errors.

### Step 5: Commit

```bash
git add deploy/ scripts/
git commit -m "feat: scaffold Helm chart, Kind config, and local dev scripts"
```
```

- [ ] **Step 2: Commit**

```bash
git add skills/setup-helm/skill.md
git commit -m "feat: add setup-helm skill — Helm chart, Kind, and dev scripts"
```

---

### Task 13: generate-claude-md skill

**Files:**
- Create: `skills/generate-claude-md/skill.md`

- [ ] **Step 1: Create the skill file**

```markdown
---
name: generate-claude-md
description: Generate a CLAUDE.md file that describes the project and enforces forge skill usage for all future Claude Code sessions. Use as the final step of project bootstrapping or when a project needs a CLAUDE.md. Trigger on "generate CLAUDE.md", "create project context", or as part of project bootstrapping.
---

# Generate CLAUDE.md

Generates a `CLAUDE.md` that enforces forge workflows in all future Claude Code sessions.

**Announce:** "Generating CLAUDE.md with project context and skill enforcement."

## Why This Matters

The `CLAUDE.md` is the bridge between "forge plugin installed" and "skills actually used." Without it, Claude only uses forge skills when explicitly asked. With it, Claude reads the enforcement rules at conversation start and automatically invokes the right skills.

## Process

### Step 1: Gather Context

Read the project to determine:
- Project name (from `go.mod` or `package.json`)
- One-liner description (ask the user if not provided)
- Which layers exist (Go backend? React frontend? Helm? BDD?)
- Current directory structure

### Step 2: Generate CLAUDE.md

Write `CLAUDE.md` at the project root using this template. Include only sections relevant to the layers that were set up.

```markdown
# <Project Name> — Claude Code Context

<One-liner description>

## Required Plugin

This project uses the `forge` plugin for development workflows.

## Mandatory Workflows

### Adding Features

You MUST invoke `forge:add-feature` before writing any feature code.
This drives the BDD→TDD pipeline: prompt → Gherkin scenarios → implementation.
Do NOT skip this workflow — even for "simple" features.

### Writing Code

All implementation follows `forge:tdd-cycle`. Write a failing test first,
then the minimum implementation, then refactor. No exceptions.

### Before Committing

Run `forge:quality-check` before every commit. All three gates must pass:
- `make lint` (zero errors)
- `make typecheck` (no TypeScript errors)
- `make test` (all tests green)

## Architecture

### Backend
- **Language:** Go, standard library preferred
- **Structure:** `cmd/` for main packages, `internal/` for non-exported packages
- **Style:** Errors returned, not panicked. Context as first argument.
- **API:** All endpoints under `/api/v1/...`. Structured JSON errors.

### Frontend
- **Framework:** React + TypeScript (strict mode)
- **Bundler:** Vite
- **Package manager:** Bun — never npm or yarn
- **Server state:** TanStack Query
- **Testing:** Vitest + React Testing Library

### BDD
- **Backend:** godog, feature files in `features/`
- **Frontend:** playwright-bdd, feature files in `web/features/`
- All features start as `.feature` files before any implementation

### Infrastructure
- **Production:** AKS (Azure Kubernetes Service)
- **Packaging:** Helm
- **Local dev:** Kind cluster
- **CI/CD:** GitHub Actions (consolidated pipeline)
- **Code quality:** SonarCloud

## Development Practices

### TDD
Write tests before or alongside implementation. Do not merge untested code.

### BDD
All new features start with Gherkin scenarios. Use `forge:bdd-feature` to generate
`.feature` files from rough prompts.

### Makefile Convention
CI/CD pipelines call Makefile targets only. Simple targets go in the Makefile;
multi-step logic extracts to `scripts/`.

### Secrets Handling
- Read secrets from mounted files (e.g., `/run/secrets/db-password`)
- Never read from environment variables for secret values

## Quality Gates

```bash
make test       # all tests
make lint       # zero errors
make typecheck  # no TypeScript errors
```

## Project Layout

```
<auto-generated tree based on actual project structure>
```
```

Adapt the template:
- Remove sections for layers that weren't set up (e.g., no Frontend section if `setup-react` wasn't invoked)
- Replace `<Project Name>` and `<One-liner description>` with actual values
- Generate the project layout tree from the actual directory structure using `find` or `ls`

### Step 3: Commit

```bash
git add CLAUDE.md
git commit -m "feat: generate CLAUDE.md with forge skill enforcement"
```
```

- [ ] **Step 2: Commit**

```bash
git add skills/generate-claude-md/skill.md
git commit -m "feat: add generate-claude-md skill — CLAUDE.md with skill enforcement"
```

---

### Task 14: add-feature orchestrator skill

**Files:**
- Create: `skills/add-feature/skill.md`

- [ ] **Step 1: Create the skill file**

```markdown
---
name: add-feature
description: Orchestrate the full feature development cycle — from rough prompt through BDD scenarios, TDD implementation, to quality verification. Use when adding any new feature, capability, or behavior. This is the primary development workflow enforced by CLAUDE.md. Trigger on any feature request, behavior description, or when the user says "add", "implement", "build", "create" followed by a feature description. Even for "simple" features, use this workflow.
---

# Add Feature

Orchestrates the full feature development cycle: prompt → BDD → TDD → quality check.

**Announce:** "Using the add-feature workflow — I'll start with BDD scenarios, then implement with TDD."

## The Workflow

This skill orchestrates three other skills in sequence. Do not skip steps.

```dot
digraph add_feature {
  "Receive prompt" -> "Invoke forge:bdd-feature";
  "Invoke forge:bdd-feature" -> "Scenarios approved?";
  "Scenarios approved?" -> "Invoke forge:bdd-feature" [label="revise"];
  "Scenarios approved?" -> "Invoke forge:tdd-cycle" [label="yes"];
  "Invoke forge:tdd-cycle" -> "All steps implemented?";
  "All steps implemented?" -> "Invoke forge:tdd-cycle" [label="next step"];
  "All steps implemented?" -> "Invoke forge:quality-check" [label="yes"];
  "Invoke forge:quality-check" -> "All gates pass?";
  "All gates pass?" -> "Fix issues" [label="no"];
  "Fix issues" -> "Invoke forge:quality-check";
  "All gates pass?" -> "Summary" [label="yes"];
}
```

### Phase 1: BDD — Define Behavior

Invoke the `forge:bdd-feature` skill with the user's prompt.

This produces:
- A `.feature` file with Gherkin scenarios (approved by the user)
- Step definition stubs that fail when run

### Phase 2: TDD — Implement

Invoke the `forge:tdd-cycle` skill.

Starting from the BDD step stubs, implement each step using red/green/refactor:
1. Verify the step stub fails (red)
2. Write minimum implementation to make it pass (green)
3. Refactor
4. Commit
5. Move to the next step

### Phase 3: Quality — Verify

Invoke the `forge:quality-check` skill.

All gates must pass:
- `make lint`
- `make typecheck`
- `make test`

If any gate fails, fix the issue and re-run.

### Phase 4: Summary

After all gates pass, summarize:
- Feature name and description
- Number of scenarios implemented
- Number of tests added (unit + BDD)
- Files created/modified
- Any design decisions made during implementation

## Important

- Do NOT skip the BDD phase, even if the feature seems simple
- Do NOT write implementation code before BDD scenarios are approved
- Do NOT skip the quality check phase
- Each red→green→refactor cycle gets its own commit
```

- [ ] **Step 2: Commit**

```bash
git add skills/add-feature/skill.md
git commit -m "feat: add add-feature orchestrator skill — BDD → TDD → quality pipeline"
```

---

### Task 15: bootstrap-project orchestrator skill

**Files:**
- Create: `skills/bootstrap-project/skill.md`

- [ ] **Step 1: Create the skill file**

```markdown
---
name: bootstrap-project
description: Scaffold a complete Go + React/TypeScript project with all layers — module, frontend, Makefile, linting, BDD, Playwright, CI/CD, SonarCloud, Helm, and CLAUDE.md. Use when starting a new project from scratch. Trigger on "new project", "bootstrap project", "start a project", "scaffold project", "create project", or "init project". This is the top-level orchestrator that invokes all setup skills.
---

# Bootstrap Project

Scaffolds a complete Go + React/TypeScript project by orchestrating all setup skills.

**Announce:** "Bootstrapping a new project — I'll set up all layers and generate CLAUDE.md."

## Process

### Step 1: Gather Project Info

Ask the user (if not already provided):

1. **Project name** — used for directory, Go binary, Helm chart, Kind cluster (e.g., `myproject`)
2. **Module path** — Go module path (e.g., `github.com/petrepopescu21/myproject`)
3. **One-liner description** — what the project does (for CLAUDE.md and README)

### Step 2: Select Layers

Ask which layers to include. All default to yes:

> "Which layers do you want? (all enabled by default, say 'skip X' to exclude)"
> - Go backend
> - React frontend
> - Helm / Kubernetes
> - BDD (godog + playwright-bdd)
> - SonarCloud

### Step 3: Initialize Git

```bash
mkdir -p <project-name>
cd <project-name>
git init
```

### Step 4: Invoke Setup Skills

Invoke each selected skill in this order. Pass the project name, module path, and description to each.

| Order | Skill | Condition |
|-------|-------|-----------|
| 1 | `forge:setup-go-module` | Go backend selected |
| 2 | `forge:setup-react` | React frontend selected |
| 3 | `forge:setup-makefile` | Always |
| 4 | `forge:setup-linting` | Always |
| 5 | `forge:setup-bdd` | BDD selected |
| 6 | `forge:setup-playwright` | React frontend selected |
| 7 | `forge:setup-ci` | Always |
| 8 | `forge:setup-sonar` | SonarCloud selected |
| 9 | `forge:setup-helm` | Helm selected |
| 10 | `forge:generate-claude-md` | Always (last) |

For skills that are skipped, the Makefile and CLAUDE.md should omit the corresponding targets/sections.

### Step 5: Create Supporting Files

**Write `.gitignore`:**

```
# Go
bin/
*.exe

# Frontend
web/node_modules/
web/dist/
web/coverage/
web/.vite/

# IDE
.idea/
.vscode/
*.swp

# Coverage
coverage.out

# OS
.DS_Store
Thumbs.db

# Secrets — never commit
.env
.env.*
!.env.example
```

**Write `Dockerfile`:**

```dockerfile
FROM golang:1.24-alpine AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/app ./cmd/<project-name>

FROM node:22-alpine AS web-builder
RUN npm i -g bun
WORKDIR /app/web
COPY web/package.json web/bun.lockb ./
RUN bun install --frozen-lockfile
COPY web/ .
RUN bun run build

FROM gcr.io/distroless/static-debian12
COPY --from=go-builder /bin/app /app
COPY --from=web-builder /app/web/dist /web/dist
EXPOSE 8080
CMD ["/app", "serve"]
```

### Step 6: Verify the Scaffold

```bash
make lint
make typecheck
make test
```

All three must pass. If any fail, fix the issue before proceeding.

### Step 7: Initial Commit

```bash
git add -A
git commit -m "feat: bootstrap <project-name> with forge — Go + React/TS + BDD + TDD + CI/CD"
```

### Step 8: Summary

Report what was set up:

- Project name and module path
- Layers included/excluded
- File count and directory structure
- How to start developing: "Run `forge:add-feature` to add your first feature"
- Reminder: "CLAUDE.md enforces the forge workflow in all future sessions"
```

- [ ] **Step 2: Commit**

```bash
git add skills/bootstrap-project/skill.md
git commit -m "feat: add bootstrap-project orchestrator — full project scaffold"
```

---

### Task 16: README and plugin.json finalization

**Files:**
- Create: `README.md`
- Modify: `plugin.json`

- [ ] **Step 1: Create README.md**

```markdown
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
| `forge:bootstrap-project` | Scaffold a complete project from scratch |
| `forge:add-feature` | Full feature cycle: prompt → BDD → TDD → quality check |

### Workflow

| Skill | Description |
|-------|-------------|
| `forge:bdd-feature` | Transform a prompt into Gherkin scenarios with step stubs |
| `forge:tdd-cycle` | Red/green/refactor with discipline enforcement |
| `forge:quality-check` | Pre-commit gate: lint, typecheck, test |

### Setup

| Skill | Description |
|-------|-------------|
| `forge:setup-go-module` | Go module + cmd/ + internal/ scaffold |
| `forge:setup-react` | Vite + React + TypeScript + Vitest |
| `forge:setup-makefile` | Makefile with standard targets |
| `forge:setup-linting` | golangci-lint + ESLint configs |
| `forge:setup-bdd` | godog + playwright-bdd infrastructure |
| `forge:setup-playwright` | Dual Playwright configs with fixtures |
| `forge:setup-ci` | Consolidated GitHub Actions workflows |
| `forge:setup-sonar` | SonarCloud with dual coverage |
| `forge:setup-helm` | Helm chart + Kind + dev scripts |
| `forge:generate-claude-md` | CLAUDE.md with skill enforcement |

## How It Works

1. **Bootstrap:** `forge:bootstrap-project` scaffolds your project and generates a `CLAUDE.md` that enforces the forge workflow
2. **Develop:** `forge:add-feature` drives every feature through BDD scenarios → TDD implementation → quality gates
3. **Verify:** `forge:quality-check` runs before every commit

The generated `CLAUDE.md` ensures all future Claude Code sessions automatically use these skills.

## Stack

- **Backend:** Go (standard library preferred)
- **Frontend:** React + TypeScript, Vite, Bun
- **BDD:** godog (Go) + playwright-bdd (frontend) + Gherkin
- **Testing:** Vitest + React Testing Library (unit), Playwright (E2E)
- **CI/CD:** GitHub Actions, SonarCloud
- **Infrastructure:** Kubernetes (AKS), Helm, Kind (local)
- **Linting:** golangci-lint, ESLint (flat config)
```

- [ ] **Step 2: Verify plugin.json matches skill directories**

Run: `ls skills/*/skill.md | wc -l`
Expected: 15 (matching the skills array in plugin.json)

- [ ] **Step 3: Commit**

```bash
git add README.md plugin.json
git commit -m "docs: add README with installation and skill reference"
```

- [ ] **Step 4: Push to GitHub**

```bash
git push -u origin main
```
