package repodocs

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	kitdocs "github.com/yarkovaleksei/go-desktop-kit/repodocs"
)

// update перезаписывает карту кода: `wails3 task docs:map`.
var update = flag.Bool("update", false, "перезаписать docs/code-map.md")

// root — корень репозитория относительно каталога пакета.
const root = "../.."

func config() kitdocs.Config {
	return kitdocs.Config{
		Root:       root,
		Areas:      Areas,
		MapCommand: "wails3 task docs:map",
		PagesDir:   "frontend/src/pages",
		// Результаты сборки и рабочие каталоги задач: пакетов и документов
		// репозитория в них нет.
		SkipDirs: []string{"bin", "dist", ".task", "frontend/dist", "build/linux/appimage/build",
			"frontend/coverage", "frontend/playwright-report", "frontend/test-results"},
	}
}

// TestCodeMap требует, чтобы карта кода покрывала все пакеты и страницы и
// совпадала с закоммиченной. Падение после добавления пакета или страницы —
// сигнал внести их в Areas, написать комментарий и выполнить
// `wails3 task docs:map`.
func TestCodeMap(t *testing.T) {
	c := config()
	if *update {
		text, err := kitdocs.RenderCodeMap(c)
		if err != nil {
			t.Fatalf("карта кода не собирается:\n%v", err)
		}
		file := filepath.Join(root, filepath.FromSlash(kitdocs.DefaultCodeMap))
		if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	for _, p := range kitdocs.CheckCodeMap(c) {
		t.Error(p)
	}
}

// TestDocs не даёт документам сослаться на то, чего в репозитории нет:
// файлы, якоря, пути, задачи Taskfile и скрипты package.json.
func TestDocs(t *testing.T) {
	for _, p := range kitdocs.CheckDocs(config()) {
		t.Error(p)
	}
}
