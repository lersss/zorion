// planet.go
// Библиотека для генерации процедурных планет с учётом климатических данных.
// Поддерживает типы: rocky, earth, ice, lava, gas (с полосами и случайным наклоном).
// Возвращает image.RGBA, который можно сохранить в PNG или использовать в игре.

package planet

import (
	"encoding/json"
	"image"
	"image/color"
	"math"
	"math/rand"
	"os"
	"time"
)

// ---------- Структуры для загрузки JSON с климатами ----------

// ClimateWeight - веса для спектральных типов звёзд (OBAFGKM)
type ClimateWeight struct {
	O float64 `json:"O"`
	B float64 `json:"B"`
	A float64 `json:"A"`
	F float64 `json:"F"`
	G float64 `json:"G"`
	K float64 `json:"K"`
	M float64 `json:"M"`
}

// Climate - описание одного климатического типа
type Climate struct {
	ID                 string        `json:"id"`
	Name               string        `json:"name"`
	Weight             ClimateWeight `json:"weight"`
	AllowedSurfaces    []string      `json:"allowed_surfaces"`
	AllowedHydrospheres []string    `json:"allowed_hydrospheres"`
	AllowedAtmospheres  []string    `json:"allowed_atmospheres"`
	AllowedBiospheres   []string    `json:"allowed_biospheres"`
	TemperatureMin      int          `json:"temperature_min"`
	TemperatureMax      int          `json:"temperature_max"`
	WaterChance         float64      `json:"water_chance"`
	LifeChance          float64      `json:"life_chance"`
}

// ClimateData - корневой объект JSON
type ClimateData struct {
	Climates []Climate `json:"climates"`
}

// ---------- Структура генератора ----------

// PlanetGenerator - основной объект для генерации планет
type PlanetGenerator struct {
	climates      []Climate
	canvasSize    int          // внутренний размер квадрата (по умолчанию 64)
	enableCache   bool
	cache         map[string]*CachedPlanet
	cacheOrder    []string
	maxCacheSize  int
	rand          *rand.Rand
}

// CachedPlanet - кэшированная планета
type CachedPlanet struct {
	Image *image.RGBA
	Meta  PlanetMeta
}

// PlanetMeta - метаданные сгенерированной планеты
type PlanetMeta struct {
	Type         string  // rocky, earth, ice, lava, gas
	Radius       int     // отображаемый радиус
	Seed         int64
	Climate      ClimateInfo
	Surface      string
	Hydrosphere  string
	Atmosphere   string
	Biosphere    string
	HasAtmosphere bool
	HasRings     bool
	StarType     string
	Temperature  float64
}

// ClimateInfo - сокращённая информация о климате
type ClimateInfo struct {
	ID   string
	Name string
	Temp float64
}

// ---------- Конструктор ----------

