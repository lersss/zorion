// internal/generator/planet/planet_data_satellites_physics.go
package planet

import (
	"math"
	"math/rand"
)

// ==================== ФИЗИКА СПУТНИКОВ ====================

// computeSatelliteTemp — температура спутника.
//
// Спутник получает тепло из трёх источников:
//  1. Остаточное излучение звезды (ослабленное, спутник в тени гиганта часть времени).
//  2. Внутренний нагрев от газового гиганта (giantTemp × 0.3).
//  3. Приливный нагрев (обратно пропорционален кубу орбитального индекса).
//
// ВАЖНО: спутник не может быть существенно горячее газового гиганта.
// Максимум — на 50 K выше (за счёт приливного разогрева, как у Ио).
// Это физически обоснованное ограничение: если гигант холодный,
// спутник тоже не может быть раскалённым.
func computeSatelliteTemp(giantTemp float64, orbitIndex int) float64 {
	if orbitIndex < 1 {
		orbitIndex = 1
	}

	// 1. Внутреннее тепло от гиганта
	internal := giantTemp * 0.3

	// 2. Приливный нагрев — скромный. Максимум ~100 K при orbitIndex=1.
	//    Раньше было 400 — отсюда и был баг «спутник в 3 раза горячее гиганта».
	tidal := 100.0 / math.Pow(float64(orbitIndex), 3)

	temp := internal + tidal

	// 3. Ограничение: не выше гиганта + 50 K
	ceiling := giantTemp + 50
	if temp > ceiling {
		temp = ceiling
	}

	// 4. Абсолютный минимум и максимум
	if temp < TempAbsoluteMin {
		temp = TempAbsoluteMin
	}
	if temp > TempAbsoluteMax {
		temp = TempAbsoluteMax
	}
	return temp
}

// computeSatelliteWater — вода на спутнике.
//
// Зависит от температуры:
//   - > 400 K — вода испарилась;
//   - 250–400 K — жидкая вода;
//   - 180–250 K — лёд + подлёдный океан;
//   - < 180 K — сплошной лёд.
func computeSatelliteWater(temp float64, rng *rand.Rand) float64 {
	switch {
	case temp > 400:
		return 0
	case temp > 350:
		return rng.Float64() * 5
	case temp > 250:
		return 10 + rng.Float64()*40
	case temp > 180:
		return 30 + rng.Float64()*50
	default:
		return 20 + rng.Float64()*60
	}
}

// pickSatelliteAtmosphere — тип атмосферы спутника.
//
// У спутников атмосфера — редкость. Если есть — соответствует температуре.
func pickSatelliteAtmosphere(temp float64, rng *rand.Rand) string {
	if rng.Float64() < 0.4 {
		return "разряженная"
	}
	switch {
	case temp > 500:
		return "ядовитая"
	case temp > 300:
		opts := []string{"метановая", "азотная", "углекислая"}
		return opts[rng.Intn(len(opts))]
	case temp > 180:
		opts := []string{"метановая", "азотная", "туманная"}
		return opts[rng.Intn(len(opts))]
	default:
		opts := []string{"разряженная", "метановая", "азотная"}
		return opts[rng.Intn(len(opts))]
	}
}

// ==================== КОМПОЗИЦИИ СПУТНИКОВ ====================

