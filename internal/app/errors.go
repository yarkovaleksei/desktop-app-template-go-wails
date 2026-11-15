package app

import (
	"errors"

	"github.com/yarkovaleksei/go-desktop-kit/apperr"
	kitsettings "github.com/yarkovaleksei/go-desktop-kit/settings"
)

// Коды ошибок сервисов. Текст для человека по коду собирает фронтенд.
// Коды проверки настроек — в пакете settings (settings.CodeInvalidTheme, …).
const (
	CodeSettingsBroken       = "settings.broken"
	CodeSettingsNewerVersion = "settings.newer-version"
	CodeSettingsWrite        = "settings.write-failed"
	CodeOpenLogFolder        = "app.open-log-folder-failed"
	CodeAutostart            = "autostart.failed"
	CodeAutostartUnavailable = "autostart.unavailable"
)

// toError переводит ошибку ядра в *apperr.Error для фронтенда. Известные
// ошибки получают свой код, *apperr.Error из цепочки проходит как есть,
// остальные — код fallback и текст в параметре detail.
//
// Возвращает error, а не *apperr.Error: типизированный nil в интерфейсе error
// фронтенд принял бы за ошибку. Для nil возвращает nil.
func toError(err error, fallback string) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, kitsettings.ErrBroken):
		return apperr.Wrap(err, CodeSettingsBroken)
	case errors.Is(err, kitsettings.ErrNewerVersion):
		return apperr.Wrap(err, CodeSettingsNewerVersion)
	default:
		return apperr.From(err, fallback)
	}
}
