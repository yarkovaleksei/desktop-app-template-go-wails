// Desktop App — десктоп-приложение на Wails 3 и React.
package main

import (
	"embed"
	"log"
	"log/slog"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// version задаётся при сборке флагом -ldflags "-X main.version=…" из тега vX.Y.Z.
var version = "0.0.1-dev"

// Размер окна при запуске и минимальный размер.
const (
	windowWidth     = 1280
	windowHeight    = 800
	windowMinWidth  = 960
	windowMinHeight = 600
)

// Собранный фронтенд встраивается в бинарь.
//
//go:embed all:frontend/dist
var assets embed.FS

// Иконка приложения. На Linux Wails 3 с GTK 4 её не применяет: иконку окна
// там берёт рабочий стол из установленного .desktop-файла.
//
//go:embed build/appicon.png
var appIcon []byte

func main() {
	slog.Info("запуск", "version", version)

	app := application.New(application.Options{
		Name:        "Desktop App",
		Description: "Desktop application on Go, Wails 3, React and MUI",
		Icon:        appIcon,
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	// MinWidth и MinHeight обязательны: без них окно GTK 4 может открыться размером 1×1.
	// На Linux (GTK 4, Wails v3.0.0) Width и Height не применяются: окно открывается
	// по минимальному размеру. Обход появится вместе с восстановлением размера окна.
	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:     "Desktop App",
		Width:     windowWidth,
		Height:    windowHeight,
		MinWidth:  windowMinWidth,
		MinHeight: windowMinHeight,
		URL:       "/",
	})

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
