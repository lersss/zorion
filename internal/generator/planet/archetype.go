// internal/generator/planet/archetype.go
package planet

import (
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"path/filepath"
)

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

func LoadArchetypes(path string) error {
	// Получаем абсолютный путь
	absPath, err := filepath.Abs(path)
	if err != nil {
		log.Printf("⚠️ Ошибка получения абсолютного пути: %v", err)
		absPath = path
	}
	log.Printf("🔍 Загрузка архетипов из: %s", absPath)

	data, err := os.ReadFile(absPath)
	if err != nil {
		// Если не нашли, пробуем искать относительно текущей директории
		cwd, _ := os.Getwd()
		log.Printf("⚠️ Текущая директория: %s", cwd)
		return err
	}
	var cfg ArchetypeConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Printf("⚠️ Ошибка парсинга JSON: %v", err)
		return err
	}
	archetypeCache = &cfg
	return nil
}

func GenerateArchetype(spectralClass string, rng *rand.Rand) *Archetype {
	if archetypeCache == nil {
		return fallbackArchetype()
	}

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