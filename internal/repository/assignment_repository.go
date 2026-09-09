package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"zorion/internal/models"
)

type AssignmentRepository struct {
	db *sql.DB
}

func NewAssignmentRepository(db *sql.DB) *AssignmentRepository {
	return &AssignmentRepository{db: db}
}

// Create создаёт контракт
func (r *AssignmentRepository) Create(a *models.Assignment) error {
	query := `
		INSERT INTO assignments (id, world_id, author_type, author_id, title, description, type, reward, expires_at, status, effects, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	now := time.Now()
	effectsJSON, err := json.Marshal(a.Effects)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(query,
		a.ID, a.WorldID, a.AuthorType, a.AuthorID,
		a.Title, a.Description, a.Type, a.Reward,
		a.ExpiresAt, a.Status, effectsJSON, now, now,
	)
	if err != nil {
		return err
	}
	a.CreatedAt = now
	a.UpdatedAt = now
	return nil
}

// GetByID возвращает контракт по ID
func (r *AssignmentRepository) GetByID(id string) (*models.Assignment, error) {
	query := `SELECT id, world_id, author_type, author_id, title, description, type, reward, expires_at, status, effects, created_at, updated_at FROM assignments WHERE id = $1`
	row := r.db.QueryRow(query, id)
	var a models.Assignment
	var effectsJSON []byte
	err := row.Scan(
		&a.ID, &a.WorldID, &a.AuthorType, &a.AuthorID,
		&a.Title, &a.Description, &a.Type, &a.Reward,
		&a.ExpiresAt, &a.Status, &effectsJSON,
		&a.CreatedAt, &a.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(effectsJSON, &a.Effects); err != nil {
		return nil, err
	}
	return &a, nil
}

// GetByWorld возвращает все контракты для мира (без фильтров)
func (r *AssignmentRepository) GetByWorld(worldID string) ([]*models.Assignment, error) {
	return r.GetByWorldWithFilters(worldID, "", "", "", 0, 0)
}

// GetByWorldWithFilters возвращает контракты мира с фильтрацией
func (r *AssignmentRepository) GetByWorldWithFilters(worldID, typeFilter, authorFilter, search string, minReward, maxReward int) ([]*models.Assignment, error) {
	query := `SELECT id, world_id, author_type, author_id, title, description, type, reward, expires_at, status, effects, created_at, updated_at FROM assignments WHERE world_id = $1 AND status = 'open'`
	args := []interface{}{worldID}
	argIdx := 2

	if typeFilter != "" && typeFilter != "all" {
		query += fmt.Sprintf(" AND type = $%d", argIdx)
		args = append(args, typeFilter)
		argIdx++
	}
	if authorFilter != "" && authorFilter != "all" {
		query += fmt.Sprintf(" AND author_id = $%d", argIdx)
		args = append(args, authorFilter)
		argIdx++
	}
	if minReward > 0 {
		query += fmt.Sprintf(" AND reward >= $%d", argIdx)
		args = append(args, minReward)
		argIdx++
	}
	if maxReward > 0 {
		query += fmt.Sprintf(" AND reward <= $%d", argIdx)
		args = append(args, maxReward)
		argIdx++
	}
	if search != "" {
		query += fmt.Sprintf(" AND (title ILIKE $%d OR description ILIKE $%d)", argIdx, argIdx+1)
		args = append(args, "%"+search+"%", "%"+search+"%")
		argIdx += 2
	}
	query += " ORDER BY created_at DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assignments []*models.Assignment
	for rows.Next() {
		var a models.Assignment
		var effectsJSON []byte
		err := rows.Scan(
			&a.ID, &a.WorldID, &a.AuthorType, &a.AuthorID,
			&a.Title, &a.Description, &a.Type, &a.Reward,
			&a.ExpiresAt, &a.Status, &effectsJSON,
			&a.CreatedAt, &a.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(effectsJSON, &a.Effects); err != nil {
			return nil, err
		}
		assignments = append(assignments, &a)
	}
	return assignments, nil
}

// UpdateStatus обновляет статус контракта
func (r *AssignmentRepository) UpdateStatus(id, status string) error {
	query := `UPDATE assignments SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(query, status, id)
	return err
}