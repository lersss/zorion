// internal/generator/planet/planet_data.go
package planet

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"zorion/internal/repository"
	"zorion/internal/resource"
)

// PlanetData — одна планета перед вставкой в БД
type PlanetData struct {
	ID         string
	WorldID    string
	Name       string
	OrbitIndex int
	Data       []byte
}

// Generator — генератор планет для мира
type Generator struct {
	db           *sql.DB
	rng          *rand.Rand
	ecoRepo      *repository.EconomyRepository
	resourceRepo *repository.ResourceRepository
	usedNames    map[string]bool
}

// NewGenerator — создаёт генератор. Если seed = 0 — берётся time.Now().
func NewGenerator(db *sql.DB, seed int64) *Generator {
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	return &Generator{
		db:           db,
		rng:          rand.New(rand.NewSource(seed)),
		ecoRepo:      repository.NewEconomyRepository(db),
		resourceRepo: repository.NewResourceRepository(db),
		usedNames:    make(map[string]bool),
	}
}

// GeneratePlanetsForWorld — генерирует все планеты мира и сохраняет их в БД.
//
// Параметр temperature (температура звезды) больше не используется —
// новая физика считает температуру планеты от светимости звезды и орбиты.
// Оставлен для обратной совместимости с хендлером.
func (g *Generator) GeneratePlanetsForWorld(worldID, spectralClass string, temperature int) (int, error) {
	planetCount := g.determinePlanetCount(spectralClass)
	if planetCount == 0 {
		return 0, nil
	}

	// Возраст системы — один на все планеты мира
	systemAge := determineSystemAge(spectralClass, g.rng)

	tx, err := g.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	batch := newBatchBuffers(planetCount)

	for i := 0; i < planetCount; i++ {
		orbitIndex := i + 1
		planet := g.generatePlanet(worldID, orbitIndex, spectralClass, systemAge)
		batch.addPlanet(planet)

		if err := g.collectEconomy(
			planet.ID, planet.Data, spectralClass,
			&batch.settlementRows, &batch.factoryRows, &batch.goodsRows,
		); err != nil {
			return 0, err
		}

		g.collectResources(planet.ID, planet.Data, spectralClass, &batch.resourceRows)
	}

	if err := g.flushBatch(tx, batch); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return planetCount, nil
}

// ==================== БАТЧ-БУФЕР ====================

type batchBuffers struct {
	planetRows     []interface{}
	settlementRows []interface{}
	factoryRows    []interface{}
	goodsRows      []interface{}
	resourceRows   []interface{}
}

func newBatchBuffers(planetCount int) *batchBuffers {
	return &batchBuffers{
		planetRows:     make([]interface{}, 0, planetCount),
		settlementRows: make([]interface{}, 0, planetCount),
		factoryRows:    make([]interface{}, 0, planetCount*2),
		goodsRows:      make([]interface{}, 0, planetCount*4),
		resourceRows:   make([]interface{}, 0, planetCount*4),
	}
}

func (b *batchBuffers) addPlanet(p *PlanetData) {
	b.planetRows = append(b.planetRows, []interface{}{
		p.ID,
		p.WorldID,
		p.Name,
		p.OrbitIndex,
		p.Data,
		time.Now(),
		time.Now(),
	})
}

func (g *Generator) flushBatch(tx *sql.Tx, b *batchBuffers) error {
	if err := g.batchInsertPlanets(tx, b.planetRows); err != nil {
		return err
	}
	if len(b.settlementRows) > 0 {
		if err := g.batchInsertSettlements(tx, b.settlementRows); err != nil {
			return err
		}
	}
	if len(b.factoryRows) > 0 {
		if err := g.batchInsertFactories(tx, b.factoryRows); err != nil {
			return err
		}
	}
	if len(b.goodsRows) > 0 {
		if err := g.batchInsertGoods(tx, b.goodsRows); err != nil {
			return err
		}
	}
	if len(b.resourceRows) > 0 {
		if err := g.batchInsertResources(tx, b.resourceRows); err != nil {
			return err
		}
	}
	return nil
}

// ==================== РЕСУРСЫ ====================

func (g *Generator) collectResources(
	planetID string,
	dataJSON []byte,
	spectralClass string,
	rows *[]interface{},
) {
	var data map[string]interface{}
	if err := json.Unmarshal(dataJSON, &data); err != nil {
		return
	}

	if isGasGiant(data) {
		return
	}

	dominant := getString(data, "surface_dominant")
	if dominant == "" {
		dominant = SurfaceRocks
	}

	subterrain := extractSubterrainComposition(data)

	resources := resource.GenerateResources(
		planetID,
		dominant,
		subterrain,
		spectralClass,
		g.rng,
	)

	for _, res := range resources {
		*rows = append(*rows, []interface{}{
			res.ID,
			res.PlanetID,
			res.Name,
			res.Category,
			res.Hardness,
			res.Elasticity,
			res.Conductivity,
			res.HeatResistance,
			res.ChemicalActivity,
			res.Density,
			res.Biocompatibility,
			res.EnergyDensity,
			res.Volatility,
			time.Now(),
			time.Now(),
		})
	}
}

func extractSubterrainComposition(data map[string]interface{}) map[string]float64 {
	result := map[string]float64{}
	raw, ok := data["subterrain_composition"].(map[string]interface{})
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

// ==================== ХЕЛПЕРЫ ДЛЯ JSON ====================

func getString(data map[string]interface{}, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}

func getFloat(data map[string]interface{}, key string) float64 {
	if val, ok := data[key].(float64); ok {
		return val
	}
	return 0
}

func getBool(data map[string]interface{}, key string) bool {
	if val, ok := data[key].(bool); ok {
		return val
	}
	return false
}

func composeToJSON(c Composition) map[string]float64 {
	if c == nil {
		return map[string]float64{}
	}
	out := make(map[string]float64, len(c))
	for k, v := range c {
		out[k] = v
	}
	return out
}

func uuidShort() string {
	return uuid.New().String()[:8]
}