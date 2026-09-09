package handlers

import (
	"encoding/json"
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
		http.Error(w, "Ошибка получения миров: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(worlds)
}

// GetWorld возвращает мир по ID с его локациями и заданиями
func (h *WorldHandlers) GetWorld(w http.ResponseWriter, r *http.Request) {
	// Извлекаем ID из URL (например, /worlds/123)
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
		http.Error(w, "Ошибка получения мира: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if world == nil {
		http.Error(w, "Мир не найден", http.StatusNotFound)
		return
	}

	locations, err := h.locationRepo.GetByWorld(id)
	if err != nil {
		http.Error(w, "Ошибка получения локаций: "+err.Error(), http.StatusInternalServerError)
		return
	}

	assignments, err := h.assignmentRepo.GetByWorld(id)
	if err != nil {
		http.Error(w, "Ошибка получения заданий: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"world":       world,
		"locations":   locations,
		"assignments": assignments,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}