import { defineConfig, devices } from '@playwright/test';

const frontendPort = 14173;
const backendPort = 18080;

export default defineConfig({
  testDir: './e2e',
  timeout: 60_000,
  expect: {
    timeout: 10_000
  },
  fullyParallel: false,
  workers: 1,
  retries: 0,
  reporter: 'list',
  use: {
    baseURL: `http://127.0.0.1:${frontendPort}`,
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure'
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] }
    }
  ],
  webServer: [
    {
      command: 'go run ./cmd/api',
      cwd: 'backend',
      url: `http://127.0.0.1:${backendPort}/healthz`,
      reuseExistingServer: false,
      timeout: 120_000,
      env: {
        APP_ENV: 'development',
        APP_PORT: String(backendPort),
        BACKEND_URL: `http://127.0.0.1:${backendPort}`,
        FRONTEND_URL: `http://127.0.0.1:${frontendPort}`,
        CORS_ALLOWED_ORIGINS: `http://127.0.0.1:${frontendPort}`,
        COOKIE_DOMAIN: '127.0.0.1',
        COOKIE_SECURE: 'false',
        DEFAULT_LOCALE: 'ru',
        ENABLE_NOTIFICATION_EMAILS: 'false',
        SESSION_SECRET: 'playwright-local-session-secret',
        DATABASE_URL: '',
        AI_API_BASE_URL: '',
        AI_API_KEY: '',
        AI_MODEL: ''
      }
    },
    {
      command: `npm run dev -- --host 127.0.0.1 --port ${frontendPort}`,
      cwd: 'frontend',
      url: `http://127.0.0.1:${frontendPort}`,
      reuseExistingServer: false,
      timeout: 120_000,
      env: {
        VITE_API_BASE_URL: `http://127.0.0.1:${backendPort}/api/v1`
      }
    }
  ]
});
