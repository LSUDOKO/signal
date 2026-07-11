package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/LSUDOKOS/signal/internal/ai"
	"github.com/LSUDOKOS/signal/internal/config"
	"github.com/LSUDOKOS/signal/internal/features"
	mcpclient "github.com/LSUDOKOS/signal/internal/mcp"
	"github.com/LSUDOKOS/signal/internal/observability"
	"github.com/LSUDOKOS/signal/internal/store"
	"github.com/LSUDOKOS/signal/internal/store/postgres"
	"github.com/LSUDOKOS/signal/internal/store/redis"
	"github.com/hibiken/asynq"
	"github.com/slack-go/slack"
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

	// Run database migrations
	if err := postgres.RunMigrations(ctx, db.Pool(), "db/migrations"); err != nil {
		slog.Error("database migration failed", "error", err)
		os.Exit(1)
	}

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
	aiClient := ai.NewClient(cfg.OpenAI.APIKey, cfg.OpenAI.Model, cfg.OpenAI.BaseURL)

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

	// Initialize Slack API client for sending digests
	slackAPI := newWorkerSlackAPI(cfg.Slack.BotToken)
	digestService := features.NewDigestService(slackAPI, aiClient, digestRepo, userRepo, prefsRepo, cache)

	// Create mux and register handlers
	mux := asynq.NewServeMux()
	mux.HandleFunc("digest:send", func(ctx context.Context, t *asynq.Task) error {
		userID := string(t.Payload())
		slog.Info("processing digest task", "user", userID)

		// In production, this would look up the user and call SendScheduledDigest
		// For now, log that the digest was queued
		slog.Info("digest sent", "user", userID)
		return nil
	})

	mux.HandleFunc("focus:check", func(ctx context.Context, t *asynq.Task) error {
		channelID := string(t.Payload())
		if channelID == "" {
			slog.Warn("focus check task: empty channel ID")
			return nil
		}

		// Check current channel velocity from Redis
		count, err := cache.GetChannelVelocity(ctx, channelID)
		if err != nil {
			slog.Warn("focus check: failed to get velocity", "channel", channelID, "error", err)
			return nil
		}

		offered, err := cache.HasFocusBeenOffered(ctx, channelID)
		if err != nil {
			slog.Warn("focus check: failed to check offered flag", "channel", channelID, "error", err)
		}

		slog.Info("focus check completed",
			"channel", channelID,
			"velocity", count,
			"focus_offered", offered,
		)

		if count >= 50 && !offered {
			slog.Info("channel velocity threshold met, awaiting next message to trigger focus mode",
				"channel", channelID, "count", count)
		}
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
	go startDigestScheduler(ctx, prefsRepo, userRepo, digestService)

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
func startDigestScheduler(ctx context.Context, prefsRepo *postgres.PreferencesRepo, userRepo store.UserRepository, digestService *features.DigestService) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	slog.Info("digest scheduler started, checking every 5 minutes")

	for {
		select {
		case <-ctx.Done():
			slog.Info("digest scheduler stopped")
			return
		case <-ticker.C:
			currentHour := time.Now().Hour()
			slog.Debug("digest scheduler checking users", "hour", currentHour)

			prefs, err := prefsRepo.GetByDigestHour(ctx, currentHour)
			if err != nil {
				slog.Error("digest scheduler: failed to get users", "error", err)
				continue
			}

			for _, pref := range prefs {
				user, err := userRepo.GetByID(ctx, pref.UserID)
				if err != nil {
					slog.Error("digest scheduler: failed to get user", "error", err, "user_id", pref.UserID)
					continue
				}

				if err := digestService.SendScheduledDigest(ctx, *user, &pref); err != nil {
					slog.Error("digest scheduler: failed to send digest", "error", err, "user", user.SlackUserID)
				} else {
					slog.Info("digest scheduler: digest sent", "user", user.SlackUserID)
				}
			}
		}
	}
}

// workerSlackAPI is a minimal SlackAPI adapter for the worker process.
// It wraps slack.Client to provide the SlackAPI interface without Socket Mode.
type workerSlackAPI struct {
	api *slack.Client
}

func newWorkerSlackAPI(botToken string) *workerSlackAPI {
	return &workerSlackAPI{
		api: slack.New(botToken),
	}
}

func (w *workerSlackAPI) PostMessage(channelID string, blocks []slack.Block, text string) error {
	_, _, err := w.api.PostMessage(channelID, slack.MsgOptionBlocks(blocks...), slack.MsgOptionText(text, false))
	return err
}

func (w *workerSlackAPI) PostEphemeral(channelID, userID string, blocks []slack.Block, text string) error {
	_, err := w.api.PostEphemeral(channelID, userID, slack.MsgOptionBlocks(blocks...), slack.MsgOptionText(text, false))
	return err
}

func (w *workerSlackAPI) OpenDMChannel(userID string) (string, error) {
	ch, _, _, err := w.api.OpenConversation(&slack.OpenConversationParameters{Users: []string{userID}})
	if err != nil {
		return "", err
	}
	return ch.ID, nil
}

func (w *workerSlackAPI) GetUser(userID string) (*slack.User, error) {
	return w.api.GetUserInfo(userID)
}

func (w *workerSlackAPI) GetChannelHistory(channelID string, limit int) ([]slack.Message, error) {
	resp, err := w.api.GetConversationHistory(&slack.GetConversationHistoryParameters{
		ChannelID: channelID,
		Limit:     limit,
	})
	if err != nil {
		return nil, err
	}
	return resp.Messages, nil
}

func (w *workerSlackAPI) SearchMessages(query string, params slack.SearchParameters) (*slack.SearchMessages, error) {
	result, err := w.api.SearchMessages(query, params)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *workerSlackAPI) SetUserStatus(userID, statusText, statusEmoji string, expiration int) error {
	// worker doesn't need user status setting; no-op
	return nil
}

func (w *workerSlackAPI) PublishView(userID string, blocks []slack.Block) error {
	_, err := w.api.PublishView(userID, slack.HomeTabViewRequest{
		Type:   "home",
		Blocks: slack.Blocks{BlockSet: blocks},
	}, "")
	return err
}
