// @ts-check
const { defineConfig } = require('@playwright/test');

/**
 * Playwright configuration for miniprogram UI automation tests.
 *
 * Environment variables:
 *   APP_BASE_URL  – backend API base (default http://localhost:8080)
 *   UI_BASE_URL   – UI server base   (default http://localhost:8081)
 */
module.exports = defineConfig({
  testDir: '.',
  testMatch: '*.spec.js',
  timeout: 60_000,
  expect: { timeout: 10_000 },
  fullyParallel: false,
  retries: 1,
  workers: 1,
  reporter: [
    ['list'],
    ['allure-playwright', { outputFolder: 'allure-results' }],
    ['html', { open: 'never', outputFolder: 'html-report' }],
  ],
  use: {
    baseURL: process.env.UI_BASE_URL || 'http://localhost:8081',
    screenshot: 'on',
    trace: 'retain-on-failure',
    video: 'on',
    actionTimeout: 10_000,
    navigationTimeout: 30_000,
  },
  outputDir: 'test-results',
  projects: [
    {
      name: 'chromium',
      use: {
        browserName: 'chromium',
        viewport: { width: 1280, height: 720 },
      },
    },
  ],
});
