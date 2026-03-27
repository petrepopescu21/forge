---
name: setup-makefile
description: Create a Makefile with standard targets for build, test, lint, typecheck, dev servers, E2E, infrastructure, and quality gates. All external tools are version-locked and installed into bin/ (no global dependencies). Use when bootstrapping a project or when the bootstrap-project orchestrator invokes this skill. Trigger on "set up Makefile", "create build targets", or as part of project bootstrapping.
---

# Setup Makefile

Creates a Makefile with all standard targets for Go + React/TypeScript projects.

**Announce:** "Setting up Makefile with standard build, test, lint, and infrastructure targets."

## Conventions

- CI/CD pipelines call Makefile targets only — never raw commands
- Single-command targets go directly in the Makefile
- Multi-step or complex logic (2+ commands with branching/loops) extracts to `scripts/`
- All targets are `.PHONY`
- **All external tools are vendored into `bin/`** — version-locked, project-scoped, no global installs. Tool targets use `bin/<tool>` as a file target so Make only downloads once.

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

# ── Pinned tool versions ─────────────────────────────────────────────────────
KIND_VERSION           := v0.27.0
TILT_VERSION           := 0.33.22
HELM_VERSION           := v3.17.3
CLOUD_PROVIDER_KIND_VERSION := v0.6.0
GOLANGCI_LINT_VERSION  := v2.1.6

# ── Paths ─────────────────────────────────────────────────────────────────────
BIN        := $(CURDIR)/bin
GO_CMD     := cmd/<project-name>
WEB_DIR    := web
CLUSTER    := <project-name>-local
KIND_CFG   := deploy/kind/kind-config.yaml

# ── Tool binaries (project-local, never global) ──────────────────────────────
KIND                := $(BIN)/kind
TILT                := $(BIN)/tilt
HELM                := $(BIN)/helm
CLOUD_PROVIDER_KIND := $(BIN)/cloud-provider-kind
GOLANGCI_LINT       := $(BIN)/golangci-lint

# ── Platform detection ────────────────────────────────────────────────────────
OS   := $(shell uname -s | tr '[:upper:]' '[:lower:]')
ARCH := $(shell uname -m)
ifeq ($(ARCH),x86_64)
  ARCH := amd64
endif
ifeq ($(ARCH),aarch64)
  ARCH := arm64
endif

# ── AKS safety guard ─────────────────────────────────────────────────────────
# Blocks destructive/local-only targets from running against an AKS cluster.
AKS_GUARD := @if kubectl get nodes -o jsonpath='{.items[*].metadata.name}' 2>/dev/null | grep -q 'aks-'; then echo 'ERROR: Refusing to run against AKS cluster. This target is for local Kind only.'; exit 1; fi

# ══════════════════════════════════════════════════════════════════════════════
# Tool installation targets (file targets — Make skips if binary exists)
# ══════════════════════════════════════════════════════════════════════════════

$(KIND):
	@mkdir -p $(BIN)
	@echo "Installing kind $(KIND_VERSION)..."
	@curl -sSLo $(KIND) "https://kind.sigs.k8s.io/dl/$(KIND_VERSION)/kind-$(OS)-$(ARCH)"
	@chmod +x $(KIND)

$(TILT):
	@mkdir -p $(BIN)
	@echo "Installing tilt $(TILT_VERSION)..."
	@curl -sSL "https://github.com/tilt-dev/tilt/releases/download/v$(TILT_VERSION)/tilt.$(TILT_VERSION).$(OS).$(shell uname -m).tar.gz" | tar xz -C $(BIN) tilt
	@chmod +x $(TILT)

$(HELM):
	@mkdir -p $(BIN)
	@echo "Installing helm $(HELM_VERSION)..."
	@curl -sSL "https://get.helm.sh/helm-$(HELM_VERSION)-$(OS)-$(ARCH).tar.gz" | tar xz --strip-components=1 -C $(BIN) $(OS)-$(ARCH)/helm
	@chmod +x $(HELM)

$(CLOUD_PROVIDER_KIND):
	@mkdir -p $(BIN)
	@echo "Installing cloud-provider-kind $(CLOUD_PROVIDER_KIND_VERSION)..."
	@GOBIN=$(BIN) go install sigs.k8s.io/cloud-provider-kind@$(CLOUD_PROVIDER_KIND_VERSION)

$(GOLANGCI_LINT):
	@mkdir -p $(BIN)
	@echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION)..."
	@GOBIN=$(BIN) go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

# ══════════════════════════════════════════════════════════════════════════════
# Main targets
# ══════════════════════════════════════════════════════════════════════════════

.PHONY: help build test lint typecheck dev-api dev-web clean clean-tools \
        e2e e2e-web e2e-web-integration e2e-cluster e2e-db e2e-deploy \
        cluster-up cluster-deps deploy migrate sonar helm-lint

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*##' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' | sort

build: ## Build Go binary and React frontend
	@go build -o $(BIN)/<project-name> ./$(GO_CMD)
	@cd $(WEB_DIR) && bun install --frozen-lockfile && bun run build

