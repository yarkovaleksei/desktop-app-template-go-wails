# Поставки

Документ описывает семь форматов поставки: что где собирается, как устроен каждый пакет, как обновляется приложение, установленное из него, и как проверить пакет. Порядок выпуска — в [release.md](release.md), сборка в CI — в [ci.md](ci.md).

## Форматы

| ОС | Формат | Файл выпуска (версия 1.2.3) | Чем собирается | Обновление из приложения |
| --- | --- | --- | --- | --- |
| Linux amd64 | deb | `desktop-app_1.2.3_amd64.deb` | `wails3 tool package` по `build/linux/nfpm/nfpm.yaml` | показывает файл и команду `sudo apt install …`; выполняет её пользователь |
| Linux amd64 | rpm | `desktop-app-1.2.3.x86_64.rpm` | то же | показывает команду `sudo dnf install …` |
| Linux amd64 | пакет Arch | `desktop-app-1.2.3-x86_64.pkg.tar.zst` | то же | показывает команду `sudo pacman -U …` |
| Linux amd64 | AppImage | `desktop-app-1.2.3-x86_64.AppImage` | `build/linux/appimage/linuxdeploy.sh`, `wails3 generate appimage`, `build/linux/appimage/finalize.sh` | заменяет файл и перезапускается после выхода |
| Windows amd64 | установщик NSIS | `desktop-app-1.2.3-windows-amd64-setup.exe` | задача Wails `create:nsis:installer` | запускает установщик, приложение выходит |
| Windows amd64 | переносимый exe | `desktop-app-1.2.3-windows-amd64.exe` | та же сборка с другим форматом | показывает папку со скачанным файлом; заменяет его пользователь |
| macOS universal | dmg | `desktop-app-1.2.3-macos-universal.dmg` | задачи Wails `build:universal`, `create:app:bundle`, `create:dmg` | открывает образ в Finder |

Имена файлов выпуска задаёт `updater.Naming` из go-desktop-kit; полный состав печатает `go tool relkit names -app desktop-app -version 1.2.3`. Список форматов проекта — переменная `RELEASE_FORMATS` в `Taskfile.yml`, единственное место, где он записан.

Обновление скачивает пакет того формата, из которого установлена сборка, и сверяет его SHA-256 с файлом сумм выпуска (`SHA256SUMS-linux.txt`, `SHA256SUMS-windows.txt`, `SHA256SUMS-macos.txt`) и с `digest` вложения на GitHub. Сборка из исходников (`wails3 task build`) формата не знает и обновляться из приложения не умеет. Подробности и коды ошибок — в документации go-desktop-kit (пакет `updater`).

## Где что собирается

`wails3 task package` собирает все пакеты текущей ОС в `bin/`; кросс-сборки нет.

| Где | Что получается | Что нужно |
| --- | --- | --- |
| Linux x86_64 | deb, rpm, пакет Arch, AppImage | Go, Node 24, yarn (`corepack enable`), `wails3`, GTK 4 и WebKitGTK 6.0 для сборки (`libgtk-4-dev`, `libwebkitgtk-6.0-dev`) |
| Windows | установщик и переносимый exe | Go, Node 24, yarn, `wails3`, NSIS (`makensis` в `PATH`) |
| macOS | dmg с универсальным `.app` | Go, Node 24, yarn, `wails3`, Xcode Command Line Tools |

Сборка выпуска: `-tags production -trimpath -buildvcs=false` и `-ldflags "-w -s -X main.version=<версия>"`. Версия — из последнего тега `vX.Y.Z` (`wails3 task version`), переопределяется `APP_VERSION=1.2.3`. Установщик Windows и `Info.plist` принимают только числовую часть `X.Y.Z`.

После `package` команда `wails3 task release:collect APP_VERSION=1.2.3` раскладывает пакеты из `bin/` в `dist/` под именами выпуска и пишет файл сумм своей ОС. Так же собирает выпуск CI.

## Как пакет знает свой формат

