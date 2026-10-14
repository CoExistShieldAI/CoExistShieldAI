import { test, expect } from '@playwright/test';

test.describe('Canvas Rendering & User Interactions', () => {
  test('canvas should have valid dimensions and fill viewport', async ({ page }) => {
    await page.goto('/', { waitUntil: 'load' });

    const canvas = page.locator('#ComposeTarget');
    await expect(canvas).toBeAttached();

    // Verify canvas bounding box matches window viewport
    const viewportSize = page.viewportSize();
    expect(viewportSize).not.toBeNull();

    if (viewportSize) {
      const box = await canvas.boundingBox();
      expect(box).not.toBeNull();
      if (box) {
        expect(box.width).toBeGreaterThanOrEqual(viewportSize.width - 5);
        expect(box.height).toBeGreaterThanOrEqual(viewportSize.height - 5);
      }
    }

    // Check canvas element attributes and computed styles
    const canvasStyles = await canvas.evaluate((el) => {
      const computed = window.getComputedStyle(el);
      return {
        position: computed.position,
        zIndex: computed.zIndex,
        width: el.clientWidth,
        height: el.clientHeight,
      };
    });

    expect(canvasStyles.position).toBe('absolute');
    expect(Number(canvasStyles.zIndex)).toBeGreaterThanOrEqual(1);
    expect(canvasStyles.width).toBeGreaterThan(0);
    expect(canvasStyles.height).toBeGreaterThan(0);
  });

  test('canvas should accept pointer events without runtime errors', async ({ page }) => {
    const pageErrors: Error[] = [];
    page.on('pageerror', (err) => pageErrors.push(err));

    await page.goto('/', { waitUntil: 'load' });
    const canvas = page.locator('#ComposeTarget');
    await expect(canvas).toBeAttached();

    // Wait for initial render stabilization
    await page.waitForTimeout(1500);

    // Simulate clicking top bar / demo buttons area
    await canvas.click({ position: { x: 100, y: 30 } });
    await page.waitForTimeout(300);

    // Simulate clicking center of canvas (HUD alert zone)
    const box = (await canvas.boundingBox()) || { width: 400, height: 600, x: 0, y: 0 };
    await canvas.click({
      position: { x: Math.floor(box.width / 2), y: Math.floor(box.height / 2) },
    });
    await page.waitForTimeout(300);

    // Simulate clicking bottom navigation bar area (e.g. GIS Map or Console tab)
    await canvas.click({
      position: { x: Math.floor(box.width * 0.75), y: Math.floor(box.height - 30) },
    });
    await page.waitForTimeout(500);

    // Ensure no uncaught JS runtime errors happened during pointer interactions
    expect(pageErrors).toHaveLength(0);
  });

  test('should handle window resize gracefully', async ({ page, isMobile }) => {
    test.skip(!!isMobile, 'Cannot resize viewport in mobile device emulation');
    await page.goto('/', { waitUntil: 'load' });
    const canvas = page.locator('#ComposeTarget');
    await expect(canvas).toBeAttached();

    // Resize viewport dynamically
    await page.setViewportSize({ width: 800, height: 600 });
    await page.waitForTimeout(500);

    const smallBox = await canvas.boundingBox();
    expect(smallBox?.width).toBeCloseTo(800, -1);
    expect(smallBox?.height).toBeCloseTo(600, -1);

    // Expand viewport dynamically
    await page.setViewportSize({ width: 1440, height: 900 });
    await page.waitForTimeout(500);

    const largeBox = await canvas.boundingBox();
    expect(largeBox?.width).toBeCloseTo(1440, -1);
    expect(largeBox?.height).toBeCloseTo(900, -1);
  });

  test('should capture visual snapshot of the demo web application', async ({ page }) => {
    await page.goto('/', { waitUntil: 'networkidle' });
    await page.waitForTimeout(2000);

    // Capture screenshot for visual inspection
    const screenshot = await page.screenshot({ fullPage: true });
    expect(screenshot).not.toBeNull();
    expect(screenshot.length).toBeGreaterThan(1000);
  });
});
