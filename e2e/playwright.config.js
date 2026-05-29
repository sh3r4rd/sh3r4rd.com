import { defineConfig, devices } from "@playwright/test";

/**
 * Playwright configuration for the recruiter dashboard E2E suite.
 *
 * This config lives in `e2e/` (not the repo root), so paths that Playwright
 * resolves relative to the config directory are set accordingly:
 *   - `testDir: '.'` -> the e2e/ folder itself.
 *   - `webServer.cwd: '..'` -> the repo root, where `npm run dev` (Vite) lives.
 *
 * Invoke via the package scripts, which pass `--config e2e/playwright.config.js`.
 */
export default defineConfig({
  testDir: ".",
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  reporter: [["html", { outputFolder: "../playwright-report", open: "never" }]],
  outputDir: "../test-results",

  use: {
    baseURL: "http://localhost:5173",
    screenshot: "only-on-failure",
    trace: "on-first-retry",
  },

  projects: [
    {
      name: "desktop-chrome",
      use: {
        ...devices["Desktop Chrome"],
        viewport: { width: 1280, height: 720 },
      },
    },
    {
      name: "mobile-chrome",
      use: {
        ...devices["Desktop Chrome"],
        viewport: { width: 375, height: 667 },
        isMobile: true,
        hasTouch: true,
      },
    },
  ],

  webServer: {
    command: "npm run dev",
    cwd: "..",
    url: "http://localhost:5173",
    reuseExistingServer: !process.env.CI,
    timeout: 120 * 1000,
  },
});
