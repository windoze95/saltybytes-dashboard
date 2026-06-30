package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/saltybytes/saltybytes-dashboard/internal/apiclient"
	"github.com/saltybytes/saltybytes-dashboard/internal/cache"
	"github.com/saltybytes/saltybytes-dashboard/internal/config"
	"github.com/saltybytes/saltybytes-dashboard/internal/db"
	"github.com/saltybytes/saltybytes-dashboard/internal/ratecard"
	"github.com/saltybytes/saltybytes-dashboard/internal/server"
	"github.com/saltybytes/saltybytes-dashboard/internal/sgsync"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Keep this host's (dynamic) IP whitelisted on the RDS security group BEFORE
	// connecting — a changed ISP IP would otherwise crash startup before the app
	// could heal itself. No-op unless SGSYNC_* env vars are set.
	sgsync.Run(context.Background(), sgsync.FromEnv())

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

	// Outbound API client for live model management (no-op unless configured).
	apiClient := apiclient.New(cfg.APIBaseURL, cfg.APIIDHeader, cfg.AdminToken)
	if apiClient.Enabled() {
		log.Printf("Live AI-model management enabled (API: %s)", cfg.APIBaseURL)
	} else {
		log.Printf("Live AI-model management disabled (set API_BASE_URL + API_ID_HEADER + ADMIN_TOKEN to enable); registry shown read-only")
	}

	// HTTP server
	srv := server.New(mc, rc, rateCardPath, cfg.DashboardPassword, apiClient)
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
