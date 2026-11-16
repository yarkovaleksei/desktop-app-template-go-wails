// Init превращает копию шаблона в новый проект: заменяет имена шаблона в
// отслеживаемых файлах и путях, ставит условие публикации сайта, сбрасывает
// spec/, пишет README проекта, перерисовывает иконки в цвет нового слага и
// удаляет служебные части шаблона, в том числе себя.
//
//	go run ./scripts/init -name "My App" -repo owner/my-app -description "…" [-dry-run]
//
// Изменения добавляются в индекс git, коммит остаётся владельцу.
// Инструмент работает только на чистом дереве шаблона: незакоммиченные
// изменения смешались бы с заменой имён, а повторный запуск не нашёл бы имён
// шаблона.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	if err := run(".", os.Args[1:], os.Stdout, time.Now().Year()); err != nil {
		fmt.Fprintln(os.Stderr, "init:", err)
		os.Exit(1)
	}
}

// Отказы на непригодном дереве.
var (
	errInitialized = errors.New("имён шаблона в go.mod нет: проект уже инициализирован, а scripts/init — остаток шаблона")
	errDirty       = errors.New("в дереве есть незакоммиченные изменения: закоммитьте или уберите их — иначе они смешаются с заменой имён")
)

// run разбирает флаги и инициализирует проект в каталоге root.
func run(root string, args []string, out io.Writer, year int) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	fs.SetOutput(out)
	var f Flags
	fs.StringVar(&f.Name, "name", "", "название продукта, например \"My App\" (обязательно)")
	fs.StringVar(&f.Slug, "slug", "", "слаг: бинарь, каталоги, пакеты; по умолчанию — из названия")
	fs.StringVar(&f.Module, "module", "", "модуль Go; по умолчанию github.com/<repo>")
	fs.StringVar(&f.Repo, "repo", "", "репозиторий GitHub владелец/имя (обязательно)")
	fs.StringVar(&f.ID, "id", "", "идентификатор продукта; по умолчанию io.github.<владелец>.<слаг без дефисов>")
	fs.StringVar(&f.Description, "description", "", "описание продукта в одну строку (обязательно)")
	fs.StringVar(&f.Author, "author", "", "правообладатель в LICENSE и файлах сборки; по умолчанию — владелец репозитория")
	fs.IntVar(&f.Year, "year", 0, "год в строке правообладателя; по умолчанию текущий")
	fs.StringVar(&f.Color, "color", "", "цвет плашки иконки #RRGGBB; по умолчанию — из слага")
	dryRun := fs.Bool("dry-run", false, "только показать изменения, ничего не меняя")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("лишние аргументы: %s", strings.Join(fs.Args(), " "))
	}
	id, err := Derive(f, year)
	if err != nil {
		return err
	}
	if err := checkTree(root); err != nil {
		return err
	}

	before, err := readTree(root)
	if err != nil {
		return err
	}
	after, err := transform(before, id)
	if err != nil {
		return err
	}
	changes := diff(before, after)

	printIdentity(out, id)
	fmt.Fprintf(out, "\nИзменения (%d):\n", len(changes))
	for _, c := range changes {
		fmt.Fprintln(out, c)
	}
	iconArgs := iconCommand(id)
	fmt.Fprintf(out, "  %-14s go %s\n", "перерисовать", strings.Join(iconArgs, " "))

	if *dryRun {
		fmt.Fprintln(out, "\nПробный запуск (-dry-run): ничего не изменено.")
		printNextSteps(out, id, "После запуска")
		return nil
	}
	if err := apply(root, before, after); err != nil {
		return fmt.Errorf("запись дерева прервана, верните его командой git checkout . && git clean -fd: %w", err)
	}
	cmd := exec.Command("go", iconArgs...)
	cmd.Dir = root
	cmd.Stdout, cmd.Stderr = out, out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("иконки не перерисованы (повторите: go %s): %w", strings.Join(iconArgs, " "), err)
	}
	// Индекс — вместе с деревом: задачи берут списки файлов из git ls-files,
	// а git status показывает переименования, а не пары «удалён — новый».
	// Коммит остаётся владельцу.
	if _, err := git(root, "add", "-A"); err != nil {
		return err
	}
	fmt.Fprintln(out, "\nИзменения добавлены в индекс git (git add -A); коммит не сделан.")
	printNextSteps(out, id, "Готово")
	return nil
}

