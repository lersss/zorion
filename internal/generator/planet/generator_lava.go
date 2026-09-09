// internal/generator/planet/generator_lava.go
package planet

import (
	"image"
	"image/color"
	"math"
	"math/rand"
)

func generateLava(img *image.RGBA, size int, rng *rand.Rand) {
	radius := float64(size)/2 - 2
	cx, cy := float64(size)/2, float64(size)/2
	baseR := uint8(20 + rng.Intn(20))
	baseG := uint8(10 + rng.Intn(10))
	baseB := uint8(10 + rng.Intn(10))

	lavaCount := 4 + rng.Intn(6)
	lavas := make([][]struct{ x, y float64 }, lavaCount)
	for i := 0; i < lavaCount; i++ {
		startAngle := rng.Float64() * 2 * math.Pi
		startDist := rng.Float64() * radius * 0.9
		x1 := cx + math.Cos(startAngle)*startDist
		y1 := cy + math.Sin(startAngle)*startDist
		points := []struct{ x, y float64 }{{x1, y1}}
		curX, curY := x1, y1
		for j := 0; j < 4+rng.Intn(3); j++ {
			angle := math.Atan2(curY-cy, curX-cx) + (rng.Float64()-0.5)*1.5
			dist := math.Sqrt((curX-cx)*(curX-cx) + (curY-cy)*(curY-cy))
			newDist := dist + (rng.Float64()-0.5)*radius*0.3
			newX := cx + math.Cos(angle)*newDist
			newY := cy + math.Sin(angle)*newDist
			if math.Sqrt((newX-cx)*(newX-cx)+(newY-cy)*(newY-cy)) > radius {
				break
			}
			points = append(points, struct{ x, y float64 }{newX, newY})
			curX, curY = newX, newY
		}
		lavas[i] = points
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
			lavaGlow := 0.0
			for _, pts := range lavas {
				for i := 0; i < len(pts)-1; i++ {
					p1, p2 := pts[i], pts[i+1]
					dx1 := float64(x) - p1.x
					dy1 := float64(y) - p1.y
					dx2 := p2.x - p1.x
					dy2 := p2.y - p1.y
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
					projX := p1.x + t*dx2
					projY := p1.y + t*dy2
					d := math.Sqrt((float64(x)-projX)*(float64(x)-projX) + (float64(y)-projY)*(float64(y)-projY))
					if d < 2.5 {
						factor := 1 - d/2.5
						if factor > lavaGlow {
							lavaGlow = factor
						}
					}
				}
			}
			if lavaGlow > 0 {
				lavaR := 255.0
				lavaG := 150.0 + rng.Float64()*50
				lavaB := 20.0 + rng.Float64()*30
				mix := lavaGlow*0.9 + 0.1
				r = r*(1-mix) + lavaR*mix
				g = g*(1-mix) + lavaG*mix
				b = b*(1-mix) + lavaB*mix
			}
			noise := math.Sin(float64(x)*0.5+float64(y)*0.6)*5 + math.Cos(float64(x)*0.8-float64(y)*0.4)*4
			r += noise
			g += noise
			b += noise
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