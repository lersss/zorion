package models

import "time"

type World struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	CoordX         float64   `json:"coord_x"`
	CoordY         float64   `json:"coord_y"`
	SpectralClass  string    `json:"spectral_class"` // O, B, A, F, G, K, M, L, T, Y
	Temperature    int       `json:"temperature"`    // в Кельвинах
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}