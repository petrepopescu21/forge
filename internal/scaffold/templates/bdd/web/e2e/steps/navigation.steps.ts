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
