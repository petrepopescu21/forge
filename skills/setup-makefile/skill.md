---
name: setup-makefile
description: Create a Makefile with standard targets for build, test, lint, typecheck, dev servers, E2E, infrastructure, and quality gates. Use when bootstrapping a project or when the bootstrap-project orchestrator invokes this skill. Trigger on "set up Makefile", "create build targets", or as part of project bootstrapping.
---

# Setup Makefile

Setting up Makefile with standard build, test, lint, and infrastructure targets.

## Conventions

- **CI/CD calls Makefile targets only** — all automation goes through `make` commands
- **Single-command targets** belong directly in `Makefile` — do not create a shell script just to wrap one line
- **Multi-step targets** (more than ~2 commands, error handling, loops, branching) go into `scripts/` and are called from the Makefile
- **All targets are .PHONY** — targets do not produce files with the same name as the target

## Prerequisites

- Go 1.22 or later installed
- Bun 1.1 or higher (for web/frontend targets)
- golangci-lint installed (for lint target)
- Docker (optional, for SonarQube scanner)

## Process

### Step 1: Determine Project Name

Extract project name from the current directory or ask the user:

```bash
PROJECT_NAME=$(basename "$(pwd)")
echo "Using project name: $PROJECT_NAME"
```

If needed, ask user to confirm or provide the project name.

### Step 2: Create scripts/ Directory

```bash
mkdir -p scripts
```

### Step 3: Create Makefile

Create a `Makefile` in the project root with the following content. **Replace `<project-name>` with the actual project name determined in Step 1.**

```makefile
.PHONY: help build test lint typecheck dev-api dev-web \
        e2e e2e-web e2e-web-integration e2e-cluster e2e-db e2e-deploy \
        cluster-up cluster-deps deploy migrate sonar clean

# Variables
GO_CMD := go
WEB_DIR := web
CLUSTER := pebblr-dev
KIND_CFG := kind-config.yaml

# Color output for help
CYAN := \033[0;36m
NC := \033[0m # No Color

# Help target — grep for ## comments and display them
help:
	@echo "$(CYAN)Available targets:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(CYAN)%-30s$(NC) %s\n", $$1, $$2}'

# =============================================================================
# BUILD & PACKAGE
# =============================================================================

build: ## Build Go backend and React frontend
	@echo "Building backend..."
	$(GO_CMD) build -o bin/<project-name> ./cmd/<project-name>
	@echo "Building frontend..."
	cd $(WEB_DIR) && bun install --frozen-lockfile && bun run build
	@echo "Build complete."

# =============================================================================
# QUALITY GATES
# =============================================================================

test: ## Run all tests (backend Go + frontend)
	@echo "Running backend tests..."
	$(GO_CMD) test -v -race -coverprofile=coverage.out ./...
	@echo "Running frontend tests..."
	cd $(WEB_DIR) && bun run test
	@echo "Tests complete."

lint: ## Run linters (golangci-lint + frontend lint)
	@echo "Linting backend..."
	golangci-lint run ./...
	@echo "Linting frontend..."
	cd $(WEB_DIR) && bun run lint
	@echo "Linting complete."

typecheck: ## Run TypeScript type checking (frontend only)
	@echo "Typechecking frontend..."
	cd $(WEB_DIR) && bun run typecheck
	@echo "Typecheck complete."

# =============================================================================
# DEVELOPMENT SERVERS
# =============================================================================

dev-api: ## Run API server with hot reload (requires air or fallback to go run)
	@which air > /dev/null 2>&1 && air -c .air.toml || $(GO_CMD) run ./cmd/<project-name>/main.go

dev-web: ## Run React dev server with hot reload
	@echo "Starting React dev server..."
	cd $(WEB_DIR) && bun run dev

# =============================================================================
# E2E TESTING
# =============================================================================

e2e: ## Run all E2E tests (web, integration, cluster, db, deploy)
	@echo "Running all E2E tests..."
	$(MAKE) e2e-web
	$(MAKE) e2e-web-integration
	$(MAKE) e2e-cluster
	$(MAKE) e2e-db
	$(MAKE) e2e-deploy

e2e-web: ## Run frontend E2E tests (Playwright or similar)
	@echo "Running web E2E tests..."
	cd $(WEB_DIR) && bun run test:e2e

e2e-web-integration: ## Run web + API integration tests
	@echo "Running web integration tests..."
	cd $(WEB_DIR) && bun run test:e2e:integration

e2e-cluster: ## Run cluster integration tests (Kind + deploy)
	@echo "Running cluster E2E tests..."
	scripts/e2e-cluster.sh

e2e-db: ## Run database integration tests
	@echo "Running database E2E tests..."
	$(GO_CMD) test -v -run TestDB ./...

e2e-deploy: ## Run deployment E2E tests (Helm + AKS simulation)
	@echo "Running deployment E2E tests..."
	scripts/e2e-deploy.sh

# =============================================================================
# INFRASTRUCTURE
# =============================================================================

cluster-up: ## Start Kind cluster for local development
	@echo "Starting Kind cluster: $(CLUSTER)..."
	kind create cluster --name $(CLUSTER) --config $(KIND_CFG) || true
	kubectl cluster-info --context kind-$(CLUSTER)

cluster-deps: ## Install dependencies on the Kind cluster (metrics-server, etc.)
	@echo "Installing cluster dependencies..."
	scripts/cluster-deps.sh

deploy: ## Deploy to Kind cluster using Helm
	@echo "Deploying to Kind cluster..."
	scripts/deploy.sh

migrate: ## Run database migrations
	@echo "Running database migrations..."
	scripts/migrate.sh

# =============================================================================
# CODE QUALITY & SCANNING
# =============================================================================

sonar: ## Run SonarQube scanner (requires SONAR_HOST_URL and SONAR_LOGIN env vars)
	@echo "Running SonarQube analysis..."
	docker run --rm \
		-e SONAR_HOST_URL=$(SONAR_HOST_URL) \
		-e SONAR_LOGIN=$(SONAR_LOGIN) \
		-e SONAR_PROJECTKEY=<project-name> \
		-v "$(PWD):/src" \
		sonarsource/sonar-scanner-cli

# =============================================================================
# CLEANUP
# =============================================================================

clean: ## Remove build artifacts and caches
	@echo "Cleaning up..."
	rm -rf bin/
	rm -rf $(WEB_DIR)/dist/
	rm -rf $(WEB_DIR)/node_modules/
	rm -f .vite/
	rm -f coverage.out
	@echo "Cleanup complete."
```

