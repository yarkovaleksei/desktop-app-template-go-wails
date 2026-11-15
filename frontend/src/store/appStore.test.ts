import { describe, expect, it } from "vitest";
import { AppError, UpdateState } from "../api";
import { createFakeSource, fakeInfo, statusOf } from "../test/fakeSource";
import { createAppStore, defaultSettings } from "./appStore";

const setup = async (init?: Parameters<typeof createFakeSource>[0]) => {
  const fake = createFakeSource(init);
  const store = createAppStore(fake.source);
  await store.getState().init();
  return { ...fake, store, s: () => store.getState() };
};

describe("загрузка", () => {
  it("читает сведения, настройки, автозапуск и статус обновления", async () => {
    const settings = { ...defaultSettings, background: true, ui: { theme: "dark", language: "ru" } };
    const { s } = await setup({ settings, autostart: true });
    expect(s().ready).toBe(true);
    expect(s().info).toEqual(fakeInfo);
    expect(s().settings).toEqual(settings);
    expect(s().settingsBroken).toBe(false);
    expect(s().autostart).toBe(true);
    expect(s().update?.state).toBe(UpdateState.StateIdle);
    expect(s().notice).toBeNull();
  });

  it("испорченный файл настроек — признак broken", async () => {
    const { s } = await setup({ broken: true });
    expect(s().settingsBroken).toBe(true);
  });

  it("ошибка загрузки — уведомление кодом, остальное загружается", async () => {
    const fake = createFakeSource();
    fake.source.getAutostart.mockRejectedValueOnce(new AppError("autostart.failed", { detail: "dbus" }));
    const store = createAppStore(fake.source);
    await store.getState().init();
    expect(store.getState().autostart).toBeNull();
    expect(store.getState().info).toEqual(fakeInfo);
    expect(store.getState().notice).toMatchObject({ code: "autostart.failed", params: { detail: "dbus" } });
    expect(store.getState().ready).toBe(true);
  });
});

describe("настройки", () => {
  it("смена темы и языка сохраняется через источник", async () => {
    const { s, source } = await setup();
    await s().setTheme("dark");
    expect(source.saveSettings).toHaveBeenLastCalledWith({
      ...defaultSettings,
      ui: { theme: "dark", language: "" },
    });
    await s().setLanguage("en");
    expect(source.saveSettings).toHaveBeenLastCalledWith({
      ...defaultSettings,
      ui: { theme: "dark", language: "en" },
    });
    expect(s().settings.ui).toEqual({ theme: "dark", language: "en" });
  });

  it("переключатели фона и автопроверки", async () => {
    const { s, source } = await setup();
    await s().setBackground(true);
    await s().setAutoCheckUpdates(true);
    expect(source.saveSettings).toHaveBeenLastCalledWith({
      ...defaultSettings,
      background: true,
      autoCheckUpdates: true,
    });
    expect(s().settings.background).toBe(true);
    expect(s().settings.autoCheckUpdates).toBe(true);
  });

  it("автозапуск переключается через источник", async () => {
    const { s, source } = await setup();
    await s().setAutostart(true);
    expect(source.setAutostart).toHaveBeenCalledWith(true);
    expect(s().autostart).toBe(true);
  });

  it("ошибка сохранения откатывает значение и попадает в уведомление кодом", async () => {
    const { s, source } = await setup();
    source.saveSettings.mockRejectedValueOnce(new AppError("settings.broken"));
    await s().setTheme("dark");
    expect(s().settings.ui.theme).toBe("");
    expect(s().notice).toMatchObject({ code: "settings.broken", params: {} });
  });

  it("ошибка автозапуска откатывает переключатель", async () => {
    const { s, source } = await setup();
    source.setAutostart.mockRejectedValueOnce(new AppError("autostart.unavailable"));
    await s().setAutostart(true);
    expect(s().autostart).toBe(false);
    expect(s().notice?.code).toBe("autostart.unavailable");
  });

  it("ошибка без кода — код unknown; повтор даёт новое уведомление", async () => {
    const { s, source } = await setup();
    source.openLogFolder.mockRejectedValue(new Error("boom"));
    await s().openLogFolder();
    const first = s().notice;
    expect(first).toMatchObject({ code: "unknown", params: { detail: "boom" } });
    await s().openLogFolder();
    expect(s().notice?.id).toBeGreaterThan(first?.id ?? 0);
    s().dismissNotice();
    expect(s().notice).toBeNull();
  });
});

describe("события Go", () => {
  it("смена настроек (из трея) обновляет состояние", async () => {
    const { s, emitSettings } = await setup();
    emitSettings({ ...defaultSettings, ui: { theme: "", language: "en" } });
    expect(s().settings.ui.language).toBe("en");
  });

  it("статус обновления: новые ревизии применяются, устаревшие отбрасываются", async () => {
    const { s, emitUpdate } = await setup();
    emitUpdate(statusOf({ revision: 5, state: UpdateState.StateDownloading, received: 10, size: 100 }));
    expect(s().update).toMatchObject({ revision: 5, received: 10 });
    emitUpdate(statusOf({ revision: 4, state: UpdateState.StateAvailable }));
    expect(s().update?.state).toBe(UpdateState.StateDownloading);
    emitUpdate(statusOf({ revision: 6, state: UpdateState.StateReady }));
    expect(s().update?.state).toBe(UpdateState.StateReady);
  });

  it("ответ вызова старше события не затирает его", async () => {
    const { s, source, emitUpdate } = await setup();
    source.checkUpdate.mockImplementationOnce(async () => {
      emitUpdate(statusOf({ revision: 9, state: UpdateState.StateAvailable }));
      return statusOf({ revision: 8, state: UpdateState.StateChecking });
    });
    await s().checkUpdate();
    expect(s().update?.state).toBe(UpdateState.StateAvailable);
  });

  it("dispose отписывает от событий", async () => {
    const { s, emitSettings } = await setup();
    s().dispose();
    emitSettings({ ...defaultSettings, background: true });
    expect(s().settings.background).toBe(false);
  });
});

describe("обновления", () => {
  it("проверка, скачивание, отмена и установка вызывают источник", async () => {
    const { s, source } = await setup();
    await s().checkUpdate();
    expect(s().update?.state).toBe(UpdateState.StateChecking);
    await s().downloadUpdate();
    expect(s().update?.state).toBe(UpdateState.StateDownloading);
    await s().cancelUpdate();
    await s().installUpdate();
    expect(source.cancelUpdate).toHaveBeenCalledOnce();
    expect(source.installUpdate).toHaveBeenCalledOnce();
  });

  it("ошибка установки — уведомление кодом", async () => {
    const { s, source } = await setup();
    source.installUpdate.mockRejectedValueOnce(new AppError("update.not-ready"));
    await s().installUpdate();
    expect(s().notice?.code).toBe("update.not-ready");
  });

  it("copyText сообщает об успехе", async () => {
    const { s, source } = await setup();
    await expect(s().copyText("sudo apt install ./x.deb")).resolves.toBe(true);
    source.copyText.mockRejectedValueOnce(new Error("denied"));
    await expect(s().copyText("x")).resolves.toBe(false);
    expect(s().notice?.code).toBe("unknown");
  });
});
