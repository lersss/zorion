// internal/generator/planet/generator_gas.go
package planet

import (
	"image"
	"image/color"
	"math"
	"math/rand"
)

func generateGas(img *image.RGBA, size int, rng *rand.Rand) {
	radius := float64(size)/2 - 2
	cx, cy := float64(size)/2, float64(size)/2
	tiltAngle := (rng.Float64() - 0.5) * 1.6
	cosA := math.Cos(tiltAngle)
	sinA := math.Sin(tiltAngle)

	numBands := 4 + rng.Intn(6)
	type band struct {
		pos   float64
		width float64
		color color.RGBA
	}
	bands := make([]band, numBands)
	baseHue := 10 + rng.Intn(40)
	for i := 0; i < numBands; i++ {
		pos := -radius + (float64(i)/float64(numBands-1))*2*radius
		width := 3 + rng.Float64()*8
		hue := baseHue + rng.Intn(30) - 15
		sat := 80
		lig := 40 + rng.Intn(30)
		r, g, b := hslToRgb(hue, sat, lig)
		bands[i] = band{pos, width, color.RGBA{r, g, b, 255}}
	}
	spotX := cx + (rng.Float64()-0.5)*radius*0.8
	spotY := cy + (rng.Float64()-0.5)*radius*0.6
	spotR := 3 + rng.Float64()*8
	spotColR, spotColG, spotColB := hslToRgb(int(baseHue+20), 90, 60)

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - cx
			dy := float64(y) - cy
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist > radius {
				continue
			}
			// Поворот для полос
			xRot := dx*cosA + dy*sinA
			yRot := -dx*sinA + dy*cosA
			_ = xRot // не используется, но нужно для компиляции
			r, g, b := 40.0, 30.0, 20.0
			for _, band := range bands {
				dyBand := yRot - band.pos
				if math.Abs(dyBand) < band.width/2 {
					factor := 1 - math.Abs(dyBand)/(band.width/2)
					mix := factor*0.8 + 0.2
					br := float64(band.color.R)
					bg := float64(band.color.G)
					bb := float64(band.color.B)
					r = r*(1-mix) + br*mix
					g = g*(1-mix) + bg*mix
					b = b*(1-mix) + bb*mix
				}
			}
			dSpot := math.Sqrt((float64(x)-spotX)*(float64(x)-spotX) + (float64(y)-spotY)*(float64(y)-spotY))
			if dSpot < spotR {
				factor := 1 - dSpot/spotR
				r = r*(1-factor) + float64(spotColR)*factor
				g = g*(1-factor) + float64(spotColG)*factor
				b = b*(1-factor) + float64(spotColB)*factor
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