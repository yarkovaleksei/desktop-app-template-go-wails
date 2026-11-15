package app

import (
	"log/slog"

	"github.com/yarkovaleksei/go-desktop-kit/apperr"
	"github.com/yarkovaleksei/go-desktop-kit/autostart"
	"github.com/yarkovaleksei/go-desktop-kit/wailsx"
	"github.com/yarkovaleksei/go-desktop-kit/wire"

	"github.com/yarkovaleksei/desktop-app-template-go-wails/internal/settings"
)

// SettingsView — что показывает экран «Настройки».
type SettingsView struct {
	Settings settings.Settings `json:"settings"`
	// Broken — файл настроек испорчен или записан более новой версией:
	// показаны умолчания, а сохранение откажет, пока файл не исправят.
	Broken bool `json:"broken"`
}

// SettingsOptions — зависимости SettingsService.
type SettingsOptions struct {
	Store *settings.Store
	Log   *slog.Logger
	// Emit отправляет событие во фронтенд; nil — wailsx.Emit.
	Emit func(name string, data any)
	// OnChange вызывается после успешного Save, до события, — Go-стороне
	// (язык меню трея, фоновая проверка обновлений).
	OnChange func(settings.Settings)
	// Autostart возвращает менеджер автозапуска. Он создаётся после
	// приложения Wails, поэтому передаётся функцией; nil или nil-результат —
	// автозапуск недоступен (smoke-режим).
	Autostart func() autostart.Manager
}

// SettingsService читает и сохраняет настройки.
type SettingsService struct {
	o SettingsOptions
}

// NewSettingsService готовит сервис.
func NewSettingsService(o SettingsOptions) *SettingsService {
	if o.Log == nil {
		o.Log = slog.Default()
	}
	if o.Emit == nil {
		o.Emit = wailsx.Emit
	}
	return &SettingsService{o: o}
}

// Get возвращает текущие настройки. Испорченный файл не ошибка вызова: экран
// показывает умолчания и признак Broken.
func (s *SettingsService) Get() SettingsView {
	v, err := s.o.Store.Load()
	if err != nil {
		s.o.Log.Warn("настройки не прочитаны, показаны умолчания", "path", s.o.Store.Path(), "error", err)
	}
	return wire.Fill(SettingsView{Settings: v, Broken: s.o.Store.Broken()})
}

// Save проверяет и сохраняет настройки, затем сообщает о смене событием
// EventSettingsChanged. Возвращает сохранённое состояние.
func (s *SettingsService) Save(v settings.Settings) (SettingsView, error) {
	if err := s.o.Store.Save(v); err != nil {
		s.o.Log.Warn("настройки не сохранены", "path", s.o.Store.Path(), "error", err)
		return SettingsView{}, toError(err, CodeSettingsWrite)
	}
	if s.o.OnChange != nil {
		s.o.OnChange(v)
	}
	s.o.Emit(EventSettingsChanged, v)
	return wire.Fill(SettingsView{Settings: v}), nil
}

// Autostart сообщает, включён ли запуск при входе в систему. Состояние
// читается из ОС, в настройках не хранится: пользователь мог выключить
// запись системными средствами.
func (s *SettingsService) Autostart() (bool, error) {
	m := s.autostart()
	if m == nil {
		return false, nil
	}
	on, err := m.Enabled()
	if err != nil {
		s.o.Log.Warn("состояние автозапуска не прочитано", "error", err)
		return false, toError(err, CodeAutostart)
	}
	return on, nil
}

// SetAutostart включает или выключает запуск при входе в систему. Запуск
// идёт с флагом --hidden: окно не показывается, если включена работа в фоне
// и есть трей (D-034).
func (s *SettingsService) SetAutostart(on bool) error {
	m := s.autostart()
	if m == nil {
		return apperr.New(CodeAutostartUnavailable)
	}
	if err := m.SetEnabled(on); err != nil {
		s.o.Log.Warn("автозапуск не изменён", "on", on, "error", err)
		return toError(err, CodeAutostart)
	}
	s.o.Log.Info("автозапуск", "on", on)
	return nil
}

func (s *SettingsService) autostart() autostart.Manager {
	if s.o.Autostart == nil {
		return nil
	}
	return s.o.Autostart()
}
