package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServerPort    string
	DBURL         string
	RedisURL      string
	TickInterval  time.Duration
	AdminPassword string
}

func Load() *Config {
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not set")
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		log.Fatal("REDIS_URL not set")
	}

	tickSec := os.Getenv("TICK_INTERVAL_SEC")
	if tickSec == "" {
		tickSec = "3"
	}
	sec, err := strconv.Atoi(tickSec)
	if err != nil {
		log.Fatalf("invalid TICK_INTERVAL_SEC: %v", err)
	}

	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		adminPassword = "admin123" // пароль по умолчанию, если не задан
		log.Println("⚠️ ADMIN_PASSWORD not set, using default: admin123")
	}

	return &Config{
		ServerPort:    port,
		DBURL:         dbURL,
		RedisURL:      redisURL,
		TickInterval:  time.Duration(sec) * time.Second,
		AdminPassword: adminPassword,
	}
}