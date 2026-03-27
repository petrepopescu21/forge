---
name: setup-sonar
description: Configure SonarCloud with sonar-project.properties, coverage wiring for Go and TypeScript, and a local scan Makefile target. Use when bootstrapping a project or adding code quality analysis. Trigger on "set up SonarCloud", "add Sonar", "configure code quality", or as part of project bootstrapping.
---

# Setup SonarCloud

Announce: "Setting up SonarCloud with Go and TypeScript coverage wiring."

## Prerequisites

- SonarCloud organization and project already created
- `SONAR_TOKEN` secret configured in GitHub (or available locally for `make sonar`)
- Go 1.22+ installed
- Bun 1.1+ installed
- `sonar-scanner` CLI installed (optional for local scans; Docker image can be used instead)

## Process

### Step 1: Gather SonarCloud Configuration

Ask the user for the following (if not already provided):

- **SonarCloud Organization Key** (e.g., `my-org`)
- **SonarCloud Project Key** (e.g., `my-org-pebblr`)

Example prompts:

```
What is your SonarCloud organization key?
What is your SonarCloud project key?
```

### Step 2: Create sonar-project.properties

Create `sonar-project.properties` in the project root with the following content. **Replace `<org>` and `<project>` with the values from Step 1.**

```properties
# SonarCloud Project Configuration
sonar.projectKey=<org>_<project>
sonar.organization=<org>
sonar.projectName=<project>
sonar.projectVersion=1.0

# Source and Test Directories
sonar.sources=cmd,internal,web/src
sonar.tests=internal,web/src
sonar.test.inclusions=**/*_test.go,**/*.test.ts,**/*.test.tsx,**/*.spec.ts,**/*.spec.tsx

# Exclusions
sonar.exclusions=**/node_modules/**,**/dist/**,**/vendor/**,**/*.gen.go

# Coverage Reports
sonar.go.coverage.reportPath=coverage.out
sonar.typescript.lcov.reportPaths=web/coverage/lcov.info

# Encoding
sonar.sourceEncoding=UTF-8
```

### Step 3: Verify Coverage Generation

#### Go Coverage

Run the Go test command to generate `coverage.out`:

```bash
go test -coverprofile=coverage.out ./...
```

Verify that `coverage.out` exists:

```bash
ls -lh coverage.out
```

#### TypeScript/React Coverage

Run the frontend test command with coverage:

```bash
cd web
bun run test -- --coverage
```

Verify that `web/coverage/lcov.info` exists:

```bash
ls -lh web/coverage/lcov.info
```

### Step 4: Add `sonar` Target to Makefile

Add the following target to the project's `Makefile` (or update the existing `sonar` target if present):

```makefile
sonar: ## Run SonarCloud analysis (requires SONAR_TOKEN)
	@echo "Running SonarCloud analysis..."
	@if [ -z "$$SONAR_TOKEN" ]; then \
		echo "ERROR: SONAR_TOKEN environment variable not set."; \
		echo "Set it before running: export SONAR_TOKEN=<your-token>"; \
		exit 1; \
	fi
	docker run --rm \
		-e SONAR_HOST_URL=https://sonarcloud.io \
		-e SONAR_LOGIN=$(SONAR_TOKEN) \
		-v "$(PWD):/src" \
		sonarsource/sonar-scanner-cli
```

### Step 5: Test Coverage Generation (Optional)

To verify that coverage files are generated correctly, run:

```bash
make test
cd web && bun run test -- --coverage
```

Then verify both files exist:

```bash
test -f coverage.out && echo "✓ Go coverage found"
test -f web/coverage/lcov.info && echo "✓ TypeScript coverage found"
```

### Step 6: Optional Local Scan

If `sonar-scanner` CLI is installed locally (not via Docker), you can run a local scan:

```bash
make sonar
```

This will:
1. Verify `SONAR_TOKEN` is set
2. Run SonarCloud analysis with coverage from both Go and TypeScript
3. Push results to SonarCloud

**Note:** Local scans require the `sonar-scanner` CLI or Docker. GitHub Actions CI/CD will run this automatically when you push.

### Step 7: Commit

Stage and commit the configuration:

```bash
git add sonar-project.properties
git add Makefile  # if you modified the sonar target
git commit -m "Configure SonarCloud with Go and TypeScript coverage"
```

## Verification

After committing, verify in SonarCloud:

1. Navigate to your project dashboard
2. Check that coverage metrics appear for both Go and TypeScript
3. Verify that code quality gates are being evaluated

You can also test locally (if `sonar-scanner` is available):

```bash
export SONAR_TOKEN=<your-sonar-token>
make test
cd web && bun run test -- --coverage
make sonar
```

## Notes

- **Coverage Reports:** Both `coverage.out` (Go) and `web/coverage/lcov.info` (TypeScript) should be generated before running the SonarCloud scan
- **Docker:** The Makefile target uses Docker to run the scanner; ensure Docker is installed if using `make sonar`
- **CI/CD:** In GitHub Actions, store `SONAR_TOKEN` as a repository secret and pass it to the `sonar` target
- **Exclusions:** The configuration excludes `node_modules`, `dist`, `vendor`, and generated files (`*.gen.go`) from analysis
