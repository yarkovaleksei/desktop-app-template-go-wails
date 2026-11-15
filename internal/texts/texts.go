// Package texts — тексты, которые показывает Go-часть приложения: пункты меню
// трея, уведомления, название записи автозапуска.
//
// Тексты интерфейса живут в словаре фронтенда; здесь только то, что строится
// в Go и не может взять строку оттуда (D-033). Словари ru и en полны: набор
// ключей у них один, это проверяет тест.
package texts

import (
	"slices"
	"strings"
)

// Поддерживаемые языки. Fallback — язык, если ни настройка, ни система не
// дали поддерживаемого.
const (
	RU       = "ru"
	EN       = "en"
	Fallback = EN
)

// Languages возвращает поддерживаемые языки.
func Languages() []string {
	return []string{RU, EN}
}

// T возвращает текст ключа key на языке lang и подставляет параметры:
// «{name}» в тексте заменяется на params["name"]. Неизвестный язык — Fallback,
// неизвестный ключ — сам ключ, чтобы пропуск был виден, а не пуст.
func T(lang, key string, params map[string]string) string {
	d, ok := dictionaries[lang]
	if !ok {
		d = dictionaries[Fallback]
	}
	s, ok := d[key]
	if !ok {
		return key
	}
	for k, v := range params {
		s = strings.ReplaceAll(s, "{"+k+"}", v)
	}
	return s
}

// Language выбирает язык текстов: настройка setting, если она задана и
// поддерживается, иначе язык системы, иначе Fallback (D-033).
func Language(setting string) string {
	return choose(setting, systemLocale())
}

// choose — Language с явной локалью системы, для тестов.
func choose(setting, locale string) string {
	if slices.Contains(Languages(), setting) {
		return setting
	}
	if l := fromLocale(locale); l != "" {
		return l
	}
	return Fallback
}

// fromLocale извлекает поддерживаемый язык из локали вида ru_RU.UTF-8,
// ru-RU или ru. Неподдерживаемая локаль, C и POSIX — пустая строка.
func fromLocale(locale string) string {
	l, _, _ := strings.Cut(locale, ".")
	l, _, _ = strings.Cut(l, "@")
	l = strings.ReplaceAll(l, "-", "_")
	l, _, _ = strings.Cut(l, "_")
	l = strings.ToLower(l)
	if slices.Contains(Languages(), l) {
		return l
	}
	return ""
}

// localeFromEnv читает локаль сообщений из окружения в порядке POSIX:
// LC_ALL, затем LC_MESSAGES, затем LANG; действует первая непустая.
func localeFromEnv(getenv func(string) string) string {
	for _, k := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if v := getenv(k); v != "" {
			return v
		}
	}
	return ""
}
