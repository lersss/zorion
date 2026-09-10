// internal/handlers/filter_worlds_handler.go
package handlers

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// worldMapItem — облегчённая структура мира для карты.
// Не включает created_at / updated_at: фронту они не нужны, а на 100k миров
// это экономит несколько мегабайт трафика.
type worldMapItem struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	CoordX        float64 `json:"coord_x"`
	CoordY        float64 `json:"coord_y"`
	SpectralClass string  `json:"spectral_class"`
	Temperature   float64 `json:"temperature"`
}

// FilterWorldsHandler возвращает миры с фильтрацией по планетам.
func (h *AdminHandlers) FilterWorldsHandler(w http.ResponseWriter, r *http.Request) {
	tStart := time.Now()

	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("🔥 PANIC in FilterWorldsHandler: %v", rec)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}()

	queryParams := r.URL.Query()

	hasPlanets := queryParams.Get("has_planets") == "true"
	hasLife := queryParams.Get("has_life") == "true"
	hasHabitable := queryParams.Get("has_habitable") == "true"
	planetType := queryParams.Get("planet_type")
	resourceCategory := queryParams.Get("resource_category")

	// --- СБОРКА SQL ---
	sqlQuery := `
		SELECT w.id, w.name, w.coord_x, w.coord_y, w.spectral_class, w.temperature
		FROM worlds w
		WHERE 1=1
	`
	args := []interface{}{}
	argCounter := 1

	if hasPlanets {
		sqlQuery += ` AND EXISTS (SELECT 1 FROM planets p WHERE p.world_id = w.id)`
	}
	if hasLife {
		sqlQuery += ` AND EXISTS (SELECT 1 FROM planets p WHERE p.world_id = w.id AND (p.data->>'life')::boolean = true)`
	}
	if hasHabitable {
		sqlQuery += ` AND EXISTS (SELECT 1 FROM planets p WHERE p.world_id = w.id AND (p.data->>'habitable')::boolean = true)`
	}
	if planetType != "" {
		planetTypeLower := strings.ToLower(planetType)
		sqlQuery += ` AND EXISTS (SELECT 1 FROM planets p WHERE p.world_id = w.id AND LOWER(p.data->>'type') = $` + strconv.Itoa(argCounter) + `)`
		args = append(args, planetTypeLower)
		argCounter++
	}
	if resourceCategory != "" {
		sqlQuery += ` AND EXISTS (
			SELECT 1 FROM planets p
			WHERE p.world_id = w.id
			  AND p.data->'resources'->>$` + strconv.Itoa(argCounter) + ` IS NOT NULL
			  AND (p.data->'resources'->>$` + strconv.Itoa(argCounter) + `)::float > 0.3
		)`
		args = append(args, resourceCategory)
		argCounter++
	}

	// --- ЗАПРОС К БД ---
	tQuery := time.Now()
	rows, err := h.db.QueryContext(r.Context(), sqlQuery, args...)
	if err != nil {
		log.Printf("❌ FilterWorldsHandler query error: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	queryDur := time.Since(tQuery)

	// --- ЧТЕНИЕ СТРОК ---
	items := make([]worldMapItem, 0, 4096)
	for rows.Next() {
		var it worldMapItem
		if err := rows.Scan(
			&it.ID,
			&it.Name,
			&it.CoordX,
			&it.CoordY,
			&it.SpectralClass,
			&it.Temperature,
		); err != nil {
			log.Printf("❌ FilterWorldsHandler scan error: %v", err)
			http.Error(w, "Scan error", http.StatusInternalServerError)
			return
		}
		items = append(items, it)
	}
	if err = rows.Err(); err != nil {
		log.Printf("❌ FilterWorldsHandler rows error: %v", err)
		http.Error(w, "Rows error", http.StatusInternalServerError)
		return
	}
	scanDur := time.Since(tQuery) - queryDur

	// --- СЕРИАЛИЗАЦИЯ + ОТПРАВКА ---
	// Заголовки нужно выставить ДО первого Write — после уже поздно.
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")

	var writer io.Writer = w
	useGzip := strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")
	if useGzip {
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		writer = gz
	}

	tEncode := time.Now()
	// Если клиент отвалился — Encode вернёт ошибку, но http.Error уже
	// вызывать нельзя (заголовки улетели). Просто логируем и выходим.
	if err := json.NewEncoder(writer).Encode(items); err != nil {
		log.Printf("⚠️ FilterWorldsHandler encode error (клиент, вероятно, отвалился): %v", err)
		return
	}
	encodeDur := time.Since(tEncode)

	log.Printf(
		"🗺️  FilterWorlds: %d миров, query=%v, scan=%v, encode=%v, total=%v, gzip=%v",
		len(items),
		queryDur.Round(time.Millisecond),
		scanDur.Round(time.Millisecond),
		encodeDur.Round(time.Millisecond),
		time.Since(tStart).Round(time.Millisecond),
		useGzip,
	)
}