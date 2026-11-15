import { describe, expect, it } from "vitest";
import { createAppTheme, resolveMode } from "./theme";
import { tokens, type ThemeMode } from "./tokens";

// Относительная яркость и контраст по WCAG 2.1.
const channel = (c: number) => {
  const s = c / 255;
  return s <= 0.04045 ? s / 12.92 : ((s + 0.055) / 1.055) ** 2.4;
};
const luminance = (hex: string) => {
  const m = /^#([0-9a-f]{2})([0-9a-f]{2})([0-9a-f]{2})$/i.exec(hex);
  if (!m) throw new Error(`не #rrggbb: ${hex}`);
  const [r, g, b] = [m[1], m[2], m[3]].map((x) => channel(parseInt(x ?? "", 16)));
  return 0.2126 * (r ?? 0) + 0.7152 * (g ?? 0) + 0.0722 * (b ?? 0);
};
const contrast = (a: string, b: string) => {
  const [hi, lo] = [luminance(a), luminance(b)].sort((x, y) => y - x);
  return ((hi ?? 0) + 0.05) / ((lo ?? 0) + 0.05);
};

const modes: ThemeMode[] = ["light", "dark"];

describe("контраст", () => {
  it("расчёт совпадает с эталоном WCAG", () => {
    expect(contrast("#000000", "#ffffff")).toBeCloseTo(21, 5);
    expect(contrast("#777777", "#ffffff")).toBeCloseTo(4.48, 2);
  });

  for (const mode of modes) {
    const t = tokens[mode];
    it(`${mode}: текст к фону не ниже 4.5:1`, () => {
      for (const [fg, bg] of [
        [t.text, t.bg],
        [t.text, t.paper],
        [t.muted, t.bg],
        [t.muted, t.paper],
        [t.onPrimary, t.primary],
      ] as const) {
        expect(contrast(fg, bg), `${fg} на ${bg}`).toBeGreaterThanOrEqual(4.5);
      }
    });
  }
});

describe("createAppTheme", () => {
  for (const mode of modes) {
    it(`${mode}: палитра MUI взята из токенов`, () => {
      const theme = createAppTheme(mode, "ru");
      const t = tokens[mode];
      expect(theme.palette.mode).toBe(mode);
      expect(theme.palette.text.primary).toBe(t.text);
      expect(theme.palette.text.secondary).toBe(t.muted);
      expect(theme.palette.background.default).toBe(t.bg);
      expect(theme.palette.background.paper).toBe(t.paper);
      expect(theme.palette.primary.contrastText).toBe(t.onPrimary);
      expect(theme.app.text).toBe(t.text);
    });
  }

  it("локаль MUI следует языку", () => {
    const ru = createAppTheme("light", "ru").components?.MuiTablePagination?.defaultProps;
    const en = createAppTheme("light", "en").components?.MuiTablePagination?.defaultProps;
    expect(ru?.labelRowsPerPage).toBe("Строк на странице:");
    // enUS в MUI пуст: английские тексты — умолчания компонентов, русских быть не должно.
    expect(en?.labelRowsPerPage).toBeUndefined();
  });
});

describe("resolveMode", () => {
  it("явная тема важнее системы, пустая — по системе", () => {
    expect(resolveMode("light", true)).toBe("light");
    expect(resolveMode("dark", false)).toBe("dark");
    expect(resolveMode("", true)).toBe("dark");
    expect(resolveMode("", false)).toBe("light");
  });
});