### Step 4: Create scripts/ Placeholder Scripts

Create placeholder scripts for multi-step targets. These will be invoked by the Makefile.

Create `scripts/e2e-cluster.sh`:

```bash
#!/bin/bash
set -euo pipefail

echo "E2E Cluster tests would run here..."
# Implement cluster E2E tests
```

Create `scripts/e2e-deploy.sh`:

```bash
#!/bin/bash
set -euo pipefail

echo "E2E Deployment tests would run here..."
# Implement deployment E2E tests
```

Create `scripts/cluster-deps.sh`:

```bash
#!/bin/bash
set -euo pipefail

echo "Installing cluster dependencies..."
# Install metrics-server, cert-manager, ingress-nginx, etc.
```

Create `scripts/deploy.sh`:

```bash
#!/bin/bash
set -euo pipefail

echo "Deploying via Helm..."
RELEASE_NAME="<project-name>"
NAMESPACE="default"

helm upgrade --install "$RELEASE_NAME" ./deploy/helm/<project-name> \
  --namespace "$NAMESPACE" \
  --create-namespace
```

Create `scripts/migrate.sh`:

```bash
#!/bin/bash
set -euo pipefail

echo "Running database migrations..."
# Use golang-migrate or custom migration tool
migrate -path ./migrations -database "$DATABASE_URL" up
```

Make all scripts executable:

```bash
chmod +x scripts/*.sh
```

### Step 5: Verify Makefile

Run the help target to verify the Makefile is working:

```bash
make help
```

Expected output: A formatted list of all available targets with descriptions.

Test basic targets:

```bash
make clean
make build
make test
make lint
make typecheck
```

All targets must run without errors (or complete gracefully if dependencies are not installed).

### Step 6: Commit

```bash
git add Makefile scripts/
git commit -m "build: create Makefile with standard targets for build, test, lint, dev, E2E, and infrastructure"
```

## Summary

The Makefile is now set up with the following target groups:

- **help** — Display all available targets
- **build** — Compile Go binary and bundle React frontend
- **test** — Run all unit and integration tests
- **lint** — Run linters on backend and frontend
- **typecheck** — Run TypeScript type checker
- **dev-api** — Start API server with hot reload
- **dev-web** — Start React dev server with hot reload
- **e2e** — Run all E2E tests (web, integration, cluster, db, deploy)
- **cluster-up** — Spin up local Kind cluster
- **cluster-deps** — Install dependencies on cluster
- **deploy** — Deploy via Helm
- **migrate** — Run database migrations
- **sonar** — Run SonarQube analysis
- **clean** — Remove build artifacts and caches

All multi-step operations are scaffolded in `scripts/` and called from the Makefile, maintaining separation of concerns and making the build system maintainable.
