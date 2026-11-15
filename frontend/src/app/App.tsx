// Корень интерфейса: словарь и тема из настроек, граница ошибок, маршруты
// и переход по событию app:navigate из Go (пункт трея, щелчок по уведомлению).
import { CssBaseline, ThemeProvider } from "@mui/material";
import { useEffect, useRef } from "react";
import { HashRouter, Navigate, Route, Routes, useNavigate } from "react-router-dom";
import type { DataSource } from "../api";
import { I18nProvider } from "../i18n";
import { HomePage } from "../pages/HomePage";
import { SettingsPage } from "../pages/SettingsPage";
import type { AppStore } from "../store/appStore";
import { AppStoreProvider, useApp } from "../store/context";
import { useAppTheme, useLang } from "../store/selectors";
import { ErrorBoundary } from "./ErrorBoundary";
import { Shell } from "./Shell";

/** Открывает страницу, которую просит Go. */
function NavigateOnEvent({ source }: { source: DataSource }) {
  const navigate = useNavigate();
  useEffect(() => source.onNavigate((path) => navigate(path)), [source, navigate]);
  return null;
}

/** Сообщает Go, что первый экран с загруженными данными отрисован (для smoke — успех). */
function ReportReady({ source }: { source: DataSource }) {
  const ready = useApp((s) => s.ready);
  const showError = useApp((s) => s.showError);
  const sent = useRef(false);
  useEffect(() => {
    if (!ready || sent.current) return;
    sent.current = true;
    source.frontendReady().catch(showError);
  }, [ready, source, showError]);
  return null;
}

function Themed({ source }: { source: DataSource }) {
  const lang = useLang();
  const theme = useAppTheme();
  const reportError = useApp((s) => s.reportError);
  return (
    <I18nProvider lang={lang}>
      <ThemeProvider theme={theme}>
        <CssBaseline />
        <ErrorBoundary onError={(r) => void reportError(r)}>
          <HashRouter>
            <NavigateOnEvent source={source} />
            <Shell>
              <Routes>
                <Route path="/" element={<HomePage />} />
                <Route path="/settings" element={<SettingsPage />} />
                <Route path="*" element={<Navigate to="/" replace />} />
              </Routes>
            </Shell>
          </HashRouter>
          <ReportReady source={source} />
        </ErrorBoundary>
      </ThemeProvider>
    </I18nProvider>
  );
}

/** Интерфейс приложения поверх стора store и источника source. */
export function App({ source, store }: { source: DataSource; store: AppStore }) {
  return (
    <AppStoreProvider store={store}>
      <Themed source={source} />
    </AppStoreProvider>
  );
}
