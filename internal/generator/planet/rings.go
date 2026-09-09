// internal/generator/planet/rings.go
package planet

import (
	"image"
	"image/color"
	"math"
	"math/rand"
)

// drawRings добавляет кольца к изображению планеты
func drawRings(img *image.RGBA, size int, rng *rand.Rand) {
	cx, cy := float64(size)/2, float64(size)/2
	radius := float64(size)/2 - 2
	ringOuter := radius * 1.6
	ringInner := radius * 1.1
	tilt := 0.3 + rng.Float64()*0.4
	rotation := 0.2 + rng.Float64()*0.3
	alpha := 0.3 + rng.Float64()*0.3

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - cx
			dy := float64(y) - cy
			cosR := math.Cos(rotation)
			sinR := math.Sin(rotation)
			xRot := dx*cosR - dy*sinR
			yRot := (dx*sinR + dy*cosR) / tilt
			dist := math.Sqrt(xRot*xRot + yRot*yRot)
			if dist >= ringInner && dist <= ringOuter {
				pos := (dist - ringInner) / (ringOuter - ringInner)
				fade := 1.0
				if pos < 0.1 {
					fade = pos / 0.1
				} else if pos > 0.8 {
					fade = 1 - (pos-0.8)/0.2
				}
				if fade < 0 {
					fade = 0
				}
				baseColor := color.RGBA{200, 180, 150, uint8(alpha * 255 * fade)}
				c := img.RGBAAt(x, y)
				if c.A == 0 {
					img.SetRGBA(x, y, baseColor)
				} else {
					sa := float64(baseColor.A) / 255.0
					da := float64(c.A) / 255.0
					outA := sa + da*(1-sa)
					if outA > 0 {
						r := (float64(baseColor.R)*sa + float64(c.R)*da*(1-sa)) / outA
						g := (float64(baseColor.G)*sa + float64(c.G)*da*(1-sa)) / outA
						b := (float64(baseColor.B)*sa + float64(c.B)*da*(1-sa)) / outA
						img.SetRGBA(x, y, color.RGBA{uint8(r), uint8(g), uint8(b), uint8(outA * 255)})
					}
				}
			}
		}
	}
}