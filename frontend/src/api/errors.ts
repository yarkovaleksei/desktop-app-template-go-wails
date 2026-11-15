// Go отдаёт ошибку кодом и параметрами без текста (пакет apperr из go-desktop-kit),
// Wails кладёт её JSON в поле cause ошибки вызова. Текст для человека собирает
// интерфейс по словарю: t(`errors.${code}`, params).

/** Код ошибки, которая пришла не из Go или без кода. */
export const UNKNOWN_CODE = "unknown";

/** Ошибка вызова Go с кодом для словаря и параметрами для подстановки. */
export class AppError extends Error {
  readonly code: string;
  readonly params: Record<string, string>;

  constructor(code: string, params: Record<string, string> = {}, options?: { cause?: unknown }) {
    super(code, options);
    this.name = "AppError";
    this.code = code;
    this.params = params;
  }
}

interface WireError {
  code: string;
  params?: Record<string, string | undefined> | null;
}

const isWireError = (v: unknown): v is WireError =>
  typeof v === "object" && v !== null && typeof (v as { code?: unknown }).code === "string";

/** Параметры без пустых значений: так их удобнее подставлять в словарь. */
const cleanParams = (params: WireError["params"]): Record<string, string> => {
  const out: Record<string, string> = {};
  for (const [k, v] of Object.entries(params ?? {})) {
    if (typeof v === "string") out[k] = v;
  }
  return out;
};

/**
 * Приводит ошибку к AppError. Ошибка вызова Go с cause.code сохраняет код
 * и параметры; любая другая получает код unknown и свой текст в параметре detail.
 */
export const toAppError = (err: unknown): AppError => {
  if (err instanceof AppError) return err;
  const cause = (err as { cause?: unknown } | null)?.cause;
  if (isWireError(cause)) return new AppError(cause.code, cleanParams(cause.params), { cause: err });
  const detail = err instanceof Error ? err.message : String(err);
  return new AppError(UNKNOWN_CODE, { detail }, { cause: err });
};

/** Ошибка из статуса обновления (updater.Failure) в виде AppError. */
export const fromFailure = (f: WireError): AppError => new AppError(f.code, cleanParams(f.params));

/** Оборачивает вызов: любая ошибка выходит наружу как AppError. */
export const call = async <T>(fn: () => Promise<T>): Promise<T> => {
  try {
    return await fn();
  } catch (err) {
    throw toAppError(err);
  }
};
