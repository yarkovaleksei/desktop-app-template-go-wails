// Desktop App — десктоп-приложение на Wails 3 и React.
//
// main.go только собирает зависимости: каталоги, журнал, настройки, сервисы,
// окно. Код, который знает о Wails, живёт здесь и в internal/app; ядро в
// internal/ от него не зависит.
package main

import (
	"embed"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/yarkovaleksei/go-desktop-kit/appdirs"
	"github.com/yarkovaleksei/go-desktop-kit/applog"
	"github.com/yarkovaleksei/go-desktop-kit/smoke"
	"github.com/yarkovaleksei/go-desktop-kit/updater"
	"github.com/yarkovaleksei/go-desktop-kit/wailsx"

	"github.com/yarkovaleksei/desktop-app-template-go-wails/internal/app"
	"github.com/yarkovaleksei/desktop-app-template-go-wails/internal/appinfo"
	"github.com/yarkovaleksei/desktop-app-template-go-wails/internal/settings"
)

// version задаётся при сборке флагом -ldflags "-X main.version=…" из тега vX.Y.Z.
var version = "0.0.1-dev"

// packageFormat — формат пакета этой сборки (deb, dmg, nsis, …). Его вшивают
// задачи упаковки флагом -ldflags "-X main.packageFormat=…"; пакеты Linux
// вместо этого ставят файл-метку (updater.Naming.Marker). Пусто — сборка из
// исходников, обновление из приложения недоступно.
var packageFormat = ""

// Размер окна при запуске и минимальный размер.
const (
	windowWidth     = 1280
	windowHeight    = 800
	windowMinWidth  = 960
	windowMinHeight = 600
)

// defaultSmokeTimeout — сколько smoke-режим ждёт ответа фронтенда.
const defaultSmokeTimeout = 30 * time.Second

// Собранный фронтенд встраивается в бинарь.
//
//go:embed all:frontend/dist
var assets embed.FS

// Иконка приложения. На Linux Wails 3 с GTK 4 её не применяет: иконку окна
// там берёт рабочий стол из установленного .desktop-файла (D-041).
//
//go:embed build/appicon.png
var appIcon []byte

func main() {
	os.Exit(run())
}