// NewPlanetGenerator создаёт новый генератор с загрузкой климатических данных из JSON-файла
func NewPlanetGenerator(climateFile string, opts ...func(*PlanetGenerator)) (*PlanetGenerator, error) {
	data, err := os.ReadFile(climateFile)
	if err != nil {
		return nil, err
	}
	var climateData ClimateData
	if err := json.Unmarshal(data, &climateData); err != nil {
		return nil, err
	}

	pg := &PlanetGenerator{
		climates:     climateData.Climates,
		canvasSize:   64,
		enableCache:  true,
		cache:        make(map[string]*CachedPlanet),
		cacheOrder:   []string{},
		maxCacheSize: 2000,
		rand:         rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	// Применяем опции
	for _, opt := range opts {
		opt(pg)
	}
	return pg, nil
}

// Опции для настройки генератора
func WithCanvasSize(size int) func(*PlanetGenerator) {
	return func(pg *PlanetGenerator) { pg.canvasSize = size }
}
func WithCacheEnabled(enabled bool) func(*PlanetGenerator) {
	return func(pg *PlanetGenerator) { pg.enableCache = enabled }
}
func WithMaxCacheSize(size int) func(*PlanetGenerator) {
	return func(pg *PlanetGenerator) { pg.maxCacheSize = size }
}
func WithSeed(seed int64) func(*PlanetGenerator) {
	return func(pg *PlanetGenerator) { pg.rand = rand.New(rand.NewSource(seed)) }
}

// ---------- Основной метод генерации ----------

// GeneratePlanet генерирует планету по заданным параметрам.
// Параметры:
//   - radius: отображаемый радиус (пикселей), рекомендуется 20-50
//   - starType: спектральный класс звезды (O,B,A,F,G,K,M) – если не указан, выбирается случайно
//   - climateID: конкретный id климата (если задан, starType игнорируется)
//   - surface: конкретная поверхность (если не задана, выбирается из allowed)
//   - hydrosphere: гидросфера (аналогично)
//   - atmosphere: атмосфера
//   - biosphere: биосфера
//   - hasRings: добавить кольца (если true, то всегда добавляются, иначе случайно для газовых)
func (pg *PlanetGenerator) GeneratePlanet(radius int, opts ...func(*GenerateOptions)) (*CachedPlanet, error) {
	// Подготовка параметров
	options := &GenerateOptions{
		Radius:     radius,
		StarType:   "",
		ClimateID:  "",
		Surface:    "",
		Hydrosphere: "",
		Atmosphere: "",
		Biosphere:  "",
		HasRings:   nil, // nil = автоматически
		Seed:       0,   // 0 = случайный
	}
	for _, opt := range opts {
		opt(options)
	}

	// Если seed не задан, генерируем случайный
	seed := options.Seed
	if seed == 0 {
		seed = pg.rand.Int63()
	}
	rng := rand.New(rand.NewSource(seed))

	// ---- Выбор климата ----
	var climate *Climate
	if options.ClimateID != "" {
		for i := range pg.climates {
			if pg.climates[i].ID == options.ClimateID {
				climate = &pg.climates[i]
				break
			}
		}
	}
	if climate == nil {
		// Выбираем на основе starType (или случайного спектрального класса)
		starType := options.StarType
		if starType == "" {
			starTypes := []string{"O", "B", "A", "F", "G", "K", "M"}
			starType = starTypes[rng.Intn(len(starTypes))]
		}
		climate = pg.selectClimateByStarType(starType, rng)
	}

	// ---- Выбор характеристик ----
	surface := options.Surface
	if surface == "" {
		surface = pickRandom(climate.AllowedSurfaces, rng)
	}
	hydrosphere := options.Hydrosphere
	if hydrosphere == "" {
		hydrosphere = pickRandom(climate.AllowedHydrospheres, rng)
	}
	atmosphere := options.Atmosphere
	if atmosphere == "" {
		atmosphere = pickRandom(climate.AllowedAtmospheres, rng)
	}
	biosphere := options.Biosphere
	if biosphere == "" {
		biosphere = pickRandom(climate.AllowedBiospheres, rng)
	}

	// ---- Вычисление температуры ----
	tempMin := float64(climate.TemperatureMin)
	tempMax := float64(climate.TemperatureMax)
	temperature := tempMin + rng.Float64()*(tempMax-tempMin)

	// ---- Определение визуального типа ----
	visualType := pg.determineVisualType(climate, surface, hydrosphere, temperature, rng)

	// ---- Наличие атмосферы ----
	hasAtmosphere := (atmosphere != "разряженная" && atmosphere != "")

	// ---- Кольца ----
	hasRings := false
	if options.HasRings != nil {
		hasRings = *options.HasRings
	} else {
		// Для газовых гигантов шанс 40%, для остальных 5%
		if visualType == "gas" && rng.Float64() < 0.4 {
			hasRings = true
		} else if visualType != "gas" && rng.Float64() < 0.05 {
			hasRings = true
		}
	}

	// ---- Генерация изображения ----
	size := pg.canvasSize
	img := pg.generateTexture(visualType, size, rng, options)

	// ---- Пост-обработка (тень, блик, атмосфера) ----
	pg.applyPostProcessing(img, size, visualType, hasAtmosphere, rng)

	// ---- Кольца (если есть) ----
	if hasRings {
		pg.drawRings(img, size, rng)
	}

	// ---- Масштабирование до нужного радиуса ----
	finalImg := pg.scaleImage(img, radius)

	// ---- Метаданные ----
	meta := PlanetMeta{
		Type:          visualType,
		Radius:        radius,
		Seed:          seed,
		Climate:       ClimateInfo{ID: climate.ID, Name: climate.Name, Temp: temperature},
		Surface:       surface,
		Hydrosphere:   hydrosphere,
		Atmosphere:    atmosphere,
		Biosphere:     biosphere,
		HasAtmosphere: hasAtmosphere,
		HasRings:      hasRings,
		StarType:      options.StarType,
		Temperature:   temperature,
	}

	planet := &CachedPlanet{
		Image: finalImg,
		Meta:  meta,
	}

	// ---- Кэширование (если включено) ----
	if pg.enableCache {
		key := hashParams(meta, options)
		if len(pg.cache) >= pg.maxCacheSize && pg.maxCacheSize > 0 {
			oldest := pg.cacheOrder[0]
			delete(pg.cache, oldest)
			pg.cacheOrder = pg.cacheOrder[1:]
		}
		pg.cache[key] = planet
		pg.cacheOrder = append(pg.cacheOrder, key)
	}

	return planet, nil
}

// ---------- Вспомогательные методы ----------

// selectClimateByStarType выбирает климат на основе спектрального типа и весов
func (pg *PlanetGenerator) selectClimateByStarType(starType string, rng *rand.Rand) *Climate {
	type weightedClimate struct {
		climate *Climate
		weight  float64
	}
	var weighted []weightedClimate
	for i := range pg.climates {
		c := &pg.climates[i]
		var w float64
		switch starType {
		case "O":
			w = c.Weight.O
		case "B":
			w = c.Weight.B
		case "A":
			w = c.Weight.A
		case "F":
			w = c.Weight.F
		case "G":
			w = c.Weight.G
		case "K":
			w = c.Weight.K
		case "M":
			w = c.Weight.M
		default:
			w = 0
		}
		if w > 0 {
			weighted = append(weighted, weightedClimate{climate: c, weight: w})
		}
	}
	if len(weighted) == 0 {
		// fallback: первый климат
		return &pg.climates[0]
	}
	// Случайный выбор с учётом весов
	total := 0.0
	for _, wc := range weighted {
		total += wc.weight
	}
	r := rng.Float64() * total
	for _, wc := range weighted {
		r -= wc.weight
		if r <= 0 {
			return wc.climate
		}
	}
	return weighted[len(weighted)-1].climate
}

// pickRandom выбирает случайный элемент из среза
func pickRandom(list []string, rng *rand.Rand) string {
	if len(list) == 0 {
		return ""
	}
	return list[rng.Intn(len(list))]
}

// determineVisualType определяет визуальный тип планеты на основе климата, поверхности, гидросферы и температуры
func (pg *PlanetGenerator) determineVisualType(climate *Climate, surface, hydrosphere string, temperature float64, rng *rand.Rand) string {
	// Приоритет: лавовая, ледяная, водная, иначе скалистая
	if surface == "лавовая" {
		return "lava"
	}
	if surface == "ледяная" {
		return "ice"
	}
	if temperature < 200 {
		return "ice"
	}
	// Если есть вода (океаны, озёра) и не сухая
	if hydrosphere == "океаны" || hydrosphere == "озёра" {
		return "earth"
	}
	// Если плотная атмосфера и есть признаки жизни -> earth, иначе rocky
	// Также можно добавить газовые гиганты, но в данных их нет, поэтому оставим rocky
	return "rocky"
}

// generateTexture создаёт текстуру планеты (без пост-эффектов) на основе визуального типа
func (pg *PlanetGenerator) generateTexture(visualType string, size int, rng *rand.Rand, opts *GenerateOptions) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	radius := float64(size)/2 - 2
	cx, cy := float64(size)/2, float64(size)/2

	// Заполняем прозрачным фоном
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - cx
			dy := float64(y) - cy
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist > radius {
				img.SetRGBA(x, y, color.RGBA{0, 0, 0, 0})
			}
		}
	}

	switch visualType {
	case "rocky":
		pg.generateRocky(img, size, rng)
	case "earth":
		pg.generateEarth(img, size, rng)
	case "ice":
		pg.generateIce(img, size, rng)
	case "lava":
		pg.generateLava(img, size, rng)
	case "gas":
		pg.generateGas(img, size, rng)
	default:
		pg.generateRocky(img, size, rng)
	}
	return img
}

