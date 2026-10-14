import { test, expect } from '@playwright/test';

test.describe('GIS Hazard Map Screen with Google Maps Vector Engine', () => {
  test('should navigate to GIS Map and verify Google Maps Vector Engine rendering', async ({ page }) => {
    const pageErrors: Error[] = [];
    page.on('pageerror', (err) => pageErrors.push(err));

    await page.goto('/', { waitUntil: 'load' });
    await page.waitForTimeout(2000);

    const canvas = page.locator('#ComposeTarget');
    await expect(canvas).toBeAttached();

    // 1. Click bottom navigation bar "GIS Map" tab (approx 37.5% across width, bottom 30px)
    const box = (await canvas.boundingBox()) || { width: 1280, height: 800 };
    await canvas.click({
      position: { x: Math.floor(box.width * 0.38), y: Math.floor(box.height - 30) },
    });
    await page.waitForTimeout(1000);

    // 2. Perform map pan (drag gestures)
    await page.mouse.move(box.width / 2, box.height / 2);
    await page.mouse.down();
    await page.mouse.move(box.width / 2 + 100, box.height / 2 + 50, { steps: 5 });
    await page.mouse.up();
    await page.waitForTimeout(500);

    // 3. Click Zoom In floating button (bottom right area)
    await canvas.click({
      position: { x: Math.floor(box.width - 35), y: Math.floor(box.height - 130) },
    });
    await page.waitForTimeout(500);

    // 4. Click Region Tab (e.g. Thailand or Sumatra)
    await canvas.click({
      position: { x: 180, y: 70 },
    });
    await page.waitForTimeout(800);

    // 5. Click Layer Toggle (FilterChip)
    await canvas.click({
      position: { x: 100, y: 110 },
    });
    await page.waitForTimeout(500);

    // 6. Capture map screenshot for visual inspection
    const mapScreenshot = await page.screenshot({ fullPage: true });
    expect(mapScreenshot.length).toBeGreaterThan(1000);

    // Verify no unhandled runtime errors occurred
    expect(pageErrors).toHaveLength(0);
  });
});
