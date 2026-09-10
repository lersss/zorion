// internal/generator/planet/composition_modifiers.go
package planet

import "math/rand"

// ==================== РАНДОМ ====================

// applyRandomJitter — умножает каждый вес на случайный множитель [0.8, 1.2].
func applyRandomJitter(c map[string]float64, rng *rand.Rand) {
	for k, v := range c {
		factor := 0.8 + rng.Float64()*0.4 // [0.8, 1.2]
		c[k] = v * factor
	}
}

// ==================== ПОВЕРХНОСТЬ: ТЕМПЕРАТУРА ====================

// applySurfaceTempModifiers — корректирует веса поверхности по температуре.
func applySurfaceTempModifiers(c map[string]float64, temperature float64) {
	switch {
	case temperature > 700:
		// Экстремально жарко
		multiplyIfExists(c, SurfaceLavaFields, 1.8)
		multiplyIfExists(c, SurfaceVolcanicFields, 1.3)
		multiplyIfExists(c, SurfaceGlassFields, 1.4)
		multiplyIfExists(c, SurfaceMetalFields, 1.2)
		multiplyIfExists(c, SurfaceOceans, 0.2)
		multiplyIfExists(c, SurfaceLakes, 0.2)
		multiplyIfExists(c, SurfaceGlaciers, 0.05)
		multiplyIfExists(c, SurfaceForests, 0.05)
		multiplyIfExists(c, SurfaceJungles, 0.05)
	case temperature > 500:
		multiplyIfExists(c, SurfaceLavaFields, 1.5)
		multiplyIfExists(c, SurfaceVolcanicFields, 1.2)
		multiplyIfExists(c, SurfaceGlassFields, 1.2)
		multiplyIfExists(c, SurfaceOceans, 0.4)
		multiplyIfExists(c, SurfaceForests, 0.3)
	case temperature < 150:
		// Очень холодно
		multiplyIfExists(c, SurfaceGlaciers, 2.0)
		multiplyIfExists(c, SurfaceFrozenGases, 1.8)
		multiplyIfExists(c, SurfaceOceans, 0.2)
		multiplyIfExists(c, SurfaceLakes, 0.3)
		multiplyIfExists(c, SurfaceForests, 0.1)
		multiplyIfExists(c, SurfaceJungles, 0.05)
	case temperature < 220:
		multiplyIfExists(c, SurfaceGlaciers, 1.4)
		multiplyIfExists(c, SurfaceFrozenGases, 1.3)
		multiplyIfExists(c, SurfaceForests, 0.5)
		multiplyIfExists(c, SurfaceJungles, 0.3)
	case temperature > 300 && temperature < 380:
		// Умеренная зона — биосфера вверх
		multiplyIfExists(c, SurfaceForests, 1.2)
		multiplyIfExists(c, SurfaceMeadows, 1.2)
		multiplyIfExists(c, SurfaceLakes, 1.1)
	}
}

// ==================== ПОВЕРХНОСТЬ: ВОДА ====================

// applySurfaceWaterModifiers — корректирует веса поверхности по проценту воды.
func applySurfaceWaterModifiers(c map[string]float64, waterPercent float64) {
	switch {
	case waterPercent > 70:
		multiplyIfExists(c, SurfaceOceans, 2.0)
		multiplyIfExists(c, SurfaceLakes, 1.3)
		multiplyIfExists(c, SurfaceCoralReefs, 1.5)
		multiplyIfExists(c, SurfaceSands, 0.5)
		multiplyIfExists(c, SurfaceRocks, 0.6)
	case waterPercent > 40:
		multiplyIfExists(c, SurfaceOceans, 1.4)
		multiplyIfExists(c, SurfaceLakes, 1.2)
	case waterPercent < 5:
		// Практически нет воды — убираем водные формы
		delete(c, SurfaceOceans)
		delete(c, SurfaceLakes)
		delete(c, SurfaceCoralReefs)
		delete(c, SurfaceSwamps)
		multiplyIfExists(c, SurfaceSands, 1.4)
		multiplyIfExists(c, SurfaceRocks, 1.2)
		multiplyIfExists(c, SurfaceCraters, 1.2)
	case waterPercent < 20:
		multiplyIfExists(c, SurfaceOceans, 0.3)
		multiplyIfExists(c, SurfaceLakes, 0.5)
		multiplyIfExists(c, SurfaceSwamps, 0.3)
		multiplyIfExists(c, SurfaceCoralReefs, 0.2)
		multiplyIfExists(c, SurfaceSands, 1.2)
	}
}

