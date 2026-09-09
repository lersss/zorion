// internal/generator/planet/planet_data.go
package planet

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
	"zorion/internal/generator/resource"
	"zorion/internal/models"
	"zorion/internal/names"
	"zorion/internal/repository"
)

type Generator struct {
	db           *sql.DB
	rng          *rand.Rand
	ecoRepo      *repository.EconomyRepository
	resourceRepo *repository.ResourceRepository
	usedNames    map[string]bool
}

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

func (g *Generator) GeneratePlanetsForWorld(worldID, spectralClass string, temperature int) (int, error) {
	planetCount := g.determinePlanetCount(spectralClass)
	if planetCount == 0 {
		return 0, nil
	}

	tx, err := g.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	planetRows := make([]interface{}, 0, planetCount)
	settlementRows := make([]interface{}, 0, planetCount)
	factoryRows := make([]interface{}, 0, planetCount*2)
	goodsRows := make([]interface{}, 0, planetCount*4)
	resourceRows := make([]interface{}, 0, planetCount*4)

	planets := make([]*PlanetData, 0, planetCount)
	for i := 0; i < planetCount; i++ {
		orbitIndex := i + 1
		planet := g.generatePlanet(worldID, orbitIndex, spectralClass, temperature)
		planets = append(planets, planet)

		planetRows = append(planetRows, []interface{}{
			planet.ID,
			planet.WorldID,
			planet.Name,
			planet.OrbitIndex,
			planet.Data,
			time.Now(),
			time.Now(),
		})

		if err := g.collectEconomy(planet.ID, planet.Data, spectralClass, &settlementRows, &factoryRows, &goodsRows); err != nil {
			return 0, err
		}

		planetType, _ := g.getPlanetType(planet.Data)
		resources := resource.GenerateResources(planet.ID, planetType, spectralClass, g.rng)
		for _, res := range resources {
			resourceRows = append(resourceRows, []interface{}{
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

	if err := g.batchInsertPlanets(tx, planetRows); err != nil {
		return 0, err
	}
	if len(settlementRows) > 0 {
		if err := g.batchInsertSettlements(tx, settlementRows); err != nil {
			return 0, err
		}
	}
	if len(factoryRows) > 0 {
		if err := g.batchInsertFactories(tx, factoryRows); err != nil {
			return 0, err
		}
	}
	if len(goodsRows) > 0 {
		if err := g.batchInsertGoods(tx, goodsRows); err != nil {
			return 0, err
		}
	}
	if len(resourceRows) > 0 {
		if err := g.batchInsertResources(tx, resourceRows); err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return planetCount, nil
}

func (g *Generator) batchInsertPlanets(tx *sql.Tx, rows []interface{}) error {
	if len(rows) == 0 {
		return nil
	}
	valueStrings := make([]string, 0, len(rows))
	valueArgs := make([]interface{}, 0, len(rows)*7)
	for _, row := range rows {
		rowSlice := row.([]interface{})
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			len(valueArgs)+1, len(valueArgs)+2, len(valueArgs)+3, len(valueArgs)+4,
			len(valueArgs)+5, len(valueArgs)+6, len(valueArgs)+7))
		valueArgs = append(valueArgs, rowSlice...)
	}
	query := fmt.Sprintf("INSERT INTO public.planets (id, world_id, name, orbit_index, data, created_at, updated_at) VALUES %s",
		strings.Join(valueStrings, ","))
	_, err := tx.Exec(query, valueArgs...)
	return err
}

func (g *Generator) batchInsertSettlements(tx *sql.Tx, rows []interface{}) error {
	if len(rows) == 0 {
		return nil
	}
	valueStrings := make([]string, 0, len(rows))
	valueArgs := make([]interface{}, 0, len(rows)*6)
	for _, row := range rows {
		rowSlice := row.([]interface{})
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d)",
			len(valueArgs)+1, len(valueArgs)+2, len(valueArgs)+3, len(valueArgs)+4,
			len(valueArgs)+5, len(valueArgs)+6))
		valueArgs = append(valueArgs, rowSlice...)
	}
	query := fmt.Sprintf("INSERT INTO public.settlements (id, planet_id, level, population, capacity, stability) VALUES %s",
		strings.Join(valueStrings, ","))
	_, err := tx.Exec(query, valueArgs...)
	return err
}

