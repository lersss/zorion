package repository

import (
	"database/sql"

	"zorion/internal/models"
)

type EconomyRepository struct {
	db *sql.DB
}

func NewEconomyRepository(db *sql.DB) *EconomyRepository {
	return &EconomyRepository{db: db}
}

// Settlement
func (r *EconomyRepository) CreateSettlement(s *models.Settlement) error {
	query := `INSERT INTO settlements (id, planet_id, level, population, capacity, stability, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())`
	_, err := r.db.Exec(query, s.ID, s.PlanetID, s.Level, s.Population, s.Capacity, s.Stability)
	return err
}

// Factory
func (r *EconomyRepository) CreateFactory(f *models.Factory) error {
	query := `INSERT INTO factories (id, planet_id, name, type, input_resource, output_product, quality, status, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())`
	_, err := r.db.Exec(query, f.ID, f.PlanetID, f.Name, f.Type, f.InputResource, f.OutputProduct, f.Quality, f.Status)
	return err
}

// GoodsBatch
func (r *EconomyRepository) CreateGoodsBatch(b *models.GoodsBatch) error {
	query := `INSERT INTO goods_batches (id, planet_id, product_name, quantity, quality, producer_id, produced_at, created_at, expires_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), $8)`
	_, err := r.db.Exec(query, b.ID, b.PlanetID, b.ProductName, b.Quantity, b.Quality, b.ProducerID, b.ProducedAt, b.ExpiresAt)
	return err
}