// DataSource на данных в памяти: интерфейс без Go — `yarn dev` в браузере и сценарии
// Playwright. Сеть, файлы и ОС не трогаются; обновление «скачивается» по таймеру.
import { AppError } from "./errors";
import { Language, Theme, UpdateFormat, UpdateInstall, UpdateState } from "./source";
import type { DataSource, Info, Settings, SettingsView, Unsubscribe, UpdateStatus } from "./source";

/** Начальное состояние моков; всё необязательно. */
export interface MockOptions {
  /** Формат сборки; пусто — сборка из исходников, обновление недоступно. */
  format?: UpdateFormat;
  trayAvailable?: boolean;
  settings?: Partial<Settings>;
  /** Файл настроек испорчен: показаны умолчания, сохранение отказывает. */
  broken?: boolean;
  autostart?: boolean;
  /** Запуск при входе в систему не поддерживается: setAutostart отказывает. */
  autostartUnavailable?: boolean;
  /** Версия последнего выпуска; совпадает с текущей — новее нет. */
  latest?: string;
  /** Проверка обновлений закончится ошибкой с этим кодом. */
  updateError?: string;
  /** Статус обновления сразу после запуска (например ready для показа команды). */
  updateState?: UpdateState;
  /** Задержка шагов проверки и скачивания, мс. */
  delay?: number;
}

const VERSION = "v1.0.0";

const PACKAGE_SIZE = 48 * 1024 * 1024;
const DOWNLOAD_STEPS = 5;

const defaultSettings = (): Settings => ({
  ui: { theme: Theme.System, language: Language.System },
  background: false,
  autoCheckUpdates: false,
});

/**
 * Начальное состояние моков из параметров адреса — так сценарии Playwright задают
 * его без правки кода: `/?lang=en&theme=dark&update=ready&format=dmg#/settings`.
 * Параметры: lang, theme (настройки), format (пусто — сборка из исходников),
 * tray=0, broken=1, autostart=1, autostart=unavailable, background=1, autocheck=1,
 * latest, update (состояние), error (код ошибки проверки), delay (мс).
 */
export function mockOptionsFromSearch(search: string): MockOptions {
  const q = new URLSearchParams(search);
  const opts: MockOptions = {};
  const settings: Partial<Settings> = {};
  const ui: Partial<Settings["ui"]> = {};
  const lang = q.get("lang");
  if (lang !== null) ui.language = lang;
  const theme = q.get("theme");
  if (theme !== null) ui.theme = theme;
  if (lang !== null || theme !== null) settings.ui = { ...defaultSettings().ui, ...ui };
  if (q.has("background")) settings.background = q.get("background") === "1";
  if (q.has("autocheck")) settings.autoCheckUpdates = q.get("autocheck") === "1";
  if (Object.keys(settings).length > 0) opts.settings = settings;
  const format = q.get("format");
  if (format !== null) opts.format = format as UpdateFormat;
  if (q.get("tray") === "0") opts.trayAvailable = false;
  if (q.get("broken") === "1") opts.broken = true;
  if (q.get("autostart") === "1") opts.autostart = true;
  if (q.get("autostart") === "unavailable") opts.autostartUnavailable = true;
  const latest = q.get("latest");
  if (latest !== null) opts.latest = latest;
  const update = q.get("update");
  if (update !== null) opts.updateState = update as UpdateState;
  const error = q.get("error");
  if (error !== null) opts.updateError = error;
  const delay = Number(q.get("delay"));
  if (q.has("delay") && Number.isFinite(delay)) opts.delay = delay;
  return opts;
}

const themes: string[] = Object.values(Theme);
const languages: string[] = Object.values(Language);

