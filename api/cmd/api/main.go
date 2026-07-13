package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/LSUDOKOS/signal/internal/ai"
	"github.com/LSUDOKOS/signal/internal/config"
	"github.com/LSUDOKOS/signal/internal/features"
	"github.com/LSUDOKOS/signal/internal/httpapi"
	mcpclient "github.com/LSUDOKOS/signal/internal/mcp"
	"github.com/LSUDOKOS/signal/internal/observability"
	"github.com/LSUDOKOS/signal/internal/rts"
	signalSlack "github.com/LSUDOKOS/signal/internal/slack"
	"github.com/LSUDOKOS/signal/internal/store/postgres"
	"github.com/LSUDOKOS/signal/internal/store/redis"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}
	observability.SetupLogging(cfg.App.LogLevel)

	// Initialize Prometheus metrics (auto-registers via promauto)
	metrics := observability.NewMetrics()
	_ = metrics // Metrics are registered via promauto; /metrics endpoint serves them

	slog.Info("starting Signal API server",
		"env", cfg.App.Env,
		"version", "1.0.0",
		"ai_model", cfg.OpenAI.Model,
	)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize database
	db, err := postgres.NewDB(ctx, cfg.DB.DSN())
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Run database migrations — check multiple paths for local dev and Railway deployment
	migrationsDirs := []string{"db/migrations", "/app/db/migrations", "../db/migrations"}
	migrationsDir := ""
	for _, d := range migrationsDirs {
		if _, err := os.Stat(d); err == nil {
			migrationsDir = d
			break
		}
	}
	if migrationsDir != "" {
		slog.Info("running database migrations", "dir", migrationsDir)
		if err := postgres.RunMigrations(ctx, db.Pool(), migrationsDir); err != nil {
			slog.Error("database migration failed", "error", err)
			os.Exit(1)
		}
	} else {
		slog.Warn("migrations directory not found in any expected path, skipping")
	}

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db)
	prefsRepo := postgres.NewPreferencesRepository(db)
	channelRepo := postgres.NewChannelRepository(db)
	digestRepo := postgres.NewDigestRepository(db)
	focusSummaryRepo := postgres.NewFocusSummaryRepository(db)
	// translationRepo initialized for future use with translation persistence
	postgres.NewTranslationRepository(db)

	// Initialize Redis cache
	// Render provides REDIS_URL as a full URL; locally we build it from REDIS_ADDR
	redisURL := cfg.Redis.Addr
	if !strings.HasPrefix(redisURL, "redis://") && !strings.HasPrefix(redisURL, "rediss://") {
		redisURL = "redis://" + redisURL
	}
	cache, err := redis.NewCache(ctx, redisURL)
	if err != nil {
		slog.Error("failed to connect to redis", "error", err)
		os.Exit(1)
	}
	defer cache.Close()

	// Initialize AI client
	aiClient := ai.NewClient(cfg.OpenAI.APIKey, cfg.OpenAI.Model, cfg.OpenAI.BaseURL)

	// Start Slack Socket Mode handler (needed before RTS since RTS needs the Slack API client)
	slackHandler, err := signalSlack.NewClient(cfg.Slack.BotToken, cfg.Slack.AppToken, nil)
	if err != nil {
		slog.Error("failed to initialize slack client", "error", err)
		os.Exit(1)
	}

	// Initialize RTS searcher with user token if available (RTS requires user token, not bot token)
	var rtsSearcher *rts.Searcher
	if cfg.Slack.UserToken != "" {
		userSlackClient := signalSlack.NewAPIClientWithToken(cfg.Slack.UserToken)
		rtsSearcher = rts.NewSearcher(userSlackClient)
		// Also set user API on the slack handler for user-scoped calls (status updates)
		slackHandler.SetUserAPI(userSlackClient)
		slog.Info("rts searcher initialized with user token")
	} else {
		rtsSearcher = rts.NewSearcher(slackHandler.GetAPI())
		slog.Warn("rts searcher using bot token (RTS will not work without user token)")
	}

	// Initialize MCP host client
	var mcpHostClient *mcpclient.HostClient
	if cfg.MCP.ServerURL != "" {
		mcpHostClient, err = mcpclient.NewHostClient(cfg.MCP.ServerURL)
		if err != nil {
			slog.Warn("failed to connect to MCP server, continuing without", "error", err)
		}
	}

	// Create feature services with the handler as SlackAPI
	slackAPI := slackHandler
	memory := features.NewMemoryService(cache)
	focusMode := features.NewFocusModeService(slackAPI, aiClient, cache, channelRepo, focusSummaryRepo)
	translator := features.NewTranslatorService(slackAPI, aiClient)
	catchup := features.NewCatchUpService(slackAPI, aiClient, rtsSearcher)
	digest := features.NewDigestService(slackAPI, aiClient, digestRepo, userRepo, prefsRepo, cache)
	deepWork := features.NewDeepWorkService(slackAPI, mcpHostClient, cache)
	modeService := features.NewModeService(slackAPI, userRepo, prefsRepo)
	decisions := features.NewDecisionService(slackAPI, aiClient, rtsSearcher)
	planner := features.NewPlannerService(slackAPI, aiClient, memory)
	threadSummary := features.NewThreadSummaryService(slackAPI, aiClient, memory)
	githubService := features.NewGitHubService(slackAPI, aiClient, memory, cfg.GitHub.Token, cfg.GitHub.Org)
	docsService := features.NewDocsService(slackAPI, aiClient, memory, cfg.Notion.Token)

	if githubService.IsConfigured() {
		slog.Info("github mcp initialized", "org", cfg.GitHub.Org)
	} else {
		slog.Warn("github mcp not configured (set GITHUB_TOKEN + GITHUB_ORG)")
	}
	if docsService.IsConfigured() {
		slog.Info("docs mcp (notion) initialized")
	} else {
		slog.Warn("docs mcp not configured (set NOTION_TOKEN)")
	}

	featureCtrl := features.NewController(
		focusMode, translator, catchup, digest, deepWork,
		modeService, decisions, planner, threadSummary, memory,
		githubService, docsService,
		userRepo, prefsRepo, rtsSearcher, slackAPI, aiClient,
	)

	// Set the feature controller on the slack handler
	slackHandler.SetFeatureCtrl(featureCtrl)

	// Start Slack event handler
	go func() {
		slog.Info("starting slack socket mode handler")
		if err := slackHandler.Start(ctx); err != nil {
			slog.Error("slack handler error", "error", err)
			cancel()
		}
	}()

	// Start HTTP server (pass AI client so Bolt app can call /api/v1/ai/chat)
	httpServer := httpapi.NewServer(&httpapi.Config{
		Port:              fmt.Sprintf("%d", cfg.App.Port),
		SlackClientID:     cfg.Slack.ClientID,
		SlackClientSecret: cfg.Slack.ClientSecret,
		FrontendURL:       cfg.App.FrontendURL,
	}, userRepo, prefsRepo, aiClient.RawClient(), cfg.OpenAI.Model)

	go func() {
		if err := httpServer.Run(ctx); err != nil {
			slog.Error("http server error", "error", err)
			cancel()
		}
	}()

	slog.Info("Signal API server running",
		"http_port", cfg.App.Port,
		"mcp_server", cfg.MCP.ServerURL,
	)

	// Wait for shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	slog.Info("shutting down...")
	cancel()
}
