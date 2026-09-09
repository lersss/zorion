// internal/generator/planet/utils_scale.go
package planet

import (
	"image"
	"image/color"
	"math"
)

// scaleImage масштабирует изображение с билинейной интерполяцией
func scaleImage(src *image.RGBA, targetRadius int) *image.RGBA {
	srcSize := src.Bounds().Dx()
	targetSize := targetRadius * 2
	if targetSize == srcSize {
		return src
	}
	dst := image.NewRGBA(image.Rect(0, 0, targetSize, targetSize))
	// Билинейная интерполяция
	for y := 0; y < targetSize; y++ {
		for x := 0; x < targetSize; x++ {
			// Координаты в исходном изображении (float)
			srcX := float64(x) * float64(srcSize) / float64(targetSize)
			srcY := float64(y) * float64(srcSize) / float64(targetSize)
			// Билинейная интерполяция
			x0 := int(math.Floor(srcX))
			x1 := int(math.Ceil(srcX))
			y0 := int(math.Floor(srcY))
			y1 := int(math.Ceil(srcY))
			// Ограничиваем
			if x0 < 0 {
				x0 = 0
			}
			if x1 >= srcSize {
				x1 = srcSize - 1
			}
			if y0 < 0 {
				y0 = 0
			}
			if y1 >= srcSize {
				y1 = srcSize - 1
			}
			// Веса
			fx := srcX - float64(x0)
			fy := srcY - float64(y0)
			// Цвета четырёх соседей
			c00 := src.RGBAAt(x0, y0)
			c01 := src.RGBAAt(x1, y0)
			c10 := src.RGBAAt(x0, y1)
			c11 := src.RGBAAt(x1, y1)
			// Интерполяция по X
			r0 := float64(c00.R)*(1-fx) + float64(c01.R)*fx
			g0 := float64(c00.G)*(1-fx) + float64(c01.G)*fx
			b0 := float64(c00.B)*(1-fx) + float64(c01.B)*fx
			a0 := float64(c00.A)*(1-fx) + float64(c01.A)*fx
			r1 := float64(c10.R)*(1-fx) + float64(c11.R)*fx
			g1 := float64(c10.G)*(1-fx) + float64(c11.G)*fx
			b1 := float64(c10.B)*(1-fx) + float64(c11.B)*fx
			a1 := float64(c10.A)*(1-fx) + float64(c11.A)*fx
			// Интерполяция по Y
			r := r0*(1-fy) + r1*fy
			g := g0*(1-fy) + g1*fy
			b := b0*(1-fy) + b1*fy
			a := a0*(1-fy) + a1*fy
			dst.SetRGBA(x, y, color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)})
		}
	}
	return dst
}