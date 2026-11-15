// Package appinfo — имена приложения. Единственное место, где они записаны в
// Go-коде: остальные пакеты берут их отсюда.
//
// Значения — заглушки шаблона. Инициализация нового проекта (scripts/init)
// заменяет их на имена проекта вместе с теми же строками в остальных файлах.
package appinfo

const (
	// Name — название продукта: заголовок окна, меню, строка версии.
	Name = "Desktop App"
	// Description — описание в одну строку.
	Description = "Desktop application on Go, Wails 3, React and MUI"
	// Slug — имя бинаря, каталогов приложения и пакетов: строчные латинские
	// буквы, цифры, дефис.
	Slug = "desktop-app"
	// Repo — репозиторий GitHub «владелец/имя», из выпусков которого приходят
	// обновления.
	Repo = "yarkovaleksei/desktop-app-template-go-wails"
	// ProductID — обратное доменное имя продукта: идентификатор пакета macOS,
	// установщика Windows и одного экземпляра.
	ProductID = "io.github.yarkovaleksei.desktopapp"
	// EnvPrefix — префикс переменных окружения: DESKTOP_APP_HOME,
	// DESKTOP_APP_CONFIG_DIR и другие (пакет appdirs kit).
	EnvPrefix = "DESKTOP_APP"
)