// checkTree отказывается работать на дереве без имён шаблона и с
// незакоммиченными изменениями.
func checkTree(root string) error {
	mod, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return fmt.Errorf("запускайте из корня шаблона: %w", err)
	}
	if !strings.HasPrefix(strings.ReplaceAll(string(mod), "\r\n", "\n"), "module "+templateModule+"\n") {
		return errInitialized
	}
	status, err := git(root, "status", "--porcelain")
	if err != nil {
		return err
	}
	if status = strings.TrimRight(status, "\n"); status != "" {
		lines := strings.Split(status, "\n")
		if len(lines) > 10 {
			lines = append(lines[:10], "…")
		}
		return fmt.Errorf("%w:\n%s", errDirty, strings.Join(lines, "\n"))
	}
	return nil
}

// iconCommand — аргументы go для перерисовки шаблонной иконки и картинок
// сайта в цвет проекта.
func iconCommand(id Identity) []string {
	args := []string{"run", "./scripts/icon", "-slug", id.Slug, "-out", "build", "-site", "www/images"}
	if id.Color != "" {
		args = append(args, "-color", id.Color)
	}
	return args
}

func printIdentity(out io.Writer, id Identity) {
	fmt.Fprintln(out, "Имена проекта:")
	for _, row := range [][2]string{
		{"название", id.Name},
		{"слаг", id.Slug},
		{"модуль Go", id.Module},
		{"репозиторий", id.Repo},
		{"идентификатор", id.ID},
		{"окружение", id.EnvPrefix + "_HOME, " + id.EnvPrefix + "_CONFIG_DIR, …"},
		{"скиллы", id.SkillPrefix + "build, " + id.SkillPrefix + "ui, …"},
		{"описание", id.Description},
		{"правообладатель", "© " + id.Year + " " + id.Author},
		{"сайт", "https://" + id.PagesHost + "/" + id.RepoName + "/"},
	} {
		fmt.Fprintf(out, "  %-16s %s\n", row[0], row[1])
	}
}

// nextSteps — что сделать после инициализации, по порядку.
func nextSteps(id Identity) []string {
	return []string{
		"wails3 task deps — зависимости фронтенда под новым именем пакета",
		"wails3 task bindings — привязки TypeScript под новым модулем",
		"wails3 task fmt — путь модуля другой длины, и prettier переносит импорты привязок иначе",
		"wails3 task icons:derive — .ico и .icns из перерисованной build/appicon.png",
		"wails3 task docs:map — карта кода без области «Инициализация шаблона»",
		"wails3 task check — всё зелёное; затем git add -A и коммит «Инициализировать проект из шаблона»",
		"заменить шаблонную иконку своей (build/appicon.png, затем wails3 task icons:derive) и снимок окна www/images/screen-main.png",
		"заполнить spec/docs/project.md: возможности, пользователи, границы",
		"поставить цель первого плана (скилл " + id.SkillPrefix + "spec, spec/AGENTS.md)",
		"на GitHub: включить Pages (Settings → Pages → Source: GitHub Actions), если нужен сайт",
	}
}

// printNextSteps печатает следующие шаги. Они меняют ещё несколько файлов:
// привязки, их импорты, .ico и .icns, карту кода.
func printNextSteps(out io.Writer, id Identity, title string) {
	fmt.Fprintf(out, "\n%s. Следующие шаги (они изменят ещё привязки, .ico и .icns, карту кода):\n", title)
	for i, s := range nextSteps(id) {
		fmt.Fprintf(out, "  %d. %s\n", i+1, s)
	}
}
