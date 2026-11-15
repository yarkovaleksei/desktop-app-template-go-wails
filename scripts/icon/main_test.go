package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// templatePlate — цвет плашки для слага шаблона. Меняется только вместе с решением о цвете иконки.
var templatePlate = color.RGBA{R: 0xAB, G: 0x2B, B: 0x90, A: 0xff}

func decode(t *testing.T, data []byte) image.Image {
	t.Helper()
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	return img
}

func TestRenderAll(t *testing.T) {
	files, err := renderAll(slugColor("desktop-app"))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]int{
		appIconFile:          appIconSize,
		trayIconFile:         trayIconSize,
		trayIconTemplateFile: trayIconSize,
	}
	if len(files) != len(want) {
		t.Fatalf("файлов %d, ожидалось %d", len(files), len(want))
	}
	for _, f := range files {
		t.Run(f.name, func(t *testing.T) {
			size, ok := want[f.name]
			if !ok {
				t.Fatalf("лишний файл %s", f.name)
			}
			img := decode(t, f.data)
			if b := img.Bounds(); b.Dx() != size || b.Dy() != size {
				t.Fatalf("размер %v, ожидалось %d×%d", b.Size(), size, size)
			}
			// Углы и края прозрачны: знак не касается края холста.
			for i := range size {
				for _, p := range []image.Point{{i, 0}, {i, size - 1}, {0, i}, {size - 1, i}} {
					if _, _, _, a := img.At(p.X, p.Y).RGBA(); a != 0 {
						t.Fatalf("непрозрачный пиксель на краю %v", p)
					}
				}
			}
		})
	}
}

func TestAppIconColors(t *testing.T) {
	img := renderApp(appIconSize, templatePlate)
	tests := []struct {
		name string
		x, y float64
		want color.RGBA
	}{
		{"плашка слева от знака", 0.15, 0.50, templatePlate},
		{"плашка внутри окна", 0.50, 0.55, templatePlate},
		{"полоса заголовка", 0.50, 0.33, white},
		{"рамка окна", 0.25, 0.55, white},
		{"прозрачный угол", 0.11, 0.11, color.RGBA{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := img.RGBAAt(int(tt.x*appIconSize), int(tt.y*appIconSize))
			if got != tt.want {
				t.Errorf("цвет %v, ожидалось %v", got, tt.want)
			}
		})
	}
}

func TestSlugColor(t *testing.T) {
	if got := slugColor("desktop-app"); got != templatePlate {
		t.Errorf("цвет слага desktop-app %v, закреплён %v", got, templatePlate)
	}
	seen := map[color.RGBA]string{}
	for _, slug := range []string{"desktop-app", "photo-viewer", "notes", "my-app", "budget"} {
		c := slugColor(slug)
		if c.A != 0xff {
			t.Errorf("%s: плашка не непрозрачна: %v", slug, c)
		}
		if other, ok := seen[c]; ok {
			t.Errorf("слаги %s и %s дали один цвет %v", slug, other, c)
		}
		seen[c] = slug
	}
}

func TestParseColor(t *testing.T) {
	tests := []struct {
		in      string
		want    color.RGBA
		wantErr bool
	}{
		{in: "#AB2B90", want: color.RGBA{0xab, 0x2b, 0x90, 0xff}},
		{in: "#00ff7f", want: color.RGBA{0x00, 0xff, 0x7f, 0xff}},
		{in: "AB2B90", wantErr: true},
		{in: "#AB2B9", wantErr: true},
		{in: "#AB2B900", wantErr: true},
		{in: "#GG0000", wantErr: true},
		{in: "red", wantErr: true},
		{in: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := parseColor(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ошибка %v, ожидалась: %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("цвет %v, ожидалось %v", got, tt.want)
			}
		})
	}
}

func TestPlateColor(t *testing.T) {
	tests := []struct {
		name, slug, color string
		want              color.RGBA
		wantErr           bool
	}{
		{name: "из слага", slug: "desktop-app", want: templatePlate},
		{name: "флаг важнее слага", slug: "desktop-app", color: "#102030", want: color.RGBA{0x10, 0x20, 0x30, 0xff}},
		{name: "недопустимый флаг", slug: "desktop-app", color: "#1020", wantErr: true},
		{name: "пустой слаг", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := plateColor(tt.slug, tt.color)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ошибка %v, ожидалась: %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("цвет %v, ожидалось %v", got, tt.want)
			}
		})
	}
}

// TestRunDeterministic проверяет, что повторный запуск даёт те же байты,
// а недопустимый цвет не создаёт файлов.
func TestRunDeterministic(t *testing.T) {
	dirs := []string{t.TempDir(), t.TempDir()}
	for _, d := range dirs {
		if err := run("desktop-app", "", d); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{appIconFile, trayIconFile, trayIconTemplateFile} {
		a, err := os.ReadFile(filepath.Join(dirs[0], name))
		if err != nil {
			t.Fatal(err)
		}
		b, err := os.ReadFile(filepath.Join(dirs[1], name))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(a, b) {
			t.Errorf("%s: повторный запуск дал другие байты", name)
		}
	}

	bad := filepath.Join(t.TempDir(), "out")
	if err := run("desktop-app", "#xyz", bad); err == nil {
		t.Fatal("недопустимый -color принят")
	}
	if _, err := os.Stat(bad); !os.IsNotExist(err) {
		t.Errorf("при ошибке создан каталог вывода: %v", err)
	}
}
