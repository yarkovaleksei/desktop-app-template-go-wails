# Desktop App

[English](README.en.md)

Шаблон десктоп-приложения на Go, Wails 3, React и MUI. Собирается под Linux, Windows и macOS, открывает окно с настройками, треем и обновлением через GitHub Releases и проходит проверки сразу после клонирования.

## Требования к окружению

| Что | Версия | Зачем |
| --- | --- | --- |
| Go | 1.26 | Приложение и инструменты |
| Node.js | 24 | Фронтенд |
| yarn | 4 (через `corepack enable`) | Зависимости фронтенда |
| Wails CLI `wails3` | v3.0.0: `go install github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0` | Сборка, задачи (`wails3 task …`), привязки |
| Linux | GTK 4 и WebKitGTK 6.0 для сборки (`libgtk-4-dev`, `libwebkitgtk-6.0-dev` в Ubuntu 24.04) | Окно |
| Windows | WebView2 (есть в Windows 10 и 11); NSIS — только для установщика | Окно, пакеты |
| macOS | Xcode Command Line Tools, macOS 12 и новее | Сборка |

Task отдельно ставить не нужно: задачи запускает `wails3 task <имя>`.

## Быстрый старт

```bash
git clone https://github.com/yarkovaleksei/desktop-app-template-go-wails.git
cd desktop-app-template-go-wails
corepack enable
wails3 task deps
wails3 task dev
```

`wails3 task dev` открывает окно в режиме разработки: изменения фронтенда применяются сразу, настройки и журнал живут в `.task/dev-home`. Интерфейс без Go, в браузере на моках, — `wails3 task ui`.

## Команды

| Команда | Что делает |
| --- | --- |
| `wails3 task deps` | Зависимости фронтенда по `yarn.lock` |
| `wails3 task dev` | Окно в режиме разработки; `REAL=true` — на настоящих каталогах пользователя |
| `wails3 task ui` | Интерфейс в браузере на моках, без Go |
| `wails3 task site` | Сайт `www/` на локальном сервере |
| `wails3 task build`, `wails3 task run` | Бинарь текущей ОС в `bin/` и его запуск |
| `wails3 task check` | Всё перед коммитом: зависимости, тесты, типы, линтеры, документы |
| `wails3 task test`, `wails3 task test:go`, `wails3 task test:frontend` | Тесты Go и фронтенда |
| `wails3 task test:e2e` | Сквозные сценарии Playwright на моках |
| `wails3 task lint`, `wails3 task typecheck`, `wails3 task fmt` | Линтеры, проверка типов, форматирование |
| `wails3 task check:race` | Тесты ядра под детектором гонок |
| `wails3 task bindings`, `wails3 task check:bindings` | Перегенерировать привязки TypeScript; проверить, что они не отстали |
| `wails3 task check:docs`, `wails3 task docs:map` | Проверить документы и карту кода; перегенерировать карту |
| `wails3 task smoke` | Запустить собранный бинарь и проверить, что окно открылось |
| `wails3 task smoke:desktop` | Linux: запустить установленное приложение как из меню |
| `wails3 task package` | Все пакеты текущей ОС в `bin/` |
| `wails3 task release:collect` | Пакеты под именами выпуска и файл сумм в `dist/` |
| `wails3 task release:plan`, `wails3 task release:verify`, `wails3 task release:promote` | Состояние перед выпуском, сверка черновика, публикация |
| `wails3 task version` | Версия, которая попадёт в сборку |
| `wails3 task icons`, `wails3 task icons:derive` | Нарисовать шаблонную иконку; обновить `.ico` и `.icns` из своей |

## Структура репозитория

```text
main.go, desktop.go   точка входа и связка с рабочим столом
internal/             app (сервисы Wails), appinfo, settings, texts, updatecheck,
                      repodocs (карта кода), archtest (правила репозитория)
frontend/             React + MUI: src/ (api, store, pages, i18n, theme), e2e/, bindings/
build/                конфигурация Wails, иконки, пакеты Linux, Windows, macOS
scripts/              icon (генератор иконки), smoke-desktop.sh
docs/                 руководства и карта кода
www/                  сайт проекта: одна страница на двух языках
spec/                 план проекта: правила ведения, этапы, решения, бэклог, журнал
.github/              workflows check, docs, release, pages; шаблоны issues и PR
Taskfile.yml          задачи
```

Подробная карта — [docs/code-map.md](docs/code-map.md).

## Установка

Файлы — на странице [выпусков](https://github.com/yarkovaleksei/desktop-app-template-go-wails/releases/latest); `<версия>` — номер выпуска без `v`. Пакеты не подписаны — как открыть их на Windows и macOS, описано в [docs/packaging.md](docs/packaging.md#подписи-нет).

| ОС | Пакет | Установка |
| --- | --- | --- |
| Debian, Ubuntu | `desktop-app_<версия>_amd64.deb` | `sudo apt install ./desktop-app_<версия>_amd64.deb` |
| Fedora | `desktop-app-<версия>.x86_64.rpm` | `sudo dnf install ./desktop-app-<версия>.x86_64.rpm` |
| Arch Linux | `desktop-app-<версия>-x86_64.pkg.tar.zst` | `sudo pacman -U desktop-app-<версия>-x86_64.pkg.tar.zst` |
| Linux, любой | `desktop-app-<версия>-x86_64.AppImage` | `chmod +x` и запуск; на Ubuntu при первом запуске спросит разрешение на профиль AppArmor |
| Windows | `desktop-app-<версия>-windows-amd64-setup.exe` | Установщик, без прав администратора |
| Windows | `desktop-app-<версия>-windows-amd64.exe` | Переносимый файл без установки |
| macOS | `desktop-app-<версия>-macos-universal.dmg` | Открыть образ и перенести приложение в «Программы» |

Суммы SHA-256 — в `SHA256SUMS-linux.txt`, `SHA256SUMS-windows.txt` и `SHA256SUMS-macos.txt` того же выпуска. Установленное из пакета приложение обновляется со страницы настроек.

## Документация

Карта документации — [docs/README.md](docs/README.md): какой документ чем владеет и когда его обновлять. Главное: устройство — [docs/development.md](docs/development.md), рабочий стол — [docs/desktop.md](docs/desktop.md), правила кода — [docs/code-style.md](docs/code-style.md), интерфейс — [docs/ui.md](docs/ui.md), поставки — [docs/packaging.md](docs/packaging.md), выпуск — [docs/release.md](docs/release.md), CI — [docs/ci.md](docs/ci.md), рецепты — [docs/recipes/](docs/recipes/README.md).

## Как начать проект из шаблона

Инструмента инициализации пока нет. Имена-заглушки шаблона — название `Desktop App`, слаг `desktop-app`, модуль Go, репозиторий, идентификатор продукта, префикс окружения `DESKTOP_APP` — записаны в `internal/appinfo/appinfo.go`. Каждое место вне Go, где они повторены, знает тест `internal/archtest/names_test.go`: после замены имён `wails3 task check` назовёт пропущенные.

## Лицензия

[MIT](LICENSE).
