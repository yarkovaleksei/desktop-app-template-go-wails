// Icon рисует шаблонную иконку приложения: белое окно на цветной плашке.
//
// Цвет плашки вычисляется из слага проекта или задаётся флагом -color.
// Программа пишет в каталог -out файлы appicon.png (1024×1024),
// trayicon.png и trayicon-template.png (64×64), а с флагом -site — картинки
// сайта: favicon.png (64×64), icon.png (256×256) и social.jpg (1280×640,
// картинка для соцсетей). Повторный запуск с теми же флагами даёт те же байты.
//
//	go run ./scripts/icon -slug desktop-app -out build -site www/images
package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"

	"github.com/yarkovaleksei/desktop-app-template-go-wails/internal/appinfo"
)

// Размеры файлов иконки.
const (
	appIconSize  = 1024
	trayIconSize = 64

	faviconSize  = 64
	siteIconSize = 256
	socialWidth  = 1280
	socialHeight = 640
)

// Имена файлов в каталоге вывода.
const (
	appIconFile          = "appicon.png"
	trayIconFile         = "trayicon.png"
	trayIconTemplateFile = "trayicon-template.png"

	faviconFile  = "favicon.png"
	siteIconFile = "icon.png"
	socialFile   = "social.jpg"
)

// socialQuality — качество JPEG картинки для соцсетей.
const socialQuality = 90

func main() {
	slug := flag.String("slug", appinfo.Slug, "слаг проекта: из него вычисляется цвет плашки")
	colorFlag := flag.String("color", "", "цвет плашки #RRGGBB вместо вычисленного из слага")
	out := flag.String("out", "build", "каталог для файлов иконки")
	site := flag.String("site", "", "каталог для картинок сайта; пусто — не писать")
	flag.Parse()

	if err := run(*slug, *colorFlag, *out, *site); err != nil {
		fmt.Fprintln(os.Stderr, "icon:", err)
		os.Exit(1)
	}
}

func run(slug, colorFlag, out, site string) error {
	plate, err := plateColor(slug, colorFlag)
	if err != nil {
		return err
	}
	files, err := renderAll(plate)
	if err != nil {
		return err
	}
	var siteFiles []iconFile
	if site != "" {
		if siteFiles, err = renderSite(plate); err != nil {
			return err
		}
	}
	if err := writeFiles(out, files); err != nil {
		return err
	}
	if site != "" {
		return writeFiles(site, siteFiles)
	}
	return nil
}

// writeFiles создаёт каталог dir и пишет в него файлы.
func writeFiles(dir string, files []iconFile) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	for _, f := range files {
		path := filepath.Join(dir, f.name)
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
	return encodePNG([]namedImage{
		{appIconFile, renderApp(appIconSize, plate)},
		{trayIconFile, renderTray(trayIconSize)},
		{trayIconTemplateFile, renderTrayTemplate(trayIconSize)},
	})
}

// renderSite рисует картинки сайта: значок вкладки, иконку и картинку для
// соцсетей.
func renderSite(plate color.RGBA) ([]iconFile, error) {
	files, err := encodePNG([]namedImage{
		{faviconFile, renderApp(faviconSize, plate)},
		{siteIconFile, renderApp(siteIconSize, plate)},
	})
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, renderSocial(socialWidth, socialHeight, plate), &jpeg.Options{Quality: socialQuality}); err != nil {
		return nil, fmt.Errorf("кодирование %s: %w", socialFile, err)
	}
	return append(files, iconFile{socialFile, buf.Bytes()}), nil
}

type namedImage struct {
	name string
	img  image.Image
}

func encodePNG(images []namedImage) ([]iconFile, error) {
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