// ---------- Генераторы конкретных типов (все работают с image.RGBA) ----------

func (pg *PlanetGenerator) generateRocky(img *image.RGBA, size int, rng *rand.Rand) {
	radius := float64(size)/2 - 2
	cx, cy := float64(size)/2, float64(size)/2
	baseR := uint8(100 + rng.Intn(60))
	baseG := uint8(80 + rng.Intn(50))
	baseB := uint8(60 + rng.Intn(40))

	// Кратеры
	craterCount := 3 + rng.Intn(8)
	type crater struct{ x, y, r float64; depth float64 }
	craters := make([]crater, craterCount)
	for i := 0; i < craterCount; i++ {
		angle := rng.Float64() * 2 * math.Pi
		dist := rng.Float64() * radius * 0.8
		craters[i] = crater{
			x: cx + math.Cos(angle)*dist,
			y: cy + math.Sin(angle)*dist,
			r: 2 + rng.Float64()*6,
			depth: 0.3 + rng.Float64()*0.5,
		}
	}

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - cx
			dy := float64(y) - cy
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist > radius {
				continue
			}
			// Шум
			noise1 := math.Sin(float64(x)*0.2+float64(y)*0.3)*10 + math.Cos(float64(x)*0.4-float64(y)*0.5)*8
			noise2 := math.Sin(float64(x)*0.7+float64(y)*0.9)*5
			r := float64(baseR) + noise1 + noise2
			g := float64(baseG) + noise1*0.8 + noise2*0.6
			b := float64(baseB) + noise1*0.5 + noise2*0.4

			// Кратеры
			for _, cr := range craters {
				ddx := float64(x) - cr.x
				ddy := float64(y) - cr.y
				d := math.Sqrt(ddx*ddx + ddy*ddy)
				if d < cr.r {
					factor := 1 - d/cr.r
					darken := factor * cr.depth * 30
					r -= darken
					g -= darken
					b -= darken
					if d > cr.r*0.7 {
						rim := (d - cr.r*0.7) / (cr.r * 0.3) * 10
						r += rim
						g += rim
						b += rim
					}
				}
			}
			clamp := func(v float64) uint8 {
				if v < 0 {
					return 0
				}
				if v > 255 {
					return 255
				}
				return uint8(v)
			}
			img.SetRGBA(x, y, color.RGBA{clamp(r), clamp(g), clamp(b), 255})
		}
	}
}

