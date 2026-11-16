# Карта кода

<!-- Сгенерировано `wails3 task docs:map` из internal/repodocs/areas.go и комментариев `// Package` и описаний страниц. Не править вручную. -->

Карта показывает, где лежит код каждой области и какие документы ей принадлежат. После изменения поведения области обновите её документы в том же коммите. Назначение пакета — первое предложение его комментария `// Package`, страницы — комментария её компонента: чтобы поправить описание, правьте комментарий и перегенерируйте карту (`wails3 task docs:map`).

## Точка входа и граница Wails

Документы: [docs/development.md#порядок-запуска](development.md#порядок-запуска), [docs/development.md#сервисы](development.md#сервисы), [docs/development.md#контракт-go--typescript](development.md#контракт-go--typescript).

| Пакет | Назначение |
| --- | --- |
| [`.`](../main.go) | Desktop App — десктоп-приложение на Wails 3 и React. |
| [`internal/app`](../internal/app) | Сервисы Wails: тонкий слой между ядром в internal/ и фронтендом. |

Ещё: [`frontend/bindings`](../frontend/bindings).

## Настройки и сведения о приложении

Документы: [docs/development.md#настройки](development.md#настройки), [docs/development.md#каталоги-и-переменные-окружения](development.md#каталоги-и-переменные-окружения).

| Пакет | Назначение |
| --- | --- |
| [`internal/appinfo`](../internal/appinfo) | Имена приложения. |
| [`internal/settings`](../internal/settings) | Схема настроек приложения, их умолчания и проверка. |
| [`internal/texts`](../internal/texts) | Тексты, которые показывает Go-часть приложения: пункты меню трея, уведомления, название записи автозапуска. |

## Рабочий стол: трей, автозапуск, уведомления, фоновая проверка

Документы: [docs/desktop.md](desktop.md).

| Пакет | Назначение |
| --- | --- |
| [`internal/updatecheck`](../internal/updatecheck) | Фоновая проверка обновлений по настройке «Проверять обновления автоматически» (D-031). |

Ещё: [`desktop.go`](../desktop.go), [`build/trayicon.png`](../build/trayicon.png), [`build/trayicon-template.png`](../build/trayicon-template.png).

## Интерфейс: каркас

Документы: [docs/development.md#контракт-go--typescript](development.md#контракт-go--typescript), [docs/development.md#ошибки](development.md#ошибки), [docs/development.md#события](development.md#события), [docs/ui.md](ui.md).

Ещё: [`frontend/src/main.tsx`](../frontend/src/main.tsx), [`frontend/src/app`](../frontend/src/app), [`frontend/src/api`](../frontend/src/api), [`frontend/src/store`](../frontend/src/store), [`frontend/src/i18n`](../frontend/src/i18n), [`frontend/src/theme`](../frontend/src/theme), [`frontend/src/components`](../frontend/src/components), [`frontend/src/logic`](../frontend/src/logic), [`frontend/src/test`](../frontend/src/test), [`frontend/index.html`](../frontend/index.html), [`frontend/package.json`](../frontend/package.json).

## Интерфейс: страницы

Документы: [docs/development.md#страница](development.md#страница), [docs/ui.md](ui.md).

| Страница | Назначение |
| --- | --- |
| [`HomePage`](../frontend/src/pages/HomePage/HomePage.tsx) | Главная страница — заглушка с названием приложения и ссылкой на документацию проекта. |
| [`SettingsPage`](../frontend/src/pages/SettingsPage/SettingsPage.tsx) | Страница настроек — язык, тема, работа в фоне, автозапуск, автопроверка обновлений, сведения о сборке, блок обновлений и папка журнала. |

Ещё: [`frontend/e2e`](../frontend/e2e).

## Сборка и поставки

Документы: [docs/packaging.md](packaging.md), [docs/release.md](release.md), [docs/ci.md](ci.md).

| Пакет | Назначение |
| --- | --- |
| [`scripts/icon`](../scripts/icon) | Icon рисует шаблонную иконку приложения: белое окно на цветной плашке. |

Ещё: [`Taskfile.yml`](../Taskfile.yml), [`build`](../build), [`scripts/smoke-desktop.sh`](../scripts/smoke-desktop.sh), [`.github/workflows/check.yml`](../.github/workflows/check.yml), [`.github/workflows/release.yml`](../.github/workflows/release.yml).

## Сайт

Документы: [docs/ci.md#сайт](ci.md#сайт).

Ещё: [`www`](../www), [`.github/workflows/pages.yml`](../.github/workflows/pages.yml).

## Проверки репозитория

Документы: [docs/development.md#тесты](development.md#тесты), [docs/code-style.md](code-style.md), [docs/ci.md#docs](ci.md#docs).

| Пакет | Назначение |
| --- | --- |
| [`internal/repodocs`](../internal/repodocs) | Хранит таблицу областей карты кода; тесты пакета сверяют карту и документы репозитория с кодом. |
| [`internal/archtest`](../internal/archtest) | Тесты правил архитектуры шаблона. |

Ещё: [`.github/workflows/docs.yml`](../.github/workflows/docs.yml), [`.github/actionlint.yaml`](../.github/actionlint.yaml), [`go.mod`](../go.mod), [`README.md`](../README.md), [`README.en.md`](../README.en.md), [`docs/README.md`](README.md).

## Инструкции для агентов

Документы: [AGENTS.md](../AGENTS.md), [CLAUDE.md](../CLAUDE.md).

Ещё: [`.agents/skills`](../.agents/skills), [`.claude/skills`](../.claude/skills), [`.claude/launch.json`](../.claude/launch.json).

## План проекта

Документы: [spec/AGENTS.md](../spec/AGENTS.md), [spec/README.md](../spec/README.md).

Ещё: [`spec`](../spec).

## Инициализация шаблона

Документы: [README.md#как-начать-проект-из-шаблона](../README.md#как-начать-проект-из-шаблона), [README.en.md#starting-a-project-from-the-template](../README.en.md#starting-a-project-from-the-template).

| Пакет | Назначение |
| --- | --- |
| [`scripts/init`](../scripts/init) | Init превращает копию шаблона в новый проект: заменяет имена шаблона в отслеживаемых файлах и путях, ставит условие публикации сайта, сбрасывает spec/, пишет README проекта, перерисовывает иконки в цвет нового слага и удаляет служебные части шаблона, в том числе себя. |

Ещё: [`.agents/skills/template-init`](../.agents/skills/template-init).
