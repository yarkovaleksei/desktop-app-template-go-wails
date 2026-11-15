import { CssBaseline, ThemeProvider, Typography, useMediaQuery } from "@mui/material";
import { useEffect, useMemo } from "react";
import { pickSource } from "./api";
import { resolveLanguage } from "./i18n";
import { createAppTheme, resolveMode } from "./theme/theme";

// Временная страница каркаса: полный интерфейс появится на шаге 09.6.
function App() {
  const prefersDark = useMediaQuery("(prefers-color-scheme: dark)");
  const theme = useMemo(
    () => createAppTheme(resolveMode("", prefersDark), resolveLanguage("")),
    [prefersDark],
  );

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
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <main style={{ padding: "2rem" }}>
        <Typography variant="h4" component="h1">
          Desktop App
        </Typography>
      </main>
    </ThemeProvider>
  );
}

export default App;
