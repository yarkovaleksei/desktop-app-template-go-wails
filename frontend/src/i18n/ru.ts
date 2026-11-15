// Русский словарь — эталон: по нему типизированы ключи, en.ts обязан иметь тот же набор.
// Ключи плоские; «{name}» заменяется параметром. Множественное число — объект
// с формами Intl.PluralRules, форма выбирается по параметру count.
// Тексты ошибок — ключи errors.<код>: коды отдаёт Go (список — error-codes.json).
import type { Message } from "./index";

export const ru = {
  "nav.home": "Главная",
  "nav.settings": "Настройки",
  "nav.label": "Разделы",

  "home.intro": "Заготовка главной страницы. Замените её интерфейсом своего приложения.",
  "home.docs": "Документация проекта",
  "home.uptime": {
    one: "Окно открыто {count} минуту",
    few: "Окно открыто {count} минуты",
    many: "Окно открыто {count} минут",
    other: "Окно открыто {count} минуты",
  },

  "settings.title": "Настройки",
  "settings.appearance": "Оформление",
  "settings.language": "Язык",
  "settings.language.system": "Как в системе",
  "settings.language.ru": "Русский",
  "settings.language.en": "English",
  "settings.theme": "Тема",
  "settings.theme.system": "Как в системе",
  "settings.theme.light": "Светлая",
  "settings.theme.dark": "Тёмная",
  "settings.behavior": "Работа приложения",
  "settings.background": "Работать в фоне при закрытии окна",
  "settings.background.noTray":
    "Недоступно: в этой системе нет области уведомлений (трея), свёрнутое окно было бы не вернуть.",
  "settings.autostart": "Запускать при входе в систему",
  "settings.autoCheck": "Проверять обновления автоматически",
  "settings.autoCheck.help":
    "При запуске и раз в сутки. Пока настройка выключена, приложение само в сеть не обращается.",
  "settings.broken":
    "Файл настроек повреждён или записан более новой версией. Показаны значения по умолчанию, сохранить изменения не получится.",
  "settings.about": "О программе",
  "settings.version": "Версия {version}",
  "settings.format": "Формат поставки: {format}",
  "settings.format.none": "сборка из исходников",
  "settings.openLogs": "Открыть папку журнала",

  "updates.title": "Обновления",
  "updates.unsupported": "Эта сборка не из пакета и обновляется вручную.",
  "updates.check": "Проверить обновления",
  "updates.checking": "Проверка…",
  "updates.upToDate": "Установлена последняя версия {version}",
  "updates.available": "Доступна версия {version}, размер {size}",
  "updates.notes": "Что нового",
  "updates.releases": "Выпуски на GitHub",
  "updates.download": "Скачать",
  "updates.downloading": "Скачано {received} из {size}",
  "updates.cancel": "Отменить",
  "updates.retry": "Повторить",
  "updates.ready.command":
    "Пакет {version} скачан и сверен. Установите его командой ниже — понадобится пароль администратора — и перезапустите приложение.",
  "updates.ready.open":
    "Пакет {version} скачан и сверен. Откройте образ диска и перетащите приложение в «Программы».",
  "updates.ready.run": "Пакет {version} скачан и сверен. Установщик заменит приложение, окно закроется.",
  "updates.ready.replace":
    "Пакет {version} скачан и сверен. Файл AppImage заменится новым, приложение перезапустится.",
  "updates.ready.reveal":
    "Пакет {version} скачан и сверен. Закройте приложение и замените его файл скачанным.",
  "updates.install.command": "Открыть папку с пакетом",
  "updates.install.open": "Открыть образ диска",
  "updates.install.run": "Запустить установщик",
  "updates.install.replace": "Установить и перезапустить",
  "updates.install.reveal": "Показать файл",
  "updates.copy": "Скопировать команду",
  "updates.copied": "Команда скопирована",

  "crash.title": "Что-то пошло не так",
  "crash.body": "В интерфейсе произошла ошибка, отчёт записан в журнал.",
  "crash.reload": "Перезагрузить интерфейс",
  "crash.previous": "Прошлый сеанс, начатый {at}, завершился неожиданно.",
  "crash.previousUnknown": "Прошлый сеанс завершился неожиданно.",

  "errors.unknown": "Неизвестная ошибка: {detail}",
  "errors.app.open-log-folder-failed": "Не удалось открыть папку журнала: {detail}",
  "errors.autostart.failed": "Не удалось изменить запуск при входе в систему: {detail}",
  "errors.autostart.unavailable": "Запуск при входе в систему в этой системе недоступен",
  "errors.settings.broken": "Файл настроек повреждён: изменения не сохранятся, пока его не исправят",
  "errors.settings.newer-version":
    "Файл настроек записан более новой версией приложения: изменения не сохранятся",
  "errors.settings.write-failed": "Не удалось сохранить настройки: {detail}",
  "errors.settings.invalid-theme": "Неизвестная тема «{value}»",
  "errors.settings.invalid-language": "Неизвестный язык «{value}»",
  "errors.update.bad-release": "GitHub вернул непонятный ответ о выпуске",
  "errors.update.canceled": "Скачивание отменено",
  "errors.update.host": "Загрузка ушла на посторонний адрес и прервана",
  "errors.update.http": "GitHub ответил ошибкой {status}",
  "errors.update.launch": "Не удалось запустить установку — поставьте пакет вручную",
  "errors.update.mismatch": "Пакет не совпал с контрольной суммой и удалён — попробуйте скачать снова",
  "errors.update.network": "Не удалось связаться с GitHub — проверьте подключение к сети",
  "errors.update.no-package": "В выпуске {version} нет пакета {package} — обновите приложение вручную",
  "errors.update.no-release": "На GitHub нет подходящего опубликованного выпуска",
  "errors.update.no-sums": "В выпуске нет контрольной суммы пакета ({file}), без неё пакет не скачивается",
  "errors.update.not-ready": "Пакет ещё не скачан",
  "errors.update.rate-limited": "GitHub временно ограничил запросы — попробуйте через час",
  "errors.update.size": "Размер пакета не совпал с заявленным или превышает допустимый",
  "errors.update.unknown": "Обновление не удалось: {detail}",
  "errors.update.write": "Не удалось сохранить пакет на диск",
} as const satisfies Record<string, Message>;

/** Ключ словаря. */
export type MessageKey = keyof typeof ru;

/** Словарь любого языка: те же ключи, у множественного числа — свои формы. */
export type Dictionary = {
  readonly [K in MessageKey]: (typeof ru)[K] extends string ? string : Exclude<Message, string>;
};
