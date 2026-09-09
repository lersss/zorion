package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"zorion/internal/auth"
	"zorion/internal/models"
	"zorion/internal/repository"
)

type ContractHandlers struct {
	assignmentRepo *repository.AssignmentRepository
	userRepo       *repository.UserRepository
}

func NewContractHandlers(assignmentRepo *repository.AssignmentRepository, userRepo *repository.UserRepository) *ContractHandlers {
	return &ContractHandlers{
		assignmentRepo: assignmentRepo,
		userRepo:       userRepo,
	}
}

// GetContracts возвращает контракты текущего мира с фильтрацией
func (h *ContractHandlers) GetContracts(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(auth.UserIDKey).(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.userRepo.GetByID(userID)
	if err != nil || user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Если у пользователя нет текущего мира — возвращаем пустой список
	if user.CurrentWorldID == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}
	worldID := *user.CurrentWorldID

	// Параметры фильтрации из query string
	query := r.URL.Query()
	typeFilter := query.Get("type")
	authorFilter := query.Get("author")
	search := query.Get("search")
	minReward, _ := strconv.Atoi(query.Get("min_reward"))
	maxReward, _ := strconv.Atoi(query.Get("max_reward"))

	contracts, err := h.assignmentRepo.GetByWorldWithFilters(worldID, typeFilter, authorFilter, search, minReward, maxReward)
	if err != nil {
		http.Error(w, "Failed to fetch contracts", http.StatusInternalServerError)
		return
	}

	// Если контрактов нет — возвращаем пустой массив
	if contracts == nil {
		contracts = []*models.Assignment{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(contracts)
}

// TakeContract берёт контракт (изменяет статус на 'taken')
func (h *ContractHandlers) TakeContract(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(auth.UserIDKey).(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		ContractID string `json:"contract_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	if req.ContractID == "" {
		http.Error(w, "contract_id required", http.StatusBadRequest)
		return
	}

	contract, err := h.assignmentRepo.GetByID(req.ContractID)
	if err != nil || contract == nil {
		http.Error(w, "Contract not found", http.StatusNotFound)
		return
	}
	if contract.Status != "open" {
		http.Error(w, "Contract is not available", http.StatusBadRequest)
		return
	}

	if err := h.assignmentRepo.UpdateStatus(req.ContractID, "taken"); err != nil {
		http.Error(w, "Failed to take contract", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"taken"}`))
}

// CompleteTestContract — для тестовых контрактов (мгновенное выполнение)
func (h *ContractHandlers) CompleteTestContract(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(auth.UserIDKey).(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		ContractID string `json:"contract_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	if req.ContractID == "" {
		http.Error(w, "contract_id required", http.StatusBadRequest)
		return
	}

	contract, err := h.assignmentRepo.GetByID(req.ContractID)
	if err != nil || contract == nil {
		http.Error(w, "Contract not found", http.StatusNotFound)
		return
	}
	if contract.Status != "taken" {
		http.Error(w, "Contract is not in taken state", http.StatusBadRequest)
		return
	}
	if contract.Type != "test" {
		http.Error(w, "Only test contracts can be completed instantly", http.StatusBadRequest)
		return
	}

	if err := h.assignmentRepo.UpdateStatus(req.ContractID, "completed"); err != nil {
		http.Error(w, "Failed to complete contract", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"completed"}`))
}