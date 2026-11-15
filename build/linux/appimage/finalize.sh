#!/usr/bin/env bash
# Встроить загрузчик appimage-bootstrap в AppDir, созданный
# wails3 generate appimage, и упаковать образ заново.
#
# Аргументы: AppDir, итоговый файл AppImage, слаг, название, linuxdeploy.
# Слаг входит в имя профиля AppArmor, название — в заголовки окон загрузчика.
#
# Wails пишет в AppDir свой AppRun, который задаёт окружение образа. Он
# становится AppRun.wails, а на его место встаёт build/linux/appimage/AppRun:
# сначала загрузчик, затем AppRun.wails (договор — README загрузчика в kit).
set -euo pipefail
[[ $# = 5 ]] || { echo "Использование: $0 <AppDir> <AppImage> <слаг> <название> <linuxdeploy>" >&2; exit 2; }
appdir=$1
output=$2
slug=$3
title=$4
linuxdeploy=$5
root=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../../.." && pwd)
apprun="$root/build/linux/appimage/AppRun"

[[ -x $appdir/AppRun ]] || { echo "В $appdir нет AppRun: сначала wails3 generate appimage" >&2; exit 1; }
# Повторный запуск на том же AppDir не должен затереть AppRun.wails своим AppRun.
if ! cmp -s "$appdir/AppRun" "$apprun"; then
  mv -- "$appdir/AppRun" "$appdir/AppRun.wails"
fi

mkdir -p "$appdir/usr/libexec"
(cd "$root" && CGO_ENABLED=0 GOOS=linux go build -trimpath -buildvcs=false \
  -ldflags "-s -w -X main.app=$slug -X 'main.title=$title'" \
  -o "$appdir/usr/libexec/appimage-bootstrap" \
  github.com/yarkovaleksei/go-desktop-kit/cmd/appimage-bootstrap)

# linuxdeploy кладёт свой AppRun из --custom-apprun и собирает образ по
# пути OUTPUT. Библиотеки Wails уже разместил; NO_STRIP — как у Wails:
# strip из linuxdeploy не понимает секции .relr.dyn новых тулчейнов.
mkdir -p "$(dirname -- "$output")"
rm -f -- "$output"
case $output in /*) ;; *) output="$PWD/$output" ;; esac
(cd "$(dirname -- "$appdir")" && OUTPUT="$output" NO_STRIP=1 ARCH=x86_64 \
  "$linuxdeploy" --appimage-extract-and-run --appdir "$appdir" \
  --custom-apprun "$apprun" --output appimage)
[[ -f $output ]] || { echo "linuxdeploy не создал $output" >&2; exit 1; }
chmod 755 "$output"
echo "Готово: $output"
