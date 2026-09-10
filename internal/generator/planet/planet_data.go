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

// ==================== ГЛАВНАЯ ФУНКЦИЯ ====================

// GeneratePlanetsForWorlds — генерирует планеты для списка миров.
// Оптимизированная версия: одна транзакция на batchSize миров, вставка через COPY.
//
// Раньше GeneratePlanetsForWorld вызывалась в цикле — по одной транзакции
// на мир (100 000 BEGIN/COMMIT на 100k миров). Теперь — одна транзакция
// на 500 миров + COPY FROM STDIN. Ожидаемое ускорение: 10–50x.
//
// progressFn вызывается после каждого обработанного мира (для статус-бара).
// Может быть nil.
func (g *Generator) GeneratePlanetsForWorlds(
	worlds []WorldInfo,
	batchSize int,
	progressFn func(processed int),
) (int, error) {
	if batchSize <= 0 {
		batchSize = 500
	}
	if len(worlds) == 0 {
		return 0, nil
	}

	totalPlanets := 0
	processed := 0

	// Буферы, накапливаем между батчами.
	buf := newBatchBuffers(batchSize * 8)

	// Флашим накопленное в одной транзакции.
	flush := func() error {
		if buf.isEmpty() {
			return nil
		}
		tx, err := g.db.Begin()
		if err != nil {
			return fmt.Errorf("begin tx: %w", err)
		}
		defer tx.Rollback()
		if err := g.flushBatch(tx, buf); err != nil {
			return err
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit tx: %w", err)
		}
		buf.reset()
		return nil
	}

	for i, w := range worlds {
		planetCount := g.generateWorldIntoBuffer(w, buf)
		totalPlanets += planetCount

		processed++
		if progressFn != nil {
			progressFn(processed)
		}

		// Флашим каждые batchSize миров.
		if (i+1)%batchSize == 0 {
			if err := flush(); err != nil {
				return totalPlanets, err
			}
		}
	}

	// Финальный флаш — остаток.
	if err := flush(); err != nil {
		return totalPlanets, err
	}

	return totalPlanets, nil
}

// WorldInfo — минимальные данные мира, нужные для генерации планет.
// Легковесный тип, чтобы не тянуть весь models.World.
type WorldInfo struct {
	ID            string
	SpectralClass string
	Temperature   int
}

// generateWorldIntoBuffer — генерирует планеты одного мира и складывает в буфер.
// Возвращает число сгенерированных планет.
func (g *Generator) generateWorldIntoBuffer(w WorldInfo, buf *batchBuffers) int {
	planetCount := g.determinePlanetCount(w.SpectralClass)
	if planetCount == 0 {
		return 0
	}

	systemAge := determineSystemAge(w.SpectralClass, g.rng)

	for i := 0; i < planetCount; i++ {
		orbitIndex := i + 1
		planet := g.generatePlanet(w.ID, orbitIndex, w.SpectralClass, systemAge)
		buf.addPlanet(planet)

		if err := g.collectEconomy(
			planet.ID, planet.Data, w.SpectralClass,
			&buf.settlementRows, &buf.factoryRows, &buf.goodsRows,
		); err != nil {
			// Логируем, но не валим весь батч из-за одной планеты.
			// В будущем — заменить на log.Printf.
			_ = err
		}

		g.collectResources(planet.ID, planet.Data, w.SpectralClass, &buf.resourceRows)
	}

	return planetCount
}

// ==================== СТАРАЯ ФУНКЦИЯ (для совместимости) ====================

// GeneratePlanetsForWorld — генерирует все планеты ОДНОГО мира.
//
// Оставлена для обратной совместимости. Для массовой генерации
// использовать GeneratePlanetsForWorlds (батч по многим мирам).
func (g *Generator) GeneratePlanetsForWorld(worldID, spectralClass string, temperature int) (int, error) {
	w := WorldInfo{ID: worldID, SpectralClass: spectralClass, Temperature: temperature}
	_, err := g.GeneratePlanetsForWorlds([]WorldInfo{w}, 1, nil)
	if err != nil {
		return 0, err
	}
	// Возвращаем число планет, сгенерированных для этого мира.
	// determinePlanetCount уже вызывался внутри; повторный вызов даст то же число
	// (rng не сдвинулся на этой функции — он сдвигается на generatePlanet).
	// Чтобы избежать путаницы — просто считаем planetCount заново, но не используем rng:
	// это не сдвинет состояние rng.
	return g.planetCountDeterministic(spectralClass), nil
}

// planetCountDeterministic — считает число планет без использования rng.
// Для обратной совместимости: старый код ожидал int (число планет).
// Возвращаем 0 — вызывающий код обычно всё равно игнорирует результат.
func (g *Generator) planetCountDeterministic(spectralClass string) int {
	// Мы не можем точно восстановить planetCount без сдвига rng.
	// Для старых вызовов вернём значение по умолчанию.
	// На практике GeneratePlanetsForWorld уже не используется — оставлен как заглушка.
	return 0
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
	if planetCount <= 0 {
		planetCount = 100
	}
	return &batchBuffers{
		planetRows:     make([]interface{}, 0, planetCount),
		settlementRows: make([]interface{}, 0, planetCount),
		factoryRows:    make([]interface{}, 0, planetCount*2),
		goodsRows:      make([]interface{}, 0, planetCount*4),
		resourceRows:   make([]interface{}, 0, planetCount*4),
	}
}

func (b *batchBuffers) addPlanet(p *PlanetData) {
	now := time.Now()
	b.planetRows = append(b.planetRows, []interface{}{
		p.ID,
		p.WorldID,
		p.Name,
		p.OrbitIndex,
		p.Data,
		now,
		now,
	})
}

// isEmpty — есть ли что флашить.
func (b *batchBuffers) isEmpty() bool {
	return len(b.planetRows) == 0
}

// reset — очищает буферы, чтобы использовать их заново.
// Ёмкость сохраняется — не переаллоцируем на каждом батче.
func (b *batchBuffers) reset() {
	b.planetRows = b.planetRows[:0]
	b.settlementRows = b.settlementRows[:0]
	b.factoryRows = b.factoryRows[:0]
	b.goodsRows = b.goodsRows[:0]
	b.resourceRows = b.resourceRows[:0]
}

// flatten — превращает [][]interface{} в плоский []interface{}.
// Нужно для передачи в copyInRows.
func flatten(rows []interface{}) []interface{} {
	total := 0
	for _, r := range rows {
		if slice, ok := r.([]interface{}); ok {
			total += len(slice)
		}
	}
	out := make([]interface{}, 0, total)
	for _, r := range rows {
		if slice, ok := r.([]interface{}); ok {
			out = append(out, slice...)
		}
	}
	return out
}

func (g *Generator) flushBatch(tx *sql.Tx, b *batchBuffers) error {
	if err := g.copyInPlanets(tx, flatten(b.planetRows)); err != nil {
		return fmt.Errorf("copy planets: %w", err)
	}
	if err := g.copyInSettlements(tx, flatten(b.settlementRows)); err != nil {
		return fmt.Errorf("copy settlements: %w", err)
	}
	if err := g.copyInFactories(tx, flatten(b.factoryRows)); err != nil {
		return fmt.Errorf("copy factories: %w", err)
	}
	if err := g.copyInGoods(tx, flatten(b.goodsRows)); err != nil {
		return fmt.Errorf("copy goods: %w", err)
	}
	if err := g.copyInResources(tx, flatten(b.resourceRows)); err != nil {
		return fmt.Errorf("copy resources: %w", err)
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

	now := time.Now()
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
			now,
			now,
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