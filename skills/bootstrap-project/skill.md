---
name: bootstrap-project
description: Scaffold a complete Go + React/TypeScript project with all layers — module, frontend, Makefile, linting, BDD, Playwright, CI/CD, SonarCloud, Helm, and CLAUDE.md. Use when starting a new project from scratch. Trigger on "new project", "bootstrap project", "start a project", "scaffold project", "create project", or "init project". This is the top-level orchestrator that invokes all setup skills.
---

# Bootstrap Project

Bootstrapping a new project — I'll set up all layers and generate CLAUDE.md.

## Prerequisites

- Go 1.22 or later
- Bun 1.1 or higher
- Git installed
- Docker (optional, for SonarQube scanner)
- kubectl and Kind (optional, for Helm/K8s support)

## Process

### Step 1: Gather Project Information

Collect the following information from the user:

1. **Project name** (e.g., `myproject`)
   - Used for directory structure, binaries, Helm charts
   - Format: lowercase alphanumeric and hyphens only

2. **Go module path** (e.g., `github.com/petrepopescu21/myproject`)
   - Used for `go.mod` and imports

3. **One-liner description** (e.g., "A self-hosted CRM for field sales")
   - Used in README and project metadata

Example prompt:
```
Enter project name: myproject
Enter Go module path (e.g., github.com/petrepopescu21/myproject): github.com/petrepopescu21/myproject
Enter one-liner description: A self-hosted CRM for field sales
```

### Step 2: Select Layers

Present a checklist of optional layers. Default all to yes; user can skip by entering `n`:

```
Select layers to scaffold:
[ ] Go backend              (default: yes)
[ ] React + TypeScript      (default: yes)
[ ] Helm + Kubernetes       (default: yes)
[ ] BDD + Gherkin scenarios (default: yes)
[ ] SonarCloud integration  (default: yes)

Enter your selections (y/n for each), or press Enter to accept defaults.
```

Store selections for conditional skill invocation in Step 4.

### Step 3: Initialize Git Repository

```bash
git init
git config user.name "Claude"
git config user.email "claude@forge.local"
```

Create an initial empty commit to establish main branch:

```bash
git commit --allow-empty -m "initial: project bootstrap"
```

### Step 4: Invoke Setup Skills in Order

Invoke the following skills sequentially based on layer selections from Step 2. All skills should be invoked with the project name and module path as parameters.

| Order | Skill | Condition | Purpose |
|-------|-------|-----------|---------|
| 1 | `forge:setup-go-module` | Go backend selected | Initialize Go module with `cmd/`, `internal/` structure |
| 2 | `forge:setup-react` | React selected | Bootstrap React + TypeScript frontend with Vite, Bun, TanStack |
| 3 | `forge:setup-makefile` | Always | Create Makefile with build, test, lint, dev, E2E, infrastructure targets |
| 4 | `forge:setup-linting` | Always | Configure golangci-lint + frontend linting (ESLint, Prettier, etc.) |
| 5 | `forge:setup-bdd` | BDD selected | Create BDD scenario structure and Cucumber/Gherkin setup |
| 6 | `forge:setup-playwright` | React selected | Add Playwright E2E test suite for frontend |
| 7 | `forge:setup-ci` | Always | Create GitHub Actions / CI/CD pipelines (lint, test, build, deploy) |
| 8 | `forge:setup-sonar` | SonarCloud selected | Add SonarCloud project configuration and scanning steps |
| 9 | `forge:setup-helm` | Helm selected | Create Helm 4 chart for Kubernetes deployment |
| 10 | `forge:generate-claude-md` | Always (last) | Generate project-specific CLAUDE.md with conventions and architecture |

### Step 5: Create Supporting Files

#### .gitignore

Create `.gitignore` in the project root with the following content:

