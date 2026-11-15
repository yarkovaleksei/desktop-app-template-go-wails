# Устройство приложения

Документ описывает, как приложение собрано из частей и как в него добавить новое: метод сервиса, событие, настройку, страницу. Где лежит код каждой области — в [карте кода](code-map.md); правила оформления кода — в [code-style.md](code-style.md); интеграция с рабочим столом — в [desktop.md](desktop.md).

## Части

```text
main.go, desktop.go      точка входа: собирает зависимости, окно, трей
internal/app             сервисы Wails — единственная граница с фронтендом
internal/…               ядро: настройки, тексты Go, фоновая проверка; без Wails
go-desktop-kit           общие пакеты: каталоги, журнал, обновления, связка с Wails
frontend/bindings        TypeScript-привязки сервисов, генерируются из Go
frontend/src             React + MUI: api → store → pages
```

Wails импортируют только пакет `main` и `internal/app`; это проверяет `internal/archtest/wails_test.go`. Ядро в `internal/` тестируется без окна и без фронтенда.

## Порядок запуска

`run()` в `main.go` выполняет шаги по порядку; каждый следующий получает результат предыдущего.

1. Флаги: `--version` (строка `Desktop App <версия> (<ОС>/<арх>; <формат>)` и выход без окна), `--hidden` (запуск без окна, из автозапуска), `--smoke-report <файл>` и `--smoke-timeout` (проверка запуска, по умолчанию 30 с).
2. Формат пакета сборки — `updater.Naming.DetectFormat`: значение `-X main.packageFormat` или файл-метка пакета Linux. Нужен до каталогов, потому что его печатает `--version`.
3. Каталоги — `appdirs.Resolve` по слагу и префиксу окружения (см. «Каталоги и переменные окружения»).
4. Журнал — `applog.Open` в каталоге журнала; при ошибке записи идут в stderr. Заодно определяется, оборвался ли прошлый сеанс.
5. Настройки — `settings.Store.Open`: при первом запуске файл создаётся с умолчаниями.
6. Сервис обновлений — `wailsx.NewUpdatesService` из kit; своего кода обновлений в шаблоне нет.
7. Сервисы `AppService` и `SettingsService` из `internal/app`, сервис уведомлений Wails (кроме smoke-режима).
8. `application.New`: название, иконка, сервисы, встроенный `frontend/dist`, один экземпляр (`wailsx.SingleInstance`).
9. Окно 1280×800, минимум 960×600. `wailsx.RestoreWindow` подставляет сохранённые размер и положение, `wailsx.TrackWindow` запоминает их при закрытии.
10. Связка с рабочим столом — `setupDesktop` в `desktop.go`: трей, автозапуск, уведомления, фоновая проверка обновлений, завершение по SIGINT и SIGTERM. В smoke-режиме вместо неё — таймер проверки.
11. `Run()`. Фронтенд загружает данные и вызывает `AppService.FrontendReady()`.

Smoke-режим (`wails3 task smoke`) проверяет, что окно открылось и фронтенд ответил: после `FrontendReady` или по таймеру пишется отчёт, код возврата ненулевой, если ответа не было.

## Сервисы

Методы сервисов — всё, что фронтенд может попросить у Go. Сервисы в `internal/app` не содержат логики: они переводят вызов в ядро или kit, ответ — в DTO через `wire.Fill` (пустые срезы и карты становятся `[]` и `{}`, а не `null`), ошибку — в `*apperr.Error`.

| Сервис | Метод | Что делает |
| --- | --- | --- |
| `AppService` | `Info()` | Название, версия, формат сборки, ОС и архитектура, каталоги, доступен ли трей |
| | `FrontendReady()` | Первый экран отрисован; в smoke-режиме завершает проверку успехом |
| | `ReportFrontendError(r)` | Записать необработанную ошибку интерфейса в журнал (поля обрезаются до 8 КиБ) |
| | `OpenLogFolder()` | Открыть каталог журнала в файловом менеджере |
| | `PreviousCrash()` | Прошлый сеанс оборвался: время его начала; `null` — завершился штатно |
| `SettingsService` | `Get()` | Настройки и признак `broken` — файл испорчен, показаны умолчания |
| | `Save(s)` | Проверить и сохранить; затем событие `settings:changed` |
| | `Autostart()`, `SetAutostart(on)` | Запуск при входе в систему; состояние читается из ОС |
| `UpdatesService` (kit `wailsx`) | `Status()`, `Check()`, `Download()`, `Cancel()`, `Install()` | Обновление через GitHub Releases; статус приходит и событием `update:status` |

