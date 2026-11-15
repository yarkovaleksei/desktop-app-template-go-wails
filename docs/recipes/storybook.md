# Storybook

## Когда нужно

Компонентов много, у них много состояний, и смотреть их через всё приложение долго: нужен стенд, где каждое состояние открывается отдельно и проверяется снимком или сценарием.

## Что добавить

- Зависимости Storybook для React и Vite с записью решения; конфигурация в каталоге `.storybook` внутри `frontend/`.
- Декоратор с темой (`createAppTheme`, обе темы), словарём (`I18nProvider`, оба языка) и стором на `fakeSource` — так истории не зовут Go.
- Истории рядом с компонентами: `<Компонент>.stories.tsx`. Правило «нет текста в JSX» на истории не распространяется, как и на тесты, — добавьте их в исключения `frontend/eslint.config.js`.
- Скрипты `storybook` и `build-storybook` в `frontend/package.json`; каталог сборки — в `.gitignore`, `frontend/.prettierignore` и исключения ESLint.
- Если истории проверяются Playwright — отдельный проект в `frontend/playwright.config.ts` с сервером Storybook.

## На что смотреть в образцах

- `songframe/frontend/.storybook/` — конфигурация и декораторы.
- `songframe/frontend/src/renderer/pages/MontagePage/MontagePage.stories.tsx` — истории страницы.
- `songframe/frontend/package.json` — скрипты и порт Storybook; `songframe/frontend/playwright.config.ts` — проверка историй.

## Как проверить

- Скрипт `build-storybook` в `frontend/` собирает стенд без ошибок; в Storybook каждая история открывается в обеих темах и на обоих языках.
- `wails3 task check`: ESLint, типы и Prettier не ругаются на истории и сборку Storybook.
