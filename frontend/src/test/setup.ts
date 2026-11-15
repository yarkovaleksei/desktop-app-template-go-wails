// Общая подготовка тестов Vitest.
import "@testing-library/jest-dom/vitest";
import { cleanup } from "@testing-library/react";
import { afterAll, afterEach } from "vitest";

// В jsdom нет matchMedia, а useMediaQuery MUI его вызывает: система «в светлой теме».
// Тесты с окружением node (i18n-lint.test.ts) окна не имеют.
if (typeof window !== "undefined" && !window.matchMedia) {
  window.matchMedia = (query: string): MediaQueryList => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: () => {},
    removeListener: () => {},
    addEventListener: () => {},
    removeEventListener: () => {},
    dispatchEvent: () => false,
  });
}

afterEach(() => {
  if (typeof window === "undefined") return;
  cleanup();
  window.location.hash = "";
});

// @wailsio/runtime при импорте запускает в окне опрос раз в 50 мс (drag.js), и
// его первый тик обращается к window. Если файл тестов закончился раньше, тик
// приходит после сноса jsdom и роняет прогон ошибкой «window is not defined»
// (на медленных машинах CI). Ждём два периода опроса, пока окно живо: первый
// тик сам останавливает опрос.
afterAll(async () => {
  if (typeof window === "undefined") return;
  await new Promise((resolve) => setTimeout(resolve, 100));
});
