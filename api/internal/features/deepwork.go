package features

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/LSUDOKOS/signal/internal/domain"
	mcpclient "github.com/LSUDOKOS/signal/internal/mcp"
	"github.com/LSUDOKOS/signal/internal/store"
	"github.com/slack-go/slack"
)

// DeepWorkService implements the Deep Work Protector feature.
type DeepWorkService struct {
	slack     SlackAPI
	mcpClient *mcpclient.HostClient
	cache     store.CacheClient
}

// NewDeepWorkService creates a new Deep Work service.
func NewDeepWorkService(slack SlackAPI, mcpClient *mcpclient.HostClient, cache store.CacheClient) *DeepWorkService {
	return &DeepWorkService{
		slack:     slack,
		mcpClient: mcpClient,
		cache:     cache,
	}
}

// HandleSlashCommand processes the /focus command via response_url.
func (d *DeepWorkService) HandleSlashCommand(ctx context.Context, cmd *slack.SlashCommand, user *domain.User, responseURL string) error {
	text := strings.TrimSpace(cmd.Text)

	duration := 120 // default 2 hours
	if text != "" {
		if parsed, err := parseDuration(text); err == nil {
			duration = parsed
		}
	}

	return d.startDeepWork(ctx, cmd.UserID, cmd.ChannelID, duration, responseURL)
}

// HandleBlockAction handles Deep Work button clicks.
func (d *DeepWorkService) HandleBlockAction(ctx context.Context, action *slack.InteractionCallback, user *domain.User) error {
	if len(action.ActionCallback.BlockActions) == 0 {
		return nil
	}
	actionID := action.ActionCallback.BlockActions[0].ActionID

	switch actionID {
	case "deepwork_start":
		duration := 120
		if len(action.ActionCallback.BlockActions) > 0 {
			if val := action.ActionCallback.BlockActions[0].Value; val != "" {
				if parsed, err := strconv.Atoi(val); err == nil {
					duration = parsed
				}
			}
		}
		return d.startDeepWork(ctx, user.SlackUserID, action.Channel.ID, duration, "")

	case "deepwork_stop":
		return d.stopDeepWork(ctx, user.SlackUserID, action.Channel.ID)

	case "deepwork_extend":
		return d.extendDeepWork(ctx, user.SlackUserID, action.Channel.ID)
	}

	return nil
}

// startDeepWork initiates a deep work session.
func (d *DeepWorkService) startDeepWork(ctx context.Context, slackUserID, channelID string, durationMinutes int, responseURL string) error {
	duration := time.Duration(durationMinutes) * time.Minute
	endTime := time.Now().Add(duration)

	if err := d.cache.SetDeepWork(ctx, slackUserID, duration); err != nil {
		slog.Error("failed to set deep work state", "error", err)
	}

	if d.mcpClient != nil {
		if _, err := d.mcpClient.BlockFocusTime(ctx, slackUserID, durationMinutes, "Deep Work"); err != nil {
			slog.Warn("mcp block focus time failed", "error", err)
		}
	}

	statusText := fmt.Sprintf("In Deep Work — back at %s", endTime.Format("3:04 PM"))
	if err := d.slack.SetUserStatus(slackUserID, statusText, ":brain:", durationMinutes); err != nil {
		slog.Warn("failed to set slack status", "error", err, "user", slackUserID)
	}

	blocks := []slack.Block{
		slack.NewHeaderBlock(
			slack.NewTextBlockObject("plain_text", "🧘 Deep Work Mode Activated", true, false),
		),
		slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn",
				fmt.Sprintf("*Duration:* %d minutes\n*Ends at:* %s\n\nYour Slack status has been updated. Non-urgent notifications are paused.", durationMinutes, endTime.Format("3:04 PM")),
				false, false,
			),
			nil, nil,
		),
		slack.NewActionBlock("deepwork_actions",
			slack.NewButtonBlockElement("deepwork_extend", "60",
				slack.NewTextBlockObject("plain_text", "Extend 1h", false, false),
			).WithStyle(slack.StylePrimary),
			slack.NewButtonBlockElement("deepwork_stop", "stop",
				slack.NewTextBlockObject("plain_text", "End Early", false, false),
			).WithStyle(slack.StyleDanger),
		),
	}

	// Prefer response_url (slash command context), fall back to posting in the channel
	if responseURL != "" {
		err := d.slack.PostWebhook(responseURL, blocks, "Deep Work Activated")
		if err == nil {
			return nil
		}
		slog.Warn("PostWebhook failed for focus, falling back to channel post", "error", err)
	}
	// Fall back: post directly to the channel where command was typed
	if channelID != "" {
		return d.slack.PostMessage(channelID, blocks, "Deep Work Activated")
	}
	return nil // Best-effort — don't fail the whole command
}

// stopDeepWork ends a deep work session.
func (d *DeepWorkService) stopDeepWork(ctx context.Context, slackUserID, channelID string) error {
	if err := d.cache.ClearDeepWork(ctx, slackUserID); err != nil {
		slog.Error("failed to clear deep work", "error", err)
	}
	if err := d.slack.SetUserStatus(slackUserID, "", "", 0); err != nil {
		slog.Warn("failed to clear slack status", "error", err)
	}
	if channelID == "" {
		return nil
	}
	return d.slack.PostMessage(channelID, []slack.Block{
		slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn", "✅ Deep Work mode ended. Notifications are back to normal.", false, false),
			nil, nil,
		),
	}, "Deep Work Ended")
}

// extendDeepWork extends the current deep work session.
func (d *DeepWorkService) extendDeepWork(ctx context.Context, slackUserID, channelID string) error {
	currentDuration, err := d.cache.GetDeepWork(ctx, slackUserID)
	if err != nil || currentDuration == 0 {
		return d.startDeepWork(ctx, slackUserID, channelID, 60, "")
	}
	newDuration := int(currentDuration.Minutes()) + 60
	return d.startDeepWork(ctx, slackUserID, channelID, newDuration, "")
}

// parseDuration parses a duration string like "2h", "90min", "60" into minutes.
func parseDuration(input string) (int, error) {
	input = strings.TrimSpace(strings.ToLower(input))

	if strings.HasSuffix(input, "h") {
		hours, err := strconv.Atoi(strings.TrimSuffix(input, "h"))
		if err != nil {
			return 0, err
		}
		return hours * 60, nil
	}

	if strings.HasSuffix(input, "min") {
		mins, err := strconv.Atoi(strings.TrimSuffix(input, "min"))
		if err != nil {
			return 0, err
		}
		return mins, nil
	}

	// Default: interpret as minutes
	return strconv.Atoi(input)
}

// IsInDeepWork checks if a user is currently in deep work mode.
func (d *DeepWorkService) IsInDeepWork(ctx context.Context, slackUserID string) bool {
	duration, err := d.cache.GetDeepWork(ctx, slackUserID)
	if err != nil {
		return false
	}
	return duration > 0
}