Код сервисов: `internal/app/app.go`, `internal/app/settings.go`. Сервис обновлений описан в документации go-desktop-kit (пакеты `updater` и `wailsx`).

## Контракт Go ↔ TypeScript

```text
internal/app  ──wails3 task bindings──▶  frontend/bindings
                                            │
frontend/src/api/bindings.ts   единственный импорт привязок и @wailsio/runtime
frontend/src/api/source.ts     интерфейс DataSource: всё, что интерфейсу нужно от Go
  ├─ wailsSource.ts            реализация привязками — в окне приложения
  └─ mockSource.ts             данные в памяти — yarn dev в браузере, Playwright
frontend/src/store/appStore.ts стор zustand поверх DataSource
```

- **Привязки.** `frontend/bindings` генерирует `wails3 task bindings` из методов сервисов и зарегистрированных событий. Каталог коммитится и руками не правится; `wails3 task check:bindings` падает, если он отстал от Go (в CI тоже).
- **Один импорт.** Длинные пути привязок встречаются только в `frontend/src/api/bindings.ts`. Остальной код берёт типы и вызовы через `frontend/src/api/index.ts`.
- **`DataSource`.** Интерфейс в `frontend/src/api/source.ts`. Источник выбирает `pickSource()` в `frontend/src/api/index.ts`: в нативном окне — `wailsSource`, в обычном браузере — `mockSource`. Ненужная реализация не загружается.
- **Моки.** `mockSource` принимает начальное состояние из адреса страницы: `/?lang=en&theme=dark&update=ready&format=dmg#/settings`. Полный список параметров — в комментарии `mockOptionsFromSearch` в `frontend/src/api/mockSource.ts`. Так сценарии Playwright задают состояние без правки кода.
- **Стор.** Один стор, создаётся `createAppStore(source)` и передаётся через контекст (`frontend/src/store/context.tsx`). Состояние живёт в Go: стор держит последний снимок, шлёт изменения через `DataSource`, а события Go обновляют снимок. Тесты подставляют `frontend/src/test/fakeSource.ts`.

## Ошибки

Go отдаёт ошибку кодом и параметрами, без текста для человека. Текст собирает интерфейс по словарю.

1. Ядро возвращает обычную ошибку с контекстом (`fmt.Errorf("…: %w", err)`) или `*apperr.Error` с кодом (как `settings.Validate`).
2. Сервис переводит её функцией `toError` из `internal/app/errors.go`: известные ошибки получают свой код, `*apperr.Error` из цепочки проходит как есть, остальные — запасной код и текст в параметре `detail`.
3. Wails кладёт JSON ошибки в поле `cause`; `call()` и `toAppError()` из `frontend/src/api/errors.ts` превращают его в `AppError` с полями `code` и `params`. Любая другая ошибка получает код `unknown`.
4. Интерфейс показывает `t("errors.<код>", params)`.

Все коды, которые Go может отдать фронтенду, перечислены в `frontend/src/i18n/error-codes.json`. Тест `internal/archtest/errorcodes_test.go` сверяет список с константами `Code*` шаблона и кодами `update.*` из kit, а тест словарей фронтенда (`frontend/src/i18n/i18n.test.tsx`) требует текст для каждого кода на обоих языках.

## События

| Событие | Данные | Кто шлёт | Подписка во фронтенде |
| --- | --- | --- | --- |
| `settings:changed` | настройки | `SettingsService.Save` | `onSettingsChanged` |
| `app:navigate` | путь маршрута, например `/settings` | пункт трея, щелчок по уведомлению | `onNavigate` |
| `update:status` | статус обновления | `UpdatesService` из kit | `onUpdateStatus` |

Имена событий шаблона и типы их данных объявлены в `internal/app/events.go` и регистрируются `application.RegisterEvent` в `init()`: из регистрации генератор выводит типы событий для TypeScript. Подписки с отпиской — в `frontend/src/api/events.ts`.

## Каталоги и переменные окружения

Каталоги вычисляет `appdirs` из kit по правилам ОС; сами каталоги создаются при первой записи.

| Каталог | Linux | macOS | Windows |
| --- | --- | --- | --- |
| Настройки | `$XDG_CONFIG_HOME/desktop-app` (`~/.config/desktop-app`) | `~/Library/Application Support/desktop-app` | `%AppData%\desktop-app` |
| Данные | `$XDG_DATA_HOME/desktop-app` (`~/.local/share/desktop-app`) | как настройки | `%LocalAppData%\desktop-app` |
| Кэш | `$XDG_CACHE_HOME/desktop-app` (`~/.cache/desktop-app`) | `~/Library/Caches/desktop-app` | `%LocalAppData%\desktop-app\cache` |
| Журнал | `<кэш>/logs` | `<кэш>/logs` | `<кэш>/logs` |

