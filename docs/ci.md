# CI

Четыре workflow GitHub Actions: что запускает каждый, на какие изменения, как читать прогон и разбирать падение. Workflow лежат в `.github/workflows/`; проверка их синтаксиса — `actionlint` (правило для `pages.yml` в `.github/actionlint.yaml`).

| Workflow | Когда | Что делает |
| --- | --- | --- |
| [check](#check) | push в `main`, PR, вручную; не на изменения только в документах | Проверки, привязки, сборка и запуск окна на трёх ОС; сквозные сценарии |
| [docs](#docs) | push в `main`, PR, вручную; на любые изменения | Документы, карта кода и устройство репозитория |
| [release](#release) | тег `v[0-9]*` | Пакеты на трёх ОС и черновик выпуска |
| [pages](#pages) | push с изменениями `www/**`, вручную | Публикация сайта; в шаблоне выключен |

Права по умолчанию — `contents: read`; запись есть только у job `release` (`contents: write`) и у `deploy` в `pages`. Секретов, кроме `github.token`, нет. Действия закреплены по мажорной версии (`actions/checkout@v7` и другие); Go — по версии из `go.mod`, Node — 24, `wails3` — версии Wails из `go.mod`.

## check

`.github/workflows/check.yml`. Те же задачи, что разработчик запускает перед коммитом, затем сборка и окно.

**Фильтр путей.** Не запускается, если в push или PR изменились только: Markdown-файлы в любой папке, `docs/**`, `spec/**`, `www/**`, шаблоны issues и PR, `pages.yml`, `docs.yml`. Любой другой файл в том же push — запускает. Ручной запуск (`gh workflow run check`) нужен перед выпуском, если последний коммит был только с документами: `release:plan` требует зелёного `check` на HEAD.

**Job `check`** — матрица `ubuntu-24.04`, `macos-latest`, `windows-latest`, `fail-fast: false`:

1. Go, Node, `corepack enable`; на Linux — GTK 4, WebKitGTK 6.0 и xvfb.
2. `wails3` той же версии, что библиотека Wails в `go.mod`.
3. `wails3 task check` — тесты, типы, линтеры, документы.
4. `wails3 task check:bindings` — закоммиченные привязки совпадают с Go.
5. `wails3 task build`.
6. Linux: если ядро запрещает user namespace без профиля, загружается профиль AppArmor из `build/linux/apparmor/desktop-app` для пути `bin/desktop-app`; песочница WebKit не отключается.
7. `wails3 task smoke` (на Linux — под `xvfb-run`): окно открылось, фронтенд ответил. При падении отчёт и журнал приложения сохраняются артефактом `smoke-<ОС>`.

**Job `e2e`** — Ubuntu: `wails3 task deps`, Chromium для Playwright (кэш по версии Playwright), `wails3 task test:e2e`. При падении — артефакт `playwright-report`: отчёт и трассы упавших сценариев.

Новый push в ветку PR отменяет прежний прогон; push в `main` и ручные запуски доходят до конца.

## docs

`.github/workflows/docs.yml`. Лёгкая сверка на любые изменения, в том числе только документов, которые `check` пропускает: `go test ./internal/repodocs ./internal/archtest` на Ubuntu, без GTK, Node и сборки фронтенда. Ловит битые ссылки, якоря и пути в документах, устаревшую карту кода, имена приложения вне Go, коды ошибок без текстов, Wails вне `main` и `internal/app`. Локально то же — `wails3 task check:docs` и `go test ./internal/archtest`.

## release

`.github/workflows/release.yml`. Запускается тегом `vX.Y.Z`; порядок выпуска вокруг него — в [release.md](release.md).

1. **Job `package`** — матрица трёх ОС: версия из тега без `v`, на Windows — NSIS, на Linux — GTK 4 и WebKitGTK; `wails3 task test`, `wails3 task package`, `wails3 task release:collect` с этой версией; содержимое `dist/` уходит артефактом `release-<ОС>`.
2. **Job `release`** — Ubuntu: собирает артефакты в один каталог, создаёт черновик выпуска (`--draft --verify-tag`, заметки `--generate-notes`), если его ещё нет, и загружает файлы. Job ждёт пакеты всех трёх ОС: без любого из них черновик не создаётся. Загрузка без `--clobber`: уже загруженный файл не перезаписывается, поэтому повтор job `release` на частичном черновике падает — черновик удаляется и job перезапускается ([release.md](release.md#если-что-то-пошло-не-так)).

Workflow ничего не публикует: черновик публикует `wails3 task release:promote` после сверки.

## pages

`.github/workflows/pages.yml` публикует каталог `www/` на GitHub Pages. В шаблоне job выключен условием `if: false`: у самого шаблона сайта нет. Проект, начатый из шаблона, получает условие на свой репозиторий при инициализации; GitHub Pages включается в настройках репозитория (источник — GitHub Actions).

## Как читать прогон

- `gh run list --limit 5` — последние запуски с итогом; `gh run view <id>` — jobs и шаги; `gh run view <id> --log-failed` — журнал только упавших шагов.
- Имя шага — подсказка, какая задача упала: шаг «Проверки» — это `wails3 task check`, «Привязки совпадают с Go-кодом» — `check:bindings`, «Запуск окна (smoke)» — `smoke`.
- Артефакты (`smoke-<ОС>`, `playwright-report`) — внизу страницы запуска или `gh run download <id>`.

## Как разбирать падение

1. Повторить локально ту же задачу: CI запускает только `wails3 task …`, своих скриптов проверок у него нет. Если локально зелёно — смотреть, чем отличается окружение (ОС, чистый клон, отсутствие кэшей).
2. По шагу:

| Где упало | Частая причина | Что делать |
| --- | --- | --- |
| `check` → тесты Go | Тест правила репозитория (`internal/archtest`): имя приложения в одном файле, код ошибки без текста | Сообщение теста называет файл и строку |
| `check` → документы | Ссылка, путь или задача в Markdown не существуют, карта кода устарела | `wails3 task check:docs`; карта — `wails3 task docs:map` и коммит |
| `check:bindings` | Изменились методы сервисов, а привязки не перегенерированы | `wails3 task bindings` и коммит `frontend/bindings` |
| `build` на одной ОС | Код для одной ОС без build-тега, зависимость CGO | Локально `GOOS=<ОС> go vet ./...` |
| smoke | Фронтенд не вызвал `FrontendReady` за 30 с, ошибка при старте | Артефакт `smoke-<ОС>`: `report.json` и журнал приложения в `home/` |
| smoke на Linux | Песочница WebKit без профиля AppArmor | Проверить, что шаг профиля нашёл путь бинаря |
| `e2e` | Изменился текст или роль элемента, ошибка в консоли или запрос в сеть | Артефакт `playwright-report`: трасса сценария; локально — `wails3 task test:e2e` |
| `release` → `package` | Нет NSIS на Windows, сборка dmg на macOS | Журнал шага «Пакеты»; перезапуск упавшего job после исправления окружения |

3. Если падение случайное (сеть раннера, таймаут скачивания) — `gh run rerun <id> --failed`. Повторяющееся падение случайным не считается.
