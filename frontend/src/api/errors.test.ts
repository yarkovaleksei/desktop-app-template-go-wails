import { describe, expect, it } from "vitest";
import { AppError, UNKNOWN_CODE, call, fromFailure, toAppError } from "./errors";

// Так выглядит ошибка вызова Wails: JSON ошибки Go (apperr) лежит в cause.
const wailsError = (cause: unknown) => Object.assign(new Error("call failed"), { cause });

describe("toAppError", () => {
  it("ошибка Go с cause.code сохраняет код и параметры", () => {
    const err = toAppError(wailsError({ code: "settings.invalid-theme", params: { value: "pink" } }));
    expect(err).toBeInstanceOf(AppError);
    expect(err.code).toBe("settings.invalid-theme");
    expect(err.params).toEqual({ value: "pink" });
  });

  it("params: null становится пустым объектом", () => {
    expect(toAppError(wailsError({ code: "settings.broken", params: null })).params).toEqual({});
  });

  it("ошибка без кода получает код unknown и текст в detail", () => {
    const err = toAppError(new Error("сломалось"));
    expect(err.code).toBe(UNKNOWN_CODE);
    expect(err.params).toEqual({ detail: "сломалось" });
  });

  it("cause без строкового кода — тоже unknown", () => {
    expect(toAppError(wailsError({ code: 42 })).code).toBe(UNKNOWN_CODE);
    expect(toAppError(wailsError("текст")).code).toBe(UNKNOWN_CODE);
  });

  it("не Error — unknown с текстом значения", () => {
    expect(toAppError("строка")).toMatchObject({ code: UNKNOWN_CODE, params: { detail: "строка" } });
    expect(toAppError(null)).toMatchObject({ code: UNKNOWN_CODE, params: { detail: "null" } });
  });

  it("AppError проходит как есть", () => {
    const err = new AppError("update.network");
    expect(toAppError(err)).toBe(err);
  });
});

describe("fromFailure", () => {
  it("отбрасывает пустые параметры", () => {
    expect(fromFailure({ code: "update.size", params: { reason: "too-large", x: undefined } })).toMatchObject(
      {
        code: "update.size",
        params: { reason: "too-large" },
      },
    );
  });
});

describe("call", () => {
  it("возвращает результат", async () => {
    await expect(call(async () => 7)).resolves.toBe(7);
  });

  it("отклоняется с AppError", async () => {
    await expect(
      call(() => Promise.reject(wailsError({ code: "autostart.unavailable" }))),
    ).rejects.toMatchObject({ name: "AppError", code: "autostart.unavailable" });
  });
});
