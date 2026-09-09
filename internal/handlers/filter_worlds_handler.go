// internal/handlers/filter_worlds_handler.go
package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"zorion/internal/models"
)

// FilterWorldsHandler возвращает миры с фильтрацией по планетам
func (h *AdminHandlers) FilterWorldsHandler(w http.ResponseWriter, r *http.Request) {
	// Перехват паники
	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("🔥 PANIC in FilterWorldsHandler: %v", rec)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}()

	queryParams := r.URL.Query()
	hasPlanets := queryParams.Get("has_planets") == "true"

	sqlQuery := `
		SELECT w.id, w.name, w.coord_x, w.coord_y, w.spectral_class, w.temperature, w.created_at, w.updated_at
		FROM worlds w
		WHERE 1=1
	`
	args := []interface{}{}

	if hasPlanets {
		sqlQuery += ` AND EXISTS (SELECT 1 FROM planets p WHERE p.world_id = w.id)`
	}

	// Логируем запрос перед выполнением
	log.Printf("🔍 FilterWorldsHandler SQL: %s, args: %v", sqlQuery, args)

	rows, err := h.db.Query(sqlQuery, args...)
	if err != nil {
		log.Printf("❌ FilterWorldsHandler query error: %v", err)
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	worlds := []models.World{}
	for rows.Next() {
		var world models.World
		if err := rows.Scan(
			&world.ID,
			&world.Name,
			&world.CoordX,
			&world.CoordY,
			&world.SpectralClass,
			&world.Temperature,
			&world.CreatedAt,
			&world.UpdatedAt,
		); err != nil {
			log.Printf("❌ FilterWorldsHandler scan error: %v", err)
			http.Error(w, "Scan error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		worlds = append(worlds, world)
	}
	if err = rows.Err(); err != nil {
		log.Printf("❌ FilterWorldsHandler rows error: %v", err)
		http.Error(w, "Rows error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(worlds); err != nil {
		log.Printf("❌ FilterWorldsHandler encode error: %v", err)
		http.Error(w, "Encode error: "+err.Error(), http.StatusInternalServerError)
	}
}