```
# Go
bin/
dist/
*.o
*.a
*.so
.DS_Store
.vscode/
.idea/

# Node/Frontend
web/node_modules/
web/dist/
web/.env.local
web/.env.*.local
web/coverage/
web/.nyc_output/
web/pnpm-lock.yaml
web/yarn.lock

# Bun
bun.lockb

# Build artifacts
*.exe
*.dll
*.dylib

# IDE
.vscode/
.idea/
*.swp
*.swo
*~
.DS_Store

# Environment
.env
.env.local
.env.*.local
/secrets/
.envrc

# Test coverage
coverage.out
coverage/
.nyc_output/

# Temporary files
tmp/
temp/
*.log
.air.toml

# Kubernetes/Helm
kubeconfig
kind-config.yaml

# SonarQube
.scannerwork/
sonar-project.properties.local
```

#### Dockerfile

Create `Dockerfile` in the project root with a multi-stage build:

```dockerfile
# Stage 1: Go Builder
FROM golang:1.22-alpine AS go-builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git

# Copy Go module files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /build/bin/app ./cmd/$(basename $(pwd))

# Stage 2: Web/Node Builder
FROM oven/bun:latest AS web-builder

WORKDIR /web

# Copy frontend files
COPY web/package.json web/bun.lockb ./

# Install dependencies
RUN bun install --frozen-lockfile

# Copy source
COPY web .

# Build frontend
RUN bun run build

# Stage 3: Runtime (Distroless)
FROM gcr.io/distroless/base-debian12

# Copy binary from go-builder
COPY --from=go-builder /build/bin/app /app

# Copy web assets from web-builder
COPY --from=web-builder /web/dist /web/dist

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD ["/app", "-health"]

# Expose port
EXPOSE 8080

# Run application
ENTRYPOINT ["/app"]
```

Make the Dockerfile executable and verifiable:

```bash
docker build -t myproject:latest .
docker run --rm myproject:latest --version
```

### Step 6: Verify Quality Gates

Run all quality gates to ensure everything is set up correctly:

```bash
make lint
```

Expected output:
- Backend linters (golangci-lint) pass with no errors
- Frontend linters (ESLint, Prettier) pass

```bash
make typecheck
```

Expected output:
- TypeScript strict mode passes with no errors

```bash
make test
```

Expected output:
- All unit and integration tests pass
- Coverage report generated (if applicable)

**Do not proceed to Step 7 if any gate fails.** Fix issues and re-run.

### Step 7: Initial Git Commits

Commit the scaffolding in logical groups:

```bash
# Commit backend structure
git add cmd/ internal/ go.mod go.sum
git commit -m "scaffold: initialize Go module with cmd/ and internal/ structure"

# Commit frontend structure (if React selected)
git add web/
git commit -m "scaffold: initialize React + TypeScript frontend with Vite and Bun"

# Commit build system
git add Makefile scripts/
git commit -m "build: create Makefile with standard targets"

# Commit linting and quality
git add .golangci.yml .eslintrc.* .prettierrc* vitest.config.ts
git commit -m "quality: configure linters and type checking"

# Commit BDD (if selected)
git add features/ cucumber.js
git commit -m "test: add BDD scenario structure"

# Commit Playwright (if React selected)
git add e2e/ playwright.config.ts
git commit -m "test: add Playwright E2E suite"

# Commit CI/CD
git add .github/workflows/
git commit -m "ci: add GitHub Actions pipelines"

# Commit Helm chart (if selected)
git add deploy/helm/
git commit -m "ops: add Helm 4 chart for Kubernetes deployment"

# Commit infrastructure files
git add Dockerfile .gitignore CLAUDE.md README.md
git commit -m "docs: add Dockerfile, .gitignore, CLAUDE.md, and README"
```

### Step 8: Project Summary

Announce completion with a summary of what was scaffolded:

