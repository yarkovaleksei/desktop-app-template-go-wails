// Словари интерфейса: перевод по ключу, множественное число через Intl.PluralRules,
// числа и даты через Intl, выбор языка (D-033). Строки интерфейса живут только
// в ru.ts и en.ts; компоненты получают t через useI18n и перерисовываются при смене языка.
import { createContext, createElement, useContext, useEffect, useMemo, type ReactNode } from "react";
import { en } from "./en";
import { ru, type Dictionary, type MessageKey } from "./ru";

export type { Dictionary, MessageKey } from "./ru";

/** Формы множественного числа; other обязательна, остальные — по правилам языка. */
export type Plural = Partial<Record<Intl.LDMLPluralRule, string>> & { other: string };
/** Значение словаря: строка или формы множественного числа. */
export type Message = string | Plural;
/** Параметры подстановки; count выбирает форму множественного числа. */
export type Params = Readonly<Record<string, string | number>>;

/** Языки интерфейса. */
export type Lang = "ru" | "en";
/** Язык, если ни настройка, ни система не дают поддерживаемого (D-033). */
export const FALLBACK_LANG: Lang = "en";

export const dictionaries: Readonly<Record<Lang, Dictionary>> = { ru, en };

const isLang = (v: string): v is Lang => v === "ru" || v === "en";

/**
 * Язык интерфейса: явная настройка, иначе язык системы (navigator.language),
 * иначе английский. Пустая настройка — «как в системе».
 */
export function resolveLanguage(setting: string, system: string | undefined = navigator.language): Lang {
  if (isLang(setting)) return setting;
  const base = (system ?? "").toLowerCase().split(/[-_]/)[0] ?? "";
  return isLang(base) ? base : FALLBACK_LANG;
}

/** Есть ли такой ключ в словаре: для ключей, собранных из данных (коды ошибок). */
export const hasKey = (key: string): key is MessageKey => Object.hasOwn(ru, key);

/** Форма множественного числа для count по правилам языка. */
export function plural(lang: Lang, count: number, forms: Plural): string {
  return forms[new Intl.PluralRules(lang).select(count)] ?? forms.other;
}

/** Дата и время по правилам языка. */
export function formatDate(
  lang: Lang,
  value: Date | string | number,
  options: Intl.DateTimeFormatOptions = { dateStyle: "medium", timeStyle: "short" },
): string {
  return new Intl.DateTimeFormat(lang, options).format(new Date(value));
}

/** Число по правилам языка. */
export function formatNumber(lang: Lang, value: number, options?: Intl.NumberFormatOptions): string {
  return new Intl.NumberFormat(lang, options).format(value);
}

/** Функция перевода для языка. */
export type Translate = (key: MessageKey, params?: Params) => string;

/**
 * Перевод для языка lang. «{name}» заменяется параметром (числа — по правилам
 * языка), отсутствующий параметр — пустой строкой.
 */
export function createTranslator(lang: Lang): Translate {
  const dict = dictionaries[lang];
  const rules = new Intl.PluralRules(lang);
  const numbers = new Intl.NumberFormat(lang);
  return (key, params = {}) => {
    const msg: Message = dict[key];
    let text: string;
    if (typeof msg === "string") {
      text = msg;
    } else {
      const count = Number(params.count ?? 0);
      text = msg[rules.select(count)] ?? msg.other;
    }
    return text.replace(/\{(\w+)\}/g, (_, name: string) => {
      const v = params[name];
      if (v === undefined) return "";
      return typeof v === "number" ? numbers.format(v) : v;
    });
  };
}

/** Текст ошибки с кодом из Go: errors.<код>, неизвестный код — errors.unknown. */
export function errorText(t: Translate, err: { code: string; params?: Params }): string {
  const key = `errors.${err.code}`;
  return hasKey(key) ? t(key, err.params) : t("errors.unknown", { detail: err.code });
}

/** Что дают компонентам словари. */
export interface I18n {
  lang: Lang;
  t: Translate;
  formatDate: (value: Date | string | number, options?: Intl.DateTimeFormatOptions) => string;
  formatNumber: (value: number, options?: Intl.NumberFormatOptions) => string;
}

const makeI18n = (lang: Lang): I18n => ({
  lang,
  t: createTranslator(lang),
  formatDate: (value, options) => formatDate(lang, value, options),
  formatNumber: (value, options) => formatNumber(lang, value, options),
});

const I18nContext = createContext<I18n>(makeI18n(FALLBACK_LANG));

/**
 * Даёт словарь языка lang всему дереву. Смена lang перерисовывает интерфейс
 * без перезагрузки и меняет атрибут lang документа. Локаль MUI меняет тема.
 */
export function I18nProvider({ lang, children }: { lang: Lang; children?: ReactNode }) {
  const value = useMemo(() => makeI18n(lang), [lang]);
  useEffect(() => {
    document.documentElement.lang = lang;
  }, [lang]);
  return createElement(I18nContext.Provider, { value }, children);
}

/** Словарь текущего языка. */
export const useI18n = (): I18n => useContext(I18nContext);
