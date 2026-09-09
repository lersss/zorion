package repository

import (
	"database/sql"

	"zorion/internal/models"
)

type ResourceRepository struct {
	db *sql.DB
}

func NewResourceRepository(db *sql.DB) *ResourceRepository {
	return &ResourceRepository{db: db}
}

func (r *ResourceRepository) Create(resource *models.PlanetResource) error {
	query := `
		INSERT INTO planet_resources (
			id, planet_id, name, category,
			hardness, elasticity, conductivity, heat_resistance,
			chemical_activity, density, biocompatibility, energy_density, volatility,
			quantity, is_known, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, NOW(), NOW())
	`
	_, err := r.db.Exec(query,
		resource.ID, resource.PlanetID, resource.Name, resource.Category,
		resource.Hardness, resource.Elasticity, resource.Conductivity, resource.HeatResistance,
		resource.ChemicalActivity, resource.Density, resource.Biocompatibility, resource.EnergyDensity, resource.Volatility,
		resource.Quantity, resource.IsKnown,
	)
	return err
}

func (r *ResourceRepository) GetByPlanet(planetID string) ([]*models.PlanetResource, error) {
	query := `SELECT id, planet_id, name, category, hardness, elasticity, conductivity, heat_resistance,
	          chemical_activity, density, biocompatibility, energy_density, volatility, quantity, is_known, created_at, updated_at
	          FROM planet_resources WHERE planet_id = $1`
	rows, err := r.db.Query(query, planetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resources []*models.PlanetResource
	for rows.Next() {
		var r models.PlanetResource
		err := rows.Scan(
			&r.ID, &r.PlanetID, &r.Name, &r.Category,
			&r.Hardness, &r.Elasticity, &r.Conductivity, &r.HeatResistance,
			&r.ChemicalActivity, &r.Density, &r.Biocompatibility, &r.EnergyDensity, &r.Volatility,
			&r.Quantity, &r.IsKnown, &r.CreatedAt, &r.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		resources = append(resources, &r)
	}
	return resources, nil
}