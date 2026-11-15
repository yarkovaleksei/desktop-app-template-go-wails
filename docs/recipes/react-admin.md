# react-admin

## Когда нужно

Интерфейс — в основном таблицы, формы, фильтры и списки записей. Для приложения из нескольких экранов со своей логикой react-admin лишний: хватает MUI и стора шаблона.

## Что добавить

- Зависимости `react-admin` и `ra-i18n-polyglot` с записью решения; версия MUI должна совпасть с той, что требует react-admin.
- `dataProvider` поверх `DataSource`: методы `getList`, `getOne`, `create`, `update`, `delete` вызывают сервисы Go через `frontend/src/api/`, а не привязки напрямую. Моки получают те же методы — сквозные сценарии работают без Go.
- Сервисы Go с постраничной выдачей, сортировкой и фильтром; ответы через `wire.Fill`.
- Словари: `ra-i18n-polyglot` поверх словарей `frontend/src/i18n/ru.ts` и `frontend/src/i18n/en.ts` или рядом с ними; правило «новый ключ в обоих языках» остаётся.
- Тема react-admin — из той же `createAppTheme`; `HashRouter` шаблона заменяет роутер react-admin.

## На что смотреть в образцах

- `ai-memory/frontend/src/dataProvider.ts` — источник данных react-admin поверх сервисов Go.
- `ai-memory/frontend/src/App.tsx` — ресурсы и маршруты.
- `ai-memory/frontend/src/i18n/index.ts` — подключение словарей через `ra-i18n-polyglot`.

## Как проверить

- Тесты `dataProvider` на `fakeSource`: список с сортировкой и фильтром, ошибка Go приходит в react-admin с кодом.
- Сквозной сценарий ресурса: список, создание, правка, удаление — в двух темах и на двух языках.
