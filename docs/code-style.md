# Правила кода

Правила кода Go и TypeScript и проверки перед коммитом. У каждого правила названо, чем оно проверяется: командой — тогда нарушение роняет её, или обзором — тогда правило держит тот, кто пишет и смотрит изменение. Устройство приложения, на которое опираются правила, — в [development.md](development.md); правила интерфейса — в [ui.md](ui.md).

## Перед коммитом

`wails3 task check` — зависимости, тесты Go и фронтенда, типы, линтеры, документы и карта кода. Дополнительно, если изменение их касается:

| Что менялось | Команда |
| --- | --- |
| Методы сервисов, события, типы их данных | `wails3 task bindings` и `frontend/bindings` в том же коммите; `wails3 task check:bindings` сверяет с закоммиченным и до коммита падает — его запускает CI или ты после коммита |
| Интерфейс: страницы, переключатели, состояния | `wails3 task test:e2e` |
| Ядро с горутинами и общими данными | `wails3 task check:race` |
| `main.go`, `desktop.go`, окно | `wails3 task build` и `wails3 task smoke` |
| Пакет Go, страница, `internal/repodocs/areas.go`, комментарий пакета | `wails3 task docs:map` |

Отформатировать всё: `wails3 task fmt`.

## Общее

| Правило | Чем проверяется |
| --- | --- |
| Комментарии, документы и сообщения коммитов — по-русски; идентификаторы, имена файлов и API — по-английски | обзором |
| Заголовок коммита — в повелительном наклонении («Добавить страницу журнала»), в теле — зачем и что изменилось | обзором |
| Имена приложения (название, слаг, модуль, репозиторий, идентификатор, префикс окружения) в Go-коде берутся только из `internal/appinfo`; копии вне Go совпадают с ними | `wails3 task test:go` (`internal/archtest/names_test.go`) |
| В отслеживаемых файлах нет абсолютных путей машины разработчика, секретов и токенов | обзором |
| Новая зависимость (модуль Go, пакет npm) — только с записью решения и причины в `spec/plan/DECISIONS.md` | обзором |
| Изменение поведения области — вместе с её документами из [карты кода](code-map.md), в том же коммите | обзором; ссылки и пути в документах — `wails3 task check:docs` |

## Go

| Правило | Чем проверяется |
| --- | --- |
| Ядро (`internal/`, кроме `internal/app`) не импортирует Wails и `kit/wailsx` и ничего не знает о фронтенде | `wails3 task test:go` (`internal/archtest/wails_test.go`) |
| `internal/app` не содержит логики: только перевод вызовов в ядро, ответов в DTO, ошибок в `*apperr.Error` | обзором |
| Код отформатирован `gofmt` | `wails3 task lint:go` |
| `go vet` и staticcheck без замечаний | `wails3 task lint:go` |
| Ошибки оборачиваются с контекстом: `fmt.Errorf("…: %w", err)` | обзором |
| Наружу, во фронтенд, уходит только `*apperr.Error`: сервис переводит ошибку через `toError` (`internal/app/errors.go`) | обзором; каждый код есть в словарях — `wails3 task test:go` (`internal/archtest/errorcodes_test.go`) и `wails3 task test:frontend` |
| `context.Context` принимает то, что может ждать неограниченно долго: сеть, внешние процессы, повторы и опросы. Короткие операции с локальными файлами обходятся без него | обзором |
| Пути — через `filepath`; каталоги приложения — только из `appdirs`; код для одной ОС — в файлах с суффиксом ОС или build-тегом | обзором |
| У каждого пакета комментарий `// Package`; его первое предложение — назначение пакета в карте кода | `wails3 task check:docs` |
| У экспортируемых деклараций — комментарий о назначении и инварианте | обзором |
| Ответы сервисов проходят `wire.Fill`: пустые срезы и карты уходят во фронтенд как `[]` и `{}`, а не `null` | обзором |
| Тексты, которые показывает Go (трей, уведомления), — в `internal/texts`, на обоих языках | `wails3 task test:go` (`internal/texts`, `internal/archtest/texts_test.go`) |

Примеры из шаблона:

- Ошибка ядра с кодом — `Validate` в `internal/settings/settings.go`; перевод ошибок для фронтенда — `toError` в `internal/app/errors.go`.
- Зависимости сервиса подставляются через опции, чтобы тест обходился без ОС: `AppOptions.OpenDir` и `SettingsOptions.Emit` в `internal/app`.
- Код для одной ОС — `internal/texts/locale_unix.go` и `internal/texts/locale_windows.go`.
- Долгая операция с `context.Context` — `AppService.OpenLogFolder` (запуск файлового менеджера).

## TypeScript и React

| Правило | Чем проверяется |
| --- | --- |
| TypeScript строгий: `strict`, `noUnusedLocals`, `noUnusedParameters`, `noFallthroughCasesInSwitch`, `noUncheckedIndexedAccess` | `wails3 task typecheck` |
| ESLint (рекомендованные правила JS, TypeScript и хуков React) без замечаний | `wails3 task lint:frontend` |
| Код отформатирован Prettier (ширина строки 110) | `wails3 task lint:frontend` |
| Привязки Wails и `@wailsio/runtime` импортирует только `frontend/src/api/bindings.ts` | обзором |
| Компоненты и стор не зовут привязки: всё, что нужно от Go, — через `DataSource` | обзором; компилятор ловит метод, которого нет в `DataSource`, — `wails3 task typecheck` |
| Вызов Go оборачивается `call()`: наружу выходит только `AppError` с кодом | обзором |
| Текст ошибки — `t("errors.<код>", params)`, а не `message` | обзором |
| Ни одной строки интерфейса в JSX — только `t("ключ")` | `wails3 task lint:frontend` (правило `i18next/no-literal-string`); что правило включено — `frontend/src/test/i18n-lint.test.ts` |
| Чистые функции (форматирование, вычисления) — в `frontend/src/logic/` с тестами | обзором |
| Подписка на событие возвращает отписку, и компонент или стор её вызывает | обзором |

Примеры из шаблона: источник данных — `frontend/src/api/source.ts` и `frontend/src/api/wailsSource.ts`; действие стора с показом ошибки — `run` в `frontend/src/store/appStore.ts`; чистая функция с тестом — `frontend/src/logic/format.ts`.
