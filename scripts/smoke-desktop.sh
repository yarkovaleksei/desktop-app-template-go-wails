#!/usr/bin/env bash
# Запуск окна так, как его запускает меню приложений: через пользовательский
# systemd, вне окружения и профиля AppArmor процесса агента или терминала.
# Из терминала Claude Code или Codex окно наследует их профиль с правом на user
# namespaces и открывается даже там, где из меню падает (WebKitGTK запускает
# bwrap). Окно на несколько секунд появится на экране. Каталоги приложения —
# во временном DESKTOP_APP_HOME: настройки пользователя не трогаются.
#
#   bash scripts/smoke-desktop.sh [путь к бинарю] [секунды]   (по умолчанию /usr/bin/desktop-app, 8)
#
# Только Linux.
set -euo pipefail
cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.."

slug="desktop-app"
bin="$(realpath -- "${1:-/usr/bin/$slug}")"
secs="${2:-8}"
[[ "$(uname -s)" == Linux ]] || { echo "Проверка запуска как из меню есть только для Linux" >&2; exit 1; }
[[ -x "$bin" ]] || { echo "Нет исполняемого $bin" >&2; exit 1; }
[[ "$secs" =~ ^[0-9]+$ ]] || { echo "Секунды — целое число, а не «$secs»" >&2; exit 1; }
command -v systemd-run > /dev/null || { echo "Нужен systemd-run (пользовательский systemd)" >&2; exit 1; }

home="$(mktemp -d "${TMPDIR:-/tmp}/$slug-smoke-desktop.XXXXXX")"
trap 'rm -rf -- "$home"' EXIT

unit="$slug-smoke-$$"
set +e
DESKTOP_APP_HOME="$home" systemd-run --user --wait --collect --quiet --unit "$unit" \
  -E DESKTOP_APP_HOME \
  timeout "$secs" "$bin"
code=$?
set -e

# timeout завершает живое окно с кодом 124: значит, оно проработало весь срок.
if [[ $code -eq 124 ]]; then
  echo "OK: $bin проработал $secs с при запуске как из меню"
  exit 0
fi
echo "ПРОВАЛ: $bin завершился с кодом $code раньше $secs с" >&2
journalctl --user -u "$unit" -o cat --no-pager 2> /dev/null | grep -m 5 -E "SIGTRAP|SIGSEGV|panic|dbus-proxy|Gtk-|error" >&2 || true
echo "Если в журнале ядра есть apparmor=\"DENIED\" … comm=\"bwrap\", не загружен профиль /etc/apparmor.d/$slug из пакета deb." >&2
exit 1
