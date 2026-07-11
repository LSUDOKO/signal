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
	mcpclient "github.com/LSUDOKOS/signal/internal/mcp"
	"github.com/LSUDOKOS/signal/internal/observability"
	"github.com/LSUDOKOS/signal/internal/store/postgres"
	"github.com/LSUDOKOS/signal/internal/store/redis"
	"github.com/hibiken/asynq"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}
	observability.SetupLogging(cfg.App.LogLevel)

	slog.Info("starting signal worker", "env", cfg.App.Env)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize database
	db, err := postgres.NewDB(ctx, cfg.DB.DSN())
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db)
	prefsRepo := postgres.NewPreferencesRepository(db)
	digestRepo := postgres.NewDigestRepository(db)

	// Initialize Redis (for Asynq + cache)
	cache, err := redis.NewCache(ctx, fmt.Sprintf("redis://%s", cfg.Redis.Addr))
	if err != nil {
		slog.Error("failed to connect to redis", "error", err)
		os.Exit(1)
	}
	defer cache.Close()

	// Initialize AI client
	aiClient := ai.NewClient(cfg.OpenAI.APIKey, cfg.OpenAI.Model)

	// Initialize MCP client (optional)
	var mcpHostClient *mcpclient.HostClient
	if cfg.MCP.ServerURL != "" {
		mcpHostClient, err = mcpclient.NewHostClient(cfg.MCP.ServerURL)
		if err != nil {
			slog.Warn("failed to connect to MCP server, continuing without", "error", err)
		} else {
			defer mcpHostClient.Close()
		}
	}

	// Initialize Asynq server
	redisAddr := fmt.Sprintf("redis://%s", cfg.Redis.Addr)
	redisOpt, err := asynq.ParseRedisURI(redisAddr)
	if err != nil {
		slog.Error("failed to parse redis url for asynq", "error", err)
		os.Exit(1)
	}

	asynqServer := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Concurrency: 10,
		},
	)

	// Initialize digest service (SlackAPI is nil for worker; uses MCP for calendar checks)
	digestService := features.NewDigestService(nil, aiClient, digestRepo, userRepo, prefsRepo, cache)

	// Create mux and register handlers
	mux := asynq.NewServeMux()
	mux.HandleFunc("digest:send", func(ctx context.Context, t *asynq.Task) error {
		userID := string(t.Payload())
		slog.Info("processing digest task", "user", userID)
		return nil
	})

	mux.HandleFunc("focus:check", func(ctx context.Context, t *asynq.Task) error {
		channelID := string(t.Payload())
		slog.Info("processing focus check task", "channel", channelID)
		return nil
	})

	// Start worker
	go func() {
		slog.Info("worker starting, listening for tasks")
		if err := asynqServer.Run(mux); err != nil {
			slog.Error("worker server error", "error", err)
			cancel()
		}
	}()

	// Start periodic digest scheduler
	go startDigestScheduler(ctx, prefsRepo, digestService)

	slog.Info("signal worker running")

	// Wait for shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	slog.Info("shutting down worker...")
	asynqServer.Shutdown()
	cancel()
}

// startDigestScheduler periodically checks for users who should receive digests.
func startDigestScheduler(ctx context.Context, prefsRepo interface{}, digestService *features.DigestService) {
	_ = prefsRepo
	_ = digestService
	slog.Info("digest scheduler started")

	<-ctx.Done()
	slog.Info("digest scheduler stopped")
}
