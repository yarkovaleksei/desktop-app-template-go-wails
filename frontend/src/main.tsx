// Точка входа: источник данных по окружению, стор, отчёты об ошибках окна, рендер.
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { pickSource } from "./api";
import { App } from "./app/App";
import { installErrorReporting } from "./app/ErrorBoundary";
import { createAppStore } from "./store/appStore";

async function start() {
  const source = await pickSource();
  const store = createAppStore(source);
  installErrorReporting((r) => void store.getState().reportError(r));
  void store.getState().init();

  const root = document.getElementById("root");
  if (!root) throw new Error("нет элемента #root");
  createRoot(root).render(
    <StrictMode>
      <App source={source} store={store} />
    </StrictMode>,
  );
}

void start();
