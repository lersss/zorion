package handlers

import (
	"database/sql"
	"zorion/internal/repository"
)

type AdminHandlers struct {
	worldRepo *repository.WorldRepository
	db        *sql.DB
}

func NewAdminHandlers(worldRepo *repository.WorldRepository, db *sql.DB) *AdminHandlers {
	return &AdminHandlers{
		worldRepo: worldRepo,
		db:        db,
	}
}