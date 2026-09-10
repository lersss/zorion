// internal/generator/planet/planet_data_generate.go
package planet

import (
	"encoding/json"
	"math"

	"github.com/google/uuid"
	"zorion/internal/names"
)

// determinePlanetCount — сколько планет у звезды данного класса
func (g *Generator) determinePlanetCount(spectralClass string) int {
	switch spectralClass {
	case "O", "B", "A":
		return g.rng.Intn(9)
	case "F", "G":
		return 2 + g.rng.Intn(7)
	case "K", "M":
		return g.rng.Intn(7)
	default:
		return g.rng.Intn(5)
	}
}

// computeEffectiveTemp — физическая температура планеты.
// Формула: T_eff = T_star * (1 / (r^2 * L))^0.25
// где r = 0.4 * 1.7^orbitIndex, L — светимость в солнечных единицах.
func computeEffectiveTemp(starTemp int, orbitIndex int, spectralClass string) float64 {
	luminosityMap := map[string]float64{
		"O": 1000, "B": 100, "A": 10, "F": 2, "G": 1,
		"K": 0.1, "M": 0.01, "L": 0.001, "T": 0.0001, "Y": 0.00001,
	}
	L := luminosityMap[spectralClass]
	if L <= 0 {
		L = 1.0
	}
	r := 0.4 * math.Pow(1.7, float64(orbitIndex))
	if r <= 0 {
		r = 0.4
	}
	ratio := 1.0 / (r * r * L)
	if ratio < 0 {
		ratio = 0
	}
	T := float64(starTemp) * math.Pow(ratio, 0.25)
	if T < 10 {
		T = 10
	}
	if T > 5000 {
		T = 5000
	}
	return T
}

// ==================== ОБЫЧНАЯ ПЛАНЕТА ====================

// generatePlanet — обычная планета на основе архетипа.
func (g *Generator) generatePlanet(
	worldID string,
	orbitIndex int,
	spectralClass string,
	starTemp int,
) *PlanetData {
	// --- ГАЗОВЫЙ ГИГАНТ (для горячих звёзд на дальних орбитах) ---
	if (spectralClass == "O" || spectralClass == "B" || spectralClass == "A") && orbitIndex >= 3 {
		if g.rng.Float64() < 0.8 {
			return g.generateGasGiant(worldID, orbitIndex, spectralClass, starTemp)
		}
	}

	// --- ОКЕАНИЧЕСКАЯ ПЛАНЕТА (2.5%) ---
	if g.rng.Float64() < 0.025 {
		return g.generateOceanicPlanet(worldID, orbitIndex, spectralClass, starTemp)
	}

	// --- РАДИОАКТИВНАЯ ПЛАНЕТА (1.5%, для горячих 3%) ---
	hotStars := map[string]bool{"O": true, "B": true, "A": true}
	chance := 0.015
	if hotStars[spectralClass] {
		chance = 0.03
	}
	if g.rng.Float64() < chance {
		return g.generateRadioactivePlanet(worldID, orbitIndex, spectralClass, starTemp)
	}

	// --- СТАНДАРТНАЯ ГЕНЕРАЦИЯ ЧЕРЕЗ АРХЕТИП ---
	archetype := GenerateArchetype(spectralClass, g.rng)
	props := GenerateProperties(archetype, orbitIndex, spectralClass, starTemp, g.rng)

	name := names.GeneratePlanetName(g.rng, g.usedNames)
	if name == "" {
		name = "Планета-" + uuidShort()
	}

	dominant := props.SurfaceComposition.DominantForm()
	if dominant == "" {
		dominant = SurfaceRocks
	}

	data := map[string]interface{}{
		// Основные параметры
		"size":              props.Size,
		"mass":              props.Mass,
		"atmosphere":        props.Atmosphere,
		"hydrosphere":       archetype.Hydrosphere,
		"biosphere":         archetype.Biosphere,
		"temperature":       props.Temperature,
		"water_percent":     props.WaterPercent,
		"habitable":         props.Habitable,
		"life":              props.Life,
		"population":        props.Population,
		"political_system":  props.Political,
		"conflict_level":    props.ConflictLevel,
		"moons":             props.Moons,
		"development_level": props.Development,
		"climate":           archetype.Climate,

		// Композиции
		"surface_composition":    composeToJSON(props.SurfaceComposition),
		"subterrain_composition": composeToJSON(props.SubterrainComposition),
		"surface_dominant":       dominant,
		"type":                   dominant, // для обратной совместимости

		// Описание
		"description": generateDescription(
			g.rng, dominant, props.Habitable, props.Life,
		),
	}

	dataJSON, _ := json.Marshal(data)

	return &PlanetData{
		ID:         uuid.New().String(),
		WorldID:    worldID,
		Name:       name,
		OrbitIndex: orbitIndex,
		Data:       dataJSON,
	}
}

// ==================== ОКЕАНИЧЕСКАЯ ПЛАНЕТА ====================

