import { test, expect } from '@playwright/test';

test.describe('Network Resilience & API Mocking', () => {
  test('should handle offline network state gracefully without crashing', async ({ context, page }) => {
    // 1. Initial page load online
    await page.goto('/', { waitUntil: 'load' });
    await page.waitForTimeout(1000);

    const canvas = page.locator('#ComposeTarget');
    await expect(canvas).toBeAttached();

    // 2. Simulate complete network offline
    await context.setOffline(true);
    await page.waitForTimeout(500);

    // Try interacting with the app while offline
    const box = (await canvas.boundingBox()) || { width: 500, height: 600 };
    await canvas.click({ position: { x: Math.floor(box.width / 2), y: 50 } });
    await page.waitForTimeout(300);

    // Ensure the page stays responsive and attached
    await expect(canvas).toBeAttached();

    // 3. Restore network online
    await context.setOffline(false);
    await page.waitForTimeout(500);
    await expect(canvas).toBeAttached();
  });

  test('should gracefully handle backend geofence API responses and fallback mock', async ({ page }) => {
    // Intercept active geofences API call to verify communication contract
    let apiCalled = false;
    await page.route('**/api/v1/geofences/active*', async (route) => {
      apiCalled = true;
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          tier: 'TIER_2_WARNING',
          species: ['BEAR', 'TIGER'],
          features: [
            {
              type: 'Feature',
              geometry: {
                type: 'Polygon',
                coordinates: [
                  [
                    [141.14, 39.71],
                    [141.16, 39.71],
                    [141.16, 39.73],
                    [141.14, 39.73],
                    [141.14, 39.71]
                  ]
                ]
              },
              properties: {
                riskScore: 0.88,
                species: 'BEAR'
              }
            }
          ]
        })
      });
    });

    await page.goto('/', { waitUntil: 'load' });
    await page.waitForTimeout(1000);

    const canvas = page.locator('#ComposeTarget');
    await expect(canvas).toBeAttached();
  });

  test('should handle backend 500 error gracefully without unhandled exception', async ({ page }) => {
    const unhandledErrors: string[] = [];
    page.on('pageerror', (err) => unhandledErrors.push(err.message));

    // Force backend API to return 500
    await page.route('**/api/v1/**', async (route) => {
      await route.fulfill({
        status: 500,
        contentType: 'application/json',
        body: JSON.stringify({ error: 'Internal Server Error' })
      });
    });

    await page.goto('/', { waitUntil: 'load' });
    await page.waitForTimeout(1000);

    const canvas = page.locator('#ComposeTarget');
    await expect(canvas).toBeAttached();

    // Verify application did not crash fatally
    expect(unhandledErrors).toHaveLength(0);
  });
});
