// internal/handlers/admin_stats_aggregator.go
package handlers

import "zorion/internal/generator/planet"

// statsAggregator — агрегирует статистику по планетам.
type statsAggregator struct {
	worldsByID map[string]worldInfo
	worldsSeen map[string]bool

	totalSize       float64
	totalMass       float64
	totalTemp       float64
	totalWater      float64
	totalPopulation int64

	sizeCount  int
	massCount  int
	tempCount  int
	waterCount int
	popCount   int
}

// newStatsAggregator — создаёт агрегатор с индексом миров.
func newStatsAggregator(worlds []worldInfo) *statsAggregator {
	byID := make(map[string]worldInfo, len(worlds))
	for _, w := range worlds {
		byID[w.ID] = w
	}
	return &statsAggregator{
		worldsByID: byID,
		worldsSeen: make(map[string]bool),
	}
}

// process — проходит по всем планетам, обновляя stats.
func (a *statsAggregator) process(planets []planetRecord, stats *PlanetStats) {
	for _, p := range planets {
		a.worldsSeen[p.WorldID] = true
		a.processOne(p, stats)
	}
	stats.WorldsWithPlanets = len(a.worldsSeen)
}

// processOne — обработка одной планеты.
func (a *statsAggregator) processOne(p planetRecord, stats *PlanetStats) {
	dominant := getString(p.Data, "surface_dominant")
	if dominant == "" {
		dominant = getString(p.Data, "type")
	}
	if dominant == "" {
		dominant = planet.SurfaceRocks
	}
	stats.PlanetsByType[dominant]++

	// По спектральному классу
	if w, ok := a.worldsByID[p.WorldID]; ok && w.SpectralClass != "" {
		if _, ok := stats.PlanetsBySpectral[w.SpectralClass]; !ok {
			stats.PlanetsBySpectral[w.SpectralClass] = make(map[string]int)
		}
		stats.PlanetsBySpectral[w.SpectralClass][dominant]++
	}

	// Гидросфера / атмосфера / биосфера
	incrementIfPresent(stats.HydrosphereCount, getString(p.Data, "hydrosphere"))
	incrementIfPresent(stats.AtmosphereCount, getString(p.Data, "atmosphere"))
	incrementIfPresent(stats.BiosphereCount, getString(p.Data, "biosphere"))

	// Композиция поверхности: count (наличие) + share (суммарный %)
	surface := extractComposition(p.Data, "surface_composition")
	for form, share := range surface {
		stats.SurfaceFormCounts[form]++
		stats.SurfaceFormShares[form] += share
	}

	// Композиция недр: count + share
	subterrain := extractComposition(p.Data, "subterrain_composition")
	for subType, share := range subterrain {
		stats.SubterrainCounts[subType]++
		stats.SubterrainShares[subType] += share
	}

	// Геймдизайнерский тип
	in := buildClassificationInput(p.Data, surface)
	gdType := planet.ClassifyGameDesignType(in)
	stats.GameDesignTypes[gdType]++

	// Счётчики
	if getBool(p.Data, "life") {
		stats.LifeCount++
	}
	if getBool(p.Data, "habitable") {
		stats.HabitableCount++
	}

	// Средние
	a.accumulateAverages(p.Data)
}

// accumulateAverages — накапливает суммы и счётчики для средних.
func (a *statsAggregator) accumulateAverages(data map[string]interface{}) {
	if size, ok := data["size"].(float64); ok {
		a.totalSize += size
		a.sizeCount++
	}
	if mass, ok := data["mass"].(float64); ok {
		a.totalMass += mass
		a.massCount++
	}
	if temp, ok := data["temperature"].(float64); ok {
		a.totalTemp += temp
		a.tempCount++
	}
	if water, ok := data["water_percent"].(float64); ok {
		a.totalWater += water
		a.waterCount++
	}
	if pop, ok := data["population"].(float64); ok {
		a.totalPopulation += int64(pop)
		a.popCount++
	}
}

// applyAverages — записывает средние значения в stats.
func (a *statsAggregator) applyAverages(stats *PlanetStats) {
	if a.sizeCount > 0 {
		stats.AvgSize = a.totalSize / float64(a.sizeCount)
	}
	if a.massCount > 0 {
		stats.AvgMass = a.totalMass / float64(a.massCount)
	}
	if a.tempCount > 0 {
		stats.AvgTemp = a.totalTemp / float64(a.tempCount)
	}
	if a.waterCount > 0 {
		stats.AvgWater = a.totalWater / float64(a.waterCount)
	}
	if a.popCount > 0 {
		stats.AvgPopulation = a.totalPopulation / int64(a.popCount)
	}
}