Переменные окружения:

- `DESKTOP_APP_HOME` — все каталоги в подпапках одного корня: `config`, `data`, `cache`, журнал в `cache/logs`. Так работают `wails3 task dev` (корень `.task/dev-home`) и `wails3 task smoke`: настоящие настройки пользователя не трогаются.
- `DESKTOP_APP_CONFIG_DIR`, `DESKTOP_APP_DATA_DIR`, `DESKTOP_APP_CACHE_DIR` — каталог по одному; важнее `DESKTOP_APP_HOME`.
- Относительный путь и `~` в переопределении отсчитываются от домашнего каталога.

Что где лежит: `settings.json` и `window.json` (размер и положение окна) — в каталоге настроек; скачанные обновления — в `<данные>/updates`; журнал `desktop-app.log` и маркер сеанса — в каталоге журнала.

WebKitGTK на Linux пишет своё хранилище в `~/.cache/desktop-app` и `~/.local/share/desktop-app` мимо `DESKTOP_APP_HOME`; на данные приложения это не влияет.

## Настройки

Схема — `internal/settings/settings.go`; файл — `settings.json` в каталоге настроек. Хранилище kit (`settings.Store`) даёт версию формата, атомарную запись и защиту испорченного файла.

| Поле | Значения | По умолчанию |
| --- | --- | --- |
| `ui.theme` | `light`, `dark`, пусто — как в системе | пусто |
| `ui.language` | `ru`, `en`, пусто — язык системы | пусто |
| `background` | закрытие окна прячет его в трей | выключено |
| `autoCheckUpdates` | проверять обновления при запуске и раз в сутки | выключено: без согласия приложение в сеть не ходит |

- Испорченный файл или файл более новой версии не перезаписывается: `Get()` отдаёт умолчания с признаком `broken`, `Save()` отказывает с кодом `settings.broken` или `settings.newer-version`, пока файл не исправят или не удалят.
- Недопустимое значение при сохранении — ошибка `settings.invalid-theme` или `settings.invalid-language`; при чтении такие значения заменяются умолчаниями.
- Автозапуск в настройках не хранится: его состояние читается из ОС.
- Язык текстов Go (трей, уведомления) — из настройки, иначе язык системы; интерфейс выбирает язык так же, а без подходящего — английский.

## Журнал

Текстовый формат `slog`, файл `desktop-app.log` в каталоге журнала. При размере больше 2 МиБ файл переименовывается в `desktop-app.log.1` (прежний заменяется). Права каталога 0700, файла 0600: в журнале бывают пути и имя пользователя.

- Ошибки интерфейса попадают в журнал через `ReportFrontendError`: их ловят граница ошибок React и обработчики `error` и `unhandledrejection` (`frontend/src/app/ErrorBoundary.tsx`).
- Маркер сеанса рядом с журналом: если прошлый сеанс не закрыл его (сбой, `kill -9`), `PreviousCrash()` об этом сообщает.
- Кнопка «Открыть папку журнала» на странице настроек открывает каталог в файловом менеджере.

## Тесты

| Вид | Где | Что покрывает | Команда |
| --- | --- | --- | --- |
| Go, модульные | `internal/app`, `internal/settings`, `internal/texts`, `internal/updatecheck`, `scripts/icon` | Сервисы с подставными зависимостями, схема настроек, словари Go, расписание фоновой проверки | `wails3 task test:go` |
| Правила репозитория | `internal/archtest` | Wails только в `main` и `internal/app`; имена приложения во всех файлах; коды ошибок; метки пакетов; ключи текстов Go | `wails3 task test:go` |
| Документы | `internal/repodocs` | Ссылки, якоря, пути, задачи и скрипты в Markdown; актуальность карты кода | `wails3 task check:docs` |
| Гонки | `internal/...` | То же ядро под `-race` | `wails3 task check:race` |
| Фронтенд, модульные | `frontend/src/**/*.test.ts(x)` | Ошибки и моки, стор на `fakeSource`, страницы и блок обновлений, словари, тема и контраст, правило «нет текста в JSX» | `wails3 task test:frontend` |
| Сквозные | `frontend/e2e` | Каждая страница в двух темах и на двух языках, переключатели настроек, все состояния блока обновлений, нет ошибок в консоли — в Chromium на моках | `wails3 task test:e2e` |
| Привязки | `frontend/bindings` | Совпадают с Go-сервисами | `wails3 task check:bindings` |
| Запуск | собранный бинарь | Окно открылось, фронтенд ответил | `wails3 task smoke` |

