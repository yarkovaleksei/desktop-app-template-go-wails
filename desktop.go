package main

import (
	"context"
	_ "embed"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/yarkovaleksei/go-desktop-kit/autostart"
	"github.com/yarkovaleksei/go-desktop-kit/wailsx"

	"github.com/yarkovaleksei/desktop-app-template-go-wails/internal/app"
	"github.com/yarkovaleksei/desktop-app-template-go-wails/internal/appinfo"
	"github.com/yarkovaleksei/desktop-app-template-go-wails/internal/settings"
	"github.com/yarkovaleksei/desktop-app-template-go-wails/internal/texts"
)

// Значок трея. На macOS — шаблонное изображение: строка меню перекрашивает
// его под светлую и тёмную тему (D-039).
var (
	//go:embed build/trayicon.png
	trayIcon []byte
	//go:embed build/trayicon-template.png
	trayIconTemplate []byte
)

// settingsPage — маршрут страницы настроек во фронтенде.
const settingsPage = "/settings"

// desktop связывает приложение с рабочим столом: трей и работа в фоне,
// автозапуск, завершение по сигналу. В smoke-режиме не создаётся.
type desktop struct {
	app     *application.App
	window  *application.WebviewWindow
	store   *settings.Store
	updates *wailsx.UpdatesService
	log     *slog.Logger

	tray      *wailsx.Tray
	quit      func()
	autostart autostart.Manager
}

// setupDesktop добавляет трей и обработку SIGINT и SIGTERM. Вызывается после
// wailsx.TrackWindow: хук трея, прячущий окно, должен идти после хука,
// который запоминает границы.
func setupDesktop(a *application.App, w *application.WebviewWindow, store *settings.Store, updates *wailsx.UpdatesService, log *slog.Logger) *desktop {
	d := &desktop{app: a, window: w, store: store, updates: updates, log: log}
	if !wailsx.TrayAvailable() {
		log.Warn("трей недоступен: закрытие окна завершает приложение, работа в фоне не действует")
	}
	d.tray, d.quit = wailsx.SetupTray(a, w, wailsx.TrayOptions{
		Icon:         trayIcon,
		TemplateIcon: trayIconTemplate,
		Tooltip:      appinfo.Name,
		Items:        d.trayItems,
		HideOnClose:  func() bool { return d.settings().Background },
	})
	// Автозапуск стартует приложение с --hidden. Описание записи — на языке
	// запуска; на Linux значок берётся из темы по слагу (.desktop пакета).
	d.autostart = wailsx.Autostart(a, autostart.Entry{
		Slug:    appinfo.Slug,
		Name:    appinfo.Name,
		Comment: texts.T(d.lang(), texts.AutostartComment, nil),
		Icon:    appinfo.Slug,
	})
	d.handleSignals()
	return d
}

// settings — текущие настройки. Испорченный файл — умолчания: о нём уже
// сообщили при запуске, а здесь важно только поведение.
func (d *desktop) settings() settings.Settings {
	v, _ := d.store.Load()
	return v
}

// lang — язык Go-текстов по настройке и системе (D-033).
func (d *desktop) lang() string {
	return texts.Language(d.settings().UI.Language)
}

// trayItems строит меню трея на текущем языке. Вызывается при создании трея
// и при каждом Refresh.
func (d *desktop) trayItems() []wailsx.TrayItem {
	lang := d.lang()
	return []wailsx.TrayItem{
		{Label: texts.T(lang, texts.TrayShow, nil), OnClick: func() { showWindow(d.window) }},
		{Label: texts.T(lang, texts.TrayCheckUpdates, nil), OnClick: d.checkUpdatesNow},
		{},
		// quit присваивается после построения меню: берётся в момент нажатия.
		{Label: texts.T(lang, texts.TrayQuit, nil), OnClick: func() { d.quit() }},
	}
}

// checkUpdatesNow показывает окно на странице настроек и сразу проверяет
// обновления; результат приходит во фронтенд событием update:status.
func (d *desktop) checkUpdatesNow() {
	d.openSettings()
	go d.updates.Check(context.Background())
}

// openSettings показывает окно и открывает страницу настроек.
func (d *desktop) openSettings() {
	showWindow(d.window)
	wailsx.Emit(app.EventNavigate, settingsPage)
}

// settingsChanged — настройки сохранены: меню трея перестраивается на
// новом языке.
func (d *desktop) settingsChanged(settings.Settings) {
	d.tray.Refresh()
}

// handleSignals завершает приложение по SIGINT и SIGTERM (выход из сеанса,
// kill) через quit трея: обычное закрытие окна при работе в фоне только
// прятало бы его.
func (d *desktop) handleSignals() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-ch
		d.log.Info("сигнал завершения", "signal", sig.String())
		d.quit()
	}()
}
