// internal/repository/compatibility_repository.go
package repository

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"zorion/internal/models"
)

// CompatibilityRepository — работа с матрицей совместимости в БД.
//
// Хранит только ЗАПРЕЩЁННЫЕ пары (compatible = FALSE).
// Пара с compatible = TRUE в БД не хранится — она удаляется,
// потому что «разрешено» — это отсутствие записи.
type CompatibilityRepository struct {
	db *sql.DB
}

// NewCompatibilityRepository — конструктор.
func NewCompatibilityRepository(db *sql.DB) *CompatibilityRepository {
	return &CompatibilityRepository{db: db}
}

// ==================== ЧТЕНИЕ ====================

// LoadAll — загружает все запрещённые пары заданной категории.
// Возвращает список пар (Compatible всегда false).
func (r *CompatibilityRepository) LoadAll(category string) ([]*models.CompatibilityPair, error) {
	rows, err := r.db.Query(`
		SELECT id, category, type_a, type_b, compatible, updated_at
		FROM compatibility_matrix
		WHERE category = $1
		ORDER BY type_a, type_b
	`, category)
	if err != nil {
		return nil, fmt.Errorf("load compatibility: %w", err)
	}
	defer rows.Close()

	result := []*models.CompatibilityPair{}
	for rows.Next() {
		p := &models.CompatibilityPair{}
		if err := rows.Scan(
			&p.ID, &p.Category, &p.TypeA, &p.TypeB,
			&p.Compatible, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan compatibility row: %w", err)
		}
		result = append(result, p)
	}
	return result, nil
}

// CountAll — сколько всего запрещённых пар в БД (для проверки «пустая ли»).
func (r *CompatibilityRepository) CountAll() (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM compatibility_matrix`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count compatibility: %w", err)
	}
	return count, nil
}

// ==================== ЗАПИСЬ ====================

// Set — устанавливает одну пару.
//   - compatible = true  → удаляет пару из БД (разрешено = отсутствие записи).
//   - compatible = false → вставляет или обновляет пару.
func (r *CompatibilityRepository) Set(
	category, typeA, typeB string,
	compatible bool,
) error {
	if compatible {
		return r.Delete(category, typeA, typeB)
	}
	return r.insertPair(category, typeA, typeB)
}

// SetBatch — батч-обновление в одной транзакции.
func (r *CompatibilityRepository) SetBatch(
	category string,
	pairs []models.CompatibilityUpdatePair,
) error {
	if len(pairs) == 0 {
		return nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	for _, p := range pairs {
		if p.Compatible {
			if err := deletePairTx(tx, category, p.TypeA, p.TypeB); err != nil {
				return err
			}
		} else {
			if err := insertPairTx(tx, category, p.TypeA, p.TypeB); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

// Delete — удаляет пару из БД (делает её разрешённой).
func (r *CompatibilityRepository) Delete(category, typeA, typeB string) error {
	_, err := r.db.Exec(`
		DELETE FROM compatibility_matrix
		WHERE category = $1 AND type_a = $2 AND type_b = $3
	`, category, typeA, typeB)
	if err != nil {
		return fmt.Errorf("delete compatibility pair: %w", err)
	}
	return nil
}

// ResetCategory — полная очистка категории (перед загрузкой дефолтов).
func (r *CompatibilityRepository) ResetCategory(category string) error {
	_, err := r.db.Exec(
		`DELETE FROM compatibility_matrix WHERE category = $1`,
		category,
	)
	if err != nil {
		return fmt.Errorf("reset compatibility category: %w", err)
	}
	return nil
}

// ==================== ВНУТРЕННЕЕ ====================

func (r *CompatibilityRepository) insertPair(category, typeA, typeB string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if err := insertPairTx(tx, category, typeA, typeB); err != nil {
		return err
	}
	return tx.Commit()
}

func insertPairTx(tx *sql.Tx, category, typeA, typeB string) error {
	// Нормализуем порядок: type_a <= type_b (чтобы избежать дублей A|B и B|A)
	if typeA > typeB {
		typeA, typeB = typeB, typeA
	}

	_, err := tx.Exec(`
		INSERT INTO compatibility_matrix (id, category, type_a, type_b, compatible, updated_at)
		VALUES ($1, $2, $3, $4, FALSE, NOW())
		ON CONFLICT (category, type_a, type_b)
		DO UPDATE SET compatible = FALSE, updated_at = NOW()
	`, uuid.New().String(), category, typeA, typeB)
	if err != nil {
		return fmt.Errorf("insert compatibility pair: %w", err)
	}
	return nil
}

func deletePairTx(tx *sql.Tx, category, typeA, typeB string) error {
	if typeA > typeB {
		typeA, typeB = typeB, typeA
	}
	_, err := tx.Exec(`
		DELETE FROM compatibility_matrix
		WHERE category = $1 AND type_a = $2 AND type_b = $3
	`, category, typeA, typeB)
	if err != nil {
		return fmt.Errorf("delete compatibility pair: %w", err)
	}
	return nil
}