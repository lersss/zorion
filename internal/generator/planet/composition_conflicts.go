// internal/generator/planet/composition_conflicts.go
package planet

import "math/rand"

// resolveConflicts — итеративно удаляет несовместимые пары из композиции.
//
// При конфликте двух форм:
//   - удаляется форма с меньшим весом;
//   - при равных весах — удаляется случайная.
//
// Если после удаления композиция опустела — возвращается копия base.
func resolveConflicts(
	category string,
	c map[string]float64,
	base map[string]float64,
	rng *rand.Rand,
) map[string]float64 {
	if len(c) == 0 {
		return c
	}

	maxIterations := len(c) * len(c) // защита от зацикливания

	for iter := 0; iter < maxIterations; iter++ {
		forms := keysWithPositive(c)
		if len(forms) <= 1 {
			break
		}

		conflictA, conflictB := findConflict(category, forms)
		if conflictA == "" {
			// Конфликтов больше нет
			break
		}

		removeWeaker(c, conflictA, conflictB, rng)
	}

	// Если всё удалили — восстановим из base
	if len(c) == 0 {
		return copyWeights(base)
	}
	return c
}

// findConflict — ищет первую несовместимую пару среди форм с ненулевым весом.
// Возвращает ("", "") если конфликтов нет.
func findConflict(category string, forms []string) (string, string) {
	for i := 0; i < len(forms); i++ {
		for j := i + 1; j < len(forms); j++ {
			if !IsCompatible(category, forms[i], forms[j]) {
				return forms[i], forms[j]
			}
		}
	}
	return "", ""
}

// removeWeaker — удаляет форму с меньшим весом. При равных весах — случайную.
func removeWeaker(c map[string]float64, a, b string, rng *rand.Rand) {
	wa, wb := c[a], c[b]
	switch {
	case wa < wb:
		delete(c, a)
	case wb < wa:
		delete(c, b)
	default:
		if rng.Intn(2) == 0 {
			delete(c, a)
		} else {
			delete(c, b)
		}
	}
}