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
//
// Пороги: 2.0× и 0.3× от ожидаемого.
// Это значит, что аномалия сработает только при серьёзном перекосе:
//   - если планета типа X в 2 раза чаще, чем ожидалось;
//   - или в 3 раза реже.
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

// expectedTypeShares — ожидаемые доли типов планет в галактике.
//
// Значения откалиброваны под текущий баланс (после правок сессии 2026-09-10).
// Реальные цифры (планет из 3113):
//
//	ледяная:       28%  (865)
//	газовый гигант: 19%  (604)
//	скалистая:      14%  (446)
//	пустынная:      12%  (374)
//	землеподобная:  12%  (366)
//	вулканическая:   9%  (286)
//	океаническая:    4%  (134)
//	радиоактивная:   1%  (38)
//	стеклянная:     <1%  (24)
//	металлическая:  <1%  (14)
//	органик:         0%  (0)
//
// Сумма близка к 1.0.
func expectedTypeShares() map[string]float64 {
	return map[string]float64{
		planet.TypeIce:         0.28,
		planet.TypeGasGiant:    0.19,
		planet.TypeRocky:       0.14,
		planet.TypeDesert:      0.12,
		planet.TypeEarthlike:   0.12,
		planet.TypeVolcanic:    0.09,
		planet.TypeOceanic:     0.04,
		planet.TypeRadioactive: 0.01,
		planet.TypeGlass:       0.005,
		planet.TypeMetal:       0.005,
		planet.TypeOrganic:     0.005,
	}
}