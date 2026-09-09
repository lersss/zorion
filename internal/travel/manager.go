package travel

import (
	"log"
	"sync"
	"time"
)

// TravelInfo содержит информацию о текущем полёте
type TravelInfo struct {
	UserID     string
	FromWorld  string
	ToWorld    string
	StartTime  time.Time
	Duration   time.Duration
	CancelChan chan struct{} // канал для отмены (пока не используется)
}

// Manager управляет активными полётами
type Manager struct {
	mu      sync.RWMutex
	flights map[string]*TravelInfo // key: userID
}

func NewManager() *Manager {
	return &Manager{
		flights: make(map[string]*TravelInfo),
	}
}

// StartFlight запускает полёт для пользователя
func (m *Manager) StartFlight(userID, fromWorldID, toWorldID string, duration time.Duration, onArrival func(userID, worldID string)) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Если уже есть активный полёт — заменяем
	if existing, ok := m.flights[userID]; ok {
		close(existing.CancelChan)
		delete(m.flights, userID)
	}

	flight := &TravelInfo{
		UserID:     userID,
		FromWorld:  fromWorldID,
		ToWorld:    toWorldID,
		StartTime:  time.Now(),
		Duration:   duration,
		CancelChan: make(chan struct{}),
	}
	m.flights[userID] = flight

	// Запускаем горутину с таймером
	go func() {
		select {
		case <-time.After(duration):
			// Полёт завершён
			m.mu.Lock()
			delete(m.flights, userID)
			m.mu.Unlock()
			log.Printf("Travel completed: user %s -> world %s", userID, toWorldID)
			// Вызываем колбэк
			if onArrival != nil {
				onArrival(userID, toWorldID)
			}
		case <-flight.CancelChan:
			// Полёт отменён (пока не используется)
			m.mu.Lock()
			delete(m.flights, userID)
			m.mu.Unlock()
			log.Printf("Travel cancelled: user %s", userID)
		}
	}()
}

// GetFlight возвращает информацию о текущем полёте пользователя
func (m *Manager) GetFlight(userID string) *TravelInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.flights[userID]
}

// IsInFlight проверяет, находится ли пользователь в полёте
func (m *Manager) IsInFlight(userID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.flights[userID]
	return ok
}