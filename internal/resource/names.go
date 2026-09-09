package resource

import (
	"math/rand"
)

// GenerateResourceName создаёт абстрактное название ресурса без пересечений с реальностью
func GenerateResourceName(category string, rng *rand.Rand, existingNames map[string]bool) string {
	roots := []string{
		"Атр", "Брол", "Векс", "Грон", "Дрей", "Зорн", "Квел", "Мрайн", "Нокс", "Прокс",
		"Свар", "Трок", "Фрин", "Хард", "Цвок", "Шрак", "Эрд", "Альт", "Брим", "Валд",
		"Гарт", "Дрейк", "Зельт", "Корн", "Линт", "Морн", "Норн", "Орн", "Парн", "Рорк",
		"Сорн", "Торн", "Урн", "Форн", "Хорн", "Цорн", "Шорн", "Юрн", "Ярн", "Аэл",
		"Берн", "Вирн", "Гарн", "Дерн", "Зерн", "Керн", "Лерн", "Мерн", "Нерн", "Перн",
	}

	suffixes := map[string][]string{
		"mineral": {"ит", "ий", "ин", "ан", "ид", "ат", "он", "ор", "ум", "ил"},
		"organic": {"ин", "ан", "ил", "он", "ор", "ит", "ум", "оз", "а", "я"},
		"energy":  {"ин", "ан", "ил", "он", "ор", "ит", "ум", "ол", "ак", "ек"},
		"rare":    {"ит", "ий", "ин", "ан", "ид", "ат", "он", "ор", "ум", "ил"},
	}

	sufList := suffixes[category]
	if len(sufList) == 0 {
		sufList = suffixes["mineral"]
	}

	maxAttempts := 100
	for attempt := 0; attempt < maxAttempts; attempt++ {
		root := roots[rng.Intn(len(roots))]
		suf := sufList[rng.Intn(len(sufList))]
		name := root + suf
		if !existingNames[name] {
			existingNames[name] = true
			return name
		}
	}
	return "Ресурс-" + string(rune(65+rng.Intn(26))) + string(rune(65+rng.Intn(26)))
}