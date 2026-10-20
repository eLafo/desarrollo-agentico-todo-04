import { defineConfig } from '@playwright/test';

// Tests de navegador: la aplicación de este directorio, arrancada para los tests con la base
// de datos de test (GO_ENV=test), en un puerto distinto del de la aplicación en desarrollo
// (3001 en el checkout principal; en un worktree de bin/tree, el E2E_PORT de su .env).
const port = Number(process.env.E2E_PORT ?? 3001);
const baseURL = `http://127.0.0.1:${port}`;

export default defineConfig({
  testDir: './e2e',
  // Los tests comparten la base de datos de test: uno detrás de otro.
  workers: 1,
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  reporter: [['list'], ['html', { open: 'never' }]],
  use: {
    baseURL,
    browserName: 'chromium',
    headless: true,
    viewport: { width: 1280, height: 800 },
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
  },
  webServer: {
    command: 'go run ./cmd/app',
    url: `${baseURL}/healthz`,
    env: { GO_ENV: 'test', ADDR: '127.0.0.1', PORT: String(port) },
    reuseExistingServer: false,
    timeout: 180_000,
    stdout: 'ignore',
    stderr: 'pipe',
  },
});