func (g *Generator) generateOceanicPlanet(
	worldID string,
	orbitIndex int,
	spectralClass string,
	starTemp int,
) *PlanetData {
	name := names.GeneratePlanetName(g.rng, g.usedNames)
	if name == "" {
		name = "Океаническая-" + uuidShort()
	}

	atmospheres := []string{"азотно-кислородная", "плотная"}
	atmosphere := atmospheres[g.rng.Intn(len(atmospheres))]
	size := 0.8 + g.rng.Float64()*1.2
	mass := 0.5 + g.rng.Float64()*2.5
	temp := 273 + g.rng.Float64()*100
	waterPercent := 70 + g.rng.Float64()*29
	life := g.rng.Float64() < 0.7
	habitable := life

	// Композиция поверхности — океаны доминируют
	surfaceComp := Composition{
		SurfaceOceans:     60 + g.rng.Float64()*15,
		SurfaceLakes:      5 + g.rng.Float64()*10,
		SurfaceRocks:      5 + g.rng.Float64()*10,
		SurfaceSands:      5 + g.rng.Float64()*10,
		SurfaceCoralReefs: 3 + g.rng.Float64()*7,
	}.Normalize().NonZero()

	// Недра — осадочные, нефть, соляные купола
	subterrainComp := Composition{
		SubterrainSedimentaryRocks: 30,
		SubterrainOilPockets:       15,
		SubterrainSaltDomes:        10,
		SubterrainGroundwater:      20,
		SubterrainOreVeins:         15,
		SubterrainEmptyRock:        10,
	}.Normalize().NonZero()

	var population int64 = 0
	political := "нет"
	if life {
		basePop := int64(1000000 + g.rng.Float64()*999000000)
		dev := 0.1 + g.rng.Float64()*0.9
		population = int64(float64(basePop) * dev)
		systems := []string{
			"демократия", "диктатура", "теократия",
			"корпоратократия", "анархия", "ИИ-управление",
		}
		political = systems[g.rng.Intn(len(systems))]
	}

	data := map[string]interface{}{
		"size":                   size,
		"mass":                   mass,
		"atmosphere":             atmosphere,
		"hydrosphere":            "океаны",
		"biosphere":              "растительная",
		"temperature":            temp,
		"water_percent":          waterPercent,
		"habitable":              habitable,
		"life":                   life,
		"population":             population,
		"political_system":       political,
		"conflict_level":         0.0,
		"moons":                  int(size / 5),
		"development_level":      0.0,
		"climate":                "temperate",
		"surface_composition":    composeToJSON(surfaceComp),
		"subterrain_composition": composeToJSON(subterrainComp),
		"surface_dominant":       SurfaceOceans,
		"type":                   SurfaceOceans,
		"description":            "Планета, почти полностью покрытая океаном. Богатая морская экосистема.",
	}
	dataJSON, _ := json.Marshal(data)

	return &PlanetData{
		ID:         uuid.New().String(),
		WorldID:    worldID,
		Name:       name,
		OrbitIndex: orbitIndex,
		Data:       dataJSON,
	}
}

// ==================== РАДИОАКТИВНАЯ ПЛАНЕТА ====================

func (g *Generator) generateRadioactivePlanet(
	worldID string,
	orbitIndex int,
	spectralClass string,
	starTemp int,
) *PlanetData {
	name := names.GeneratePlanetName(g.rng, g.usedNames)
	if name == "" {
		name = "Радиоактивная-" + uuidShort()
	}

	surfaces := []string{SurfaceMetalFields, SurfaceGlassFields}
	surface := surfaces[g.rng.Intn(len(surfaces))]
	atmospheres := []string{"плотная", "ядовитая"}
	atmosphere := atmospheres[g.rng.Intn(len(atmospheres))]
	size := 0.5 + g.rng.Float64()*14.5
	mass := 0.1 + g.rng.Float64()*19.9

	baseTemp := computeEffectiveTemp(starTemp, orbitIndex, spectralClass)
	temp := baseTemp + 200 + g.rng.Float64()*200
	if temp > 1200 {
		temp = 1200
	}

	waterPercent := 0.0
	if g.rng.Float64() < 0.1 {
		waterPercent = g.rng.Float64() * 20
	}
	life := g.rng.Float64() < 0.05

	// Композиция поверхности
	surfaceComp := Composition{
		surface:              50 + g.rng.Float64()*20,
		SurfaceRocks:         15 + g.rng.Float64()*10,
		SurfaceCraters:       10 + g.rng.Float64()*10,
		SurfaceVolcanicFields: 5 + g.rng.Float64()*10,
	}.Normalize().NonZero()

	// Недра — радиоактивные, металлические ядра
	subterrainComp := Composition{
		SubterrainRadioactiveZones: 30,
		SubterrainMetalCores:       20,
		SubterrainRareEarthVeins:   20,
		SubterrainMagmaticRocks:    15,
		SubterrainOreVeins:         15,
	}.Normalize().NonZero()

	data := map[string]interface{}{
		"size":                   size,
		"mass":                   mass,
		"atmosphere":             atmosphere,
		"hydrosphere":            "сухая",
		"biosphere":              "стерильная",
		"temperature":            temp,
		"water_percent":          waterPercent,
		"habitable":              false,
		"life":                   life,
		"population":             0,
		"political_system":       "нет",
		"conflict_level":         0.0,
		"moons":                  int(size / 8),
		"development_level":      0.0,
		"climate":                "extreme",
		"radioactive":            true,
		"surface_composition":    composeToJSON(surfaceComp),
		"subterrain_composition": composeToJSON(subterrainComp),
		"surface_dominant":       surface,
		"type":                   surface,
		"description":            "Планета с высоким радиационным фоном, богатая редкими элементами.",
	}
	dataJSON, _ := json.Marshal(data)

	return &PlanetData{
		ID:         uuid.New().String(),
		WorldID:    worldID,
		Name:       name,
		OrbitIndex: orbitIndex,
		Data:       dataJSON,
	}
}