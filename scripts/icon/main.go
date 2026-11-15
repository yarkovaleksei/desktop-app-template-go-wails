// Icon рисует шаблонную иконку приложения: белое окно на цветной плашке.
//
// Цвет плашки вычисляется из слага проекта или задаётся флагом -color.
// Программа пишет в каталог -out файлы appicon.png (1024×1024),
// trayicon.png и trayicon-template.png (64×64). Повторный запуск с теми же
// флагами даёт те же байты.
//
//	go run ./scripts/icon -slug desktop-app -out build
package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"

	"github.com/yarkovaleksei/desktop-app-template-go-wails/internal/appinfo"
)

// Размеры файлов иконки.
const (
	appIconSize  = 1024
	trayIconSize = 64
)

// Имена файлов в каталоге вывода.
const (
	appIconFile          = "appicon.png"
	trayIconFile         = "trayicon.png"
	trayIconTemplateFile = "trayicon-template.png"
)

func main() {
	slug := flag.String("slug", appinfo.Slug, "слаг проекта: из него вычисляется цвет плашки")
	colorFlag := flag.String("color", "", "цвет плашки #RRGGBB вместо вычисленного из слага")
	out := flag.String("out", "build", "каталог для файлов иконки")
	flag.Parse()

	if err := run(*slug, *colorFlag, *out); err != nil {
		fmt.Fprintln(os.Stderr, "icon:", err)
		os.Exit(1)
	}
}

func run(slug, colorFlag, out string) error {
	plate, err := plateColor(slug, colorFlag)
	if err != nil {
		return err
	}
	files, err := renderAll(plate)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		return err
	}
	for _, f := range files {
		path := filepath.Join(out, f.name)
		if err := os.WriteFile(path, f.data, 0o644); err != nil {
			return err
		}
		fmt.Println(path)
	}
	return nil
}

// plateColor выбирает цвет плашки: из флага, если он задан, иначе из слага.
func plateColor(slug, colorFlag string) (color.RGBA, error) {
	if colorFlag != "" {
		return parseColor(colorFlag)
	}
	if slug == "" {
		return color.RGBA{}, fmt.Errorf("пустой слаг: задайте -slug или -color")
	}
	return slugColor(slug), nil
}

type iconFile struct {
	name string
	data []byte
}

// renderAll рисует все файлы иконки и кодирует их в PNG.
func renderAll(plate color.RGBA) ([]iconFile, error) {
	images := []struct {
		name string
		img  image.Image
	}{
		{appIconFile, renderApp(appIconSize, plate)},
		{trayIconFile, renderTray(trayIconSize)},
		{trayIconTemplateFile, renderTrayTemplate(trayIconSize)},
	}
	files := make([]iconFile, 0, len(images))
	enc := png.Encoder{CompressionLevel: png.BestCompression}
	for _, im := range images {
		var buf bytes.Buffer
		if err := enc.Encode(&buf, im.img); err != nil {
			return nil, fmt.Errorf("кодирование %s: %w", im.name, err)
		}
		files = append(files, iconFile{im.name, buf.Bytes()})
	}
	return files, nil
}
