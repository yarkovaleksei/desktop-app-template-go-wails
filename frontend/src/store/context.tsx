import { createContext, useContext, type ReactNode } from "react";
import { useStore } from "zustand";
import type { AppState, AppStore } from "./appStore";

const StoreContext = createContext<AppStore | null>(null);

/** Передаёт стор дереву компонентов; тесты подставляют стор с fakeSource. */
export function AppStoreProvider({ store, children }: { store: AppStore; children?: ReactNode }) {
  return <StoreContext.Provider value={store}>{children}</StoreContext.Provider>;
}

/** Стор целиком: для действий вне рендера. */
export function useAppStore(): AppStore {
  const store = useContext(StoreContext);
  if (!store) throw new Error("AppStoreProvider не найден");
  return store;
}

/** Подписка на часть стора. Селектор должен возвращать стабильное значение. */
export function useApp<T>(selector: (s: AppState) => T): T {
  return useStore(useAppStore(), selector);
}