func (pg *PlanetGenerator) generateEarth(img *image.RGBA, size int, rng *rand.Rand) {
	radius := float64(size)/2 - 2
	cx, cy := float64(size)/2, float64(size)/2
	oceanR := uint8(20 + rng.Intn(30))
	oceanG := uint8(60 + rng.Intn(50))
	oceanB := uint8(120 + rng.Intn(60))

	// Континенты
	type continent struct{ x, y, r float64; color [3]uint8 }
	numContinents := 3 + rng.Intn(4)
	continents := make([]continent, numContinents)
	for i := 0; i < numContinents; i++ {
		angle := rng.Float64() * 2 * math.Pi
		dist := rng.Float64() * radius * 0.7
		continents[i] = continent{
			x: cx + math.Cos(angle)*dist,
			y: cy + math.Sin(angle)*dist,
			r: 5 + rng.Float64()*15,
			color: [3]uint8{
				uint8(50 + rng.Intn(80)),
				uint8(100 + rng.Intn(70)),
				uint8(30 + rng.Intn(50)),
			},
		}
	}

	// Облака
	type cloud struct{ x, y, r, alpha float64 }
	numClouds := 2 + rng.Intn(5)
	clouds := make([]cloud, numClouds)
	for i := 0; i < numClouds; i++ {
		angle := rng.Float64() * 2 * math.Pi
		dist := rng.Float64() * radius * 0.8
		clouds[i] = cloud{
			x: cx + math.Cos(angle)*dist,
			y: cy + math.Sin(angle)*dist,
			r: 4 + rng.Float64()*12,
			alpha: 0.1 + rng.Float64()*0.3,
		}
	}

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - cx
			dy := float64(y) - cy
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist > radius {
				continue
			}
			r := float64(oceanR)
			g := float64(oceanG)
			b := float64(oceanB)

			// Континенты
			for _, cont := range continents {
				ddx := float64(x) - cont.x
				ddy := float64(y) - cont.y
				d := math.Sqrt(ddx*ddx + ddy*ddy)
				if d < cont.r {
					factor := 1 - d/cont.r
					blend := factor*factor*0.9 + 0.1
					r = r*(1-blend) + float64(cont.color[0])*blend
					g = g*(1-blend) + float64(cont.color[1])*blend
					b = b*(1-blend) + float64(cont.color[2])*blend
				}
			}
			// Шум
			noise := math.Sin(float64(x)*0.5+float64(y)*0.7)*5 + math.Cos(float64(x)*0.9-float64(y)*0.3)*4
			r += noise
			g += noise
			b += noise

			// Облака
			for _, cl := range clouds {
				ddx := float64(x) - cl.x
				ddy := float64(y) - cl.y
				d := math.Sqrt(ddx*ddx + ddy*ddy)
				if d < cl.r {
					factor := 1 - d/cl.r
					alpha := factor * cl.alpha
					r = r*(1-alpha) + 255*alpha
					g = g*(1-alpha) + 255*alpha
					b = b*(1-alpha) + 255*alpha
				}
			}
			clamp := func(v float64) uint8 {
				if v < 0 {
					return 0
				}
				if v > 255 {
					return 255
				}
				return uint8(v)
			}
			img.SetRGBA(x, y, color.RGBA{clamp(r), clamp(g), clamp(b), 255})
		}
	}
}

func (pg *PlanetGenerator) generateIce(img *image.RGBA, size int, rng *rand.Rand) {
	radius := float64(size)/2 - 2
	cx, cy := float64(size)/2, float64(size)/2
	baseR := uint8(180 + rng.Intn(40))
	baseG := uint8(200 + rng.Intn(40))
	baseB := uint8(220 + rng.Intn(30))

	// Трещины
	type crack struct{ x1, y1, x2, y2, width float64 }
	crackCount := 5 + rng.Intn(10)
	cracks := make([]crack, crackCount)
	for i := 0; i < crackCount; i++ {
		angle := rng.Float64() * 2 * math.Pi
		dist := rng.Float64() * radius * 0.8
		x1 := cx + math.Cos(angle)*dist
		y1 := cy + math.Sin(angle)*dist
		angle2 := angle + (rng.Float64()-0.5)*1.2
		dist2 := dist + (rng.Float64()-0.5)*radius*0.5
		x2 := cx + math.Cos(angle2)*dist2
		y2 := cy + math.Sin(angle2)*dist2
		cracks[i] = crack{x1, y1, x2, y2, 1 + rng.Float64()*2}
	}

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - cx
			dy := float64(y) - cy
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist > radius {
				continue
			}
			r := float64(baseR)
			g := float64(baseG)
			b := float64(baseB)
			noise := math.Sin(float64(x)*0.3+float64(y)*0.5)*8 + math.Cos(float64(x)*0.7-float64(y)*0.2)*6
			r += noise
			g += noise
			b += noise

			// Трещины
			for _, cr := range cracks {
				// расстояние от точки до отрезка
				dx1 := float64(x) - cr.x1
				dy1 := float64(y) - cr.y1
				dx2 := cr.x2 - cr.x1
				dy2 := cr.y2 - cr.y1
				len2 := dx2*dx2 + dy2*dy2
				if len2 == 0 {
					continue
				}
				t := (dx1*dx2 + dy1*dy2) / len2
				if t < 0 {
					t = 0
				}
				if t > 1 {
					t = 1
				}
				projX := cr.x1 + t*dx2
				projY := cr.y1 + t*dy2
				d := math.Sqrt((float64(x)-projX)*(float64(x)-projX) + (float64(y)-projY)*(float64(y)-projY))
				if d < cr.width {
					factor := 1 - d/cr.width
					darken := factor * 30
					r -= darken
					g -= darken
					b -= darken
				}
			}
			clamp := func(v float64) uint8 {
				if v < 0 {
					return 0
				}
				if v > 255 {
					return 255
				}
				return uint8(v)
			}
			img.SetRGBA(x, y, color.RGBA{clamp(r), clamp(g), clamp(b), 255})
		}
	}
}

