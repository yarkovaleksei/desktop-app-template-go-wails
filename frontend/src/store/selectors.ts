// Производные значения стора для компонентов: язык и тема с учётом «как в системе».
import { useMediaQuery } from "@mui/material";
import type { Theme } from "@mui/material/styles";
import { useMemo } from "react";
import { resolveLanguage, type Lang } from "../i18n";
import { createAppTheme, resolveMode, type ThemeMode } from "../theme/theme";
import { useApp } from "./context";

/** Язык интерфейса: из настройки, иначе язык системы (D-033). */
export function useLang(): Lang {
  const setting = useApp((s) => s.settings.ui.language);
  return resolveLanguage(setting);
}

/** Режим темы: из настройки, иначе по prefers-color-scheme. */
export function useThemeMode(): ThemeMode {
  const setting = useApp((s) => s.settings.ui.theme);
  const prefersDark = useMediaQuery("(prefers-color-scheme: dark)", { noSsr: true });
  return resolveMode(setting, prefersDark);
}

/** Тема MUI для текущих режима и языка. */
export function useAppTheme(): Theme {
  const mode = useThemeMode();
  const lang = useLang();
  return useMemo(() => createAppTheme(mode, lang), [mode, lang]);
}

/** Работа в фоне возможна только при трее: без него окно было бы не вернуть. */
export function useTrayAvailable(): boolean {
  return useApp((s) => s.info?.trayAvailable ?? false);
}
