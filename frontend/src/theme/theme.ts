// Тема MUI из токенов: светлая или тёмная, с локалью компонентов MUI по языку
// интерфейса. Шрифт Roboto подключается из пакета — без обращения к сети.
import "@fontsource/roboto/latin-400.css";
import "@fontsource/roboto/latin-500.css";
import "@fontsource/roboto/latin-700.css";
import "@fontsource/roboto/latin-ext-400.css";
import "@fontsource/roboto/latin-ext-500.css";
import "@fontsource/roboto/latin-ext-700.css";
import "@fontsource/roboto/cyrillic-400.css";
import "@fontsource/roboto/cyrillic-500.css";
import "@fontsource/roboto/cyrillic-700.css";
import "@fontsource/roboto/cyrillic-ext-400.css";
import "@fontsource/roboto/cyrillic-ext-500.css";
import "@fontsource/roboto/cyrillic-ext-700.css";

import { enUS, ruRU } from "@mui/material/locale";
import { createTheme, type Theme } from "@mui/material/styles";
import type { Lang } from "../i18n";
import { fontFamily, monoFontFamily, tokens, type AppTokens, type ThemeMode } from "./tokens";

export type { AppTokens, ThemeMode } from "./tokens";

declare module "@mui/material/styles" {
  interface Theme {
    /** Токены текущей темы и моноширинный шрифт: для цветов вне палитры MUI. */
    app: AppTokens & { monoFontFamily: string };
  }
  interface ThemeOptions {
    app?: AppTokens & { monoFontFamily: string };
  }
}

const muiLocales = { ru: ruRU, en: enUS } as const;

/**
 * Режим темы по настройке: "light" и "dark" — как заданы, пустая строка
 * (как в системе) — по prefers-color-scheme.
 */
export function resolveMode(setting: string, prefersDark: boolean): ThemeMode {
  if (setting === "light" || setting === "dark") return setting;
  return prefersDark ? "dark" : "light";
}

/** Тема MUI для режима mode и языка интерфейса language. */
export function createAppTheme(mode: ThemeMode, language: Lang): Theme {
  const t = tokens[mode];
  return createTheme(
    {
      app: { ...t, monoFontFamily },
      palette: {
        mode,
        primary: { main: t.primary, contrastText: t.onPrimary },
        success: { main: t.success },
        warning: { main: t.warning },
        error: { main: t.error },
        info: { main: t.info },
        background: { default: t.bg, paper: t.paper },
        text: { primary: t.text, secondary: t.muted },
        divider: t.line,
      },
      shape: { borderRadius: 8 },
      typography: {
        fontFamily,
        button: { textTransform: "none", fontWeight: 500 },
      },
      components: {
        MuiPaper: {
          // Карточки с рамкой вместо тени: размытые тени WebKitGTK перерисовывает
          // медленно. У меню и диалогов тень своя, по elevation.
          defaultProps: { elevation: 0 },
          styleOverrides: { root: { backgroundImage: "none" }, outlined: { borderColor: t.line } },
        },
        MuiButton: { defaultProps: { disableElevation: true } },
      },
    },
    muiLocales[language],
  );
}
