#!/usr/bin/env bash
# Положить в каталог $1 linuxdeploy закреплённой версии и проверить его SHA-256.
#
# wails3 generate appimage берёт linuxdeploy-x86_64.AppImage из своего
# каталога сборки, а если его там нет — скачивает сборку continuous без
# проверки. Этот скрипт запускается до него, поэтому Wails получает уже
# проверенный файл. Новую версию брать вместе с её суммой: digest вложения
# в выпуске https://github.com/linuxdeploy/linuxdeploy/releases.
set -euo pipefail
[[ $# = 1 ]] || { echo "Использование: $0 <каталог сборки AppImage>" >&2; exit 2; }
[[ $(uname -m) = x86_64 ]] || { echo "AppImage собирается только на x86_64" >&2; exit 1; }

version=1-alpha-20251107-1
sha256=c20cd71e3a4e3b80c3483cef793cda3f4e990aca14014d23c544ca3ce1270b4d
url="https://github.com/linuxdeploy/linuxdeploy/releases/download/$version/linuxdeploy-x86_64.AppImage"

dir=$1
file="$dir/linuxdeploy-x86_64.AppImage"
mkdir -p "$dir"
# Файл от прежней версии скрипта или чужой continuous не проходит сверку и
# скачивается заново.
if [[ -f $file ]] && ! printf '%s  %s\n' "$sha256" "$file" | sha256sum --check --status; then
  rm -f -- "$file"
fi
if [[ ! -f $file ]]; then
  echo "Скачивается linuxdeploy $version"
  curl --fail --location --silent --show-error --retry 2 "$url" -o "$file.part"
  if ! printf '%s  %s\n' "$sha256" "$file.part" | sha256sum --check --status; then
    rm -f -- "$file.part"
    echo "SHA-256 linuxdeploy $version не совпадает с закреплённой: $url" >&2
    exit 1
  fi
  mv -- "$file.part" "$file"
fi
chmod 755 "$file"
