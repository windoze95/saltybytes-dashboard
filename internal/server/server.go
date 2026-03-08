package server

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/saltybytes/saltybytes-dashboard/internal/cache"
	"github.com/saltybytes/saltybytes-dashboard/internal/ratecard"
)

//go:embed static/*
var staticFiles embed.FS

type Server struct {
	cache        *cache.MetricCache
	rateCard     *ratecard.RateCard
	rateCardPath string
	password     string
	sessions     map[string]time.Time
	sessionMu    sync.RWMutex
}

func New(mc *cache.MetricCache, rc *ratecard.RateCard, rateCardPath, password string) *Server {
	return &Server{
		cache:        mc,
		rateCard:     rc,
		rateCardPath: rateCardPath,
		password:     password,
		sessions:     make(map[string]time.Time),
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	// Public routes
	mux.HandleFunc("/api/auth/login", s.handleLogin)
	mux.HandleFunc("/api/auth/logout", s.handleLogout)
	mux.HandleFunc("/api/auth/check", s.handleCheckAuth)
	mux.HandleFunc("/api/health", s.handleStatus)

	// Protected API routes
	api := http.NewServeMux()
	api.HandleFunc("/api/overview", s.handleOverview)
	api.HandleFunc("/api/users", s.handleUsers)
	api.HandleFunc("/api/recipes", s.handleRecipes)
	api.HandleFunc("/api/canonical", s.handleCanonical)
	api.HandleFunc("/api/search-cache", s.handleSearchCache)
	api.HandleFunc("/api/subscriptions", s.handleSubscriptions)
	api.HandleFunc("/api/allergens", s.handleAllergens)
	api.HandleFunc("/api/families", s.handleFamilies)
	api.HandleFunc("/api/infrastructure", s.handleInfrastructure)
	api.HandleFunc("/api/health-checks", s.handleHealthChecks)
	api.HandleFunc("/api/cost-center", s.handleCostCenter)
	api.HandleFunc("/api/rate-card", s.handleGetRateCard)
	api.HandleFunc("/api/rate-card/update", s.handleUpdateRateCard)
	api.HandleFunc("/api/refresh", s.handleRefresh)

	mux.Handle("/api/", s.authMiddleware(api))

	// Serve embedded React frontend
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatal("Failed to create sub filesystem:", err)
	}
	fileServer := http.FileServer(http.FS(staticFS))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Try to serve static file first
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}

		// Check if file exists in embedded FS
		f, err := staticFS.Open(path[1:]) // strip leading /
		if err != nil {
			// SPA fallback: serve index.html for client-side routing
			r.URL.Path = "/"
			fileServer.ServeHTTP(w, r)
			return
		}
		f.Close()

		fileServer.ServeHTTP(w, r)
	})

	return mux
}
