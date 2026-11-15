# Второй бинарник в cmd/

## Когда нужно

Рядом с окном нужна отдельная программа из того же модуля: консольный интерфейс к тем же данным, MCP-сервер для агентов, помощник, который запускает основное приложение. Если код нужен только внутри приложения, второй бинарник не нужен — достаточно пакета в `internal/`.

## Что добавить

- `cmd/<имя>/main.go` — только разбор флагов и сборка зависимостей; логика — в пакетах `internal/`, общих с окном. Пакет `cmd/<имя>` не импортирует Wails: тест `internal/archtest/wails_test.go` разрешает Wails только корневому `main` и `internal/app` и проверит новый пакет сам.
- Каталоги и настройки — тем же `appdirs.Resolve` с тем же слагом и префиксом, что в `main.go`, иначе программы разойдутся по разным каталогам.
- Задача сборки в `Taskfile.yml` (бинарь в `bin/`) и зависимость `build` и `package` от неё.
- Поставки: строка `contents` в `build/linux/nfpm/nfpm.yaml` (`/usr/bin/<имя>`); копия в `.app/Contents/MacOS` в `build/darwin/Taskfile.yml`; файл в установщике `build/windows/nsis/project.nsi`. В AppImage — в `usr/bin` AppDir в `build/linux/appimage/finalize.sh`.
- Строка пакета в `internal/repodocs/areas.go` и `wails3 task docs:map`.

Пакеты go-desktop-kit: [appdirs](https://github.com/yarkovaleksei/go-desktop-kit/tree/v0.1.1/appdirs), [applog](https://github.com/yarkovaleksei/go-desktop-kit/tree/v0.1.1/applog) — свой файл журнала второй программы.

## На что смотреть в образцах

- `ai-memory/cmd/ai-memory-mcp/main.go` — MCP-мост поверх ядра приложения, с тестом рядом.
- `ai-memory/Taskfile.yml` — задача сборки второго бинаря; `ai-memory/build/linux/nfpm/nfpm.yaml` — его место в пакетах Linux; `ai-memory/build/darwin/Taskfile.yml` — универсальный бинарь и копия в `.app`.

## Как проверить

- `go test ./cmd/<имя>` и `wails3 task check` (тест архитектуры и карта кода видят новый пакет).
- `wails3 task package`: бинарь есть в каждом пакете (`dpkg-deb -c`, `--appimage-extract`), и `<имя> --help` работает из установленного пакета.
