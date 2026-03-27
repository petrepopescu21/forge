---
name: generate-claude-md
description: Generate a CLAUDE.md file that describes the project and enforces forge skill usage for all future Claude Code sessions. Use as the final step of project bootstrapping or when a project needs a CLAUDE.md. Trigger on "generate CLAUDE.md", "create project context", or as part of project bootstrapping.
---

# Generate CLAUDE.md

Generating CLAUDE.md with project context and skill enforcement.

## Announce

"Generating CLAUDE.md with project context and skill enforcement."

## Why This Matters

CLAUDE.md is the bridge between "plugin installed" and "skills actually used." Without it, Claude only uses forge skills when explicitly asked. This file becomes the source of truth for the project — documenting architecture, conventions, constraints, and mandatory workflows. It enforces that every feature starts with `forge:add-feature`, every implementation uses `forge:tdd-cycle`, and every commit is checked with `forge:quality-check`.

## Process

### Step 1: Gather Context

Determine what exists in the project:

1. **Read project metadata:**
   - Check if `go.mod` exists → project uses Go backend
   - Check if `package.json` exists → project uses Node/TypeScript frontend
   - Extract project name from module path or package name

2. **Ask for project description:**
   - Prompt: "Describe this project in 1-2 sentences (e.g., domain, purpose, key users)"
   - Store for the CLAUDE.md intro

3. **Check which layers exist:**
   - Backend: Is there a `cmd/` or `internal/` directory?
   - Frontend: Is there a `web/` directory with React setup?
   - Testing: Are there `features/` directories or test files?
   - Infrastructure: Is there a `deploy/`, `helm/`, or `k8s/` directory?

4. **Read existing conventions:**
   - Check if `Makefile` exists and examine its targets
   - Look for any `.editorconfig`, `tsconfig.json`, or `go.mod` to infer existing decisions
   - If a partial CLAUDE.md exists, preserve existing sections and merge

### Step 2: Generate CLAUDE.md Using Template

Create a CLAUDE.md file with the following structure. Include only sections for layers that exist.

**TEMPLATE:**

