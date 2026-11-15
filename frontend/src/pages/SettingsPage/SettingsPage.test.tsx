import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { renderApp, ruSettings } from "../../test/render";

describe("SettingsPage", () => {
  it("без трея работа в фоне недоступна с пояснением", async () => {
    await renderApp(
      { settings: { ...ruSettings, background: true }, info: { trayAvailable: false } },
      "/settings",
    );
    const sw = screen.getByRole("switch", { name: "Работать в фоне при закрытии окна" });
    expect(sw).toBeDisabled();
    expect(sw).not.toBeChecked();
    expect(sw).toHaveAccessibleDescription(/нет области уведомлений/);
  });

  it("с треем переключатель фона сохраняет настройку", async () => {
    const { source } = await renderApp(undefined, "/settings");
    const sw = screen.getByRole("switch", { name: "Работать в фоне при закрытии окна" });
    expect(sw).toBeEnabled();
    await userEvent.click(sw);
    expect(source.saveSettings).toHaveBeenLastCalledWith(expect.objectContaining({ background: true }));
    expect(sw).toBeChecked();
  });

  it("автозапуск и автопроверка вызывают источник", async () => {
    const { source } = await renderApp(undefined, "/settings");
    await userEvent.click(screen.getByRole("switch", { name: "Запускать при входе в систему" }));
    expect(source.setAutostart).toHaveBeenCalledWith(true);
    const auto = screen.getByRole("switch", { name: "Проверять обновления автоматически" });
    expect(auto).toHaveAccessibleDescription(/раз в сутки/);
    await userEvent.click(auto);
    expect(source.saveSettings).toHaveBeenLastCalledWith(expect.objectContaining({ autoCheckUpdates: true }));
  });

  it("испорченный файл настроек — предупреждение", async () => {
    await renderApp({ settings: ruSettings, broken: true }, "/settings");
    expect(screen.getByText(/Файл настроек повреждён или записан более новой версией/)).toBeInTheDocument();
  });

  it("версия, формат поставки и папка журнала", async () => {
    const { source } = await renderApp({ settings: ruSettings, info: { format: "" } }, "/settings");
    expect(screen.getByText(/Формат поставки: сборка из исходников/)).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "Открыть папку журнала" }));
    expect(source.openLogFolder).toHaveBeenCalledOnce();
  });
});
