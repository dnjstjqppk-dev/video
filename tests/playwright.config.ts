import { defineConfig } from '@playwright/test';

const apiPort = Number(process.env.GVP_API_PORT || 8080);

export default defineConfig({
  testDir: './e2e',
  testMatch: '**/*.spec.ts',
  timeout: 60_000,
  expect: { timeout: 10_000 },
  fullyParallel: false,
  workers: 1,
  reporter: [['list']],
  use: {
    baseURL: `http://localhost:${apiPort}/v1`,
  },
  webServer: {
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
      GVP_LOCAL_MEDIA_DIR: './test-results/e2e-media',
    } as Record<string, string>,
  },
});