test: ## Run Go tests and frontend tests
	@go test ./...
	@cd $(WEB_DIR) && bun run test

lint: $(GOLANGCI_LINT) ## Run golangci-lint and ESLint
	@$(GOLANGCI_LINT) run ./...
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

e2e-cluster: $(KIND) ## Create Kind cluster for E2E (lightweight, no extras)
	$(AKS_GUARD)
	@$(KIND) create cluster --name $(CLUSTER) --config $(KIND_CFG) --wait 120s

e2e-db: ## Deploy PostgreSQL, run migrations, seed data for E2E namespace
	$(AKS_GUARD)
	@scripts/cluster-db.sh <project-name>-e2e

e2e-deploy: $(TILT) ## Build, load images, and deploy to E2E namespace via Tilt CI
	$(AKS_GUARD)
	@$(TILT) ci -- --namespace <project-name>-e2e

cluster-up: $(KIND) $(CLOUD_PROVIDER_KIND) ## Recreate local Kind cluster with all dependencies
	$(AKS_GUARD)
	@$(KIND) delete cluster --name $(CLUSTER) 2>/dev/null || true
	@$(KIND) create cluster --name $(CLUSTER) --config $(KIND_CFG)
	@nohup sudo $(CLOUD_PROVIDER_KIND) > /dev/null 2>&1 &
	@$(MAKE) cluster-deps

cluster-deps: $(HELM) ## Install cert-manager, ESO, and other cluster dependencies (idempotent)
	$(AKS_GUARD)
	@scripts/cluster-deps.sh

deploy: $(TILT) ## Start Tilt for local development (interactive mode)
	$(AKS_GUARD)
	@$(TILT) up

migrate: ## Run database migrations
	$(AKS_GUARD)
	@go run ./cmd/migrate

helm-lint: $(HELM) ## Validate Helm chart
	@$(HELM) lint deploy/helm/*/

sonar: ## Run SonarCloud analysis locally
	@docker run --rm --network=host -v $(CURDIR):/usr/src -w /usr/src \
		sonarsource/sonar-scanner-cli \
		-Dsonar.host.url=https://sonarcloud.io \
		-Dsonar.token=$${SONAR_TOKEN:?Set SONAR_TOKEN}

clean: ## Clean build artifacts (preserves bin/ tools)
	@rm -rf $(BIN)/<project-name> web/dist/ web/node_modules/.vite

clean-tools: ## Remove all downloaded tools from bin/
	@rm -f $(KIND) $(TILT) $(HELM) $(CLOUD_PROVIDER_KIND) $(GOLANGCI_LINT)
```

### Step 3: Update .gitignore

Ensure `bin/` is gitignored (tools are downloaded per-machine, not committed):

```
# Project-local tools (downloaded by make)
bin/
```

### Step 4: Create scripts/ Directory

```bash
mkdir -p scripts
```

The scripts referenced by Makefile targets (`cluster-db.sh`, `cluster-deps.sh`, `e2e-web.sh`) are created by the `setup-helm` and `setup-playwright` skills. This skill does not create them — it only references them. If invoked standalone without those skills, create stub scripts that explain what's needed:

```bash
#!/usr/bin/env bash
echo "ERROR: This script is created by forge:setup-helm. Run that skill first."
exit 1
```

Note: `scripts/cluster-deps.sh` should use `$(BIN)/helm` or the `HELM` env var rather than assuming a global `helm` binary. The Makefile exports `PATH=$(BIN):$(PATH)` implicitly through the tool variables, but scripts should be aware that tools live in `bin/`.

### Step 5: Verify

```bash
make help
```

Should list all targets with descriptions. Running any target that needs a tool (e.g., `make lint`) will auto-download it on first run.

### Step 6: Commit

```bash
git add Makefile scripts/ .gitignore
git commit -m "feat: add Makefile with vendored tools in bin/ and standard targets"
```

## Key Design Points

- **All tools vendored in `bin/`** — Kind, Tilt, Helm, cloud-provider-kind, golangci-lint are downloaded once into `bin/` with pinned versions. No global installs required. Make's file-target mechanism skips the download if the binary already exists.
- **Version-locked** — tool versions are pinned at the top of the Makefile. Update a version → delete the old binary (or `make clean-tools`) → next run downloads the new version.
- **Platform-aware** — OS and ARCH are auto-detected for download URLs (macOS/Linux, amd64/arm64).
- **AKS safety guard** on all destructive/local-only targets prevents accidentally running against a production cluster.
- **Tilt** for local development — `deploy` runs `tilt up` (interactive), `e2e-deploy` runs `tilt ci` (headless).
- **cloud-provider-kind** enables LoadBalancer services in Kind — started by `cluster-up` with `sudo` (needed on macOS for privileged ports).
- **`clean` preserves tools** — only removes build artifacts. Use `clean-tools` to remove downloaded binaries.
- **Scripts are owned by other skills** — `setup-helm` creates `cluster-db.sh`, `cluster-deps.sh`; `setup-playwright` creates `e2e-web.sh`.
