---
name: setup-playwright
description: Configure Playwright with dual configs (unit E2E with auto-start dev server, integration E2E against K8s cluster), fixtures for API mocking, and example specs. Use when bootstrapping a project or adding E2E testing. Trigger on "set up Playwright", "add E2E tests", "configure E2E", or as part of project bootstrapping.
---

# Setup Playwright

Setting up Playwright with dual E2E configs and fixtures.

## Process

### 1. Install Dependencies

Install Playwright and browser binaries:

```bash
bun add -d @playwright/test
bunx playwright install chromium
```

### 2. Unit E2E Config

Create `web/playwright.config.ts` for local development testing with auto-start dev server:

```typescript
import { defineConfig, devices } from "@playwright/test";

export default defineConfig({
  testDir: "./e2e",
  testIgnore: "**/integration/**",
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: "html",
  use: {
    baseURL: "http://localhost:5174",
    trace: "on-first-retry",
  },
  projects: [
    {
      name: "chromium",
      use: { ...devices.chromiumLinux },
    },
  ],
  webServer: {
    command: "bun run dev",
    url: "http://localhost:5174",
    reuseExistingServer: !process.env.CI,
  },
});
```

### 3. Integration E2E Config

Create `web/playwright-integration.config.ts` for testing against a running K8s cluster:

```typescript
import { defineConfig, devices } from "@playwright/test";

export default defineConfig({
  testDir: "./e2e/integration",
  fullyParallel: false,
  workers: 1,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  reporter: "html",
  use: {
    baseURL: process.env.E2E_BASE_URL || (() => {
      throw new Error("E2E_BASE_URL environment variable is required");
    })(),
    trace: "on-first-retry",
  },
  projects: [
    {
      name: "chromium",
      use: { ...devices.chromiumLinux },
    },
  ],
});
```

### 4. Fixtures File

Create `web/e2e/fixtures.ts` with API mocking and base test extension:

```typescript
import { test as base, expect } from "@playwright/test";
import type { Page } from "@playwright/test";

interface TestFixtures {
  mockApi: void;
}

export const test = base.extend<TestFixtures>({
  mockApi: async ({ page }, use) => {
    await page.route(/\/api\/v1\/.*/, async (route) => {
      const request = route.request();
      const method = request.method();
      const url = request.url();

      // Example: mock GET /api/v1/leads
      if (method === "GET" && url.includes("/api/v1/leads")) {
        await route.abort("blockedbyclient");
      } else {
        // Pass through all other requests
        await route.continue();
      }
    });

    await use();
  },
});

export { expect };
```

### 5. Create Directories

Create the necessary directory structure:

```bash
mkdir -p web/e2e/integration
mkdir -p web/e2e/steps
```

### 6. Example Spec

Create `web/e2e/navigation.spec.ts` as a reference implementation:

```typescript
import { test, expect } from "./fixtures";

test.describe("Navigation", () => {
  test.beforeEach(async ({ mockApi, page }) => {
    // mockApi fixture is used but doesn't need explicit setup
    await page.goto("/");
  });

  test("loads home page successfully", async ({ page }) => {
    await expect(page).toHaveTitle(/.*/, { timeout: 5000 });
    await expect(page.locator("body")).toBeVisible();
  });

  test("navigation links are present", async ({ page }) => {
    // Example: check if navigation element exists
    const nav = page.locator("nav");
    await expect(nav).toBeVisible();
  });
});
```

### 7. E2E Web Script

Create `scripts/e2e-web.sh` to port-forward to K8s cluster and run integration E2E tests:

```bash
#!/bin/bash
set -euo pipefail

# Configuration
E2E_NAMESPACE="${E2E_NAMESPACE:-default}"
E2E_SERVICE="${E2E_SERVICE:-pebblr-api}"
E2E_SERVICE_PORT="${E2E_SERVICE_PORT:-8080}"
E2E_WAIT_TIMEOUT="${E2E_WAIT_TIMEOUT:-30}"

# Find a free port
find_free_port() {
  local port=9000
  while netstat -ln 2>/dev/null | grep -q ":$port "; do
    ((port++))
  done
  echo $port
}

FREE_PORT=$(find_free_port)
LOCAL_URL="http://localhost:$FREE_PORT"

echo "Setting up port-forward from $E2E_SERVICE:$E2E_SERVICE_PORT to localhost:$FREE_PORT..."

# Start port-forward in background
kubectl port-forward \
  -n "$E2E_NAMESPACE" \
  "svc/$E2E_SERVICE" \
  "$FREE_PORT:$E2E_SERVICE_PORT" &

PF_PID=$!

# Cleanup trap
cleanup() {
  echo "Cleaning up port-forward (PID: $PF_PID)..."
  kill $PF_PID 2>/dev/null || true
  wait $PF_PID 2>/dev/null || true
}
trap cleanup EXIT

# Wait for service to be ready
echo "Waiting for $LOCAL_URL/healthz (timeout: ${E2E_WAIT_TIMEOUT}s)..."
ELAPSED=0
while [ $ELAPSED -lt $E2E_WAIT_TIMEOUT ]; do
  if curl -sf "$LOCAL_URL/healthz" >/dev/null 2>&1; then
    echo "Service is ready!"
    break
  fi
  sleep 1
  ((ELAPSED++))
done

if [ $ELAPSED -ge $E2E_WAIT_TIMEOUT ]; then
  echo "ERROR: Service did not become ready within ${E2E_WAIT_TIMEOUT}s"
  exit 1
fi

# Run Playwright integration tests
echo "Running Playwright integration E2E tests..."
cd "$(dirname "$0")/../web"
E2E_BASE_URL="$LOCAL_URL" bunx playwright test -c playwright-integration.config.ts "$@"
```

Make the script executable:

```bash
chmod +x scripts/e2e-web.sh
```

### 8. Verify

Run the quality gates:

```bash
bun run typecheck
```

Run unit E2E tests locally:

```bash
cd web
bunx playwright test
```

Run a single spec:

```bash
cd web
bunx playwright test e2e/navigation.spec.ts
```

Generate and open HTML report:

```bash
cd web
bunx playwright show-report
```

## Summary

You now have:

- **Unit E2E config** (`playwright.config.ts`) — runs against `localhost:5174`, auto-starts dev server, fully parallel, CI-aware retries
- **Integration E2E config** (`playwright-integration.config.ts`) — requires `E2E_BASE_URL` env var, serial execution, no auto-start
- **Fixtures file** (`e2e/fixtures.ts`) — mockApi fixture for request routing, re-exported expect
- **Example spec** (`e2e/navigation.spec.ts`) — demonstrates home page navigation test
- **E2E web script** (`scripts/e2e-web.sh`) — port-forwards to K8s cluster, waits for readiness, runs integration tests with cleanup
- **Directories** — `e2e/integration` and `e2e/steps` for future test organization

Both configs use Chromium only, enable tracing on first retry, and are CI-aware. The integration script handles port-finding, health checks, and graceful cleanup.