```
✓ Bootstrap complete!

Project: myproject
Module:  github.com/petrepopescu21/myproject
Layers:  Go backend, React frontend, Helm/K8s, BDD, SonarCloud

Scaffolded:
  [✓] Go backend (cmd/, internal/, handlers, domain models)
  [✓] React frontend (web/, Vite, TanStack Query/Table, strict TypeScript)
  [✓] Makefile (build, test, lint, dev, E2E, cluster, sonar targets)
  [✓] Linting (golangci-lint, ESLint, Prettier, TypeScript strict)
  [✓] BDD (Gherkin scenarios, step definitions, Cucumber integration)
  [✓] Playwright (E2E browser testing, API integration tests)
  [✓] CI/CD (GitHub Actions: lint, test, build, deploy)
  [✓] SonarCloud (code quality scanning and reporting)
  [✓] Helm 4 (Kubernetes manifests for deployment)
  [✓] CLAUDE.md (project conventions, architecture, decisions)
  [✓] .gitignore (Go, Node, IDE, secrets, coverage)
  [✓] Dockerfile (multi-stage: Go builder, web builder, distroless runtime)

Quality gates: PASS (lint, typecheck, test all passing)

Next steps:
  • Run: make help
      Display all available Makefile targets
  • Run: make dev-api
      Start the API server with hot reload
  • Run: make dev-web
      Start the React dev server with hot reload
  • Use: forge:add-feature
      Begin adding your first feature with BDD + TDD

Reminder:
  CLAUDE.md has been generated with your project's conventions, architecture,
  and development practices. This file is your source of truth — all future
  work should adhere to the patterns defined here. Update CLAUDE.md as you
  learn and evolve the architecture.
```

### Step 9: Documentation

Create or update `README.md` with project overview:

```markdown
# myproject

A self-hosted CRM for field sales.

## Quick Start

### Prerequisites

- Go 1.22+
- Bun 1.1+
- Docker (optional, for containerized builds)
- kubectl + Kind (optional, for local K8s development)

### Development

Clone the repository:

\`\`\`bash
git clone https://github.com/petrepopescu21/myproject.git
cd myproject
\`\`\`

Install dependencies and start development servers:

\`\`\`bash
# Backend API (http://localhost:8080)
make dev-api

# Frontend (http://localhost:5174)
make dev-web
\`\`\`

Run quality gates:

\`\`\`bash
make lint       # Linters
make typecheck  # TypeScript
make test       # Unit + integration tests
\`\`\`

Run E2E tests:

\`\`\`bash
make e2e-web              # Playwright tests
make e2e-web-integration  # Web + API integration
\`\`\`

### Production Deployment

Build and deploy to AKS using Helm:

\`\`\`bash
make build
docker build -t myproject:latest .
docker push myproject:latest

helm upgrade --install myproject ./deploy/helm/myproject \
  --namespace default \
  --create-namespace
\`\`\`

## Architecture

See [CLAUDE.md](./CLAUDE.md) for:

- Project conventions (Go, TypeScript, Kubernetes)
- Development practices (TDD, BDD, Makefile)
- Database and auth architecture
- RBAC and security model
- Helm and deployment strategy

## Contributing

1. Write BDD scenarios first (`.feature` files)
2. Implement with TDD (red/green/refactor)
3. Ensure all quality gates pass
4. Commit with clear messages (one feature per commit)

See CLAUDE.md for full development workflow.

## License

[Add your license here]
```

## Summary

Bootstrap is complete! The project now has:

- **Backend:** Go module with clean architecture (`cmd/`, `internal/`, handlers, domain)
- **Frontend:** React + TypeScript with Vite, Bun, TanStack Query/Table, strict types
- **Build system:** Makefile with targets for build, test, lint, dev, E2E, cluster, deploy, sonar
- **Linting:** golangci-lint, ESLint, Prettier, TypeScript strict mode
- **Testing:** BDD (Gherkin), Vitest (unit), Playwright (E2E)
- **CI/CD:** GitHub Actions pipelines for lint, test, build, deploy
- **Code quality:** SonarCloud integration for static analysis
- **Infrastructure:** Helm 4 chart for Kubernetes deployment
- **Documentation:** CLAUDE.md with conventions, architecture, and development practices
- **Secrets & ignore:** `.gitignore` and Dockerfile with best practices

All quality gates pass. Next step: use `forge:add-feature` to begin feature development with BDD + TDD.

CLAUDE.md is your enforcement layer — future work must follow the patterns defined there.
