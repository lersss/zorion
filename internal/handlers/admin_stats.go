// internal/handlers/admin_stats.go
package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

// PlanetStats — агрегированная статистика по планетам.
type PlanetStats struct {
	TotalWorlds       int                       `json:"total_worlds"`
	WorldsWithPlanets int                       `json:"worlds_with_planets"`
	TotalPlanets      int                       `json:"total_planets"`
	PlanetsByType     map[string]int            `json:"planets_by_type"`
	PlanetsBySpectral map[string]map[string]int `json:"planets_by_spectral"`
	GameDesignTypes   map[string]int            `json:"game_design_types"`
	SurfaceFormCounts map[string]int            `json:"surface_form_counts"`
	SubterrainCounts  map[string]int            `json:"subterrain_counts"`
	HydrosphereCount  map[string]int            `json:"hydrosphereCount"`
	AtmosphereCount   map[string]int            `json:"atmosphereCount"`
	BiosphereCount    map[string]int            `json:"biosphereCount"`
	HabitableCount    int                       `json:"habitable_count"`
	LifeCount         int                       `json:"life_count"`
	AvgSize           float64                   `json:"avg_size"`
	AvgMass           float64                   `json:"avg_mass"`
	AvgTemp           float64                   `json:"avg_temp"`
	AvgWater          float64                   `json:"avg_water"`
	AvgPopulation     int64                     `json:"avg_population"`
	Anomalies         []Anomaly                 `json:"anomalies"`
}

// Anomaly — отклонение от ожидаемого распределения.
type Anomaly struct {
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Value       float64 `json:"value"`
	Expected    float64 `json:"expected"`
	Severity    string  `json:"severity"`
}

// GetPlanetStatsHandler — HTTP-обработчик статистики.
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

// calculatePlanetStats — основной расчёт статистики.
func (h *AdminHandlers) calculatePlanetStats() (*PlanetStats, error) {
	stats := newPlanetStats()

	worlds, err := h.loadWorlds()
	if err != nil {
		return nil, err
	}
	stats.TotalWorlds = len(worlds)

	planets, err := h.loadPlanets()
	if err != nil {
		return nil, err
	}
	stats.TotalPlanets = len(planets)

	aggregator := newStatsAggregator(worlds)
	aggregator.process(planets, stats)
	aggregator.applyAverages(stats)

	detectWorldAnomalies(stats)
	detectTypeAnomalies(stats)

	return stats, nil
}

// newPlanetStats — инициализация со всеми map-полями.
func newPlanetStats() *PlanetStats {
	return &PlanetStats{
		PlanetsByType:     make(map[string]int),
		PlanetsBySpectral: make(map[string]map[string]int),
		GameDesignTypes:   make(map[string]int),
		SurfaceFormCounts: make(map[string]int),
		SubterrainCounts:  make(map[string]int),
		HydrosphereCount:  make(map[string]int),
		AtmosphereCount:   make(map[string]int),
		BiosphereCount:    make(map[string]int),
		Anomalies:         []Anomaly{},
	}
}