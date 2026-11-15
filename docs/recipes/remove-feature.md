# Как убрать ненужную возможность

## Когда нужно

Проекту не нужны трей, автозапуск или AppImage, и лишний код только мешает. Убирать стоит целиком — с настройками, текстами, тестами и документами, — а не оставлять выключенным.

## Что убрать

Как устроена каждая возможность рабочего стола и что именно убрать, написано в разделе «Как убрать» каждой возможности в [desktop.md](../desktop.md). Кратко:

- **Трей и работа в фоне.** Вызов `wailsx.SetupTray` в `desktop.go`, настройка `background` (Go, фронтенд, моки, сценарии), тексты трея в `internal/texts`, значки `build/trayicon.png` и `build/trayicon-template.png`. Без трея убрать и `--hidden`.
- **Автозапуск.** Менеджер в `setupDesktop`, методы `Autostart` и `SetAutostart` в `internal/app/settings.go`, их коды ошибок, переключатель на странице настроек, текст `AutostartComment`.
- **AppImage.** Формат `appimage` из `RELEASE_FORMATS` в `Taskfile.yml`; задача `create:appimage` в `build/linux/Taskfile.yml`; каталог `build/linux/appimage/`; строка AppImage в таблице установки `README.md` и `README.en.md`, на сайте и в `docs/packaging.md`. Загрузчик `appimage-bootstrap` из kit тогда не нужен.
- Уведомления и фоновая проверка обновлений убираются вместе: фоновой проверке нечем сообщить о версии.

После удаления: `wails3 task bindings`, если менялись сервисы; `internal/repodocs/areas.go` и `wails3 task docs:map`, если удалён пакет; строки возможности в документах `docs/`.

## На что смотреть в образцах

- `forktopus/internal/desktop/tray.go` и `forktopus/internal/desktop/autostart.go` — трей и автозапуск, подключённые из `forktopus/main.go`: видно, сколько мест затрагивает возможность.
- `ai-memory/build/linux/Taskfile.yml` — проект без AppImage: пакеты Linux только deb, rpm и Arch.

## Как проверить

- `wails3 task check` и `wails3 task check:bindings`: тесты имён, форматов и меток (`internal/archtest`) ловят рассогласование списка форматов с задачами и документами; проверка документов — ссылки на удалённое.
- `wails3 task build` и `wails3 task smoke`; для удалённого формата — `wails3 task package` без него и `wails3 task release:collect APP_VERSION=0.0.1`.
- Вживую: закрытие окна завершает приложение (без трея), на странице настроек нет переключателя удалённой возможности.
