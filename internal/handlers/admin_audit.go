// internal/handlers/admin_audit.go
package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"zorion/internal/audit/planet"
)

// GetAuditHandler — HTTP-обработчик аудита планет.
//
// Маршрут: GET /admin/audit
//
// Загружает все планеты из БД, прогоняет через правила аудита,
// возвращает агрегированный результат (счётчики, примеры проблем).
func (h *AdminHandlers) GetAuditHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rows, err := loadPlanetRowsForAudit(h)
	if err != nil {
		log.Printf("❌ audit: failed to load planets: %v", err)
		http.Error(w, "failed to load planets", http.StatusInternalServerError)
		return
	}

	result := planet.AuditRows(rows)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf("❌ audit: JSON encode failed: %v", err)
	}
}

// loadPlanetRowsForAudit — читает планеты из БД в минимальный формат.
//
// Возвращает id, world_id, name, data (JSON) — всё, что нужно аудитору.
func loadPlanetRowsForAudit(h *AdminHandlers) ([]planet.Row, error) {
	rows, err := h.db.Query(`
		SELECT id, world_id, name, data
		FROM planets
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []planet.Row{}
	for rows.Next() {
		var r planet.Row
		if err := rows.Scan(&r.ID, &r.WorldID, &r.Name, &r.Data); err != nil {
			continue
		}
		result = append(result, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}