// run запускает приложение и возвращает код выхода: ненулевой, если
// приложение не запустилось или smoke-проверка не пройдена.
func run() int {
	showVersion := flag.Bool("version", false, "напечатать версию и выйти")
	smokeReport := flag.String("smoke-report", "", "проверка запуска: дождаться ответа фронтенда, записать отчёт JSON в этот файл и выйти")
	smokeTimeout := flag.Duration("smoke-timeout", defaultSmokeTimeout, "сколько проверка запуска ждёт ответа фронтенда")
	flag.Parse()

	format := detectFormat()
	if *showVersion {
		fmt.Println(versionLine(format))
		return 0
	}
	if *smokeReport != "" && *smokeTimeout <= 0 {
		fmt.Fprintln(os.Stderr, "--smoke-timeout должен быть положительным")
		return 2
	}

	dirs, err := appdirs.Resolve(appinfo.Slug, appinfo.EnvPrefix, appdirs.FromEnvironment())
	if err != nil {
		fmt.Fprintf(os.Stderr, "каталоги приложения не определены: %v\n", err)
		return 1
	}

	// Без журнала приложение работает: записи идут в stderr.
	var crash *applog.Crash
	appLog, err := applog.Open(applog.Options{Path: filepath.Join(dirs.Logs, appinfo.Slug+".log")})
	if err != nil {
		slog.Error("журнал не открыт, записи идут в stderr", "error", err)
	} else {
		defer appLog.Close()
		slog.SetDefault(appLog.Logger())
		crash = appLog.PreviousCrash()
	}
	log := slog.Default()
	log.Info("запуск", "version", version, "format", string(format), "os", runtime.GOOS, "arch", runtime.GOARCH)
	if crash != nil {
		log.Warn("прошлый сеанс завершился не штатно", "startedAt", crash.StartedAt)
	}

	store := settings.New(filepath.Join(dirs.Config, "settings.json"))
	if _, err := store.Open(); err != nil {
		log.Warn("настройки не прочитаны, действуют умолчания", "path", store.Path(), "error", err)
	}

	updates := wailsx.NewUpdatesService(updater.Options{
		App:      appinfo.Slug,
		Repo:     appinfo.Repo,
		Current:  version,
		Format:   format,
		Dir:      filepath.Join(dirs.Data, "updates"),
		AppImage: os.Getenv("APPIMAGE"),
		Log:      log,
	})

	var wailsApp *application.App
	var check *smokeCheck
	if *smokeReport != "" {
		check = &smokeCheck{result: smoke.New(*smokeReport), log: log}
	}

	appService := app.NewAppService(app.AppOptions{
		Version:         version,
		Format:          format,
		Dirs:            dirs,
		Log:             log,
		Crash:           crash,
		OnFrontendReady: check.onReady(func() { wailsApp.Quit() }),
	})
	settingsService := app.NewSettingsService(app.SettingsOptions{Store: store, Log: log})

	// window объявлен до application.New: повторный запуск показывает окно,
	// которое создаётся после.
	var window *application.WebviewWindow

	wailsApp = application.New(application.Options{
		Name:        appinfo.Name,
		Description: appinfo.Description,
		Icon:        appIcon,
		Logger:      log,
		Services: []application.Service{
			application.NewService(appService),
			application.NewService(settingsService),
			application.NewService(updates),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
		// Один экземпляр на каталог настроек: повторный запуск (ярлык, меню)
		// показывает окно первого. Другой DESKTOP_APP_HOME — другой экземпляр,
		// поэтому smoke и разработка не мешают установленному приложению.
		SingleInstance: wailsx.SingleInstance(appinfo.ProductID, dirs.Config, func([]string) {
			log.Info("повторный запуск: показать окно")
			if window != nil {
				showWindow(window)
			}
		}),
	})

	// MinWidth и MinHeight обязательны: без них окно GTK 4 может открыться размером 1×1.
	// Сохранённые размер, положение и разворот подставляет RestoreWindow, а
	// TrackWindow запоминает их при закрытии и выходе (D-037). На Linux GTK 4
	// не применяет начальный размер надёжно, и до запуска экраны неизвестны;
	// обход — внутри wailsx (D-040), здесь кода для этого нет.
	windowFile := filepath.Join(dirs.Config, "window.json")
	windowOptions := application.WebviewWindowOptions{
		Title:     appinfo.Name,
		Width:     windowWidth,
		Height:    windowHeight,
		MinWidth:  windowMinWidth,
		MinHeight: windowMinHeight,
		URL:       "/",
	}
	wailsx.RestoreWindow(wailsApp, &windowOptions, windowFile)
	window = wailsApp.Window.NewWithOptions(windowOptions)
	wailsx.TrackWindow(window, windowFile)

	if check != nil {
		timer := time.AfterFunc(*smokeTimeout, func() {
			check.finish(smoke.Report{Error: fmt.Sprintf("фронтенд не ответил за %s", *smokeTimeout)})
			wailsApp.Quit()
		})
		defer timer.Stop()
	}

	runErr := wailsApp.Run()
	if runErr != nil {
		log.Error("приложение завершилось с ошибкой", "error", runErr)
	}
	if check != nil {
		// Исход принимается один: если фронтенд или таймер уже ответили, это ничего не меняет.
		check.finish(smoke.Report{Error: "приложение завершилось до ответа фронтенда"})
		log.Info("проверка запуска завершена", "passed", check.result.Passed(), "report", *smokeReport)
		if !check.result.Passed() {
			return 1
		}
	}
	if runErr != nil {
		return 1
	}
	return 0
}

// showWindow показывает окно и выводит его на передний план.
func showWindow(w *application.WebviewWindow) {
	w.Show()
	w.Focus()
}

// detectFormat определяет формат пакета этой сборки (D-047).
func detectFormat() updater.Format {
	exe, err := os.Executable()
	if err == nil {
		if real, err := filepath.EvalSymlinks(exe); err == nil {
			exe = real
		}
	}
	naming := updater.Naming{App: appinfo.Slug}
	return naming.DetectFormat(packageFormat, runtime.GOOS, exe, os.Getenv("APPIMAGE"), os.ReadFile)
}

// versionLine — строка --version: «Desktop App 1.2.3 (linux/amd64; deb)».
func versionLine(f updater.Format) string {
	format := string(f)
	if f == updater.FormatNone {
		format = "source"
	}
	return fmt.Sprintf("%s %s (%s/%s; %s)", appinfo.Name, version, runtime.GOOS, runtime.GOARCH, format)
}

// smokeCheck — проверка запуска: окно открылось и фронтенд вызвал
// AppService.FrontendReady. Исход пишет kit/smoke в файл отчёта.
type smokeCheck struct {
	result *smoke.Result
	log    *slog.Logger
}

// onReady возвращает обработчик готовности фронтенда: записать успех и выйти
// через quit. Без smoke-режима (nil) — nil, и FrontendReady ничего не делает.
func (c *smokeCheck) onReady(quit func()) func() {
	if c == nil {
		return nil
	}
	return func() {
		c.finish(smoke.Report{FrontendReady: true})
		quit()
	}
}

// finish принимает исход проверки; принят будет только первый.
func (c *smokeCheck) finish(rep smoke.Report) {
	rep.App = appinfo.Name
	rep.Version = version
	if err := c.result.Complete(rep); err != nil {
		c.log.Error("отчёт проверки запуска не записан", "error", err)
	}
}
