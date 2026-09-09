// internal/generator/planet/generator_rocky.go
package planet

import (
	"image"
	"image/color"
	"math"
	"math/rand"
)

func generateRocky(img *image.RGBA, size int, rng *rand.Rand) {
	radius := float64(size)/2 - 2
	cx, cy := float64(size)/2, float64(size)/2
	baseR := uint8(100 + rng.Intn(60))
	baseG := uint8(80 + rng.Intn(50))
	baseB := uint8(60 + rng.Intn(40))

	craterCount := 3 + rng.Intn(8)
	type crater struct {
		x, y, r float64
		depth   float64
	}
	craters := make([]crater, craterCount)
	for i := 0; i < craterCount; i++ {
		angle := rng.Float64() * 2 * math.Pi
		dist := rng.Float64() * radius * 0.8
		craters[i] = crater{
			x:     cx + math.Cos(angle)*dist,
			y:     cy + math.Sin(angle)*dist,
			r:     2 + rng.Float64()*6,
			depth: 0.3 + rng.Float64()*0.5,
		}
	}
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - cx
			dy := float64(y) - cy
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist > radius {
				continue
			}
			noise1 := math.Sin(float64(x)*0.2+float64(y)*0.3)*10 + math.Cos(float64(x)*0.4-float64(y)*0.5)*8
			noise2 := math.Sin(float64(x)*0.7+float64(y)*0.9)*5
			r := float64(baseR) + noise1 + noise2
			g := float64(baseG) + noise1*0.8 + noise2*0.6
			b := float64(baseB) + noise1*0.5 + noise2*0.4
			for _, cr := range craters {
				ddx := float64(x) - cr.x
				ddy := float64(y) - cr.y
				d := math.Sqrt(ddx*ddx + ddy*ddy)
				if d < cr.r {
					factor := 1 - d/cr.r
					darken := factor * cr.depth * 30
					r -= darken
					g -= darken
					b -= darken
					if d > cr.r*0.7 {
						rim := (d - cr.r*0.7) / (cr.r * 0.3) * 10
						r += rim
						g += rim
						b += rim
					}
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