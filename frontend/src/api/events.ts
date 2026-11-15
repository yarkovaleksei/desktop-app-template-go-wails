// Типизированные подписки на события Go. Имена и типы данных генерирует wails3
// из регистрации событий в internal/app и wailsx; каждая подписка возвращает отписку.
import {
  AppNavigate,
  Events,
  SettingsChanged,
  UpdateStatusEvent,
  type Settings,
  type UpdateStatus,
} from "./bindings";

/** Отменяет подписку. */
export type Unsubscribe = () => void;

/** Новый снимок статуса обновления. */
export const onUpdateStatus = (cb: (s: UpdateStatus) => void): Unsubscribe =>
  Events.On(UpdateStatusEvent, (e) => cb(e.data));

/** Настройки сохранены, в том числе из меню трея. */
export const onSettingsChanged = (cb: (s: Settings) => void): Unsubscribe =>
  Events.On(SettingsChanged, (e) => cb(e.data));

/** Go просит открыть страницу; данные — путь маршрута, например "/settings". */
export const onNavigate = (cb: (path: string) => void): Unsubscribe =>
  Events.On(AppNavigate, (e) => cb(e.data));
