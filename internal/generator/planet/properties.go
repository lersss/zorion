package planet

import (
	"math"
	"math/rand"
)

type Properties struct {
	Type          string
	Size          float64
	Mass          float64
	Atmosphere    string
	Temperature   float64
	WaterPercent  float64
	Moons         int
	Habitable     bool
	Life          bool
	Population    int64
	Political     string
	ConflictLevel float64
	Development   float64
}

// GenerateProperties создаёт параметры планеты на основе архетипа и орбитальных данных
func GenerateProperties(archetype *Archetype, orbitIndex int, spectralClass string, starTemp int, rng *rand.Rand) *Properties {
	// 1. Орбитальный радиус (в условных единицах, 1 AU = 1)
	orbitRadius := 0.4 * math.Pow(1.7, float64(orbitIndex))

	// 2. Светимость звезды (в солнечных единицах)
	luminosityMap := map[string]float64{
		"O": 1000, "B": 100, "A": 10, "F": 2, "G": 1, "K": 0.1, "M": 0.01,
	}
	luminosity := luminosityMap[spectralClass]
	if luminosity == 0 {
		luminosity = 1.0 // fallback
	}

	// 3. Эффективная температура планеты (без атмосферы)
	ratio := 1.0 / (orbitRadius * orbitRadius * luminosity)
	effectiveTemp := float64(starTemp) * math.Sqrt(math.Sqrt(ratio))
	if effectiveTemp < 10 {
		effectiveTemp = 10
	}

	// 4. Базовые параметры из архетипа
	size := archetype.SizeMin + rng.Float64()*(archetype.SizeMax-archetype.SizeMin)
	mass := archetype.MassMin + rng.Float64()*(archetype.MassMax-archetype.MassMin)

	// 5. Температура: effectiveTemp с разбросом ±10%
	temp := effectiveTemp * (0.9 + rng.Float64()*0.2)

	// Ограничиваем диапазоном архетипа
	if temp < archetype.TemperatureMin {
		temp = archetype.TemperatureMin
	}
	if temp > archetype.TemperatureMax {
		temp = archetype.TemperatureMax
	}

	// 6. Вода
	waterPercent := 0.0
	if archetype.Hydrosphere != "сухая" && archetype.Hydrosphere != "кислотная" {
		if temp > 250 && temp < 400 {
			if rng.Float64() < archetype.WaterChance {
				waterPercent = 30 + rng.Float64()*60
			}
		} else if temp > 150 && temp <= 250 {
			if rng.Float64() < archetype.WaterChance*0.7 {
				waterPercent = 20 + rng.Float64()*40
			}
		} else {
			if rng.Float64() < archetype.WaterChance*0.3 {
				waterPercent = 10 + rng.Float64()*20
			}
		}
	}
	// Переопределяем для конкретных гидросфер
	switch archetype.Hydrosphere {
	case "океаны":
		waterPercent = 70 + rng.Float64()*25
	case "озёра":
		waterPercent = 20 + rng.Float64()*40
	case "ледяной покров":
		waterPercent = 5 + rng.Float64()*20
	case "подлёдная":
		waterPercent = 50 + rng.Float64()*40
	}

	// 7. Жизнь
	life := false
	if archetype.Biosphere != "стерильная" && waterPercent > 10 && temp > 200 && temp < 400 {
		if rng.Float64() < archetype.LifeChance {
			life = true
		}
	}

	// 8. Обитаемость
	habitable := false
	if life && waterPercent > 10 && temp > 200 && temp < 350 && archetype.Atmosphere != "ядовитая" {
		habitable = true
	}

	// 9. Атмосфера (из архетипа)
	atmosphere := archetype.Atmosphere

	// 10. Спутники
	moons := 0
	switch archetype.Surface {
	case "скалистая", "песчаная", "глинистая", "стеклянная":
		moons = int(size / 5)
	case "ледяная", "реголитовая", "металлическая":
		moons = int(size / 8)
	default:
		moons = 0
	}

	// 11. Население
	var population int64 = 0
	if life && habitable {
		basePop := int64(1000000 + rng.Float64()*999000000)
		dev := 0.1 + rng.Float64()*0.9
		population = int64(float64(basePop) * dev)
	}

	// 12. Политика
	political := "нет"
	if population > 0 {
		systems := []string{"демократия", "диктатура", "теократия", "корпоратократия", "анархия", "ИИ-управление"}
		political = systems[rng.Intn(len(systems))]
	}

	// 13. Конфликт
	conflict := 0.0
	if population > 0 {
		conflict = rng.Float64()
	}

	// 14. Развитие
	devLevel := 0.0
	if population > 0 {
		devLevel = 0.1 + rng.Float64()*0.9
	}

	return &Properties{
		Type:          archetype.Surface,
		Size:          size,
		Mass:          mass,
		Atmosphere:    atmosphere,
		Temperature:   temp,
		WaterPercent:  waterPercent,
		Moons:         moons,
		Habitable:     habitable,
		Life:          life,
		Population:    population,
		Political:     political,
		ConflictLevel: conflict,
		Development:   devLevel,
	}
}