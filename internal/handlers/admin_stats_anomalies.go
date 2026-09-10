// internal/handlers/admin_stats_anomalies.go
package handlers

import "zorion/internal/generator/planet"

// detectWorldAnomalies — проверяет аномалии на уровне миров.
func detectWorldAnomalies(stats *PlanetStats) {
	noPlanets := stats.TotalWorlds - stats.WorldsWithPlanets
	if stats.TotalWorlds > 0 && float64(noPlanets)/float64(stats.TotalWorlds) > 0.5 {
		stats.Anomalies = append(stats.Anomalies, Anomaly{
			Type:        "world",
			Description: "Слишком много миров без планет",
			Value:       float64(noPlanets),
			Expected:    float64(stats.TotalWorlds) * 0.3,
			Severity:    "high",
		})
	}
}

// detectTypeAnomalies — проверяет отклонения в распределении типов.
func detectTypeAnomalies(stats *PlanetStats) {
	if stats.TotalPlanets == 0 {
		return
	}

	expected := expectedTypeShares()

	for gdType, count := range stats.GameDesignTypes {
		exp, ok := expected[gdType]
		if !ok || exp == 0 {
			continue
		}
		expectedCount := float64(stats.TotalPlanets) * exp
		if expectedCount < 1 {
			continue
		}
		ratio := float64(count) / expectedCount
		if ratio > 2.0 {
			stats.Anomalies = append(stats.Anomalies, Anomaly{
				Type:        "game_design_type",
				Description: "Слишком много планет типа " + gdType,
				Value:       float64(count),
				Expected:    expectedCount,
				Severity:    "medium",
			})
		} else if ratio < 0.3 {
			stats.Anomalies = append(stats.Anomalies, Anomaly{
				Type:        "game_design_type",
				Description: "Слишком мало планет типа " + gdType,
				Value:       float64(count),
				Expected:    expectedCount,
				Severity:    "medium",
			})
		}
	}
}

// expectedTypeShares — ожидаемые доли типов (сумма = 1.0).
func expectedTypeShares() map[string]float64 {
	return map[string]float64{
		planet.TypeEarthlike:   0.03,
		planet.TypeOceanic:     0.05,
		planet.TypeIce:         0.15,
		planet.TypeVolcanic:    0.10,
		planet.TypeDesert:      0.10,
		planet.TypeGasGiant:    0.15,
		planet.TypeRadioactive: 0.02,
		planet.TypeGlass:       0.05,
		planet.TypeMetal:       0.03,
		planet.TypeOrganic:     0.10,
		planet.TypeRocky:       0.22,
	}
}