func (g *Generator) batchInsertFactories(tx *sql.Tx, rows []interface{}) error {
	if len(rows) == 0 {
		return nil
	}
	valueStrings := make([]string, 0, len(rows))
	valueArgs := make([]interface{}, 0, len(rows)*8)
	for _, row := range rows {
		rowSlice := row.([]interface{})
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			len(valueArgs)+1, len(valueArgs)+2, len(valueArgs)+3, len(valueArgs)+4,
			len(valueArgs)+5, len(valueArgs)+6, len(valueArgs)+7, len(valueArgs)+8))
		valueArgs = append(valueArgs, rowSlice...)
	}
	query := fmt.Sprintf("INSERT INTO public.factories (id, planet_id, name, type, input_resource, output_product, quality, status) VALUES %s",
		strings.Join(valueStrings, ","))
	_, err := tx.Exec(query, valueArgs...)
	return err
}

func (g *Generator) batchInsertGoods(tx *sql.Tx, rows []interface{}) error {
	if len(rows) == 0 {
		return nil
	}
	valueStrings := make([]string, 0, len(rows))
	valueArgs := make([]interface{}, 0, len(rows)*7)
	for _, row := range rows {
		rowSlice := row.([]interface{})
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			len(valueArgs)+1, len(valueArgs)+2, len(valueArgs)+3, len(valueArgs)+4,
			len(valueArgs)+5, len(valueArgs)+6, len(valueArgs)+7))
		valueArgs = append(valueArgs, rowSlice...)
	}
	query := fmt.Sprintf("INSERT INTO public.goods_batches (id, planet_id, product_name, quantity, quality, producer_id, produced_at) VALUES %s",
		strings.Join(valueStrings, ","))
	_, err := tx.Exec(query, valueArgs...)
	return err
}

func (g *Generator) batchInsertResources(tx *sql.Tx, rows []interface{}) error {
	if len(rows) == 0 {
		return nil
	}
	valueStrings := make([]string, 0, len(rows))
	valueArgs := make([]interface{}, 0, len(rows)*15)
	for _, row := range rows {
		rowSlice := row.([]interface{})
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			len(valueArgs)+1, len(valueArgs)+2, len(valueArgs)+3, len(valueArgs)+4,
			len(valueArgs)+5, len(valueArgs)+6, len(valueArgs)+7, len(valueArgs)+8,
			len(valueArgs)+9, len(valueArgs)+10, len(valueArgs)+11, len(valueArgs)+12,
			len(valueArgs)+13, len(valueArgs)+14, len(valueArgs)+15))
		valueArgs = append(valueArgs, rowSlice...)
	}
	query := fmt.Sprintf(`INSERT INTO public.resources (
		id, planet_id, name, category, hardness, elasticity, conductivity,
		heat_resistance, chemical_activity, density, biocompatibility,
		energy_density, volatility, created_at, updated_at
	) VALUES %s`, strings.Join(valueStrings, ","))
	_, err := tx.Exec(query, valueArgs...)
	return err
}

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

