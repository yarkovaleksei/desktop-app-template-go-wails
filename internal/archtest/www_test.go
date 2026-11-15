package archtest

import (
	"encoding/json"
	"html"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/yarkovaleksei/go-desktop-kit/updater"

	"github.com/yarkovaleksei/desktop-app-template-go-wails/internal/appinfo"
)

// Сайт www/index.html живёт без сборки, поэтому его согласованность с kit и с
// самим собой проверяют эти тесты: тексты двух языков (D-028), суффиксы
// вложений выпуска, адреса и размеры картинок.

const siteFile = "www/index.html"

func readSite(t *testing.T) (root, page string) {
	t.Helper()
	root = moduleRoot(t)
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(siteFile)))
	if err != nil {
		t.Fatal(err)
	}
	return root, string(data)
}

// siteTexts — словари из блока <script type="application/json" id="i18n">.
func siteTexts(t *testing.T, page string) map[string]map[string]string {
	t.Helper()
	m := regexp.MustCompile(`(?s)<script type="application/json" id="i18n">(.*?)</script>`).FindStringSubmatch(page)
	if m == nil {
		t.Fatalf("%s: нет блока словарей <script type=\"application/json\" id=\"i18n\">", siteFile)
	}
	var texts map[string]map[string]string
	if err := json.Unmarshal([]byte(m[1]), &texts); err != nil {
		t.Fatalf("%s: словари — не JSON: %v", siteFile, err)
	}
	return texts
}

// siteScript — текст основного скрипта страницы (последний <script> без type).
func siteScript(t *testing.T, page string) string {
	t.Helper()
	all := regexp.MustCompile(`(?s)<script>(.*?)</script>`).FindAllStringSubmatch(page, -1)
	if len(all) == 0 {
		t.Fatalf("%s: нет скрипта", siteFile)
	}
	return all[len(all)-1][1]
}

var placeholderRe = regexp.MustCompile(`\{\w+\}`)

// TestSiteTexts: наборы ключей en и ru совпадают, подстановки {имя} в
// переводе те же; каждый ключ из разметки и скрипта есть в словаре, каждый
// ключ словаря где-то используется; английский в разметке — дословно из
// словаря en, чтобы страница без JavaScript читалась так же.
func TestSiteTexts(t *testing.T) {
	_, page := readSite(t)
	texts := siteTexts(t, page)
	en, ru := texts["en"], texts["ru"]
	if len(texts) != 2 || en == nil || ru == nil {
		t.Fatalf("%s: словари должны быть ровно en и ru, а есть %v", siteFile, slices.Sorted(maps.Keys(texts)))
	}
	for _, pair := range [][2]string{{"en", "ru"}, {"ru", "en"}} {
		a, b := texts[pair[0]], texts[pair[1]]
		for _, k := range slices.Sorted(maps.Keys(a)) {
			if _, ok := b[k]; !ok {
				t.Errorf("ключ %q есть в %s, но нет в %s", k, pair[0], pair[1])
			}
			if strings.TrimSpace(a[k]) == "" {
				t.Errorf("%s: пустой текст ключа %q", pair[0], k)
			}
		}
	}
	for k, v := range en {
		if w, ok := ru[k]; ok {
			got, want := placeholderRe.FindAllString(w, -1), placeholderRe.FindAllString(v, -1)
			slices.Sort(got)
			slices.Sort(want)
			if !slices.Equal(got, want) {
				t.Errorf("ключ %q: подстановки ru %v, en %v", k, got, want)
			}
		}
	}

	// Словари вырезаются из страницы: их строки не считаются использованием.
	markup := regexp.MustCompile(`(?s)<script type="application/json" id="i18n">.*?</script>`).ReplaceAllString(page, "")
	used := map[string]bool{}

	// Элемент с data-i18n содержит только текст, равный en.
	elem := regexp.MustCompile(`(?s)<(\w+)\b([^>]*)\bdata-i18n="([^"]+)"([^>]*)>(.*?)</(\w+)>`)
	for _, m := range elem.FindAllStringSubmatch(markup, -1) {
		key, inner := m[3], m[5]
		used[key] = true
		want, ok := en[key]
		switch {
		case !ok:
			t.Errorf("разметка: ключа data-i18n=%q нет в словаре", key)
		case m[1] != m[6] || strings.Contains(inner, "<"):
			t.Errorf("разметка: элемент с data-i18n=%q должен содержать только текст: скрипт заменит его целиком", key)
		case normSpace(html.UnescapeString(inner)) != want:
			t.Errorf("разметка: data-i18n=%q — %q, а в словаре en — %q", key, normSpace(html.UnescapeString(inner)), want)
		}
	}
	if len(used) == 0 {
		t.Fatalf("%s: не найдено ни одного data-i18n", siteFile)
	}

	// Атрибут с data-i18n-<имя>="ключ" рядом с <имя>="текст из en".
	tag := regexp.MustCompile(`<\w+\b[^>]*\bdata-i18n-[a-z-]+="[^"]*"[^>]*>`)
	attr := regexp.MustCompile(`\bdata-i18n-([a-z-]+)="([^"]+)"`)
	for _, tg := range tag.FindAllString(markup, -1) {
		for _, a := range attr.FindAllStringSubmatch(tg, -1) {
			name, key := a[1], a[2]
			used[key] = true
			want, ok := en[key]
			if !ok {
				t.Errorf("разметка: ключа data-i18n-%s=%q нет в словаре", name, key)
				continue
			}
			v := regexp.MustCompile(`\s` + regexp.QuoteMeta(name) + `="([^"]*)"`).FindStringSubmatch(tg)
			if v == nil {
				t.Errorf("разметка: data-i18n-%s=%q без атрибута %s", name, key, name)
			} else if got := html.UnescapeString(v[1]); got != want {
				t.Errorf("разметка: %s у data-i18n-%s=%q — %q, а в словаре en — %q", name, name, key, got, want)
			}
		}
	}

	for _, m := range regexp.MustCompile(`\bt\('([^']+)'`).FindAllStringSubmatch(siteScript(t, page), -1) {
		used[m[1]] = true
		if _, ok := en[m[1]]; !ok {
			t.Errorf("скрипт: ключа t(%q) нет в словаре", m[1])
		}
	}
	for _, k := range slices.Sorted(maps.Keys(en)) {
		if !used[k] {
			t.Errorf("ключ %q словаря нигде не используется", k)
		}
	}
}

