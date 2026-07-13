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
	"github.com/sashabaranov/go-openai"
)

// Server is the HTTP API server.
type Server struct {
	router    *chi.Mux
	userRepo  store.UserRepository
	prefsRepo store.PreferencesRepository
	aiClient  *openai.Client
	aiModel   string
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
func NewServer(cfg *Config, userRepo store.UserRepository, prefsRepo store.PreferencesRepository, aiClient *openai.Client, aiModel string) *Server {
	s := &Server{
		router:    chi.NewRouter(),
		userRepo:  userRepo,
		prefsRepo: prefsRepo,
		aiClient:  aiClient,
		aiModel:   aiModel,
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
	s.router.Get("/oauth/start", s.handleOAuthStart)
	s.router.Get("/oauth/slack", s.handleSlackOAuth)

	// API v1
	s.router.Route("/api/v1", func(r chi.Router) {
		r.Get("/users/{userID}/preferences", s.handleGetPreferences)
		r.Put("/users/{userID}/preferences", s.handleUpdatePreferences)
		r.Get("/user/by-slack", s.handleGetUserBySlack)
		r.Put("/user/by-slack", s.handleUpdateUserBySlack)
		r.Get("/preferences/by-slack", s.handleGetPreferencesBySlack)
		r.Put("/preferences/by-slack", s.handleUpdatePreferencesBySlack)

		// AI chat endpoint — used by the signal-agent Bolt app
		r.Post("/ai/chat", s.handleAIChat)
	})

	// Prometheus metrics (real metrics defined in observability/metrics.go)
	s.router.Handle("/metrics", promhttp.Handler())
}

// Handler returns the HTTP handler for use with http.Server.
func (s *Server) Handler() http.Handler {
	return s.router
}

func (s *Server) handleOAuthStart(w http.ResponseWriter, r *http.Request) {
	scopes := []string{
		"channels:history",
		"channels:read",
		"chat:write",
		"commands",
		"groups:history",
		"groups:read",
		"im:history",
		"im:read",
		"im:write",
		"mpim:history",
		"mpim:read",
		"reactions:read",
		"search:read",
		"team:read",
		"users:read",
		"users:read.email",
		"users.profile:read",
		"users.profile:write",
	}
	// Slack redirects to the frontend's OAuth callback page, which calls the API
	redirectURI := fmt.Sprintf("%s/oauth/callback", s.config.FrontendURL)
	authURL := GetSlackAuthURL(s.config.SlackClientID, redirectURI, scopes)
	http.Redirect(w, r, authURL, http.StatusFound)
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
			"redirect_uri":  {fmt.Sprintf("%s/oauth/callback", s.config.FrontendURL)},
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
		Team      struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"team"`
		AuthedUser struct {
			ID string `json:"id"`
		} `json:"authed_user"`
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
		"team_id", tokenResp.Team.ID,
		"team_name", tokenResp.Team.Name,
	)

	// Create a placeholder user record for the bot in this team so the install
	// is tracked in the database. The actual user records are created on first
	// interaction via ensureUser().
	if tokenResp.BotUserID != "" {
		placeholderUser := &domain.User{
			SlackUserID:  tokenResp.BotUserID,
			SlackTeamID:  tokenResp.Team.ID,
			DisplayName:  fmt.Sprintf("Signal Bot (%s)", tokenResp.Team.Name),
		}
		if err := s.userRepo.Create(r.Context(), placeholderUser); err != nil {
			// User may already exist from a previous install; log but don't fail
			slog.Warn("oauth: could not create placeholder user (may already exist)", "error", err)
		}
	}

	// Redirect to frontend with success and user identifiers
	redirectURL := fmt.Sprintf("%s/app-home?install=success&slack_user_id=%s&team_id=%s",
		s.config.FrontendURL, tokenResp.AuthedUser.ID, tokenResp.Team.ID)
	http.Redirect(w, r, redirectURL, http.StatusFound)
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

func (s *Server) handleGetUserBySlack(w http.ResponseWriter, r *http.Request) {
	slackUserID := r.URL.Query().Get("slack_user_id")
	teamID := r.URL.Query().Get("team_id")
	if slackUserID == "" {
		respondError(w, http.StatusBadRequest, "missing slack_user_id query parameter")
		return
	}

	user, err := s.userRepo.GetBySlackID(r.Context(), slackUserID, teamID)
	if err != nil {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}

	respondJSON(w, http.StatusOK, user)
}

func (s *Server) handleUpdateUserBySlack(w http.ResponseWriter, r *http.Request) {
	slackUserID := r.URL.Query().Get("slack_user_id")
	teamID := r.URL.Query().Get("team_id")
	if slackUserID == "" {
		respondError(w, http.StatusBadRequest, "missing slack_user_id query parameter")
		return
	}

	var update struct {
		Neurotype string `json:"neurotype"`
	}
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := s.userRepo.GetBySlackID(r.Context(), slackUserID, teamID)
	if err != nil {
		user = &domain.User{
			SlackUserID: slackUserID,
			SlackTeamID: teamID,
			Neurotype:   update.Neurotype,
		}
		if err := s.userRepo.Create(r.Context(), user); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to create user")
			return
		}
		respondJSON(w, http.StatusOK, user)
		return
	}

	user.Neurotype = update.Neurotype
	if err := s.userRepo.Update(r.Context(), user); err != nil {
		slog.Error("failed to update user", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to update user")
		return
	}

	respondJSON(w, http.StatusOK, user)
}

func (s *Server) handleGetPreferencesBySlack(w http.ResponseWriter, r *http.Request) {
	slackUserID := r.URL.Query().Get("slack_user_id")
	teamID := r.URL.Query().Get("team_id")
	if slackUserID == "" {
		respondError(w, http.StatusBadRequest, "missing slack_user_id query parameter")
		return
	}

	user, err := s.userRepo.GetBySlackID(r.Context(), slackUserID, teamID)
	if err != nil {
		// User not found; return defaults
		respondJSON(w, http.StatusOK, domain.UserPreferences{
			FocusModeEnabled:  true,
			FocusThreshold:    50,
			TranslatorEnabled: true,
			DigestEnabled:     false,
			DigestHour:        16,
			DeepWorkAutoDetect: false,
			QuietHoursStart:   "22:00",
			QuietHoursEnd:     "08:00",
		})
		return
	}

	prefs, err := s.prefsRepo.Get(r.Context(), user.ID)
	if err != nil {
		respondJSON(w, http.StatusOK, domain.UserPreferences{UserID: user.ID})
		return
	}

	respondJSON(w, http.StatusOK, prefs)
}

func (s *Server) handleUpdatePreferencesBySlack(w http.ResponseWriter, r *http.Request) {
	slackUserID := r.URL.Query().Get("slack_user_id")
	teamID := r.URL.Query().Get("team_id")
	if slackUserID == "" {
		respondError(w, http.StatusBadRequest, "missing slack_user_id query parameter")
		return
	}

	user, err := s.userRepo.GetBySlackID(r.Context(), slackUserID, teamID)
	if err != nil {
		// Create user record if not exists
		user = &domain.User{
			SlackUserID: slackUserID,
			SlackTeamID: teamID,
		}
		if err := s.userRepo.Create(r.Context(), user); err != nil {
			slog.Error("failed to create user", "error", err, "slack_user_id", slackUserID)
			respondError(w, http.StatusInternalServerError, "failed to create user")
			return
		}
	}

	var prefs domain.UserPreferences
	if err := json.NewDecoder(r.Body).Decode(&prefs); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	prefs.UserID = user.ID
	if err := s.prefsRepo.Upsert(r.Context(), &prefs); err != nil {
		slog.Error("failed to update preferences", "error", err, "user", user.ID)
		respondError(w, http.StatusInternalServerError, "failed to update preferences")
		return
	}

	slog.Info("preferences updated via slack", "user", user.ID)
	respondJSON(w, http.StatusOK, prefs)
}

func (s *Server) handleAIChat(w http.ResponseWriter, r *http.Request) {
	if s.aiClient == nil {
		respondError(w, http.StatusServiceUnavailable, "AI client not configured")
		return
	}

	var req struct {
		SystemPrompt string `json:"system_prompt"`
		UserPrompt   string `json:"user_prompt"`
		MaxTokens    int    `json:"max_tokens"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.SystemPrompt == "" {
		respondError(w, http.StatusBadRequest, "system_prompt is required")
		return
	}
	if req.UserPrompt == "" {
		respondError(w, http.StatusBadRequest, "user_prompt is required")
		return
	}
	if req.MaxTokens <= 0 {
		req.MaxTokens = 1000
	}

	resp, err := s.aiClient.CreateChatCompletion(r.Context(), openai.ChatCompletionRequest{
		Model: s.aiModel,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: req.SystemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: req.UserPrompt},
		},
		Temperature: 0.3,
		MaxTokens:   req.MaxTokens,
	})
	if err != nil {
		slog.Error("AI chat failed", "error", err)
		respondError(w, http.StatusInternalServerError, "AI processing failed")
		return
	}

	if len(resp.Choices) == 0 {
		respondError(w, http.StatusInternalServerError, "no response from AI")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"text": resp.Choices[0].Message.Content,
	})
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
