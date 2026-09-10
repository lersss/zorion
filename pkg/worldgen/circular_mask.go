package worldgen

import (
	"math"
)

// ApplyCircularMask накладывает круглую маску на карту высот.
// Карта — двумерный срез [rows][cols] float64.
// cx, cy — координаты центра (обычно rows/2, cols/2).
// radius — радиус круга.
// outsideValue — значение за пределами круга (0 — вода).
func ApplyCircularMask(heightMap [][]float64, cx, cy, radius float64, outsideValue float64) {
	rows := len(heightMap)
	if rows == 0 {
		return
	}
	cols := len(heightMap[0])

	radiusSq := radius * radius

	for y := 0; y < rows; y++ {
		for x := 0; x < cols; x++ {
			dx := float64(x) - cx
			dy := float64(y) - cy
			distSq := dx*dx + dy*dy
			if distSq > radiusSq {
				heightMap[y][x] = outsideValue
			}
		}
	}
}

// ApplySmoothCircularMask — то же самое, но с плавным переходом на границе.
// smoothWidth — ширина зоны сглаживания.
func ApplySmoothCircularMask(heightMap [][]float64, cx, cy, radius, smoothWidth float64, outsideValue float64) {
	rows := len(heightMap)
	if rows == 0 {
		return
	}
	cols := len(heightMap[0])

	for y := 0; y < rows; y++ {
		for x := 0; x < cols; x++ {
			dx := float64(x) - cx
			dy := float64(y) - cy
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist > radius {
				heightMap[y][x] = outsideValue
			} else if dist > radius-smoothWidth {
				t := (dist - (radius - smoothWidth)) / smoothWidth
				smooth := t * t * (3 - 2*t)
				// Смешиваем исходное значение с outsideValue
				heightMap[y][x] = heightMap[y][x]*(1-smooth) + outsideValue*smooth
			}
			// иначе оставляем без изменений
		}
	}
}