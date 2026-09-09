package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"zorion/internal/models"
	"zorion/internal/repository"
)

type TestHandlers struct {
	worldRepo      *repository.WorldRepository
	locationRepo   *repository.LocationRepository
	assignmentRepo *repository.AssignmentRepository
}

func NewTestHandlers(
	worldRepo *repository.WorldRepository,
	locationRepo *repository.LocationRepository,
	assignmentRepo *repository.AssignmentRepository,
) *TestHandlers {
	return &TestHandlers{
		worldRepo:      worldRepo,
		locationRepo:   locationRepo,
		assignmentRepo: assignmentRepo,
	}
}

// CreateTestData создаёт тестовый мир, локацию и задание
func (h *TestHandlers) CreateTestData(w http.ResponseWriter, r *http.Request) {
	// 1. Создаём мир
	worldID := uuid.New().String()
	world := &models.World{
		ID:     worldID,
		Name:   "Тестовый мир",
		CoordX: 0,
		CoordY: 0,
	}
	if err := h.worldRepo.Create(world); err != nil {
		http.Error(w, "Ошибка создания мира: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 2. Создаём локацию внутри мира
	locationID := uuid.New().String()
	location := &models.Location{
		ID:          locationID,
		WorldID:     worldID,
		Name:        "Тестовая локация",
		IsInhabited: true,
		State: map[string]interface{}{
			"population": 1000,
			"gdp":        5000.0,
		},
	}
	if err := h.locationRepo.Create(location); err != nil {
		http.Error(w, "Ошибка создания локации: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 3. Создаём задание
	assignmentID := uuid.New().String()
	assignment := &models.Assignment{
		ID:          assignmentID,
		WorldID:     worldID,
		AuthorType:  "location",
		AuthorID:    locationID,
		Title:       "Тестовое задание",
		Description: "Это задание создано для проверки",
		ExpiresAt:   time.Now().Add(24 * time.Hour),
		Status:      "open",
		Effects: map[string]interface{}{
			"population": 100,
		},
	}
	if err := h.assignmentRepo.Create(assignment); err != nil {
		http.Error(w, "Ошибка создания задания: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Ответ
	response := map[string]interface{}{
		"world_id":      worldID,
		"location_id":   locationID,
		"assignment_id": assignmentID,
		"message":       "Тестовые данные созданы успешно",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}