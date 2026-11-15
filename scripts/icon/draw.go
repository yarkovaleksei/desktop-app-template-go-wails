package main

import (
	"fmt"
	"hash/fnv"
	"image"
	"image/color"
	"image/draw"
	"math"
	"regexp"
	"strconv"
)

// Насыщенность и светлота плашки: белый знак на ней остаётся контрастным при любом оттенке.
const (
	plateSaturation = 0.60
	plateLightness  = 0.42
)

// samplesPerAxis — выборок на пиксель по каждой оси (сглаживание краёв).
const samplesPerAxis = 4

var (
	white     = color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	trayLight = color.RGBA{R: 0xf5, G: 0xf5, B: 0xf5, A: 0xff}
	trayDark  = color.RGBA{R: 0x1f, G: 0x1f, B: 0x1f, A: 0xff}
	black     = color.RGBA{A: 0xff}
)

var hexColor = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// parseColor разбирает цвет в виде #RRGGBB.
func parseColor(s string) (color.RGBA, error) {
	if !hexColor.MatchString(s) {
		return color.RGBA{}, fmt.Errorf("цвет %q не в виде #RRGGBB", s)
	}
	v, err := strconv.ParseUint(s[1:], 16, 32)
	if err != nil {
		return color.RGBA{}, err
	}
	return color.RGBA{R: uint8(v >> 16), G: uint8(v >> 8), B: uint8(v), A: 0xff}, nil
}

// slugColor вычисляет цвет плашки из слага: оттенок — из хэша FNV-1a,
// насыщенность и светлота фиксированы.
func slugColor(slug string) color.RGBA {
	h := fnv.New32a()
	h.Write([]byte(slug))
	hue := float64(h.Sum32() % 360)
	return hslToRGB(hue, plateSaturation, plateLightness)
}

func hslToRGB(hue, s, l float64) color.RGBA {
	c := (1 - math.Abs(2*l-1)) * s
	hp := hue / 60
	x := c * (1 - math.Abs(math.Mod(hp, 2)-1))
	var r, g, b float64
	switch {
	case hp < 1:
		r, g = c, x
	case hp < 2:
		r, g = x, c
	case hp < 3:
		g, b = c, x
	case hp < 4:
		g, b = x, c
	case hp < 5:
		r, b = x, c
	default:
		r, b = c, x
	}
	m := l - c/2
	to8 := func(v float64) uint8 { return uint8(math.Round((v + m) * 255)) }
	return color.RGBA{R: to8(r), G: to8(g), B: to8(b), A: 0xff}
}

// roundRect — прямоугольник со скруглёнными углами в долях холста.
type roundRect struct {
	x0, y0, x1, y1, r float64
}

// sdf — расстояние со знаком до границы: меньше нуля внутри.
func (q roundRect) sdf(x, y float64) float64 {
	cx, cy := (q.x0+q.x1)/2, (q.y0+q.y1)/2
	dx := math.Abs(x-cx) - ((q.x1-q.x0)/2 - q.r)
	dy := math.Abs(y-cy) - ((q.y1-q.y0)/2 - q.r)
	outside := math.Hypot(math.Max(dx, 0), math.Max(dy, 0))
	inside := math.Min(math.Max(dx, dy), 0)
	return outside + inside - q.r
}

// windowGlyph — знак окна: рамка с толстой полосой заголовка.
type windowGlyph struct {
	outer, inner roundRect
}

func newWindowGlyph(x0, y0, x1, y1, radius, stroke, title float64) windowGlyph {
	return windowGlyph{
		outer: roundRect{x0, y0, x1, y1, radius},
		inner: roundRect{x0 + stroke, y0 + title, x1 - stroke, y1 - stroke, math.Max(radius-stroke, radius/4)},
	}
}

func (g windowGlyph) sdf(x, y float64) float64 {
	return math.Max(g.outer.sdf(x, y), -g.inner.sdf(x, y))
}

