// internal/handlers/admin_stats.go
package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

// PlanetStats содержит агрегированную статистику по планетам
type PlanetStats struct {
	TotalWorlds       int                      `json:"total_worlds"`
	WorldsWithPlanets int                      `json:"worlds_with_planets"`
	TotalPlanets      int                      `json:"total_planets"`
	PlanetsByType     map[string]int           `json:"planets_by_type"`
	PlanetsBySpectral map[string]map[string]int `json:"planets_by_spectral"`
	HabitableCount    int                      `json:"habitable_count"`
	LifeCount         int                      `json:"life_count"`
	AvgSize           float64                  `json:"avg_size"`
	AvgMass           float64                  `json:"avg_mass"`
	AvgTemp           float64                  `json:"avg_temp"`
	AvgWater          float64                  `json:"avg_water"`
	AvgPopulation     int64                    `json:"avg_population"`
	Anomalies         []Anomaly                `json:"anomalies"`
}

type Anomaly struct {
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Value       float64 `json:"value"`
	Expected    float64 `json:"expected"`
	Severity    string  `json:"severity"` // low, medium, high
}

// GetPlanetStatsHandler возвращает детальную статистику по планетам
func (h *AdminHandlers) GetPlanetStatsHandler(w http.ResponseWriter, r *http.Request) {
	stats, err := h.calculatePlanetStats()
	if err != nil {
		log.Printf("❌ Failed to calculate planet stats: %v", err)
		http.Error(w, "Failed to calculate stats", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (h *AdminHandlers) calculatePlanetStats() (*PlanetStats, error) {
	stats := &PlanetStats{
		PlanetsByType:      make(map[string]int),
		PlanetsBySpectral:  make(map[string]map[string]int),
		Anomalies:          []Anomaly{},
	}

	// Получаем все миры
	rows, err := h.db.Query(`
		SELECT id, coord_x, coord_y, spectral_class, temperature
		FROM worlds
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	worlds := []struct {
		ID            string
		SpectralClass string
		Temperature   int
	}{}
	for rows.Next() {
		var w struct {
			ID            string
			CoordX        float64
			CoordY        float64
			SpectralClass string
			Temperature   int
		}
		if err := rows.Scan(&w.ID, &w.CoordX, &w.CoordY, &w.SpectralClass, &w.Temperature); err != nil {
			continue
		}
		worlds = append(worlds, struct {
			ID            string
			SpectralClass string
			Temperature   int
		}{ID: w.ID, SpectralClass: w.SpectralClass, Temperature: w.Temperature})
	}
	stats.TotalWorlds = len(worlds)

	// Получаем все планеты
	planetRows, err := h.db.Query(`
		SELECT world_id, data
		FROM planets
	`)
	if err != nil {
		return nil, err
	}
	defer planetRows.Close()

	var planets []struct {
		WorldID string
		Data    map[string]interface{}
	}
	for planetRows.Next() {
		var worldID string
		var dataJSON []byte
		if err := planetRows.Scan(&worldID, &dataJSON); err != nil {
			continue
		}
		var data map[string]interface{}
		if err := json.Unmarshal(dataJSON, &data); err != nil {
			continue
		}
		planets = append(planets, struct {
			WorldID string
			Data    map[string]interface{}
		}{WorldID: worldID, Data: data})
	}
	stats.TotalPlanets = len(planets)

	// Считаем планеты с жизнью и обитаемые
	var totalSize, totalMass, totalTemp, totalWater float64
	var totalPopulation int64
	var sizeCount, massCount, tempCount, waterCount, popCount int

	worldsWithPlanets := make(map[string]bool)
	for _, p := range planets {
		worldsWithPlanets[p.WorldID] = true

		// Тип планеты
		if pType, ok := p.Data["type"].(string); ok {
			stats.PlanetsByType[pType]++
			// По спектральному классу
			spectral := ""
			for _, w := range worlds {
				if w.ID == p.WorldID {
					spectral = w.SpectralClass
					break
				}
			}
			if spectral != "" {
				if _, ok := stats.PlanetsBySpectral[spectral]; !ok {
					stats.PlanetsBySpectral[spectral] = make(map[string]int)
				}
				stats.PlanetsBySpectral[spectral][pType]++
			}
		}

		// Жизнь
		if life, ok := p.Data["life"].(bool); ok && life {
			stats.LifeCount++
		}
		// Обитаемость
		if habitable, ok := p.Data["habitable"].(bool); ok && habitable {
			stats.HabitableCount++
		}

		// Средние значения
		if size, ok := p.Data["size"].(float64); ok {
			totalSize += size
			sizeCount++
		}
		if mass, ok := p.Data["mass"].(float64); ok {
			totalMass += mass
			massCount++
		}
		if temp, ok := p.Data["temperature"].(float64); ok {
			totalTemp += temp
			tempCount++
		}
		if water, ok := p.Data["water_percent"].(float64); ok {
			totalWater += water
			waterCount++
		}
		if pop, ok := p.Data["population"].(float64); ok {
			totalPopulation += int64(pop)
			popCount++
		}
	}
	stats.WorldsWithPlanets = len(worldsWithPlanets)

	if sizeCount > 0 {
		stats.AvgSize = totalSize / float64(sizeCount)
	}
	if massCount > 0 {
		stats.AvgMass = totalMass / float64(massCount)
	}
	if tempCount > 0 {
		stats.AvgTemp = totalTemp / float64(tempCount)
	}
	if waterCount > 0 {
		stats.AvgWater = totalWater / float64(waterCount)
	}
	if popCount > 0 {
		stats.AvgPopulation = totalPopulation / int64(popCount)
	}

	// Аномалии
	noPlanets := stats.TotalWorlds - stats.WorldsWithPlanets
	if float64(noPlanets)/float64(stats.TotalWorlds) > 0.5 {
		stats.Anomalies = append(stats.Anomalies, Anomaly{
			Type:        "world",
			Description: "Слишком много миров без планет",
			Value:       float64(noPlanets),
			Expected:    float64(stats.TotalWorlds) * 0.3,
			Severity:    "high",
		})
	}

	expectedTypes := map[string]float64{
		"землеподобная":   0.2,
		"пустынная":       0.15,
		"ледяная":         0.15,
		"газовый гигант":  0.15,
		"океаническая":    0.1,
		"вулканическая":   0.1,
		"скалистая":       0.15,
	}
	if stats.TotalPlanets > 0 {
		for pType, count := range stats.PlanetsByType {
			expected := float64(stats.TotalPlanets) * expectedTypes[pType]
			if expected > 0 {
				ratio := float64(count) / expected
				if ratio > 1.5 {
					stats.Anomalies = append(stats.Anomalies, Anomaly{
						Type:        "planet_type",
						Description: "Слишком много планет типа " + pType,
						Value:       float64(count),
						Expected:    expected,
						Severity:    "medium",
					})
				} else if ratio < 0.5 {
					stats.Anomalies = append(stats.Anomalies, Anomaly{
						Type:        "planet_type",
						Description: "Слишком мало планет типа " + pType,
						Value:       float64(count),
						Expected:    expected,
						Severity:    "medium",
					})
				}
			}
		}
	}

	return stats, nil
}