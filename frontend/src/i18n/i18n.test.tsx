import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import codes from "./error-codes.json";
import {
  I18nProvider,
  createTranslator,
  dictionaries,
  errorText,
  formatDate,
  hasKey,
  plural,
  resolveLanguage,
  useI18n,
  type Lang,
} from "./index";

const langs: Lang[] = ["ru", "en"];

describe("createTranslator", () => {
  it("подставляет параметры, числа — по правилам языка", () => {
    const t = createTranslator("ru");
    expect(t("settings.version", { version: "v1.2.0" })).toBe("Версия v1.2.0");
    // Разделитель разрядов берётся из Intl: символ зависит от версии ICU.
    const grouped = new Intl.NumberFormat("ru").format(1502);
    expect(grouped).not.toBe("1502");
    expect(t("errors.update.http", { status: 1502 })).toBe(`GitHub ответил ошибкой ${grouped}`);
    expect(createTranslator("en")("errors.update.http", { status: 1502 })).toBe(
      "GitHub responded with error 1,502",
    );
  });

  it("отсутствующий параметр — пустая строка", () => {
    expect(createTranslator("en")("settings.version")).toBe("Version ");
  });

  it("русские формы множественного числа", () => {
    const t = createTranslator("ru");
    expect(t("home.uptime", { count: 1 })).toBe("Окно открыто 1 минуту");
    expect(t("home.uptime", { count: 3 })).toBe("Окно открыто 3 минуты");
    expect(t("home.uptime", { count: 5 })).toBe("Окно открыто 5 минут");
    expect(t("home.uptime", { count: 21 })).toBe("Окно открыто 21 минуту");
    expect(t("home.uptime", { count: 1.5 })).toBe("Окно открыто 1,5 минуты");
  });

  it("английские формы множественного числа", () => {
    const t = createTranslator("en");
    expect(t("home.uptime", { count: 1 })).toBe("The window has been open for 1 minute");
    expect(t("home.uptime", { count: 2 })).toBe("The window has been open for 2 minutes");
    expect(t("home.uptime", { count: 0 })).toBe("The window has been open for 0 minutes");
  });

  it("ключ не из словаря не компилируется", () => {
    const t = createTranslator("ru");
    // @ts-expect-error ключа нет в ru.ts
    expect(() => t("no.such.key")).toThrow();
  });
});

describe("plural и formatDate", () => {
  it("plural выбирает форму по языку", () => {
    const forms = { one: "файл", few: "файла", many: "файлов", other: "файла" };
    expect(plural("ru", 2, forms)).toBe("файла");
    expect(plural("ru", 11, forms)).toBe("файлов");
    expect(plural("en", 1, { one: "file", other: "files" })).toBe("file");
  });

  it("formatDate форматирует по языку", () => {
    const d = new Date(2026, 10, 15, 9, 5);
    expect(formatDate("ru", d, { dateStyle: "long" })).toBe("15 ноября 2026 г.");
    expect(formatDate("en", d, { dateStyle: "long" })).toBe("November 15, 2026");
  });
});

describe("словари", () => {
  it("наборы ключей совпадают", () => {
    expect(Object.keys(dictionaries.en).sort()).toEqual(Object.keys(dictionaries.ru).sort());
  });

  it("у множественного числа есть все формы языка, у строки — строка", () => {
    for (const lang of langs) {
      const categories = new Intl.PluralRules(lang).resolvedOptions().pluralCategories;
      for (const [key, msg] of Object.entries(dictionaries[lang])) {
        const ruMsg = dictionaries.ru[key as keyof typeof dictionaries.ru];
        expect(typeof msg, `${lang} ${key}`).toBe(typeof ruMsg);
        if (typeof msg === "string") continue;
        for (const c of categories) expect(msg[c], `${lang} ${key} ${c}`).toBeTypeOf("string");
      }
    }
  });

  it("у каждого кода ошибки из Go есть текст", () => {
    expect(codes.length).toBeGreaterThan(0);
    for (const code of codes) expect(hasKey(`errors.${code}`), code).toBe(true);
    expect(hasKey("errors.unknown")).toBe(true);
  });

  it("errorText: текст по коду, неизвестный код — errors.unknown", () => {
    const t = createTranslator("ru");
    expect(errorText(t, { code: "settings.invalid-theme", params: { value: "pink" } })).toBe(
      "Неизвестная тема «pink»",
    );
    expect(errorText(t, { code: "nope.nope" })).toBe("Неизвестная ошибка: nope.nope");
  });
});

describe("resolveLanguage", () => {
  it("язык системы, если настройка пустая", () => {
    expect(resolveLanguage("", "ru-RU")).toBe("ru");
    expect(resolveLanguage("", "en-US")).toBe("en");
    expect(resolveLanguage("", "de-DE")).toBe("en");
    expect(resolveLanguage("", undefined)).toBe("en");
  });

  it("явная настройка важнее системы", () => {
    expect(resolveLanguage("en", "ru-RU")).toBe("en");
    expect(resolveLanguage("ru", "de-DE")).toBe("ru");
  });

  it("по умолчанию берёт navigator.language", () => {
    expect(resolveLanguage("")).toBe(navigator.language.startsWith("ru") ? "ru" : "en");
  });
});

describe("I18nProvider", () => {
  function Title() {
    const { t } = useI18n();
    return <h1>{t("settings.title")}</h1>;
  }

  it("смена языка перерисовывает текст и меняет lang документа", () => {
    const { rerender } = render(
      <I18nProvider lang="ru">
        <Title />
      </I18nProvider>,
    );
    expect(screen.getByRole("heading")).toHaveTextContent("Настройки");
    expect(document.documentElement.lang).toBe("ru");

    rerender(
      <I18nProvider lang="en">
        <Title />
      </I18nProvider>,
    );
    expect(screen.getByRole("heading")).toHaveTextContent("Settings");
    expect(document.documentElement.lang).toBe("en");
  });
});
