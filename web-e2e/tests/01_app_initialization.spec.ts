import { test, expect } from '@playwright/test';

test.describe('Web App Initialization & Core Assets', () => {
  test('should load page with valid title and metadata', async ({ page }) => {
    const consoleErrors: string[] = [];
    page.on('console', (msg) => {
      if (msg.type() === 'error') {
        // Exclude benign font or network warnings if any
        consoleErrors.push(msg.text());
      }
    });

    const response = await page.goto('/', { waitUntil: 'domcontentloaded' });
    expect(response?.status()).toBe(200);

    // Verify Title
    await expect(page).toHaveTitle(/CoExist Shield AI/);

    // Verify Viewport and Meta tags
    const metaDescription = page.locator('meta[name="description"]');
    await expect(metaDescription).toHaveAttribute('content', /predictive geofencing/i);

    // Verify Root Container and Splash Element
    const rootElement = page.locator('#root');
    await expect(rootElement).toBeVisible();

    const splashElement = page.locator('#splash');
    await expect(splashElement).toBeAttached();

    // Verify Splash details
    const titleText = page.locator('#splash .title');
    await expect(titleText).toHaveText('CoExist Shield AI');

    const statusText = page.locator('#splash .status');
    await expect(statusText).toHaveText('Initializing Autonomous Defense System...');

    // Verify ComposeTarget Canvas Element
    const canvas = page.locator('#ComposeTarget');
    await expect(canvas).toBeAttached();

    // Wait briefly for initial scripts to parse
    await page.waitForTimeout(2000);

    // Filter out known benign errors (e.g. cross-origin font warnings in headless mode)
    const fatalErrors = consoleErrors.filter(
      (err) =>
        !err.includes('WebGL') &&
        !err.includes('favicon') &&
        !err.includes('net::ERR_FAILED')
    );
    expect(fatalErrors).toEqual([]);
  });

  test('should serve critical static assets with correct status codes', async ({ request, baseURL }) => {
    // 1. index.html
    const indexRes = await request.get('/');
    expect(indexRes.status()).toBe(200);
    const indexHtml = await indexRes.text();
    expect(indexHtml).toContain('ComposeTarget');
    expect(indexHtml).toContain('shared.js');

    // 2. shared.js
    const sharedJsRes = await request.get('/shared.js');
    expect(sharedJsRes.status()).toBe(200);
    expect((await sharedJsRes.body()).length).toBeGreaterThan(1000);

    // 3. skiko.js
    const skikoJsRes = await request.get('/skiko.js');
    expect(skikoJsRes.status()).toBe(200);

    // 4. skiko.wasm
    const skikoWasmRes = await request.get('/skiko.wasm');
    expect(skikoWasmRes.status()).toBe(200);
    expect((await skikoWasmRes.body()).length).toBeGreaterThan(100000);
  });
});
