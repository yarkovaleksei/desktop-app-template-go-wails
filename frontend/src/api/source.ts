// DataSource — всё, что интерфейсу нужно от Go. Реализации: wailsSource (привязки
// в окне) и mockSource (данные в памяти для браузера и сценариев Playwright).
// Каждый метод отклоняется только с AppError (см. errors.ts).
import type { Unsubscribe } from "./events";
import type { Crash, FrontendError, Info, Settings, SettingsView, UpdateStatus } from "./bindings";

export type { Unsubscribe } from "./events";
export type {
  Crash,
  Dirs,
  Failure,
  FrontendError,
  Info,
  Settings,
  SettingsView,
  UI,
  UpdateStatus,
} from "./bindings";
export { UpdateFormat, UpdateInstall, UpdateState } from "./bindings";

/** Значения настроек оформления; пустая строка — как в системе. Совпадают с internal/settings. */
export const Theme = { System: "", Light: "light", Dark: "dark" } as const;
export const Language = { System: "", RU: "ru", EN: "en" } as const;

/** Источник данных интерфейса. */
export interface DataSource {
  readonly kind: "wails" | "mock";

  /** Название, версия, формат сборки, ОС, каталоги и наличие трея. */
  info(): Promise<Info>;
  /** Первый экран отрисован; в smoke-режиме это завершает проверку запуска. */
  frontendReady(): Promise<void>;
  /** Прошлый сеанс оборвался; null — завершился штатно. */
  previousCrash(): Promise<Crash | null>;
  /** Записать необработанную ошибку интерфейса в журнал Go. */
  reportError(r: FrontendError): Promise<void>;
  /** Открыть каталог журнала в файловом менеджере. */
  openLogFolder(): Promise<void>;

  /** Текущие настройки и признак испорченного файла. */
  getSettings(): Promise<SettingsView>;
  /** Проверить и сохранить настройки; после сохранения приходит onSettingsChanged. */
  saveSettings(s: Settings): Promise<SettingsView>;
  /** Запуск при входе в систему: состояние читается из ОС. */
  getAutostart(): Promise<boolean>;
  setAutostart(on: boolean): Promise<void>;

  /** Текущий статус обновления, без сети. */
  updateStatus(): Promise<UpdateStatus>;
  /** Спросить о последнем выпуске; ошибка — в поле error статуса. */
  checkUpdate(): Promise<UpdateStatus>;
  /** Начать скачивание в фоне; прогресс приходит в onUpdateStatus. */
  downloadUpdate(): Promise<UpdateStatus>;
  cancelUpdate(): Promise<void>;
  /** Поставить скачанный пакет; некоторые форматы закрывают приложение. */
  installUpdate(): Promise<void>;

  /** Открыть адрес в браузере системы (заметки к выпуску). */
  openURL(url: string): Promise<void>;
  /** Скопировать текст в буфер обмена (команда установки пакета). */
  copyText(text: string): Promise<void>;

  onUpdateStatus(cb: (s: UpdateStatus) => void): Unsubscribe;
  onSettingsChanged(cb: (s: Settings) => void): Unsubscribe;
  /** Go просит открыть страницу: пункт трея, щелчок по уведомлению. */
  onNavigate(cb: (path: string) => void): Unsubscribe;
}
