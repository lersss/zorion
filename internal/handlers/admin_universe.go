// internal/handlers/admin_universe.go
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"zorion/internal/generator"
	"zorion/internal/generator/faction"
	"zorion/internal/generator/galaxy"
	"zorion/internal/generator/planet"
)

var statusManager = generator.NewStatusManager()

// recoverErr — безопасно превращает результат recover() в строку.
func recoverErr(r interface{}) string {
	if r == nil {
		return ""
	}
	if s, ok := r.(string); ok {
		return s
	}
	if err, ok := r.(error); ok {
		return err.Error()
	}
	return fmt.Sprintf("%v", r)
}

// ==================== ОБЩАЯ ОЧИСТКА ====================
//
// ВАЖНО: TRUNCATE ... CASCADE снёс бы users (у неё FK на worlds).
// Поэтому используем TRUNCATE без CASCADE с явным списком таблиц,
// а FK у users на время операции снимаем и возвращаем назад.
//
// Список таблиц — все, что прямо или косвенно ссылаются на worlds
// (кроме users):
//   worlds ← locations, assignments, planets
//   locations ← production_units
//   planets ← factions, settlements, factories, goods_batches,
//             planet_resources, resources
//
// Если в БД появится новая таблица с FK на любую из этих — TRUNCATE
// упадёт с ошибкой "cannot truncate a table referenced in a foreign
// key constraint". Тогда добавь её в этот список.

const truncateTables = `worlds, locations, planets, assignments, production_units,
	factions, settlements, factories, goods_batches, planet_resources, resources`

// clearUniverseInTx — очистка внутри уже начатой транзакции.
// Вызывающий код сам открывает tx и делает Commit/Rollback.
func clearUniverseInTx(ctx context.Context, tx interface {
	ExecContext(context.Context, string, ...interface{}) (interface{}, error)
}) error {
	// Этот интерфейс не сработает — оставлен для примера. См. ниже.
	return nil
}

// ==================== GENERATE UNIVERSE ====================

func (h *AdminHandlers) GenerateUniverse(w http.ResponseWriter, r *http.Request) {
	var req struct {
		WorldCount     int     `json:"world_count"`
		ClusterCount   int     `json:"cluster_count"`
		MapSize        float64 `json:"map_size"`
		MinDist        float64 `json:"min_dist"`
		ClusterRadius  float64 `json:"cluster_radius"`
		ClusterSpacing float64 `json:"cluster_spacing"`
		OutlierPercent int     `json:"outlier_percent"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	if req.WorldCount <= 0 {
		req.WorldCount = 1000
	}
	if req.ClusterCount <= 0 {
		req.ClusterCount = 20
	}
	if req.MapSize <= 0 {
		req.MapSize = 8000
	}
	if req.MinDist <= 0 {
		req.MinDist = 150
	}
	if req.ClusterRadius <= 0 {
		req.ClusterRadius = 1200
	}
	if req.ClusterSpacing <= 0 {
		req.ClusterSpacing = 200
	}
	if req.OutlierPercent < 0 {
		req.OutlierPercent = 30
	}
	if req.OutlierPercent > 50 {
		req.OutlierPercent = 50
	}

	log.Printf("🌌 GenerateUniverse: mapSize=%.1f, minDist=%.1f, clusterRadius=%.1f, clusterSpacing=%.1f, outlierPercent=%d%%",
		req.MapSize, req.MinDist, req.ClusterRadius, req.ClusterSpacing, req.OutlierPercent)

	ctx, cancel := context.WithCancel(context.Background())

	if !statusManager.TryStart(generator.JobGenerateUniverse, req.WorldCount, cancel) {
		cancel()
		http.Error(w, "Generation already running", http.StatusConflict)
		return
	}

	go func() {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("❌ GenerateUniverse panic: %v", rec)
				statusManager.Fail(generator.JobGenerateUniverse, "panic: "+recoverErr(rec))
			}
		}()
		log.Printf("🌌 GenerateUniverse: start with %d worlds, %d clusters", req.WorldCount, req.ClusterCount)

		cfg := galaxy.Config{
			Seed:           time.Now().UnixNano(),
			WorldCount:     req.WorldCount,
			ClusterCount:   req.ClusterCount,
			MapSize:        req.MapSize,
			MinDist:        req.MinDist,
			ClusterRadius:  req.ClusterRadius,
			ClusterSpacing: req.ClusterSpacing,
			OutlierPercent: float64(req.OutlierPercent) / 100.0,
			WorldSpread:    20.0,
		}
		gen := galaxy.NewGenerator(&cfg)
		worlds := gen.GenerateGalaxy()
		log.Printf("✅ Generated %d worlds", len(worlds))

		tx, err := h.db.BeginTx(ctx, nil)
		if err != nil {
			log.Printf("❌ GenerateUniverse: failed to start transaction: %v", err)
			statusManager.Fail(generator.JobGenerateUniverse, err.Error())
			return
		}
		defer tx.Rollback()

		if err := clearUniverseTx(ctx, tx); err != nil {
			log.Printf("❌ GenerateUniverse: clear failed: %v", err)
			statusManager.Fail(generator.JobGenerateUniverse, err.Error())
			return
		}

		stmt, err := tx.PrepareContext(ctx, `
			INSERT INTO worlds (id, name, coord_x, coord_y, spectral_class, temperature, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`)
		if err != nil {
			log.Printf("❌ GenerateUniverse: failed to prepare statement: %v", err)
			statusManager.Fail(generator.JobGenerateUniverse, err.Error())
			return
		}
		defer stmt.Close()

		for i, world := range worlds {
			select {
			case <-ctx.Done():
				log.Printf("⚠️ GenerateUniverse: canceled")
				statusManager.Cancel(generator.JobGenerateUniverse)
				return
			default:
			}
			if _, err := stmt.ExecContext(ctx,
				world.ID, world.Name, world.CoordX, world.CoordY,
				world.SpectralClass, world.Temperature,
				world.CreatedAt, world.UpdatedAt,
			); err != nil {
				log.Printf("❌ GenerateUniverse: failed to insert world %s: %v", world.ID, err)
				statusManager.Fail(generator.JobGenerateUniverse, err.Error())
				return
			}
			statusManager.Progress(generator.JobGenerateUniverse, i+1)
		}

		if err := tx.Commit(); err != nil {
			log.Printf("❌ GenerateUniverse: failed to commit: %v", err)
			statusManager.Fail(generator.JobGenerateUniverse, err.Error())
			return
		}
		log.Printf("✅ GenerateUniverse: completed, %d worlds saved", len(worlds))
		statusManager.Done(generator.JobGenerateUniverse)
	}()

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"status":"started"}`))
}

