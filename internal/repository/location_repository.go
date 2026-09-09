package repository

import (
	"database/sql"
	"encoding/json"
	"time"

	"zorion/internal/models"
)

type LocationRepository struct {
	db *sql.DB
}

func NewLocationRepository(db *sql.DB) *LocationRepository {
	return &LocationRepository{db: db}
}

// Create создаёт новую локацию
func (r *LocationRepository) Create(location *models.Location) error {
	query := `
		INSERT INTO locations (id, world_id, name, is_inhabited, state, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	now := time.Now()
	stateJSON, err := json.Marshal(location.State)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(query,
		location.ID,
		location.WorldID,
		location.Name,
		location.IsInhabited,
		stateJSON,
		now,
		now,
	)
	if err != nil {
		return err
	}
	location.CreatedAt = now
	location.UpdatedAt = now
	return nil
}

// GetByID возвращает локацию по ID
func (r *LocationRepository) GetByID(id string) (*models.Location, error) {
	query := `SELECT id, world_id, name, is_inhabited, state, created_at, updated_at FROM locations WHERE id = $1`
	row := r.db.QueryRow(query, id)

	var loc models.Location
	var stateJSON []byte
	err := row.Scan(
		&loc.ID,
		&loc.WorldID,
		&loc.Name,
		&loc.IsInhabited,
		&stateJSON,
		&loc.CreatedAt,
		&loc.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(stateJSON, &loc.State); err != nil {
		return nil, err
	}
	return &loc, nil
}

// GetByWorld возвращает все локации для мира
func (r *LocationRepository) GetByWorld(worldID string) ([]*models.Location, error) {
	query := `SELECT id, world_id, name, is_inhabited, state, created_at, updated_at FROM locations WHERE world_id = $1 ORDER BY name`
	rows, err := r.db.Query(query, worldID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locations []*models.Location
	for rows.Next() {
		var loc models.Location
		var stateJSON []byte
		err := rows.Scan(
			&loc.ID,
			&loc.WorldID,
			&loc.Name,
			&loc.IsInhabited,
			&stateJSON,
			&loc.CreatedAt,
			&loc.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(stateJSON, &loc.State); err != nil {
			return nil, err
		}
		locations = append(locations, &loc)
	}
	return locations, nil
}