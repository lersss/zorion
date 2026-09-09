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
	// T_eff = T_star * (1 / (r^2 * L))^0.25
	// Упрощённо: sqrt(sqrt(1.0 / (r^2 * L)))
	ratio := 1.0 / (orbitRadius * orbitRadius * luminosity)
	effectiveTemp := float64(starTemp) * math.Sqrt(math.Sqrt(ratio))
	if effectiveTemp < 10 {
		effectiveTemp = 10 // минимальная температура
	}

	// 4. Выбор климата на основе эффективной температуры
	var climateID string
	switch {
	case effectiveTemp > 500:
		climateID = "hot"
	case effectiveTemp > 250:
		climateID = "temperate"
	case effectiveTemp > 150:
		climateID = "cold"
	default:
		climateID = "extreme"
	}
	// Если температура близка к границе, иногда выбираем "variable"
	if effectiveTemp > 230 && effectiveTemp < 270 && rng.Float64() < 0.3 {
		climateID = "variable"
	}

	// 5. Если переданный archetype уже содержит климат, но он может не совпадать с effectiveTemp,
	// мы всё равно используем effectiveTemp для выбора климата (переопределяем).
	// В новой логике мы не используем archetype для выбора климата, только для черт.
	// Но у нас archetype уже содержит климат. Чтобы не ломать старую логику, мы создадим новый Archetype
	// на основе effectiveTemp и разрешённых списков из конфига.
	// Для этого нужно получить ClimateConfig по climateID из загруженного конфига.
	// Это проще сделать в generatePlanet, но я покажу, как адаптировать здесь.

	// Вместо этого мы будем использовать archetype только для черт, но климат переопределим.
	// Для простоты я создам fallback-архетип на основе effectiveTemp, но у нас уже есть archetype с чертами.
	// В данном случае мы можем просто скорректировать температуру и воду, а климат оставить как есть.
	// Но это не совсем физично. Поэтому я предлагаю переделать generatePlanet так, чтобы он вызывал новую функцию,
	// которая выбирает климат по температуре, а архетип получает только черты.
	// Однако, чтобы не переписывать всю структуру, я оставлю старую логику с архетипом, но скорректирую
	// физику на основе effectiveTemp.

	// Итак, мы будем использовать эффективную температуру как основу, а архетип (черты) как дополнение.
	// В результате температура будет определяться эффективной температурой (с небольшим разбросом).

	// Базовые параметры из архетипа
	size := archetype.SizeMin + rng.Float64()*(archetype.SizeMax-archetype.SizeMin)
	mass := archetype.MassMin + rng.Float64()*(archetype.MassMax-archetype.MassMin)

	// Температура: берём effectiveTemp, но добавляем случайный разброс ±10%
	temp := effectiveTemp * (0.9 + rng.Float64()*0.2)

	// Ограничиваем температуру диапазоном архетипа (если архетип имеет ограничения)
	if temp < archetype.TemperatureMin {
		temp = archetype.TemperatureMin
	}
	if temp > archetype.TemperatureMax {
		temp = archetype.TemperatureMax
	}

	// Вода: зависит от температуры и климата
	waterPercent := 0.0
	if archetype.Hydrosphere != "сухая" && archetype.Hydrosphere != "кислотная" {
		// Вероятность воды зависит от температуры: 0°C < T < 100°C -> больше воды
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
	if archetype.Hydrosphere == "океаны" {
		waterPercent = 70 + rng.Float64()*25
	} else if archetype.Hydrosphere == "озёра" {
		waterPercent = 20 + rng.Float64()*40
	} else if archetype.Hydrosphere == "ледяной покров" {
		waterPercent = 5 + rng.Float64()*20
	} else if archetype.Hydrosphere == "подлёдная" {
		waterPercent = 50 + rng.Float64()*40
	}

	// Жизнь: зависит от температуры, воды и биосферы
	life := false
	if archetype.Biosphere != "стерильная" && waterPercent > 10 && temp > 200 && temp < 400 {
		if rng.Float64() < archetype.LifeChance {
			life = true
		}
	}

	// Обитаемость
	habitable := false
	if life && waterPercent > 10 && temp > 200 && temp < 350 && archetype.Atmosphere != "ядовитая" {
		habitable = true
	}

	// Атмосфера
	atmosphere := archetype.Atmosphere
	// Если температура очень высокая, атмосфера может быть плотнее, но оставим как есть

	// Спутники
	moons := 0
	switch archetype.Surface {
	case "скалистая", "песчаная", "глинистая", "стеклянная":
		moons = int(size / 5)
	case "ледяная", "реголитовая", "металлическая":
		moons = int(size / 8)
	default:
		moons = 0
	}

	// Население
	var population int64 = 0
	if life && habitable {
		basePop := int64(1000000 + rng.Float64()*999000000)
		dev := 0.1 + rng.Float64()*0.9
		population = int64(float64(basePop) * dev)
	}

	// Политика
	political := "нет"
	if population > 0 {
		systems := []string{"демократия", "диктатура", "теократия", "корпоратократия", "анархия", "ИИ-управление"}
		political = systems[rng.Intn(len(systems))]
	}

	// Конфликт
	conflict := 0.0
	if population > 0 {
		conflict = rng.Float64()
	}

	// Развитие
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