// Рендер интерфейса для тестов: стор поверх fakeSource, загрузка до первой проверки.
import { render, screen } from "@testing-library/react";
import { App } from "../app/App";
import { createAppStore } from "../store/appStore";
import { createFakeSource } from "./fakeSource";

/** Русский интерфейс по умолчанию: jsdom сообщает системный язык en-US. */
export const ruSettings = {
  ui: { theme: "", language: "ru" },
  background: false,
  autoCheckUpdates: false,
};

/** Полное приложение с маршрутом route; ждёт, пока стор загрузится. */
export async function renderApp(
  init: Parameters<typeof createFakeSource>[0] = { settings: ruSettings },
  route = "/",
) {
  window.location.hash = `#${route}`;
  const fake = createFakeSource(init);
  const store = createAppStore(fake.source);
  await store.getState().init();
  const view = render(<App source={fake.source} store={store} />);
  await screen.findByRole("heading", { level: 1 });
  return { ...fake, store, view };
}
