package archtest

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/yarkovaleksei/desktop-app-template-go-wails/internal/appinfo"
)

// nameUse — место, где имя из appinfo записано вне Go-кода: файл от корня
// модуля и выражение, первая группа которого — само имя.
type nameUse struct {
	file string
	re   *regexp.Regexp
	want string
	what string
}

// nameUses — все места, где имена приложения повторены вне Go. Инициализация
// проекта меняет их вместе с appinfo; тест ловит пропущенное.
var nameUses = []nameUse{
	{"build/config.yml", regexp.MustCompile(`(?m)^\s*productName:\s*"([^"]*)"`), appinfo.Name, "Name"},
	{"build/config.yml", regexp.MustCompile(`(?m)^\s*productIdentifier:\s*"([^"]*)"`), appinfo.ProductID, "ProductID"},
	{"build/config.yml", regexp.MustCompile(`(?m)^\s*description:\s*"([^"]*)"`), appinfo.Description, "Description"},
	{"Taskfile.yml", regexp.MustCompile(`(?m)^\s*APP_NAME:\s*"([^"]*)"`), appinfo.Slug, "Slug"},
	{"Taskfile.yml", regexp.MustCompile(`(?m)^\s*([A-Z0-9_]+)_HOME:\s`), appinfo.EnvPrefix, "EnvPrefix"},
	{"frontend/index.html", regexp.MustCompile(`<title>([^<]*)</title>`), appinfo.Name, "Name"},
	{"build/linux/nfpm/nfpm.yaml", regexp.MustCompile(`(?m)^name:\s*"([^"]*)"`), appinfo.Slug, "Slug"},
	{"build/linux/nfpm/nfpm.yaml", regexp.MustCompile(`(?m)^description:\s*"([^"]*)"`), appinfo.Description, "Description"},
	{"build/linux/nfpm/nfpm.yaml", regexp.MustCompile(`(?m)^homepage:\s*"https://github\.com/([^"]*)"`), appinfo.Repo, "Repo"},
	{"build/linux/" + appinfo.Slug + ".desktop", regexp.MustCompile(`(?m)^Name=(.+)$`), appinfo.Name, "Name"},
	{"build/linux/" + appinfo.Slug + ".desktop", regexp.MustCompile(`(?m)^Comment=(.+)$`), appinfo.Description, "Description"},
	{"build/linux/" + appinfo.Slug + ".desktop", regexp.MustCompile(`(?m)^Exec=(\S+)`), appinfo.Slug, "Slug"},
	{"build/linux/" + appinfo.Slug + ".desktop", regexp.MustCompile(`(?m)^Icon=(\S+)`), appinfo.Slug, "Slug"},
	{"scripts/smoke-desktop.sh", regexp.MustCompile(`(?m)^slug="([^"]*)"`), appinfo.Slug, "Slug"},
	{"scripts/smoke-desktop.sh", regexp.MustCompile(`(?m)\b([A-Z0-9_]+)_HOME="`), appinfo.EnvPrefix, "EnvPrefix"},
}

// TestNamesMatchAppinfo сверяет имена приложения в файлах сборки, задачах и
// фронтенде с internal/appinfo. Каждое выражение должно найтись хотя бы раз,
// и все найденные значения — совпасть с appinfo.
func TestNamesMatchAppinfo(t *testing.T) {
	root := moduleRoot(t)
	for _, u := range nameUses {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(u.file)))
		if err != nil {
			t.Errorf("%s: %v", u.file, err)
			continue
		}
		matches := u.re.FindAllStringSubmatch(string(data), -1)
		if len(matches) == 0 {
			t.Errorf("%s: не найдено %s (%s) — поправьте файл или выражение теста", u.file, u.what, u.re)
			continue
		}
		for _, m := range matches {
			if m[1] != u.want {
				t.Errorf("%s: %s = %q, а в internal/appinfo — %q", u.file, u.what, m[1], u.want)
			}
		}
	}
}

// moduleRoot — корень модуля шаблона.
func moduleRoot(t *testing.T) string {
	t.Helper()
	_, root, _ := strings.Cut(strings.TrimSpace(goCmd(t, "list", "-f", "{{.Module.Path}} {{.Module.Dir}}", ".")), " ")
	return root
}
