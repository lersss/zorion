// internal/generator/planet/archetype.go
package planet

import (
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"path/filepath"
)

// Archetype — результат выбора архетипа для конкретной планеты.
// Вместо одиночной Surface теперь хранит базовые веса композиции.
type Archetype struct {
	ID             string
	Name           string
	Climate        string             // id климата (hot/temperate/...)
	BaseSurface    map[string]float64 // форма → базовый вес (не нормализовано)
	BaseSubterrain map[string]float64 // тип недр → базовый вес
	Hydrosphere    string
	Atmosphere     string
	Biosphere      string
	TemperatureMin float64
	TemperatureMax float64
	WaterChance    float64
	LifeChance     float64
	SizeMin        float64
	SizeMax        float64
	MassMin        float64
	MassMax        float64
}

// ClimateConfig — конфиг климата из JSON
type ClimateConfig struct {
	ID                  string             `json:"id"`
	Name                string             `json:"name"`
	Weight              map[string]float64 `json:"weight"`
	BaseSurface         map[string]float64 `json:"base_surface"`
	BaseSubterrain      map[string]float64 `json:"base_subterrain"`
	AllowedHydrospheres []string           `json:"allowed_hydrospheres"`
	AllowedAtmospheres  []string           `json:"allowed_atmospheres"`
	AllowedBiospheres   []string           `json:"allowed_biospheres"`
	TemperatureMin      float64            `json:"temperature_min"`
	TemperatureMax      float64            `json:"temperature_max"`
	WaterChance         float64            `json:"water_chance"`
	LifeChance          float64            `json:"life_chance"`
}

type ArchetypeConfig struct {
	Climates []ClimateConfig `json:"climates"`
}

var archetypeCache *ArchetypeConfig

// LoadArchetypes — загружает архетипы из JSON
func LoadArchetypes(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		log.Printf("⚠️ Ошибка получения абсолютного пути: %v", err)
		absPath = path
	}
	log.Printf("🔍 Загрузка архетипов из: %s", absPath)

	data, err := os.ReadFile(absPath)
	if err != nil {
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

// GenerateArchetype — выбирает климат по весу спектрального класса
// и формирует архетип с базовыми весами композиции.
func GenerateArchetype(spectralClass string, rng *rand.Rand) *Archetype {
	if archetypeCache == nil || len(archetypeCache.Climates) == 0 {
		return fallbackArchetype()
	}

	// 1. Выбор климата по весам
	var selectedClimate *ClimateConfig
	totalWeight := 0.0
	for i := range archetypeCache.Climates {
		if w, ok := archetypeCache.Climates[i].Weight[spectralClass]; ok {
			totalWeight += w
		}
	}

	if totalWeight == 0 {
		idx := rng.Intn(len(archetypeCache.Climates))
		selectedClimate = &archetypeCache.Climates[idx]
	} else {
		r := rng.Float64() * totalWeight
		for i := range archetypeCache.Climates {
			c := &archetypeCache.Climates[i]
			if w, ok := c.Weight[spectralClass]; ok {
				r -= w
				if r <= 0 {
					selectedClimate = c
					break
				}
			}
		}
		if selectedClimate == nil {
			selectedClimate = &archetypeCache.Climates[0]
		}
	}

	// 2. Гидросфера / атмосфера / биосфера — одиночный выбор из разрешённых
	hydro := pickRandom(selectedClimate.AllowedHydrospheres, rng, "сухая")
	atmo := pickRandom(selectedClimate.AllowedAtmospheres, rng, "разряженная")
	bio := pickRandom(selectedClimate.AllowedBiospheres, rng, "стерильная")

	// 3. Базовые веса композиции (копируем, чтобы не мутировать кэш)
	baseSurface := copyWeights(selectedClimate.BaseSurface)
	baseSubterrain := copyWeights(selectedClimate.BaseSubterrain)

	return &Archetype{
		ID:             selectedClimate.ID,
		Name:           selectedClimate.Name,
		Climate:        selectedClimate.ID,
		BaseSurface:    baseSurface,
		BaseSubterrain: baseSubterrain,
		Hydrosphere:    hydro,
		Atmosphere:     atmo,
		Biosphere:      bio,
		TemperatureMin: selectedClimate.TemperatureMin,
		TemperatureMax: selectedClimate.TemperatureMax,
		WaterChance:    selectedClimate.WaterChance,
		LifeChance:     selectedClimate.LifeChance,
		SizeMin:        0.5,
		SizeMax:        14.5,
		MassMin:        0.1,
		MassMax:        19.9,
	}
}

// fallbackArchetype — используется, если JSON не загрузился
func fallbackArchetype() *Archetype {
	return &Archetype{
		ID:      "fallback",
		Name:    "Землеподобная (fallback)",
		Climate: "temperate",
		BaseSurface: map[string]float64{
			SurfaceRocks:   0.4,
			SurfaceSands:   0.2,
			SurfaceOceans:  0.2,
			SurfaceLakes:   0.1,
			SurfaceForests: 0.1,
		},
		BaseSubterrain: map[string]float64{
			SubterrainEmptyRock:       0.3,
			SubterrainMagmaticRocks:   0.15,
			SubterrainSedimentaryRocks: 0.15,
			SubterrainOreVeins:        0.15,
			SubterrainGroundwater:     0.15,
			SubterrainCrystalVeins:    0.1,
		},
		Hydrosphere:    "океаны",
		Atmosphere:     "азотно-кислородная",
		Biosphere:      "растительная",
		TemperatureMin: 200,
		TemperatureMax: 350,
		WaterChance:    0.7,
		LifeChance:     0.4,
		SizeMin:        0.5,
		SizeMax:        14.5,
		MassMin:        0.1,
		MassMax:        19.9,
	}
}

// pickRandom — случайный элемент из списка, либо fallback если пусто
func pickRandom(items []string, rng *rand.Rand, fallback string) string {
	if len(items) == 0 {
		return fallback
	}
	return items[rng.Intn(len(items))]
}

// copyWeights — глубокая копия map, чтобы не мутировать кэш
func copyWeights(src map[string]float64) map[string]float64 {
	if src == nil {
		return map[string]float64{}
	}
	dst := make(map[string]float64, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}