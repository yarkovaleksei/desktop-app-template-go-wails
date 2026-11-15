// Package app — сервисы Wails: тонкий слой между ядром в internal/ и
// фронтендом.
//
// Здесь и в main.go живёт весь код, который знает о Wails. Сервисы только
// переводят вызовы фронтенда в вызовы ядра и kit, ответы — в DTO через
// wire.Fill, ошибки — в *apperr.Error (errors.go). Пути от фронтенда сервисы
// не принимают: каталоги известны Go-стороне. Методы сервисов становятся
// TypeScript-привязками во frontend/bindings (`wails3 task bindings`).
package app

import (
	"context"
	"log/slog"
	"runtime"
	"time"

	"github.com/yarkovaleksei/go-desktop-kit/appdirs"
	"github.com/yarkovaleksei/go-desktop-kit/applog"
	"github.com/yarkovaleksei/go-desktop-kit/opener"
	"github.com/yarkovaleksei/go-desktop-kit/updater"
	"github.com/yarkovaleksei/go-desktop-kit/wailsx"
	"github.com/yarkovaleksei/go-desktop-kit/wire"

	"github.com/yarkovaleksei/desktop-app-template-go-wails/internal/appinfo"
)

// maxReportField — предел длины поля отчёта об ошибке фронтенда в журнале:
// зациклившийся рендер не должен раздувать файл.
const maxReportField = 8 << 10

// Info — сведения о приложении для экрана «О программе» и диагностики.
type Info struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	// Format — формат пакета этой сборки (deb, dmg, …); пусто — сборка из
	// исходников, обновление из приложения недоступно.
	Format        string `json:"format"`
	OS            string `json:"os"`
	Arch          string `json:"arch"`
	Dirs          Dirs   `json:"dirs"`
	TrayAvailable bool   `json:"trayAvailable"`
}

// Dirs — каталоги приложения.
type Dirs struct {
	Config string `json:"config"`
	Data   string `json:"data"`
	Cache  string `json:"cache"`
	Logs   string `json:"logs"`
}

// FrontendError — отчёт фронтенда о необработанной ошибке.
type FrontendError struct {
	// Source — где поймана: render (граница ошибок React), error,
	// unhandledrejection.
	Source  string `json:"source"`
	Message string `json:"message"`
	Stack   string `json:"stack"`
	// ComponentStack — стек компонентов React, если ошибка из рендера.
	ComponentStack string `json:"componentStack"`
}

// Crash — прошлый сеанс завершился не штатно.
type Crash struct {
	// StartedAt — время начала прошлого сеанса в RFC 3339; пусто, если маркер
	// сеанса повреждён.
	StartedAt string `json:"startedAt"`
}

// AppOptions — зависимости AppService. Пустые функции заменяются системными.
type AppOptions struct {
	Version string
	Format  updater.Format
	Dirs    appdirs.Dirs
	Log     *slog.Logger
	// Crash — applog.Log.PreviousCrash: nil, если прошлый сеанс завершён.
	Crash *applog.Crash
	// TrayAvailable — есть ли трей; nil — wailsx.TrayAvailable.
	TrayAvailable func() bool
	// OpenDir открывает каталог в файловом менеджере; nil — opener.Open.
	OpenDir func(ctx context.Context, path string) error
	// OnFrontendReady вызывается из FrontendReady; его задаёт smoke-режим.
	OnFrontendReady func()
}

// AppService — сведения о приложении, готовность фронтенда, журнал.
type AppService struct {
	o AppOptions
}

// NewAppService готовит сервис.
func NewAppService(o AppOptions) *AppService {
	if o.Log == nil {
		o.Log = slog.Default()
	}
	if o.TrayAvailable == nil {
		o.TrayAvailable = wailsx.TrayAvailable
	}
	if o.OpenDir == nil {
		o.OpenDir = opener.Open
	}
	return &AppService{o: o}
}

// Info возвращает название, версию, формат сборки, ОС, каталоги и наличие трея.
func (s *AppService) Info() Info {
	return wire.Fill(Info{
		Name:    appinfo.Name,
		Version: s.o.Version,
		Format:  string(s.o.Format),
		OS:      runtime.GOOS,
		Arch:    runtime.GOARCH,
		Dirs: Dirs{
			Config: s.o.Dirs.Config,
			Data:   s.o.Dirs.Data,
			Cache:  s.o.Dirs.Cache,
			Logs:   s.o.Dirs.Logs,
		},
		TrayAvailable: s.o.TrayAvailable(),
	})
}

// FrontendReady — фронтенд отрисовал первый экран. В обычной работе ничего не
// делает; в smoke-режиме завершает проверку успехом.
func (s *AppService) FrontendReady() {
	if s.o.OnFrontendReady != nil {
		s.o.OnFrontendReady()
	}
}

// ReportFrontendError пишет ошибку фронтенда в журнал.
func (s *AppService) ReportFrontendError(r FrontendError) {
	s.o.Log.Error("ошибка фронтенда",
		"source", clip(r.Source),
		"message", clip(r.Message),
		"stack", clip(r.Stack),
		"componentStack", clip(r.ComponentStack),
	)
}

// OpenLogFolder открывает каталог журнала в файловом менеджере.
func (s *AppService) OpenLogFolder(ctx context.Context) error {
	if err := s.o.OpenDir(ctx, s.o.Dirs.Logs); err != nil {
		s.o.Log.Warn("каталог журнала не открыт", "path", s.o.Dirs.Logs, "error", err)
		return toError(err, CodeOpenLogFolder)
	}
	return nil
}

// PreviousCrash сообщает, что прошлый сеанс оборвался; nil — завершился штатно.
func (s *AppService) PreviousCrash() *Crash {
	if s.o.Crash == nil {
		return nil
	}
	c := &Crash{}
	if !s.o.Crash.StartedAt.IsZero() {
		c.StartedAt = s.o.Crash.StartedAt.Format(time.RFC3339)
	}
	return c
}

func clip(s string) string {
	if len(s) <= maxReportField {
		return s
	}
	return s[:maxReportField] + "…"
}
