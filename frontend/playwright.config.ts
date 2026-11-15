import { defineConfig } from "@playwright/test";

// Сквозные сценарии интерфейса на моках (D-036): только Chromium, без Go и без сети.
// Окно на трёх ОС проверяет smoke, здесь — сам интерфейс.
const PORT = Number(process.env.E2E_PORT) || 9255;

export default defineConfig({
  testDir: "./e2e",
  outputDir: "./test-results",
  fullyParallel: true,
  forbidOnly: Boolean(process.env.CI),
  retries: process.env.CI ? 1 : 0,
  reporter: process.env.CI
    ? [["list"], ["html", { outputFolder: "playwright-report", open: "never" }]]
    : "list",
  use: {
    baseURL: `http://127.0.0.1:${PORT}`,
    // Полный Chromium в режиме headless: отдельный headless shell не нужен.
    browserName: "chromium",
    channel: "chromium",
    locale: "ru-RU",
    trace: "retain-on-failure",
  },
  webServer: {
    command: `yarn dev --port ${PORT}`,
    url: `http://127.0.0.1:${PORT}`,
    reuseExistingServer: !process.env.CI,
    timeout: 60_000,
  },
});
