package batch

import (
	"time"
	"github.com/google/uuid"
)

// Effect — это заявка на изменение мира от одного Assignment
type Effect struct {
	ID             uuid.UUID
	LocationID     uuid.UUID
	ActorID        uuid.UUID   // кто инициировал
	AssignmentID   uuid.UUID   // ссылка на задание
	ResourceType   string      // "wood", "gold", "food"
	Delta          int64       // положительное = добавление, отрицательное = изъятие
	Priority       int         // чем больше, тем важнее (для конфликтов)
	CreatedAt      time.Time
	Metadata       map[string]interface{} // для гибкости (координаты, качество и т.д.)
}

// ResolvedEffect — финальное изменение после разрешения конфликтов
type ResolvedEffect struct {
	ResourceType   string
	Delta          int64
	// SourceIDs — список ID эффектов, которые вошли в этот резолв (для аудита)
	SourceIDs      []uuid.UUID
}

// CommitBundle — атомарный пакет изменений для одной локации
type CommitBundle struct {
	LocationID     uuid.UUID
	TickNumber     int64
	Effects        []ResolvedEffect
	Version        int64  // для optimistic locking в БД
}