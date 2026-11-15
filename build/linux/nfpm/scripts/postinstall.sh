#!/bin/sh

# Профиль AppArmor (только в deb): загрузить сразу, чтобы окно запускалось из
# меню без перезагрузки. Где AppArmor нет или он выключен, ничего не делаем.
if [ -f /etc/apparmor.d/desktop-app ] && command -v apparmor_parser >/dev/null 2>&1 &&
  [ "$(cat /sys/module/apparmor/parameters/enabled 2>/dev/null)" = "Y" ]; then
  apparmor_parser -r -T -W /etc/apparmor.d/desktop-app ||
    echo "Warning: AppArmor profile for desktop-app was not loaded." >&2
fi

# Обновить базу ярлыков: приложение появится в меню сразу.
if command -v update-desktop-database >/dev/null 2>&1; then
  update-desktop-database -q /usr/share/applications
fi

exit 0
