// Каркас окна: шапка с навигацией, уведомление о прошлом сбое, страница
// и всплывающее уведомление об ошибке.
import { Alert, AppBar, Box, Container, Snackbar, Tab, Tabs, Toolbar, Typography } from "@mui/material";
import { useState, type ReactNode } from "react";
import { Link, useLocation } from "react-router-dom";
import { errorText, useI18n } from "../i18n";
import { useApp } from "../store/context";

const tabs = [
  { path: "/", key: "nav.home" },
  { path: "/settings", key: "nav.settings" },
] as const;

/** Сообщение о прошлом сеансе, завершившемся не штатно; закрывается до перезапуска. */
function PreviousCrash() {
  const { t, formatDate } = useI18n();
  const crash = useApp((s) => s.previousCrash);
  const [hidden, setHidden] = useState(false);
  if (!crash || hidden) return null;
  return (
    <Alert severity="warning" onClose={() => setHidden(true)} sx={{ mb: 2 }}>
      {crash.startedAt
        ? t("crash.previous", { at: formatDate(crash.startedAt) })
        : t("crash.previousUnknown")}
    </Alert>
  );
}

/** Ошибка из стора: текст по коду из словаря. */
function ErrorNotice() {
  const { t } = useI18n();
  const notice = useApp((s) => s.notice);
  const dismiss = useApp((s) => s.dismissNotice);
  return (
    <Snackbar key={notice?.id} open={notice !== null} autoHideDuration={8000} onClose={dismiss}>
      <Alert severity="error" variant="filled" onClose={dismiss} sx={{ maxWidth: 560 }}>
        {notice ? errorText(t, notice) : null}
      </Alert>
    </Snackbar>
  );
}

/** Каркас окна вокруг страницы children. */
export function Shell({ children }: { children?: ReactNode }) {
  const { t } = useI18n();
  const name = useApp((s) => s.info?.name ?? "");
  const { pathname } = useLocation();
  const current = tabs.some((tab) => tab.path === pathname) ? pathname : false;

  return (
    <Box sx={{ minHeight: "100vh", bgcolor: "background.default" }}>
      <AppBar
        position="sticky"
        color="inherit"
        elevation={0}
        sx={{ borderBottom: 1, borderColor: "divider" }}
      >
        <Toolbar sx={{ gap: 3 }}>
          <Typography variant="h6" component="div" noWrap>
            {name}
          </Typography>
          <Tabs value={current} component="nav" aria-label={t("nav.label")}>
            {tabs.map((tab) => (
              <Tab key={tab.path} value={tab.path} label={t(tab.key)} component={Link} to={tab.path} />
            ))}
          </Tabs>
        </Toolbar>
      </AppBar>
      <Container maxWidth="md" component="main" sx={{ py: 3 }}>
        <PreviousCrash />
        {children}
      </Container>
      <ErrorNotice />
    </Box>
  );
}
