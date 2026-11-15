// Форматирование значений для интерфейса.
import type { Lang } from "../i18n";

/** Размер в мегабайтах по правилам языка: «48,2 МБ», «48.2 MB». */
export function formatBytes(lang: Lang, bytes: number): string {
  return new Intl.NumberFormat(lang, { style: "unit", unit: "megabyte", maximumFractionDigits: 1 }).format(
    bytes / 1e6,
  );
}

/** Доля скачанного в процентах, 0–100; без размера — 0. */
export function progressPercent(received: number | undefined, size: number | undefined): number {
  if (!size || size <= 0) return 0;
  return Math.min(100, Math.max(0, ((received ?? 0) / size) * 100));
}

/** Полных минут с момента since до now. */
export function minutesSince(since: number, now: number): number {
  return Math.max(0, Math.floor((now - since) / 60_000));
}
