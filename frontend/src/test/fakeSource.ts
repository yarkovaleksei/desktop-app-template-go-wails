// DataSource для тестов: каждый метод — vi.fn с поведением по умолчанию поверх
// состояния в памяти, события отправляет сам тест через emit*.
import { vi } from "vitest";
import {
  UpdateFormat,
  UpdateState,
  type DataSource,
  type Info,
  type Settings,
  type SettingsView,
  type UpdateStatus,
} from "../api";
import { defaultSettings } from "../store/appStore";

export const fakeInfo: Info = {
  name: "Desktop App",
  version: "v1.0.0",
  format: UpdateFormat.FormatDeb,
  os: "linux",
  arch: "amd64",
  dirs: { config: "/c", data: "/d", cache: "/k", logs: "/l" },
  trayAvailable: true,
};

export const statusOf = (s: Partial<UpdateStatus> & { revision: number }): UpdateStatus => ({
  state: UpdateState.StateIdle,
  current: "v1.0.0",
  format: UpdateFormat.FormatDeb,
  ...s,
});

/** Источник с данными по умолчанию; init — начальные значения. */
export function createFakeSource(
  init: { info?: Partial<Info>; settings?: Settings; broken?: boolean; autostart?: boolean } = {},
) {
  let settings: Settings = structuredClone(init.settings ?? defaultSettings);
  const broken = init.broken ?? false;
  let autostart = init.autostart ?? false;
  const update = new Set<(s: UpdateStatus) => void>();
  const changed = new Set<(s: Settings) => void>();
  const navigate = new Set<(p: string) => void>();
  const view = (): SettingsView => ({ settings: structuredClone(settings), broken });
  const sub =
    <T>(set: Set<(v: T) => void>) =>
    (cb: (v: T) => void) => {
      set.add(cb);
      return () => {
        set.delete(cb);
      };
    };

  const source = {
    kind: "mock" as const,
    info: vi.fn(async () => ({ ...fakeInfo, ...init.info })),
    frontendReady: vi.fn(async () => {}),
    previousCrash: vi.fn(async () => null),
    reportError: vi.fn(async () => {}),
    openLogFolder: vi.fn(async () => {}),
    getSettings: vi.fn(async () => view()),
    saveSettings: vi.fn(async (s: Settings) => {
      settings = structuredClone(s);
      for (const cb of changed) cb(structuredClone(settings));
      return view();
    }),
    getAutostart: vi.fn(async () => autostart),
    setAutostart: vi.fn(async (on: boolean) => {
      autostart = on;
    }),
    updateStatus: vi.fn(async () => statusOf({ revision: 1 })),
    checkUpdate: vi.fn(async () => statusOf({ revision: 2, state: UpdateState.StateChecking })),
    downloadUpdate: vi.fn(async () => statusOf({ revision: 3, state: UpdateState.StateDownloading })),
    cancelUpdate: vi.fn(async () => {}),
    installUpdate: vi.fn(async () => {}),
    openURL: vi.fn(async () => {}),
    copyText: vi.fn(async () => {}),
    onUpdateStatus: vi.fn(sub(update)),
    onSettingsChanged: vi.fn(sub(changed)),
    onNavigate: vi.fn(sub(navigate)),
  } satisfies DataSource;

  return {
    /** Источник: методы — vi.fn, их можно переопределять в тесте. */
    source,
    emitUpdate: (s: UpdateStatus) => update.forEach((cb) => cb(s)),
    emitSettings: (s: Settings) => changed.forEach((cb) => cb(s)),
    emitNavigate: (path: string) => navigate.forEach((cb) => cb(path)),
  };
}
