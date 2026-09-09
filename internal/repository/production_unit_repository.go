package repository

import (
	"database/sql"
	"time"

	"zorion/internal/models"
)

type ProductionUnitRepository struct {
	db *sql.DB
}

func NewProductionUnitRepository(db *sql.DB) *ProductionUnitRepository {
	return &ProductionUnitRepository{db: db}
}

// Create создаёт новую производственную единицу
func (r *ProductionUnitRepository) Create(unit *models.ProductionUnit) error {
	query := `
		INSERT INTO production_units (id, location_id, input_resource, output_resource, cycle_duration_ticks, remaining_ticks, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	now := time.Now()
	_, err := r.db.Exec(query,
		unit.ID,
		unit.LocationID,
		unit.InputResource,
		unit.OutputResource,
		unit.CycleDurationTicks,
		unit.RemainingTicks,
		unit.IsActive,
		now,
		now,
	)
	if err != nil {
		return err
	}
	unit.CreatedAt = now
	unit.UpdatedAt = now
	return nil
}

// GetByID возвращает единицу по ID
func (r *ProductionUnitRepository) GetByID(id string) (*models.ProductionUnit, error) {
	query := `SELECT id, location_id, input_resource, output_resource, cycle_duration_ticks, remaining_ticks, is_active, created_at, updated_at FROM production_units WHERE id = $1`
	row := r.db.QueryRow(query, id)

	var u models.ProductionUnit
	err := row.Scan(
		&u.ID,
		&u.LocationID,
		&u.InputResource,
		&u.OutputResource,
		&u.CycleDurationTicks,
		&u.RemainingTicks,
		&u.IsActive,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// GetByLocation возвращает все единицы для локации
func (r *ProductionUnitRepository) GetByLocation(locationID string) ([]*models.ProductionUnit, error) {
	query := `SELECT id, location_id, input_resource, output_resource, cycle_duration_ticks, remaining_ticks, is_active, created_at, updated_at FROM production_units WHERE location_id = $1 ORDER BY created_at`
	rows, err := r.db.Query(query, locationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var units []*models.ProductionUnit
	for rows.Next() {
		var u models.ProductionUnit
		err := rows.Scan(
			&u.ID,
			&u.LocationID,
			&u.InputResource,
			&u.OutputResource,
			&u.CycleDurationTicks,
			&u.RemainingTicks,
			&u.IsActive,
			&u.CreatedAt,
			&u.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		units = append(units, &u)
	}
	return units, nil
}

// UpdateRemainingTicks обновляет оставшиеся тики
func (r *ProductionUnitRepository) UpdateRemainingTicks(id string, remainingTicks int) error {
	query := `UPDATE production_units SET remaining_ticks = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(query, remainingTicks, id)
	return err
}