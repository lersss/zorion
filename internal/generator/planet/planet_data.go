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

func (g *Generator) generatePlanet(worldID string, orbitIndex int, spectralClass string, starTemp int) *PlanetData {
	// --- ГАЗОВЫЙ ГИГАНТ (для горячих звёзд на дальних орбитах) ---
	if (spectralClass == "O" || spectralClass == "B" || spectralClass == "A") && orbitIndex >= 3 {
		if g.rng.Float64() < 0.8 {
			return g.generateGasGiant(worldID, orbitIndex, spectralClass, starTemp)
		}
	}

	// --- ОКЕАНИЧЕСКАЯ ПЛАНЕТА (2.5% шанс) ---
	if g.rng.Float64() < 0.025 {
		return g.generateOceanicPlanet(worldID, orbitIndex, spectralClass, starTemp)
	}

	// --- РАДИОАКТИВНАЯ ПЛАНЕТА (1.5% шанс, для горячих звёзд 3%) ---
	hotStars := map[string]bool{"O": true, "B": true, "A": true}
	if hotStars[spectralClass] {
		if g.rng.Float64() < 0.03 {
			return g.generateRadioactivePlanet(worldID, orbitIndex, spectralClass, starTemp)
		}
	} else {
		if g.rng.Float64() < 0.015 {
			return g.generateRadioactivePlanet(worldID, orbitIndex, spectralClass, starTemp)
		}
	}

	// --- ОБЫЧНАЯ ПЛАНЕТА (архетипы) ---
	archetype := GenerateArchetype(spectralClass, g.rng)
	props := GenerateProperties(archetype, orbitIndex, spectralClass, starTemp, g.rng)

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

func (g *Generator) generateGasGiant(worldID string, orbitIndex int, spectralClass string, starTemp int) *PlanetData {
	name := names.GeneratePlanetName(g.rng, g.usedNames)
	if name == "" {
		name = "Газовый гигант-" + uuid.New().String()[:8]
	}
	size := 8 + g.rng.Float64()*20
	mass := 5 + g.rng.Float64()*15
	atmospheres := []string{"водородно-гелиевая", "водородная", "гелиевая"}
	atmosphere := atmospheres[g.rng.Intn(len(atmospheres))]
	baseTemp := float64(starTemp) * 0.3
	temp := baseTemp + g.rng.Float64()*100 - 50
	waterPercent := 0.0
	habitable := false
	life := false
	moons := 3 + g.rng.Intn(8)
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

func (g *Generator) generateOceanicPlanet(worldID string, orbitIndex int, spectralClass string, starTemp int) *PlanetData {
	name := names.GeneratePlanetName(g.rng, g.usedNames)
	if name == "" {
		name = "Океаническая-" + uuid.New().String()[:8]
	}
	// Поверхность: песчаная или глинистая
	surfaces := []string{"песчаная", "глинистая"}
	surface := surfaces[g.rng.Intn(len(surfaces))]
	// Гидросфера: океаны
	hydrosphere := "океаны"
	// Атмосфера: азотно-кислородная или плотная
	atmospheres := []string{"азотно-кислородная", "плотная"}
	atmosphere := atmospheres[g.rng.Intn(len(atmospheres))]
	// Размер и масса (0.8–2.0 R⊕, 0.5–3.0 M⊕)
	size := 0.8 + g.rng.Float64()*1.2
	mass := 0.5 + g.rng.Float64()*2.5
	// Температура: 0–100 °C -> 273–373 K
	temp := 273 + g.rng.Float64()*100
	// Вода: почти вся поверхность
	waterPercent := 70 + g.rng.Float64()*29
	// Жизнь: вероятна
	life := g.rng.Float64() < 0.7
	habitable := life // для океанических планет обитаемость = наличие жизни
	// Ресурсы: органические и минералы
	resources := map[string]float64{
		"минералы": 0.3 + g.rng.Float64()*0.5,
		"энергия":  0.1 + g.rng.Float64()*0.3,
		"органика": 0.6 + g.rng.Float64()*0.4,
		"редкие":   0.1 + g.rng.Float64()*0.2,
	}
	// Население и политика (если есть жизнь)
	var population int64 = 0
	political := "нет"
	if life {
		basePop := int64(1000000 + g.rng.Float64()*999000000)
		dev := 0.1 + g.rng.Float64()*0.9
		population = int64(float64(basePop) * dev)
		systems := []string{"демократия", "диктатура", "теократия", "корпоратократия", "анархия", "ИИ-управление"}
		political = systems[g.rng.Intn(len(systems))]
	}
	description := "Планета, почти полностью покрытая океаном. Богатая морская экосистема и влажный климат."
	data := map[string]interface{}{
		"type":              surface,
		"size":              size,
		"mass":              mass,
		"atmosphere":        atmosphere,
		"temperature":       temp,
		"water_percent":     waterPercent,
		"habitable":         habitable,
		"life":              life,
		"resources":         resources,
		"population":        population,
		"political_system":  political,
		"conflict_level":    0.0,
		"moons":             int(size / 5),
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

func (g *Generator) generateRadioactivePlanet(worldID string, orbitIndex int, spectralClass string, starTemp int) *PlanetData {
	name := names.GeneratePlanetName(g.rng, g.usedNames)
	if name == "" {
		name = "Радиоактивная-" + uuid.New().String()[:8]
	}
	// Поверхность: металлическая или реголитовая
	surfaces := []string{"металлическая", "реголитовая"}
	surface := surfaces[g.rng.Intn(len(surfaces))]
	// Атмосфера: плотная или ядовитая
	atmospheres := []string{"плотная", "ядовитая"}
	atmosphere := atmospheres[g.rng.Intn(len(atmospheres))]
	// Размер и масса (как у скалистых)
	size := 0.5 + g.rng.Float64()*14.5
	mass := 0.1 + g.rng.Float64()*19.9
	// Температура: высокая из-за внутреннего нагрева
	temp := 300 + g.rng.Float64()*400
	// Вода: маловероятна
	waterPercent := 0.0
	if g.rng.Float64() < 0.1 {
		waterPercent = g.rng.Float64() * 20
	}
	// Жизнь: маловероятна (микробная)
	life := false
	if g.rng.Float64() < 0.05 {
		life = true
	}
	habitable := false
	// Ресурсы: много редких и энергии
	resources := map[string]float64{
		"минералы": 0.2 + g.rng.Float64()*0.3,
		"энергия":  0.7 + g.rng.Float64()*0.3,
		"органика": 0.0 + g.rng.Float64()*0.1,
		"редкие":   0.8 + g.rng.Float64()*0.2,
	}
	description := "Планета с высоким радиационным фоном, богатая редкими радиоактивными элементами. Опасна для жизни без специальной защиты."
	data := map[string]interface{}{
		"type":              surface,
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
		"moons":             int(size / 8),
		"description":       description,
		"development_level": 0.0,
		"radioactive":       true, // маркер для классификации
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
	if data["type"] == "газовый гигант" || data["radioactive"] == true {
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