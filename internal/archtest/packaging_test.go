package archtest

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/yarkovaleksei/desktop-app-template-go-wails/internal/appinfo"
	"github.com/yarkovaleksei/go-desktop-kit/updater"
)

// Формат пакета приложение узнаёт двумя путями (updater.Naming.DetectFormat):
// пакеты Linux кладут файл-метку, остальные вшивают формат флагом
// -X main.packageFormat. Тесты ниже сверяют оба пути с kit, чтобы опечатка
// в файлах сборки не оставила установленное приложение без обновлений.

// markerFormats — какой формат должна содержать метка пакета каждого
// упаковщика nfpm.
var markerFormats = map[string]updater.Format{
	"deb":       updater.FormatDeb,
	"rpm":       updater.FormatRPM,
	"archlinux": updater.FormatArch,
}

// TestPackageFormatMarkers: в nfpm.yaml у каждого упаковщика есть своя метка
// по пути updater.Naming.Marker, и в ней записан его формат.
func TestPackageFormatMarkers(t *testing.T) {
	root := moduleRoot(t)
	marker := updater.Naming{App: appinfo.Slug}.Marker()
	data, err := os.ReadFile(filepath.Join(root, "build", "linux", "nfpm", "nfpm.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	src := regexp.MustCompile(`(?m)^\s*src:\s*"([^"]*)"`)
	dst := regexp.MustCompile(`(?m)^\s*dst:\s*"([^"]*)"`)
	packager := regexp.MustCompile(`(?m)^\s*packager:\s*(\S+)`)
	seen := map[string]bool{}
	// Записи contents начинаются с «- src:»; прочие разделы файла их не содержат.
	for _, entry := range regexp.MustCompile(`(?m)^\s*- src:`).Split(string(data), -1)[1:] {
		entry = "src:" + entry
		d := dst.FindStringSubmatch(entry)
		if d == nil || d[1] != marker {
			continue
		}
		s, p := src.FindStringSubmatch(entry), packager.FindStringSubmatch(entry)
		if p == nil {
			t.Errorf("метка %s без packager: один бинарь в трёх пакетах должен получить три разных метки", marker)
			continue
		}
		want, ok := markerFormats[p[1]]
		if !ok {
			t.Errorf("метка для неизвестного упаковщика %q", p[1])
			continue
		}
		seen[p[1]] = true
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(s[1])))
		if err != nil {
			t.Errorf("метка %s: %v", p[1], err)
			continue
		}
		if got := strings.TrimSpace(string(content)); got != string(want) {
			t.Errorf("метка %s (%s) содержит %q, а нужно %q", p[1], s[1], got, want)
		}
	}
	for p := range markerFormats {
		if !seen[p] {
			t.Errorf("в nfpm.yaml нет метки %s для упаковщика %s", marker, p)
		}
	}
}

// builtFormats — форматы, которые вшивают задачи упаковки каждой ОС.
var builtFormats = map[string][]updater.Format{
	"build/linux/Taskfile.yml":   {updater.FormatAppImage},
	"build/darwin/Taskfile.yml":  {updater.FormatDMG},
	"build/windows/Taskfile.yml": {updater.FormatNSIS, updater.FormatExe},
}

// TestBuiltPackageFormats: задачи упаковки вшивают известные kit форматы
// своей ОС, сборка передаёт значение в -X main.packageFormat, а Taskfile.yml
// перечисляет в составе выпуска только известные форматы.
func TestBuiltPackageFormats(t *testing.T) {
	root := moduleRoot(t)
	value := regexp.MustCompile(`(?m)^\s*PACKAGE_FORMAT:\s*"?([a-z]+)"?\s*$`)
	for file, want := range builtFormats {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(file)))
		if err != nil {
			t.Fatal(err)
		}
		var got []updater.Format
		for _, m := range value.FindAllStringSubmatch(string(data), -1) {
			if !slices.Contains(got, updater.Format(m[1])) {
				got = append(got, updater.Format(m[1]))
			}
		}
		slices.Sort(got)
		w := slices.Clone(want)
		slices.Sort(w)
		if !slices.Equal(got, w) {
			t.Errorf("%s: вшиваются форматы %v, а нужно %v", file, got, w)
		}
		if !strings.Contains(string(data), "-X main.packageFormat={{.PACKAGE_FORMAT}}") {
			t.Errorf("%s: сборка не передаёт PACKAGE_FORMAT в -X main.packageFormat", file)
		}
	}
	for _, fs := range builtFormats {
		for _, f := range fs {
			if f.GOOS() == "" {
				t.Errorf("формат %q неизвестен updater", f)
			}
		}
	}

	main, err := os.ReadFile(filepath.Join(root, "main.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !regexp.MustCompile(`(?m)^var packageFormat = ""`).Match(main) {
		t.Error(`main.go: нет var packageFormat = "" — флаг -X main.packageFormat некуда вшить`)
	}

	tf, err := os.ReadFile(filepath.Join(root, "Taskfile.yml"))
	if err != nil {
		t.Fatal(err)
	}
	m := regexp.MustCompile(`(?m)^\s*RELEASE_FORMATS:\s*"([^"]*)"`).FindStringSubmatch(string(tf))
	if m == nil {
		t.Fatal("Taskfile.yml: нет RELEASE_FORMATS")
	}
	for _, f := range strings.Split(m[1], ",") {
		if updater.Format(f).GOOS() == "" {
			t.Errorf("Taskfile.yml: RELEASE_FORMATS содержит неизвестный формат %q", f)
		}
	}
}
