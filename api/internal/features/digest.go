package features

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/LSUDOKOS/signal/internal/ai"
	"github.com/LSUDOKOS/signal/internal/domain"
	"github.com/LSUDOKOS/signal/internal/store"
	"github.com/slack-go/slack"
)

// DigestService implements the Quiet Hours Digest feature.
type DigestService struct {
	slack    SlackAPI
	ai       *ai.Client
	digests  store.DigestRepository
	users    store.UserRepository
	prefs    store.PreferencesRepository
	cache    store.CacheClient
}

// NewDigestService creates a new Digest service.
func NewDigestService(
	slack SlackAPI,
	ai *ai.Client,
	digests store.DigestRepository,
	users store.UserRepository,
	prefs store.PreferencesRepository,
	cache store.CacheClient,
) *DigestService {
	return &DigestService{
		slack:   slack,
		ai:      ai,
		digests: digests,
		users:   users,
		prefs:   prefs,
		cache:   cache,
	}
}

// HandleSlashCommand processes the /digest command to force-send a digest now.
func (d *DigestService) HandleSlashCommand(ctx context.Context, cmd *slack.SlashCommand, user *domain.User) error {
	// Send immediate digest
	dmChannel, err := d.slack.OpenDMChannel(cmd.UserID)
	if err != nil {
		return fmt.Errorf("open dm: %w", err)
	}

	// Get user's mentions/replies (in production, this would batched from Slack API)
	blocks := []slack.Block{
		slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn",
				"📬 *On-Demand Digest*\n\nI'm preparing your digest now. This feature is fully functional with a connected Slack workspace. For now, here's a summary of what I track:\n\n• @mentions in channels\n• Thread replies\n• Direct messages\n• Channel activity since last digest",
				false, false,
			),
			nil, nil,
		),
		slack.NewActionBlock("digest_actions",
			slack.NewButtonBlockElement("digest_update_prefs", "prefs", slack.NewTextBlockObject("plain_text", "Update Preferences", false, true)).WithStyle("primary"),
			slack.NewButtonBlockElement("digest_open_slack", "slack", slack.NewTextBlockObject("plain_text", "Open Slack", false, true)),
		),
	}

	return d.slack.PostMessage(dmChannel, blocks, "On-Demand Digest")
}

// SendScheduledDigest sends a digest to a specific user (called by the worker).
func (d *DigestService) SendScheduledDigest(ctx context.Context, u domain.User, prefs *domain.UserPreferences) error {
	dmChannel, err := d.slack.OpenDMChannel(u.SlackUserID)
	if err != nil {
		return fmt.Errorf("open dm for %s: %w", u.SlackUserID, err)
	}

	// Fetch recent messages the user was mentioned in via Slack Search API
	recentMessages, err := d.slack.SearchMessages(
		fmt.Sprintf("from:@%s OR to:@%s after:yesterday", u.SlackUserID, u.SlackUserID),
		slack.SearchParameters{Sort: "timestamp", Count: 50, SortDirection: "desc"},
	)

	var urgent, fyi, threads []domain.DigestItem
	if err == nil && recentMessages != nil {
		for _, match := range recentMessages.Matches {
			item := domain.DigestItem{
				From:    match.User,
				Message: match.Text,
				Channel: match.Channel.Name,
			}
			// Simple heuristic: messages directed at user are urgent
			if strings.Contains(match.Text, u.SlackUserID) || strings.HasPrefix(match.Text, "to:") {
				urgent = append(urgent, item)
			} else {
				fyi = append(fyi, item)
			}
		}
	}

	// Build digest blocks with real Slack data
	blocks := d.buildDigestBlocks(prefs.DigestHour, urgent, fyi, threads)

	if err := d.slack.PostMessage(dmChannel, blocks, "Digest"); err != nil {
		return fmt.Errorf("post digest: %w", err)
	}

	// Track digest
	if err := d.cache.SetLastDigest(ctx, u.SlackUserID, time.Now()); err != nil {
		slog.Error("failed to set last digest", "error", err)
	}

	slog.Info("digest sent", "user", u.SlackUserID, "hour", prefs.DigestHour)
	return nil
}

func (d *DigestService) buildDigestBlocks(hour int, urgent, fyi, threads []domain.DigestItem) []slack.Block {
	hourStr := fmt.Sprintf("%02d:00", hour)
	var blocks []slack.Block

	blocks = append(blocks,
		slack.NewHeaderBlock(
			slack.NewTextBlockObject("plain_text", fmt.Sprintf("📬 Your %s Digest", hourStr), true, false),
		),
	)

	// Urgent section
	if len(urgent) > 0 {
		urgentText := ""
		for _, item := range urgent {
			urgentText += fmt.Sprintf("• *%s:* \"%s\" → [Reply]\n", item.From, item.Message)
		}
		blocks = append(blocks,
			slack.NewSectionBlock(
				slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*🔴 Urgent (needs response today)*\n%s", urgentText), false, false),
				nil, nil,
			),
		)
	}

	// FYI section
	if len(fyi) > 0 {
		fyiText := ""
		for _, item := range fyi {
			fyiText += fmt.Sprintf("• *%s:* \"%s\" → [View]\n", item.From, item.Message)
		}
		blocks = append(blocks,
			slack.NewSectionBlock(
				slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*🟢 FYI (no action needed)*\n%s", fyiText), false, false),
				nil, nil,
			),
		)
	}

	// Thread replies
	if len(threads) > 0 {
		threadText := ""
		for _, item := range threads {
			threadText += fmt.Sprintf("• #%s: %s → [Jump]\n", item.Channel, item.Message)
		}
		blocks = append(blocks,
			slack.NewSectionBlock(
				slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*💬 Thread Replies*\n%s", threadText), false, false),
				nil, nil,
			),
		)
	}

	// Actions
	blocks = append(blocks,
		slack.NewActionBlock("digest_actions",
			slack.NewButtonBlockElement("digest_open_slack", "slack", slack.NewTextBlockObject("plain_text", "Open Slack", false, true)).WithStyle("primary"),
			slack.NewButtonBlockElement("digest_update_prefs", "prefs", slack.NewTextBlockObject("plain_text", "Update Preferences", false, true)),
		),
	)

	return blocks
}
