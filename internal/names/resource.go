// internal/names/resource.go
package names

import "math/rand"

// GenerateResourceName — генерирует имя ресурса в двух вариантах:
// кириллица (Cyr) и латиница (Lat).
//
// Пример: {"Атрокс", "Atrox"}, {"Бролин", "Brolin"}.
//
// Параметры:
//   - category — категория ресурса ("mineral", "organic", "rare",
//     "fuel", "water", "gas").
//   - rng — источник случайности.
//   - usedNames — карта занятых имён (ключи — кириллические варианты).
//
// Если category неизвестна — используется набор "mineral".
func GenerateResourceName(
	category string,
	rng *rand.Rand,
	usedNames map[string]bool,
) LocalizedName {
	sufList, ok := resourceSuffixes[category]
	if !ok || len(sufList) == 0 {
		sufList = resourceSuffixes["mineral"]
	}

	for attempt := 0; attempt < 100; attempt++ {
		root := resourceRoots[rng.Intn(len(resourceRoots))]
		suf := sufList[rng.Intn(len(sufList))]

		cyr := root.Cyr + suf.Cyr
		if usedNames[cyr] {
			continue
		}
		usedNames[cyr] = true

		return LocalizedName{
			Cyr: cyr,
			Lat: root.Lat + suf.Lat,
		}
	}

	// Fallback — если 100 попыток не хватило (например, пространство имён исчерпано)
	fallbackCyr := "Ресурс-" + randomSuffix(rng)
	fallbackLat := "Resource-" + randomSuffix(rng)
	usedNames[fallbackCyr] = true

	return LocalizedName{
		Cyr: fallbackCyr,
		Lat: fallbackLat,
	}
}