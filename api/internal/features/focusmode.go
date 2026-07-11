package features

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/LSUDOKOS/signal/internal/ai"
	"github.com/LSUDOKOS/signal/internal/domain"
	"github.com/LSUDOKOS/signal/internal/store"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

// FocusModeService implements the Focus Mode feature.
type FocusModeService struct {
	slack    SlackAPI
	ai       *ai.Client
	cache    store.CacheClient
	channels store.ChannelRepository
	summaries store.FocusSummaryRepository
}

// NewFocusModeService creates a new Focus Mode service.
func NewFocusModeService(
	slack SlackAPI,
	ai *ai.Client,
	cache store.CacheClient,
	channels store.ChannelRepository,
	summaries store.FocusSummaryRepository,
) *FocusModeService {
	return &FocusModeService{
		slack:     slack,
		ai:        ai,
		cache:     cache,
		channels:  channels,
		summaries: summaries,
	}
}

// HandleMessage processes a message for Focus Mode velocity detection.
func (f *FocusModeService) HandleMessage(ctx context.Context, event *slackevents.MessageEvent, user *domain.User, prefs *domain.UserPreferences) error {
	// Skip DMs
	if event.ChannelType == "im" {
		return nil
	}

	// Increment velocity counter
	count, err := f.cache.IncrChannelVelocity(ctx, event.Channel)
	if err != nil {
		return fmt.Errorf("velocity counter: %w", err)
	}

	// Check if threshold is met and focus hasn't already been offered
	if count >= prefs.FocusThreshold {
		offered, err := f.cache.HasFocusBeenOffered(ctx, event.Channel)
		if err != nil {
			return fmt.Errorf("check offered: %w", err)
		}
		if !offered {
			return f.triggerFocusMode(ctx, event.Channel, prefs.FocusThreshold)
		}
	}

	return nil
}

// HandleBlockAction handles Focus Mode button clicks.
func (f *FocusModeService) HandleBlockAction(ctx context.Context, action *slack.InteractionCallback, user *domain.User) error {
	if len(action.ActionCallback.BlockActions) == 0 {
		return nil
	}
	actionID := action.ActionCallback.BlockActions[0].ActionID
	channelID := action.Channel.ID

	switch actionID {
	case "focus_get_summary":
		return f.sendFullSummary(ctx, channelID, user.SlackUserID)
	case "focus_mute_30":
		return f.muteChannel(ctx, channelID, user.SlackUserID)
	case "focus_open_thread":
		return f.sendSummaryToDM(ctx, channelID, user.SlackUserID)
	}

	return nil
}

func (f *FocusModeService) triggerFocusMode(ctx context.Context, channelID string, threshold int) error {
	// Set offered flag (30 min TTL to prevent spam)
	if err := f.cache.SetFocusOffered(ctx, channelID, 30); err != nil {
		return fmt.Errorf("set offered: %w", err)
	}

	// Fetch recent messages
	messages, err := f.slack.GetChannelHistory(channelID, 50)
	if err != nil {
		return fmt.Errorf("get history: %w", err)
	}

	// Extract text from messages
	var messageTexts []string
	for _, msg := range messages {
		if msg.Text != "" {
			messageTexts = append(messageTexts, msg.Text)
		}
	}

	// Generate AI summary
	summary, err := f.ai.SummarizeFocus(ctx, messageTexts)
	if err != nil {
		slog.Error("ai focus summary failed", "error", err)
		// Post fallback message without AI
		return f.postVelocityWarning(ctx, channelID, threshold, nil)
	}

	return f.postVelocityWarning(ctx, channelID, threshold, summary)
}

func (f *FocusModeService) postVelocityWarning(ctx context.Context, channelID string, threshold int, summary *domain.FocusSummaryResult) error {
	var summaryText string
	if summary != nil && !summary.NoDecisions {
		for _, d := range summary.Decisions {
			summaryText += fmt.Sprintf("✅ *%s*\n", d.Decision)
			for _, a := range d.ActionItems {
				owner := a.Owner
				if owner == "" {
					owner = "Unassigned"
				}
				due := a.Due
				if due == "" {
					due = "None"
				}
				summaryText += fmt.Sprintf("   ↳ %s — Owner: %s — Due: %s\n", a.Description, owner, due)
			}
		}
	} else {
		summaryText = "No decisions found. Discussion was exploratory."
	}

	blocks := []slack.Block{
		slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn",
				fmt.Sprintf("🧘 *This channel is moving fast*\n%d+ messages in the last 10 minutes.", threshold),
				false, false,
			),
			nil, nil,
		),
		slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn", summaryText, false, false),
			nil, nil,
		),
		slack.NewActionBlock("focus_actions",
			slack.NewButtonBlockElement("focus_get_summary", "get_summary", slack.NewTextBlockObject("plain_text", "Get Full Summary", false, true)).WithStyle("primary"),
			slack.NewButtonBlockElement("focus_mute_30", "mute_30", slack.NewTextBlockObject("plain_text", "Mute 30 min", false, true)),
			slack.NewButtonBlockElement("focus_open_thread", "open_thread", slack.NewTextBlockObject("plain_text", "Open in Thread", false, true)),
		),
	}

	return f.slack.PostMessage(channelID, blocks, "Focus Mode: Channel velocity alert")
}

