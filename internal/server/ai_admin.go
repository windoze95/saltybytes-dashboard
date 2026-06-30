package server

import (
	"encoding/json"
	"io"
	"net/http"
)

// handleAIRegistry returns the light-tier model registry + active selection
// (read-only mirror of the API tables) and whether live management is wired up.
func (s *Server) handleAIRegistry(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]interface{}{
		"registry":           s.cache.GetAIRegistry(),
		"management_enabled": s.api != nil && s.api.Enabled(),
	})
}

// requireManagement guards the mutating routes; returns false (and writes a 503)
// when the dashboard isn't configured to call the API admin endpoints.
func (s *Server) requireManagement(w http.ResponseWriter) bool {
	if s.api == nil || !s.api.Enabled() {
		writeError(w, http.StatusServiceUnavailable,
			"live model management is not configured (set API_BASE_URL + API_ID_HEADER + ADMIN_TOKEN on the dashboard)")
		return false
	}
	return true
}

// relay forwards the API's status + body to the browser so probe/validation
// errors (400 + {"error":...}) surface verbatim; on a successful mutation it
// kicks a cache refresh so the registry view updates promptly.
func (s *Server) relay(w http.ResponseWriter, status int, body []byte, err error) {
	if err != nil {
		writeError(w, http.StatusBadGateway, "API request failed: "+err.Error())
		return
	}
	if status >= 200 && status < 300 {
		go s.cache.Refresh()
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(body)
}

func (s *Server) handleAIModelAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !s.requireManagement(w) {
		return
	}
	body, _ := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	status, resp, err := s.api.CreateModel(r.Context(), body)
	s.relay(w, status, resp, err)
}

func (s *Server) handleAIModelUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !s.requireManagement(w) {
		return
	}
	body, _ := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	var idOnly struct {
		ID uint `json:"id"`
	}
	_ = json.Unmarshal(body, &idOnly)
	if idOnly.ID == 0 {
		writeError(w, http.StatusBadRequest, "id is required")
		return
	}
	status, resp, err := s.api.UpdateModel(r.Context(), idOnly.ID, body)
	s.relay(w, status, resp, err)
}

func (s *Server) handleAIModelDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !s.requireManagement(w) {
		return
	}
	var req struct {
		ID uint `json:"id"`
	}
	if err := decodeJSON(r, &req); err != nil || req.ID == 0 {
		writeError(w, http.StatusBadRequest, "id is required")
		return
	}
	status, resp, err := s.api.DeleteModel(r.Context(), req.ID)
	s.relay(w, status, resp, err)
}

func (s *Server) handleAIModelActivate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !s.requireManagement(w) {
		return
	}
	body, _ := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	var req struct {
		ID uint `json:"id"`
	}
	_ = json.Unmarshal(body, &req)
	if req.ID == 0 {
		writeError(w, http.StatusBadRequest, "id is required")
		return
	}
	// Forward {id} verbatim; the API probes the model and only switches on green.
	status, resp, err := s.api.SetActive(r.Context(), body)
	s.relay(w, status, resp, err)
}

// writeError writes a JSON error with a status code.
func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
