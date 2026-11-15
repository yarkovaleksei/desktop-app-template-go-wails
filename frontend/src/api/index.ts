// Выбор источника данных: привязки Wails в окне приложения, моки в обычном браузере
// (`yarn dev`, сценарии Playwright). Код источника, который не нужен, не загружается.
import type { DataSource } from "./source";

export * from "./source";
export { AppError, UNKNOWN_CODE, call, fromFailure, toAppError } from "./errors";

// Признаки нативного окна Wails v3: WebView2 в Windows, WebKit в Linux и macOS.
// window._wails не годится: его создаёт сам рантайм и в обычном браузере.
function hasNativeHost(): boolean {
  const w = window as unknown as {
    chrome?: { webview?: unknown };
    webkit?: { messageHandlers?: { external?: unknown } };
  };
  return Boolean(w.chrome?.webview || w.webkit?.messageHandlers?.external);
}

/** Источник данных для текущего окружения. */
export async function pickSource(): Promise<DataSource> {
  if (hasNativeHost()) {
    const { wailsSource } = await import("./wailsSource");
    return wailsSource;
  }
  const { createMockSource } = await import("./mockSource");
  return createMockSource();
}
