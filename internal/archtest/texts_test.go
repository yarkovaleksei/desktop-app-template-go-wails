package archtest

import (
	"go/ast"
	"go/token"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/yarkovaleksei/desktop-app-template-go-wails/internal/texts"
)

const textsPath = "github.com/yarkovaleksei/desktop-app-template-go-wails/internal/texts"

// TestTextKeysExist находит в Go-коде модуля вызовы texts.T и проверяет, что
// у каждого ключа есть текст на всех языках. Ключ задаётся константой пакета
// texts: строковый литерал или выражение тест не разберёт и сочтёт ошибкой.
// texts.T возвращает сам ключ, если текста нет, — так пропуск и виден.
func TestTextKeysExist(t *testing.T) {
	root := moduleRoot(t)
	keys := textKeyConsts(t, filepath.Join(root, "internal", "texts"))

	used := 0
	for _, file := range goFiles(t, root, true) {
		fset, f := parseFile(t, file)
		alias := importName(f, textsPath)
		if alias == "" {
			continue
		}
		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok || !isSelector(call.Fun, alias, "T") || len(call.Args) < 2 {
				return true
			}
			pos := fset.Position(call.Pos())
			if rel, err := filepath.Rel(root, pos.Filename); err == nil {
				pos.Filename = filepath.ToSlash(rel)
			}
			sel, ok := call.Args[1].(*ast.SelectorExpr)
			if !ok || !isSelector(sel, alias, sel.Sel.Name) {
				t.Errorf("%s: ключ texts.T — не константа пакета texts", pos)
				return true
			}
			key, ok := keys[sel.Sel.Name]
			if !ok {
				t.Errorf("%s: texts.%s — не строковая константа пакета texts", pos, sel.Sel.Name)
				return true
			}
			used++
			for _, lang := range texts.Languages() {
				if texts.T(lang, key, nil) == key {
					t.Errorf("%s: нет текста %q (texts.%s) на языке %s", pos, key, sel.Sel.Name, lang)
				}
			}
			return true
		})
	}
	if used < 5 {
		t.Fatalf("найдено %d вызовов texts.T: тест ничего не проверяет", used)
	}
}

// textKeyConsts — строковые константы пакета texts: имя → значение.
func textKeyConsts(t *testing.T, dir string) map[string]string {
	out := map[string]string{}
	for _, file := range goFiles(t, dir, false) {
		_, f := parseFile(t, file)
		for _, decl := range f.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.CONST {
				continue
			}
			for _, spec := range gd.Specs {
				vs := spec.(*ast.ValueSpec)
				for i, name := range vs.Names {
					if i < len(vs.Values) {
						if v, ok := stringLit(vs.Values[i]); ok {
							out[name.Name] = v
						}
					}
				}
			}
		}
	}
	return out
}

// importName — имя, под которым файл импортирует пакет path, или "".
func importName(f *ast.File, path string) string {
	for _, imp := range f.Imports {
		if p, _ := strconv.Unquote(imp.Path.Value); p == path {
			if imp.Name != nil {
				return imp.Name.Name
			}
			return filepath.Base(path)
		}
	}
	return ""
}

func isSelector(e ast.Expr, pkg, name string) bool {
	sel, ok := e.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	id, ok := sel.X.(*ast.Ident)
	return ok && id.Name == pkg && sel.Sel.Name == name
}
