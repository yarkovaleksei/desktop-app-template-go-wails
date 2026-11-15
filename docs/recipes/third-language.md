# Третий язык интерфейса

## Когда нужно

Кроме русского и английского нужен ещё язык. Тексты живут в трёх местах: словари фронтенда, тексты Go (трей, уведомления, автозапуск) и сайт.

## Что добавить

- **Фронтенд.** Словарь `frontend/src/i18n/<код>.ts` с тем же набором ключей, что `frontend/src/i18n/ru.ts` (тип ключей не даст пропустить); выбор языка и формы множественного числа — в `frontend/src/i18n/index.ts`; локаль MUI — в `createAppTheme` (`frontend/src/theme/theme.ts`); кнопка языка на странице настроек и её подпись во всех словарях.
- **Go.** Константа языка и проверка в `internal/settings/settings.go` (`languages`, `Validate`), словарь в `internal/texts/dictionaries.go`, выбор языка системы в `internal/texts/texts.go`.
- **Коды ошибок.** Текст каждого кода из `frontend/src/i18n/error-codes.json` в новом словаре.
- **Сайт.** Тексты нового языка в `www/index.html` и переключатель.
- `wails3 task bindings`, если менялись типы настроек.

## На что смотреть в образцах

Третьего языка нет ни в одном образце. Устройство словарей с типизированными ключами — `forktopus/frontend/src/i18n/index.ts`; подключение списка языков через `ra-i18n-polyglot` — `ai-memory/frontend/src/i18n/index.ts`.

## Как проверить

- `wails3 task check`: типы и тест словарей фронтенда (`frontend/src/i18n/i18n.test.tsx`) требуют полный набор ключей, тесты `internal/texts` и `internal/archtest/texts_test.go` — тексты Go на всех языках.
- Сквозной сценарий страниц на новом языке в `frontend/e2e/pages.spec.ts`; меню трея меняет язык без перезапуска.
