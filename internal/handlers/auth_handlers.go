package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"zorion/internal/auth"
	"zorion/internal/models"
	"zorion/internal/repository"
)

type AuthHandlers struct {
	userRepo  *repository.UserRepository
	worldRepo *repository.WorldRepository
}

func NewAuthHandlers(userRepo *repository.UserRepository, worldRepo *repository.WorldRepository) *AuthHandlers {
	return &AuthHandlers{
		userRepo:  userRepo,
		worldRepo: worldRepo,
	}
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email,omitempty"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  struct {
		ID       string `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email,omitempty"`
	} `json:"user"`
}

// Register обрабатывает регистрацию нового пользователя
func (h *AuthHandlers) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Username == "" || req.Password == "" {
		http.Error(w, "Username and password required", http.StatusBadRequest)
		return
	}

	existing, _ := h.userRepo.GetByUsername(req.Username)
	if existing != nil {
		http.Error(w, "Username already taken", http.StatusConflict)
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	user := &models.User{
		ID:           uuid.New().String(),
		Username:     req.Username,
		PasswordHash: string(hashed),
	}
	if req.Email != "" {
		user.Email = &req.Email
	}
	// Устанавливаем текущий мир в nil — позже будет установлен при первом полёте
	if err := h.userRepo.Create(user); err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	token, err := auth.GenerateToken(user.ID)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	resp := AuthResponse{
		Token: token,
	}
	resp.User.ID = user.ID
	resp.User.Username = user.Username
	if user.Email != nil {
		resp.User.Email = *user.Email
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// Login обрабатывает вход пользователя
func (h *AuthHandlers) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Username == "" || req.Password == "" {
		http.Error(w, "Username and password required", http.StatusBadRequest)
		return
	}

	user, err := h.userRepo.GetByUsername(req.Username)
	if err != nil || user == nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := auth.GenerateToken(user.ID)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	resp := AuthResponse{
		Token: token,
	}
	resp.User.ID = user.ID
	resp.User.Username = user.Username
	if user.Email != nil {
		resp.User.Email = *user.Email
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// GetMe возвращает данные текущего пользователя с названием текущего мира
func (h *AuthHandlers) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(auth.UserIDKey).(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.userRepo.GetByID(userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	var currentWorldName string
	if user.CurrentWorldID != nil {
		world, err := h.worldRepo.GetByID(*user.CurrentWorldID)
		if err == nil && world != nil {
			currentWorldName = world.Name
		}
	}

	response := map[string]interface{}{
		"id":                 user.ID,
		"username":           user.Username,
		"email":              user.Email,
		"current_world_id":   user.CurrentWorldID,
		"current_world_name": currentWorldName,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}