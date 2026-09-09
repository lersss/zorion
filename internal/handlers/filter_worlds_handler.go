// internal/handlers/filter_worlds_handler.go
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"zorion/internal/models"
)

// FilterWorldsHandler возвращает миры с фильтрацией по планетам
func (h *AdminHandlers) FilterWorldsHandler(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	hasPlanets := queryParams.Get("has_planets") == "true"
	hasLife := queryParams.Get("has_life") == "true"
	hasHabitable := queryParams.Get("has_habitable") == "true"
	planetType := queryParams.Get("planet_type")
	resourceCategory := queryParams.Get("resource_category")

	sqlQuery := `
		SELECT w.id, w.name, w.coord_x, w.coord_y, w.spectral_class, w.temperature, w.created_at, w.updated_at
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
		sqlQuery += ` AND EXISTS (SELECT 1 FROM planets p WHERE p.world_id = w.id AND p.data->>'type' = $` + strconv.Itoa(argCounter) + `)`
		args = append(args, planetType)
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

	rows, err := h.db.Query(sqlQuery, args...)
	if err != nil {
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
			http.Error(w, "Scan error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		worlds = append(worlds, world)
	}
	if err = rows.Err(); err != nil {
		http.Error(w, "Rows error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(worlds)
}