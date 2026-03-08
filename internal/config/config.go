package config

import (
	"fmt"
	"os"
)

type Config struct {
	DatabaseURL       string
	Port              string
	DashboardPassword string
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
	}, nil
}
