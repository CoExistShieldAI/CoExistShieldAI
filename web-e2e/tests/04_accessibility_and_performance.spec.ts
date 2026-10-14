import { test, expect } from '@playwright/test';

test.describe('Accessibility & Web Performance Metrics', () => {
  test('html document should meet essential SEO and Accessibility criteria', async ({ page }) => {
    await page.goto('/', { waitUntil: 'domcontentloaded' });

    // HTML lang attribute
    const htmlLang = await page.locator('html').getAttribute('lang');
    expect(htmlLang).toBe('en');

    // Title tag presence and length
    const title = await page.title();
    expect(title.length).toBeGreaterThan(10);
    expect(title).toContain('CoExist Shield AI');

    // Meta viewport presence
    const viewportMeta = page.locator('meta[name="viewport"]');
    await expect(viewportMeta).toBeAttached();

    // Canvas element has proper ID
    const canvas = page.locator('#ComposeTarget');
    await expect(canvas).toBeAttached();
    await expect(canvas).toHaveId('ComposeTarget');
  });

  test('page performance metrics should be within reasonable bounds', async ({ page }) => {
    await page.goto('/', { waitUntil: 'load' });

    // Extract Navigation Timing API metrics
    const timing = await page.evaluate(() => {
      const nav = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming;
      return {
        dnsLookupTime: nav.domainLookupEnd - nav.domainLookupStart,
        tcpConnectTime: nav.connectEnd - nav.connectStart,
        domContentLoadedTime: nav.domContentLoadedEventEnd - nav.startTime,
        loadTime: nav.loadEventEnd - nav.startTime,
      };
    });

    expect(timing.domContentLoadedTime).toBeGreaterThanOrEqual(0);
    // On local server or fast CDN, DOMContentLoaded is typically within 10 seconds
    expect(timing.domContentLoadedTime).toBeLessThan(10000);
  });

  test('should render properly across multiple device viewport form factors', async ({ page, isMobile }) => {
    test.skip(!!isMobile, 'Cannot resize viewport in mobile device emulation');
    const viewports = [
      { name: 'Tablet (iPad Landscape)', width: 1024, height: 768 },
      { name: 'Tablet (iPad Portrait)', width: 768, height: 1024 },
      { name: 'Ultra-wide Desktop', width: 1920, height: 1080 },
      { name: 'Compact Mobile', width: 360, height: 640 },
    ];

    for (const vp of viewports) {
      await page.setViewportSize({ width: vp.width, height: vp.height });
      await page.goto('/', { waitUntil: 'load' });
      await page.waitForTimeout(500);

      const canvas = page.locator('#ComposeTarget');
      await expect(canvas).toBeAttached();

      const box = await canvas.boundingBox();
      expect(box).not.toBeNull();
      if (box) {
        expect(box.width).toBeCloseTo(vp.width, -1);
        expect(box.height).toBeCloseTo(vp.height, -1);
      }
    }
  });
});
