package main

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// testArgs — флаги пробного проекта. Имена без имён шаблона внутри: иначе
// поиск остатков шаблона не отличит их от имён проекта (вложенные имена
// проверяет TestReplaceAll).
var testArgs = []string{
	"-name", "Photo Viewer",
	"-slug", "photo-viewer",
	"-repo", "acme-org/photo-viewer",
	"-description", "View photos on the desktop",
	"-author", "ACME Org",
	"-year", "2031",
}

const testSlug = "photo-viewer"

// copyTemplate копирует файлы шаблона (отслеживаемые и новые неигнорируемые)
// во временный каталог с новым репозиторием git и одним коммитом.
func copyTemplate(t *testing.T) string {
	t.Helper()
	src := moduleRootDir(t)
	out, err := git(src, "ls-files", "-z", "--cached", "--others", "--exclude-standard")
	if err != nil {
		t.Fatal(err)
	}
	dst := t.TempDir()
	for _, p := range strings.Split(strings.TrimRight(out, "\x00"), "\x00") {
		from := filepath.Join(src, filepath.FromSlash(p))
		to := filepath.Join(dst, filepath.FromSlash(p))
		info, err := os.Lstat(from)
		if errors.Is(err, fs.ErrNotExist) {
			continue // удалён в рабочем дереве, но ещё не из индекса
		}
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
			t.Fatal(err)
		}
		if info.Mode()&fs.ModeSymlink != 0 {
			target, err := os.Readlink(from)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(target, to); err != nil {
				t.Fatal(err)
			}
			continue
		}
		data, err := os.ReadFile(from)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(to, data, info.Mode().Perm()); err != nil {
			t.Fatal(err)
		}
	}
	commitAll(t, dst, "Шаблон")
	return dst
}

func commitAll(t *testing.T, dir, msg string) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(dir, ".git")); err != nil {
		mustGit(t, dir, "init", "-q")
	}
	mustGit(t, dir, "add", "-A")
	mustGit(t, dir, "-c", "user.name=test", "-c", "user.email=test@example.com", "-c", "commit.gpgsign=false", "commit", "-q", "-m", msg)
}

func mustGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, err := git(dir, args...)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func goRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("go", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}

func TestRefuseDirty(t *testing.T) {
	if testing.Short() {
		t.Skip("копия шаблона")
	}
	dir := copyTemplate(t)
	for name, spoil := range map[string]func() error{
		"изменённый файл": func() error {
			return os.WriteFile(filepath.Join(dir, "README.md"), []byte("# правка\n"), 0o644)
		},
		"новый файл": func() error {
			return os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("заметка\n"), 0o644)
		},
	} {
		t.Run(name, func(t *testing.T) {
			if err := spoil(); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				mustGit(t, dir, "checkout", "-q", ".")
				mustGit(t, dir, "clean", "-qfd")
			})
			var out bytes.Buffer
			err := run(dir, testArgs, &out, 2031)
			if !errors.Is(err, errDirty) {
				t.Fatalf("ожидался отказ на незакоммиченных изменениях, получено %v", err)
			}
			if status := mustGit(t, dir, "status", "--porcelain"); strings.Count(status, "\n") != 1 {
				t.Errorf("инструмент изменил дерево при отказе:\n%s", status)
			}
		})
	}
}

func TestRefuseInitialized(t *testing.T) {
	if testing.Short() {
		t.Skip("копия шаблона")
	}
	dir := copyTemplate(t)
	mod := filepath.Join(dir, "go.mod")
	data, err := os.ReadFile(mod)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(mod, bytes.Replace(data, []byte(templateModule), []byte("github.com/acme-org/photo-viewer"), 1), 0o644); err != nil {
		t.Fatal(err)
	}
	commitAll(t, dir, "Проект")
	err = run(dir, testArgs, &bytes.Buffer{}, 2031)
	if !errors.Is(err, errInitialized) {
		t.Fatalf("ожидался отказ на дереве без имён шаблона, получено %v", err)
	}
}

