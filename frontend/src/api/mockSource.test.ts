import { describe, expect, it, vi } from "vitest";
import { UpdateFormat, UpdateInstall, UpdateState } from "./source";
import { createMockSource, mockOptionsFromSearch } from "./mockSource";

describe("mockOptionsFromSearch", () => {
  it("пустой адрес — пустые опции", () => {
    expect(mockOptionsFromSearch("")).toEqual({});
  });

  it("разбирает все параметры", () => {
    expect(
      mockOptionsFromSearch(
        "?lang=en&theme=dark&background=1&autocheck=1&format=dmg&tray=0&broken=1&autostart=unavailable&latest=v2.0.0&update=ready&error=update.network&delay=5",
      ),
    ).toEqual({
      settings: { ui: { language: "en", theme: "dark" }, background: true, autoCheckUpdates: true },
      format: "dmg",
      trayAvailable: false,
      broken: true,
      autostartUnavailable: true,
      latest: "v2.0.0",
      updateState: "ready",
      updateError: "update.network",
      delay: 5,
    });
  });

  it("format= — сборка из исходников", () => {
    expect(mockOptionsFromSearch("?format=").format).toBe("");
  });
});

describe("createMockSource", () => {
  it("проводит обновление до ready по таймеру", async () => {
    vi.useFakeTimers();
    const src = createMockSource({ format: UpdateFormat.FormatRPM, delay: 10 });
    const states: string[] = [];
    src.onUpdateStatus((s) => states.push(s.state));
    const check = src.checkUpdate();
    await vi.advanceTimersByTimeAsync(10);
    expect((await check).state).toBe(UpdateState.StateAvailable);
    await src.downloadUpdate();
    await vi.advanceTimersByTimeAsync(100);
    const st = await src.updateStatus();
    expect(st.state).toBe(UpdateState.StateReady);
    expect(st.install).toBe(UpdateInstall.InstallCommand);
    expect(st.command).toMatch(/^sudo dnf install .+\.rpm$/);
    expect(states).toContain(UpdateState.StateDownloading);
    vi.useRealTimers();
  });

  it("сборка из исходников — unsupported", async () => {
    const src = createMockSource({ format: UpdateFormat.FormatNone });
    expect((await src.updateStatus()).state).toBe(UpdateState.StateUnsupported);
  });

  it("отказ сохранения при испорченном файле", async () => {
    const src = createMockSource({ broken: true });
    const { settings } = await src.getSettings();
    await expect(src.saveSettings(settings)).rejects.toMatchObject({ code: "settings.broken" });
  });
});
