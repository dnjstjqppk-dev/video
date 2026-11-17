import { defineConfig } from '@playwright/test';

// P1-19 responsive sweep: boots the web console (dev) plus the Go API and
// asserts zero horizontal overflow at 8 viewport widths. Run explicitly:
//   npx playwright test --config=responsive.config.ts
// (Outside tests/e2e so the API release-gate suite stays untouched.)

const apiPort = Number(process.env.GVP_API_PORT || 18087);
const webPort = Number(process.env.GVP_WEB_PORT || 13173);

export default defineConfig({
  testDir: './e2e-responsive',
  testMatch: '**/*.spec.ts',
  timeout: 120_000,
  expect: { timeout: 15_000 },
  fullyParallel: false,
  workers: 1,
  reporter: [['list']],
  use: {
    baseURL: `http://localhost:${webPort}`,
  },
  webServer: [
    {
      command: 'go run ./cmd/server',
      cwd: '../services/api',
      port: apiPort,
      timeout: 120_000,
      reuseExistingServer: !process.env.CI,
      env: {
        ...process.env,
        GVP_ENV: 'test',
        GVP_ALLOW_DEV_TOKENS: 'true',
        PORT: String(apiPort),
        GVP_LOCAL_MEDIA_DIR: './test-results/responsive-media',
      } as Record<string, string>,
    },
    {
      command: `pnpm exec nuxt dev --port ${webPort}`,
      cwd: '../apps/web',
      port: webPort,
      timeout: 240_000,
      reuseExistingServer: !process.env.CI,
      env: {
        ...process.env,
        PORT: String(webPort),
        NUXT_PUBLIC_API_BASE_URL: '/api',
        NUXT_INTERNAL_API_URL: `http://localhost:${apiPort}`,
      } as Record<string, string>,
    },
  ],
});
