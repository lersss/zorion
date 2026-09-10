// internal/generator/planet/descriptions_manager.go
package planet

import (
	"log"
	"strings"
)

// ==================== ГЛОБАЛЬНЫЙ МЕНЕДЖЕР ====================
//
// Инициализируется один раз при старте сервера (LoadDescriptions),
// после чего только читается из горутин. RWMutex внутри менеджера
// защищает карты от конкурентного чтения.

var globalDescriptions *descriptionsManager

// LoadDescriptionsGlobal — инициализация глобального менеджера.
// Вызывается один раз из main.go. При ошибке — log.Fatal на стороне
// вызывающего (сервер не должен стартовать без описаний).
func LoadDescriptionsGlobal(baseDir string) error {
	m, err := LoadDescriptions(baseDir)
	if err != nil {
		return err
	}
	globalDescriptions = m
	return nil
}

// ==================== ТОЧКА ВХОДА ====================

// GenerateDescription — собирает описание планеты: зачин + концовка.
// Если для типа нет подходящих записей — возвращает fallback
// (детерминированный по planetID), чтобы поле description никогда
// не было пустым.
func GenerateDescription(ctx DescriptionContext) string {
	if globalDescriptions == nil {
		return fallbackDescription(ctx.PlanetID)
	}
	return globalDescriptions.generate(ctx)
}

// generate — внутренняя реализация.
func (m *descriptionsManager) generate(ctx DescriptionContext) string {
	if ctx.PlanetID == "" || ctx.Type == "" {
		return fallbackDescription(ctx.PlanetID)
	}

	tags := computeTags(ctx)

	m.mu.RLock()
	defer m.mu.RUnlock()

	opening, okOpen := m.pickOpening(ctx.PlanetID, ctx.Type, tags)
	closing, okClose := m.pickClosing(ctx.PlanetID, ctx.Type, tags)

	// Если не нашлось ни одной подходящей записи — fallback.
	if !okOpen && !okClose {
		log.Printf(
			"descriptions: fallback для планеты %s (тип %s, тегов=%d)",
			ctx.PlanetID, ctx.Type, len(tags),
		)
		return fallbackDescription(ctx.PlanetID)
	}

	return joinDescription(opening, closing)
}

// ==================== СКЛЕЙКА ====================

// joinDescription — склеивает зачин и концовку.
// Если одна из частей пустая — возвращает только вторую.
//
// Когда появятся средние блоки (climate, surface, ...), логика
// расширится здесь: pieces := []string{opening, climate, surface,
// interior, atmosphere, biosphere, closing}; strings.Join(pieces,
// descriptionSeparator).
func joinDescription(opening, closing string) string {
	switch {
	case opening == "" && closing == "":
		return ""
	case opening == "":
		return closing
	case closing == "":
		return opening
	default:
		return opening + descriptionSeparator + closing
	}
}

// ==================== FALLBACK ====================
//
// Используется, когда для планеты не нашлось ни одного подходящего
// зачина и ни одной концовки. Варианты подобраны в нейтральном тоне —
// подходят любому типу планеты, не противоречат ни одному тегу.
//
// Выбор — детерминированный по planetID, чтобы одна и та же планета
// всегда получала один и тот же fallback.

var fallbackVariants = []string{
	"Ничем не примечательная планета. Ни ресурсов, ни жизни, ни истории — только камень и время.",

	"Обычный мир без ярких особенностей. Таких в галактике много, и большинство экспедиций проходит мимо, не задерживаясь.",

	"Планета не запомнилась ничем. Ровная, спокойная, предсказуемая — из тех, что служат фоном для более интересных мест.",

	"Здесь нечего искать. Мир, который существует сам по себе и не предлагает ничего, кроме самого факта своего существования.",

	"Непримечательный мир. Ни воды, ни биосферы, ни следов разума — только порода, пыль и медленное время.",

	"Планета без истории. Она была здесь до нас и будет после — тихая, серая, никому не нужная.",

	"Один из тех миров, что редко попадают на карты. Не потому что опасны или недоступны, а потому что нечем зацепиться.",
}

// fallbackDescription — выбирает вариант по хэшу от planetID.
// Пустой planetID → первый вариант (детерминированно).
func fallbackDescription(planetID string) string {
	if len(fallbackVariants) == 0 {
		return ""
	}
	if planetID == "" {
		return fallbackVariants[0]
	}
	idx := hashIndex(planetID+":fallback", len(fallbackVariants))
	return fallbackVariants[idx]
}

// ==================== ДИАГНОСТИКА ====================

// DescriptionsStats — публичная сводка для /admin или healthcheck.
// Возвращает копию, чтобы вызывающий не мог испортить внутреннее
// состояние менеджера.
func DescriptionsStats() map[string]descriptionsStats {
	if globalDescriptions == nil {
		return nil
	}
	globalDescriptions.mu.RLock()
	defer globalDescriptions.mu.RUnlock()

	out := make(map[string]descriptionsStats, len(globalDescriptions.stats))
	for k, v := range globalDescriptions.stats {
		out[k] = v
	}
	return out
}

// ==================== УТИЛИТЫ ====================

// trimDescription — убирает лишние пробелы по краям и склеивает
// множественные переносы. На случай, если в JSON кто-то оставил
// лишние \n в начале/конце text.
func trimDescription(s string) string {
	s = strings.TrimSpace(s)
	for strings.Contains(s, "\n\n\n") {
		s = strings.ReplaceAll(s, "\n\n\n", "\n\n")
	}
	return s
}