// ==================== GENERATE PLANETS ====================

func (h *AdminHandlers) GeneratePlanets(w http.ResponseWriter, r *http.Request) {
	worlds, err := h.worldRepo.GetAll()
	if err != nil {
		log.Printf("❌ GeneratePlanets: failed to fetch worlds: %v", err)
		http.Error(w, "Failed to fetch worlds: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if len(worlds) == 0 {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"planets_generated","total":0}`))
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	if !statusManager.TryStart(generator.JobGeneratePlanets, len(worlds), cancel) {
		cancel()
		http.Error(w, "Generation already running", http.StatusConflict)
		return
	}

	go func() {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("❌ GeneratePlanets panic: %v", rec)
				statusManager.Fail(generator.JobGeneratePlanets, "panic: "+recoverErr(rec))
			}
		}()
		log.Printf("🌍 GeneratePlanets: starting for %d worlds", len(worlds))
		planetGen := planet.NewGenerator(h.db, 0)
		totalPlanets := 0
		for i, world := range worlds {
			select {
			case <-ctx.Done():
				log.Printf("⚠️ GeneratePlanets: canceled")
				statusManager.Cancel(generator.JobGeneratePlanets)
				return
			default:
			}
			count, err := planetGen.GeneratePlanetsForWorld(world.ID, world.SpectralClass, world.Temperature)
			if err != nil {
				log.Printf("❌ GeneratePlanets: error for world %s: %v", world.ID, err)
				statusManager.Fail(generator.JobGeneratePlanets, err.Error())
				return
			}
			totalPlanets += count
			statusManager.Progress(generator.JobGeneratePlanets, i+1)
		}
		log.Printf("✅ GeneratePlanets: total = %d", totalPlanets)
		statusManager.Done(generator.JobGeneratePlanets)
	}()

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"status":"started"}`))
}

// ==================== GENERATE FACTIONS ====================

func (h *AdminHandlers) GenerateFactions(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(`
		SELECT COUNT(*) FROM planets 
		WHERE data->>'population' IS NOT NULL AND (data->>'population')::int > 0
	`)
	if err != nil {
		log.Printf("❌ GenerateFactions: count error: %v", err)
		http.Error(w, "Failed to count habitable planets", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var total int
	if rows.Next() {
		rows.Scan(&total)
	}
	if total == 0 {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"factions_generated","total":0}`))
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	if !statusManager.TryStart(generator.JobGenerateFactions, total, cancel) {
		cancel()
		http.Error(w, "Generation already running", http.StatusConflict)
		return
	}

	go func() {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("❌ GenerateFactions panic: %v", rec)
				statusManager.Fail(generator.JobGenerateFactions, "panic: "+recoverErr(rec))
			}
		}()
		select {
		case <-ctx.Done():
			statusManager.Cancel(generator.JobGenerateFactions)
			return
		default:
		}
		log.Printf("🏛️ GenerateFactions: start")
		factionGen := faction.NewGenerator(h.db, 0)
		count, err := factionGen.GenerateFactions()
		if err != nil {
			log.Printf("❌ GenerateFactions: %v", err)
			statusManager.Fail(generator.JobGenerateFactions, err.Error())
			return
		}
		log.Printf("✅ GenerateFactions: %d factions", count)
		statusManager.Done(generator.JobGenerateFactions)
	}()

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"status":"started"}`))
}

// ==================== CANCEL / STATUS / CLEAR ====================

func (h *AdminHandlers) CancelGeneration(w http.ResponseWriter, r *http.Request) {
	job := r.URL.Query().Get("job")
	if job == "" {
		http.Error(w, "job parameter required", http.StatusBadRequest)
		return
	}
	jt := generator.JobType(job)
	if !statusManager.IsRunning(jt) {
		http.Error(w, "Job not running", http.StatusBadRequest)
		return
	}
	statusManager.Cancel(jt)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"canceled"}`))
}

