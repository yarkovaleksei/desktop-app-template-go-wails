package main

import (
	"bytes"
	"embed"
	"fmt"
	"slices"
	"strings"
	"text/template"
)

// templates — заготовки файлов проекта: шапки README.
//
//go:embed templates/*.tmpl
var templates embed.FS

// removed — служебные части шаблона, которых в проекте нет: путь целиком
// или каталог с «/» на конце.
var removed = []string{
	"scripts/init/",
	".agents/skills/template-init/",
	".claude/skills/template-init",
}

// verbatim — файлы, в которых имена шаблона остаются намеренно.
// scripts/icon/main_test.go закрепляет цвет плашки для слага desktop-app:
// это проверка алгоритма цвета, а не имя проекта.
var verbatim = []string{"scripts/icon/main_test.go"}

// edit — точечная правка служебного места шаблона. Тексты записаны в именах
// шаблона: правила замены применяются после правок. Фрагмент old должен
// найтись в файле ровно один раз — иначе шаблон разошёлся с инструментом, и
// инициализация останавливается, ничего не изменив.
type edit struct {
	file, old, new string
}

var edits = []edit{
	{
		file: "AGENTS.md",
		old:  "| [template-init](.agents/skills/template-init/SKILL.md) | Создать проект из шаблона: имена, иконки, следующие шаги; удаляется инициализацией |\n",
		new:  "",
	},
	{
		file: "internal/repodocs/areas.go",
		old: "\t{\n\t\tTitle:    \"Инициализация шаблона\",\n" +
			"\t\tDocs:     []string{\"README.md#как-начать-проект-из-шаблона\", \"README.en.md#starting-a-project-from-the-template\"},\n" +
			"\t\tPackages: []string{\"scripts/init\"},\n" +
			"\t\tPaths:    []string{\".agents/skills/template-init\"},\n" +
			"\t},\n",
		new: "",
	},
	{
		file: "internal/appinfo/appinfo.go",
		old: "// Значения — заглушки шаблона. Инициализация нового проекта (scripts/init)\n" +
			"// заменяет их на имена проекта вместе с теми же строками в остальных файлах.\n",
		new: "// Те же имена повторены вне Go — в файлах сборки, пакетов и сайта; их\n" +
			"// совпадение с этими константами проверяет internal/archtest.\n",
	},
	{
		file: ".github/workflows/pages.yml",
		old: "    # Сайт-заготовка шаблона не публикуется, поэтому job выключен.\n" +
			"    # scripts/init находит строку по метке в комментарии и ставит условие\n" +
			"    # github.repository == '<владелец>/<репозиторий>': форки сайт не публикуют.\n" +
			"    if: false # template-init: pages-condition\n",
		new: "    # Сайт публикует только репозиторий проекта: форки его не публикуют.\n" +
			"    if: github.repository == '" + templateRepo + "'\n",
	},
	{
		file: ".github/actionlint.yaml",
		old: "paths:\n" +
			"  # В шаблоне публикация сайта выключена намеренно (`if: false`); scripts/init\n" +
			"  # ставит вместо него условие на репозиторий проекта.\n" +
			"  .github/workflows/pages.yml:\n" +
			"    ignore:\n" +
			"      - 'constant expression \"false\" in condition'\n",
		new: "# Исключение для файла: paths: {<workflow>: {ignore: [<текст ошибки>]}}.\n" +
			"paths: {}\n",
	},
	{
		file: "docs/ci.md",
		old: "В шаблоне job выключен условием `if: false`: у самого шаблона сайта нет. " +
			"Проект, начатый из шаблона, получает условие на свой репозиторий при инициализации; " +
			"GitHub Pages включается в настройках репозитория (источник — GitHub Actions).",
		new: "Job выполняется только в репозитории проекта (условие на `github.repository`): форки сайт не публикуют. " +
			"GitHub Pages включается в настройках репозитория (источник — GitHub Actions).",
	},
	{
		file: "docs/ci.md",
		old: "Включают публикацию действия владельца репозитория: GitHub Pages в настройках (Settings → Pages → Source: GitHub Actions) " +
			"и условие `if: github.repository == '<владелец>/<репозиторий>'` в `pages.yml`, которое ставит инициализация проекта.",
		new: "Включает публикацию владелец репозитория: GitHub Pages в настройках (Settings → Pages → Source: GitHub Actions); " +
			"условие на репозиторий в `pages.yml` уже стоит.",
	},
	{
		file: "docs/README.md",
		old:  "установка пакетов, как начать проект из шаблона |",
		new:  "установка пакетов |",
	},
	{
		file: "README.md",
		old:  "scripts/              init (инициализация проекта из шаблона), icon (генератор иконки), smoke-desktop.sh",
		new:  "scripts/              icon (генератор иконки), smoke-desktop.sh",
	},
	{
		file: "README.en.md",
		old:  "scripts/              init (project initialization from the template), icon (icon generator), smoke-desktop.sh",
		new:  "scripts/              icon (icon generator), smoke-desktop.sh",
	},
}

