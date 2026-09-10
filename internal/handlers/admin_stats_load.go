// internal/handlers/admin_stats_load.go
package handlers

import "encoding/json"

// worldInfo — минимальная информация о мире, нужная статистике.
type worldInfo struct {
	ID            string
	SpectralClass string
	Temperature   int
}

// planetRecord — одна планета из БД в виде map.
type planetRecord struct {
	WorldID string
	Data    map[string]interface{}
}

// loadWorlds — загружает список миров.
func (h *AdminHandlers) loadWorlds() ([]worldInfo, error) {
	rows, err := h.db.Query(`SELECT id, spectral_class, temperature FROM worlds`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []worldInfo{}
	for rows.Next() {
		var w worldInfo
		var coordX, coordY float64
		if err := rows.Scan(&w.ID, &coordX, &coordY, &w.SpectralClass, &w.Temperature); err != nil {
			continue
		}
		result = append(result, w)
	}
	return result, nil
}

// loadPlanets — загружает все планеты с распарсенным JSON.
func (h *AdminHandlers) loadPlanets() ([]planetRecord, error) {
	rows, err := h.db.Query(`SELECT world_id, data FROM planets`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []planetRecord{}
	for rows.Next() {
		var worldID string
		var dataJSON []byte
		if err := rows.Scan(&worldID, &dataJSON); err != nil {
			continue
		}
		var data map[string]interface{}
		if err := json.Unmarshal(dataJSON, &data); err != nil {
			continue
		}
		result = append(result, planetRecord{WorldID: worldID, Data: data})
	}
	return result, nil
}