// generateSatelliteSurface — композиция поверхности спутника.
func generateSatelliteSurface(
	temp, waterPercent float64,
	orbitIndex int,
	rng *rand.Rand,
) Composition {
	c := map[string]float64{
		SurfaceRocks:          30,
		SurfaceCraters:        20,
		SurfaceGlaciers:       0,
		SurfaceFrozenGases:    0,
		SurfaceOceans:         0,
		SurfaceLakes:          0,
		SurfaceVolcanicFields: 0,
		SurfaceLavaFields:     0,
		SurfaceSands:          10,
	}

	switch {
	case temp > 700:
		c[SurfaceLavaFields] = 30
		c[SurfaceVolcanicFields] = 20
		c[SurfaceRocks] = 20
		c[SurfaceCraters] = 10
		delete(c, SurfaceGlaciers)
		delete(c, SurfaceSands)
	case temp > 400:
		c[SurfaceVolcanicFields] = 20
		c[SurfaceLavaFields] = 10
		c[SurfaceSands] = 15
		c[SurfaceRocks] = 25
	case temp > 250:
		c[SurfaceRocks] = 30
		c[SurfaceSands] = 20
		c[SurfaceLakes] = 10
		c[SurfaceCraters] = 15
	case temp > 180:
		c[SurfaceGlaciers] = 25
		c[SurfaceOceans] = 15 // подлёдный океан
		c[SurfaceRocks] = 25
		c[SurfaceCraters] = 15
	default:
		c[SurfaceGlaciers] = 40
		c[SurfaceFrozenGases] = 20
		c[SurfaceRocks] = 20
		c[SurfaceCraters] = 15
	}

	// Водные корректировки
	if waterPercent < 5 {
		delete(c, SurfaceOceans)
		delete(c, SurfaceLakes)
	} else if waterPercent > 50 {
		c[SurfaceOceans] = c[SurfaceOceans] + 20
	}

	// Абсолютные запреты
	if temp < 500 {
		delete(c, SurfaceLavaFields)
	}
	if temp > 320 {
		delete(c, SurfaceGlaciers)
		delete(c, SurfaceFrozenGases)
	}

	// Рандом ±20%
	for k, v := range c {
		c[k] = v * (0.8 + rng.Float64()*0.4)
	}

	c = resolveConflicts("surface", c, c, rng)

	return Composition(c).Normalize().NonZero()
}

// generateSatelliteSubterrain — композиция недр спутника.
func generateSatelliteSubterrain(
	temp float64,
	surface Composition,
	rng *rand.Rand,
) Composition {
	c := map[string]float64{
		SubterrainEmptyRock:     30,
		SubterrainMagmaticRocks: 15,
		SubterrainOreVeins:      15,
		SubterrainGroundIce:     0,
		SubterrainGroundwater:   0,
		SubterrainMagmaChambers: 0,
		SubterrainMetalCores:    10,
		SubterrainCrystalVeins:  10,
	}

	switch {
	case temp > 500:
		c[SubterrainMagmaChambers] = 20
		c[SubterrainMagmaticRocks] = 25
		c[SubterrainMetalCores] = 15
		c[SubterrainRadioactiveZones] = 10
	case temp < 200:
		c[SubterrainGroundIce] = 30
		c[SubterrainCrystalVeins] = 15
	}

	if surface.ShareOf(SurfaceOceans)+surface.ShareOf(SurfaceLakes) > 20 {
		c[SubterrainGroundwater] = 25
	}

	for k, v := range c {
		c[k] = v * (0.8 + rng.Float64()*0.4)
	}

	c = resolveConflicts("subterrain", c, c, rng)

	return Composition(c).Normalize().NonZero()
}

// ==================== ЖИЗНЬ И ОПИСАНИЕ ====================

// determineSatelliteLife — есть ли жизнь на спутнике.
//
// Ключевое условие: подлёдный океан при умеренной T.
// Шанс 15% — жизнь на спутниках редкость.
func determineSatelliteLife(
	temp, waterPercent float64,
	surface Composition,
	rng *rand.Rand,
) bool {
	if waterPercent < 20 {
		return false
	}
	if temp < 250 || temp > 350 {
		return false
	}
	return rng.Float64() < 0.15
}

// satelliteDescription — описание спутника.
func satelliteDescription(temp, waterPercent float64, life bool) string {
	if life {
		return "Спутник с подлёдным океаном и признаками микробной жизни."
	}
	switch {
	case temp > 700:
		return "Раскалённый спутник с активным вулканизмом."
	case temp > 400:
		return "Горячий спутник, поверхность покрыта лавой и вулканами."
	case temp > 250:
		return "Умеренный спутник с жидкой водой и плотной атмосферой."
	case temp > 180:
		return "Ледяной спутник с подповерхностным океаном."
	default:
		return "Холодный спутник, покрытый льдом и мёрзлыми газами."
	}
}