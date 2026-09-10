// internal/generator/planet/descriptions_pick.go
package planet

// ==================== ФИЛЬТРАЦИЯ ПО ТЕГАМ ====================

// matchesAll — проверяет, что все теги из requires присутствуют в tags.
// Пустой requires — всегда true.
func matchesAll(requires []string, tags map[string]bool) bool {
	for _, req := range requires {
		if !tags[req] {
			return false
		}
	}
	return true
}

// filterOpenings — возвращает зачины, подходящие планете по тегам.
// Порядок сохраняется как в исходном срезе (файлы уже отсортированы
// лексикографически при загрузке).
func filterOpenings(all []Opening, tags map[string]bool) []Opening {
	if len(all) == 0 {
		return nil
	}
	result := make([]Opening, 0, len(all))
	for _, op := range all {
		if matchesAll(op.Requires, tags) {
			result = append(result, op)
		}
	}
	return result
}

// filterClosings — то же для концовок.
func filterClosings(all []Closing, tags map[string]bool) []Closing {
	if len(all) == 0 {
		return nil
	}
	result := make([]Closing, 0, len(all))
	for _, cl := range all {
		if matchesAll(cl.Requires, tags) {
			result = append(result, cl)
		}
	}
	return result
}

// ==================== ВЫБОР ПО ХЭШУ ====================

// pickOpening — выбирает один зачин из подходящих по тегам.
// Возвращает (текст, true) при успехе и ("", false), если ни один
// зачин не подошёл. В этом случае вызывающий код использует fallback.
func (m *descriptionsManager) pickOpening(
	planetID, typeName string,
	tags map[string]bool,
) (string, bool) {
	all, ok := m.openings[typeName]
	if !ok || len(all) == 0 {
		return "", false
	}

	candidates := filterOpenings(all, tags)
	if len(candidates) == 0 {
		return "", false
	}

	idx := hashIndex(planetID+openingHashSalt, len(candidates))
	return candidates[idx].Text, true
}

// pickClosing — то же для концовок.
func (m *descriptionsManager) pickClosing(
	planetID, typeName string,
	tags map[string]bool,
) (string, bool) {
	all, ok := m.closings[typeName]
	if !ok || len(all) == 0 {
		return "", false
	}

	candidates := filterClosings(all, tags)
	if len(candidates) == 0 {
		return "", false
	}

	idx := hashIndex(planetID+closingHashSalt, len(candidates))
	return candidates[idx].Text, true
}

// ==================== ХЭШ ====================

// hashIndex — стабильный индекс в диапазоне [0, size).
// Использует FNV-1a. Одна и та же строка → один и тот же индекс,
// в любом процессе, в любой момент времени.
func hashIndex(key string, size int) int {
	if size <= 0 {
		return 0
	}
	h := fnv1a(key)
	return int(h % uint64(size))
}

// fnv1a — 64-битный FNV-1a. Простой, быстрый, стабильный.
// Не криптостойкий — нам это и не нужно, только детерминизм.
func fnv1a(s string) uint64 {
	const (
		offset64 = 14695981039346656037
		prime64  = 1099511628211
	)
	h := uint64(offset64)
	for i := 0; i < len(s); i++ {
		h ^= uint64(s[i])
		h *= prime64
	}
	return h
}