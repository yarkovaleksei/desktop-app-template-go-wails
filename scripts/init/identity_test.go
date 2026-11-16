package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/yarkovaleksei/desktop-app-template-go-wails/internal/appinfo"
)

// TestTemplateNamesMatchAppinfo держит имена шаблона в identity.go равными
// internal/appinfo: иначе инициализация искала бы не те строки.
func TestTemplateNamesMatchAppinfo(t *testing.T) {
	for _, c := range []struct{ what, got, want string }{
		{"Name", templateName, appinfo.Name},
		{"Slug", templateSlug, appinfo.Slug},
		{"Repo", templateRepo, appinfo.Repo},
		{"ProductID", templateID, appinfo.ProductID},
		{"EnvPrefix", templateEnvPrefix, appinfo.EnvPrefix},
		{"Description", templateDescription, appinfo.Description},
	} {
		if c.got != c.want {
			t.Errorf("%s: в scripts/init %q, в internal/appinfo %q", c.what, c.got, c.want)
		}
	}
	mod, err := os.ReadFile(filepath.Join(moduleRootDir(t), "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(mod), "module "+templateModule+"\n") {
		t.Errorf("go.mod: модуль не %s", templateModule)
	}
}

// TestTemplateSkillsMatchDirs сверяет templateSkills с каталогами скиллов.
func TestTemplateSkillsMatchDirs(t *testing.T) {
	entries, err := os.ReadDir(filepath.Join(moduleRootDir(t), ".agents", "skills"))
	if err != nil {
		t.Fatal(err)
	}
	var dirs []string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), templateSkillPrefix) {
			dirs = append(dirs, e.Name())
		}
	}
	if !slices.Equal(dirs, templateSkills) {
		t.Errorf("скиллы с префиксом %s: каталоги %v, в templateSkills %v", templateSkillPrefix, dirs, templateSkills)
	}
}

func TestDerive(t *testing.T) {
	got, err := Derive(Flags{
		Name:        "My App",
		Repo:        "Some-Owner/my-app",
		Description: "Notes on the desktop",
	}, 2031)
	if err != nil {
		t.Fatal(err)
	}
	want := Identity{
		Name:        "My App",
		Slug:        "my-app",
		Module:      "github.com/Some-Owner/my-app",
		Repo:        "Some-Owner/my-app",
		Owner:       "Some-Owner",
		RepoName:    "my-app",
		PagesHost:   "some-owner.github.io",
		ID:          "io.github.someowner.myapp",
		EnvPrefix:   "MY_APP",
		SkillPrefix: "my-app-",
		Description: "Notes on the desktop",
		Author:      "Some-Owner",
		Year:        "2031",
	}
	if got != want {
		t.Errorf("Derive:\n got %+v\nwant %+v", got, want)
	}
}

func TestDeriveExplicit(t *testing.T) {
	got, err := Derive(Flags{
		Name:        "Photo Viewer 2",
		Slug:        "photos",
		Module:      "example.com/tools/photos",
		Repo:        "acme/photo-viewer",
		ID:          "com.example.photos",
		Description: "Photo viewer",
		Author:      "ACME Ltd.",
		Year:        2030,
		Color:       "#2b6cb0",
	}, 2031)
	if err != nil {
		t.Fatal(err)
	}
	if got.Slug != "photos" || got.Module != "example.com/tools/photos" || got.ID != "com.example.photos" ||
		got.Author != "ACME Ltd." || got.Year != "2030" || got.Color != "#2b6cb0" ||
		got.EnvPrefix != "PHOTOS" || got.SkillPrefix != "photos-" || got.RepoName != "photo-viewer" {
		t.Errorf("явные значения не сохранены: %+v", got)
	}
}

