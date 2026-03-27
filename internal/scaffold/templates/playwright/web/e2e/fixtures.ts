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
