package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/saltybytes/saltybytes-dashboard/internal/cache"
	"github.com/saltybytes/saltybytes-dashboard/internal/config"
	"github.com/saltybytes/saltybytes-dashboard/internal/db"
	"github.com/saltybytes/saltybytes-dashboard/internal/ratecard"
	"github.com/saltybytes/saltybytes-dashboard/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Rate card
	rc := ratecard.Default()
	rateCardPath := filepath.Join(dataDir(), "ratecard.json")
	if err := rc.Load(rateCardPath); err != nil {
		log.Printf("Warning: failed to load rate card: %v (using defaults)", err)
	}

	// Metric cache
	mc := cache.New(database, rc)
	mc.Start()

	// HTTP server
	srv := server.New(mc, rc, rateCardPath, cfg.DashboardPassword)
	handler := srv.Handler()

	log.Printf("Dashboard starting on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func dataDir() string {
	dir := os.Getenv("DATA_DIR")
	if dir == "" {
		dir = "/data"
	}
	os.MkdirAll(dir, 0755)
	return dir
}
