// internal/audit/parser.go
package audit

// parsePlanetView — превращает JSON планеты из БД в PlanetView.
// Безопасно обрабатывает отсутствующие поля (возвращает нулевые значения).
func parsePlanetView(data map[string]interface{}) *PlanetView {
	p := &PlanetView{
		Raw:        data,
		Surface:    map[string]float64{},
		Subterrain: map[string]float64{},
	}

	// Идентификация (может быть перезаписано снаружи)
	p.ID = aStr(data, "id")
	p.Name = aStr(data, "name")
	p.WorldID = aStr(data, "world_id")

	// Физика
	p.Size = aFloat(data, "size")
	p.Mass = aFloat(data, "mass")
	p.Density = aFloat(data, "density")
	p.Temperature = aFloat(data, "temperature")
	p.WaterPercent = aFloat(data, "water_percent")

	// Типы и флаги
	p.Type = aStr(data, "type")
	p.SurfaceDominant = aStr(data, "surface_dominant")
	p.Climate = aStr(data, "climate")
	p.Atmosphere = aStr(data, "atmosphere")
	p.Hydrosphere = aStr(data, "hydrosphere")
	p.Biosphere = aStr(data, "biosphere")

	p.IsGasGiant = aBool(data, "is_gas_giant")
	p.IsRadioactive = aBool(data, "radioactive")
	p.Habitable = aBool(data, "habitable")
	p.Life = aBool(data, "life")
	p.Population = int64(aFloat(data, "population"))

	// Композиции
	p.Surface = aFloatMap(data, "surface_composition")
	p.Subterrain = aFloatMap(data, "subterrain_composition")

	// Ядро
	p.Core = parseCoreView(data)

	// Спутники
	p.Satellites = parseSatelliteViews(data)

	return p
}

// ==================== ЯДРО ====================

func parseCoreView(data map[string]interface{}) *CoreView {
	raw, ok := data["core"].(map[string]interface{})
	if !ok {
		return nil
	}
	return &CoreView{
		Type:          aStr(raw, "type"),
		MassPercent:   aFloat(raw, "mass_percent"),
		Activity:      aFloat(raw, "activity"),
		Radioactivity: aFloat(raw, "radioactivity"),
		Age:           aFloat(raw, "age"),
		IsActive:      aBool(raw, "is_active"),
		IsMetallic:    aBool(raw, "is_metallic"),
	}
}

// ==================== СПУТНИКИ ====================

func parseSatelliteViews(data map[string]interface{}) []SatelliteView {
	raw, ok := data["satellites"].([]interface{})
	if !ok {
		return nil
	}
	result := make([]SatelliteView, 0, len(raw))
	for _, item := range raw {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		result = append(result, SatelliteView{
			ID:           aStr(m, "id"),
			Name:         aStr(m, "name"),
			Mass:         aFloat(m, "mass"),
			Size:         aFloat(m, "size"),
			Temperature:  aFloat(m, "temperature"),
			WaterPercent: aFloat(m, "water_percent"),
			Atmosphere:   aStr(m, "atmosphere"),
			Habitable:    aBool(m, "habitable"),
			Life:         aBool(m, "life"),
		})
	}
	return result
}

// ==================== ХЕЛПЕРЫ ДОСТУПА ====================

// aStr — безопасное чтение string.
func aStr(data map[string]interface{}, key string) string {
	if v, ok := data[key].(string); ok {
		return v
	}
	return ""
}

// aFloat — безопасное чтение float64.
func aFloat(data map[string]interface{}, key string) float64 {
	if v, ok := data[key].(float64); ok {
		return v
	}
	return 0
}

// aBool — безопасное чтение bool.
func aBool(data map[string]interface{}, key string) bool {
	if v, ok := data[key].(bool); ok {
		return v
	}
	return false
}

// aFloatMap — безопасное чтение map[string]float64.
func aFloatMap(data map[string]interface{}, key string) map[string]float64 {
	result := map[string]float64{}
	raw, ok := data[key].(map[string]interface{})
	if !ok {
		return result
	}
	for k, v := range raw {
		if f, ok := v.(float64); ok {
			result[k] = f
		}
	}
	return result
}