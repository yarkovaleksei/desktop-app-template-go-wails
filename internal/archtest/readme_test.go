package archtest

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/yarkovaleksei/go-desktop-kit/updater"

	"github.com/yarkovaleksei/desktop-app-template-go-wails/internal/appinfo"
)

// readmes — README и его перевод. version — как в таблице установки
// записана версия на языке файла.
var readmes = []struct {
	file    string
	version string
}{
	{"README.md", "<версия>"},
	{"README.en.md", "<version>"},
}

// headingRe — заголовок Markdown вне блока кода: уровень и текст.
var headingRe = regexp.MustCompile(`^(#{1,6})\s+(.+?)\s*#*\s*$`)

// readDocLines читает файл и возвращает строки вне блоков кода с номерами.
func readDocLines(t *testing.T, file string) (lines []string, numbers []int) {
	t.Helper()
	f, err := os.Open(file)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	fence := ""
	sc := bufio.NewScanner(f)
	for n := 1; sc.Scan(); n++ {
		l := sc.Text()
		trim := strings.TrimSpace(l)
		if fence != "" {
			if strings.HasPrefix(trim, fence) {
				fence = ""
			}
			continue
		}
		if strings.HasPrefix(trim, "```") || strings.HasPrefix(trim, "~~~") {
			fence = trim[:3]
			continue
		}
		lines = append(lines, l)
		numbers = append(numbers, n)
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	return lines, numbers
}

// TestReadmeSections требует, чтобы README.md и README.en.md состояли из
// одинаковых разделов: то же число заголовков тех же уровней в том же
// порядке (D-028). Меняешь один — приведи в соответствие второй в том же
// коммите.
func TestReadmeSections(t *testing.T) {
	root := moduleRoot(t)
	var outlines [][]string
	for _, r := range readmes {
		lines, numbers := readDocLines(t, filepath.Join(root, r.file))
		var outline []string
		for i, l := range lines {
			if m := headingRe.FindStringSubmatch(l); m != nil {
				outline = append(outline, fmt.Sprintf("%s (%s:%d %q)", m[1], r.file, numbers[i], m[2]))
			}
		}
		if len(outline) == 0 {
			t.Fatalf("%s: нет заголовков", r.file)
		}
		outlines = append(outlines, outline)
	}
	a, b := outlines[0], outlines[1]
	for i := range max(len(a), len(b)) {
		switch {
		case i >= len(a):
			t.Errorf("лишний раздел %s: в %s его нет", b[i], readmes[0].file)
		case i >= len(b):
			t.Errorf("лишний раздел %s: в %s его нет", a[i], readmes[1].file)
		case level(a[i]) != level(b[i]):
			t.Errorf("разделы не совпадают по уровню: %s и %s", a[i], b[i])
		}
	}
}

func level(heading string) string { return heading[:strings.IndexByte(heading, ' ')] }

// TestReadmeInstallNames сверяет имена пакетов и файлов сумм в обоих README
// с updater.Naming для форматов из RELEASE_FORMATS: каждое имя выпуска
// названо, а имени, которого в выпуске нет (опечатка, удалённый формат), нет.
func TestReadmeInstallNames(t *testing.T) {
	root := moduleRoot(t)
	naming := updater.Naming{App: appinfo.Slug}
	formats := releaseFormats(t, root)
	for _, r := range readmes {
		data, err := os.ReadFile(filepath.Join(root, r.file))
		if err != nil {
			t.Fatal(err)
		}
		text := string(data)
		want := map[string]bool{}
		for _, f := range formats {
			name, ok := naming.Asset(r.version, f)
			if !ok {
				t.Fatalf("updater.Naming не знает формат %q", f)
			}
			want[name] = true
			want[naming.Sums(f)] = true
		}
		for name := range want {
			if !strings.Contains(text, name) {
				t.Errorf("%s: нет файла выпуска %s", r.file, name)
			}
		}
		// Любое имя файла приложения с версией-заполнителем или файл сумм
		// должны быть из состава выпуска.
		found := regexp.MustCompile(regexp.QuoteMeta(appinfo.Slug)+`[-_]`+regexp.QuoteMeta(r.version)+"[^\\s`|]*|SHA256SUMS-[a-z]+\\.txt").FindAllString(text, -1)
		for _, name := range found {
			name = strings.TrimPrefix(name, "./")
			if !want[name] {
				t.Errorf("%s: файла %s нет в выпуске (updater.Naming, RELEASE_FORMATS)", r.file, name)
			}
		}
	}
}

// releaseFormats — форматы выпуска из RELEASE_FORMATS в Taskfile.yml.
func releaseFormats(t *testing.T, root string) []updater.Format {
	t.Helper()
	tf, err := os.ReadFile(filepath.Join(root, "Taskfile.yml"))
	if err != nil {
		t.Fatal(err)
	}
	m := regexp.MustCompile(`(?m)^\s*RELEASE_FORMATS:\s*"([^"]*)"`).FindStringSubmatch(string(tf))
	if m == nil {
		t.Fatal("Taskfile.yml: нет RELEASE_FORMATS")
	}
	var out []updater.Format
	for _, f := range strings.Split(m[1], ",") {
		if !slices.Contains(updater.Formats(), updater.Format(f)) {
			t.Fatalf("Taskfile.yml: неизвестный формат %q", f)
		}
		out = append(out, updater.Format(f))
	}
	return out
}
