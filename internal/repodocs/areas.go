// Package repodocs хранит таблицу областей карты кода; тесты пакета сверяют карту и документы репозитория с кодом.
//
// Новый Go-пакет или новая страница интерфейса без строки в Areas роняет
// TestCodeMap: так карта не отстаёт от кода. После правки таблицы, комментария
// пакета или JSDoc страницы карту перегенерирует `wails3 task docs:map`.
// TestDocs проверяет ссылки, якоря, пути и команды во всех Markdown-файлах
// репозитория (`wails3 task check:docs`).
package repodocs

import kitdocs "github.com/yarkovaleksei/go-desktop-kit/repodocs"

// Areas — области шаблона в порядке карты. Docs — документы-владельцы:
// изменение поведения области описывается в них в том же коммите.
var Areas = []kitdocs.Area{
	{
		Title:    "Точка входа и граница Wails",
		Docs:     []string{"docs/development.md#порядок-запуска", "docs/development.md#сервисы", "docs/development.md#контракт-go--typescript"},
		Packages: []string{".", "internal/app"},
		Paths:    []string{"frontend/bindings"},
	},
	{
		Title:    "Настройки и сведения о приложении",
		Docs:     []string{"docs/development.md#настройки", "docs/development.md#каталоги-и-переменные-окружения"},
		Packages: []string{"internal/appinfo", "internal/settings", "internal/texts"},
	},
	{
		Title:    "Рабочий стол: трей, автозапуск, уведомления, фоновая проверка",
		Docs:     []string{"docs/desktop.md"},
		Packages: []string{"internal/updatecheck"},
		Paths:    []string{"desktop.go", "build/trayicon.png", "build/trayicon-template.png"},
	},
	{
		Title: "Интерфейс: каркас",
		Docs:  []string{"docs/development.md#контракт-go--typescript", "docs/development.md#ошибки", "docs/development.md#события", "docs/ui.md"},
		Paths: []string{
			"frontend/src/main.tsx", "frontend/src/app", "frontend/src/api", "frontend/src/store",
			"frontend/src/i18n", "frontend/src/theme", "frontend/src/components", "frontend/src/logic",
			"frontend/src/test", "frontend/index.html", "frontend/package.json",
		},
	},
	{
		Title: "Интерфейс: страницы",
		Docs:  []string{"docs/development.md#страница", "docs/ui.md"},
		Pages: []string{"HomePage", "SettingsPage"},
		Paths: []string{"frontend/e2e"},
	},
	{
		Title:    "Сборка и поставки",
		Docs:     []string{"docs/packaging.md", "docs/release.md", "docs/ci.md"},
		Packages: []string{"scripts/icon"},
		Paths: []string{
			"Taskfile.yml", "build", "scripts/smoke-desktop.sh",
			".github/workflows/check.yml", ".github/workflows/release.yml", ".github/workflows/pages.yml",
		},
	},
	{
		Title:    "Проверки репозитория",
		Docs:     []string{"docs/development.md#тесты", "docs/code-style.md", "docs/ci.md#docs"},
		Packages: []string{"internal/repodocs", "internal/archtest"},
		Paths:    []string{".github/workflows/docs.yml", ".github/actionlint.yaml", "go.mod", "README.md", "README.en.md", "docs/README.md"},
	},
	{
		Title: "Инструкции для агентов",
		Docs:  []string{"AGENTS.md", "CLAUDE.md"},
		Paths: []string{".agents/skills", ".claude/skills", ".claude/launch.json"},
	},
	{
		Title: "План проекта",
		Docs:  []string{"spec/AGENTS.md", "spec/README.md"},
		Paths: []string{"spec"},
	},
	{
		Title: "Инициализация шаблона",
		Docs:  []string{"README.md#как-начать-проект-из-шаблона", "README.en.md#starting-a-project-from-the-template"},
	},
}