// TestDryRun — пробный запуск печатает изменения и ничего не меняет.
func TestDryRun(t *testing.T) {
	if testing.Short() {
		t.Skip("копия шаблона")
	}
	dir := copyTemplate(t)
	var out bytes.Buffer
	if err := run(dir, append(testArgs, "-dry-run"), &out, 2031); err != nil {
		t.Fatalf("%v\n%s", err, out.String())
	}
	for _, want := range []string{
		"переименовать  build/linux/desktop-app.desktop → build/linux/photo-viewer.desktop",
		"переименовать  .claude/skills/app-ui → .claude/skills/photo-viewer-ui",
		"удалить        scripts/init/main.go",
		"изменить       go.mod",
		"go run ./scripts/icon -slug photo-viewer",
		"ничего не изменено",
	} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("в выводе нет %q:\n%s", want, out.String())
		}
	}
	if status := mustGit(t, dir, "status", "--porcelain"); status != "" {
		t.Errorf("после -dry-run дерево изменено:\n%s", status)
	}
}

// TestInit инициализирует копию шаблона и проверяет, что получился
// собираемый проект без имён шаблона.
func TestInit(t *testing.T) {
	if testing.Short() {
		t.Skip("копия шаблона, сборка и тесты в ней")
	}
	dir := copyTemplate(t)
	templateIcon, err := os.ReadFile(filepath.Join(dir, "build", "appicon.png"))
	if err != nil {
		t.Fatal(err)
	}
	// Остатки плана, которых нет в начальном состоянии spec/.
	for _, p := range []string{"spec/plan/stages/01-old.md", "spec/archive/old/01-old.md"} {
		if err := os.MkdirAll(filepath.Join(dir, filepath.Dir(p)), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, p), []byte("# Этап\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	commitAll(t, dir, "План")

	var out bytes.Buffer
	if err := run(dir, testArgs, &out, 2031); err != nil {
		t.Fatalf("%v\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), "wails3 task docs:map") {
		t.Errorf("в выводе нет следующих шагов:\n%s", out.String())
	}

	// Служебного шаблона нет, план — в начальном состоянии.
	for _, p := range []string{"scripts/init", ".agents/skills/template-init", ".claude/skills/template-init",
		"spec/plan/stages/01-old.md", "spec/archive/old"} {
		if _, err := os.Lstat(filepath.Join(dir, filepath.FromSlash(p))); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("%s не удалён", p)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "spec", "plan", "stages", ".gitkeep")); err != nil {
		t.Errorf("spec/plan/stages/.gitkeep: %v", err)
	}
	fileContains(t, dir, "spec/plan/README.md", "**Цель:** нет.")
	fileContains(t, dir, "spec/plan/JOURNAL.md", "Записей нет.")
	fileContains(t, dir, "spec/docs/project.md", "Photo Viewer — View photos on the desktop.")
	fileContains(t, dir, ".github/workflows/pages.yml", "if: github.repository == 'acme-org/photo-viewer'\n")
	fileContains(t, dir, "LICENSE", "Copyright (c) 2031 ACME Org\n")
	fileContains(t, dir, "README.md", "# Photo Viewer\n")
	fileContains(t, dir, "README.en.md", "git clone https://github.com/acme-org/photo-viewer.git\ncd photo-viewer\n")

	// Скиллы переименованы, симлинки ведут на них.
	for _, s := range templateSkills {
		name := testSlug + "-" + strings.TrimPrefix(s, templateSkillPrefix)
		if _, err := os.Stat(filepath.Join(dir, ".agents", "skills", name, "SKILL.md")); err != nil {
			t.Errorf("скилл %s: %v", name, err)
		}
		link := filepath.Join(dir, ".claude", "skills", name)
		target, err := os.Readlink(link)
		if err != nil {
			// Без симлинков (Windows) git хранит цель в файле.
			data, rerr := os.ReadFile(link)
			if rerr != nil {
				t.Errorf("%s: %v", link, err)
				continue
			}
			target = string(data)
		}
		if want := "../../.agents/skills/" + name; filepath.ToSlash(target) != want {
			t.Errorf(".claude/skills/%s ведёт на %q, а не на %q", name, target, want)
		}
	}

	// Иконка перерисована в цвет нового слага.
	icon, err := os.ReadFile(filepath.Join(dir, "build", "appicon.png"))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(icon, templateIcon) {
		t.Error("build/appicon.png не перерисована: цвет плашки — шаблонный")
	}
	fresh := t.TempDir()
	goRun(t, dir, "run", "./scripts/icon", "-slug", testSlug, "-out", fresh)
	if want, _ := os.ReadFile(filepath.Join(fresh, "appicon.png")); !bytes.Equal(icon, want) {
		t.Error("build/appicon.png не совпадает с иконкой для слага " + testSlug)
	}

	// Проект собирается, правила репозитория выполняются.
	goRun(t, dir, "build", "./internal/...")
	goRun(t, dir, "vet", "./internal/...")
	goRun(t, dir, "test", "-count=1", "./internal/repodocs", "-run", "TestCodeMap", "-update")
	goRun(t, dir, "test", "-count=1", "./internal/archtest", "./internal/repodocs")

	checkNoTemplateNames(t, dir)

	// Индекс совпадает с деревом: удалённых, но отслеживаемых файлов нет.
	if deleted := mustGit(t, dir, "ls-files", "--deleted"); deleted != "" {
		t.Errorf("в индексе остались удалённые файлы:\n%s", deleted)
	}

	// Повторный запуск отказывается: имён шаблона нет.
	if err := run(dir, testArgs, &bytes.Buffer{}, 2031); !errors.Is(err, errInitialized) {
		t.Errorf("повторный запуск: ожидался отказ, получено %v", err)
	}
}

func fileContains(t *testing.T, dir, file, want string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(file)))
	if err != nil {
		t.Error(err)
		return
	}
	if !strings.Contains(string(data), want) {
		t.Errorf("%s: нет %q", file, want)
	}
}

