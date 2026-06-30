package server

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/saltybytes/saltybytes-dashboard/internal/apiclient"
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
	api          *apiclient.Client // nil-safe; live model switching when Enabled()
	sessions     map[string]time.Time
	sessionMu    sync.RWMutex
}

func New(mc *cache.MetricCache, rc *ratecard.RateCard, rateCardPath, password string, api *apiclient.Client) *Server {
	return &Server{
		cache:        mc,
		rateCard:     rc,
		rateCardPath: rateCardPath,
		password:     password,
		api:          api,
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
	api.HandleFunc("/api/ai-models", s.handleAIModels)
	api.HandleFunc("/api/ai-ops", s.handleAIOps)
	api.HandleFunc("/api/cache-economics", s.handleCacheEconomics)
	api.HandleFunc("/api/video-economics", s.handleVideoEconomics)
	api.HandleFunc("/api/growth", s.handleGrowth)
	api.HandleFunc("/api/recipe-quality", s.handleRecipeQuality)
	// Light-tier model registry + live switch (registry read from DB; mutations
	// proxied to the API admin endpoints).
	api.HandleFunc("/api/ai-registry", s.handleAIRegistry)
	api.HandleFunc("/api/ai-models/add", s.handleAIModelAdd)
	api.HandleFunc("/api/ai-models/update", s.handleAIModelUpdate)
	api.HandleFunc("/api/ai-models/delete", s.handleAIModelDelete)
	api.HandleFunc("/api/ai-models/activate", s.handleAIModelActivate)
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
