// internal/generator/planet/planet_data_description.go
package planet

import (
	"fmt"
	"math/rand"
)

// generateDescription — текстовое описание планеты на основе её характеристик.
func generateDescription(
	rng *rand.Rand,
	dominant string,
	habitable, life bool,
) string {
	if life && habitable {
		adjectives := []string{
			"цветущий", "развитый", "мирный",
			"технологичный", "экологичный",
		}
		return fmt.Sprintf(
			"%s мир с богатой биосферой",
			adjectives[rng.Intn(len(adjectives))],
		)
	}
	if habitable {
		return "Потенциально пригодная для терраформирования планета."
	}
	if life {
		return "Планета с признаками микробной жизни."
	}
	return dominantDescription(dominant)
}

// dominantDescription — описание по доминирующей форме поверхности.
func dominantDescription(dominant string) string {
	switch dominant {
	case SurfaceOceans:
		return "Безжизненный океанический мир."
	case SurfaceLavaFields:
		return "Раскалённый мир с морями лавы."
	case SurfaceVolcanicFields:
		return "Вулканический мир, покрытый застывшей лавой."
	case SurfaceGlaciers:
		return "Ледяной мир, скованный вечным холодом."
	case SurfaceFrozenGases:
		return "Мёртвый мир из замёрзших газов и льда."
	case SurfaceSands:
		return "Пустынный мир с бескрайними песками."
	case SurfaceGlassFields:
		return "Оплавленная поверхность, превращённая в стекло."
	case SurfaceMetalFields:
		return "Металлический мир с обнажённой корой."
	case SurfaceCraters:
		return "Изрытый кратерами безжизненный мир."
	case SurfaceRocks:
		return "Скалистый мир без признаков жизни."
	default:
		return "Безжизненный и суровый мир."
	}
}

// clamp — ограничивает значение диапазоном [min, max].
// Общая утилита для генераторов.
func clamp(val, min, max float64) float64 {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}