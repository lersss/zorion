// internal/generator/planet/properties.go
package planet

import (
	"math"
	"math/rand"
)

// Properties — физические параметры конкретной планеты.
// Вместо одного Type теперь две композиции: поверхность и недра.
type Properties struct {
	Size                  float64
	Mass                  float64
	Atmosphere            string
	Temperature           float64
	WaterPercent          float64
	Moons                 int
	Habitable             bool
	Life                  bool
	Population            int64
	Political             string
	ConflictLevel         float64
	Development           float64
	SurfaceComposition    Composition
	SubterrainComposition Composition
}

// GenerateProperties — рассчитывает параметры планеты на основе архетипа,
// орбитальных данных и звёздных характеристик.
func GenerateProperties(
	archetype *Archetype,
	orbitIndex int,
	spectralClass string,
	starTemp int,
	rng *rand.Rand,
) *Properties {
	// 1. Эффективная температура (физика)
	temp := computeEffectiveTemp(starTemp, orbitIndex, spectralClass)

	// Разброс ±10% и ограничение по архетипу
	temp = temp * (0.9 + rng.Float64()*0.2)
	if temp < archetype.TemperatureMin {
		temp = archetype.TemperatureMin
	}
	if temp > archetype.TemperatureMax {
		temp = archetype.TemperatureMax
	}

	// 2. Размер и масса
	size := archetype.SizeMin + rng.Float64()*(archetype.SizeMax-archetype.SizeMin)
	mass := archetype.MassMin + rng.Float64()*(archetype.MassMax-archetype.MassMin)

	// 3. Вода
	waterPercent := generateWater(archetype, temp, rng)

	// 4. Жизнь
	life := generateLife(archetype, waterPercent, temp, rng)

	// 5. Обитаемость
	habitable := generateHabitable(life, waterPercent, temp, archetype.Atmosphere)

	// 6. Спутники (у газовых гигантов — своя логика, здесь — обычные планеты)
	moons := determineMoons(size, archetype.Climate, rng)

	// 7. Население
	var population int64 = 0
	if life && habitable {
		basePop := int64(1000000 + rng.Float64()*999000000)
		dev := 0.1 + rng.Float64()*0.9
		population = int64(float64(basePop) * dev)
	}

	// 8. Политика
	political := "нет"
	if population > 0 {
		systems := []string{
			"демократия", "диктатура", "теократия",
			"корпоратократия", "анархия", "ИИ-управление",
		}
		political = systems[rng.Intn(len(systems))]
	}

	// 9. Конфликт и развитие
	conflict := 0.0
	devLevel := 0.0
	if population > 0 {
		conflict = rng.Float64()
		devLevel = 0.1 + rng.Float64()*0.9
	}

	// 10. Композиция поверхности
	surfaceComp := GenerateSurfaceComposition(
		archetype.BaseSurface,
		temp,
		waterPercent,
		rng,
	)

	// 11. Композиция недр (зависит от поверхности)
	subterrainComp := GenerateSubterrainComposition(
		archetype.BaseSubterrain,
		surfaceComp,
		temp,
		waterPercent,
		rng,
	)

	return &Properties{
		Size:                  size,
		Mass:                  mass,
		Atmosphere:            archetype.Atmosphere,
		Temperature:           temp,
		WaterPercent:          waterPercent,
		Moons:                 moons,
		Habitable:             habitable,
		Life:                  life,
		Population:            population,
		Political:             political,
		ConflictLevel:         conflict,
		Development:           devLevel,
		SurfaceComposition:    surfaceComp,
		SubterrainComposition: subterrainComp,
	}
}

// ==================== ХЕЛПЕРЫ ====================

// generateWater — определяет процент воды на планете.
func generateWater(archetype *Archetype, temp float64, rng *rand.Rand) float64 {
	// Конкретные гидросферы переопределяют значение
	switch archetype.Hydrosphere {
	case "океаны":
		return 70 + rng.Float64()*25
	case "озёра":
		return 20 + rng.Float64()*40
	case "ледяной покров":
		return 5 + rng.Float64()*20
	case "подлёдная":
		return 50 + rng.Float64()*40
	case "кислотная":
		return 0
	case "сухая":
		return 0
	}

	// Обычный расчёт по температуре
	if temp > 250 && temp < 400 {
		if rng.Float64() < archetype.WaterChance {
			return 30 + rng.Float64()*60
		}
	} else if temp > 150 && temp <= 250 {
		if rng.Float64() < archetype.WaterChance*0.7 {
			return 20 + rng.Float64()*40
		}
	} else {
		if rng.Float64() < archetype.WaterChance*0.3 {
			return 10 + rng.Float64()*20
		}
	}
	return 0
}

// generateLife — определяет, есть ли жизнь на планете.
func generateLife(archetype *Archetype, waterPercent, temp float64, rng *rand.Rand) bool {
	if archetype.Biosphere == "стерильная" {
		return false
	}
	if waterPercent <= 10 {
		return false
	}
	if temp <= 200 || temp >= 400 {
		return false
	}
	return rng.Float64() < archetype.LifeChance
}

// generateHabitable — определяет, пригодна ли планета для жизни.
func generateHabitable(life bool, waterPercent, temp float64, atmosphere string) bool {
	if !life {
		return false
	}
	if waterPercent <= 10 {
		return false
	}
	if temp <= 200 || temp >= 350 {
		return false
	}
	if atmosphere == "ядовитая" {
		return false
	}
	return true
}

// determineMoons — количество спутников у обычной планеты.
// Зависит от размера и климата.
func determineMoons(size float64, climate string, rng *rand.Rand) int {
	base := 0
	switch climate {
	case "hot", "extreme":
		base = int(size / 10) // меньше спутников у горячих
	case "cold":
		base = int(size / 6) // больше у холодных
	default:
		base = int(size / 8)
	}
	if base < 0 {
		base = 0
	}
	// ±1 случайность
	jitter := rng.Intn(3) - 1
	moons := base + jitter
	if moons < 0 {
		moons = 0
	}
	return moons
}