package main

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// Имена шаблона — значения-заглушки, которые инициализация заменяет. Те же
// значения записаны в internal/appinfo; совпадение проверяет тест.
const (
	templateName        = "Desktop App"
	templateSlug        = "desktop-app"
	templateModule      = "github.com/yarkovaleksei/desktop-app-template-go-wails"
	templateRepo        = "yarkovaleksei/desktop-app-template-go-wails"
	templateID          = "io.github.yarkovaleksei.desktopapp"
	templateEnvPrefix   = "DESKTOP_APP"
	templateSkillPrefix = "app-"
	templateDescription = "Desktop application on Go, Wails 3, React and MUI"
	templateAuthor      = "yarkovaleksei"
	templateYear        = "2026"
)

// templateSkills — скиллы шаблона с префиксом templateSkillPrefix: каталоги
// .agents/skills/<имя> и симлинки .claude/skills/<имя>. Список сверяет тест.
var templateSkills = []string{
	"app-build", "app-ci", "app-code-map", "app-code-style",
	"app-docs", "app-release", "app-spec", "app-ui",
}

// Flags — значения нового проекта из командной строки. Пустое поле, кроме
// Name, Repo и Description, выводится из остальных (Derive).
type Flags struct {
	Name        string
	Slug        string
	Module      string
	Repo        string
	ID          string
	Description string
	Author      string
	Year        int
	Color       string
}

// Identity — все варианты записи имён нового проекта.
type Identity struct {
	Name        string // название продукта: «My App»
	Slug        string // бинарь, каталоги, пакеты: my-app
	Module      string // модуль Go: github.com/owner/my-app
	Repo        string // репозиторий GitHub: owner/my-app
	Owner       string // владелец репозитория: owner
	RepoName    string // имя репозитория: my-app
	PagesHost   string // хост GitHub Pages: owner.github.io
	ID          string // идентификатор продукта: io.github.owner.myapp
	EnvPrefix   string // префикс окружения: MY_APP
	SkillPrefix string // префикс скиллов: my-app-
	Description string // описание в одну строку
	Author      string // правообладатель
	Year        string // год в строке правообладателя
	Color       string // цвет плашки иконки #RRGGBB; пусто — из слага
}