`wails3 task check` запускает всё, кроме сквозных сценариев, гонок, привязок и smoke; их запускает CI (см. ci.md). Браузер для сквозных сценариев ставится один раз: `yarn playwright install chromium --no-shell` в `frontend/`.

## Как добавить

Каждый рецепт кончается `wails3 task check`. Если менялись методы сервисов, события или типы их данных — ещё `wails3 task bindings`, и `frontend/bindings` коммитится вместе с кодом.

### Метод сервиса

На примере метода, который вызывается кнопкой.

1. Метод в сервисе `internal/app`: принимает только данные от фронтенда (не пути: каталоги знает Go), зовёт ядро, возвращает DTO через `wire.Fill` и ошибку через `toError`. Долгая операция (сеть, внешние процессы) принимает первым параметром `context.Context`.
2. Новый код ошибки — константа `Code…` в `internal/app/errors.go`, строка в `frontend/src/i18n/error-codes.json` (коды идут по алфавиту — это проверяет тест), тексты `errors.<код>` в `frontend/src/i18n/ru.ts` и `frontend/src/i18n/en.ts`.
3. Тест метода рядом с сервисом (`internal/app/app_test.go` — образец с подставными зависимостями).
4. `wails3 task bindings`. Новый сервис или новый тип — реэкспорт в `frontend/src/api/bindings.ts`.
5. Метод в интерфейсе `DataSource` (`frontend/src/api/source.ts`) и в трёх реализациях: `wailsSource.ts` (вызов привязки через `call`), `mockSource.ts`, `frontend/src/test/fakeSource.ts`. Компилятор TypeScript покажет, где реализации не хватает.
6. Действие в сторе (`frontend/src/store/appStore.ts`): объявление в `AppState` и реализация; ошибка уходит в уведомление через `run`.
7. Кнопка на странице, текст — ключом словаря в обоих языках. Тест страницы на `fakeSource` проверяет, что нажатие вызывает метод источника (образец — `frontend/src/pages/SettingsPage/SettingsPage.test.tsx`).

Новый сервис, кроме этого, создаётся в `main.go` и добавляется в список `services`.

### Событие Go → фронтенд

1. Константа имени в `internal/app/events.go` и `application.RegisterEvent[T](имя)` в `init()`; данные — тип, который генератор привязок умеет вывести.
2. Отправка — `wailsx.Emit(имя, данные)`; в сервисе — через функцию `Emit` из его опций, чтобы тест мог её подменить (как в `SettingsService`).
3. `wails3 task bindings`; константа события — реэкспорт в `frontend/src/api/bindings.ts`.
4. Подписка в `frontend/src/api/events.ts`, метод `on…` в `DataSource` и реализациях, обработка в сторе (`init` подписывается, `dispose` отписывается).

### Настройка

1. Поле в `Settings` (`internal/settings/settings.go`) с тегом `json`, умолчание в `Defaults`, проверка в `Validate` и `normalize`, если набор значений ограничен (с кодом ошибки — как `settings.invalid-theme`). Тест в `internal/settings/settings_test.go`.
2. Если от настройки зависит Go-сторона (трей, фоновые задачи) — реакция в `settingsChanged` в `desktop.go`.
3. `wails3 task bindings` — тип `Settings` во фронтенде обновится.
4. Умолчание в `defaultSettings` (`frontend/src/store/appStore.ts`) и в `mockSource.ts`; при необходимости — параметр адреса в `mockOptionsFromSearch`.
5. Действие `set…` в сторе, переключатель на странице настроек, тексты в обоих словарях, сценарий Playwright в `frontend/e2e/settings.spec.ts`.

Если старый файл нельзя прочитать новой схемой без преобразования, повышается `FormatVersion`.

### Страница

1. Папка `frontend/src/pages/<Имя>Page/` с компонентом `<Имя>Page.tsx` и `index.ts`. Перед `export function` — комментарий `/** … */`: его первое предложение попадает в карту кода.
2. Маршрут в `frontend/src/app/App.tsx`; вкладка — в списке `tabs` в `frontend/src/app/Shell.tsx`, ключ `nav.…` в обоих словарях.
3. Страница — в `Pages` подходящей области `internal/repodocs/areas.go`, затем `wails3 task docs:map`.
4. Тест страницы на `fakeSource` и сквозной сценарий в `frontend/e2e/`: страница в двух темах и на двух языках, без ошибок в консоли. Правила интерфейса — в [ui.md](ui.md).

Пакет Go добавляется так же: комментарий `// Package`, строка в `areas.go`, `wails3 task docs:map`.
