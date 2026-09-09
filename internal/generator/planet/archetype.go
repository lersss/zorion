package planet

import (
	"encoding/json"
	"math/rand"
	"os"
)

// Archetype — готовый набор черт планеты
type Archetype struct {
	ID              string
	Name            string
	Surface         string
	Hydrosphere     string
	Atmosphere      string
	Biosphere       string
	TemperatureMin  float64
	TemperatureMax  float64
	WaterChance     float64
	LifeChance      float64
	SizeMin         float64
	SizeMax         float64
	MassMin         float64
	MassMax         float64
}

// ClimateConfig — структура из JSON
type ClimateConfig struct {
	ID                  string            `json:"id"`
	Name                string            `json:"name"`
	Weight              map[string]float64 `json:"weight"`
	AllowedSurfaces     []string          `json:"allowed_surfaces"`
	AllowedHydrospheres []string          `json:"allowed_hydrospheres"`
	AllowedAtmospheres  []string          `json:"allowed_atmospheres"`
	AllowedBiospheres   []string          `json:"allowed_biospheres"`
	TemperatureMin      float64           `json:"temperature_min"`
	TemperatureMax      float64           `json:"temperature_max"`
	WaterChance         float64           `json:"water_chance"`
	LifeChance          float64           `json:"life_chance"`
}

type ArchetypeConfig struct {
	Climates []ClimateConfig `json:"climates"`
}

var archetypeCache *ArchetypeConfig

// LoadArchetypes загружает конфиг из JSON
func LoadArchetypes(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var cfg ArchetypeConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return err
	}
	archetypeCache = &cfg
	return nil
}

// GenerateArchetype создаёт архетип на основе спектрального класса
func GenerateArchetype(spectralClass string, rng *rand.Rand) *Archetype {
	if archetypeCache == nil {
		return fallbackArchetype()
	}

	// 1. Выбираем климат с учётом весов
	var selectedClimate *ClimateConfig
	totalWeight := 0.0
	for _, c := range archetypeCache.Climates {
		if w, ok := c.Weight[spectralClass]; ok {
			totalWeight += w
		}
	}
	if totalWeight == 0 {
		selectedClimate = &archetypeCache.Climates[rng.Intn(len(archetypeCache.Climates))]
	} else {
		r := rng.Float64() * totalWeight
		for _, c := range archetypeCache.Climates {
			if w, ok := c.Weight[spectralClass]; ok {
				r -= w
				if r <= 0 {
					selectedClimate = &c
					break
				}
			}
		}
		if selectedClimate == nil {
			selectedClimate = &archetypeCache.Climates[0]
		}
	}

	// 2. Выбираем черты из разрешённых списков
	surface := selectedClimate.AllowedSurfaces[rng.Intn(len(selectedClimate.AllowedSurfaces))]
	hydro := selectedClimate.AllowedHydrospheres[rng.Intn(len(selectedClimate.AllowedHydrospheres))]
	atmo := selectedClimate.AllowedAtmospheres[rng.Intn(len(selectedClimate.AllowedAtmospheres))]
	bio := selectedClimate.AllowedBiospheres[rng.Intn(len(selectedClimate.AllowedBiospheres))]

	return &Archetype{
		ID:              selectedClimate.ID + "_" + surface + "_" + hydro + "_" + atmo + "_" + bio,
		Name:            selectedClimate.Name + " (" + surface + ")",
		Surface:         surface,
		Hydrosphere:     hydro,
		Atmosphere:      atmo,
		Biosphere:       bio,
		TemperatureMin:  selectedClimate.TemperatureMin,
		TemperatureMax:  selectedClimate.TemperatureMax,
		WaterChance:     selectedClimate.WaterChance,
		LifeChance:      selectedClimate.LifeChance,
		SizeMin:         0.5,
		SizeMax:         14.5,
		MassMin:         0.1,
		MassMax:         19.9,
	}
}

func fallbackArchetype() *Archetype {
	return &Archetype{
		ID:              "fallback",
		Name:            "Землеподобная",
		Surface:         "скалистая",
		Hydrosphere:     "океаны",
		Atmosphere:      "азотно-кислородная",
		Biosphere:       "растительная",
		TemperatureMin:  200,
		TemperatureMax:  350,
		WaterChance:     0.7,
		LifeChance:      0.4,
		SizeMin:         0.5,
		SizeMax:         14.5,
		MassMin:         0.1,
		MassMax:         19.9,
	}
}