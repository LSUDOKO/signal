package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/LSUDOKOS/signal/internal/domain"
	"github.com/LSUDOKOS/signal/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server is the HTTP API server.
type Server struct {
	router    *chi.Mux
	userRepo  store.UserRepository
	prefsRepo store.PreferencesRepository
	config    *Config
}

// Config holds HTTP API configuration.
type Config struct {
	Port            string
	SlackClientID   string
	SlackClientSecret string
	FrontendURL     string
}

// NewServer creates a new HTTP API server.
func NewServer(cfg *Config, userRepo store.UserRepository, prefsRepo store.PreferencesRepository) *Server {
	s := &Server{
		router:    chi.NewRouter(),
		userRepo:  userRepo,
		prefsRepo: prefsRepo,
		config:    cfg,
	}
	s.registerRoutes()
	return s
}

func (s *Server) registerRoutes() {
	// Middleware
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.RequestID)
	s.router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{s.config.FrontendURL, "https://slack.com"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health
	s.router.Get("/health", s.handleHealth)

	// OAuth
	s.router.Get("/oauth/slack", s.handleSlackOAuth)

	// API v1
	s.router.Route("/api/v1", func(r chi.Router) {
		r.Get("/users/{userID}/preferences", s.handleGetPreferences)
		r.Put("/users/{userID}/preferences", s.handleUpdatePreferences)
	})

	// Prometheus metrics (real metrics defined in observability/metrics.go)
	s.router.Handle("/metrics", promhttp.Handler())
}

// Handler returns the HTTP handler for use with http.Server.
func (s *Server) Handler() http.Handler {
	return s.router
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "ok",
		"version": "1.0.0",
	})
}

func (s *Server) handleSlackOAuth(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing code parameter", http.StatusBadRequest)
		return
	}

	slog.Info("oauth callback received", "code_prefix", code[:min(10, len(code))])

	// Exchange authorization code for access token via Slack OAuth API
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.PostForm("https://slack.com/api/oauth.v2.access",
		url.Values{
			"client_id":     {s.config.SlackClientID},
			"client_secret": {s.config.SlackClientSecret},
			"code":          {code},
			"redirect_uri":  {fmt.Sprintf("%s/oauth/slack", s.config.FrontendURL)},
		},
	)
	if err != nil {
		slog.Error("oauth token exchange failed", "error", err)
		http.Redirect(w, r, fmt.Sprintf("%s/app-home?install=error&reason=token_exchange_failed", s.config.FrontendURL), http.StatusFound)
		return
	}
	defer resp.Body.Close()

	var tokenResp struct {
		OK        bool   `json:"ok"`
		Error     string `json:"error,omitempty"`
		BotToken  string `json:"access_token"`
		BotUserID string `json:"bot_user_id"`
		TeamName  string `json:"team_name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		slog.Error("failed to decode token response", "error", err)
		http.Redirect(w, r, fmt.Sprintf("%s/app-home?install=error&reason=parse_failed", s.config.FrontendURL), http.StatusFound)
		return
	}

	if !tokenResp.OK {
		slog.Error("oauth token exchange denied", "error", tokenResp.Error)
		http.Redirect(w, r, fmt.Sprintf("%s/app-home?install=error&reason=%s", s.config.FrontendURL, tokenResp.Error), http.StatusFound)
		return
	}

	slog.Info("oauth successful",
		"bot_user", tokenResp.BotUserID,
		"team", tokenResp.TeamName,
	)

	// Store session token in database (future: create user record)
	_ = tokenResp

	// Redirect to frontend with success
	http.Redirect(w, r, fmt.Sprintf("%s/app-home?install=success", s.config.FrontendURL), http.StatusFound)
}

func (s *Server) handleGetPreferences(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "userID")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	prefs, err := s.prefsRepo.Get(r.Context(), userID)
	if err != nil {
		// Return default preferences
		prefs = &domain.UserPreferences{
			UserID:            userID,
			FocusModeEnabled:  true,
			FocusThreshold:    50,
			TranslatorEnabled: true,
			DigestEnabled:     false,
			DigestHour:        16,
			DeepWorkAutoDetect: false,
			QuietHoursStart:   "22:00",
			QuietHoursEnd:     "08:00",
		}
	}

	respondJSON(w, http.StatusOK, prefs)
}

func (s *Server) handleUpdatePreferences(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "userID")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	var prefs domain.UserPreferences
	if err := json.NewDecoder(r.Body).Decode(&prefs); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	prefs.UserID = userID
	if err := s.prefsRepo.Upsert(r.Context(), &prefs); err != nil {
		slog.Error("failed to update preferences", "error", err, "user", userID)
		respondError(w, http.StatusInternalServerError, "failed to update preferences")
		return
	}

	slog.Info("preferences updated", "user", userID)
	respondJSON(w, http.StatusOK, prefs)
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Run starts the HTTP server.
func (s *Server) Run(ctx context.Context) error {
	addr := fmt.Sprintf(":%s", s.config.Port)
	slog.Info("starting http server", "addr", addr)

	httpServer := &http.Server{
		Addr:    addr,
		Handler: s.Handler(),
	}

	go func() {
		<-ctx.Done()
		slog.Info("shutting down http server")
		httpServer.Shutdown(context.Background())
	}()

	return httpServer.ListenAndServe()
}

// GetSlackAuthURL returns the Slack OAuth authorization URL.
func GetSlackAuthURL(clientID string, redirectURI string, scopes []string) string {
	return fmt.Sprintf(
		"https://slack.com/oauth/v2/authorize?client_id=%s&scope=%s&redirect_uri=%s",
		clientID,
		strings.Join(scopes, " "),
		redirectURI,
	)
}
