package settings

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/yarkovaleksei/go-desktop-kit/apperr"
	kitsettings "github.com/yarkovaleksei/go-desktop-kit/settings"
)

func newStore(t *testing.T) *Store {
	t.Helper()
	return New(filepath.Join(t.TempDir(), "settings.json"))
}

func TestDefaults(t *testing.T) {
	got, err := newStore(t).Load()
	if err != nil {
		t.Fatalf("Load без файла: %v", err)
	}
	want := Settings{UI: UI{Theme: ThemeSystem, Language: LanguageSystem}, Background: false, AutoCheckUpdates: false}
	if got != want {
		t.Errorf("умолчания %+v, ожидались %+v", got, want)
	}
	if err := Defaults().Validate(); err != nil {
		t.Errorf("умолчания не проходят проверку: %v", err)
	}
}

func TestSaveLoadEachField(t *testing.T) {
	cases := []struct {
		name string
		edit func(*Settings)
	}{
		{"тема", func(s *Settings) { s.UI.Theme = ThemeDark }},
		{"язык", func(s *Settings) { s.UI.Language = LanguageRU }},
		{"работа в фоне", func(s *Settings) { s.Background = true }},
		{"проверка обновлений", func(s *Settings) { s.AutoCheckUpdates = true }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			st := newStore(t)
			want := Defaults()
			c.edit(&want)
			if want == Defaults() {
				t.Fatal("правка не меняет умолчаний: тест ничего не проверяет")
			}
			if err := st.Save(want); err != nil {
				t.Fatalf("Save: %v", err)
			}
			got, err := New(st.Path()).Load()
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			if got != want {
				t.Errorf("прочитано %+v, записано %+v", got, want)
			}
		})
	}
}

func TestSaveRejectsInvalid(t *testing.T) {
	cases := []struct {
		name  string
		edit  func(*Settings)
		code  string
		value string
	}{
		{"тема", func(s *Settings) { s.UI.Theme = "blue" }, CodeInvalidTheme, "blue"},
		{"язык", func(s *Settings) { s.UI.Language = "de" }, CodeInvalidLanguage, "de"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			st := newStore(t)
			v := Defaults()
			c.edit(&v)
			err := st.Save(v)
			var ae *apperr.Error
			if !errors.As(err, &ae) {
				t.Fatalf("ошибка %v, ожидалась *apperr.Error", err)
			}
			if ae.Code != c.code || ae.Params["value"] != c.value {
				t.Errorf("код %q, value %q; ожидались %q, %q", ae.Code, ae.Params["value"], c.code, c.value)
			}
			if _, err := os.Stat(st.Path()); !errors.Is(err, os.ErrNotExist) {
				t.Errorf("файл создан несмотря на ошибку: %v", err)
			}
		})
	}
}

func TestLoadNormalizesHandEdited(t *testing.T) {
	st := newStore(t)
	data := `{"version":1,"data":{"ui":{"theme":"blue","language":"de"},"background":true}}`
	if err := os.WriteFile(st.Path(), []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := st.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	want := Settings{Background: true}
	if got != want {
		t.Errorf("прочитано %+v, ожидалось %+v", got, want)
	}
}

func TestOpen(t *testing.T) {
	st := newStore(t)
	got, err := st.Open()
	if err != nil || got != Defaults() {
		t.Fatalf("Open без файла: %+v, %v", got, err)
	}
	if _, err := os.Stat(st.Path()); err != nil {
		t.Fatalf("файл умолчаний не создан: %v", err)
	}
	want := Defaults()
	want.Background = true
	if err := st.Save(want); err != nil {
		t.Fatal(err)
	}
	if got, err := New(st.Path()).Open(); err != nil || got != want {
		t.Errorf("Open существующего файла: %+v, %v; ожидалось %+v", got, err, want)
	}

	broken := newStore(t)
	if err := os.WriteFile(broken.Path(), []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := broken.Open(); !errors.Is(err, kitsettings.ErrBroken) {
		t.Errorf("Open испорченного файла: %v", err)
	}
	if data, _ := os.ReadFile(broken.Path()); string(data) != "{" {
		t.Errorf("испорченный файл перезаписан: %q", data)
	}
}

func TestBrokenFileKept(t *testing.T) {
	st := newStore(t)
	if err := os.WriteFile(st.Path(), []byte("{не json"), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := st.Load()
	if !errors.Is(err, kitsettings.ErrBroken) {
		t.Fatalf("ошибка %v, ожидалась ErrBroken", err)
	}
	if got != Defaults() || !st.Broken() {
		t.Errorf("испорченный файл: %+v, Broken=%v", got, st.Broken())
	}
	if err := st.Save(Defaults()); !errors.Is(err, kitsettings.ErrBroken) {
		t.Errorf("Save поверх испорченного файла: %v", err)
	}
}
