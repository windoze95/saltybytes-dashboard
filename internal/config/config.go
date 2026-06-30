package config

import (
	"fmt"
	"os"
)

type Config struct {
	DatabaseURL       string
	Port              string
	DashboardPassword string

	// Optional: enable live AI-model switching by letting the dashboard call the
	// API's admin endpoints. All three must be set; otherwise the registry is
	// shown read-only and switching is disabled. The admin token + ID header are
	// never sent to the browser — the dashboard backend proxies the calls.
	APIBaseURL  string // API base URL, e.g. https://api.saltybytes.ai
	APIIDHeader string // value for the X-SaltyBytes-Identifier header (API's ID_HEADER)
	AdminToken  string // value for the X-Admin-Token header (API's ADMIN_TOKEN)
}

func Load() (*Config, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}

	password := os.Getenv("DASHBOARD_PASSWORD")
	if password == "" {
		return nil, fmt.Errorf("DASHBOARD_PASSWORD is required")
	}

	return &Config{
		DatabaseURL:       dbURL,
		Port:              port,
		DashboardPassword: password,
		APIBaseURL:        os.Getenv("API_BASE_URL"),
		APIIDHeader:       os.Getenv("API_ID_HEADER"),
		AdminToken:        os.Getenv("ADMIN_TOKEN"),
	}, nil
}
