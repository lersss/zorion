// internal/handlers/admin_compatibility_reset.go
package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"zorion/internal/models"
)

// compatibilityDefaultsPath — путь к JSON с дефолтами.
const compatibilityDefaultsPath = "config/compatibility_defaults.json"

// compatibilityDefaults — структура файла config/compatibility_defaults.json.
type compatibilityDefaults struct {
	Surface    map[string][]string `json:"surface"`
	Subterrain map[string][]string `json:"subterrain"`
}

// ResetMatrix — сбрасывает матрицу выбранной категории к дефолтам из JSON.
//
// Маршрут: POST /admin/compatibility/reset?category=surface|subterrain
//
// Шаги:
//  1. Читаем config/compatibility_defaults.json.
//  2. Очищаем категорию в БД.
//  3. Вставляем все запрещённые пары из дефолтов.
//  4. Пересобираем кеш.
func (h *CompatibilityHandlers) ResetMatrix(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	category := r.URL.Query().Get("category")
	if category == "" {
		category = models.CompatCategorySurface
	}
	if !isValidCategory(category) {
		http.Error(w, "invalid category", http.StatusBadRequest)
		return
	}

	// 1. Читаем дефолты
	defaults, err := loadCompatibilityDefaults()
	if err != nil {
		log.Printf("❌ load defaults: %v", err)
		http.Error(w, "failed to load defaults", http.StatusInternalServerError)
		return
	}

	// 2. Выбираем карту для категории
	var forbidMap map[string][]string
	switch category {
	case models.CompatCategorySurface:
		forbidMap = defaults.Surface
	case models.CompatCategorySubterrain:
		forbidMap = defaults.Subterrain
	}

	// 3. Очищаем категорию в БД
	if err := h.repo.ResetCategory(category); err != nil {
		log.Printf("❌ reset category: %v", err)
		http.Error(w, "failed to reset category", http.StatusInternalServerError)
		return
	}

	// 4. Вставляем дефолтные пары
	pairs := flattenForbidMap(forbidMap)
	if len(pairs) > 0 {
		if err := h.repo.SetBatch(category, pairs); err != nil {
			log.Printf("❌ insert defaults: %v", err)
			http.Error(w, "failed to insert defaults", http.StatusInternalServerError)
			return
		}
	}

	// 5. Пересобираем кеш
	if err := h.rebuildCache(); err != nil {
		log.Printf("⚠️ Cache rebuild failed: %v", err)
	}

	writeJSON(w, map[string]interface{}{
		"status":   "ok",
		"category": category,
		"restored": len(pairs),
	})
}

// ==================== ВСПОМОГАТЕЛЬНОЕ ====================

// loadCompatibilityDefaults — читает JSON-файл с дефолтами.
func loadCompatibilityDefaults() (*compatibilityDefaults, error) {
	absPath, err := filepath.Abs(compatibilityDefaultsPath)
	if err != nil {
		absPath = compatibilityDefaultsPath
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, err
	}

	var d compatibilityDefaults
	if err := json.Unmarshal(data, &d); err != nil {
		return nil, err
	}
	return &d, nil
}

// flattenForbidMap — превращает карту «форма → [запрещённые]» в плоский
// список пар для репозитория.
func flattenForbidMap(m map[string][]string) []models.CompatibilityUpdatePair {
	if len(m) == 0 {
		return nil
	}
	result := make([]models.CompatibilityUpdatePair, 0, len(m)*2)
	for a, list := range m {
		for _, b := range list {
			if a == b {
				continue // защита от самой себя
			}
			result = append(result, models.CompatibilityUpdatePair{
				TypeA:      a,
				TypeB:      b,
				Compatible: false, // все дефолты — запрещённые пары
			})
		}
	}
	return result
}