func (pg *PlanetGenerator) generateLava(img *image.RGBA, size int, rng *rand.Rand) {
	radius := float64(size)/2 - 2
	cx, cy := float64(size)/2, float64(size)/2
	baseR := uint8(20 + rng.Intn(20))
	baseG := uint8(10 + rng.Intn(10))
	baseB := uint8(10 + rng.Intn(10))

	// Лавовые трещины (извилистые линии)
	lavaCount := 4 + rng.Intn(6)
	lavas := make([][]struct{ x, y float64 }, lavaCount)
	for i := 0; i < lavaCount; i++ {
		startAngle := rng.Float64() * 2 * math.Pi
		startDist := rng.Float64() * radius * 0.9
		x1 := cx + math.Cos(startAngle)*startDist
		y1 := cy + math.Sin(startAngle)*startDist
		points := []struct{ x, y float64 }{{x1, y1}}
		curX, curY := x1, y1
		for j := 0; j < 4+rng.Intn(3); j++ {
			angle := math.Atan2(curY-cy, curX-cx) + (rng.Float64()-0.5)*1.5
			dist := math.Sqrt((curX-cx)*(curX-cx) + (curY-cy)*(curY-cy))
			newDist := dist + (rng.Float64()-0.5)*radius*0.3
			newX := cx + math.Cos(angle)*newDist
			newY := cy + math.Sin(angle)*newDist
			if math.Sqrt((newX-cx)*(newX-cx)+(newY-cy)*(newY-cy)) > radius {
				break
			}
			points = append(points, struct{ x, y float64 }{newX, newY})
			curX, curY = newX, newY
		}
		lavas[i] = points
	}

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - cx
			dy := float64(y) - cy
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist > radius {
				continue
			}
			r := float64(baseR)
			g := float64(baseG)
			b := float64(baseB)

			lavaGlow := 0.0
			for _, pts := range lavas {
				for i := 0; i < len(pts)-1; i++ {
					p1, p2 := pts[i], pts[i+1]
					dx1 := float64(x) - p1.x
					dy1 := float64(y) - p1.y
					dx2 := p2.x - p1.x
					dy2 := p2.y - p1.y
					len2 := dx2*dx2 + dy2*dy2
					if len2 == 0 {
						continue
					}
					t := (dx1*dx2 + dy1*dy2) / len2
					if t < 0 {
						t = 0
					}
					if t > 1 {
						t = 1
					}
					projX := p1.x + t*dx2
					projY := p1.y + t*dy2
					d := math.Sqrt((float64(x)-projX)*(float64(x)-projX) + (float64(y)-projY)*(float64(y)-projY))
					if d < 2.5 {
						factor := 1 - d/2.5
						if factor > lavaGlow {
							lavaGlow = factor
						}
					}
				}
			}
			if lavaGlow > 0 {
				lavaR := 255.0
				lavaG := 150.0 + rng.Float64()*50
				lavaB := 20.0 + rng.Float64()*30
				mix := lavaGlow*0.9 + 0.1
				r = r*(1-mix) + lavaR*mix
				g = g*(1-mix) + lavaG*mix
				b = b*(1-mix) + lavaB*mix
			}
			noise := math.Sin(float64(x)*0.5+float64(y)*0.6)*5 + math.Cos(float64(x)*0.8-float64(y)*0.4)*4
			r += noise
			g += noise
			b += noise
			clamp := func(v float64) uint8 {
				if v < 0 {
					return 0
				}
				if v > 255 {
					return 255
				}
				return uint8(v)
			}
			img.SetRGBA(x, y, color.RGBA{clamp(r), clamp(g), clamp(b), 255})
		}
	}
}

