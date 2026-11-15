import { Alert, Box, Button, LinearProgress, Stack, Typography } from "@mui/material";
import { useState } from "react";
import { UpdateInstall, UpdateState } from "../api";
import { RELEASES_URL } from "../app/links";
import { errorText, useI18n, type MessageKey } from "../i18n";
import { formatBytes, progressPercent } from "../logic/format";
import { useApp, useAppStore } from "../store/context";

// Пустой способ установки (нулевое значение Go) текста не имеет.
const readyKey: Partial<Record<UpdateInstall, MessageKey>> = {
  [UpdateInstall.InstallCommand]: "updates.ready.command",
  [UpdateInstall.InstallOpen]: "updates.ready.open",
  [UpdateInstall.InstallRun]: "updates.ready.run",
  [UpdateInstall.InstallReplace]: "updates.ready.replace",
  [UpdateInstall.InstallReveal]: "updates.ready.reveal",
};

const installKey: Partial<Record<UpdateInstall, MessageKey>> = {
  [UpdateInstall.InstallCommand]: "updates.install.command",
  [UpdateInstall.InstallOpen]: "updates.install.open",
  [UpdateInstall.InstallRun]: "updates.install.run",
  [UpdateInstall.InstallReplace]: "updates.install.replace",
  [UpdateInstall.InstallReveal]: "updates.install.reveal",
};

/**
 * Блок обновлений: проверка выпуска, скачивание с прогрессом, отмена и установка.
 * В сеть ходит только по кнопке или по автопроверке в Go; состояние живёт в Go
 * и приходит событием update:status, поэтому уход со страницы его не сбрасывает.
 */
export function AppUpdates() {
  const { t, lang } = useI18n();
  const st = useApp((s) => s.update);
  const actions = useAppStore().getState();
  const [installing, setInstalling] = useState(false);
  const [copied, setCopied] = useState(false);

  if (!st) return null;
  const size = (bytes: number | undefined) => formatBytes(lang, bytes ?? 0);
  const busy = st.state === UpdateState.StateChecking || st.state === UpdateState.StateDownloading;
  const install = async () => {
    setInstalling(true);
    await actions.installUpdate();
    setInstalling(false);
  };
  const copy = async (text: string) => setCopied(await actions.copyText(text));
  const ready = st.state === UpdateState.StateReady && st.install ? readyKey[st.install] : undefined;
  const installLabel = st.install ? installKey[st.install] : undefined;

  return (
    <Stack spacing={1.5}>
      <Typography variant="body2" color="text.secondary">
        {t("settings.version", { version: st.current })}
      </Typography>

      {st.state === UpdateState.StateUnsupported && (
        <Alert
          severity="info"
          action={
            <Button color="inherit" onClick={() => void actions.openURL(RELEASES_URL)}>
              {t("updates.notes")}
            </Button>
          }
        >
          {t("updates.unsupported")}
        </Alert>
      )}
      {st.state === UpdateState.StateCurrent && (
        <Alert severity="success">{t("updates.upToDate", { version: st.latest ?? st.current })}</Alert>
      )}
      {st.state === UpdateState.StateFailed && st.error && (
        <Alert severity="error">
          {errorText(t, { code: st.error.code, params: cleanParams(st.error.params) })}
        </Alert>
      )}
      {st.state === UpdateState.StateAvailable && (
        <Alert severity="info">
          {t("updates.available", { version: st.latest ?? "", size: size(st.size) })}
        </Alert>
      )}
      {st.state === UpdateState.StateDownloading && (
        <Stack spacing={0.5}>
          <Typography variant="body2">
            {t("updates.downloading", { received: size(st.received), size: size(st.size) })}
          </Typography>
          <LinearProgress
            variant="determinate"
            value={progressPercent(st.received, st.size)}
            aria-label={t("updates.download")}
          />
        </Stack>
      )}
      {ready && (
        <Alert severity="success">
          <Stack spacing={1}>
            <span>{t(ready, { version: st.latest ?? "" })}</span>
            {st.command && (
              <Stack direction="row" spacing={1} sx={{ alignItems: "center", flexWrap: "wrap" }}>
                <Box
                  component="code"
                  sx={{
                    fontFamily: (th) => th.app.monoFontFamily,
                    wordBreak: "break-all",
                    userSelect: "all",
                  }}
                >
                  {st.command}
                </Box>
                <Button size="small" color="inherit" onClick={() => void copy(st.command ?? "")}>
                  {t("updates.copy")}
                </Button>
                {copied && <Typography variant="caption">{t("updates.copied")}</Typography>}
              </Stack>
            )}
          </Stack>
        </Alert>
      )}

      <Stack direction="row" spacing={1} sx={{ flexWrap: "wrap" }}>
        {st.state !== UpdateState.StateUnsupported && st.state !== UpdateState.StateReady && (
          <Button variant="outlined" disabled={busy} onClick={() => void actions.checkUpdate()}>
            {t(st.state === UpdateState.StateChecking ? "updates.checking" : "updates.check")}
          </Button>
        )}
        {(st.state === UpdateState.StateAvailable ||
          (st.state === UpdateState.StateFailed && st.package)) && (
          <Button variant="contained" onClick={() => void actions.downloadUpdate()}>
            {t(st.state === UpdateState.StateFailed ? "updates.retry" : "updates.download")}
          </Button>
        )}
        {st.state === UpdateState.StateDownloading && (
          <Button onClick={() => void actions.cancelUpdate()}>{t("updates.cancel")}</Button>
        )}
        {ready && installLabel && (
          <Button variant="contained" disabled={installing} onClick={() => void install()}>
            {t(installLabel)}
          </Button>
        )}
        {st.notesUrl && (
          <Button onClick={() => void actions.openURL(st.notesUrl ?? RELEASES_URL)}>
            {t("updates.notes")}
          </Button>
        )}
      </Stack>
    </Stack>
  );
}

/** Параметры ошибки статуса без пустых значений. */
function cleanParams(params: Record<string, string | undefined> | null | undefined): Record<string, string> {
  const out: Record<string, string> = {};
  for (const [k, v] of Object.entries(params ?? {})) if (typeof v === "string") out[k] = v;
  return out;
}
