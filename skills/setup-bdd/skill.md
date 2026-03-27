---
name: setup-bdd
description: Set up BDD infrastructure with godog for Go and playwright-bdd for the frontend, including feature directories, example feature files, and step definition scaffolds. Use when bootstrapping a project or adding BDD to an existing one. Trigger on "set up BDD", "add Cucumber", "add Gherkin", "set up godog", "set up playwright-bdd", or as part of project bootstrapping.
---

# Setup BDD

Announce: "Setting up BDD with godog and playwright-bdd."

## Process

### 1. Go BDD Setup

1. Install godog:
   ```bash
   go get github.com/cucumber/godog@latest
   ```

2. Create features directory:
   ```bash
   mkdir -p features
   ```

3. Create `features/health.feature`:
   ```gherkin
   Feature: Health Check
     As a user
     I want the API to provide a health check endpoint
     So that I can verify the service is running

     Scenario: API should return healthy status
       Given the API server is running
       When I make a GET request to "/api/v1/health"
       Then the response status should be 200
       And the response body should contain "healthy": true
   ```

4. Create `internal/api/health_bdd_test.go`:
   ```go
   package api

   import (
       "bytes"
       "encoding/json"
       "fmt"
       "io"
       "net/http"
       "net/http/httptest"
       "testing"

       "github.com/cucumber/godog"
   )

   type apiContext struct {
       server   *httptest.Server
       response *http.Response
       body     []byte
       err      error
   }

   func (ac *apiContext) theAPIServerIsRunning() error {
       // In a real application, you would initialize your actual API router here
       // For now, we'll create a minimal health check handler
       mux := http.NewServeMux()
       mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
           w.Header().Set("Content-Type", "application/json")
           w.WriteHeader(http.StatusOK)
           json.NewEncoder(w).Encode(map[string]bool{"healthy": true})
       })

       ac.server = httptest.NewServer(mux)
       return nil
   }

   func (ac *apiContext) iMakeAGETRequestToEndpoint(endpoint string) error {
       url := ac.server.URL + endpoint
       resp, err := http.Get(url)
       if err != nil {
           return fmt.Errorf("making GET request: %w", err)
       }

       ac.response = resp
       body, err := io.ReadAll(resp.Body)
       if err != nil {
           return fmt.Errorf("reading response body: %w", err)
       }
       defer resp.Body.Close()
       ac.body = body
       return nil
   }

   func (ac *apiContext) theResponseStatusShouldBe(status int) error {
       if ac.response.StatusCode != status {
           return fmt.Errorf("expected status %d, got %d", status, ac.response.StatusCode)
       }
       return nil
   }

   func (ac *apiContext) theResponseBodyShouldContain(expected string) error {
       if !bytes.Contains(ac.body, []byte(expected)) {
           return fmt.Errorf("expected response to contain %q, got %s", expected, string(ac.body))
       }
       return nil
   }

   func (ac *apiContext) closeServer() {
       if ac.server != nil {
           ac.server.Close()
       }
   }

   // InitializeScenario is called before each scenario to reset context
   func (ac *apiContext) InitializeScenario(ctx *godog.ScenarioContext) {
       ctx.Given(`the API server is running`, ac.theAPIServerIsRunning)
       ctx.When(`I make a GET request to "([^"]*)"`, ac.iMakeAGETRequestToEndpoint)
       ctx.Then(`the response status should be (\d+)`, ac.theResponseStatusShouldBe)
       ctx.Then(`the response body should contain "([^"]*)"`, ac.theResponseBodyShouldContain)
       ctx.After(func(ctx context.Context, scenario *godog.Scenario, err error) context.Context {
           ac.closeServer()
           return ctx
       })
   }

   // TestHealthFeatures runs all feature tests for health checks
   func TestHealthFeatures(t *testing.T) {
       suite := godog.TestSuite{
           ScenarioInitializer: func(ctx *godog.ScenarioContext) {
               ac := &apiContext{}
               ac.InitializeScenario(ctx)
           },
           Options: &godog.Options{
               Format:   "pretty",
               Paths:    []string{"../../features"},
               TestingT: t,
           },
       }

       if suite.Run() != 0 {
           t.Fatal("godog scenarios failed")
       }
   }
   ```

   Note: Add `context` import if needed:
   ```go
   import "context"
   ```

### 2. Frontend BDD Setup

1. Install playwright-bdd:
   ```bash
   cd web
   bun add -d playwright-bdd
   ```

2. Create features directory:
   ```bash
   mkdir -p web/features
   ```

3. Create `web/features/navigation.feature`:
   ```gherkin
   Feature: Navigation
     As a user
     I want to navigate between pages
     So that I can access different parts of the application

     Scenario: User can navigate to the leads page
       Given I am on the home page
       When I click the "Leads" navigation link
       Then I should be on the "/leads" page
       And the page title should contain "Leads"

     Scenario: User can navigate to the dashboard
       Given I am on the home page
       When I click the "Dashboard" navigation link
       Then I should be on the "/dashboard" page
   ```

4. Create `web/e2e/steps/navigation.steps.ts`:
   ```typescript
   import { createBdd } from 'playwright-bdd';
   import { expect, test } from '@playwright/test';

   const { Given, When, Then } = createBdd(test);

   Given('I am on the home page', async ({ page }) => {
       await page.goto('/');
       // Optionally wait for specific element to ensure page is loaded
       await page.waitForURL(/^\/$|^$/);
   });

   When('I click the {string} navigation link', async ({ page }, linkText: string) => {
       await page.click(`a:has-text("${linkText}")`);
   });

   Then('I should be on the {string} page', async ({ page }, path: string) => {
       await page.waitForURL(new RegExp(`${path}$`));
       expect(page.url()).toContain(path);
   });

   Then('the page title should contain {string}', async ({ page }, text: string) => {
       const title = await page.title();
       expect(title).toContain(text);
   });
   ```

5. Update `web/playwright.config.ts` to use defineBddConfig:
   ```typescript
   import { defineConfig, devices } from '@playwright/test';
   import { defineBddConfig } from 'playwright-bdd';

   const testDir = './e2e/features';

   export default defineConfig(
       defineBddConfig({
           testDir,
           /* Run tests in files in parallel */
           fullyParallel: true,

           /* Fail the build on CI if you accidentally left test.only in the source code. */
           forbidOnly: !!process.env.CI,

           /* Retry on CI only */
           retries: process.env.CI ? 2 : 0,

           /* Opt out of parallel tests on CI. */
           workers: process.env.CI ? 1 : undefined,

           /* Reporter to use. See https://playwright.dev/docs/test-reporters */
           reporter: 'html',

           use: {
               /* Base URL to use in actions like `await page.goto('/')`. */
               baseURL: 'http://localhost:5173',
               trace: 'on-first-retry',
           },

           /* Configure projects for major browsers */
           projects: [
               {
                   name: 'chromium',
                   use: { ...devices['Desktop Chrome'] },
               },
               {
                   name: 'firefox',
                   use: { ...devices['Desktop Firefox'] },
               },
           ],

           /* Run your local dev server before starting the tests */
           webServer: {
               command: 'bun run dev',
               url: 'http://localhost:5173',
               reuseExistingServer: !process.env.CI,
           },
       })
   );
   ```

### 3. Verify Both Pass

Run Go BDD tests:
```bash
go test ./internal/api -v -run TestHealthFeatures
```

Run frontend BDD tests:
```bash
cd web
bun run test:e2e
```

Both test suites should complete successfully.

### 4. Commit

```bash
git add features/ internal/api/health_bdd_test.go web/features/ web/e2e/steps/navigation.steps.ts web/playwright.config.ts
git commit -m "Set up BDD infrastructure with godog and playwright-bdd"
```

## Summary

BDD infrastructure is now in place:
- Go: godog installed with example health check feature and step definitions
- Frontend: playwright-bdd configured with example navigation feature
- Both test suites verified and passing
- Ready for incremental feature development in BDD style
