import { act, render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { AppError } from "../api";
import { renderApp, ruSettings } from "../test/render";
import { I18nProvider } from "../i18n";
import { ErrorBoundary, installErrorReporting } from "./ErrorBoundary";

const heading = () => screen.getByRole("heading", { level: 1 });

describe("навигация", () => {
  it("переход между страницами по вкладкам", async () => {
    await renderApp();
    expect(heading()).toHaveTextContent("Desktop App");
    await userEvent.click(screen.getByRole("tab", { name: "Настройки" }));
    expect(heading()).toHaveTextContent("Настройки");
    await userEvent.click(screen.getByRole("tab", { name: "Главная" }));
    expect(heading()).toHaveTextContent("Desktop App");
  });

  it("событие app:navigate открывает страницу", async () => {
    const { emitNavigate } = await renderApp();
    act(() => emitNavigate("/settings"));
    await waitFor(() => expect(heading()).toHaveTextContent("Настройки"));
    act(() => emitNavigate("/"));
    await waitFor(() => expect(heading()).toHaveTextContent("Desktop App"));
  });

  it("неизвестный адрес ведёт на главную", async () => {
    await renderApp(undefined, "/nope");
    expect(heading()).toHaveTextContent("Desktop App");
  });

  it("FrontendReady вызывается один раз после загрузки", async () => {
    const { source } = await renderApp();
    await waitFor(() => expect(source.frontendReady).toHaveBeenCalledOnce());
  });
});

describe("язык и тема", () => {
  it("смена языка перерисовывает интерфейс и сохраняется", async () => {
    const { source } = await renderApp(undefined, "/settings");
    await userEvent.click(screen.getByRole("button", { name: "English" }));
    expect(heading()).toHaveTextContent("Settings");
    expect(screen.getByRole("tab", { name: "Home" })).toBeInTheDocument();
    expect(document.documentElement.lang).toBe("en");
    expect(source.saveSettings).toHaveBeenLastCalledWith(
      expect.objectContaining({ ui: { theme: "", language: "en" } }),
    );
  });

  it("смена темы применяется и сохраняется", async () => {
    const { source } = await renderApp(undefined, "/settings");
    const before = getComputedStyle(document.body).backgroundColor;
    await userEvent.click(screen.getByRole("button", { name: "Тёмная" }));
    expect(screen.getByRole("button", { name: "Тёмная" })).toHaveAttribute("aria-pressed", "true");
    expect(source.saveSettings).toHaveBeenLastCalledWith(
      expect.objectContaining({ ui: { theme: "dark", language: "ru" } }),
    );
    await waitFor(() => expect(getComputedStyle(document.body).backgroundColor).not.toBe(before));
  });

  it("смена настроек из трея меняет язык сразу", async () => {
    const { emitSettings } = await renderApp(undefined, "/settings");
    act(() => emitSettings({ ...ruSettings, ui: { theme: "", language: "en" } }));
    await waitFor(() => expect(heading()).toHaveTextContent("Settings"));
  });
});

describe("ошибки", () => {
  it("текст ошибки сохранения берётся из словаря по коду", async () => {
    const { source } = await renderApp(undefined, "/settings");
    source.saveSettings.mockRejectedValueOnce(new AppError("settings.invalid-theme", { value: "pink" }));
    await userEvent.click(screen.getByRole("button", { name: "Тёмная" }));
    expect(await screen.findByRole("alert")).toHaveTextContent("Неизвестная тема «pink»");
    expect(screen.getByRole("button", { name: "Тёмная" })).toHaveAttribute("aria-pressed", "false");
  });

  it("неизвестный код — общий текст с кодом", async () => {
    const { source } = await renderApp(undefined, "/settings");
    source.openLogFolder.mockRejectedValueOnce(new AppError("app.something-new"));
    await userEvent.click(screen.getByRole("button", { name: "Открыть папку журнала" }));
    expect(await screen.findByRole("alert")).toHaveTextContent("Неизвестная ошибка: app.something-new");
  });

  it("сообщение о прошлом сбое", async () => {
    await renderApp({ settings: ruSettings, crash: { startedAt: "" } });
    expect(screen.getByText("Прошлый сеанс завершился неожиданно.")).toBeInTheDocument();
  });

  it("сообщение о прошлом сбое со временем начала", async () => {
    await renderApp({ settings: ruSettings, crash: { startedAt: "2026-11-15T09:05:00Z" } });
    expect(screen.getByText(/Прошлый сеанс, начатый .*2026.*, завершился неожиданно/)).toBeInTheDocument();
  });

  it("ошибка рендера показывает экран сбоя и уходит в журнал", () => {
    const onError = vi.fn();
    const Boom = () => {
      throw new Error("бах");
    };
    vi.spyOn(console, "error").mockImplementation(() => {});
    render(
      <I18nProvider lang="ru">
        <ErrorBoundary onError={onError}>
          <Boom />
        </ErrorBoundary>
      </I18nProvider>,
    );
    expect(screen.getByRole("heading", { name: "Что-то пошло не так" })).toBeInTheDocument();
    expect(onError).toHaveBeenCalledWith(
      expect.objectContaining({ source: "render", message: "Error: бах" }),
    );
    vi.restoreAllMocks();
  });

  it("ошибки окна и отклонённые промисы уходят в журнал", () => {
    const report = vi.fn();
    const off = installErrorReporting(report);
    window.dispatchEvent(new ErrorEvent("error", { error: new TypeError("x") }));
    const rejection = new Event("unhandledrejection") as PromiseRejectionEvent;
    Object.defineProperty(rejection, "reason", { value: "отказ" });
    window.dispatchEvent(rejection);
    off();
    expect(report).toHaveBeenCalledWith(
      expect.objectContaining({ source: "error", message: "TypeError: x" }),
    );
    expect(report).toHaveBeenCalledWith(
      expect.objectContaining({ source: "unhandledrejection", message: "отказ" }),
    );
  });
});

describe("главная", () => {
  it("ссылка на документацию открывается через источник", async () => {
    const { source } = await renderApp();
    await userEvent.click(screen.getByRole("button", { name: "Документация проекта" }));
    expect(source.openURL).toHaveBeenCalledWith(expect.stringContaining("github.com/"));
    expect(within(screen.getByRole("main")).getByText("Окно открыто 0 минут")).toBeInTheDocument();
  });
});
