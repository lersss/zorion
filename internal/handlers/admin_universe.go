package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"zorion/internal/generator"
	"zorion/internal/generator/faction"
	"zorion/internal/generator/galaxy"
	"zorion/internal/generator/planet"
)

var statusManager = generator.NewStatusManager()

func (h *AdminHandlers) GenerateUniverse(w http.ResponseWriter, r *http.Request) {
	if s := statusManager.Get(generator.JobGenerateUniverse); s != nil && s.Status == "running" {
		http.Error(w, "Generation already running", http.StatusConflict)
		return
	}

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
	statusManager.Start(generator.JobGenerateUniverse, req.WorldCount, cancel)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("❌ GenerateUniverse panic: %v", r)
				statusManager.Fail(generator.JobGenerateUniverse, "panic: "+r.(string))
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
			WorldSpread:    20.0, // Можно сделать параметром, но пока фиксированно
		}
		gen := galaxy.NewGenerator(&cfg)

		// Генерация миров
		worlds := gen.GenerateGalaxy()
		log.Printf("✅ Generated %d worlds", len(worlds))

		// Сохраняем миры в БД (вставка по одному или батчем)
		// Для простоты используем h.db, но можно и h.worldRepo, если есть метод Create
		tx, err := h.db.Begin()
		if err != nil {
			log.Printf("❌ GenerateUniverse: failed to start transaction: %v", err)
			statusManager.Fail(generator.JobGenerateUniverse, err.Error())
			return
		}
		defer tx.Rollback()

		// Очищаем старые миры (как в ClearUniverse)
		_, err = tx.Exec("DELETE FROM worlds")
		if err != nil {
			log.Printf("❌ GenerateUniverse: failed to clear worlds: %v", err)
			statusManager.Fail(generator.JobGenerateUniverse, err.Error())
			return
		}

		stmt, err := tx.Prepare(`
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
			_, err := stmt.Exec(
				world.ID,
				world.Name,
				world.CoordX,
				world.CoordY,
				world.SpectralClass,
				world.Temperature,
				world.CreatedAt,
				world.UpdatedAt,
			)
			if err != nil {
				log.Printf("❌ GenerateUniverse: failed to insert world %s: %v", world.ID, err)
				statusManager.Fail(generator.JobGenerateUniverse, err.Error())
				return
			}
			statusManager.Progress(generator.JobGenerateUniverse, i+1)
		}

		err = tx.Commit()
		if err != nil {
			log.Printf("❌ GenerateUniverse: failed to commit: %v", err)
			statusManager.Fail(generator.JobGenerateUniverse, err.Error())
			return
		}
		log.Printf("✅ GenerateUniverse: completed successfully, %d worlds saved", len(worlds))
		statusManager.Done(generator.JobGenerateUniverse)
	}()

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"status":"started"}`))
}

func (h *AdminHandlers) CancelGeneration(w http.ResponseWriter, r *http.Request) {
	job := r.URL.Query().Get("job")
	if job == "" {
		http.Error(w, "job parameter required", http.StatusBadRequest)
		return
	}
	jt := generator.JobType(job)
	s := statusManager.Get(jt)
	if s == nil || s.Status != "running" {
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
	total, processed, status, err := statusManager.GetStatus(jt)
	resp := map[string]interface{}{
		"total":     total,
		"processed": processed,
		"status":    status,
	}
	if err != "" {
		resp["error"] = err
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *AdminHandlers) ClearUniverse(w http.ResponseWriter, r *http.Request) {
	log.Printf("🗑️ ClearUniverse: deleting all worlds")
	_, err := h.db.Exec("DELETE FROM worlds")
	if err != nil {
		log.Printf("❌ ClearUniverse: error: %v", err)
		http.Error(w, "Failed to clear universe: "+err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("✅ ClearUniverse: cleared")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"cleared"}`))
}

func (h *AdminHandlers) GeneratePlanets(w http.ResponseWriter, r *http.Request) {
	if s := statusManager.Get(generator.JobGeneratePlanets); s != nil && s.Status == "running" {
		http.Error(w, "Generation already running", http.StatusConflict)
		return
	}

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
	statusManager.Start(generator.JobGeneratePlanets, len(worlds), cancel)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("❌ GeneratePlanets panic: %v", r)
				statusManager.Fail(generator.JobGeneratePlanets, "panic: "+r.(string))
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
				log.Printf("❌ GeneratePlanets: error generating planets for world %s: %v", world.ID, err)
				statusManager.Fail(generator.JobGeneratePlanets, err.Error())
				return
			}
			totalPlanets += count
			statusManager.Progress(generator.JobGeneratePlanets, i+1)
		}
		log.Printf("✅ GeneratePlanets: total planets generated = %d", totalPlanets)
		statusManager.Done(generator.JobGeneratePlanets)
	}()

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"status":"started"}`))
}

func (h *AdminHandlers) GenerateFactions(w http.ResponseWriter, r *http.Request) {
	if s := statusManager.Get(generator.JobGenerateFactions); s != nil && s.Status == "running" {
		http.Error(w, "Generation already running", http.StatusConflict)
		return
	}

	rows, err := h.db.Query(`
		SELECT COUNT(*) FROM planets 
		WHERE data->>'population' IS NOT NULL AND (data->>'population')::int > 0
	`)
	if err != nil {
		log.Printf("❌ GenerateFactions: failed to count habitable planets: %v", err)
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

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("❌ GenerateFactions panic: %v", r)
				statusManager.Fail(generator.JobGenerateFactions, "panic: "+r.(string))
			}
		}()
		log.Printf("🏛️ GenerateFactions: start")
		factionGen := faction.NewGenerator(h.db, 0)
		totalFactions, err := factionGen.GenerateFactions()
		if err != nil {
			log.Printf("❌ GenerateFactions: error: %v", err)
			statusManager.Fail(generator.JobGenerateFactions, err.Error())
			return
		}
		log.Printf("✅ GenerateFactions: generated %d factions", totalFactions)
		statusManager.Done(generator.JobGenerateFactions)
	}()

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"status":"started"}`))
}

func (h *AdminHandlers) GetStats(w http.ResponseWriter, r *http.Request) {
	var worldsCount, planetsCount int
	err := h.db.QueryRow("SELECT COUNT(*) FROM worlds").Scan(&worldsCount)
	if err != nil {
		http.Error(w, "Failed to count worlds", http.StatusInternalServerError)
		return
	}
	err = h.db.QueryRow("SELECT COUNT(*) FROM planets").Scan(&planetsCount)
	if err != nil {
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