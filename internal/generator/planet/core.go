// internal/generator/planet/core.go
package planet

import "math/rand"

// Core — ядро планеты. Первичная сущность, определяющая температуру
// поверхности (через HeatContribution), наличие магнитного поля
// (через IsActive) и геологическую активность.
type Core struct {
	Type          string  // "металлическое" | "силикатное" | "ледяное" | "экзотическое"
	MassPercent   float64 // доля от массы планеты (5–60%)
	Activity      float64 // 0–100: жидкое ядро, конвекция
	Radioactivity float64 // 0–100: U, Th, K
	Age           float64 // млрд лет
}

// ==================== ТИПЫ ЯДРА ====================

const (
	CoreMetallic  = "металлическое"
	CoreSilicate  = "силикатное"
	CoreIce       = "ледяное"
	CoreExotic    = "экзотическое"
)

// ==================== ГЕНЕРАЦИЯ ====================

// GenerateCore — создаёт ядро для планеты.
//
// Параметры:
//   - mass         — масса планеты (в земных)
//   - climate      — климат из архетипа ("hot", "temperate", ...)
//   - subterrain   — композиция недр (для связи с радиоактивностью и магмой)
//   - systemAge    — возраст системы (млрд лет, один на мир)
//   - rng          — источник случайности
func GenerateCore(
	mass float64,
	climate string,
	subterrain Composition,
	systemAge float64,
	rng *rand.Rand,
) Core {
	coreType := determineCoreType(mass, rng)
	massPercent := determineCoreMassPercent(mass, rng)
	activity := determineCoreActivity(climate, subterrain, rng)
	radioactivity := determineCoreRadioactivity(subterrain, rng)

	return Core{
		Type:          coreType,
		MassPercent:   massPercent,
		Activity:      activity,
		Radioactivity: radioactivity,
		Age:           systemAge,
	}
}

// determineCoreType — тип ядра по массе.
func determineCoreType(mass float64, rng *rand.Rand) string {
	switch {
	case mass > 10:
		// Газовые гиганты и суперземли — всегда металлическое
		return CoreMetallic
	case mass > 1:
		// Землеподобные: металлическое или силикатное
		if rng.Float64() < 0.7 {
			return CoreMetallic
		}
		return CoreSilicate
	default:
		// Малые планеты: силикатное, ледяное, редко экзотика
		r := rng.Float64()
		switch {
		case r < 0.5:
			return CoreSilicate
		case r < 0.9:
			return CoreIce
		default:
			return CoreExotic
		}
	}
}

// determineCoreMassPercent — доля ядра от массы планеты.
func determineCoreMassPercent(mass float64, rng *rand.Rand) float64 {
	switch {
	case mass > 10:
		return 30 + rng.Float64()*30 // 30–60%
	case mass > 1:
		return 15 + rng.Float64()*20 // 15–35%
	default:
		return 5 + rng.Float64()*15 // 5–20%
	}
}

// determineCoreActivity — активность ядра.
//
// Зависит от:
//   - климата (hot/extreme → активнее);
//   - доли магматических камер в недрах;
//   - доли лавовых/вулканических полей (через композицию недр).
func determineCoreActivity(
	climate string,
	subterrain Composition,
	rng *rand.Rand,
) float64 {
	// База по климату
	var base float64
	switch climate {
	case "extreme":
		base = 70 + rng.Float64()*30 // 70–100
	case "hot":
		base = 40 + rng.Float64()*30 // 40–70
	case "temperate":
		base = 30 + rng.Float64()*30 // 30–60
	case "variable":
		base = 20 + rng.Float64()*30 // 20–50
	case "cold":
		base = 5 + rng.Float64()*20 // 5–25
	default:
		base = 20 + rng.Float64()*30
	}

	// Бонус за магматические камеры
	magmaShare := subterrain.ShareOf(SubterrainMagmaChambers)
	base += magmaShare * 1.5

	// Бонус за магматические породы
	magmaticShare := subterrain.ShareOf(SubterrainMagmaticRocks)
	base += magmaticShare * 0.5

	// Ограничение
	if base > 100 {
		base = 100
	}
	if base < 0 {
		base = 0
	}
	return base
}

// determineCoreRadioactivity — радиоактивность ядра.
//
// Прямо связана с долей радиоактивных зон в недрах.
func determineCoreRadioactivity(subterrain Composition, rng *rand.Rand) float64 {
	share := subterrain.ShareOf(SubterrainRadioactiveZones)

	// Линейная связь: 0% → 0–10, 15% → 70–100
	base := share * 6.0 // 15% → 90
	base += rng.Float64() * 10

	if base > 100 {
		base = 100
	}
	if base < 0 {
		base = 0
	}
	return base
}

// ==================== ВКЛАД В ТЕМПЕРАТУРУ ====================

// HeatContribution — сколько кельвинов добавляет ядро к температуре поверхности.
//
// Формула:
//
//	base_heat = Activity * 1.5 + Radioactivity * 2.5   (0–400 K)
//	age_factor = max(0.15, 1.0 - Age/15)               (старение)
//	mass_factor = MassPercent / 30                     (масштаб ядра)
//	contribution = base_heat * age_factor * mass_factor
func (c Core) HeatContribution() float64 {
	baseHeat := c.Activity*1.5 + c.Radioactivity*2.5

	ageFactor := 1.0 - c.Age/15.0
	if ageFactor < 0.15 {
		ageFactor = 0.15
	}

	massFactor := c.MassPercent / 30.0
	if massFactor < 0.1 {
		massFactor = 0.1
	}

	result := baseHeat * ageFactor * massFactor
	if result < 0 {
		result = 0
	}
	return result
}

// ==================== ФЛАГИ ====================

// IsMetallic — металлическое ли ядро (для UI, фильтров, магнитного поля).
func (c Core) IsMetallic() bool {
	return c.Type == CoreMetallic
}

// IsActive — активное ли ядро (для магнитного поля, вулканизма).
// Порог 40 — эмпирический: Земля ~70, Марс ~20.
func (c Core) IsActive() bool {
	return c.Activity > 40
}

// IsRadioactive — сильно ли радиоактивное ядро.
func (c Core) IsRadioactive() bool {
	return c.Radioactivity > 50
}