// internal/handlers/planet_image_handler.go
package handlers

import (
	"encoding/base64"
	"image/png"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"zorion/internal/generator/planet"
)

// PlanetImageHandler возвращает изображение планеты по параметрам
func PlanetImageHandler(w http.ResponseWriter, r *http.Request) {
	// Параметры запроса
	seedStr := r.URL.Query().Get("seed")
	starType := r.URL.Query().Get("starType")
	climateID := r.URL.Query().Get("climateId")
	radiusStr := r.URL.Query().Get("radius")
	if radiusStr == "" {
		radiusStr = "20"
	}
	radius, err := strconv.Atoi(radiusStr)
	if err != nil || radius < 1 || radius > 80 {
		radius = 20
	}
	var seed int64 = 0
	if seedStr != "" {
		seed, err = strconv.ParseInt(seedStr, 10, 64)
		if err != nil {
			seed = 0
		}
	}
	// Если seed не задан, используем случайный (но для детерминизма лучше передавать)
	if seed == 0 {
		seed = time.Now().UnixNano()
	}

	// Загружаем генератор (синглтон)
	gen, err := getPlanetGenerator()
	if err != nil {
		log.Printf("❌ Failed to get planet generator: %v", err)
		http.Error(w, "Generator not available", http.StatusInternalServerError)
		return
	}

	// Опции для генерации
	opts := []func(*planet.GenerateOptions){
		planet.WithRadius(radius),
		planet.WithSeed(seed),
	}
	if starType != "" {
		opts = append(opts, planet.WithStarType(starType))
	}
	if climateID != "" {
		opts = append(opts, planet.WithClimateID(climateID))
	}

	// Генерация
	cachedPlanet, err := gen.GeneratePlanet(radius, opts...)
	if err != nil {
		log.Printf("❌ Failed to generate planet: %v", err)
		http.Error(w, "Failed to generate planet", http.StatusInternalServerError)
		return
	}

	// Кодируем в PNG
	w.Header().Set("Content-Type", "image/png")
	if err := png.Encode(w, cachedPlanet.Image); err != nil {
		log.Printf("❌ Failed to encode image: %v", err)
		http.Error(w, "Failed to encode image", http.StatusInternalServerError)
		return
	}
}

// planetGeneratorSingleton — ленивая инициализация генератора
var (
	planetGenOnce sync.Once
	planetGen     *planet.PlanetGenerator
	planetGenErr  error
)

func getPlanetGenerator() (*planet.PlanetGenerator, error) {
	planetGenOnce.Do(func() {
		// Загружаем климатические данные из файла
		const climateFile = "config/planet_archetypes.json"
		pg, err := planet.NewPlanetGenerator(climateFile,
			planet.WithCanvasSize(64),
			planet.WithCacheEnabled(true),
			planet.WithMaxCacheSize(1000),
		)
		if err != nil {
			log.Printf("⚠️ Failed to load planet generator: %v", err)
			planetGenErr = err
			return
		}
		planetGen = pg
	})
	return planetGen, planetGenErr
}