//go:build !windows

package texts

import "os"

// systemLocale — локаль сообщений из окружения процесса.
func systemLocale() string {
	return localeFromEnv(os.Getenv)
}