func normSpace(s string) string { return strings.Join(strings.Fields(s), " ") }

// TestSiteAssets: таблица SUFFIXES скрипта — ровно форматы RELEASE_FORMATS
// с окончаниями имён из updater.Naming, и для каждого формата на странице
// есть ссылка data-asset на страницу выпусков.
func TestSiteAssets(t *testing.T) {
	root, page := readSite(t)
	script := siteScript(t, page)

	block := regexp.MustCompile(`(?s)const SUFFIXES = \{(.*?)\};`).FindStringSubmatch(script)
	if block == nil {
		t.Fatalf("%s: нет константы SUFFIXES", siteFile)
	}
	got := map[string]string{}
	for _, m := range regexp.MustCompile(`(?m)^\s*([a-z]+): '([^']*)',?\s*$`).FindAllStringSubmatch(block[1], -1) {
		got[m[1]] = m[2]
	}

	const version = "1.2.3"
	naming := updater.Naming{App: appinfo.Slug}
	want := map[string]string{}
	for _, f := range releaseFormats(t, root) {
		name, _ := naming.Asset(version, f)
		// Имя — слаг, «-» или «_», версия, окончание.
		_, suffix, ok := strings.Cut(name, version)
		if !ok || !strings.HasPrefix(name, appinfo.Slug) {
			t.Fatalf("имя %s формата %s не вида <слаг><-|_><версия><окончание>: поправьте скрипт сайта вместе с тестом", name, f)
		}
		want[string(f)] = suffix
	}
	for _, k := range slices.Sorted(maps.Keys(want)) {
		switch g, ok := got[k]; {
		case !ok:
			t.Errorf("SUFFIXES: нет формата %s (%q)", k, want[k])
		case g != want[k]:
			t.Errorf("SUFFIXES: %s = %q, а по updater.Naming — %q", k, g, want[k])
		}
	}
	for k := range got {
		if _, ok := want[k]; !ok {
			t.Errorf("SUFFIXES: формата %s нет в RELEASE_FORMATS", k)
		}
	}

	releases := "https://github.com/" + appinfo.Repo + "/releases/latest"
	links := regexp.MustCompile(`<a\b[^>]*\bdata-asset="([^"]*)"[^>]*>`).FindAllStringSubmatch(page, -1)
	seen := map[string]bool{}
	for _, l := range links {
		seen[l[1]] = true
		if _, ok := want[l[1]]; !ok {
			t.Errorf("ссылка data-asset=%q: такого формата нет в RELEASE_FORMATS", l[1])
		}
		if !strings.Contains(l[0], `href="`+releases+`"`) {
			t.Errorf("ссылка data-asset=%q должна вести на %s, пока скрипт не подставил вложение", l[1], releases)
		}
	}
	for k := range want {
		if !seen[k] {
			t.Errorf("на странице нет ссылки data-asset=%q", k)
		}
	}
}

