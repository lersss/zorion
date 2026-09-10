// internal/handlers/filter_worlds_handler.go
package handlers

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ==================== ТИПЫ ====================

// worldCluster — один кластер миров для карты.
// CellX/CellY — идентификатор ячейки (для отладки).
// X/Y — центроид кластера в мировых координатах.
// Count — сколько миров в ячейке.
// Sample* — данные представителя для cnt=1 (для tooltip и цвета).
type worldCluster struct {
	CellX          int64   `json:"cx"`
	CellY          int64   `json:"cy"`
	Count          int     `json:"cnt"`
	X              float64 `json:"x"`
	Y              float64 `json:"y"`
	SampleID       string  `json:"sid,omitempty"`
	SampleName     string  `json:"sname,omitempty"`
	SampleSpectral string  `json:"sspec,omitempty"`
}

// ==================== ЛИМИТЫ ====================

const (
	// Минимальный размер ячейки в мировых координатах.
	minCellSize = 0.5

	// Максимум ячеек по ширине viewport — защита от DoS,
	// если клиент пришлёт микроскопический cell.
	maxCellCols = 250

	// Максимум кластеров в ответе (страховка на случай, если что-то
	// пойдёт не так и ячеек окажется больше, чем мы ожидаем).
	maxClusterReturn = 20000
)

// ==================== ХЕНДЛЕР ====================

