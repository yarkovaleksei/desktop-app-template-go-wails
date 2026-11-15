import { useEffect } from "react";
import { AppService } from "../bindings/github.com/yarkovaleksei/desktop-app-template-go-wails/internal/app";

// Временная страница каркаса: полный интерфейс появится на этапе фронтенда.
function App() {
  // Первый экран отрисован: сообщить Go-стороне. В обычной работе вызов ничего
  // не делает, в smoke-режиме завершает проверку запуска успехом.
  useEffect(() => {
    AppService.FrontendReady().catch((err: unknown) => {
      console.error("FrontendReady", err);
    });
  }, []);

  return (
    <main style={{ fontFamily: "system-ui, sans-serif", padding: "2rem" }}>
      <h1>Desktop App</h1>
    </main>
  );
}

export default App;
