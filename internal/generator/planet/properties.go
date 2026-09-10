// internal/generator/planet/properties.go
package planet

import "math/rand"

// Properties — физические параметры конкретной планеты.
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
	Core                  Core
}

// GenerateProperties — рассчитывает параметры планеты на основе архетипа,
// орбитальных данных, звёздных характеристик и возраста системы.
//
// Порядок:
//  1. Размер, масса (от архетипа).
//  2. Базовая композиция поверхности (без T, т.к. T ещё не посчитана).
//  3. Ядро (по массе, климату, недрам).
//  4. Итоговая температура (с учётом альбедо, парника, ядра).
//  5. Вода (по температуре).
//  6. Жизнь, обитаемость.
//  7. Пересборка композиций с финальной температурой и водой.
func GenerateProperties(
	archetype *Archetype,
	orbitIndex int,
	spectralClass string,
	systemAge float64,
	rng *rand.Rand,
) *Properties {
	// 1. Размер и масса
	size := archetype.SizeMin + rng.Float64()*(archetype.SizeMax-archetype.SizeMin)
	mass := archetype.MassMin + rng.Float64()*(archetype.MassMax-archetype.MassMin)

	// 2. Первичная композиция поверхности (без поправок по T и воде,
	//    но с проверкой совместимости) — нужна для альбедо.
	preliminarySurface := GenerateSurfaceComposition(
		archetype.BaseSurface,
		0, // T ещё не известна — корректировки по T будут применены позже
		0, // water тоже
		rng,
	)

	// 3. Предварительная композиция недр — нужна для генерации ядра
	//    (радиоактивность и магма берутся из недр).
	preliminarySubterrain := GenerateSubterrainComposition(
		archetype.BaseSubterrain,
		preliminarySurface,
		0,
		0,
		rng,
	)

	// 4. Ядро
	core := GenerateCore(mass, archetype.Climate, preliminarySubterrain, systemAge, rng)

	// 5. Итоговая температура
	luminosity := luminosityBySpectral(spectralClass)
	orbitRadius := orbitRadiusByIndex(orbitIndex)

	temp := computeSurfaceTemp(SurfaceTempInput{
		Luminosity:   luminosity,
		OrbitRadius:  orbitRadius,
		Surface:      preliminarySurface,
		Atmosphere:   archetype.Atmosphere,
		Core:         core,
		TidalHeat:    0,
		ArchetypeMin: archetype.TemperatureMin,
		ArchetypeMax: archetype.TemperatureMax,
		SkipInternal: false,
	})

	// 6. Вода (по финальной температуре)
	waterPercent := generateWater(archetype, temp, rng)

	// 7. Пересборка композиции поверхности с учётом реальных T и воды
	surfaceComp := GenerateSurfaceComposition(
		archetype.BaseSurface,
		temp,
		waterPercent,
		rng,
	)

	// 8. Пересборка композиции недр с учётом финальной поверхности
	subterrainComp := GenerateSubterrainComposition(
		archetype.BaseSubterrain,
		surfaceComp,
		temp,
		waterPercent,
		rng,
	)

	// 9. Жизнь
	life := generateLife(archetype, waterPercent, temp, rng)

	// 10. Обитаемость
	habitable := generateHabitable(life, waterPercent, temp, archetype.Atmosphere)

	// 11. Спутники
	moons := determineMoons(size, archetype.Climate, rng)

	// 12. Население
	var population int64 = 0
	if life && habitable {
		basePop := int64(1000000 + rng.Float64()*999000000)
		dev := 0.1 + rng.Float64()*0.9
		population = int64(float64(basePop) * dev)
	}

	// 13. Политика
	political := "нет"
	if population > 0 {
		systems := []string{
			"демократия", "диктатура", "теократия",
			"корпоратократия", "анархия", "ИИ-управление",
		}
		political = systems[rng.Intn(len(systems))]
	}

	// 14. Конфликт и развитие
	conflict := 0.0
	devLevel := 0.0
	if population > 0 {
		conflict = rng.Float64()
		devLevel = 0.1 + rng.Float64()*0.9
	}

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
		Core:                  core,
	}
}

// ==================== ХЕЛПЕРЫ ====================

// generateWater — определяет процент воды на планете.
func generateWater(archetype *Archetype, temp float64, rng *rand.Rand) float64 {
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
func determineMoons(size float64, climate string, rng *rand.Rand) int {
	base := 0
	switch climate {
	case "hot", "extreme":
		base = int(size / 10)
	case "cold":
		base = int(size / 6)
	default:
		base = int(size / 8)
	}
	if base < 0 {
		base = 0
	}
	jitter := rng.Intn(3) - 1
	moons := base + jitter
	if moons < 0 {
		moons = 0
	}
	return moons
}