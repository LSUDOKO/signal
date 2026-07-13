package features

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/LSUDOKOS/signal/internal/ai"
	"github.com/LSUDOKOS/signal/internal/domain"
	"github.com/LSUDOKOS/signal/internal/rts"
	"github.com/slack-go/slack"
)

// DecisionService handles the /decisions command.
// It searches a channel's recent history for decisions made and surfaces them.
type DecisionService struct {
	slack    SlackAPI
	ai       *ai.Client
	searcher *rts.Searcher
}

// NewDecisionService creates a new DecisionService.
func NewDecisionService(slack SlackAPI, ai *ai.Client, searcher *rts.Searcher) *DecisionService {
	return &DecisionService{slack: slack, ai: ai, searcher: searcher}
}

// HandleSlashCommand handles /decisions [channel] [days].
// Usage: /decisions #engineering 7
func (d *DecisionService) HandleSlashCommand(ctx context.Context, cmd *slack.SlashCommand, user *domain.User, responseURL string) error {
	args := strings.Fields(strings.TrimSpace(cmd.Text))

	// Parse optional args: channel and days
	channelID := cmd.ChannelID // default: current channel
	daysBack := 7

	for _, arg := range args {
		if strings.HasPrefix(arg, "<#") {
			// Strip Slack channel mention format: <#C123|name>
			arg = strings.TrimPrefix(arg, "<#")
			arg = strings.Split(arg, "|")[0]
			arg = strings.TrimSuffix(arg, ">")
			channelID = arg
		} else if arg == "7" || arg == "14" || arg == "30" {
			fmt.Sscanf(arg, "%d", &daysBack)
		}
	}

	// Fetch channel history
	messages, err := d.slack.GetChannelHistory(channelID, 100)
	if err != nil {
		slog.Error("decisions: failed to get channel history", "error", err, "channel", channelID)
		return d.slack.PostWebhook(responseURL, []slack.Block{
			slack.NewSectionBlock(
				slack.NewTextBlockObject("mrkdwn",
					fmt.Sprintf("❌ I couldn't read <#%s>. Make sure Signal is a member of that channel (`/invite @Signal`).", channelID),
					false, false,
				),
				nil, nil,
			),
		}, "Cannot Read Channel")
	}

	// Filter to daysBack
	cutoff := time.Now().AddDate(0, 0, -daysBack)
	var recentTexts []string
	for _, msg := range messages {
		if msg.BotID != "" {
			continue
		}
		// Parse Slack timestamp
		var ts float64
		fmt.Sscanf(msg.Timestamp, "%f", &ts)
		msgTime := time.Unix(int64(ts), 0)
		if msgTime.After(cutoff) && strings.TrimSpace(msg.Text) != "" {
			recentTexts = append(recentTexts, msg.Text)
		}
	}

	if len(recentTexts) == 0 {
		return d.slack.PostWebhook(responseURL, []slack.Block{
			slack.NewSectionBlock(
				slack.NewTextBlockObject("mrkdwn",
					fmt.Sprintf("No messages found in <#%s> in the last %d days.", channelID, daysBack),
					false, false,
				),
				nil, nil,
			),
		}, "No Messages")
	}

	// Use AI to extract decisions
	summary, err := d.ai.ExtractDecisions(ctx, recentTexts, channelID, daysBack)
	if err != nil {
		slog.Error("decisions: AI extraction failed", "error", err)
		return d.slack.PostWebhook(responseURL, []slack.Block{
			slack.NewSectionBlock(
				slack.NewTextBlockObject("mrkdwn", "❌ AI decision extraction failed. Please try again.", false, false),
				nil, nil,
			),
		}, "AI Error")
	}

	blocks := []slack.Block{
		slack.NewHeaderBlock(
			slack.NewTextBlockObject("plain_text", "✅ Decision Log", true, false),
		),
		slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn",
				fmt.Sprintf("*Channel:* <#%s> • *Last %d days* • *%d messages scanned*", channelID, daysBack, len(recentTexts)),
				false, false,
			),
			nil, nil,
		),
		slack.NewDividerBlock(),
		slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn", summary, false, false),
			nil, nil,
		),
		slack.NewContextBlock("decisions_footer",
			slack.NewTextBlockObject("mrkdwn", "_Use `/decisions #channel 14` to scan the last 14 days_", false, false),
		),
	}

	return d.slack.PostWebhook(responseURL, blocks, "Decision Log")
}