// shader возвращает цвет точки холста в долях [0, 1]; нулевая альфа — пусто.
type shader func(x, y float64) color.RGBA

// render рисует квадратный холст size×size, усредняя выборки внутри пикселя.
func render(size int, shade shader) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	const n = samplesPerAxis * samplesPerAxis
	for py := range size {
		for px := range size {
			var r, g, b, a int
			for sy := range samplesPerAxis {
				for sx := range samplesPerAxis {
					x := (float64(px) + (float64(sx)+0.5)/samplesPerAxis) / float64(size)
					y := (float64(py) + (float64(sy)+0.5)/samplesPerAxis) / float64(size)
					c := shade(x, y)
					// Накапливаем цвет, умноженный на альфу: так хранит image.RGBA.
					r += int(c.R) * int(c.A)
					g += int(c.G) * int(c.A)
					b += int(c.B) * int(c.A)
					a += int(c.A)
				}
			}
			img.SetRGBA(px, py, color.RGBA{
				R: uint8((r + 255*n/2) / (255 * n)),
				G: uint8((g + 255*n/2) / (255 * n)),
				B: uint8((b + 255*n/2) / (255 * n)),
				A: uint8((a + n/2) / n),
			})
		}
	}
	return img
}

// Геометрия иконки приложения: плашка занимает 80% холста, края прозрачные.
var (
	appPlate = roundRect{0.10, 0.10, 0.90, 0.90, 0.18}
	appGlyph = newWindowGlyph(0.22, 0.28, 0.78, 0.72, 0.05, 0.07, 0.15)
)

// Геометрия значка трея: знак окна без плашки с тёмной обводкой.
var (
	trayGlyph   = newWindowGlyph(0.08, 0.18, 0.92, 0.82, 0.08, 0.10, 0.22)
	trayOutline = 0.05
)

// renderApp рисует иконку приложения: белое окно на цветной плашке.
func renderApp(size int, plate color.RGBA) *image.RGBA {
	return render(size, func(x, y float64) color.RGBA {
		switch {
		case appPlate.sdf(x, y) >= 0:
			return color.RGBA{}
		case appGlyph.sdf(x, y) < 0:
			return white
		default:
			return plate
		}
	})
}

// renderTray рисует значок трея для Linux и Windows: светлый знак с тёмной обводкой.
func renderTray(size int) *image.RGBA {
	return render(size, func(x, y float64) color.RGBA {
		d := trayGlyph.sdf(x, y)
		switch {
		case d < 0:
			return trayLight
		case d < trayOutline:
			return trayDark
		default:
			return color.RGBA{}
		}
	})
}

// renderTrayTemplate рисует одноцветный знак для строки меню macOS:
// система использует только альфу шаблонного изображения.
func renderTrayTemplate(size int) *image.RGBA {
	return render(size, func(x, y float64) color.RGBA {
		if trayGlyph.sdf(x, y) < 0 {
			return black
		}
		return color.RGBA{}
	})
}

// Картинка для соцсетей: иконка по центру на фоне цвета плашки, разбавленного
// белым. Текста нет — для него нужен шрифт; проект заменяет картинку своей.
const (
	socialTint     = 0.14 // доля цвета плашки в фоне
	socialIconPart = 0.70 // высота иконки в долях высоты картинки
)

// renderSocial рисует картинку для соцсетей width×height.
func renderSocial(width, height int, plate color.RGBA) *image.RGBA {
	mix := func(c uint8) uint8 { return uint8(math.Round(255 - (255-float64(c))*socialTint)) }
	bg := color.RGBA{R: mix(plate.R), G: mix(plate.G), B: mix(plate.B), A: 0xff}
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: bg}, image.Point{}, draw.Src)
	size := int(float64(height) * socialIconPart)
	icon := renderApp(size, plate)
	at := image.Pt((width-size)/2, (height-size)/2)
	draw.Draw(img, icon.Bounds().Add(at), icon, image.Point{}, draw.Over)
	return img
}