// readme — README проекта: шапка из templates/<file>.tmpl, тело — README
// шаблона без шапки и без раздела section о шаблоне.
type readme struct {
	file, section string
}

var readmes = []readme{
	{"README.md", "## Как начать проект из шаблона"},
	{"README.en.md", "## Starting a project from the template"},
}

// yarnLock — lock-файл yarn. Записи в нём идут по алфавиту ключей, и
// `yarn install --immutable` отвергает файл, где переименованная запись
// рабочей области (<слаг>@workspace:.) стоит не на своём месте.
const yarnLock = "frontend/yarn.lock"

// projectDoc — описание продукта, которое получает проект: образец
// spec/templates/project.md с названием и описанием из флагов.
const (
	projectDoc         = "spec/docs/project.md"
	projectDocTemplate = "spec/templates/project.md"
)

// transform строит дерево проекта из дерева шаблона. Исходное дерево не
// меняется.
func transform(before []*entry, id Identity) ([]*entry, error) {
	rules := id.Rules()
	byPath := map[string]*entry{}
	var after []*entry
	for _, e := range before {
		if isRemoved(e.path) || isPlanLeftover(e.path) {
			continue
		}
		c := *e
		after = append(after, &c)
		byPath[c.path] = &c
	}

	for _, ed := range edits {
		e := byPath[ed.file]
		if e == nil {
			return nil, fmt.Errorf("шаблон разошёлся с scripts/init: нет файла %s", ed.file)
		}
		if n := bytes.Count(e.data, []byte(ed.old)); n != 1 {
			return nil, fmt.Errorf("шаблон разошёлся с scripts/init: в %s фрагмент найден %d раз вместо одного:\n%s", ed.file, n, ed.old)
		}
		e.data = bytes.Replace(e.data, []byte(ed.old), []byte(ed.new), 1)
	}

	bodies := map[string]string{}
	for _, r := range readmes {
		e := byPath[r.file]
		if e == nil {
			return nil, fmt.Errorf("нет файла %s", r.file)
		}
		body, err := readmeBody(string(e.data), r.section)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", r.file, err)
		}
		bodies[r.file] = body
	}

	for _, e := range after {
		if e.link || e.text() && !slices.Contains(verbatim, e.path) {
			e.data = []byte(replaceAll(string(e.data), rules))
		}
		e.path = replaceAll(e.path, rules)
	}
	byPath = map[string]*entry{}
	for _, e := range after {
		byPath[e.path] = e
	}
	if lock := byPath[yarnLock]; lock != nil {
		sorted, err := sortYarnLock(string(lock.data))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", yarnLock, err)
		}
		lock.data = []byte(sorted)
	}

	for _, r := range readmes {
		head, err := renderHead(r.file+".tmpl", id)
		if err != nil {
			return nil, err
		}
		byPath[r.file].data = []byte(head + replaceAll(bodies[r.file], rules))
	}

	tmpl := byPath[projectDocTemplate]
	doc := byPath[projectDoc]
	if tmpl == nil || doc == nil {
		return nil, fmt.Errorf("нет %s или %s", projectDocTemplate, projectDoc)
	}
	text, err := projectText(string(tmpl.data), id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", projectDocTemplate, err)
	}
	doc.data = []byte(text)
	return after, nil
}

