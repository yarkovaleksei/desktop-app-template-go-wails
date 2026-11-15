// Английский словарь: тот же набор ключей, что в ru.ts (проверяют тип Dictionary и тест).
import type { Dictionary } from "./ru";

export const en: Dictionary = {
  "nav.home": "Home",
  "nav.settings": "Settings",
  "nav.label": "Sections",

  "home.intro": "A placeholder home page. Replace it with your application's interface.",
  "home.docs": "Project documentation",
  "home.uptime": {
    one: "The window has been open for {count} minute",
    other: "The window has been open for {count} minutes",
  },

  "settings.title": "Settings",
  "settings.appearance": "Appearance",
  "settings.language": "Language",
  "settings.language.system": "System default",
  "settings.language.ru": "Русский",
  "settings.language.en": "English",
  "settings.theme": "Theme",
  "settings.theme.system": "System default",
  "settings.theme.light": "Light",
  "settings.theme.dark": "Dark",
  "settings.behavior": "Behavior",
  "settings.background": "Keep running in the background when the window is closed",
  "settings.background.noTray":
    "Unavailable: this system has no notification area (tray), so a hidden window could not be brought back.",
  "settings.autostart": "Start at login",
  "settings.autoCheck": "Check for updates automatically",
  "settings.autoCheck.help":
    "At startup and once a day. While this is off, the application never goes online on its own.",
  "settings.broken":
    "The settings file is damaged or was written by a newer version. Defaults are shown and changes cannot be saved.",
  "settings.about": "About",
  "settings.version": "Version {version}",
  "settings.format": "Package format: {format}",
  "settings.format.none": "built from source",
  "settings.openLogs": "Open log folder",

  "updates.title": "Updates",
  "updates.unsupported": "This build is not from a package and is updated manually.",
  "updates.check": "Check for updates",
  "updates.checking": "Checking…",
  "updates.upToDate": "You have the latest version {version}",
  "updates.available": "Version {version} is available, size {size}",
  "updates.notes": "What's new",
  "updates.download": "Download",
  "updates.downloading": "Downloaded {received} of {size}",
  "updates.cancel": "Cancel",
  "updates.retry": "Retry",
  "updates.ready.command":
    "Package {version} is downloaded and verified. Install it with the command below (an administrator password is required) and restart the application.",
  "updates.ready.open":
    "Package {version} is downloaded and verified. Open the disk image and drag the application to Applications.",
  "updates.ready.run":
    "Package {version} is downloaded and verified. The installer will replace the application; this window will close.",
  "updates.ready.replace":
    "Package {version} is downloaded and verified. The AppImage file will be replaced and the application restarted.",
  "updates.ready.reveal":
    "Package {version} is downloaded and verified. Quit the application and replace its file with the downloaded one.",
  "updates.install.command": "Open package folder",
  "updates.install.open": "Open disk image",
  "updates.install.run": "Run installer",
  "updates.install.replace": "Install and restart",
  "updates.install.reveal": "Show file",
  "updates.copy": "Copy command",
  "updates.copied": "Command copied",

  "crash.title": "Something went wrong",
  "crash.body": "The interface ran into an error; a report was written to the log.",
  "crash.reload": "Reload interface",
  "crash.previous": "The previous session, started {at}, ended unexpectedly.",
  "crash.previousUnknown": "The previous session ended unexpectedly.",

  "errors.unknown": "Unknown error: {detail}",
  "errors.app.open-log-folder-failed": "Could not open the log folder: {detail}",
  "errors.autostart.failed": "Could not change start at login: {detail}",
  "errors.autostart.unavailable": "Start at login is not available on this system",
  "errors.settings.broken": "The settings file is damaged: changes will not be saved until it is fixed",
  "errors.settings.newer-version":
    "The settings file was written by a newer version of the application: changes will not be saved",
  "errors.settings.write-failed": "Could not save settings: {detail}",
  "errors.settings.invalid-theme": "Unknown theme “{value}”",
  "errors.settings.invalid-language": "Unknown language “{value}”",
  "errors.update.bad-release": "GitHub returned an unexpected release description",
  "errors.update.canceled": "Download canceled",
  "errors.update.host": "The download was redirected to an unexpected address and stopped",
  "errors.update.http": "GitHub responded with error {status}",
  "errors.update.launch": "Could not start the installation; install the package manually",
  "errors.update.mismatch": "The package did not match its checksum and was deleted; try downloading again",
  "errors.update.network": "Could not reach GitHub; check your network connection",
  "errors.update.no-package": "Release {version} has no package {package}; update the application manually",
  "errors.update.no-release": "GitHub has no suitable published release",
  "errors.update.no-sums": "The release has no checksum for the package ({file}), so it is not downloaded",
  "errors.update.not-ready": "The package has not been downloaded yet",
  "errors.update.rate-limited": "GitHub is temporarily limiting requests; try again in an hour",
  "errors.update.size": "The package size did not match the declared size or exceeds the limit",
  "errors.update.unknown": "Update failed: {detail}",
  "errors.update.write": "Could not save the package to disk",
};
