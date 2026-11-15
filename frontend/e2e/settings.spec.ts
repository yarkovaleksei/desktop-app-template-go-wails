import { expect, open, test } from "./fixtures";

test("переключатели настроек", async ({ page }) => {
  await open(page, "/settings", { lang: "ru" });
  for (const name of [
    "Работать в фоне при закрытии окна",
    "Запускать при входе в систему",
    "Проверять обновления автоматически",
  ]) {
    const sw = page.getByRole("switch", { name });
    await expect(sw).not.toBeChecked();
    await sw.click();
    await expect(sw).toBeChecked();
  }
});

test("без трея работа в фоне недоступна с пояснением", async ({ page }) => {
  await open(page, "/settings", { lang: "ru", tray: "0" });
  const sw = page.getByRole("switch", { name: "Работать в фоне при закрытии окна" });
  await expect(sw).toBeDisabled();
  await expect(sw).toHaveAccessibleDescription(/нет области уведомлений/);
});

test("испорченный файл: предупреждение и ошибка сохранения из словаря", async ({ page }) => {
  await open(page, "/settings", { lang: "ru", broken: "1" });
  await expect(page.getByText(/Файл настроек повреждён или записан более новой версией/)).toBeVisible();
  await page.getByRole("button", { name: "Тёмная" }).click();
  await expect(page.getByRole("alert").filter({ hasText: "изменения не сохранятся" })).toBeVisible();
  await expect(page.getByRole("button", { name: "Тёмная" })).toHaveAttribute("aria-pressed", "false");
});

test("автозапуск недоступен: ошибка из словаря, переключатель возвращается", async ({ page }) => {
  await open(page, "/settings", { lang: "ru", autostart: "unavailable" });
  const sw = page.getByRole("switch", { name: "Запускать при входе в систему" });
  await sw.click();
  await expect(page.getByRole("alert").filter({ hasText: "в этой системе недоступен" })).toBeVisible();
  await expect(sw).not.toBeChecked();
});

test("сведения о сборке", async ({ page }) => {
  await open(page, "/settings", { lang: "en", format: "nsis" });
  await expect(page.getByText("Version v1.0.0")).toBeVisible();
  await expect(page.getByText("Package format: nsis")).toBeVisible();
});