/** Как ставится пакет и чем: то же соответствие, что в updater.installFor. */
const installFor = (format: UpdateFormat): { install: UpdateInstall; command?: string } => {
  switch (format) {
    case UpdateFormat.FormatDeb:
      return { install: UpdateInstall.InstallCommand, command: "sudo apt install" };
    case UpdateFormat.FormatRPM:
      return { install: UpdateInstall.InstallCommand, command: "sudo dnf install" };
    case UpdateFormat.FormatArch:
      return { install: UpdateInstall.InstallCommand, command: "sudo pacman -U" };
    case UpdateFormat.FormatAppImage:
      return { install: UpdateInstall.InstallReplace };
    case UpdateFormat.FormatDMG:
      return { install: UpdateInstall.InstallOpen };
    case UpdateFormat.FormatNSIS:
      return { install: UpdateInstall.InstallRun };
    default:
      return { install: UpdateInstall.InstallReveal };
  }
};

const packageName = (format: UpdateFormat, version: string): string => {
  const v = version.replace(/^v/, "");
  switch (format) {
    case UpdateFormat.FormatRPM:
      return `desktop-app-${v}-1.x86_64.rpm`;
    case UpdateFormat.FormatArch:
      return `desktop-app-${v}-1-x86_64.pkg.tar.zst`;
    case UpdateFormat.FormatAppImage:
      return `desktop-app-${v}-x86_64.AppImage`;
    case UpdateFormat.FormatDMG:
      return `desktop-app-${v}-universal.dmg`;
    case UpdateFormat.FormatNSIS:
      return `desktop-app-${v}-amd64-installer.exe`;
    case UpdateFormat.FormatExe:
      return `desktop-app-${v}-amd64.exe`;
    default:
      return `desktop-app_${v}_amd64.deb`;
  }
};

const sleep = (ms: number) => new Promise<void>((resolve) => setTimeout(resolve, ms));

