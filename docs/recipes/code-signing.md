# Подпись кода и нотаризация

## Когда нужно

Пакеты шаблона не подписаны (см. [packaging.md](../packaging.md#подписи-нет)): Windows показывает SmartScreen, macOS не открывает приложение без обхода Gatekeeper. Подпись нужна, когда приложение раздаётся людям, которым нельзя объяснять обход, или когда этого требует организация.

## Что добавить

- **macOS.** Сертификат Developer ID Application и учётная запись Apple Developer. Подпись `.app` с hardened runtime вместо ad-hoc в `build/darwin/Taskfile.yml`, нотаризация `xcrun notarytool submit --wait` и `xcrun stapler staple` для `.dmg`. В CI — сертификат и пароль приложения в секретах GitHub, временная связка ключей на раннере.
- **Windows.** Сертификат подписи кода (OV или EV; EV быстрее снимает SmartScreen). `signtool sign` с меткой времени для `desktop-app.exe` до упаковки, для установщика — после (`build/windows/nsis/project.nsi` содержит место для подписи удаления), и для переносимого exe. В CI — сертификат в секретах или облачная служба подписи.
- **Linux.** Подпись deb и rpm ключом GPG (`nfpm` умеет подписывать) и публикация открытого ключа; AppImage подписывается отдельно (`--sign` у appimagetool) или остаётся с проверкой по суммам.
- Секреты — только в секретах GitHub, у job `release`; `docs/ci.md` и `docs/packaging.md` обновляются в том же коммите.
- Обновление из приложения (`updater` из kit) подписи не проверяет: оно сверяет SHA-256 с файлом сумм выпуска. Подпись защищает первую установку и снимает предупреждения ОС.

## На что смотреть в образцах

- `ai-memory/build/darwin/Taskfile.yml` и `forktopus/build/darwin/Taskfile.yml` — ad-hoc подпись и задача нотаризации через `wails3 tool sign`.
- `ai-memory/build/linux/Taskfile.yml` — подпись пакетов Linux ключом GPG.
- `ai-memory/build/windows/nsis/project.nsi` — место подписи в установщике.
- `forktopus/scripts/release.sh` — подпись при сборке выпуска.

## Как проверить

- macOS: `codesign --verify --deep --strict` и `spctl --assess --type execute` для `.app`, `xcrun stapler validate` для `.dmg`; скачанный образ открывается без предупреждения.
- Windows: `signtool verify /pa` для exe и установщика; свойства файла показывают подпись.
- Linux: `dpkg-sig --verify` или `rpm --checksig`.
