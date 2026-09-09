package models

import "time"

type Planet struct {
	ID           string    `json:"id"`
	WorldID      string    `json:"world_id"`
	Name         string    `json:"name"`
	OrbitIndex   int       `json:"orbit_index"`
	Type         string    `json:"type"`
	Size         float64   `json:"size"`
	Mass         float64   `json:"mass"`
	Atmosphere   string    `json:"atmosphere"`
	Temperature  float64   `json:"temperature"`
	WaterPercent float64   `json:"water_percent"`
	Habitable    bool      `json:"habitable"`
	Life         bool      `json:"life"`
	Population   int64     `json:"population"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}