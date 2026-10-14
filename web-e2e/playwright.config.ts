import { defineConfig, devices } from '@playwright/test';

const isProduction = !!process.env.BASE_URL;
const baseURL = process.env.BASE_URL || 'http://localhost:8085';

export default defineConfig({
  testDir: './tests',
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  workers: 1,
  reporter: [
    ['list'],
    ['html', { outputFolder: 'playwright-report', open: 'never' }]
  ],
  timeout: 60000,
  expect: {
    timeout: 15000,
  },
  use: {
    baseURL,
    trace: 'on-first-retry',
    screenshot: 'on',
    video: 'off',
  },
  projects: [
    {
      name: 'Desktop Chrome',
      use: {
        ...devices['Desktop Chrome'],
        viewport: { width: 1280, height: 800 },
      },
    },
    {
      name: 'Mobile Chrome',
      use: {
        ...devices['Pixel 7'],
      },
    },
  ],
  webServer: isProduction
    ? undefined
    : {
        command: 'node static-server.mjs',
        url: 'http://localhost:8085',
        reuseExistingServer: !process.env.CI,
        timeout: 30000,
      },
});
