// DataSource поверх привязок Wails: используется в окне приложения.
import { AppService, Browser, Clipboard, SettingsService, UpdatesService } from "./bindings";
import { call } from "./errors";
import { onNavigate, onSettingsChanged, onUpdateStatus } from "./events";
import type { DataSource } from "./source";

export const wailsSource: DataSource = {
  kind: "wails",

  info: () => call(() => AppService.Info()),
  frontendReady: () => call(() => AppService.FrontendReady()),
  previousCrash: () => call(() => AppService.PreviousCrash()),
  reportError: (r) => call(() => AppService.ReportFrontendError(r)),
  openLogFolder: () => call(() => AppService.OpenLogFolder()),

  getSettings: () => call(() => SettingsService.Get()),
  saveSettings: (s) => call(() => SettingsService.Save(s)),
  getAutostart: () => call(() => SettingsService.Autostart()),
  setAutostart: (on) => call(() => SettingsService.SetAutostart(on)),

  updateStatus: () => call(() => UpdatesService.Status()),
  checkUpdate: () => call(() => UpdatesService.Check()),
  downloadUpdate: () => call(() => UpdatesService.Download()),
  cancelUpdate: () => call(() => UpdatesService.Cancel()),
  installUpdate: () => call(() => UpdatesService.Install()),

  openURL: (url) => call(() => Browser.OpenURL(url)),
  copyText: (text) => call(() => Clipboard.SetText(text)),

  onUpdateStatus,
  onSettingsChanged,
  onNavigate,
};
