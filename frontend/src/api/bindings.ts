// Единственная точка импорта сгенерированных привязок Wails и рантайма. Длинные пути
// пакетов Go живут только здесь; остальной код берёт привязки через DataSource.
export {
  AppService,
  SettingsService,
} from "../../bindings/github.com/yarkovaleksei/desktop-app-template-go-wails/internal/app";
export type {
  Crash,
  Dirs,
  FrontendError,
  Info,
  SettingsView,
} from "../../bindings/github.com/yarkovaleksei/desktop-app-template-go-wails/internal/app";
export type {
  Settings,
  UI,
} from "../../bindings/github.com/yarkovaleksei/desktop-app-template-go-wails/internal/settings";
export { UpdatesService } from "../../bindings/github.com/yarkovaleksei/go-desktop-kit/wailsx";
export {
  Format as UpdateFormat,
  Install as UpdateInstall,
  State as UpdateState,
} from "../../bindings/github.com/yarkovaleksei/go-desktop-kit/updater";
export type {
  Failure,
  Status as UpdateStatus,
} from "../../bindings/github.com/yarkovaleksei/go-desktop-kit/updater";
// Модуль событий заодно регистрирует в рантайме создание данных событий.
export { AppNavigate, SettingsChanged, UpdateStatus as UpdateStatusEvent } from "../../bindings/events";
export { Browser, Clipboard, Events } from "@wailsio/runtime";
