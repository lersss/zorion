// internal/generator/planet/planet_data_batch.go
package planet

import (
	"database/sql"
	"fmt"
	"strings"
)

// batchInsertPlanets — вставляет пачку планет в public.planets
func (g *Generator) batchInsertPlanets(tx *sql.Tx, rows []interface{}) error {
	if len(rows) == 0 {
		return nil
	}
	query := fmt.Sprintf(
		"INSERT INTO public.planets (id, world_id, name, orbit_index, data, created_at, updated_at) VALUES %s",
		buildPlaceholders(len(rows), 7),
	)
	args := flattenRows(rows)
	_, err := tx.Exec(query, args...)
	return err
}

// batchInsertSettlements — вставляет пачку поселений
func (g *Generator) batchInsertSettlements(tx *sql.Tx, rows []interface{}) error {
	if len(rows) == 0 {
		return nil
	}
	query := fmt.Sprintf(
		"INSERT INTO public.settlements (id, planet_id, level, population, capacity, stability) VALUES %s",
		buildPlaceholders(len(rows), 6),
	)
	args := flattenRows(rows)
	_, err := tx.Exec(query, args...)
	return err
}

// batchInsertFactories — вставляет пачку заводов
func (g *Generator) batchInsertFactories(tx *sql.Tx, rows []interface{}) error {
	if len(rows) == 0 {
		return nil
	}
	query := fmt.Sprintf(
		"INSERT INTO public.factories (id, planet_id, name, type, input_resource, output_product, quality, status) VALUES %s",
		buildPlaceholders(len(rows), 8),
	)
	args := flattenRows(rows)
	_, err := tx.Exec(query, args...)
	return err
}

// batchInsertGoods — вставляет пачку партий товаров
func (g *Generator) batchInsertGoods(tx *sql.Tx, rows []interface{}) error {
	if len(rows) == 0 {
		return nil
	}
	query := fmt.Sprintf(
		"INSERT INTO public.goods_batches (id, planet_id, product_name, quantity, quality, producer_id, produced_at) VALUES %s",
		buildPlaceholders(len(rows), 7),
	)
	args := flattenRows(rows)
	_, err := tx.Exec(query, args...)
	return err
}

// batchInsertResources — вставляет пачку ресурсов
func (g *Generator) batchInsertResources(tx *sql.Tx, rows []interface{}) error {
	if len(rows) == 0 {
		return nil
	}
	query := fmt.Sprintf(`INSERT INTO public.resources (
		id, planet_id, name, category, hardness, elasticity, conductivity,
		heat_resistance, chemical_activity, density, biocompatibility,
		energy_density, volatility, created_at, updated_at
	) VALUES %s`, buildPlaceholders(len(rows), 15))
	args := flattenRows(rows)
	_, err := tx.Exec(query, args...)
	return err
}

// ==================== УТИЛИТЫ ДЛЯ БАТЧ-ВСТАВКИ ====================

// buildPlaceholders — строит "$1, $2, ... , $N" для одной строки
// и повторяет её для каждой строки.
// Пример: rows=2, cols=3 → "($1, $2, $3), ($4, $5, $6)"
func buildPlaceholders(rows, cols int) string {
	rowPlaceholder := "("
	for i := 1; i <= cols; i++ {
		if i > 1 {
			rowPlaceholder += ", "
		}
		rowPlaceholder += fmt.Sprintf("$%d", i)
	}
	rowPlaceholder += ")"

	valueStrings := make([]string, 0, rows)
	for i := 0; i < rows; i++ {
		// Сдвигаем номера $N на (i * cols)
		shifted := rowPlaceholder
		if i > 0 {
			shifted = shiftPlaceholders(rowPlaceholder, i*cols)
		}
		valueStrings = append(valueStrings, shifted)
	}
	return strings.Join(valueStrings, ", ")
}

// shiftPlaceholders — увеличивает все $N в строке на delta.
// Пример: "( $1, $2 )" + delta 2 → "( $3, $4 )"
func shiftPlaceholders(s string, delta int) string {
	if delta == 0 {
		return s
	}
	var b strings.Builder
	b.Grow(len(s) + 8)

	i := 0
	for i < len(s) {
		if s[i] == '$' {
			b.WriteByte('$')
			i++
			numStart := i
			for i < len(s) && s[i] >= '0' && s[i] <= '9' {
				i++
			}
			if numStart == i {
				// Не число после $ — оставляем как есть
				continue
			}
			numStr := s[numStart:i]
			num := 0
			for _, ch := range numStr {
				num = num*10 + int(ch-'0')
			}
			fmt.Fprintf(&b, "%d", num+delta)
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

// flattenRows — превращает []interface{} из [][]interface{} в плоский []interface{}
func flattenRows(rows []interface{}) []interface{} {
	total := 0
	for _, r := range rows {
		if rowSlice, ok := r.([]interface{}); ok {
			total += len(rowSlice)
		}
	}
	result := make([]interface{}, 0, total)
	for _, r := range rows {
		if rowSlice, ok := r.([]interface{}); ok {
			result = append(result, rowSlice...)
		}
	}
	return result
}