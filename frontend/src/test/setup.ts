// Общая подготовка тестов Vitest.
import "@testing-library/jest-dom/vitest";
import { cleanup } from "@testing-library/react";
import { afterEach } from "vitest";

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