// templateNameRe — записи имён шаблона. Правообладатель шаблона допустим
// только в путях его репозиториев (go-desktop-kit, проекты-источники);
// репозиторий самого шаблона ловит слаг desktop-app в его имени.
var templateNameRe = regexp.MustCompile(`Desktop App[^l]|desktop-app|DESKTOP_APP|desktopapp|template-init|scripts/init|` +
	`(?:^|[^a-z0-9_-])app-(?:build|ci|code-map|code-style|docs|release|spec|ui)(?:[^a-z0-9_-]|$)|` +
	`(?:^|[^/])yarkovaleksei`)

// checkNoTemplateNames ищет имена шаблона во всех файлах проекта, кроме
// ссылки «создано из шаблона» в README и закреплённого цвета иконки.
func checkNoTemplateNames(t *testing.T, dir string) {
	t.Helper()
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel := filepath.ToSlash(strings.TrimPrefix(p, dir+string(filepath.Separator)))
		if d.IsDir() {
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if rel == "scripts/icon/main_test.go" {
			return nil
		}
		var data []byte
		if d.Type()&fs.ModeSymlink != 0 {
			target, err := os.Readlink(p)
			if err != nil {
				return err
			}
			data = []byte(target)
		} else if data, err = os.ReadFile(p); err != nil {
			return err
		}
		if bytes.IndexByte(data, 0) >= 0 {
			return nil
		}
		for n, line := range strings.Split(rel+"\n"+string(data), "\n") {
			if (rel == "README.md" || rel == "README.en.md") && strings.Contains(line, "](https://github.com/"+templateRepo+")") {
				continue
			}
			if m := templateNameRe.FindString(line); m != "" {
				t.Errorf("%s:%d: имя шаблона %q: %s", rel, n, m, line)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