func (pg *PlanetGenerator) generateGas(img *image.RGBA, size int, rng *rand.Rand) {
	radius := float64(size)/2 - 2
	cx, cy := float64(size)/2, float64(size)/2

	// Полосы с разным наклоном (случайный угол)
	// Угол наклона (в радианах) от -0.8 до 0.8
	tiltAngle := (rng.Float64() - 0.5) * 1.6
	cosA := math.Cos(tiltAngle)
	sinA := math.Sin(tiltAngle)

	numBands := 4 + rng.Intn(6)
	type band struct {
		pos   float64 // положение вдоль оси, перпендикулярной полосам
		width float64
		color color.RGBA
	}
	bands := make([]band, numBands)
	baseHue := 10 + rng.Intn(40)
	for i := 0; i < numBands; i++ {
		// Позиция: от -radius до radius
		pos := -radius + (float64(i)/float64(numBands-1))*2*radius
		width := 3 + rng.Float64()*8
		// Цвет: вариация оттенка
		hue := baseHue + rng.Intn(30) - 15
		sat := 80
		lig := 40 + rng.Intn(30)
		// Преобразуем HSL в RGB (упрощённо)
		r, g, b := hslToRgb(hue, sat, lig)
		bands[i] = band{pos, width, color.RGBA{r, g, b, 255}}
	}

	// Большое пятно (вихрь)
	spotX := cx + (rng.Float64()-0.5)*radius*0.8
	spotY := cy + (rng.Float64()-0.5)*radius*0.6
	spotR := 3 + rng.Float64()*8
	spotCol := hslToRgb(int(baseHue+20), 90, 60)

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - cx
			dy := float64(y) - cy
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist > radius {
				continue
			}
			// Преобразуем координаты в систему, где полосы горизонтальны
			// Поворачиваем точку на угол -tiltAngle, чтобы полосы стали горизонтальными
			// Координата вдоль оси X не важна, важна координата Y' (перпендикулярно полосам)
			xRot := dx*cosA + dy*sinA
			yRot := -dx*sinA + dy*cosA

			// Определяем, в какой полосе находится пиксель
			r, g, b := float64(40), float64(30), float64(20) // фон
			for _, band := range bands {
				// Расстояние по Y' от центра полосы
				dyBand := yRot - band.pos
				if math.Abs(dyBand) < band.width/2 {
					factor := 1 - math.Abs(dyBand)/(band.width/2)
					mix := factor*0.8 + 0.2
					br := float64(band.color.R)
					bg := float64(band.color.G)
					bb := float64(band.color.B)
					r = r*(1-mix) + br*mix
					g = g*(1-mix) + bg*mix
					b = b*(1-mix) + bb*mix
				}
			}

			// Пятно
			dSpot := math.Sqrt((float64(x)-spotX)*(float64(x)-spotX) + (float64(y)-spotY)*(float64(y)-spotY))
			if dSpot < spotR {
				factor := 1 - dSpot/spotR
				r = r*(1-factor) + float64(spotCol.R)*factor
				g = g*(1-factor) + float64(spotCol.G)*factor
				b = b*(1-factor) + float64(spotCol.B)*factor
			}

			clamp := func(v float64) uint8 {
				if v < 0 {
					return 0
				}
				if v > 255 {
					return 255
				}
				return uint8(v)
			}
			img.SetRGBA(x, y, color.RGBA{clamp(r), clamp(g), clamp(b), 255})
		}
	}
}

// Вспомогательная функция преобразования HSL в RGB (упрощённая)
func hslToRgb(h, s, l int) (uint8, uint8, uint8) {
	// h: 0-360, s: 0-100, l: 0-100
	H := float64(h) / 360.0
	S := float64(s) / 100.0
	L := float64(l) / 100.0
	var r, g, b float64
	if S == 0 {
		r, g, b = L, L, L
	} else {
		var hue2rgb func(p, q, t float64) float64
		hue2rgb = func(p, q, t float64) float64 {
			if t < 0 {
				t += 1
			}
			if t > 1 {
				t -= 1
			}
			if t < 1.0/6.0 {
				return p + (q-p)*6*t
			}
			if t < 1.0/2.0 {
				return q
			}
			if t < 2.0/3.0 {
				return p + (q-p)*(2.0/3.0-t)*6
			}
			return p
		}
		q := 0.0
		if L < 0.5 {
			q = L * (1 + S)
		} else {
			q = L + S - L*S
		}
		p := 2*L - q
		r = hue2rgb(p, q, H+1.0/3.0)
		g = hue2rgb(p, q, H)
		b = hue2rgb(p, q, H-1.0/3.0)
	}
	return uint8(r * 255), uint8(g * 255), uint8(b * 255)
}

// ---------- Пост-обработка ----------

