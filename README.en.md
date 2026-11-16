# Desktop App

[Русский](README.md)

A desktop application template on Go, Wails 3, React and MUI. It builds for Linux, Windows and macOS, opens a window with settings, a tray icon and updates from GitHub Releases, and passes its checks right after cloning.

## Requirements

| What | Version | Why |
| --- | --- | --- |
| Go | 1.26 | Application and tools |
| Node.js | 24 | Frontend |
| yarn | 4 (via `corepack enable`) | Frontend dependencies |
| Wails CLI `wails3` | v3.0.0: `go install github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0` | Builds, tasks (`wails3 task …`), bindings |
| Linux | GTK 4 and WebKitGTK 6.0 for building (`libgtk-4-dev`, `libwebkitgtk-6.0-dev` on Ubuntu 24.04) | Window |
| Windows | WebView2 (bundled with Windows 10 and 11); NSIS only for the installer | Window, packages |
| macOS | Xcode Command Line Tools, macOS 12 or later | Builds |

You don't need to install Task separately: `wails3 task <name>` runs the tasks.

## Quick start

```bash
git clone https://github.com/yarkovaleksei/desktop-app-template-go-wails.git
cd desktop-app-template-go-wails
corepack enable
wails3 task deps
wails3 task dev
```

`wails3 task dev` opens the window in development mode: frontend changes apply immediately, settings and the log live in `.task/dev-home`. The interface without Go, in a browser on mocks, is `wails3 task ui`.

## Commands

| Command | What it does |
| --- | --- |
| `wails3 task deps` | Frontend dependencies from `yarn.lock` |
| `wails3 task dev` | Window in development mode; `REAL=true` uses the real user directories |
| `wails3 task ui` | Interface in a browser on mocks, without Go |
| `wails3 task site` | The `www/` site on a local server |
| `wails3 task build`, `wails3 task run` | Binary for the current OS in `bin/` and running it |
| `wails3 task check` | Everything before a commit: dependencies, tests, types, linters, docs |
| `wails3 task test`, `wails3 task test:go`, `wails3 task test:frontend` | Go and frontend tests |
| `wails3 task test:e2e` | Playwright end-to-end scenarios on mocks |
| `wails3 task lint`, `wails3 task typecheck`, `wails3 task fmt` | Linters, type checking, formatting |
| `wails3 task check:race` | Core tests under the race detector |
| `wails3 task bindings`, `wails3 task check:bindings` | Regenerate TypeScript bindings; check that they are up to date |
| `wails3 task check:docs`, `wails3 task docs:map` | Check docs and the code map; regenerate the map |
| `wails3 task smoke` | Run the built binary and check that the window opened |
| `wails3 task smoke:desktop` | Linux: run the installed application as if from the menu |
| `wails3 task package` | All packages for the current OS in `bin/` |
| `wails3 task release:collect` | Packages under release names and the checksum file in `dist/` |
| `wails3 task release:plan`, `wails3 task release:verify`, `wails3 task release:promote` | State before a release, draft verification, publishing |
| `wails3 task version` | The version that goes into the build |
| `wails3 task icons`, `wails3 task icons:derive` | Draw the template icon; update `.ico` and `.icns` from your own |

## Repository layout

```text
main.go, desktop.go   entry point and desktop integration
internal/             app (Wails services), appinfo, settings, texts, updatecheck,
                      repodocs (code map), archtest (repository rules)
frontend/             React + MUI: src/ (api, store, pages, i18n, theme), e2e/, bindings/
build/                Wails configuration, icons, Linux, Windows and macOS packages
scripts/              init (project initialization from the template), icon (icon generator), smoke-desktop.sh
docs/                 guides and the code map
www/                  project site: one page in two languages
spec/                 project plan: rules, stages, decisions, backlog, journal (in Russian)
.github/              check, docs, release, pages workflows; issue and PR templates
Taskfile.yml          tasks
```

The detailed map is [docs/code-map.md](docs/code-map.md).

## Installation

