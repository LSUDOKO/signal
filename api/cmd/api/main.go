package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
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

	// Run database migrations
	migrationsDir := "db/migrations"
	if _, err := os.Stat(migrationsDir); err == nil {
		slog.Info("running database migrations", "dir", migrationsDir)
		if err := postgres.RunMigrations(ctx, db.Pool(), migrationsDir); err != nil {
			slog.Error("database migration failed", "error", err)
			os.Exit(1)
		}
	} else {
		slog.Warn("migrations directory not found, skipping", "dir", migrationsDir)
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
	cache, err := redis.NewCache(ctx, fmt.Sprintf("redis://%s", cfg.Redis.Addr))
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

	// Initialize RTS searcher with real Slack API client
	rtsSearcher := rts.NewSearcher(slackHandler.GetAPI())

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
	focusMode := features.NewFocusModeService(slackAPI, aiClient, cache, channelRepo, focusSummaryRepo)
	translator := features.NewTranslatorService(slackAPI, aiClient)
	catchup := features.NewCatchUpService(slackAPI, aiClient, rtsSearcher)
	digest := features.NewDigestService(slackAPI, aiClient, digestRepo, userRepo, prefsRepo, cache)
	deepWork := features.NewDeepWorkService(slackAPI, mcpHostClient, cache)

	featureCtrl := features.NewController(
		focusMode, translator, catchup, digest, deepWork,
		userRepo, prefsRepo, rtsSearcher, slackAPI,
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

	// Start HTTP server
	httpServer := httpapi.NewServer(&httpapi.Config{
		Port:              fmt.Sprintf("%d", cfg.App.Port),
		SlackClientID:     cfg.Slack.ClientID,
		SlackClientSecret: cfg.Slack.ClientSecret,
		FrontendURL:       cfg.App.FrontendURL,
	}, userRepo, prefsRepo)

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
