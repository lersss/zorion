package core

import (
	"context"
	"sync"
	"time"
	"github.com/google/uuid"
	"your-project/pkg/simulation/batch"
)

// LocationWithBatch — обертка над вашей существующей Location
type LocationWithBatch struct {
	*Location // ваша существующая структура
	
	mu           sync.RWMutex
	pendingEffects []batch.Effect
	stateSnapshot batch.ResourceState // кэш состояния на начало тика
	
	resolverConfig batch.ResolverConfig
	eventStore     EventStore // ваш интерфейс для сохранения событий
}

// NewLocationWithBatch — конструктор
func NewLocationWithBatch(
	loc *Location,
	eventStore EventStore,
	config batch.ResolverConfig,
) *LocationWithBatch {
	return &LocationWithBatch{
		Location:       loc,
		pendingEffects: make([]batch.Effect, 0, 100),
		stateSnapshot:  make(batch.ResourceState),
		resolverConfig: config,
		eventStore:     eventStore,
	}
}

// AddEffect — вызывается из Assignment'а (не блокирует, просто складывает в буфер)
func (l *LocationWithBatch) AddEffect(ctx context.Context, effect batch.Effect) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	// Простая валидация
	if effect.LocationID != l.ID {
		return ErrInvalidLocation
	}
	
	l.pendingEffects = append(l.pendingEffects, effect)
	return nil
}

// OnTickStart — вызывается в начале каждого тика, фиксируем снэпшот состояния
func (l *LocationWithBatch) OnTickStart(ctx context.Context) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	// Снимаем снепшот текущего состояния ресурсов
	l.stateSnapshot = l.getResourceState() // реализуйте этот метод для вашей Location
	l.pendingEffects = l.pendingEffects[:0] // очищаем буфер, но сохраняем capacity
	return nil
}

// OnTickEnd — вызывается в конце тика, резолвим и применяем
func (l *LocationWithBatch) OnTickEnd(ctx context.Context, tickNumber int64) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	if len(l.pendingEffects) == 0 {
		return nil
	}
	
	// 1. Резолвим конфликты
	bundle, err := batch.Resolve(
		l.ID,
		tickNumber,
		l.stateSnapshot,
		l.pendingEffects,
		l.resolverConfig,
	)
	if err != nil {
		return err
	}
	
	// 2. Применяем к состоянию локации
	if err := l.applyBundle(ctx, bundle); err != nil {
		return err
	}
	
	// 3. Сохраняем события в Event Store (одной пачкой)
	if err := l.eventStore.SaveEvents(ctx, bundle); err != nil {
		// Здесь нужна компенсирующая транзакция, но для простоты откатываем состояние
		return err
	}
	
	// 4. Очищаем буфер
	l.pendingEffects = l.pendingEffects[:0]
	
	return nil
}

// applyBundle — применяет ResolvedEffect к состоянию локации
func (l *LocationWithBatch) applyBundle(ctx context.Context, bundle *batch.CommitBundle) error {
	for _, effect := range bundle.Effects {
		// Ваша логика изменения ресурсов локации
		// Например:
		// l.Resources[effect.ResourceType] += effect.Delta
	}
	return nil
}

// getResourceState — снимает снепшот состояния ресурсов
func (l *LocationWithBatch) getResourceState() batch.ResourceState {
	state := make(batch.ResourceState)
	// Заполните из вашей структуры Location
	// Например:
	// for k, v := range l.Resources {
	//     state[k] = v
	// }
	return state
}