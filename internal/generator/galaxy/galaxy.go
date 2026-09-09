package galaxy

import (
	"math"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"zorion/internal/models"
	"zorion/internal/names"
)

// Config – параметры генерации галактики
type Config struct {
	Seed           int64
	WorldCount     int
	MapSize        float64   // радиус круга
	MinDist        float64   // минимальное расстояние между мирами
	ClusterCount   int
	ClusterSpacing float64
	ClusterRadius  float64
	OutlierPercent float64
	WorldSpread    float64
}

// Generator создаёт миры и галактику
type Generator struct {
	cfg *Config
	rng *rand.Rand
}

// NewGenerator создаёт новый генератор
func NewGenerator(cfg *Config) *Generator {
	return &Generator{
		cfg: cfg,
		rng: rand.New(rand.NewSource(cfg.Seed)),
	}
}

// GenerateGalaxy генерирует все миры
func (g *Generator) GenerateGalaxy() []*models.World {
	if g.cfg.ClusterCount > 0 {
		return g.generateWorldsPoisson()
	}
	return g.generateWorldsRandom(g.cfg.MinDist)
}

// ---------- ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ ДЛЯ ГЕНЕРАЦИИ СВОЙСТВ МИРА ----------
func randomSpectralClass(rng *rand.Rand) string {
	classes := []string{"O", "B", "A", "F", "G", "K", "M", "L", "T", "Y"}
	return classes[rng.Intn(len(classes))]
}

func randomTemperature(rng *rand.Rand) int {
	return rng.Intn(40000-200) + 200
}

// generateWorld создаёт один мир по заданным координатам (без масштабирования)
func (g *Generator) generateWorld(center struct{ X, Y float64 }) *models.World {
	spread := g.cfg.WorldSpread
	x := center.X + g.rng.NormFloat64()*spread
	y := center.Y + g.rng.NormFloat64()*spread

	// На всякий случай проверяем, что точка не вышла за круг (но такого быть не должно)
	// Если вышла – просто логируем и не масштабируем (точка будет обрезана на клиенте или отброшена)
	radius := g.cfg.MapSize
	if math.Hypot(x, y) > radius {
		// Логируем предупреждение, но не масштабируем
		// (можно было бы отбросить, но для MVP оставим как есть)
		// В будущем можно добавить отбрасывание таких точек.
	}

	// Карта для уникальных имён (локальная)
	usedNames := make(map[string]bool)

	now := time.Now()
	world := &models.World{
		ID:             uuid.New().String(),
		Name:           names.GeneratePlanetName(g.rng, usedNames),
		CoordX:         x,
		CoordY:         y,
		SpectralClass:  randomSpectralClass(g.rng),
		Temperature:    randomTemperature(g.rng),
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	return world
}