func TestSlugify(t *testing.T) {
	for in, want := range map[string]string{
		"My App":         "my-app",
		"Photo Viewer 2": "photo-viewer-2",
		"  ACME  tools ": "acme-tools",
		"3D Editor":      "d-editor",
		"Заметки":        "",
		"Notes+":         "notes",
	} {
		if got := slugify(in); got != want {
			t.Errorf("slugify(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestDeriveRejects — недопустимые значения отклоняются с ошибкой, которая
// называет флаг.
func TestDeriveRejects(t *testing.T) {
	base := Flags{Name: "My App", Repo: "owner/my-app", Description: "Notes"}
	for _, c := range []struct {
		name string
		edit func(*Flags)
		flag string
	}{
		{"нет названия", func(f *Flags) { f.Name = "" }, "-name"},
		{"кавычка в названии", func(f *Flags) { f.Name = `My "App"` }, "-name"},
		{"слаг с пробелом", func(f *Flags) { f.Slug = "my app" }, "-slug"},
		{"слаг с заглавной", func(f *Flags) { f.Slug = "My-app" }, "-slug"},
		{"слаг с двумя дефисами", func(f *Flags) { f.Slug = "my--app" }, "-slug"},
		{"слаг с цифры", func(f *Flags) { f.Slug = "1app" }, "-slug"},
		{"слаг не выводится", func(f *Flags) { f.Name = "Заметки" }, "-slug"},
		{"нет репозитория", func(f *Flags) { f.Repo = "" }, "-repo"},
		{"репозиторий без владельца", func(f *Flags) { f.Repo = "my-app" }, "-repo"},
		{"модуль без домена", func(f *Flags) { f.Module = "my-app" }, "-module"},
		{"модуль без пути", func(f *Flags) { f.Module = "example.com" }, "-module"},
		{"модуль с хостом без точки", func(f *Flags) { f.Module = "owner/my-app" }, "-module"},
		{"идентификатор без точек", func(f *Flags) { f.ID = "myapp" }, "-id"},
		{"идентификатор из двух частей", func(f *Flags) { f.ID = "com.myapp" }, "-id"},
		{"идентификатор с дефисом", func(f *Flags) { f.ID = "io.github.owner.my-app" }, "-id"},
		{"идентификатор не обратный", func(f *Flags) { f.ID = "myapp.owner.github.io" }, "-id"},
		{"нет описания", func(f *Flags) { f.Description = "" }, "-description"},
		{"описание в две строки", func(f *Flags) { f.Description = "a\nb" }, "-description"},
		{"амперсанд у автора", func(f *Flags) { f.Author = "A & B" }, "-author"},
		{"год", func(f *Flags) { f.Year = 99 }, "-year"},
		{"цвет", func(f *Flags) { f.Color = "#12345" }, "-color"},
		{"имя шаблона", func(f *Flags) { f.Name = templateName }, "-name"},
		{"слаг шаблона", func(f *Flags) { f.Slug = templateSlug }, "-slug"},
		{"репозиторий шаблона", func(f *Flags) { f.Repo = templateRepo }, "-repo"},
	} {
		t.Run(c.name, func(t *testing.T) {
			f := base
			c.edit(&f)
			_, err := Derive(f, 2031)
			if err == nil {
				t.Fatal("значение принято")
			}
			if !strings.Contains(err.Error(), c.flag+" ") && !strings.Contains(err.Error(), c.flag+":") {
				t.Errorf("ошибка не называет %s: %v", c.flag, err)
			}
		})
	}
}

func TestReplaceAll(t *testing.T) {
	id, err := Derive(Flags{Name: "Desktop App Pro", Slug: "desktop-app-pro", Repo: "owner/desktop-app-pro", Description: "Pro"}, 2031)
	if err != nil {
		t.Fatal(err)
	}
	for in, want := range map[string]string{
		"module github.com/yarkovaleksei/desktop-app-template-go-wails\n":      "module github.com/owner/desktop-app-pro\n",
		"https://yarkovaleksei.github.io/desktop-app-template-go-wails/":       "https://owner.github.io/desktop-app-pro/",
		"cd desktop-app-template-go-wails":                                     "cd desktop-app-pro",
		"Name=Desktop App\nExec=desktop-app\n":                                 "Name=Desktop App Pro\nExec=desktop-app-pro\n",
		"DESKTOP_APP_HOME":                                                     "DESKTOP_APP_PRO_HOME",
		"io.github.yarkovaleksei.desktopapp":                                   "io.github.owner.desktopapppro",
		`companyName: "yarkovaleksei"`:                                         `companyName: "owner"`,
		"Copyright (c) 2026 yarkovaleksei":                                     "Copyright (c) 2031 owner",
		`maintainer: "yarkovaleksei <yarkovaleksei@users.noreply.github.com>"`: `maintainer: "owner <owner@users.noreply.github.com>"`,
		"github.com/yarkovaleksei/go-desktop-kit/updater":                      "github.com/yarkovaleksei/go-desktop-kit/updater",
		"[ai-memory](https://github.com/yarkovaleksei/ai-memory)":              "[ai-memory](https://github.com/yarkovaleksei/ai-memory)",
		".agents/skills/app-code-map/SKILL.md":                                 ".agents/skills/desktop-app-pro-code-map/SKILL.md",
		"скилл `app-ui`, не happ-ui и не app-uix":                              "скилл `desktop-app-pro-ui`, не happ-ui и не app-uix",
		"name: app-build":                                                      "name: desktop-app-pro-build",
	} {
		if got := replaceAll(in, id.Rules()); got != want {
			t.Errorf("replaceAll(%q)\n got %q\nwant %q", in, got, want)
		}
	}
}

// TestSortYarnLockKeepsTemplate — на lock-файле шаблона, который записал сам
// yarn, сортировка ничего не меняет: порядок совпадает с порядком yarn.
func TestSortYarnLockKeepsTemplate(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(moduleRootDir(t), filepath.FromSlash(yarnLock)))
	if err != nil {
		t.Fatal(err)
	}
	got, err := sortYarnLock(string(data))
	if err != nil {
		t.Fatal(err)
	}
	if got != string(data) {
		t.Error("сортировка изменила lock-файл шаблона: порядок не совпадает с порядком yarn")
	}
}

func TestSortYarnLock(t *testing.T) {
	in := "# yarn\n\n__metadata:\n  version: 10\n\n\"zeta@npm:1\":\n  version: 1\n\n\"photo-viewer@workspace:.\":\n  version: 0\n\n\"@a/b@npm:1, @a/b@npm:^1\":\n  version: 1\n"
	want := "# yarn\n\n__metadata:\n  version: 10\n\n\"@a/b@npm:1, @a/b@npm:^1\":\n  version: 1\n\n\"photo-viewer@workspace:.\":\n  version: 0\n\n\"zeta@npm:1\":\n  version: 1\n"
	got, err := sortYarnLock(in)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("sortYarnLock:\n%s\nwant:\n%s", got, want)
	}
}

// moduleRootDir — корень модуля шаблона: тесты запускаются из scripts/init.
func moduleRootDir(t *testing.T) string {
	t.Helper()
	dir, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return dir
}
