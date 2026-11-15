package archtest

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// errorCodesFile — список кодов ошибок, которые Go отдаёт фронтенду, от корня
// модуля. По нему тест словарей фронтенда проверяет, что у каждого кода есть
// текст errors.<код> на обоих языках.
const errorCodesFile = "frontend/src/i18n/error-codes.json"

// kitCodePackages — пакеты kit, чьи коды ошибок доходят до интерфейса.
// В kit коды — строковые литералы update.*: списка кодов пакет не экспортирует.
var kitCodePackages = []string{
	"github.com/yarkovaleksei/go-desktop-kit/updater",
	"github.com/yarkovaleksei/go-desktop-kit/wailsx",
}

var kitCode = regexp.MustCompile(`^update\.[a-z][a-z-]*$`)

// TestErrorCodesListed сверяет список кодов для словарей фронтенда с кодами
// в Go: константами Code* шаблона и литералами update.* пакетов kit. Новый код
// без строки в списке (и, значит, без текста) или код, которого больше нет,
// роняют тест.
func TestErrorCodesListed(t *testing.T) {
	_, root, _ := strings.Cut(strings.TrimSpace(goCmd(t, "list", "-f", "{{.Module.Path}} {{.Module.Dir}}", ".")), " ")

	found := map[string]string{} // код → где найден
	for _, file := range goFiles(t, root, true) {
		for code, pos := range codeConsts(t, file) {
			found[code] = pos
		}
	}
	dirs := strings.Fields(goCmd(t, append([]string{"list", "-f", "{{.Dir}}"}, kitCodePackages...)...))
	for _, dir := range dirs {
		for _, file := range goFiles(t, dir, false) {
			for code, pos := range kitCodes(t, file) {
				found[code] = pos
			}
		}
	}
	if len(found) < 10 {
		t.Fatalf("найдено %d кодов ошибок: тест ничего не проверяет", len(found))
	}

	data, err := os.ReadFile(filepath.Join(root, errorCodesFile))
	if err != nil {
		t.Fatalf("список кодов: %v", err)
	}
	var listed []string
	if err := json.Unmarshal(data, &listed); err != nil {
		t.Fatalf("разбор %s: %v", errorCodesFile, err)
	}
	if !slices.IsSorted(listed) || len(slices.Compact(slices.Clone(listed))) != len(listed) {
		t.Errorf("%s: коды должны идти по алфавиту без повторов", errorCodesFile)
	}

	for _, code := range slices.Sorted(maps.Keys(found)) {
		if !slices.Contains(listed, code) {
			t.Errorf("код %s (%s) не внесён в %s: добавьте его и тексты errors.%s в словари", code, found[code], errorCodesFile, code)
		}
	}
	for _, code := range listed {
		if _, ok := found[code]; !ok {
			t.Errorf("код %s из %s в Go не найден: уберите его из списка и словарей", code, errorCodesFile)
		}
	}
}

// goFiles возвращает Go-файлы без тестов. recursive — обойти подкаталоги,
// пропуская скрытые, node_modules, testdata и фронтенд.
func goFiles(t *testing.T, dir string, recursive bool) []string {
	t.Helper()
	var files []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := d.Name()
		if d.IsDir() {
			if path == dir {
				return nil
			}
			if !recursive || strings.HasPrefix(name, ".") || name == "node_modules" || name == "testdata" || name == "frontend" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("обход %s: %v", dir, err)
	}
	return files
}

func parseFile(t *testing.T, path string) (*token.FileSet, *ast.File) {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("разбор %s: %v", path, err)
	}
	return fset, f
}

// codeConsts — строковые константы с именем Code*: так шаблон объявляет коды ошибок.
func codeConsts(t *testing.T, path string) map[string]string {
	fset, f := parseFile(t, path)
	out := map[string]string{}
	for _, decl := range f.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.CONST {
			continue
		}
		for _, spec := range gd.Specs {
			vs := spec.(*ast.ValueSpec)
			for i, name := range vs.Names {
				if !strings.HasPrefix(name.Name, "Code") || i >= len(vs.Values) {
					continue
				}
				if v, ok := stringLit(vs.Values[i]); ok {
					out[v] = fset.Position(name.Pos()).String()
				}
			}
		}
	}
	return out
}

// kitCodes — строковые литералы вида update.* в файле kit.
func kitCodes(t *testing.T, path string) map[string]string {
	fset, f := parseFile(t, path)
	out := map[string]string{}
	ast.Inspect(f, func(n ast.Node) bool {
		if v, ok := stringLit(n); ok && kitCode.MatchString(v) {
			out[v] = filepath.Base(path) + ":" + strconv.Itoa(fset.Position(n.Pos()).Line)
		}
		return true
	})
	return out
}

func stringLit(n ast.Node) (string, bool) {
	lit, ok := n.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return "", false
	}
	v, err := strconv.Unquote(lit.Value)
	return v, err == nil
}
