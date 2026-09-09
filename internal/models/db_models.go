package models

import "time"

type Location struct {
	ID          string                 `json:"id"`
	WorldID     string                 `json:"world_id"`
	Name        string                 `json:"name"`
	IsInhabited bool                   `json:"is_inhabited"`
	State       map[string]interface{} `json:"state"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type ProductionUnit struct {
	ID                 string    `json:"id"`
	LocationID         string    `json:"location_id"`
	InputResource      string    `json:"input_resource"`
	OutputResource     string    `json:"output_resource"`
	CycleDurationTicks int       `json:"cycle_duration_ticks"`
	RemainingTicks     int       `json:"remaining_ticks"`
	IsActive           bool      `json:"is_active"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type Event struct {
	ID            string                 `json:"id"`
	AggregateID   string                 `json:"aggregate_id"`
	AggregateType string                 `json:"aggregate_type"`
	EventType     string                 `json:"event_type"`
	Data          map[string]interface{} `json:"data"`
	CreatedAt     time.Time              `json:"created_at"`
}