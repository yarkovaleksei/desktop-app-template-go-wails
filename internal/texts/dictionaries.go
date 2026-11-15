package texts

// Ключи текстов. Параметры в фигурных скобках перечислены у ключа.
const (
	TrayShow         = "tray.show"
	TrayCheckUpdates = "tray.check-updates"
	TrayQuit         = "tray.quit"
	// UpdateAvailableTitle и UpdateAvailableBody — уведомление о новой версии;
	// в тексте — {version}.
	UpdateAvailableTitle = "notify.update-available.title"
	UpdateAvailableBody  = "notify.update-available.body"
	// AutostartComment — описание записи автозапуска в системе.
	AutostartComment = "autostart.comment"
)

var dictionaries = map[string]map[string]string{
	RU: {
		TrayShow:             "Показать",
		TrayCheckUpdates:     "Проверить обновления",
		TrayQuit:             "Выйти",
		UpdateAvailableTitle: "Доступна новая версия",
		UpdateAvailableBody:  "Версия {version} готова к установке. Нажмите, чтобы открыть настройки.",
		AutostartComment:     "Запуск при входе в систему",
	},
	EN: {
		TrayShow:             "Show",
		TrayCheckUpdates:     "Check for updates",
		TrayQuit:             "Quit",
		UpdateAvailableTitle: "New version available",
		UpdateAvailableBody:  "Version {version} is ready to install. Click to open settings.",
		AutostartComment:     "Start at login",
	},
}
