package app

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/yarkovaleksei/go-desktop-kit/appdirs"
	"github.com/yarkovaleksei/go-desktop-kit/apperr"
	"github.com/yarkovaleksei/go-desktop-kit/applog"
	"github.com/yarkovaleksei/go-desktop-kit/updater"

	"github.com/yarkovaleksei/desktop-app-template-go-wails/internal/appinfo"
	"github.com/yarkovaleksei/desktop-app-template-go-wails/internal/settings"
)

func testDirs(t *testing.T) appdirs.Dirs {
	t.Helper()
	root := t.TempDir()
	return appdirs.Dirs{
		Config: filepath.Join(root, "config"),
		Data:   filepath.Join(root, "data"),
		Cache:  filepath.Join(root, "cache"),
		Logs:   filepath.Join(root, "cache", "logs"),
	}
}

func bufLog() (*slog.Logger, *bytes.Buffer) {
	var b bytes.Buffer
	return slog.New(slog.NewTextHandler(&b, nil)), &b
}

func TestInfo(t *testing.T) {
	dirs := testDirs(t)
	s := NewAppService(AppOptions{
		Version:       "1.2.3",
		Format:        updater.FormatDeb,
		Dirs:          dirs,
		TrayAvailable: func() bool { return true },
	})
	got := s.Info()
	want := Info{
		Name:          appinfo.Name,
		Version:       "1.2.3",
		Format:        "deb",
		OS:            runtime.GOOS,
		Arch:          runtime.GOARCH,
		Dirs:          Dirs{Config: dirs.Config, Data: dirs.Data, Cache: dirs.Cache, Logs: dirs.Logs},
		TrayAvailable: true,
	}
	if got != want {
		t.Errorf("Info = %+v, ожидалось %+v", got, want)
	}
}

func TestFrontendReady(t *testing.T) {
	NewAppService(AppOptions{}).FrontendReady() // без smoke — ничего не делает
	called := 0
	NewAppService(AppOptions{OnFrontendReady: func() { called++ }}).FrontendReady()
	if called != 1 {
		t.Errorf("OnFrontendReady вызван %d раз", called)
	}
}

func TestReportFrontendError(t *testing.T) {
	log, buf := bufLog()
	s := NewAppService(AppOptions{Log: log})
	s.ReportFrontendError(FrontendError{Source: "render", Message: "boom", Stack: strings.Repeat("x", maxReportField+100)})
	out := buf.String()
	for _, want := range []string{"level=ERROR", "source=render", "message=boom"} {
		if !strings.Contains(out, want) {
			t.Errorf("в журнале нет %q: %s", want, out)
		}
	}
	if strings.Contains(out, strings.Repeat("x", maxReportField+1)) {
		t.Error("длинный стек не обрезан")
	}
}

func TestOpenLogFolder(t *testing.T) {
	dirs := testDirs(t)
	var opened string
	s := NewAppService(AppOptions{Dirs: dirs, OpenDir: func(_ context.Context, p string) error { opened = p; return nil }})
	if err := s.OpenLogFolder(context.Background()); err != nil {
		t.Fatalf("OpenLogFolder: %v", err)
	}
	if opened != dirs.Logs {
		t.Errorf("открыт %q, ожидался %q", opened, dirs.Logs)
	}

	log, _ := bufLog()
	s = NewAppService(AppOptions{Dirs: dirs, Log: log, OpenDir: func(context.Context, string) error { return errors.New("нет файлового менеджера") }})
	err := s.OpenLogFolder(context.Background())
	var ae *apperr.Error
	if !errors.As(err, &ae) || ae.Code != CodeOpenLogFolder {
		t.Errorf("ошибка %v, ожидался код %s", err, CodeOpenLogFolder)
	}
}

func TestPreviousCrash(t *testing.T) {
	if c := NewAppService(AppOptions{}).PreviousCrash(); c != nil {
		t.Errorf("без сбоя: %+v", c)
	}
	at := time.Date(2026, 11, 15, 10, 0, 0, 0, time.UTC)
	c := NewAppService(AppOptions{Crash: &applog.Crash{StartedAt: at}}).PreviousCrash()
	if c == nil || c.StartedAt != "2026-11-15T10:00:00Z" {
		t.Errorf("сбой: %+v", c)
	}
	c = NewAppService(AppOptions{Crash: &applog.Crash{}}).PreviousCrash()
	if c == nil || c.StartedAt != "" {
		t.Errorf("сбой с повреждённым маркером: %+v", c)
	}
}

type emitted struct {
	name string
	data any
}

func newSettingsService(t *testing.T) (*SettingsService, *settings.Store, *[]emitted) {
	t.Helper()
	st := settings.New(filepath.Join(testDirs(t).Config, "settings.json"))
	var events []emitted
	log, _ := bufLog()
	s := NewSettingsService(SettingsOptions{Store: st, Log: log, Emit: func(n string, d any) { events = append(events, emitted{n, d}) }})
	return s, st, &events
}

func TestSettingsSave(t *testing.T) {
	s, _, events := newSettingsService(t)
	v := settings.Defaults()
	v.UI.Theme = settings.ThemeDark
	v.AutoCheckUpdates = true
	got, err := s.Save(v)
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	if got.Settings != v || got.Broken {
		t.Errorf("Save вернул %+v", got)
	}
	if again := s.Get(); again.Settings != v {
		t.Errorf("Get после Save: %+v", again)
	}
	if len(*events) != 1 || (*events)[0].name != EventSettingsChanged || (*events)[0].data != v {
		t.Errorf("события: %+v", *events)
	}
}

func TestSettingsSaveInvalidTheme(t *testing.T) {
	s, _, events := newSettingsService(t)
	v := settings.Defaults()
	v.UI.Theme = "blue"
	_, err := s.Save(v)
	var ae *apperr.Error
	if !errors.As(err, &ae) || ae.Code != settings.CodeInvalidTheme {
		t.Fatalf("ошибка %v, ожидался код %s", err, settings.CodeInvalidTheme)
	}
	if len(*events) != 0 {
		t.Errorf("событие после ошибки: %+v", *events)
	}
}

func TestSettingsBrokenFile(t *testing.T) {
	s, st, _ := newSettingsService(t)
	if err := os.MkdirAll(filepath.Dir(st.Path()), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(st.Path(), []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	if v := s.Get(); !v.Broken || v.Settings != settings.Defaults() {
		t.Errorf("Get испорченного файла: %+v", v)
	}
	_, err := s.Save(settings.Defaults())
	var ae *apperr.Error
	if !errors.As(err, &ae) || ae.Code != CodeSettingsBroken {
		t.Errorf("ошибка %v, ожидался код %s", err, CodeSettingsBroken)
	}
}

func TestToErrorNil(t *testing.T) {
	if err := toError(nil, "x"); err != nil {
		t.Errorf("toError(nil) = %#v", err)
	}
}