// generatePlanet создаёт одну планету с учётом газовых гигантов и архетипов
func (g *Generator) generatePlanet(worldID string, orbitIndex int, spectralClass string, starTemp int) *PlanetData {
	// --- ЛОГИКА ГАЗОВЫХ ГИГАНТОВ (для горячих звёзд на дальних орбитах) ---
	if (spectralClass == "O" || spectralClass == "B" || spectralClass == "A") && orbitIndex >= 3 {
		if g.rng.Float64() < 0.8 {
			return g.generateGasGiant(worldID, orbitIndex, spectralClass, starTemp)
		}
	}

	// --- ЛОГИКА АРХЕТИПОВ (для всех остальных) ---
	archetype := GenerateArchetype(spectralClass, g.rng)
	props := g.GenerateProperties(archetype, orbitIndex, spectralClass, starTemp)

	name := names.GeneratePlanetName(g.rng, g.usedNames)
	if name == "" {
		name = "Планета-" + uuid.New().String()[:8]
	}

	data := map[string]interface{}{
		"type":              props.Type,
		"size":              props.Size,
		"mass":              props.Mass,
		"atmosphere":        props.Atmosphere,
		"temperature":       props.Temperature,
		"water_percent":     props.WaterPercent,
		"habitable":         props.Habitable,
		"life":              props.Life,
		"resources":         g.generateResourceCategories(spectralClass),
		"population":        props.Population,
		"political_system":  props.Political,
		"conflict_level":    props.ConflictLevel,
		"moons":             props.Moons,
		"description":       generateDescription(g.rng, props.Type, props.Habitable, props.Life),
		"development_level": props.Development,
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

// generateGasGiant создаёт газовый гигант с соответствующими параметрами
func (g *Generator) generateGasGiant(worldID string, orbitIndex int, spectralClass string, starTemp int) *PlanetData {
	name := names.GeneratePlanetName(g.rng, g.usedNames)
	if name == "" {
		name = "Газовый гигант-" + uuid.New().String()[:8]
	}

	// Размер и масса для газового гиганта (большие значения)
	size := 8 + g.rng.Float64()*20   // 8–28
	mass := 5 + g.rng.Float64()*15   // 5–20

	// Атмосфера — водородно-гелиевая
	atmospheres := []string{"водородно-гелиевая", "водородная", "гелиевая"}
	atmosphere := atmospheres[g.rng.Intn(len(atmospheres))]

	// Температура зависит от орбиты
	// Для газовых гигантов температура ниже, чем у архетипов
	baseTemp := float64(starTemp) * 0.3
	temp := baseTemp + g.rng.Float64()*100 - 50

	waterPercent := 0.0
	habitable := false
	life := false

	// У газовых гигантов может быть много спутников
	moons := 3 + g.rng.Intn(8)

	// Ресурсы — газовые гиганты богаты энергией и редкими газами
	resources := map[string]float64{
		"минералы": 0.0 + g.rng.Float64()*0.3,
		"энергия":  0.7 + g.rng.Float64()*0.3,
		"органика": 0.0 + g.rng.Float64()*0.2,
		"редкие":   0.5 + g.rng.Float64()*0.5,
	}

	description := "Огромная планета, состоящая в основном из водорода и гелия, с мощной атмосферой и множеством спутников."

	data := map[string]interface{}{
		"type":              "газовый гигант",
		"size":              size,
		"mass":              mass,
		"atmosphere":        atmosphere,
		"temperature":       temp,
		"water_percent":     waterPercent,
		"habitable":         habitable,
		"life":              life,
		"resources":         resources,
		"population":        0,
		"political_system":  "нет",
		"conflict_level":    0.0,
		"moons":             moons,
		"description":       description,
		"development_level": 0.0,
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

func (g *Generator) generateResourceCategories(spectralClass string) map[string]float64 {
	res := map[string]float64{
		"минералы": 0.0,
		"энергия":  0.0,
		"органика": 0.0,
		"редкие":   0.0,
	}
	var mineralsBase, energyBase, organicsBase, rareBase float64
	dispersion := 0.3

	switch spectralClass {
	case "O", "B", "A":
		mineralsBase = 0.7
		energyBase = 0.4
		organicsBase = 0.2
		rareBase = 0.8
	case "F", "G":
		mineralsBase = 0.5
		energyBase = 0.5
		organicsBase = 0.5
		rareBase = 0.5
	case "K", "M":
		mineralsBase = 0.3
		energyBase = 0.7
		organicsBase = 0.8
		rareBase = 0.3
	default:
		mineralsBase = 0.5
		energyBase = 0.5
		organicsBase = 0.5
		rareBase = 0.5
	}

	res["минералы"] = clamp(mineralsBase+(g.rng.Float64()-0.5)*dispersion, 0, 1)
	res["энергия"] = clamp(energyBase+(g.rng.Float64()-0.5)*dispersion, 0, 1)
	res["органика"] = clamp(organicsBase+(g.rng.Float64()-0.5)*dispersion, 0, 1)
	res["редкие"] = clamp(rareBase+(g.rng.Float64()-0.5)*dispersion, 0, 1)

	return res
}

func clamp(val, min, max float64) float64 {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

func (g *Generator) getPlanetType(data []byte) (string, error) {
	var d map[string]interface{}
	if err := json.Unmarshal(data, &d); err != nil {
		return "", err
	}
	t, ok := d["type"].(string)
	if !ok {
		return "землеподобная", nil
	}
	return t, nil
}

func (g *Generator) collectEconomy(planetID string, dataJSON []byte, spectralClass string,
	settlementRows, factoryRows, goodsRows *[]interface{}) error {
	var data map[string]interface{}
	if err := json.Unmarshal(dataJSON, &data); err != nil {
		return err
	}
	// Если планета — газовый гигант, пропускаем экономику
	if data["type"] == "газовый гигант" {
		return nil
	}

	habitable, _ := data["habitable"].(bool)
	life, _ := data["life"].(bool)
	populationRaw, _ := data["population"].(float64)
	population := int(populationRaw)

	if !life && population == 0 {
		return nil
	}

	level := 1
	if population > 1000000 {
		level = 2
	}
	if population > 10000000 {
		level = 3
	}
	if !habitable {
		level = 1
	}

	capacity := population * 2
	if capacity < 1000 {
		capacity = 1000
	}
	stability := 40 + g.rng.Intn(41)

	settlementID := uuid.New().String()
	*settlementRows = append(*settlementRows, []interface{}{
		settlementID,
		planetID,
		level,
		population,
		capacity,
		stability,
	})

	resources, ok := data["resources"].(map[string]interface{})
	if ok {
		recipes := map[string]struct {
			factoryType string
			output      string
		}{
			"минералы": {"добывающий", "металл"},
			"энергия":  {"добывающий", "энергоноситель"},
			"органика": {"перерабатывающий", "еда"},
			"редкие":   {"перерабатывающий", "компоненты"},
		}
		for resourceKey, value := range resources {
			if val, ok := value.(float64); ok && val > 0.3 {
				recipe, exists := recipes[resourceKey]
				if !exists {
					continue
				}
				factoryCount := 1 + g.rng.Intn(2)
				for i := 0; i < factoryCount; i++ {
					quality := 30 + g.rng.Intn(41)
					*factoryRows = append(*factoryRows, []interface{}{
						uuid.New().String(),
						planetID,
						fmt.Sprintf("%s завод %d", recipe.output, i+1),
						recipe.factoryType,
						resourceKey,
						recipe.output,
						quality,
						"active",
					})
				}
			}
		}
	}

	goods := []struct {
		name    string
		quality int
		qty     int
	}{
		{"еда", 30 + g.rng.Intn(41), 100 + g.rng.Intn(401)},
	}
	possibleGoods := []string{"металл", "энергоноситель", "компоненты", "инструменты"}
	for i := 0; i < 1+g.rng.Intn(3); i++ {
		goods = append(goods, struct {
			name    string
			quality int
			qty     int
		}{
			name:    possibleGoods[g.rng.Intn(len(possibleGoods))],
			quality: 30 + g.rng.Intn(41),
			qty:     50 + g.rng.Intn(201),
		})
	}
	for _, gd := range goods {
		*goodsRows = append(*goodsRows, []interface{}{
			uuid.New().String(),
			planetID,
			gd.name,
			gd.qty,
			gd.quality,
			nil,
			time.Now().Add(-24 * time.Hour),
		})
	}
	return nil
}

type PlanetData struct {
	ID         string
	WorldID    string
	Name       string
	OrbitIndex int
	Data       []byte
}

func generateDescription(rng *rand.Rand, planetType string, habitable, life bool) string {
	if life && habitable {
		adj := []string{"цветущий", "развитый", "мирный", "технологичный", "экологичный"}
		return fmt.Sprintf("%s мир с богатой биосферой", adj[rng.Intn(len(adj))])
	}
	if habitable {
		return "Потенциально пригодная для терраформирования планета."
	}
	return "Безжизненный и суровый мир."
}

// ---------- Функции для работы с архетипами (уже есть в других файлах) ----------
// Здесь предполагается, что GenerateArchetype и GenerateProperties определены в archetype.go и properties.go.
// Они не дублируются в этом файле.