// FilterWorldsHandler — возвращает кластеры миров для видимой области карты.
// Клиент присылает границы viewport'а и размер ячейки — сервер делает
// GROUP BY и отдаёт компактный набор кластеров вместо всех миров.
func (h *AdminHandlers) FilterWorldsHandler(w http.ResponseWriter, r *http.Request) {
	tStart := time.Now()

	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("🔥 PANIC in FilterWorldsHandler: %v", rec)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}()

	q := r.URL.Query()

	// --- Геометрия viewport ---
	xMin, err := parseFloatParam(q.Get("x_min"))
	if err != nil {
		http.Error(w, "invalid x_min", http.StatusBadRequest)
		return
	}
	xMax, err := parseFloatParam(q.Get("x_max"))
	if err != nil {
		http.Error(w, "invalid x_max", http.StatusBadRequest)
		return
	}
	yMin, err := parseFloatParam(q.Get("y_min"))
	if err != nil {
		http.Error(w, "invalid y_min", http.StatusBadRequest)
		return
	}
	yMax, err := parseFloatParam(q.Get("y_max"))
	if err != nil {
		http.Error(w, "invalid y_max", http.StatusBadRequest)
		return
	}
	cell, err := parseFloatParam(q.Get("cell"))
	if err != nil || cell <= 0 {
		http.Error(w, "invalid cell", http.StatusBadRequest)
		return
	}
	if xMax <= xMin || yMax <= yMin {
		http.Error(w, "invalid bounds (max must be > min)", http.StatusBadRequest)
		return
	}

	// Ограничиваем ячейку снизу: не больше maxCellCols ячеек по ширине.
	minCell := (xMax - xMin) / float64(maxCellCols)
	if cell < minCell {
		cell = minCell
	}
	if cell < minCellSize {
		cell = minCellSize
	}

	// --- Опциональные фильтры ---
	hasPlanets := q.Get("has_planets") == "true"
	hasLife := q.Get("has_life") == "true"
	hasHabitable := q.Get("has_habitable") == "true"
	planetType := q.Get("planet_type")
	resourceCategory := q.Get("resource_category")

	// --- Сборка SQL ---
	// $1 = xMin, $2 = xMax, $3 = yMin, $4 = yMax, $5 = cell
	args := []interface{}{xMin, xMax, yMin, yMax, cell}
	argN := 6

	filterSQL := ""
	if hasPlanets {
		filterSQL += ` AND EXISTS (SELECT 1 FROM planets p WHERE p.world_id = w.id)`
	}
	if hasLife {
		filterSQL += ` AND EXISTS (SELECT 1 FROM planets p WHERE p.world_id = w.id AND (p.data->>'life')::boolean = true)`
	}
	if hasHabitable {
		filterSQL += ` AND EXISTS (SELECT 1 FROM planets p WHERE p.world_id = w.id AND (p.data->>'habitable')::boolean = true)`
	}
	if planetType != "" {
		filterSQL += ` AND EXISTS (SELECT 1 FROM planets p WHERE p.world_id = w.id AND LOWER(p.data->>'type') = $` + strconv.Itoa(argN) + `)`
		args = append(args, strings.ToLower(planetType))
		argN++
	}
	if resourceCategory != "" {
		filterSQL += ` AND EXISTS (
			SELECT 1 FROM planets p
			WHERE p.world_id = w.id
			  AND p.data->'resources'->>$` + strconv.Itoa(argN) + ` IS NOT NULL
			  AND (p.data->'resources'->>$` + strconv.Itoa(argN) + `)::float > 0.3
		)`
		args = append(args, resourceCategory)
		argN++
	}

	// Основной запрос: фильтрация → группировка по ячейкам.
	sqlQuery := `
		WITH filtered AS (
			SELECT id, name, coord_x, coord_y, spectral_class
			FROM worlds w
			WHERE coord_x BETWEEN $1 AND $2
			  AND coord_y BETWEEN $3 AND $4
			  ` + filterSQL + `
		)
		SELECT
			FLOOR(coord_x / $5)::bigint                AS cell_x,
			FLOOR(coord_y / $5)::bigint                AS cell_y,
			COUNT(*)::int                              AS cnt,
			AVG(coord_x)::float8                       AS avg_x,
			AVG(coord_y)::float8                       AS avg_y,
			(ARRAY_AGG(id ORDER BY id))[1]             AS sample_id,
			(ARRAY_AGG(name ORDER BY id))[1]           AS sample_name,
			(ARRAY_AGG(spectral_class ORDER BY id))[1] AS sample_spectral
		FROM filtered
		GROUP BY cell_x, cell_y
		LIMIT ` + strconv.Itoa(maxClusterReturn)

	// --- Запрос ---
	tQuery := time.Now()
	rows, err := h.db.QueryContext(r.Context(), sqlQuery, args...)
	if err != nil {
		log.Printf("❌ FilterWorlds query error: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	queryDur := time.Since(tQuery)

	// --- Чтение ---
	clusters := make([]worldCluster, 0, 512)
	for rows.Next() {
		var c worldCluster
		var sampleID, sampleName, sampleSpec *string
		if err := rows.Scan(
			&c.CellX, &c.CellY, &c.Count, &c.X, &c.Y,
			&sampleID, &sampleName, &sampleSpec,
		); err != nil {
			log.Printf("❌ FilterWorlds scan error: %v", err)
			http.Error(w, "Scan error", http.StatusInternalServerError)
			return
		}
		// Sample* — только для cnt=1, чтобы не раздувать payload.
		if c.Count == 1 {
			if sampleID != nil {
				c.SampleID = *sampleID
			}
			if sampleName != nil {
				c.SampleName = *sampleName
			}
			if sampleSpec != nil {
				c.SampleSpectral = *sampleSpec
			}
		}
		clusters = append(clusters, c)
	}
	if err = rows.Err(); err != nil {
		log.Printf("❌ FilterWorlds rows error: %v", err)
		http.Error(w, "Rows error", http.StatusInternalServerError)
		return
	}
	scanDur := time.Since(tQuery) - queryDur

	// --- Ответ ---
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")

	var writer io.Writer = w
	useGzip := strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")
	if useGzip {
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		writer = gz
	}

	tEncode := time.Now()
	if err := json.NewEncoder(writer).Encode(clusters); err != nil {
		// Заголовки уже улетели — http.Error нельзя. Просто логируем.
		log.Printf("⚠️ FilterWorlds encode error (клиент отвалился?): %v", err)
		return
	}
	encodeDur := time.Since(tEncode)

	log.Printf(
		"🗺️  FilterWorlds: bounds=(%.0f,%.0f)-(%.0f,%.0f) cell=%.2f → %d кластеров, query=%v scan=%v encode=%v total=%v gzip=%v",
		xMin, yMin, xMax, yMax, cell,
		len(clusters),
		queryDur.Round(time.Millisecond),
		scanDur.Round(time.Millisecond),
		encodeDur.Round(time.Millisecond),
		time.Since(tStart).Round(time.Millisecond),
		useGzip,
	)
}

// ==================== УТИЛИТЫ ====================

func parseFloatParam(s string) (float64, error) {
	if s == "" {
		return 0, strconv.ErrSyntax
	}
	return strconv.ParseFloat(s, 64)
}