// ==================== НЕДРА: ТЕМПЕРАТУРА ====================

// applySubterrainTempModifiers — корректировки недр по температуре.
func applySubterrainTempModifiers(c map[string]float64, temperature float64) {
	switch {
	case temperature > 700:
		multiplyIfExists(c, SubterrainMagmaChambers, 2.0)
		multiplyIfExists(c, SubterrainMagmaticRocks, 1.3)
		multiplyIfExists(c, SubterrainMetalCores, 1.3)
		multiplyIfExists(c, SubterrainRadioactiveZones, 1.2)
		multiplyIfExists(c, SubterrainGroundIce, 0.05)
		multiplyIfExists(c, SubterrainCoalSeams, 0.2)
		multiplyIfExists(c, SubterrainOilPockets, 0.2)
	case temperature > 400:
		multiplyIfExists(c, SubterrainMagmaChambers, 1.4)
		multiplyIfExists(c, SubterrainMagmaticRocks, 1.2)
		multiplyIfExists(c, SubterrainGroundIce, 0.2)
	case temperature < 150:
		multiplyIfExists(c, SubterrainGroundIce, 1.8)
		multiplyIfExists(c, SubterrainMagmaChambers, 0.3)
		multiplyIfExists(c, SubterrainCoalSeams, 0.5)
	case temperature < 220:
		multiplyIfExists(c, SubterrainGroundIce, 1.3)
	}
}

// ==================== НЕДРА: ВОДА ====================

// applySubterrainWaterModifiers — корректировки недр по воде.
func applySubterrainWaterModifiers(c map[string]float64, waterPercent float64) {
	switch {
	case waterPercent > 60:
		multiplyIfExists(c, SubterrainGroundwater, 1.6)
		multiplyIfExists(c, SubterrainSedimentaryRocks, 1.3)
		multiplyIfExists(c, SubterrainSaltDomes, 1.2)
	case waterPercent < 5:
		multiplyIfExists(c, SubterrainGroundwater, 0.3)
		multiplyIfExists(c, SubterrainGroundIce, 0.3)
		multiplyIfExists(c, SubterrainOilPockets, 0.5)
		multiplyIfExists(c, SubterrainCoalSeams, 0.5)
	}
}

// ==================== НЕДРА: СВЯЗЬ С ПОВЕРХНОСТЬЮ ====================

// applySurfaceToSubterrainLinks — мягкое влияние поверхности на недра.
func applySurfaceToSubterrainLinks(c map[string]float64, surface Composition) {
	if len(surface) == 0 {
		return
	}

	// Вулканические/лавовые поля → магма, руды, сера
	if surface.ShareOf(SurfaceLavaFields)+surface.ShareOf(SurfaceVolcanicFields) > 25 {
		multiplyIfExists(c, SubterrainMagmaChambers, 1.4)
		multiplyIfExists(c, SubterrainMagmaticRocks, 1.2)
		multiplyIfExists(c, SubterrainOreVeins, 1.2)
	}

	// Океаны → нефть на шельфе, соляные купола, осадочные породы
	if surface.ShareOf(SurfaceOceans) > 30 {
		multiplyIfExists(c, SubterrainOilPockets, 1.4)
		multiplyIfExists(c, SubterrainSaltDomes, 1.3)
		multiplyIfExists(c, SubterrainSedimentaryRocks, 1.2)
	}

	// Леса/болота → уголь (захоронённая биомасса)
	if surface.ShareOf(SurfaceForests)+surface.ShareOf(SurfaceSwamps) > 20 {
		multiplyIfExists(c, SubterrainCoalSeams, 1.5)
		multiplyIfExists(c, SubterrainOilPockets, 1.2)
	}

	// Ледники → подземные льды, кристаллические жилы
	if surface.ShareOf(SurfaceGlaciers) > 30 {
		multiplyIfExists(c, SubterrainGroundIce, 1.5)
		multiplyIfExists(c, SubterrainCrystalVeins, 1.2)
	}

	// Кратеры → руды от метеоритов, редкие металлы
	if surface.ShareOf(SurfaceCraters) > 15 {
		multiplyIfExists(c, SubterrainOreVeins, 1.3)
		multiplyIfExists(c, SubterrainRareEarthVeins, 1.3)
	}
}