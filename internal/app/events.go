package app

import (
	"github.com/wailsapp/wails/v3/pkg/application"

	"github.com/yarkovaleksei/desktop-app-template-go-wails/internal/settings"
)

// События Go → фронтенд. Типы данных регистрируются в init: генератор
// привязок выводит из них типы событий для TypeScript.
const (
	// EventSettingsChanged — настройки сохранены; данные — settings.Settings.
	EventSettingsChanged = "settings:changed"
	// EventNavigate — открыть страницу интерфейса; данные — путь маршрута,
	// например "/settings". Шлёт Go-сторона: уведомление, пункт трея.
	EventNavigate = "app:navigate"
)

func init() {
	application.RegisterEvent[settings.Settings](EventSettingsChanged)
	application.RegisterEvent[string](EventNavigate)
}