Files are on the [releases](https://github.com/yarkovaleksei/desktop-app-template-go-wails/releases/latest) page; `<version>` is the release number without `v`. Packages are not signed — how to open them on Windows and macOS is described in [docs/packaging.md](docs/packaging.md#подписи-нет) (in Russian).

| OS | Package | Installation |
| --- | --- | --- |
| Debian, Ubuntu | `desktop-app_<version>_amd64.deb` | `sudo apt install ./desktop-app_<version>_amd64.deb` |
| Fedora | `desktop-app-<version>.x86_64.rpm` | `sudo dnf install ./desktop-app-<version>.x86_64.rpm` |
| Arch Linux | `desktop-app-<version>-x86_64.pkg.tar.zst` | `sudo pacman -U desktop-app-<version>-x86_64.pkg.tar.zst` |
| Any Linux | `desktop-app-<version>-x86_64.AppImage` | `chmod +x` and run; on Ubuntu the first run asks to install an AppArmor profile |
| Windows | `desktop-app-<version>-windows-amd64-setup.exe` | Installer, no administrator rights needed |
| Windows | `desktop-app-<version>-windows-amd64.exe` | Portable file, no installation |
| macOS | `desktop-app-<version>-macos-universal.dmg` | Open the image and move the app to Applications |

SHA-256 checksums are in `SHA256SUMS-linux.txt`, `SHA256SUMS-windows.txt` and `SHA256SUMS-macos.txt` of the same release. An application installed from a package updates from its settings page.

## Documentation

The documentation map is [docs/README.md](docs/README.md): which document owns what and when to update it. The guides are in Russian. The main ones: architecture — [docs/development.md](docs/development.md), desktop integration — [docs/desktop.md](docs/desktop.md), code rules — [docs/code-style.md](docs/code-style.md), interface — [docs/ui.md](docs/ui.md), packages — [docs/packaging.md](docs/packaging.md), releases — [docs/release.md](docs/release.md), CI — [docs/ci.md](docs/ci.md), recipes — [docs/recipes/](docs/recipes/README.md).

## Starting a project from the template

1. Get a copy of the template: on GitHub — “Use this template” → “Create a new repository” and clone the new repository, or clone the template itself under a new name and replace `origin`:

   ```bash
   git clone https://github.com/yarkovaleksei/desktop-app-template-go-wails.git my-app
   cd my-app
   git remote set-url origin https://github.com/owner/my-app.git
   ```

2. Initialize the project. With an agent — the `template-init` skill: it asks for the names, shows a dry run and runs the initialization once you agree. Without an agent — a dry run first, then the same command without `-dry-run`:

   ```bash
   go run ./scripts/init -name "My App" -repo owner/my-app -description "One-line description" -dry-run
   go run ./scripts/init -name "My App" -repo owner/my-app -description "One-line description"
   ```

3. Follow the next steps the tool prints at the end — from `wails3 task deps` to a green `wails3 task check` — and commit the result.

| Flag | What it sets | Default |
| --- | --- | --- |
| `-name` | Product name: window, menu, packages, site | required |
| `-repo` | GitHub repository `owner/name`: updates, site, links | required |
| `-description` | One-line description: build files, packages, launcher, README | required |
| `-slug` | Binary, directory and package name, skill prefix | from the name: `my-app` |
| `-module` | Go module | `github.com/<repository>` |
| `-id` | Product identifier in reverse domain notation | `io.github.<owner>.<slug without hyphens>` |
| `-author`, `-year` | Copyright holder and year in `LICENSE` and build files | repository owner, current year |
| `-color` | Plate colour of the template icon, `#RRGGBB` | from the slug |

The tool only works on a clean template tree: it refuses with uncommitted changes and on an already initialized project. It replaces the template names in every tracked file and path (the environment prefix comes from the slug: `MY_APP_HOME`), renames the launcher, the AppArmor profile and the `app-*` skills to `<slug>-*`, points site publishing at the project repository, resets `spec/` to its initial state with the product description from the flag, rewrites this README, redraws the icon in the new slug's colour and removes `scripts/init`, the `template-init` skill and this section. It stages its changes in git; the commit is yours.

## License

[MIT](LICENSE).