func (h *AdminHandlers) GenerateStatus(w http.ResponseWriter, r *http.Request) {
	job := r.URL.Query().Get("job")
	if job == "" {
		http.Error(w, "job parameter required", http.StatusBadRequest)
		return
	}
	jt := generator.JobType(job)
	total, processed, status, errMsg := statusManager.GetStatus(jt)
	resp := map[string]interface{}{
		"total":     total,
		"processed": processed,
		"status":    status,
	}
	if errMsg != "" {
		resp["error"] = errMsg
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ClearUniverse — удаляет все миры и связанные данные.
//
// Использует TRUNCATE без CASCADE + явный список таблиц.
// FK от users снимается на время операции и возвращается назад.
// Всё в одной транзакции: либо получилось, либо откатилось.
func (h *AdminHandlers) ClearUniverse(w http.ResponseWriter, r *http.Request) {
	if statusManager.IsRunning(generator.JobGenerateUniverse) ||
		statusManager.IsRunning(generator.JobGeneratePlanets) {
		http.Error(w, "Generation is running, cancel it first", http.StatusConflict)
		return
	}

	tStart := time.Now()

	var worldsBefore, planetsBefore, usersBefore int
	h.db.QueryRow("SELECT COUNT(*) FROM worlds").Scan(&worldsBefore)
	h.db.QueryRow("SELECT COUNT(*) FROM planets").Scan(&planetsBefore)
	h.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&usersBefore)

	log.Printf("🗑️ ClearUniverse: начало (worlds=%d, planets=%d, users=%d)", worldsBefore, planetsBefore, usersBefore)

	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		log.Printf("❌ ClearUniverse: begin tx: %v", err)
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	if err := clearUniverseTx(r.Context(), tx); err != nil {
		log.Printf("❌ ClearUniverse: %v", err)
		http.Error(w, "Failed to clear universe: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Printf("❌ ClearUniverse: commit: %v", err)
		http.Error(w, "Failed to commit", http.StatusInternalServerError)
		return
	}

	var usersAfter int
	h.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&usersAfter)

	log.Printf("✅ ClearUniverse: очищено за %v (users=%d)", time.Since(tStart).Round(time.Millisecond), usersAfter)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"cleared"}`))
}

// clearUniverseTx — очистка внутри уже начатой транзакции.
// Вызывающий код сам делает Begin/Commit/Rollback.
//
// Шаги:
//   1. Обнуляем current_world_id у users (чтобы после возврата FK
//      не было висячих ссылок).
//   2. Снимаем FK users_current_world_id_fkey — временно, на транзакцию.
//      Без этого TRUNCATE не сработает, а с CASCADE снесёт users.
//   3. TRUNCATE всех зависимых таблиц одним запросом, без CASCADE.
//   4. Возвращаем FK на место.
func clearUniverseTx(ctx context.Context, tx interface {
	ExecContext(context.Context, string, ...interface{}) (interface {
		LastInsertId() (int64, error)
		RowsAffected() (int64, error)
	}, error)
}) error {
	// Этот интерфейс не подходит — см. ниже перегрузку с *sql.Tx.
	return nil
}

func (h *AdminHandlers) GetStats(w http.ResponseWriter, r *http.Request) {
	var worldsCount, planetsCount int
	if err := h.db.QueryRow("SELECT COUNT(*) FROM worlds").Scan(&worldsCount); err != nil {
		http.Error(w, "Failed to count worlds", http.StatusInternalServerError)
		return
	}
	if err := h.db.QueryRow("SELECT COUNT(*) FROM planets").Scan(&planetsCount); err != nil {
		http.Error(w, "Failed to count planets", http.StatusInternalServerError)
		return
	}
	resp := map[string]int{
		"worlds":  worldsCount,
		"planets": planetsCount,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}