func (pg *PlanetGenerator) applyPostProcessing(img *image.RGBA, size int, visualType string, hasAtmosphere bool, rng *rand.Rand) {
	radius := float64(size)/2 - 2
	cx, cy := float64(size)/2, float64(size)/2

	// Определяем параметры эффектов в зависимости от типа
	shadowStrength := 0.6
	highlightStrength := 0.3
	atmStrength := 0.0
	atmColor := color.RGBA{100, 150, 255, 50}
	highlightColor := color.RGBA{255, 255, 255, 230}

	if hasAtmosphere {
		atmStrength = 0.2
	}
	switch visualType {
	case "ice":
		atmColor = color.RGBA{200, 230, 255, 40}
		highlightStrength = 0.8
		shadowStrength = 0.4
	case "lava":
		atmColor = color.RGBA{255, 100, 50, 60}
		highlightStrength = 0.2
		shadowStrength = 0.6
		if hasAtmosphere {
			atmStrength = 0.15
		}
	case "earth":
		atmColor = color.RGBA{70, 150, 255, 50}
		highlightStrength = 0.3
		shadowStrength = 0.6
	case "gas":
		atmColor = color.RGBA{200, 180, 150, 40}
		highlightStrength = 0.2
		shadowStrength = 0.6
	}

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - cx
			dy := float64(y) - cy
			dist := math.Sqrt(dx*dx + dy*dy)
			normDist := dist / radius

			if normDist > 1.2 {
				// За пределами планеты делаем прозрачным
				img.SetRGBA(x, y, color.RGBA{0, 0, 0, 0})
				continue
			}
			if normDist > 1 {
				// Внешняя атмосфера (свечение)
				if hasAtmosphere && atmStrength > 0 {
					outer := (normDist - 1) / 0.2
					alpha := float64(atmColor.A) / 255.0 * (1 - outer) * 0.5
					c := img.RGBAAt(x, y)
					r := float64(c.R)*(1-alpha) + float64(atmColor.R)*alpha
					g := float64(c.G)*(1-alpha) + float64(atmColor.G)*alpha
					b := float64(c.B)*(1-alpha) + float64(atmColor.B)*alpha
					img.SetRGBA(x, y, color.RGBA{uint8(r), uint8(g), uint8(b), uint8(float64(c.A)*(1-alpha) + 255*alpha)})
				}
				continue
			}

			// Получаем текущий цвет
			c := img.RGBAAt(x, y)

			// ---- Сферическая тень ----
			// Направление света: (-0.5, -0.4, 0.2)
			lightX, lightY, lightZ := -0.5, -0.4, 0.2
			lenL := math.Sqrt(lightX*lightX + lightY*lightY + lightZ*lightZ)
			nx, ny, nz := lightX/lenL, lightY/lenL, lightZ/lenL

			// Нормаль к сфере
			z := math.Sqrt(math.Max(0, radius*radius-dx*dx-dy*dy))
			normLen := math.Sqrt(dx*dx + dy*dy + z*z)
			if normLen == 0 {
				continue
			}
			normDx, normDy, normDz := dx/normLen, dy/normLen, z/normLen
			diffuse := normDx*nx + normDy*ny + normDz*nz
			if diffuse < 0 {
				diffuse = 0
			}
			if diffuse > 1 {
				diffuse = 1
			}
			shadow := 1 - shadowStrength*(1-diffuse)
			r := float64(c.R) * shadow
			g := float64(c.G) * shadow
			b := float64(c.B) * shadow

			// ---- Блик ----
			reflectX := 2*diffuse*normDx - nx
			reflectY := 2*diffuse*normDy - ny
			reflectZ := 2*diffuse*normDz - nz
			spec := math.Max(0, reflectZ)
			specIntensity := math.Pow(spec, 20) * highlightStrength * 2
			if specIntensity > 0.01 {
				hr := float64(highlightColor.R)
				hg := float64(highlightColor.G)
				hb := float64(highlightColor.B)
				r = r + (hr-r)*specIntensity
				g = g + (hg-g)*specIntensity
				b = b + (hb-b)*specIntensity
			}

			// ---- Внутренняя атмосфера (у края) ----
			if hasAtmosphere && atmStrength > 0 && normDist > 0.7 {
				edge := (normDist - 0.7) / 0.3
				atmAlpha := atmStrength * edge * 0.8
				ar := float64(atmColor.R)
				ag := float64(atmColor.G)
				ab := float64(atmColor.B)
				r = r*(1-atmAlpha) + ar*atmAlpha
				g = g*(1-atmAlpha) + ag*atmAlpha
				b = b*(1-atmAlpha) + ab*atmAlpha
			}

			clamp := func(v float64) uint8 {
				if v < 0 {
					return 0
				}
				if v > 255 {
					return 255
				}
				return uint8(v)
			}
			img.SetRGBA(x, y, color.RGBA{clamp(r), clamp(g), clamp(b), 255})
		}
	}
}

// ---------- Кольца ----------