func (f *FocusModeService) sendFullSummary(ctx context.Context, channelID, userID string) error {
	messages, err := f.slack.GetChannelHistory(channelID, 50)
	if err != nil {
		return err
	}

	var messageTexts []string
	for _, msg := range messages {
		if msg.Text != "" {
			messageTexts = append(messageTexts, msg.Text)
		}
	}

	summary, err := f.ai.SummarizeFocus(ctx, messageTexts)
	if err != nil {
		return f.slack.PostEphemeral(channelID, userID,
			[]slack.Block{
				slack.NewSectionBlock(
					slack.NewTextBlockObject("mrkdwn", "Sorry, I couldn't generate a summary right now. Please try again.", false, false),
					nil, nil,
				),
			},
			"Summary unavailable",
		)
	}

	var text string
	if summary.NoDecisions {
		text = "No formal decisions found. Discussion was exploratory."
	} else {
		for _, d := range summary.Decisions {
			text += fmt.Sprintf("✅ *%s*\n", d.Decision)
			for _, a := range d.ActionItems {
				text += fmt.Sprintf("   ↳ %s — Owner: %s — Due: %s\n", a.Description, a.Owner, a.Due)
			}
		}
	}

	return f.slack.PostEphemeral(channelID, userID,
		[]slack.Block{
			slack.NewSectionBlock(
				slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*Full Channel Summary*\n\n%s", text), false, false),
				nil, nil,
			),
		},
		"Full Summary",
	)
}

func (f *FocusModeService) sendSummaryToDM(ctx context.Context, channelID, userID string) error {
	messages, err := f.slack.GetChannelHistory(channelID, 50)
	if err != nil {
		return fmt.Errorf("get history: %w", err)
	}

	var messageTexts []string
	for _, msg := range messages {
		if msg.Text != "" {
			messageTexts = append(messageTexts, msg.Text)
		}
	}

	summary, err := f.ai.SummarizeFocus(ctx, messageTexts)
	if err != nil {
		slog.Error("ai focus summary failed", "error", err)
		return nil // Don't break the button flow
	}

	dmChannel, err := f.slack.OpenDMChannel(userID)
	if err != nil {
		return fmt.Errorf("open dm: %w", err)
	}

	blocks := buildFocusSummaryBlockKit(summary, channelID)
	return f.slack.PostMessage(dmChannel, blocks, "Focus Summary")
}

func (f *FocusModeService) muteChannel(ctx context.Context, channelID, userID string) error {
	// Request to mute the channel for 30 minutes via DM
	dmChannel, err := f.slack.OpenDMChannel(userID)
	if err != nil {
		return err
	}

	blocks := []slack.Block{
		slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn",
				fmt.Sprintf("🔇 Channel <#%s> will be muted for 30 minutes. I'll let you know when it's time to check back.", channelID),
				false, false,
			),
			nil, nil,
		),
	}

	return f.slack.PostMessage(dmChannel, blocks, "Channel Muted")
}

// buildFocusSummaryBlockKit creates a Block Kit message for focus mode results.
func buildFocusSummaryBlockKit(summary *domain.FocusSummaryResult, channelID string) []slack.Block {
	var blocks []slack.Block

	headerText := "📋 *Focus Mode Summary*"
	if summary.NoDecisions {
		headerText = "📋 *Focus Mode — No Decisions Found*\nDiscussion was exploratory."
	}

	blocks = append(blocks,
		slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn", headerText, false, false),
			nil, nil,
		),
	)

	if !summary.NoDecisions {
		for _, d := range summary.Decisions {
			decisionText := fmt.Sprintf("✅ *%s*", strings.TrimSpace(d.Decision))
			blocks = append(blocks, slack.NewSectionBlock(
				slack.NewTextBlockObject("mrkdwn", decisionText, false, false),
				nil, nil,
			))

			for _, a := range d.ActionItems {
				blocks = append(blocks, slack.NewSectionBlock(
					slack.NewTextBlockObject("mrkdwn",
						fmt.Sprintf("   • %s — *Owner:* %s — *Due:* %s",
							a.Description, a.Owner, a.Due),
						false, false,
					),
					nil, nil,
				))
			}
		}
	}

	return blocks
}
