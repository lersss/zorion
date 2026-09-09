package main

import (
	"database/sql"
	"log"
	"net/http"

	"context"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"

	"zorion/internal/auth"
	"zorion/internal/config"
	"zorion/internal/generator/planet"
	"zorion/internal/handlers"
	"zorion/internal/repository"
	"zorion/internal/travel"
)

var db *sql.DB
var rdb *redis.Client

func main() {
	cfg := config.Load()
	log.Printf("Запуск сервера Zorion на порту %s", cfg.ServerPort)
	log.Printf("Интервал тика: %v", cfg.TickInterval)

	var err error
	db, err = sql.Open("postgres", cfg.DBURL)
	if err != nil {
		log.Fatalf("Ошибка подключения к PostgreSQL: %v", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatalf("PostgreSQL не отвечает: %v", err)
	}
	log.Println("✅ PostgreSQL подключен")

	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatalf("Ошибка парсинга Redis URL: %v", err)
	}
	rdb = redis.NewClient(opt)
	if err = rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Redis не отвечает: %v", err)
	}
	log.Println("✅ Redis подключен")

	// ------------------------------------------------------------------
	// ЗАГРУЗКА АРХЕТИПОВ ПЛАНЕТ (новая система типов)
	if err := planet.LoadArchetypes("config/planet_archetypes.json"); err != nil {
		log.Printf("⚠️ Не удалось загрузить архетипы планет: %v, использую fallback", err)
	} else {
		log.Println("✅ Архетипы планет загружены")
	}
	// ------------------------------------------------------------------

	worldRepo := repository.NewWorldRepository(db)
	locationRepo := repository.NewLocationRepository(db)
	assignmentRepo := repository.NewAssignmentRepository(db)
	userRepo := repository.NewUserRepository(db)

	travelManager := travel.NewManager()
	wsHub := handlers.NewWebSocketHub()

	testHandlers := handlers.NewTestHandlers(worldRepo, locationRepo, assignmentRepo)
	worldHandlers := handlers.NewWorldHandlers(worldRepo, locationRepo, assignmentRepo)
	authHandlers := handlers.NewAuthHandlers(userRepo, worldRepo)
	travelHandlers := handlers.NewTravelHandlers(worldRepo, userRepo, travelManager)
	wsHandler := handlers.NewWebSocketHandler(wsHub)
	contractHandlers := handlers.NewContractHandlers(assignmentRepo, userRepo)
	adminHandlers := handlers.NewAdminHandlers(worldRepo, db)

	// API открытые
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/status", statusHandler)
	http.HandleFunc("/register", authHandlers.Register)
	http.HandleFunc("/login", authHandlers.Login)

	// API защищённые JWT
	http.HandleFunc("/create-test-data", auth.AuthMiddleware(testHandlers.CreateTestData))
	http.HandleFunc("/worlds", auth.AuthMiddleware(worldHandlers.GetAllWorlds))
	http.HandleFunc("/worlds/", auth.AuthMiddleware(worldHandlers.GetWorld))
	http.HandleFunc("/travel", auth.AuthMiddleware(travelHandlers.StartTravel))
	http.HandleFunc("/me", auth.AuthMiddleware(authHandlers.GetMe))

	// API контрактов
	http.HandleFunc("/api/contracts", auth.AuthMiddleware(contractHandlers.GetContracts))
	http.HandleFunc("/api/contracts/take", auth.AuthMiddleware(contractHandlers.TakeContract))
	http.HandleFunc("/api/contracts/complete-test", auth.AuthMiddleware(contractHandlers.CompleteTestContract))

	// API планет
	http.HandleFunc("/api/worlds/", auth.AuthMiddleware(adminHandlers.GetPlanetsByWorld))

	// WebSocket
	http.HandleFunc("/ws", auth.AuthMiddleware(wsHandler.ServeWS))

	// Админка (пароль)
	http.HandleFunc("/admin/worlds", auth.AdminAuth(adminHandlers.GetAllWorlds))
	http.HandleFunc("/admin/worlds/delete", auth.AdminAuth(adminHandlers.DeleteWorld))
	http.HandleFunc("/admin/worlds/create", auth.AdminAuth(adminHandlers.CreateWorld))
	http.HandleFunc("/admin/generate", auth.AdminAuth(adminHandlers.GenerateUniverse))
	http.HandleFunc("/admin/stats", auth.AdminAuth(adminHandlers.GetStats))
	http.HandleFunc("/admin/generate-status", auth.AdminAuth(adminHandlers.GenerateStatus))
	http.HandleFunc("/admin/clear", auth.AdminAuth(adminHandlers.ClearUniverse))
	http.HandleFunc("/admin/generate-planets", auth.AdminAuth(adminHandlers.GeneratePlanets))
	http.HandleFunc("/admin/generate-factions", auth.AdminAuth(adminHandlers.GenerateFactions))
	http.HandleFunc("/admin/generate-cancel", auth.AdminAuth(adminHandlers.CancelGeneration))
	http.HandleFunc("/admin", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/admin.html")
	})

	// Статика
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static"))))

	// Страницы
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "./web/index.html")
			return
		}
		http.FileServer(http.Dir("./web")).ServeHTTP(w, r)
	})
	http.HandleFunc("/assignments", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/assignments.html")
	})
	http.HandleFunc("/login-page", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/login.html")
	})
	http.HandleFunc("/register-page", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/register.html")
	})
	http.HandleFunc("/map", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/map.html")
	})
	http.HandleFunc("/contracts", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/contracts.html")
	})

	log.Println("🚀 Сервер Zorion запущен и работает")
	log.Fatal(http.ListenAndServe(":"+cfg.ServerPort, nil))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	if err := db.Ping(); err != nil {
		http.Error(w, "DB connection failed", http.StatusInternalServerError)
		return
	}
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		http.Error(w, "Redis connection failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("All systems ready"))
}