```markdown
# [PROJECT_NAME] — Claude Code Project Context

[PROJECT_DESCRIPTION]

## Product

- **Domain:** [category/purpose]
- **Users:** [who uses this?]
- [any other key product attributes]

## Required Plugin: forge

This project uses the **forge** plugin for Claude Code. All developers must have the plugin installed and enabled. Key skills:
- `forge:add-feature` — Start every feature with this skill
- `forge:tdd-cycle` — Implement every feature using TDD
- `forge:quality-check` — Run before every commit

## Mandatory Workflows

**EVERY feature follows this workflow:**

1. **`forge:add-feature`** — Specify what to build, get a plan, create the branch
2. **`forge:tdd-cycle`** — Red/Green/Refactor loop for each test/implementation pair
3. **`forge:quality-check`** — Lint, typecheck, test, verify everything passes
4. **Create pull request** — Link to feature plan, request review

No code should be written without starting with `forge:add-feature`. No implementation without `forge:tdd-cycle`. No commit without `forge:quality-check`.

## Architecture

### Backend

[Include only if Go backend exists]

- **Language:** Go
- **Package manager:** Go modules
- **REST API:** Clean, versioned at `/api/v1/...`
- **Database:** [PostgreSQL / SQLite / etc., if applicable]
- **Auth:** [None / Azure AD / OAuth / etc.]
- **Secrets:** [File mounts / environment variables / etc.]

Directory structure:
- `cmd/` — Executable binaries (main packages)
- `internal/` — Non-exported packages (API handlers, business logic, repositories)
- [other key directories based on project]

### Frontend

[Include only if React/TypeScript frontend exists]

- **Language:** TypeScript (strict mode)
- **Framework:** React with functional components and hooks
- **Bundler:** Vite
- **Package manager:** Bun — use `bun install`, `bun run`. Never npm or yarn.
- **Server state:** TanStack Query (`@tanstack/react-query`)
- **Data grids:** TanStack Table (`@tanstack/react-table`)
- **Styling:** [Tailwind / styled-components / CSS Modules / etc.]

Directory structure:
- `web/src/components/` — React `.tsx` components
- `web/src/hooks/` — Custom React hooks
- `web/src/pages/` — Page-level components
- [other key directories]

TypeScript config: `"strict": true` enforced. No implicit `any`.

### Testing

[Include only if BDD/testing setup exists]

- **Backend:** [godog / Go testing package]
- **Frontend:** [Playwright / Vitest / Jest / etc.]
- **Feature files:** `features/` directory (`.feature` files in Gherkin)
- **Step definitions:** `internal/*_test.go` (backend) or `web/e2e/steps/` (frontend)

### Infrastructure

[Include only if deployment/k8s setup exists]

- **Deployment:** [Kubernetes / Docker / Heroku / etc.]
- **Package:** [Helm / Kustomize / Docker Compose / etc.]
- **Configuration:** [ConfigMaps / ExternalSecrets / environment files]

Directory structure:
- `deploy/helm/` or `k8s/` — Infrastructure as code
- `scripts/` — Deployment and utility scripts

## Development Practices

### TDD (Test-Driven Development)

Write tests before or alongside implementation. Every feature uses the red/green/refactor cycle:

1. **Red:** Write a failing test
2. **Green:** Write minimal code to pass the test
3. **Refactor:** Improve code quality without changing behavior

Do not merge untested code.

### BDD (Behavior-Driven Development)

[Include only if features/ directory exists]

Features are written in Gherkin (`.feature` files) before implementation. Each feature describes the acceptance criteria using business language. Developers implement step definitions to satisfy the feature scenarios.

### Makefile Convention

CI/CD pipelines call **Makefile targets only**. Follow these rules:

- Simple, single-command targets belong directly in the `Makefile` — do not create a shell script just to wrap one line.
- Multi-step or complex logic (more than ~2 commands) must be extracted into a script under `scripts/`, then called from the Makefile target.
- Never inline multi-step logic directly in a `Makefile` recipe — extract it.

When in doubt: if it fits on one line and has no branching or loops, put it in the Makefile. If it needs error handling, loops, or multiple steps, make it a script.

### Quality Gates (Run Before Every Commit)

All three gates must pass before committing:

```bash
make lint       # Linting (golangci-lint, eslint, etc.)
make typecheck  # TypeScript type checking (tsc --noEmit)
make test       # Unit and integration tests
```

[Include only if E2E tests exist]

```bash
make test:e2e   # End-to-end tests (Playwright, Cypress, etc.)
```

### Languages

- **Go** for all backend code (if backend exists)
- **TypeScript** for all frontend code (if frontend exists)
- No other languages without explicit decision

### Secrets Handling

[Customize based on actual secret strategy]

- Never commit secrets to version control
- Use environment variables for local development (optional)
- Use [file mounts / ExternalSecrets / managed identities] in production
- Audit secret usage regularly

### Error Handling

[Include only if applicable based on backend language]

Errors are returned, not panicked. Always wrap errors with context:

```go
// Good
return fmt.Errorf("doing X: %w", err)

// Bad
panic(err)
return err // no context
```

### Dependency Injection

No global state. Pass all dependencies via constructors or function parameters.

## Conventions

### Go

[Include only if Go backend exists]

- Standard library preferred; minimize third-party dependencies
- Errors returned, not panicked
- `internal/` for non-exported packages
- HTTP handlers thin — business logic in service layer
- Context threading: always pass `context.Context` as first arg
- **Error wrapping:** Use `fmt.Errorf("doing X: %w", err)` — always wrap with context describing the operation
- **Dependency injection:** No global state; pass all dependencies via constructors (e.g., `NewService(db, cache)`)

### TypeScript / React

[Include only if React frontend exists]

- **Strict mode:** `"strict": true` in tsconfig; no implicit `any`
- **Functional components only** — no class components
- **Hooks:** Use React hooks for state, side effects, and context
- **Server state:** All API data fetching via TanStack Query hooks (never ad-hoc `fetch` calls)
- **Component files:** `.tsx` extension for all React components
- **File structure:** One component per file; collocate styles and tests
- **No global state managers:** Use React Context or TanStack Query cache
- **Type safety:** Export types explicitly; use discriminated unions for state machines

### API Design

[Include only if REST API exists]

- **Versioning:** All endpoints under `/api/v1/...` — increment on breaking changes
- **Error responses:** Structured JSON: `{"error": {"code": "ERROR_CODE", "message": "human readable"}}`
- **Pagination:** List endpoints support `?page=1&limit=50` with consistent response shape
- **Filtering:** Collection routes support filtering, e.g., `?status=active&assignee=alice`

### Project Layout

[Auto-generate from actual directory structure; include only existing directories]

```
[PROJECT_NAME]/
├── cmd/               # Go binaries [if backend exists]
│   └── [service]/
├── internal/          # Go internal packages [if backend exists]
│   ├── api/           # HTTP handlers, routing
│   ├── domain/        # Business entities
│   └── store/         # Repository interfaces
├── features/          # Gherkin feature files [if testing exists]
├── migrations/        # Database migrations [if applicable]
├── web/               # React + TypeScript frontend [if frontend exists]
│   ├── src/
│   │   ├── components/
│   │   ├── hooks/
│   │   ├── pages/
│   │   └── App.tsx
│   ├── vite.config.ts
│   └── tsconfig.json
├── deploy/            # Infrastructure as code [if applicable]
│   └── helm/
├── scripts/           # Shell scripts called by Makefile
├── Makefile
├── go.mod             # Go module definition [if backend exists]
├── package.json       # Node dependencies [if frontend exists]
└── README.md
```

## What Claude Should Know

- Always run `forge:add-feature` before implementing any new feature
- Every implementation must use `forge:tdd-cycle` (red/green/refactor)
- Before committing, run `forge:quality-check` to verify all gates pass
- This project has a Makefile — always check available targets before suggesting build commands
- Do not create new languages or frameworks without explicit decision in this file
- If you encounter violations of these conventions (e.g., untested code, missing error wrapping, implicit `any`), flag them immediately
```

---

### Step 3: Customize and Write

1. Replace all `[PLACEHOLDERS]` with actual project details
2. Remove sections marked `[Include only if X]` that don't apply to this project
3. Generate an accurate project layout tree using the actual directory structure
4. If this is a retrofit (project already exists), preserve any existing project-specific context and merge it

### Step 4: Verify and Commit

Run this verification:

```bash
# Verify the file exists and is valid Markdown
ls -l CLAUDE.md
head -20 CLAUDE.md

# Check for any remaining [PLACEHOLDERS]
grep -i '\[.*\]' CLAUDE.md | grep -v http
```

If grep finds unresolved placeholders, fix them.

Then commit:

```bash
git add CLAUDE.md
git commit -m "docs: generate CLAUDE.md with project context and mandatory workflows"
```

## After This Skill

The project now has:
- Clear architecture documentation
- Enforced development workflows (add-feature → tdd-cycle → quality-check)
- Conventions for all layers (backend, frontend, testing, infrastructure)
- A reference for all future Claude Code sessions

Every future feature starts with `forge:add-feature`, which will reference this CLAUDE.md for context and constraints.

## Notes for Implementation

- **Auto-detection:** Use Go/Node/directory checks to determine what layers exist
- **Preservation:** If CLAUDE.md already exists, ask before overwriting; offer to merge instead
- **Tree generation:** Use `tree` command or manual directory inspection to build the accurate layout section
- **Gherkin note:** Include only testing sections if `.feature` files are found
- **Secrets note:** Customize the "Secrets Handling" section to match the actual strategy used in the project (env vars, file mounts, managed identities, etc.)
- **Error handling:** Go-specific conventions only if Go backend exists; similarly for TypeScript/React conventions
