package texts

import (
	"maps"
	"slices"
	"testing"
)

func TestDictionariesHaveSameKeys(t *testing.T) {
	if got := slices.Sorted(maps.Keys(dictionaries)); !slices.Equal(got, slices.Sorted(slices.Values(Languages()))) {
		t.Fatalf("словари %v не совпадают с языками %v", got, Languages())
	}
	want := slices.Sorted(maps.Keys(dictionaries[Fallback]))
	for _, lang := range Languages() {
		got := slices.Sorted(maps.Keys(dictionaries[lang]))
		if !slices.Equal(got, want) {
			t.Errorf("ключи %s: %v, ключи %s: %v", lang, got, Fallback, want)
		}
		for k, v := range dictionaries[lang] {
			if v == "" {
				t.Errorf("%s: пустой текст ключа %s", lang, k)
			}
		}
	}
}

func TestT(t *testing.T) {
	params := map[string]string{"version": "1.2.3"}
	cases := []struct {
		lang, key string
		params    map[string]string
		want      string
	}{
		{RU, TrayQuit, nil, "Выйти"},
		{EN, TrayQuit, nil, "Quit"},
		{RU, UpdateAvailableBody, params, "Версия 1.2.3 готова к установке. Нажмите, чтобы открыть настройки."},
		{EN, UpdateAvailableBody, params, "Version 1.2.3 is ready to install. Click to open settings."},
		{"de", TrayQuit, nil, "Quit"},
		{RU, "no.such.key", nil, "no.such.key"},
	}
	for _, c := range cases {
		if got := T(c.lang, c.key, c.params); got != c.want {
			t.Errorf("T(%q, %q) = %q, ожидалось %q", c.lang, c.key, got, c.want)
		}
	}
}

func TestChoose(t *testing.T) {
	cases := []struct {
		name, setting, env string
		want               string
	}{
		{"русская система", "", "ru_RU.UTF-8", RU},
		{"английская система", "", "en_US.UTF-8", EN},
		{"немецкая система", "", "de_DE.UTF-8", EN},
		{"пустое окружение", "", "", EN},
		{"локаль C", "", "C", EN},
		{"локаль Windows", "", "ru-RU", RU},
		{"модификатор локали", "", "ru_RU@euro", RU},
		{"настройка важнее системы", EN, "ru_RU.UTF-8", EN},
		{"настройка на немецкой системе", RU, "de_DE.UTF-8", RU},
		{"неизвестная настройка", "de", "ru_RU.UTF-8", RU},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			env := map[string]string{"LANG": c.env}
			if got := choose(c.setting, localeFromEnv(func(k string) string { return env[k] })); got != c.want {
				t.Errorf("choose(%q, LANG=%q) = %q, ожидалось %q", c.setting, c.env, got, c.want)
			}
		})
	}
}

func TestLocaleFromEnvOrder(t *testing.T) {
	cases := []struct {
		name string
		env  map[string]string
		want string
	}{
		{"LC_ALL важнее всех", map[string]string{"LC_ALL": "ru_RU.UTF-8", "LC_MESSAGES": "en_US.UTF-8", "LANG": "de_DE.UTF-8"}, "ru_RU.UTF-8"},
		{"LC_MESSAGES важнее LANG", map[string]string{"LC_MESSAGES": "ru_RU.UTF-8", "LANG": "en_US.UTF-8"}, "ru_RU.UTF-8"},
		{"только LANG", map[string]string{"LANG": "en_US.UTF-8"}, "en_US.UTF-8"},
		{"пусто", map[string]string{}, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := localeFromEnv(func(k string) string { return c.env[k] }); got != c.want {
				t.Errorf("localeFromEnv = %q, ожидалось %q", got, c.want)
			}
		})
	}
}
