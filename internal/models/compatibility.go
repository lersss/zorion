// internal/models/compatibility.go
package models

import "time"

// CompatibilityPair — одна запись в матрице совместимости.
//
// Хранит ТОЛЬКО запрещённые пары (Compatible = false).
// Если пары нет в БД — формы совместимы по умолчанию.
type CompatibilityPair struct {
	ID         string    `json:"id"`
	Category   string    `json:"category"`   // "surface" | "subterrain"
	TypeA      string    `json:"type_a"`
	TypeB      string    `json:"type_b"`
	Compatible bool      `json:"compatible"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ==================== КОНСТАНТЫ КАТЕГОРИЙ ====================

const (
	CompatCategorySurface    = "surface"
	CompatCategorySubterrain = "subterrain"
)

// ==================== ПУБЛИЧНЫЕ ХЕЛПЕРЫ ====================

// IsSurface — относится ли пара к поверхности.
func (p *CompatibilityPair) IsSurface() bool {
	return p.Category == CompatCategorySurface
}

// IsSubterrain — относится ли пара к недрам.
func (p *CompatibilityPair) IsSubterrain() bool {
	return p.Category == CompatCategorySubterrain
}

// ==================== API-ФОРМАТЫ ====================

// CompatibilityMatrixDTO — формат ответа админке.
// Удобно для фронта: {категория: {"форма_а": ["форма_б", "форма_в"]}}.
type CompatibilityMatrixDTO struct {
	Category string              `json:"category"`
	Forbid   map[string][]string `json:"forbid"`
	Total    int                 `json:"total"`
}

// CompatibilityUpdateRequest — тело POST-запроса на обновление.
// Батч: массив пар, которые надо установить.
type CompatibilityUpdateRequest struct {
	Category string                    `json:"category"`
	Pairs    []CompatibilityUpdatePair `json:"pairs"`
}

// CompatibilityUpdatePair — одна пара в запросе на обновление.
type CompatibilityUpdatePair struct {
	TypeA      string `json:"type_a"`
	TypeB      string `json:"type_b"`
	Compatible bool   `json:"compatible"`
}