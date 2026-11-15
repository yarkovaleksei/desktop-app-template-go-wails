// Цветовые токены светлой и тёмной темы. Компоненты берут цвета из темы MUI,
// а не отсюда; контраст текста к фону проверяет theme.test.ts (не ниже 4.5:1).

export type ThemeMode = "light" | "dark";

export interface AppTokens {
  /** Фон окна. */
  bg: string;
  /** Фон карточек, меню и диалогов. */
  paper: string;
  /** Рамки и разделители. */
  line: string;
  /** Основной текст. */
  text: string;
  /** Второстепенный текст: подсказки, подписи. */
  muted: string;
  primary: string;
  /** Текст на фоне primary. */
  onPrimary: string;
  success: string;
  warning: string;
  error: string;
  info: string;
}

/** Roboto лежит в сборке (@fontsource/roboto); системные — запасные. */
export const fontFamily = 'Roboto, system-ui, -apple-system, "Segoe UI", Arial, sans-serif';
export const monoFontFamily = 'ui-monospace, SFMono-Regular, "Cascadia Mono", Consolas, monospace';

export const tokens: Readonly<Record<ThemeMode, AppTokens>> = {
  light: {
    bg: "#f5f6f8",
    paper: "#ffffff",
    line: "#dde1e6",
    text: "#1b2430",
    muted: "#55606e",
    primary: "#2f5fa8",
    onPrimary: "#ffffff",
    success: "#1e7f5c",
    warning: "#9a6212",
    error: "#b3364a",
    info: "#2671a8",
  },
  dark: {
    bg: "#0f141c",
    paper: "#171e29",
    line: "#2a3442",
    text: "#e6eaf0",
    muted: "#a3adbb",
    primary: "#8fb0ff",
    onPrimary: "#0f141c",
    success: "#4cc79a",
    warning: "#f0b866",
    error: "#f78a99",
    info: "#7cc2ee",
  },
};
