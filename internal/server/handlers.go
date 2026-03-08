package server

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/saltybytes/saltybytes-dashboard/internal/ratecard"
)

func (s *Server) handleOverview(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, s.cache.GetOverview())
}

func (s *Server) handleUsers(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, s.cache.GetUsers())
}

func (s *Server) handleRecipes(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, s.cache.GetRecipes())
}

func (s *Server) handleCanonical(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, s.cache.GetCanonical())
}

func (s *Server) handleSearchCache(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, s.cache.GetSearchCache())
}

func (s *Server) handleSubscriptions(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, s.cache.GetSubscriptions())
}

func (s *Server) handleAllergens(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, s.cache.GetAllergens())
}

func (s *Server) handleFamilies(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, s.cache.GetFamilies())
}

func (s *Server) handleInfrastructure(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, s.cache.GetInfrastructure())
}

func (s *Server) handleHealthChecks(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, s.cache.GetHealthChecks())
}

func (s *Server) handleCostCenter(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, s.cache.GetCostCenter())
}

func (s *Server) handleGetRateCard(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, s.rateCard.Get())
}

func (s *Server) handleUpdateRateCard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var updated ratecard.RateCard
	if err := decodeJSON(r, &updated); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	s.rateCard.Update(updated)
	if err := s.rateCard.Save(s.rateCardPath); err != nil {
		http.Error(w, "Failed to save rate card", http.StatusInternalServerError)
		return
	}

	// Trigger cache refresh to recalculate costs
	go s.cache.Refresh()

	writeJSON(w, s.rateCard.Get())
}

func (s *Server) handleRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	go s.cache.Refresh()
	writeJSON(w, map[string]string{"status": "refreshing"})
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]interface{}{
		"status":       "ok",
		"last_refresh": s.cache.GetLastRefresh(),
	})
}

// Helpers

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func decodeJSON(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}
