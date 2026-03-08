package server

import (
	"crypto/subtle"
	"net/http"
	"time"
)

const sessionCookieName = "dashboard_session"
const sessionDuration = 24 * time.Hour

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookieName)
		if err != nil || !s.isValidSession(cookie.Value) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) isValidSession(token string) bool {
	s.sessionMu.RLock()
	defer s.sessionMu.RUnlock()
	expiry, ok := s.sessions[token]
	if !ok {
		return false
	}
	return time.Now().Before(expiry)
}

func (s *Server) createSession() string {
	token := generateToken()
	s.sessionMu.Lock()
	s.sessions[token] = time.Now().Add(sessionDuration)
	s.sessionMu.Unlock()
	return token
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Password string `json:"password"`
	}
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if subtle.ConstantTimeCompare([]byte(req.Password), []byte(s.password)) != 1 {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	token := s.createSession()
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(sessionDuration.Seconds()),
	})

	writeJSON(w, map[string]string{"status": "ok"})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookieName)
	if err == nil {
		s.sessionMu.Lock()
		delete(s.sessions, cookie.Value)
		s.sessionMu.Unlock()
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	writeJSON(w, map[string]string{"status": "ok"})
}

func (s *Server) handleCheckAuth(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || !s.isValidSession(cookie.Value) {
		writeJSON(w, map[string]bool{"authenticated": false})
		return
	}
	writeJSON(w, map[string]bool{"authenticated": true})
}
