import { expect, open, test } from "./fixtures";

const updates = (page: import("@playwright/test").Page) =>
  page.locator("section").filter({ has: page.getByRole("heading", { name: /^(Обновления|Updates)$/ }) });

test("unsupported: сборка из исходников", async ({ page }) => {
  await open(page, "/settings", { lang: "ru", format: "" });
  await expect(updates(page).getByText("Эта сборка не из пакета и обновляется вручную.")).toBeVisible();
  await expect(updates(page).getByRole("button", { name: "Проверить обновления" })).toHaveCount(0);
});

test("current: новее нет", async ({ page }) => {
  await open(page, "/settings", { lang: "ru", latest: "v1.0.0" });
  await updates(page).getByRole("button", { name: "Проверить обновления" }).click();
  await expect(updates(page).getByText("Установлена последняя версия v1.0.0")).toBeVisible();
});

test("проверка → доступно → скачивание → готово с командой", async ({ page }) => {
  await open(page, "/settings", { lang: "ru", format: "deb" });
  const box = updates(page);
  await box.getByRole("button", { name: "Проверить обновления" }).click();
  await expect(box.getByText(/Доступна версия v1\.1\.0/)).toBeVisible();
  await box.getByRole("button", { name: "Скачать" }).click();
  await expect(box.getByRole("progressbar")).toBeVisible();
  await expect(box.getByText(/Пакет v1\.1\.0 скачан и сверен\. Установите его командой ниже/)).toBeVisible();
  await expect(box.locator("code")).toHaveText(/^sudo apt install .+\.deb$/);
  await expect(box.getByRole("button", { name: "Открыть папку с пакетом" })).toBeEnabled();
});

test("downloading: отмена даёт ошибку из словаря и повтор", async ({ page }) => {
  await open(page, "/settings", { lang: "ru", update: "available", delay: "1000" });
  const box = updates(page);
  await box.getByRole("button", { name: "Скачать" }).click();
  await box.getByRole("button", { name: "Отменить" }).click();
  await expect(box.getByText("Скачивание отменено")).toBeVisible();
  await expect(box.getByRole("button", { name: "Повторить" })).toBeVisible();
});

test("failed: текст ошибки по коду", async ({ page }) => {
  await open(page, "/settings", { lang: "en", update: "failed", error: "update.rate-limited" });
  await expect(
    updates(page).getByText("GitHub is temporarily limiting requests; try again in an hour"),
  ).toBeVisible();
});

for (const [format, text, button] of [
  ["dmg", /drag the application to Applications/, "Open disk image"],
  ["nsis", /The installer will replace the application/, "Run installer"],
  ["appimage", /The AppImage file will be replaced/, "Install and restart"],
  ["exe", /replace its file with the downloaded one/, "Show file"],
] as const) {
  test(`ready (${format}): подсказка и кнопка установки`, async ({ page }) => {
    await open(page, "/settings", { lang: "en", update: "ready", format });
    await expect(updates(page).getByText(text)).toBeVisible();
    await expect(updates(page).getByRole("button", { name: button })).toBeEnabled();
  });
}
