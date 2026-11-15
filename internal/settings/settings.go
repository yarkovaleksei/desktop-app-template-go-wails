// Package settings — схема настроек приложения, их умолчания и проверка.
//
// Файл хранит kit/settings.Store: версия формата, атомарная запись, защита
// испорченного файла. Здесь — только то, что относится к этому приложению.
// Новая настройка добавляется полем в Settings, умолчанием в Defaults и, если
// у неё ограниченный набор значений, проверкой в Validate.
package settings

import (
	"slices"

	"github.com/yarkovaleksei/go-desktop-kit/apperr"
	kitsettings "github.com/yarkovaleksei/go-desktop-kit/settings"
)

// FormatVersion — версия формата файла настроек. Повышается, когда старый
// файл нельзя прочитать новой схемой без преобразования.
const FormatVersion = 1

// Значения темы оформления. Пустая тема — как в системе.
const (
	ThemeSystem = ""
	ThemeLight  = "light"
	ThemeDark   = "dark"
)

// Значения языка интерфейса. Пустой язык — язык системы (D-033).
const (
	LanguageSystem = ""
	LanguageRU     = "ru"
	LanguageEN     = "en"
)

// Коды ошибок проверки настроек. Параметр value — отклонённое значение.
const (
	CodeInvalidTheme    = "settings.invalid-theme"
	CodeInvalidLanguage = "settings.invalid-language"
)

var (
	themes    = []string{ThemeSystem, ThemeLight, ThemeDark}
	languages = []string{LanguageSystem, LanguageRU, LanguageEN}
)

// UI — оформление интерфейса.
type UI struct {
	Theme    string `json:"theme"`    // ThemeSystem, ThemeLight или ThemeDark
	Language string `json:"language"` // LanguageSystem, LanguageRU или LanguageEN
}

// Settings — все настройки приложения.
type Settings struct {
	UI UI `json:"ui"`
	// Background — закрытие окна прячет его в трей, приложение продолжает
	// работать. По умолчанию выключено (D-026).
	Background bool `json:"background"`
	// AutoCheckUpdates — проверять обновления при запуске и раз в сутки.
	// По умолчанию выключено: без согласия приложение в сеть не ходит (D-031).
	AutoCheckUpdates bool `json:"autoCheckUpdates"`
}

// Defaults возвращает настройки первого запуска: тема и язык как в системе,
// работа в фоне и проверка обновлений выключены.
func Defaults() Settings {
	return Settings{}
}

// Validate проверяет значения с ограниченным набором. Ошибка — *apperr.Error
// с кодом CodeInvalidTheme или CodeInvalidLanguage.
func (s Settings) Validate() error {
	if !slices.Contains(themes, s.UI.Theme) {
		return apperr.New(CodeInvalidTheme, "value", s.UI.Theme)
	}
	if !slices.Contains(languages, s.UI.Language) {
		return apperr.New(CodeInvalidLanguage, "value", s.UI.Language)
	}
	return nil
}

// normalize заменяет недопустимые значения умолчаниями. Нужна при чтении:
// файл могли поправить руками, а интерфейс должен получить то, что умеет
// показать.
func (s Settings) normalize() Settings {
	d := Defaults()
	if !slices.Contains(themes, s.UI.Theme) {
		s.UI.Theme = d.UI.Theme
	}
	if !slices.Contains(languages, s.UI.Language) {
		s.UI.Language = d.UI.Language
	}
	return s
}

// Store читает и записывает файл настроек. Методы безопасны для
// одновременного вызова.
type Store struct {
	store *kitsettings.Store[Settings]
}

// New создаёт хранилище файла path. Файл не создаётся до первого Save.
func New(path string) *Store {
	return &Store{store: kitsettings.New(path, FormatVersion, Defaults)}
}

// Load читает настройки. Нет файла — умолчания без ошибки. Непригодный файл —
// умолчания и ошибка с kit/settings.ErrBroken или ErrNewerVersion: Save его
// не перезапишет, пока файл не исправят или не удалят. Недопустимые значения
// из файла заменяются умолчаниями.
func (s *Store) Load() (Settings, error) {
	v, err := s.store.Load()
	return v.normalize(), err
}

// Save проверяет и записывает настройки. Недопустимое значение — ошибка из
// Validate, файл не меняется.
func (s *Store) Save(v Settings) error {
	if err := v.Validate(); err != nil {
		return err
	}
	return s.store.Save(v)
}

// Broken сообщает, что файл настроек непригоден и Save его не перезапишет.
func (s *Store) Broken() bool {
	return s.store.Broken()
}

// Path возвращает путь к файлу настроек.
func (s *Store) Path() string {
	return s.store.Path()
}