/** Создаёт источник моков с начальным состоянием opts. */
export function createMockSource(opts: MockOptions = {}): DataSource {
  const format = opts.format ?? UpdateFormat.FormatDeb;
  const latest = opts.latest ?? "v1.1.0";
  const delay = opts.delay ?? 400;

  const info: Info = {
    name: "Desktop App",
    version: VERSION,
    format,
    os: "linux",
    arch: "amd64",
    dirs: {
      config: "/home/user/.config/desktop-app",
      data: "/home/user/.local/share/desktop-app",
      cache: "/home/user/.cache/desktop-app",
      logs: "/home/user/.local/state/desktop-app/logs",
    },
    trayAvailable: opts.trayAvailable ?? true,
  };

  const base = defaultSettings();
  let settings: Settings = { ...base, ...opts.settings, ui: { ...base.ui, ...opts.settings?.ui } };
  const broken = opts.broken ?? false;
  let autostart = opts.autostart ?? false;

  const updateListeners = new Set<(s: UpdateStatus) => void>();
  const settingsListeners = new Set<(s: Settings) => void>();
  const navigateListeners = new Set<(path: string) => void>();

  let status: UpdateStatus = { revision: 1, state: UpdateState.StateIdle, current: VERSION, format };
  let download: ReturnType<typeof setInterval> | undefined;

  const setStatus = (next: Omit<UpdateStatus, "revision" | "current" | "format">) => {
    status = { ...next, revision: status.revision + 1, current: VERSION, format };
    for (const cb of updateListeners) cb(status);
  };

  const available = (): Omit<UpdateStatus, "revision" | "current" | "format"> => ({
    state: UpdateState.StateAvailable,
    latest,
    notesUrl: `https://github.com/example/desktop-app/releases/tag/${latest}`,
    package: packageName(format, latest),
    size: PACKAGE_SIZE,
  });

  const ready = (): Omit<UpdateStatus, "revision" | "current" | "format"> => {
    const how = installFor(format);
    const pkg = packageName(format, latest);
    return {
      ...available(),
      state: UpdateState.StateReady,
      received: PACKAGE_SIZE,
      install: how.install,
      ...(how.command ? { command: `${how.command} /home/user/.cache/desktop-app/updates/${pkg}` } : {}),
    };
  };

  if (format === UpdateFormat.FormatNone) {
    status = { ...status, state: UpdateState.StateUnsupported };
  } else if (opts.updateState === UpdateState.StateAvailable) {
    setStatus(available());
  } else if (opts.updateState === UpdateState.StateReady) {
    setStatus(ready());
  } else if (opts.updateState === UpdateState.StateCurrent) {
    setStatus({ state: UpdateState.StateCurrent, latest: VERSION });
  } else if (opts.updateState === UpdateState.StateFailed) {
    setStatus({ state: UpdateState.StateFailed, error: { code: opts.updateError ?? "update.network" } });
  } else if (opts.updateState === UpdateState.StateDownloading) {
    setStatus({ ...available(), state: UpdateState.StateDownloading, received: PACKAGE_SIZE / 2 });
  }

  const view = (): SettingsView => ({ settings, broken });

  const subscribe = <T>(set: Set<(v: T) => void>, cb: (v: T) => void): Unsubscribe => {
    set.add(cb);
    return () => set.delete(cb);
  };

  const stopDownload = () => {
    clearInterval(download);
    download = undefined;
  };

  return {
    kind: "mock",

    info: async () => info,
    frontendReady: async () => {},
    previousCrash: async () => null,
    reportError: async (r) => {
      console.warn("mock: отчёт об ошибке интерфейса", r);
    },
    openLogFolder: async () => {},

    getSettings: async () => view(),
    saveSettings: async (s) => {
      if (broken) throw new AppError("settings.broken");
      if (!themes.includes(s.ui.theme)) throw new AppError("settings.invalid-theme", { value: s.ui.theme });
      if (!languages.includes(s.ui.language)) {
        throw new AppError("settings.invalid-language", { value: s.ui.language });
      }
      settings = structuredClone(s);
      for (const cb of settingsListeners) cb(structuredClone(settings));
      return view();
    },
    getAutostart: async () => autostart,
    setAutostart: async (on) => {
      if (opts.autostartUnavailable) throw new AppError("autostart.unavailable");
      autostart = on;
    },

    updateStatus: async () => status,
    checkUpdate: async () => {
      if (format === UpdateFormat.FormatNone) return status;
      setStatus({ state: UpdateState.StateChecking });
      await sleep(delay);
      if (opts.updateError) {
        setStatus({ state: UpdateState.StateFailed, error: { code: opts.updateError } });
      } else if (latest === VERSION) {
        setStatus({ state: UpdateState.StateCurrent, latest });
      } else {
        setStatus(available());
      }
      return status;
    },
    downloadUpdate: async () => {
      if (status.state !== UpdateState.StateAvailable && status.state !== UpdateState.StateFailed) {
        return status;
      }
      let step = 0;
      setStatus({ ...available(), state: UpdateState.StateDownloading, received: 0 });
      download = setInterval(() => {
        step += 1;
        if (step >= DOWNLOAD_STEPS) {
          stopDownload();
          setStatus(ready());
          return;
        }
        setStatus({
          ...available(),
          state: UpdateState.StateDownloading,
          received: (PACKAGE_SIZE * step) / DOWNLOAD_STEPS,
        });
      }, delay);
      return status;
    },
    cancelUpdate: async () => {
      if (download === undefined) return;
      stopDownload();
      setStatus({ ...available(), state: UpdateState.StateFailed, error: { code: "update.canceled" } });
    },
    installUpdate: async () => {
      if (status.state !== UpdateState.StateReady) throw new AppError("update.not-ready");
    },

    openURL: async (url) => {
      window.open(url, "_blank", "noopener");
    },
    copyText: async (text) => {
      await navigator.clipboard?.writeText(text);
    },

    onUpdateStatus: (cb) => subscribe(updateListeners, cb),
    onSettingsChanged: (cb) => subscribe(settingsListeners, cb),
    onNavigate: (cb) => subscribe(navigateListeners, cb),
  };
}
