//go:build windows

package texts

import (
	"syscall"
	"unsafe"
)

// localeNameMaxLength — LOCALE_NAME_MAX_LENGTH из winnls.h.
const localeNameMaxLength = 85

var procGetUserDefaultLocaleName = syscall.NewLazyDLL("kernel32.dll").NewProc("GetUserDefaultLocaleName")

// systemLocale — локаль пользователя Windows вида ru-RU. Переменные LANG и LC_*
// там обычно не заданы, поэтому окружение не читается. Ошибка — пустая строка.
func systemLocale() string {
	buf := make([]uint16, localeNameMaxLength)
	n, _, _ := procGetUserDefaultLocaleName.Call(uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if n == 0 {
		return ""
	}
	return syscall.UTF16ToString(buf)
}