- **deb, rpm, Arch.** Бинарь один на три пакета; формат лежит в файле-метке `/usr/share/desktop-app/package-format` (`deb`, `rpm` или `archlinux`), его ставит пакет. Метки — `build/linux/nfpm/package-format/`.
- **AppImage, NSIS, exe, dmg.** Формат вшит при сборке: `-X main.packageFormat=<формат>`; переменная `PACKAGE_FORMAT` в задачах `build/linux/Taskfile.yml`, `build/windows/Taskfile.yml`, `build/darwin/Taskfile.yml`.

`desktop-app --version` печатает формат последним полем: `Desktop App 1.2.3 (linux/amd64; deb)`; `source` — сборка из исходников. Согласованность меток и форматов проверяет `internal/archtest/packaging_test.go`.

## Linux

### deb, rpm, пакет Arch

`build/linux/nfpm/nfpm.yaml`: бинарь в `/usr/bin/desktop-app`, ярлык `build/linux/desktop-app.desktop` в `/usr/share/applications`, иконка в `hicolor/512x512`, метка формата. Зависимости — GTK 4 и WebKitGTK 6.0 (`libgtk-4-1`, `libwebkitgtk-6.0-4` для deb). Скрипты `build/linux/nfpm/scripts/postinstall.sh` и `build/linux/nfpm/scripts/preremove.sh` обновляют базу ярлыков и загружают или выгружают профиль AppArmor.

Иконку окна на Linux даёт установленный ярлык: Wails v3.0.0 на GTK 4 иконку из бинаря не применяет, поэтому окно, запущенное из `bin/`, иконки не имеет.

### AppArmor

В Ubuntu 23.10 и новее AppArmor запрещает программам без профиля создавать user namespace, а WebKitGTK запускает в нём свою песочницу. Без профиля окно падает при старте. Песочница не отключается ни в одном из форматов.

