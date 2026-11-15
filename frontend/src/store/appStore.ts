// Один стор приложения (zustand vanilla). Источник данных передаётся в конструктор:
// в окне — wailsSource, в браузере — mockSource, в тестах — fakeSource.
// Состояние живёт в Go; стор держит последний известный снимок и шлёт изменения
// через DataSource, а события Go (смена настроек из трея, статус обновления)
// обновляют снимок.
import { createStore, type StoreApi } from "zustand/vanilla";
import {
  Language,
  Theme,
  toAppError,
  type Crash,
  type DataSource,
  type FrontendError,
  type Info,
  type Settings,
  type Unsubscribe,
  type UpdateStatus,
} from "../api";

/** Ошибка для уведомления: код и параметры для словаря (errors.<код>). */
export interface ErrorNotice {
  /** растёт на каждое уведомление: повтор той же ошибки показывается заново */
  id: number;
  code: string;
  params: Record<string, string>;
}

export interface AppState {
  sourceKind: DataSource["kind"];
  /** первая загрузка закончилась (успешно или с ошибкой в notice) */
  ready: boolean;
  info: Info | null;
  /** прошлый сеанс оборвался; null — завершился штатно */
  previousCrash: Crash | null;
  settings: Settings;
  /** файл настроек испорчен: показаны умолчания, сохранение откажет */
  settingsBroken: boolean;
  /** запуск при входе в систему; null — состояние не прочитано */
  autostart: boolean | null;
  update: UpdateStatus | null;
  notice: ErrorNotice | null;

  /** Загрузить всё и подписаться на события Go. */
  init(): Promise<void>;
  /** Отписаться от событий. */
  dispose(): void;

  setTheme(theme: string): Promise<void>;
  setLanguage(language: string): Promise<void>;
  setBackground(on: boolean): Promise<void>;
  setAutoCheckUpdates(on: boolean): Promise<void>;
  setAutostart(on: boolean): Promise<void>;

  checkUpdate(): Promise<void>;
  downloadUpdate(): Promise<void>;
  cancelUpdate(): Promise<void>;
  installUpdate(): Promise<void>;

  openLogFolder(): Promise<void>;
  openURL(url: string): Promise<void>;
  /** Скопировать текст; false — не удалось (ошибка уже в notice). */
  copyText(text: string): Promise<boolean>;
  /** Отправить отчёт о необработанной ошибке в журнал Go; сбой отправки не показывается. */
  reportError(r: FrontendError): Promise<void>;

  /** Показать ошибку в уведомлении. */
  showError(err: unknown): void;
  dismissNotice(): void;
}

export type AppStore = StoreApi<AppState>;

/** Настройки до загрузки и при испорченном файле — как умолчания Go. */
export const defaultSettings: Settings = {
  ui: { theme: Theme.System, language: Language.System },
  background: false,
  autoCheckUpdates: false,
};

/** Создаёт стор поверх источника данных. */
export function createAppStore(source: DataSource): AppStore {
  let unsubscribe: Unsubscribe[] = [];
  let noticeId = 0;

  return createStore<AppState>()((set, get) => {
    const showError = (err: unknown) => {
      const e = toAppError(err);
      noticeId += 1;
      set({ notice: { id: noticeId, code: e.code, params: e.params } });
    };

    // Снимки статуса приходят ответами вызовов и событиями в любом порядке:
    // устаревший (с меньшей ревизией) отбрасывается.
    const applyStatus = (s: UpdateStatus) => {
      const cur = get().update;
      if (cur && s.revision < cur.revision) return;
      set({ update: s });
    };

    // Сохранить настройки: интерфейс меняется сразу, при отказе — откат и уведомление.
    const saveSettings = async (next: Settings) => {
      const prev = get().settings;
      set({ settings: next });
      try {
        const view = await source.saveSettings(next);
        set({ settings: view.settings, settingsBroken: view.broken });
      } catch (err) {
        set({ settings: prev });
        showError(err);
      }
    };

    const run = async (fn: () => Promise<unknown>) => {
      try {
        await fn();
      } catch (err) {
        showError(err);
      }
    };

    const updateCall = async (fn: () => Promise<UpdateStatus>) => {
      try {
        applyStatus(await fn());
      } catch (err) {
        showError(err);
      }
    };

    return {
      sourceKind: source.kind,
      ready: false,
      info: null,
      previousCrash: null,
      settings: defaultSettings,
      settingsBroken: false,
      autostart: null,
      update: null,
      notice: null,

      async init() {
        get().dispose();
        unsubscribe = [
          source.onUpdateStatus(applyStatus),
          source.onSettingsChanged((settings) => set({ settings })),
        ];
        const [info, view, autostart, update, crash] = await Promise.allSettled([
          source.info(),
          source.getSettings(),
          source.getAutostart(),
          source.updateStatus(),
          source.previousCrash(),
        ]);
        if (info.status === "fulfilled") set({ info: info.value });
        else showError(info.reason);
        if (view.status === "fulfilled")
          set({ settings: view.value.settings, settingsBroken: view.value.broken });
        else showError(view.reason);
        if (autostart.status === "fulfilled") set({ autostart: autostart.value });
        else showError(autostart.reason);
        if (update.status === "fulfilled") applyStatus(update.value);
        else showError(update.reason);
        // Отметка о прошлом сбое не главное: её ошибку не показываем.
        if (crash.status === "fulfilled") set({ previousCrash: crash.value });
        set({ ready: true });
      },

      dispose() {
        for (const off of unsubscribe) off();
        unsubscribe = [];
      },

      setTheme: (theme) => {
        const s = get().settings;
        return saveSettings({ ...s, ui: { ...s.ui, theme } });
      },
      setLanguage: (language) => {
        const s = get().settings;
        return saveSettings({ ...s, ui: { ...s.ui, language } });
      },
      setBackground: (background) => saveSettings({ ...get().settings, background }),
      setAutoCheckUpdates: (autoCheckUpdates) => saveSettings({ ...get().settings, autoCheckUpdates }),

      async setAutostart(on) {
        const prev = get().autostart;
        set({ autostart: on });
        try {
          await source.setAutostart(on);
        } catch (err) {
          set({ autostart: prev });
          showError(err);
        }
      },

      checkUpdate: () => updateCall(() => source.checkUpdate()),
      downloadUpdate: () => updateCall(() => source.downloadUpdate()),
      cancelUpdate: () => run(() => source.cancelUpdate()),
      installUpdate: () => run(() => source.installUpdate()),

      openLogFolder: () => run(() => source.openLogFolder()),
      openURL: (url) => run(() => source.openURL(url)),
      async copyText(text) {
        try {
          await source.copyText(text);
          return true;
        } catch (err) {
          showError(err);
          return false;
        }
      },
      async reportError(r) {
        try {
          await source.reportError(r);
        } catch (err) {
          console.error("отчёт об ошибке не отправлен", err);
        }
      },

      showError,
      dismissNotice: () => set({ notice: null }),
    };
  });
}
