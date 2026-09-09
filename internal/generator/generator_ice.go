// internal/generator/planet/generator_ice.go
package planet

import (
	"image"
	"image/color"
	"math"
	"math/rand"
)

func generateIce(img *image.RGBA, size int, rng *rand.Rand) {
	radius := float64(size)/2 - 2
	cx, cy := float64(size)/2, float64(size)/2
	baseR := uint8(180 + rng.Intn(40))
	baseG := uint8(200 + rng.Intn(40))
	baseB := uint8(220 + rng.Intn(30))

	type crack struct {
		x1, y1, x2, y2, width float64
	}
	crackCount := 5 + rng.Intn(10)
	cracks := make([]crack, crackCount)
	for i := 0; i < crackCount; i++ {
		angle := rng.Float64() * 2 * math.Pi
		dist := rng.Float64() * radius * 0.8
		x1 := cx + math.Cos(angle)*dist
		y1 := cy + math.Sin(angle)*dist
		angle2 := angle + (rng.Float64()-0.5)*1.2
		dist2 := dist + (rng.Float64()-0.5)*radius*0.5
		x2 := cx + math.Cos(angle2)*dist2
		y2 := cy + math.Sin(angle2)*dist2
		cracks[i] = crack{x1, y1, x2, y2, 1 + rng.Float64()*2}
	}
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - cx
			dy := float64(y) - cy
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist > radius {
				continue
			}
			r := float64(baseR)
			g := float64(baseG)
			b := float64(baseB)
			noise := math.Sin(float64(x)*0.3+float64(y)*0.5)*8 + math.Cos(float64(x)*0.7-float64(y)*0.2)*6
			r += noise
			g += noise
			b += noise
			for _, cr := range cracks {
				dx1 := float64(x) - cr.x1
				dy1 := float64(y) - cr.y1
				dx2 := cr.x2 - cr.x1
				dy2 := cr.y2 - cr.y1
				len2 := dx2*dx2 + dy2*dy2
				if len2 == 0 {
					continue
				}
				t := (dx1*dx2 + dy1*dy2) / len2
				if t < 0 {
					t = 0
				}
				if t > 1 {
					t = 1
				}
				projX := cr.x1 + t*dx2
				projY := cr.y1 + t*dy2
				d := math.Sqrt((float64(x)-projX)*(float64(x)-projX) + (float64(y)-projY)*(float64(y)-projY))
				if d < cr.width {
					factor := 1 - d/cr.width
					darken := factor * 30
					r -= darken
					g -= darken
					b -= darken
				}
			}
			clamp := func(v float64) uint8 {
				if v < 0 {
					return 0
				}
				if v > 255 {
					return 255
				}
				return uint8(v)
			}
			img.SetRGBA(x, y, color.RGBA{clamp(r), clamp(g), clamp(b), 255})
		}
	}
}