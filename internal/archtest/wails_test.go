package archtest

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// wailsModule — модуль Wails. Пакет считается зависящим от Wails, если тот
// есть среди его зависимостей, прямых или через kit/wailsx.
const wailsModule = "github.com/wailsapp/wails/v3"

// wailsAllowed — пакеты модуля, которым Wails разрешён, относительно его
// корня: "" — сам main.
var wailsAllowed = []string{"", "internal/app"}

type goPackage struct {
	ImportPath string
	Deps       []string
}

// modulePackages возвращает пакеты модуля шаблона и путь модуля.
func modulePackages(t *testing.T) (string, []goPackage) {
	t.Helper()
	module := strings.TrimSpace(goCmd(t, "list", "-m"))
	trackSources(t, strings.TrimSpace(goCmd(t, "list", "-m", "-f", "{{.Dir}}")))
	out := goCmd(t, "list", "-e", "-json=ImportPath,Deps", module+"/...")
	var pkgs []goPackage
	dec := json.NewDecoder(strings.NewReader(out))
	for {
		var p goPackage
		err := dec.Decode(&p)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatalf("разбор go list: %v", err)
		}
		pkgs = append(pkgs, p)
	}
	if len(pkgs) < 2 {
		t.Fatalf("go list нашёл %d пакетов: тест ничего не проверяет", len(pkgs))
	}
	return module, pkgs
}

// trackSources читает сведения о Go-файлах модуля. go list идёт отдельным
// процессом, и кэш go test не знает, что результат зависит от исходников:
// без этого тест оставался бы зелёным из кэша после запрещённого импорта.
// Файлы, которые тест открыл или проверил через os.Stat, входят в ключ кэша.
func trackSources(t *testing.T, root string) {
	t.Helper()
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := d.Name()
		if d.IsDir() {
			if path != root && (strings.HasPrefix(name, ".") || name == "node_modules" || name == "testdata") {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(name, ".go") || name == "go.mod" || name == "go.sum" {
			_, err = os.Stat(path)
		}
		return err
	})
	if err != nil {
		t.Fatalf("обход исходников: %v", err)
	}
}

func goCmd(t *testing.T, args ...string) string {
	t.Helper()
	cmd := exec.Command("go", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("go %s: %v\n%s", strings.Join(args, " "), err, stderr.String())
	}
	return string(out)
}

func TestWailsOnlyInMainAndApp(t *testing.T) {
	module, pkgs := modulePackages(t)
	for _, p := range pkgs {
		rel := strings.TrimPrefix(strings.TrimPrefix(p.ImportPath, module), "/")
		if slices.Contains(wailsAllowed, rel) {
			continue
		}
		for _, d := range p.Deps {
			if d == wailsModule || strings.HasPrefix(d, wailsModule+"/") {
				t.Errorf("%s зависит от Wails (%s): Wails разрешён только main и internal/app", p.ImportPath, d)
				break
			}
		}
	}
}
