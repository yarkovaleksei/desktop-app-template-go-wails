import { Button, Stack, Typography } from "@mui/material";
import { useEffect, useState } from "react";
import { DOCS_URL } from "../../app/links";
import { useI18n } from "../../i18n";
import { minutesSince } from "../../logic/format";
import { useApp } from "../../store/context";

/**
 * Главная страница — заглушка с названием приложения и ссылкой на документацию проекта.
 * Новый проект заменяет её своим интерфейсом.
 */
export function HomePage() {
  const { t } = useI18n();
  const name = useApp((s) => s.info?.name ?? "");
  const openURL = useApp((s) => s.openURL);
  const [openedAt] = useState(() => Date.now());
  const [now, setNow] = useState(openedAt);

  useEffect(() => {
    const timer = setInterval(() => setNow(Date.now()), 60_000);
    return () => clearInterval(timer);
  }, []);

  return (
    <Stack spacing={2} sx={{ alignItems: "flex-start" }}>
      <Typography variant="h4" component="h1">
        {name}
      </Typography>
      <Typography>{t("home.intro")}</Typography>
      <Typography variant="body2" color="text.secondary">
        {t("home.uptime", { count: minutesSince(openedAt, now) })}
      </Typography>
      <Button variant="outlined" onClick={() => void openURL(DOCS_URL)}>
        {t("home.docs")}
      </Button>
    </Stack>
  );
}
