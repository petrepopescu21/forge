---
name: setup-ci
description: Create consolidated GitHub Actions workflows — one CI pipeline (lint, test, coverage, E2E, SonarCloud), one deploy pipeline, and Dependabot auto-merge. Use when bootstrapping a project or adding CI/CD. Trigger on "set up CI", "add GitHub Actions", "configure CI/CD", "add pipelines", or as part of project bootstrapping.
---

# Setup CI

Setting up GitHub Actions — consolidated CI, deploy, and Dependabot auto-merge.

## Why Consolidated?

A single source of truth for "is this green" — lint, test, coverage, E2E, and SonarCloud analysis all in one pipeline. Clear status at a glance; fast feedback loop.

## Workflows

This creates three workflows:

### 1. ci.yml — Consolidated Pipeline

Runs on: `push` to `main`/`develop`, `pull_request` to `main`, `workflow_call`

Four parallel jobs:

1. **lint** — Go, frontend, and Helm linting
2. **test** — Unit tests with coverage (Go + TypeScript)
3. **e2e** — Kind cluster, Tilt CI, end-to-end tests
4. **sonar** — SonarCloud analysis (depends on test)

### 2. deploy.yml — Build and Deploy

Runs on: `workflow_run` from `ci.yml` (on main)

Concurrency group prevents simultaneous deployments.

Jobs:

1. **build-and-push** — Docker build, push to ACR with SHA tag
2. **deploy** — Helm upgrade, rollout verification

### 3. dependabot-auto-merge.yml — Automatic Merging

Runs on: `pull_request_target` from `dependabot[bot]`

Auto-merges non-major dependency updates using squash merge.

---

## Workflow Files

Create these files in `.github/workflows/`:

### `.github/workflows/ci.yml`

```yaml
name: CI

on:
  push:
    branches:
      - main
      - develop
  pull_request:
    branches:
      - main
  workflow_call:

concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - name: Lint Go
        uses: golangci/golangci-lint-action@v4
        with:
          version: latest

      - name: Set up Bun
        uses: oven-sh/setup-bun@v1

      - name: Install frontend dependencies
        run: cd web && bun install --frozen-lockfile

      - name: Lint frontend
        run: cd web && bun run lint

      - name: Lint Helm charts
        run: helm lint deploy/helm/pebblr

  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - name: Set up Bun
        uses: oven-sh/setup-bun@v1

      - name: Run Go tests with coverage
        run: go test -coverprofile=coverage.out ./...

      - name: Install frontend dependencies
        run: cd web && bun install --frozen-lockfile

      - name: Run frontend tests with coverage
        run: cd web && bun run test -- --coverage

      - name: TypeScript type check
        run: cd web && bun run typecheck

      - name: Upload coverage artifacts
        uses: actions/upload-artifact@v4
        with:
          name: coverage-reports
          path: |
            coverage.out
            web/coverage

  e2e:
    runs-on: ubuntu-latest
    timeout-minutes: 20
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - name: Set up Bun
        uses: oven-sh/setup-bun@v1

      - name: Install Kind v0.27.0
        run: |
          curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.27.0/kind-linux-amd64
          chmod +x ./kind
          sudo mv ./kind /usr/local/bin/kind

      - name: Install Tilt
        run: |
          curl -fsSL https://raw.githubusercontent.com/tilt-dev/tilt/master/scripts/install.sh | bash

      - name: Create Kind cluster
        run: make e2e-cluster

      - name: Setup test database
        run: make e2e-db

      - name: Deploy services
        run: make e2e-deploy

      - name: Run backend E2E tests
        run: make e2e

      - name: Install Playwright
        run: cd web && bun install --frozen-lockfile && bun add -d @playwright/test

      - name: Run frontend E2E tests
        run: cd web && make e2e-web-integration

      - name: Debug on failure
        if: failure()
        run: |
          echo "=== Kind Cluster Status ==="
          kubectl get nodes
          kubectl get pods --all-namespaces
          kubectl logs -n pebblr --all-containers=true --tail=50

  sonar:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Download coverage artifacts
        uses: actions/download-artifact@v4
        with:
          name: coverage-reports

      - name: SonarCloud Scan
        uses: SonarSource/sonarqube-scan-action@v6
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          SONAR_TOKEN: ${{ secrets.SONAR_TOKEN }}
```

