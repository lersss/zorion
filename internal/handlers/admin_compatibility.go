// internal/handlers/admin_compatibility.go
package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"zorion/internal/generator/planet"
	"zorion/internal/models"
	"zorion/internal/repository"
)

// CompatibilityHandlers — HTTP-хендлеры для матрицы совместимости.
//
// Маршруты:
//   - GET  /admin/compatibility?category=surface|subterrain
//   - POST /admin/compatibility
//   - POST /admin/compatibility/reset?category=surface|subterrain
type CompatibilityHandlers struct {
	repo *repository.CompatibilityRepository
}

// NewCompatibilityHandlers — конструктор.
func NewCompatibilityHandlers(db *sql.DB) *CompatibilityHandlers {
	return &CompatibilityHandlers{
		repo: repository.NewCompatibilityRepository(db),
	}
}

// HandleMatrix — диспетчер: GET → GetMatrix, POST → UpdateMatrix.
func (h *CompatibilityHandlers) HandleMatrix(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.GetMatrix(w, r)
	case http.MethodPost:
		h.UpdateMatrix(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// ==================== GET: ТЕКУЩАЯ МАТРИЦА ====================

// GetMatrix — отдаёт текущую матрицу заданной категории в формате
// {"category": "...", "forbid": {"форма": ["запрещённая1", ...]}, "total": N}.
func (h *CompatibilityHandlers) GetMatrix(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	if category == "" {
		category = models.CompatCategorySurface
	}
	if !isValidCategory(category) {
		http.Error(w, "invalid category", http.StatusBadRequest)
		return
	}

	pairs, err := h.repo.LoadAll(category)
	if err != nil {
		log.Printf("❌ LoadAll compatibility: %v", err)
		http.Error(w, "failed to load matrix", http.StatusInternalServerError)
		return
	}

	dto := models.CompatibilityMatrixDTO{
		Category: category,
		Forbid:   pairsToForbidMap(pairs),
		Total:    len(pairs),
	}

	writeJSON(w, dto)
}

// ==================== POST: ОБНОВЛЕНИЕ ====================

// UpdateMatrix — батч-обновление пар в категории.
func (h *CompatibilityHandlers) UpdateMatrix(w http.ResponseWriter, r *http.Request) {
	var req models.CompatibilityUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if !isValidCategory(req.Category) {
		http.Error(w, "invalid category", http.StatusBadRequest)
		return
	}
	if len(req.Pairs) == 0 {
		http.Error(w, "empty pairs", http.StatusBadRequest)
		return
	}

	for _, p := range req.Pairs {
		if p.TypeA == "" || p.TypeB == "" {
			http.Error(w, "type_a and type_b required", http.StatusBadRequest)
			return
		}
		if p.TypeA == p.TypeB {
			http.Error(w, "type_a and type_b must differ", http.StatusBadRequest)
			return
		}
	}

	if err := h.repo.SetBatch(req.Category, req.Pairs); err != nil {
		log.Printf("❌ SetBatch compatibility: %v", err)
		http.Error(w, "failed to update matrix", http.StatusInternalServerError)
		return
	}

	if err := h.rebuildCache(); err != nil {
		log.Printf("⚠️ Cache rebuild failed: %v", err)
	}

	writeJSON(w, map[string]interface{}{
		"status":  "ok",
		"updated": len(req.Pairs),
	})
}

// ==================== ВСПОМОГАТЕЛЬНОЕ ====================

// rebuildCache — читает обе категории из БД и пересобирает кеш в памяти.
func (h *CompatibilityHandlers) rebuildCache() error {
	surfaceMap, err := h.loadCategoryMap(models.CompatCategorySurface)
	if err != nil {
		return err
	}
	subterrainMap, err := h.loadCategoryMap(models.CompatCategorySubterrain)
	if err != nil {
		return err
	}
	planet.RebuildCompatibilityMatrix(surfaceMap, subterrainMap)
	return nil
}

// loadCategoryMap — загружает одну категорию из БД в формате «форма → [запрещённые]».
func (h *CompatibilityHandlers) loadCategoryMap(category string) (map[string][]string, error) {
	pairs, err := h.repo.LoadAll(category)
	if err != nil {
		return nil, err
	}
	return pairsToForbidMap(pairs), nil
}

// pairsToForbidMap — []Pair → map «type_a → [type_b, ...]».
func pairsToForbidMap(pairs []*models.CompatibilityPair) map[string][]string {
	result := map[string][]string{}
	for _, p := range pairs {
		result[p.TypeA] = append(result[p.TypeA], p.TypeB)
	}
	return result
}

// isValidCategory — проверка допустимых категорий.
func isValidCategory(cat string) bool {
	return cat == models.CompatCategorySurface || cat == models.CompatCategorySubterrain
}

// writeJSON — общий ответ JSON.
func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("⚠️ JSON encode failed: %v", err)
	}
}