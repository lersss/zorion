// internal/generator/planet/composition_compatibility_reload.go
package planet

// RebuildCompatibilityMatrix — полностью пересобирает кеш матрицы совместимости
// из переданных данных. Формат совпадает с compatibility_defaults.json:
//   - surface:    {"лавовые_поля": ["ледники", "океаны"], ...}
//   - subterrain: {"магматические_камеры": ["подземные_льды"], ...}
//
// Используется после изменений в админке.
func RebuildCompatibilityMatrix(
	surface map[string][]string,
	subterrain map[string][]string,
) {
	if surface == nil {
		surface = map[string][]string{}
	}
	if subterrain == nil {
		subterrain = map[string][]string{}
	}

	m := &CompatibilityMatrix{
		Surface:    surface,
		Subterrain: subterrain,
	}
	m.buildIndexes()

	compatMu.Lock()
	compatMatrix = m
	compatMu.Unlock()
}

// ResetCompatibilityToDefaults — сбрасывает матрицу к встроенным дефолтам.
func ResetCompatibilityToDefaults() {
	compatMu.Lock()
	compatMatrix = defaultCompatibilityMatrix()
	compatMu.Unlock()
}

// ==================== ХЕЛПЕРЫ ДЛЯ ГРУППИРОВКИ ====================

// CompatibilityPairInput — минимальная структура пары для построения матрицы.
type CompatibilityPairInput struct {
	Category string
	TypeA    string
	TypeB    string
}

// PairsToForbidMap — превращает список пар в map «форма → [запрещённые формы]».
// Экспортируется для использования в handlers.
func PairsToForbidMap(pairs []CompatibilityPairInput) map[string][]string {
	result := map[string][]string{}
	for _, p := range pairs {
		result[p.TypeA] = append(result[p.TypeA], p.TypeB)
	}
	return result
}

// ==================== API ДЛЯ ХЕНДЛЕРОВ ====================

// GetSurfaceForbidMap — возвращает текущую матрицу поверхности в виде
// «форма → список запрещённых» (для ответа админке).
func GetSurfaceForbidMap() map[string][]string {
	compatMu.RLock()
	defer compatMu.RUnlock()
	if compatMatrix == nil {
		return map[string][]string{}
	}
	return copyForbidMap(compatMatrix.Surface)
}

// GetSubterrainForbidMap — то же для недр.
func GetSubterrainForbidMap() map[string][]string {
	compatMu.RLock()
	defer compatMu.RUnlock()
	if compatMatrix == nil {
		return map[string][]string{}
	}
	return copyForbidMap(compatMatrix.Subterrain)
}

// copyForbidMap — глубокая копия карты запретов.
func copyForbidMap(src map[string][]string) map[string][]string {
	if src == nil {
		return map[string][]string{}
	}
	dst := make(map[string][]string, len(src))
	for k, list := range src {
		cp := make([]string, len(list))
		copy(cp, list)
		dst[k] = cp
	}
	return dst
}