// internal/generator/planet/planet_data_gas.go
package planet

import (
	"encoding/json"

	"github.com/google/uuid"
	"zorion/internal/names"
)

// generateGasGiant — газовый гигант. У него нет композиции поверхности
// (только атмосфера), но есть спутники — каждый полноценная локация.
func (g *Generator) generateGasGiant(
	worldID string,
	orbitIndex int,
	spectralClass string,
	starTemp int,
) *PlanetData {
	name := names.GeneratePlanetName(g.rng, g.usedNames)
	if name == "" {
		name = "Газовый гигант-" + uuidShort()
	}

	size := 8 + g.rng.Float64()*20
	mass := 5 + g.rng.Float64()*15

	atmospheres := []string{"водородно-гелиевая", "водородная", "гелиевая"}
	atmosphere := atmospheres[g.rng.Intn(len(atmospheres))]

	// Физическая температура: эффективная + внутренний нагрев
	baseTemp := computeEffectiveTemp(starTemp, orbitIndex, spectralClass)
	temp := baseTemp*(0.9+g.rng.Float64()*0.2) + 30

	// Количество спутников: 3–10
	satelliteCount := 3 + g.rng.Intn(8)

	// Генерируем спутники
	satellites := g.generateSatellites(satelliteCount, size, temp, spectralClass)

	// Сериализуем спутники в JSON
	satellitesJSON := make([]map[string]interface{}, 0, len(satellites))
	for _, sat := range satellites {
		satellitesJSON = append(satellitesJSON, satelliteToMap(sat))
	}

	// Ресурсы газового гиганта (атмосферные)
	resources := map[string]float64{
		"энергия": 0.7 + g.rng.Float64()*0.3,
		"редкие":  0.5 + g.rng.Float64()*0.5,
		"газы":    0.9 + g.rng.Float64()*0.1,
	}

	data := map[string]interface{}{
		"size":              size,
		"mass":              mass,
		"atmosphere":        atmosphere,
		"hydrosphere":       "сухая",
		"biosphere":         "стерильная",
		"temperature":       temp,
		"water_percent":     0.0,
		"habitable":         false,
		"life":              false,
		"population":        0,
		"political_system":  "нет",
		"conflict_level":    0.0,
		"moons":             satelliteCount,
		"development_level": 0.0,
		"climate":           "hot",
		"is_gas_giant":      true,
		"resources":         resources,
		"satellites":        satellitesJSON,
		// Обратная совместимость: dominant = "газовый_гигант"
		"surface_dominant": "газовый_гигант",
		"type":             "газовый гигант",
		"description":      "Огромная планета из водорода и гелия с множеством спутников.",
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

// satelliteToMap — превращает Satellite в map для JSON-сериализации.
func satelliteToMap(s *Satellite) map[string]interface{} {
	return map[string]interface{}{
		"id":                     s.ID,
		"name":                   s.Name,
		"orbit_index":            s.OrbitIndex,
		"size":                   s.Size,
		"mass":                   s.Mass,
		"temperature":            s.Temperature,
		"water_percent":          s.WaterPercent,
		"habitable":              s.Habitable,
		"life":                   s.Life,
		"atmosphere":             s.Atmosphere,
		"biosphere":              s.Biosphere,
		"surface_composition":    composeToJSON(s.SurfaceComposition),
		"subterrain_composition": composeToJSON(s.SubterrainComposition),
		"surface_dominant":       s.SurfaceComposition.DominantForm(),
		"description":            s.Description,
	}
}