var (
	slugRe   = regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`)
	repoRe   = regexp.MustCompile(`^[A-Za-z0-9](?:[A-Za-z0-9-]*[A-Za-z0-9])?/[A-Za-z0-9._-]+$`)
	moduleRe = regexp.MustCompile(`^[a-z0-9-]+(\.[a-z0-9-]+)*\.[a-z]{2,}(/[A-Za-z0-9._~-]+)+$`)
	idRe     = regexp.MustCompile(`^[a-z]{2,}(\.[a-z0-9]+){2,}$`)
	colorRe  = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)
)

// Derive проверяет флаги и выводит из них все имена проекта. Ошибка называет
// флаг и допустимый вид значения; ошибки всех флагов собираются вместе.
func Derive(f Flags, year int) (Identity, error) {
	var errs []error
	fail := func(format string, args ...any) { errs = append(errs, fmt.Errorf(format, args...)) }

	id := Identity{
		Name:        strings.TrimSpace(f.Name),
		Slug:        strings.TrimSpace(f.Slug),
		Module:      strings.TrimSpace(f.Module),
		Repo:        strings.TrimSpace(f.Repo),
		ID:          strings.TrimSpace(f.ID),
		Description: strings.TrimSpace(f.Description),
		Author:      strings.TrimSpace(f.Author),
		Color:       strings.TrimSpace(f.Color),
	}

	switch {
	case id.Name == "":
		fail("-name: задайте название продукта, например \"My App\"")
	case !validName(id.Name):
		fail("-name %q: название — буквы, цифры, пробел и знаки . _ + -, начинается с буквы или цифры, не длиннее 64 символов", id.Name)
	}

	if id.Slug == "" && id.Name != "" {
		id.Slug = slugify(id.Name)
		if id.Slug == "" {
			fail("-slug: из названия %q слаг не выводится (нужны латинские буквы) — задайте его, например my-app", id.Name)
		}
	}
	if id.Slug != "" && (!slugRe.MatchString(id.Slug) || len(id.Slug) < 2 || len(id.Slug) > 50) {
		fail("-slug %q: слаг — строчные латинские буквы и цифры, начинается с буквы, части через один дефис, 2–50 символов, например my-app", id.Slug)
	}

	switch {
	case id.Repo == "":
		fail("-repo: задайте репозиторий GitHub в виде владелец/имя, например owner/my-app")
	case !repoRe.MatchString(id.Repo):
		fail("-repo %q: репозиторий — владелец/имя, например owner/my-app", id.Repo)
	default:
		id.Owner, id.RepoName, _ = strings.Cut(id.Repo, "/")
		id.PagesHost = strings.ToLower(id.Owner) + ".github.io"
	}

	if id.Module == "" && id.Repo != "" {
		id.Module = "github.com/" + id.Repo
	}
	if id.Module != "" && !moduleRe.MatchString(id.Module) {
		fail("-module %q: путь модуля начинается с домена и содержит путь, например github.com/owner/my-app", id.Module)
	}

	if id.ID == "" && id.Owner != "" && id.Slug != "" {
		id.ID = "io.github." + idPart(id.Owner) + "." + idPart(id.Slug)
	}
	if id.ID != "" && (!idRe.MatchString(id.ID) || !isTopLevel(id.ID[:strings.IndexByte(id.ID, '.')])) {
		fail("-id %q: идентификатор — обратная доменная запись из трёх и более частей: строчные латинские буквы и цифры через точку, первая часть — домен верхнего уровня, например io.github.owner.myapp", id.ID)
	}

	switch {
	case id.Description == "":
		fail("-description: задайте описание продукта в одну строку")
	case !validText(id.Description, 200):
		fail("-description %q: описание — одна строка до 200 символов без кавычек, обратной косой черты и знаков < > & $ `", id.Description)
	}

	if id.Author == "" {
		id.Author = id.Owner
	}
	if id.Author != "" && !validText(id.Author, 100) {
		fail("-author %q: правообладатель — одна строка до 100 символов без кавычек, обратной косой черты и знаков < > & $ `", id.Author)
	}

	if f.Year != 0 {
		year = f.Year
	}
	if year < 2000 || year > 9999 {
		fail("-year %d: год — четыре цифры, не раньше 2000", year)
	}
	id.Year = strconv.Itoa(year)

	if id.Color != "" && !colorRe.MatchString(id.Color) {
		fail("-color %q: цвет — #RRGGBB, например #2b6cb0", id.Color)
	}

	id.EnvPrefix = strings.ToUpper(strings.ReplaceAll(id.Slug, "-", "_"))
	id.SkillPrefix = id.Slug + "-"

	for _, same := range []struct{ flag, got, tmpl string }{
		{"-name", id.Name, templateName},
		{"-slug", id.Slug, templateSlug},
		{"-module", id.Module, templateModule},
		{"-repo", id.Repo, templateRepo},
		{"-id", id.ID, templateID},
	} {
		if same.got != "" && same.got == same.tmpl {
			fail("%s %q: это имя шаблона — задайте имя нового проекта", same.flag, same.got)
		}
	}

	if len(errs) > 0 {
		return Identity{}, errors.Join(errs...)
	}
	return id, nil
}

// isTopLevel — похожа ли часть на домен верхнего уровня: двухбуквенный код
// страны или распространённый общий домен. Так ловится прямая запись
// (myapp.owner.github.io) вместо обратной.
func isTopLevel(s string) bool {
	switch s {
	case "com", "org", "net", "io", "dev", "app", "info", "biz", "xyz", "tech", "pro", "name", "edu", "gov":
		return true
	}
	return len(s) == 2
}

// validName — название: буквы и цифры любых алфавитов, пробел и . _ + -.
func validName(s string) bool {
	if len([]rune(s)) > 64 {
		return false
	}
	for i, r := range s {
		ok := unicode.IsLetter(r) || unicode.IsDigit(r)
		if i > 0 && strings.ContainsRune(" ._+-", r) {
			ok = true
		}
		if !ok {
			return false
		}
	}
	return !strings.Contains(s, "  ")
}

