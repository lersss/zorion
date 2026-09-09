// internal/generator/planet/generator_earth.go
package planet

import (
	"image"
	"image/color"
	"math"
	"math/rand"
)

func generateEarth(img *image.RGBA, size int, rng *rand.Rand) {
	radius := float64(size)/2 - 2
	cx, cy := float64(size)/2, float64(size)/2
	oceanR := uint8(20 + rng.Intn(30))
	oceanG := uint8(60 + rng.Intn(50))
	oceanB := uint8(120 + rng.Intn(60))

	type continent struct {
		x, y, r float64
		color   [3]uint8
	}
	numContinents := 3 + rng.Intn(4)
	continents := make([]continent, numContinents)
	for i := 0; i < numContinents; i++ {
		angle := rng.Float64() * 2 * math.Pi
		dist := rng.Float64() * radius * 0.7
		continents[i] = continent{
			x: cx + math.Cos(angle)*dist,
			y: cy + math.Sin(angle)*dist,
			r: 5 + rng.Float64()*15,
			color: [3]uint8{
				uint8(50 + rng.Intn(80)),
				uint8(100 + rng.Intn(70)),
				uint8(30 + rng.Intn(50)),
			},
		}
	}
	type cloud struct {
		x, y, r, alpha float64
	}
	numClouds := 2 + rng.Intn(5)
	clouds := make([]cloud, numClouds)
	for i := 0; i < numClouds; i++ {
		angle := rng.Float64() * 2 * math.Pi
		dist := rng.Float64() * radius * 0.8
		clouds[i] = cloud{
			x:     cx + math.Cos(angle)*dist,
			y:     cy + math.Sin(angle)*dist,
			r:     4 + rng.Float64()*12,
			alpha: 0.1 + rng.Float64()*0.3,
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
			r := float64(oceanR)
			g := float64(oceanG)
			b := float64(oceanB)
			for _, cont := range continents {
				ddx := float64(x) - cont.x
				ddy := float64(y) - cont.y
				d := math.Sqrt(ddx*ddx + ddy*ddy)
				if d < cont.r {
					factor := 1 - d/cont.r
					blend := factor*factor*0.9 + 0.1
					r = r*(1-blend) + float64(cont.color[0])*blend
					g = g*(1-blend) + float64(cont.color[1])*blend
					b = b*(1-blend) + float64(cont.color[2])*blend
				}
			}
			noise := math.Sin(float64(x)*0.5+float64(y)*0.7)*5 + math.Cos(float64(x)*0.9-float64(y)*0.3)*4
			r += noise
			g += noise
			b += noise
			for _, cl := range clouds {
				ddx := float64(x) - cl.x
				ddy := float64(y) - cl.y
				d := math.Sqrt(ddx*ddx + ddy*ddy)
				if d < cl.r {
					factor := 1 - d/cl.r
					alpha := factor * cl.alpha
					r = r*(1-alpha) + 255*alpha
					g = g*(1-alpha) + 255*alpha
					b = b*(1-alpha) + 255*alpha
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