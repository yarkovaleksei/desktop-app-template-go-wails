module github.com/yarkovaleksei/desktop-app-template-go-wails

go 1.26.0

// Зависимости фронтенда бывают с Go-кодом (flatted): ./... не должен их видеть.
ignore ./frontend/node_modules

require (
	github.com/wailsapp/wails/v3 v3.0.0
	github.com/yarkovaleksei/go-desktop-kit v0.1.1
)

require (
	git.sr.ht/~jackmordaunt/go-toast/v2 v2.0.3 // indirect
	github.com/BurntSushi/toml v1.6.0 // indirect
	github.com/adrg/xdg v0.5.3 // indirect
	github.com/coder/websocket v1.8.14 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/godbus/dbus/v5 v5.2.2 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	golang.org/x/exp/typeparams v0.0.0-20260112195511-716be5621a96 // indirect
	golang.org/x/mod v0.41.0 // indirect
	golang.org/x/sync v0.23.0 // indirect
	golang.org/x/sys v0.50.0 // indirect
	golang.org/x/tools v0.51.0 // indirect
	honnef.co/go/tools v0.8.2 // indirect
)

tool (
	github.com/yarkovaleksei/go-desktop-kit/cmd/relkit
	honnef.co/go/tools/cmd/staticcheck
)