### `.github/workflows/deploy.yml`

```yaml
name: Deploy

on:
  workflow_run:
    workflows:
      - CI
    types:
      - completed
    branches:
      - main

concurrency:
  group: deploy-${{ github.ref }}
  cancel-in-progress: false

jobs:
  build-and-push:
    if: github.event.workflow_run.conclusion == 'success'
    runs-on: ubuntu-latest
    outputs:
      image-tag: ${{ steps.meta.outputs.tags }}
    steps:
      - uses: actions/checkout@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Log in to Azure Container Registry
        run: |
          az login --service-principal \
            -u ${{ secrets.ACR_CLIENT_ID }} \
            -p ${{ secrets.ACR_CLIENT_SECRET }} \
            --tenant ${{ secrets.ACR_TENANT_ID }}
          az acr login --name ${{ secrets.ACR_NAME }}

      - name: Generate image tag
        id: meta
        run: |
          TAG="${{ secrets.ACR_NAME }}.azurecr.io/pebblr:${{ github.sha }}"
          echo "tags=${TAG}" >> $GITHUB_OUTPUT

      - name: Build and push Docker image
        uses: docker/build-push-action@v5
        with:
          context: .
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          cache-from: type=registry,ref=${{ secrets.ACR_NAME }}.azurecr.io/pebblr:buildcache
          cache-to: type=registry,ref=${{ secrets.ACR_NAME }}.azurecr.io/pebblr:buildcache,mode=max

  deploy:
    needs: build-and-push
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up kubectl
        uses: azure/setup-kubectl@v3

      - name: Authenticate with AKS
        run: |
          az login --service-principal \
            -u ${{ secrets.AKS_CLIENT_ID }} \
            -p ${{ secrets.AKS_CLIENT_SECRET }} \
            --tenant ${{ secrets.AKS_TENANT_ID }}
          az aks get-credentials \
            --resource-group ${{ secrets.AKS_RESOURCE_GROUP }} \
            --name ${{ secrets.AKS_CLUSTER_NAME }}

      - name: Set up Helm
        uses: azure/setup-helm@v3

      - name: Helm upgrade
        run: |
          helm upgrade \
            --install pebblr deploy/helm/pebblr \
            --namespace pebblr \
            --create-namespace \
            --values deploy/helm/pebblr/values.yaml \
            --set image.tag=${{ github.sha }}

      - name: Verify rollout
        run: |
          kubectl rollout status deployment/pebblr-api \
            --namespace pebblr \
            --timeout=5m
          kubectl rollout status deployment/pebblr-web \
            --namespace pebblr \
            --timeout=5m
```

### `.github/workflows/dependabot-auto-merge.yml`

```yaml
name: Dependabot Auto-Merge

on:
  pull_request_target:
    types:
      - opened
      - synchronize

permissions:
  pull-requests: write
  contents: write

jobs:
  auto-merge:
    if: github.actor == 'dependabot[bot]'
    runs-on: ubuntu-latest
    steps:
      - name: Fetch Dependabot metadata
        id: dependabot-metadata
        uses: dependabot/fetch-metadata@v1

      - name: Auto-merge for non-major updates
        if: steps.dependabot-metadata.outputs.update-type != 'version-update:semver-major'
        run: |
          gh pr merge --auto --squash "${{ github.event.pull_request.number }}"
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

---

## Next Steps

1. Copy these three workflow files to `.github/workflows/`
2. Set GitHub Actions secrets:
   - `ACR_CLIENT_ID`, `ACR_CLIENT_SECRET`, `ACR_TENANT_ID`, `ACR_NAME` (Azure Container Registry)
   - `AKS_CLIENT_ID`, `AKS_CLIENT_SECRET`, `AKS_TENANT_ID`, `AKS_RESOURCE_GROUP`, `AKS_CLUSTER_NAME` (AKS)
   - `SONAR_TOKEN` (SonarCloud)
3. Verify Makefile targets: `make lint`, `make test`, `make e2e`, `make e2e-cluster`, `make e2e-db`, `make e2e-deploy`
4. Test by pushing to `develop` or opening a PR to `main`

