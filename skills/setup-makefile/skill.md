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

Replace all `<project-name>` placeholders with the actual project name.

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
# Blocks destructive/local-only targets from running against an AKS cluster.
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

e2e: ## Run Go E2E tests against a running cluster
	@go test -v -tags=e2e -count=1 -timeout=10m ./e2e/...

e2e-web: ## Run Playwright E2E tests (starts Vite dev server automatically)
	@cd $(WEB_DIR) && bun run test:e2e

e2e-web-integration: ## Run Playwright integration tests against Kind cluster
	@scripts/e2e-web.sh

e2e-cluster: ## Create Kind cluster for E2E (lightweight, no extras)
	$(AKS_GUARD)
	@kind create cluster --name $(CLUSTER) --config $(KIND_CFG) --wait 120s

e2e-db: ## Deploy PostgreSQL, run migrations, seed data for E2E namespace
	$(AKS_GUARD)
	@scripts/cluster-db.sh <project-name>-e2e

e2e-deploy: ## Build, load images, and deploy to E2E namespace via Tilt CI
	$(AKS_GUARD)
	@tilt ci -- --namespace <project-name>-e2e

cluster-up: ## Recreate local Kind cluster with all dependencies
	$(AKS_GUARD)
	@kind delete cluster --name $(CLUSTER) 2>/dev/null || true
	@kind create cluster --name $(CLUSTER) --config $(KIND_CFG)
	@cloud-provider-kind &
	@$(MAKE) cluster-deps

cluster-deps: ## Install cert-manager, ESO, and other cluster dependencies (idempotent)
	$(AKS_GUARD)
	@scripts/cluster-deps.sh

deploy: ## Start Tilt for local development (interactive mode)
	$(AKS_GUARD)
	@tilt up

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

### Step 3: Create scripts/ Directory

```bash
mkdir -p scripts
```

The scripts referenced by Makefile targets (`cluster-db.sh`, `cluster-deps.sh`, `e2e-web.sh`) are created by the `setup-helm` and `setup-playwright` skills. This skill does not create them — it only references them. If invoked standalone without those skills, create stub scripts that explain what's needed:

```bash
#!/usr/bin/env bash
echo "ERROR: This script is created by forge:setup-helm. Run that skill first."
exit 1
```

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

## Key Design Points

- **AKS safety guard** on all destructive/local-only targets prevents accidentally running `cluster-up` or `deploy` against a production AKS cluster
- **Tilt** is the development/deployment tool for local Kind clusters — `deploy` runs `tilt up` (interactive), `e2e-deploy` runs `tilt ci` (headless)
- **cloud-provider-kind** runs alongside Kind to enable LoadBalancer services locally — started by `cluster-up`
- **Scripts are owned by other skills** — `setup-helm` creates `cluster-db.sh`, `cluster-deps.sh`; `setup-playwright` creates `e2e-web.sh`
- **E2E targets** are split: `e2e` runs Go E2E tests, `e2e-web` runs Playwright with dev server, `e2e-web-integration` runs Playwright against a live cluster
