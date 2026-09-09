package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"zorion/internal/models"
)

type PlanetRepository struct {
	db *sql.DB
}

func NewPlanetRepository(db *sql.DB) *PlanetRepository {
	return &PlanetRepository{db: db}
}

func (r *PlanetRepository) GetPlanetsByWorldID(worldID string) ([]models.Planet, error) {
	query := `
		SELECT id, world_id, name, orbit_index, data, created_at, updated_at
		FROM planets
		WHERE world_id = $1
		ORDER BY orbit_index ASC
	`
	rows, err := r.db.Query(query, worldID)
	if err != nil {
		return nil, fmt.Errorf("failed to query planets: %w", err)
	}
	defer rows.Close()

	var planets []models.Planet
	for rows.Next() {
		var p models.Planet
		var dataJSON []byte
		err := rows.Scan(&p.ID, &p.WorldID, &p.Name, &p.OrbitIndex, &dataJSON, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan planet: %w", err)
		}

		// --- ЛОГИРОВАНИЕ ДЛЯ ДИАГНОСТИКИ ---
		var data map[string]interface{}
		if err := json.Unmarshal(dataJSON, &data); err != nil {
			log.Printf("❌ JSON unmarshal error for planet %s (world %s): %v", p.ID, p.WorldID, err)
			log.Printf("   Raw JSON (first 200 chars): %s", string(dataJSON[:min(200, len(dataJSON))]))
			return nil, fmt.Errorf("failed to unmarshal planet data: %w", err)
		}
		// ------------------------------------

		// Заполняем поля из JSON
		if val, ok := data["type"].(string); ok {
			p.Type = val
		}
		if val, ok := data["size"].(float64); ok {
			p.Size = val
		}
		if val, ok := data["mass"].(float64); ok {
			p.Mass = val
		}
		if val, ok := data["atmosphere"].(string); ok {
			p.Atmosphere = val
		}
		if val, ok := data["temperature"].(float64); ok {
			p.Temperature = val
		}
		if val, ok := data["water_percent"].(float64); ok {
			p.WaterPercent = val
		}
		if val, ok := data["habitable"].(bool); ok {
			p.Habitable = val
		}
		if val, ok := data["life"].(bool); ok {
			p.Life = val
		}
		if val, ok := data["population"].(float64); ok {
			p.Population = int64(val)
		}
		planets = append(planets, p)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	return planets, nil
}

// вспомогательная функция min для безопасного обрезания строки
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}