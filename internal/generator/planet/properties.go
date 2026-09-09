package planet

import (
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

// GenerateProperties создаёт параметры планеты на основе архетипа
func GenerateProperties(archetype *Archetype, orbitIndex int, spectralClass string, starTemp int, rng *rand.Rand) *Properties {
	size := archetype.SizeMin + rng.Float64()*(archetype.SizeMax-archetype.SizeMin)
	mass := archetype.MassMin + rng.Float64()*(archetype.MassMax-archetype.MassMin)

	temp := archetype.TemperatureMin + rng.Float64()*(archetype.TemperatureMax-archetype.TemperatureMin)

	waterPercent := 0.0
	if archetype.Hydrosphere != "сухая" && archetype.Hydrosphere != "кислотная" {
		if rng.Float64() < archetype.WaterChance {
			waterPercent = 10 + rng.Float64()*80
		}
	}
	if archetype.Hydrosphere == "океаны" {
		waterPercent = 70 + rng.Float64()*25
	} else if archetype.Hydrosphere == "озёра" {
		waterPercent = 20 + rng.Float64()*40
	} else if archetype.Hydrosphere == "ледяной покров" {
		waterPercent = 5 + rng.Float64()*20
	} else if archetype.Hydrosphere == "подлёдная" {
		waterPercent = 50 + rng.Float64()*40
	}

	life := false
	if rng.Float64() < archetype.LifeChance && archetype.Biosphere != "стерильная" {
		life = true
	}

	habitable := false
	if life && waterPercent > 10 && temp > 200 && temp < 350 && archetype.Atmosphere != "ядовитая" {
		habitable = true
	}

	atmosphere := archetype.Atmosphere

	moons := 0
	switch archetype.Surface {
	case "скалистая", "песчаная", "глинистая", "стеклянная":
		moons = int(size / 5)
	case "ледяная", "реголитовая", "металлическая":
		moons = int(size / 8)
	default:
		moons = 0
	}

	var population int64 = 0
	if life && habitable {
		basePop := int64(1000000 + rng.Float64()*999000000)
		dev := 0.1 + rng.Float64()*0.9
		population = int64(float64(basePop) * dev)
	}

	political := "нет"
	if population > 0 {
		systems := []string{"демократия", "диктатура", "теократия", "корпоратократия", "анархия", "ИИ-управление"}
		political = systems[rng.Intn(len(systems))]
	}

	conflict := 0.0
	if population > 0 {
		conflict = rng.Float64()
	}

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