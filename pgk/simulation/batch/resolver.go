package batch

import (
	"errors"
	"fmt"
)

var (
	ErrNegativeResource = errors.New("resource cannot go below zero")
)

// ResourceState — текущее состояние ресурсов локации на момент начала тика
type ResourceState map[string]int64

// ResolverConfig — стратегии разрешения конфликтов
type ResolverConfig struct {
	AllowNegative      bool   // разрешать ли уход в минус (если false — пропорциональное урезание)
	StrictPriority     bool   // если true — высокий приоритет забирает всё, остальным ничего
	MaxDeltaPerTick    map[string]int64 // лимит изменения за тик (например, нельзя вырубить >1000 дерева за тик)
}

// DefaultResolverConfig — безопасные настройки для MMO
func DefaultResolverConfig() ResolverConfig {
	return ResolverConfig{
		AllowNegative:  false,
		StrictPriority: false,
		MaxDeltaPerTick: map[string]int64{
			"wood": 1000,
			"gold": 500,
			"food": 2000,
		},
	}
}

// Resolve — главная функция: на входе эффекты и состояние, на выходе CommitBundle
func Resolve(
	locationID uuid.UUID,
	tickNumber int64,
	currentState ResourceState,
	effects []Effect,
	config ResolverConfig,
) (*CommitBundle, error) {
	if len(effects) == 0 {
		return &CommitBundle{
			LocationID: locationID,
			TickNumber: tickNumber,
			Effects:    []ResolvedEffect{},
			Version:    0,
		}, nil
	}

	// 1. Группируем эффекты по типу ресурса
	groups := groupByResource(effects)

	// 2. Для каждой группы вычисляем финальный дельта
	resolved := make([]ResolvedEffect, 0, len(groups))
	newState := make(ResourceState)

	for resourceType, currentAmount := range currentState {
		newState[resourceType] = currentAmount
	}

	for resourceType, groupEffects := range groups {
		finalDelta, sourceIDs, err := calculateDelta(
			resourceType,
			currentState[resourceType],
			groupEffects,
			config,
		)
		if err != nil {
			return nil, fmt.Errorf("resolve resource %s: %w", resourceType, err)
		}

		// Проверяем лимит изменения за тик
		if maxDelta, ok := config.MaxDeltaPerTick[resourceType]; ok {
			if abs(finalDelta) > maxDelta {
				// Пропорционально урезаем до лимита
				finalDelta = clampDelta(finalDelta, maxDelta)
			}
		}

		// Проверяем, не уходим ли в минус
		newAmount := currentState[resourceType] + finalDelta
		if !config.AllowNegative && newAmount < 0 {
			// Урезаем дельту так, чтобы не уйти в минус
			finalDelta = -currentState[resourceType]
		}

		resolved = append(resolved, ResolvedEffect{
			ResourceType: resourceType,
			Delta:        finalDelta,
			SourceIDs:    sourceIDs,
		})

		// Обновляем состояние для последующих проверок (если ресурсы взаимозависимы)
		newState[resourceType] = currentState[resourceType] + finalDelta
	}

	return &CommitBundle{
		LocationID: locationID,
		TickNumber: tickNumber,
		Effects:    resolved,
		Version:    0, // будет проставлен при сохранении в БД
	}, nil
}

// groupByResource — группируем эффекты по типу ресурса
func groupByResource(effects []Effect) map[string][]Effect {
	groups := make(map[string][]Effect)
	for _, e := range effects {
		groups[e.ResourceType] = append(groups[e.ResourceType], e)
	}
	return groups
}

// calculateDelta — вычисляем итоговую дельту для одного ресурса
func calculateDelta(
	resourceType string,
	currentAmount int64,
	effects []Effect,
	config ResolverConfig,
) (int64, []uuid.UUID, error) {
	var totalDelta int64
	sourceIDs := make([]uuid.UUID, 0, len(effects))

	if config.StrictPriority {
		// Сортируем по приоритету (высший — первый)
		sorted := sortEffectsByPriority(effects)
		remaining := currentAmount

		for _, e := range sorted {
			if e.Delta < 0 { // только забор ресурсов имеет смысл для приоритетов
				available := remaining
				if available <= 0 {
					break
				}
				toTake := min(-e.Delta, available)
				totalDelta -= toTake
				sourceIDs = append(sourceIDs, e.ID)
				remaining -= toTake
				// Положительные дельты (добавление) применяются всем всегда
			} else {
				totalDelta += e.Delta
				sourceIDs = append(sourceIDs, e.ID)
			}
		}
	} else {
		// Демократичный режим: суммируем все дельты
		for _, e := range effects {
			totalDelta += e.Delta
			sourceIDs = append(sourceIDs, e.ID)
		}
	}

	return totalDelta, sourceIDs, nil
}

// вспомогательные функции
func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func clampDelta(delta, maxAbs int64) int64 {
	if delta > maxAbs {
		return maxAbs
	}
	if delta < -maxAbs {
		return -maxAbs
	}
	return delta
}

// sortEffectsByPriority — простейшая сортировка (можно заменить на sort.Slice)
func sortEffectsByPriority(effects []Effect) []Effect {
	// TODO: реализовать сортировку по убыванию Priority
	// Пока возвращаем как есть
	return effects
}