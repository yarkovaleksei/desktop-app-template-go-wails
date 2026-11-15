import { expect, open, test } from "./fixtures";

const themes = [
  { theme: "light", bg: "rgb(245, 246, 248)" },
  { theme: "dark", bg: "rgb(15, 20, 28)" },
];
const langs = [
  { lang: "ru", home: "Главная", settings: "Настройки", intro: /Заготовка главной страницы/ },
  { lang: "en", home: "Home", settings: "Settings", intro: /placeholder home page/ },
];

for (const { theme, bg } of themes) {
  for (const l of langs) {
    test(`страницы: тема ${theme}, язык ${l.lang}`, async ({ page }) => {
      await open(page, "/", { theme, lang: l.lang });
      await expect(page.locator("html")).toHaveAttribute("lang", l.lang);
      await expect(page.locator("body")).toHaveCSS("background-color", bg);
      await expect(page.getByRole("heading", { level: 1 })).toHaveText("Desktop App");
      await expect(page.getByText(l.intro)).toBeVisible();

      await page.getByRole("tab", { name: l.settings }).click();
      await expect(page.getByRole("heading", { level: 1 })).toHaveText(l.settings);
      await expect(page).toHaveURL(/#\/settings$/);

      await page.getByRole("tab", { name: l.home }).click();
      await expect(page.getByRole("heading", { level: 1 })).toHaveText("Desktop App");
    });
  }
}

test("смена языка и темы применяется сразу", async ({ page }) => {
  await open(page, "/settings", { lang: "ru", theme: "light" });
  await page.getByRole("button", { name: "English" }).click();
  await expect(page.getByRole("heading", { level: 1 })).toHaveText("Settings");
  await expect(page.locator("html")).toHaveAttribute("lang", "en");
  await page.getByRole("button", { name: "Dark" }).click();
  await expect(page.locator("body")).toHaveCSS("background-color", "rgb(15, 20, 28)");
  await expect(page.getByRole("button", { name: "Dark" })).toHaveAttribute("aria-pressed", "true");
});

test("неизвестный адрес ведёт на главную", async ({ page }) => {
  await page.goto("/?lang=ru#/nope");
  await expect(page.getByRole("heading", { level: 1 })).toHaveText("Desktop App");
});
