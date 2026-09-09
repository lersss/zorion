package models

import "time"

type Assignment struct {
	ID          string                 `json:"id"`
	WorldID     string                 `json:"world_id"`
	AuthorType  string                 `json:"author_type"`
	AuthorID    string                 `json:"author_id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"`          // tech, agri, military, trade, mixed, test
	Reward      int                    `json:"reward"`        // награда в кредитах
	ExpiresAt   time.Time              `json:"expires_at"`
	Status      string                 `json:"status"`        // open, taken, completed, expired
	Effects     map[string]interface{} `json:"effects"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}