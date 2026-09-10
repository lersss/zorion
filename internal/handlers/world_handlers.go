package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"zorion/internal/repository"
)

type WorldHandlers struct {
	worldRepo      *repository.WorldRepository
	locationRepo   *repository.LocationRepository
	assignmentRepo *repository.AssignmentRepository
}

func NewWorldHandlers(
	worldRepo *repository.WorldRepository,
	locationRepo *repository.LocationRepository,
	assignmentRepo *repository.AssignmentRepository,
) *WorldHandlers {
	return &WorldHandlers{
		worldRepo:      worldRepo,
		locationRepo:   locationRepo,
		assignmentRepo: assignmentRepo,
	}
}

// GetAllWorlds возвращает список всех миров
func (h *WorldHandlers) GetAllWorlds(w http.ResponseWriter, r *http.Request) {
	worlds, err := h.worldRepo.GetAll()
	if err != nil {
		log.Printf("GetAllWorlds error: %v", err)
		http.Error(w, "Ошибка получения миров", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(worlds)
}

// GetWorld возвращает мир по ID с его локациями и заданиями.
// Ключевая особенность: если locations или assignments падают —
// это не повод возвращать 500. Мир важнее. Логируем и отдаём что есть.
func (h *WorldHandlers) GetWorld(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/worlds/")
	if path == "" || path == r.URL.Path {
		http.Error(w, "ID не указан", http.StatusBadRequest)
		return
	}
	id := strings.Split(path, "/")[0]
	if id == "" {
		http.Error(w, "ID не указан", http.StatusBadRequest)
		return
	}

	world, err := h.worldRepo.GetByID(id)
	if err != nil {
		log.Printf("GetWorld: worldRepo.GetByID(%s) error: %v", id, err)
		http.Error(w, "Ошибка получения мира", http.StatusInternalServerError)
		return
	}
	if world == nil {
		http.Error(w, "Мир не найден", http.StatusNotFound)
		return
	}

	// locations — вспомогательные данные. Их падение не блокирует ответ.
	locations, err := h.locationRepo.GetByWorld(id)
	if err != nil {
		log.Printf("GetWorld: locationRepo.GetByWorld(%s) error (пропускаем): %v", id, err)
		locations = nil
	}

	// assignments — тоже вспомогательные.
	assignments, err := h.assignmentRepo.GetByWorld(id)
	if err != nil {
		log.Printf("GetWorld: assignmentRepo.GetByWorld(%s) error (пропускаем): %v", id, err)
		assignments = nil
	}

	response := map[string]interface{}{
		"world":       world,
		"locations":   locations,
		"assignments": assignments,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}