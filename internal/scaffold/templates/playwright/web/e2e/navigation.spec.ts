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