func (pg *PlanetGenerator) drawRings(img *image.RGBA, size int, rng *rand.Rand) {
	// Рисуем кольца поверх планеты
	// Используем отдельный слой и смешиваем
	// Создаём временное изображение для колец
	ringImg := image.NewRGBA(img.Bounds())
	cx, cy := float64(size)/2, float64(size)/2
	radius := float64(size)/2 - 2

	// Параметры колец
	ringOuter := radius * 1.6
	ringInner := radius * 1.1
	tilt := 0.3 + rng.Float64()*0.4
	rotation := 0.2 + rng.Float64()*0.3
	alpha := 0.3 + rng.Float64()*0.3

	// Создаём градиент для колец
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			// Преобразуем координаты с учётом наклона и поворота
			dx := float64(x) - cx
			dy := float64(y) - cy
			// Поворот
			cosR := math.Cos(rotation)
			sinR := math.Sin(rotation)
			xRot := dx*cosR - dy*sinR
			yRot := dx*sinR + dy*cosR
			// Наклон (сжатие по Y)
			yRot = yRot * (1 / tilt)

			// Расстояние от центра
			dist := math.Sqrt(xRot*xRot + yRot*yRot)
			if dist >= ringInner && dist <= ringOuter {
				// Нормализованное положение в кольце (0..1)
				pos := (dist - ringInner) / (ringOuter - ringInner)
				// Прозрачность: fade in/out
				fade := 1.0
				if pos < 0.1 {
					fade = pos / 0.1
				} else if pos > 0.8 {
					fade = 1 - (pos-0.8)/0.2
				}
				if fade < 0 {
					fade = 0
				}
				// Цвет колец (желтовато-серый)
				baseColor := color.RGBA{200, 180, 150, uint8(alpha * 255 * fade)}
				// Смешиваем с существующим пикселем
				c := img.RGBAAt(x, y)
				// Если пиксель за пределами планеты, он прозрачный, тогда просто ставим цвет колец
				if c.A == 0 {
					img.SetRGBA(x, y, baseColor)
				} else {
					// Смешивание по alpha
					sa := float64(baseColor.A) / 255.0
					da := float64(c.A) / 255.0
					outA := sa + da*(1-sa)
					if outA > 0 {
						r := (float64(baseColor.R)*sa + float64(c.R)*da*(1-sa)) / outA
						g := (float64(baseColor.G)*sa + float64(c.G)*da*(1-sa)) / outA
						b := (float64(baseColor.B)*sa + float64(c.B)*da*(1-sa)) / outA
						img.SetRGBA(x, y, color.RGBA{uint8(r), uint8(g), uint8(b), uint8(outA * 255)})
					}
				}
			}
		}
	}
}

// ---------- Масштабирование ----------

func (pg *PlanetGenerator) scaleImage(src *image.RGBA, targetRadius int) *image.RGBA {
	srcSize := src.Bounds().Dx()
	targetSize := targetRadius * 2
	if targetSize == srcSize {
		return src
	}
	dst := image.NewRGBA(image.Rect(0, 0, targetSize, targetSize))
	// Простое масштабирование (ближайший сосед для скорости, можно добавить билинейную)
	for y := 0; y < targetSize; y++ {
		for x := 0; x < targetSize; x++ {
			srcX := int(float64(x) * float64(srcSize) / float64(targetSize))
			srcY := int(float64(y) * float64(srcSize) / float64(targetSize))
			if srcX >= srcSize {
				srcX = srcSize - 1
			}
			if srcY >= srcSize {
				srcY = srcSize - 1
			}
			dst.SetRGBA(x, y, src.RGBAAt(srcX, srcY))
		}
	}
	return dst
}

// ---------- Кэширование ----------

func hashParams(meta PlanetMeta, opts *GenerateOptions) string {
	// Простой хеш: конкатенация ключевых полей
	return fmt.Sprintf("%d-%s-%s-%s-%s-%s-%v-%v",
		meta.Seed,
		meta.Type,
		meta.Surface,
		meta.Hydrosphere,
		meta.Atmosphere,
		meta.Biosphere,
		meta.HasAtmosphere,
		meta.HasRings,
	)
}

// ---------- Дополнительные типы ----------

// GenerateOptions - параметры для генерации (аналогичны передаваемым опциям)
type GenerateOptions struct {
	Radius      int
	StarType    string
	ClimateID   string
	Surface     string
	Hydrosphere string
	Atmosphere  string
	Biosphere   string
	HasRings    *bool
	Seed        int64
}

// Опции для удобства
func WithStarType(starType string) func(*GenerateOptions) {
	return func(o *GenerateOptions) { o.StarType = starType }
}
func WithClimateID(id string) func(*GenerateOptions) {
	return func(o *GenerateOptions) { o.ClimateID = id }
}
func WithSurface(s string) func(*GenerateOptions) {
	return func(o *GenerateOptions) { o.Surface = s }
}
func WithHydrosphere(s string) func(*GenerateOptions) {
	return func(o *GenerateOptions) { o.Hydrosphere = s }
}
func WithAtmosphere(s string) func(*GenerateOptions) {
	return func(o *GenerateOptions) { o.Atmosphere = s }
}
func WithBiosphere(s string) func(*GenerateOptions) {
	return func(o *GenerateOptions) { o.Biosphere = s }
}
func WithRings(has bool) func(*GenerateOptions) {
	return func(o *GenerateOptions) { o.HasRings = &has }
}
func WithSeed(seed int64) func(*GenerateOptions) {
	return func(o *GenerateOptions) { o.Seed = seed }
}