func isRemoved(p string) bool {
	for _, r := range removed {
		if p == r || strings.HasSuffix(r, "/") && strings.HasPrefix(p, r) {
			return true
		}
	}
	return false
}

// isPlanLeftover — файл плана, которого нет в начальном состоянии `spec/`:
// этапы (кроме .gitkeep) и архив (кроме его README).
func isPlanLeftover(p string) bool {
	switch {
	case strings.HasPrefix(p, "spec/plan/stages/"):
		return p != "spec/plan/stages/.gitkeep"
	case strings.HasPrefix(p, "spec/archive/"):
		return p != "spec/archive/README.md"
	}
	return false
}

// readmeBody — README без шапки (всего до первого раздела) и без раздела
// section.
func readmeBody(text, section string) (string, error) {
	start := strings.Index(text, "\n## ")
	if start < 0 {
		return "", fmt.Errorf("нет разделов")
	}
	body := text[start+1:]
	i := strings.Index(body, section+"\n")
	if i < 0 || i > 0 && body[i-1] != '\n' {
		return "", fmt.Errorf("нет раздела %q", section)
	}
	end := strings.Index(body[i+len(section):], "\n## ")
	if end < 0 {
		return body[:i], nil
	}
	return body[:i] + body[i+len(section)+end+1:], nil
}

// renderHead — шапка README проекта из заготовки.
func renderHead(name string, id Identity) (string, error) {
	t, err := template.ParseFS(templates, "templates/"+name)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	err = t.Execute(&b, struct {
		Identity
		Sentence     string
		TemplateRepo string
	}{id, sentence(id.Description), templateRepo})
	return b.String(), err
}

// projectText — образец описания проекта из блока markdown в
// spec/templates/project.md: пояснение раздела «Продукт» заменено
// названием и описанием, остальные пояснения остаются заданием на
// заполнение.
func projectText(tmpl string, id Identity) (string, error) {
	const fence, product = "```markdown\n", "## Продукт\n\n"
	i := strings.Index(tmpl, fence)
	if i < 0 {
		return "", fmt.Errorf("нет блока %q", strings.TrimSpace(fence))
	}
	block := tmpl[i+len(fence):]
	end := strings.Index(block, "\n```")
	if end < 0 {
		return "", fmt.Errorf("блок markdown не закрыт")
	}
	block = block[:end+1]
	p := strings.Index(block, product)
	if p < 0 {
		return "", fmt.Errorf("нет раздела «Продукт»")
	}
	rest := block[p+len(product):]
	q := strings.Index(rest, "\n\n")
	if q < 0 {
		return "", fmt.Errorf("в разделе «Продукт» нет пояснения")
	}
	return block[:p+len(product)] + id.Name + " — " + sentence(id.Description) + rest[q:], nil
}

// sentence — описание с точкой на конце.
func sentence(s string) string {
	if strings.HasSuffix(s, ".") || strings.HasSuffix(s, "!") || strings.HasSuffix(s, "?") {
		return s
	}
	return s + "."
}

// sortYarnLock упорядочивает записи lock-файла yarn так же, как yarn: шапка
// с __metadata остаётся первой, записи — по ключу без кавычек.
func sortYarnLock(text string) (string, error) {
	blocks := strings.Split(strings.TrimRight(text, "\n"), "\n\n")
	head := 0
	for head < len(blocks) && !strings.Contains(blocks[head], "__metadata:") {
		head++
	}
	if head == len(blocks) {
		return "", fmt.Errorf("нет __metadata")
	}
	entries := blocks[head+1:]
	key := func(b string) string {
		line, _, _ := strings.Cut(b, "\n")
		return strings.Trim(strings.TrimSuffix(line, ":"), `"`)
	}
	slices.SortStableFunc(entries, func(a, b string) int { return strings.Compare(key(a), key(b)) })
	return strings.Join(blocks, "\n\n") + "\n", nil
}
