// Необработанные ошибки интерфейса уходят в журнал Go: ошибки рендера ловит
// граница ErrorBoundary, прочие — обработчики error и unhandledrejection окна.
import { Alert, Box, Button, Typography } from "@mui/material";
import { Component, type ErrorInfo, type ReactNode } from "react";
import type { FrontendError } from "../api";
import { useI18n } from "../i18n";

type Report = (r: FrontendError) => void;

const describe = (value: unknown): Pick<FrontendError, "message" | "stack"> =>
  value instanceof Error
    ? { message: `${value.name}: ${value.message}`, stack: value.stack ?? "" }
    : { message: String(value), stack: "" };

/** Подключает отчёты об ошибках окна; вызывается один раз в main.tsx. Возвращает отключение. */
export function installErrorReporting(report: Report, target: Window = window): () => void {
  const onError = (e: ErrorEvent) =>
    report({ source: "error", componentStack: "", ...describe(e.error ?? e.message) });
  const onRejection = (e: PromiseRejectionEvent) =>
    report({ source: "unhandledrejection", componentStack: "", ...describe(e.reason) });
  target.addEventListener("error", onError);
  target.addEventListener("unhandledrejection", onRejection);
  return () => {
    target.removeEventListener("error", onError);
    target.removeEventListener("unhandledrejection", onRejection);
  };
}

/** Экран после ошибки рендера: что случилось и как продолжить. */
function CrashScreen() {
  const { t } = useI18n();
  return (
    <Box sx={{ p: 4, maxWidth: 640 }}>
      <Alert
        severity="error"
        action={
          <Button color="inherit" onClick={() => window.location.reload()}>
            {t("crash.reload")}
          </Button>
        }
      >
        <Typography component="h1" variant="subtitle1" sx={{ fontWeight: 500 }}>
          {t("crash.title")}
        </Typography>
        {t("crash.body")}
      </Alert>
    </Box>
  );
}

/** Граница ошибок рендера: показывает CrashScreen и отправляет отчёт в Go. */
export class ErrorBoundary extends Component<{ onError: Report; children?: ReactNode }, { failed: boolean }> {
  state = { failed: false };

  static getDerivedStateFromError() {
    return { failed: true };
  }

  componentDidCatch(error: unknown, info: ErrorInfo) {
    this.props.onError({ source: "render", componentStack: info.componentStack ?? "", ...describe(error) });
  }

  render() {
    return this.state.failed ? <CrashScreen /> : this.props.children;
  }
}
