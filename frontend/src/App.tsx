import { useEffect } from "react";
import { pickSource } from "./api";

// Временная страница каркаса: полный интерфейс появится на шаге 09.6.
function App() {
  // Первый экран отрисован: сообщить Go-стороне. В обычной работе вызов ничего
  // не делает, в smoke-режиме завершает проверку запуска успехом.
  useEffect(() => {
    pickSource()
      .then((source) => source.frontendReady())
      .catch((err: unknown) => {
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
