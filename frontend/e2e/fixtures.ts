// Общая подготовка сценариев: страница не ходит в сеть и не пишет в консоль
// ошибок и предупреждений (D-036). Нарушение роняет сценарий после его шагов.
import { expect, test as base, type Page } from "@playwright/test";

// Предупреждение рантайма Wails в обычном браузере: штатное для режима моков.
const allowedConsole = [/Browser Environment Detected/];

export const test = base.extend<{ consoleProblems: string[] }>({
  consoleProblems: [
    async ({ page, baseURL }, use) => {
      const problems: string[] = [];
      const origin = new URL(baseURL ?? "http://127.0.0.1").origin;
      await page.route("**/*", (route) => {
        const url = route.request().url();
        if (url.startsWith(origin) || url.startsWith("data:") || url.startsWith("blob:"))
          return route.continue();
        problems.push(`запрос в сеть: ${url}`);
        return route.abort();
      });
      page.on("console", (msg) => {
        if (msg.type() !== "error" && msg.type() !== "warning") return;
        const text = msg.text();
        if (allowedConsole.some((re) => re.test(text))) return;
        problems.push(`console.${msg.type()}: ${text}`);
      });
      page.on("pageerror", (err) => problems.push(`исключение: ${err.message}`));
      await use(problems);
      expect(problems, "консоль и сеть").toEqual([]);
    },
    { auto: true },
  ],
});

export { expect };

/** Открыть страницу route с начальным состоянием моков из параметров params. */
export async function open(page: Page, route: string, params: Record<string, string> = {}) {
  const query = new URLSearchParams({ delay: "50", ...params }).toString();
  await page.goto(`/?${query}#${route}`);
  await expect(page.getByRole("heading", { level: 1 })).toBeVisible();
}