// TestSiteMeta: адрес сайта — Pages репозитория appinfo.Repo; картинки, на
// которые ссылается страница, существуют и их размеры совпадают с указанными.
func TestSiteMeta(t *testing.T) {
	root, page := readSite(t)
	owner, name, _ := strings.Cut(appinfo.Repo, "/")
	site := "https://" + strings.ToLower(owner) + ".github.io/" + name + "/"
	for _, re := range []string{
		`<link rel="canonical" href="([^"]*)"`,
		`<meta property="og:url" content="([^"]*)"`,
	} {
		m := regexp.MustCompile(re).FindStringSubmatch(page)
		if m == nil || m[1] != site {
			t.Errorf("%s: %s — ожидался адрес %s", siteFile, re, site)
		}
	}

	img := regexp.MustCompile(`<meta property="og:image" content="([^"]*)"`).FindStringSubmatch(page)
	if img == nil || !strings.HasPrefix(img[1], site) {
		t.Fatalf("%s: og:image должен лежать на %s", siteFile, site)
	}
	ogW := metaInt(t, page, "og:image:width")
	ogH := metaInt(t, page, "og:image:height")
	checkImage(t, root, strings.TrimPrefix(img[1], site), ogW, ogH)

	for _, m := range regexp.MustCompile(`<img\b[^>]*>`).FindAllString(page, -1) {
		src := regexp.MustCompile(`\ssrc="([^"]*)"`).FindStringSubmatch(m)
		w := regexp.MustCompile(`\swidth="(\d+)"`).FindStringSubmatch(m)
		h := regexp.MustCompile(`\sheight="(\d+)"`).FindStringSubmatch(m)
		if src == nil || w == nil || h == nil {
			t.Errorf("%s: у %s нет src, width или height", siteFile, m)
			continue
		}
		wi, _ := strconv.Atoi(w[1])
		hi, _ := strconv.Atoi(h[1])
		// Иконка показывается уменьшенной: важны пропорции, а не размер.
		checkImageRatio(t, root, src[1], wi, hi)
	}
	for _, m := range regexp.MustCompile(`<link rel="icon"[^>]*href="([^"]*)"`).FindAllStringSubmatch(page, -1) {
		if _, err := os.Stat(filepath.Join(root, "www", filepath.FromSlash(m[1]))); err != nil {
			t.Errorf("%s: значок вкладки: %v", siteFile, err)
		}
	}
}

func metaInt(t *testing.T, page, property string) int {
	t.Helper()
	m := regexp.MustCompile(`<meta property="` + regexp.QuoteMeta(property) + `" content="(\d+)"`).FindStringSubmatch(page)
	if m == nil {
		t.Fatalf("%s: нет %s", siteFile, property)
	}
	v, _ := strconv.Atoi(m[1])
	return v
}

func imageSize(t *testing.T, root, rel string) (int, int, bool) {
	t.Helper()
	f, err := os.Open(filepath.Join(root, "www", filepath.FromSlash(rel)))
	if err != nil {
		t.Errorf("%s: картинка %s: %v", siteFile, rel, err)
		return 0, 0, false
	}
	defer f.Close()
	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		t.Errorf("%s: картинка %s: %v", siteFile, rel, err)
		return 0, 0, false
	}
	return cfg.Width, cfg.Height, true
}

func checkImage(t *testing.T, root, rel string, w, h int) {
	t.Helper()
	if gw, gh, ok := imageSize(t, root, rel); ok && (gw != w || gh != h) {
		t.Errorf("%s: %s — %d×%d, а на странице указано %d×%d", siteFile, rel, gw, gh, w, h)
	}
}

func checkImageRatio(t *testing.T, root, rel string, w, h int) {
	t.Helper()
	if gw, gh, ok := imageSize(t, root, rel); ok && gw*h != gh*w {
		t.Errorf("%s: %s — %d×%d, пропорции не те, что на странице (%d×%d)", siteFile, rel, gw, gh, w, h)
	}
}