- **deb** ставит профиль `build/linux/apparmor/desktop-app` в `/etc/apparmor.d/desktop-app` (как conffile). Профиль ничего не ограничивает и разрешает только `userns`; местные дополнения — `/etc/apparmor.d/local/desktop-app`. `apt purge` удаляет и файл профиля, `apt remove` только выгружает его.
- **rpm и Arch** профиля не ставят: Fedora живёт на SELinux, в Arch AppArmor по умолчанию выключен. Где AppArmor с ограничением включён вручную, профиль ставится так же, как для deb.
- **AppImage** несёт загрузчик `appimage-bootstrap` из kit. При запуске он проверяет, может ли песочница создать user namespace, и если нет — с согласия пользователя ставит профиль для пути этого файла AppImage (`desktop-app-appimage-<sha256 пути>`) через `pkexec`, затем перезапускает образ. Без окна согласия (`zenity`) и при запуске с `--hidden` печатает команду для ручной установки. Поставить профиль заранее: `./desktop-app-1.2.3-x86_64.AppImage --setup-apparmor`. Перенесённый в другое место файл получает новый профиль.
- **Сборка из исходников** под Ubuntu: профиль для пути `bin/desktop-app` загружается вручную — так делает CI перед smoke (см. [ci.md](ci.md#check)).

`wails3 task smoke:desktop` запускает установленный бинарь как из меню, через пользовательский systemd, и проверяет, что окно проработало несколько секунд.

### AppImage

`build/linux/appimage/linuxdeploy.sh` кладёт linuxdeploy закреплённой версии с проверкой SHA-256 в каталог сборки до `wails3 generate appimage`: без этого Wails скачал бы последнюю версию без проверки. Бинарь AppImage собирается отдельно, в `build/linux/appimage/build`, с вшитым форматом `appimage`. `build/linux/appimage/finalize.sh` переименовывает `AppRun` Wails в `AppRun.wails`, кладёт загрузчик kit и `build/linux/appimage/AppRun` и упаковывает образ заново. Сборка — только на Linux x86_64.

## Windows

Установщик NSIS ставит приложение без прав администратора (`INSTALL_SCOPE=user`), в ресурсах — числовая версия. Переносимый exe — один файл без установки. Сценарий установщика — `build/windows/nsis/project.nsi`, манифест и сведения о файле — `build/windows/`. Запуск окна на Windows проверяет smoke в CI; установка и обновление на Windows не проверены.

## macOS

Универсальный `desktop-app.app` (arm64 и amd64) с ad-hoc подписью в образе dmg; минимальная версия macOS — 12.0. Версия записывается в `CFBundleShortVersionString` и `CFBundleVersion` при сборке; шаблоны — `build/darwin/Info.plist` и `build/darwin/Info.dev.plist`. Собирается только на macOS. Запуск окна на macOS проверяет smoke в CI; установка из dmg и обновление не проверены.

## Подписи нет

Пакеты не подписаны: ни ключом deb и rpm, ни сертификатом Windows, ни Developer ID с нотаризацией Apple (ad-hoc подпись macOS не в счёт). Целостность при обновлении из приложения обеспечивает сверка SHA-256 с файлом сумм выпуска и `digest` GitHub; доверие — к владельцу репозитория. Подпись и нотаризация — отдельный рецепт.

Как открыть неподписанный пакет:

- **Windows.** SmartScreen показывает «Система Windows защитила ваш компьютер»: «Подробнее» → «Выполнить в любом случае». Браузер может предупредить о редко скачиваемом файле — «Сохранить».
- **macOS.** Gatekeeper не даёт открыть приложение неизвестного разработчика: открыть его через контекстное меню Finder → «Открыть» → «Открыть», или «Системные настройки» → «Конфиденциальность и безопасность» → «Всё равно открыть». В терминале: `xattr -dr com.apple.quarantine /Applications/desktop-app.app`.
- **Linux.** deb и rpm ставятся из файла без проверки ключа; AppImage нужно сделать исполняемым: `chmod +x desktop-app-1.2.3-x86_64.AppImage`.

## Иконка

Шаблонная иконка — белое окно на цветной плашке; цвет плашки выводится из слага. Её рисует `scripts/icon` (стандартная библиотека Go, без внешних программ).

| Файл | Что это |
| --- | --- |
| `build/appicon.png` | 1024×1024, иконка приложения; из неё делаются `.ico` и `.icns` |
| `build/trayicon.png` | 64×64, значок трея Linux и Windows |
| `build/trayicon-template.png` | 64×64, одноцветный значок строки меню macOS |
| `build/windows/icon.ico`, `build/darwin/icons.icns` | производные, генерируются |

- `wails3 task icons` — нарисовать шаблонную иконку заново и обновить производные; `COLOR=#RRGGBB` задаёт цвет плашки вручную.
- **Своя иконка.** Положить её в `build/appicon.png` (квадрат 1024×1024, PNG с прозрачностью) и выполнить `wails3 task icons:derive`: обновятся `.ico` и `.icns`, сама иконка не тронется. Значки трея проект рисует сам и кладёт на место `build/trayicon.png` и `build/trayicon-template.png`. После своей иконки `wails3 task icons` больше не запускать — он перерисует шаблонную.

## Проверка пакета

| Формат | Состав | Формат в приложении |
| --- | --- | --- |
| deb | `dpkg-deb -I` (версия, зависимости), `dpkg-deb -c` (файлы) | после установки `desktop-app --version` — `deb` |
| rpm | `rpm -qpl` или `7z x` и `cpio -it` | `rpm` |
| Arch | `tar -tf` (`zstd`) и `.PKGINFO` | `archlinux` |
| AppImage | `./<файл>.AppImage --appimage-extract` | `./<файл>.AppImage --version` — `appimage` |
| NSIS | `7z l` | после установки — `nsis` |
| exe | — | `desktop-app.exe --version` — `exe` |
| dmg | смонтировать образ | `desktop-app.app/Contents/MacOS/desktop-app --version` — `dmg` |

Суммы файлов выпуска: `sha256sum -c SHA256SUMS-linux.txt` в каталоге `dist/` (на macOS — `shasum -a 256 -c`).