// validText — строка, которую можно вставить в YAML, JSON, XML, NSIS, shell
// и JavaScript без экранирования.
func validText(s string, maxLen int) bool {
	if len([]rune(s)) > maxLen {
		return false
	}
	for _, r := range s {
		if unicode.IsControl(r) || strings.ContainsRune("\"'\\<>&$`", r) {
			return false
		}
	}
	return true
}

// slugify выводит слаг из названия: латинские буквы и цифры в нижнем
// регистре, остальное — дефис между частями.
func slugify(name string) string {
	var parts []string
	var cur strings.Builder
	flush := func() {
		if cur.Len() > 0 {
			parts = append(parts, cur.String())
			cur.Reset()
		}
	}
	for _, r := range strings.ToLower(name) {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			cur.WriteRune(r)
		} else {
			flush()
		}
	}
	flush()
	s := strings.Join(parts, "-")
	return strings.TrimLeft(s, "0123456789-")
}

// idPart — часть идентификатора продукта: строчные латинские буквы и цифры.
func idPart(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// rule — замена записи имени шаблона на запись имени проекта. word требует,
// чтобы вокруг найденного не было букв, цифр, дефиса и подчёркивания.
type rule struct {
	old, new string
	word     bool
}

// Rules — таблица замен в порядке приоритета: в каждой позиции текста
// срабатывает первое подходящее правило, поэтому длинные записи стоят
// раньше коротких, которые в них входят. Правообладатель заменяется только
// в строках правообладателя и компании: остальные github.com/yarkovaleksei/…
// — ссылки на go-desktop-kit и проекты-источники.
func (id Identity) Rules() []rule {
	rs := []rule{
		{old: templateModule, new: id.Module},
		{old: templateRepo, new: id.Repo},
		{old: "yarkovaleksei.github.io", new: id.PagesHost},
		{old: "desktop-app-template-go-wails", new: id.RepoName},
		{old: templateID, new: id.ID},
		{old: templateDescription, new: id.Description},
		{old: templateAuthor + " <" + templateAuthor + "@users.noreply.github.com>", new: id.Author + " <" + id.Owner + "@users.noreply.github.com>"},
		{old: "Copyright (c) " + templateYear + " " + templateAuthor, new: "Copyright (c) " + id.Year + " " + id.Author},
		{old: "© " + templateYear + " " + templateAuthor, new: "© " + id.Year + " " + id.Author},
		{old: "(c) " + templateYear + " " + templateAuthor, new: "(c) " + id.Year + " " + id.Author},
		{old: `"` + templateAuthor + `"`, new: `"` + id.Author + `"`},
		{old: templateName, new: id.Name},
		{old: templateEnvPrefix, new: id.EnvPrefix},
		{old: templateSlug, new: id.Slug},
	}
	for _, s := range templateSkills {
		rs = append(rs, rule{old: s, new: id.SkillPrefix + strings.TrimPrefix(s, templateSkillPrefix), word: true})
	}
	return rs
}

// replaceAll применяет правила за один проход: заменённый текст повторно не
// просматривается, поэтому имя проекта может содержать имя шаблона.
func replaceAll(s string, rules []rule) string {
	var b strings.Builder
	changed := false
	for i := 0; i < len(s); {
		if r, ok := matchAt(s, i, rules); ok {
			if !changed {
				b.Grow(len(s))
				b.WriteString(s[:i])
				changed = true
			}
			b.WriteString(r.new)
			i += len(r.old)
			continue
		}
		if changed {
			b.WriteByte(s[i])
		}
		i++
	}
	if !changed {
		return s
	}
	return b.String()
}

func matchAt(s string, i int, rules []rule) (rule, bool) {
	for _, r := range rules {
		if !strings.HasPrefix(s[i:], r.old) {
			continue
		}
		if r.word {
			end := i + len(r.old)
			if i > 0 && isNameByte(s[i-1]) || end < len(s) && isNameByte(s[end]) {
				continue
			}
		}
		return r, true
	}
	return rule{}, false
}

func isNameByte(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-' || c == '_'
}
