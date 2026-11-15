import {
  Alert,
  Button,
  FormControlLabel,
  FormHelperText,
  Paper,
  Stack,
  Switch,
  ToggleButton,
  ToggleButtonGroup,
  Typography,
} from "@mui/material";
import type { ReactNode } from "react";
import { Language, Theme } from "../../api";
import { AppUpdates } from "../../components/AppUpdates";
import { useI18n, type MessageKey } from "../../i18n";
import { useApp, useAppStore } from "../../store/context";
import { useTrayAvailable } from "../../store/selectors";

const languages: { value: string; key: MessageKey }[] = [
  { value: Language.System, key: "settings.language.system" },
  { value: Language.RU, key: "settings.language.ru" },
  { value: Language.EN, key: "settings.language.en" },
];

const themes: { value: string; key: MessageKey }[] = [
  { value: Theme.System, key: "settings.theme.system" },
  { value: Theme.Light, key: "settings.theme.light" },
  { value: Theme.Dark, key: "settings.theme.dark" },
];

function Section({ title, children }: { title: string; children?: ReactNode }) {
  return (
    <Paper variant="outlined" component="section" sx={{ p: 2.5 }}>
      <Typography variant="h6" component="h2" sx={{ mb: 1.5 }}>
        {title}
      </Typography>
      <Stack spacing={1.5}>{children}</Stack>
    </Paper>
  );
}

/** Выбор одного значения из нескольких кнопками; подпись — для чтения с экрана. */
function Choice(props: {
  label: string;
  value: string;
  options: { value: string; key: MessageKey }[];
  onChange: (v: string) => void;
}) {
  const { t } = useI18n();
  return (
    <Stack spacing={0.5}>
      <Typography variant="body2" color="text.secondary" component="div">
        {props.label}
      </Typography>
      <ToggleButtonGroup
        exclusive
        size="small"
        value={props.value}
        aria-label={props.label}
        onChange={(_, v: string | null) => v !== null && props.onChange(v)}
      >
        {props.options.map((o) => (
          <ToggleButton key={o.value} value={o.value}>
            {t(o.key)}
          </ToggleButton>
        ))}
      </ToggleButtonGroup>
    </Stack>
  );
}

/**
 * Страница настроек — язык, тема, работа в фоне, автозапуск, автопроверка
 * обновлений, сведения о сборке, блок обновлений и папка журнала.
 */
export function SettingsPage() {
  const { t } = useI18n();
  const settings = useApp((s) => s.settings);
  const broken = useApp((s) => s.settingsBroken);
  const autostart = useApp((s) => s.autostart);
  const info = useApp((s) => s.info);
  const tray = useTrayAvailable();
  // Действия берутся из стора без подписки: страница перерисовывается только от данных.
  const actions = useAppStore().getState();

  return (
    <Stack spacing={2}>
      <Typography variant="h4" component="h1">
        {t("settings.title")}
      </Typography>
      {broken && <Alert severity="warning">{t("settings.broken")}</Alert>}

      <Section title={t("settings.appearance")}>
        <Choice
          label={t("settings.language")}
          value={settings.ui.language}
          options={languages}
          onChange={(v) => void actions.setLanguage(v)}
        />
        <Choice
          label={t("settings.theme")}
          value={settings.ui.theme}
          options={themes}
          onChange={(v) => void actions.setTheme(v)}
        />
      </Section>

      <Section title={t("settings.behavior")}>
        <div>
          <FormControlLabel
            disabled={!tray}
            control={
              <Switch
                checked={tray && settings.background}
                onChange={(e) => void actions.setBackground(e.target.checked)}
                slotProps={{ input: { "aria-describedby": tray ? undefined : "background-help" } }}
              />
            }
            label={t("settings.background")}
          />
          {!tray && <FormHelperText id="background-help">{t("settings.background.noTray")}</FormHelperText>}
        </div>
        <FormControlLabel
          disabled={autostart === null}
          control={
            <Switch
              checked={autostart ?? false}
              onChange={(e) => void actions.setAutostart(e.target.checked)}
            />
          }
          label={t("settings.autostart")}
        />
        <div>
          <FormControlLabel
            control={
              <Switch
                checked={settings.autoCheckUpdates}
                onChange={(e) => void actions.setAutoCheckUpdates(e.target.checked)}
                slotProps={{ input: { "aria-describedby": "auto-check-help" } }}
              />
            }
            label={t("settings.autoCheck")}
          />
          <FormHelperText id="auto-check-help">{t("settings.autoCheck.help")}</FormHelperText>
        </div>
      </Section>

      <Section title={t("updates.title")}>
        <AppUpdates />
      </Section>

      <Section title={t("settings.about")}>
        {info && (
          <Typography variant="body2">
            {t("settings.version", { version: info.version })}
            <br />
            {t("settings.format", { format: info.format || t("settings.format.none") })}
          </Typography>
        )}
        <div>
          <Button variant="outlined" onClick={() => void actions.openLogFolder()}>
            {t("settings.openLogs")}
          </Button>
        </div>
      </Section>
    </Stack>
  );
}
